package applier

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

func TestAppliesAreCountedByActionAndTrigger(t *testing.T) {
	applyCounter.Reset()
	countApply(nil, "Succeeded")
	countApply(&v1alpha1.ApplyRequest{Action: v1alpha1.ActionApply}, "Rejected")
	countApply(&v1alpha1.ApplyRequest{Action: v1alpha1.ActionRevert}, "Succeeded")
	countApply(&v1alpha1.ApplyRequest{Action: "Delete"}, "Rejected")

	for _, tc := range []struct {
		action, trigger, result string
		want                    float64
	}{
		{"Apply", "auto", "Succeeded", 1},
		{"Apply", "manual", "Rejected", 2},
		{"Revert", "manual", "Succeeded", 1},
	} {
		if got := testutil.ToFloat64(applyCounter.WithLabelValues(tc.action, tc.trigger, tc.result)); got != tc.want {
			t.Errorf("%s %s %s = %v, want %v", tc.action, tc.trigger, tc.result, got, tc.want)
		}
	}
	if n := testutil.CollectAndCount(applyCounter); n != 3 {
		t.Errorf("an unknown action made its own series: %d series, want 3", n)
	}
}
