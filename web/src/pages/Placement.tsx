import { ChevronRight, Info, Server } from "lucide-react";
import { useState } from "react";
import { Bar, BarChart, CartesianGrid, Tooltip, XAxis, YAxis } from "recharts";
import type { Detail, Placement } from "../api";
import { useWidth } from "../chart";
import { fmtCores, fmtMoney, Gi } from "../format";
import { useInfo, usePlacement } from "../hooks";
import {
  capacityLabel,
  capacitySlices,
  fmtWindow,
  nodeMonthly,
  placementSummary,
  typeRows,
  UNPRICED,
  workloadRates,
  type Measure,
  type TypeRow,
} from "../placement";

const COLORS: Record<string, string> = {
  "on-demand": "var(--c-curlim)",
  reserved: "var(--c-cur)",
  spot: "var(--c-rec)",
  [UNPRICED]: "var(--c-alloc)",
  mixed: "var(--c-usage50)",
};
const color = (c: string) => COLORS[c] ?? "var(--c-usage)";
const pc = (v: number) => (v > 0 && v < 0.01 ? "<1%" : `${Math.round(v * 100)}%`);
const rel = (v: number) => `${v.toFixed(2)}×`;

/** How the workload's monthly cost is calculated, and where its requests ran during the learning period. */
export function CostSection({ d }: { d: Detail }) {
  const info = useInfo().data;
  const [open, setOpen] = useState(false);
  const { data: p, isLoading } = usePlacement(d, open);
  const [m, setM] = useState<Measure>("requests");
  const rates = info ? workloadRates(info.cost, d.pricing) : undefined;
  const placed = p?.available ? p : undefined;
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <button className="rz-costtoggle" aria-expanded={open} onClick={() => setOpen(!open)}>
          <ChevronRight size={16} className="chev" />
          <span>
            <b>How the cost is calculated</b>
            <span className="rz-dim rz-xs">
              {" "}
              · requests × {d.avgReplicas.toFixed(1)} replicas ×{" "}
              {d.pricing?.allocation === "node" && rates
                ? `its nodes' prices (${rel(rates.relative)} the cluster average)`
                : "the cluster's rates"}{" "}
              × 730 h, and where it ran
            </span>
          </span>
        </button>
        {open && placed && (
          <div className="rz-zoom" role="group" aria-label="Share of">
            {(["requests", "cost"] as const).map((k) => (
              <button key={k} aria-pressed={m === k} onClick={() => setM(k)}>
                {k === "requests" ? "Requests" : "Cost"}
              </button>
            ))}
          </div>
        )}
      </div>
      {open && (
        <div className="rz-place">
          {rates && <CostFormula d={d} rates={rates} />}
          {isLoading && <p className="rz-dim rz-xs">Loading where it ran…</p>}
          {p && !p.available && (
            <p className="rz-place-note rz-dim">
              <Info size={14} />
              {p.reason?.includes("Karpenter") ? p.reason : `Could not read where the pods ran: ${p.reason}`}
            </p>
          )}
          {placed && <Where p={placed} d={d} m={m} />}
        </div>
      )}
    </section>
  );
}

/** Requests × rates × 730 hours, now and with the recommendation, with the rates this workload pays. */
function CostFormula({ d, rates }: { d: Detail; rates: ReturnType<typeof workloadRates> }) {
  const row = (label: string, t: Detail["current"]) => {
    const gib = t.memory / Gi;
    return (
      <tr>
        <th>{label}</th>
        <td>
          {fmtCores(t.cpu)} × {fmtMoney(rates.cpuCoreHour, 4)}
        </td>
        <td>
          + {gib.toFixed(2)} GiB × {fmtMoney(rates.memoryGiBHour, 4)}
        </td>
        <td>× 730 h</td>
        <td className="num">
          <b>{fmtMoney(t.monthlyCost)}</b>
        </td>
      </tr>
    );
  };
  return (
    <div className="rz-place-formula">
      <div className="rz-tablewrap">
        <table>
          <tbody>
            {row("Now", d.current)}
            {row("Recommended", d.recommended)}
          </tbody>
        </table>
      </div>
      <p className="rz-dim rz-xs">
        Requests summed over containers × {d.avgReplicas.toFixed(1)} average replicas, at{" "}
        {d.pricing?.allocation === "node"
          ? `the prices of the nodes it ran on (${rel(rates.relative)} the cluster average).`
          : "the cluster's average rates."}{" "}
        <a href="#/settings">Rates</a>
      </p>
    </div>
  );
}

function Where({ p, d, m }: { p: Placement; d: Detail; m: Measure }) {
  const monthly = nodeMonthly(p, d.current.monthlyCost);
  const { rows, others } = typeRows(p);
  return (
    <>
      <div className="rz-sechd">
        <h3>
          <Server size={16} /> Where it ran
        </h3>
        <span className="rz-dim rz-xs">
          last {fmtWindow((p.to ?? 0) - (p.from ?? 0))} · {pc(p.coverage ?? 0)} on priced nodes
        </span>
      </div>
      <p className="rz-place-sum">
        <b>{placementSummary(p)}</b>
        {p.allocation === "cluster" && (
          <span className="rz-dim">
            {" "}
            Costs use the cluster average (cost.allocation: cluster); at its nodes&apos; prices it would be{" "}
            {fmtMoney(monthly)}/mo.
          </span>
        )}
      </p>
      <CapacityBar p={p} m={m} />
      <Timeline p={p} m={m} />
      <TypeTable rows={rows} others={others} m={m} monthly={monthly} />
    </>
  );
}

function CapacityBar({ p, m }: { p: Placement; m: Measure }) {
  const slices = capacitySlices(p, m);
  return (
    <>
      <div className="rz-place-stack" role="img" aria-label={slices.map((s) => `${s.label} ${pc(s.share)}`).join(", ")}>
        {slices.map((s) => (
          <i
            key={s.key}
            title={`${s.label}: ${pc(s.share)}`}
            style={{ width: `${s.share * 100}%`, background: color(s.key) }}
          />
        ))}
      </div>
      <div className="rz-place-legend">
        {slices.map((s) => (
          <span key={s.key}>
            <i style={{ background: color(s.key) }} /> {s.label} <b>{pc(s.share)}</b>
          </span>
        ))}
      </div>
    </>
  );
}

function Timeline({ p, m }: { p: Placement; m: Measure }) {
  const keys = capacitySlices(p, "requests").map((s) => s.key);
  const rows = (p.timeline ?? []).map((b) => ({ t: b.t, ...b[m] }));
  const span = (p.to ?? 0) - (p.from ?? 0);
  const tick = (t: number) =>
    new Date(t * 1000).toLocaleString(
      [],
      span > 2 * 86400 ? { weekday: "short", day: "numeric" } : { hour: "numeric", minute: "2-digit" },
    );
  const [ref, width] = useWidth();
  if (rows.length < 2) return <p className="rz-dim rz-xs">Not enough history for a timeline yet.</p>;
  return (
    <div ref={ref} className="rz-place-tl">
      <BarChart
        width={width}
        height={150}
        data={rows}
        barCategoryGap={1}
        margin={{ top: 4, right: 18, bottom: 0, left: 0 }}
      >
        <CartesianGrid vertical={false} stroke="var(--track)" strokeDasharray="3 3" />
        <XAxis dataKey="t" tickFormatter={tick} className="g-axis" tickLine={false} axisLine={false} minTickGap={40} />
        <YAxis
          domain={[0, 1]}
          ticks={[0, 0.5, 1]}
          tickFormatter={(v: number) => `${v * 100}%`}
          className="g-axis"
          tickLine={false}
          axisLine={false}
          width={46}
        />
        <Tooltip
          cursor={{ fill: "var(--accent-soft)" }}
          wrapperStyle={{ outline: "none", zIndex: 5 }}
          content={({ active, payload }) => {
            const row = payload?.[0]?.payload as ({ t: number } & Record<string, number>) | undefined;
            if (!active || !row) return null;
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
                {keys
                  .filter((k) => row[k] > 0)
                  .map((k) => (
                    <div key={k} className="tt-r">
                      <span style={{ color: `color-mix(in srgb, ${color(k)} 60%, var(--text))` }}>
                        {capacityLabel(k)}
                      </span>
                      <b>{pc(row[k])}</b>
                    </div>
                  ))}
              </div>
            );
          }}
        />
        {keys.map((k) => (
          <Bar key={k} dataKey={k} stackId="c" fill={color(k)} isAnimationActive={false} />
        ))}
      </BarChart>
    </div>
  );
}

function TypeTable({ rows, others, m, monthly }: { rows: TypeRow[]; others: TypeRow[]; m: Measure; monthly: number }) {
  const [open, setOpen] = useState(false);
  const shown = open ? rows.flatMap((r) => (r.count ? others : [r])) : rows;
  const max = Math.max(...shown.map((r) => r[m]), 0.0001);
  return (
    <div className="rz-tablewrap">
      <table className="rz-place-types">
        <thead>
          <tr>
            <th>Instance type</th>
            <th>Capacity</th>
            <th>{m === "requests" ? "Requests" : "Cost"}</th>
            <th className="num" title="Price of its capacity relative to the cluster average">
              vs average
            </th>
            <th className="num">$/mo</th>
          </tr>
        </thead>
        <tbody>
          {shown.map((r) => (
            <tr key={r.capacity + r.name} className={r.capacity === UNPRICED ? "rz-dim" : undefined}>
              <td>
                {r.count ? (
                  <button className="rz-link" onClick={() => setOpen(true)}>
                    {r.name}
                  </button>
                ) : (
                  r.name
                )}
              </td>
              <td>
                <i className="rz-place-dot" style={{ background: color(r.capacity) }} /> {capacityLabel(r.capacity)}
              </td>
              <td>
                <span className="rz-place-mini">
                  <i style={{ width: `${(r[m] / max) * 100}%`, background: color(r.capacity) }} />
                </span>
                {pc(r[m])}
              </td>
              <td className="num">{r.capacity === UNPRICED ? "average" : rel(r.relative)}</td>
              <td className="num">{fmtMoney(monthly * r.cost)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {open && (
        <button className="rz-link small" onClick={() => setOpen(false)}>
          Fold the smaller types
        </button>
      )}
    </div>
  );
}
