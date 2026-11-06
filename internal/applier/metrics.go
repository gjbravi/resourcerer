package applier

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// These show applies happening, applies being held, applies being undone and how far a policy is from its
// concurrency limit: together they show a herd forming when a window opens, the counterpart of Karpenter's
// allowed disruptions.
var (
	applyCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_apply_total",
		Help: "Applies and reverts by action (Apply, Revert), trigger (manual: a person; auto: auto-deploy) and result: Succeeded, Failed or Rejected.",
	}, []string{"action", "trigger", "result"})
	deferredCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_apply_deferred_total",
		Help: "Automatic applies held by an apply policy, by the rule that held them.",
	}, []string{"reason"})
	revertedCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_apply_reverted_total",
		Help: "Reverts by reason. Every revert in this version is one a person asked for.",
	}, []string{"reason"})
	slotsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_applypolicy_slots_available",
		Help: "Applies a policy's maxConcurrent still allows among the workloads it selects.",
	}, []string{"policy"})
)

func init() {
	metrics.Registry.MustRegister(applyCounter, deferredCounter, revertedCounter, slotsGauge)
}

// countApply counts an apply or a revert; ar is the request a person made, nil for auto-deploy.
func countApply(ar *v1alpha1.ApplyRequest, result string) {
	action, trigger := string(v1alpha1.ActionApply), "auto"
	if ar != nil {
		trigger = "manual"
		if ar.Action == v1alpha1.ActionRevert {
			action = string(v1alpha1.ActionRevert)
		}
	}
	applyCounter.WithLabelValues(action, trigger, result).Inc()
}

// deferApply counts a held apply under the short reason a metric reader filters by, rather than the spec
// field path the status carries.
func deferApply(blockedBy string) {
	reason := map[string]string{
		"windows":            "window",
		"freezes":            "freeze",
		"rate.cooldown":      "cooldown",
		"rate.maxConcurrent": "concurrency",
		"rate.perNamespace":  "namespace",
		"rate.maxPerHour":    "rate",
		"rate.spread":        "spread",
		"rollout":            "rollout",
	}[blockedBy]
	if reason == "" {
		reason = "policy"
	}
	deferredCounter.WithLabelValues(reason).Inc()
}
