package server

import (
	"testing"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func pod(name string, qos corev1.PodQOSClass, cs ...corev1.ContainerStatus) corev1.Pod {
	p := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}}
	p.Status.QOSClass = qos
	p.Status.ContainerStatuses = cs
	return p
}

func terminated(reason string, code int32, at time.Time) corev1.ContainerState {
	return corev1.ContainerState{Terminated: &corev1.ContainerStateTerminated{Reason: reason, ExitCode: code, FinishedAt: metav1.NewTime(at)}}
}

func byID(cs []Check) map[string]Check {
	out := map[string]Check{}
	for _, c := range cs {
		out[c.ID] = c
	}
	return out
}

func TestPodChecks(t *testing.T) {
	now := time.Now()
	since := now.Add(-15 * time.Minute)
	pods := []corev1.Pod{
		pod("api-1", corev1.PodQOSBurstable, corev1.ContainerStatus{Name: "api", LastTerminationState: terminated("OOMKilled", 137, now.Add(-5*time.Minute))}),
		pod("api-2", corev1.PodQOSBurstable, corev1.ContainerStatus{
			Name:                 "api",
			State:                corev1.ContainerState{Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"}},
			LastTerminationState: terminated("Error", 1, now.Add(-time.Minute)),
		}),
		pod("api-3", corev1.PodQOSBurstable, corev1.ContainerStatus{Name: "api", LastTerminationState: terminated("OOMKilled", 137, now.Add(-time.Hour))}),
	}
	got := byID(podChecks(pods, since, map[string]float64{"api": 2}, "q"))
	if c := got["oom"]; c.Status != "bad" || c.Value != 1 || c.Pod != "api-1" {
		t.Errorf("oom = %+v, want 1 kill in the window, in api-1", c)
	}
	if c := got["restarts"]; c.Status != "warn" || c.Value != 2 || c.Pod != "api-2" {
		t.Errorf("restarts = %+v, want 2 with the last exit from api-2", c)
	}
	if c := got["crashloop"]; c.Status != "bad" || c.Container != "api" {
		t.Errorf("crashloop = %+v", c)
	}

	quiet := byID(podChecks(pods[2:], since, map[string]float64{}, "q"))
	for _, id := range []string{"oom", "restarts", "crashloop"} {
		if quiet[id].Status != "ok" {
			t.Errorf("%s = %s, want ok when the only kill is older than the window", id, quiet[id].Status)
		}
	}
}

func TestHPAAndQoSChecks(t *testing.T) {
	h := &autoscalingv2.HorizontalPodAutoscaler{ObjectMeta: metav1.ObjectMeta{Name: "web"}}
	h.Spec.MaxReplicas, h.Status.CurrentReplicas = 3, 3
	if c := hpaCheck(h); c.Status != "warn" || c.Display != "3 / 3 replicas" {
		t.Errorf("hpa at max = %+v", c)
	}
	h.Status.CurrentReplicas = 2
	if c := hpaCheck(h); c.Status != "ok" {
		t.Errorf("hpa with headroom = %+v", c)
	}
	if c := hpaCheck(nil); c.Status != "ok" || c.Display != "no HPA" {
		t.Errorf("no hpa = %+v", c)
	}
	if c := qosCheck([]corev1.Pod{pod("a", corev1.PodQOSBestEffort), pod("b", corev1.PodQOSBurstable)}); c.Status != "warn" || c.Display != "BestEffort" {
		t.Errorf("qos = %+v", c)
	}
}
