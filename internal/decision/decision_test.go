package decision

import (
	"encoding/json"
	"math"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
)

var t0 = time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)

// series samples f every 5 minutes for the given number of days.
func series(days int, f func(t time.Time) float64) []prom.Point {
	var out []prom.Point
	for t := t0; t.Before(t0.Add(time.Duration(days) * 24 * time.Hour)); t = t.Add(5 * time.Minute) {
		out = append(out, prom.Point{T: float64(t.Unix()), V: f(t)})
	}
	return out
}

func TestShapeOf(t *testing.T) {
	evening := func(t time.Time) float64 {
		if h := t.Hour(); h >= 18 && h < 21 {
			return 1
		}
		return 0.2
	}
	cases := []struct {
		name               string
		pts                []prom.Point
		daily, trend       bool
		peakFrom, peakTo   int
		spikeAbove, cvOver float64
	}{
		{name: "flat", pts: series(7, func(time.Time) float64 { return 0.5 })},
		{name: "evening peak", pts: series(7, evening), daily: true, peakFrom: 18, peakTo: 21, cvOver: 0.5},
		{name: "growing", pts: series(7, func(t time.Time) float64 { return 100 + 10*t.Sub(t0).Hours()/24 }), trend: true},
		{name: "one spike", pts: series(3, func(t time.Time) float64 {
			if t.Equal(t0.Add(30 * time.Hour)) {
				return 5
			}
			return 1
		}), spikeAbove: 4},
		{name: "too short for a daily pattern", pts: series(1, evening), cvOver: 0.5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := ShapeOf(c.pts)
			if got := s.Daily >= DailyPattern; got != c.daily {
				t.Fatalf("daily = %.2f, want pattern %v", s.Daily, c.daily)
			}
			if c.daily && (s.PeakFrom != c.peakFrom || s.PeakTo != c.peakTo) {
				t.Errorf("peak %d-%d, want %d-%d", s.PeakFrom, s.PeakTo, c.peakFrom, c.peakTo)
			}
			if got := math.Abs(s.Trend) >= 0.01; got != c.trend {
				t.Errorf("trend = %.3f, want trend %v", s.Trend, c.trend)
			}
			if s.Spike < c.spikeAbove {
				t.Errorf("spike = %.2f, want at least %.2f", s.Spike, c.spikeAbove)
			}
			if s.CV < c.cvOver {
				t.Errorf("cv = %.2f, want at least %.2f", s.CV, c.cvOver)
			}
		})
	}
}

func qty(s string) resource.Quantity { return resource.MustParse(s) }

// status is a Ready recommendation for one container, app, cutting CPU 1 -> 0.55 and memory 1Gi -> 460Mi.
func status() v1alpha1.RecommendationStatus {
	return v1alpha1.RecommendationStatus{
		Phase: v1alpha1.PhaseReady, Pending: true,
		Settings: map[string]v1alpha1.Setting{
			settings.CPUGoal: {Value: "Balanced", Source: "cluster"}, settings.MemoryGoal: {Value: "Balanced", Source: "cluster"},
		},
		Containers: []v1alpha1.ContainerRecommendation{{
			Name: "app",
			Current: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: qty("1"), corev1.ResourceMemory: qty("1Gi")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: qty("1Gi")},
			},
			Recommended: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: qty("550m"), corev1.ResourceMemory: qty("460Mi")},
				Limits:   corev1.ResourceList{corev1.ResourceMemory: qty("1Gi")},
			},
		}},
		Impact:   &v1alpha1.Impact{AvgReplicas: "3.00", Current: v1alpha1.Totals{MonthlyCost: "74.46"}, Recommended: v1alpha1.Totals{MonthlyCost: "40.07"}},
		Evidence: &v1alpha1.Evidence{HistorySeconds: 7 * 86400, LearningSeconds: 7 * 86400, CPUAbovePercent: 2, CPUPeakPercent: 91, MemoryPeakPercent: 87},
	}
}

func usage() map[string]Usage {
	return map[string]Usage{"app": {
		CPU:    series(7, func(t time.Time) float64 { return 0.2 + 0.3*math.Sin(float64(t.Unix())/3600) }),
		Memory: series(7, func(time.Time) float64 { return 400 * engine.Mi }),
	}}
}

func TestBuild(t *testing.T) {
	st := status()
	st.OOMKills = 2
	st.Reliability = &v1alpha1.Reliability{Signals: []v1alpha1.SignalStatus{{Name: "HTTP 5xx", Display: "0.2%", Status: "ok"}}}
	st.History = []v1alpha1.HistoryEntry{{Time: metav1Time(t0.Add(-48 * time.Hour)), Action: v1alpha1.ActionApply, By: "jane@example.com"}}
	in := Input{Kind: "Deployment", Replicas: 3, Status: st, Usage: usage(), Now: t0}
	req, h := Build(in)
	b, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	doc := string(b)
	for _, want := range []string{
		`"workload":"Deployment, 3 replicas now, 3 on average"`,
		`"window":"7d learning period, fully covered"`,
		`"intent":"CPU goal Balanced, memory goal Balanced, set at cluster level"`,
		`"app memory":"busiest pod p50 400Mi`,
		`no daily pattern`,
		`"recommendation":"app: cpu request 1 -\u003e 0.55 (-45%), memory request 1Gi -\u003e 460Mi (-55%)"`,
		`used more than the recommended requests 2% of the time for CPU`,
		`$74.46 -\u003e $40.07 a month, saves $34.39/month`,
		`"fixes":"nothing`,
		`"reliability":"2 OOM kills; HTTP 5xx 0.2% (ok)"`,
		`applied 2d ago by a person`,
		`"criteria":{"apply":"safe to apply now","review":`,
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("state lacks %s\n%s", want, doc)
		}
	}
	if len(req.Questions) != 1 {
		t.Errorf("questions %v, want the verdict only", req.Questions)
	}
	if strings.Contains(doc, "jane") {
		t.Errorf("state names a person: %s", doc)
	}
	if _, h2 := Build(in); h2 != h || len(h) != 12 {
		t.Errorf("hash %q then %q", h, h2)
	}
}

func TestBuildLeavesOutSkippedContainersUsage(t *testing.T) {
	st := status()
	st.Containers = append(st.Containers, v1alpha1.ContainerRecommendation{Name: "istio-proxy", Skipped: true})
	u := usage()
	u["istio-proxy"] = u["app"]
	req, _ := Build(Input{Kind: "Deployment", Status: st, Usage: u, Now: t0})
	if _, ok := req.State.Usage["istio-proxy cpu"]; ok || len(req.State.Usage) != 2 {
		t.Errorf("usage %v, want the app container only", req.State.Usage)
	}
	if !strings.Contains(req.State.Recommendation, "istio-proxy not changed") {
		t.Errorf("recommendation %q, want the sidecar listed as unchanged", req.State.Recommendation)
	}
}

func TestWindow(t *testing.T) {
	ev := &v1alpha1.Evidence{HistorySeconds: 1200, LearningSeconds: 7 * 86400}
	if got := window(ev); got != "only 20m of the 7d learning period collected: preliminary" {
		t.Errorf("window %q", got)
	}
}

func TestFixes(t *testing.T) {
	raise := func(st *v1alpha1.RecommendationStatus, kind corev1.ResourceName, req, lim string) {
		c := &st.Containers[0]
		c.Recommended.Requests[kind] = qty(req)
		if lim != "" {
			c.Recommended.Limits[kind] = qty(lim)
		}
	}
	cases := []struct {
		name        string
		edit        func(*v1alpha1.RecommendationStatus)
		fixes, tags string
	}{
		{name: "over-provisioned", fixes: "nothing: no reliability problem the recommendation addresses", tags: "no OOM kills"},
		{name: "OOM killed, memory raised", edit: func(st *v1alpha1.RecommendationStatus) {
			st.OOMKills = 3
			raise(st, corev1.ResourceMemory, "1200Mi", "1500Mi")
			st.Reliability = &v1alpha1.Reliability{Containers: []v1alpha1.ContainerReliability{{Name: "app", MemoryPeakPercent: 100}}}
		}, fixes: "raises memory after OOM kills", tags: "3 OOM kills (fixed by applying); app memory peaked at 100% of its limit (fixed by applying)"},
		{name: "OOM killed, memory cut", edit: func(st *v1alpha1.RecommendationStatus) {
			st.OOMKills = 3
		}, fixes: "nothing", tags: "3 OOM kills"},
		{name: "throttled, limit raised", edit: func(st *v1alpha1.RecommendationStatus) {
			st.Containers[0].Current.Limits[corev1.ResourceCPU] = qty("400m")
			raise(st, corev1.ResourceCPU, "550m", "600m")
			st.Reliability = &v1alpha1.Reliability{Containers: []v1alpha1.ContainerReliability{{Name: "app", CPUThrottledPercent: 30}}}
		}, fixes: "app: raises the CPU limit it was throttled by 30% of the time", tags: "no OOM kills; app throttled by its CPU limit 30% of the time (fixed by applying)"},
		{name: "requests under usage", edit: func(st *v1alpha1.RecommendationStatus) {
			st.Containers[0].Current.Requests[corev1.ResourceCPU] = qty("50m")
		}, fixes: "app: raises the cpu request, which p95 usage was above", tags: "no OOM kills"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := status()
			if c.edit != nil {
				c.edit(&st)
			}
			req, _ := Build(Input{Kind: "Deployment", Status: st, Usage: usage(), Now: t0})
			if !strings.HasPrefix(req.State.Fixes, c.fixes) {
				t.Errorf("fixes %q, want %q", req.State.Fixes, c.fixes)
			}
			if req.State.Reliability != c.tags {
				t.Errorf("reliability %q, want %q", req.State.Reliability, c.tags)
			}
		})
	}
}

func TestVerdict(t *testing.T) {
	req, _ := Build(Input{Kind: "Deployment", Status: status(), Usage: usage(), Now: t0})
	ok := `{"model":"decider-4b","answers":{"verdict":{"choice":"review","probabilities":{"apply":0.3,"review":0.6,"wait":0.1}}}}`
	cases := []struct {
		name, body, err string
	}{
		{name: "valid", body: ok},
		{name: "missing answer", body: `{"answers":{}}`, err: `no answer to "verdict"`},
		{name: "choice not offered", body: strings.Replace(ok, `"choice":"review"`, `"choice":"maybe"`, 1), err: `answer "maybe" to "verdict" is not an option`},
		{name: "missing probability", body: strings.Replace(ok, `"wait":0.1`, `"later":0.1`, 1), err: `no probability for "wait" in "verdict"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var resp Response
			if err := json.Unmarshal([]byte(c.body), &resp); err != nil {
				t.Fatal(err)
			}
			got, err := Verdict(req, resp)
			if c.err != "" {
				if err == nil || err.Error() != c.err {
					t.Fatalf("err = %v, want %s", err, c.err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Choice != Review || len(got.Options) != 3 || got.Options[0].Name != Apply || got.Options[1].Probability != 0.6 {
				t.Errorf("answer %+v", got)
			}
		})
	}
}

func metav1Time(t time.Time) metav1.Time { return metav1.Time{Time: t} }
