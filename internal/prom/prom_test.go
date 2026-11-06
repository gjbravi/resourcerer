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

func TestNodeAllocationQueries(t *testing.T) {
	nodes := KarpenterNodes("memory", 7*24*time.Hour)
	for _, want := range []string{
		`max_over_time(karpenter_nodes_allocatable{resource_type="memory"}[1w])`,
		`topk by (node_name) (1,`,
		`max by (node_name, instance_type, zone, capacity_type)`,
	} {
		if !strings.Contains(nodes, want) {
			t.Errorf("nodes query lacks %s:\n%s", want, nodes)
		}
	}
	prices := KarpenterOfferingPrices([]string{"m5.large", "c7i-flex.xlarge"})
	if !strings.Contains(prices, `offering_price_estimate{instance_type=~"m5\\.large|c7i-flex\\.xlarge"}[5m]`) {
		t.Errorf("prices must read only the nodes' instance types, over the last minutes:\n%s", prices)
	}
	if all := KarpenterOfferingPrices(nil); !strings.Contains(all, "offering_price_estimate[5m]") {
		t.Errorf("without types every offering is read:\n%s", all)
	}
	w := WorkloadRequestsByNode(Ref{Namespace: "demo", Kind: "StatefulSet", Name: "cache"}, 0.0316, 0.0042)
	for _, want := range []string{
		`sum by (node) (`,
		`kube_pod_container_resource_requests{namespace="demo", resource="cpu"} * 0.0316`,
		`topk by (namespace, pod) (1,`,
		`owner_kind="StatefulSet", owner_name="cache"`,
	} {
		if !strings.Contains(w, want) {
			t.Errorf("requests by node query lacks %s:\n%s", want, w)
		}
	}
	if strings.Contains(w, "karpenter") {
		t.Errorf("a workload's query must not read Karpenter's series:\n%s", w)
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
