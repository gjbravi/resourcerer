package server

import (
	"context"
	"fmt"
	"maps"
	"math"
	"net/http"
	"slices"
	"strings"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

const (
	throttledWarn  = 0.10
	memoryNearWarn = 0.90
	// Points per chart; each one is the worst value within its step.
	reliabilityPoints = 240
)

var liveWindows = map[string]time.Duration{"15m": 15 * time.Minute, "1h": time.Hour, "6h": 6 * time.Hour, "1d": 24 * time.Hour}

// reliability checks the workload over a window that ends now. Unlike the status, which the controller
// writes on its schedule, it is computed on request, so a fix shows up within minutes.
func (s *Server) reliability(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	name := r.URL.Query().Get("window")
	win, ok := liveWindows[name]
	if !ok {
		name = "learning"
		win = 7 * 24 * time.Hour
		if v, err := settings.ParseDuration(rec.Status.Settings[settings.LearningPeriod].Value); err == nil {
			win = v
		}
	}
	end := time.Now().Truncate(s.MinStep)
	key := strings.Join([]string{"reliability", rec.Namespace, rec.Name, name, end.String()}, "|")
	v := s.cache.get(key, 15*time.Second, func() (any, error) {
		return s.reliabilityReport(r.Context(), rec, name, end.Add(-win), end)
	})
	if err, isErr := v.(error); isErr {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, v)
}

func (s *Server) reliabilityReport(ctx context.Context, rec *v1alpha1.Recommendation, name string, start, end time.Time) (*ReliabilityReport, error) {
	win := end.Sub(start)
	step := max(win/reliabilityPoints, s.MinStep).Truncate(time.Second)
	ref := prom.Ref{Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name}
	rep := &ReliabilityReport{Window: name, From: float64(start.Unix()), To: float64(end.Unix()), Step: step.Seconds(),
		GeneratedAt: time.Now().UTC().Format(time.RFC3339), Checks: []Check{}}

	rateWindow := step
	if d, err := time.ParseDuration(s.Prom.RateWindow); err == nil {
		rateWindow = max(d, step)
	}
	throttledQ := s.Prom.Throttled(ref, win)
	throttled, err := s.Prom.Instant(ctx, throttledQ, "container")
	if err != nil {
		return nil, fmt.Errorf("query Prometheus: %w", err)
	}
	throttledSeriesQ := s.Prom.ThrottledRate(ref, rateWindow)
	throttledSeries, err := s.Prom.Range(ctx, throttledSeriesQ, start, end, step, "container")
	if err != nil {
		return nil, fmt.Errorf("query Prometheus: %w", err)
	}
	memoryQ := s.Prom.MemoryOfLimit(ref, step, s.MinStep)
	memory, err := s.Prom.Range(ctx, memoryQ, start, end, step, "container")
	if err != nil {
		return nil, fmt.Errorf("query Prometheus: %w", err)
	}
	restartsQ := prom.Restarts(ref, win)
	restarts, err := s.Prom.Instant(ctx, restartsQ, "container")
	if err != nil {
		return nil, fmt.Errorf("query Prometheus: %w", err)
	}

	for _, c := range rec.Status.Containers {
		if c.Skipped {
			continue
		}
		limits := workload.ToValues(c.Current.Limits)
		chk := Check{ID: "throttling", Name: "CPU throttling", Container: c.Name, Unit: "ratio", Threshold: engine.ThrottledShare,
			Source: "Share of CFS periods in which the kernel paused the container for hitting its CPU limit (cAdvisor).",
			Query:  throttledQ, Series: nonNil(throttledSeries[c.Name])}
		if _, has := limits[engine.CPU]; !has {
			chk.Status, chk.Display, chk.Detail, chk.Series = "ok", "no limit", "Without a CPU limit the container is never throttled.", nil
		} else {
			t := throttled[c.Name]
			chk.Value, chk.Display = t, percent(t)
			chk.Status = grade(t, throttledWarn, engine.ThrottledShare)
			chk.Detail = fmt.Sprintf("Paused in %s of 100ms periods.", percent(t))
		}
		rep.Checks = append(rep.Checks, chk)

		mem := Check{ID: "memory", Name: "Memory headroom", Container: c.Name, Unit: "ratio", Threshold: memoryNearWarn,
			Source: "Highest working set of any pod of the workload against the container's memory limit (cAdvisor).",
			Query:  memoryQ, Series: nonNil(memory[c.Name])}
		if lim, has := limits[engine.Memory]; !has || lim == 0 {
			mem.Status, mem.Display, mem.Detail, mem.Series = "ok", "no limit", "Without a memory limit the container is not OOM-killed for its own usage.", nil
		} else {
			peak := 0.0
			for _, p := range memory[c.Name] {
				peak = max(peak, p.V)
			}
			mem.Value, mem.Display = peak, percent(peak)
			mem.Status = "ok"
			if peak >= memoryNearWarn {
				mem.Status = "warn"
			}
			mem.Detail = fmt.Sprintf("Peak %s of the %s limit.", fmtBytes(peak*lim), fmtBytes(lim))
		}
		rep.Checks = append(rep.Checks, mem)
	}

	pods, hpa := s.workloadPods(ctx, rec), s.workloadHPA(ctx, rec)
	rep.Checks = append(rep.Checks, podChecks(pods, start, restarts, restartsQ)...)
	rep.Checks = append(rep.Checks, hpaCheck(hpa), qosCheck(pods))
	return rep, nil
}

// podChecks reads OOM kills, restarts and crash loops from pod status, the source the kubelet writes.
func podChecks(pods []corev1.Pod, since time.Time, restarts map[string]float64, restartsQ string) []Check {
	type exit struct {
		pod string
		t   *corev1.ContainerStateTerminated
	}
	oom := map[string][]exit{}
	last := map[string]exit{}
	crashing := map[string]string{}
	for _, p := range pods {
		for _, cs := range p.Status.ContainerStatuses {
			if w := cs.State.Waiting; w != nil && w.Reason == "CrashLoopBackOff" {
				crashing[cs.Name] = p.Name
			}
			t := cs.LastTerminationState.Terminated
			if t == nil || t.Reason == "Completed" || !t.FinishedAt.After(since) {
				continue
			}
			if t.Reason == "OOMKilled" {
				oom[cs.Name] = append(oom[cs.Name], exit{p.Name, t})
			}
			if l, seen := last[cs.Name]; !seen || t.FinishedAt.After(l.t.FinishedAt.Time) {
				last[cs.Name] = exit{p.Name, t}
			}
		}
	}

	n := 0
	var oomPod string
	var oomAt metav1.Time
	for _, es := range oom {
		for _, e := range es {
			n++
			if e.t.FinishedAt.After(oomAt.Time) {
				oomAt, oomPod = e.t.FinishedAt, e.pod
			}
		}
	}
	o := Check{ID: "oom", Name: "OOM kills", Status: "ok", Unit: "count", Value: float64(n), Display: fmt.Sprint(n),
		Detail: "No container was killed for exceeding its memory limit.",
		Source: "Containers whose last termination reason is OOMKilled (pod status, lastState.terminated)."}
	if n > 0 {
		o.Status, o.Pod = "bad", oomPod
		o.Detail = fmt.Sprintf("Last one at %s in pod %s.", oomAt.UTC().Format("Jan 2 15:04 UTC"), oomPod)
	}

	total := 0.0
	for _, v := range restarts {
		total += math.Round(v)
	}
	re := Check{ID: "restarts", Name: "Restarts", Status: "ok", Unit: "count", Value: total, Display: fmt.Sprint(int(total)),
		Detail: "No container restarted.", Query: restartsQ,
		Source: "Increase of kube_pod_container_status_restarts_total (kube-state-metrics); the last exit comes from pod status."}
	if total > 0 {
		re.Status = "warn"
		var newest exit
		for _, e := range last {
			if newest.t == nil || e.t.FinishedAt.After(newest.t.FinishedAt.Time) {
				newest = e
			}
		}
		re.Detail = fmt.Sprintf("%d in the window.", int(total))
		if newest.t != nil {
			re.Pod = newest.pod
			re.Detail = fmt.Sprintf("Last exit: %s, code %d, at %s in pod %s.", newest.t.Reason, newest.t.ExitCode,
				newest.t.FinishedAt.UTC().Format("Jan 2 15:04 UTC"), newest.pod)
		}
	}

	cl := Check{ID: "crashloop", Name: "Crash loop", Status: "ok", Unit: "text", Display: "none",
		Detail: "No container is waiting in CrashLoopBackOff.", Source: "Current container state (pod status, state.waiting.reason)."}
	if len(crashing) > 0 {
		cs := slices.Sorted(maps.Keys(crashing))
		cl.Status, cl.Container, cl.Pod, cl.Display = "bad", strings.Join(cs, ", "), crashing[cs[0]], "CrashLoopBackOff"
		cl.Detail = fmt.Sprintf("%s keeps crashing in pod %s.", cs[0], crashing[cs[0]])
	}
	return []Check{o, re, cl}
}

func hpaCheck(h *autoscalingv2.HorizontalPodAutoscaler) Check {
	c := Check{ID: "hpa", Name: "HPA headroom", Status: "ok", Unit: "text", Display: "no HPA",
		Detail: "No HorizontalPodAutoscaler targets the workload.", Source: "HorizontalPodAutoscaler status.currentReplicas and spec.maxReplicas."}
	if h == nil {
		return c
	}
	c.Container = h.Name
	c.Display = fmt.Sprintf("%d / %d replicas", h.Status.CurrentReplicas, h.Spec.MaxReplicas)
	c.Detail = fmt.Sprintf("HPA %s can still add %d replicas.", h.Name, max(0, h.Spec.MaxReplicas-h.Status.CurrentReplicas))
	if h.Status.CurrentReplicas >= h.Spec.MaxReplicas {
		c.Status = "warn"
		c.Detail = fmt.Sprintf("HPA %s runs at its maximum of %d replicas.", h.Name, h.Spec.MaxReplicas)
	}
	return c
}

func qosCheck(pods []corev1.Pod) Check {
	c := Check{ID: "qos", Name: "QoS class", Status: "ok", Unit: "text", Display: "no pods",
		Detail: "No running pods.", Source: "Pod status.qosClass, set by Kubernetes from requests and limits."}
	count := map[corev1.PodQOSClass]int{}
	for _, p := range pods {
		count[p.Status.QOSClass]++
	}
	for _, q := range []corev1.PodQOSClass{corev1.PodQOSBestEffort, corev1.PodQOSBurstable, corev1.PodQOSGuaranteed} {
		if count[q] > 0 {
			c.Display = string(q)
			c.Detail = fmt.Sprintf("%d of %d pods are %s.", count[q], len(pods), q)
			break
		}
	}
	if count[corev1.PodQOSBestEffort] > 0 {
		c.Status = "warn"
		c.Detail = fmt.Sprintf("%d of %d pods have no requests or limits and are evicted first under node pressure.", count[corev1.PodQOSBestEffort], len(pods))
	}
	return c
}

// workloadPods lists pods straight from the API server: caching every pod in the cluster would cost more
// memory than these occasional reads.
func (s *Server) workloadPods(ctx context.Context, rec *v1alpha1.Recommendation) []corev1.Pod {
	obj, err := workload.NewObject(rec.Spec.TargetRef.Kind)
	if err != nil || s.Client.Get(ctx, client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name}, obj) != nil {
		return nil
	}
	wl, err := workload.FromObject(obj)
	if err != nil || workload.Resolve(ctx, s.Client, wl) != nil || wl.Selector == nil {
		return nil
	}
	sel, err := metav1.LabelSelectorAsSelector(wl.Selector)
	if err != nil || sel == labels.Nothing() {
		return nil
	}
	var l corev1.PodList
	if s.APIReader.List(ctx, &l, client.InNamespace(rec.Namespace), client.MatchingLabelsSelector{Selector: sel}) != nil {
		return nil
	}
	return l.Items
}

func (s *Server) workloadHPA(ctx context.Context, rec *v1alpha1.Recommendation) *autoscalingv2.HorizontalPodAutoscaler {
	var l autoscalingv2.HorizontalPodAutoscalerList
	if s.APIReader.List(ctx, &l, client.InNamespace(rec.Namespace)) != nil {
		return nil
	}
	for i := range l.Items {
		t := l.Items[i].Spec.ScaleTargetRef
		if t.Kind == rec.Spec.TargetRef.Kind && t.Name == rec.Spec.TargetRef.Name {
			return &l.Items[i]
		}
	}
	return nil
}

func grade(v, warn, bad float64) string {
	switch {
	case v >= bad:
		return "bad"
	case v >= warn:
		return "warn"
	}
	return "ok"
}

func percent(v float64) string { return fmt.Sprintf("%d%%", int(math.Round(v*100))) }

func fmtBytes(b float64) string {
	if b >= engine.Gi {
		return fmt.Sprintf("%.2fGiB", b/engine.Gi)
	}
	return fmt.Sprintf("%dMiB", int(math.Round(b/engine.Mi)))
}
