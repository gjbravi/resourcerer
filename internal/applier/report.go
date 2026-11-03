package applier

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
)

// policyReporter keeps each policy's status current: how many managed workloads it governs, and why it
// cannot be read when it cannot. A policy that fails closed would otherwise only show on the
// Recommendations it holds, one reason at a time.
type policyReporter struct {
	*Reconciler
}

func (p *policyReporter) NeedLeaderElection() bool { return true }

func (p *policyReporter) Start(ctx context.Context) error {
	// Often enough that a policy created or labelled onto a namespace shows a status while the person who
	// wrote it is still looking at kubectl: the lists come from the cache, so a tick costs nothing.
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		p.sync(ctx)
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
		}
	}
}

func (p *policyReporter) sync(ctx context.Context) {
	var caps v1alpha1.ClusterApplyPolicyList
	if err := p.APIReader.List(ctx, &caps); err != nil {
		if !notInstalled(err) {
			log.FromContext(ctx).Error(err, "list cluster apply policies")
		}
		return
	}
	var aps v1alpha1.ApplyPolicyList
	if err := p.APIReader.List(ctx, &aps); err != nil {
		if !notInstalled(err) {
			log.FromContext(ctx).Error(err, "list apply policies")
		}
		return
	}
	var recs v1alpha1.RecommendationList
	if err := p.List(ctx, &recs); err != nil {
		log.FromContext(ctx).Error(err, "list recommendations")
		return
	}
	nsLabels, err := p.namespaceLabels(ctx)
	if err != nil {
		log.FromContext(ctx).Error(err, "list namespaces")
		return
	}
	fleet := make([]policy.Workload, 0, len(recs.Items))
	for i := range recs.Items {
		rec := &recs.Items[i]
		fleet = append(fleet, policy.Workload{
			Namespace:       rec.Namespace,
			Kind:            rec.Spec.TargetRef.Kind,
			Name:            rec.Spec.TargetRef.Name,
			Labels:          rec.Status.TargetLabels,
			NamespaceLabels: nsLabels[rec.Namespace],
		})
	}
	now := metav1.NewTime(p.Now())
	for i := range caps.Items {
		c := &caps.Items[i]
		p.report(ctx, c, "ClusterApplyPolicy/"+c.Name, c.Spec.NamespaceSelector, c.Spec.ApplyRules, fleet, now)
	}
	for i := range aps.Items {
		a := &aps.Items[i]
		p.report(ctx, a, "ApplyPolicy/"+a.Namespace+"/"+a.Name, nil, a.Spec.ApplyRules, fleet, now)
	}
}

func (p *policyReporter) report(ctx context.Context, obj client.Object, name string, namespaces *metav1.LabelSelector,
	rules v1alpha1.ApplyRules, fleet []policy.Workload, now metav1.Time) {
	var current *v1alpha1.ApplyPolicyHealth
	switch o := obj.(type) {
	case *v1alpha1.ClusterApplyPolicy:
		current = &o.Status
	case *v1alpha1.ApplyPolicy:
		current = &o.Status
	}
	pl := policy.Policy{Name: name, Namespace: obj.GetNamespace(), Namespaces: namespaces, Spec: rules}
	workloads := int32(0)
	for _, w := range fleet {
		if policy.Selects(pl, w) {
			workloads++
		}
	}
	next := v1alpha1.ApplyPolicyHealth{Workloads: &workloads}
	if err := policy.Validate(pl); err != nil {
		next.Conditions = []metav1.Condition{{
			Type: "Valid", Status: metav1.ConditionFalse, Reason: "InvalidPolicy", Message: err.Error(),
			LastTransitionTime: now, ObservedGeneration: obj.GetGeneration(),
		}}
	} else {
		next.Conditions = []metav1.Condition{{
			Type: "Valid", Status: metav1.ConditionTrue, Reason: "Valid", Message: "the policy is applied to the workloads it selects",
			LastTransitionTime: now, ObservedGeneration: obj.GetGeneration(),
		}}
	}
	if current != nil {
		if sameCondition(current.Conditions, next.Conditions) {
			next.Conditions[0].LastTransitionTime = current.Conditions[0].LastTransitionTime
		}
		if current.Workloads != nil && *current.Workloads == workloads && sameCondition(current.Conditions, next.Conditions) {
			return
		}
	}
	switch o := obj.(type) {
	case *v1alpha1.ClusterApplyPolicy:
		o.Status = next
	case *v1alpha1.ApplyPolicy:
		o.Status = next
	}
	// The whole status, not a patch: a count of zero is information, and a diff would leave it out.
	if err := p.Status().Update(ctx, obj); err != nil {
		log.FromContext(ctx).Error(err, "update "+name+" status")
	}
}

func sameCondition(a, b []metav1.Condition) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type || a[i].Status != b[i].Status || a[i].Reason != b[i].Reason ||
			a[i].Message != b[i].Message || a[i].ObservedGeneration != b[i].ObservedGeneration {
			return false
		}
	}
	return true
}
