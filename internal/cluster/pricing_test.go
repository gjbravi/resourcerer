package cluster

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
)

var static = engine.Rates{CPUCoreHour: 0.0316, MemoryGiBHour: 0.0042}

func TestCalibrate(t *testing.T) {
	// develop-pe: 115.9 cores and 417.3GiB of Karpenter nodes cost $2.1751/h, $5.41/h at the static rates.
	got := Calibrate(static, 2.1751, 115.9, 417.3*engine.Gi)
	if math.Abs(got.CPUCoreHour-0.01271) > 0.0001 || math.Abs(got.MemoryGiBHour-0.00169) > 0.00001 {
		t.Errorf("rates = %+v, want about 0.0127/core-h and 0.0017/GiB-h", got)
	}
	if r := got.CPUCoreHour / got.MemoryGiBHour; math.Abs(r-static.CPUCoreHour/static.MemoryGiBHour) > 1e-9 {
		t.Errorf("CPU to memory ratio changed to %v", r)
	}
	if got := Calibrate(static, 0, 115.9, 417.3*engine.Gi); got != static {
		t.Errorf("no price must keep the static rates, got %+v", got)
	}
}

// pricer answers from results in order, one per recalibration.
func pricer(results ...error) *Pricer {
	i := 0
	return &Pricer{fetch: func(context.Context) (Pricing, error) {
		err := results[min(i, len(results)-1)]
		i++
		if err != nil {
			return Pricing{}, err
		}
		return Pricing{HourlyCost: 2.1751, CPUCores: 115.9, MemoryBytes: 417.3 * engine.Gi, Nodes: 27, Spot: 26, OnDemand: 1}, nil
	}}
}

// next makes the following call recalibrate.
func (p *Pricer) next() { p.attempted = time.Time{} }

func TestPricingStates(t *testing.T) {
	ctx := context.Background()
	down := errors.New("prometheus down")

	if got := pricer(nil).Pricing(ctx, static, CostStatic, 0); got.Source != CostStatic || got.State != PricingOK || got.Rates != static {
		t.Errorf("static = %+v", got)
	}

	p := pricer(nil, down, down, down)
	got := p.Pricing(ctx, static, CostAuto, time.Minute)
	if got.Source != CostKarpenter || got.State != PricingOK || got.Rates == static || got.Spot != 26 {
		t.Fatalf("auto with Karpenter = %+v", got)
	}
	calibrated := got.Rates
	for i := 1; i <= 3; i++ {
		p.next()
		got = p.Pricing(ctx, static, CostAuto, time.Minute)
		if got.State != PricingStale || got.Rates != calibrated {
			t.Errorf("auto after %d failures = %s with %+v, want stale with the last calibration", i, got.State, got.Rates)
		}
	}
	if got := p.Pricing(ctx, static, CostKarpenter, time.Minute); got.State != PricingError || got.Rates != calibrated {
		t.Errorf("karpenter after 3 failures = %s with %+v, want error with the last calibration", got.State, got.Rates)
	}

	none := errors.New("no Karpenter node prices")
	if got := pricer(none).Pricing(ctx, static, CostAuto, time.Minute); got.State != PricingFallback || got.Rates != static || got.Reason == "" {
		t.Errorf("auto without Karpenter = %+v, want fallback to the static rates with a reason", got)
	}
	if got := pricer(none).Pricing(ctx, static, CostKarpenter, time.Minute); got.State != PricingError || got.Rates != static {
		t.Errorf("karpenter without Karpenter = %+v, want error on the static rates", got)
	}

	recovered := pricer(nil, down, nil)
	recovered.Pricing(ctx, static, CostKarpenter, time.Minute)
	recovered.next()
	recovered.Pricing(ctx, static, CostKarpenter, time.Minute)
	recovered.next()
	if got := recovered.Pricing(ctx, static, CostKarpenter, time.Minute); got.State != PricingOK || got.Reason != "" {
		t.Errorf("after recovering = %+v, want ok", got)
	}
}

func TestFailureRetriesSooner(t *testing.T) {
	ctx := context.Background()
	p := pricer(errors.New("prometheus down"), nil)
	if got := p.Pricing(ctx, static, CostAuto, time.Hour); got.State != PricingFallback {
		t.Fatalf("first call = %s, want fallback", got.State)
	}
	p.attempted = time.Now().Add(-2 * retryAfter)
	if got := p.Pricing(ctx, static, CostAuto, time.Hour); got.State != PricingOK {
		t.Errorf("after %v = %s, want a retry before the 1h refresh", 2*retryAfter, got.State)
	}
	p.attempted = time.Now().Add(-2 * retryAfter)
	before := p.attempted
	p.Pricing(ctx, static, CostAuto, time.Hour)
	if p.attempted != before {
		t.Errorf("a good calibration must wait for the refresh interval")
	}
}

func TestNilPricerUsesStaticRates(t *testing.T) {
	var p *Pricer
	if got := p.Pricing(context.Background(), static, CostAuto, time.Minute); got.Rates != static || got.Source != CostStatic {
		t.Errorf("a nil pricer must return the static rates, got %+v", got)
	}
}

func TestRefreshInterval(t *testing.T) {
	cases := map[string]time.Duration{"": DefaultRefresh, "1h": time.Hour, "5s": MinRefresh, "soon": DefaultRefresh, "-1m": DefaultRefresh}
	for in, want := range cases {
		if got := refreshInterval(in); got != want {
			t.Errorf("refreshInterval(%q) = %v, want %v", in, got, want)
		}
	}
}
