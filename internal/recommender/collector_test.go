package recommender

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/settings"
)

func rec(ns, kind, name string, st v1alpha1.RecommendationStatus) v1alpha1.Recommendation {
	return v1alpha1.Recommendation{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{Kind: kind, Name: name}},
		Status:     st,
	}
}

func fleet() []v1alpha1.Recommendation {
	generated := metav1.NewTime(time.Unix(1_800_000_000, 0))
	auto := map[string]v1alpha1.Setting{settings.AutoDeploy: {Value: "Enabled"}}
	return []v1alpha1.Recommendation{
		rec("payments", "Deployment", "checkout", v1alpha1.RecommendationStatus{
			Phase: v1alpha1.PhaseReady, Pending: true, Settings: auto, GeneratedAt: &generated,
			Impact: &v1alpha1.Impact{
				Current:     v1alpha1.Totals{CPU: "2.000", Memory: "4294967296", MonthlyCost: "120.00"},
				Recommended: v1alpha1.Totals{CPU: "0.500", Memory: "2147483648", MonthlyCost: "40.00"},
			},
			Evidence: &v1alpha1.Evidence{HistorySeconds: 302400, LearningSeconds: 604800},
		}),
		rec("demo", "Deployment", "idle", v1alpha1.RecommendationStatus{Phase: v1alpha1.PhasePending}),
		rec("demo", "StatefulSet", "cache", v1alpha1.RecommendationStatus{
			Phase:    v1alpha1.PhaseReady,
			Warnings: []string{"memory.optimization-goal: unknown goal Cheap"},
			Containers: []v1alpha1.ContainerRecommendation{
				{Name: "redis", Current: corev1.ResourceRequirements{Requests: list("cpu", "500m")}},
			},
			History: []v1alpha1.HistoryEntry{{Action: v1alpha1.ActionApply, Containers: []v1alpha1.ContainerChange{
				{Name: "redis", To: corev1.ResourceRequirements{Requests: list("cpu", "100m")}},
			}}},
		}),
	}
}

func TestFleetCollectorExportsWhatTheOverviewShows(t *testing.T) {
	c := FleetCollector{List: func(context.Context) ([]v1alpha1.Recommendation, error) { return fleet(), nil }}
	want := `
# HELP resourcerer_recommendation_generated_timestamp_seconds When the recommendation was last computed, in Unix seconds.
# TYPE resourcerer_recommendation_generated_timestamp_seconds gauge
resourcerer_recommendation_generated_timestamp_seconds{namespace="payments",workload="checkout",workload_kind="Deployment"} 1.8e+09
# HELP resourcerer_recommendation_history_coverage_ratio Usage history the recommendation was computed from, over its learning period: below 1 when Prometheus does not keep enough.
# TYPE resourcerer_recommendation_history_coverage_ratio gauge
resourcerer_recommendation_history_coverage_ratio{namespace="payments",workload="checkout",workload_kind="Deployment"} 0.5
# HELP resourcerer_settings_warnings Settings of the workload that could not be used and fell back to their default; only workloads with some.
# TYPE resourcerer_settings_warnings gauge
resourcerer_settings_warnings{namespace="demo",workload="cache",workload_kind="StatefulSet"} 1
# HELP resourcerer_workload_drifted Containers whose resources no longer match the last recommendation applied: something else changed them since. Only workloads with some.
# TYPE resourcerer_workload_drifted gauge
resourcerer_workload_drifted{namespace="demo",workload="cache",workload_kind="StatefulSet"} 1
# HELP resourcerer_workload_monthly_cost Monthly cost of the workload's requests now (current) and with the recommendation (recommended): requests × average replicas × rates × 730.
# TYPE resourcerer_workload_monthly_cost gauge
resourcerer_workload_monthly_cost{namespace="payments",state="current",workload="checkout",workload_kind="Deployment"} 120
resourcerer_workload_monthly_cost{namespace="payments",state="recommended",workload="checkout",workload_kind="Deployment"} 40
# HELP resourcerer_workload_requests Requests × average replicas now (current) and with the recommendation (recommended), in cores and bytes.
# TYPE resourcerer_workload_requests gauge
resourcerer_workload_requests{namespace="payments",resource="cpu",state="current",workload="checkout",workload_kind="Deployment"} 2
resourcerer_workload_requests{namespace="payments",resource="cpu",state="recommended",workload="checkout",workload_kind="Deployment"} 0.5
resourcerer_workload_requests{namespace="payments",resource="memory",state="current",workload="checkout",workload_kind="Deployment"} 4.294967296e+09
resourcerer_workload_requests{namespace="payments",resource="memory",state="recommended",workload="checkout",workload_kind="Deployment"} 2.147483648e+09
# HELP resourcerer_workloads Managed workloads by kind, phase, whether the recommendation is pending an apply and whether auto-deploy is on.
# TYPE resourcerer_workloads gauge
resourcerer_workloads{auto_deploy="false",pending="false",phase="Pending",workload_kind="Deployment"} 1
resourcerer_workloads{auto_deploy="false",pending="false",phase="Ready",workload_kind="StatefulSet"} 1
resourcerer_workloads{auto_deploy="true",pending="true",phase="Ready",workload_kind="Deployment"} 1
`
	if err := testutil.CollectAndCompare(c, strings.NewReader(want)); err != nil {
		t.Error(err)
	}
}

func TestFleetCollectorWaitsToLead(t *testing.T) {
	elected := make(chan struct{})
	c := FleetCollector{Elected: elected, List: func(context.Context) ([]v1alpha1.Recommendation, error) { return fleet(), nil }}
	if n := testutil.CollectAndCount(c); n != 0 {
		t.Errorf("a replica that does not lead exported %d series", n)
	}
	close(elected)
	if n := testutil.CollectAndCount(c, "resourcerer_workloads"); n != 3 {
		t.Errorf("the leader exported %d workload groups, want 3", n)
	}
}

func TestFleetCollectorReportsAFailedList(t *testing.T) {
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(FleetCollector{List: func(context.Context) ([]v1alpha1.Recommendation, error) {
		return nil, errors.New("cache not synced")
	}})
	if _, err := reg.Gather(); err == nil || !strings.Contains(err.Error(), "cache not synced") {
		t.Errorf("a failed list should fail the scrape with its reason, got %v", err)
	}
}
