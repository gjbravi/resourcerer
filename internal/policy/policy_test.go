package policy

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// mondayNoon is 2026-11-02 12:00 in São Paulo (UTC-3), a Monday, used as "now" unless a case says
// otherwise. São Paulo had no daylight saving since 2019, so the arithmetic in these cases is stable.
var mondayNoon = time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)

func freeze(from, to time.Time, reason string) v1alpha1.ApplyFreeze {
	return v1alpha1.ApplyFreeze{From: metav1.NewTime(from), To: metav1.NewTime(to), Reason: reason}
}

func rules(name string, spec v1alpha1.ApplyRules) Rules { return Rules{Name: name, Spec: spec} }

func TestEvaluate(t *testing.T) {
	businessHours := []v1alpha1.ApplyWindow{{Days: []string{"Mon", "Tue", "Wed", "Thu", "Fri"}, From: "10:00", To: "16:00", TimeZone: "America/Sao_Paulo"}}
	weekend := []v1alpha1.ApplyWindow{{Days: []string{"Sat", "Sun"}, From: "00:00", To: "23:59", TimeZone: "America/Sao_Paulo"}}
	overnight := []v1alpha1.ApplyWindow{{From: "22:00", To: "02:00"}}

	tests := []struct {
		name      string
		in        Input
		allowed   bool
		blockedBy string
		policy    string
		next      time.Time
		inReason  string
		wouldStop []string
	}{
		{
			name:    "no policy allows",
			in:      Input{Now: mondayNoon},
			allowed: true,
		},
		{
			name: "inside a business hour window allows",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: businessHours}}),
			}},
			allowed: true,
		},
		{
			name: "a window closes: allowed again at the next opening",
			in: Input{Now: mondayNoon.Add(5 * time.Hour), Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: businessHours}}),
			}},
			blockedBy: "when.windows",
			policy:    "ClusterApplyPolicy/platform",
			next:      time.Date(2026, 11, 3, 13, 0, 0, 0, time.UTC),
			inReason:  "outside the apply windows",
		},
		{
			name: "a window on another day waits for that day",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/weekend", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: weekend}}),
			}},
			blockedBy: "when.windows",
			policy:    "ClusterApplyPolicy/weekend",
			next:      time.Date(2026, 11, 7, 3, 0, 0, 0, time.UTC),
		},
		{
			name:    "an overnight window covers after midnight",
			in:      Input{Now: time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC), Rules: []Rules{rules("ApplyPolicy/a/b", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: overnight}})}},
			allowed: true,
		},
		{
			name:      "an overnight window waits for its opening",
			in:        Input{Now: time.Date(2026, 11, 2, 21, 0, 0, 0, time.UTC), Rules: []Rules{rules("ApplyPolicy/a/b", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: overnight}})}},
			blockedBy: "when.windows",
			policy:    "ApplyPolicy/a/b",
			next:      time.Date(2026, 11, 2, 22, 0, 0, 0, time.UTC),
		},
		{
			name: "a freeze blocks and reports its reason",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon.Add(-time.Hour), mondayNoon.Add(time.Hour), "Sales event")},
				}}),
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/platform",
			next:      mondayNoon.Add(time.Hour),
			inReason:  "Sales event",
		},
		{
			name: "a freeze that ended does not block",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon.Add(-2*time.Hour), mondayNoon.Add(-time.Hour), "release")},
				}}),
			}},
			allowed: true,
		},
		{
			name: "a freeze ending outside the window waits for the window",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
					Windows: []v1alpha1.ApplyWindow{{Days: []string{"Mon", "Tue", "Wed", "Thu", "Fri"}, From: "10:00", To: "16:00"}},
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(5*time.Hour), "")},
				}}),
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/platform",
			next:      time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "the later of two freezes decides when it may apply",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(time.Hour), "")}}}),
				rules("ApplyPolicy/ns/b", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(3*time.Hour), "")}}}),
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/a",
			next:      mondayNoon.Add(3 * time.Hour),
		},
		{
			name: "cooldown counts from the last change",
			in: Input{Now: mondayNoon, LastChange: mondayNoon.Add(-time.Hour), Rules: []Rules{
				rules("ApplyPolicy/ns/checkout", v1alpha1.ApplyRules{Limits: &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: 24 * time.Hour}}}),
			}},
			blockedBy: "limits.cooldown",
			policy:    "ApplyPolicy/ns/checkout",
			next:      mondayNoon.Add(23 * time.Hour),
		},
		{
			name: "an expired cooldown does not block",
			in: Input{Now: mondayNoon, LastChange: mondayNoon.Add(-25 * time.Hour), Rules: []Rules{
				rules("ApplyPolicy/ns/checkout", v1alpha1.ApplyRules{Limits: &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: 24 * time.Hour}}}),
			}},
			allowed: true,
		},
		{
			name: "audit mode allows and reports what it would block",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/platform", v1alpha1.ApplyRules{Mode: v1alpha1.PolicyModeAudit, When: &v1alpha1.ApplyWhen{
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(time.Hour), "release")},
				}}),
			}},
			allowed:   true,
			wouldStop: []string{"ClusterApplyPolicy/platform: frozen: release (until 2026-11-02T16:00:00Z)"},
		},
		{
			name: "a policy that governs manual applies only leaves automatic ones alone",
			in: Input{Now: mondayNoon, Action: v1alpha1.AppliesToAuto, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{AppliesTo: []v1alpha1.AppliesTo{v1alpha1.AppliesToManual}, When: &v1alpha1.ApplyWhen{
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(time.Hour), "")},
				}}),
			}},
			allowed: true,
		},
		{
			name: "a policy that governs manual applies blocks them",
			in: Input{Now: mondayNoon, Action: v1alpha1.AppliesToManual, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{AppliesTo: []v1alpha1.AppliesTo{v1alpha1.AppliesToManual, v1alpha1.AppliesToAuto}, When: &v1alpha1.ApplyWhen{
					Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(time.Hour), "")},
				}}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(time.Hour),
		},
		{
			name: "a workload outside the selector is not governed",
			in: Input{Now: mondayNoon, Labels: map[string]string{"tier": "web"}, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{
					WorkloadSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "batch"}},
					When:             &v1alpha1.ApplyWhen{Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(time.Hour), "")}},
				}),
			}},
			allowed: true,
		},
		{
			name: "an unknown time zone blocks instead of opening the door",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
					Windows: []v1alpha1.ApplyWindow{{From: "10:00", To: "16:00", TimeZone: "Nowhere/Nothing"}},
				}}),
			}},
			blockedBy: "policy",
			policy:    "ApplyPolicy/ns/a",
			inReason:  "unknown time zone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.in)
			if err != nil {
				t.Fatalf("Evaluate: %v", err)
			}
			if got.Allowed != tt.allowed {
				t.Fatalf("Allowed = %v, want %v (decision %+v)", got.Allowed, tt.allowed, got)
			}
			if got.BlockedBy != tt.blockedBy {
				t.Errorf("BlockedBy = %q, want %q", got.BlockedBy, tt.blockedBy)
			}
			if got.Policy != tt.policy {
				t.Errorf("Policy = %q, want %q", got.Policy, tt.policy)
			}
			if !tt.next.IsZero() && !got.NextAllowed.Equal(tt.next) {
				t.Errorf("NextAllowed = %s, want %s", got.NextAllowed.UTC().Format(time.RFC3339), tt.next.UTC().Format(time.RFC3339))
			}
			if tt.next.IsZero() && !got.NextAllowed.IsZero() {
				t.Errorf("NextAllowed = %s, want none", got.NextAllowed.UTC().Format(time.RFC3339))
			}
			if tt.inReason != "" && !strings.Contains(got.Reason, tt.inReason) {
				t.Errorf("Reason = %q, want it to mention %q", got.Reason, tt.inReason)
			}
			if len(got.WouldBlock) != len(tt.wouldStop) {
				t.Fatalf("WouldBlock = %v, want %v", got.WouldBlock, tt.wouldStop)
			}
			for i := range tt.wouldStop {
				if got.WouldBlock[i] != tt.wouldStop[i] {
					t.Errorf("WouldBlock[%d] = %q, want %q", i, got.WouldBlock[i], tt.wouldStop[i])
				}
			}
		})
	}
}

func TestWindowContainsRejectsEqualEndpoints(t *testing.T) {
	w := v1alpha1.ApplyWindow{From: "10:00", To: "10:00"}
	ok, err := windowContains(w, mondayNoon)
	if err != nil {
		t.Fatalf("windowContains: %v", err)
	}
	if ok {
		t.Fatal("a window that opens and closes at the same minute must never be open")
	}
}
