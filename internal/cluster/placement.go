package cluster

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Unpriced is the capacity type of requests on nodes without a Karpenter price.
const Unpriced = "unpriced"

// placementBuckets caps the timeline, so a chart stays readable whatever the window.
const placementBuckets = 84

// Placement is where a workload's requests ran over a window, by instance type and over time, and what each part
// cost. Shares are of the workload's requests or of its cost; unpriced requests cost the cluster factor.
type Placement struct {
	Allocation
	Start, End time.Time
	Step       time.Duration
	// Types are every instance type the workload ran on, the largest share of requests first.
	Types    []TypeUse
	Unpriced Use
	Capacity map[string]Use
	Timeline []Bucket
}

type Use struct{ Requests, Cost float64 }

type TypeUse struct {
	Capacity, Name string
	Factor         float64
	Use
}

// Bucket shares the workload's requests and cost between capacity types (and Unpriced) over one period.
type Bucket struct {
	T              float64
	Requests, Cost map[string]float64
}

// Place reads where the workload ran over the window ending at end.
func Place(ctx context.Context, p *prom.Client, ref prom.Ref, static engine.Rates, clusterFactor float64, window time.Duration, end time.Time, minStep time.Duration) (Placement, error) {
	h, err := fetch(ctx, p, ref, static, window, end, minStep)
	if err != nil {
		return Placement{}, err
	}
	return place(h, clusterFactor)
}

func place(h history, k float64) (Placement, error) {
	a, err := combine(sums(h.cost), sums(h.priced), sum(h.total[""]), k)
	if err != nil {
		return Placement{}, err
	}
	pl := Placement{Allocation: a, Start: h.start, End: h.end, Step: h.step, Capacity: map[string]Use{}}
	total := sum(h.total[""])
	priced, cost := 0.0, 0.0
	for key, pts := range h.priced {
		capacity, name, _ := strings.Cut(key, "|")
		r, c := sum(pts), sum(h.cost[key])
		t := TypeUse{Capacity: capacity, Name: name, Use: Use{Requests: r, Cost: c}}
		if r > 0 {
			t.Factor = c / r
		}
		pl.Types = append(pl.Types, t)
		priced += r
		cost += c
	}
	unpriced := max(total-priced, 0)
	cost += unpriced * k
	pl.Unpriced = Use{Requests: unpriced / total, Cost: share(unpriced*k, cost)}
	for i := range pl.Types {
		t := &pl.Types[i]
		t.Requests, t.Cost = t.Requests/total, share(t.Cost, cost)
		u := pl.Capacity[t.Capacity]
		u.Requests += t.Requests
		u.Cost += t.Cost
		pl.Capacity[t.Capacity] = u
	}
	sort.Slice(pl.Types, func(i, j int) bool {
		if pl.Types[i].Requests != pl.Types[j].Requests {
			return pl.Types[i].Requests > pl.Types[j].Requests
		}
		return pl.Types[i].Capacity+pl.Types[i].Name < pl.Types[j].Capacity+pl.Types[j].Name
	})
	pl.Timeline = timeline(h, k)
	return pl, nil
}

// timeline sums requests and cost per capacity type over at most placementBuckets periods, then shares them.
func timeline(h history, k float64) []Bucket {
	steps := int(h.end.Sub(h.start)/h.step) + 1
	per := int(math.Ceil(float64(steps) / placementBuckets))
	at := func(t float64) int { return int((t-float64(h.start.Unix()))/h.step.Seconds()) / per }
	n := (steps + per - 1) / per
	req := make([]map[string]float64, n)
	cost := make([]map[string]float64, n)
	for i := range req {
		req[i], cost[i] = map[string]float64{}, map[string]float64{}
	}
	add := func(m []map[string]float64, key string, pts []prom.Point, f float64) {
		for _, p := range pts {
			if i := at(p.T); i >= 0 && i < n {
				m[i][key] += p.V * f
			}
		}
	}
	for key, pts := range h.priced {
		capacity, _, _ := strings.Cut(key, "|")
		add(req, capacity, pts, 1)
		add(cost, capacity, h.cost[key], 1)
		add(req, Unpriced, pts, -1)
		add(cost, Unpriced, pts, -k)
	}
	add(req, Unpriced, h.total[""], 1)
	add(cost, Unpriced, h.total[""], k)
	var out []Bucket
	for i := range n {
		r, c := normalize(req[i]), normalize(cost[i])
		if r == nil {
			continue
		}
		out = append(out, Bucket{T: float64(h.start.Unix()) + float64(i*per)*h.step.Seconds(), Requests: r, Cost: c})
	}
	return out
}

// normalize turns sums into shares, dropping rounding noise below zero; nil when there is nothing.
func normalize(m map[string]float64) map[string]float64 {
	t := 0.0
	for k, v := range m {
		if v < 1e-12 {
			delete(m, k)
			continue
		}
		t += v
	}
	if t <= 0 {
		return nil
	}
	for k, v := range m {
		m[k] = v / t
	}
	return m
}

func share(v, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return v / total
}
