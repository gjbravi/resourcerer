// Package policy decides when an apply may start, from the apply policies that govern a workload.
// Decide is pure: the applier reads the policies, the workloads and their history once and calls it, so
// the rules are tested in tables and with a clock instead of against a cluster.
package policy

import (
	"fmt"
	"hash/fnv"
	"sort"
	"strings"
	"time"

	// Windows carry IANA time zones and the container image has no tz database of its own.
	_ "time/tzdata"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// Policy is one apply policy that may govern a workload.
type Policy struct {
	// Name identifies it in the decision: Kind/name, or Kind/namespace/name for a namespaced policy.
	Name string
	// Namespace is set on an ApplyPolicy only, which governs its own namespace.
	Namespace string
	// Namespaces selects namespaces, on a ClusterApplyPolicy only.
	Namespaces *metav1.LabelSelector
	Spec       v1alpha1.ApplyRules
}

// Cluster reports whether the policy is cluster-scoped, which decides the order the decision names them in.
func (p Policy) Cluster() bool { return p.Namespace == "" }

// Workload is one managed workload as the rate limits see it.
type Workload struct {
	Namespace, Kind, Name string
	// Labels are the workload's and its pod template's; NamespaceLabels select cluster policies.
	Labels          map[string]string
	NamespaceLabels map[string]string
	// AutoDeploy is whether the auto-deploy setting is on: the base a percentage bound applies to.
	AutoDeploy bool
	// Applying is an apply of ours whose rollout has not finished; RollingOut is any rollout in flight,
	// whoever started it.
	Applying   bool
	RollingOut bool
	// Applied are the apply times within the last hour, for maxPerHour.
	Applied []time.Time
	// Changed is the last apply or revert, which the cooldown counts from.
	Changed time.Time
}

// Key identifies the workload, as the offset of a spread is computed from.
func (w Workload) Key() string { return w.Namespace + "/" + w.Kind + "/" + w.Name }

// Input is everything Decide needs.
type Input struct {
	Now      time.Time
	Target   Workload
	Policies []Policy
	// Fleet is the managed workloads a rate counts, cluster-wide, the target included. Rate limits are
	// ignored without it, so a caller that needs none can leave it out.
	Fleet []Workload
}

// Result is the decision for each kind of apply, and when to look again.
type Result struct {
	Auto   v1alpha1.ApplyDecision
	Manual v1alpha1.ApplyDecision
	// NextChange is the earliest instant a rule changes (a window opening or closing, a freeze, a
	// cooldown ending), or zero when nothing is scheduled. Rate limits waiting for a slot have none.
	NextChange time.Time
}

// Decide returns whether an automatic and a manual apply may start now. Both are decided apart, because
// "auto-deploy is held, a person may still apply" is a common state.
func Decide(in Input) Result {
	var res Result
	sel := make([]compiled, 0, len(in.Policies))
	for _, p := range in.Policies {
		c := compile(p)
		if c.err == nil && !c.matches(in.Target) {
			continue
		}
		sel = append(sel, c)
	}
	// The same rules always name the same policy: cluster policies first, then by namespace and name.
	sort.SliceStable(sel, func(i, j int) bool {
		if sel[i].p.Cluster() != sel[j].p.Cluster() {
			return sel[i].p.Cluster()
		}
		return sel[i].p.Name < sel[j].p.Name
	})
	res.Auto, res.NextChange = decideKind(sel, in, v1alpha1.ApplyKindAuto)
	var manualAt time.Time
	res.Manual, manualAt = decideKind(sel, in, v1alpha1.ApplyKindManual)
	res.NextChange = earliest(res.NextChange, manualAt)
	return res
}

// Validate reports why a policy cannot be interpreted, so the applier can publish it instead of only
// failing closed on every workload it governs.
func Validate(p Policy) error { return compile(p).err }

// Slots reports how many applies the policy's maxConcurrent still allows, and how many it allows in all, so
// the applier can export it the way Karpenter exports allowed disruptions.
func Slots(p Policy, in Input) (available, limit int, ok bool) {
	c := compile(p)
	if c.err != nil || c.rate.maxConcurrent == nil {
		return 0, 0, false
	}
	base, applying := 0, 0
	for _, w := range in.Fleet {
		if !c.matches(w) {
			continue
		}
		if w.AutoDeploy {
			base++
		}
		if w.Applying {
			applying++
		}
	}
	limit, err := resolve(*c.rate.maxConcurrent, base)
	if err != nil {
		return 0, 0, false
	}
	return max(0, limit-applying), limit, true
}

// Selects reports whether the policy governs the workload's namespace and labels. A policy that cannot be
// read selects everything: Decide is what holds the applies then.
func Selects(p Policy, w Workload) bool {
	c := compile(p)
	if c.err != nil {
		return true
	}
	return c.matches(w)
}

// block is one rule that holds an apply, and when it stops holding.
type block struct {
	field  string
	reason string
	// next is when the rule stops blocking; zero when that depends on other rollouts finishing.
	next time.Time
}

func decideKind(sel []compiled, in Input, kind v1alpha1.ApplyKind) (v1alpha1.ApplyDecision, time.Time) {
	d := v1alpha1.ApplyDecision{State: v1alpha1.ApplyAllowed}
	var enforcing []compiled
	var nextChange time.Time
	for _, c := range sel {
		if !governs(c.p.Spec, kind) {
			continue
		}
		if c.err != nil {
			if c.p.Spec.Mode == v1alpha1.ApplyPolicyAudit {
				d.WouldBlock = append(d.WouldBlock, c.p.Name+": "+c.err.Error())
				continue
			}
			d.State = v1alpha1.ApplyBlocked
			d.Policy, d.BlockedBy, d.Reason = c.p.Name, "policy", c.err.Error()
			return d, nextChange
		}
		nextChange = earliest(nextChange, c.nextChange(in, in.Now))
		if b, blocked := c.blocks(kind, in, in.Now); blocked {
			if c.p.Spec.Mode == v1alpha1.ApplyPolicyAudit {
				d.WouldBlock = append(d.WouldBlock, c.p.Name+": "+b.reason)
				continue
			}
			enforcing = append(enforcing, c)
			if !hasBlocked(d) {
				d.State = v1alpha1.ApplyBlocked
				d.Policy, d.BlockedBy, d.Reason = c.p.Name, b.field, b.reason
			}
		}
	}
	if len(enforcing) == 0 {
		return d, nextChange
	}
	// Every enforcing policy has to allow the apply, so the instant to report is the first one none of
	// them blocks; a rule that waits for another rollout to finish has none.
	if at, ok := nextAllowed(in, enforcing, kind); ok {
		d.NextAllowed = &metav1.Time{Time: at}
	}
	return d, nextChange
}

func hasBlocked(d v1alpha1.ApplyDecision) bool { return d.State == v1alpha1.ApplyBlocked }

// maxSteps bounds the search for the instant every policy allows: each step jumps to the end of a blocker,
// so it converges in a few steps and the bound only guards a policy whose windows never open.
const maxSteps = 64

func nextAllowed(in Input, enforcing []compiled, kind v1alpha1.ApplyKind) (time.Time, bool) {
	t := in.Now.Truncate(time.Second).Add(time.Second)
	for range maxSteps {
		var jump time.Time
		for _, c := range enforcing {
			b, blocked := c.blocks(kind, in, t)
			if !blocked {
				continue
			}
			if b.next.IsZero() {
				return time.Time{}, false
			}
			if b.next.After(jump) {
				jump = b.next
			}
		}
		if jump.IsZero() {
			return t, true
		}
		if !jump.After(t) {
			jump = t.Add(time.Second)
		}
		t = jump
	}
	return t, true
}

/* ── One policy ──────────────────────────────────────────────────────────── */

type compiled struct {
	p     Policy
	err   error
	nsSel labels.Selector
	wlSel labels.Selector

	windows []window
	freezes []v1alpha1.ApplyFreeze
	rate    rate
}

type rate struct {
	// maxConcurrent and perNamespace are a whole number or a percentage; nil means unset.
	maxConcurrent *intstr.IntOrString
	maxPerHour    *int32
	perNamespace  *intstr.IntOrString
	cooldown      time.Duration
	spread        time.Duration
}

func compile(p Policy) compiled {
	c := compiled{p: p}
	spec := p.Spec
	if p.Namespaces != nil {
		if c.nsSel, c.err = metav1.LabelSelectorAsSelector(p.Namespaces); c.err != nil {
			c.err = fmt.Errorf("namespaceSelector: %w", c.err)
			return c
		}
	}
	if spec.WorkloadSelector != nil {
		if c.wlSel, c.err = metav1.LabelSelectorAsSelector(spec.WorkloadSelector); c.err != nil {
			c.err = fmt.Errorf("workloadSelector: %w", c.err)
			return c
		}
	}
	for _, a := range spec.AppliesTo {
		switch a {
		case v1alpha1.ApplyKindAuto, v1alpha1.ApplyKindManual:
		default:
			c.err = fmt.Errorf("appliesTo: %q is neither Auto nor Manual", a)
			return c
		}
	}
	switch spec.Mode {
	case "", v1alpha1.ApplyPolicyEnforce, v1alpha1.ApplyPolicyAudit:
	default:
		c.err = fmt.Errorf("mode: %q is neither Enforce nor Audit", spec.Mode)
		return c
	}
	for i, w := range spec.Windows {
		parsed, err := parseWindow(w)
		if err != nil {
			c.err = fmt.Errorf("windows[%d]: %w", i, err)
			return c
		}
		c.windows = append(c.windows, parsed)
	}
	for i, f := range spec.Freezes {
		if !f.To.After(f.From.Time) {
			c.err = fmt.Errorf("freezes[%d]: to must be after from", i)
			return c
		}
		c.freezes = append(c.freezes, f)
	}
	if r := spec.Rate; r != nil {
		c.rate.maxConcurrent = r.MaxConcurrent
		c.rate.maxPerHour = r.MaxPerHour
		if r.PerNamespace != nil {
			c.rate.perNamespace = r.PerNamespace.MaxConcurrent
		}
		if r.Cooldown != nil {
			c.rate.cooldown = r.Cooldown.Duration
		}
		if r.Spread != nil {
			c.rate.spread = r.Spread.Duration
		}
		if c.err = c.rate.validate(); c.err != nil {
			c.err = fmt.Errorf("rate: %w", c.err)
			return c
		}
	}
	return c
}

func (r rate) validate() error {
	for _, f := range []struct {
		name string
		v    *intstr.IntOrString
	}{{"maxConcurrent", r.maxConcurrent}, {"perNamespace.maxConcurrent", r.perNamespace}} {
		if f.v == nil {
			continue
		}
		if _, err := resolve(*f.v, 100); err != nil {
			return fmt.Errorf("%s: %w", f.name, err)
		}
	}
	if r.maxPerHour != nil && *r.maxPerHour < 0 {
		return fmt.Errorf("maxPerHour: %d is negative", *r.maxPerHour)
	}
	if r.cooldown < 0 {
		return fmt.Errorf("cooldown: %s is negative", r.cooldown)
	}
	if r.spread < 0 {
		return fmt.Errorf("spread: %s is negative", r.spread)
	}
	return nil
}

// resolve reads a bound that is a whole number or a percentage of base, rounded up as Karpenter and
// PodDisruptionBudgets do, so 10% of 5 workloads still allows one.
func resolve(v intstr.IntOrString, base int) (int, error) {
	if v.Type == intstr.Int {
		if v.IntVal < 0 {
			return 0, fmt.Errorf("%d is negative", v.IntVal)
		}
		return int(v.IntVal), nil
	}
	s := strings.TrimSuffix(strings.TrimSpace(v.StrVal), "%")
	var pct int
	if _, err := fmt.Sscanf(s, "%d", &pct); err != nil || !strings.HasSuffix(strings.TrimSpace(v.StrVal), "%") {
		return 0, fmt.Errorf("%q is neither a whole number nor a percentage", v.StrVal)
	}
	if pct < 0 || pct > 100 {
		return 0, fmt.Errorf("%q is outside 0-100%%", v.StrVal)
	}
	return (pct*base + 99) / 100, nil
}

// governs reports whether the policy decides this kind of apply, read from the spec so a policy that
// cannot be parsed still holds the applies it says it covers. Empty means automatic applies only.
func governs(spec v1alpha1.ApplyRules, kind v1alpha1.ApplyKind) bool {
	if len(spec.AppliesTo) == 0 {
		return kind == v1alpha1.ApplyKindAuto
	}
	for _, a := range spec.AppliesTo {
		if a == kind {
			return true
		}
	}
	return false
}

func (c compiled) matches(w Workload) bool {
	if c.p.Namespace != "" && c.p.Namespace != w.Namespace {
		return false
	}
	if c.nsSel != nil && !c.nsSel.Matches(labels.Set(w.NamespaceLabels)) {
		return false
	}
	if c.wlSel != nil && !c.wlSel.Matches(labels.Set(w.Labels)) {
		return false
	}
	return true
}

// blocks reports why the policy holds an apply at now, and when the rule stops.
func (c compiled) blocks(kind v1alpha1.ApplyKind, in Input, now time.Time) (block, bool) {
	// A freeze is the strongest promise a team makes: it holds both kinds of apply, windows or not.
	if f, ok := c.frozen(now); ok {
		reason := "frozen until " + f.To.UTC().Format(time.RFC3339)
		if f.Reason != "" {
			reason = fmt.Sprintf("frozen: %s (until %s)", f.Reason, f.To.UTC().Format(time.RFC3339))
		}
		return block{field: "freezes", reason: reason, next: f.To.Time}, true
	}
	span, inside := c.windowAt(now)
	if len(c.windows) > 0 && !inside {
		next := c.nextOpening(now)
		return block{field: "windows", reason: "outside the apply windows (" + c.describeWindows() + ")", next: next}, true
	}
	if kind == v1alpha1.ApplyKindManual {
		// A person is not a queue: the rules about how many run at once never hold a reviewed apply.
		return block{}, false
	}
	if in.Target.RollingOut {
		return block{field: "rollout", reason: "its rollout has not finished"}, true
	}
	if c.rate.spread > 0 && inside {
		if at := spreadStart(in.Target.Key(), c.opening(span, now), span.close, c.rate.spread); now.Before(at) {
			return block{field: "rate.spread", reason: "held until its offset in the window opens (" + at.UTC().Format(time.RFC3339) + ")", next: at}, true
		}
	}
	if b, ok := c.slot(in, now); ok {
		return b, true
	}
	if c.rate.cooldown > 0 && !in.Target.Changed.IsZero() {
		if until := in.Target.Changed.Add(c.rate.cooldown); now.Before(until) {
			return block{field: "rate.cooldown", reason: fmt.Sprintf("changed less than %s ago, cooling down until %s",
				c.rate.cooldown, until.UTC().Format(time.RFC3339)), next: until}, true
		}
	}
	return block{}, false
}

// slot reports a rate limit that has no room for another apply. The cooldown is a separate rule because it
// holds manual applies too.
func (c compiled) slot(in Input, now time.Time) (block, bool) {
	if c.rate.maxConcurrent == nil && c.rate.maxPerHour == nil && c.rate.perNamespace == nil {
		return block{}, false
	}
	selected := make([]Workload, 0, len(in.Fleet))
	for _, w := range in.Fleet {
		if c.matches(w) {
			selected = append(selected, w)
		}
	}
	if c.rate.maxConcurrent != nil {
		base := 0
		applying := 0
		for _, w := range selected {
			if w.AutoDeploy {
				base++
			}
			if w.Applying {
				applying++
			}
		}
		limit, err := resolve(*c.rate.maxConcurrent, base)
		if err != nil {
			return block{field: "policy", reason: "rate.maxConcurrent: " + err.Error()}, true
		}
		if applying >= limit {
			return block{field: "rate.maxConcurrent", reason: fmt.Sprintf("waiting for a slot: %d of %d in progress in %s", applying, limit, c.p.Name)}, true
		}
	}
	if c.rate.perNamespace != nil {
		base, rolling, ns := 0, 0, in.Target.Namespace
		for _, w := range in.Fleet {
			if w.Namespace != ns {
				continue
			}
			base++
			if w.RollingOut {
				rolling++
			}
		}
		limit, err := resolve(*c.rate.perNamespace, base)
		if err != nil {
			return block{field: "policy", reason: "rate.perNamespace.maxConcurrent: " + err.Error()}, true
		}
		if rolling >= limit {
			return block{field: "rate.perNamespace", reason: fmt.Sprintf("waiting for a slot: %d of %d rollouts in progress in namespace %s", rolling, limit, ns)}, true
		}
	}
	if c.rate.maxPerHour != nil {
		since := now.Add(-time.Hour)
		count := 0
		var oldest time.Time
		for _, w := range selected {
			for _, t := range w.Applied {
				if t.After(since) && !t.After(now) {
					count++
					if oldest.IsZero() || t.Before(oldest) {
						oldest = t
					}
				}
			}
		}
		if count >= int(*c.rate.maxPerHour) {
			b := block{field: "rate.maxPerHour", reason: fmt.Sprintf("waiting for a slot: %d applies in the last hour, %s allows %d",
				count, c.p.Name, *c.rate.maxPerHour)}
			if !oldest.IsZero() {
				b.next = oldest.Add(time.Hour).Add(time.Second)
			}
			return b, true
		}
	}
	return block{}, false
}

// nextChange is the earliest instant any of the policy's rules changes, so the applier can look again then
// instead of polling.
func (c compiled) nextChange(in Input, now time.Time) time.Time {
	var out time.Time
	for _, f := range c.freezes {
		if f.From.After(now) {
			out = earliest(out, f.From.Time)
		}
		if f.To.After(now) {
			out = earliest(out, f.To.Time)
		}
	}
	if len(c.windows) > 0 {
		if _, inside := c.windowAt(now); inside {
			out = earliest(out, c.windowClose(now))
		} else {
			out = earliest(out, c.nextOpening(now))
		}
	}
	if c.rate.cooldown > 0 && !in.Target.Changed.IsZero() {
		if until := in.Target.Changed.Add(c.rate.cooldown); until.After(now) {
			out = earliest(out, until)
		}
	}
	return out
}

func (c compiled) frozen(now time.Time) (v1alpha1.ApplyFreeze, bool) {
	for _, f := range c.freezes {
		if !now.Before(f.From.Time) && now.Before(f.To.Time) {
			return f, true
		}
	}
	return v1alpha1.ApplyFreeze{}, false
}

func (c compiled) describeWindows() string {
	out := make([]string, 0, len(c.windows))
	for _, w := range c.windows {
		out = append(out, w.describe())
	}
	return strings.Join(out, ", ")
}

// spreadStart is when the workload's apply may start inside the span: the opening plus a stable offset, so
// the same workload always gets the same one and different workloads spread out.
func spreadStart(key string, opening, close time.Time, spread time.Duration) time.Time {
	if length := close.Sub(opening); spread > length {
		spread = length
	}
	secs := int64(spread / time.Second)
	if secs <= 0 {
		return opening
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	return opening.Add(time.Duration(int64(h.Sum32())%secs) * time.Second)
}

func earliest(a, b time.Time) time.Time {
	switch {
	case a.IsZero():
		return b
	case b.IsZero():
		return a
	case b.Before(a):
		return b
	}
	return a
}
