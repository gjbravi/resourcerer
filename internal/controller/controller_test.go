package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/scheme"
)

var key = types.NamespacedName{Namespace: "shop", Name: "api"}
var vpaKey = types.NamespacedName{Namespace: "shop", Name: "deployment-api"}

func setup(t *testing.T, sel config.Selection, nsLabels, wlLabels map[string]string) (*Reconciler, client.Client) {
	t.Helper()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "shop", Labels: nsLabels}}
	d := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "shop", UID: "u1", Labels: wlLabels}}
	c := fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(ns, d).Build()
	return &Reconciler{
		Client:   c,
		Kind:     "Deployment",
		Options:  config.Options{Selection: sel, DefaultMode: config.ModeOff},
		Recorder: events.NewFakeRecorder(10),
	}, c
}

func run(t *testing.T, r *Reconciler) {
	t.Helper()
	if _, err := r.Reconcile(context.Background(), ctrl.Request{NamespacedName: key}); err != nil {
		t.Fatal(err)
	}
}

func hasVPA(c client.Client) bool {
	return c.Get(context.Background(), vpaKey, &vpav1.VerticalPodAutoscaler{}) == nil
}

func TestSelectionAll(t *testing.T) {
	r, c := setup(t, config.SelectionAll, nil, nil)
	run(t, r)
	if !hasVPA(c) {
		t.Fatal("expected VPA")
	}

	var d appsv1.Deployment
	_ = c.Get(context.Background(), key, &d)
	d.Labels = map[string]string{config.LabelEnabled: "false"}
	_ = c.Update(context.Background(), &d)
	run(t, r)
	if hasVPA(c) {
		t.Error("VPA should be removed after opt-out")
	}
}

func TestSelectionLabel(t *testing.T) {
	r, c := setup(t, config.SelectionLabel, nil, nil)
	run(t, r)
	if hasVPA(c) {
		t.Fatal("no VPA expected without label")
	}

	r, c = setup(t, config.SelectionLabel, map[string]string{config.LabelEnabled: "true"}, nil)
	run(t, r)
	if !hasVPA(c) {
		t.Fatal("namespace label should opt in")
	}

	r, c = setup(t, config.SelectionLabel, nil, map[string]string{config.LabelEnabled: "true"})
	run(t, r)
	if !hasVPA(c) {
		t.Fatal("workload label should opt in")
	}
}

func TestModeAnnotation(t *testing.T) {
	r, c := setup(t, config.SelectionAll, nil, nil)
	var d appsv1.Deployment
	_ = c.Get(context.Background(), key, &d)
	d.Annotations = map[string]string{config.AnnotationMode: "Auto"}
	_ = c.Update(context.Background(), &d)
	run(t, r)

	var v vpav1.VerticalPodAutoscaler
	_ = c.Get(context.Background(), vpaKey, &v)
	if *v.Spec.UpdatePolicy.UpdateMode != vpav1.UpdateModeInPlaceOrRecreate {
		t.Errorf("updateMode = %s", *v.Spec.UpdatePolicy.UpdateMode)
	}
}

func TestForeignVPAIsNotTouched(t *testing.T) {
	r, c := setup(t, config.SelectionAll, nil, nil)
	foreign := &vpav1.VerticalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "deployment-api", Namespace: "shop"}}
	_ = c.Create(context.Background(), foreign)
	run(t, r)
	var v vpav1.VerticalPodAutoscaler
	_ = c.Get(context.Background(), vpaKey, &v)
	if v.Labels[config.LabelManaged] == "true" || v.Spec.UpdatePolicy != nil {
		t.Errorf("foreign VPA was modified: %+v", v)
	}
}
