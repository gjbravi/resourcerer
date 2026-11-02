// Package policy decides when an apply may start, from the ApplyPolicy objects that govern a workload.
// Evaluate is pure: the applier gathers the workloads, policies and history once and calls it, so the
// rules are tested in tables instead of against a cluster.
package policy

import (
	"fmt"
	"time"

	// The windows carry IANA time zones; embedding the database keeps them working on any base image.
	_ "time/tzdata"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// Rules is one policy that may govern the workload, with the name the decision reports.
type Rules struct {
	// Name identifies the policy: Kind/namespace/name.
	Name string
	// NamespaceSelector is a ClusterApplyPolicy's, matched against Input.NamespaceLabels; nil for an
	// ApplyPolicy, which only ever reaches Evaluate for its own namespace.
	NamespaceSelector *metav1.LabelSelector
	Spec              v1alpha1.ApplyRules
}

// Input is everything Evaluate needs.
type Input struct {
	Now time.Time
	// Action is the kind of apply being decided.
	Action v1alpha1.AppliesTo
	// NamespaceLabels are the labels of the workload's namespace, matched by NamespaceSelector.
	NamespaceLabels map[string]string
	// Labels of the workload, matched by WorkloadSelector.
	Labels map[string]string
	// LastChange is when resourcerer last applied or reverted this workload; zero when it never did.
	LastChange time.Time
	// Rules are the policies that may govern the workload, most authoritative first, so the decision
	// reports the same policy for the same input.
	Rules []Rules
}

// Decision is when an apply may start.
type Decision struct {
	Allowed bool
	// BlockedBy is the field of the blocking policy, such as when.windows or limits.cooldown, so the
	// UI can point at it instead of parsing a message. "policy" when the policy cannot be interpreted.
	BlockedBy string
	// Policy is the Name of the policy that blocks; empty when Allowed.
	Policy string
	Reason string
	// NextAllowed is the first instant every enforcing policy allows an apply. Zero when Allowed, and
	// when no such instant is known: a policy that cannot be interpreted blocks until someone fixes it.
	NextAllowed time.Time
	// WouldBlock lists what Audit-mode policies would have blocked, as "<policy>: <reason>", in the
	// order of Input.Rules.
	WouldBlock []string
}

// FieldPolicy is the BlockedBy of a policy that cannot be interpreted.
const FieldPolicy = "policy"

const (
	// maxSteps bounds the search for the next instant every policy allows. Each step jumps to the end
	// of the latest blocker, so it takes as many steps as blockers hand over to one another.
	maxSteps = 64
	// searchDays is how many days ahead a window's opening is looked for: a weekly window opens within
	// seven.
	searchDays = 8
)

// Evaluate returns when an apply may start. A policy resourcerer cannot interpret, such as one with an
// unknown time zone or a malformed selector, blocks when it enforces: a policy that cannot be read must
// not open the door. In Audit mode it is reported in WouldBlock instead, since auditing never blocks.
func Evaluate(in Input) Decision {
	action := in.Action
	if action == "" {
		action = v1alpha1.AppliesToAuto
	}
	var d Decision
	var enforcing []compiled
	var broken bool
	for _, r := range in.Rules {
		if !governs(r.Spec, action) {
			continue
		}
		audit := r.Spec.Mode == v1alpha1.PolicyModeAudit
		c, selected, err := compile(r, in.NamespaceLabels, in.Labels)
		if err != nil {
			if audit {
				d.WouldBlock = append(d.WouldBlock, r.Name+": "+err.Error())
				continue
			}
			// Reported over any other blocker: it is the one that does not end on its own.
			if !broken {
				broken = true
				d.BlockedBy, d.Policy, d.Reason = FieldPolicy, r.Name, err.Error()
			}
			continue
		}
		if !selected {
			continue
		}
		b, blocks := c.blocksAt(in.Now, in.LastChange)
		if audit {
			if blocks {
				d.WouldBlock = append(d.WouldBlock, r.Name+": "+b.reason)
			}
			continue
		}
		enforcing = append(enforcing, c)
		if blocks && d.Policy == "" {
			d.BlockedBy, d.Policy, d.Reason = b.field, r.Name, b.reason
		}
	}
	if broken {
		return d
	}
	if d.Policy == "" {
		d.Allowed = true
		return d
	}
	d.NextAllowed = nextAllowed(enforcing, in.Now, in.LastChange)
	return d
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

// selects reports whether the labels match the selector; an absent selector matches everything. A
// selector that cannot be parsed is an error, not a mismatch: a mismatch would quietly lift the policy.
func selects(sel *metav1.LabelSelector, l map[string]string) (bool, error) {
	if sel == nil {
		return true, nil
	}
	s, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return false, err
	}
	return s.Matches(labels.Set(l)), nil
}

// compiled is a policy with its times parsed and its time zones loaded, once per evaluation instead of
// at every instant tried.
type compiled struct {
	spec    v1alpha1.ApplyRules
	windows []window
}

// window is an ApplyWindow with its times in minutes since local midnight.
type window struct {
	days     [7]bool
	from, to int
	loc      *time.Location
}

type block struct {
	field  string
	reason string
	// next is when this condition stops blocking; zero when it is not known.
	next time.Time
}

func compile(r Rules, nsLabels, wlLabels map[string]string) (compiled, bool, error) {
	ns, err := selects(r.NamespaceSelector, nsLabels)
	if err != nil {
		return compiled{}, false, fmt.Errorf("namespaceSelector: %w", err)
	}
	wl, err := selects(r.Spec.WorkloadSelector, wlLabels)
	if err != nil {
		return compiled{}, false, fmt.Errorf("workloadSelector: %w", err)
	}
	c := compiled{spec: r.Spec}
	if w := r.Spec.When; w != nil {
		for _, aw := range w.Windows {
			cw, err := compileWindow(aw)
			if err != nil {
				return compiled{}, false, err
			}
			c.windows = append(c.windows, cw)
		}
	}
	return c, ns && wl, nil
}

func compileWindow(aw v1alpha1.ApplyWindow) (window, error) {
	w := window{loc: time.UTC}
	if aw.TimeZone != "" {
		l, err := time.LoadLocation(aw.TimeZone)
		if err != nil {
			return window{}, fmt.Errorf("unknown time zone %q", aw.TimeZone)
		}
		w.loc = l
	}
	var err error
	if w.from, err = minutes(aw.From); err != nil {
		return window{}, err
	}
	if w.to, err = minutes(aw.To); err != nil {
		return window{}, err
	}
	if w.from == w.to {
		return window{}, fmt.Errorf("the window from %s to %s never opens", aw.From, aw.To)
	}
	if len(aw.Days) == 0 {
		for i := range w.days {
			w.days[i] = true
		}
	}
	for _, s := range aw.Days {
		d, ok := weekdays[s]
		if !ok {
			return window{}, fmt.Errorf("%q is not a day such as Mon", s)
		}
		w.days[d] = true
	}
	return w, nil
}

var weekdays = map[string]time.Weekday{
	"Sun": time.Sunday, "Mon": time.Monday, "Tue": time.Tuesday, "Wed": time.Wednesday,
	"Thu": time.Thursday, "Fri": time.Friday, "Sat": time.Saturday,
}

// blocksAt reports why a policy blocks an apply at t, and when that reason stops holding.
func (c compiled) blocksAt(t, lastChange time.Time) (block, bool) {
	if w := c.spec.When; w != nil {
		for _, f := range w.Freezes {
			if !t.Before(f.From.Time) && t.Before(f.To.Time) {
				reason := "frozen until " + f.To.UTC().Format(time.RFC3339)
				if f.Reason != "" {
					reason = fmt.Sprintf("frozen: %s (until %s)", f.Reason, f.To.UTC().Format(time.RFC3339))
				}
				return block{field: "when.freezes", reason: reason, next: f.To.Time}, true
			}
		}
	}
	if len(c.windows) > 0 && !c.inWindows(t) {
		return block{field: "when.windows", reason: "outside the apply windows", next: c.nextOpening(t)}, true
	}
	if l := c.spec.Limits; l != nil && l.Cooldown != nil && !lastChange.IsZero() {
		until := lastChange.Add(l.Cooldown.Duration)
		if t.Before(until) {
			reason := fmt.Sprintf("changed less than %s ago, cooling down until %s", l.Cooldown.Duration, until.UTC().Format(time.RFC3339))
			return block{field: "limits.cooldown", reason: reason, next: until}, true
		}
	}
	return block{}, false
}

// nextAllowed is the first instant no enforcing policy blocks, counting the ones that allow now: a
// policy whose window is open now may be closed by the time another's freeze ends. Each step jumps to
// the latest instant a blocker stops blocking, then re-evaluates. Every blocker ends strictly after the
// instant it blocks, so the search only moves forward. Zero when it does not settle.
func nextAllowed(enforcing []compiled, now, lastChange time.Time) time.Time {
	t := now
	for range maxSteps {
		var next time.Time
		blocked := false
		for _, c := range enforcing {
			b, ok := c.blocksAt(t, lastChange)
			if !ok {
				continue
			}
			if b.next.IsZero() {
				return time.Time{}
			}
			blocked = true
			if b.next.After(next) {
				next = b.next
			}
		}
		if !blocked {
			return t
		}
		t = next
	}
	return time.Time{}
}

func (c compiled) inWindows(t time.Time) bool {
	for _, w := range c.windows {
		if w.contains(t) {
			return true
		}
	}
	return false
}

// nextOpening is the first instant after t inside any of the windows; zero when none opens within
// searchDays.
func (c compiled) nextOpening(t time.Time) time.Time {
	var best time.Time
	for _, w := range c.windows {
		if o := w.nextOpening(t); !o.IsZero() && (best.IsZero() || o.Before(best)) {
			best = o
		}
	}
	return best
}

// A window is decided from the instants it opens and closes, the way KEDA's cron scaler does: each
// listed local day has one occurrence, from time.Date(day, from) to time.Date(day, to), on the next day
// when to is earlier than from. Being inside and the next opening both come from those instants, so
// they always agree. On the day the clocks jump forward, a time in the skipped hour is whatever
// time.Date makes of it, an hour earlier or later depending on the zone; on the day they fall back, a
// repeated time is its first occurrence.

// occurrence is the window opening on the local day y-m-d; ok is false when it does not open that day.
func (w window) occurrence(y int, m time.Month, d int) (open, closed time.Time, ok bool) {
	if !w.days[time.Date(y, m, d, 12, 0, 0, 0, w.loc).Weekday()] {
		return time.Time{}, time.Time{}, false
	}
	open = time.Date(y, m, d, w.from/60, w.from%60, 0, 0, w.loc)
	closeDay := d
	if w.to < w.from {
		closeDay++
	}
	closed = time.Date(y, m, closeDay, w.to/60, w.to%60, 0, 0, w.loc)
	return open, closed, closed.After(open)
}

// contains reports whether t falls in the window: in the occurrence of its own local day, of the day
// before for an overnight window, or of the day after when a jump moved that opening back across
// midnight.
func (w window) contains(t time.Time) bool {
	y, m, d := t.In(w.loc).Date()
	for i := -1; i <= 1; i++ {
		if open, closed, ok := w.occurrence(y, m, d+i); ok && !t.Before(open) && t.Before(closed) {
			return true
		}
	}
	return false
}

// nextOpening is the first opening after t; the caller asks only when t is outside the window.
func (w window) nextOpening(t time.Time) time.Time {
	y, m, d := t.In(w.loc).Date()
	for i := -1; i <= searchDays; i++ {
		if open, _, ok := w.occurrence(y, m, d+i); ok && open.After(t) {
			return open
		}
	}
	return time.Time{}
}

func minutes(hhmm string) (int, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, fmt.Errorf("%q is not a time of day like 10:00", hhmm)
	}
	return t.Hour()*60 + t.Minute(), nil
}
