package applier

import (
	"fmt"
	"slices"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// Selection checks the containers an apply is limited to. Empty means every container the recommendation changes.
// A partial apply is refused when the recommendation retargets an HPA: the new target keeps the HPA scaling at the
// same CPU per pod only with every container's CPU applied.
func Selection(st v1alpha1.RecommendationStatus, names []string) error {
	if len(names) == 0 {
		return nil
	}
	var optimized []string
	for _, c := range st.Containers {
		if !c.Skipped {
			optimized = append(optimized, c.Name)
		}
	}
	seen := map[string]bool{}
	for _, n := range names {
		switch {
		case seen[n]:
			return fmt.Errorf("container %q is listed twice", n)
		case !slices.ContainsFunc(st.Containers, func(c v1alpha1.ContainerRecommendation) bool { return c.Name == n }):
			return fmt.Errorf("the workload has no container %q", n)
		case !slices.Contains(optimized, n):
			return fmt.Errorf("container %q is not optimized", n)
		}
		seen[n] = true
	}
	if st.HPA != nil && len(seen) < len(optimized) {
		return fmt.Errorf("the recommendation changes HPA %s's target, which needs every container's CPU: apply all containers", st.HPA.Name)
	}
	return nil
}
