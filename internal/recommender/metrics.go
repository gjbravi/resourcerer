package recommender

import (
	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// Exported so Prometheus keeps the history of recommendations; the UI charts it next to usage.
var (
	recommendationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_recommendation",
		Help: "Recommended requests and limits per container, in cores and bytes.",
	}, []string{"namespace", "workload_kind", "workload", "container", "resource", "type"})
	scoreGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_optimization_score",
		Help: "How closely current requests match the recommendation, 0-100.",
	}, []string{"namespace", "workload_kind", "workload"})
)

// What computing recommendations costs and how it ends: a reconcile that only finds the next one is not due yet is
// not counted, so these are the Prometheus-heavy ones. Failures to read Prometheus end in the Error phase, never in a
// reconcile error, so controller-runtime's own counters do not see them.
var (
	generationsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_recommendation_generations_total",
		Help: "Recommendations computed, by workload kind and the phase they ended in; Error when Prometheus could not be read.",
	}, []string{"workload_kind", "phase"})
	generationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "resourcerer_recommendation_generation_duration_seconds",
		Help:    "Time to compute one recommendation, Prometheus queries included.",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60, 120},
	}, []string{"workload_kind"})
	signalErrorsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_signal_errors_total",
		Help: "Reliability signal queries that failed; the signal is left out of the workload's status.",
	}, []string{"signal"})
	overwrittenCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "resourcerer_apply_overwritten_total",
		Help: "Applied recommendations whose resources another field manager changed afterwards, by that manager.",
	}, []string{"manager"})
)

func init() {
	metrics.Registry.MustRegister(recommendationGauge, scoreGauge, generationsCounter, generationSeconds, signalErrorsCounter,
		overwrittenCounter)
}

func export(rec *v1alpha1.Recommendation) {
	ns, kind, name := rec.Namespace, rec.Spec.TargetRef.Kind, rec.Spec.TargetRef.Name
	forget(ns, kind, name)
	if rec.Status.Phase == v1alpha1.PhasePending || rec.Status.Phase == v1alpha1.PhaseUnsupported {
		return
	}
	for _, c := range rec.Status.Containers {
		for typ, l := range map[string]corev1.ResourceList{"request": c.Recommended.Requests, "limit": c.Recommended.Limits} {
			for _, r := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
				if q, ok := l[r]; ok {
					recommendationGauge.WithLabelValues(ns, kind, name, c.Name, string(r), typ).Set(q.AsApproximateFloat64())
				}
			}
		}
	}
	if rec.Status.Score != nil {
		scoreGauge.WithLabelValues(ns, kind, name).Set(float64(rec.Status.Score.Overall))
	}
}

func forget(ns, kind, name string) {
	l := prometheus.Labels{"namespace": ns, "workload_kind": kind, "workload": name}
	recommendationGauge.DeletePartialMatch(l)
	scoreGauge.DeletePartialMatch(l)
}
