package applier

import (
	"errors"
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
)

func TestSortRules(t *testing.T) {
	tests := []struct {
		name  string
		rules []policy.Rules
		want  []string
	}{{
		name: "higher priority first, then name",
		rules: []policy.Rules{
			{Name: "ApplyPolicy/demo/b", Spec: v1alpha1.ApplyRules{Priority: 10}},
			{Name: "ClusterApplyPolicy/platform", Spec: v1alpha1.ApplyRules{Priority: 50}},
			{Name: "ApplyPolicy/demo/a", Spec: v1alpha1.ApplyRules{Priority: 10}},
		},
		want: []string{"ClusterApplyPolicy/platform", "ApplyPolicy/demo/a", "ApplyPolicy/demo/b"},
	}, {
		name: "equal priorities and names keep their order",
		rules: []policy.Rules{
			{Name: "ApplyPolicy/demo/a"},
			{Name: "ApplyPolicy/demo/a"},
		},
		want: []string{"ApplyPolicy/demo/a", "ApplyPolicy/demo/a"},
	}, {
		name:  "no policies",
		rules: nil,
		want:  nil,
	}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sortRules(tc.rules)
			if len(tc.rules) != len(tc.want) {
				t.Fatalf("got %d rules, want %d", len(tc.rules), len(tc.want))
			}
			for i, want := range tc.want {
				if tc.rules[i].Name != want {
					t.Errorf("rule %d is %q, want %q", i, tc.rules[i].Name, want)
				}
			}
		})
	}
}

func TestSameDecision(t *testing.T) {
	now := metav1.NewTime(time.Date(2026, 11, 2, 12, 0, 0, 0, time.UTC))
	later := metav1.NewTime(now.Add(time.Hour))
	blocked := func() *v1alpha1.ApplyDecision {
		return &v1alpha1.ApplyDecision{
			State:       v1alpha1.PolicyStateBlocked,
			BlockedBy:   "when.freezes",
			Policy:      "ApplyPolicy/demo/freeze",
			Reason:      "frozen",
			EvaluatedAt: &now,
		}
	}

	tests := []struct {
		name string
		a, b *v1alpha1.ApplyDecision
		want bool
	}{
		{name: "both absent", want: true},
		{name: "one absent", a: blocked(), want: false},
		{name: "one absent, the other way", b: blocked(), want: false},
		{name: "same decision, later evaluation", a: blocked(), b: func() *v1alpha1.ApplyDecision {
			d := blocked()
			d.EvaluatedAt = &later
			return d
		}(), want: true},
		{name: "different state", a: blocked(), b: &v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateAllowed}, want: false},
		{name: "different blocking field", a: blocked(), b: func() *v1alpha1.ApplyDecision {
			d := blocked()
			d.BlockedBy = "when.windows"
			return d
		}(), want: false},
		{name: "a window opens later", a: blocked(), b: func() *v1alpha1.ApplyDecision {
			d := blocked()
			d.NextAllowed = &later
			return d
		}(), want: false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := sameDecision(tc.a, tc.b); got != tc.want {
				t.Errorf("sameDecision() = %v, want %v", got, tc.want)
			}
			if got := sameDecision(tc.b, tc.a); got != tc.want {
				t.Errorf("sameDecision() reversed = %v, want %v", got, tc.want)
			}
		})
	}
}

// A policy that cannot be read has to stop both kinds of apply: the alternative is applying while the
// rules are unknown.
func TestFailClosed(t *testing.T) {
	d := failClosed(errors.New("list apply policies in demo: unknown unit"))
	for name, got := range map[string]policy.Decision{"auto": d.auto, "manual": d.manual} {
		if got.Allowed {
			t.Errorf("%s: allowed, want blocked", name)
		}
		if !strings.Contains(got.Reason, "could not be read") || !strings.Contains(got.Reason, "unknown unit") {
			t.Errorf("%s: reason %q does not say what could not be read", name, got.Reason)
		}
	}
}
