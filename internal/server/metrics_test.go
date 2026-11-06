package server

import (
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestRouteNames(t *testing.T) {
	for pattern, want := range map[string]string{
		"GET /api/v1/workloads/{ns}/{kind}/{name}/series": "workload/series",
		"GET /api/v1/workloads/{ns}/{kind}/{name}":        "workload",
		"GET /api/v1/workloads":                           "workloads",
		"GET /api/v1/cluster/series":                      "cluster/series",
		"PUT /api/v1/namespaces/{ns}/settings":            "namespace/settings",
		"GET /api/v1/me":                                  "me",
		"/api/":                                           "unknown",
		"/auth/":                                          "auth",
		"/":                                               "ui",
		"":                                                "unmatched",
	} {
		if got := routeName(pattern); got != want {
			t.Errorf("routeName(%q) = %q, want %q", pattern, got, want)
		}
	}
}

func TestRequestsAreCountedByRouteAndCode(t *testing.T) {
	httpRequestsCounter.Reset()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/workloads/{ns}/{kind}/{name}/placement", func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusBadGateway, "query Prometheus")
	})
	mux.HandleFunc("GET /api/v1/workloads", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, []string{}) })
	h := logRequests(mux)
	for _, path := range []string{"/api/v1/workloads/demo/Deployment/api/placement", "/api/v1/workloads", "/api/v1/workloads"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("BREW", "/api/v1/workloads", nil))

	for _, tc := range []struct {
		route, method, code string
		want                float64
	}{
		{"workload/placement", "GET", "502", 1},
		{"workloads", "GET", "200", 2},
		{"unmatched", "other", "405", 1},
	} {
		if got := testutil.ToFloat64(httpRequestsCounter.WithLabelValues(tc.route, tc.method, tc.code)); got != tc.want {
			t.Errorf("%s %s %s = %v, want %v", tc.method, tc.route, tc.code, got, tc.want)
		}
	}
}

func TestCacheCountsHitsSharesAndMisses(t *testing.T) {
	var c cache
	release := make(chan struct{})
	started := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.get("test-cache|a", time.Minute, func() (any, error) { close(started); <-release; return 1, nil })
	}()
	<-started
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.get("test-cache|a", time.Minute, func() (any, error) { t.Error("a concurrent lookup ran the query again"); return 2, nil })
	}()
	// The second lookup must find the first one running before it is released.
	for testutil.ToFloat64(cacheCounter.WithLabelValues("test-cache", "shared")) == 0 {
		runtime.Gosched()
	}
	close(release)
	wg.Wait()
	c.get("test-cache|a", time.Minute, func() (any, error) { t.Error("a cached value was fetched again"); return 3, nil })

	for result, want := range map[string]float64{"miss": 1, "shared": 1, "hit": 1} {
		if got := testutil.ToFloat64(cacheCounter.WithLabelValues("test-cache", result)); got != want {
			t.Errorf("%s = %v, want %v", result, got, want)
		}
	}
}
