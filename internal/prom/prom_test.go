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
