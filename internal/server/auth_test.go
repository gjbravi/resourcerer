package server

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"maps"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/server/auth"
)

// stub answers with a fixed identity, or none. cookie marks a strategy that authenticates with a session cookie,
// like oidc; the name stays "stub" so a test can show the server does not special-case strategy names.
type stub struct {
	id     auth.Identity
	ok     bool
	cookie bool
}

func (stub) Strategy() string                               { return "stub" }
func (p stub) Identify(*http.Request) (auth.Identity, bool) { return p.id, p.ok }
func (stub) Handle(http.ResponseWriter, *http.Request) bool { return false }
func (stub) Status() auth.Status                            { return auth.Status{} }
func (p stub) UsesSessionCookie() bool                      { return p.cookie }

var (
	nobody = stub{}
	viewer = stub{id: auth.Identity{User: "vera", Viewer: true}, ok: true}
	editor = stub{id: auth.Identity{User: "eddie", Viewer: true, Editor: true}, ok: true}
	denied = stub{id: auth.Identity{User: "dana"}, ok: true}
	// cookieEditor is a provider that authenticates with a cookie, so mutations must come from the UI.
	cookieEditor = stub{id: auth.Identity{User: "eddie", Viewer: true, Editor: true}, ok: true, cookie: true}
)

func path(pattern string) (method, p string) {
	method, p, _ = strings.Cut(pattern, " ")
	return method, strings.NewReplacer("{ns}", "demo", "{kind}", "Deployment", "{name}", "api").Replace(p)
}

func call(t *testing.T, h http.Handler, method, target, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	return callHTTPS(t, h, false, method, target, body, cookies...)
}

// callHTTPS is call, through a TLS-terminating proxy when https is set.
func callHTTPS(t *testing.T, h http.Handler, https bool, method, target, body string, cookies ...*http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if https {
		r.Header.Set("X-Forwarded-Proto", "https")
	}
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func reason(w *httptest.ResponseRecorder) string {
	var b map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &b)
	return b["reason"]
}

func newServer(p auth.Provider, objs ...client.Object) *Server {
	c := fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(objs...).Build()
	return &Server{Client: c, APIReader: c, Auth: p}
}

func TestEveryRouteHasAMethod(t *testing.T) {
	for _, rt := range (&Server{}).routes() {
		method, p, ok := strings.Cut(rt.pattern, " ")
		if !ok || method == "" || !strings.HasPrefix(p, "/api/v1/") {
			t.Errorf("%q: routes need a method and an /api/v1/ path", rt.pattern)
		}
	}
}

func TestWriteGate(t *testing.T) {
	var mutations []string
	for _, rt := range (&Server{}).routes() {
		if mutating(rt.pattern) {
			mutations = append(mutations, rt.pattern)
		}
	}
	if len(mutations) < 5 {
		t.Fatalf("found %d mutating routes, want at least the 5 known ones", len(mutations))
	}
	for _, pattern := range mutations {
		method, target := path(pattern)
		for _, tc := range []struct {
			name     string
			p        auth.Provider
			readOnly bool
			code     int
			reason   string
		}{
			{"unauthenticated", nobody, false, http.StatusUnauthorized, "unauthenticated"},
			{"no role", denied, false, http.StatusForbidden, "forbidden"},
			{"viewer", viewer, false, http.StatusForbidden, "editor-required"},
			{"editor", editor, false, 0, ""},
			{"anonymous", auth.Anonymous{}, false, 0, ""},
			{"editor, read-only", editor, true, http.StatusForbidden, "read-only"},
			{"viewer, read-only", viewer, true, http.StatusForbidden, "read-only"},
		} {
			s := newServer(tc.p)
			s.ReadOnly = tc.readOnly
			w := call(t, s.Handler(), method, target, "{}")
			switch {
			case tc.code == 0 && (w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden):
				t.Errorf("%s as %s = %d %s, want it past the gate", pattern, tc.name, w.Code, w.Body)
			case tc.code != 0 && (w.Code != tc.code || reason(w) != tc.reason):
				t.Errorf("%s as %s = %d %q, want %d %q", pattern, tc.name, w.Code, reason(w), tc.code, tc.reason)
			}
		}
	}
}

func TestReadsNeedAuthentication(t *testing.T) {
	s := newServer(nobody)
	for _, rt := range s.routes() {
		if mutating(rt.pattern) {
			continue
		}
		method, target := path(rt.pattern)
		if w := call(t, s.Handler(), method, target, ""); w.Code != http.StatusUnauthorized {
			t.Errorf("%s unauthenticated = %d, want 401", rt.pattern, w.Code)
		}
	}
	if w := call(t, newServer(denied).Handler(), http.MethodGet, "/api/v1/workloads", ""); w.Code != http.StatusForbidden {
		t.Errorf("list without a role = %d, want 403", w.Code)
	}
	if w := call(t, newServer(viewer).Handler(), http.MethodGet, "/api/v1/workloads", ""); w.Code != http.StatusOK {
		t.Errorf("list as viewer = %d, want 200", w.Code)
	}
}

func TestMe(t *testing.T) {
	for _, tc := range []struct {
		p              auth.Provider
		user, role, st string
	}{
		{auth.Anonymous{}, "anonymous", auth.RoleEditor, auth.StrategyAnonymous},
		{viewer, "vera", auth.RoleViewer, "stub"},
		{denied, "dana", auth.RoleNone, "stub"},
	} {
		w := call(t, newServer(tc.p).Handler(), http.MethodGet, "/api/v1/me", "")
		var m Me
		if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil || w.Code != http.StatusOK {
			t.Fatalf("me = %d %s", w.Code, w.Body)
		}
		if !m.Authenticated || m.User != tc.user || m.Role != tc.role || m.Strategy != tc.st || m.Groups == nil || m.Impersonating {
			t.Errorf("me = %+v, want %s as %s with %s", m, tc.user, tc.role, tc.st)
		}
	}

	// The UI mirrors an authenticated `me` by hand, in `web/src/api.ts`, and nothing else compares the two: a tag
	// renamed here disappears from the page with every test still green. Naming the key set makes whoever changes it
	// change the mirror in the same commit, which is a tripwire rather than a guard, and the only kind available here.
	w := call(t, newServer(viewer).Handler(), http.MethodGet, "/api/v1/me", "")
	var signed map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &signed)
	got := slices.Sorted(maps.Keys(signed))
	want := []string{"authenticated", "configStale", "groups", "impersonating", "role", "strategy", "user"}
	if !slices.Equal(got, want) {
		t.Errorf("me keys = %v, want %v, which is what web/src/api.ts mirrors", got, want)
	}

	w = call(t, newServer(nobody).Handler(), http.MethodGet, "/api/v1/me", "")
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if w.Code != http.StatusOK || len(body) != 2 || body["authenticated"] != false || body["strategy"] != "stub" ||
		w.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("me unauthenticated = %d %v, want 200 with authenticated and strategy only", w.Code, body)
	}
}

func TestHandlerRequiresAProvider(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Handler without Auth did not panic")
		}
	}()
	(&Server{}).Handler()
}

func recommendation() *v1alpha1.Recommendation {
	return &v1alpha1.Recommendation{
		ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Name: "deployment-api"},
		Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"}},
		Status:     v1alpha1.RecommendationStatus{Hash: "h1"},
	}
}

func TestRequestedByComesFromTheIdentity(t *testing.T) {
	for _, tc := range []struct {
		p    auth.Provider
		want string
	}{{editor, "eddie"}, {auth.Anonymous{}, "anonymous"}} {
		for _, action := range []string{"apply", "revert"} {
			s := newServer(tc.p, recommendation())
			w := call(t, s.Handler(), http.MethodPost, "/api/v1/workloads/demo/Deployment/api/"+action,
				`{"hash":"h1","requestedBy":"mallory"}`)
			if w.Code != http.StatusOK {
				t.Fatalf("%s = %d %s", action, w.Code, w.Body)
			}
			var got v1alpha1.Recommendation
			_ = s.Client.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "deployment-api"}, &got)
			if got.Spec.ApplyRequest == nil || got.Spec.ApplyRequest.RequestedBy != tc.want {
				t.Errorf("%s by %s recorded %+v", action, tc.want, got.Spec.ApplyRequest)
			}
		}
	}
}

/* ── OIDC against a test issuer ────────────────────────────────────────── */

type issuer struct {
	*httptest.Server
	key       *rsa.PrivateKey
	challenge string
	nonce     string
	claims    map[string]any
}

func newIssuer(t *testing.T) *issuer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	is := &issuer{key: key, claims: map[string]any{"preferred_username": "alice", "groups": []string{"platform"}}}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{"issuer": is.URL, "authorization_endpoint": is.URL + "/authorize",
			"token_endpoint": is.URL + "/token", "jwks_uri": is.URL + "/jwks", "id_token_signing_alg_values_supported": []string{"RS256"}})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		enc := base64.RawURLEncoding
		writeJSON(w, map[string]any{"keys": []map[string]string{{"kty": "RSA", "kid": "k1", "alg": "RS256", "use": "sig",
			"n": enc.EncodeToString(key.N.Bytes()), "e": enc.EncodeToString(big.NewInt(int64(key.E)).Bytes())}}})
	})
	mux.HandleFunc("POST /token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
		if r.Form.Get("code") != "good-code" || base64.RawURLEncoding.EncodeToString(sum[:]) != is.challenge {
			writeError(w, http.StatusBadRequest, "invalid_grant")
			return
		}
		claims := map[string]any{"iss": is.URL, "aud": "resourcerer", "sub": "u1", "nonce": is.nonce,
			"iat": time.Now().Unix(), "exp": time.Now().Add(5 * time.Minute).Unix()}
		for k, v := range is.claims {
			claims[k] = v
		}
		writeJSON(w, map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 300, "id_token": is.sign(t, claims)})
	})
	is.Server = httptest.NewServer(mux)
	t.Cleanup(is.Close)
	return is
}

func (is *issuer) sign(t *testing.T, claims map[string]any) string {
	t.Helper()
	enc := base64.RawURLEncoding
	h, _ := json.Marshal(map[string]string{"alg": "RS256", "kid": "k1", "typ": "JWT"})
	c, _ := json.Marshal(claims)
	unsigned := enc.EncodeToString(h) + "." + enc.EncodeToString(c)
	sum := sha256.Sum256([]byte(unsigned))
	sig, err := rsa.SignPKCS1v15(rand.Reader, is.key, crypto.SHA256, sum[:])
	if err != nil {
		t.Fatal(err)
	}
	return unsigned + "." + enc.EncodeToString(sig)
}

type configMaps struct {
	client.Reader
	cm *corev1.ConfigMap
}

func (r configMaps) Get(_ context.Context, _ types.NamespacedName, obj client.Object, _ ...client.GetOption) error {
	*obj.(*corev1.ConfigMap) = *r.cm
	return nil
}

func oidcServer(t *testing.T, is *issuer, allowed string) http.Handler {
	t.Helper()
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	cm := &corev1.ConfigMap{Data: map[string]string{auth.ConfigKey: "issuer: " + is.URL + "\nclientID: resourcerer\n" +
		"allowedGroups: [" + allowed + "]\nroles:\n  editorGroups: [platform]\n"}}
	p, err := auth.NewOIDC(context.Background(), auth.OIDCOptions{Reader: configMaps{cm: cm},
		ConfigMap: types.NamespacedName{Namespace: "rz", Name: "rz-auth"}, ClientSecret: "secret",
		SessionKey: base64.StdEncoding.EncodeToString(key)})
	if err != nil {
		t.Fatal(err)
	}
	return newServer(p).Handler()
}

func cookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name && c.MaxAge >= 0 && c.Value != "" {
			return c
		}
	}
	return nil
}

// login starts a sign-in and returns the state the IdP would echo back and the state cookie.
func login(t *testing.T, h http.Handler, is *issuer) (string, *http.Cookie) {
	t.Helper()
	return loginHTTPS(t, h, is, false)
}

func loginHTTPS(t *testing.T, h http.Handler, is *issuer, https bool) (string, *http.Cookie) {
	t.Helper()
	w := callHTTPS(t, h, https, http.MethodGet, "/auth/login?redirect=/workloads/demo", "")
	if w.Code != http.StatusFound {
		t.Fatalf("login = %d %s", w.Code, w.Body)
	}
	to, _ := url.Parse(w.Header().Get("Location"))
	q := to.Query()
	if !strings.HasPrefix(to.String(), is.URL+"/authorize") || q.Get("code_challenge_method") != "S256" ||
		q.Get("code_challenge") == "" || q.Get("state") == "" || q.Get("nonce") == "" || q.Get("client_id") != "resourcerer" ||
		!strings.HasSuffix(q.Get("redirect_uri"), "://example.com/auth/callback") {
		t.Fatalf("login redirects to %s", to)
	}
	is.challenge, is.nonce = q.Get("code_challenge"), q.Get("nonce")
	name := "resourcerer_login"
	if https {
		name = "__Host-" + name
	}
	st := cookie(w, name)
	if st == nil || !st.HttpOnly || st.Path != "/" || st.Secure != https {
		t.Fatalf("login set no HttpOnly state cookie: %v", w.Result().Cookies())
	}
	return q.Get("state"), st
}

func TestOIDCFlow(t *testing.T) {
	is := newIssuer(t)
	h := oidcServer(t, is, "platform")

	if w := call(t, h, http.MethodGet, "/api/v1/me", ""); w.Code != http.StatusOK ||
		strings.TrimSpace(w.Body.String()) != `{"authenticated":false,"strategy":"oidc"}` {
		t.Fatalf("me before sign-in = %d %s", w.Code, w.Body)
	}
	if w := call(t, h, http.MethodGet, "/api/v1/info", ""); w.Code != http.StatusUnauthorized {
		t.Fatalf("info before sign-in = %d, want 401", w.Code)
	}
	state, st := login(t, h, is)

	w := call(t, h, http.MethodGet, "/auth/callback?code=good-code&state=forged", "", st)
	if w.Code != http.StatusBadRequest || cookie(w, auth.SessionCookie) != nil {
		t.Errorf("callback with a mismatched state = %d, session %v", w.Code, cookie(w, auth.SessionCookie))
	}
	if w := call(t, h, http.MethodGet, "/auth/callback?code=good-code&state="+state, ""); w.Code != http.StatusBadRequest {
		t.Errorf("callback without the state cookie = %d, want 400", w.Code)
	}

	w = call(t, h, http.MethodGet, "/auth/callback?code=good-code&state="+state, "", st)
	sess := cookie(w, auth.SessionCookie)
	if w.Code != http.StatusFound || w.Header().Get("Location") != "/workloads/demo" || sess == nil {
		t.Fatalf("callback = %d to %q, session %v: %s", w.Code, w.Header().Get("Location"), sess, w.Body)
	}
	if !sess.HttpOnly || sess.SameSite != http.SameSiteLaxMode || sess.Path != "/" || sess.Secure {
		t.Errorf("session cookie = %+v, want HttpOnly, SameSite=Lax, Path=/, not Secure over plain HTTP", sess)
	}

	w = call(t, h, http.MethodGet, "/api/v1/me", "", sess)
	var m Me
	_ = json.Unmarshal(w.Body.Bytes(), &m)
	if w.Code != http.StatusOK || !m.Authenticated || m.User != "alice" || m.Role != auth.RoleEditor || m.Strategy != auth.StrategyOIDC ||
		len(m.Groups) != 1 || m.ExpiresAt == "" {
		t.Errorf("me = %d %+v", w.Code, m)
	}

	w = call(t, h, http.MethodGet, "/auth/logout", "", sess)
	// The session cookie goes and the browser lands on the signed-out screen: back at "/" it would be sent to the
	// provider again and the user would never see that signing out worked.
	if w.Code != http.StatusFound || w.Header().Get("Location") != "/?signed-out" || cookie(w, auth.SessionCookie) != nil {
		t.Errorf("logout = %d, location %q, cookies %v", w.Code, w.Header().Get("Location"), w.Result().Cookies())
	}
}

func TestOIDCRejects(t *testing.T) {
	is := newIssuer(t)
	h := oidcServer(t, is, "platform")

	state, st := login(t, h, is)
	is.nonce = "replayed"
	if w := call(t, h, http.MethodGet, "/auth/callback?code=good-code&state="+state, "", st); w.Code != http.StatusUnauthorized ||
		cookie(w, auth.SessionCookie) != nil {
		t.Errorf("callback with another nonce = %d", w.Code)
	}

	state, st = login(t, h, is)
	if w := call(t, h, http.MethodGet, "/auth/callback?code=stolen&state="+state, "", st); w.Code != http.StatusUnauthorized {
		t.Errorf("callback with a bad code = %d", w.Code)
	}

	is.claims["groups"] = []string{"marketing"}
	state, st = login(t, h, is)
	if w := call(t, h, http.MethodGet, "/auth/callback?code=good-code&state="+state, "", st); w.Code != http.StatusForbidden ||
		cookie(w, auth.SessionCookie) != nil {
		t.Errorf("callback outside allowedGroups = %d", w.Code)
	}
}

func TestOIDCSecureCookies(t *testing.T) {
	is := newIssuer(t)
	h := oidcServer(t, is, "platform")
	state, st := loginHTTPS(t, h, is, true)
	w := callHTTPS(t, h, true, http.MethodGet, "/auth/callback?code=good-code&state="+state, "", st)
	sess := cookie(w, auth.SecureSessionCookie)
	if w.Code != http.StatusFound || sess == nil || !sess.Secure || sess.Path != "/" || sess.Domain != "" || !sess.HttpOnly {
		t.Fatalf("callback over HTTPS = %d, session %+v", w.Code, sess)
	}
	if w := callHTTPS(t, h, true, http.MethodGet, "/api/v1/workloads", "", sess); w.Code != http.StatusOK {
		t.Errorf("list with the __Host- session = %d %s", w.Code, w.Body)
	}
	plain := &http.Cookie{Name: auth.SessionCookie, Value: sess.Value}
	if w := callHTTPS(t, h, true, http.MethodGet, "/api/v1/workloads", "", plain); w.Code != http.StatusUnauthorized {
		t.Errorf("unprefixed session over HTTPS = %d, want 401", w.Code)
	}
}
