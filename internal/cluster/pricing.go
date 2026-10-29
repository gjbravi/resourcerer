package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Cost sources: static uses the configured rates; karpenter scales them to what the nodes cost according to
// Karpenter's price estimates and reports an error when it cannot; auto uses Karpenter when its metrics are
// there and the configured rates otherwise.
const (
	CostStatic    = "static"
	CostKarpenter = "karpenter"
	CostAuto      = "auto"
)

// States of the pricing, from the configured source and the last calibration.
const (
	// PricingOK: the rates are what the source asks for, up to date.
	PricingOK = "ok"
	// PricingStale: the last recalibration failed, so the previous calibration is still used.
	PricingStale = "stale"
	// PricingFallback: auto found no Karpenter prices and uses the configured rates.
	PricingFallback = "fallback"
	// PricingError: karpenter was asked for but cannot be used, or has failed for too long.
	PricingError = "error"
)

const (
	// DefaultRefresh is how long a calibration is reused; MinRefresh keeps a short setting from loading Prometheus.
	DefaultRefresh = 10 * time.Minute
	MinRefresh     = 30 * time.Second
	// staleFailures is how many failed recalibrations in a row turn a stale karpenter source into an error.
	staleFailures = 3
	// retryAfter is how soon a failed calibration is tried again, so a Prometheus that was briefly unreachable
	// does not leave the rates wrong for a whole refresh interval.
	retryAfter = time.Minute
)

// Pricing is the rates every cost estimate uses, and where they come from.
type Pricing struct {
	Rates engine.Rates
	// Static is the configured rates; Rates over Static is the cluster factor.
	Static engine.Rates
	// Allocation is how the rates are split between workloads: cluster or node.
	Allocation string
	// Mode is the configured source; Source is the one in use, karpenter or static.
	Mode, Source string
	State        string
	// Reason explains a state other than ok.
	Reason string
	// Fields below describe the Karpenter nodes behind calibrated rates.
	Nodes, Spot, OnDemand int
	HourlyCost            float64
	OnDemandHourlyCost    float64
	CPUCores, MemoryBytes float64
	// CalibratedAt is when the rates were last calibrated successfully.
	CalibratedAt time.Time
	// Refresh is how often the rates are recalibrated.
	Refresh time.Duration
}

// Pricer calibrates rates from Prometheus and keeps the last good calibration, so a failed query does not swing
// every cost back to the configured rates.
type Pricer struct {
	Prom *prom.Client
	// fetch replaces the Prometheus queries in tests.
	fetch func(context.Context) (Pricing, error)

	mu        sync.Mutex
	attempted time.Time
	good      *Pricing
	err       error
	failures  int
}

// Pricing returns the rates for the configured source, recalibrating once refresh has passed.
func (p *Pricer) Pricing(ctx context.Context, static engine.Rates, mode string, refresh time.Duration) Pricing {
	if mode != CostKarpenter && mode != CostAuto {
		mode = CostStatic
	}
	if mode == CostStatic || p == nil || (p.Prom == nil && p.fetch == nil) {
		return Pricing{Rates: static, Mode: mode, Source: CostStatic, State: PricingOK}
	}
	refresh = max(refresh, MinRefresh)

	p.mu.Lock()
	defer p.mu.Unlock()
	wait := refresh
	if p.err != nil {
		wait = min(refresh, retryAfter)
	}
	if time.Since(p.attempted) > wait {
		p.attempted = time.Now()
		fetch := p.fetch
		if fetch == nil {
			fetch = p.karpenter
		}
		got, err := fetch(ctx)
		if err == nil {
			got.CalibratedAt = p.attempted
			p.good, p.err, p.failures = &got, nil, 0
		} else {
			p.err = err
			p.failures++
			lg := log.FromContext(ctx)
			if mode == CostKarpenter {
				lg.Error(err, "calibrate cost rates from Karpenter prices")
			} else {
				lg.V(1).Info("calibrate cost rates from Karpenter prices", "error", err.Error())
			}
		}
	}

	if p.good == nil {
		out := Pricing{Rates: static, Mode: mode, Source: CostStatic, State: PricingFallback, Refresh: refresh}
		if p.err != nil {
			out.Reason = p.err.Error()
		}
		if mode == CostKarpenter {
			out.State = PricingError
		}
		return out
	}
	out := *p.good
	out.Rates = Calibrate(static, out.HourlyCost, out.CPUCores, out.MemoryBytes)
	out.Mode, out.Source, out.State, out.Refresh = mode, CostKarpenter, PricingOK, refresh
	if p.err != nil {
		out.State, out.Reason = PricingStale, p.err.Error()
		if mode == CostKarpenter && p.failures >= staleFailures {
			out.State = PricingError
		}
	}
	return out
}

func (p *Pricer) karpenter(ctx context.Context) (Pricing, error) {
	var out Pricing
	cost, err := p.Prom.Instant(ctx, prom.KarpenterHourlyCost(), "")
	if err != nil {
		return out, fmt.Errorf("query Karpenter prices: %w", err)
	}
	out.HourlyCost = cost[""]
	if out.HourlyCost <= 0 {
		return out, fmt.Errorf("no Karpenter node prices in Prometheus (karpenter_nodes_allocatable, karpenter_cloudprovider_instance_type_offering_price_estimate)")
	}
	if od, err := p.Prom.Instant(ctx, prom.KarpenterOnDemandHourlyCost(), ""); err == nil {
		out.OnDemandHourlyCost = od[""]
	}
	capacity, err := p.Prom.Instant(ctx, prom.KarpenterCapacity(), "resource_type")
	if err != nil {
		return out, fmt.Errorf("query Karpenter capacity: %w", err)
	}
	out.CPUCores, out.MemoryBytes = capacity["cpu"], capacity["memory"]
	if out.CPUCores <= 0 || out.MemoryBytes <= 0 {
		return out, fmt.Errorf("no allocatable capacity for the priced Karpenter nodes")
	}
	if counts, err := p.Prom.Instant(ctx, prom.KarpenterNodeCounts(), "capacity_type"); err == nil {
		for k, v := range counts {
			out.Nodes += int(v)
			switch k {
			case "spot":
				out.Spot += int(v)
			case "on-demand":
				out.OnDemand += int(v)
			}
		}
	}
	return out, nil
}

// ClusterFactor is how much the rates in use differ from the configured ones.
func (p Pricing) ClusterFactor() float64 {
	if p.Static.CPUCoreHour > 0 {
		return p.Rates.CPUCoreHour / p.Static.CPUCoreHour
	}
	if p.Static.MemoryGiBHour > 0 {
		return p.Rates.MemoryGiBHour / p.Static.MemoryGiBHour
	}
	return 1
}

// Calibrate scales configured rates so the capacity they price costs what the nodes cost per hour. The CPU to
// memory ratio of the configured rates is kept: node prices do not say how much of them is CPU.
func Calibrate(static engine.Rates, hourly, cores, memBytes float64) engine.Rates {
	base := cores*static.CPUCoreHour + memBytes/engine.Gi*static.MemoryGiBHour
	if hourly <= 0 || base <= 0 {
		return static
	}
	k := hourly / base
	return engine.Rates{CPUCoreHour: static.CPUCoreHour * k, MemoryGiBHour: static.MemoryGiBHour * k}
}
