package vpa

import (
	"context"
	"errors"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/workload"
)

func deployment() *workload.Workload {
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "shop", UID: "uid-1"},
		Spec:       appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "api"}}}}},
	}
	w, _ := workload.FromObject(d)
	return w
}

func newClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(objs...).Build()
}

func get(t *testing.T, c client.Client) *vpav1.VerticalPodAutoscaler {
	t.Helper()
	v := &vpav1.VerticalPodAutoscaler{}
	if err := c.Get(context.Background(), types.NamespacedName{Namespace: "shop", Name: "deployment-api"}, v); err != nil {
		t.Fatal(err)
	}
	return v
}

func TestEnsureCreatesAndSwitchesMode(t *testing.T) {
	c := newClient()
	ctx := context.Background()
	if _, err := Ensure(ctx, c, deployment(), config.ModeOff); err != nil {
		t.Fatal(err)
	}
	v := get(t, c)
	if v.Labels[config.LabelManaged] != "true" || v.OwnerReferences[0].UID != "uid-1" || ModeOf(v) != config.ModeOff {
		t.Fatalf("created %+v", v)
	}
	if _, err := Ensure(ctx, c, deployment(), config.ModeAuto); err != nil {
		t.Fatal(err)
	}
	if got := *get(t, c).Spec.UpdatePolicy.UpdateMode; got != vpav1.UpdateModeInPlaceOrRecreate {
		t.Errorf("updateMode = %s", got)
	}
}

func TestEnsureKeepsUserFields(t *testing.T) {
	c := newClient()
	ctx := context.Background()
	_, _ = Ensure(ctx, c, deployment(), config.ModeOff)
	v := get(t, c)
	off := vpav1.ContainerScalingModeOff
	v.Spec.ResourcePolicy = &vpav1.PodResourcePolicy{ContainerPolicies: []vpav1.ContainerResourcePolicy{{ContainerName: "istio-proxy", Mode: &off}}}
	if err := c.Update(ctx, v); err != nil {
		t.Fatal(err)
	}
	_, _ = Ensure(ctx, c, deployment(), config.ModeOff)
	if get(t, c).Spec.ResourcePolicy == nil {
		t.Error("resourcePolicy set by the user was dropped")
	}
}

func TestEnsureRefusesForeignVPA(t *testing.T) {
	foreign := &vpav1.VerticalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "deployment-api", Namespace: "shop"}}
	c := newClient(foreign)
	if _, err := Ensure(context.Background(), c, deployment(), config.ModeAuto); !errors.Is(err, ErrNotManaged) {
		t.Fatalf("err = %v, want ErrNotManaged", err)
	}
	if err := Delete(context.Background(), c, "shop", "deployment-api"); err != nil {
		t.Fatal(err)
	}
	if get(t, c) == nil {
		t.Error("foreign VPA deleted")
	}
}
