import {
  Activity,
  Boxes,
  ChevronDown,
  ChevronUp,
  Cpu,
  Flame,
  Info,
  MemoryStick,
  PiggyBank,
  Search,
  ShieldCheck,
  Zap,
} from "lucide-react";
import { useState, type ReactNode } from "react";
import type { Summary } from "../api";
import { NoteArt, WandArt } from "../art";
import { Delta, FitBar, Metric, Money, PHASE_NAMES, phaseTone, Ring, StatusBadge, WlCell } from "../components";
import { costSource, fmtCores, fmtMem, fmtMoney } from "../format";
import { navigate, useInfo, workloadHash } from "../hooks";
import { ClusterTimeline } from "./ClusterTimeline";
import { signals } from "../reliability";
import { applicable, atRisk } from "./Workloads";

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

interface NamespaceRow {
  ns: string;
  n: number;
  cur: number;
  rec: number;
  auto: number;
  annotated: boolean;
}

const TOP_NAMESPACES = 8;

/** The namespaces with the most to change; the rest behind a search and a scrollable list. */
function NamespacesCard({ namespaces }: { namespaces: NamespaceRow[] }) {
  const [q, setQ] = useState("");
  const [all, setAll] = useState(false);
  const matches = namespaces.filter((n) => n.ns.includes(q.trim().toLowerCase()));
  const shown = all || q ? matches : matches.slice(0, TOP_NAMESPACES);
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <h2>Namespaces</h2>
        {namespaces.length > TOP_NAMESPACES && (
          <label className="rz-input small">
            <Search size={13} />
            <input
              placeholder={`Search ${namespaces.length} namespaces`}
              value={q}
              onChange={(e) => setQ(e.target.value)}
              aria-label="Search namespaces"
            />
          </label>
        )}
      </div>
      <div className={all || q ? "rz-scrolltable" : undefined}>
        <table className="rz-table rz-nstable">
          <thead>
            <tr>
              <th>Namespace</th>
              <th className="num">Cost / mo</th>
              <th className="num">Change</th>
            </tr>
          </thead>
          <tbody>
            {shown.map((n) => (
              <tr key={n.ns} className="rz-click" onClick={() => navigate(`#/workloads?ns=${n.ns}`)}>
                <td>
                  <span className="rz-nsname">
                    <b title={n.ns}>{n.ns}</b>
                    {n.annotated && (
                      <span className="rz-src s-namespace" title="The namespace has resourcerer.io annotations">
                        annotated
                      </span>
                    )}
                  </span>
                  <span className="rz-dim rz-xs">
                    {n.n} {n.n === 1 ? "workload" : "workloads"} · {n.auto} auto-deploy
                  </span>
                </td>
                <td className="num">{fmtMoney(n.cur)}</td>
                <td className="num">
                  <Delta from={n.cur} to={n.rec} money />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {!shown.length && <p className="rz-empty">No namespace matches “{q}”.</p>}
      </div>
      {!q && namespaces.length > TOP_NAMESPACES && (
        <button className="rz-showall" onClick={() => setAll(!all)}>
          {all ? `Show top ${TOP_NAMESPACES}` : `Show all ${namespaces.length} namespaces`}
          {all ? <ChevronUp size={14} /> : <ChevronDown size={14} />}
        </button>
      )}
    </section>
  );
}

export function Overview({ rows, cluster }: { rows: Summary[]; cluster: string }) {
  const info = useInfo().data;
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
    .filter(atRisk)
    .sort((a, b) => b.oomKills - a.oomKills || signals(b).length - signals(a).length || net(a) - net(b))
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

      <div className="rz-pillbar">
        <a className="rz-btn primary" href="#/workloads?view=pending">
          Review {ready} ready to apply
        </a>
        <a className="rz-btn ghost" href="#/workloads?view=reliability">
          <ShieldCheck size={15} /> {risks.length} reliability risks
        </a>
        <a className="rz-btn ghost" href="#/workloads?view=savings">
          <PiggyBank size={15} className="t-save" /> {savings.length} savings
        </a>
      </div>

      <section className="rz-metrics">
        <Metric
          label={
            <>
              Reclaimable <Info size={15} aria-label="Sum of savings where requests are above the recommendation" />
            </>
          }
          art={<NoteArt />}
          trend={
            <>
              <Delta from={cur.cost} to={rec.cost} money />{" "}
              <span className="rz-dim">
                cluster total · {fmtMoney(cur.cost)} → {fmtMoney(rec.cost)}
              </span>
              {info && (
                <a className="rz-costsrc" href="#/settings" title="How costs are estimated">
                  {costSource(info.cost)}
                </a>
              )}
            </>
          }
        >
          <Money value={reclaim} tone="save" />
        </Metric>
        <Metric
          label={
            <>
              Fit score{" "}
              <Info size={15} aria-label="How closely current requests match the recommendations, weighted by cost" />
            </>
          }
          art={<WandArt />}
          trend={
            <span className="rz-dim">
              {scored.length} of {rows.length} workloads scored
            </span>
          }
        >
          <div className="rz-money">
            {score == null ? "–" : Math.round(score)}
            <span className="dec">/100</span>
          </div>
          <FitBar value={score} />
        </Metric>
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

      <ClusterTimeline />

      <div className="rz-grid2">
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <PiggyBank size={17} className="t-save" /> Top savings
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
                    {signals(w)[0] ? (
                      <span className={`t-${signals(w)[0].tone}`}>
                        <Flame size={12} /> {signals(w)[0].title}
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
        <NamespacesCard namespaces={namespaces} />
      </div>
    </div>
  );
}
