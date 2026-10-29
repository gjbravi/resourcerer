package cluster

import (
	"errors"
	"math"
	"testing"

	"github.com/gjbravi/resourcerer/internal/engine"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCombine(t *testing.T) {
	const k = 0.5
	tests := []struct {
		name          string
		cost, priced  map[string]float64
		total         float64
		factor, cover float64
		spot          float64
	}{
		{
			name:   "all on one on-demand node",
			cost:   map[string]float64{"on-demand|m7g.xlarge": 84},
			priced: map[string]float64{"on-demand|m7g.xlarge": 100},
			total:  100, factor: 0.84, cover: 1,
		},
		{
			name:   "time and request weighted across spot and on-demand",
			cost:   map[string]float64{"on-demand|m7g.xlarge": 25 * 0.84, "spot|t3.xlarge": 75 * 0.26},
			priced: map[string]float64{"on-demand|m7g.xlarge": 25, "spot|t3.xlarge": 75},
			total:  100, factor: 0.25*0.84 + 0.75*0.26, cover: 1, spot: 0.75,
		},
		{
			name:   "requests on unpriced nodes count at the cluster factor",
			cost:   map[string]float64{"spot|t3.xlarge": 60 * 0.26},
			priced: map[string]float64{"spot|t3.xlarge": 60},
			total:  100, factor: 0.6*0.26 + 0.4*k, cover: 0.6, spot: 1,
		},
		{
			name:  "no priced node keeps the cluster factor",
			total: 100, factor: k,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := combine(tt.cost, tt.priced, tt.total, k)
			if err != nil {
				t.Fatal(err)
			}
			if !near(a.Factor, tt.factor) || !near(a.Coverage, tt.cover) || !near(a.Capacity["spot"], tt.spot) {
				t.Errorf("got factor %v coverage %v spot %v, want %v %v %v", a.Factor, a.Coverage, a.Capacity["spot"], tt.factor, tt.cover, tt.spot)
			}
		})
	}
}

func TestCombineNoRequests(t *testing.T) {
	if _, err := combine(nil, nil, 0, 0.5); !errors.Is(err, errNoRequests) {
		t.Errorf("err = %v, want errNoRequests", err)
	}
}

func TestCombineTopInstanceTypes(t *testing.T) {
	priced := map[string]float64{"spot|a": 10, "spot|b": 40, "spot|c": 30, "on-demand|d": 20}
	a, err := combine(priced, priced, 100, 1)
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, s := range a.InstanceTypes {
		got = append(got, s.Name)
	}
	if len(got) != 3 || got[0] != "b" || got[1] != "c" || got[2] != "d" {
		t.Errorf("instance types = %v, want [b c d]", got)
	}
}

func TestAllocationRates(t *testing.T) {
	r := Allocation{Factor: 2}.Rates(static)
	if !near(r.CPUCoreHour, 2*static.CPUCoreHour) || !near(r.MemoryGiBHour, 2*static.MemoryGiBHour) {
		t.Errorf("rates = %+v", r)
	}
}

func TestWithAllocation(t *testing.T) {
	calibrated := Pricing{Rates: Calibrate(static, 2.1751, 115.9, 417.3*engine.Gi)}
	for in, want := range map[string]string{"": AllocationCluster, "node": AllocationNode, "nodes": AllocationCluster} {
		p := withAllocation(calibrated, static, in)
		if p.Allocation != want {
			t.Errorf("allocation %q = %q, want %q", in, p.Allocation, want)
		}
		if f := p.ClusterFactor(); math.Abs(f-0.402) > 0.001 {
			t.Errorf("cluster factor = %v, want about 0.402", f)
		}
	}
}
