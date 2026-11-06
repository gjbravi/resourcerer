package recommender

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	corev1 "k8s.io/api/core/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/settings"
)

var (
	workloadLabels = []string{"namespace", "workload_kind", "workload"}

	workloadsDesc = prometheus.NewDesc("resourcerer_workloads",
		"Managed workloads by kind, phase, whether the recommendation is pending an apply and whether auto-deploy is on.",
		[]string{"workload_kind", "phase", "pending", "auto_deploy"}, nil)
	costDesc = prometheus.NewDesc("resourcerer_workload_monthly_cost",
		"Monthly cost of the workload's requests now (current) and with the recommendation (recommended): requests × average replicas × rates × 730.",
		append(workloadLabels, "state"), nil)
	requestsDesc = prometheus.NewDesc("resourcerer_workload_requests",
		"Requests × average replicas now (current) and with the recommendation (recommended), in cores and bytes.",
		append(workloadLabels, "resource", "state"), nil)
	generatedDesc = prometheus.NewDesc("resourcerer_recommendation_generated_timestamp_seconds",
		"When the recommendation was last computed, in Unix seconds.", workloadLabels, nil)
	coverageDesc = prometheus.NewDesc("resourcerer_recommendation_history_coverage_ratio",
		"Usage history the recommendation was computed from, over its learning period: below 1 when Prometheus does not keep enough.",
		workloadLabels, nil)
	warningsDesc = prometheus.NewDesc("resourcerer_settings_warnings",
		"Settings of the workload that could not be used and fell back to their default; only workloads with some.", workloadLabels, nil)
	driftedDesc = prometheus.NewDesc("resourcerer_workload_drifted",
		"Containers whose resources no longer match the last recommendation applied: something else changed them since. Only workloads with some.",
		workloadLabels, nil)
)

// FleetCollector exports what the Overview shows from the Recommendations in the cache, when Prometheus scrapes:
// costs and requests now and with the recommendations, workloads by kind and phase, and whether each recommendation
// is fresh, complete and still applied. Read at scrape time rather than set by the reconciler, deleted workloads leave
// no series behind and the totals always match the per-workload series.
type FleetCollector struct {
	List func(context.Context) ([]v1alpha1.Recommendation, error)
	// Elected closes once this replica leads; until then it exports nothing, so replicas do not double every series.
	Elected <-chan struct{}
}

func (c FleetCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, d := range []*prometheus.Desc{workloadsDesc, costDesc, requestsDesc, generatedDesc, coverageDesc, warningsDesc, driftedDesc} {
		ch <- d
	}
}

func (c FleetCollector) Collect(ch chan<- prometheus.Metric) {
	if c.Elected != nil {
		select {
		case <-c.Elected:
		default:
			return
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	recs, err := c.List(ctx)
	if err != nil {
		ch <- prometheus.NewInvalidMetric(workloadsDesc, fmt.Errorf("list recommendations: %w", err))
		return
	}
	for _, m := range fleetMetrics(recs) {
		ch <- m
	}
}

func fleetMetrics(recs []v1alpha1.Recommendation) []prometheus.Metric {
	type group struct{ kind, phase, pending, auto string }
	counts := map[group]float64{}
	var out []prometheus.Metric
	gauge := func(d *prometheus.Desc, v float64, labels ...string) {
		out = append(out, prometheus.MustNewConstMetric(d, prometheus.GaugeValue, v, labels...))
	}
	for i := range recs {
		rec := &recs[i]
		st := rec.Status
		id := []string{rec.Namespace, rec.Spec.TargetRef.Kind, rec.Spec.TargetRef.Name}
		phase := string(st.Phase)
		if phase == "" {
			phase = "Unknown"
		}
		counts[group{rec.Spec.TargetRef.Kind, phase, strconv.FormatBool(st.Pending),
			strconv.FormatBool(st.Settings[settings.AutoDeploy].Value == "Enabled")}]++

		if im := st.Impact; im != nil {
			for state, t := range map[string]v1alpha1.Totals{"current": im.Current, "recommended": im.Recommended} {
				if v, err := strconv.ParseFloat(t.MonthlyCost, 64); err == nil {
					gauge(costDesc, v, append(id, state)...)
				}
				if v, err := strconv.ParseFloat(t.CPU, 64); err == nil {
					gauge(requestsDesc, v, append(id, "cpu", state)...)
				}
				if v, err := strconv.ParseFloat(t.Memory, 64); err == nil {
					gauge(requestsDesc, v, append(id, "memory", state)...)
				}
			}
		}
		if st.GeneratedAt != nil {
			gauge(generatedDesc, float64(st.GeneratedAt.Unix()), id...)
		}
		if ev := st.Evidence; ev != nil && ev.LearningSeconds > 0 {
			gauge(coverageDesc, min(1, float64(ev.HistorySeconds)/float64(ev.LearningSeconds)), id...)
		}
		if n := len(st.Warnings); n > 0 {
			gauge(warningsDesc, float64(n), id...)
		}
		live := make(map[string]corev1.ResourceRequirements, len(st.Containers))
		for _, c := range st.Containers {
			live[c.Name] = c.Current
		}
		if d := drifted(st.History, live); len(d) > 0 {
			gauge(driftedDesc, float64(len(d)), id...)
		}
	}
	for g, n := range counts {
		gauge(workloadsDesc, n, g.kind, g.phase, g.pending, g.auto)
	}
	return out
}
