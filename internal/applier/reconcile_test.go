package applier

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/policy"
	"github.com/gjbravi/resourcerer/internal/settings"
)

var testNow = time.Date(2026, 11, 2, 15, 0, 0, 0, time.UTC)

func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	if err := v1alpha1.AddToScheme(s); err != nil {
		t.Fatal(err)
	}
	return s
}

// fixture is a Ready, pending recommendation for demo/worker, the Deployment it targets and its namespace.
type fixture struct {
	rec        *v1alpha1.Recommendation
	objects    []client.Object
	autoDeploy string
}

func newFixture() *fixture {
	rec := &v1alpha1.Recommendation{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment-worker", Namespace: "demo"},
		Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "worker"}},
		Status: v1alpha1.RecommendationStatus{
			Phase:   v1alpha1.PhaseReady,
			Pending: true,
			Hash:    "h2",
			Containers: []v1alpha1.ContainerRecommendation{{Name: "app", Recommended: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")},
			}}},
		},
	}
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: "demo", Labels: map[string]string{"app": "worker"}},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"tier": "batch"}},
			Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "app", Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")},
			}}}},
		}},
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "demo", Labels: map[string]string{"env": "production"}}}
	return &fixture{rec: rec, objects: []client.Object{dep, ns}, autoDeploy: "Enabled"}
}

func (f *fixture) with(objs ...client.Object) *fixture {
	f.objects = append(f.objects, objs...)
	return f
}

func (f *fixture) request(action v1alpha1.Action) *fixture {
	f.rec.Spec.ApplyRequest = &v1alpha1.ApplyRequest{ID: "r1", Action: action, Hash: f.rec.Status.Hash, RequestedBy: "someone", RequestedAt: metav1.NewTime(testNow)}
	return f
}

// applied is a previous apply of h1, an hour ago: something to revert, and a change the cooldown counts from.
func (f *fixture) applied() *fixture {
	f.rec.Status.History = []v1alpha1.HistoryEntry{{
		Time: metav1.NewTime(testNow.Add(-time.Hour)), Action: v1alpha1.ActionApply, By: AutoDeployUser, Hash: "h1",
		Containers: []v1alpha1.ContainerChange{{Name: "app",
			From: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}},
			To:   corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("500m")}},
		}},
	}}
	return f
}

// patches counts the patches resourcerer sends to workloads; the fake client does not run server-side
// apply, and what matters here is whether one was sent.
type patches struct{ workload int }

func (f *fixture) client(t *testing.T, p *patches, funcs interceptor.Funcs) client.WithWatch {
	t.Helper()
	f.rec.Status.Settings = map[string]v1alpha1.Setting{settings.AutoDeploy: {Value: f.autoDeploy, Source: "namespace"}}
	patch := funcs.Patch
	funcs.Patch = func(ctx context.Context, c client.WithWatch, obj client.Object, pt client.Patch, opts ...client.PatchOption) error {
		if _, ok := obj.(*appsv1.Deployment); ok {
			p.workload++
			return nil
		}
		if patch != nil {
			return patch(ctx, c, obj, pt, opts...)
		}
		return c.Patch(ctx, obj, pt, opts...)
	}
	return fake.NewClientBuilder().
		WithScheme(testScheme(t)).
		WithObjects(append([]client.Object{f.rec.DeepCopy()}, f.objects...)...).
		WithStatusSubresource(&v1alpha1.Recommendation{}).
		WithInterceptorFuncs(funcs).
		Build()
}

func newReconciler(c client.Client, api client.Reader) *Reconciler {
	return &Reconciler{Client: c, APIReader: api, Recorder: events.NewFakeRecorder(10), Now: func() time.Time { return testNow }, policies: true}
}

func reconcileOnce(t *testing.T, r *Reconciler) (ctrl.Result, *v1alpha1.Recommendation) {
	t.Helper()
	key := types.NamespacedName{Namespace: "demo", Name: "deployment-worker"}
	res, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: key})
	if err != nil {
		t.Fatalf("Reconcile: %v", err)
	}
	var rec v1alpha1.Recommendation
	if err := r.Get(context.Background(), key, &rec); err != nil {
		t.Fatal(err)
	}
	return res, &rec
}

func applyPolicy(name string, spec v1alpha1.ApplyRules) *v1alpha1.ApplyPolicy {
	return &v1alpha1.ApplyPolicy{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "demo"}, Spec: v1alpha1.ApplyPolicySpec{ApplyRules: spec}}
}

func freezeRules(until time.Time, kinds ...v1alpha1.AppliesTo) v1alpha1.ApplyRules {
	return v1alpha1.ApplyRules{AppliesTo: kinds, When: &v1alpha1.ApplyWhen{Freezes: []v1alpha1.ApplyFreeze{{
		From: metav1.NewTime(testNow.Add(-time.Hour)), To: metav1.NewTime(until), Reason: "launch",
	}}}}
}

func TestReconcileAppliesWhenNothingHolds(t *testing.T) {
	var p patches
	c := newFixture().client(t, &p, interceptor.Funcs{})
	_, rec := reconcileOnce(t, newReconciler(c, c))
	if p.workload != 1 {
		t.Fatalf("patched the workload %d times, want 1", p.workload)
	}
	if len(rec.Status.History) != 1 || rec.Status.History[0].By != AutoDeployUser {
		t.Errorf("history = %+v, want the automatic apply", rec.Status.History)
	}
	if d := rec.Status.ApplyPolicy; d == nil || d.Auto.State != v1alpha1.PolicyStateAllowed || d.Manual.State != v1alpha1.PolicyStateAllowed {
		t.Errorf("applyPolicy = %+v, want both allowed", d)
	}
}

func TestReconcileHoldsAutoDeployDuringAFreeze(t *testing.T) {
	var p patches
	c := newFixture().with(applyPolicy("freeze", freezeRules(testNow.Add(30*time.Minute)))).client(t, &p, interceptor.Funcs{})
	res, rec := reconcileOnce(t, newReconciler(c, c))
	if p.workload != 0 {
		t.Fatal("applied during a freeze")
	}
	d := rec.Status.ApplyPolicy
	if d == nil || d.Auto.State != v1alpha1.PolicyStateBlocked || d.Auto.BlockedBy != "when.freezes" || d.Auto.Policy != "ApplyPolicy/demo/freeze" {
		t.Fatalf("applyPolicy = %+v, want auto blocked by the freeze", d)
	}
	if d.Manual.State != v1alpha1.PolicyStateAllowed {
		t.Errorf("manual = %+v, want allowed: the policy lists no Manual", d.Manual)
	}
	if res.RequeueAfter != 30*time.Minute {
		t.Errorf("RequeueAfter = %s, want the end of the freeze", res.RequeueAfter)
	}
}

func TestReconcileRejectsAManualApplyAPolicyHolds(t *testing.T) {
	var p patches
	f := newFixture().request(v1alpha1.ActionApply).with(applyPolicy("freeze", freezeRules(testNow.Add(time.Hour), v1alpha1.AppliesToManual)))
	f.autoDeploy = "Disabled"
	c := f.client(t, &p, interceptor.Funcs{})
	_, rec := reconcileOnce(t, newReconciler(c, c))
	if p.workload != 0 {
		t.Fatal("applied a request a policy holds")
	}
	r := rec.Status.LastApplyResult
	if r == nil || r.Result != "Rejected" || !strings.Contains(r.Message, "launch") {
		t.Errorf("lastApplyResult = %+v, want a rejection that gives the freeze's reason", r)
	}
	if rec.Status.LastHandledRequestID != "r1" {
		t.Errorf("the request is not marked handled")
	}
}

// A revert undoes a bad apply, and right after that apply is when the cooldown and a freeze hold.
func TestReconcileRevertsWhatAPolicyHolds(t *testing.T) {
	all := []v1alpha1.AppliesTo{v1alpha1.AppliesToAuto, v1alpha1.AppliesToManual}
	rules := freezeRules(testNow.Add(time.Hour), all...)
	rules.Limits = &v1alpha1.ApplyLimits{Cooldown: &metav1.Duration{Duration: 24 * time.Hour}}
	for name, funcs := range map[string]interceptor.Funcs{
		"held by a freeze and the cooldown": {},
		"the policies cannot be read": {List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			if _, ok := list.(*v1alpha1.ApplyPolicyList); ok {
				return errors.New("apiserver unavailable")
			}
			return c.List(ctx, list, opts...)
		}},
	} {
		t.Run(name, func(t *testing.T) {
			var p patches
			c := newFixture().applied().request(v1alpha1.ActionRevert).with(applyPolicy("strict", rules)).client(t, &p, funcs)
			_, rec := reconcileOnce(t, newReconciler(c, c))
			if p.workload != 1 {
				t.Fatalf("patched the workload %d times, want the revert", p.workload)
			}
			if h := rec.Status.History; len(h) != 2 || h[0].Action != v1alpha1.ActionRevert {
				t.Errorf("history = %+v, want the revert on top", h)
			}
		})
	}
}

// The cache may not have seen a policy created a moment ago; the apply about to happen must.
func TestReconcileReadsThePoliciesAgainBeforeApplying(t *testing.T) {
	var p, q patches
	f := newFixture()
	cached := f.client(t, &p, interceptor.Funcs{})
	live := newFixture().with(applyPolicy("just-created", freezeRules(testNow.Add(time.Hour)))).client(t, &q, interceptor.Funcs{})
	_, rec := reconcileOnce(t, newReconciler(cached, live))
	if p.workload != 0 {
		t.Fatal("applied through a freeze the cache had not seen")
	}
	if d := rec.Status.ApplyPolicy; d == nil || d.Auto.Policy != "ApplyPolicy/demo/just-created" {
		t.Errorf("applyPolicy = %+v, want the freeze from the API server", d)
	}
}

func TestReconcileDoesNotReadUncachedWhenNothingIsDue(t *testing.T) {
	var p patches
	f := newFixture()
	f.autoDeploy = "Disabled"
	c := f.client(t, &p, interceptor.Funcs{})
	reads := 0
	api := interceptor.NewClient(fake.NewClientBuilder().WithScheme(testScheme(t)).Build(), interceptor.Funcs{
		List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			reads++
			return c.List(ctx, list, opts...)
		},
	})
	reconcileOnce(t, newReconciler(c, api))
	if reads != 0 {
		t.Errorf("listed %d times from the API server with nothing to apply, want 0", reads)
	}
}

// Anything the decision depends on that cannot be read blocks: no labels would quietly lift every policy
// with a selector, which is how a ClusterApplyPolicy for production namespaces would stop applying.
func TestReconcileFailsClosed(t *testing.T) {
	prod := &v1alpha1.ClusterApplyPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "production"},
		Spec: v1alpha1.ClusterApplyPolicySpec{
			NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"env": "production"}},
			ApplyRules:        freezeRules(testNow.Add(time.Hour)),
		},
	}
	failGet := func(match func(client.Object) bool) interceptor.Funcs {
		return interceptor.Funcs{Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
			if match(obj) {
				return errors.New("apiserver unavailable")
			}
			return c.Get(ctx, key, obj, opts...)
		}}
	}
	tests := map[string]interceptor.Funcs{
		"the cluster policies": {List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
			if _, ok := list.(*v1alpha1.ClusterApplyPolicyList); ok {
				return errors.New("apiserver unavailable")
			}
			return c.List(ctx, list, opts...)
		}},
		"the namespace": failGet(func(o client.Object) bool { _, ok := o.(*corev1.Namespace); return ok }),
		"the workload":  failGet(func(o client.Object) bool { _, ok := o.(*appsv1.Deployment); return ok }),
	}
	for name, funcs := range tests {
		t.Run(name, func(t *testing.T) {
			var p patches
			c := newFixture().with(prod).client(t, &p, funcs)
			res, rec := reconcileOnce(t, newReconciler(c, c))
			if p.workload != 0 {
				t.Fatal("applied without knowing the policies")
			}
			for kind, d := range map[string]v1alpha1.ApplyDecision{"auto": rec.Status.ApplyPolicy.Auto, "manual": rec.Status.ApplyPolicy.Manual} {
				if d.State != v1alpha1.PolicyStateBlocked || d.BlockedBy != policy.FieldPolicy || !strings.Contains(d.Reason, "apiserver unavailable") {
					t.Errorf("%s = %+v, want blocked with the read error", kind, d)
				}
			}
			if res.RequeueAfter != policyRetry {
				t.Errorf("RequeueAfter = %s, want %s", res.RequeueAfter, policyRetry)
			}
		})
	}
}

func TestReconcileWithoutThePolicyKinds(t *testing.T) {
	var p patches
	c := newFixture().client(t, &p, interceptor.Funcs{List: func(ctx context.Context, c client.WithWatch, list client.ObjectList, opts ...client.ListOption) error {
		switch list.(type) {
		case *v1alpha1.ApplyPolicyList, *v1alpha1.ClusterApplyPolicyList:
			t.Error("listed a policy kind that is not installed")
		}
		return c.List(ctx, list, opts...)
	}})
	r := newReconciler(c, c)
	r.policies = false
	_, rec := reconcileOnce(t, r)
	if p.workload != 1 {
		t.Errorf("patched the workload %d times, want the apply", p.workload)
	}
	if rec.Status.ApplyPolicy != nil {
		t.Errorf("applyPolicy = %+v, want none recorded when policies do not exist", rec.Status.ApplyPolicy)
	}
}

// The server refuses a click on what the status says, so a manual block has to be decided again when it
// ends even with auto-deploy off.
func TestReconcileRequeuesForAManualBlock(t *testing.T) {
	var p patches
	f := newFixture().with(applyPolicy("window", v1alpha1.ApplyRules{
		AppliesTo: []v1alpha1.AppliesTo{v1alpha1.AppliesToManual},
		When:      &v1alpha1.ApplyWhen{Windows: []v1alpha1.ApplyWindow{{From: "15:20", To: "16:00"}}},
	}))
	f.autoDeploy = "Disabled"
	c := f.client(t, &p, interceptor.Funcs{})
	res, rec := reconcileOnce(t, newReconciler(c, c))
	if rec.Status.ApplyPolicy.Manual.State != v1alpha1.PolicyStateBlocked {
		t.Fatalf("manual = %+v, want blocked outside the window", rec.Status.ApplyPolicy.Manual)
	}
	if res.RequeueAfter != 20*time.Minute {
		t.Errorf("RequeueAfter = %s, want when the window opens", res.RequeueAfter)
	}
}

func TestRequeueFor(t *testing.T) {
	blocked := func(d time.Duration) policy.Decision { return policy.Decision{NextAllowed: testNow.Add(d)} }
	allowed := policy.Decision{Allowed: true}
	tests := []struct {
		name string
		d    decide
		want time.Duration
	}{
		{"nothing blocks", decide{auto: allowed, manual: allowed}, 0},
		{"the automatic apply", decide{auto: blocked(10 * time.Minute), manual: allowed}, 10 * time.Minute},
		{"the manual apply", decide{auto: allowed, manual: blocked(5 * time.Minute)}, 5 * time.Minute},
		{"both: the earlier", decide{auto: blocked(40 * time.Minute), manual: blocked(5 * time.Minute)}, 5 * time.Minute},
		{"far away: capped", decide{auto: blocked(72 * time.Hour), manual: allowed}, maxRequeue},
		{"already due", decide{auto: blocked(-time.Minute), manual: allowed}, time.Second},
		{"no end known", decide{auto: policy.Decision{}, manual: allowed}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := requeueFor(testNow, tt.d); got != tt.want {
				t.Errorf("requeueFor() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRecommendationsIn(t *testing.T) {
	recs := []client.Object{
		&v1alpha1.Recommendation{ObjectMeta: metav1.ObjectMeta{Name: "a", Namespace: "demo"}},
		&v1alpha1.Recommendation{ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "payments"}},
	}
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(recs...).Build()
	r := newReconciler(c, c)
	if got := r.recommendationsIn(true)(context.Background(), applyPolicy("p", v1alpha1.ApplyRules{})); len(got) != 1 || got[0].Namespace != "demo" {
		t.Errorf("an ApplyPolicy enqueued %v, want only its namespace", got)
	}
	if got := r.recommendationsIn(false)(context.Background(), &v1alpha1.ClusterApplyPolicy{}); len(got) != 2 {
		t.Errorf("a ClusterApplyPolicy enqueued %v, want every recommendation", got)
	}
}

func TestWorkloadLabels(t *testing.T) {
	f := newFixture()
	c := fake.NewClientBuilder().WithScheme(testScheme(t)).WithObjects(f.objects...).Build()
	got, err := workloadLabels(context.Background(), c, f.rec)
	if err != nil {
		t.Fatal(err)
	}
	if got["app"] != "worker" || got["tier"] != "batch" {
		t.Errorf("labels = %v, want the workload's and its pod template's", got)
	}
	f.rec.Spec.TargetRef.Name = "gone"
	if got, err := workloadLabels(context.Background(), c, f.rec); err != nil || got != nil {
		t.Errorf("a workload that is gone: labels %v, error %v, want none and no error", got, err)
	}
}
