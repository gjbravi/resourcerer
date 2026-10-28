// Package engine computes resource recommendations from usage history.
// It is pure: no I/O, so every rule can be tested with synthetic series.
package engine

import (
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"time"

	"github.com/gjbravi/resourcerer/internal/settings"
)

const (
	CPU    = "cpu"
	Memory = "memory"
	Mi     = 1 << 20
	Gi     = 1 << 30
)

var Resources = []string{CPU, Memory}

// Goal maps an optimization goal to a usage percentile (1 = max) and a headroom margin.
type Goal struct {
	Percentile float64
	Margin     float64
}

var Goals = map[string]map[string]Goal{
	"Savings":     {CPU: {0.85, 0.05}, Memory: {0.99, 0.10}},
	"Balanced":    {CPU: {0.95, 0.10}, Memory: {1, 0.15}},
	"Reliability": {CPU: {0.99, 0.20}, Memory: {1, 0.30}},
}

func (g Goal) String() string {
	p := "max"
	if g.Percentile < 1 {
		p = fmt.Sprintf("p%d", int(math.Round(g.Percentile*100)))
	}
	return fmt.Sprintf("%s + %d%%", p, int(math.Round(g.Margin*100)))
}

// Values are cores for CPU and bytes for memory; a missing key means unset.
type Values map[string]float64

type ContainerInput struct {
	Name     string
	Requests Values
	Limits   Values
	// Usage of the busiest pod at each sample, oldest first.
	CPU    []float64
	Memory []float64
	// OOMKilled within the learning period: memory usage was capped by the limit.
	OOMKilled bool
	// CPUThrottled is the share of CFS periods throttled by the CPU limit over the learning period.
	CPUThrottled float64
}

// ThrottledShare is the share of throttled CFS periods from which the CPU limit counts as too tight.
const ThrottledShare = 0.25

type HPA struct {
	Name string
	// Utilization metrics the HPA scales on.
	CPU, Memory bool
	CPUTarget   int32
}

type Input struct {
	Settings   settings.Effective
	Containers []ContainerInput
	HPA        *HPA
	// History is how far back usage samples go.
	History  time.Duration
	Replicas int32
	// Expected is how many samples the learning period holds at the query step; 0 means unknown.
	Expected int
}

type ContainerOutput struct {
	Name     string
	Skipped  bool
	Requests Values
	Limits   Values
	Basis    map[string]string
}

type HPAOutput struct {
	Name        string
	Current     int32
	Recommended int32
	Clamped     bool
}

type Output struct {
	Phase      string
	Message    string
	Containers []ContainerOutput
	HPA        *HPAOutput
}

// MinSamples is the least history per series needed for any recommendation.
const MinSamples = 10

// minSamples relaxes MinSamples for learning periods so short they hold fewer samples.
func minSamples(expected int) int {
	if expected > 0 && expected/2 < MinSamples {
		return max(3, expected/2)
	}
	return MinSamples
}

func Recommend(in Input) Output {
	s := in.Settings
	learning := s.Duration(settings.LearningPeriod)
	out := Output{Phase: "Ready", Message: "Based on the full learning period."}

	if in.HPA != nil && in.HPA.CPU && in.HPA.Memory {
		out.Phase = "Unsupported"
		out.Message = fmt.Sprintf("HPA %s scales on CPU and memory utilization, so changing requests would change replica counts.", in.HPA.Name)
		for _, c := range in.Containers {
			out.Containers = append(out.Containers, keep(c, "HPA scales on CPU and memory"))
		}
		return out
	}

	need := minSamples(in.Expected)
	enough := false
	for _, c := range in.Containers {
		if len(c.CPU) >= need && len(c.Memory) >= need {
			enough = true
		}
	}
	switch {
	case !enough:
		out.Phase = "Pending"
		out.Message = "Collecting usage history."
	case in.Replicas == 0:
		out.Phase = "Warning"
		out.Message = "Scaled to zero: there are no pods to apply to. The recommendation uses earlier history."
	case in.History < learning*95/100:
		out.Phase = "Learning"
		out.Message = fmt.Sprintf("Preliminary: %s of the %s learning period collected.", short(in.History), short(learning))
	}

	skip := s.SkipList()
	for _, c := range in.Containers {
		if slices.Contains(skip, c.Name) {
			out.Containers = append(out.Containers, keep(c, "not optimized ("+settings.SkipContainers+")"))
			continue
		}
		o := ContainerOutput{Name: c.Name, Requests: Values{}, Limits: Values{}, Basis: map[string]string{}}
		for _, r := range Resources {
			policy := s.Get(pick(r, settings.CPULimits, settings.MemoryLimits))
			usage := pickSeries(c, r)
			keepCurrent := func(why string) {
				copyValue(o.Requests, c.Requests, r)
				copyValue(o.Limits, c.Limits, r)
				o.Basis[r] = why
			}
			switch {
			case len(usage) < need:
				keepCurrent("not enough data")
				continue
			case policy == "DoNotOptimize":
				keepCurrent("DoNotOptimize")
				continue
			case r == CPU && in.HPA != nil && in.HPA.CPU && s.Get(settings.HPAPreserve) != "true":
				keepCurrent("HPA scales on CPU")
				continue
			}

			goal := Goals[s.Get(pick(r, settings.CPUGoal, settings.MemoryGoal))][r]
			v := Quantile(usage, goal.Percentile) * (1 + goal.Margin)
			basis := goal.String() + " over " + short(min(in.History, learning))
			if r == Memory && c.OOMKilled {
				// Usage is capped at the limit when the kernel kills the container, so the history underestimates.
				if lim, ok := c.Limits[Memory]; ok && v < lim*1.2 {
					v = lim * 1.2
				}
				basis += " · raised after OOM kill"
			}
			minQ := s.Quantity(pick(r, settings.CPUMin, settings.MemoryMin))
			floor := minQ.AsApproximateFloat64()
			if v < floor {
				v = floor
			}
			v = round(r, v)
			o.Requests[r] = v
			o.Basis[r] = basis

			ratio := s.Float(pick(r, settings.CPURatio, settings.MemoryRatio))
			cur, hasLim := c.Limits[r]
			switch policy {
			case "RequestsOnly":
				if hasLim {
					o.Limits[r] = cur
				}
			case "RequestsAndLimits":
				o.Limits[r] = round(r, v*ratio)
			default:
				if hasLim {
					o.Limits[r] = cur
					if v > cur {
						o.Limits[r] = round(r, v*ratio)
					}
				}
			}
			if hasLim && policy != "RequestsOnly" {
				raiseLimit(&o, c, r, cur)
			}
		}
		out.Containers = append(out.Containers, o)
	}

	if in.HPA != nil && in.HPA.CPU && s.Get(settings.HPAPreserve) == "true" && in.HPA.CPUTarget > 0 {
		cur, rec := 0.0, 0.0
		for i, c := range in.Containers {
			cur += c.Requests[CPU]
			rec += out.Containers[i].Requests[CPU]
		}
		if cur > 0 && rec > 0 {
			raw := int32(math.Round(float64(in.HPA.CPUTarget) * cur / rec))
			h := &HPAOutput{Name: in.HPA.Name, Current: in.HPA.CPUTarget, Recommended: min(max(raw, 50), 90)}
			h.Clamped = h.Recommended != raw
			out.HPA = h
		}
	}
	return out
}

// raiseLimit lifts a CPU limit that throttles the container. Throttling happens in 100ms periods, so it shows
// up even when sampled usage stays below the limit and the request rule would leave the limit alone.
func raiseLimit(o *ContainerOutput, c ContainerInput, r string, cur float64) {
	if r != CPU || c.CPUThrottled < ThrottledShare {
		return
	}
	o.Limits[CPU] = max(o.Limits[CPU], round(CPU, cur*1.5))
	o.Basis[CPU] += fmt.Sprintf(" · limit raised: throttled %d%% of the time", int(math.Round(c.CPUThrottled*100)))
}

func keep(c ContainerInput, why string) ContainerOutput {
	o := ContainerOutput{Name: c.Name, Skipped: true, Requests: Values{}, Limits: Values{}, Basis: map[string]string{CPU: why, Memory: why}}
	for _, r := range Resources {
		copyValue(o.Requests, c.Requests, r)
		copyValue(o.Limits, c.Limits, r)
	}
	return o
}

func copyValue(dst, src Values, r string) {
	if v, ok := src[r]; ok {
		dst[r] = v
	}
}

func pick(r, cpu, mem string) string {
	if r == CPU {
		return cpu
	}
	return mem
}

func pickSeries(c ContainerInput, r string) []float64 {
	if r == CPU {
		return c.CPU
	}
	return c.Memory
}

// round keeps CPU at 1m precision below 100m and 5m above; memory at 1Mi.
func round(r string, v float64) float64 {
	if r == Memory {
		return math.Ceil(v/Mi) * Mi
	}
	if v < 0.1 {
		return math.Ceil(v*1000) / 1000
	}
	return math.Ceil(v*200) / 200
}

// Quantile uses linear interpolation; p >= 1 returns the maximum.
func Quantile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	s := slices.Clone(values)
	slices.Sort(s)
	if p >= 1 {
		return s[len(s)-1]
	}
	i := p * float64(len(s)-1)
	lo := int(math.Floor(i))
	hi := min(lo+1, len(s)-1)
	return s[lo] + (s[hi]-s[lo])*(i-float64(lo))
}

func short(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dm", max(1, int(d.Minutes())))
	}
}

// Hash identifies a set of recommended values, so a reviewed recommendation can be told apart from a newer one.
func Hash(out Output) string {
	h := fnv.New64a()
	for _, c := range out.Containers {
		_, _ = fmt.Fprintf(h, "%s|", c.Name)
		for _, r := range Resources {
			_, _ = fmt.Fprintf(h, "%v/%v|", c.Requests[r], c.Limits[r])
		}
	}
	if out.HPA != nil {
		_, _ = fmt.Fprintf(h, "hpa=%d", out.HPA.Recommended)
	}
	return fmt.Sprintf("%012x", h.Sum64())[:12]
}

// Pending reports whether any request changes by at least threshold and by a meaningful absolute amount.
func Pending(in []ContainerInput, out Output, threshold float64) bool {
	for i, c := range in {
		o := out.Containers[i]
		if o.Skipped {
			continue
		}
		for _, r := range Resources {
			to, ok := o.Requests[r]
			if !ok {
				continue
			}
			from := c.Requests[r]
			minAbs := 0.01
			if r == Memory {
				minAbs = 16 * Mi
			}
			if math.Abs(to-from) >= minAbs && (from == 0 || math.Abs(to-from)/from >= threshold) {
				return true
			}
		}
	}
	return false
}

type Rates struct {
	CPUCoreHour   float64
	MemoryGiBHour float64
}

const hoursPerMonth = 730

type Totals struct {
	CPU, Memory, MonthlyCost float64
}

type Summary struct {
	Current, Recommended Totals
	// Scores are 0-100: how closely current requests match the recommendation.
	Score, CPUScore, MemoryScore int32
}

func (r Rates) monthly(cpu, mem float64) float64 {
	return (cpu*r.CPUCoreHour + mem/Gi*r.MemoryGiBHour) * hoursPerMonth
}

// Summarize totals requests × average replicas and scores alignment, weighting CPU and memory by cost.
func Summarize(in []ContainerInput, out Output, avgReplicas float64, rates Rates) Summary {
	var s Summary
	num := map[string]float64{}
	den := map[string]float64{}
	for i, c := range in {
		o := out.Containers[i]
		for _, r := range Resources {
			a := c.Requests[r]
			b, ok := o.Requests[r]
			if !ok {
				b = a
			}
			if r == CPU {
				s.Current.CPU += a * avgReplicas
				s.Recommended.CPU += b * avgReplicas
			} else {
				s.Current.Memory += a * avgReplicas
				s.Recommended.Memory += b * avgReplicas
			}
			if !o.Skipped {
				num[r] += math.Min(a, b)
				den[r] += math.Max(a, b)
			}
		}
	}
	s.Current.MonthlyCost = rates.monthly(s.Current.CPU, s.Current.Memory)
	s.Recommended.MonthlyCost = rates.monthly(s.Recommended.CPU, s.Recommended.Memory)
	score := func(r string) float64 {
		if den[r] == 0 {
			return 100
		}
		return 100 * num[r] / den[r]
	}
	wCPU := s.Recommended.CPU * rates.CPUCoreHour
	wMem := s.Recommended.Memory / Gi * rates.MemoryGiBHour
	overall := 100.0
	if wCPU+wMem > 0 {
		overall = (score(CPU)*wCPU + score(Memory)*wMem) / (wCPU + wMem)
	}
	s.Score = int32(math.Round(overall))
	s.CPUScore = int32(math.Round(score(CPU)))
	s.MemoryScore = int32(math.Round(score(Memory)))
	return s
}
