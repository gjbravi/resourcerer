package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
)

func TestSafeRedirect(t *testing.T) {
	for in, want := range map[string]string{
		"":                    "/",
		"/":                   "/",
		"/workloads/demo/api": "/workloads/demo/api",
		"/workloads/":         "/workloads/",
		"/w?range=1d":         "/w?range=1d",
		"/w#impact":           "/w#impact",
		"/?x=//evil.com":      "/?x=//evil.com",
		// The marker sign-out puts on the address the browser lands on has to survive here too, or signing in again
		// from the signed-out screen would lose the user where they were.
		"/?signed-out":           "/?signed-out",
		"/w?range=1d&x=2":        "/w?range=1d&x=2",
		"/a/../b":                "/b",
		"https://evil.com":       "/",
		"//evil.com":             "/",
		"///evil.com":            "/",
		"/\\evil.com":            "/",
		"/./\\evil.com":          "/",
		"/a/../\\evil.com":       "/",
		"/%5Cevil.com":           "/",
		"/./%5Cevil.com":         "/",
		"workloads":              "/",
		"javascript:alert(1)":    "/",
		"/ok\r\nSet-Cookie: x=y": "/",
		"/ok\tthere":             "/",
		"/%0d%0aSet-Cookie:x":    "/",
	} {
		got := SafeRedirect(in)
		if got != want {
			t.Errorf("SafeRedirect(%q) = %q, want %q", in, got, want)
		}
		w := httptest.NewRecorder()
		http.Redirect(w, httptest.NewRequest(http.MethodGet, "/auth/callback", nil), got, http.StatusFound)
		if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, "/") || strings.HasPrefix(loc, "//") ||
			strings.ContainsRune(loc, '\\') {
			t.Errorf("SafeRedirect(%q) sends the browser to %q", in, loc)
		}
	}
}

func TestCookieSecure(t *testing.T) {
	plain := httptest.NewRequest(http.MethodGet, "http://rz.example/", nil)
	proxied := httptest.NewRequest(http.MethodGet, "http://rz.example/", nil)
	proxied.Header.Set("X-Forwarded-Proto", "https")
	direct := httptest.NewRequest(http.MethodGet, "https://rz.example/", nil)
	direct.TLS = &tls.ConnectionState{}
	for _, tc := range []struct {
		mode CookieSecure
		r    *http.Request
		want bool
	}{
		{CookieSecureAuto, plain, false},
		{CookieSecureAuto, proxied, true},
		{CookieSecureAuto, direct, true},
		{CookieSecureTrue, plain, true},
		{CookieSecureFalse, direct, false},
	} {
		if got := tc.mode.For(tc.r); got != tc.want {
			t.Errorf("%s with TLS %v, proto %q = %v", tc.mode, tc.r.TLS != nil, tc.r.Header.Get("X-Forwarded-Proto"), got)
		}
	}
	if _, err := ParseCookieSecure("yes"); err == nil {
		t.Error("ParseCookieSecure accepted yes")
	}
}

func TestIdentify(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	codec, _ := NewCodec(newKey(t), "")
	cfg, _ := ParseConfig(validConfig)
	p := &OIDC{codec: codec, now: func() time.Time { return now }}
	p.config.state.Store(&configState{cfg: cfg})
	valid := session{User: "alice", Groups: []string{"platform"}, Issuer: cfg.Issuer, ClientID: cfg.ClientID, Expires: now.Add(time.Hour).Unix()}

	request := func(s session, expires time.Time) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
		v, _ := codec.Seal(SessionCookie, s, expires)
		r.AddCookie(&http.Cookie{Name: SessionCookie, Value: v})
		return r
	}
	id, ok := p.Identify(request(valid, now.Add(time.Hour)))
	if !ok || id.User != "alice" || !id.Editor || !id.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("valid session = %+v, %v", id, ok)
	}

	otherIssuer, otherClient := valid, valid
	otherIssuer.Issuer = "https://attacker.example"
	otherClient.ClientID = "other"
	for name, r := range map[string]*http.Request{
		"no cookie":        httptest.NewRequest(http.MethodGet, "/", nil),
		"expired":          request(valid, now),
		"another issuer":   request(otherIssuer, now.Add(time.Hour)),
		"another clientID": request(otherClient, now.Add(time.Hour)),
	} {
		if _, ok := p.Identify(r); ok {
			t.Errorf("%s: identified", name)
		}
	}

	stricter := cfg
	stricter.Roles.EditorGroups = []string{"sre"}
	p.config.state.Store(&configState{cfg: stricter})
	if id, ok := p.Identify(request(valid, now.Add(time.Hour))); !ok || id.Editor || !id.Viewer {
		t.Errorf("after editorGroups changed = %+v, want a viewer without signing in again", id)
	}
}

// The browser drops a cookie it cannot take without saying so, which leaves a signed-in user with no session and no
// message. So a session that will not fit is refused out loud, after the groups no rule mentions are dropped.
func TestSealSessionFitsTheCookie(t *testing.T) {
	codec, _ := NewCodec(newKey(t), "")
	now := time.Unix(1_700_000_000, 0)
	exp := now.Add(time.Hour)
	p := &OIDC{codec: codec, now: func() time.Time { return now }}
	cfg, err := ParseConfig(validConfig)
	if err != nil {
		t.Fatal(err)
	}
	// About 50 bytes a group: 200 of them are well past what a cookie carries.
	groups := func(n int) []string {
		out := make([]string, n)
		for i := range out {
			out[i] = fmt.Sprintf("team-%03d-%s", i, strings.Repeat("x", 40))
		}
		return out
	}
	// Every group one rule mentions is one the session keeps, so this is the user trimming cannot help.
	allReferenced := cfg
	allReferenced.AllowedGroups = groups(200)

	for _, tc := range []struct {
		name   string
		cfg    Config
		groups []string
		// The groups the session is expected to carry; nil means no session at all.
		want []string
	}{
		{"one group fits as it is", cfg, []string{"platform"}, []string{"platform"}},
		{"the rest are trimmed away", cfg, append([]string{"platform"}, groups(200)...), []string{"platform"}},
		{"too many of the referenced ones is refused", allReferenced, groups(200), nil},
	} {
		s := session{User: "alice", Groups: tc.groups, Issuer: cfg.Issuer, ClientID: cfg.ClientID, Expires: exp.Unix()}
		sealed, err := p.seal(s, tc.cfg, exp)
		if tc.want == nil {
			if err == nil || !strings.Contains(err.Error(), "browser accepts") || sealed != "" {
				t.Errorf("%s: sealed %d bytes with err %v, want a refusal that names the limit", tc.name, len(sealed), err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", tc.name, err)
			continue
		}
		if len(sealed) > maxCookie {
			t.Errorf("%s: the cookie is %d bytes, over the %d a browser takes", tc.name, len(sealed), maxCookie)
		}
		var got session
		if err := p.codec.Open(SessionCookie, sealed, now, &got); err != nil {
			t.Errorf("%s: open: %v", tc.name, err)
			continue
		}
		if !slices.Equal(got.Groups, tc.want) {
			t.Errorf("%s: the session carries %v, want %v", tc.name, got.Groups, tc.want)
		}
	}
}

func TestDiscoveryFailureIsRemembered(t *testing.T) {
	var hits atomic.Int32
	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		http.Error(w, "down", http.StatusBadGateway)
	}))
	defer down.Close()
	cm := configMap("issuer: " + down.URL + "\nclientID: rz\n")
	p, err := NewOIDC(context.Background(), OIDCOptions{Reader: &reader{cm: cm}, ConfigMap: types.NamespacedName{
		Namespace: cm.Namespace, Name: cm.Name}, ClientSecret: "s", SessionKey: newKey(t)})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	p.now = func() time.Time { return now }
	login := func() int {
		w := httptest.NewRecorder()
		p.Handle(w, httptest.NewRequest(http.MethodGet, "/auth/login", nil))
		return w.Code
	}
	for range 3 {
		if code := login(); code != http.StatusServiceUnavailable {
			t.Fatalf("login with the IdP down = %d, want 503", code)
		}
	}
	if hits.Load() != 1 {
		t.Errorf("discovery tried %d times within %s, want 1", hits.Load(), discoveryRetry)
	}
	now = now.Add(discoveryRetry)
	_ = login()
	if hits.Load() != 2 {
		t.Errorf("discovery tried %d times after %s, want a retry", hits.Load(), discoveryRetry)
	}
}
