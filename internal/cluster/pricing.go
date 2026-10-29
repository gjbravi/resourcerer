package cluster

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Cost sources: static uses the configured rates; karpenter scales them to what the nodes cost according to
// Karpenter's price estimates; auto uses karpenter when its metrics are there and static otherwise.
const (
	CostStatic    = "static"
	CostKarpenter = "karpenter"
	CostAuto      = "auto"
)

const pricingTTL = 10 * time.Minute

// Pricing is the rates every cost estimate uses, and where they come from.
type Pricing struct {
	Rates  engine.Rates
	Source string
	// Reason explains why configured rates are used when Karpenter prices were asked for.
	Reason string
	// Fields below describe the Karpenter nodes behind calibrated rates.
	Nodes, Spot, OnDemand int
	HourlyCost            float64
	OnDemandHourlyCost    float64
	CPUCores, MemoryBytes float64
	UpdatedAt             time.Time
}

// Pricer calibrates rates from Prometheus and caches them, since node prices change slowly.
type Pricer struct {
	Prom *prom.Client

	mu     sync.Mutex
	at     time.Time
	cached Pricing
	err    error
}

// Pricing returns rates for the configured source; static are the configured rates.
func (p *Pricer) Pricing(ctx context.Context, static engine.Rates, source string) Pricing {
	if source == CostStatic || p == nil || p.Prom == nil {
		return Pricing{Rates: static, Source: CostStatic}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if time.Since(p.at) > pricingTTL {
		p.cached, p.err = p.karpenter(ctx)
		p.at = time.Now()
	}
	if p.err != nil {
		return Pricing{Rates: static, Source: CostStatic, Reason: p.err.Error()}
	}
	out := p.cached
	out.Rates = Calibrate(static, out.HourlyCost, out.CPUCores, out.MemoryBytes)
	out.Source = CostKarpenter
	return out
}

func (p *Pricer) karpenter(ctx context.Context) (Pricing, error) {
	out := Pricing{UpdatedAt: time.Now()}
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
