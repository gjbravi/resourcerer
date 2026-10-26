import { useQueryClient } from "@tanstack/react-query";
import { ArrowDown, ArrowRight, ArrowUp, Flame, Search, Zap, ZapOff } from "lucide-react";
import { useState } from "react";
import { api, refKey, type Phase, type Resource, type Summary } from "../api";
import { Delta, Modal, PHASE_NAMES, Ring, StatusBadge, WlCell } from "../components";
import { fmt } from "../format";
import { navigate, toast, workloadHash } from "../hooks";

const net = (w: Summary) => w.current.monthlyCost - w.recommended.monthlyCost;

export const VIEWS: Record<string, [string, (w: Summary) => boolean]> = {
  all: ["All", () => true],
  savings: ["Savings", (w) => net(w) > 0.5],
  reliability: ["Reliability", (w) => w.under || w.oomKills > 0 || w.noRequests],
  pending: ["Ready to apply", (w) => applicable(w) && !w.autoDeploy],
  auto: ["Auto-deploy on", (w) => w.autoDeploy],
};

export const applicable = (w: Summary) => !!w.hash && w.pending && (w.phase === "Ready" || w.phase === "Learning");

const COLS: Record<string, (w: Summary) => number | string> = {
  name: (w) => refKey(w),
  status: (w) => w.phase,
  score: (w) => w.score?.overall ?? -1,
  cpu: (w) => w.current.cpu - w.recommended.cpu,
  memory: (w) => w.current.memory - w.recommended.memory,
  net,
};

function ReqDiff({ r, a, b, noReq }: { r: Resource; a: number; b: number; noReq: boolean }) {
  const same = Math.abs(a - b) < (r === "cpu" ? 0.0005 : 2 ** 19);
  return (
    <span className="rz-reqdiff">
      <span className={noReq && !a ? "t-warn" : ""}>{noReq && !a ? "none" : fmt(r, a)}</span>
      {!same && (
        <>
          <ArrowRight size={12} className="rz-dim" />
          <b>{fmt(r, b)}</b>
        </>
      )}
    </span>
  );
}

export function Workloads({ rows, query, readOnly }: { rows: Summary[]; query: string; readOnly: boolean }) {
  const params = new URLSearchParams(query);
  const qc = useQueryClient();
  const [view, setView] = useState(params.get("view") ?? "all");
  const [ns, setNs] = useState(params.get("ns") ?? "");
  const [status, setStatus] = useState(params.get("status") ?? "");
  const [q, setQ] = useState("");
  const [sort, setSort] = useState({ col: "net", dir: -1 });
  const [sel, setSel] = useState<Set<string>>(() => new Set());
  const [bulk, setBulk] = useState(false);
  const [busy, setBusy] = useState(false);

  const namespaces = [...new Set(rows.map((w) => w.namespace))].sort();
  const scoped = rows.filter(
    (w) =>
      (!ns || w.namespace === ns) &&
      (!status || w.phase === status) &&
      (!q || refKey(w).toLowerCase().includes(q.toLowerCase())),
  );
  const shown = scoped.filter(VIEWS[view]?.[1] ?? (() => true)).sort((a, b) => {
    const x = COLS[sort.col](a);
    const y = COLS[sort.col](b);
    return (typeof x === "string" ? x.localeCompare(y as string) : x - (y as number)) * sort.dir;
  });
  const selected = shown.filter((w) => sel.has(refKey(w)));
  const toApply = selected.filter(applicable);
  const toggle = (k: string) =>
    setSel((s) => {
      const n = new Set(s);
      if (n.has(k)) n.delete(k);
      else n.add(k);
      return n;
    });
  const head = (col: string, label: string, num?: boolean) => (
    <th
      className={num ? "num" : ""}
      aria-sort={sort.col === col ? (sort.dir > 0 ? "ascending" : "descending") : "none"}
    >
      <button
        className="rz-sort"
        onClick={() => setSort((s) => ({ col, dir: s.col === col ? -s.dir : col === "name" ? 1 : -1 }))}
      >
        {label} {sort.col === col && (sort.dir > 0 ? <ArrowUp size={12} /> : <ArrowDown size={12} />)}
      </button>
    </th>
  );

  const applyAll = async () => {
    setBusy(true);
    const results = await Promise.allSettled(toApply.map((w) => api.apply(w, w.hash!)));
    const failed = results.filter((r) => r.status === "rejected").length;
    setBusy(false);
    setBulk(false);
    setSel(new Set());
    qc.invalidateQueries({ queryKey: ["workloads"] });
    toast(
      failed
        ? {
            kind: "error",
            title: `${failed} of ${toApply.length} apply requests failed`,
            body: "Open the workloads to see why.",
          }
        : {
            kind: "ok",
            title: `Requested ${toApply.length} applies`,
            body: "The applier rolls them out; each workload shows the result.",
          },
    );
  };

  return (
    <div className="rz-page">
      <div className="rz-pagehd">
        <div>
          <h1>Workloads</h1>
          <p className="rz-dim">Deployments, StatefulSets and DaemonSets with a recommendation.</p>
        </div>
      </div>
      <section className="rz-card">
        <div className="rz-toolbar">
          <div className="rz-tabs small" role="tablist">
            {Object.entries(VIEWS).map(([id, [label, f]]) => (
              <button key={id} role="tab" aria-selected={view === id} onClick={() => setView(id)}>
                {label} <span className="n">{scoped.filter(f).length}</span>
              </button>
            ))}
          </div>
          <div className="rz-tools">
            <label className="rz-input">
              <Search size={14} />
              <input
                placeholder="Search workloads"
                value={q}
                onChange={(e) => setQ(e.target.value)}
                aria-label="Search workloads"
              />
            </label>
            <select className="rz-select" value={ns} onChange={(e) => setNs(e.target.value)} aria-label="Namespace">
              <option value="">All namespaces</option>
              {namespaces.map((n) => (
                <option key={n}>{n}</option>
              ))}
            </select>
            <select
              className="rz-select"
              value={status}
              onChange={(e) => setStatus(e.target.value as Phase)}
              aria-label="Status"
            >
              <option value="">All statuses</option>
              {PHASE_NAMES.map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </div>
        </div>
        {selected.length > 0 && (
          <div className="rz-bulk">
            <span>
              <b>{selected.length}</b> selected · net impact{" "}
              <Delta
                from={selected.reduce((a, w) => a + w.current.monthlyCost, 0)}
                to={selected.reduce((a, w) => a + w.recommended.monthlyCost, 0)}
                money
              />
            </span>
            <span className="rz-grow" />
            <button className="rz-btn ghost small" onClick={() => setSel(new Set())}>
              Clear
            </button>
            <button
              className="rz-btn primary small"
              disabled={readOnly || !toApply.length}
              onClick={() => setBulk(true)}
            >
              Apply {toApply.length} recommendations
            </button>
          </div>
        )}
        <div className="rz-tablewrap">
          <table className="rz-table big">
            <thead>
              <tr>
                <th className="chk">
                  <input
                    type="checkbox"
                    aria-label="Select all ready to apply"
                    disabled={readOnly}
                    checked={shown.some(applicable) && shown.filter(applicable).every((w) => sel.has(refKey(w)))}
                    onChange={(e) =>
                      setSel(e.target.checked ? new Set(shown.filter(applicable).map(refKey)) : new Set())
                    }
                  />
                </th>
                {head("name", "Workload")}
                {head("status", "Status")}
                {head("score", "Score")}
                {head("cpu", "CPU requests", true)}
                {head("memory", "Memory requests", true)}
                {head("net", "Net impact /mo", true)}
                <th>Auto</th>
              </tr>
            </thead>
            <tbody>
              {shown.map((w) => (
                <tr
                  key={refKey(w)}
                  className="rz-click"
                  tabIndex={0}
                  onClick={() => navigate(workloadHash(w))}
                  onKeyDown={(e) => e.key === "Enter" && navigate(workloadHash(w))}
                >
                  <td className="chk" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      disabled={readOnly || !applicable(w)}
                      checked={sel.has(refKey(w))}
                      aria-label={`Select ${w.name}`}
                      title={applicable(w) ? "" : "Nothing to apply"}
                      onChange={() => toggle(refKey(w))}
                    />
                  </td>
                  <td>
                    <WlCell w={w} />
                  </td>
                  <td>
                    <StatusBadge phase={w.phase} title={w.message} />
                    {w.oomKills > 0 && (
                      <span className="rz-oom" title="OOM kills during the learning period">
                        <Flame size={12} />
                        {w.oomKills}
                      </span>
                    )}
                  </td>
                  <td>
                    <Ring value={w.score?.overall} size={32} stroke={3.5} />
                  </td>
                  <td className="num">
                    <ReqDiff r="cpu" a={w.current.cpu} b={w.recommended.cpu} noReq={w.noRequests} />
                  </td>
                  <td className="num">
                    <ReqDiff r="memory" a={w.current.memory} b={w.recommended.memory} noReq={w.noRequests} />
                  </td>
                  <td className="num">
                    <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />
                  </td>
                  <td>
                    {w.autoDeploy ? (
                      <span className="rz-auto on" title="Automatic deployment on">
                        <Zap size={15} />
                      </span>
                    ) : (
                      <span className="rz-auto" title="Automatic deployment off">
                        <ZapOff size={15} />
                      </span>
                    )}
                  </td>
                </tr>
              ))}
              {!shown.length && (
                <tr>
                  <td colSpan={8} className="rz-empty">
                    No workloads match.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
        <p className="rz-foot rz-dim rz-xs">
          Requests are per-container totals × average replicas. Costs use the rates in Settings.
        </p>
      </section>
      {bulk && (
        <Modal
          title={`Apply ${toApply.length} recommendations?`}
          onClose={() => setBulk(false)}
          actions={
            <>
              <button className="rz-btn ghost" onClick={() => setBulk(false)}>
                Cancel
              </button>
              <button className="rz-btn primary" disabled={busy} onClick={() => void applyAll()}>
                Apply
              </button>
            </>
          }
        >
          <p>Each workload gets a rolling update with its recommended requests and limits.</p>
          <table className="rz-table compact">
            <thead>
              <tr>
                <th>Workload</th>
                <th className="num">Net impact</th>
              </tr>
            </thead>
            <tbody>
              {toApply.map((w) => (
                <tr key={refKey(w)}>
                  <td>
                    {w.namespace}/{w.name}
                    {w.phase === "Learning" && <span className="rz-src s-workload">preliminary</span>}
                  </td>
                  <td className="num">
                    <Delta from={w.current.monthlyCost} to={w.recommended.monthlyCost} money />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {selected.length > toApply.length && (
            <p className="rz-dim rz-xs">
              {selected.length - toApply.length} selected workloads are skipped: nothing to apply.
            </p>
          )}
        </Modal>
      )}
    </div>
  );
}
