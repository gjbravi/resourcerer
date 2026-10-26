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

func init() {
	metrics.Registry.MustRegister(recommendationGauge, scoreGauge)
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
