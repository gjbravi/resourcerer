import { Activity, Boxes, Cpu, Flame, Info, MemoryStick, PiggyBank, ShieldCheck, Zap } from "lucide-react";
import type { ReactNode } from "react";
import type { Summary } from "../api";
import { Art, Gauge } from "../art";
import { Delta, Money, PHASE_NAMES, phaseTone, Ring, StatusBadge, WlCell } from "../components";
import { fmtCores, fmtMem, fmtMoney } from "../format";
import { navigate, workloadHash } from "../hooks";
import { applicable } from "./Workloads";

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
  const cur = {
    cost: sum((w) => w.current.monthlyCost),
    cpu: sum((w) => w.current.cpu),
    memory: sum((w) => w.current.memory),
  };
  const rec = {
    cost: sum((w) => w.recommended.monthlyCost),
    cpu: sum((w) => w.recommended.cpu),
    memory: sum((w) => w.recommended.memory),
  };
  const reclaim = sum((w) => Math.max(0, net(w)));
  const savings = rows
    .filter((w) => net(w) > 0.5)
    .sort((a, b) => net(b) - net(a))
    .slice(0, 5);
  const risks = rows
    .filter((w) => w.under || w.oomKills > 0 || w.noRequests)
    .sort((a, b) => b.oomKills - a.oomKills || net(a) - net(b))
    .slice(0, 5);
  const ready = rows.filter((w) => applicable(w) && !w.autoDeploy).length;
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
            {rows.length} workloads in {cluster}, compared with what their usage history recommends.
          </p>
        </div>
      </div>

      <section className="rz-card rz-hero">
        <Art kind="potion" />
        <div className="rz-herobody">
          <h3>
            Reclaimable from over-provisioned workloads{" "}
            <Info size={14} aria-label="Sum of savings where requests are above the recommendation" />
          </h3>
          <Money value={reclaim} />
          <div className="rz-heromsg">
            Cluster total <Delta from={cur.cost} to={rec.cost} money />{" "}
            <span className="rz-dim">
              · {fmtMoney(cur.cost)} → {fmtMoney(rec.cost)} per month
            </span>
          </div>
          <div className="rz-recactions">
            <a className="rz-btn primary" href="#/workloads?view=pending">
              Review {ready} ready to apply
            </a>
            <a className="rz-btn ghost" href="#/workloads?view=reliability">
              <ShieldCheck size={15} /> {risks.length} reliability risks
            </a>
          </div>
        </div>
        <div className="rz-heroside">
          <Gauge value={score} />
          <span className="rz-dim rz-xs">
            Fit score · {scored.length} of {rows.length} scored
          </span>
        </div>
      </section>

      <section className="rz-card rz-kpis">
        <div className="rz-kpi">
          <span>
            <Cpu size={15} /> CPU requested
          </span>
          <b>
            {fmtCores(rec.cpu)} <Delta from={cur.cpu} to={rec.cpu} r="cpu" />
          </b>
          <span className="rz-dim rz-xs">from {fmtCores(cur.cpu)}</span>
        </div>
        <div className="rz-kpi">
          <span>
            <MemoryStick size={15} /> Memory requested
          </span>
          <b>
            {fmtMem(rec.memory)} <Delta from={cur.memory} to={rec.memory} r="memory" />
          </b>
          <span className="rz-dim rz-xs">from {fmtMem(cur.memory)}</span>
        </div>
        <div className="rz-kpi">
          <span>
            <Boxes size={15} /> Workloads
          </span>
          <b>{rows.length}</b>
          <span className="rz-dim rz-xs">
            {rows.filter((w) => w.phase === "Ready").length} ready · {rows.filter((w) => w.phase === "Learning").length}{" "}
            learning
          </span>
        </div>
        <div className="rz-kpi">
          <span>
            <Zap size={15} /> Auto-deploy
          </span>
          <b>
            {rows.filter((w) => w.autoDeploy).length}
            <span className="rz-dim rz-xs">of {rows.length}</span>
          </b>
          <span className="rz-dim rz-xs">applied without review</span>
        </div>
      </section>

      <div className="rz-grid2">
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <PiggyBank size={17} className="t-accent" /> Top savings
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
              ["Fit", (w) => <Ring value={w.score?.overall} size={30} stroke={3.5} />],
              ["Per month", (w) => <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />],
            ]}
          />
        </section>
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <ShieldCheck size={17} className="t-accent" /> Reliability risks
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
              ["Per month", (w) => <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />],
            ]}
          />
        </section>
      </div>

      <div className="rz-grid2">
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <Activity size={17} className="t-accent" /> Recommendation status
            </h2>
          </div>
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
        </section>
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>Namespaces</h2>
          </div>
          <table className="rz-table">
            <thead>
              <tr>
                <th>Namespace</th>
                <th className="num">Workloads</th>
                <th className="num">Cost / mo</th>
                <th className="num">Change</th>
                <th className="num">Auto</th>
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
                  <td className="num">{fmtMoney(n.cur)}</td>
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
    </div>
  );
}
