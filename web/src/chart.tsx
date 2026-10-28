import { useEffect, useId, useRef, useState, type ReactNode } from "react";
import type { Point, Series } from "./api";

function useWidth() {
  const ref = useRef<HTMLDivElement>(null);
  const [w, setW] = useState(800);
  useEffect(() => {
    if (!ref.current) return;
    const ro = new ResizeObserver(([e]) => setW(Math.max(260, e.contentRect.width)));
    ro.observe(ref.current);
    return () => ro.disconnect();
  }, []);
  return [ref, w] as const;
}

function niceMax(v: number) {
  const e = 10 ** Math.floor(Math.log10(v));
  const m = v / e;
  return ([1, 1.25, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10].find((s) => m <= s) ?? 10) * e;
}

const ease = (k: number) => 1 - (1 - k) ** 3;

interface View {
  from: number;
  to: number;
  top: number;
}

/**
 * Animates the visible window towards the data's: on the first load the chart draws in from the left, and on a
 * new range the axes zoom from the old window to the new one instead of jumping.
 */
function useView(target: View | null) {
  const [view, setView] = useState<View | null>(null);
  const [drawn, setDrawn] = useState(0);
  const cur = useRef<View | null>(null);
  const key = target ? `${target.from}|${target.to}|${target.top}` : "";
  useEffect(() => {
    if (!target) return;
    const start = cur.current;
    const first = start == null;
    const t0 = performance.now();
    const dur = first ? 900 : 650;
    let raf = 0;
    const tick = (now: number) => {
      const k = ease(Math.min(1, (now - t0) / dur));
      const v = first
        ? target
        : {
            from: start.from + (target.from - start.from) * k,
            to: start.to + (target.to - start.to) * k,
            top: start.top + (target.top - start.top) * k,
          };
      cur.current = v;
      setView(v);
      if (first) setDrawn(k);
      if (k < 1) raf = requestAnimationFrame(tick);
    };
    raf = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(raf);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key]);
  return [view, drawn] as const;
}

/** Monotone cubic through the points (no overshoot), as a path continuing from the current pen position. */
function monotone(pts: [number, number][], move = true): string {
  const n = pts.length;
  if (n === 0) return "";
  let d = move ? `M${pts[0][0]},${pts[0][1]}` : `L${pts[0][0]},${pts[0][1]}`;
  if (n < 3)
    return (
      d +
      pts
        .slice(1)
        .map((p) => `L${p[0]},${p[1]}`)
        .join("")
    );
  const m: number[] = [];
  for (let i = 0; i < n - 1; i++) m.push((pts[i + 1][1] - pts[i][1]) / (pts[i + 1][0] - pts[i][0] || 1));
  const t = [m[0]];
  for (let i = 1; i < n - 1; i++) t.push(m[i - 1] * m[i] <= 0 ? 0 : (m[i - 1] + m[i]) / 2);
  t.push(m[n - 2]);
  for (let i = 0; i < n - 1; i++) {
    if (m[i] === 0) {
      t[i] = t[i + 1] = 0;
      continue;
    }
    const a = t[i] / m[i];
    const b = t[i + 1] / m[i];
    const s = a * a + b * b;
    if (s > 9) {
      const r = 3 / Math.sqrt(s);
      t[i] = r * a * m[i];
      t[i + 1] = r * b * m[i];
    }
  }
  for (let i = 0; i < n - 1; i++) {
    const dx = (pts[i + 1][0] - pts[i][0]) / 3;
    d += `C${pts[i][0] + dx},${pts[i][1] + t[i] * dx} ${pts[i + 1][0] - dx},${pts[i + 1][1] - t[i + 1] * dx} ${pts[i + 1][0]},${pts[i + 1][1]}`;
  }
  return d;
}

/** Splits points where samples are further apart than gap: missing data, not a slope. */
function segments<T extends { t: number }>(pts: T[], gap: number): T[][] {
  const out: T[][] = [];
  for (const p of pts) {
    const last = out[out.length - 1];
    if (last && p.t - last[last.length - 1].t <= gap) last.push(p);
    else out.push([p]);
  }
  return out;
}

export interface Line {
  name: string;
  color: string;
  points: Point[];
  dash?: string;
  /** step: values that change at discrete moments (requests, limits); smooth: measured values. */
  shape?: "step" | "smooth";
  /** Fill under the line, or down to another line's name (a savings gap). */
  fill?: boolean | string;
  width?: number;
}

export interface ChartEvent {
  t: number;
  label: string;
}

const BAND_KEYS = ["p95", "p75", "p50"] as const;
const BAND_LABELS: Record<(typeof BAND_KEYS)[number], string> = {
  p95: "95th percentile",
  p75: "75th percentile",
  p50: "50th percentile",
};

/** Usage percentile bands with lines for requests, limits or totals; t in unix seconds. */
export function TimeChart({
  title,
  unit,
  series,
  lines,
  fmtY,
  height = 300,
  loading,
  events = [],
  bandsLabel = "Usage",
  footer,
}: {
  title: ReactNode;
  unit: string;
  series?: Series;
  lines: Line[];
  fmtY: (v: number) => string;
  height?: number;
  loading?: boolean;
  events?: ChartEvent[];
  bandsLabel?: string;
  footer?: (t: number) => ReactNode;
}) {
  const [ref, width] = useWidth();
  const [hover, setHover] = useState<number | null>(null);
  const [hidden, setHidden] = useState<Set<string>>(() => new Set());
  const toggle = (k: string) =>
    setHidden((h) => {
      const n = new Set(h);
      if (n.has(k)) n.delete(k);
      else n.add(k);
      return n;
    });
  const bands = series?.bands ?? [];
  const showBands = bands.length > 0 && !hidden.has(bandsLabel);
  const shown = lines.filter((l) => !hidden.has(l.name));
  const maxV = Math.max(
    1e-9,
    ...(showBands ? bands.map((b) => b.p95) : []),
    ...shown.flatMap((l) => l.points.map((p) => p.v)),
  );
  const [view, drawn] = useView(series ? { from: series.from, to: series.to, top: niceMax(maxV * 1.1) } : null);
  const from = view?.from ?? 0;
  const to = view?.to ?? 1;
  const top = view?.top ?? 1;
  const pad = { l: 64, r: 16, t: 18, b: 28 };
  const iw = width - pad.l - pad.r;
  const ih = height - pad.t - pad.b;
  const x = (t: number) => pad.l + ((t - from) / (to - from)) * iw;
  const y = (v: number) => pad.t + ih - (Math.max(0, v) / top) * ih;
  const span = to - from;

  const step = (pts: Point[]) => {
    let d = "";
    for (const p of pts) d += d ? `H${x(p.t)}V${y(p.v)}` : `M${x(p.t)},${y(p.v)}`;
    return d && `${d}H${x(Math.max(pts[pts.length - 1].t, series?.to ?? 0))}`;
  };
  const lineGap = Math.max(series ? (series.to - series.from) / 60 : Infinity, 900);
  const path = (l: Line) =>
    l.shape === "smooth"
      ? segments(l.points, lineGap)
          .map((s) => monotone(s.map((p) => [x(p.t), y(p.v)])))
          .join("")
      : step(l.points);
  const fillPath = (l: Line) => {
    const base = typeof l.fill === "string" ? lines.find((o) => o.name === l.fill) : undefined;
    return segments(l.points, lineGap)
      .map((s) => {
        const upper = monotone(s.map((p) => [x(p.t), y(p.v)]));
        if (!base) return `${upper}L${x(s[s.length - 1].t)},${y(0)}L${x(s[0].t)},${y(0)}Z`;
        const lower = s.map((p) => [x(p.t), y(at(base.points, p.t)?.v ?? p.v)] as [number, number]).reverse();
        return `${upper}${monotone(lower, false)}Z`;
      })
      .join("");
  };
  // Buckets further apart than three widths are a gap in data, not a slope.
  const bandGap = series ? ((series.to - series.from) / 150) * 3 : Infinity;
  const area = (k: (typeof BAND_KEYS)[number]) =>
    segments(bands, bandGap)
      .map((s) => {
        const upper = monotone(s.map((b) => [x(b.t), y(b[k])]));
        return `${upper}L${x(s[s.length - 1].t)},${y(0)}L${x(s[0].t)},${y(0)}Z`;
      })
      .join("");

  const tickEvery =
    span <= 3600
      ? 600
      : span <= 86400
        ? 4 * 3600
        : span <= 7 * 86400
          ? 86400
          : span <= 14 * 86400
            ? 2 * 86400
            : 5 * 86400;
  const ticks: number[] = [];
  if (view) for (let t = Math.ceil(from / tickEvery) * tickEvery; t < to; t += tickEvery) ticks.push(t);
  const label = (t: number) =>
    span <= 86400
      ? new Date(t * 1000).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })
      : new Date(t * 1000).toLocaleDateString([], { month: "short", day: "numeric" });
  const hb =
    hover != null && showBands ? bands.reduce((b, p) => (Math.abs(p.t - hover) < Math.abs(b.t - hover) ? p : b)) : null;
  const empty = !loading && bands.length === 0 && shown.every((l) => l.points.length === 0);
  const visibleEvents = events
    .filter((e) => series && e.t >= series.from && e.t <= series.to)
    .sort((a, b) => a.t - b.t);
  // Markers closer than a label's width share one label, e.g. an apply reverted minutes later.
  const eventGroups: { key: string; x: number; xs: number[]; label: string; w: number }[] = [];
  for (const e of visibleEvents) {
    const ex = x(e.t);
    const g = eventGroups[eventGroups.length - 1];
    if (g && ex - g.xs[0] < 70) {
      g.xs.push(ex);
      g.label = `${g.xs.length} events`;
    } else eventGroups.push({ key: `${e.t}-${e.label}`, x: ex, xs: [ex], label: e.label, w: 0 });
  }
  for (const g of eventGroups) g.w = g.label.length * 6 + 14;

  const rows = [
    ...(hb
      ? BAND_KEYS.map((k, i) => ({ key: k, label: BAND_LABELS[k], v: hb[k], sw: `sw b${2 - i}`, color: "" }))
      : []),
    ...(hover != null
      ? shown.flatMap((l) => {
          const p = at(l.points, hover);
          return p ? [{ key: l.name, label: l.name, v: p.v, sw: "sw", color: l.color }] : [];
        })
      : []),
  ].sort((a, b) => b.v - a.v);
  const dots =
    hover != null
      ? [
          ...(hb
            ? [
                { v: hb.p95, color: "var(--band1)" },
                { v: hb.p50, color: "var(--band0)" },
              ]
            : []),
          ...shown.flatMap((l) => {
            const p = at(l.points, hover);
            return p ? [{ v: p.v, color: l.color }] : [];
          }),
        ]
      : [];
  const clip = `rz-clip-${useId().replace(/:/g, "")}`;

  return (
    <div className="rz-tchart">
      <div className="rz-tcharthd">
        <h4>{title}</h4>
      </div>
      <div ref={ref} style={{ position: "relative" }}>
        <svg
          width={width}
          height={height}
          onMouseLeave={() => setHover(null)}
          onMouseMove={(e) => {
            const r = e.currentTarget.getBoundingClientRect();
            const t = from + ((e.clientX - r.left - pad.l) / iw) * span;
            setHover(view && t >= from && t <= to ? t : null);
          }}
        >
          <defs>
            <clipPath id={clip}>
              <rect x={pad.l} y={0} width={iw * drawn} height={height} />
            </clipPath>
            {lines
              .filter((l) => l.fill === true)
              .map((l, i) => (
                <linearGradient key={l.name} id={`${clip}-g${i}`} x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0" style={{ stopColor: l.color, stopOpacity: 0.35 }} />
                  <stop offset="1" style={{ stopColor: l.color, stopOpacity: 0.02 }} />
                </linearGradient>
              ))}
          </defs>
          {[0, 0.25, 0.5, 0.75, 1].map((f) => (
            <g key={f}>
              <line x1={pad.l} x2={pad.l + iw} y1={y(top * f)} y2={y(top * f)} className="g-grid" />
              <text x={pad.l - 8} y={y(top * f) + 4} textAnchor="end" className="g-axis">
                {fmtY(top * f)}
              </text>
            </g>
          ))}
          {ticks.map((t) => (
            <g key={t}>
              <line x1={x(t)} x2={x(t)} y1={pad.t} y2={pad.t + ih} className="g-grid" />
              <text x={x(t)} y={height - 8} textAnchor="middle" className="g-axis">
                {label(t)}
              </text>
            </g>
          ))}
          <text transform={`translate(14 ${pad.t + ih / 2}) rotate(-90)`} textAnchor="middle" className="g-axis">
            {unit}
          </text>
          {view && (
            <g clipPath={`url(#${clip})`}>
              {BAND_KEYS.map((k, i) => (
                <path key={k} d={area(k)} className={`g-band b${2 - i}${showBands ? "" : " off"}`} />
              ))}
              {lines.map((l) =>
                l.fill && l.points.length ? (
                  <path
                    key={`${l.name}-fill`}
                    d={fillPath(l)}
                    className={`g-fill${hidden.has(l.name) ? " off" : ""}`}
                    style={{
                      fill:
                        l.fill === true ? `url(#${clip}-g${lines.filter((o) => o.fill === true).indexOf(l)})` : l.color,
                    }}
                  />
                ) : null,
              )}
              {lines.map((l) => (
                <path
                  key={l.name}
                  d={path(l)}
                  className={`g-line${hidden.has(l.name) ? " off" : ""}`}
                  style={{ stroke: l.color, strokeDasharray: l.dash, strokeWidth: l.width }}
                />
              ))}
              {eventGroups.map((g) => {
                const lx = Math.min(Math.max(g.x, pad.l + g.w / 2), pad.l + iw - g.w / 2);
                return (
                  <g key={g.key} className="g-ev">
                    {g.xs.map((ex) => (
                      <line key={ex} x1={ex} x2={ex} y1={pad.t + 4} y2={pad.t + ih} />
                    ))}
                    <rect x={lx - g.w / 2} y={pad.t - 12} width={g.w} height={15} rx={7.5} />
                    <text x={lx} y={pad.t - 1.5} textAnchor="middle">
                      {g.label}
                    </text>
                  </g>
                );
              })}
            </g>
          )}
          {(loading || empty) && !view && (
            <text x={pad.l + iw / 2} y={pad.t + ih / 2} textAnchor="middle" className="g-axis">
              {loading ? "Loading…" : "No samples in this window yet"}
            </text>
          )}
          {empty && view && (
            <text x={pad.l + iw / 2} y={pad.t + ih / 2} textAnchor="middle" className="g-axis">
              No samples in this window yet
            </text>
          )}
          {hover != null && (
            <>
              <line x1={x(hover)} x2={x(hover)} y1={pad.t} y2={pad.t + ih} className="g-cross" />
              {dots.map((d, i) => (
                <circle key={i} cx={x(hover)} cy={y(d.v)} r={4} className="g-dot" style={{ fill: d.color }} />
              ))}
            </>
          )}
        </svg>
        {hover != null && rows.length > 0 && (
          <div
            className="rz-tt"
            style={x(hover) > width - 280 ? { right: width - x(hover) + 14 } : { left: x(hover) + 14 }}
          >
            <div className="tt-h">
              {new Date((hb?.t ?? hover) * 1000).toLocaleString([], {
                month: "short",
                day: "numeric",
                hour: "numeric",
                minute: "2-digit",
              })}
            </div>
            {rows.map((r) => (
              <div key={r.key} className="tt-r">
                <i className={r.sw} style={r.color ? { background: r.color } : undefined} />
                <span>{r.label}</span>
                <b>{fmtY(r.v)}</b>
              </div>
            ))}
            {visibleEvents
              .filter((e) => Math.abs(e.t - hover) < span / 80)
              .map((e) => (
                <div key={e.t} className="tt-ev">
                  {e.label} at {new Date(e.t * 1000).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}
                </div>
              ))}
            {footer?.(hover)}
          </div>
        )}
      </div>
      <div className="rz-glegend" role="group" aria-label="Series">
        {bands.length > 0 && (
          <button aria-pressed={!hidden.has(bandsLabel)} onClick={() => toggle(bandsLabel)}>
            <i className="sw b1" />
            {bandsLabel}
          </button>
        )}
        {lines.map((l) => (
          <button key={l.name} aria-pressed={!hidden.has(l.name)} onClick={() => toggle(l.name)}>
            <i className={l.fill === true ? "sw" : "sw line"} style={{ background: l.color }} />
            {l.name}
          </button>
        ))}
      </div>
    </div>
  );
}

export function at(pts: Point[], t: number) {
  let best: Point | null = null;
  for (const p of pts) if (p.t <= t) best = p;
  return best;
}
