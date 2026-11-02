package server

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

func TestHeldByPolicy(t *testing.T) {
	now := time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)
	at := func(d time.Duration) *metav1.Time { t := metav1.NewTime(now.Add(d)); return &t }
	blocked := func(policy string, next *metav1.Time) *v1alpha1.ApplyPolicyStatus {
		return &v1alpha1.ApplyPolicyStatus{
			Auto:   v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateAllowed},
			Manual: v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateBlocked, Policy: policy, BlockedBy: "when.freezes", Reason: "frozen: launch", NextAllowed: next},
		}
	}
	tests := []struct {
		name   string
		status *v1alpha1.ApplyPolicyStatus
		want   string
	}{
		{name: "no decision recorded", status: nil},
		{name: "manual allowed", status: &v1alpha1.ApplyPolicyStatus{Manual: v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateAllowed}}},
		{
			name: "only the automatic apply is held",
			status: &v1alpha1.ApplyPolicyStatus{
				Auto:   v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateBlocked, Policy: "ApplyPolicy/demo/freeze"},
				Manual: v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateAllowed},
			},
		},
		{name: "held, and names the policy", status: blocked("ApplyPolicy/demo/freeze", at(time.Hour)), want: "ApplyPolicy/demo/freeze holds this apply: frozen: launch"},
		{name: "held until no known time", status: blocked("ApplyPolicy/demo/freeze", nil), want: "holds this apply"},
		{name: "held, by a policy that could not be read", status: blocked("", nil), want: "an apply policy holds this apply"},
		{name: "the block ended: left to the applier", status: blocked("ApplyPolicy/demo/freeze", at(-time.Second))},
		{name: "the block ends now: left to the applier", status: blocked("ApplyPolicy/demo/freeze", at(0))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := &v1alpha1.Recommendation{Status: v1alpha1.RecommendationStatus{ApplyPolicy: tt.status}}
			err := heldByPolicy(rec, now)
			if tt.want == "" {
				if err != nil {
					t.Errorf("heldByPolicy() = %v, want no refusal", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("heldByPolicy() = %v, want %q", err, tt.want)
			}
		})
	}
}
