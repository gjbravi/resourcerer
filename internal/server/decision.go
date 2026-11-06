package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/decision"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// decisionCache keeps an evaluation for the recommendation it was made for; a regeneration changes the key.
const decisionCache = time.Hour

// decisionModel reads the DecisionModel without the cache, as the controller does not watch it; nil without one.
func (s *Server) decisionModel(ctx context.Context) (*v1alpha1.DecisionModel, error) {
	var dm v1alpha1.DecisionModel
	if err := s.APIReader.Get(ctx, client.ObjectKey{Name: v1alpha1.DecisionModelDefault}, &dm); err != nil {
		if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read the DecisionModel: %w", err)
	}
	return &dm, nil
}

// decision is the decision model's opinion on whether the recommendation is safe to apply. It only helps the user
// decide: nothing is applied or stored. Without evaluate it returns the last evaluation of this recommendation, if
// any; with it, it asks the model, once per recommendation unless fresh asks again.
func (s *Server) decision(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	dm, err := s.decisionModel(ctx)
	switch {
	case err != nil:
		writeJSON(w, DecisionReport{Reason: err.Error()})
		return
	case dm == nil:
		writeJSON(w, DecisionReport{Reason: "No decision model is configured."})
		return
	}
	out := DecisionReport{Configured: true, Model: dm.Spec.Model}
	st := rec.Status
	if !st.Pending || (st.Phase != v1alpha1.PhaseReady && st.Phase != v1alpha1.PhaseLearning) || st.GeneratedAt == nil {
		out.Reason = "Nothing to apply."
		writeJSON(w, out)
		return
	}
	key := strings.Join([]string{"decision", rec.Namespace, rec.Name, st.Hash, st.GeneratedAt.UTC().String(),
		dm.Spec.Endpoint, dm.Spec.Model}, "|")
	q := r.URL.Query()
	if q.Get("evaluate") == "" {
		if v, ok := s.cache.peek(key); ok {
			if e, isEval := v.(*Evaluation); isEval {
				out.Evaluation = e
			}
		}
		writeJSON(w, out)
		return
	}
	if q.Get("fresh") != "" {
		s.cache.drop(key)
	}
	v := s.cache.get(key, decisionCache, func() (any, error) { return s.evaluate(ctx, rec, dm) })
	switch e := v.(type) {
	case *Evaluation:
		out.Evaluation = e
	case error:
		out.Reason = e.Error()
	}
	writeJSON(w, out)
}

// evaluate asks the model about the recommendation, from its status and the usage history it was computed from.
func (s *Server) evaluate(ctx context.Context, rec *v1alpha1.Recommendation, dm *v1alpha1.DecisionModel) (*Evaluation, error) {
	st := rec.Status
	learning := s.learningOf(ctx, rec)
	end := st.GeneratedAt.Time
	step := prom.Step(learning, s.MinStep)
	ref := prom.Ref{Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name}
	cpu, err := s.Prom.Range(ctx, s.Prom.BusiestPod(ref, "cpu"), end.Add(-learning), end, step, "container")
	if err != nil {
		return nil, fmt.Errorf("read CPU usage: %w", err)
	}
	mem, err := s.Prom.Range(ctx, s.Prom.BusiestPod(ref, "memory"), end.Add(-learning), end, step, "container")
	if err != nil {
		return nil, fmt.Errorf("read memory usage: %w", err)
	}
	usage := map[string]decision.Usage{}
	for _, c := range st.Containers {
		usage[c.Name] = decision.Usage{CPU: cpu[c.Name], Memory: mem[c.Name]}
	}
	now := time.Now()
	req, _ := decision.Build(decision.Input{Kind: rec.Spec.TargetRef.Kind, Replicas: s.replicas(ctx, rec), Status: st, Usage: usage, Now: now})
	req.Model = dm.Spec.Model

	token := ""
	if ref := dm.Spec.CredentialsSecretRef; ref != nil {
		var sec corev1.Secret
		if err := s.APIReader.Get(ctx, client.ObjectKey{Namespace: s.Cluster.DefaultsConfigMap.Namespace, Name: ref.Name}, &sec); err != nil {
			return nil, fmt.Errorf("read the decision model's credentials: %w", err)
		}
		v, ok := sec.Data[ref.Key]
		if !ok {
			return nil, fmt.Errorf("secret %s has no key %s", ref.Name, ref.Key)
		}
		token = string(v)
	}
	timeout := time.Duration(max(dm.Spec.TimeoutSeconds, 1)) * time.Second
	asked := time.Now()
	resp, err := decision.Ask(ctx, http.DefaultClient, dm.Spec.Endpoint, token, timeout, req)
	decisionSeconds.Observe(time.Since(asked).Seconds())
	if err != nil {
		decisionCounter.WithLabelValues("error").Inc()
		return nil, err
	}
	a, err := decision.Verdict(req, resp)
	if err != nil {
		decisionCounter.WithLabelValues("invalid").Inc()
		return nil, err
	}
	decisionCounter.WithLabelValues("ok").Inc()
	return &Evaluation{Version: resp.Model, EvaluatedAt: now.UTC().Format(time.RFC3339), Hash: st.Hash, Choice: a.Choice, Options: a.Options}, nil
}

// learningOf is the learning period the recommendation was computed over.
func (s *Server) learningOf(ctx context.Context, rec *v1alpha1.Recommendation) time.Duration {
	if ev := rec.Status.Evidence; ev != nil && ev.LearningSeconds > 0 {
		return time.Duration(ev.LearningSeconds) * time.Second
	}
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	return s.learning(r, rec)
}
