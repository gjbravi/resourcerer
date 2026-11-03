package auth

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const validConfig = `
issuer: https://sso.example.com/realms/example
clientID: resourcerer
scopes: [openid, profile, email, groups]
usernameClaim: preferred_username
groupsClaim: groups
allowedGroups: [platform]
roles:
  editorGroups: [platform]
  anyAuthenticatedIsViewer: true
sessionTTL: 8h
`

func TestParseConfig(t *testing.T) {
	c, err := ParseConfig(validConfig)
	if err != nil {
		t.Fatal(err)
	}
	if c.TTL() != 8*time.Hour || c.ClientID != "resourcerer" || !slices.Equal(c.Roles.EditorGroups, []string{"platform"}) {
		t.Errorf("parsed %+v", c)
	}

	d, err := ParseConfig("issuer: http://idp.example.com\nclientID: rz\n")
	if err != nil {
		t.Fatal(err)
	}
	if d.TTL() != 12*time.Hour || d.UsernameClaim != "preferred_username" || d.GroupsClaim != "groups" ||
		!slices.Contains(d.Scopes, "openid") || !d.Roles.anyViewer() {
		t.Errorf("defaults = %+v", d)
	}

	for name, data := range map[string]string{
		"unknown key":         validConfig + "allowedGroup: [everyone]\n",
		"unknown nested key":  "issuer: https://idp.example.com\nclientID: rz\nroles:\n  editors: [a]\n",
		"no issuer":           "clientID: rz\n",
		"issuer not a URL":    "issuer: idp.example.com\nclientID: rz\n",
		"issuer without host": "issuer: https://\nclientID: rz\n",
		"issuer not http(s)":  "issuer: ssh://idp.example.com\nclientID: rz\n",
		"no client":           "issuer: https://idp.example.com\n",
		"no openid scope":     "issuer: https://idp.example.com\nclientID: rz\nscopes: [profile]\n",
		"bad ttl":             "issuer: https://idp.example.com\nclientID: rz\nsessionTTL: forever\n",
		"ttl too short":       "issuer: https://idp.example.com\nclientID: rz\nsessionTTL: 1m\n",
		"ttl too long":        "issuer: https://idp.example.com\nclientID: rz\nsessionTTL: 200h\n",
		"relative redirect":   "issuer: https://idp.example.com\nclientID: rz\nredirectURL: /auth/callback\n",
		"redirect no host":    "issuer: https://idp.example.com\nclientID: rz\nredirectURL: https://\n",
		"redirect no scheme":  "issuer: https://idp.example.com\nclientID: rz\nredirectURL: //cb.example.com/auth/callback\n",
		"redirect not http":   "issuer: https://idp.example.com\nclientID: rz\nredirectURL: ftp://cb.example.com/auth/callback\n",
		"redirect whitespace": "issuer: https://idp.example.com\nclientID: rz\nredirectURL: https://cb.example.com /x\n",
		"not yaml":            "issuer: [",
	} {
		if _, err := ParseConfig(data); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}

	for name, data := range map[string]string{
		"issuer with a port":   "issuer: https://idp.example.com:8443\nclientID: rz\n",
		"issuer trailing path": "issuer: https://idp.example.com/realms/example\nclientID: rz\n",
		"plain http":           "issuer: http://localhost:8080/realms/x\nclientID: rz\n",
		"redirect with port":   "issuer: https://idp.example.com\nclientID: rz\nredirectURL: http://localhost:8080/auth/callback\n",
		"ttl boundaries":       "issuer: https://idp.example.com\nclientID: rz\nsessionTTL: 5m\nroles: {}\n",
	} {
		if _, err := ParseConfig(data); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// reader serves one ConfigMap, or fails.
type reader struct {
	client.Reader
	cm   *corev1.ConfigMap
	err  error
	hang bool
}

func (r *reader) Get(ctx context.Context, key types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
	if r.hang {
		<-ctx.Done()
		return ctx.Err()
	}
	if r.err != nil {
		return r.err
	}
	if r.cm == nil || key.Name != r.cm.Name || key.Namespace != r.cm.Namespace {
		return errors.New("not found")
	}
	*obj.(*corev1.ConfigMap) = *r.cm.DeepCopy()
	return nil
}

var configKey = types.NamespacedName{Namespace: "resourcerer", Name: "resourcerer-auth"}

func configMap(data string) *corev1.ConfigMap {
	return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: configKey.Namespace, Name: configKey.Name},
		Data: map[string]string{ConfigKey: data}}
}

func TestNewOIDCFailsClosed(t *testing.T) {
	key := newKey(t)
	ok := OIDCOptions{Reader: &reader{cm: configMap(validConfig)}, ConfigMap: configKey, ClientSecret: "s", SessionKey: key}
	if _, err := NewOIDC(context.Background(), ok); err != nil {
		t.Fatalf("valid options: %v", err)
	}
	for name, mutate := range map[string]func(*OIDCOptions){
		"ConfigMap missing":    func(o *OIDCOptions) { o.Reader = &reader{} },
		"ConfigMap unreadable": func(o *OIDCOptions) { o.Reader = &reader{err: errors.New("forbidden")} },
		"key missing": func(o *OIDCOptions) {
			cm := configMap("")
			cm.Data = map[string]string{"other.yaml": validConfig}
			o.Reader = &reader{cm: cm}
		},
		"malformed":         func(o *OIDCOptions) { o.Reader = &reader{cm: configMap("issuer: [")} },
		"unknown key":       func(o *OIDCOptions) { o.Reader = &reader{cm: configMap(validConfig + "typo: 1\n")} },
		"no client secret":  func(o *OIDCOptions) { o.ClientSecret = "" },
		"no session key":    func(o *OIDCOptions) { o.SessionKey = "" },
		"short session key": func(o *OIDCOptions) { o.SessionKey = "c2hvcnQ=" },
	} {
		o := ok
		mutate(&o)
		if p, err := NewOIDC(context.Background(), o); err == nil || p != nil {
			t.Errorf("%s: started", name)
		}
	}
}

func TestConfigKeepsLastKnownGood(t *testing.T) {
	r := &reader{cm: configMap(validConfig)}
	s := configSource{reader: r, key: configKey}
	ctx := context.Background()
	if err := s.load(ctx); err != nil {
		t.Fatal(err)
	}

	r.err = errors.New("connection refused")
	if err := s.load(ctx); err == nil {
		t.Fatal("failed read reported no error")
	}
	if st := s.current(); !st.stale || !strings.Contains(st.err, "connection refused") || st.cfg.ClientID != "resourcerer" {
		t.Errorf("after a failed read = %+v, want the previous configuration, stale", st)
	}

	r.err, r.cm = nil, configMap(validConfig+"typo: 1\n")
	_ = s.load(ctx)
	if st := s.current(); !st.stale || st.cfg.ClientID != "resourcerer" {
		t.Errorf("after an invalid update = %+v, want the previous configuration, stale", st)
	}

	r.cm = configMap(strings.Replace(validConfig, "clientID: resourcerer", "clientID: renamed", 1))
	if err := s.load(ctx); err != nil {
		t.Fatal(err)
	}
	if st := s.current(); st.stale || st.err != "" || st.cfg.ClientID != "renamed" {
		t.Errorf("after recovering = %+v, want the new configuration", st)
	}
}

func TestConfigReadTimesOut(t *testing.T) {
	r := &reader{cm: configMap(validConfig)}
	s := configSource{reader: r, key: configKey, timeout: 20 * time.Millisecond}
	if err := s.load(context.Background()); err != nil {
		t.Fatal(err)
	}
	r.hang = true
	if err := s.load(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("hung read = %v, want a deadline error", err)
	}
	if st := s.current(); !st.stale || st.cfg.ClientID != "resourcerer" {
		t.Errorf("after a hung read = %+v, want the previous configuration, stale", st)
	}
}
