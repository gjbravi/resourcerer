import { useEffect, useId, useMemo, useRef, useState, type ReactNode } from "react";
import { Area, CartesianGrid, ComposedChart, Line as RLine, ReferenceLine, Tooltip, XAxis, YAxis } from "recharts";
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

const BAND_KEYS = ["p95", "p50"] as const;
const BAND_LINES = [
  { key: "p95", color: "var(--c-usage)", width: 2 },
  { key: "p50", color: "var(--c-usage50)", width: 1.5 },
] as const;

type Row = { t: number } & Record<string, number | null>;

/**
 * Value of a series at t: the last point for steps, linear between neighbours for measurements, and null outside
 * the series or across a gap in data, so the line breaks instead of bridging it.
 */
function valueAt(pts: Point[], t: number, step: boolean, gap: number): number | null {
  if (!pts.length || t < pts[0].t) return null;
  let lo = 0;
  let hi = pts.length - 1;
  while (lo < hi) {
    const mid = (lo + hi + 1) >> 1;
    if (pts[mid].t <= t) lo = mid;
    else hi = mid - 1;
  }
  const p = pts[lo];
  const n = pts[lo + 1];
  if (p.t === t) return p.v;
  if (step) return (n ? n.t - p.t : t - p.t) <= gap ? p.v : null;
  if (!n || n.t - p.t > gap) return null;
  return p.v + ((n.v - p.v) * (t - p.t)) / (n.t - p.t);
}

/** Usage percentile bands with lines for requests, limits or totals; t in unix seconds. */
export function TimeChart({
  title,
  unit,
  series,
  lines,
  fmtY,
  fmtTip = fmtY,
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
  /** Tooltip values, with units; defaults to the axis format. */
  fmtTip?: (v: number) => string;
  height?: number;
  loading?: boolean;
  events?: ChartEvent[];
  bandsLabel?: string;
  footer?: (t: number) => ReactNode;
}) {
  const [ref, width] = useWidth();
  const [hidden, setHidden] = useState<Set<string>>(() => new Set());
  const toggle = (k: string) =>
    setHidden((h) => {
      const n = new Set(h);
      if (n.has(k)) n.delete(k);
      else n.add(k);
      return n;
    });
  const id = `rz${useId().replace(/:/g, "")}`;
  const from = series?.from ?? 0;
  const to = series?.to ?? 1;
  const span = to - from;
  const bands = useMemo(() => series?.bands ?? [], [series]);
  const shownBands = BAND_LINES.filter((b) => bands.length > 0 && !hidden.has(`${bandsLabel} ${b.key}`));
  const key = (i: number) => `l${i}`;

  const rows = useMemo(() => {
    const lineGap = Math.max(span / 60, 900);
    const bandGap = (span / 150) * 3;
    const ts = new Set<number>();
    // A row inside each gap gets null for that series; without it the chart joins the points across the gap.
    const add = (pts: { t: number }[], gap: number) =>
      pts.forEach((p, i) => {
        ts.add(p.t);
        if (i > 0 && p.t - pts[i - 1].t > gap) ts.add((p.t + pts[i - 1].t) / 2);
      });
    add(bands, bandGap);
    for (const l of lines) add(l.points, lineGap);
    return [...ts]
      .sort((a, b) => a - b)
      .map((t) => {
        const row: Row = { t };
        for (const k of BAND_KEYS)
          row[k] = valueAt(
            bands.map((b) => ({ t: b.t, v: b[k] })),
            t,
            false,
            bandGap,
          );
        lines.forEach((l, i) => (row[key(i)] = valueAt(l.points, t, l.shape !== "smooth", lineGap)));
        return row;
      });
  }, [bands, lines, span]);

  const shown = lines.filter((l) => !hidden.has(l.name));
  const maxV = Math.max(
    1e-9,
    ...shownBands.flatMap((sb) => bands.map((b) => b[sb.key])),
    ...shown.flatMap((l) => l.points.map((p) => p.v)),
  );
  const top = niceMax(maxV * 1.1);
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
  for (let t = Math.ceil(from / tickEvery) * tickEvery; t < to; t += tickEvery) ticks.push(t);
  const label = (t: number) =>
    span <= 86400
      ? new Date(t * 1000).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })
      : new Date(t * 1000).toLocaleDateString([], { month: "short", day: "numeric" });

  // Markers closer than a label's width share one label, e.g. an apply reverted minutes later.
  const plotW = width - 80;
  const groups: { t: number; ts: number[]; label: string }[] = [];
  for (const e of events.filter((e) => series && e.t >= from && e.t <= to).sort((a, b) => a.t - b.t)) {
    const g = groups[groups.length - 1];
    if (g && ((e.t - g.ts[0]) / span) * plotW < 70) {
      g.ts.push(e.t);
      g.label = `${g.ts.length} events`;
    } else groups.push({ t: e.t, ts: [e.t], label: e.label });
  }

  const empty = !loading && rows.length === 0;
  const fillBase = (l: Line) => (typeof l.fill === "string" ? lines.findIndex((o) => o.name === l.fill) : -1);
  const anim = { isAnimationActive: true, animationDuration: 700, animationEasing: "ease-out" as const };

  return (
    <div className="rz-tchart">
      <div className="rz-tcharthd">
        <h4>{title}</h4>
      </div>
      <div ref={ref} className="rz-tchartbody">
        <ComposedChart width={width} height={height} data={rows} margin={{ top: 20, right: 16, bottom: 4, left: 8 }}>
          <defs>
            {lines.map((l, i) => (
              <linearGradient key={l.name} id={`${id}-g${i}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" style={{ stopColor: l.color, stopOpacity: 0.35 }} />
                <stop offset="1" style={{ stopColor: l.color, stopOpacity: 0.02 }} />
              </linearGradient>
            ))}
          </defs>
          <CartesianGrid stroke="var(--track)" strokeDasharray="3 3" />
          <XAxis
            dataKey="t"
            type="number"
            domain={[from, to]}
            ticks={ticks}
            tickFormatter={label}
            className="g-axis"
            tickLine={false}
            axisLine={false}
            allowDataOverflow
          />
          <YAxis
            domain={[0, top]}
            ticks={[0, top / 4, top / 2, (top * 3) / 4, top]}
            tickFormatter={fmtY}
            className="g-axis"
            tickLine={false}
            axisLine={false}
            width={56}
            allowDataOverflow
            label={{ value: unit, angle: -90, position: "insideLeft", className: "g-axis", dy: 30 }}
          />
          {BAND_LINES.map((b) => (
            <RLine
              key={b.key}
              type="monotone"
              dataKey={b.key}
              stroke={b.color}
              strokeWidth={b.width}
              dot={false}
              activeDot={{ r: 3.5, stroke: "var(--panel)", strokeWidth: 1.5, fill: b.color }}
              hide={!bands.length || hidden.has(`${bandsLabel} ${b.key}`)}
              {...anim}
            />
          ))}
          {lines.map((l, i) =>
            l.fill ? (
              <Area
                key={`${l.name}-fill`}
                type={l.shape === "smooth" ? "monotone" : "stepAfter"}
                dataKey={
                  fillBase(l) >= 0
                    ? (r: Row) =>
                        r[key(i)] == null || r[key(fillBase(l))] == null ? null : [r[key(fillBase(l))], r[key(i)]]
                    : key(i)
                }
                stroke="none"
                fill={l.fill === true ? `url(#${id}-g${i})` : l.color}
                fillOpacity={l.fill === true ? 1 : 0.1}
                hide={hidden.has(l.name)}
                activeDot={false}
                {...anim}
              />
            ) : null,
          )}
          {lines.map((l, i) => (
            <RLine
              key={l.name}
              type={l.shape === "smooth" ? "monotone" : "stepAfter"}
              dataKey={key(i)}
              stroke={l.color}
              strokeWidth={l.width ?? 2.2}
              strokeDasharray={l.dash}
              strokeLinecap="round"
              dot={false}
              activeDot={{ r: 3.5, stroke: "var(--panel)", strokeWidth: 1.5, fill: l.color }}
              hide={hidden.has(l.name)}
              {...anim}
            />
          ))}
          {groups.map((g) => (
            <ReferenceLine
              key={g.t}
              x={g.t}
              className="g-ev"
              stroke="var(--accent)"
              strokeDasharray="2 3"
              label={{
                position: "top",
                content: (props) => {
                  const viewBox = props.viewBox as { x?: number; y?: number } | undefined;
                  const w = g.label.length * 6 + 14;
                  const x = Math.min(Math.max(viewBox?.x ?? 0, 64 + w / 2), width - 16 - w / 2);
                  return (
                    <g className="g-ev">
                      <rect x={x - w / 2} y={(viewBox?.y ?? 20) - 16} width={w} height={15} rx={7.5} />
                      <text x={x} y={(viewBox?.y ?? 20) - 5} textAnchor="middle">
                        {g.label}
                      </text>
                    </g>
                  );
                },
              }}
            />
          ))}
          <Tooltip
            cursor={{ stroke: "var(--accent-line)", strokeWidth: 1 }}
            isAnimationActive
            animationDuration={180}
            wrapperStyle={{ outline: "none", zIndex: 5 }}
            content={({ active, payload }) => {
              const row = payload?.[0]?.payload as Row | undefined;
              if (!active || !row) return null;
              const items = [
                ...shownBands.map((b) => ({
                  key: b.key,
                  label: `${bandsLabel} ${b.key}`,
                  v: row[b.key],
                  color: b.color,
                })),
                ...lines.map((l, i) => ({ key: l.name, label: l.name, v: row[key(i)], color: l.color })),
              ]
                .filter((r): r is typeof r & { v: number } => r.v != null && !hidden.has(r.key))
                .sort((a, b) => b.v - a.v);
              const near = groups.filter((g) => Math.abs(g.t - row.t) < span / 80);
              return (
                <div className="rz-tt">
                  <div className="tt-h">
                    {new Date(row.t * 1000).toLocaleString([], {
                      month: "short",
                      day: "numeric",
                      hour: "numeric",
                      minute: "2-digit",
                    })}
                  </div>
                  {items.map((r) => (
                    <div key={r.key} className="tt-r">
                      <span style={{ color: `color-mix(in srgb, ${r.color} 60%, var(--text))` }}>{r.label}</span>
                      <b>{fmtTip(r.v)}</b>
                    </div>
                  ))}
                  {near.map((g) => (
                    <div key={g.t} className="tt-ev">
                      {g.ts.length > 1 ? `${g.ts.length} applies or reverts` : g.label} around{" "}
                      {new Date(g.t * 1000).toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })}
                    </div>
                  ))}
                  {footer?.(row.t)}
                </div>
              );
            }}
          />
        </ComposedChart>
        {(loading || empty) && (
          <div className="rz-tchartmsg">{loading ? "Loading…" : "No samples in this window yet"}</div>
        )}
      </div>
      <div className="rz-glegend" role="group" aria-label="Series">
        {bands.length > 0 &&
          BAND_LINES.map((b) => (
            <button
              key={b.key}
              aria-pressed={!hidden.has(`${bandsLabel} ${b.key}`)}
              onClick={() => toggle(`${bandsLabel} ${b.key}`)}
            >
              <i className="sw line" style={{ background: b.color }} />
              {bandsLabel} {b.key}
            </button>
          ))}
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
