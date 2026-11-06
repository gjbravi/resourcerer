package recommender

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/applier"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// reportOverwrite counts and announces an applied recommendation whose resources another field manager changed, such
// as a GitOps tool syncing the workload back to its manifests: what the apply saved is gone, and nothing else says so.
// It is called when the pod template's resources change, and reports each apply once.
func (r *Reconciler) reportOverwrite(ctx context.Context, obj client.Object, rec *v1alpha1.Recommendation, w *workload.Workload) {
	live := map[string]corev1.ResourceRequirements{}
	for _, c := range w.Containers() {
		live[c.Name] = c.Resources
	}
	drift := drifted(rec.Status.History, live)
	if len(drift) == 0 {
		return
	}
	applied := rec.Status.History[0]
	id := client.ObjectKeyFromObject(rec).String() + "|" + applied.Time.UTC().String()
	if _, done := r.overwritten.Load(id); done {
		return
	}
	kind, name := w.Kind, w.Name()
	if t := w.TemplateRef; t != nil {
		kind, name = t.Kind, t.Name
	}
	tmpl, err := workload.NewObject(kind)
	if err != nil {
		return
	}
	if err := r.APIReader.Get(ctx, client.ObjectKey{Namespace: w.Namespace(), Name: name}, tmpl); err != nil {
		log.FromContext(ctx).V(1).Info("read managed fields", "error", err.Error())
		return
	}
	manager := overwriter(tmpl.GetManagedFields(), drift, applier.FieldManager)
	if manager == "" {
		return
	}
	r.overwritten.Store(id, true)
	overwrittenCounter.WithLabelValues(manager).Inc()
	r.Recorder.Eventf(obj, rec, corev1.EventTypeWarning, "RecommendationOverwritten", "Overwrite",
		"Recommendation %s applied by %s was changed by %s: %s", applied.Hash, applied.By, manager,
		strings.Join(slices.Sorted(maps.Keys(drift)), ", "))
}

// drifted is what no longer matches the last apply, per container: the values it set (requests/cpu, limits/memory…)
// that something other than resourcerer changed since. Nothing drifts when the last change was a revert, or nothing
// was ever applied. Only the values the apply set are compared, so a limit it left to its owner, or another resource,
// can change freely.
func drifted(history []v1alpha1.HistoryEntry, live map[string]corev1.ResourceRequirements) map[string][]string {
	if len(history) == 0 || history[0].Action != v1alpha1.ActionApply {
		return nil
	}
	out := map[string][]string{}
	for _, c := range history[0].Containers {
		l, ok := live[c.Name]
		if !ok {
			continue
		}
		fields := append(changed("requests", c.To.Requests, l.Requests), changed("limits", c.To.Limits, l.Limits)...)
		if len(fields) > 0 {
			out[c.Name] = fields
		}
	}
	return out
}

// changed lists the CPU and memory values applied set that live no longer has, as kind/resource.
func changed(kind string, applied, live corev1.ResourceList) []string {
	var out []string
	for _, r := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
		a, ok := applied[r]
		if !ok {
			continue
		}
		if l, ok := live[r]; !ok || a.Cmp(l) != 0 {
			out = append(out, kind+"/"+string(r))
		}
	}
	return out
}

// overwriter is the field manager that most recently took one of the drifted values from resourcerer, from the pod
// template's managed fields; empty when resourcerer still owns them all. That is the case right after an apply of ours
// that history does not show yet, even though the manifests' manager still shares the values the apply left alike. A
// removed value has no owner left, so it is not attributed to anyone.
func overwriter(fields []metav1.ManagedFieldsEntry, drift map[string][]string, self string) string {
	manager, latest := "", metav1.Time{}
	for _, f := range fields {
		if f.Manager == self || !owns(f.FieldsV1.GetRawBytes(), drift) {
			continue
		}
		if manager == "" || (f.Time != nil && latest.Before(f.Time)) {
			manager = f.Manager
			if f.Time != nil {
				latest = *f.Time
			}
		}
	}
	return manager
}

// owns reports whether a managed fields set includes one of the drifted values in the pod template.
func owns(raw []byte, drift map[string][]string) bool {
	var set map[string]any
	if json.Unmarshal(raw, &set) != nil {
		return false
	}
	list, ok := dig(set, "f:spec", "f:template", "f:spec", "f:containers")
	if !ok {
		return false
	}
	for key, v := range list {
		var id struct {
			Name string `json:"name"`
		}
		if !strings.HasPrefix(key, "k:") || json.Unmarshal([]byte(key[2:]), &id) != nil {
			continue
		}
		c, ok := v.(map[string]any)
		if !ok {
			continue
		}
		for _, field := range drift[id.Name] {
			kind, res, _ := strings.Cut(field, "/")
			if _, ok := dig(c, "f:resources", "f:"+kind, "f:"+res); ok {
				return true
			}
		}
	}
	return false
}

func dig(m map[string]any, keys ...string) (map[string]any, bool) {
	for _, k := range keys {
		next, ok := m[k].(map[string]any)
		if !ok {
			return nil, false
		}
		m = next
	}
	return m, true
}
