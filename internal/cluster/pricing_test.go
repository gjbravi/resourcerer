package cluster

import (
	"context"
	"math"
	"testing"

	"github.com/gjbravi/resourcerer/internal/engine"
)

func TestCalibrate(t *testing.T) {
	static := engine.Rates{CPUCoreHour: 0.0316, MemoryGiBHour: 0.0042}
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

func TestStaticSourceSkipsPrometheus(t *testing.T) {
	static := engine.Rates{CPUCoreHour: 1, MemoryGiBHour: 2}
	var p *Pricer
	if got := p.Pricing(context.Background(), static, CostAuto); got.Rates != static || got.Source != CostStatic {
		t.Errorf("a nil pricer must return the static rates, got %+v", got)
	}
}
