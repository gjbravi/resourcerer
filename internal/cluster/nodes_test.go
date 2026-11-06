package cluster

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

func sample(v float64, kv ...string) prom.Sample {
	l := map[string]string{}
	for i := 0; i < len(kv); i += 2 {
		l[kv[i]] = kv[i+1]
	}
	return prom.Sample{Labels: l, Value: v}
}

func node(name, typ, zone, capacity string, v float64) prom.Sample {
	return sample(v, "node_name", name, "instance_type", typ, "zone", zone, "capacity_type", capacity)
}

func TestNodePrices(t *testing.T) {
	// 4 cores and 16 GiB cost 4×0.0316 + 16×0.0042 = 0.1936 an hour at the configured rates.
	value := 4*static.CPUCoreHour + 16*static.MemoryGiBHour
	cpu := []prom.Sample{
		node("spot-1", "t3.xlarge", "us-east-1a", "spot", 4),
		node("od-1", "m7g.xlarge", "us-east-1a", "on-demand", 4),
		node("ri-1", "m7g.xlarge", "us-east-1a", "reserved", 4),
		node("unknown-1", "x9.huge", "us-east-1a", "spot", 4),
		node("empty-1", "t3.xlarge", "us-east-1a", "spot", 0),
	}
	mem := []prom.Sample{
		node("spot-1", "", "", "", 16*engine.Gi), node("od-1", "", "", "", 16*engine.Gi), node("ri-1", "", "", "", 16*engine.Gi),
		node("unknown-1", "", "", "", 16*engine.Gi),
	}
	prices := []prom.Sample{
		sample(0.05, "instance_type", "t3.xlarge", "zone", "us-east-1a", "capacity_type", "spot"),
		sample(0.17, "instance_type", "m7g.xlarge", "zone", "us-east-1a", "capacity_type", "on-demand"),
		sample(0.03, "instance_type", "m7g.xlarge", "zone", "us-east-1b", "capacity_type", "on-demand"),
	}
	got := nodePrices(cpu, mem, prices, static)
	for name, want := range map[string]NodePrice{
		"spot-1": {Capacity: "spot", Type: "t3.xlarge", Factor: 0.05 / value},
		"od-1":   {Capacity: "on-demand", Type: "m7g.xlarge", Factor: 0.17 / value},
		"ri-1":   {Capacity: "reserved", Type: "m7g.xlarge", Factor: 0.17 / value},
	} {
		g, ok := got[name]
		if !ok || g.Capacity != want.Capacity || g.Type != want.Type || !near(g.Factor, want.Factor) {
			t.Errorf("%s = %+v, want %+v", name, g, want)
		}
	}
	for _, name := range []string{"unknown-1", "empty-1"} {
		if _, ok := got[name]; ok {
			t.Errorf("%s has no offering price or no capacity and must stay unpriced", name)
		}
	}
}

func TestSplit(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	nodes := map[string]NodePrice{
		"spot-1": {Capacity: "spot", Type: "t3.xlarge", Factor: 0.25},
		"spot-2": {Capacity: "spot", Type: "t3.xlarge", Factor: 0.5},
		"od-1":   {Capacity: "on-demand", Type: "m7g.xlarge", Factor: 1},
	}
	byNode := map[string][]prom.Point{
		"spot-1":  series(start, time.Hour, 2, 2, -1),
		"spot-2":  series(start, time.Hour, -1, 2, 2),
		"od-1":    series(start, time.Hour, 1, -1, -1),
		"fargate": series(start, time.Hour, -1, -1, 1),
	}
	cost, priced, total := split(byNode, nodes)

	if got := sum(total[""]); !near(got, 2+2+2+2+1+1) {
		t.Errorf("total = %v, want every request on any node, 10", got)
	}
	if got := sum(priced["spot|t3.xlarge"]); !near(got, 8) {
		t.Errorf("priced spot = %v, want both spot nodes' requests, 8", got)
	}
	if got := sum(cost["spot|t3.xlarge"]); !near(got, 2*0.25+2*0.25+2*0.5+2*0.5) {
		t.Errorf("spot cost = %v, want each node at its own factor, 3", got)
	}
	if got := sum(cost["on-demand|m7g.xlarge"]); !near(got, 1) {
		t.Errorf("on-demand cost = %v, want 1", got)
	}
	if len(total[""]) != 3 || total[""][0].T >= total[""][1].T || !near(total[""][1].V, 4) {
		t.Errorf("total must be one series summed per step, in order: %+v", total[""])
	}
	if _, ok := priced["|"]; ok {
		t.Error("a node without a price must not be keyed as priced")
	}
}

// fakeProm answers the node, price and workload queries of an allocation, and counts each kind.
func fakeProm(t *testing.T, warn bool) (*prom.Client, *atomic.Int32, *atomic.Int32) {
	var nodes, workloads atomic.Int32
	vector := func(w http.ResponseWriter, result []map[string]any) {
		body := map[string]any{"status": "success", "data": map[string]any{"resultType": "vector", "result": result}}
		if warn {
			body["warnings"] = []string{"receive series from store: pool exhausted"}
		}
		_ = json.NewEncoder(w).Encode(body)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		q := r.Form.Get("query")
		w.Header().Set("Content-Type", "application/json")
		now := float64(time.Now().Unix())
		nodeLabels := func(name, typ, capacity string) map[string]any {
			return map[string]any{"node_name": name, "instance_type": typ, "zone": "us-east-1a", "capacity_type": capacity}
		}
		switch {
		case strings.Contains(q, `karpenter_nodes_allocatable{resource_type="cpu"}`):
			nodes.Add(1)
			vector(w, []map[string]any{
				{"metric": nodeLabels("spot-1", "t3.xlarge", "spot"), "value": []any{now, "4"}},
				{"metric": nodeLabels("od-1", "m7g.xlarge", "on-demand"), "value": []any{now, "4"}},
			})
		case strings.Contains(q, `karpenter_nodes_allocatable{resource_type="memory"}`):
			vector(w, []map[string]any{
				{"metric": nodeLabels("spot-1", "", ""), "value": []any{now, "17179869184"}},
				{"metric": nodeLabels("od-1", "", ""), "value": []any{now, "17179869184"}},
			})
		case strings.Contains(q, "offering_price_estimate"):
			vector(w, []map[string]any{
				{"metric": map[string]any{"instance_type": "t3.xlarge", "zone": "us-east-1a", "capacity_type": "spot"}, "value": []any{now, "0.05"}},
				{"metric": map[string]any{"instance_type": "m7g.xlarge", "zone": "us-east-1a", "capacity_type": "on-demand"}, "value": []any{now, "0.17"}},
			})
		case strings.Contains(q, "sum by (node)"):
			workloads.Add(1)
			matrix := func(n string) map[string]any {
				return map[string]any{"metric": map[string]any{"node": n}, "values": [][]any{{now - 60, "1"}, {now, "1"}}}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"status": "success", "data": map[string]any{
				"resultType": "matrix", "result": []map[string]any{matrix("spot-1"), matrix("od-1")}}})
		default:
			t.Errorf("unexpected query %s", q)
		}
	}))
	t.Cleanup(ts.Close)
	c, err := prom.New(ts.URL, "5m", prom.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return c, &nodes, &workloads
}

func TestAllocatePricesOnceForEveryWorkload(t *testing.T) {
	c, nodes, workloads := fakeProm(t, false)
	pr := &Pricer{Prom: c}
	ctx, end := context.Background(), time.Now()
	value := 4*static.CPUCoreHour + 16*static.MemoryGiBHour
	for _, name := range []string{"api", "web", "worker"} {
		a, err := Allocate(ctx, pr, prom.Ref{Namespace: "demo", Kind: "Deployment", Name: name}, static, 0.5, 7*24*time.Hour, end, time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		// Half the requests on each node: the factor is the average of the two nodes' factors.
		if want := (0.05/value + 0.17/value) / 2; !near(a.Factor, want) || !near(a.Coverage, 1) || !near(a.Capacity["spot"], 0.5) {
			t.Errorf("%s: factor %v coverage %v spot %v, want %v 1 0.5", name, a.Factor, a.Coverage, a.Capacity["spot"], want)
		}
	}
	if n := nodes.Load(); n != 1 {
		t.Errorf("the nodes were read %d times for three workloads, want once", n)
	}
	if n := workloads.Load(); n != 3 {
		t.Errorf("workload queries = %d, want one per workload", n)
	}
}

func TestNodePricesRefuseAPartialAnswer(t *testing.T) {
	c, _, _ := fakeProm(t, true)
	_, err := (&Pricer{Prom: c}).Nodes(context.Background(), static, 7*24*time.Hour, time.Now())
	if err == nil || !strings.Contains(err.Error(), "incomplete result") {
		t.Errorf("a partial answer must not price the nodes, got %v", err)
	}
}
