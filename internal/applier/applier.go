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
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
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
	r := &Reconciler{Client: mgr.GetClient(), APIReader: mgr.GetAPIReader(), Recorder: mgr.GetEventRecorder("resourcerer-applier"), Now: time.Now}
	return ctrl.NewControllerManagedBy(mgr).Named("applier").For(&v1alpha1.Recommendation{}).Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	rec := &v1alpha1.Recommendation{}
	if err := r.Get(ctx, req.NamespacedName, rec); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	key := req.String()
	if ar := rec.Spec.ApplyRequest; ar != nil && ar.ID != "" && ar.ID != rec.Status.LastHandledRequestID {
		id := "request:" + ar.ID
		if r.seen(key, id) {
			return ctrl.Result{}, nil
		}
		r.remember(key, id)
		return ctrl.Result{}, r.handle(ctx, rec, *ar)
	}
	if r.autoDue(rec) {
		id := "auto:" + rec.Status.Hash
		if r.seen(key, id) {
			return ctrl.Result{}, nil
		}
		r.remember(key, id)
		return ctrl.Result{}, r.apply(ctx, rec, AutoDeployUser, nil)
	}
	return ctrl.Result{}, nil
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

func (r *Reconciler) handle(ctx context.Context, rec *v1alpha1.Recommendation, ar v1alpha1.ApplyRequest) error {
	switch ar.Action {
	case v1alpha1.ActionApply:
		if ar.Hash != rec.Status.Hash {
			return r.finish(ctx, rec, &ar, "Rejected", "the recommendation changed since it was reviewed; reload and review it again")
		}
		if rec.Status.Phase != v1alpha1.PhaseReady && rec.Status.Phase != v1alpha1.PhaseLearning {
			return r.finish(ctx, rec, &ar, "Rejected", fmt.Sprintf("cannot apply a %s recommendation", rec.Status.Phase))
		}
		return r.apply(ctx, rec, ar.RequestedBy, &ar)
	case v1alpha1.ActionRevert:
		return r.revert(ctx, rec, ar)
	}
	return r.finish(ctx, rec, &ar, "Rejected", "unknown action "+string(ar.Action))
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
	entry := v1alpha1.HistoryEntry{Time: metav1.Time{Time: r.Now()}, Action: v1alpha1.ActionApply, By: by, Hash: rec.Status.Hash, Containers: changes}
	r.event(ctx, rec, corev1.EventTypeNormal, "RecommendationApplied", "Apply", "Applied %s by %s: %s", rec.Status.Hash, by, describe(changes))
	return r.record(ctx, rec, ar, entry)
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
	return r.record(ctx, rec, &ar, entry)
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
