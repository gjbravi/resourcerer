package decision

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// Usage is a container's busiest-pod usage over the learning period, oldest first.
type Usage struct {
	CPU, Memory []prom.Point
}

// Input is what the state is built from: the Recommendation's status as the controller wrote it, and the usage
// series it was computed from. Names of the workload, its namespace and the people who applied it are left out:
// they do not help the decision and should not leave the cluster.
type Input struct {
	Kind     string
	Replicas int32
	Status   v1alpha1.RecommendationStatus
	Usage    map[string]Usage
	Now      time.Time
}

// State is the document the model reads. Every field is a short phrase with units.
type State struct {
	Workload       string            `json:"workload"`
	Window         string            `json:"window"`
	Intent         string            `json:"intent"`
	Usage          map[string]string `json:"usage"`
	Recommendation string            `json:"recommendation"`
	Replay         string            `json:"replay"`
	Fixes          string            `json:"fixes"`
	Reliability    string            `json:"reliability"`
	History        string            `json:"history"`
}

// Question is a choice question in the /v1/systemone format; Criteria keeps the options in order.
type Question struct {
	Type         string   `json:"type"`
	Instructions string   `json:"instructions"`
	Criteria     criteria `json:"criteria"`
}

type option struct{ Name, Description string }

type criteria []option

func (c criteria) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, o := range c {
		if i > 0 {
			b.WriteByte(',')
		}
		k, _ := json.Marshal(o.Name)
		v, _ := json.Marshal(o.Description)
		b.Write(k)
		b.WriteByte(':')
		b.Write(v)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// Request is the body of POST /v1/systemone.
type Request struct {
	State     State               `json:"state"`
	Questions map[string]Question `json:"questions"`
}

// QuestionVerdict asks whether the recommendation is safe to apply now; its options are Apply, Review and Wait.
const (
	QuestionVerdict = "verdict"
	Apply           = "apply"
	Review          = "review"
	Wait            = "wait"
)

// DailyPattern is the share of the variance the hour of day must explain for usage to follow the time of day.
const DailyPattern = 0.5

// Build returns the request for one workload and a hash of it.
func Build(in Input) (Request, string) {
	cs := containers(in.Status)
	req := Request{
		State: State{
			Workload:       describeWorkload(in),
			Window:         window(in.Status.Evidence),
			Intent:         intent(in.Status.Settings),
			Usage:          map[string]string{},
			Recommendation: recommendation(cs),
			Replay:         replay(in.Status),
			Fixes:          fixes(in, cs),
			Reliability:    reliability(in, cs),
			History:        history(in),
		},
		Questions: map[string]Question{
			QuestionVerdict: {
				Type: "choice",
				Instructions: "Is it safe to apply the recommendation to this workload now? Weigh how the recommended " +
					"requests held up against the usage history, how much of the learning period that history covers, " +
					"the usage shape, what applying fixes and the reliability signals it does not fix.",
				Criteria: criteria{
					{Apply, "safe to apply now"},
					{Review, "a person should review it before applying"},
					{Wait, "wait for more history or for the reliability signals to settle"},
				},
			},
		},
	}
	for _, c := range cs {
		u := in.Usage[c.Name]
		req.State.Usage[c.Name+" cpu"] = describe(ShapeOf(u.CPU), cores)
		req.State.Usage[c.Name+" memory"] = describe(ShapeOf(u.Memory), mebibytes)
	}
	b, _ := json.Marshal(req)
	h := fnv.New64a()
	_, _ = h.Write(b)
	return req, fmt.Sprintf("%012x", h.Sum64())[:12]
}

// container is one container's values before and after applying, in cores and bytes.
type container struct {
	Name                   string
	Skipped                bool
	CurRequests, CurLimits engine.Values
	RecRequests, RecLimits engine.Values
}

func containers(st v1alpha1.RecommendationStatus) []container {
	out := make([]container, 0, len(st.Containers))
	for _, c := range st.Containers {
		out = append(out, container{
			Name: c.Name, Skipped: c.Skipped,
			CurRequests: values(c.Current.Requests), CurLimits: values(c.Current.Limits),
			RecRequests: values(c.Recommended.Requests), RecLimits: values(c.Recommended.Limits),
		})
	}
	return out
}

func values(l corev1.ResourceList) engine.Values { return workload.ToValues(l) }

func describeWorkload(in Input) string {
	st := in.Status
	var b strings.Builder
	fmt.Fprintf(&b, "%s, %d replicas now", in.Kind, in.Replicas)
	if st.Impact != nil {
		if avg, err := strconv.ParseFloat(st.Impact.AvgReplicas, 64); err == nil {
			fmt.Fprintf(&b, ", %s on average", num(avg))
		}
	}
	if h := st.HPA; h != nil {
		fmt.Fprintf(&b, "; HPA on CPU at %d%%", h.Current)
		if h.Recommended != h.Current {
			fmt.Fprintf(&b, ", target to become %d%% to keep scaling at the same CPU per pod", h.Recommended)
		}
	}
	if st.Pricing != nil && len(st.Pricing.Capacity) > 0 {
		parts := []string{}
		for _, t := range []string{"spot", "on-demand", "reserved"} {
			if v, ok := st.Pricing.Capacity[t]; ok {
				parts = append(parts, fmt.Sprintf("%d%% %s", v, t))
			}
		}
		if len(parts) > 0 {
			fmt.Fprintf(&b, "; runs on %s nodes", strings.Join(parts, ", "))
		}
	}
	return b.String()
}

func window(ev *v1alpha1.Evidence) string {
	if ev == nil || ev.LearningSeconds == 0 {
		return "unknown history"
	}
	h, l := time.Duration(ev.HistorySeconds)*time.Second, time.Duration(ev.LearningSeconds)*time.Second
	if h >= l*95/100 {
		return fmt.Sprintf("%s learning period, fully covered", span(l))
	}
	return fmt.Sprintf("only %s of the %s learning period collected: preliminary", span(h), span(l))
}

func intent(s map[string]v1alpha1.Setting) string {
	cpu, mem := s[settings.CPUGoal], s[settings.MemoryGoal]
	src := cpu.Source
	if mem.Source != src {
		src = cpu.Source + " (CPU) and " + mem.Source + " (memory)"
	}
	return fmt.Sprintf("CPU goal %s, memory goal %s, set at %s level", cpu.Value, mem.Value, src)
}

func describe(s Shape, f func(float64) string) string {
	if s.Samples == 0 {
		return "no data"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "busiest pod p50 %s, p95 %s, p99 %s, max %s; max/p95 %sx; variation %d%%",
		f(s.P50), f(s.P95), f(s.P99), f(s.Max), num(s.Spike), pct(s.CV))
	switch {
	case s.Daily >= DailyPattern:
		fmt.Fprintf(&b, "; daily pattern (the hour of day explains %d%% of the variation), busiest %02d:00-%02d:00 UTC",
			pct(s.Daily), s.PeakFrom, s.PeakTo)
	case s.Daily > 0:
		b.WriteString("; no daily pattern")
	}
	if math.Abs(s.Trend) >= 0.01 {
		dir := "growing"
		if s.Trend < 0 {
			dir = "shrinking"
		}
		fmt.Fprintf(&b, "; %s %d%% a day", dir, pct(math.Abs(s.Trend)))
	} else {
		b.WriteString("; no trend")
	}
	return b.String()
}

// recommendation lists every value the apply changes, from and to.
func recommendation(cs []container) string {
	parts := []string{}
	for _, c := range cs {
		if c.Skipped {
			parts = append(parts, c.Name+" not changed")
			continue
		}
		ch := []string{}
		for _, r := range engine.Resources {
			f := cores
			if r == engine.Memory {
				f = mebibytes
			}
			ch = append(ch, change(r+" request", c.CurRequests, c.RecRequests, r, f)...)
			ch = append(ch, change(r+" limit", c.CurLimits, c.RecLimits, r, f)...)
		}
		if len(ch) == 0 {
			ch = append(ch, "no change")
		}
		parts = append(parts, c.Name+": "+strings.Join(ch, ", "))
	}
	return strings.Join(parts, "; ")
}

func change(what string, from, to engine.Values, r string, f func(float64) string) []string {
	a, okA := from[r]
	b, okB := to[r]
	switch {
	case !okA && !okB, okA && okB && a == b:
		return nil
	case !okA:
		return []string{fmt.Sprintf("%s none -> %s", what, f(b))}
	case !okB:
		return []string{fmt.Sprintf("%s %s -> none", what, f(a))}
	}
	return []string{fmt.Sprintf("%s %s -> %s (%+d%%)", what, f(a), f(b), pct(b/a-1))}
}

func replay(st v1alpha1.RecommendationStatus) string {
	parts := []string{}
	if ev := st.Evidence; ev != nil {
		parts = append(parts, fmt.Sprintf("over the history, the busiest pod used more than the recommended requests %d%% "+
			"of the time for CPU and %d%% for memory, and peaked at %d%% of the CPU and %d%% of the memory request",
			ev.CPUAbovePercent, ev.MemoryAbovePercent, ev.CPUPeakPercent, ev.MemoryPeakPercent))
	}
	if im := st.Impact; im != nil {
		cur, _ := strconv.ParseFloat(im.Current.MonthlyCost, 64)
		rec, _ := strconv.ParseFloat(im.Recommended.MonthlyCost, 64)
		effect := fmt.Sprintf("saves %s/month", money(cur-rec))
		if rec > cur {
			effect = fmt.Sprintf("costs %s/month more", money(rec-cur))
		}
		parts = append(parts, fmt.Sprintf("%s -> %s a month, %s", money(cur), money(rec), effect))
	}
	if len(parts) == 0 {
		return "no replay"
	}
	return strings.Join(parts, "; ")
}

// fixed tells which reliability problems the recommendation itself addresses.
type fixed struct {
	oom              bool
	throttle, memLim map[string]bool
	underRequested   []string
}

func whatItFixes(in Input, cs []container) fixed {
	f := fixed{throttle: map[string]bool{}, memLim: map[string]bool{}}
	throttled := map[string]int32{}
	if r := in.Status.Reliability; r != nil {
		for _, c := range r.Containers {
			throttled[c.Name] = c.CPUThrottledPercent
		}
	}
	for _, c := range cs {
		if c.Skipped {
			continue
		}
		raisedMem := c.RecRequests[engine.Memory] > c.CurRequests[engine.Memory] || c.RecLimits[engine.Memory] > c.CurLimits[engine.Memory]
		if in.Status.OOMKills > 0 && raisedMem {
			f.oom = true
		}
		if c.RecLimits[engine.CPU] > c.CurLimits[engine.CPU] {
			f.throttle[c.Name] = true
		}
		if c.RecLimits[engine.Memory] > c.CurLimits[engine.Memory] {
			f.memLim[c.Name] = true
		}
		u := in.Usage[c.Name]
		for _, r := range engine.Resources {
			pts := u.CPU
			if r == engine.Memory {
				pts = u.Memory
			}
			req := c.CurRequests[r]
			if req > 0 && len(pts) > 0 && engine.Quantile(prom.Values(pts), 0.95) > req && c.RecRequests[r] > req {
				f.underRequested = append(f.underRequested, fmt.Sprintf("%s: raises the %s request, which p95 usage was above", c.Name, r))
			}
		}
		if t := throttled[c.Name]; float64(t)/100 >= engine.ThrottledShare && f.throttle[c.Name] {
			f.underRequested = append(f.underRequested, fmt.Sprintf("%s: raises the CPU limit it was throttled by %d%% of the time", c.Name, t))
		}
	}
	return f
}

// fixes names the reliability problems the recommendation itself addresses.
func fixes(in Input, cs []container) string {
	f := whatItFixes(in, cs)
	parts := []string{}
	if f.oom {
		parts = append(parts, "raises memory after OOM kills")
	}
	parts = append(parts, f.underRequested...)
	if len(parts) == 0 {
		return "nothing: no reliability problem the recommendation addresses"
	}
	return strings.Join(parts, "; ")
}

// reliability lists the signals, tagging the ones the recommendation fixes: listed apart from the signals, the model
// counted them against applying.
func reliability(in Input, cs []container) string {
	f := whatItFixes(in, cs)
	tag := func(ok bool) string {
		if ok {
			return " (fixed by applying)"
		}
		return ""
	}
	parts := []string{}
	if n := in.Status.OOMKills; n > 0 {
		parts = append(parts, fmt.Sprintf("%d OOM kills%s", n, tag(f.oom)))
	} else {
		parts = append(parts, "no OOM kills")
	}
	if r := in.Status.Reliability; r != nil {
		for _, c := range r.Containers {
			if c.CPUThrottledPercent > 0 {
				parts = append(parts, fmt.Sprintf("%s throttled by its CPU limit %d%% of the time%s", c.Name, c.CPUThrottledPercent, tag(f.throttle[c.Name])))
			}
			if c.MemoryPeakPercent > 0 {
				parts = append(parts, fmt.Sprintf("%s memory peaked at %d%% of its limit%s", c.Name, c.MemoryPeakPercent, tag(f.memLim[c.Name])))
			}
		}
		if r.Restarts > 0 {
			parts = append(parts, fmt.Sprintf("%d other restarts", r.Restarts))
		}
		if len(r.CrashLooping) > 0 {
			parts = append(parts, "crash looping now: "+strings.Join(r.CrashLooping, ", "))
		}
		if r.BestEffortPods > 0 {
			parts = append(parts, fmt.Sprintf("%d pods without requests", r.BestEffortPods))
		}
		if r.HPAAtMax {
			parts = append(parts, "HPA at its maximum replicas")
		}
		for _, s := range r.Signals {
			parts = append(parts, fmt.Sprintf("%s %s (%s)", s.Name, s.Display, s.Status))
		}
	}
	return strings.Join(parts, "; ")
}

func history(in Input) string {
	parts := []string{}
	for _, h := range in.Status.History {
		by := "by a person"
		if strings.HasPrefix(h.By, "auto-deploy") {
			by = "by auto-deploy"
		}
		did := "applied"
		if h.Action == v1alpha1.ActionRevert {
			did = "reverted"
		}
		parts = append(parts, fmt.Sprintf("%s %s ago %s", did, span(in.Now.Sub(h.Time.Time)), by))
		if len(parts) == 3 {
			break
		}
	}
	if a := in.Status.LastApplyResult; a != nil && a.Result != "Succeeded" {
		parts = append(parts, fmt.Sprintf("last apply %s %s ago", strings.ToLower(a.Result), span(in.Now.Sub(a.Time.Time))))
	}
	if len(parts) == 0 {
		return "never applied"
	}
	return strings.Join(parts, "; ")
}

func cores(v float64) string { return num(v) }

func mebibytes(v float64) string {
	mi := v / engine.Mi
	if mi >= 1024 {
		return num(mi/1024) + "Gi"
	}
	return fmt.Sprintf("%dMi", int(math.Round(mi)))
}

func num(v float64) string {
	switch {
	case v >= 10:
		return fmt.Sprintf("%.0f", v)
	case v >= 1:
		return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.1f", v), "0"), ".")
	}
	return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.3f", v), "0"), ".")
}

func pct(f float64) int { return int(math.Round(f * 100)) }

func money(v float64) string { return fmt.Sprintf("$%.2f", v) }

func span(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dm", max(1, int(d.Minutes())))
}
