package workload

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func rollout(spec map[string]any) *unstructured.Unstructured {
	u := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "shop", "namespace": "demo"}, "spec": spec}}
	u.SetGroupVersionKind(RolloutGVK)
	return u
}

func TestRolloutWithTemplate(t *testing.T) {
	w, err := FromObject(rollout(map[string]any{
		"replicas": int64(3),
		"selector": map[string]any{"matchLabels": map[string]any{"app": "shop"}},
		"template": map[string]any{"spec": map[string]any{"containers": []any{
			map[string]any{"name": "shop", "resources": map[string]any{"requests": map[string]any{"cpu": "300m"}}},
		}}},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if w.Kind != Rollout || w.APIVersion != "argoproj.io/v1alpha1" || w.Replicas != 3 || w.TemplateRef != nil {
		t.Fatalf("workload = %+v", w)
	}
	if got := w.Containers()[0].Resources.Requests.Cpu().String(); got != "300m" {
		t.Errorf("cpu request = %s", got)
	}
	if w.ResourceName() != "rollout-shop" {
		t.Errorf("resource name = %s", w.ResourceName())
	}
}

func TestRolloutWithWorkloadRefResolvesTheDeployment(t *testing.T) {
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "shop-template", Namespace: "demo"},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name:      "shop",
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("256Mi")}},
		}}}}},
	}
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	c := fake.NewClientBuilder().WithScheme(s).WithObjects(d).Build()

	w, err := FromObject(rollout(map[string]any{
		"workloadRef": map[string]any{"apiVersion": "apps/v1", "kind": "Deployment", "name": "shop-template"},
	}))
	if err != nil {
		t.Fatal(err)
	}
	if w.TemplateRef == nil || w.TemplateRef.Name != "shop-template" || w.Replicas != 1 {
		t.Fatalf("workload = %+v", w)
	}
	if err := Resolve(context.Background(), c, w); err != nil {
		t.Fatal(err)
	}
	if got := w.Containers()[0].Resources.Requests.Memory().String(); got != "256Mi" {
		t.Errorf("memory request from the referenced deployment = %s", got)
	}
}
