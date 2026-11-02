package applier

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// maxRequeue caps how far ahead the applier asks to be woken for a policy that blocks: recomputing once
// an hour keeps the decision fresh when a policy changes, and covers a long freeze without depending on
// a single long timer.
const maxRequeue = time.Hour

// policyRetry is how soon to try again after the policies could not be read: short enough that fixing a
// typo unblocks applies without waiting for the next regeneration of every recommendation.
const policyRetry = time.Minute

const clusterPolicyPrefix = "ClusterApplyPolicy/"

// decide is when an apply may start, for both kinds of apply. The rules are read once and evaluated
// twice: a policy may govern one and not the other.
type decide struct {
	auto   policy.Decision
	manual policy.Decision
}

// decideFor reads the policies, the namespace and the workload through c. Anything it cannot read is an
// error, never an empty set: no labels would quietly lift every policy with a selector.
func (r *Reconciler) decideFor(ctx context.Context, c client.Reader, rec *v1alpha1.Recommendation) (decide, error) {
	if !r.policies {
		return decide{auto: policy.Decision{Allowed: true}, manual: policy.Decision{Allowed: true}}, nil
	}
	rules, err := rulesFor(ctx, c, rec.Namespace)
	if err != nil {
		return decide{}, err
	}
	in := policy.Input{Now: r.Now(), LastChange: lastChange(rec), Rules: rules}
	if len(rules) > 0 {
		if in.NamespaceLabels, err = namespaceLabels(ctx, c, rec.Namespace); err != nil {
			return decide{}, err
		}
		if in.Labels, err = workloadLabels(ctx, c, rec); err != nil {
			return decide{}, err
		}
	}
	return decide{auto: policy.Evaluate(withAction(in, v1alpha1.AppliesToAuto)), manual: policy.Evaluate(withAction(in, v1alpha1.AppliesToManual))}, nil
}

// failClosed is what a policy read failure means: no apply of either kind until the policies can be read
// again, and a reason that says what could not be read.
func failClosed(err error) decide {
	d := policy.Decision{BlockedBy: policy.FieldPolicy, Reason: "the apply policies could not be read: " + err.Error()}
	return decide{auto: d, manual: d}
}

func withAction(in policy.Input, a v1alpha1.AppliesTo) policy.Input {
	in.Action = a
	return in
}

// rulesFor lists the policies that may govern a workload in ns: every cluster policy, whose
// namespaceSelector Evaluate matches, and the namespace's own, ordered by sortRules.
func rulesFor(ctx context.Context, c client.Reader, ns string) ([]policy.Rules, error) {
	var clusters v1alpha1.ClusterApplyPolicyList
	if err := c.List(ctx, &clusters); err != nil {
		return nil, fmt.Errorf("list cluster apply policies: %w", err)
	}
	var namespaced v1alpha1.ApplyPolicyList
	if err := c.List(ctx, &namespaced, client.InNamespace(ns)); err != nil {
		return nil, fmt.Errorf("list apply policies in %s: %w", ns, err)
	}
	rules := make([]policy.Rules, 0, len(clusters.Items)+len(namespaced.Items))
	for _, p := range clusters.Items {
		rules = append(rules, policy.Rules{Name: clusterPolicyPrefix + p.Name, NamespaceSelector: p.Spec.NamespaceSelector, Spec: p.Spec.ApplyRules})
	}
	for _, p := range namespaced.Items {
		rules = append(rules, policy.Rules{Name: "ApplyPolicy/" + p.Namespace + "/" + p.Name, Spec: p.Spec.ApplyRules})
	}
	sortRules(rules)
	return rules, nil
}

// sortRules orders the policies the way the decision reports them: highest priority first, then cluster
// policies before namespaced ones, then by name, so the same input always names the same policy.
func sortRules(rules []policy.Rules) {
	sort.SliceStable(rules, func(i, j int) bool {
		a, b := rules[i], rules[j]
		if a.Spec.Priority != b.Spec.Priority {
			return a.Spec.Priority > b.Spec.Priority
		}
		ac, bc := strings.HasPrefix(a.Name, clusterPolicyPrefix), strings.HasPrefix(b.Name, clusterPolicyPrefix)
		if ac != bc {
			return ac
		}
		return a.Name < b.Name
	})
}

func namespaceLabels(ctx context.Context, c client.Reader, name string) (map[string]string, error) {
	var ns corev1.Namespace
	if err := c.Get(ctx, types.NamespacedName{Name: name}, &ns); err != nil {
		return nil, fmt.Errorf("read namespace %s for the apply policies: %w", name, err)
	}
	return ns.Labels, nil
}

// workloadLabels are what a workloadSelector matches: the workload's own labels and its pod template's,
// since a manifest often carries the labels selectors are written against only in the template. A
// workload that no longer exists has none: there is nothing to apply to, and the apply fails on its own.
func workloadLabels(ctx context.Context, c client.Reader, rec *v1alpha1.Recommendation) (map[string]string, error) {
	t := rec.Spec.TargetRef
	w, err := workload.Get(ctx, c, t.Kind, types.NamespacedName{Namespace: rec.Namespace, Name: t.Name})
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s %s for the apply policies: %w", t.Kind, t.Name, err)
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
	return out, nil
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
	if !r.policies {
		return nil
	}
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

// requeueFor is how long to wait before deciding again: until the earliest block ends, of either kind.
// The manual decision counts too, since the server refuses a click on what the status records, so a
// window opening has to reach the status even when nothing would be applied automatically.
func requeueFor(now time.Time, d decide) time.Duration {
	var wait time.Duration
	for _, x := range []policy.Decision{d.auto, d.manual} {
		if x.Allowed || x.NextAllowed.IsZero() {
			continue
		}
		w := x.NextAllowed.Sub(now)
		if w <= 0 {
			w = time.Second
		}
		if wait == 0 || w < wait {
			wait = w
		}
	}
	return min(wait, maxRequeue)
}
