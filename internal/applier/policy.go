package applier

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// maxRequeue caps how far ahead the applier asks to be woken for a policy that blocks: recomputing once
// an hour keeps the decision fresh when a policy changes, and covers a long freeze without depending on
// a single long timer.
const maxRequeue = time.Hour

// policyRetry is how soon to try again after the policies could not be read: short enough that fixing a
// typo unblocks applies without waiting for the next regeneration of every recommendation.
const policyRetry = time.Minute

// decideAt is when an apply may start, for both kinds of apply. The rules are read once and evaluated
// twice: a policy may govern one and not the other.
type decide struct {
	auto   policy.Decision
	manual policy.Decision
}

func (r *Reconciler) decideFor(ctx context.Context, rec *v1alpha1.Recommendation) (decide, error) {
	rules, err := r.rulesFor(ctx, rec)
	if err != nil {
		return decide{}, err
	}
	in := policy.Input{
		Now:        r.Now(),
		Labels:     r.workloadLabels(ctx, rec),
		LastChange: lastChange(rec),
		Rules:      rules,
	}
	auto, err := policy.Evaluate(withAction(in, v1alpha1.AppliesToAuto))
	if err != nil {
		return decide{}, err
	}
	manual, err := policy.Evaluate(withAction(in, v1alpha1.AppliesToManual))
	if err != nil {
		return decide{}, err
	}
	return decide{auto: auto, manual: manual}, nil
}

// failClosed is what a policy read failure means: no apply of either kind until the policies can be read
// again, and a reason that names the policy that could not be read.
func failClosed(err error) decide {
	reason := "the apply policies could not be read: " + err.Error()
	return decide{
		auto:   policy.Decision{Reason: reason},
		manual: policy.Decision{Reason: reason},
	}
}

func withAction(in policy.Input, a v1alpha1.AppliesTo) policy.Input {
	in.Action = a
	return in
}

// rulesFor lists the policies that govern the workload: cluster policies first, then namespaced ones,
// each ordered by priority and name so the same input always reports the same policy.
func (r *Reconciler) rulesFor(ctx context.Context, rec *v1alpha1.Recommendation) ([]policy.Rules, error) {
	if !r.policies {
		return nil, nil
	}
	// Read from the API server, not the cache: right after a policy is created its effect must already
	// be there, and a cache that has not caught up would let an apply through a freeze that exists.
	var clusters v1alpha1.ClusterApplyPolicyList
	if err := r.APIReader.List(ctx, &clusters); err != nil {
		return nil, fmt.Errorf("list cluster apply policies: %w", err)
	}
	nsLabels := r.namespaceLabels(ctx, rec.Namespace)
	rules := make([]policy.Rules, 0, len(clusters.Items))
	for _, c := range clusters.Items {
		if !policy.Selects(c.Spec.NamespaceSelector, nsLabels) {
			continue
		}
		rules = append(rules, policy.Rules{Name: "ClusterApplyPolicy/" + c.Name, Spec: c.Spec.ApplyRules})
	}
	sortRules(rules)

	var namespaced v1alpha1.ApplyPolicyList
	if err := r.APIReader.List(ctx, &namespaced, client.InNamespace(rec.Namespace)); err != nil {
		return nil, fmt.Errorf("list apply policies in %s: %w", rec.Namespace, err)
	}
	own := make([]policy.Rules, 0, len(namespaced.Items))
	for _, p := range namespaced.Items {
		own = append(own, policy.Rules{Name: "ApplyPolicy/" + p.Namespace + "/" + p.Name, Spec: p.Spec.ApplyRules})
	}
	sortRules(own)
	return append(rules, own...), nil
}

func sortRules(rules []policy.Rules) {
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Spec.Priority != rules[j].Spec.Priority {
			return rules[i].Spec.Priority > rules[j].Spec.Priority
		}
		return rules[i].Name < rules[j].Name
	})
}

func (r *Reconciler) namespaceLabels(ctx context.Context, name string) map[string]string {
	var ns corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: name}, &ns); err != nil {
		return nil
	}
	return ns.Labels
}

// workloadLabels are what a workloadSelector matches: the workload's own labels and its pod template's,
// since a manifest often carries the labels selectors are written against only in the template.
func (r *Reconciler) workloadLabels(ctx context.Context, rec *v1alpha1.Recommendation) map[string]string {
	t := rec.Spec.TargetRef
	w, err := workload.Get(ctx, r.Client, t.Kind, types.NamespacedName{Namespace: rec.Namespace, Name: t.Name})
	if err != nil {
		return nil
	}
	out := map[string]string{}
	if w.Template != nil {
		for k, v := range w.Template.Labels {
			out[k] = v
		}
	}
	for k, v := range w.Object.GetLabels() {
		out[k] = v
	}
	return out
}

// lastChange is when resourcerer last changed the workload, for the cooldown; history is newest first
// and holds reverts too, since a revert is also a change its effects take time to show.
func lastChange(rec *v1alpha1.Recommendation) time.Time {
	if len(rec.Status.History) == 0 {
		return time.Time{}
	}
	return rec.Status.History[0].Time.Time
}

// recordDecision writes the decision on the Recommendation, so the UI and the API show the reason the
// applier used instead of guessing it. Both kinds of apply are recorded: a policy that holds auto-deploy
// back while a person may still apply is a real state, and one field could not say it.
func (r *Reconciler) recordDecision(ctx context.Context, rec *v1alpha1.Recommendation, d decide) error {
	next := &v1alpha1.ApplyPolicyStatus{Auto: decision(d.auto), Manual: decision(d.manual)}
	if sameDecision(rec.Status.ApplyPolicy, next) {
		return nil
	}
	now := metav1.NewTime(r.Now())
	next.EvaluatedAt = &now
	patch := client.MergeFrom(rec.DeepCopy())
	rec.Status.ApplyPolicy = next
	return r.Status().Patch(ctx, rec, patch)
}

func decision(d policy.Decision) v1alpha1.ApplyDecision {
	out := v1alpha1.ApplyDecision{State: v1alpha1.PolicyStateAllowed, WouldBlock: d.WouldBlock}
	if !d.Allowed {
		out.State = v1alpha1.PolicyStateBlocked
		out.BlockedBy = d.BlockedBy
		out.Policy = d.Policy
		out.Reason = d.Reason
	}
	if !d.NextAllowed.IsZero() {
		t := metav1.NewTime(d.NextAllowed)
		out.NextAllowed = &t
	}
	return out
}

// sameDecision compares everything but EvaluatedAt, which changes on every evaluation and would make
// the applier write the status of every workload on every pass.
func sameDecision(a, b *v1alpha1.ApplyPolicyStatus) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	x, y := *a, *b
	x.EvaluatedAt, y.EvaluatedAt = nil, nil
	return reflect.DeepEqual(x, y)
}

// requeueFor is how long to wait before asking again, when a policy blocks an automatic apply that is
// otherwise due. Anything else is left to the watch on the Recommendation.
func (r *Reconciler) requeueFor(rec *v1alpha1.Recommendation, d policy.Decision) time.Duration {
	if d.Allowed || d.NextAllowed.IsZero() || !r.autoCandidate(rec) {
		return 0
	}
	wait := d.NextAllowed.Sub(r.Now())
	if wait <= 0 {
		return time.Second
	}
	if wait > maxRequeue {
		return maxRequeue
	}
	return wait
}

// autoCandidate reports whether auto-deploy would apply this recommendation if no policy blocked it,
// so the applier does not reschedule for policies that govern nothing that is about to happen.
func (r *Reconciler) autoCandidate(rec *v1alpha1.Recommendation) bool {
	st := rec.Status
	return st.Settings[settings.AutoDeploy].Value == "Enabled" && st.Phase == v1alpha1.PhaseReady && st.Pending && st.Hash != ""
}
