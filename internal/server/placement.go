package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/gjbravi/resourcerer/internal/cluster"
	"github.com/gjbravi/resourcerer/internal/prom"
	"github.com/gjbravi/resourcerer/internal/settings"
)

// placementCache keeps a placement for a while: it covers the whole learning period, so it barely moves.
const placementCache = 5 * time.Minute

// placement shows where the workload's requests ran over its learning period and what each part cost. It is
// computed on request from the same queries as node allocation, so the status stays small.
func (s *Server) placement(w http.ResponseWriter, r *http.Request) {
	rec, ok := s.get(w, r)
	if !ok {
		return
	}
	defaults, pricing := s.Cluster.Defaults(r.Context(), s.Client)
	out := Placement{Allocation: pricing.Allocation, ClusterFactor: pricing.ClusterFactor()}
	if pricing.Source != cluster.CostKarpenter {
		out.Reason = "Node prices need Karpenter; costs use the configured rates."
		writeJSON(w, out)
		return
	}
	win := 7 * 24 * time.Hour
	if v, err := settings.ParseDuration(rec.Status.Settings[settings.LearningPeriod].Value); err == nil {
		win = v
	} else if v := settings.Resolve(defaults, nil, nil).Duration(settings.LearningPeriod); v > 0 {
		win = v
	}
	end := time.Now().Truncate(max(s.MinStep, time.Minute))
	key := strings.Join([]string{"placement", rec.Namespace, rec.Name, end.Truncate(placementCache).String()}, "|")
	ref := prom.Ref{Namespace: rec.Namespace, Kind: rec.Spec.TargetRef.Kind, Name: rec.Spec.TargetRef.Name}
	v := s.cache.get(key, placementCache, func() (any, error) {
		return cluster.Place(r.Context(), s.Prom, ref, pricing.Static, pricing.ClusterFactor(), win, end, s.MinStep)
	})
	pl, isPlacement := v.(cluster.Placement)
	if !isPlacement {
		if err, isErr := v.(error); isErr {
			out.Reason = err.Error()
		}
		writeJSON(w, out)
		return
	}
	out.Available = true
	out.From, out.To, out.Step = float64(pl.Start.Unix()), float64(pl.End.Unix()), pl.Step.Seconds()
	out.Factor, out.Coverage = pl.Factor, pl.Coverage
	out.Unpriced = PlacementUse{Requests: pl.Unpriced.Requests, Cost: pl.Unpriced.Cost}
	out.Capacity = map[string]PlacementUse{}
	for c, u := range pl.Capacity {
		out.Capacity[c] = PlacementUse{Requests: u.Requests, Cost: u.Cost}
	}
	out.Types = make([]PlacementType, 0, len(pl.Types))
	for _, t := range pl.Types {
		out.Types = append(out.Types, PlacementType{Name: t.Name, Capacity: t.Capacity, Factor: t.Factor,
			PlacementUse: PlacementUse{Requests: t.Requests, Cost: t.Cost}})
	}
	out.Timeline = make([]PlacementBucket, 0, len(pl.Timeline))
	for _, b := range pl.Timeline {
		out.Timeline = append(out.Timeline, PlacementBucket{T: b.T, Requests: b.Requests, Cost: b.Cost})
	}
	writeJSON(w, out)
}
