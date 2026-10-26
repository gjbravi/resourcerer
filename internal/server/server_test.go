package server

import (
	"errors"
	"math"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/prom"
)

func TestBandsBucketsPercentiles(t *testing.T) {
	start := time.Unix(0, 0)
	end := start.Add(150 * time.Minute)
	var pts []prom.Point
	for i := 0; i < 150*60; i += 15 {
		pts = append(pts, prom.Point{T: float64(i), V: float64(i % 60)})
	}
	b := bands(pts, start, end)
	if len(b) != buckets {
		t.Fatalf("got %d buckets, want %d", len(b), buckets)
	}
	if !near(b[0].P50, 22.5) || !near(b[0].P95, 42.75) {
		t.Errorf("bucket 0 = %+v, want p50 22.5 and p95 42.75 for samples 0,15,30,45", b[0])
	}
}

func TestCacheCollapsesAndExpires(t *testing.T) {
	var c cache
	var calls atomic.Int32
	fetch := func() (any, error) { calls.Add(1); return "v", nil }
	for range 3 {
		if v := c.get("k", time.Hour, fetch); v != "v" {
			t.Fatalf("got %v", v)
		}
	}
	if calls.Load() != 1 {
		t.Errorf("fetched %d times, want 1", calls.Load())
	}
	c.get("e", time.Hour, func() (any, error) { return nil, errors.New("down") })
	if v := c.get("e", time.Hour, fetch); v != "v" {
		t.Errorf("errors must not be cached, got %v", v)
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
