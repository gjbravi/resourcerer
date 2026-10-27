// Package workload abstracts the pod-owning workload kinds resourcerer manages.
package workload

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Rollout is read as unstructured, so resourcerer does not depend on the Argo Rollouts Go module.
const Rollout = "Rollout"

var RolloutGVK = schema.GroupVersionKind{Group: "argoproj.io", Version: "v1alpha1", Kind: Rollout}

// Kinds are the workload kinds always available; Rollout is added when its CRD is installed.
var Kinds = []string{"Deployment", "StatefulSet", "DaemonSet"}

// AllKinds includes kinds that depend on optional CRDs.
var AllKinds = []string{"Deployment", "StatefulSet", "DaemonSet", Rollout}

// Workload is a uniform view over Deployment, StatefulSet, DaemonSet and Argo Rollout.
type Workload struct {
	Object     client.Object
	Kind       string
	APIVersion string
	Template   *corev1.PodTemplateSpec
	Selector   *metav1.LabelSelector
	Replicas   int32
	// TemplateRef names a Deployment holding the pod template (Rollout spec.workloadRef). Nil means the
	// workload's own template, which is also where recommendations are applied.
	TemplateRef *Ref
}

type Ref struct {
	APIVersion, Kind, Name string
}

func (w *Workload) Name() string      { return w.Object.GetName() }
func (w *Workload) Namespace() string { return w.Object.GetNamespace() }

// ResourceName is the name of the workload's Recommendation.
func (w *Workload) ResourceName() string {
	return ResourceName(w.Kind, w.Name())
}

func ResourceName(kind, name string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(kind), name)
}

func (w *Workload) Containers() []corev1.Container {
	if w.Template == nil {
		return nil
	}
	return w.Template.Spec.Containers
}

func APIVersion(kind string) string {
	if kind == Rollout {
		return RolloutGVK.GroupVersion().String()
	}
	return "apps/v1"
}

func FromObject(obj client.Object) (*Workload, error) {
	desired := func(p *int32) int32 {
		if p == nil {
			return 1
		}
		return *p
	}
	switch o := obj.(type) {
	case *appsv1.Deployment:
		return &Workload{Object: o, Kind: "Deployment", APIVersion: "apps/v1", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: desired(o.Spec.Replicas)}, nil
	case *appsv1.StatefulSet:
		return &Workload{Object: o, Kind: "StatefulSet", APIVersion: "apps/v1", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: desired(o.Spec.Replicas)}, nil
	case *appsv1.DaemonSet:
		return &Workload{Object: o, Kind: "DaemonSet", APIVersion: "apps/v1", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: o.Status.DesiredNumberScheduled}, nil
	case *unstructured.Unstructured:
		if o.GetKind() == Rollout {
			return fromRollout(o)
		}
	}
	return nil, fmt.Errorf("unsupported workload type %T", obj)
}

// rolloutSpec is the part of an Argo Rollout resourcerer reads.
type rolloutSpec struct {
	Replicas    *int32                  `json:"replicas"`
	Selector    *metav1.LabelSelector   `json:"selector"`
	Template    *corev1.PodTemplateSpec `json:"template"`
	WorkloadRef *struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Name       string `json:"name"`
	} `json:"workloadRef"`
}

func fromRollout(u *unstructured.Unstructured) (*Workload, error) {
	raw, _, _ := unstructured.NestedMap(u.Object, "spec")
	var spec rolloutSpec
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw, &spec); err != nil {
		return nil, fmt.Errorf("read rollout spec: %w", err)
	}
	w := &Workload{Object: u, Kind: Rollout, APIVersion: RolloutGVK.GroupVersion().String(), Template: spec.Template, Selector: spec.Selector, Replicas: 1}
	if spec.Replicas != nil {
		w.Replicas = *spec.Replicas
	}
	if r := spec.WorkloadRef; r != nil && r.Name != "" {
		w.TemplateRef = &Ref{APIVersion: r.APIVersion, Kind: r.Kind, Name: r.Name}
	}
	return w, nil
}

// Resolve loads the pod template a Rollout references with spec.workloadRef.
func Resolve(ctx context.Context, c client.Reader, w *Workload) error {
	if w.TemplateRef == nil {
		return nil
	}
	if w.TemplateRef.Kind != "Deployment" {
		return fmt.Errorf("workloadRef kind %q is not supported", w.TemplateRef.Kind)
	}
	var d appsv1.Deployment
	if err := c.Get(ctx, client.ObjectKey{Namespace: w.Namespace(), Name: w.TemplateRef.Name}, &d); err != nil {
		return fmt.Errorf("get workloadRef deployment %s: %w", w.TemplateRef.Name, err)
	}
	w.Template = &d.Spec.Template
	if w.Selector == nil {
		w.Selector = d.Spec.Selector
	}
	return nil
}

// Get fetches and resolves a workload.
func Get(ctx context.Context, c client.Reader, kind string, key client.ObjectKey) (*Workload, error) {
	obj, err := NewObject(kind)
	if err != nil {
		return nil, err
	}
	if err := c.Get(ctx, key, obj); err != nil {
		return nil, err
	}
	w, err := FromObject(obj)
	if err != nil {
		return nil, err
	}
	return w, Resolve(ctx, c, w)
}

// NewObject returns an empty object of the given kind.
func NewObject(kind string) (client.Object, error) {
	switch kind {
	case "Deployment":
		return &appsv1.Deployment{}, nil
	case "StatefulSet":
		return &appsv1.StatefulSet{}, nil
	case "DaemonSet":
		return &appsv1.DaemonSet{}, nil
	case Rollout:
		u := &unstructured.Unstructured{}
		u.SetGroupVersionKind(RolloutGVK)
		return u, nil
	}
	return nil, fmt.Errorf("unsupported workload kind %q", kind)
}

// OwnerReference makes the workload own resourcerer objects so they are garbage-collected with it.
func (w *Workload) OwnerReference() metav1.OwnerReference {
	t := true
	return metav1.OwnerReference{
		APIVersion:         w.APIVersion,
		Kind:               w.Kind,
		Name:               w.Name(),
		UID:                w.Object.GetUID(),
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}
}

// ReferencedDeployments returns the names of Deployments used as templates by Rollouts in a namespace.
// A missing Rollout CRD yields none.
func ReferencedDeployments(ctx context.Context, c client.Reader, namespace string) map[string]string {
	out := map[string]string{}
	l := &unstructured.UnstructuredList{}
	l.SetGroupVersionKind(RolloutGVK.GroupVersion().WithKind("RolloutList"))
	if err := c.List(ctx, l, client.InNamespace(namespace)); err != nil {
		return out
	}
	for i := range l.Items {
		w, err := fromRollout(&l.Items[i])
		if err == nil && w.TemplateRef != nil && w.TemplateRef.Kind == "Deployment" {
			out[w.TemplateRef.Name] = w.Name()
		}
	}
	return out
}
