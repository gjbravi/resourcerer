package policy

import (
	"fmt"
	"strings"
	"time"

	"github.com/gjbravi/resourcerer/api/v1alpha1"
)

// maxSearch is how far ahead an opening is looked for: a window that opens once a week is found within
// eight days, and one that never opens is reported far away instead of forever.
const maxSearch = 9 * 24 * time.Hour

// span is a window's opening and closing instants.
type span struct{ open, close time.Time }

// window is one parsed ApplyWindow.
type window struct {
	loc  *time.Location
	tz   string
	week []v1alpha1.Weekday
	// days is a bit per time.Weekday; zero means every day.
	days uint8
	// from and to are minutes after local midnight.
	from, to int
}

func parseWindow(w v1alpha1.ApplyWindow) (window, error) {
	out := window{tz: w.TimeZone, loc: time.UTC, week: w.Days}
	if w.TimeZone != "" {
		l, err := time.LoadLocation(w.TimeZone)
		if err != nil {
			return out, fmt.Errorf("unknown time zone %q", w.TimeZone)
		}
		out.loc = l
	}
	from, err := minutes(w.From)
	if err != nil {
		return out, err
	}
	to, err := minutes(w.To)
	if err != nil {
		return out, err
	}
	if from == to {
		return out, fmt.Errorf("from and to are both %s, so the window would never close", w.From)
	}
	for _, d := range w.Days {
		wd, ok := weekday(d)
		if !ok {
			return out, fmt.Errorf("%q is not a day of the week", d)
		}
		out.days |= 1 << wd
	}
	out.from, out.to = from, to
	return out, nil
}

func minutes(hhmm string) (int, error) {
	t, err := time.Parse("15:04", hhmm)
	if err != nil {
		return 0, fmt.Errorf("%q is not a time of day like 10:00", hhmm)
	}
	return t.Hour()*60 + t.Minute(), nil
}

func weekday(d v1alpha1.Weekday) (time.Weekday, bool) {
	for i, name := range [...]v1alpha1.Weekday{v1alpha1.Sunday, v1alpha1.Monday, v1alpha1.Tuesday,
		v1alpha1.Wednesday, v1alpha1.Thursday, v1alpha1.Friday, v1alpha1.Saturday} {
		if d == name {
			return time.Weekday(i), true
		}
	}
	return 0, false
}

// describe writes the window the way a person reads it, for the decision's reason.
func (w window) describe() string {
	days := "every day"
	if len(w.week) > 0 {
		names := make([]string, len(w.week))
		for i, d := range w.week {
			names[i] = string(d)
		}
		days = strings.Join(names, ",")
	}
	tz := w.tz
	if tz == "" {
		tz = "UTC"
	}
	return fmt.Sprintf("%s %02d:%02d-%02d:%02d %s", days, w.from/60, w.from%60, w.to/60, w.to%60, tz)
}

// spans returns the openings that touch [from, to], one per listed local day. It never scans the clock: a
// minute-by-minute walk over a week in a named zone is far too slow to run per decision.
func (w window) spans(from, to time.Time) []span {
	var out []span
	first := dateOf(from, w.loc).AddDate(0, 0, -1)
	last := dateOf(to, w.loc).AddDate(0, 0, 1)
	for d := first; !d.After(last); d = d.AddDate(0, 0, 1) {
		if w.days != 0 && w.days&(1<<uint(d.Weekday())) == 0 {
			continue
		}
		open := instantOf(d, w.from, w.loc)
		closeDay := d
		if w.to < w.from {
			closeDay = d.AddDate(0, 0, 1)
		}
		close := instantOf(closeDay, w.to, w.loc)
		if !close.After(open) || !close.After(from) || open.After(to) {
			continue
		}
		out = append(out, span{open: open, close: close})
	}
	return out
}

// dateOf is the local calendar date of t, as UTC midnight of that date.
func dateOf(t time.Time, loc *time.Location) time.Time {
	y, m, d := t.In(loc).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// instantOf returns the instant whose local clock reads that date and minute: the first occurrence when the
// clock reads it twice, and the instant the clock resumes when it skips it. So a window never opens before,
// nor closes after, what it says, in any zone.
func instantOf(date time.Time, minute int, loc *time.Location) time.Time {
	y, m, d := date.Date()
	hh, mm := minute/60, minute%60
	wall := time.Date(y, m, d, hh, mm, 0, 0, time.UTC)
	var best time.Time
	// One candidate per offset in force around that day; for a repeated time the earlier one is the answer.
	for _, probe := range []time.Time{wall.Add(-24 * time.Hour), wall.Add(24 * time.Hour)} {
		_, off := probe.In(loc).Zone()
		cand := wall.Add(-time.Duration(off) * time.Second)
		if !sameClock(cand, loc, y, m, d, hh, mm) {
			continue
		}
		if best.IsZero() || cand.Before(best) {
			best = cand
		}
	}
	if best.IsZero() {
		return resume(y, m, d, hh, mm, loc)
	}
	return best
}

func sameClock(t time.Time, loc *time.Location, y int, m time.Month, d, hh, mm int) bool {
	l := t.In(loc)
	ly, lm, ld := l.Date()
	lh, lmi, ls := l.Clock()
	return ly == y && lm == m && ld == d && lh == hh && lmi == mm && ls == 0
}

// resume is the instant the clock jumps past a time it skips: the transition that would have shown it.
// time.Date normalizes a skipped time, sometimes an hour early, which would open a window too early.
func resume(y int, m time.Month, d, hh, mm int, loc *time.Location) time.Time {
	guess := time.Date(y, m, d, hh, mm, 0, 0, loc)
	lo, hi := guess.Add(-3*time.Hour), guess.Add(3*time.Hour)
	_, before := lo.In(loc).Zone()
	if _, after := hi.In(loc).Zone(); before == after {
		return guess
	}
	for hi.Sub(lo) > time.Second {
		mid := lo.Add(hi.Sub(lo) / 2)
		if _, off := mid.In(loc).Zone(); off == before {
			lo = mid
		} else {
			hi = mid
		}
	}
	return hi.Truncate(time.Second)
}

// windowAt is the opening that contains now, the most recent one when windows overlap.
func (c compiled) windowAt(now time.Time) (span, bool) {
	var best span
	found := false
	for _, w := range c.windows {
		for _, s := range w.spans(now, now) {
			if now.Before(s.open) || !now.Before(s.close) {
				continue
			}
			if !found || s.open.After(best.open) {
				best, found = s, true
			}
		}
	}
	return best, found
}

// windowClose is when the opening that contains now closes.
func (c compiled) windowClose(now time.Time) time.Time {
	if s, ok := c.windowAt(now); ok {
		return s.close
	}
	return time.Time{}
}

// nextOpening is when a closed set of windows opens next, or the end of the search when it never does.
func (c compiled) nextOpening(now time.Time) time.Time {
	var out time.Time
	for _, w := range c.windows {
		for _, s := range w.spans(now, now.Add(maxSearch)) {
			if s.open.After(now) {
				out = earliest(out, s.open)
			}
		}
	}
	if out.IsZero() {
		return now.Add(maxSearch)
	}
	return out
}

// opening is when the held applies of this span may start: the window opening, or the end of a freeze that
// ran into it, since the spread of a herd counts from the moment it is free to move.
func (c compiled) opening(s span, now time.Time) time.Time {
	open := s.open
	for _, f := range c.freezes {
		to := f.To.Time
		if to.After(open) && to.Before(s.close) && !to.After(now) {
			open = f.To.Time
		}
	}
	return open
}
