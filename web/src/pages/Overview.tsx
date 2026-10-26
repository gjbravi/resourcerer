import { Flame, Info, PiggyBank, ShieldCheck, Zap } from "lucide-react";
import type { ReactNode } from "react";
import type { Summary } from "../api";
import { Delta, PHASE_NAMES, phaseTone, Ring, StatusBadge } from "../components";
import { fmtCores, fmtMem, fmtMoney } from "../format";
import { navigate, workloadHash } from "../hooks";
import { ArrowRight } from "lucide-react";

export function SummaryGrid({
  cur,
  rec,
}: {
  cur: { cost: number; cpu: number; memory: number };
  rec: { cost: number; cpu: number; memory: number };
}) {
  return (
    <div className="rz-sumgrid">
      <span className="h">Current Requests</span>
      <span />
      <span className="h">Recommended</span>
      <span className="h">Net Impact</span>
      <b>
        {fmtMoney(cur.cost)} <small>/mo.</small>
      </b>
      <ArrowRight size={14} className="rz-dim" />
      <b>
        {fmtMoney(rec.cost)} <small>/mo.</small>
      </b>
      <Delta from={cur.cost} to={rec.cost} money />
      <b>{fmtCores(cur.cpu)}</b>
      <ArrowRight size={14} className="rz-dim" />
      <b>{fmtCores(rec.cpu)}</b>
      <Delta from={cur.cpu} to={rec.cpu} r="cpu" />
      <b>{fmtMem(cur.memory)}</b>
      <ArrowRight size={14} className="rz-dim" />
      <b>{fmtMem(rec.memory)}</b>
      <Delta from={cur.memory} to={rec.memory} r="memory" />
    </div>
  );
}

export const WlCell = ({ w }: { w: Summary }) => (
  <a className="rz-wl" href={workloadHash(w)} onClick={(e) => e.stopPropagation()}>
    <b>{w.name}</b>
    <span className="rz-dim rz-xs">
      {w.namespace} · {w.kind}
    </span>
  </a>
);

const net = (w: Summary) => w.current.monthlyCost - w.recommended.monthlyCost;

function MiniTable({
  rows,
  cols,
  empty,
}: {
  rows: Summary[];
  cols: [string, (w: Summary) => ReactNode][];
  empty: string;
}) {
  if (!rows.length) return <p className="rz-empty">{empty}</p>;
  return (
    <table className="rz-table">
      <thead>
        <tr>
          {cols.map(([h]) => (
            <th key={h}>{h}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((w) => (
          <tr key={`${w.namespace}/${w.kind}/${w.name}`} className="rz-click" onClick={() => navigate(workloadHash(w))}>
            {cols.map(([h, f]) => (
              <td key={h}>{f(w)}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function Overview({ rows, cluster }: { rows: Summary[]; cluster: string }) {
  const sum = (f: (w: Summary) => number) => rows.reduce((a, w) => a + f(w), 0);
  const scored = rows.filter((w) => w.score);
  const weight = scored.reduce((a, w) => a + w.recommended.monthlyCost, 0);
  const score = scored.length
    ? scored.reduce((a, w) => a + w.score!.overall * w.recommended.monthlyCost, 0) / (weight || 1)
    : undefined;
  const savings = rows
    .filter((w) => net(w) > 0.5)
    .sort((a, b) => net(b) - net(a))
    .slice(0, 5);
  const risks = rows
    .filter((w) => w.under || w.oomKills > 0 || w.noRequests)
    .sort((a, b) => b.oomKills - a.oomKills || net(a) - net(b))
    .slice(0, 5);
  const namespaces = [...new Set(rows.map((w) => w.namespace))]
    .map((ns) => {
      const r = rows.filter((w) => w.namespace === ns);
      return {
        ns,
        n: r.length,
        cur: r.reduce((a, w) => a + w.current.monthlyCost, 0),
        rec: r.reduce((a, w) => a + w.recommended.monthlyCost, 0),
        auto: r.filter((w) => w.autoDeploy).length,
        annotated: r.some((w) => w.namespaceAnnotated),
      };
    })
    .sort((a, b) => b.cur - b.rec - (a.cur - a.rec));

  return (
    <div className="rz-page">
      <div className="rz-pagehd">
        <div>
          <h1>Overview</h1>
          <p className="rz-dim">
            Requests across {rows.length} workloads in {cluster}, compared with what their usage history recommends.
          </p>
        </div>
      </div>
      <div className="rz-hero">
        <div className="rz-herocell score">
          <h3>
            Optimization Score{" "}
            <Info
              size={14}
              className="rz-dim"
              aria-label="How well current requests match recommendations, weighted by cost"
            />
          </h3>
          <div className="rz-scorerow">
            <Ring value={score} size={96} stroke={9} />
            <div className="rz-dim">
              {scored.length} scored
              <br />
              {rows.length - scored.length} without score
            </div>
          </div>
        </div>
        <div className="rz-herocell summary">
          <h3>Recommendation Summary</h3>
          <SummaryGrid
            cur={{
              cost: sum((w) => w.current.monthlyCost),
              cpu: sum((w) => w.current.cpu),
              memory: sum((w) => w.current.memory),
            }}
            rec={{
              cost: sum((w) => w.recommended.monthlyCost),
              cpu: sum((w) => w.recommended.cpu),
              memory: sum((w) => w.recommended.memory),
            }}
          />
        </div>
        <div className="rz-herocell">
          <h3>Workloads</h3>
          <div className="rz-statusbars">
            {PHASE_NAMES.filter((p) => rows.some((w) => w.phase === p) || ["Ready", "Learning"].includes(p)).map(
              (p) => {
                const n = rows.filter((w) => w.phase === p).length;
                return (
                  <a key={p} href={`#/workloads?status=${p}`} className="rz-statusbar">
                    <StatusBadge phase={p} />
                    <span className="bar">
                      <span
                        className={`fill t-${phaseTone(p)}`}
                        style={{ width: `${rows.length ? (n / rows.length) * 100 : 0}%` }}
                      />
                    </span>
                    <b>{n}</b>
                  </a>
                );
              },
            )}
          </div>
          <p className="rz-dim rz-xs" style={{ marginTop: 10 }}>
            <Zap size={12} /> Automatic deployment on {rows.filter((w) => w.autoDeploy).length} of {rows.length}
          </p>
        </div>
      </div>

      <div className="rz-grid2">
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <PiggyBank size={16} /> Top savings opportunities
            </h2>
            <a className="rz-link" href="#/workloads?view=savings">
              View all
            </a>
          </div>
          <MiniTable
            rows={savings}
            empty="Nothing to reclaim."
            cols={[
              ["Workload", (w) => <WlCell w={w} />],
              ["Score", (w) => <Ring value={w.score?.overall} size={30} stroke={3.5} />],
              ["Net impact", (w) => <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />],
            ]}
          />
        </section>
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <ShieldCheck size={16} /> Reliability risks
            </h2>
            <a className="rz-link" href="#/workloads?view=reliability">
              View all
            </a>
          </div>
          <MiniTable
            rows={risks}
            empty="No workload is short on resources."
            cols={[
              ["Workload", (w) => <WlCell w={w} />],
              [
                "Why",
                (w) => (
                  <span className="rz-xs">
                    {w.oomKills > 0 ? (
                      <span className="t-bad">
                        <Flame size={12} /> {w.oomKills} OOM kills
                      </span>
                    ) : w.noRequests ? (
                      "No requests set"
                    ) : (
                      "Requests below usage"
                    )}
                  </span>
                ),
              ],
              ["Net impact", (w) => <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />],
            ]}
          />
        </section>
      </div>

      <section className="rz-card">
        <div className="rz-cardhd">
          <h2>Namespaces</h2>
        </div>
        <table className="rz-table">
          <thead>
            <tr>
              <th>Namespace</th>
              <th className="num">Workloads</th>
              <th className="num">Current cost</th>
              <th className="num">Net impact</th>
              <th className="num">Auto-deploy</th>
            </tr>
          </thead>
          <tbody>
            {namespaces.map((n) => (
              <tr key={n.ns} className="rz-click" onClick={() => navigate(`#/workloads?ns=${n.ns}`)}>
                <td>
                  <b>{n.ns}</b>
                  {n.annotated && (
                    <span className="rz-src s-namespace" title="The namespace has resourcerer.io annotations">
                      annotated
                    </span>
                  )}
                </td>
                <td className="num">{n.n}</td>
                <td className="num">{fmtMoney(n.cur)}/mo</td>
                <td className="num">
                  <Delta from={n.cur} to={n.rec} money />
                </td>
                <td className="num">
                  {n.auto}/{n.n}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  );
}
