package applier

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// notInstalled reports whether the apply policy CRDs are missing, which an installation upgraded without
// applying charts/resourcerer/crds hits: it runs ungated and says so, instead of refusing every apply.
func notInstalled(err error) bool {
	return meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err)
}

// policyDecision reads what governs the workload and decides when an automatic and a manual apply may
// start. The policies are read uncached, right before the decision, so a change holds the next apply.
func (r *Reconciler) policyDecision(ctx context.Context, rec *v1alpha1.Recommendation) (policy.Result, error) {
	w, err := workload.Get(ctx, r.Client, rec.Spec.TargetRef.Kind,
		client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name})
	if err != nil {
		return policy.Result{}, fmt.Errorf("read the workload: %w", err)
	}
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: rec.Namespace}, ns); err != nil {
		return policy.Result{}, fmt.Errorf("read namespace %s: %w", rec.Namespace, err)
	}
	policies, err := r.policies(ctx, rec.Namespace)
	if err != nil {
		return policy.Result{}, err
	}
	target := r.target(rec, w, ns.Labels)
	in := policy.Input{Now: r.Now(), Target: target, Policies: policies}
	if needsFleet(policies, target) {
		fleet, err := r.fleet(ctx, rec.Namespace, policies)
		if err != nil {
			return policy.Result{}, err
		}
		in.Fleet = fleet
		for _, p := range policies {
			if available, _, ok := policy.Slots(p, in); ok {
				slotsGauge.WithLabelValues(p.Name).Set(float64(available))
			}
		}
	}
	return policy.Decide(in), nil
}

// policies returns the cluster policies and the ones in the workload's namespace.
func (r *Reconciler) policies(ctx context.Context, namespace string) ([]policy.Policy, error) {
	var caps v1alpha1.ClusterApplyPolicyList
	if err := r.APIReader.List(ctx, &caps); err != nil {
		if notInstalled(err) {
			log.FromContext(ctx).Info("the apply policy CRDs are not installed: applies are ungated")
			return nil, nil
		}
		return nil, fmt.Errorf("list cluster apply policies: %w", err)
	}
	var aps v1alpha1.ApplyPolicyList
	if err := r.APIReader.List(ctx, &aps, client.InNamespace(namespace)); err != nil {
		if notInstalled(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list apply policies in %s: %w", namespace, err)
	}
	out := make([]policy.Policy, 0, len(caps.Items)+len(aps.Items))
	for _, c := range caps.Items {
		out = append(out, policy.Policy{
			Name: "ClusterApplyPolicy/" + c.Name, Namespaces: c.Spec.NamespaceSelector, Spec: c.Spec.ApplyRules,
		})
	}
	for _, a := range aps.Items {
		out = append(out, policy.Policy{
			Name: "ApplyPolicy/" + a.Namespace + "/" + a.Name, Namespace: a.Namespace, Spec: a.Spec.ApplyRules,
		})
	}
	return out, nil
}

// target is the workload being decided, with the history the cooldown and the hourly limit count.
func (r *Reconciler) target(rec *v1alpha1.Recommendation, w *workload.Workload, nsLabels map[string]string) policy.Workload {
	t := policy.Workload{
		Namespace:       rec.Namespace,
		Kind:            rec.Spec.TargetRef.Kind,
		Name:            rec.Spec.TargetRef.Name,
		Labels:          w.Labels(),
		NamespaceLabels: nsLabels,
		AutoDeploy:      rec.Status.Settings[settings.AutoDeploy].Value == "Enabled",
		RollingOut:      w.InProgress(),
	}
	t.Changed, t.Applied = history(rec, r.Now())
	return t
}

// history returns the last change to the workload and the applies of the hour before now.
func history(rec *v1alpha1.Recommendation, now time.Time) (time.Time, []time.Time) {
	var changed time.Time
	var applied []time.Time
	since := now.Add(-time.Hour)
	for _, h := range rec.Status.History {
		if changed.IsZero() || h.Time.After(changed) {
			changed = h.Time.Time
		}
		if h.Action == v1alpha1.ActionApply && h.Time.After(since) && !h.Time.After(now) {
			applied = append(applied, h.Time.Time)
		}
	}
	return changed, applied
}

// needsFleet reports whether a rate limit the workload falls under counts other workloads too.
func needsFleet(policies []policy.Policy, target policy.Workload) bool {
	for _, p := range policies {
		rate := p.Spec.Rate
		if rate == nil || !policy.Selects(p, target) {
			continue
		}
		if rate.MaxConcurrent != nil || rate.MaxPerHour != nil ||
			(rate.PerNamespace != nil && rate.PerNamespace.MaxConcurrent != nil) {
			return true
		}
	}
	return false
}

// fleet is every managed workload, since a rate limit counts what the policy selects wherever it runs.
func (r *Reconciler) fleet(ctx context.Context, namespace string, policies []policy.Policy) ([]policy.Workload, error) {
	var recs v1alpha1.RecommendationList
	if err := r.List(ctx, &recs); err != nil {
		return nil, fmt.Errorf("list recommendations: %w", err)
	}
	nsLabels, err := r.namespaceLabels(ctx)
	if err != nil {
		return nil, err
	}
	rolling, err := r.rollingOut(ctx, namespace, policies)
	if err != nil {
		return nil, err
	}
	out := make([]policy.Workload, 0, len(recs.Items))
	for i := range recs.Items {
		rec := &recs.Items[i]
		fw := policy.Workload{
			Namespace:       rec.Namespace,
			Kind:            rec.Spec.TargetRef.Kind,
			Name:            rec.Spec.TargetRef.Name,
			Labels:          rec.Status.TargetLabels,
			NamespaceLabels: nsLabels[rec.Namespace],
			AutoDeploy:      rec.Status.Settings[settings.AutoDeploy].Value == "Enabled",
			Applying:        r.isApplying(rec),
			RollingOut:      rolling[rec.Spec.TargetRef.Kind+"/"+rec.Namespace+"/"+rec.Spec.TargetRef.Name],
		}
		fw.Changed, fw.Applied = history(rec, r.Now())
		out = append(out, fw)
	}
	return out, nil
}

// namespaceLabels is the labels of every namespace, which a ClusterApplyPolicy's namespaceSelector matches.
func (r *Reconciler) namespaceLabels(ctx context.Context) (map[string]map[string]string, error) {
	var list corev1.NamespaceList
	if err := r.List(ctx, &list); err != nil {
		return nil, fmt.Errorf("list namespaces: %w", err)
	}
	out := make(map[string]map[string]string, len(list.Items))
	for i := range list.Items {
		out[list.Items[i].Name] = list.Items[i].Labels
	}
	return out, nil
}

// rollingOut reports which workloads in the namespace are replacing pods right now, ours or anyone
// else's: perNamespace counts them all, as Karpenter counts nodes being deleted for any reason.
func (r *Reconciler) rollingOut(ctx context.Context, namespace string, policies []policy.Policy) (map[string]bool, error) {
	wanted := false
	for _, p := range policies {
		if p.Spec.Rate != nil && p.Spec.Rate.PerNamespace != nil && p.Spec.Rate.PerNamespace.MaxConcurrent != nil {
			wanted = true
			break
		}
	}
	if !wanted {
		return nil, nil
	}
	out := map[string]bool{}
	for _, kind := range workload.AllKinds {
		objs, err := workload.List(ctx, r.APIReader, kind, client.InNamespace(namespace))
		if err != nil {
			if notInstalled(err) {
				continue
			}
			return nil, fmt.Errorf("list %s in %s: %w", kind, namespace, err)
		}
		for _, obj := range objs {
			w, err := workload.FromObject(obj)
			if err != nil || !w.InProgress() {
				continue
			}
			out[w.Kind+"/"+w.Namespace()+"/"+w.Name()] = true
		}
	}
	return out, nil
}

func (r *Reconciler) isApplying(rec *v1alpha1.Recommendation) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.applying[types.NamespacedName{Namespace: rec.Namespace, Name: rec.Name}]
	return ok
}
