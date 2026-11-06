package cluster

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Whether the costs everywhere else can be trusted: a calibration that keeps failing leaves them on the last good
// rates or the configured ones, and only the Settings page would say so.
var (
	calibrationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_cost_calibration",
		Help: "1 for where cost rates come from now (source: static or karpenter) and the state of the last calibration (ok, stale, fallback, error).",
	}, []string{"source", "state"})
	rateGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "resourcerer_cost_rate",
		Help: "Cost rates in use: per core-hour for cpu, per GiB-hour for memory.",
	}, []string{"resource"})
	factorGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "resourcerer_cost_cluster_factor",
		Help: "Rates in use over the configured rates: what capacity costs on this cluster relative to cost.rates.",
	})
)

func init() {
	metrics.Registry.MustRegister(calibrationGauge, rateGauge, factorGauge)
}

func report(p Pricing) Pricing {
	calibrationGauge.Reset()
	calibrationGauge.WithLabelValues(p.Source, p.State).Set(1)
	rateGauge.WithLabelValues("cpu").Set(p.Rates.CPUCoreHour)
	rateGauge.WithLabelValues("memory").Set(p.Rates.MemoryGiBHour)
	factorGauge.Set(p.ClusterFactor())
	return p
}
