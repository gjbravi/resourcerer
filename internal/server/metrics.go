package server

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// The server answers people: each page view runs live Prometheus queries, and Evaluate calls a model outside the
// cluster. These show how often, how fast, and how much of it the cache absorbs.
var (
	httpRequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_http_requests_total",
		Help: "API and UI requests by route, method and status code.",
	}, []string{"route", "method", "code"})
	httpSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "resourcerer_http_request_duration_seconds",
		Help:    "Time to answer API and UI requests, by route and method.",
		Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30},
	}, []string{"route", "method"})
	cacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_server_cache_requests_total",
		Help: "Lookups of the server's query cache by query and result: hit (answered from the cache), shared (waited for the same query already running) or miss (ran it).",
	}, []string{"query", "result"})
	decisionCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_decision_requests_total",
		Help: "Calls to the decision model by result: ok, error (no usable response) or invalid (an answer that does not fit the question).",
	}, []string{"result"})
	decisionSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "resourcerer_decision_request_duration_seconds",
		Help:    "Time the decision model took to answer.",
		Buckets: []float64{0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
	})
)

func init() {
	metrics.Registry.MustRegister(httpRequestsCounter, httpSeconds, cacheCounter, decisionCounter, decisionSeconds)
}

// routeName is the route a ServeMux pattern serves, as metrics label it and as the purpose of its Prometheus queries:
// workload/series for GET /api/v1/workloads/{ns}/{kind}/{name}/series. A request no pattern matched is unmatched.
func routeName(pattern string) string {
	path := pattern
	if _, p, ok := strings.Cut(pattern, " "); ok {
		path = p
	}
	switch path {
	case "":
		return "unmatched"
	case "/":
		return "ui"
	case "/api/":
		return "unknown"
	case "/auth/":
		return "auth"
	}
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.Replace(path, "workloads/{ns}/{kind}/{name}", "workload", 1)
	return strings.Replace(path, "namespaces/{ns}", "namespace", 1)
}

type statusWriter struct {
	http.ResponseWriter
	code int
}

func (w *statusWriter) WriteHeader(code int) {
	if w.code == 0 {
		w.code = code
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func observe(r *http.Request, code int, took time.Duration) {
	route := routeName(r.Pattern)
	if code == 0 {
		code = http.StatusOK
	}
	// The method comes from the client: anything the API does not serve is one label value, not one per string sent.
	method := r.Method
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut:
	default:
		method = "other"
	}
	httpRequestsCounter.WithLabelValues(route, method, strconv.Itoa(code)).Inc()
	httpSeconds.WithLabelValues(route, method).Observe(took.Seconds())
}
