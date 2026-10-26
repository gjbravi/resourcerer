// Package server serves the REST API and the embedded UI.
//
// Authentication is "anonymous", like Kiali's anonymous strategy: anyone who reaches the
// server can use it, with the server's own ServiceAccount. Use ReadOnly to forbid changes.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/cluster"
	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
	"github.com/gjbravi/resourcerer/internal/workload"
)

// User is recorded as the requester of every change, since authentication is anonymous.
const User = "anonymous"

type Server struct {
	Client    client.Client
	APIReader client.Reader
	Cluster   cluster.Config
	Prom      *prom.Client
	PromURL   string
	MinStep   time.Duration
	Version   string
	ReadOnly  bool
	UI        fs.FS

	cache cache
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/info", s.info)
	mux.HandleFunc("GET /api/v1/workloads", s.list)
	mux.HandleFunc("GET /api/v1/workloads/{ns}/{kind}/{name}", s.detail)
	mux.HandleFunc("GET /api/v1/workloads/{ns}/{kind}/{name}/series", s.series)
	mux.HandleFunc("GET /api/v1/workloads/{ns}/{kind}/{name}/impact", s.impact)
	mux.HandleFunc("POST /api/v1/workloads/{ns}/{kind}/{name}/apply", s.write(s.apply))
	mux.HandleFunc("POST /api/v1/workloads/{ns}/{kind}/{name}/revert", s.write(s.revert))
	mux.HandleFunc("POST /api/v1/workloads/{ns}/{kind}/{name}/regenerate", s.write(s.regenerate))
	mux.HandleFunc("PUT /api/v1/workloads/{ns}/{kind}/{name}/settings", s.write(s.workloadSetting))
	mux.HandleFunc("PUT /api/v1/namespaces/{ns}/settings", s.write(s.namespaceSetting))
	mux.HandleFunc("/api/", func(w http.ResponseWriter, _ *http.Request) { writeError(w, http.StatusNotFound, "not found") })
	if s.UI != nil {
		mux.Handle("/", spa(s.UI))
	}
	return logRequests(mux)
}

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
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}
}

func (s *Server) info(w http.ResponseWriter, r *http.Request) {
	defaults, rates := s.Cluster.Defaults(r.Context(), s.Client)
	eff := settings.Resolve(defaults, nil, nil)
	st := s.cache.get("prom-status", 30*time.Second, func() (any, error) {
		return s.Prom.Status(r.Context(), s.PromURL), nil
	})
	writeJSON(w, Info{
		Version:            s.Version,
		Cluster:            s.Cluster.Name,
		Auth:               "anonymous",
		ReadOnly:           s.ReadOnly,
		Selection:          string(s.Cluster.Selection),
		ExcludedNamespaces: s.Cluster.ExcludedNamespaces,
		Defaults:           settingInfos(eff, nil),
		Cost:               Cost{CPUCoreHour: rates.CPUCoreHour, MemoryGiBHour: rates.MemoryGiBHour},
		Prometheus:         st.(prom.Status),
	})
}

func settingInfos(eff settings.Effective, st map[string]v1alpha1.Setting) []SettingInfo {
	out := make([]SettingInfo, 0, len(settings.All))
	for _, d := range settings.All {
		v := eff.Values[d.Key]
		if st != nil {
			v = settings.Value{Value: st[d.Key].Value, Source: st[d.Key].Source}
		}
		out = append(out, SettingInfo{Key: d.Key, Value: v.Value, Source: v.Source, Annotation: d.Annotation(), Options: d.Options})
	}
	return out
}

func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	var recs v1alpha1.RecommendationList
	if err := s.Client.List(r.Context(), &recs); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	nsAnnotated := map[string]bool{}
	var nss corev1.NamespaceList
	if err := s.Client.List(r.Context(), &nss); err == nil {
		for _, ns := range nss.Items {
			nsAnnotated[ns.Name] = hasSettings(ns.Annotations)
		}
	}
	out := make([]Summary, 0, len(recs.Items))
	for i := range recs.Items {
		sum := summarize(&recs.Items[i], s.replicas(r.Context(), &recs.Items[i]))
		sum.NamespaceSet = nsAnnotated[sum.Namespace]
		out = append(out, sum)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Namespace+out[i].Name < out[j].Namespace+out[j].Name })
	writeJSON(w, out)
}

func hasSettings(ann map[string]string) bool {
	for k := range ann {
		if strings.HasPrefix(k, settings.Domain+"/") {
			return true
		}
	}
	return false
}

func (s *Server) replicas(ctx context.Context, rec *v1alpha1.Recommendation) int32 {
	obj, err := workload.NewObject(rec.Spec.TargetRef.Kind)
	if err != nil || s.Client.Get(ctx, client.ObjectKey{Namespace: rec.Namespace, Name: rec.Spec.TargetRef.Name}, obj) != nil {
		return 0
	}
	wl, _ := workload.FromObject(obj)
	return wl.Replicas
}

func summarize(rec *v1alpha1.Recommendation, replicas int32) Summary {
	st := rec.Status
	sum := Summary{
		Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name,
		Replicas: replicas, Phase: string(st.Phase), Message: st.Message, Pending: st.Pending,
		AutoDeploy: st.Settings[settings.AutoDeploy].Value == "Enabled", OOMKills: st.OOMKills, Hash: st.Hash,
	}
	if sum.Phase == "" {
		sum.Phase = string(v1alpha1.PhasePending)
		sum.Message = "Waiting for the first recommendation."
	}
	if st.Score != nil {
		sum.Score = &Score{Overall: st.Score.Overall, CPU: st.Score.CPU, Memory: st.Score.Memory}
	}
	if st.Impact != nil {
		sum.AvgReplicas, _ = strconv.ParseFloat(st.Impact.AvgReplicas, 64)
		sum.Current = totals(st.Impact.Current)
		sum.Recommended = totals(st.Impact.Recommended)
	}
	if st.GeneratedAt != nil {
		sum.GeneratedAt = st.GeneratedAt.UTC().Format(time.RFC3339)
	}
	for _, h := range st.History {
		if h.Action == v1alpha1.ActionApply {
			sum.LastApplied = h.Time.UTC().Format(time.RFC3339)
			break
		}
	}
	for _, c := range st.Containers {
		if c.Skipped {
			continue
		}
		cur := workload.ToValues(c.Current.Requests)
		rec := workload.ToValues(c.Recommended.Requests)
		if len(cur) == 0 {
			sum.NoRequests = true
		}
		for _, r := range engine.Resources {
			if rec[r] > cur[r]*1.1 {
				sum.Under = true
			}
		}
	}
	if st.HPA != nil {
		sum.HPA = []string{"cpu"}
	}
	if st.Phase == v1alpha1.PhaseUnsupported {
		sum.HPA = []string{"cpu", "memory"}
	}
	return sum
}

func totals(t v1alpha1.Totals) Totals {
	cpu, _ := strconv.ParseFloat(t.CPU, 64)
	mem, _ := strconv.ParseFloat(t.Memory, 64)
	cost, _ := strconv.ParseFloat(t.MonthlyCost, 64)
	return Totals{CPU: cpu, Memory: mem, MonthlyCost: cost}
}

func (s *Server) get(w http.ResponseWriter, r *http.Request) (*v1alpha1.Recommendation, bool) {
	kind := r.PathValue("kind")
	if !slices.Contains(workload.Kinds, kind) {
		writeError(w, http.StatusBadRequest, "unsupported kind "+kind)
		return nil, false
	}
	rec := &v1alpha1.Recommendation{}
	key := client.ObjectKey{Namespace: r.PathValue("ns"), Name: workload.ResourceName(kind, r.PathValue("name"))}
	if err := s.Client.Get(r.Context(), key, rec); err != nil {
		if apierrors.IsNotFound(err) {
			writeError(w, http.StatusNotFound, "workload is not managed by resourcerer")
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return nil, false
	}
	return rec, true
}

func (s *Server) detail(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	st := rec.Status
	d := Detail{Summary: summarize(rec, s.replicas(r.Context(), rec)), Warnings: st.Warnings, History: []History{}, Events: []Event{}}
	d.ApplyMethod = st.Settings[settings.ApplyMethod].Value
	var ns corev1.Namespace
	if s.Client.Get(r.Context(), client.ObjectKey{Name: rec.Namespace}, &ns) == nil {
		d.NamespaceSet = hasSettings(ns.Annotations)
	}
	if st.NextScheduleAt != nil {
		d.NextScheduleAt = st.NextScheduleAt.UTC().Format(time.RFC3339)
	}
	for _, c := range st.Containers {
		d.Containers = append(d.Containers, Container{
			Name: c.Name, Skipped: c.Skipped, Basis: c.Basis,
			CurrentRequests: res(c.Current.Requests), CurrentLimits: res(c.Current.Limits),
			RecommendedRequest: res(c.Recommended.Requests), RecommendedLimits: res(c.Recommended.Limits),
		})
	}
	if h := st.HPA; h != nil {
		d.HPARec = &HPA{Name: h.Name, Current: h.Current, Recommended: h.Recommended, Clamped: h.Clamped}
	}
	if st.Settings != nil {
		d.Settings = settingInfos(settings.Effective{}, st.Settings)
	}
	for _, h := range st.History {
		e := History{Time: h.Time.UTC().Format(time.RFC3339), Action: string(h.Action), By: h.By, Hash: h.Hash}
		for _, c := range h.Containers {
			e.Containers = append(e.Containers, Change{Name: c.Name,
				FromRequests: res(c.From.Requests), ToRequests: res(c.To.Requests), FromLimits: res(c.From.Limits), ToLimits: res(c.To.Limits)})
		}
		d.History = append(d.History, e)
	}
	if a := st.LastApplyResult; a != nil {
		d.LastApplyResult = &ApplyResult{Time: a.Time.UTC().Format(time.RFC3339), Result: a.Result, Message: a.Message}
	}
	d.Events = s.events(r.Context(), rec)
	writeJSON(w, d)
}

func res(l corev1.ResourceList) Resources { return Resources(workload.ToValues(l)) }

func (s *Server) events(ctx context.Context, rec *v1alpha1.Recommendation) []Event {
	var l corev1.EventList
	sel := fields.SelectorFromSet(fields.Set{"involvedObject.name": rec.Spec.TargetRef.Name, "involvedObject.kind": rec.Spec.TargetRef.Kind})
	if err := s.APIReader.List(ctx, &l, client.InNamespace(rec.Namespace), client.MatchingFieldsSelector{Selector: sel}); err != nil {
		log.FromContext(ctx).Error(err, "list events")
		return []Event{}
	}
	out := make([]Event, 0, len(l.Items))
	for _, e := range l.Items {
		t := e.EventTime.Time
		if t.IsZero() {
			t = e.LastTimestamp.Time
		}
		if t.IsZero() {
			t = e.CreationTimestamp.Time
		}
		out = append(out, Event{Type: e.Type, Reason: e.Reason, Message: e.Message, Time: t.UTC().Format(time.RFC3339)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Time > out[j].Time })
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

var windows = map[string]time.Duration{"1h": time.Hour, "1d": 24 * time.Hour, "1w": 7 * 24 * time.Hour, "2w": 14 * 24 * time.Hour, "1m": 30 * 24 * time.Hour}

const buckets = 150

func (s *Server) window(r *http.Request) (time.Time, time.Time, time.Duration) {
	win, ok := windows[r.URL.Query().Get("range")]
	if !ok {
		win = windows["1w"]
	}
	end := time.Now().Truncate(s.MinStep)
	return end.Add(-win), end, win
}

// series charts one container: usage percentiles per bucket plus current and recommended requests and limits.
func (s *Server) series(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	container, resource := r.URL.Query().Get("container"), r.URL.Query().Get("resource")
	if resource != engine.CPU && resource != engine.Memory {
		writeError(w, http.StatusBadRequest, "resource must be cpu or memory")
		return
	}
	start, end, win := s.window(r)
	ref := prom.Ref{Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name}
	key := strings.Join([]string{"series", rec.Namespace, ref.Kind, ref.Name, container, resource, r.URL.Query().Get("range"), end.String()}, "|")
	v := s.cache.get(key, 30*time.Second, func() (any, error) {
		ctx := r.Context()
		raw, err := s.Prom.Range(ctx, s.Prom.BusiestPod(ref, resource), start, end, prom.Step(win, s.MinStep), "container")
		if err != nil {
			return nil, err
		}
		lineStep := max(win/300, s.MinStep)
		lines := map[string][]prom.Point{}
		for name, q := range map[string]string{
			"currentRequests":     s.Prom.Requests(ref, "requests", resource),
			"currentLimits":       s.Prom.Requests(ref, "limits", resource),
			"recommendedRequests": prom.Recommended(ref, "request", resource),
			"recommendedLimits":   prom.Recommended(ref, "limit", resource),
		} {
			m, err := s.Prom.Range(ctx, q, start, end, lineStep, "container")
			if err != nil {
				return nil, err
			}
			lines[name] = nonNil(m[container])
		}
		return Series{From: float64(start.Unix()), To: float64(end.Unix()), Bands: bands(raw[container], start, end), Lines: lines}, nil
	})
	if err, isErr := v.(error); isErr {
		writeError(w, http.StatusBadGateway, "query Prometheus: "+err.Error())
		return
	}
	writeJSON(w, v)
}

// impact charts the whole workload: total usage across pods, total requests now, and requests with the recommendation.
func (s *Server) impact(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	resource := r.URL.Query().Get("resource")
	if resource != engine.CPU && resource != engine.Memory {
		writeError(w, http.StatusBadRequest, "resource must be cpu or memory")
		return
	}
	start, end, win := s.window(r)
	ref := prom.Ref{Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name}
	key := strings.Join([]string{"impact", rec.Namespace, ref.Kind, ref.Name, resource, r.URL.Query().Get("range"), end.String()}, "|")
	v := s.cache.get(key, 30*time.Second, func() (any, error) {
		ctx := r.Context()
		raw, err := s.Prom.Range(ctx, s.Prom.Total(ref, resource), start, end, prom.Step(win, s.MinStep), "")
		if err != nil {
			return nil, err
		}
		lineStep := max(win/300, s.MinStep)
		requested, err := s.Prom.Range(ctx, "sum("+s.Prom.Requests(ref, "requests", resource)+")", start, end, lineStep, "")
		if err != nil {
			return nil, err
		}
		pods, err := s.Prom.Range(ctx, s.Prom.Pods(ref), start, end, lineStep, "")
		if err != nil {
			return nil, err
		}
		recSum := 0.0
		for _, c := range rec.Status.Containers {
			cur := workload.ToValues(c.Current.Requests)
			val, ok := workload.ToValues(c.Recommended.Requests)[resource]
			if !ok {
				val = cur[resource]
			}
			recSum += val
		}
		withRec := make([]prom.Point, 0, len(pods[""]))
		for _, p := range pods[""] {
			withRec = append(withRec, prom.Point{T: p.T, V: p.V * recSum})
		}
		return Series{From: float64(start.Unix()), To: float64(end.Unix()), Bands: bands(raw[""], start, end),
			Lines: map[string][]prom.Point{"requested": nonNil(requested[""]), "withRecommendation": withRec}}, nil
	})
	if err, isErr := v.(error); isErr {
		writeError(w, http.StatusBadGateway, "query Prometheus: "+err.Error())
		return
	}
	writeJSON(w, v)
}

func nonNil(p []prom.Point) []prom.Point {
	if p == nil {
		return []prom.Point{}
	}
	return p
}

// bands turns raw samples into p50/p75/p95 per bucket, so a month of samples stays readable.
func bands(pts []prom.Point, start, end time.Time) []Band {
	out := []Band{}
	size := float64(end.Sub(start).Seconds()) / buckets
	from := float64(start.Unix())
	groups := make([][]float64, buckets)
	for _, p := range pts {
		i := int((p.T - from) / size)
		if i >= 0 && i < buckets {
			groups[i] = append(groups[i], p.V)
		}
	}
	for i, g := range groups {
		if len(g) == 0 {
			continue
		}
		out = append(out, Band{T: from + (float64(i)+0.5)*size,
			P50: engine.Quantile(g, 0.5), P75: engine.Quantile(g, 0.75), P95: engine.Quantile(g, 0.95)})
	}
	return out
}

/* ── Mutations ─────────────────────────────────────────────────────────── */

func (s *Server) write(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.ReadOnly {
			writeError(w, http.StatusForbidden, "resourcerer is read-only")
			return
		}
		h(w, r)
	}
}

func (s *Server) patchSpec(w http.ResponseWriter, r *http.Request, mutate func(*v1alpha1.Recommendation) error) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	patch := client.MergeFrom(rec.DeepCopy())
	if err := mutate(rec); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err := s.Client.Patch(r.Context(), rec, patch); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "accepted"})
}

func (s *Server) apply(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Hash string `json:"hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Hash == "" {
		writeError(w, http.StatusBadRequest, "body must be {\"hash\": \"<recommendation hash>\"}")
		return
	}
	s.patchSpec(w, r, func(rec *v1alpha1.Recommendation) error {
		if rec.Status.Hash != body.Hash {
			return errors.New("the recommendation changed since it was displayed; reload and review it again")
		}
		rec.Spec.ApplyRequest = &v1alpha1.ApplyRequest{Action: v1alpha1.ActionApply, Hash: body.Hash, RequestedBy: User, RequestedAt: metav1.NewTime(time.Now())}
		return nil
	})
}

func (s *Server) revert(w http.ResponseWriter, r *http.Request) {
	s.patchSpec(w, r, func(rec *v1alpha1.Recommendation) error {
		rec.Spec.ApplyRequest = &v1alpha1.ApplyRequest{Action: v1alpha1.ActionRevert, RequestedBy: User, RequestedAt: metav1.NewTime(time.Now())}
		return nil
	})
}

func (s *Server) regenerate(w http.ResponseWriter, r *http.Request) {
	s.patchSpec(w, r, func(rec *v1alpha1.Recommendation) error {
		now := metav1.NewTime(time.Now())
		rec.Spec.RegenerateRequestedAt = &now
		return nil
	})
}

type settingBody struct {
	Key string `json:"key"`
	// Value nil removes the annotation, so the setting is inherited again.
	Value *string `json:"value"`
}

func decodeSetting(w http.ResponseWriter, r *http.Request) (settings.Definition, *string, bool) {
	var b settingBody
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return settings.Definition{}, nil, false
	}
	d, ok := settings.Lookup(b.Key)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown setting "+b.Key)
		return d, nil, false
	}
	if b.Value != nil {
		if err := d.Check(*b.Value); err != nil {
			writeError(w, http.StatusBadRequest, d.Key+": "+err.Error())
			return d, nil, false
		}
	}
	return d, b.Value, true
}

func annotationPatch(key string, value *string) client.Patch {
	body, _ := json.Marshal(map[string]any{"metadata": map[string]any{"annotations": map[string]any{key: value}}})
	return client.RawPatch(types.MergePatchType, body)
}

func (s *Server) workloadSetting(w http.ResponseWriter, r *http.Request) {
	d, value, ok := decodeSetting(w, r)
	if !ok {
		return
	}
	obj, err := workload.NewObject(r.PathValue("kind"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	obj.SetNamespace(r.PathValue("ns"))
	obj.SetName(r.PathValue("name"))
	if err := s.Client.Patch(r.Context(), obj, annotationPatch(d.Annotation(), value)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}

func (s *Server) namespaceSetting(w http.ResponseWriter, r *http.Request) {
	d, value, ok := decodeSetting(w, r)
	if !ok {
		return
	}
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: r.PathValue("ns")}}
	if err := s.Client.Patch(r.Context(), ns, annotationPatch(d.Annotation(), value)); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]string{"status": "saved"})
}

/* ── Plumbing ──────────────────────────────────────────────────────────── */

// cache keeps Prometheus results briefly and collapses concurrent identical queries into one.
type cache struct {
	mu      sync.Mutex
	entries map[string]*entry
}

type entry struct {
	done    chan struct{}
	value   any
	expires time.Time
}

func (c *cache) get(key string, ttl time.Duration, fetch func() (any, error)) any {
	c.mu.Lock()
	if c.entries == nil {
		c.entries = map[string]*entry{}
	}
	now := time.Now()
	for k, e := range c.entries {
		if !e.expires.IsZero() && now.After(e.expires) {
			delete(c.entries, k)
		}
	}
	if e, ok := c.entries[key]; ok {
		c.mu.Unlock()
		<-e.done
		return e.value
	}
	e := &entry{done: make(chan struct{})}
	c.entries[key] = e
	c.mu.Unlock()

	v, err := fetch()
	c.mu.Lock()
	if err != nil {
		e.value = err
		delete(c.entries, key)
	} else {
		e.value = v
		e.expires = time.Now().Add(ttl)
	}
	c.mu.Unlock()
	close(e.done)
	return e.value
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func logRequests(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Log.WithName("api").V(1).Info("request", "method", r.Method, "path", r.URL.Path, "duration", time.Since(start).String())
		}
	})
}

// spa serves built assets and falls back to index.html for client-side routes.
func spa(ui fs.FS) http.Handler {
	files := http.FileServer(http.FS(ui))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p != "" {
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
		r.URL.Path = "/"
		files.ServeHTTP(w, r)
	})
}
