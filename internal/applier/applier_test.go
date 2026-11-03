package applier

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/settings"
)

type nopRecorder struct{}

func (nopRecorder) Eventf(runtime.Object, runtime.Object, string, string, string, string, ...interface{}) {
}

// fixture is a kind environment in one fake client, with a clock the test moves.
type fixture struct {
	reconciler *Reconciler
	now        time.Time
	rec        *v1alpha1.Recommendation
}

func newFixture(t *testing.T, now time.Time, policies ...client.Object) *fixture {
	t.Helper()
	return newFixtureWith(t, now, interceptor.Funcs{}, policies...)
}

// newFixtureWith lets a test break single API calls, to see what the applier does with a failure.
func newFixtureWith(t *testing.T, now time.Time, funcs interceptor.Funcs, policies ...client.Object) *fixture {
	t.Helper()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo", Labels: map[string]string{"env": "production"}}}
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "demo", Generation: 1, Labels: map[string]string{"app": "api"}},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr(int32(1)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api", "tier": "web"}},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "api",
					Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}}}}},
			},
		},
		Status: appsv1.DeploymentStatus{ObservedGeneration: 1, Replicas: 1, UpdatedReplicas: 1, AvailableReplicas: 1, ReadyReplicas: 1},
	}
	rec := &v1alpha1.Recommendation{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment-api", Namespace: "demo"},
		Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"}},
		Status: v1alpha1.RecommendationStatus{
			Phase: v1alpha1.PhaseReady, Pending: true, Hash: "abc",
			TargetLabels: map[string]string{"app": "api"},
			Settings:     map[string]v1alpha1.Setting{settings.AutoDeploy: {Value: "Enabled", Source: "cluster"}},
			Containers: []v1alpha1.ContainerRecommendation{{Name: "api",
				Current:     corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}},
				Recommended: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}},
			}},
		},
	}
	objs := append([]client.Object{ns, deploy, rec}, policies...)
	c := fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(objs...).
		WithStatusSubresource(&v1alpha1.Recommendation{}, &v1alpha1.ApplyPolicy{}, &v1alpha1.ClusterApplyPolicy{}).
		WithInterceptorFuncs(funcs).Build()
	f := &fixture{now: now, rec: rec}
	f.reconciler = &Reconciler{Client: c, APIReader: c, Recorder: nopRecorder{}, Now: func() time.Time { return f.now },
		applying: map[types.NamespacedName]time.Time{}}
	return f
}

func (f *fixture) reconcile(t *testing.T) ctrl.Result {
	t.Helper()
	res, err := f.reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "demo", Name: "deployment-api"}})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	return res
}

func (f *fixture) cpu(t *testing.T) string {
	t.Helper()
	var d appsv1.Deployment
	if err := f.reconciler.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "api"}, &d); err != nil {
		t.Fatal(err)
	}
	return d.Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().String()
}

// update changes the Recommendation the way the server does: read, mutate, write. Status has its own
// subresource, so a change there goes through Status().Update.
func (f *fixture) update(t *testing.T, spec bool, mutate func(*v1alpha1.Recommendation)) {
	t.Helper()
	var rec v1alpha1.Recommendation
	if err := f.reconciler.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "deployment-api"}, &rec); err != nil {
		t.Fatal(err)
	}
	mutate(&rec)
	var err error
	if spec {
		err = f.reconciler.Update(context.Background(), &rec)
	} else {
		err = f.reconciler.Status().Update(context.Background(), &rec)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func (f *fixture) status(t *testing.T) v1alpha1.RecommendationStatus {
	t.Helper()
	var rec v1alpha1.Recommendation
	if err := f.reconciler.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "deployment-api"}, &rec); err != nil {
		t.Fatal(err)
	}
	return rec.Status
}

func businessHours() *v1alpha1.ClusterApplyPolicy {
	return &v1alpha1.ClusterApplyPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "business-hours"},
		Spec: v1alpha1.ClusterApplyPolicySpec{ApplyRules: v1alpha1.ApplyRules{
			Windows: []v1alpha1.ApplyWindow{{From: "10:00", To: "16:00"}},
		}},
	}
}

// ptr is shared with the policy tests' helpers via a local copy: the applier package has no generic one.
func ptr[T any](v T) *T { return &v }

func TestBlockedApplyIsRequeuedAtTheWindowAndThenApplied(t *testing.T) {
	f := newFixture(t, time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC), businessHours())
	res := f.reconcile(t)

	st := f.status(t)
	if st.ApplyPolicy == nil || st.ApplyPolicy.Auto.State != v1alpha1.ApplyBlocked || st.ApplyPolicy.Auto.BlockedBy != "windows" {
		t.Fatalf("the decision should hold the apply: %+v", st.ApplyPolicy)
	}
	if st.ApplyPolicy.Manual.State != v1alpha1.ApplyAllowed {
		t.Fatalf("a policy that governs Auto leaves a person free: %+v", st.ApplyPolicy.Manual)
	}
	if got := f.cpu(t); got != "1" {
		t.Fatalf("the workload changed while held: cpu = %s", got)
	}
	// The applier looks again at the earliest rule change, and at most an hour later.
	if res.RequeueAfter != time.Hour {
		t.Fatalf("RequeueAfter = %v, want 1h", res.RequeueAfter)
	}

	// The window opens: the same recommendation is due, and the decision no longer holds it.
	f.now = time.Date(2026, 11, 3, 10, 0, 30, 0, time.UTC)
	f.reconcile(t)
	if got := f.cpu(t); got != "500m" {
		t.Fatalf("cpu = %s, want the recommendation applied", got)
	}
	st = f.status(t)
	if st.ApplyPolicy.Auto.State != v1alpha1.ApplyAllowed {
		t.Fatalf("decision = %+v, want allowed", st.ApplyPolicy.Auto)
	}
	if len(st.History) == 0 || st.History[0].By != AutoDeployUser {
		t.Fatalf("the apply should be in the history: %+v", st.History)
	}
	if !f.reconciler.isApplying(f.rec) {
		t.Fatal("the apply should count as in progress until its rollout finishes")
	}
}

func TestAuditPolicyRecordsWithoutHolding(t *testing.T) {
	p := businessHours()
	p.Spec.Mode = v1alpha1.ApplyPolicyAudit
	f := newFixture(t, time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC), p)
	f.reconcile(t)
	if got := f.cpu(t); got != "500m" {
		t.Fatalf("an audit policy must not hold: cpu = %s", got)
	}
	st := f.status(t)
	if st.ApplyPolicy == nil || len(st.ApplyPolicy.Auto.WouldBlock) == 0 {
		t.Fatalf("wouldBlock = %+v", st.ApplyPolicy)
	}
}

func TestManualApplyIsRejectedWhenThePolicyHoldsIt(t *testing.T) {
	p := businessHours()
	p.Spec.AppliesTo = []v1alpha1.ApplyKind{v1alpha1.ApplyKindAuto, v1alpha1.ApplyKindManual}
	f := newFixture(t, time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC), p)
	f.reconcile(t)

	// A person asks for it anyway: the applier decides again, on the live rules.
	f.update(t, true, func(rec *v1alpha1.Recommendation) {
		rec.Spec.ApplyRequest = &v1alpha1.ApplyRequest{ID: "r1", Action: v1alpha1.ActionApply, Hash: "abc",
			RequestedBy: "alice", RequestedAt: metav1.NewTime(f.now)}
	})
	f.reconcile(t)

	st := f.status(t)
	if st.LastApplyResult == nil || st.LastApplyResult.Result != "Rejected" {
		t.Fatalf("result = %+v, want Rejected", st.LastApplyResult)
	}
	if !strings.Contains(st.LastApplyResult.Message, "held by ClusterApplyPolicy/business-hours") {
		t.Fatalf("message = %q", st.LastApplyResult.Message)
	}
	if got := f.cpu(t); got != "1" {
		t.Fatalf("cpu = %s, want it unchanged", got)
	}
}

func TestRevertIsNeverHeld(t *testing.T) {
	frozen := &v1alpha1.ClusterApplyPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "event"},
		Spec: v1alpha1.ClusterApplyPolicySpec{ApplyRules: v1alpha1.ApplyRules{
			AppliesTo: []v1alpha1.ApplyKind{v1alpha1.ApplyKindAuto, v1alpha1.ApplyKindManual},
			Freezes: []v1alpha1.ApplyFreeze{{
				From:   metav1.NewTime(time.Date(2026, 11, 3, 8, 0, 0, 0, time.UTC)),
				To:     metav1.NewTime(time.Date(2026, 11, 3, 18, 0, 0, 0, time.UTC)),
				Reason: "sales event",
			}},
		}},
	}
	f := newFixture(t, time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), frozen)
	// The workload carries the recommendation already, so there is something to revert.
	f.update(t, false, func(rec *v1alpha1.Recommendation) {
		rec.Status.History = []v1alpha1.HistoryEntry{{
			Time: metav1.NewTime(time.Date(2026, 11, 3, 7, 0, 0, 0, time.UTC)), Action: v1alpha1.ActionApply, By: "alice", Hash: "abc",
			Containers: []v1alpha1.ContainerChange{{Name: "api",
				From: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}},
				To:   corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}}}},
		}}
	})
	var d appsv1.Deployment
	if err := f.reconciler.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "api"}, &d); err != nil {
		t.Fatal(err)
	}
	d.Spec.Template.Spec.Containers[0].Resources.Requests[corev1.ResourceCPU] = resource.MustParse("500m")
	if err := f.reconciler.Update(context.Background(), &d); err != nil {
		t.Fatal(err)
	}
	f.reconcile(t) // the freeze is in force and holds every apply

	f.update(t, true, func(rec *v1alpha1.Recommendation) {
		rec.Spec.ApplyRequest = &v1alpha1.ApplyRequest{ID: "r2", Action: v1alpha1.ActionRevert,
			RequestedBy: "alice", RequestedAt: metav1.NewTime(f.now)}
	})
	f.reconcile(t)
	if got := f.cpu(t); got != "1" {
		t.Fatalf("cpu = %s, want the revert to have restored the previous request during the freeze", got)
	}
}

func TestRequeueAfterIsQuietUnlessAnApplyWaits(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	until := metav1.NewTime(now.Add(10 * time.Minute))
	window := policy.Result{NextChange: until.Time,
		Auto: v1alpha1.ApplyDecision{State: v1alpha1.ApplyBlocked, BlockedBy: "windows", NextAllowed: &until}}
	slot := policy.Result{Auto: v1alpha1.ApplyDecision{State: v1alpha1.ApplyBlocked, BlockedBy: "rate.maxConcurrent"}}
	allowed := policy.Result{Auto: v1alpha1.ApplyDecision{State: v1alpha1.ApplyAllowed},
		Manual: v1alpha1.ApplyDecision{State: v1alpha1.ApplyAllowed}, NextChange: now.Add(time.Hour)}

	// Nothing to apply: wait for the next reconcile instead of polling.
	if got := requeueAfter(window, now, false); got != 0 {
		t.Errorf("idle requeue = %v, want 0", got)
	}
	// An apply waits for the window, and the applier looks again when it opens.
	if got := requeueAfter(window, now, true); got != 10*time.Minute {
		t.Errorf("window requeue = %v, want 10m", got)
	}
	// A slot frees when another rollout finishes, not at a known instant: look sooner.
	if got := requeueAfter(slot, now, true); got != 30*time.Second {
		t.Errorf("slot requeue = %v, want 30s", got)
	}
	// Nothing holds it: it is applied now.
	if got := requeueAfter(allowed, now, true); got != 0 {
		t.Errorf("allowed requeue = %v, want 0", got)
	}
	// A window far away is looked at again at most an hour later.
	far := metav1.NewTime(now.Add(48 * time.Hour))
	if got := requeueAfter(policy.Result{NextChange: far.Time, Auto: v1alpha1.ApplyDecision{
		State: v1alpha1.ApplyBlocked, BlockedBy: "windows", NextAllowed: &far}}, now, true); got != time.Hour {
		t.Errorf("far requeue = %v, want 1h", got)
	}
}

// An apply that fails must be tried again: the applier remembers a recommendation only once it worked, or a
// conflict with the controller writing the Recommendation would skip that hash until the applier restarts.
func TestAFailedAutoApplyIsRetried(t *testing.T) {
	fail := true
	funcs := interceptor.Funcs{Patch: func(ctx context.Context, c client.WithWatch, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
		if _, isDeployment := obj.(*appsv1.Deployment); isDeployment && fail {
			fail = false
			return fmt.Errorf("the server said no")
		}
		return c.Patch(ctx, obj, patch, opts...)
	}}
	f := newFixtureWith(t, time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC), funcs)

	res, err := f.reconciler.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Namespace: "demo", Name: "deployment-api"}})
	if err != nil {
		t.Fatalf("a failed apply is recorded, not returned: %v", err)
	}
	_ = res
	if got := f.cpu(t); got != "1" {
		t.Fatalf("cpu = %s, want the workload untouched after the failure", got)
	}
	if r := f.status(t).LastApplyResult; r == nil || r.Result != "Failed" {
		t.Fatalf("the failure should be recorded: %+v", r)
	}

	f.reconcile(t)
	if got := f.cpu(t); got != "500m" {
		t.Fatalf("cpu = %s, want the recommendation applied on the next attempt", got)
	}
	if st := f.status(t); len(st.History) == 0 || st.History[0].By != AutoDeployUser {
		t.Fatalf("the second attempt should be in the history: %+v", st.History)
	}
}

// A Recommendation deleted and written again is a new object with an empty history, so the applier may apply the
// same hash again: the resource that opt-out deletes would otherwise never be applied while the applier lives.
func TestARecreatedRecommendationIsAppliedAgain(t *testing.T) {
	f := newFixture(t, time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC))
	f.reconcile(t)
	if got := f.cpu(t); got != "500m" {
		t.Fatalf("cpu = %s, want the recommendation applied", got)
	}

	// The same values, a new object: what the controller writes after a user opts out and back in.
	var rec v1alpha1.Recommendation
	if err := f.reconciler.Get(context.Background(), types.NamespacedName{Namespace: "demo", Name: "deployment-api"}, &rec); err != nil {
		t.Fatal(err)
	}
	if err := f.reconciler.Delete(context.Background(), &rec); err != nil {
		t.Fatal(err)
	}
	fresh := &v1alpha1.Recommendation{ObjectMeta: metav1.ObjectMeta{Name: "deployment-api", Namespace: "demo"},
		Spec: rec.Spec, Status: *rec.Status.DeepCopy()}
	fresh.UID = "e2e-second-life"
	fresh.Status.History = nil
	fresh.Status.ApplyPolicy = nil
	fresh.Status.Containers[0].Current = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}}
	if err := f.reconciler.Create(context.Background(), fresh); err != nil {
		t.Fatal(err)
	}

	f.reconcile(t)
	st := f.status(t)
	if len(st.History) == 0 || st.History[0].By != AutoDeployUser {
		t.Fatalf("the recreated Recommendation should be applied: %+v", st.History)
	}
}

// An installation that upgrades before the chart's CRDs are applied must not watch a kind that does not exist:
// the informer never syncs and the applier stops reconciling while looking healthy.
func TestPoliciesInstalled(t *testing.T) {
	mapper := meta.NewDefaultRESTMapper([]schema.GroupVersion{v1alpha1.GroupVersion})
	if policiesInstalled(mapper) {
		t.Fatal("neither policy kind is mapped, so neither can be watched")
	}
	mapper.Add(v1alpha1.GroupVersion.WithKind("ClusterApplyPolicy"), meta.RESTScopeRoot)
	if policiesInstalled(mapper) {
		t.Fatal("a mapping for one kind is not both of them")
	}
	mapper.Add(v1alpha1.GroupVersion.WithKind("ApplyPolicy"), meta.RESTScopeNamespace)
	if !policiesInstalled(mapper) {
		t.Fatal("both kinds are mapped")
	}
}
