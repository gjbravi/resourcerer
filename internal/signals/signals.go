// Package signals holds the reliability signals: built-in ones, which can be tuned or turned off, and signals
// declared as PromQL templates in the chart values, evaluated per workload.
package signals

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/prometheus/common/model"
	"sigs.k8s.io/yaml"

	"github.com/gjbravi/resourcerer/internal/prom"
)

// ConfigKey is the ConfigMap key the chart renders the reliability values into.
const ConfigKey = "reliability.yaml"

const (
	// MaxSignals caps declared signals, each one a query per workload and regeneration.
	MaxSignals = 20
	// QueryTimeout bounds each signal query.
	QueryTimeout = 10 * time.Second
)

// Statuses of a signal's value against its levels.
const (
	OK       = "ok"
	Warn     = "warn"
	Critical = "critical"
)

// Units a signal's value is in, for display.
var units = map[string]bool{"ratio": true, "count": true, "seconds": true, "bytes": true}

var idPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,40}[a-z0-9])?$`)

// Config is the reliability block of the chart values.
type Config struct {
	Builtin Builtin  `json:"builtin"`
	Signals []Signal `json:"signals"`
	// Errors are problems with the configuration; signals with errors are left out of Signals.
	Errors []string `json:"-"`
}

// Builtin tunes the signals resourcerer computes itself. OOM kills are always on: the memory recommendation
// depends on them.
type Builtin struct {
	Throttling     Toggle `json:"throttling"`
	MemoryHeadroom Toggle `json:"memoryHeadroom"`
	Restarts       Toggle `json:"restarts"`
	CrashLoop      Toggle `json:"crashLoop"`
	HPAAtMax       Toggle `json:"hpaAtMax"`
	BestEffort     Toggle `json:"bestEffort"`
}

// Toggle turns a built-in signal on or off and sets its warning level, where it has one.
type Toggle struct {
	Enabled *bool    `json:"enabled,omitempty"`
	Warn    *float64 `json:"warn,omitempty"`
}

// On reports whether the signal is enabled; unset means enabled.
func (t Toggle) On() bool { return t.Enabled == nil || *t.Enabled }

// WarnOr is the warning level, or def when unset.
func (t Toggle) WarnOr(def float64) float64 {
	if t.Warn == nil {
		return def
	}
	return *t.Warn
}

// Levels are the values from which a signal warns and turns critical; either may be unset.
type Levels struct {
	Warn     *float64 `json:"warn,omitempty"`
	Critical *float64 `json:"critical,omitempty"`
}

// Signal is a PromQL template evaluated per workload. The query returns the signal's value over {{ .Window }}:
// one series, or one per container label.
type Signal struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Unit        string `json:"unit"`
	// Above flags values over its levels, Below values under them; exactly one is set.
	Above   *Levels `json:"above,omitempty"`
	Below   *Levels `json:"below,omitempty"`
	Query   string  `json:"query"`
	Runbook string  `json:"runbook,omitempty"`

	tmpl *template.Template
}

// Vars fill a signal's query template.
type Vars struct {
	Namespace, Name, Kind string
	// Window is a PromQL duration, such as 7d.
	Window string
	// OwnedPods is a vector of 1s per pod of the workload, keyed by namespace and pod.
	OwnedPods string
}

// VarsFor fills the template variables for a workload and window.
func VarsFor(r prom.Ref, window time.Duration) Vars {
	return Vars{Namespace: r.Namespace, Name: r.Name, Kind: r.Kind, Window: model.Duration(window).String(), OwnedPods: prom.Owner(r)}
}

var cache sync.Map

// Parse reads the reliability configuration, keeping every valid signal and reporting the others in Errors.
// Results are cached by content: the controller parses it on every reconcile.
func Parse(data string) Config {
	if v, ok := cache.Load(data); ok {
		return v.(Config)
	}
	var c Config
	if err := yaml.UnmarshalStrict([]byte(data), &c); err != nil {
		c = Config{Errors: []string{fmt.Sprintf("reliability configuration: %v", err)}}
	}
	seen := map[string]bool{}
	valid := c.Signals[:0]
	for i, s := range c.Signals {
		if err := s.validate(seen); err != nil {
			c.Errors = append(c.Errors, fmt.Sprintf("signal %d (%s): %v", i+1, s.ID, err))
			continue
		}
		if len(valid) == MaxSignals {
			c.Errors = append(c.Errors, fmt.Sprintf("signal %s: more than %d signals, left out", s.ID, MaxSignals))
			continue
		}
		valid = append(valid, s)
	}
	c.Signals = valid
	cache.Store(data, c)
	return c
}

func (s *Signal) validate(seen map[string]bool) error {
	switch {
	case !idPattern.MatchString(s.ID):
		return errors.New("id must be lowercase letters, digits and dashes")
	case seen[s.ID]:
		return errors.New("duplicate id")
	case s.Name == "":
		return errors.New("name is required")
	case !units[s.Unit]:
		return errors.New("unit must be ratio, count, seconds or bytes")
	case (s.Above == nil) == (s.Below == nil):
		return errors.New("set exactly one of above and below")
	case strings.TrimSpace(s.Query) == "":
		return errors.New("query is required")
	}
	l := s.levels()
	if l.Warn == nil && l.Critical == nil {
		return errors.New("set warn, critical or both")
	}
	t, err := template.New(s.ID).Option("missingkey=error").Parse(s.Query)
	if err != nil {
		return fmt.Errorf("query template: %w", err)
	}
	if _, err := s.render(t, Vars{Namespace: "ns", Name: "name", Kind: "Deployment", Window: "5m", OwnedPods: "vector(1)"}); err != nil {
		return fmt.Errorf("query template: %w", err)
	}
	s.tmpl = t
	seen[s.ID] = true
	return nil
}

func (s Signal) levels() Levels {
	if s.Above != nil {
		return *s.Above
	}
	return *s.Below
}

// Overrides are per-namespace or per-workload changes to the signals, from annotations.
type Overrides struct {
	Disabled map[string]bool
	Levels   map[string]Levels
}

const (
	annotationPrefix   = "resourcerer.io/reliability."
	annotationDisabled = annotationPrefix + "disabled"
)

// OverridesFrom merges annotations: signals disabled on the namespace or the workload are off, and the workload's
// levels win over the namespace's.
func OverridesFrom(annotations ...map[string]string) Overrides {
	o := Overrides{Disabled: map[string]bool{}, Levels: map[string]Levels{}}
	for _, ann := range annotations {
		for k, v := range ann {
			if !strings.HasPrefix(k, annotationPrefix) {
				continue
			}
			if k == annotationDisabled {
				for _, id := range strings.Split(v, ",") {
					if id = strings.TrimSpace(id); id != "" {
						o.Disabled[id] = true
					}
				}
				continue
			}
			id, level, ok := strings.Cut(strings.TrimPrefix(k, annotationPrefix), ".")
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if !ok || err != nil {
				continue
			}
			l := o.Levels[id]
			switch level {
			case "warn":
				l.Warn = &f
			case "critical":
				l.Critical = &f
			default:
				continue
			}
			o.Levels[id] = l
		}
	}
	return o
}

// LevelsWith returns the signal's levels with the overrides applied.
func (s Signal) LevelsWith(o Overrides) Levels {
	l := s.levels()
	if v, ok := o.Levels[s.ID]; ok {
		if v.Warn != nil {
			l.Warn = v.Warn
		}
		if v.Critical != nil {
			l.Critical = v.Critical
		}
	}
	return l
}

// Grade compares a value with the signal's levels.
func (s Signal) Grade(v float64, l Levels) string {
	past := func(t *float64) bool {
		if t == nil {
			return false
		}
		if s.Above != nil {
			return v >= *t
		}
		return v <= *t
	}
	switch {
	case past(l.Critical):
		return Critical
	case past(l.Warn):
		return Warn
	}
	return OK
}

// worse reports whether a is a worse value than b for the signal.
func (s Signal) worse(a, b float64) bool {
	if s.Above != nil {
		return a > b
	}
	return a < b
}

// Render fills the signal's query for a workload and window.
func (s Signal) Render(v Vars) (string, error) { return s.render(s.tmpl, v) }

func (s Signal) render(t *template.Template, v Vars) (string, error) {
	if t == nil {
		return "", errors.New("signal not validated")
	}
	var b bytes.Buffer
	if err := t.Execute(&b, v); err != nil {
		return "", err
	}
	return b.String(), nil
}

// Result is a signal's worst value for a workload, and the container it came from when the query breaks it down.
type Result struct {
	Value     float64
	Container string
	Status    string
}

// Evaluate runs the signal over the window ending now. No data means the signal does not apply: nil, no error.
func (s Signal) Evaluate(ctx context.Context, p *prom.Client, v Vars, l Levels) (*Result, error) {
	q, err := s.Render(v)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, QueryTimeout)
	defer cancel()
	vals, err := p.Instant(ctx, q, "container")
	if err != nil {
		return nil, err
	}
	var r *Result
	for c, val := range vals {
		if r == nil || s.worse(val, r.Value) {
			r = &Result{Value: val, Container: c}
		}
	}
	if r != nil {
		r.Status = s.Grade(r.Value, l)
	}
	return r, nil
}

// Format shows a value in the signal's unit.
func Format(unit string, v float64) string {
	switch unit {
	case "ratio":
		return strconv.FormatFloat(v*100, 'f', precision(v*100), 64) + "%"
	case "seconds":
		if v < 1 {
			return strconv.FormatFloat(v*1000, 'f', 0, 64) + "ms"
		}
		return strconv.FormatFloat(v, 'f', precision(v), 64) + "s"
	case "bytes":
		return strconv.FormatFloat(v/(1<<20), 'f', 0, 64) + "MiB"
	}
	return strconv.FormatFloat(v, 'f', precision(v), 64)
}

func precision(v float64) int {
	switch a := math.Abs(v); {
	case a == math.Trunc(a) || a >= 100:
		return 0
	case a >= 10:
		return 1
	}
	return 2
}
