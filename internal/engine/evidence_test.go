package engine

import (
	"math"
	"testing"
)

func TestReplay(t *testing.T) {
	in := []ContainerInput{
		{Name: "app", Requests: Values{CPU: 1, Memory: 4 * Gi},
			CPU: []float64{0.2, 0.4, 0.5, 0.9}, Memory: []float64{1 * Gi, 1 * Gi, 2 * Gi, 2 * Gi}},
		{Name: "sidecar", Requests: Values{CPU: 0.1}, CPU: []float64{5, 5, 5, 5}},
	}
	out := Output{Containers: []ContainerOutput{
		{Requests: Values{CPU: 0.5, Memory: 2 * Gi}},
		{Skipped: true, Requests: Values{CPU: 0.1}},
	}}
	ev := Replay(in, out, 2, Rates{CPUCoreHour: 0.01, MemoryGiBHour: 0.001})
	if ev.Above[CPU] != 0.25 || ev.Above[Memory] != 0 {
		t.Errorf("above = %v, want cpu 0.25 and memory 0 (a skipped container does not count)", ev.Above)
	}
	if math.Abs(ev.Peak[CPU]-1.8) > 1e-9 || ev.Peak[Memory] != 1 {
		t.Errorf("peak = %v, want cpu 1.8 and memory 1", ev.Peak)
	}
	// p95 of the busiest pod × 2 replicas, the skipped sidecar included: what the pods use.
	if math.Abs(ev.Used.CPU-2*(Quantile(in[0].CPU, 0.95)+5)) > 1e-9 || ev.Used.MonthlyCost <= 0 {
		t.Errorf("used = %+v", ev.Used)
	}
}
