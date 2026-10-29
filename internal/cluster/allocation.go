package cluster

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Cost allocations: cluster prices every workload at the cluster's average rates; node prices each workload at the
// rates of the nodes its pods ran on, weighted by time and requests over the learning period.
const (
	AllocationCluster = "cluster"
	AllocationNode    = "node"
)

const (
	// allocationWindow caps the history read for node allocation, and allocationPoints the samples in it, so
	// long learning periods do not load Prometheus.
	allocationWindow  = 7 * 24 * time.Hour
	allocationPoints  = 336
	allocationTimeout = 30 * time.Second
	// topInstanceTypes is how many instance types an allocation lists.
	topInstanceTypes = 3
)

// Allocation is what a workload's nodes cost relative to the configured rates.
type Allocation struct {
	// Factor multiplies the configured rates; requests not on priced nodes count at the cluster factor.
	Factor float64
	// Coverage is the share of the workload's requests that ran on nodes with a price.
	Coverage float64
	// Capacity and InstanceTypes split the priced requests by capacity type and by instance type (top few).
	Capacity      map[string]float64
	InstanceTypes []Share
}

type Share struct {
	Name  string
	Share float64
}

// Rates returns the configured rates scaled by the allocation's factor.
func (a Allocation) Rates(static engine.Rates) engine.Rates {
	return engine.Rates{CPUCoreHour: static.CPUCoreHour * a.Factor, MemoryGiBHour: static.MemoryGiBHour * a.Factor}
}

var errNoRequests = errors.New("no running pods with requests in the window")

// Allocate reads where the workload's pods ran over the window ending at end and prices them at their nodes.
// clusterFactor is the calibrated rates over the configured ones, used for requests on nodes without a price.
func Allocate(ctx context.Context, p *prom.Client, ref prom.Ref, static engine.Rates, clusterFactor float64, window time.Duration, end time.Time, minStep time.Duration) (Allocation, error) {
	h, err := fetch(ctx, p, ref, static, window, end, minStep)
	if err != nil {
		return Allocation{}, err
	}
	return combine(sums(h.cost), sums(h.priced), sum(h.total[""]), clusterFactor)
}

// history is the workload's requests over the window, valued at the configured rates: times the node factor
// (cost) and on priced nodes (priced), keyed "capacity_type|instance_type", and on any node (total).
type history struct {
	cost, priced, total map[string][]prom.Point
	start, end          time.Time
	step                time.Duration
}

func fetch(ctx context.Context, p *prom.Client, ref prom.Ref, static engine.Rates, window time.Duration, end time.Time, minStep time.Duration) (history, error) {
	window = min(window, allocationWindow)
	h := history{step: max(window/allocationPoints, minStep).Truncate(time.Second), start: end.Add(-window), end: end}
	ctx, cancel := context.WithTimeout(ctx, allocationTimeout)
	defer cancel()
	var err error
	if h.cost, err = p.RangeComplete(ctx, prom.WorkloadNodeCost(ref, static.CPUCoreHour, static.MemoryGiBHour, false), h.start, end, h.step, "key"); err != nil {
		return h, fmt.Errorf("query node prices of the workload's pods: %w", err)
	}
	if h.priced, err = p.RangeComplete(ctx, prom.WorkloadNodeCost(ref, static.CPUCoreHour, static.MemoryGiBHour, true), h.start, end, h.step, "key"); err != nil {
		return h, fmt.Errorf("query the workload's priced requests: %w", err)
	}
	if h.total, err = p.RangeComplete(ctx, prom.WorkloadRequestValue(ref, static.CPUCoreHour, static.MemoryGiBHour), h.start, end, h.step, ""); err != nil {
		return h, fmt.Errorf("query the workload's requests: %w", err)
	}
	return h, nil
}

// combine turns request-weighted sums into an allocation. cost and priced are keyed "capacity_type|instance_type":
// requests times node factor, and requests, on priced nodes; total is requests on any node.
func combine(cost, priced map[string]float64, total, clusterFactor float64) (Allocation, error) {
	if total <= 0 {
		return Allocation{}, errNoRequests
	}
	var c, p float64
	byCapacity, byType := map[string]float64{}, map[string]float64{}
	for k, v := range priced {
		capacity, typ, _ := strings.Cut(k, "|")
		byCapacity[capacity] += v
		byType[typ] += v
		p += v
		c += cost[k]
	}
	a := Allocation{Factor: clusterFactor, Coverage: min(p/total, 1), Capacity: map[string]float64{}}
	if p <= 0 {
		return a, nil
	}
	a.Factor = a.Coverage*(c/p) + (1-a.Coverage)*clusterFactor
	for k, v := range byCapacity {
		a.Capacity[k] = v / p
	}
	for k, v := range byType {
		a.InstanceTypes = append(a.InstanceTypes, Share{Name: k, Share: v / p})
	}
	sort.Slice(a.InstanceTypes, func(i, j int) bool {
		if a.InstanceTypes[i].Share != a.InstanceTypes[j].Share {
			return a.InstanceTypes[i].Share > a.InstanceTypes[j].Share
		}
		return a.InstanceTypes[i].Name < a.InstanceTypes[j].Name
	})
	if len(a.InstanceTypes) > topInstanceTypes {
		a.InstanceTypes = a.InstanceTypes[:topInstanceTypes]
	}
	return a, nil
}

func sums(series map[string][]prom.Point) map[string]float64 {
	out := make(map[string]float64, len(series))
	for k, pts := range series {
		out[k] = sum(pts)
	}
	return out
}

func sum(pts []prom.Point) float64 {
	s := 0.0
	for _, p := range pts {
		s += p.V
	}
	return s
}
