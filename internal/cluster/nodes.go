package cluster

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// nodeRefresh is how long the node prices are reused for every workload: they are the cluster's, not a workload's.
const nodeRefresh = 10 * time.Minute

// NodePrice is what a Karpenter node costs relative to the configured rates: its offering's price now over its
// allocatable capacity valued at those rates.
type NodePrice struct {
	Capacity, Type string
	Factor         float64
}

type nodeEntry struct {
	at    time.Time
	nodes map[string]NodePrice
	err   error
}

// Nodes prices every Karpenter node seen over the window ending at end, at its offering's current price. The
// result is the cluster's, so it is computed once for every workload and kept for nodeRefresh, a minute after a
// failure. Nodes launched since are left unpriced until the next refresh: their requests count at the cluster factor.
func (p *Pricer) Nodes(ctx context.Context, static engine.Rates, window time.Duration, end time.Time) (map[string]NodePrice, error) {
	if p == nil || p.Prom == nil {
		return nil, fmt.Errorf("node prices need Prometheus")
	}
	key := fmt.Sprint(window, static)
	p.nodesMu.Lock()
	defer p.nodesMu.Unlock()
	if e, ok := p.nodes[key]; ok {
		ttl := nodeRefresh
		if e.err != nil {
			ttl = retryAfter
		}
		if time.Since(e.at) < ttl {
			return e.nodes, e.err
		}
	}
	nodes, err := fetchNodes(prom.For(ctx, "node-prices"), p.Prom, static, window, end)
	if p.nodes == nil {
		p.nodes = map[string]nodeEntry{}
	}
	p.nodes[key] = nodeEntry{at: time.Now(), nodes: nodes, err: err}
	return nodes, err
}

func fetchNodes(ctx context.Context, p *prom.Client, static engine.Rates, window time.Duration, end time.Time) (map[string]NodePrice, error) {
	ctx, cancel := context.WithTimeout(ctx, allocationTimeout)
	defer cancel()
	cpu, err := p.Vector(ctx, prom.KarpenterNodes("cpu", window), end)
	if err != nil {
		return nil, fmt.Errorf("query the Karpenter nodes: %w", err)
	}
	mem, err := p.Vector(ctx, prom.KarpenterNodes("memory", window), end)
	if err != nil {
		return nil, fmt.Errorf("query the Karpenter nodes' memory: %w", err)
	}
	types := map[string]bool{}
	for _, s := range cpu {
		if t := s.Labels["instance_type"]; t != "" {
			types[t] = true
		}
	}
	if len(types) == 0 {
		return map[string]NodePrice{}, nil
	}
	names := make([]string, 0, len(types))
	for t := range types {
		names = append(names, t)
	}
	sort.Strings(names)
	prices, err := p.Vector(ctx, prom.KarpenterOfferingPrices(names), end)
	if err != nil {
		return nil, fmt.Errorf("query Karpenter offering prices: %w", err)
	}
	return nodePrices(cpu, mem, prices, static), nil
}

// nodePrices prices each node at its offering: instance type, zone and capacity type, reserved capacity priced as
// on-demand. The factor is the price over what the node's allocatable CPU and memory cost at the configured rates.
// A node without an offering price, or without capacity, is left out: requests on it count as unpriced.
func nodePrices(cpu, mem, prices []prom.Sample, static engine.Rates) map[string]NodePrice {
	type offering struct{ typ, zone, capacity string }
	price := map[offering]float64{}
	for _, s := range prices {
		price[offering{s.Labels["instance_type"], s.Labels["zone"], s.Labels["capacity_type"]}] = s.Value
	}
	memory := map[string]float64{}
	for _, s := range mem {
		memory[s.Labels["node_name"]] = s.Value
	}
	out := map[string]NodePrice{}
	for _, s := range cpu {
		name, typ, capacity := s.Labels["node_name"], s.Labels["instance_type"], s.Labels["capacity_type"]
		priceType := capacity
		if capacity == "reserved" {
			priceType = "on-demand"
		}
		hourly, ok := price[offering{typ, s.Labels["zone"], priceType}]
		value := s.Value*static.CPUCoreHour + memory[name]/engine.Gi*static.MemoryGiBHour
		if !ok || name == "" || value <= 0 {
			continue
		}
		out[name] = NodePrice{Capacity: capacity, Type: typ, Factor: hourly / value}
	}
	return out
}

// split values the workload's requests on each node, keyed "capacity_type|instance_type": times the node's factor
// (cost) and as they are (priced) on priced nodes, and on any node (total, keyed ""). Each comes back as a series
// over time, its points in order.
func split(byNode map[string][]prom.Point, nodes map[string]NodePrice) (cost, priced, total map[string][]prom.Point) {
	c, p, t := map[string]map[float64]float64{}, map[string]map[float64]float64{}, map[string]map[float64]float64{}
	add := func(m map[string]map[float64]float64, key string, pt prom.Point, f float64) {
		if m[key] == nil {
			m[key] = map[float64]float64{}
		}
		m[key][pt.T] += pt.V * f
	}
	for node, pts := range byNode {
		n, ok := nodes[node]
		for _, pt := range pts {
			add(t, "", pt, 1)
			if ok {
				key := n.Capacity + "|" + n.Type
				add(p, key, pt, 1)
				add(c, key, pt, n.Factor)
			}
		}
	}
	return inOrder(c), inOrder(p), inOrder(t)
}

// inOrder turns values by time into series with their points in order.
func inOrder(m map[string]map[float64]float64) map[string][]prom.Point {
	out := make(map[string][]prom.Point, len(m))
	for key, byT := range m {
		ts := make([]float64, 0, len(byT))
		for ts0 := range byT {
			ts = append(ts, ts0)
		}
		slices.Sort(ts)
		pts := make([]prom.Point, len(ts))
		for i, ts0 := range ts {
			pts[i] = prom.Point{T: ts0, V: byT[ts0]}
		}
		out[key] = pts
	}
	return out
}
