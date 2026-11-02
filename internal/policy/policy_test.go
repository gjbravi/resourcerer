package policy

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// mondayNoon is a Monday, 2026-11-02 15:00 UTC, used as "now" unless a case says otherwise.
var mondayNoon = time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)

func freeze(from, to time.Time, reason string) v1alpha1.ApplyFreeze {
	return v1alpha1.ApplyFreeze{From: metav1.NewTime(from), To: metav1.NewTime(to), Reason: reason}
}

func rules(name string, spec v1alpha1.ApplyRules) Rules { return Rules{Name: name, Spec: spec} }

func TestEvaluate(t *testing.T) {
	businessHours := []v1alpha1.ApplyWindow{{Days: []string{"Mon", "Tue", "Wed", "Thu", "Fri"}, From: "13:00", To: "19:00"}}
	weekend := []v1alpha1.ApplyWindow{{Days: []string{"Sat", "Sun"}, From: "00:00", To: "23:59"}}
	overnight := []v1alpha1.ApplyWindow{{From: "22:00", To: "02:00"}}
	fridayNight := []v1alpha1.ApplyWindow{{Days: []string{"Fri"}, From: "22:00", To: "02:00"}}
	freezeFor := func(from, to time.Time) *v1alpha1.ApplyWhen {
		return &v1alpha1.ApplyWhen{Freezes: []v1alpha1.ApplyFreeze{freeze(from, to, "")}}
	}
	cooldown := func(d time.Duration) *v1alpha1.ApplyLimits {
		return &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: d}}
	}
	badSelector := &metav1.LabelSelector{MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "tier", Operator: "Sometimes"}}}

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
			next:      time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC),
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
			name: "a window that cannot be read blocks instead of opening the door",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
					Windows: []v1alpha1.ApplyWindow{{From: "9am", To: "16:00"}},
				}}),
			}},
			blockedBy: "policy",
			policy:    "ApplyPolicy/ns/a",
			inReason:  "not a time of day",
		},
		{
			name: "a policy that allows now still counts when another's freeze ends",
			in: Input{Now: time.Date(2026, 11, 2, 8, 30, 0, 0, time.UTC), Rules: []Rules{
				rules("ClusterApplyPolicy/freeze", v1alpha1.ApplyRules{When: freezeFor(time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC), time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC))}),
				rules("ApplyPolicy/ns/morning", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "08:00", To: "09:00"}}}}),
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/freeze",
			next:      time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "a freeze hands over to a cooldown",
			in: Input{Now: mondayNoon, LastChange: mondayNoon.Add(-time.Hour), Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(2*time.Hour)), Limits: cooldown(6 * time.Hour)}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(5 * time.Hour),
		},
		{
			name: "windows are a union: the earliest opening wins",
			in: Input{Now: time.Date(2026, 11, 2, 7, 0, 0, 0, time.UTC), Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "14:00", To: "15:00"}, {From: "09:00", To: "10:00"}}}}),
			}},
			blockedBy: "when.windows",
			policy:    "ApplyPolicy/ns/a",
			next:      time.Date(2026, 11, 2, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "to is exclusive",
			in: Input{Now: time.Date(2026, 11, 2, 16, 0, 0, 0, time.UTC), Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "10:00", To: "16:00"}}}}),
			}},
			blockedBy: "when.windows",
			policy:    "ApplyPolicy/ns/a",
			next:      time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC),
		},
		{
			name: "a weekly window a week away",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{Days: []string{"Mon"}, From: "05:00", To: "06:00"}}}}),
			}},
			blockedBy: "when.windows",
			policy:    "ApplyPolicy/ns/a",
			next:      time.Date(2026, 11, 9, 5, 0, 0, 0, time.UTC),
		},
		{
			name:    "an overnight window belongs to the day it opens: its tail runs into the next day",
			in:      Input{Now: time.Date(2026, 11, 7, 1, 0, 0, 0, time.UTC), Rules: []Rules{rules("ApplyPolicy/a/b", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: fridayNight}})}},
			allowed: true,
		},
		{
			name:      "an overnight window does not open on the day after its day",
			in:        Input{Now: time.Date(2026, 11, 7, 22, 30, 0, 0, time.UTC), Rules: []Rules{rules("ApplyPolicy/a/b", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: fridayNight}})}},
			blockedBy: "when.windows",
			policy:    "ApplyPolicy/a/b",
			next:      time.Date(2026, 11, 13, 22, 0, 0, 0, time.UTC),
		},
		{
			name: "a freeze without a reason still says until when",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(time.Hour),
			inReason:  "frozen until 2026-11-02T16:00:00Z",
		},
		{
			name: "a freeze's to is exclusive",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon.Add(-time.Hour), mondayNoon)}),
			}},
			allowed: true,
		},
		{
			name: "the first blocking policy in order is named",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(3*time.Hour))}),
				rules("ApplyPolicy/ns/b", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/a",
			next:      mondayNoon.Add(3 * time.Hour),
		},
		{
			name: "an audit policy that cannot be interpreted does not block",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/try", v1alpha1.ApplyRules{Mode: v1alpha1.PolicyModeAudit, When: &v1alpha1.ApplyWhen{
					Windows: []v1alpha1.ApplyWindow{{From: "9am", To: "16:00"}},
				}}),
			}},
			allowed:   true,
			wouldStop: []string{`ApplyPolicy/ns/try: "9am" is not a time of day like 10:00`},
		},
		{
			name: "audit and enforce together: blocked, and the audit is still reported",
			in: Input{Now: mondayNoon, LastChange: mondayNoon.Add(-time.Hour), Rules: []Rules{
				rules("ClusterApplyPolicy/try", v1alpha1.ApplyRules{Mode: v1alpha1.PolicyModeAudit, Limits: cooldown(48 * time.Hour)}),
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(time.Hour),
			wouldStop: []string{"ClusterApplyPolicy/try: changed less than 48h0m0s ago, cooling down until 2026-11-04T14:00:00Z"},
		},
		{
			name: "an audit policy that allows reports nothing",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/try", v1alpha1.ApplyRules{Mode: v1alpha1.PolicyModeAudit, When: &v1alpha1.ApplyWhen{Windows: businessHours}}),
			}},
			allowed: true,
		},
		{
			name: "a window that never opens blocks, with no time to wait for",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "10:00", To: "10:00"}}}}),
			}},
			blockedBy: "policy",
			policy:    "ApplyPolicy/ns/a",
			inReason:  "never opens",
		},
		{
			name: "a policy that cannot be interpreted is named over one that blocks for a while",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ClusterApplyPolicy/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
				rules("ApplyPolicy/ns/broken", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "25:00", To: "26:00"}}}}),
			}},
			blockedBy: "policy",
			policy:    "ApplyPolicy/ns/broken",
			inReason:  "is not a time of day",
		},
		{
			name: "a malformed workload selector blocks instead of selecting nothing",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{WorkloadSelector: badSelector, When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			blockedBy: "policy",
			policy:    "ApplyPolicy/ns/a",
			inReason:  "workloadSelector",
		},
		{
			name: "a malformed namespace selector blocks instead of selecting nothing",
			in: Input{Now: mondayNoon, Rules: []Rules{
				{Name: "ClusterApplyPolicy/a", NamespaceSelector: badSelector, Spec: v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}},
			}},
			blockedBy: "policy",
			policy:    "ClusterApplyPolicy/a",
			inReason:  "namespaceSelector",
		},
		{
			name: "a cluster policy for other namespaces does not govern",
			in: Input{Now: mondayNoon, NamespaceLabels: map[string]string{"team": "web"}, Rules: []Rules{
				{Name: "ClusterApplyPolicy/a", NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"team": "payments"}}, Spec: v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}},
			}},
			allowed: true,
		},
		{
			name: "a cluster policy for this namespace governs",
			in: Input{Now: mondayNoon, NamespaceLabels: map[string]string{"team": "payments"}, Rules: []Rules{
				{Name: "ClusterApplyPolicy/a", NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"team": "payments"}}, Spec: v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}},
			}},
			blockedBy: "when.freezes",
			policy:    "ClusterApplyPolicy/a",
			next:      mondayNoon.Add(time.Hour),
		},
		{
			name: "a selector that matches the workload governs it",
			in: Input{Now: mondayNoon, Labels: map[string]string{"tier": "batch"}, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{
					WorkloadSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "batch"}},
					When:             freezeFor(mondayNoon, mondayNoon.Add(time.Hour)),
				}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(time.Hour),
		},
		{
			name: "no action means an automatic apply",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			blockedBy: "when.freezes",
			policy:    "ApplyPolicy/ns/a",
			next:      mondayNoon.Add(time.Hour),
		},
		{
			name: "a policy with no appliesTo leaves a person free",
			in: Input{Now: mondayNoon, Action: v1alpha1.AppliesToManual, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: freezeFor(mondayNoon, mondayNoon.Add(time.Hour))}),
			}},
			allowed: true,
		},
		{
			name: "a cooldown with no change yet does not block",
			in: Input{Now: mondayNoon, Rules: []Rules{
				rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{Limits: cooldown(24 * time.Hour)}),
			}},
			allowed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.in)
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
			if !got.Allowed && !got.NextAllowed.IsZero() {
				// What NextAllowed promises: allowed then, and still blocked a minute before.
				if at := Evaluate(withNow(tt.in, got.NextAllowed)); !at.Allowed {
					t.Errorf("at NextAllowed %s still blocked by %s (%s)", got.NextAllowed.UTC().Format(time.RFC3339), at.Policy, at.BlockedBy)
				}
				if before := got.NextAllowed.Add(-time.Minute); before.After(tt.in.Now) && Evaluate(withNow(tt.in, before)).Allowed {
					t.Errorf("allowed at %s, before NextAllowed", before.UTC().Format(time.RFC3339))
				}
			}
		})
	}
}

func withNow(in Input, t time.Time) Input {
	in.Now = t
	return in
}

// The decision a reconcile records must not change while nothing changes: one that moved with the clock
// would be written to the status on every pass, and every write is another reconcile.
func TestNextAllowedDoesNotMoveWithTheClock(t *testing.T) {
	in := Input{Now: mondayNoon, LastChange: mondayNoon.Add(-time.Hour), Rules: []Rules{
		rules("ClusterApplyPolicy/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
			Windows: []v1alpha1.ApplyWindow{{Days: []string{"Tue", "Thu"}, From: "10:00", To: "11:00"}},
			Freezes: []v1alpha1.ApplyFreeze{freeze(mondayNoon, mondayNoon.Add(30*time.Hour), "")},
		}}),
		rules("ApplyPolicy/ns/b", v1alpha1.ApplyRules{Limits: &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: 72 * time.Hour}}}),
	}}
	first := Evaluate(in)
	if first.Allowed || first.NextAllowed.IsZero() {
		t.Fatalf("want a block with a time to wait for, got %+v", first)
	}
	for _, d := range []time.Duration{time.Second, 7 * time.Minute, 3 * time.Hour} {
		if got := Evaluate(withNow(in, in.Now.Add(d))); !got.NextAllowed.Equal(first.NextAllowed) || got.Reason != first.Reason {
			t.Errorf("%s later: NextAllowed %s and reason %q, want %s and %q", d, got.NextAllowed, got.Reason, first.NextAllowed, first.Reason)
		}
	}
}

// nextAllowed jumps from blocker to blocker; the walk tries every minute. They have to agree on random
// policies mixing windows, overnight ones included, freezes, cooldowns and audit.
func TestNextAllowedMatchesAMinuteByMinuteWalk(t *testing.T) {
	starts := []time.Time{
		time.Date(2026, 2, 26, 0, 0, 0, 0, time.UTC),  // across the end of a month
		time.Date(2026, 12, 29, 0, 0, 0, 0, time.UTC), // across the end of a year
		mondayNoon,
	}
	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	rng := rand.New(rand.NewPCG(1, 2))
	hhmm := func() string { return fmt.Sprintf("%02d:%02d", rng.IntN(24), rng.IntN(4)*15) }
	const horizon = 10 * 24 * time.Hour

	for i := range 300 {
		now := starts[rng.IntN(len(starts))].Add(time.Duration(rng.IntN(7*24*60)) * time.Minute)
		var rs []Rules
		for j := range 1 + rng.IntN(3) {
			spec := v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{}}
			for range rng.IntN(3) {
				w := v1alpha1.ApplyWindow{From: hhmm(), To: hhmm()}
				if w.From == w.To {
					continue
				}
				for _, d := range days {
					if rng.IntN(2) == 0 {
						w.Days = append(w.Days, d)
					}
				}
				spec.When.Windows = append(spec.When.Windows, w)
			}
			for range rng.IntN(2) {
				from := now.Add(time.Duration(rng.IntN(72*60)-24*60) * time.Minute)
				spec.When.Freezes = append(spec.When.Freezes, freeze(from, from.Add(time.Duration(1+rng.IntN(72*60))*time.Minute), ""))
			}
			if rng.IntN(3) == 0 {
				spec.Limits = &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: time.Duration(1+rng.IntN(72)) * time.Hour}}
			}
			if rng.IntN(5) == 0 {
				spec.Mode = v1alpha1.PolicyModeAudit
			}
			rs = append(rs, rules(fmt.Sprintf("ApplyPolicy/ns/p%d", j), spec))
		}
		in := Input{Now: now, LastChange: now.Add(-time.Duration(rng.IntN(48*60)) * time.Minute), Rules: rs}

		got := Evaluate(in)
		want, found := walk(t, in, horizon)
		switch {
		case !found && got.Allowed:
			t.Fatalf("case %d: allowed, but the walk finds no open minute in %s%s", i, horizon, describe(rs))
		case !found:
			continue // beyond the walk: the search may still have found one
		case want.Equal(now) != got.Allowed:
			t.Fatalf("case %d: Allowed = %v, the walk opens at %s (now %s)%s", i, got.Allowed, want, now, describe(rs))
		case !got.Allowed && !got.NextAllowed.Equal(want):
			t.Fatalf("case %d at %s: NextAllowed = %s, the walk opens at %s%s", i, now.UTC().Format(time.RFC3339),
				got.NextAllowed.UTC().Format(time.RFC3339), want.UTC().Format(time.RFC3339), describe(rs))
		}
	}
}

// walk is the first minute from in.Now at which every enforcing policy allows, tried one minute at a
// time. Windows and freezes in the random cases fall on whole minutes, so a minute is fine enough.
func walk(t *testing.T, in Input, horizon time.Duration) (time.Time, bool) {
	t.Helper()
	var enforcing []compiled
	for _, r := range in.Rules {
		if r.Spec.Mode == v1alpha1.PolicyModeAudit {
			continue
		}
		c, _, err := compile(r, nil, nil)
		if err != nil {
			t.Fatalf("compile %s: %v", r.Name, err)
		}
		enforcing = append(enforcing, c)
	}
	open := func(at time.Time) bool {
		for _, c := range enforcing {
			if _, blocks := c.blocksAt(at, in.LastChange); blocks {
				return false
			}
		}
		return true
	}
	if open(in.Now) {
		return in.Now, true
	}
	for at := in.Now.Truncate(time.Minute).Add(time.Minute); at.Before(in.Now.Add(horizon)); at = at.Add(time.Minute) {
		if open(at) {
			return at, true
		}
	}
	return time.Time{}, false
}

func describe(rs []Rules) string {
	var b strings.Builder
	for _, r := range rs {
		fmt.Fprintf(&b, "\n  %s mode=%s", r.Name, r.Spec.Mode)
		for _, w := range r.Spec.When.Windows {
			fmt.Fprintf(&b, " window=%v %s-%s", w.Days, w.From, w.To)
		}
		for _, f := range r.Spec.When.Freezes {
			fmt.Fprintf(&b, " freeze=%s..%s", f.From.UTC().Format(time.RFC3339), f.To.UTC().Format(time.RFC3339))
		}
		if r.Spec.Limits != nil {
			fmt.Fprintf(&b, " cooldown=%s", r.Spec.Limits.Cooldown.Duration)
		}
	}
	return b.String()
}

func TestCompileWindowRejects(t *testing.T) {
	for _, tt := range []struct {
		name string
		w    v1alpha1.ApplyWindow
		want string
	}{
		{"equal endpoints", v1alpha1.ApplyWindow{From: "10:00", To: "10:00"}, "never opens"},
		{"a time that is not one", v1alpha1.ApplyWindow{From: "9am", To: "10:00"}, "not a time of day"},
		{"an unknown day", v1alpha1.ApplyWindow{Days: []string{"Monday"}, From: "09:00", To: "10:00"}, "not a day"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := compileWindow(tt.w); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("compileWindow() error = %v, want one mentioning %q", err, tt.want)
			}
		})
	}
}

// A weekly window, decided right after it closed: the longest search, and what every
// reconcile pays for a workload such a policy holds.
func BenchmarkEvaluateWeeklyWindow(b *testing.B) {
	in := Input{Now: mondayNoon, Rules: []Rules{
		rules("ApplyPolicy/ns/a", v1alpha1.ApplyRules{When: &v1alpha1.ApplyWhen{
			Windows: []v1alpha1.ApplyWindow{{Days: []string{"Mon"}, From: "05:00", To: "06:00"}},
		}}),
	}}
	for b.Loop() {
		Evaluate(in)
	}
}
