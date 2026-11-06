package prom

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestQueriesAreCountedByPurpose(t *testing.T) {
	var agent string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		agent = r.UserAgent()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/query":
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"vector","result":[
				{"metric":{"container":"app"},"value":[1,"0.5"]},{"metric":{"container":"sidecar"},"value":[1,"0.1"]}]}}`))
		case "/api/v1/query_range":
			_, _ = w.Write([]byte(`{"status":"success","data":{"resultType":"matrix","result":[
				{"metric":{"container":"app"},"values":[[1,"1"],[2,"2"],[3,"3"]]}]}}`))
		}
	}))
	defer ts.Close()
	c, err := New(ts.URL, "5m", Options{UserAgent: "resourcerer/test (controller)"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := For(context.Background(), "test-usage")
	if _, err := c.Instant(ctx, "up", "container"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Range(ctx, "up", time.Unix(0, 0), time.Unix(3, 0), time.Second, "container"); err != nil {
		t.Fatal(err)
	}

	if agent != "resourcerer/test (controller)" {
		t.Errorf("User-Agent = %q", agent)
	}
	for _, typ := range []string{"instant", "range"} {
		if got := testutil.ToFloat64(requestsCounter.WithLabelValues("test-usage", typ, "200")); got != 1 {
			t.Errorf("%s requests = %v, want 1", typ, got)
		}
	}
	if got := testutil.ToFloat64(samplesCounter.WithLabelValues("test-usage")); got != 5 {
		t.Errorf("samples = %v, want 2 from the instant query and 3 from the range", got)
	}
	if got := testutil.ToFloat64(bytesCounter.WithLabelValues("test-usage")); got <= 0 {
		t.Errorf("response bytes = %v, want some", got)
	}
	if got := testutil.ToFloat64(inflightGauge); got != 0 {
		t.Errorf("in flight after the responses were read = %v, want 0", got)
	}
}

func TestATimeoutIsCountedAsOne(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(2 * time.Second):
		}
	}))
	defer ts.Close()
	c, err := New(ts.URL, "5m", Options{Timeout: 50 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	timeouts := requestsCounter.WithLabelValues("test-timeout", "instant", "timeout")
	before := testutil.ToFloat64(timeouts)
	if _, err := c.Instant(For(context.Background(), "test-timeout"), "up", ""); err == nil {
		t.Fatal("want an error from a query slower than the timeout")
	}
	if got := testutil.ToFloat64(timeouts) - before; got != 1 {
		t.Errorf("timeouts = %v, want 1", got)
	}
}

func TestUntaggedQueriesCountAsOther(t *testing.T) {
	if got := purposeOf(context.Background()); got != "other" {
		t.Errorf("purpose = %q, want other", got)
	}
}
