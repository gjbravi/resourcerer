package applier

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// These show applies happening, applies being held, applies being undone and how far a policy is from its
// concurrency limit: together they show a herd forming when a window opens, the counterpart of Karpenter's
// allowed disruptions.
var (
	applyCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_apply_total",
		Help: "Applies and reverts by result: Succeeded, Failed or Rejected.",
	}, []string{"result"})
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
	// An installation that skipped the chart's CRDs, or kept an older one, accepts writes and drops what it does
	// not know: policies are not read at all, or a decision never reaches the status. Neither logs an error on its
	// own, so this is the one signal an alert can be built on.
	crdCurrentGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_crd_current",
		Help: "1 when the CRD is installed and keeps the fields this binary writes to it, 0 when it is missing or older than the binary.",
	}, []string{"crd"})
)

func init() {
	metrics.Registry.MustRegister(applyCounter, deferredCounter, revertedCounter, slotsGauge, crdCurrentGauge)
}

func countApply(result string) { applyCounter.WithLabelValues(result).Inc() }

func setCRDCurrent(crd string, current bool) {
	if current {
		crdCurrentGauge.WithLabelValues(crd).Set(1)
		return
	}
	crdCurrentGauge.WithLabelValues(crd).Set(0)
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
