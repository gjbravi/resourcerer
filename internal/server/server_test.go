package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gjbravi/resourcerer/internal/config"
	"github.com/gjbravi/resourcerer/internal/scheme"
)

func fixture(t *testing.T, readOnly bool) (*Server, client.Client) {
	t.Helper()
	ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "shop"}}
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "shop"},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
			Name:      "api",
			Resources: corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1"), corev1.ResourceMemory: resource.MustParse("1Gi")}},
		}}}}},
	}
	target := corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("126m"), corev1.ResourceMemory: resource.MustParse("77594624")}
	v := &vpav1.VerticalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{Name: "deployment-api", Namespace: "shop", Labels: map[string]string{config.LabelManaged: "true"}},
		Status: vpav1.VerticalPodAutoscalerStatus{
			Conditions: []vpav1.VerticalPodAutoscalerCondition{{Type: vpav1.RecommendationProvided, Status: corev1.ConditionTrue}},
			Recommendation: &vpav1.RecommendedPodResources{ContainerRecommendations: []vpav1.RecommendedContainerResources{
				{ContainerName: "api", Target: target, LowerBound: target, UpperBound: target, UncappedTarget: target},
			}},
		},
	}
	c := fake.NewClientBuilder().WithScheme(scheme.New()).WithObjects(ns, d, v).Build()
	return &Server{
		Client:   c,
		Options:  config.Options{Selection: config.SelectionAll, DefaultMode: config.ModeOff, ExcludedNamespaces: []string{"kube-system"}},
		ReadOnly: readOnly,
	}, c
}

func do(t *testing.T, s *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	rr := httptest.NewRecorder()
	s.Handler().ServeHTTP(rr, httptest.NewRequest(method, path, strings.NewReader(body)))
	return rr
}

func TestList(t *testing.T) {
	s, _ := fixture(t, false)
	rr := do(t, s, http.MethodGet, "/api/v1/workloads", "")
	var out []Workload
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil || len(out) != 1 {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body)
	}
	w := out[0]
	if w.Mode != ModeOff || w.Status != StatusReady || w.VPA != "deployment-api" {
		t.Errorf("workload = %+v", w)
	}
	c := w.Containers[0]
	if c.Target == nil || c.Target.CPU != "126m" || c.Target.Memory != "74Mi" || c.Requests.Memory != "1Gi" {
		t.Errorf("container = %+v target=%+v", c, c.Target)
	}
}

func TestSetMode(t *testing.T) {
	s, c := fixture(t, false)
	rr := do(t, s, http.MethodPut, "/api/v1/workloads/shop/Deployment/api/mode", `{"mode":"Auto"}`)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body)
	}
	var d appsv1.Deployment
	_ = c.Get(t.Context(), client.ObjectKey{Namespace: "shop", Name: "api"}, &d)
	if d.Annotations[config.AnnotationMode] != "Auto" || d.Labels[config.LabelEnabled] != "true" {
		t.Errorf("labels=%v annotations=%v", d.Labels, d.Annotations)
	}

	do(t, s, http.MethodPut, "/api/v1/workloads/shop/Deployment/api/mode", `{"mode":"Disabled"}`)
	_ = c.Get(t.Context(), client.ObjectKey{Namespace: "shop", Name: "api"}, &d)
	if d.Labels[config.LabelEnabled] != "false" {
		t.Errorf("labels=%v", d.Labels)
	}

	if rr := do(t, s, http.MethodPut, "/api/v1/workloads/shop/Deployment/api/mode", `{"mode":"Recreate"}`); rr.Code != http.StatusBadRequest {
		t.Errorf("invalid mode: status=%d", rr.Code)
	}
}

func TestReadOnlyAndExcluded(t *testing.T) {
	s, _ := fixture(t, true)
	if rr := do(t, s, http.MethodPut, "/api/v1/workloads/shop/Deployment/api/mode", `{"mode":"Auto"}`); rr.Code != http.StatusForbidden {
		t.Errorf("read-only: status=%d", rr.Code)
	}
	if rr := do(t, s, http.MethodGet, "/api/v1/workloads/kube-system/Deployment/coredns", ""); rr.Code != http.StatusNotFound {
		t.Errorf("excluded: status=%d", rr.Code)
	}
}

func TestDetailListsAreNeverNull(t *testing.T) {
	s, _ := fixture(t, false)
	body := do(t, s, http.MethodGet, "/api/v1/workloads/shop/Deployment/api", "").Body.String()
	for _, k := range []string{`"pods":null`, `"events":null`, `"containers":null`} {
		if strings.Contains(body, k) {
			t.Errorf("%s in %s", k, body)
		}
	}
}

func TestSPA(t *testing.T) {
	h := spa(fstest.MapFS{"index.html": {Data: []byte("<html>")}, "assets/app-1.js": {Data: []byte("js")}})
	cases := []struct {
		path, cache string
		code        int
	}{
		{"/", "no-cache", 200},
		{"/workloads/shop/Deployment/api", "no-cache", 200},
		{"/assets/app-1.js", "public, max-age=31536000, immutable", 200},
		{"/assets/gone.js", "", 404},
	}
	for _, c := range cases {
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, c.path, nil))
		if rr.Code != c.code || rr.Header().Get("Cache-Control") != c.cache {
			t.Errorf("%s: %d %q", c.path, rr.Code, rr.Header().Get("Cache-Control"))
		}
	}
}
