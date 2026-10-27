// Package recommender keeps one Recommendation per managed workload and regenerates
// it from Prometheus history on each workload's schedule.
package recommender

import (
	"context"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/cluster"
	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

type Reconciler struct {
	client.Client
	// Rollouts is true when the Argo Rollouts CRD is installed.
	Rollouts bool
	// emitted remembers the last announced hash per workload; the cache can lag behind status patches.
	emitted  sync.Map
	Kind     string
	Cluster  cluster.Config
	Prom     *prom.Client
	MinStep  time.Duration
	Recorder events.EventRecorder
	Now      func() time.Time
}

// RolloutsInstalled reports whether the API server serves Argo Rollouts.
func RolloutsInstalled(mgr ctrl.Manager) bool {
	_, err := mgr.GetRESTMapper().RESTMapping(workload.RolloutGVK.GroupKind(), workload.RolloutGVK.Version)
	return err == nil
}

func Setup(mgr ctrl.Manager, cfg cluster.Config, p *prom.Client, minStep time.Duration) error {
	rollouts := RolloutsInstalled(mgr)
	kinds := workload.Kinds
	if rollouts {
		kinds = workload.AllKinds
	}
	log.Log.WithName("recommender").Info("watching workload kinds", "kinds", kinds, "argoRollouts", rollouts)
	for _, kind := range kinds {
		r := &Reconciler{Client: mgr.GetClient(), Kind: kind, Cluster: cfg, Prom: p, MinStep: minStep, Rollouts: rollouts,
			Recorder: mgr.GetEventRecorder("resourcerer"), Now: time.Now}
		if err := r.setup(mgr); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) setup(mgr ctrl.Manager) error {
	obj, _ := workload.NewObject(r.Kind)
	notExcluded := predicate.NewPredicateFuncs(func(o client.Object) bool { return !r.Cluster.Excluded(o.GetNamespace()) })
	b := ctrl.NewControllerManagedBy(mgr)
	switch {
	case r.Kind == workload.Rollout:
		// A Rollout's template can live in a referenced Deployment; changes there change the Rollout's current values.
		b = b.Watches(&appsv1.Deployment{}, handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, d client.Object) []reconcile.Request {
			if name, ok := workload.ReferencedDeployments(ctx, r.Client, d.GetNamespace())[d.GetName()]; ok {
				return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: d.GetNamespace(), Name: name}}}
			}
			return nil
		}), builder.WithPredicates(notExcluded, predicate.GenerationChangedPredicate{}))
	case r.Kind == "Deployment" && r.Rollouts:
		// Deployments used as a Rollout's workloadRef are optimized through the Rollout.
		rollout, _ := workload.NewObject(workload.Rollout)
		b = b.Watches(rollout, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, o client.Object) []reconcile.Request {
			w, err := workload.FromObject(o)
			if err != nil || w.TemplateRef == nil {
				return nil
			}
			return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: w.TemplateRef.Name}}}
		}), builder.WithPredicates(notExcluded))
	}
	return b.
		Named("recommender-"+r.Kind).
		For(obj, builder.WithPredicates(notExcluded, predicate.Or(
			predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{}, predicate.AnnotationChangedPredicate{},
		))).
		// Spec changes on the Recommendation are regenerate requests.
		Owns(&v1alpha1.Recommendation{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.all(func(ns client.Object) []client.ListOption {
			return []client.ListOption{client.InNamespace(ns.GetName())}
		})), builder.WithPredicates(predicate.Or(predicate.LabelChangedPredicate{}, predicate.AnnotationChangedPredicate{}))).
		Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(r.all(func(client.Object) []client.ListOption { return nil })),
			builder.WithPredicates(predicate.NewPredicateFuncs(func(o client.Object) bool {
				return o.GetNamespace() == r.Cluster.DefaultsConfigMap.Namespace && o.GetName() == r.Cluster.DefaultsConfigMap.Name
			}))).
		Watches(&autoscalingv2.HorizontalPodAutoscaler{}, handler.EnqueueRequestsFromMapFunc(func(_ context.Context, o client.Object) []reconcile.Request {
			h := o.(*autoscalingv2.HorizontalPodAutoscaler)
			if h.Spec.ScaleTargetRef.Kind != r.Kind {
				return nil
			}
			return []reconcile.Request{{NamespacedName: client.ObjectKey{Namespace: h.Namespace, Name: h.Spec.ScaleTargetRef.Name}}}
		})).
		Complete(r)
}

func (r *Reconciler) all(opts func(client.Object) []client.ListOption) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		objs, err := workload.List(ctx, r.Client, r.Kind, opts(o)...)
		if err != nil {
			log.FromContext(ctx).Error(err, "list workloads")
			return nil
		}
		reqs := make([]reconcile.Request, 0, len(objs))
		for _, w := range objs {
			if !r.Cluster.Excluded(w.GetNamespace()) {
				reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(w)})
			}
		}
		return reqs
	}
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	obj, _ := workload.NewObject(r.Kind)
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			forget(req.Namespace, r.Kind, req.Name)
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	w, err := workload.FromObject(obj)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := workload.Resolve(ctx, r.Client, w); err != nil {
		return ctrl.Result{}, err
	}
	var ns corev1.Namespace
	if err := r.Get(ctx, client.ObjectKey{Name: w.Namespace()}, &ns); err != nil {
		return ctrl.Result{}, err
	}

	rec := &v1alpha1.Recommendation{}
	key := client.ObjectKey{Namespace: w.Namespace(), Name: w.ResourceName()}
	err = r.Get(ctx, key, rec)
	exists := err == nil
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}

	rolloutTemplate := false
	if r.Kind == "Deployment" && r.Rollouts {
		_, rolloutTemplate = workload.ReferencedDeployments(ctx, r.Client, w.Namespace())[w.Name()]
	}
	if !r.Cluster.Managed(&ns, obj.GetLabels()) || !obj.GetDeletionTimestamp().IsZero() || rolloutTemplate {
		forget(w.Namespace(), w.Kind, w.Name())
		if exists {
			return ctrl.Result{}, client.IgnoreNotFound(r.Delete(ctx, rec))
		}
		return ctrl.Result{}, nil
	}
	if !exists {
		rec = &v1alpha1.Recommendation{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name, OwnerReferences: []metav1.OwnerReference{w.OwnerReference()}},
			Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{APIVersion: w.APIVersion, Kind: w.Kind, Name: w.Name()}},
		}
		if err := r.Create(ctx, rec); err != nil {
			return ctrl.Result{}, err
		}
	}

	defaults, rates := r.Cluster.Defaults(ctx, r.Client)
	eff := settings.Resolve(defaults, ns.Annotations, obj.GetAnnotations())
	period, _ := settings.ParseSchedule(eff.Get(settings.Schedule))
	now := r.Now()
	inputs := r.containerInputs(w)

	st := rec.Status
	due := st.GeneratedAt == nil || st.NextScheduleAt == nil || !now.Before(st.NextScheduleAt.Time) ||
		(rec.Spec.RegenerateRequestedAt != nil && rec.Spec.RegenerateRequestedAt.After(st.GeneratedAt.Time)) ||
		!sameSettings(st.Settings, eff) || !sameCurrent(st.Containers, inputs) || st.Phase == v1alpha1.PhaseError
	if !due {
		export(rec)
		return ctrl.Result{RequeueAfter: time.Until(st.NextScheduleAt.Time) + time.Second}, nil
	}

	in, avg, err := r.gather(ctx, w, eff, inputs, now)
	patch := client.MergeFrom(rec.DeepCopy())
	if err != nil {
		logger.Error(err, "read usage history")
		rec.Status.Phase = v1alpha1.PhaseError
		rec.Status.Message = "Metrics unavailable: " + err.Error()
		if perr := r.Status().Patch(ctx, rec, patch); perr != nil {
			return ctrl.Result{}, perr
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	out := engine.Recommend(in)
	sum := engine.Summarize(in.Containers, out, avg, rates)
	threshold, _ := settings.ParsePercent(eff.Get(settings.AutoThreshold))
	hash := engine.Hash(out)
	changed := hash != rec.Status.Hash

	next := nextSlot(now, period, w.Namespace()+"/"+w.Kind+"/"+w.Name())
	if out.Phase == "Pending" {
		// Retry soon instead of waiting a full period for the first usable history.
		next = now.Add(min(period, time.Minute))
	}
	rec.Status.TemplateRef = nil
	if t := w.TemplateRef; t != nil {
		rec.Status.TemplateRef = &v1alpha1.TargetRef{APIVersion: t.APIVersion, Kind: t.Kind, Name: t.Name}
	}
	rec.Status.Phase = v1alpha1.Phase(out.Phase)
	rec.Status.Message = out.Message
	rec.Status.GeneratedAt = &metav1.Time{Time: now}
	rec.Status.NextScheduleAt = &metav1.Time{Time: next}
	rec.Status.Hash = hash
	rec.Status.Pending = out.Phase != "Pending" && engine.Pending(in.Containers, out, threshold)
	rec.Status.Settings = toSettings(eff)
	rec.Status.Warnings = eff.Warnings
	rec.Status.Containers = toContainers(in.Containers, out)
	rec.Status.HPA = nil
	if out.HPA != nil {
		rec.Status.HPA = &v1alpha1.HPARecommendation{Name: out.HPA.Name, Current: out.HPA.Current, Recommended: out.HPA.Recommended, Clamped: out.HPA.Clamped}
	}
	rec.Status.Score = nil
	if out.Phase != "Pending" && out.Phase != "Unsupported" {
		rec.Status.Score = &v1alpha1.Score{Overall: sum.Score, CPU: sum.CPUScore, Memory: sum.MemoryScore}
	}
	rec.Status.Impact = &v1alpha1.Impact{
		AvgReplicas: strconv.FormatFloat(avg, 'f', 2, 64),
		Current:     totals(sum.Current),
		Recommended: totals(sum.Recommended),
	}
	rec.Status.OOMKills = int32(r.oomKills(ctx, w, now.Add(-eff.Duration(settings.LearningPeriod))))
	if err := r.Status().Patch(ctx, rec, patch); err != nil {
		return ctrl.Result{}, err
	}
	export(rec)
	last, _ := r.emitted.Swap(key.String(), hash)
	if changed && last != hash && out.Phase != "Pending" {
		r.Recorder.Eventf(obj, rec, corev1.EventTypeNormal, "RecommendationGenerated", "Generate",
			"%s recommendation %s (score %d)", out.Phase, hash, sum.Score)
	}
	return ctrl.Result{RequeueAfter: time.Until(next) + time.Second}, nil
}

func (r *Reconciler) containerInputs(w *workload.Workload) []engine.ContainerInput {
	cs := w.Containers()
	out := make([]engine.ContainerInput, len(cs))
	for i, c := range cs {
		out[i] = engine.ContainerInput{Name: c.Name, Requests: workload.ToValues(c.Resources.Requests), Limits: workload.ToValues(c.Resources.Limits)}
	}
	return out
}

// gather reads usage history, OOM kills and the HPA, and returns the engine input and average replicas.
func (r *Reconciler) gather(ctx context.Context, w *workload.Workload, eff settings.Effective, cs []engine.ContainerInput, now time.Time) (engine.Input, float64, error) {
	learning := eff.Duration(settings.LearningPeriod)
	start := now.Add(-learning)
	step := prom.Step(learning, r.MinStep)
	ref := prom.Ref{Namespace: w.Namespace(), Kind: w.Kind, Name: w.Name()}

	cpu, err := r.Prom.Range(ctx, r.Prom.BusiestPod(ref, "cpu"), start, now, step, "container")
	if err != nil {
		return engine.Input{}, 0, err
	}
	mem, err := r.Prom.Range(ctx, r.Prom.BusiestPod(ref, "memory"), start, now, step, "container")
	if err != nil {
		return engine.Input{}, 0, err
	}
	pods, err := r.Prom.Range(ctx, r.Prom.Pods(ref), now.Add(-min(learning, 7*24*time.Hour)), now, step, "")
	if err != nil {
		return engine.Input{}, 0, err
	}

	oom := r.oomContainers(ctx, w, start)
	first := now
	for i := range cs {
		cs[i].CPU = prom.Values(cpu[cs[i].Name])
		cs[i].Memory = prom.Values(mem[cs[i].Name])
		cs[i].OOMKilled = oom[cs[i].Name]
		for _, s := range [][]prom.Point{cpu[cs[i].Name], mem[cs[i].Name]} {
			if len(s) > 0 && time.Unix(int64(s[0].T), 0).Before(first) {
				first = time.Unix(int64(s[0].T), 0)
			}
		}
	}
	avg := prom.Mean(pods[""])
	if len(pods[""]) == 0 {
		avg = float64(w.Replicas)
	}
	return engine.Input{
		Settings:   eff,
		Containers: cs,
		HPA:        r.hpa(ctx, w),
		History:    now.Sub(first),
		Replicas:   w.Replicas,
		Expected:   int(learning / step),
	}, avg, nil
}

func (r *Reconciler) pods(ctx context.Context, w *workload.Workload) []corev1.Pod {
	sel, err := metav1.LabelSelectorAsSelector(w.Selector)
	if err != nil || sel == labels.Nothing() {
		return nil
	}
	var l corev1.PodList
	if err := r.List(ctx, &l, client.InNamespace(w.Namespace()), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil
	}
	return l.Items
}

func (r *Reconciler) oomContainers(ctx context.Context, w *workload.Workload, since time.Time) map[string]bool {
	out := map[string]bool{}
	for _, p := range r.pods(ctx, w) {
		for _, cs := range p.Status.ContainerStatuses {
			if t := cs.LastTerminationState.Terminated; t != nil && t.Reason == "OOMKilled" && t.FinishedAt.After(since) {
				out[cs.Name] = true
			}
		}
	}
	return out
}

// oomKills approximates kills in the window by restarts of containers whose last termination was an OOM kill.
func (r *Reconciler) oomKills(ctx context.Context, w *workload.Workload, since time.Time) int {
	n := 0
	for _, p := range r.pods(ctx, w) {
		for _, cs := range p.Status.ContainerStatuses {
			if t := cs.LastTerminationState.Terminated; t != nil && t.Reason == "OOMKilled" && t.FinishedAt.After(since) {
				n += int(max(cs.RestartCount, 1))
			}
		}
	}
	return n
}

func (r *Reconciler) hpa(ctx context.Context, w *workload.Workload) *engine.HPA {
	var l autoscalingv2.HorizontalPodAutoscalerList
	if err := r.List(ctx, &l, client.InNamespace(w.Namespace())); err != nil {
		return nil
	}
	for _, h := range l.Items {
		if h.Spec.ScaleTargetRef.Kind != w.Kind || h.Spec.ScaleTargetRef.Name != w.Name() {
			continue
		}
		out := &engine.HPA{Name: h.Name}
		for _, m := range h.Spec.Metrics {
			if m.Type != autoscalingv2.ResourceMetricSourceType || m.Resource == nil || m.Resource.Target.Type != autoscalingv2.UtilizationMetricType {
				continue
			}
			switch m.Resource.Name {
			case corev1.ResourceCPU:
				out.CPU = true
				if m.Resource.Target.AverageUtilization != nil {
					out.CPUTarget = *m.Resource.Target.AverageUtilization
				}
			case corev1.ResourceMemory:
				out.Memory = true
			}
		}
		return out
	}
	return nil
}

// nextSlot spreads workloads over the schedule period by hashing their name, so they do not all query at once.
func nextSlot(now time.Time, period time.Duration, key string) time.Time {
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	offset := time.Duration(h.Sum32()%uint32(period/time.Second)) * time.Second
	base := now.Truncate(period).Add(offset)
	for !base.After(now) {
		base = base.Add(period)
	}
	return base
}

func toSettings(e settings.Effective) map[string]v1alpha1.Setting {
	out := make(map[string]v1alpha1.Setting, len(e.Values))
	for k, v := range e.Values {
		out[k] = v1alpha1.Setting{Value: v.Value, Source: v.Source}
	}
	return out
}

func sameSettings(a map[string]v1alpha1.Setting, e settings.Effective) bool {
	if len(a) != len(e.Values) {
		return false
	}
	for k, v := range e.Values {
		if a[k].Value != v.Value || a[k].Source != v.Source {
			return false
		}
	}
	return true
}

func sameCurrent(st []v1alpha1.ContainerRecommendation, in []engine.ContainerInput) bool {
	if len(st) != len(in) {
		return false
	}
	for i, c := range in {
		if st[i].Name != c.Name ||
			fmt.Sprint(workload.ToValues(st[i].Current.Requests)) != fmt.Sprint(workload.ToValues(workload.ToList(c.Requests))) ||
			fmt.Sprint(workload.ToValues(st[i].Current.Limits)) != fmt.Sprint(workload.ToValues(workload.ToList(c.Limits))) {
			return false
		}
	}
	return true
}

func toContainers(in []engine.ContainerInput, out engine.Output) []v1alpha1.ContainerRecommendation {
	res := make([]v1alpha1.ContainerRecommendation, len(in))
	for i, c := range in {
		o := out.Containers[i]
		res[i] = v1alpha1.ContainerRecommendation{
			Name:        c.Name,
			Skipped:     o.Skipped,
			Current:     corev1.ResourceRequirements{Requests: workload.ToList(c.Requests), Limits: workload.ToList(c.Limits)},
			Recommended: corev1.ResourceRequirements{Requests: workload.ToList(o.Requests), Limits: workload.ToList(o.Limits)},
			Basis:       o.Basis,
		}
	}
	return res
}

func totals(t engine.Totals) v1alpha1.Totals {
	return v1alpha1.Totals{
		CPU:         strconv.FormatFloat(t.CPU, 'f', 3, 64),
		Memory:      strconv.FormatFloat(t.Memory, 'f', 0, 64),
		MonthlyCost: strconv.FormatFloat(t.MonthlyCost, 'f', 2, 64),
	}
}
