package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// SessionCookie is the session cookie's name over plain HTTP; with Secure it is SecureSessionCookie, whose
	// __Host- prefix makes browsers refuse it unless it is Secure, host-only and on Path=/.
	SessionCookie       = "resourcerer_session"
	SecureSessionCookie = "__Host-" + SessionCookie
	stateCookie         = "resourcerer_login"
	stateTTL            = 10 * time.Minute
	// maxCookie keeps a session cookie under the 4 KiB browsers accept, with room for its attributes.
	maxCookie = 3800
	// ConfigRefresh bounds how long a change to the ConfigMap takes to apply.
	ConfigRefresh = 30 * time.Second
	// discoveryRetry is how long a failed discovery is remembered, so sign-ins during an IdP outage fail fast
	// instead of each waiting on the IdP.
	discoveryRetry   = 30 * time.Second
	discoveryTimeout = 10 * time.Second
)

type OIDCOptions struct {
	// Reader reads the ConfigMap directly from the API server, so the first read works before caches start.
	Reader             client.Reader
	ConfigMap          types.NamespacedName
	ClientSecret       string
	SessionKey         string
	PreviousSessionKey string
	CookieSecure       CookieSecure
}

var logger = log.Log.WithName("auth")

// OIDC signs users in at an OpenID Connect provider with the authorization code flow and PKCE, and keeps the
// identity in a sealed cookie. Tokens from the IdP never reach the browser and are not kept: at the end of the
// session the user signs in again, usually without a prompt while the IdP's own session lasts.
//
// Cookies are SameSite=Lax, so a cross-site form cannot carry the session into a mutation. Mutations must also carry
// the header the UI sends and a body declared as JSON, which closes what Lax leaves open: a page on a sibling origin
// posting a JSON body as text/plain, since the API's decoders do not look at Content-Type. The server skips that
// check for a strategy that issues no cookie, because there is no session there for another page to borrow.
type OIDC struct {
	config       configSource
	codec        *Codec
	clientSecret string
	secure       CookieSecure
	now          func() time.Time

	mu        sync.Mutex
	providers map[string]*oidc.Provider
	failed    map[string]discoveryFailure
	// discover serializes discovery, so concurrent sign-ins during an outage wait for one attempt, not one each.
	discover sync.Mutex
	insecure sync.Once
}

type discoveryFailure struct {
	at  time.Time
	err error
}

// NewOIDC reads the configuration once and fails when it cannot, so a broken install refuses to serve.
func NewOIDC(ctx context.Context, o OIDCOptions) (*OIDC, error) {
	if o.ClientSecret == "" {
		return nil, errors.New("the OIDC client secret is empty")
	}
	codec, err := NewCodec(o.SessionKey, o.PreviousSessionKey)
	if err != nil {
		return nil, fmt.Errorf("session key: %w", err)
	}
	if o.CookieSecure == "" {
		o.CookieSecure = CookieSecureAuto
	}
	p := &OIDC{config: configSource{reader: o.Reader, key: o.ConfigMap}, codec: codec, clientSecret: o.ClientSecret,
		secure: o.CookieSecure, now: time.Now, providers: map[string]*oidc.Provider{}, failed: map[string]discoveryFailure{}}
	if err := p.config.load(ctx); err != nil {
		return nil, err
	}
	return p, nil
}

// Run rereads the configuration until ctx ends.
func (p *OIDC) Run(ctx context.Context) error {
	t := time.NewTicker(ConfigRefresh)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := p.config.load(ctx); err != nil {
				logger.Error(err, "reload OIDC configuration; keeping the previous one")
			}
		}
	}
}

func (p *OIDC) Strategy() string { return StrategyOIDC }

// The session lives in a cookie, so the server guards mutations against another page borrowing it.
func (p *OIDC) UsesSessionCookie() bool { return true }

func (p *OIDC) Status() Status {
	st := p.config.current()
	return Status{ConfigStale: st.stale, ConfigError: st.err}
}

type session struct {
	User   string   `json:"u"`
	Groups []string `json:"g,omitempty"`
	// Issuer and ClientID end the session when the configuration points elsewhere.
	Issuer   string `json:"iss"`
	ClientID string `json:"aud"`
	Expires  int64  `json:"exp"`
}

func (p *OIDC) Identify(r *http.Request) (Identity, bool) {
	c, err := r.Cookie(p.name(r, SessionCookie))
	if err != nil {
		return Identity{}, false
	}
	var s session
	if p.codec.Open(SessionCookie, c.Value, p.now(), &s) != nil {
		return Identity{}, false
	}
	cfg := p.config.current().cfg
	if s.Issuer != cfg.Issuer || s.ClientID != cfg.ClientID || s.User == "" {
		return Identity{}, false
	}
	viewer, editor := cfg.Access(s.Groups)
	return Identity{User: s.User, Groups: s.Groups, Viewer: viewer, Editor: editor, ExpiresAt: time.Unix(s.Expires, 0)}, true
}

func (p *OIDC) Handle(w http.ResponseWriter, r *http.Request) bool {
	var h http.HandlerFunc
	switch r.URL.Path {
	case "/auth/login":
		h = p.login
	case "/auth/callback":
		h = p.callback
	case "/auth/logout":
		h = p.logout
	default:
		return false
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return true
	}
	h(w, r)
	return true
}

type loginState struct {
	State    string `json:"s"`
	Nonce    string `json:"n"`
	Verifier string `json:"v"`
	Redirect string `json:"r"`
}

func (p *OIDC) login(w http.ResponseWriter, r *http.Request) {
	cfg := p.config.current().cfg
	oc, _, err := p.client(r, cfg)
	if err != nil {
		logger.Error(err, "discover OIDC provider", "issuer", cfg.Issuer)
		http.Error(w, "the identity provider is unavailable", http.StatusServiceUnavailable)
		return
	}
	st := loginState{State: random(), Nonce: random(), Verifier: oauth2.GenerateVerifier(), Redirect: SafeRedirect(r.URL.Query().Get("redirect"))}
	sealed, err := p.codec.Seal(stateCookie, st, p.now().Add(stateTTL))
	if err != nil {
		http.Error(w, "start sign-in: "+err.Error(), http.StatusInternalServerError)
		return
	}
	p.setCookie(w, r, stateCookie, sealed, int(stateTTL.Seconds()))
	http.Redirect(w, r, oc.AuthCodeURL(st.State, oauth2.S256ChallengeOption(st.Verifier), oidc.Nonce(st.Nonce)), http.StatusFound)
}

func (p *OIDC) callback(w http.ResponseWriter, r *http.Request) {
	p.setCookie(w, r, stateCookie, "", -1)
	var st loginState
	c, err := r.Cookie(p.name(r, stateCookie))
	if err != nil || p.codec.Open(stateCookie, c.Value, p.now(), &st) != nil {
		http.Error(w, "the sign-in expired or was started elsewhere; sign in again", http.StatusBadRequest)
		return
	}
	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		http.Error(w, "the identity provider refused the sign-in: "+e+" "+q.Get("error_description"), http.StatusUnauthorized)
		return
	}
	if subtle.ConstantTimeCompare([]byte(q.Get("state")), []byte(st.State)) != 1 {
		http.Error(w, "the sign-in state does not match; sign in again", http.StatusBadRequest)
		return
	}
	cfg := p.config.current().cfg
	s, err := p.exchange(r, cfg, q.Get("code"), st)
	if err != nil {
		logger.Error(err, "complete OIDC sign-in")
		http.Error(w, "sign-in failed: "+err.Error(), http.StatusUnauthorized)
		return
	}
	if viewer, editor := cfg.Access(s.Groups); !viewer && !editor {
		http.Error(w, s.User+" is not allowed to use resourcerer", http.StatusForbidden)
		return
	}
	expires := time.Unix(s.Expires, 0)
	sealed, err := p.codec.Seal(SessionCookie, s, expires)
	if err == nil && len(sealed) > maxCookie {
		s.Groups = cfg.referenced(s.Groups)
		sealed, err = p.codec.Seal(SessionCookie, s, expires)
	}
	if err != nil {
		http.Error(w, "create session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	p.setCookie(w, r, SessionCookie, sealed, int(cfg.TTL().Seconds()))
	http.Redirect(w, r, SafeRedirect(st.Redirect), http.StatusFound)
}

func (p *OIDC) exchange(r *http.Request, cfg Config, code string, st loginState) (session, error) {
	oc, idp, err := p.client(r, cfg)
	if err != nil {
		return session{}, fmt.Errorf("discover provider: %w", err)
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	tok, err := oc.Exchange(ctx, code, oauth2.VerifierOption(st.Verifier))
	if err != nil {
		return session{}, fmt.Errorf("exchange code: %w", err)
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok {
		return session{}, errors.New("the token response has no id_token")
	}
	idt, err := idp.Verifier(&oidc.Config{ClientID: cfg.ClientID, Now: p.now}).Verify(ctx, raw)
	if err != nil {
		return session{}, fmt.Errorf("verify ID token: %w", err)
	}
	if subtle.ConstantTimeCompare([]byte(idt.Nonce), []byte(st.Nonce)) != 1 {
		return session{}, errors.New("the ID token nonce does not match")
	}
	var claims map[string]any
	if err := idt.Claims(&claims); err != nil {
		return session{}, fmt.Errorf("read claims: %w", err)
	}
	user, _ := claims[cfg.UsernameClaim].(string)
	if user == "" {
		return session{}, fmt.Errorf("the ID token has no %s claim", cfg.UsernameClaim)
	}
	return session{User: user, Groups: stringList(claims[cfg.GroupsClaim]), Issuer: cfg.Issuer, ClientID: cfg.ClientID,
		Expires: p.now().Add(cfg.TTL()).Unix()}, nil
}

func (p *OIDC) logout(w http.ResponseWriter, r *http.Request) {
	p.setCookie(w, r, SessionCookie, "", -1)
	http.Redirect(w, r, "/", http.StatusFound)
}

// client discovers the issuer once and builds the OAuth2 client. A failed discovery is retried only after
// discoveryRetry.
func (p *OIDC) client(r *http.Request, cfg Config) (*oauth2.Config, *oidc.Provider, error) {
	idp, err := p.discovered(r.Context(), cfg.Issuer)
	if err != nil {
		return nil, nil, err
	}
	redirect := cfg.RedirectURL
	if redirect == "" {
		scheme := "http"
		if p.secure.For(r) {
			scheme = "https"
		}
		redirect = scheme + "://" + r.Host + "/auth/callback"
	}
	return &oauth2.Config{ClientID: cfg.ClientID, ClientSecret: p.clientSecret, Endpoint: idp.Endpoint(),
		RedirectURL: redirect, Scopes: cfg.Scopes}, idp, nil
}

func (p *OIDC) discovered(ctx context.Context, issuer string) (*oidc.Provider, error) {
	lookup := func() (*oidc.Provider, bool, error) {
		p.mu.Lock()
		defer p.mu.Unlock()
		if idp, ok := p.providers[issuer]; ok {
			return idp, true, nil
		}
		if f, ok := p.failed[issuer]; ok && p.now().Sub(f.at) < discoveryRetry {
			return nil, true, f.err
		}
		return nil, false, nil
	}
	if idp, ok, err := lookup(); ok {
		return idp, err
	}
	p.discover.Lock()
	defer p.discover.Unlock()
	if idp, ok, err := lookup(); ok {
		return idp, err
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), discoveryTimeout)
	defer cancel()
	idp, err := oidc.NewProvider(ctx, issuer)
	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		p.failed[issuer] = discoveryFailure{at: p.now(), err: err}
		return nil, err
	}
	delete(p.failed, issuer)
	p.providers[issuer] = idp
	return idp, nil
}

// name is the cookie's name for this request: __Host- prefixed when the cookie is Secure.
func (p *OIDC) name(r *http.Request, base string) string {
	if p.secure.For(r) {
		return "__Host-" + base
	}
	return base
}

// setCookie sets a host-only cookie on Path=/, as the __Host- prefix requires.
func (p *OIDC) setCookie(w http.ResponseWriter, r *http.Request, base, value string, maxAge int) {
	secure := p.secure.For(r)
	if !secure && maxAge > 0 {
		p.insecure.Do(func() {
			logger.Info("issuing auth cookies without Secure because the request is not HTTPS; serve resourcerer over " +
				"TLS, and set --auth-cookie-secure=true behind a TLS-terminating proxy")
		})
	}
	http.SetCookie(w, &http.Cookie{Name: p.name(r, base), Value: value, Path: "/", MaxAge: maxAge, HttpOnly: true,
		Secure: secure, SameSite: http.SameSiteLaxMode})
}

// SafeRedirect keeps a post-login destination on this site: a path on this host, never another one. It checks the
// path as http.Redirect will send it, cleaned, because a cleaned "/./\host" is "/\host", which browsers follow as
// "//host".
func SafeRedirect(s string) string {
	if s == "" || strings.HasPrefix(s, "//") || strings.ContainsRune(s, '\\') || hasControl(s) {
		return "/"
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme != "" || u.Host != "" || u.Opaque != "" || u.User != nil ||
		strings.ContainsRune(u.Path, '\\') || hasControl(u.Path) {
		return "/"
	}
	clean := path.Clean(u.Path)
	if !strings.HasPrefix(clean, "/") || strings.HasPrefix(clean, "//") {
		return "/"
	}
	if strings.HasSuffix(u.Path, "/") && clean != "/" {
		clean += "/"
	}
	out := (&url.URL{Path: clean, RawQuery: u.RawQuery, Fragment: u.Fragment}).String()
	if !strings.HasPrefix(out, "/") || strings.HasPrefix(out, "//") {
		return "/"
	}
	return out
}

func hasControl(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f })
}

func stringList(v any) []string {
	switch t := v.(type) {
	case string:
		return []string{t}
	case []any:
		out := make([]string, 0, len(t))
		for _, x := range t {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func random() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
