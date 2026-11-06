package recommender

import (
	"maps"
	"slices"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

func list(kv ...string) corev1.ResourceList {
	out := corev1.ResourceList{}
	for i := 0; i < len(kv); i += 2 {
		out[corev1.ResourceName(kv[i])] = resource.MustParse(kv[i+1])
	}
	return out
}

func TestDrifted(t *testing.T) {
	applied := []v1alpha1.HistoryEntry{{Action: v1alpha1.ActionApply, Containers: []v1alpha1.ContainerChange{
		{Name: "app", To: corev1.ResourceRequirements{Requests: list("cpu", "100m", "memory", "128Mi"), Limits: list("memory", "256Mi")}},
		{Name: "worker", To: corev1.ResourceRequirements{Requests: list("cpu", "50m")}},
	}}}
	still := map[string]corev1.ResourceRequirements{
		"app":    {Requests: list("cpu", "0.1", "memory", "128Mi"), Limits: list("memory", "256Mi", "cpu", "1")},
		"worker": {Requests: list("cpu", "50m", "memory", "64Mi")},
	}
	for _, tc := range []struct {
		name    string
		history []v1alpha1.HistoryEntry
		live    map[string]corev1.ResourceRequirements
		want    map[string][]string
	}{
		{"nothing applied", nil, still, nil},
		{"the last change was a revert", []v1alpha1.HistoryEntry{{Action: v1alpha1.ActionRevert}, applied[0]}, nil, nil},
		{"the same values, written differently, and others the apply did not set", applied, still, map[string][]string{}},
		{"a request changed back", applied, map[string]corev1.ResourceRequirements{
			"app": {Requests: list("cpu", "500m", "memory", "128Mi"), Limits: list("memory", "256Mi")}, "worker": still["worker"],
		}, map[string][]string{"app": {"requests/cpu"}}},
		{"an applied limit removed", applied, map[string]corev1.ResourceRequirements{
			"app": {Requests: list("cpu", "100m", "memory", "128Mi")}, "worker": still["worker"],
		}, map[string][]string{"app": {"limits/memory"}}},
		{"a container gone from the template", applied, map[string]corev1.ResourceRequirements{"app": still["app"]}, map[string][]string{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := drifted(tc.history, tc.live)
			if len(got) != len(tc.want) || !maps.EqualFunc(got, tc.want, slices.Equal) {
				t.Errorf("drifted = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestOverwriterIsTheLatestOtherManagerOfTheDriftedValues(t *testing.T) {
	at := func(m int) *metav1.Time {
		t := metav1.NewTime(time.Date(2026, 11, 6, 12, m, 0, 0, time.UTC))
		return &t
	}
	container := func(name, resources string) *metav1.FieldsV1 {
		return metav1.NewFieldsV1(`{"f:spec":{"f:template":{"f:spec":{"f:containers":{"k:{\"name\":\"` + name +
			`\"}":{".":{},"f:name":{},"f:resources":` + resources + `}}}}}}`)
	}
	cpu := `{"f:requests":{"f:cpu":{}}}`
	memoryLimit := `{"f:limits":{"f:memory":{}}}`
	drift := map[string][]string{"app": {"requests/cpu"}}
	for _, tc := range []struct {
		name   string
		fields []metav1.ManagedFieldsEntry
		want   string
	}{
		{"resourcerer owns the changed value, the manifests' manager only others", []metav1.ManagedFieldsEntry{
			{Manager: "resourcerer", Time: at(5), FieldsV1: container("app", cpu)},
			{Manager: "kubectl", Time: at(1), FieldsV1: container("app", memoryLimit)},
		}, ""},
		{"a GitOps sync took the value back", []metav1.ManagedFieldsEntry{
			{Manager: "argocd-controller", Time: at(9), FieldsV1: container("app", cpu)},
		}, "argocd-controller"},
		{"the latest of two", []metav1.ManagedFieldsEntry{
			{Manager: "helm", Time: at(2), FieldsV1: container("app", cpu)},
			{Manager: "kubectl-set", Time: at(7), FieldsV1: container("app", cpu)},
		}, "kubectl-set"},
		{"another container's values do not count", []metav1.ManagedFieldsEntry{
			{Manager: "kubectl-edit", Time: at(3), FieldsV1: container("sidecar", cpu)},
		}, ""},
		{"unreadable fields are skipped", []metav1.ManagedFieldsEntry{
			{Manager: "broken", Time: at(3), FieldsV1: metav1.NewFieldsV1(`not json`)},
		}, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := overwriter(tc.fields, drift, "resourcerer"); got != tc.want {
				t.Errorf("overwriter = %q, want %q", got, tc.want)
			}
		})
	}
}
