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
