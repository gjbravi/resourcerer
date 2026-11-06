package prom

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Every query resourcerer sends is counted by what it was for, so the load it puts on Prometheus can be watched and
// attributed as a cluster grows. The component that sent it is the scrape target's own labels.
var (
	requestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_prometheus_requests_total",
		Help: "Queries sent to Prometheus, by purpose, type (range or instant) and HTTP status code: error when no response came back, timeout when it took too long.",
	}, []string{"purpose", "type", "code"})
	durationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "resourcerer_prometheus_request_duration_seconds",
		Help:    "Time from sending a query to Prometheus to reading the whole response.",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"purpose", "type"})
	inflightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "resourcerer_prometheus_requests_in_flight",
		Help: "Queries sent to Prometheus that have not been answered yet.",
	})
	bytesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_prometheus_response_bytes_total",
		Help: "Bytes read from Prometheus responses, by purpose.",
	}, []string{"purpose"})
	samplesCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_prometheus_samples_total",
		Help: "Samples returned by Prometheus queries, by purpose: what a query cost to evaluate and to send.",
	}, []string{"purpose"})
)

func init() {
	metrics.Registry.MustRegister(requestsCounter, durationHistogram, inflightGauge, bytesCounter, samplesCounter)
}

type purposeKey struct{}

// For tags the queries sent with the returned context with what they are for, such as usage or signals; queries
// without one are counted as other.
func For(ctx context.Context, purpose string) context.Context {
	return context.WithValue(ctx, purposeKey{}, purpose)
}

func purposeOf(ctx context.Context) string {
	if p, ok := ctx.Value(purposeKey{}).(string); ok && p != "" {
		return p
	}
	return "other"
}

func countSamples(ctx context.Context, n int) {
	samplesCounter.WithLabelValues(purposeOf(ctx)).Add(float64(n))
}

// instrumented counts each request and its response, and names resourcerer in the User-Agent so whoever runs
// Prometheus can tell its queries apart in their own logs.
type instrumented struct {
	next      http.RoundTripper
	userAgent string
}

func (t instrumented) RoundTrip(req *http.Request) (*http.Response, error) {
	purpose, typ := purposeOf(req.Context()), queryType(req.URL.Path)
	if t.userAgent != "" {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", t.userAgent)
	}
	start := time.Now()
	inflightGauge.Inc()
	resp, err := t.next.RoundTrip(req)
	if err != nil {
		inflightGauge.Dec()
		requestsCounter.WithLabelValues(purpose, typ, errorCode(err)).Inc()
		durationHistogram.WithLabelValues(purpose, typ).Observe(time.Since(start).Seconds())
		return nil, err
	}
	requestsCounter.WithLabelValues(purpose, typ, strconv.Itoa(resp.StatusCode)).Inc()
	resp.Body = &countedBody{ReadCloser: resp.Body, done: func(n int64) {
		inflightGauge.Dec()
		bytesCounter.WithLabelValues(purpose).Add(float64(n))
		durationHistogram.WithLabelValues(purpose, typ).Observe(time.Since(start).Seconds())
	}}
	return resp, nil
}

func queryType(path string) string {
	switch {
	case strings.HasSuffix(path, "/query_range"):
		return "range"
	case strings.HasSuffix(path, "/query"):
		return "instant"
	}
	return "other"
}

func errorCode(err error) string {
	var ne net.Error
	if errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()) {
		return "timeout"
	}
	return "error"
}

// countedBody reports the bytes read once the client closes it: a large response takes as long to read as to evaluate.
type countedBody struct {
	io.ReadCloser
	n    int64
	once sync.Once
	done func(int64)
}

func (b *countedBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.n += int64(n)
	return n, err
}

func (b *countedBody) Close() error {
	err := b.ReadCloser.Close()
	b.once.Do(func() { b.done(b.n) })
	return err
}
