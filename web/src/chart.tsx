import { useEffect, useRef, useState, type ReactNode } from "react";
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
  return (m <= 1 ? 1 : m <= 2 ? 2 : m <= 2.5 ? 2.5 : m <= 5 ? 5 : 10) * e;
}

export interface Line {
  name: string;
  color: string;
  points: Point[];
  dash?: string;
}

const BAND_KEYS = ["p95", "p75", "p50"] as const;
const BAND_LABELS: Record<(typeof BAND_KEYS)[number], string> = {
  p95: "95th percentile",
  p75: "75th percentile",
  p50: "50th percentile",
};

/** Usage percentile bands with step lines for requests and limits; t in unix seconds. */
export function TimeChart({
  title,
  unit,
  series,
  lines,
  fmtY,
  height = 300,
  loading,
}: {
  title: ReactNode;
  unit: string;
  series?: Series;
  lines: Line[];
  fmtY: (v: number) => string;
  height?: number;
  loading?: boolean;
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
  const from = series?.from ?? 0;
  const to = series?.to ?? 1;
  const bands = series?.bands ?? [];
  const pad = { l: 64, r: 16, t: 14, b: 28 };
  const iw = width - pad.l - pad.r;
  const ih = height - pad.t - pad.b;
  const x = (t: number) => pad.l + ((t - from) / (to - from)) * iw;
  const shown = lines.filter((l) => !hidden.has(l.name));
  const maxV = Math.max(
    1e-9,
    ...(hidden.has("Usage") ? [] : bands.map((b) => b.p95)),
    ...shown.flatMap((l) => l.points.map((p) => p.v)),
  );
  const top = niceMax(maxV * 1.1);
  const y = (v: number) => pad.t + ih - (v / top) * ih;

  const step = (pts: Point[]) => {
    let d = "";
    for (const p of pts) {
      const px = Math.max(pad.l, Math.min(pad.l + iw, x(p.t)));
      d += d ? `H${px}V${y(p.v)}` : `M${px},${y(p.v)}`;
    }
    return d && `${d}H${x(pts[pts.length - 1].t)}`;
  };
  // Buckets further apart than three widths are a gap in data, not a slope.
  const gap = bands.length > 1 ? ((to - from) / 150) * 3 : Infinity;
  const area = (k: (typeof BAND_KEYS)[number]) => {
    let d = "";
    let prev: number | null = null;
    let lastX = 0;
    for (const b of bands) {
      const px = x(b.t);
      if (prev != null && b.t - prev > gap) {
        d += `L${lastX},${y(0)}Z`;
        prev = null;
      }
      d += prev == null ? `M${px},${y(0)}L${px},${y(b[k])}` : `L${px},${y(b[k])}`;
      prev = b.t;
      lastX = px;
    }
    return prev == null ? d : `${d}L${lastX},${y(0)}Z`;
  };

  const span = to - from;
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
  const at = (pts: Point[], t: number) => {
    let best: Point | null = null;
    for (const p of pts) if (p.t <= t) best = p;
    return best;
  };
  const hb =
    hover != null && bands.length
      ? bands.reduce((b, p) => (Math.abs(p.t - hover) < Math.abs(b.t - hover) ? p : b))
      : null;
  const empty = !loading && bands.length === 0 && shown.every((l) => l.points.length === 0);

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
            setHover(t >= from && t <= to ? t : null);
          }}
        >
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
          {!hidden.has("Usage") && BAND_KEYS.map((k, i) => <path key={k} d={area(k)} className={`g-band b${2 - i}`} />)}
          {shown.map((l) => (
            <path
              key={l.name}
              d={step(l.points)}
              className="g-line"
              style={{ stroke: l.color, strokeDasharray: l.dash }}
            />
          ))}
          {(loading || empty) && (
            <text x={pad.l + iw / 2} y={pad.t + ih / 2} textAnchor="middle" className="g-axis">
              {loading ? "Loading…" : "No samples in this window yet"}
            </text>
          )}
          {hover != null && <line x1={x(hover)} x2={x(hover)} y1={pad.t} y2={pad.t + ih} className="g-cross" />}
        </svg>
        {hover != null && (hb || shown.length > 0) && (
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
            {!hidden.has("Usage") && hb && (
              <>
                <div className="tt-s">Usage:</div>
                {BAND_KEYS.map((k, i) => (
                  <div key={k} className="tt-r">
                    <i className={`sw b${2 - i}`} />
                    {BAND_LABELS[k]}: <b>{fmtY(hb[k])}</b>
                  </div>
                ))}
              </>
            )}
            {shown.map((l) => {
              const p = at(l.points, hover);
              return (
                p && (
                  <div key={l.name} className="tt-r">
                    <i className="sw" style={{ background: l.color }} />
                    {l.name}: <b>{fmtY(p.v)}</b>
                  </div>
                )
              );
            })}
          </div>
        )}
      </div>
      <div className="rz-glegend" role="group" aria-label="Series">
        <button aria-pressed={!hidden.has("Usage")} onClick={() => toggle("Usage")}>
          <i className="sw b1" />
          Usage
        </button>
        {lines.map((l) => (
          <button key={l.name} aria-pressed={!hidden.has(l.name)} onClick={() => toggle(l.name)}>
            <i className="sw line" style={{ background: l.color }} />
            {l.name}
          </button>
        ))}
      </div>
    </div>
  );
}
