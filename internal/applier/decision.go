package applier

import (
	"context"
	"errors"
	"slices"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// recordPolicy writes the decision when it changes. The status is small and the recommender patches it
// too, so writing it on every pass would only cause conflicts.
func (r *Reconciler) recordPolicy(ctx context.Context, rec *v1alpha1.Recommendation, res policy.Result) error {
	if samePolicy(rec.Status.ApplyPolicy, res) {
		return nil
	}
	now := metav1.NewTime(r.Now())
	patch := client.MergeFrom(rec.DeepCopy())
	rec.Status.ApplyPolicy = &v1alpha1.ApplyPolicyStatus{Auto: res.Auto, Manual: res.Manual, EvaluatedAt: &now}
	if err := r.Status().Patch(ctx, rec, patch); err != nil {
		return err
	}
	r.verifyPolicyRecorded(ctx, rec)
	return nil
}

// verifyPolicyRecorded checks once per process that the API server keeps the decision written above. An
// installation whose Recommendation CRD predates this binary prunes unknown status fields without an error: the
// apply stays held, the status says nothing about why, and the server answers the click with a plain apply.
func (r *Reconciler) verifyPolicyRecorded(ctx context.Context, rec *v1alpha1.Recommendation) {
	if r.verifiedPolicy.Load() {
		return
	}
	var written v1alpha1.Recommendation
	if err := r.APIReader.Get(ctx, client.ObjectKeyFromObject(rec), &written); err != nil {
		// The next written decision tries again: a read that failed says nothing about the CRD.
		return
	}
	r.verifiedPolicy.Store(true)
	if written.Status.ApplyPolicy != nil {
		setCRDCurrent("recommendations", true)
		return
	}
	setCRDCurrent("recommendations", false)
	log.FromContext(ctx).Error(errors.New("status.applyPolicy did not survive the write"),
		"the Recommendation CRD is older than this binary, so apply decisions never reach the status: apply charts/resourcerer/crds",
		"crd", "recommendations")
}

// policyRefused is the decision when the rules could not be read: everything is held, and the reason says
// why. A policy resourcerer cannot read never opens the door.
func policyRefused(err error) policy.Result {
	d := v1alpha1.ApplyDecision{State: v1alpha1.ApplyBlocked, BlockedBy: "policy", Reason: err.Error()}
	return policy.Result{Auto: d, Manual: d}
}

func samePolicy(cur *v1alpha1.ApplyPolicyStatus, res policy.Result) bool {
	if cur == nil {
		return false
	}
	return sameDecision(cur.Auto, res.Auto) && sameDecision(cur.Manual, res.Manual)
}

func sameDecision(a, b v1alpha1.ApplyDecision) bool {
	if a.State != b.State || a.Policy != b.Policy || a.BlockedBy != b.BlockedBy || a.Reason != b.Reason {
		return false
	}
	if (a.NextAllowed == nil) != (b.NextAllowed == nil) {
		return false
	}
	if a.NextAllowed != nil && !a.NextAllowed.Time.Equal(b.NextAllowed.Time) {
		return false
	}
	return slices.Equal(a.WouldBlock, b.WouldBlock)
}

// requeueAfter is when to look again: at the earliest rule change, capped at an hour, and sooner while the
// only thing in the way is a slot another rollout will free. A workload with nothing to apply waits for its
// next reconcile instead, so an idle cluster does not poll.
func requeueAfter(res policy.Result, now time.Time, due bool) time.Duration {
	// An apply that nothing holds is handled now; one that a rule holds is looked at again at the earliest
	// instant the rules change.
	if !due || !held(res) {
		return 0
	}
	wait := time.Hour
	if !res.NextChange.IsZero() {
		wait = min(wait, res.NextChange.Sub(now))
	}
	if waitingForSlot(res) {
		wait = min(wait, 30*time.Second)
	}
	return max(wait, time.Second)
}

func held(res policy.Result) bool {
	return res.Auto.State == v1alpha1.ApplyBlocked || res.Manual.State == v1alpha1.ApplyBlocked
}

func waitingForSlot(res policy.Result) bool {
	for _, d := range []v1alpha1.ApplyDecision{res.Auto, res.Manual} {
		switch d.BlockedBy {
		case "rate.maxConcurrent", "rate.maxPerHour", "rate.perNamespace", "rollout":
			return true
		}
	}
	return false
}

// heldReason is what a person sees when a policy holds the apply they asked for.
func heldReason(d v1alpha1.ApplyDecision) string {
	out := "held by " + d.Policy + ": " + d.Reason
	if d.NextAllowed != nil {
		out += " (next allowed at " + d.NextAllowed.UTC().Format(time.RFC3339) + ")"
	}
	return out
}

/* ── Applies in progress ─────────────────────────────────────────────────── */

func (r *Reconciler) markApplying(rec *v1alpha1.Recommendation) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.applying == nil {
		r.applying = map[types.NamespacedName]time.Time{}
	}
	r.applying[client.ObjectKeyFromObject(rec)] = r.Now()
}

// releaseApplying forgets an apply once its rollout has finished: that is when its slot frees.
func (r *Reconciler) releaseApplying(ctx context.Context, rec *v1alpha1.Recommendation) {
	if !r.isApplying(rec) {
		return
	}
	w, err := workload.Get(ctx, r.Client, rec.Spec.TargetRef.Kind,
		client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name})
	if err == nil && w.InProgress() {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.applying, client.ObjectKeyFromObject(rec))
}
