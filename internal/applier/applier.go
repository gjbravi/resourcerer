// Package applier changes workload resources: on request from the UI (Apply, Revert)
// and automatically when auto-deploy is enabled. It is the only component that writes
// to workloads, so read-only installs simply leave it out.
package applier

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// FieldManager owns the resources fields resourcerer sets, so GitOps tools can ignore them by manager.
const FieldManager = "resourcerer"

const AutoDeployUser = "auto-deploy"

const maxHistory = 10

type Reconciler struct {
	client.Client
	APIReader client.Reader
	Recorder  events.EventRecorder
	Now       func() time.Time

	// done remembers handled requests and auto-applied hashes per Recommendation: right after a
	// status patch the cache can still serve the old object, which would apply the same thing twice.
	mu   sync.Mutex
	done map[string]string
	// applying remembers the applies whose rollout has not finished, which a rate limit counts. The
	// applier is leader-elected, so one process sees every apply it starts.
	applying map[types.NamespacedName]time.Time
}

func (r *Reconciler) seen(key, id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.done[key+"|"+id] != ""
}

func (r *Reconciler) remember(key, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done == nil {
		r.done = map[string]string{}
	}
	r.done[key+"|"+id] = id
}

func Setup(mgr ctrl.Manager) error {
	r := &Reconciler{Client: mgr.GetClient(), APIReader: mgr.GetAPIReader(), Recorder: mgr.GetEventRecorder("resourcerer-applier"),
		Now: time.Now, applying: map[types.NamespacedName]time.Time{}}
	if err := ctrl.NewControllerManagedBy(mgr).Named("applier").For(&v1alpha1.Recommendation{}).
		// A policy change has to reach the workloads it governs at once: a freeze or a narrower window is
		// often written because something is about to happen.
		Watches(&v1alpha1.ClusterApplyPolicy{}, handler.EnqueueRequestsFromMapFunc(r.everyWorkload)).
		Watches(&v1alpha1.ApplyPolicy{}, handler.EnqueueRequestsFromMapFunc(r.namespaceWorkloads)).
		Complete(r); err != nil {
		return err
	}
	return mgr.Add(&policyReporter{Reconciler: r})
}

// everyWorkload enqueues every Recommendation, for a change to a cluster-scoped policy.
func (r *Reconciler) everyWorkload(ctx context.Context, _ client.Object) []reconcile.Request {
	var recs v1alpha1.RecommendationList
	if err := r.List(ctx, &recs); err != nil {
		log.FromContext(ctx).Error(err, "list recommendations for a policy change")
		return nil
	}
	out := make([]reconcile.Request, 0, len(recs.Items))
	for i := range recs.Items {
		out = append(out, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&recs.Items[i])})
	}
	return out
}

// namespaceWorkloads enqueues the Recommendations of the policy's namespace.
func (r *Reconciler) namespaceWorkloads(ctx context.Context, obj client.Object) []reconcile.Request {
	var recs v1alpha1.RecommendationList
	if err := r.List(ctx, &recs, client.InNamespace(obj.GetNamespace())); err != nil {
		log.FromContext(ctx).Error(err, "list recommendations for a policy change")
		return nil
	}
	out := make([]reconcile.Request, 0, len(recs.Items))
	for i := range recs.Items {
		out = append(out, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&recs.Items[i])})
	}
	return out
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec := &v1alpha1.Recommendation{}
	if err := r.Get(ctx, req.NamespacedName, rec); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	key := req.String()
	// A revert is never held by a policy: it is how a bad apply is undone, and a freeze or the cooldown
	// right after that apply is exactly when it has to stay possible.
	if ar := rec.Spec.ApplyRequest; ar != nil && ar.ID != "" && ar.ID != rec.Status.LastHandledRequestID && ar.Action == v1alpha1.ActionRevert {
		if r.seen(key, "request:"+ar.ID) {
			return ctrl.Result{}, nil
		}
		r.remember(key, "request:"+ar.ID)
		return ctrl.Result{}, r.revert(ctx, rec, *ar)
	}

	res, err := r.policyDecision(ctx, rec)
	if err != nil {
		// "Could not read the rules" never means "no rules": nothing is applied, and the reason is recorded.
		log.FromContext(ctx).Error(err, "read the apply policies")
		if perr := r.recordPolicy(ctx, rec, policyRefused(err)); perr != nil {
			return ctrl.Result{}, perr
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	if err := r.recordPolicy(ctx, rec, res); err != nil {
		return ctrl.Result{}, err
	}
	r.releaseApplying(ctx, rec)
	due := r.autoDue(rec) || (rec.Spec.ApplyRequest != nil && rec.Spec.ApplyRequest.ID != rec.Status.LastHandledRequestID)
	requeue := ctrl.Result{RequeueAfter: requeueAfter(res, r.Now(), due)}

	if ar := rec.Spec.ApplyRequest; ar != nil && ar.ID != "" && ar.ID != rec.Status.LastHandledRequestID {
		if r.seen(key, "request:"+ar.ID) {
			return requeue, nil
		}
		r.remember(key, "request:"+ar.ID)
		return requeue, r.handle(ctx, rec, *ar, res)
	}
	if r.autoDue(rec) {
		if d := res.Auto; d.State == v1alpha1.ApplyBlocked {
			// Held, not skipped: the applier looks again when the rule that holds it changes.
			deferApply(d.BlockedBy)
			r.event(ctx, rec, corev1.EventTypeNormal, "ApplyDeferred", "Defer", "Automatic apply held by %s: %s", d.Policy, d.Reason)
			return requeue, nil
		}
		id := "auto:" + rec.Status.Hash
		if r.seen(key, id) {
			return requeue, nil
		}
		r.remember(key, id)
		return requeue, r.apply(ctx, rec, AutoDeployUser, nil)
	}
	return requeue, nil
}

func (r *Reconciler) autoDue(rec *v1alpha1.Recommendation) bool {
	st := rec.Status
	if st.Settings[settings.AutoDeploy].Value != "Enabled" || st.Phase != v1alpha1.PhaseReady || !st.Pending || st.Hash == "" {
		return false
	}
	// Do not re-apply a recommendation that was already applied, or reverted by someone.
	for _, h := range st.History {
		if h.Hash == st.Hash {
			return false
		}
	}
	return true
}

func (r *Reconciler) handle(ctx context.Context, rec *v1alpha1.Recommendation, ar v1alpha1.ApplyRequest, res policy.Result) error {
	if ar.Action != v1alpha1.ActionApply {
		return r.finish(ctx, rec, &ar, "Rejected", "unknown action "+string(ar.Action))
	}
	if ar.Hash != rec.Status.Hash {
		return r.finish(ctx, rec, &ar, "Rejected", "the recommendation changed since it was reviewed; reload and review it again")
	}
	if rec.Status.Phase != v1alpha1.PhaseReady && rec.Status.Phase != v1alpha1.PhaseLearning {
		return r.finish(ctx, rec, &ar, "Rejected", fmt.Sprintf("cannot apply a %s recommendation", rec.Status.Phase))
	}
	// The server refuses a held click before writing the request, but a policy may have changed since the
	// decision in the status was recorded, so the applier decides once more on what it reads now.
	if d := res.Manual; d.State == v1alpha1.ApplyBlocked {
		deferApply(d.BlockedBy)
		return r.finish(ctx, rec, &ar, "Rejected", heldReason(d))
	}
	return r.apply(ctx, rec, ar.RequestedBy, &ar)
}

func (r *Reconciler) apply(ctx context.Context, rec *v1alpha1.Recommendation, by string, ar *v1alpha1.ApplyRequest) error {
	live, err := r.liveResources(ctx, rec)
	if err != nil {
		return r.finish(ctx, rec, ar, "Failed", err.Error())
	}
	var changes []v1alpha1.ContainerChange
	for _, c := range rec.Status.Containers {
		if c.Skipped {
			continue
		}
		changes = append(changes, v1alpha1.ContainerChange{Name: c.Name, From: live[c.Name], To: c.Recommended})
	}
	if err := r.patchWorkload(ctx, rec, changes); err != nil {
		return r.finish(ctx, rec, ar, "Failed", err.Error())
	}
	if h := rec.Status.HPA; h != nil {
		if err := r.patchHPA(ctx, rec.Namespace, h.Name, h.Recommended); err != nil {
			return r.finish(ctx, rec, ar, "Failed", "resources applied, HPA target not updated: "+err.Error())
		}
	}
	r.markApplying(rec)
	entry := v1alpha1.HistoryEntry{Time: metav1.Time{Time: r.Now()}, Action: v1alpha1.ActionApply, By: by, Hash: rec.Status.Hash, Containers: changes}
	r.event(ctx, rec, corev1.EventTypeNormal, "RecommendationApplied", "Apply", "Applied %s by %s: %s", rec.Status.Hash, by, describe(changes))
	if err := r.record(ctx, rec, ar, entry); err != nil {
		return err
	}
	countApply("Succeeded")
	return nil
}

func (r *Reconciler) revert(ctx context.Context, rec *v1alpha1.Recommendation, ar v1alpha1.ApplyRequest) error {
	var last *v1alpha1.HistoryEntry
	for i := range rec.Status.History {
		if rec.Status.History[i].Action == v1alpha1.ActionApply {
			last = &rec.Status.History[i]
			break
		}
	}
	if last == nil {
		return r.finish(ctx, rec, &ar, "Rejected", "nothing to revert")
	}
	changes := make([]v1alpha1.ContainerChange, len(last.Containers))
	for i, c := range last.Containers {
		changes[i] = v1alpha1.ContainerChange{Name: c.Name, From: c.To, To: c.From}
	}
	if err := r.patchWorkload(ctx, rec, changes); err != nil {
		return r.finish(ctx, rec, &ar, "Failed", err.Error())
	}
	entry := v1alpha1.HistoryEntry{Time: metav1.Time{Time: r.Now()}, Action: v1alpha1.ActionRevert, By: ar.RequestedBy, Hash: last.Hash, Containers: changes}
	r.event(ctx, rec, corev1.EventTypeNormal, "RecommendationReverted", "Revert", "Reverted %s by %s: %s", last.Hash, ar.RequestedBy, describe(changes))
	if err := r.record(ctx, rec, &ar, entry); err != nil {
		return err
	}
	// A revert starts a rollout like any other apply, and counts against the limits while it runs.
	r.markApplying(rec)
	revertedCounter.WithLabelValues("manual").Inc()
	return nil
}

// liveResources reads container resources from the API server, not the cache, so history records what was really replaced.
func (r *Reconciler) liveResources(ctx context.Context, rec *v1alpha1.Recommendation) (map[string]corev1.ResourceRequirements, error) {
	w, err := workload.Get(ctx, r.APIReader, rec.Spec.TargetRef.Kind, client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name})
	if err != nil {
		return nil, err
	}
	out := map[string]corev1.ResourceRequirements{}
	for _, c := range w.Containers() {
		out[c.Name] = c.Resources
	}
	return out, nil
}

// patchWorkload server-side applies only the resources of the listed containers; other fields keep their owners.
func (r *Reconciler) patchWorkload(ctx context.Context, rec *v1alpha1.Recommendation, changes []v1alpha1.ContainerChange) error {
	// A Rollout with spec.workloadRef takes its template from a Deployment; changing it starts a new rollout.
	t := rec.Spec.TargetRef
	if rec.Status.TemplateRef != nil {
		t = *rec.Status.TemplateRef
	}
	containers := make([]map[string]any, 0, len(changes))
	for _, c := range changes {
		res := map[string]any{}
		if len(c.To.Requests) > 0 {
			res["requests"] = c.To.Requests
		}
		if len(c.To.Limits) > 0 {
			res["limits"] = c.To.Limits
		}
		containers = append(containers, map[string]any{"name": c.Name, "resources": res})
	}
	body, err := json.Marshal(map[string]any{
		"apiVersion": t.APIVersion,
		"kind":       t.Kind,
		"metadata":   map[string]any{"name": t.Name, "namespace": rec.Namespace},
		"spec":       map[string]any{"template": map[string]any{"spec": map[string]any{"containers": containers}}},
	})
	if err != nil {
		return err
	}
	obj, err := workload.NewObject(t.Kind)
	if err != nil {
		return err
	}
	obj.SetName(t.Name)
	obj.SetNamespace(rec.Namespace)
	force := true
	return r.Patch(ctx, obj, client.RawPatch(types.ApplyPatchType, body), &client.PatchOptions{FieldManager: FieldManager, Force: &force})
}

func (r *Reconciler) patchHPA(ctx context.Context, ns, name string, target int32) error {
	var h autoscalingv2.HorizontalPodAutoscaler
	if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: name}, &h); err != nil {
		return err
	}
	patch := client.MergeFrom(h.DeepCopy())
	for i, m := range h.Spec.Metrics {
		if m.Type == autoscalingv2.ResourceMetricSourceType && m.Resource != nil && m.Resource.Name == corev1.ResourceCPU {
			h.Spec.Metrics[i].Resource.Target.AverageUtilization = &target
		}
	}
	return r.Patch(ctx, &h, patch)
}

func (r *Reconciler) record(ctx context.Context, rec *v1alpha1.Recommendation, ar *v1alpha1.ApplyRequest, entry v1alpha1.HistoryEntry) error {
	patch := client.MergeFrom(rec.DeepCopy())
	rec.Status.History = append([]v1alpha1.HistoryEntry{entry}, rec.Status.History...)
	if len(rec.Status.History) > maxHistory {
		rec.Status.History = rec.Status.History[:maxHistory]
	}
	r.mark(rec, ar, "Succeeded", "")
	return r.Status().Patch(ctx, rec, patch)
}

func (r *Reconciler) finish(ctx context.Context, rec *v1alpha1.Recommendation, ar *v1alpha1.ApplyRequest, result, msg string) error {
	patch := client.MergeFrom(rec.DeepCopy())
	r.mark(rec, ar, result, msg)
	countApply(result)
	if result != "Succeeded" {
		r.event(ctx, rec, corev1.EventTypeWarning, "ApplyFailed", "Apply", "%s: %s", result, msg)
		log.FromContext(ctx).Info("apply not done", "result", result, "message", msg)
	}
	return r.Status().Patch(ctx, rec, patch)
}

func (r *Reconciler) mark(rec *v1alpha1.Recommendation, ar *v1alpha1.ApplyRequest, result, msg string) {
	if ar != nil {
		t := ar.RequestedAt
		rec.Status.LastHandledRequestAt = &t
		rec.Status.LastHandledRequestID = ar.ID
	}
	rec.Status.LastApplyResult = &v1alpha1.ApplyResult{Time: metav1.Time{Time: r.Now()}, Result: result, Message: msg}
}

func (r *Reconciler) event(ctx context.Context, rec *v1alpha1.Recommendation, typ, reason, action, note string, args ...any) {
	obj, err := workload.NewObject(rec.Spec.TargetRef.Kind)
	if err != nil {
		return
	}
	if err := r.Get(ctx, client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name}, obj); err != nil {
		if !apierrors.IsNotFound(err) {
			log.FromContext(ctx).Error(err, "get workload for event")
		}
		return
	}
	r.Recorder.Eventf(obj, rec, typ, reason, action, note, args...)
}

func describe(changes []v1alpha1.ContainerChange) string {
	parts := make([]string, 0, len(changes))
	for _, c := range changes {
		parts = append(parts, fmt.Sprintf("%s cpu %s→%s memory %s→%s", c.Name,
			q(c.From.Requests, corev1.ResourceCPU), q(c.To.Requests, corev1.ResourceCPU),
			q(c.From.Requests, corev1.ResourceMemory), q(c.To.Requests, corev1.ResourceMemory)))
	}
	return strings.Join(parts, "; ")
}

func q(l corev1.ResourceList, r corev1.ResourceName) string {
	if v, ok := l[r]; ok {
		return v.String()
	}
	return "none"
}
