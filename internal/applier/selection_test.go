package applier

import (
	"testing"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

func TestSelection(t *testing.T) {
	st := v1alpha1.RecommendationStatus{Containers: []v1alpha1.ContainerRecommendation{
		{Name: "app"}, {Name: "worker"}, {Name: "istio-proxy", Skipped: true},
	}}
	withHPA := st
	withHPA.HPA = &v1alpha1.HPARecommendation{Name: "orders", Current: 70, Recommended: 80}
	cases := []struct {
		name  string
		st    v1alpha1.RecommendationStatus
		names []string
		err   string
	}{
		{name: "empty applies all", st: st},
		{name: "one container", st: st, names: []string{"worker"}},
		{name: "every optimized container", st: st, names: []string{"app", "worker"}},
		{name: "unknown container", st: st, names: []string{"db"}, err: `the workload has no container "db"`},
		{name: "skipped container", st: st, names: []string{"istio-proxy"}, err: `container "istio-proxy" is not optimized`},
		{name: "twice", st: st, names: []string{"app", "app"}, err: `container "app" is listed twice`},
		{name: "partial with an HPA", st: withHPA, names: []string{"app"},
			err: "the recommendation changes HPA orders's target, which needs every container's CPU: apply all containers"},
		{name: "all with an HPA", st: withHPA, names: []string{"worker", "app"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Selection(c.st, c.names)
			if c.err == "" && err != nil || c.err != "" && (err == nil || err.Error() != c.err) {
				t.Errorf("Selection(%v) = %v, want %q", c.names, err, c.err)
			}
		})
	}
}
