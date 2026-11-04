package decision

import (
	"math"
	"time"

	"github.com/gjbravi/resourcerer/internal/engine"
	"github.com/gjbravi/resourcerer/internal/prom"
)

// Shape describes a busiest-pod usage series beyond its percentiles: how spiky it is, whether it follows the time
// of day, and whether it grows.
type Shape struct {
	Samples            int
	P50, P95, P99, Max float64
	Mean               float64
	// Spike is max over p95: how far the rare peaks reach above what the workload needs most of the time.
	Spike float64
	// CV is the coefficient of variation.
	CV float64
	// Daily is the share of the variance the hour of day explains; PeakFrom and PeakTo (UTC hours, PeakTo
	// exclusive) bound the busiest hours. Zero without two days of history.
	Daily            float64
	PeakFrom, PeakTo int
	// Trend is the slope of the daily means as a share of the mean, per day. Zero without three days of history.
	Trend float64
}

const (
	// minDailyHistory is the least history a time-of-day pattern is looked for in.
	minDailyHistory = 48 * time.Hour
	// minTrendDays is the least number of whole days a trend is measured over, so the time of day cannot pass for one.
	minTrendDays = 3
)

// ShapeOf describes a series of points, oldest first.
func ShapeOf(pts []prom.Point) Shape {
	vs := prom.Values(pts)
	s := Shape{Samples: len(vs)}
	if len(vs) == 0 {
		return s
	}
	s.P50 = engine.Quantile(vs, 0.5)
	s.P95 = engine.Quantile(vs, 0.95)
	s.P99 = engine.Quantile(vs, 0.99)
	s.Max = engine.Quantile(vs, 1)
	s.Mean = prom.Mean(pts)
	if s.P95 > 0 {
		s.Spike = s.Max / s.P95
	}
	if s.Mean <= 0 {
		return s
	}
	var ss float64
	for _, v := range vs {
		ss += (v - s.Mean) * (v - s.Mean)
	}
	s.CV = math.Sqrt(ss/float64(len(vs))) / s.Mean
	s.Trend = trend(pts, s.Mean)
	if span := time.Duration(pts[len(pts)-1].T-pts[0].T) * time.Second; span >= minDailyHistory && ss > 0 {
		s.Daily, s.PeakFrom, s.PeakTo = daily(pts, s.Mean, ss)
	}
	return s
}

// trend fits a line through the means of whole days since the first sample.
func trend(pts []prom.Point, mean float64) float64 {
	var sum []float64
	var n []int
	for _, p := range pts {
		d := int((p.T - pts[0].T) / 86400)
		for len(sum) <= d {
			sum, n = append(sum, 0), append(n, 0)
		}
		sum[d] += p.V
		n[d]++
	}
	if (pts[len(pts)-1].T-pts[0].T)/86400 < float64(len(sum)) {
		sum, n = sum[:len(sum)-1], n[:len(n)-1]
	}
	var days []prom.Point
	for d := range sum {
		if n[d] > 0 {
			days = append(days, prom.Point{T: float64(d), V: sum[d] / float64(n[d])})
		}
	}
	if len(days) < minTrendDays {
		return 0
	}
	return slope(days) / mean
}

// slope is the least-squares slope of the values over T.
func slope(pts []prom.Point) float64 {
	var mt, mv float64
	for _, p := range pts {
		mt += p.T
		mv += p.V
	}
	n := float64(len(pts))
	mt, mv = mt/n, mv/n
	var num, den float64
	for _, p := range pts {
		num += (p.T - mt) * (p.V - mv)
		den += (p.T - mt) * (p.T - mt)
	}
	if den == 0 {
		return 0
	}
	return num / den
}

// daily returns the share of the variance (ss, the sum of squares around mean) the hour-of-day means explain, and
// the busiest run of hours: those within a quarter of the spread from the busiest hour, contiguous around it.
func daily(pts []prom.Point, mean, ss float64) (float64, int, int) {
	var sum [24]float64
	var n [24]int
	for _, p := range pts {
		h := time.Unix(int64(p.T), 0).UTC().Hour()
		sum[h] += p.V
		n[h]++
	}
	var avg [24]float64
	var between float64
	lo, hi, top := math.Inf(1), math.Inf(-1), -1
	for h := range 24 {
		if n[h] == 0 {
			return 0, 0, 0
		}
		avg[h] = sum[h] / float64(n[h])
		between += float64(n[h]) * (avg[h] - mean) * (avg[h] - mean)
		lo = math.Min(lo, avg[h])
		if avg[h] > hi {
			hi, top = avg[h], h
		}
	}
	cut := hi - (hi-lo)/4
	from, to := top, (top+1)%24
	for i := 1; i < 24 && avg[(top-i+24)%24] >= cut; i++ {
		from = (top - i + 24) % 24
	}
	for i := 1; i < 24 && avg[(top+i)%24] >= cut; i++ {
		to = (top + i + 1) % 24
	}
	return between / ss, from, to
}
