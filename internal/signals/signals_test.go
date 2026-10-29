package signals

import (
	"strings"
	"testing"
	"time"

	"github.com/gjbravi/resourcerer/internal/prom"
)

const config = `
builtin:
  throttling: { warn: 0.2 }
  bestEffort: { enabled: false }
signals:
  - id: http-5xx
    name: HTTP 5xx rate
    unit: ratio
    above: { warn: 0.01, critical: 0.05 }
    query: |
      sum(rate(errors{namespace="{{ .Namespace }}", workload="{{ .Name }}"}[{{ .Window }}]))
  - id: availability
    name: Availability
    unit: ratio
    below: { critical: 0.99 }
    query: avg_over_time(up{namespace="{{ .Namespace }}"}[{{ .Window }}])
  - id: Bad_ID
    name: Bad
    unit: ratio
    above: { warn: 1 }
    query: x
  - id: no-levels
    name: No levels
    unit: count
    above: {}
    query: x
  - id: broken
    name: Broken template
    unit: count
    above: { warn: 1 }
    query: '{{ .Nope }}'
`

func TestParse(t *testing.T) {
	c := Parse(config)
	if len(c.Signals) != 2 || c.Signals[0].ID != "http-5xx" || c.Signals[1].ID != "availability" {
		t.Fatalf("signals = %+v", c.Signals)
	}
	if len(c.Errors) != 3 {
		t.Errorf("errors = %q, want 3", c.Errors)
	}
	if c.Builtin.Throttling.WarnOr(0.1) != 0.2 || c.Builtin.BestEffort.On() || !c.Builtin.Restarts.On() {
		t.Errorf("builtin = %+v", c.Builtin)
	}
	if bad := Parse("signals: [{id: x, unknown: 1}]"); len(bad.Errors) != 1 || len(bad.Signals) != 0 {
		t.Errorf("unknown fields must be rejected: %+v", bad)
	}
}

func TestRender(t *testing.T) {
	s := Parse(config).Signals[0]
	q, err := s.Render(VarsFor(prom.Ref{Namespace: "payments", Kind: "Deployment", Name: "checkout"}, 7*24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q, `namespace="payments", workload="checkout"}[1w]`) {
		t.Errorf("query = %s", q)
	}
}

func TestGradeAndOverrides(t *testing.T) {
	c := Parse(config)
	up, down := c.Signals[0], c.Signals[1]
	for v, want := range map[float64]string{0.001: OK, 0.02: Warn, 0.05: Critical} {
		if got := up.Grade(v, up.LevelsWith(Overrides{})); got != want {
			t.Errorf("above %v = %s, want %s", v, got, want)
		}
	}
	if got := down.Grade(0.95, down.LevelsWith(Overrides{})); got != Critical {
		t.Errorf("below 0.95 = %s, want critical", got)
	}
	o := OverridesFrom(
		map[string]string{"resourcerer.io/reliability.disabled": "availability", "resourcerer.io/reliability.http-5xx.critical": "0.2"},
		map[string]string{"resourcerer.io/reliability.http-5xx.critical": "0.1", "resourcerer.io/reliability.http-5xx.bogus": "1"},
	)
	if !o.Disabled["availability"] {
		t.Error("namespace annotation must disable availability")
	}
	if got := up.Grade(0.08, up.LevelsWith(o)); got != Warn {
		t.Errorf("0.08 with the workload's critical 0.1 = %s, want warn", got)
	}
}

func TestFormat(t *testing.T) {
	for _, c := range []struct {
		unit string
		v    float64
		want string
	}{{"ratio", 0.062, "6.20%"}, {"ratio", 0.5, "50%"}, {"seconds", 0.25, "250ms"}, {"seconds", 12.5, "12.5s"}, {"count", 3, "3"}, {"bytes", 64 << 20, "64MiB"}} {
		if got := Format(c.unit, c.v); got != c.want {
			t.Errorf("Format(%s, %v) = %s, want %s", c.unit, c.v, got, c.want)
		}
	}
}
