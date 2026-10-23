package workload

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// List returns all objects of the given workload kind matching opts.
func List(ctx context.Context, c client.Reader, kind string, opts ...client.ListOption) ([]client.Object, error) {
	var out []client.Object
	switch kind {
	case "Deployment":
		var l appsv1.DeploymentList
		if err := c.List(ctx, &l, opts...); err != nil {
			return nil, err
		}
		for i := range l.Items {
			out = append(out, &l.Items[i])
		}
	case "StatefulSet":
		var l appsv1.StatefulSetList
		if err := c.List(ctx, &l, opts...); err != nil {
			return nil, err
		}
		for i := range l.Items {
			out = append(out, &l.Items[i])
		}
	case "DaemonSet":
		var l appsv1.DaemonSetList
		if err := c.List(ctx, &l, opts...); err != nil {
			return nil, err
		}
		for i := range l.Items {
			out = append(out, &l.Items[i])
		}
	default:
		return nil, fmt.Errorf("unsupported workload kind %q", kind)
	}
	return out, nil
}
