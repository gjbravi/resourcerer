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
	q := WorkloadNodeCost(Ref{Namespace: "demo", Kind: "StatefulSet", Name: "cache"}, 0.0316, 0.0042, false)
	for _, want := range []string{
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
	if p := WorkloadNodeCost(Ref{Namespace: "demo", Kind: "StatefulSet", Name: "cache"}, 0.0316, 0.0042, true); !strings.Contains(p, "* 0 + 1)") {
		t.Errorf("priced query must leave the factor out:\n%s", p)
	}
}
