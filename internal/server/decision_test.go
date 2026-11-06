package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/scheme"
	"github.com/gjbravi/resourcerer/internal/server/auth"
)

// fakeProm answers every range query with a flat series for container app.
func fakeProm(t *testing.T) *prom.Client {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		now := time.Now().Unix()
		vals := []string{}
		for i := int64(0); i < 20; i++ {
			vals = append(vals, fmt.Sprintf(`[%d,"0.2"]`, now-600+i*30))
		}
		_, _ = fmt.Fprintf(w, `{"status":"success","data":{"resultType":"matrix","result":[{"metric":{"container":"app"},"values":[%s]}]}}`,
			strings.Join(vals, ","))
	}))
	t.Cleanup(ts.Close)
	p, err := prom.New(ts.URL, "5m", prom.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// fakeDecider answers the verdict with choice, counting the requests and keeping the last state.
func fakeDecider(t *testing.T, choice string, calls *atomic.Int32, state *string) string {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var body map[string]json.RawMessage
		_ = json.NewDecoder(r.Body).Decode(&body)
		*state = string(body["state"])
		_, _ = fmt.Fprintf(w, `{"model":"decider-test","answers":{"verdict":{"choice":%q,"probabilities":{"apply":0.81,"review":0.15,"wait":0.04}}}}`, choice)
	}))
	t.Cleanup(ts.Close)
	return ts.URL + "/v1/systemone"
}

func decisionServer(t *testing.T, pending bool, objs ...client.Object) http.Handler {
	t.Helper()
	gen := metav1.NewTime(time.Now().Add(-time.Minute))
	rec := &v1alpha1.Recommendation{
		ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Name: "deployment-api"},
		Spec:       v1alpha1.RecommendationSpec{TargetRef: v1alpha1.TargetRef{APIVersion: "apps/v1", Kind: "Deployment", Name: "api"}},
		Status: v1alpha1.RecommendationStatus{
			Phase: v1alpha1.PhaseReady, Pending: pending, Hash: "abc", GeneratedAt: &gen,
			Containers: []v1alpha1.ContainerRecommendation{{
				Name:        "app",
				Current:     corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1")}},
				Recommended: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("250m")}},
			}},
			Evidence: &v1alpha1.Evidence{HistorySeconds: 600, LearningSeconds: 600},
		},
	}
	replicas := int32(2)
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "demo", Name: "api"}, Spec: appsv1.DeploymentSpec{Replicas: &replicas}}
	c := fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(append([]client.Object{rec, dep}, objs...)...).Build()
	s := &Server{Client: c, APIReader: c, Prom: fakeProm(t), MinStep: 30 * time.Second, Auth: auth.Anonymous{}}
	return s.Handler()
}

func getDecision(t *testing.T, h http.Handler, query string) DecisionReport {
	t.Helper()
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/workloads/demo/Deployment/api/decision"+query, nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body)
	}
	var out DecisionReport
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return out
}

func TestDecision(t *testing.T) {
	model := func(endpoint string) *v1alpha1.DecisionModel {
		return &v1alpha1.DecisionModel{ObjectMeta: metav1.ObjectMeta{Name: "default"},
			Spec: v1alpha1.DecisionModelSpec{Provider: "decider", Endpoint: endpoint, Model: "decider-test", TimeoutSeconds: 5}}
	}

	t.Run("without a model", func(t *testing.T) {
		got := getDecision(t, decisionServer(t, true), "?evaluate=1")
		if got.Configured || got.Evaluation != nil || got.Reason != "No decision model is configured." {
			t.Errorf("got %+v", got)
		}
	})

	t.Run("nothing to apply", func(t *testing.T) {
		var calls atomic.Int32
		var state string
		got := getDecision(t, decisionServer(t, false, model(fakeDecider(t, "apply", &calls, &state))), "?evaluate=1")
		if !got.Configured || got.Evaluation != nil || got.Reason != "Nothing to apply." || calls.Load() != 0 {
			t.Errorf("got %+v after %d calls", got, calls.Load())
		}
	})

	t.Run("evaluates on request only, once per recommendation", func(t *testing.T) {
		var calls atomic.Int32
		var state string
		h := decisionServer(t, true, model(fakeDecider(t, "apply", &calls, &state)))
		if got := getDecision(t, h, ""); !got.Configured || got.Evaluation != nil || calls.Load() != 0 {
			t.Fatalf("asked without evaluate: %+v after %d calls", got, calls.Load())
		}
		got := getDecision(t, h, "?evaluate=1")
		e := got.Evaluation
		if e == nil || e.Choice != "apply" || e.Hash != "abc" || e.Version != "decider-test" || len(e.Options) != 3 || e.Options[0].Probability != 0.81 {
			t.Fatalf("got %+v", got)
		}
		if !strings.Contains(state, `"workload":"Deployment, 2 replicas now"`) || !strings.Contains(state, `"app cpu":"busiest pod p50 0.2`) {
			t.Errorf("state %s", state)
		}
		if getDecision(t, h, "").Evaluation == nil || getDecision(t, h, "?evaluate=1").Evaluation == nil || calls.Load() != 1 {
			t.Errorf("the evaluation is not kept: %d calls", calls.Load())
		}
		if getDecision(t, h, "?evaluate=1&fresh=1"); calls.Load() != 2 {
			t.Errorf("fresh did not ask again: %d calls", calls.Load())
		}
	})

	t.Run("an answer outside the options is not shown", func(t *testing.T) {
		var calls atomic.Int32
		var state string
		got := getDecision(t, decisionServer(t, true, model(fakeDecider(t, "maybe", &calls, &state))), "?evaluate=1")
		if got.Evaluation != nil || got.Reason != `answer "maybe" to "verdict" is not an option` {
			t.Errorf("got %+v", got)
		}
	})
}
