package recommender

import (
	"testing"
	"time"
)

func TestNextSlotIsStableAndInTheFuture(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	a := nextSlot(now, 24*time.Hour, "demo/Deployment/api")
	b := nextSlot(now.Add(time.Minute), 24*time.Hour, "demo/Deployment/api")
	if !a.After(now) || a.Sub(now) > 24*time.Hour {
		t.Fatalf("slot %v not within the next period", a)
	}
	if !a.Equal(b) && b.Sub(a) != 24*time.Hour {
		t.Errorf("slot moved: %v then %v", a, b)
	}
	if nextSlot(now, 24*time.Hour, "demo/Deployment/web").Equal(a) {
		t.Error("different workloads should usually get different slots")
	}
}

func TestRetryBacksOffUpToThePeriod(t *testing.T) {
	for _, tc := range []struct {
		attempts int
		period   time.Duration
		want     time.Duration
	}{
		{1, 24 * time.Hour, time.Minute},
		{2, 24 * time.Hour, 2 * time.Minute},
		{3, 24 * time.Hour, 4 * time.Minute},
		{11, 24 * time.Hour, 1024 * time.Minute},
		{12, 24 * time.Hour, 24 * time.Hour},
		{50, 24 * time.Hour, 24 * time.Hour},
		{5, time.Minute, time.Minute},
		{3, 3 * time.Minute, 3 * time.Minute},
		{1, 0, time.Minute},
	} {
		if got := retryAfter(tc.attempts, tc.period); got != tc.want {
			t.Errorf("retryAfter(%d, %v) = %v, want %v", tc.attempts, tc.period, got, tc.want)
		}
	}
}

func TestRetryCountsAttemptsInARow(t *testing.T) {
	r := &Reconciler{}
	for want := 1; want <= 3; want++ {
		if got := r.retry("demo/api"); got != want {
			t.Fatalf("attempt %d counted as %d", want, got)
		}
	}
	if got := r.retry("demo/web"); got != 1 {
		t.Errorf("another workload starts at %d, want 1", got)
	}
}
