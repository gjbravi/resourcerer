// Package server serves the resourcerer HTTP API and embedded UI.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	eventsv1 "k8s.io/api/events/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/vpa"
	"github.com/gjbravi/resourcerer/internal/workload"
)

type Server struct {
	Client  client.Client
	Options config.Options
	Version string
	// ReadOnly disables mode changes from the UI.
	ReadOnly bool
	// UI is the built frontend; nil serves the API only.
	UI fs.FS
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/info", s.info)
	mux.HandleFunc("GET /api/v1/workloads", s.list)
	mux.HandleFunc("GET /api/v1/workloads/{namespace}/{kind}/{name}", s.get)
	mux.HandleFunc("PUT /api/v1/workloads/{namespace}/{kind}/{name}/mode", s.setMode)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if s.UI != nil {
		mux.Handle("/", spa(s.UI))
	}
	return mux
}

// Start serves until ctx is done; it runs on every replica, not only the leader.
func (s *Server) Start(addr string) func(context.Context) error {
	return func(ctx context.Context) error {
		srv := &http.Server{Addr: addr, Handler: s.Handler(), ReadHeaderTimeout: 10 * time.Second}
		go func() {
			<-ctx.Done()
			shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = srv.Shutdown(shutdown)
		}()
		log.FromContext(ctx).Info("serving API and UI", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func (s *Server) info(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, Info{
		Version: s.Version, Selection: string(s.Options.Selection), DefaultMode: string(s.Options.DefaultMode), ReadOnly: s.ReadOnly,
	})
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var opts []client.ListOption
	if ns := r.URL.Query().Get("namespace"); ns != "" {
		opts = append(opts, client.InNamespace(ns))
	}
	var vpas vpav1.VerticalPodAutoscalerList
	if err := s.Client.List(ctx, &vpas, opts...); err != nil {
		writeErr(w, err)
		return
	}
	byName := map[types.NamespacedName]*vpav1.VerticalPodAutoscaler{}
	for i := range vpas.Items {
		v := &vpas.Items[i]
		byName[types.NamespacedName{Namespace: v.Namespace, Name: v.Name}] = v
	}
	namespaces := map[string]*corev1.Namespace{}

	out := []Workload{}
	for _, kind := range workload.Kinds {
		objs, err := workload.List(ctx, s.Client, kind, opts...)
		if err != nil {
			writeErr(w, err)
			return
		}
		for _, obj := range objs {
			if s.Options.Excluded(obj.GetNamespace()) {
				continue
			}
			wl, _ := workload.FromObject(obj)
			ns, err := s.namespace(ctx, namespaces, wl.Namespace())
			if err != nil {
				writeErr(w, err)
				return
			}
			out = append(out, s.summarize(wl, ns, byName[types.NamespacedName{Namespace: wl.Namespace(), Name: vpa.Name(wl)}]))
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Namespace+"/"+out[i].Kind+"/"+out[i].Name < out[j].Namespace+"/"+out[j].Kind+"/"+out[j].Name
	})
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	wl, err := s.load(ctx, r)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, s.detail(ctx, wl))
}

func (s *Server) detail(ctx context.Context, wl *workload.Workload) WorkloadDetail {
	ns, _ := s.namespace(ctx, map[string]*corev1.Namespace{}, wl.Namespace())
	v := &vpav1.VerticalPodAutoscaler{}
	if err := s.Client.Get(ctx, types.NamespacedName{Namespace: wl.Namespace(), Name: vpa.Name(wl)}, v); err != nil {
		v = nil
	}
	return WorkloadDetail{Workload: s.summarize(wl, ns, v), Pods: s.pods(ctx, wl), Events: s.events(ctx, wl)}
}

// setMode writes the opt-out label or the mode annotation on the workload; the
// controller reconciles the VPA from there, so GitOps users can set the same keys.
func (s *Server) setMode(w http.ResponseWriter, r *http.Request) {
	if s.ReadOnly {
		writeJSON(w, http.StatusForbidden, errBody("resourcerer is running in read-only mode"))
		return
	}
	ctx := r.Context()
	var body ModeRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	wl, err := s.load(ctx, r)
	if err != nil {
		writeErr(w, err)
		return
	}

	labels := map[string]any{}
	annotations := map[string]any{}
	switch body.Mode {
	case ModeDisabled:
		labels[config.LabelEnabled] = "false"
	case ModeOff, ModeAuto:
		labels[config.LabelEnabled] = "true"
		annotations[config.AnnotationMode] = string(body.Mode)
	default:
		writeJSON(w, http.StatusBadRequest, errBody(fmt.Sprintf("mode must be %s, %s or %s", ModeOff, ModeAuto, ModeDisabled)))
		return
	}
	patch, _ := json.Marshal(map[string]any{"metadata": map[string]any{"labels": labels, "annotations": annotations}})
	if err := s.Client.Patch(ctx, wl.Object, client.RawPatch(types.MergePatchType, patch), client.FieldOwner("resourcerer-ui")); err != nil {
		writeErr(w, err)
		return
	}
	log.FromContext(ctx).Info("mode changed", "workload", wl.Namespace()+"/"+wl.Name(), "kind", wl.Kind, "mode", body.Mode, "user", user(r))
	writeJSON(w, http.StatusOK, s.detail(ctx, wl))
}

func (s *Server) load(ctx context.Context, r *http.Request) (*workload.Workload, error) {
	ns, kind, name := r.PathValue("namespace"), r.PathValue("kind"), r.PathValue("name")
	if s.Options.Excluded(ns) {
		return nil, &httpError{http.StatusNotFound, fmt.Sprintf("namespace %s is excluded", ns)}
	}
	obj, err := workload.NewObject(kind)
	if err != nil {
		return nil, &httpError{http.StatusBadRequest, err.Error()}
	}
	if err := s.Client.Get(ctx, types.NamespacedName{Namespace: ns, Name: name}, obj); err != nil {
		return nil, err
	}
	return workload.FromObject(obj)
}

func (s *Server) namespace(ctx context.Context, cache map[string]*corev1.Namespace, name string) (*corev1.Namespace, error) {
	if ns, ok := cache[name]; ok {
		return ns, nil
	}
	ns := &corev1.Namespace{}
	if err := s.Client.Get(ctx, types.NamespacedName{Name: name}, ns); err != nil {
		return nil, err
	}
	cache[name] = ns
	return ns, nil
}

func (s *Server) summarize(wl *workload.Workload, ns *corev1.Namespace, v *vpav1.VerticalPodAutoscaler) Workload {
	out := Workload{Namespace: wl.Namespace(), Kind: wl.Kind, Name: wl.Name(), Replicas: wl.Replicas, Containers: []Container{}}
	var nsLabels, nsAnn map[string]string
	if ns != nil {
		nsLabels, nsAnn = ns.Labels, ns.Annotations
	}
	d := config.Decide(s.Options, wl.Namespace(), nsLabels, nsAnn, wl.Object.GetLabels(), wl.Object.GetAnnotations())
	out.Message = d.Warning
	switch {
	case !d.Managed:
		out.Mode, out.Status, out.ModeSource = ModeDisabled, StatusNone, d.Reason
	default:
		out.Mode = Mode(d.Mode)
		out.ModeSource = modeSource(nsAnn, wl.Object.GetAnnotations())
	}

	recs := map[string]vpav1.RecommendedContainerResources{}
	if d.Managed {
		out.Status, out.Message = vpaStatus(v, out.Message)
		if v != nil {
			out.VPA = v.Name
			if v.Status.Recommendation != nil {
				for _, c := range v.Status.Recommendation.ContainerRecommendations {
					recs[c.ContainerName] = c
				}
			}
		}
	}
	for _, c := range wl.Containers() {
		ct := Container{Name: c.Name, Requests: resources(c.Resources.Requests), Limits: resources(c.Resources.Limits)}
		if rc, ok := recs[c.Name]; ok {
			ct.Target, ct.LowerBound, ct.UpperBound, ct.Uncapped = ptr(resources(rc.Target)), ptr(resources(rc.LowerBound)), ptr(resources(rc.UpperBound)), ptr(resources(rc.UncappedTarget))
		}
		out.Containers = append(out.Containers, ct)
	}
	return out
}

func modeSource(nsAnn, wlAnn map[string]string) string {
	if _, err := config.ParseMode(wlAnn[config.AnnotationMode]); err == nil {
		return "workload"
	}
	if _, err := config.ParseMode(nsAnn[config.AnnotationMode]); err == nil {
		return "namespace"
	}
	return "default"
}

func vpaStatus(v *vpav1.VerticalPodAutoscaler, warning string) (Status, string) {
	if v == nil {
		return StatusPending, firstNonEmpty(warning, "VPA not created yet")
	}
	if v.Labels[config.LabelManaged] != "true" {
		return StatusConflict, fmt.Sprintf("VPA %s exists and is not managed by resourcerer", v.Name)
	}
	cond := map[vpav1.VerticalPodAutoscalerConditionType]vpav1.VerticalPodAutoscalerCondition{}
	for _, c := range v.Status.Conditions {
		cond[c.Type] = c
	}
	isTrue := func(t vpav1.VerticalPodAutoscalerConditionType) bool { return cond[t].Status == corev1.ConditionTrue }
	switch {
	case isTrue(vpav1.ConfigUnsupported):
		return StatusError, cond[vpav1.ConfigUnsupported].Message
	case isTrue(vpav1.NoPodsMatched):
		return StatusError, firstNonEmpty(cond[vpav1.NoPodsMatched].Message, "no pods matched")
	case isTrue(vpav1.RecommendationProvided) && isTrue(vpav1.LowConfidence):
		return StatusLowConfidence, "not enough usage history yet"
	case isTrue(vpav1.RecommendationProvided):
		return StatusReady, warning
	}
	return StatusPending, firstNonEmpty(warning, "waiting for the VPA recommender")
}

func (s *Server) pods(ctx context.Context, wl *workload.Workload) []Pod {
	out := []Pod{}
	if wl.Selector == nil {
		return out
	}
	sel, err := metav1.LabelSelectorAsSelector(wl.Selector)
	if err != nil || sel.Empty() {
		return out
	}
	var list corev1.PodList
	if err := s.Client.List(ctx, &list, client.InNamespace(wl.Namespace()), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return out
	}
	for _, p := range list.Items {
		if p.DeletionTimestamp != nil {
			continue
		}
		pod := Pod{Name: p.Name, Containers: []Container{}}
		for _, c := range p.Spec.Containers {
			pod.Containers = append(pod.Containers, Container{Name: c.Name, Requests: resources(c.Resources.Requests), Limits: resources(c.Resources.Limits)})
		}
		out = append(out, pod)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *Server) events(ctx context.Context, wl *workload.Workload) []Event {
	out := []Event{}
	var list eventsv1.EventList
	if err := s.Client.List(ctx, &list, client.InNamespace(wl.Namespace())); err != nil {
		return out
	}
	for _, e := range list.Items {
		if e.Regarding.Name != wl.Name() || e.Regarding.Kind != wl.Kind {
			continue
		}
		t := e.EventTime.Time
		if t.IsZero() {
			t = e.DeprecatedLastTimestamp.Time
		}
		if t.IsZero() {
			t = e.CreationTimestamp.Time
		}
		out = append(out, Event{Type: e.Type, Reason: e.Reason, Message: e.Note, Time: t.UTC().Format(time.RFC3339)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time > out[j].Time })
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

func resources(l corev1.ResourceList) Resources {
	r := Resources{}
	if q, ok := l[corev1.ResourceCPU]; ok {
		r.CPU = q.String()
	}
	if q, ok := l[corev1.ResourceMemory]; ok {
		r.Memory = normalizeMemory(q)
	}
	return r
}

// normalizeMemory renders plain byte counts (as VPA reports them) in binary units.
func normalizeMemory(q resource.Quantity) string {
	if q.Format == resource.BinarySI {
		return q.String()
	}
	return resource.NewQuantity(q.Value(), resource.BinarySI).String()
}

func ptr[T any](v T) *T { return &v }

func firstNonEmpty(s ...string) string {
	for _, v := range s {
		if v != "" {
			return v
		}
	}
	return ""
}

func user(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-Email", "X-Forwarded-User", "X-Auth-Request-Email"} {
		if v := r.Header.Get(h); v != "" {
			return v
		}
	}
	return "anonymous"
}

type httpError struct {
	code int
	msg  string
}

func (e *httpError) Error() string { return e.msg }

func errBody(msg string) map[string]string { return map[string]string{"error": msg} }

func writeErr(w http.ResponseWriter, err error) {
	var he *httpError
	switch {
	case errors.As(err, &he):
		writeJSON(w, he.code, errBody(he.msg))
	case apierrors.IsNotFound(err):
		writeJSON(w, http.StatusNotFound, errBody(err.Error()))
	case apierrors.IsForbidden(err):
		writeJSON(w, http.StatusForbidden, errBody(err.Error()))
	default:
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// spa serves static files and falls back to index.html for client-side routes.
// Hashed assets are cached forever, index.html never, so a deploy is picked up on reload.
func spa(ui fs.FS) http.Handler {
	files := http.FileServer(http.FS(ui))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" && p != "index.html" {
			if _, err := fs.Stat(ui, p); err == nil {
				if strings.HasPrefix(p, "assets/") {
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
				files.ServeHTTP(w, r)
				return
			}
			if strings.HasPrefix(p, "assets/") {
				http.NotFound(w, r)
				return
			}
		}
		w.Header().Set("Cache-Control", "no-cache")
		r2 := r.Clone(r.Context())
		r2.URL.Path = "/"
		files.ServeHTTP(w, r2)
	})
}
