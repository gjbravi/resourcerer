// Package controller reconciles workloads into managed VPAs.
package controller

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/client-go/tools/events"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/vpa"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// Reconciler keeps one managed VPA per selected workload of a given kind.
type Reconciler struct {
	client.Client
	Kind     string
	Options  config.Options
	Recorder events.EventRecorder
}

// Setup registers a reconciler for every supported workload kind.
func Setup(mgr ctrl.Manager, opts config.Options) error {
	for _, kind := range workload.Kinds {
		r := &Reconciler{Client: mgr.GetClient(), Kind: kind, Options: opts, Recorder: mgr.GetEventRecorder("resourcerer")}
		if err := r.setup(mgr); err != nil {
			return err
		}
	}
	return nil
}

func (r *Reconciler) setup(mgr ctrl.Manager) error {
	obj, _ := workload.NewObject(r.Kind)
	notExcluded := predicate.NewPredicateFuncs(func(o client.Object) bool { return !r.Options.Excluded(o.GetNamespace()) })
	return ctrl.NewControllerManagedBy(mgr).
		Named("resourcerer-"+r.Kind).
		For(obj, builder.WithPredicates(notExcluded, predicate.Or(
			predicate.GenerationChangedPredicate{},
			predicate.LabelChangedPredicate{},
			predicate.AnnotationChangedPredicate{},
		))).
		Owns(&vpav1.VerticalPodAutoscaler{}, builder.WithPredicates(notExcluded, predicate.GenerationChangedPredicate{})).
		Watches(&corev1.Namespace{}, handler.EnqueueRequestsFromMapFunc(r.mapNamespace),
			builder.WithPredicates(predicate.Or(predicate.LabelChangedPredicate{}, predicate.AnnotationChangedPredicate{}))).
		Complete(r)
}

func (r *Reconciler) mapNamespace(ctx context.Context, ns client.Object) []reconcile.Request {
	if r.Options.Excluded(ns.GetName()) {
		return nil
	}
	objs, err := workload.List(ctx, r.Client, r.Kind, client.InNamespace(ns.GetName()))
	if err != nil {
		log.FromContext(ctx).Error(err, "list workloads for namespace", "namespace", ns.GetName())
		return nil
	}
	reqs := make([]reconcile.Request, 0, len(objs))
	for _, o := range objs {
		reqs = append(reqs, reconcile.Request{NamespacedName: client.ObjectKeyFromObject(o)})
	}
	return reqs
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj, err := workload.NewObject(r.Kind)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		// The managed VPA is garbage-collected through its owner reference.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	if obj.GetDeletionTimestamp() != nil {
		return ctrl.Result{}, nil
	}
	w, err := workload.FromObject(obj)
	if err != nil {
		return ctrl.Result{}, err
	}

	var ns corev1.Namespace
	if err := r.Get(ctx, types.NamespacedName{Name: w.Namespace()}, &ns); err != nil {
		return ctrl.Result{}, err
	}
	d := config.Decide(r.Options, w.Namespace(), ns.Labels, ns.Annotations, obj.GetLabels(), obj.GetAnnotations())
	if d.Warning != "" {
		r.Recorder.Eventf(obj, nil, corev1.EventTypeWarning, "InvalidConfiguration", "Reconcile", d.Warning)
	}
	if !d.Managed {
		return ctrl.Result{}, vpa.Delete(ctx, r.Client, w.Namespace(), vpa.Name(w))
	}

	_, err = vpa.Ensure(ctx, r.Client, w, d.Mode)
	if errors.Is(err, vpa.ErrNotManaged) {
		r.Recorder.Eventf(obj, nil, corev1.EventTypeWarning, "VPAConflict", "Reconcile",
			"VPA %s exists without label %s=true; resourcerer will not modify it", vpa.Name(w), config.LabelManaged)
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, err
}
