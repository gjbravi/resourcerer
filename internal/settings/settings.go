// Package settings resolves optimization settings from cluster defaults, namespace
// annotations and workload annotations, in that order of precedence.
package settings

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

const Domain = "resourcerer.io"

// LabelEnabled opts a namespace or workload in ("true") or out ("false").
const LabelEnabled = Domain + "/enabled"

const (
	SourceCluster   = "cluster"
	SourceNamespace = "namespace"
	SourceWorkload  = "workload"
)

// Definition describes one setting. Key is the annotation suffix after resourcerer.io/.
type Definition struct {
	Key     string
	Default string
	// Options, when set, are the only accepted values.
	Options  []string
	Validate func(string) error
}

func (d Definition) Annotation() string { return Domain + "/" + d.Key }

const (
	CPUGoal        = "cpu.optimization-goal"
	MemoryGoal     = "memory.optimization-goal"
	LearningPeriod = "learning-period"
	Schedule       = "schedule"
	AutoDeploy     = "auto-deploy"
	AutoThreshold  = "auto-deploy.threshold"
	ApplyMethod    = "apply-method"
	CPULimits      = "cpu.limits-policy"
	MemoryLimits   = "memory.limits-policy"
	CPURatio       = "cpu.limit-request-ratio"
	MemoryRatio    = "memory.limit-request-ratio"
	CPUMin         = "cpu.requests.min"
	MemoryMin      = "memory.requests.min"
	SkipContainers = "containers.skip"
	HPAPreserve    = "hpa.preserve-scale-point"
)

var (
	goals    = []string{"Savings", "Balanced", "Reliability"}
	policies = []string{"RequestsOnly", "RequestsRaiseLimitsIfNeeded", "RequestsAndLimits", "DoNotOptimize"}
)

// All lists every setting in display order.
var All = []Definition{
	{Key: CPUGoal, Default: "Balanced", Options: goals},
	{Key: MemoryGoal, Default: "Balanced", Options: goals},
	{Key: LearningPeriod, Default: "P7D", Validate: func(s string) error { _, err := ParseDuration(s); return err }},
	{Key: Schedule, Default: "@daily", Validate: func(s string) error { _, err := ParseSchedule(s); return err }},
	{Key: AutoDeploy, Default: "Disabled", Options: []string{"Enabled", "Disabled"}},
	{Key: AutoThreshold, Default: "10%", Validate: func(s string) error { _, err := ParsePercent(s); return err }},
	{Key: ApplyMethod, Default: "PatchImmediateRollout", Options: []string{"PatchImmediateRollout"}},
	{Key: CPULimits, Default: "RequestsRaiseLimitsIfNeeded", Options: policies},
	{Key: MemoryLimits, Default: "RequestsRaiseLimitsIfNeeded", Options: policies},
	{Key: CPURatio, Default: "2.0", Validate: validateRatio},
	{Key: MemoryRatio, Default: "1.5", Validate: validateRatio},
	{Key: CPUMin, Default: "10m", Validate: validateQuantity},
	{Key: MemoryMin, Default: "16Mi", Validate: validateQuantity},
	{Key: SkipContainers, Default: "istio-proxy,linkerd-proxy"},
	{Key: HPAPreserve, Default: "true", Options: []string{"true", "false"}},
}

func Lookup(key string) (Definition, bool) {
	i := slices.IndexFunc(All, func(d Definition) bool { return d.Key == key })
	if i < 0 {
		return Definition{}, false
	}
	return All[i], true
}

func (d Definition) Check(v string) error {
	if len(d.Options) > 0 && !slices.Contains(d.Options, v) {
		return fmt.Errorf("want one of %s", strings.Join(d.Options, ", "))
	}
	if d.Validate != nil {
		return d.Validate(v)
	}
	return nil
}

type Value struct {
	Value  string
	Source string
}

// Effective holds resolved settings plus warnings for invalid values that were ignored.
type Effective struct {
	Values   map[string]Value
	Warnings []string
}

func (e Effective) Get(key string) string { return e.Values[key].Value }

// Resolve applies cluster defaults, then namespace annotations, then workload annotations.
// Invalid values are skipped with a warning, so the previous level wins.
func Resolve(defaults, nsAnnotations, wlAnnotations map[string]string) Effective {
	e := Effective{Values: map[string]Value{}}
	for _, d := range All {
		v := Value{Value: d.Default, Source: SourceCluster}
		if s, ok := defaults[d.Key]; ok && d.Check(s) == nil {
			v.Value = s
		}
		for _, lvl := range []struct {
			src string
			ann map[string]string
		}{{SourceNamespace, nsAnnotations}, {SourceWorkload, wlAnnotations}} {
			raw, ok := lvl.ann[d.Annotation()]
			if !ok {
				continue
			}
			raw = strings.TrimSpace(raw)
			if err := d.Check(raw); err != nil {
				e.Warnings = append(e.Warnings, fmt.Sprintf("ignoring %s annotation %s=%q: %v", lvl.src, d.Annotation(), raw, err))
				continue
			}
			v = Value{Value: raw, Source: lvl.src}
		}
		e.Values[d.Key] = v
	}
	return e
}

// ParseDuration accepts the ISO 8601 subset used by settings: P<n>D, PT<n>H, PT<n>M, P<n>DT<n>H.
func ParseDuration(s string) (time.Duration, error) {
	if !strings.HasPrefix(s, "P") {
		return 0, fmt.Errorf("%q is not an ISO 8601 duration like P7D or PT30M", s)
	}
	rest := s[1:]
	var d time.Duration
	inTime := false
	num := ""
	for _, r := range rest {
		switch {
		case r == 'T':
			inTime = true
		case r >= '0' && r <= '9':
			num += string(r)
		default:
			n, err := strconv.Atoi(num)
			if err != nil {
				return 0, fmt.Errorf("%q is not an ISO 8601 duration", s)
			}
			num = ""
			switch {
			case r == 'D' && !inTime:
				d += time.Duration(n) * 24 * time.Hour
			case r == 'W' && !inTime:
				d += time.Duration(n) * 7 * 24 * time.Hour
			case r == 'H' && inTime:
				d += time.Duration(n) * time.Hour
			case r == 'M' && inTime:
				d += time.Duration(n) * time.Minute
			default:
				return 0, fmt.Errorf("%q is not a supported ISO 8601 duration", s)
			}
		}
	}
	if num != "" || d <= 0 {
		return 0, fmt.Errorf("%q is not a positive ISO 8601 duration", s)
	}
	return d, nil
}

// ParseSchedule accepts @daily, @hourly, @weekly and "@every <go duration>". It returns the period.
func ParseSchedule(s string) (time.Duration, error) {
	switch s {
	case "@hourly":
		return time.Hour, nil
	case "@daily":
		return 24 * time.Hour, nil
	case "@weekly":
		return 7 * 24 * time.Hour, nil
	}
	if d, ok := strings.CutPrefix(s, "@every "); ok {
		p, err := time.ParseDuration(strings.TrimSpace(d))
		if err != nil || p < time.Minute {
			return 0, fmt.Errorf("%q: want @every with a duration of at least 1m", s)
		}
		return p, nil
	}
	return 0, fmt.Errorf("%q: want @hourly, @daily, @weekly or @every <duration>", s)
}

func ParsePercent(s string) (float64, error) {
	v, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
	if err != nil || v < 0 || v > 100 || !strings.HasSuffix(s, "%") {
		return 0, fmt.Errorf("%q: want a percentage like 10%%", s)
	}
	return v / 100, nil
}

func validateRatio(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil || v < 1 {
		return fmt.Errorf("%q: want a number >= 1", s)
	}
	return nil
}

func validateQuantity(s string) error {
	_, err := resource.ParseQuantity(s)
	return err
}

func (e Effective) Float(key string) float64 {
	v, _ := strconv.ParseFloat(e.Get(key), 64)
	return v
}

func (e Effective) Quantity(key string) resource.Quantity {
	q, _ := resource.ParseQuantity(e.Get(key))
	return q
}

func (e Effective) Duration(key string) time.Duration {
	d, _ := ParseDuration(e.Get(key))
	return d
}

func (e Effective) SkipList() []string {
	v := e.Get(SkipContainers)
	if v == "" || v == "none" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
