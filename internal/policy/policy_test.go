package policy

import (
	"strings"
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

func mustTime(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatalf("parse %q: %v", s, err)
	}
	return parsed
}

func daily(from, to string) v1alpha1.ApplyWindow {
	return v1alpha1.ApplyWindow{From: from, To: to}
}

func io(v string) *intstr.IntOrString {
	s := intstr.FromString(v)
	return &s
}

func ptr[T any](v T) *T { return &v }

func windowPolicy(w ...v1alpha1.ApplyWindow) []Policy {
	return []Policy{{Name: "ClusterApplyPolicy/production", Spec: v1alpha1.ApplyRules{Windows: w}}}
}

func target() Workload {
	return Workload{Namespace: "payments", Kind: "Deployment", Name: "checkout",
		Labels:          map[string]string{"app": "checkout"},
		NamespaceLabels: map[string]string{"env": "production"}}
}

func decide(t *testing.T, now time.Time, policies []Policy, mutate func(*Input)) Result {
	t.Helper()
	in := Input{Now: now, Target: target(), Policies: policies}
	if mutate != nil {
		mutate(&in)
	}
	return Decide(in)
}

func TestWindowContainment(t *testing.T) {
	sp, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		window v1alpha1.ApplyWindow
		at     string
		inside bool
	}{
		{"inside the window", daily("10:00", "16:00"), "2026-11-03T12:00:00Z", true},
		{"on the opening minute", daily("10:00", "16:00"), "2026-11-03T10:00:00Z", true},
		{"the closing minute is exclusive", daily("10:00", "16:00"), "2026-11-03T16:00:00Z", false},
		{"before the window", daily("10:00", "16:00"), "2026-11-03T09:59:00Z", false},
		{"overnight, before midnight", daily("22:00", "02:00"), "2026-11-03T23:00:00Z", true},
		{"overnight, after midnight", daily("22:00", "02:00"), "2026-11-04T01:00:00Z", true},
		{"overnight ends at to", daily("22:00", "02:00"), "2026-11-04T02:00:00Z", false},
		{
			"a listed day", v1alpha1.ApplyWindow{Days: []v1alpha1.Weekday{v1alpha1.Monday}, From: "10:00", To: "16:00"},
			"2026-11-02T11:00:00Z", true,
		},
		{
			"another day", v1alpha1.ApplyWindow{Days: []v1alpha1.Weekday{v1alpha1.Monday}, From: "10:00", To: "16:00"},
			"2026-11-03T11:00:00Z", false,
		},
		{
			"the tail of an overnight window belongs to the day it opened on",
			v1alpha1.ApplyWindow{Days: []v1alpha1.Weekday{v1alpha1.Monday}, From: "22:00", To: "02:00"},
			"2026-11-03T01:00:00Z", true,
		},
		{
			"an overnight window does not open on an unlisted day",
			v1alpha1.ApplyWindow{Days: []v1alpha1.Weekday{v1alpha1.Monday}, From: "22:00", To: "02:00"},
			"2026-11-03T23:00:00Z", false,
		},
		{
			"a named time zone reads the local clock",
			v1alpha1.ApplyWindow{From: "10:00", To: "16:00", TimeZone: sp.String()},
			"2026-11-03T13:00:00Z", true,
		},
		{
			"UTC is outside a São Paulo afternoon",
			v1alpha1.ApplyWindow{From: "10:00", To: "16:00", TimeZone: sp.String()},
			"2026-11-03T20:00:00Z", false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			now := mustTime(t, tc.at)
			got := decide(t, now, windowPolicy(tc.window), nil)
			inside := got.Auto.State == v1alpha1.ApplyAllowed
			if inside != tc.inside {
				t.Fatalf("inside=%v, want %v (decision %+v)", inside, tc.inside, got.Auto)
			}
		})
	}
}

func TestWindowsAreAUnion(t *testing.T) {
	policies := windowPolicy(daily("00:00", "08:00"), daily("20:00", "23:00"))
	for _, at := range []string{"2026-11-03T03:00:00Z", "2026-11-03T21:00:00Z"} {
		if got := decide(t, mustTime(t, at), policies, nil); got.Auto.State != v1alpha1.ApplyAllowed {
			t.Fatalf("%s should be inside the union: %+v", at, got.Auto)
		}
	}
	got := decide(t, mustTime(t, "2026-11-03T12:00:00Z"), policies, nil)
	if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "windows" {
		t.Fatalf("12:00 is outside both windows: %+v", got.Auto)
	}
	if want := mustTime(t, "2026-11-03T20:00:00Z"); got.Auto.NextAllowed == nil || !got.Auto.NextAllowed.Time.Equal(want) {
		t.Fatalf("nextAllowed = %v, want %v", got.Auto.NextAllowed, want)
	}
}

func TestDaylightSaving(t *testing.T) {
	lisbon := v1alpha1.ApplyWindow{From: "02:00", To: "05:00", TimeZone: "Europe/Lisbon"}
	lordHowe := v1alpha1.ApplyWindow{From: "02:00", To: "03:00", TimeZone: "Australia/Lord_Howe"}

	t.Run("a skipped opening is the instant the clock resumes", func(t *testing.T) {
		// Europe/Lisbon jumps from 01:00 to 02:00 on 2026-03-29; time.Date would put 02:00 an hour early.
		c := compile(Policy{Name: "p", Spec: v1alpha1.ApplyRules{Windows: []v1alpha1.ApplyWindow{lisbon}}})
		if c.err != nil {
			t.Fatal(c.err)
		}
		s, ok := c.windowAt(mustTime(t, "2026-03-29T01:00:00Z"))
		if !ok {
			t.Fatal("the window should be open when the clock resumes at 02:00")
		}
		if want := mustTime(t, "2026-03-29T01:00:00Z"); !s.open.Equal(want) {
			t.Fatalf("opened at %v, want %v (not before what is written)", s.open, want)
		}
		if inside, _ := c.windowAt(mustTime(t, "2026-03-29T00:59:00Z")); inside.open.IsZero() {
			// 00:59 UTC is 00:59 local, before the window.
		} else {
			t.Fatal("the window must not be open at 00:59Z, which is 00:59 local")
		}
	})

	t.Run("a repeated closing time closes at its first occurrence", func(t *testing.T) {
		// On 2026-10-25 Lisbon repeats 01:00-01:59; 01:30 local happens at 00:30Z and again at 01:30Z.
		c := compile(Policy{Name: "p", Spec: v1alpha1.ApplyRules{
			Windows: []v1alpha1.ApplyWindow{{From: "00:30", To: "01:30", TimeZone: "Europe/Lisbon"}}}})
		if c.err != nil {
			t.Fatal(c.err)
		}
		s, ok := c.windowAt(mustTime(t, "2026-10-25T00:10:00Z"))
		if !ok {
			t.Fatal("the window should be open at 00:10Z (01:10 local, first pass)")
		}
		if want := mustTime(t, "2026-10-25T00:30:00Z"); !s.close.Equal(want) {
			t.Fatalf("closed at %v, want %v (the first 01:30 local)", s.close, want)
		}
		if _, open := c.windowAt(mustTime(t, "2026-10-25T01:10:00Z")); open {
			t.Fatal("the window must not reopen for the repeated hour")
		}
	})

	t.Run("a thirty minute shift", func(t *testing.T) {
		// Lord Howe moves 02:00 to 02:30 on 2026-10-04, so a 02:00-03:00 window lasts thirty minutes.
		c := compile(Policy{Name: "p", Spec: v1alpha1.ApplyRules{Windows: []v1alpha1.ApplyWindow{lordHowe}}})
		if c.err != nil {
			t.Fatal(c.err)
		}
		s, ok := c.windowAt(mustTime(t, "2026-10-03T15:45:00Z"))
		if !ok {
			t.Fatal("the window should be open once the clock resumes at 02:30")
		}
		if want := mustTime(t, "2026-10-03T15:30:00Z"); !s.open.Equal(want) {
			t.Fatalf("opened at %v, want %v", s.open, want)
		}
		if d := s.close.Sub(s.open); d != 30*time.Minute {
			t.Fatalf("window lasted %v, want 30m", d)
		}
	})
}

// TestWindowNeverOpensOutsideTheClock walks every minute over the transitions of zones with a shift, a
// half-hour shift and no shift at all, and checks that the window is never open while the local clock reads
// a time outside [from, to).
func TestWindowNeverOpensOutsideTheClock(t *testing.T) {
	zones := map[string][]string{
		"Europe/Lisbon":       {"2026-03-28T00:00:00Z", "2026-03-30T00:00:00Z"},
		"Australia/Lord_Howe": {"2026-10-03T00:00:00Z", "2026-10-05T00:00:00Z"},
		"America/Sao_Paulo":   {"2026-11-02T00:00:00Z", "2026-11-04T00:00:00Z"},
	}
	for zone, bounds := range zones {
		for _, w := range []v1alpha1.ApplyWindow{
			{From: "01:00", To: "02:00", TimeZone: zone},
			{From: "02:00", To: "05:00", TimeZone: zone},
			{From: "01:30", To: "01:45", TimeZone: zone},
			{From: "22:00", To: "02:30", TimeZone: zone},
			{From: "02:00", To: "03:00", TimeZone: zone, Days: []v1alpha1.Weekday{v1alpha1.Sunday}},
		} {
			c := compile(Policy{Name: "p", Spec: v1alpha1.ApplyRules{Windows: []v1alpha1.ApplyWindow{w}}})
			if c.err != nil {
				t.Fatalf("%s %v: %v", zone, w, c.err)
			}
			parsed := c.windows[0]
			start, end := mustTime(t, bounds[0]), mustTime(t, bounds[1])
			for m := start; m.Before(end); m = m.Add(time.Minute) {
				if _, inside := c.windowAt(m); !inside {
					continue
				}
				if !clockInside(parsed, m) {
					t.Fatalf("%s %s: open at %s when the local clock reads %s", zone, parsed.describe(), m,
						m.In(parsed.loc).Format(time.RFC3339))
				}
			}
		}
	}
}

// clockInside is the definition the spans have to agree with: the local clock shows a time in [from, to) on a
// listed day, or in the tail of an overnight window opened the day before.
func clockInside(w window, t time.Time) bool {
	l := t.In(w.loc)
	day := uint(1) << uint(l.Weekday())
	minute := l.Hour()*60 + l.Minute()
	listed := func(d uint) bool { return w.days == 0 || w.days&uint8(d) != 0 }
	if w.from < w.to {
		return listed(day) && minute >= w.from && minute < w.to
	}
	if listed(day) && minute >= w.from {
		return true
	}
	return listed(1<<uint((l.Weekday()+6)%7)) && minute < w.to
}

func TestFreeze(t *testing.T) {
	now := mustTime(t, "2026-11-20T12:00:00Z")
	policies := []Policy{{Name: "ApplyPolicy/payments/launch", Spec: v1alpha1.ApplyRules{
		Freezes: []v1alpha1.ApplyFreeze{{
			From: metav1.NewTime(mustTime(t, "2026-11-20T00:00:00Z")), To: metav1.NewTime(mustTime(t, "2026-11-21T00:00:00Z")), Reason: "Checkout v5 launch",
		}},
	}}}
	got := decide(t, now, policies, nil)
	if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "freezes" {
		t.Fatalf("frozen apply: %+v", got.Auto)
	}
	if !strings.Contains(got.Auto.Reason, "Checkout v5 launch") {
		t.Fatalf("the reason should name the freeze: %q", got.Auto.Reason)
	}
	if want := mustTime(t, "2026-11-21T00:00:00Z"); got.Auto.NextAllowed == nil || !got.Auto.NextAllowed.Time.Equal(want) {
		t.Fatalf("nextAllowed = %v, want the end of the freeze", got.Auto.NextAllowed)
	}
	if got.Manual.State != v1alpha1.ApplyAllowed {
		t.Fatalf("a policy that governs Auto must leave a person free: %+v", got.Manual)
	}
}

// The cooldown is about the workload settling, not about a queue of applies: it holds a person too, while the
// limits that count concurrent applies never do.
func TestCooldownHoldsBothKindsOfApply(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	policies := []Policy{{Name: "p", Spec: v1alpha1.ApplyRules{
		AppliesTo: []v1alpha1.ApplyKind{v1alpha1.ApplyKindAuto, v1alpha1.ApplyKindManual},
		Rate:      &v1alpha1.ApplyRate{Cooldown: &metav1.Duration{Duration: 24 * time.Hour}}}}}
	got := decide(t, now, policies, func(in *Input) {
		in.Target.Changed = mustTime(t, "2026-11-03T06:00:00Z")
	})
	for _, d := range []v1alpha1.ApplyDecision{got.Auto, got.Manual} {
		if d.State != v1alpha1.ApplyBlocked || d.BlockedBy != "rate.cooldown" {
			t.Fatalf("both kinds should be cooling down: %+v", d)
		}
	}
}

func TestCooldown(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	policies := []Policy{{Name: "p", Spec: v1alpha1.ApplyRules{
		Rate: &v1alpha1.ApplyRate{Cooldown: &metav1.Duration{Duration: 24 * time.Hour}}}}}
	got := decide(t, now, policies, func(in *Input) {
		in.Target.Changed = mustTime(t, "2026-11-03T06:00:00Z")
	})
	if got.Auto.BlockedBy != "rate.cooldown" {
		t.Fatalf("blockedBy = %q, want rate.cooldown (%+v)", got.Auto.BlockedBy, got.Auto)
	}
	if want := mustTime(t, "2026-11-04T06:00:00Z"); got.Auto.NextAllowed == nil || !got.Auto.NextAllowed.Time.Equal(want) {
		t.Fatalf("nextAllowed = %v, want the cooldown to end", got.Auto.NextAllowed)
	}
	if next := got.NextChange; !next.Equal(mustTime(t, "2026-11-04T06:00:00Z")) {
		t.Fatalf("nextChange = %v, want the cooldown end", next)
	}
}

func TestAppliesTo(t *testing.T) {
	outside := daily("00:00", "01:00")
	tests := []struct {
		name string
		to   []v1alpha1.ApplyKind
		auto v1alpha1.ApplyDecisionState
		man  v1alpha1.ApplyDecisionState
	}{
		{"empty governs automatic applies only", nil, v1alpha1.ApplyBlocked, v1alpha1.ApplyAllowed},
		{"Auto", []v1alpha1.ApplyKind{v1alpha1.ApplyKindAuto}, v1alpha1.ApplyBlocked, v1alpha1.ApplyAllowed},
		{"Manual", []v1alpha1.ApplyKind{v1alpha1.ApplyKindManual}, v1alpha1.ApplyAllowed, v1alpha1.ApplyBlocked},
		{"both", []v1alpha1.ApplyKind{v1alpha1.ApplyKindAuto, v1alpha1.ApplyKindManual}, v1alpha1.ApplyBlocked, v1alpha1.ApplyBlocked},
	}
	now := mustTime(t, "2026-11-03T12:00:00Z")
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policies := []Policy{{Name: "p", Spec: v1alpha1.ApplyRules{AppliesTo: tc.to, Windows: []v1alpha1.ApplyWindow{outside}}}}
			got := decide(t, now, policies, nil)
			if got.Auto.State != tc.auto || got.Manual.State != tc.man {
				t.Fatalf("auto=%s manual=%s, want auto=%s manual=%s", got.Auto.State, got.Manual.State, tc.auto, tc.man)
			}
		})
	}
}

func TestAuditReportsWithoutBlocking(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	policies := []Policy{{Name: "ClusterApplyPolicy/quiet", Spec: v1alpha1.ApplyRules{
		Mode:    v1alpha1.ApplyPolicyAudit,
		Windows: []v1alpha1.ApplyWindow{daily("00:00", "01:00")},
	}}}
	got := decide(t, now, policies, nil)
	if got.Auto.State != v1alpha1.ApplyAllowed {
		t.Fatalf("audit must not block: %+v", got.Auto)
	}
	if len(got.Auto.WouldBlock) != 1 || !strings.HasPrefix(got.Auto.WouldBlock[0], "ClusterApplyPolicy/quiet: outside the apply windows") {
		t.Fatalf("wouldBlock = %q", got.Auto.WouldBlock)
	}
}

func TestFailsClosed(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	tests := []struct {
		name   string
		spec   v1alpha1.ApplyRules
		mutate func(*Input)
		want   string
	}{
		{"unknown time zone", v1alpha1.ApplyRules{Windows: []v1alpha1.ApplyWindow{
			{From: "10:00", To: "16:00", TimeZone: "Mars/Olympus"}}}, nil, "unknown time zone"},
		{"malformed selector", v1alpha1.ApplyRules{WorkloadSelector: &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{{Key: "a", Operator: "Bogus"}}}}, nil, "workloadSelector"},
		{"malformed percentage", v1alpha1.ApplyRules{Rate: &v1alpha1.ApplyRate{
			MaxConcurrent: io("ten%")}}, nil, "maxConcurrent"},
		{"equal from and to", v1alpha1.ApplyRules{Windows: []v1alpha1.ApplyWindow{
			daily("10:00", "10:00")}}, nil, "never close"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := decide(t, now, []Policy{{Name: "p", Spec: tc.spec}}, nil)
			if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "policy" {
				t.Fatalf("a policy that cannot be read must hold: %+v", got.Auto)
			}
			if !strings.Contains(got.Auto.Reason, tc.want) {
				t.Fatalf("reason = %q, want it to mention %q", got.Auto.Reason, tc.want)
			}
		})
	}
}

func TestRateSlots(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	policies := func(r *v1alpha1.ApplyRate) []Policy {
		return []Policy{{Name: "ClusterApplyPolicy/production", Spec: v1alpha1.ApplyRules{Rate: r}}}
	}
	applying := target()
	applying.Applying, applying.AutoDeploy = true, true
	other := Workload{Namespace: "payments", Kind: "Deployment", Name: "ledger", AutoDeploy: true,
		Labels: map[string]string{"app": "ledger"}, NamespaceLabels: map[string]string{"env": "production"}}

	t.Run("a percentage is rounded up", func(t *testing.T) {
		// 10% of one workload still allows one, as Karpenter rounds a budget up; 34% of three allows two.
		p := policies(&v1alpha1.ApplyRate{MaxConcurrent: io("10%")})
		in := Input{Now: now, Target: target(), Policies: p}
		in.Target.AutoDeploy = true
		in.Fleet = []Workload{in.Target}
		if got := Decide(in); got.Auto.State != v1alpha1.ApplyAllowed {
			t.Fatalf("one workload, nothing in progress: %+v", got.Auto)
		}
		in.Policies = policies(&v1alpha1.ApplyRate{MaxConcurrent: io("34%")})
		in.Fleet = []Workload{in.Target, other, applying}
		if got := Decide(in); got.Auto.State != v1alpha1.ApplyAllowed {
			t.Fatalf("34%% of three allows two, one is in progress: %+v", got.Auto)
		}
		in.Fleet = append(in.Fleet, applying)
		if got := Decide(in); got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "rate.maxConcurrent" {
			t.Fatalf("two of two in progress: %+v", got.Auto)
		}
	})

	t.Run("maxConcurrent counts applies in progress", func(t *testing.T) {
		p := policies(&v1alpha1.ApplyRate{MaxConcurrent: ptr(intstr.FromInt32(1))})
		got := decide(t, now, p, func(in *Input) {
			in.Target.AutoDeploy = true
			in.Fleet = []Workload{in.Target, other, applying}
		})
		if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "rate.maxConcurrent" {
			t.Fatalf("%+v", got.Auto)
		}
		if got.Auto.NextAllowed != nil {
			t.Fatal("waiting for a slot has no instant to promise")
		}
		if got.Manual.State != v1alpha1.ApplyAllowed {
			t.Fatalf("a person never waits for a slot: %+v", got.Manual)
		}
	})

	t.Run("a workload outside the selector does not count", func(t *testing.T) {
		p := []Policy{{Name: "p", Spec: v1alpha1.ApplyRules{
			WorkloadSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "checkout"}},
			Rate:             &v1alpha1.ApplyRate{MaxConcurrent: ptr(intstr.FromInt32(1))}}}}
		foreign := applying
		foreign.Name, foreign.Labels = "reports", map[string]string{"app": "reports"}
		got := decide(t, now, p, func(in *Input) {
			in.Target.AutoDeploy = true
			in.Fleet = []Workload{in.Target, foreign}
		})
		if got.Auto.State != v1alpha1.ApplyAllowed {
			t.Fatalf("%+v", got.Auto)
		}
	})

	t.Run("perNamespace counts rollouts resourcerer did not start", func(t *testing.T) {
		p := policies(&v1alpha1.ApplyRate{PerNamespace: &v1alpha1.ApplyNamespaceRate{
			MaxConcurrent: ptr(intstr.FromInt32(1))}})
		foreign := other
		foreign.RollingOut = true // a deploy from CI, not ours
		got := decide(t, now, p, func(in *Input) {
			in.Target.AutoDeploy = true
			in.Fleet = []Workload{in.Target, foreign}
		})
		if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "rate.perNamespace" {
			t.Fatalf("%+v", got.Auto)
		}
	})

	t.Run("maxPerHour counts the last hour", func(t *testing.T) {
		p := policies(&v1alpha1.ApplyRate{MaxPerHour: ptr(int32(2))})
		got := decide(t, now, p, func(in *Input) {
			in.Target.AutoDeploy = true
			in.Fleet = []Workload{in.Target, {Namespace: "payments", Kind: "Deployment", Name: "ledger", Labels: target().Labels,
				NamespaceLabels: target().NamespaceLabels,
				Applied:         []time.Time{now.Add(-30 * time.Minute), now.Add(-10 * time.Minute)}}}
		})
		if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "rate.maxPerHour" {
			t.Fatalf("%+v", got.Auto)
		}
		if want := now.Add(-30 * time.Minute).Add(time.Hour).Add(time.Second); got.Auto.NextAllowed == nil || !got.Auto.NextAllowed.Time.Equal(want) {
			t.Fatalf("nextAllowed = %v, want %v", got.Auto.NextAllowed, want)
		}
	})

	t.Run("its own rollout is not applied on top of", func(t *testing.T) {
		rolling := target()
		rolling.RollingOut = true
		p := policies(&v1alpha1.ApplyRate{MaxConcurrent: ptr(intstr.FromInt32(2))})
		got := decide(t, now, p, func(in *Input) {
			in.Target = rolling
			in.Fleet = []Workload{rolling}
		})
		if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.BlockedBy != "rollout" {
			t.Fatalf("%+v", got.Auto)
		}
		if got.Manual.State != v1alpha1.ApplyAllowed {
			t.Fatalf("a person reviewing is not held by a rollout: %+v", got.Manual)
		}
	})
}

func TestSpread(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z") // inside a window that opened at 10:00
	policies := []Policy{{Name: "p", Spec: v1alpha1.ApplyRules{
		Windows: []v1alpha1.ApplyWindow{daily("10:00", "18:00")},
		Rate:    &v1alpha1.ApplyRate{Spread: &metav1.Duration{Duration: 4 * time.Hour}},
	}}}
	first := decide(t, now, policies, nil)
	if first.Auto.State != v1alpha1.ApplyBlocked || first.Auto.BlockedBy != "rate.spread" {
		t.Fatalf("%+v", first.Auto)
	}
	if first.Auto.NextAllowed == nil || first.Auto.NextAllowed.After(mustTime(t, "2026-11-03T14:00:00Z")) {
		t.Fatalf("the offset must stay inside the spread: %v", first.Auto.NextAllowed)
	}
	again := decide(t, now, policies, nil)
	if !again.Auto.NextAllowed.Time.Equal(first.Auto.NextAllowed.Time) {
		t.Fatal("the same workload must always get the same offset")
	}
	if !first.Auto.NextAllowed.After(mustTime(t, "2026-11-03T09:59:00Z")) {
		t.Fatal("the offset counts from the opening")
	}
	// Once its offset has passed the apply is due, and the decision stops blocking.
	at := first.Auto.NextAllowed.Add(time.Second)
	if got := decide(t, at, policies, nil); got.Auto.State != v1alpha1.ApplyAllowed {
		t.Fatalf("after its offset the apply should be due: %+v", got.Auto)
	}
	if got := decide(t, now, policies, func(in *Input) { in.Target.Name = "ledger" }); got.Auto.NextAllowed == nil ||
		got.Auto.NextAllowed.Time.Equal(first.Auto.NextAllowed.Time) {
		t.Fatal("different workloads should spread out")
	}
}

func TestMostRestrictiveAndThePolicyNamed(t *testing.T) {
	now := mustTime(t, "2026-11-03T12:00:00Z")
	policies := []Policy{
		{Name: "ClusterApplyPolicy/business-hours", Spec: v1alpha1.ApplyRules{
			Windows: []v1alpha1.ApplyWindow{daily("10:00", "16:00")}}},
		{Name: "ApplyPolicy/payments/checkout", Namespace: "payments", Spec: v1alpha1.ApplyRules{
			Windows: []v1alpha1.ApplyWindow{daily("11:00", "15:00")}}},
	}
	if got := decide(t, now, policies, nil); got.Auto.State != v1alpha1.ApplyAllowed {
		t.Fatalf("the namespaced window only narrows the cluster's: %+v", got.Auto)
	}
	// The cluster window is the stricter one in the morning.
	got := decide(t, mustTime(t, "2026-11-03T09:00:00Z"), policies, nil)
	if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.Policy != "ClusterApplyPolicy/business-hours" {
		t.Fatalf("%+v", got.Auto)
	}
	// The namespaced policy is the stricter one in the afternoon.
	got = decide(t, mustTime(t, "2026-11-03T15:30:00Z"), policies, nil)
	if got.Auto.State != v1alpha1.ApplyBlocked || got.Auto.Policy != "ApplyPolicy/payments/checkout" {
		t.Fatalf("%+v", got.Auto)
	}
	// The instant to report is the first one both allow: 11:00 the next day.
	if want := mustTime(t, "2026-11-04T11:00:00Z"); got.Auto.NextAllowed == nil || !got.Auto.NextAllowed.Time.Equal(want) {
		t.Fatalf("nextAllowed = %v, want %v", got.Auto.NextAllowed, want)
	}
}

func TestSelectsAndValidate(t *testing.T) {
	p := Policy{Name: "ClusterApplyPolicy/production", Namespaces: &metav1.LabelSelector{
		MatchLabels: map[string]string{"env": "production"}}}
	if !Selects(p, target()) {
		t.Fatal("the namespace label matches")
	}
	other := target()
	other.NamespaceLabels = map[string]string{"env": "staging"}
	if Selects(p, other) {
		t.Fatal("a staging namespace is not selected")
	}
	if err := Validate(p); err != nil {
		t.Fatalf("valid policy: %v", err)
	}
	broken := Policy{Name: "bad", Spec: v1alpha1.ApplyRules{
		Windows: []v1alpha1.ApplyWindow{{From: "10:00", To: "16:00", TimeZone: "Mars/Olympus"}}}}
	if err := Validate(broken); err == nil || !strings.Contains(err.Error(), "unknown time zone") {
		t.Fatalf("err = %v", err)
	}
}
