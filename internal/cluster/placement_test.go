package cluster

import (
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/prom"
)

// series is v at each of n steps from start.
func series(start time.Time, step time.Duration, vs ...float64) []prom.Point {
	out := make([]prom.Point, 0, len(vs))
	for i, v := range vs {
		if v >= 0 {
			out = append(out, prom.Point{T: float64(start.Add(time.Duration(i) * step).Unix()), V: v})
		}
	}
	return out
}

func TestPlace(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	step := time.Hour
	// Four hours of the same requests: two on spot, one on on-demand, one on a node without a price.
	h := history{
		start: start, end: start.Add(3 * step), step: step,
		priced: map[string][]prom.Point{
			"spot|t3.xlarge":       series(start, step, 1, 1, -1, -1),
			"on-demand|m7g.xlarge": series(start, step, -1, -1, 1, -1),
		},
		cost: map[string][]prom.Point{
			"spot|t3.xlarge":       series(start, step, 0.25, 0.25, -1, -1),
			"on-demand|m7g.xlarge": series(start, step, -1, -1, 0.85, -1),
		},
		total: map[string][]prom.Point{"": series(start, step, 1, 1, 1, 1)},
	}
	pl, err := place(h, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	// Cost: 0.5 spot + 0.85 on-demand + 0.5 unpriced at the cluster factor = 1.85.
	if !near(pl.Capacity["spot"].Requests, 0.5) || !near(pl.Capacity["on-demand"].Cost, 0.85/1.85) {
		t.Errorf("capacity = %+v", pl.Capacity)
	}
	if !near(pl.Unpriced.Requests, 0.25) || !near(pl.Unpriced.Cost, 0.5/1.85) {
		t.Errorf("unpriced = %+v", pl.Unpriced)
	}
	if len(pl.Types) != 2 || pl.Types[0].Name != "t3.xlarge" || !near(pl.Types[1].Factor, 0.85) {
		t.Errorf("types = %+v", pl.Types)
	}
	if len(pl.Timeline) != 4 {
		t.Fatalf("timeline has %d buckets, want 4", len(pl.Timeline))
	}
	if b := pl.Timeline[2]; !near(b.Requests["on-demand"], 1) || len(b.Requests) != 1 {
		t.Errorf("third hour = %+v, want all on-demand", b.Requests)
	}
	if b := pl.Timeline[3]; !near(b.Requests[Unpriced], 1) {
		t.Errorf("fourth hour = %+v, want all unpriced", b.Requests)
	}
}

func TestTimelineIsBucketed(t *testing.T) {
	start := time.Unix(1_700_000_000, 0)
	step := 30 * time.Minute
	vs := make([]float64, 336)
	for i := range vs {
		vs[i] = 1
	}
	h := history{start: start, end: start.Add(335 * step), step: step,
		priced: map[string][]prom.Point{"spot|t3.xlarge": series(start, step, vs...)},
		cost:   map[string][]prom.Point{"spot|t3.xlarge": series(start, step, vs...)},
		total:  map[string][]prom.Point{"": series(start, step, vs...)},
	}
	pl, err := place(h, 1)
	if err != nil {
		t.Fatal(err)
	}
	if n := len(pl.Timeline); n > placementBuckets || n < placementBuckets/2 {
		t.Errorf("timeline has %d buckets, want at most %d", n, placementBuckets)
	}
}
