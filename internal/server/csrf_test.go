package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/internal/server/auth"
)

// ajaxPair is what the UI sends with a mutation that carries a body.
const ajaxPair = "X-Resourcerer-Request: 1\nContent-Type: application/json"

// rawCall sends a mutation as a browser puts it on the wire: with the header lines given and, when chunked, with no
// Content-Length, which is how fetch streams a body.
func rawCall(t *testing.T, h http.Handler, method, target, body, headers string, chunked bool) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	for _, line := range strings.Split(headers, "\n") {
		if line != "" {
			name, value, _ := strings.Cut(line, ": ")
			r.Header.Set(name, value)
		}
	}
	if chunked {
		r.ContentLength = -1
		r.TransferEncoding = []string{"chunked"}
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

// A provider that authenticates with a cookie is asked to prove a mutation came from the UI, on every route that
// changes something, whoever the caller is.
func TestMutationsNeedTheUIHeaderUnderACookieStrategy(t *testing.T) {
	s := newServer(cookieEditor)
	h := s.Handler()
	const body = `{"hash":"h1","key":"goal","value":"Savings"}`
	var mutations int
	for _, rt := range s.routes() {
		if !mutating(rt.pattern) {
			continue
		}
		mutations++
		method, target := path(rt.pattern)
		for _, tc := range []struct{ name, headers string }{
			{"no header", "Content-Type: application/json"},
			{"no header, form content type", "Content-Type: application/x-www-form-urlencoded"},
			{"header with another value", "X-Resourcerer-Request: fetch\nContent-Type: application/json"},
		} {
			w := rawCall(t, h, method, target, body, tc.headers, false)
			if w.Code != http.StatusForbidden || reason(w) != "ajax-required" {
				t.Errorf("%s with %s = %d %q, want 403 ajax-required", rt.pattern, tc.name, w.Code, w.Body)
			}
		}
	}
	if mutations < 5 {
		t.Fatalf("found %d mutating routes, want at least the 5 known ones", mutations)
	}
}

func TestWriteAsksACookieStrategyForTheUIHeader(t *testing.T) {
	const apply = "/api/v1/workloads/demo/Deployment/api/apply"
	const settingsRoute = "/api/v1/workloads/demo/Deployment/api/settings"
	for _, tc := range []struct {
		name, method, target, headers, body string
		chunked                             bool
		code                                int
		reason                              string
	}{
		{"json from the UI", http.MethodPost, apply, ajaxPair, `{"hash":"h1"}`, false, http.StatusOK, ""},
		{"json with a charset", http.MethodPost, apply, "X-Resourcerer-Request: 1\nContent-Type: application/json; charset=utf-8", `{"hash":"h1"}`, false, http.StatusOK, ""},
		{"chunked json from the UI", http.MethodPost, apply, ajaxPair, `{"hash":"h1"}`, true, http.StatusOK, ""},
		// A mutation with no body has nothing to parse, so only the header is required of it.
		{"no body", http.MethodPost, "/api/v1/workloads/demo/Deployment/api/revert", "X-Resourcerer-Request: 1", "", false, http.StatusOK, ""},
		{"no header at all", http.MethodPost, apply, "Content-Type: application/json", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		{"no header and no content type", http.MethodPost, apply, "", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		// The hole this closes: a same-site page posting a JSON body the browser sends without a preflight.
		{"json body posted as text/plain", http.MethodPost, apply, "Content-Type: text/plain", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		{"json body posted as a form", http.MethodPost, apply, "Content-Type: application/x-www-form-urlencoded", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		{"header, body posted as a form", http.MethodPost, apply, "X-Resourcerer-Request: 1\nContent-Type: application/x-www-form-urlencoded", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		{"header, chunked body posted as text/plain", http.MethodPost, apply, "X-Resourcerer-Request: 1\nContent-Type: text/plain", `{"hash":"h1"}`, true, http.StatusForbidden, "ajax-required"},
		{"header, no content type and a body", http.MethodPost, apply, "X-Resourcerer-Request: 1", `{"hash":"h1"}`, false, http.StatusForbidden, "ajax-required"},
		// The settings routes are PUT, the other method the UI mutates with, and they carry a body like apply does.
		{"settings from the UI", http.MethodPut, settingsRoute, ajaxPair, `{"key":"cpu.optimization-goal","value":"Savings"}`, false, http.StatusOK, ""},
		{"settings from the UI, chunked", http.MethodPut, settingsRoute, ajaxPair, `{"key":"cpu.optimization-goal","value":"Savings"}`, true, http.StatusOK, ""},
		{"settings with no header", http.MethodPut, settingsRoute, "Content-Type: application/json", `{"key":"cpu.optimization-goal","value":"Savings"}`, false, http.StatusForbidden, "ajax-required"},
		{"settings posted as a form", http.MethodPut, settingsRoute, "X-Resourcerer-Request: 1\nContent-Type: application/x-www-form-urlencoded", `key=cpu.optimization-goal&value=Savings`, false, http.StatusForbidden, "ajax-required"},
	} {
		// A case that reaches the handler has to find the workload there: the settings route patches its annotations.
		s := newServer(cookieEditor, recommendation(), deployment())
		w := rawCall(t, s.Handler(), tc.method, tc.target, tc.body, tc.headers, tc.chunked)
		if w.Code != tc.code || reason(w) != tc.reason {
			t.Errorf("%s: %d %q, want %d %q", tc.name, w.Code, w.Body, tc.code, tc.reason)
		}
	}
}

// deployment is the workload the settings route in the table above writes to.
func deployment() *appsv1.Deployment {
	return &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Name: "api"}}
}

// The caller is identified first, then asked where the request came from, and only then does the gate decide what
// that caller may do. The stub's strategy is named "stub", so this also shows the check follows the provider's
// capability and not a strategy name.
func TestHeaderIsCheckedAfterTheIdentityAndBeforeTheRestOfTheGate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		p        auth.Provider
		readOnly bool
		headers  string
		code     int
		reason   string
	}{
		{"unauthenticated", stub{cookie: true}, false, "", http.StatusUnauthorized, "unauthenticated"},
		{"no role", stub{id: auth.Identity{User: "dana"}, ok: true, cookie: true}, false, "", http.StatusForbidden, "forbidden"},
		{"viewer, no header", stub{id: auth.Identity{User: "vera", Viewer: true}, ok: true, cookie: true}, false, "", http.StatusForbidden, "ajax-required"},
		{"read-only, no header", cookieEditor, true, "", http.StatusForbidden, "ajax-required"},
		{"viewer, with the header", stub{id: auth.Identity{User: "vera", Viewer: true}, ok: true, cookie: true}, false, ajaxPair, http.StatusForbidden, "editor-required"},
		{"read-only, with the header", cookieEditor, true, ajaxPair, http.StatusForbidden, "read-only"},
	} {
		s := newServer(tc.p)
		s.ReadOnly = tc.readOnly
		w := rawCall(t, s.Handler(), http.MethodPost, "/api/v1/workloads/demo/Deployment/api/revert", "", tc.headers, false)
		if w.Code != tc.code || reason(w) != tc.reason {
			t.Errorf("%s: %d %q, want %d %q", tc.name, w.Code, w.Body, tc.code, tc.reason)
		}
	}
}

// Nothing changes for a strategy with no session cookie: with nothing to carry across origins there is nothing to
// guard, and the curl checks that drive the kind environment keep working as they are.
func TestWritesWithoutACookieStayOpen(t *testing.T) {
	for _, tc := range []struct {
		name, headers, body string
		code                int
	}{
		{"as the UI sends it", ajaxPair, `{"hash":"h1"}`, http.StatusOK},
		{"as curl sends it", "Content-Type: application/x-www-form-urlencoded", `{"hash":"h1"}`, http.StatusOK},
		{"with nothing", "", `{"hash":"h1"}`, http.StatusOK},
		{"the hash is still checked", ajaxPair, `{"hash":"other"}`, http.StatusConflict},
	} {
		s := newServer(auth.Anonymous{}, recommendation())
		w := rawCall(t, s.Handler(), http.MethodPost, "/api/v1/workloads/demo/Deployment/api/apply", tc.body, tc.headers, false)
		if w.Code != tc.code {
			t.Errorf("anonymous %s: %d %s, want %d", tc.name, w.Code, w.Body, tc.code)
		}
	}
}

// Reads keep their shape: the guard belongs to mutations, where a cross-site request could change something.
func TestReadsNeverNeedTheHeader(t *testing.T) {
	s := newServer(cookieEditor)
	if w := call(t, s.Handler(), http.MethodGet, "/api/v1/workloads", ""); w.Code != http.StatusOK {
		t.Errorf("list without the header = %d %s, want 200", w.Code, w.Body)
	}
	if w := call(t, s.Handler(), http.MethodGet, "/api/v1/me", ""); w.Code != http.StatusOK {
		t.Errorf("me without the header = %d, want 200", w.Code)
	}
}

func TestExplicitMutation(t *testing.T) {
	req := func(header, value, contentType string, length int64) *http.Request {
		r, err := http.NewRequest(http.MethodPost, "/api/v1/x", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		if header != "" {
			r.Header.Set(header, value)
		}
		if contentType != "" {
			r.Header.Set("Content-Type", contentType)
		}
		r.ContentLength = length
		return r
	}
	for _, tc := range []struct {
		name string
		r    *http.Request
		want bool
	}{
		{"json from the UI", req(ajaxHeader, ajaxValue, "application/json", 2), true},
		{"json with a charset", req(ajaxHeader, ajaxValue, "application/json;charset=utf-8", 2), true},
		{"header only, no body", req(ajaxHeader, ajaxValue, "", 0), true},
		{"no header", req("", "", "application/json", 2), false},
		{"header with another value", req(ajaxHeader, "0", "application/json", 2), false},
		{"empty header", req(ajaxHeader, "", "application/json", 2), false},
		{"no content type with a body", req(ajaxHeader, ajaxValue, "", 2), false},
		{"text/plain", req(ajaxHeader, ajaxValue, "text/plain", 2), false},
		{"a form", req(ajaxHeader, ajaxValue, "application/x-www-form-urlencoded", 2), false},
		{"a type with no subtype", req(ajaxHeader, ajaxValue, "application", 2), false},
		{"a broken content type", req(ajaxHeader, ajaxValue, "application/json; q", 2), false},
		{"chunked json", req(ajaxHeader, ajaxValue, "application/json", -1), true},
		{"chunked text/plain", req(ajaxHeader, ajaxValue, "text/plain", -1), false},
	} {
		if got := explicitMutation(tc.r); got != tc.want {
			t.Errorf("%s = %v, want %v", tc.name, got, tc.want)
		}
	}
}
