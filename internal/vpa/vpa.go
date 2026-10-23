// Package vpa builds and reads the VerticalPodAutoscaler objects resourcerer manages.
package vpa

import (
	"context"
	"fmt"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// ErrNotManaged means a VPA with the expected name exists but was not created by resourcerer.
var ErrNotManaged = fmt.Errorf("VPA exists and is not managed by resourcerer")

// Name is the VPA name for a workload: <kind>-<name>, lowercase.
func Name(w *workload.Workload) string { return w.ResourceName() }

func updateMode(m config.Mode) vpav1.UpdateMode {
	if m == config.ModeAuto {
		return vpav1.UpdateModeInPlaceOrRecreate
	}
	return vpav1.UpdateModeOff
}

// ModeOf maps a VPA update mode back to a resourcerer mode.
func ModeOf(v *vpav1.VerticalPodAutoscaler) config.Mode {
	if v.Spec.UpdatePolicy != nil && v.Spec.UpdatePolicy.UpdateMode != nil && *v.Spec.UpdatePolicy.UpdateMode != vpav1.UpdateModeOff {
		return config.ModeAuto
	}
	return config.ModeOff
}

// Desired renders the VPA for a workload.
func Desired(w *workload.Workload, mode config.Mode) *vpav1.VerticalPodAutoscaler {
	apiVersion, kind, name := w.TargetRef()
	um := updateMode(mode)
	return &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            Name(w),
			Namespace:       w.Namespace(),
			Labels:          map[string]string{config.LabelManaged: "true"},
			OwnerReferences: []metav1.OwnerReference{w.OwnerReference()},
		},
		Spec: vpav1.VerticalPodAutoscalerSpec{
			TargetRef:    &autoscalingv1.CrossVersionObjectReference{APIVersion: apiVersion, Kind: kind, Name: name},
			UpdatePolicy: &vpav1.PodUpdatePolicy{UpdateMode: &um},
		},
	}
}

// Ensure creates or updates the managed VPA. It never modifies VPAs without the managed label.
func Ensure(ctx context.Context, c client.Client, w *workload.Workload, mode config.Mode) (*vpav1.VerticalPodAutoscaler, error) {
	desired := Desired(w, mode)
	current := &vpav1.VerticalPodAutoscaler{}
	err := c.Get(ctx, types.NamespacedName{Namespace: desired.Namespace, Name: desired.Name}, current)
	if apierrors.IsNotFound(err) {
		return desired, c.Create(ctx, desired)
	}
	if err != nil {
		return nil, err
	}
	if current.Labels[config.LabelManaged] != "true" {
		return current, ErrNotManaged
	}
	// Only the fields resourcerer owns are reconciled; anything else set on the VPA by hand is kept.
	changed := false
	if !equality.Semantic.DeepEqual(current.Spec.TargetRef, desired.Spec.TargetRef) {
		current.Spec.TargetRef = desired.Spec.TargetRef
		changed = true
	}
	if ModeOf(current) != mode || current.Spec.UpdatePolicy == nil {
		if current.Spec.UpdatePolicy == nil {
			current.Spec.UpdatePolicy = &vpav1.PodUpdatePolicy{}
		}
		current.Spec.UpdatePolicy.UpdateMode = desired.Spec.UpdatePolicy.UpdateMode
		changed = true
	}
	if !equality.Semantic.DeepEqual(current.OwnerReferences, desired.OwnerReferences) {
		current.OwnerReferences = desired.OwnerReferences
		changed = true
	}
	if !changed {
		return current, nil
	}
	return current, c.Update(ctx, current)
}

// Delete removes the managed VPA for a workload, if any.
func Delete(ctx context.Context, c client.Client, namespace, name string) error {
	v := &vpav1.VerticalPodAutoscaler{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, v); err != nil {
		return client.IgnoreNotFound(err)
	}
	if v.Labels[config.LabelManaged] != "true" {
		return nil
	}
	return client.IgnoreNotFound(c.Delete(ctx, v))
}
