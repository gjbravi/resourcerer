package engine

import (
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/settings"
)

func series(n int, f func(i int) float64) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = f(i)
	}
	return out
}

func flat(v float64) []float64 { return series(100, func(int) float64 { return v }) }

func input(ann map[string]string, c ContainerInput) Input {
	return Input{
		Settings:   settings.Resolve(nil, nil, ann),
		Containers: []ContainerInput{c},
		History:    8 * 24 * time.Hour,
		Replicas:   2,
	}
}

func TestQuantile(t *testing.T) {
	v := series(101, func(i int) float64 { return float64(i) })
	if got := Quantile(v, 0.95); got != 95 {
		t.Fatalf("p95 = %v, want 95", got)
	}
	if got := Quantile(v, 1); got != 100 {
		t.Fatalf("max = %v, want 100", got)
	}
}

func TestBalancedGoal(t *testing.T) {
	out := Recommend(input(nil, ContainerInput{
		Name:     "api",
		Requests: Values{CPU: 1, Memory: 1 * Gi},
		Limits:   Values{CPU: 2, Memory: 2 * Gi},
		CPU:      series(101, func(i int) float64 { return float64(i) / 1000 }),
		Memory:   flat(100 * Mi),
	}))
	c := out.Containers[0]
	if out.Phase != "Ready" {
		t.Fatalf("phase = %s", out.Phase)
	}
	// p95 of 0..100m is 95m, +10% = 104.5m, rounded up to 105m.
	if got := c.Requests[CPU]; got != 0.105 {
		t.Errorf("cpu request = %v, want 0.105", got)
	}
	if got := c.Requests[Memory]; got != 115*Mi {
		t.Errorf("memory request = %v MiB, want 115", got/Mi)
	}
	if c.Limits[CPU] != 2 || c.Limits[Memory] != 2*Gi {
		t.Errorf("limits changed: %v", c.Limits)
	}
	if c.Basis[CPU] != "p95 + 10% over 7d" {
		t.Errorf("basis = %q", c.Basis[CPU])
	}
}

func TestRaiseLimitsIfNeeded(t *testing.T) {
	out := Recommend(input(map[string]string{settings.Domain + "/" + settings.CPUGoal: "Reliability"}, ContainerInput{
		Name:     "worker",
		Requests: Values{CPU: 0.05, Memory: 64 * Mi},
		Limits:   Values{CPU: 0.5, Memory: 512 * Mi},
		CPU:      flat(0.5),
		Memory:   flat(100 * Mi),
	}))
	c := out.Containers[0]
	if c.Requests[CPU] != 0.6 {
		t.Fatalf("cpu request = %v, want 0.6 (max + 20%%)", c.Requests[CPU])
	}
	if c.Limits[CPU] != 1.2 {
		t.Errorf("cpu limit = %v, want request × 2.0 because the request exceeds the old limit", c.Limits[CPU])
	}
	if c.Limits[Memory] != 512*Mi {
		t.Errorf("memory limit = %v, want unchanged", c.Limits[Memory])
	}
}

func TestOOMRaisesMemoryAboveLimit(t *testing.T) {
	out := Recommend(input(nil, ContainerInput{
		Name:      "ledger",
		Requests:  Values{CPU: 0.25, Memory: 256 * Mi},
		Limits:    Values{CPU: 0.5, Memory: 320 * Mi},
		CPU:       flat(0.2),
		Memory:    flat(320 * Mi),
		OOMKilled: true,
	}))
	c := out.Containers[0]
	if c.Requests[Memory] != 384*Mi {
		t.Fatalf("memory request = %v MiB, want 384 (limit × 1.2)", c.Requests[Memory]/Mi)
	}
	if c.Limits[Memory] != 576*Mi {
		t.Errorf("memory limit = %v MiB, want 576 (request × 1.5)", c.Limits[Memory]/Mi)
	}
}

func TestSkipAndMinimums(t *testing.T) {
	in := input(nil, ContainerInput{Name: "istio-proxy", Requests: Values{CPU: 0.1}, CPU: flat(0.01), Memory: flat(Mi)})
	in.Containers = append(in.Containers, ContainerInput{Name: "tiny", CPU: flat(0.0001), Memory: flat(Mi)})
	out := Recommend(in)
	if !out.Containers[0].Skipped || out.Containers[0].Requests[CPU] != 0.1 {
		t.Errorf("istio-proxy should be skipped and keep 100m: %+v", out.Containers[0])
	}
	if got := out.Containers[1].Requests; got[CPU] != 0.01 || got[Memory] != 16*Mi {
		t.Errorf("tiny should be clamped to 10m/16Mi: %v", got)
	}
}

func TestPhases(t *testing.T) {
	c := ContainerInput{Name: "a", Requests: Values{CPU: 1}, CPU: flat(0.1), Memory: flat(Mi)}
	in := input(nil, c)
	in.History = 2 * 24 * time.Hour
	if p := Recommend(in).Phase; p != "Learning" {
		t.Errorf("2d of 7d: phase = %s, want Learning", p)
	}
	in.Replicas = 0
	if p := Recommend(in).Phase; p != "Warning" {
		t.Errorf("scaled to zero: phase = %s, want Warning", p)
	}
	in.Containers[0].CPU = flat(0.1)[:3]
	in.Containers[0].Memory = flat(Mi)[:3]
	if p := Recommend(in).Phase; p != "Pending" {
		t.Errorf("3 samples: phase = %s, want Pending", p)
	}
	in = input(nil, c)
	in.HPA = &HPA{Name: "h", CPU: true, Memory: true}
	if p := Recommend(in).Phase; p != "Unsupported" {
		t.Errorf("HPA on cpu+memory: phase = %s, want Unsupported", p)
	}
}

func TestHPAPreserveScalePoint(t *testing.T) {
	in := input(nil, ContainerInput{Name: "web", Requests: Values{CPU: 0.3, Memory: 512 * Mi}, CPU: flat(0.2), Memory: flat(50 * Mi)})
	in.HPA = &HPA{Name: "web", CPU: true, CPUTarget: 70}
	out := Recommend(in)
	if out.HPA == nil || out.HPA.Recommended != 90 || !out.HPA.Clamped {
		t.Fatalf("hpa = %+v, want 70%% × 300m / 220m = 95%%, clamped to 90%%", out.HPA)
	}

	in.Settings = settings.Resolve(map[string]string{settings.HPAPreserve: "false"}, nil, nil)
	out = Recommend(in)
	if out.HPA != nil || out.Containers[0].Requests[CPU] != 0.3 {
		t.Errorf("without preserve, CPU stays and no HPA change: %+v %+v", out.HPA, out.Containers[0].Requests)
	}
}

func TestSummarizeAndPending(t *testing.T) {
	in := []ContainerInput{{Name: "a", Requests: Values{CPU: 1, Memory: 1 * Gi}}}
	out := Output{Containers: []ContainerOutput{{Name: "a", Requests: Values{CPU: 0.5, Memory: 1 * Gi}}}}
	s := Summarize(in, out, 2, Rates{CPUCoreHour: 0.03, MemoryGiBHour: 0.004})
	if s.Current.CPU != 2 || s.Recommended.CPU != 1 {
		t.Errorf("totals = %+v", s)
	}
	if s.CPUScore != 50 || s.MemoryScore != 100 {
		t.Errorf("scores = %d/%d, want 50/100", s.CPUScore, s.MemoryScore)
	}
	if !Pending(in, out, 0.1) {
		t.Error("50% change should be pending")
	}
	out.Containers[0].Requests[CPU] = 0.95
	if Pending(in, out, 0.1) {
		t.Error("5% change should not be pending at a 10% threshold")
	}
}

func TestShortLearningPeriodNeedsFewerSamples(t *testing.T) {
	in := input(nil, ContainerInput{Name: "a", Requests: Values{CPU: 1}, CPU: flat(0.1)[:6], Memory: flat(Mi)[:6]})
	if p := Recommend(in).Phase; p != "Pending" {
		t.Errorf("6 samples with unknown expectation: phase = %s, want Pending", p)
	}
	in.Expected = 8
	if p := Recommend(in).Phase; p == "Pending" {
		t.Errorf("6 of 8 expected samples should be enough, got Pending")
	}
}
