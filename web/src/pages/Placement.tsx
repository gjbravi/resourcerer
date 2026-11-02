import { Info, Server } from "lucide-react";
import { useState } from "react";
import type { Detail, Placement } from "../api";
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
  const { data: p, isLoading } = usePlacement(d);
  const rates = info ? workloadRates(info.cost, d.pricing) : undefined;
  const placed = p?.available ? p : undefined;
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <h2>How the cost is calculated</h2>
        {placed && <span className="rz-dim rz-xs">over the learning period, {fmtWindow(span(placed))}</span>}
      </div>
      <div className="rz-place">
        {rates && <CostFormula d={d} rates={rates} />}
        {isLoading && <WhereSkeleton />}
        {p && !p.available && (
          <p className="rz-place-note rz-dim">
            <Info size={14} />
            {p.reason?.includes("Karpenter") ? p.reason : `Could not read where the pods ran: ${p.reason}`}
          </p>
        )}
        {placed && <Where p={placed} d={d} />}
      </div>
    </section>
  );
}

const span = (p: Placement) => (p.to ?? 0) - (p.from ?? 0);

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

/** Where the requests ran and what each instance type cost: one sentence, one bar, one table. */
function Where({ p, d }: { p: Placement; d: Detail }) {
  const monthly = nodeMonthly(p, d.current.monthlyCost);
  const { rows, others } = typeRows(p);
  const req = capacitySlices(p, "requests");
  const cost = capacitySlices(p, "cost");
  return (
    <>
      <div className="rz-sechd">
        <h3>
          <Server size={16} /> Where it ran
        </h3>
        <span className="rz-dim rz-xs">{pc(p.coverage ?? 0)} of requests on nodes with a known price</span>
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
      <div
        className="rz-place-stack"
        role="img"
        aria-label={cost.map((s) => `${s.label} ${pc(s.share)} of the cost`).join(", ")}
      >
        {cost.map((s) => (
          <i key={s.key} style={{ width: `${s.share * 100}%`, background: color(s.key) }} />
        ))}
      </div>
      <div className="rz-place-legend">
        {cost.map((s) => (
          <span key={s.key}>
            <i style={{ background: color(s.key) }} /> {s.label} <b>{pc(s.share)}</b> of the cost
            <span className="rz-dim"> · {pc(req.find((r) => r.key === s.key)?.share ?? 0)} of requests</span>
          </span>
        ))}
      </div>
      <TypeTable rows={rows} others={others} monthly={monthly} />
    </>
  );
}

function TypeTable({ rows, others, monthly }: { rows: TypeRow[]; others: TypeRow[]; monthly: number }) {
  const [open, setOpen] = useState(false);
  const shown = open ? rows.flatMap((r) => (r.count ? others : [r])) : rows;
  const max = Math.max(...shown.map((r) => r.requests), 0.0001);
  return (
    <div className="rz-tablewrap">
      <table className="rz-place-types">
        <thead>
          <tr>
            <th>Instance type</th>
            <th>Capacity</th>
            <th>Requests</th>
            <th className="num" title="At the prices of these nodes">
              Cost / month
            </th>
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
                  <i style={{ width: `${(r.requests / max) * 100}%`, background: color(r.capacity) }} />
                </span>
                {pc(r.requests)}
              </td>
              <td className="num">
                <b>{fmtMoney(monthly * r.cost)}</b> <span className="rz-dim rz-xs">{pc(r.cost)}</span>
              </td>
            </tr>
          ))}
        </tbody>
        <tfoot>
          <tr>
            <th colSpan={2}>Total</th>
            <th>100%</th>
            <th className="num">{fmtMoney(monthly)}</th>
          </tr>
        </tfoot>
      </table>
      {open && (
        <button className="rz-link small" onClick={() => setOpen(false)}>
          Fold the smaller types
        </button>
      )}
    </div>
  );
}

/** The shape of "Where it ran" while the placement loads. */
function WhereSkeleton() {
  return (
    <div className="rz-skelgroup" aria-busy="true" aria-label="Loading where it ran">
      <i className="rz-skel" style={{ width: "30%", height: 16 }} />
      <i className="rz-skel" style={{ width: "55%", height: 12 }} />
      <i className="rz-skel" style={{ height: 12, borderRadius: 99 }} />
      {[0, 1, 2].map((i) => (
        <i key={i} className="rz-skel" style={{ height: 22 }} />
      ))}
    </div>
  );
}
