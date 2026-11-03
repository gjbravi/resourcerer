package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

// ConfigKey is the key of the OIDC configuration in its ConfigMap.
const ConfigKey = "oidc.yaml"

const (
	defaultSessionTTL = 12 * time.Hour
	minSessionTTL     = 5 * time.Minute
	maxSessionTTL     = 7 * 24 * time.Hour
)

// Config holds the OIDC parameters that are not secret. Secrets come from the environment.
type Config struct {
	Issuer   string `json:"issuer"`
	ClientID string `json:"clientID"`
	// RedirectURL is the callback registered at the IdP; empty derives it from the request's host.
	RedirectURL   string   `json:"redirectURL,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	UsernameClaim string   `json:"usernameClaim,omitempty"`
	GroupsClaim   string   `json:"groupsClaim,omitempty"`
	// AllowedGroups restricts access to members of these groups; empty admits anyone the IdP authenticates.
	AllowedGroups []string `json:"allowedGroups,omitempty"`
	Roles         Roles    `json:"roles"`
	SessionTTL    string   `json:"sessionTTL,omitempty"`

	ttl time.Duration
}

type Roles struct {
	EditorGroups []string `json:"editorGroups,omitempty"`
	// AnyAuthenticatedIsViewer lets every admitted user read; defaults to true.
	AnyAuthenticatedIsViewer *bool `json:"anyAuthenticatedIsViewer,omitempty"`
}

func (c Config) TTL() time.Duration { return c.ttl }

var parsed sync.Map

// ParseConfig reads oidc.yaml strictly, so a misspelled key fails instead of silently weakening access, and caches
// the result by content. Keys bind as JSON tags do, case aside, so EditorGroups parses as editorGroups: an unknown key
// is refused, a differently capitalised one is not.
func ParseConfig(data string) (Config, error) {
	if v, ok := parsed.Load(data); ok {
		return v.(Config), nil
	}
	var c Config
	if err := yaml.UnmarshalStrict([]byte(data), &c); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", ConfigKey, err)
	}
	if err := c.complete(); err != nil {
		return Config{}, fmt.Errorf("%s: %w", ConfigKey, err)
	}
	parsed.Store(data, c)
	return c, nil
}

func (c *Config) complete() error {
	if err := httpURL("issuer", c.Issuer); err != nil {
		return err
	}
	if c.ClientID == "" {
		return errors.New("clientID is required")
	}
	if c.RedirectURL != "" {
		if err := httpURL("redirectURL", c.RedirectURL); err != nil {
			return err
		}
	}
	if len(c.Scopes) == 0 {
		c.Scopes = []string{"openid", "profile", "email"}
	}
	if !slices.Contains(c.Scopes, "openid") {
		return errors.New("scopes must include openid")
	}
	if c.UsernameClaim == "" {
		c.UsernameClaim = "preferred_username"
	}
	if c.GroupsClaim == "" {
		c.GroupsClaim = "groups"
	}
	c.ttl = defaultSessionTTL
	if c.SessionTTL != "" {
		d, err := time.ParseDuration(c.SessionTTL)
		if err != nil || d < minSessionTTL || d > maxSessionTTL {
			return fmt.Errorf("sessionTTL must be a duration between %s and %s", minSessionTTL, maxSessionTTL)
		}
		c.ttl = d
	}
	return nil
}

// httpURL is the server's own gate on the two URLs of the OIDC configuration: an http(s) URL whose authority is a real
// host, carrying neither credentials nor a character a browser would resolve somewhere else. values.schema.json checks
// the same shapes earlier and is the stricter side of the pair, which is the rule chart_test.go enforces: the schema
// admits nothing this refuses. That only covers what helm renders with validation on. The flag is off the path
// altogether for a ConfigMap an operator hand-edits or writes with kubectl, and --skip-schema-validation turns it off
// by choice, so this is what reads those values. url.Parse alone would not hold the line: it takes userinfo, and it
// accepts a backslash or a control character in places where a browser reads a different address.
func httpURL(field, raw string) error {
	if strings.ContainsRune(raw, '\\') || hasControl(raw) {
		return fmt.Errorf("%s must be an http(s) URL with a host: it contains a backslash or a control character", field)
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("%s must be an http(s) URL with a host", field)
	}
	// The client secret goes to the token endpoint in the request body, so credentials in the issuer would only ever
	// be a way to point the server at a host someone else named.
	if u.User != nil {
		return fmt.Errorf("%s must be an http(s) URL with a host: it must not carry credentials", field)
	}
	return nil
}

// configSource reads the configuration from its ConfigMap. The first read must succeed: authentication fails
// closed and never falls back to anonymous. Later failures keep the last good configuration and report it stale.
type configSource struct {
	reader client.Reader
	key    types.NamespacedName
	// timeout bounds each read, so a hung API call marks the configuration stale instead of stalling reloads.
	timeout time.Duration
	state   atomic.Pointer[configState]
}

const configReadTimeout = 10 * time.Second

type configState struct {
	cfg   Config
	stale bool
	err   string
}

func (s *configSource) load(ctx context.Context) error {
	timeout := s.timeout
	if timeout == 0 {
		timeout = configReadTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cfg, err := s.read(ctx)
	prev := s.state.Load()
	switch {
	case err == nil:
		s.state.Store(&configState{cfg: cfg})
	case prev == nil:
		return err
	default:
		s.state.Store(&configState{cfg: prev.cfg, stale: true, err: err.Error()})
	}
	return err
}

func (s *configSource) read(ctx context.Context) (Config, error) {
	var cm corev1.ConfigMap
	if err := s.reader.Get(ctx, s.key, &cm); err != nil {
		return Config{}, fmt.Errorf("read ConfigMap %s: %w", s.key, err)
	}
	data, ok := cm.Data[ConfigKey]
	if !ok {
		return Config{}, fmt.Errorf("ConfigMap %s has no %s key", s.key, ConfigKey)
	}
	return ParseConfig(data)
}

func (s *configSource) current() *configState { return s.state.Load() }
