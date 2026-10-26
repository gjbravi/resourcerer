// Package workload abstracts the pod-owning workload kinds resourcerer manages.
package workload

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Workload is a uniform view over Deployment, StatefulSet and DaemonSet.
type Workload struct {
	Object   client.Object
	Kind     string
	Template *corev1.PodTemplateSpec
	Selector *metav1.LabelSelector
	Replicas int32
}

func (w *Workload) Name() string      { return w.Object.GetName() }
func (w *Workload) Namespace() string { return w.Object.GetNamespace() }

func (w *Workload) TargetRef() (apiVersion, kind, name string) {
	return "apps/v1", w.Kind, w.Name()
}

// ResourceName is the name shared by the Recommendation and VPA of a workload.
func (w *Workload) ResourceName() string {
	return ResourceName(w.Kind, w.Name())
}

func ResourceName(kind, name string) string {
	return fmt.Sprintf("%s-%s", strings.ToLower(kind), name)
}

func (w *Workload) Containers() []corev1.Container { return w.Template.Spec.Containers }

func FromObject(obj client.Object) (*Workload, error) {
	desired := func(p *int32) int32 {
		if p == nil {
			return 1
		}
		return *p
	}
	switch o := obj.(type) {
	case *appsv1.Deployment:
		return &Workload{Object: o, Kind: "Deployment", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: desired(o.Spec.Replicas)}, nil
	case *appsv1.StatefulSet:
		return &Workload{Object: o, Kind: "StatefulSet", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: desired(o.Spec.Replicas)}, nil
	case *appsv1.DaemonSet:
		return &Workload{Object: o, Kind: "DaemonSet", Template: &o.Spec.Template, Selector: o.Spec.Selector, Replicas: o.Status.DesiredNumberScheduled}, nil
	}
	return nil, fmt.Errorf("unsupported workload type %T", obj)
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
	}
	return nil, fmt.Errorf("unsupported workload kind %q", kind)
}

var Kinds = []string{"Deployment", "StatefulSet", "DaemonSet"}

// OwnerReference makes the workload own resourcerer-managed objects so they are garbage-collected with it.
func (w *Workload) OwnerReference() metav1.OwnerReference {
	t := true
	return metav1.OwnerReference{
		APIVersion:         "apps/v1",
		Kind:               w.Kind,
		Name:               w.Name(),
		UID:                w.Object.GetUID(),
		Controller:         &t,
		BlockOwnerDeletion: &t,
	}
}
