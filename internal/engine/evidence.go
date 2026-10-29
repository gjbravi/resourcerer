package engine

import "math"

// usedPercentile is the usage the "used" totals count: what the workload needs most of the time.
const usedPercentile = 0.95

// Evidence is how the recommendation would have held up against the history it was computed from.
type Evidence struct {
	// Above is the share of samples in which the busiest pod's usage passed the recommended request, for the
	// worst container; Peak is that container's highest usage over the recommended request.
	Above, Peak map[string]float64
	// Used totals each container's p95 usage (busiest pod) × average replicas, priced like the requests.
	Used Totals
}

// Replay checks the recommended requests against the usage history. Skipped containers and containers
// without a recommended request are left out.
func Replay(in []ContainerInput, out Output, avgReplicas float64, rates Rates) Evidence {
	ev := Evidence{Above: map[string]float64{}, Peak: map[string]float64{}}
	for i, c := range in {
		o := out.Containers[i]
		for _, r := range Resources {
			usage := pickSeries(c, r)
			if len(usage) == 0 {
				continue
			}
			used := Quantile(usage, usedPercentile) * avgReplicas
			if r == CPU {
				ev.Used.CPU += used
			} else {
				ev.Used.Memory += used
			}
			req := o.Requests[r]
			if o.Skipped || req <= 0 {
				continue
			}
			above, peak := 0, 0.0
			for _, v := range usage {
				if v > req {
					above++
				}
				peak = math.Max(peak, v)
			}
			ev.Above[r] = math.Max(ev.Above[r], float64(above)/float64(len(usage)))
			ev.Peak[r] = math.Max(ev.Peak[r], peak/req)
		}
	}
	ev.Used.MonthlyCost = rates.monthly(ev.Used.CPU, ev.Used.Memory)
	return ev
}
