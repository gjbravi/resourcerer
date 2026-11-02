// Package policy decides when an apply may start, from the ApplyPolicy objects that govern a workload.
// Evaluate is pure: the applier gathers the workloads, policies and history once and calls it, so the
// rules are tested in tables instead of against a cluster.
package policy

import (
	"fmt"
	"time"

	// The windows carry IANA time zones, and the container image has no tz database of its own.
	_ "time/tzdata"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// Rules is one policy that governs the workload, with the name the decision reports.
type Rules struct {
	// Name identifies the policy: Kind/namespace/name.
	Name string
	Spec v1alpha1.ApplyRules
}

// Input is everything Evaluate needs.
type Input struct {
	Now time.Time
	// Action is the kind of apply being decided.
	Action v1alpha1.AppliesTo
	// Labels of the workload, matched by WorkloadSelector.
	Labels map[string]string
	// LastChange is when resourcerer last applied or reverted this workload; zero when it never did.
	LastChange time.Time
	// Rules are the policies that govern the workload, most authoritative first, so the decision
	// reports the same policy for the same input.
	Rules []Rules
}

// Decision is when an apply may start.
type Decision struct {
	Allowed bool
	// BlockedBy is the field of the blocking policy, such as when.windows or limits.cooldown, so the
	// UI can point at it instead of parsing a message.
	BlockedBy string
	// Policy is the Name of the policy that blocks; empty when Allowed.
	Policy string
	Reason string
	// NextAllowed is the first instant every enforcing policy allows an apply. Zero when Allowed.
	NextAllowed time.Time
	// WouldBlock lists what Audit-mode policies would have blocked, as "<policy>: <reason>".
	WouldBlock []string
}

type block struct {
	field  string
	reason string
	// next is when this condition stops blocking.
	next time.Time
}

const (
	// maxSteps bounds the search for the next instant every policy allows; each step jumps to the next
	// instant a blocker stops blocking, so the loop converges in a few steps in practice.
	maxSteps = 64
	// maxSearch is how far ahead a window is looked for: a window that opens once a week is found in
	// seven days, and a policy whose windows never open is reported as far away rather than forever.
	maxSearch = 14 * 24 * time.Hour
)

// Evaluate returns when an apply may start. An error is a policy resourcerer cannot interpret, such as
// an unknown time zone: it blocks, since a policy that cannot be read must not open the door.
func Evaluate(in Input) (Decision, error) {
	var d Decision
	var enforcing []Rules
	action := in.Action
	if action == "" {
		action = v1alpha1.AppliesToAuto
	}
	blocked := false
	for _, r := range in.Rules {
		if !governs(r.Spec, action) || !Selects(r.Spec.WorkloadSelector, in.Labels) {
			continue
		}
		b, ok, err := blocksAt(r.Spec, in.Now, in.LastChange)
		if err != nil {
			return Decision{BlockedBy: "policy", Policy: r.Name, Reason: err.Error()}, nil
		}
		if !ok {
			continue
		}
		if r.Spec.Mode == v1alpha1.PolicyModeAudit {
			d.WouldBlock = append(d.WouldBlock, r.Name+": "+b.reason)
			continue
		}
		enforcing = append(enforcing, r)
		if !blocked {
			blocked = true
			d.BlockedBy, d.Policy, d.Reason = b.field, r.Name, b.reason
		}
	}
	if !blocked {
		d.Allowed = true
		return d, nil
	}
	next, err := nextAllowed(in, enforcing)
	if err != nil {
		return Decision{BlockedBy: "policy", Policy: d.Policy, Reason: err.Error()}, nil
	}
	d.NextAllowed = next
	return d, nil
}

// governs reports whether a policy decides this kind of apply. A policy that lists nothing governs
// automatic applies only, so it never blocks a person without saying so.
func governs(spec v1alpha1.ApplyRules, action v1alpha1.AppliesTo) bool {
	if len(spec.AppliesTo) == 0 {
		return action == v1alpha1.AppliesToAuto
	}
	for _, a := range spec.AppliesTo {
		if a == action {
			return true
		}
	}
	return false
}

// Selects reports whether the labels match the selector; an absent selector matches everything.
func Selects(sel *metav1.LabelSelector, l map[string]string) bool {
	if sel == nil {
		return true
	}
	s, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return false
	}
	return s.Matches(labels.Set(l))
}

// blocksAt reports why a policy blocks an apply at t, and when that reason stops holding.
func blocksAt(spec v1alpha1.ApplyRules, t, lastChange time.Time) (block, bool, error) {
	if w := spec.When; w != nil {
		for _, f := range w.Freezes {
			if !t.Before(f.From.Time) && t.Before(f.To.Time) {
				reason := "frozen until " + f.To.UTC().Format(time.RFC3339)
				if f.Reason != "" {
					reason = fmt.Sprintf("frozen: %s (until %s)", f.Reason, f.To.UTC().Format(time.RFC3339))
				}
				return block{field: "when.freezes", reason: reason, next: f.To.Time}, true, nil
			}
		}
		if len(w.Windows) > 0 {
			inside, err := inWindows(w.Windows, t)
			if err != nil {
				return block{}, false, err
			}
			if !inside {
				next, err := nextWindow(w.Windows, t)
				if err != nil {
					return block{}, false, err
				}
				return block{field: "when.windows", reason: "outside the apply windows", next: next}, true, nil
			}
		}
	}
	if l := spec.Limits; l != nil && l.Cooldown != nil && !lastChange.IsZero() {
		until := lastChange.Add(l.Cooldown.Duration)
		if t.Before(until) {
			reason := fmt.Sprintf("changed less than %s ago, cooling down until %s", l.Cooldown.Duration, until.UTC().Format(time.RFC3339))
			return block{field: "limits.cooldown", reason: reason, next: until}, true, nil
		}
	}
	return block{}, false, nil
}

// nextAllowed is the first instant no enforcing policy blocks. Each step jumps to the latest instant a
// blocker stops blocking, then re-evaluates, since a policy that just allowed one condition may hit
// another (a freeze ending outside a window, for instance).
func nextAllowed(in Input, enforcing []Rules) (time.Time, error) {
	t := in.Now.Truncate(time.Minute).Add(time.Minute)
	for i := 0; i < maxSteps; i++ {
		var next time.Time
		for _, r := range enforcing {
			b, ok, err := blocksAt(r.Spec, t, in.LastChange)
			if err != nil {
				return time.Time{}, err
			}
			if ok && b.next.After(next) {
				next = b.next
			}
		}
		if next.IsZero() {
			return t, nil
		}
		if !next.After(t) {
			next = t.Add(time.Minute)
		}
		t = next
	}
	return t, nil
}

func inWindows(ws []v1alpha1.ApplyWindow, t time.Time) (bool, error) {
	for _, w := range ws {
		ok, err := windowContains(w, t)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

func nextWindow(ws []v1alpha1.ApplyWindow, t time.Time) (time.Time, error) {
	cursor := t.Truncate(time.Minute).Add(time.Minute)
	for i := time.Duration(0); i < maxSearch/time.Minute; i++ {
		ok, err := inWindows(ws, cursor)
		if err != nil {
			return time.Time{}, err
		}
		if ok {
			return cursor, nil
		}
		cursor = cursor.Add(time.Minute)
	}
	return t.Add(maxSearch), nil
}

// windowContains reports whether t falls in the window. A window that opens and closes on the same day
// covers [from, to) that day; one whose To is earlier than its From crosses midnight, opening on a
// listed day and closing on the next.
func windowContains(w v1alpha1.ApplyWindow, t time.Time) (bool, error) {
	loc := time.UTC
	if w.TimeZone != "" {
		l, err := time.LoadLocation(w.TimeZone)
		if err != nil {
			return false, fmt.Errorf("unknown time zone %q", w.TimeZone)
		}
		loc = l
	}
	from, err := minutes(w.From)
	if err != nil {
		return false, err
	}
	to, err := minutes(w.To)
	if err != nil {
		return false, err
	}
	local := t.In(loc)
	day := local.Weekday()
	minute := local.Hour()*60 + local.Minute()

	if from < to {
		if !listsDay(w.Days, day) {
			return false, nil
		}
		return minute >= from && minute < to, nil
	}
	if from == to {
		// Validation rejects it; treat an equal pair as a window that never opens rather than all day.
		return false, nil
	}
	if listsDay(w.Days, day) && minute >= from {
		return true, nil
	}
	// The tail of an overnight window opened the day before.
	if listsDay(w.Days, time.Weekday((int(day)+6)%7)) && minute < to {
		return true, nil
	}
	return false, nil
}

func listsDay(days []string, d time.Weekday) bool {
	if len(days) == 0 {
		return true
	}
	want := [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}[d]
	for _, s := range days {
		if s == want {
			return true
		}
	}
	return false
}

func minutes(hhmm string) (int, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, fmt.Errorf("%q is not a time of day like 10:00", hhmm)
	}
	return t.Hour()*60 + t.Minute(), nil
}
