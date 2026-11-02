package prom

import (
	"strings"
	"testing"
	"time"
)

func TestOwnerJoin(t *testing.T) {
	d := owner(Ref{Namespace: "demo", Kind: "Deployment", Name: "api"})
	for _, want := range []string{`kube_replicaset_owner{namespace="demo", owner_kind="Deployment", owner_name="api"}`, `owner_kind="ReplicaSet"`} {
		if !strings.Contains(d, want) {
			t.Errorf("deployment owner query lacks %s:\n%s", want, d)
		}
	}
	s := owner(Ref{Namespace: "demo", Kind: "StatefulSet", Name: `ca"che`})
	if !strings.Contains(s, `owner_kind="StatefulSet", owner_name="ca\"che"`) {
		t.Errorf("statefulset owner query not quoted: %s", s)
	}
}

func TestStepStaysUnderPointLimit(t *testing.T) {
	if s := Step(30*24*time.Hour, 15*time.Second); 30*24*time.Hour/s > 11000 {
		t.Errorf("step %v gives too many points", s)
	}
	if s := Step(15*time.Minute, 15*time.Second); s != 15*time.Second {
		t.Errorf("short windows use the minimum step, got %v", s)
	}
}

func TestRolloutOwnerUsesReplicaSets(t *testing.T) {
	q := owner(Ref{Namespace: "demo", Kind: "Rollout", Name: "shop"})
	if !strings.Contains(q, `kube_replicaset_owner{namespace="demo", owner_kind="Rollout", owner_name="shop"}`) {
		t.Errorf("rollout owner query:\n%s", q)
	}
}

func TestWorkloadNodeCost(t *testing.T) {
	q := WorkloadNodeCost(Ref{Namespace: "demo", Kind: "StatefulSet", Name: "cache"}, 0.0316, 0.0042, false, []string{"m5.large", "c7i-flex.xlarge"})
	for _, want := range []string{
		`offering_price_estimate{instance_type=~"m5\\.large|c7i-flex\\.xlarge"}[5m]`,
		`kube_pod_container_resource_requests{namespace="demo", resource="cpu"} * 0.0316`,
		`topk by (namespace, pod) (1,`,
		`topk by (node_name) (1,`,
		`"price_type", "on-demand", "capacity_type", "reserved"`,
		`owner_kind="StatefulSet", owner_name="cache"`,
	} {
		if !strings.Contains(q, want) {
			t.Errorf("node cost query lacks %s:\n%s", want, q)
		}
	}
	if p := WorkloadNodeCost(Ref{Namespace: "demo", Kind: "StatefulSet", Name: "cache"}, 0.0316, 0.0042, true, nil); !strings.Contains(p, "* 0 + 1)") || !strings.Contains(p, "offering_price_estimate[5m]") {
		t.Errorf("priced query must leave the factor out, and read every offering without types:\n%s", p)
	}
}

func TestClusterUsageCoversTheStep(t *testing.T) {
	c := &Client{RateWindow: "5m"}
	if q := c.ClusterUsage("cpu", time.Minute); !strings.Contains(q, "[5m]") {
		t.Errorf("a short step keeps the rate window:\n%s", q)
	}
	if q := c.ClusterUsage("cpu", 2*time.Hour); !strings.Contains(q, "[2h]") {
		t.Errorf("a long step rates over the whole step:\n%s", q)
	}
	if q := c.ClusterUsage("memory", time.Minute); strings.Contains(q, "avg_over_time") {
		t.Errorf("a short step reads memory as is:\n%s", q)
	}
	if q := c.ClusterUsage("memory", 2*time.Hour); !strings.Contains(q, "avg_over_time(container_memory_working_set_bytes{container!=\"\", container!=\"POD\"}[2h])") {
		t.Errorf("a long step averages memory over the step:\n%s", q)
	}
}
