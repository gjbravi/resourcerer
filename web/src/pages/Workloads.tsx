import { useQueryClient } from "@tanstack/react-query";
import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ChevronLeft,
  ChevronRight,
  Flame,
  Search,
  ShieldCheck,
  TriangleAlert,
  Zap,
  ZapOff,
} from "lucide-react";
import { useState } from "react";
import { api, refKey, type Phase, type Resource, type Summary } from "../api";
import { Delta, Modal, PHASE_NAMES, Ring, StatusBadge, WlCell } from "../components";
import { fmt } from "../format";
import { navigate, toast, workloadHash } from "../hooks";
import { severity, signals } from "../reliability";

const net = (w: Summary) => w.current.monthlyCost - w.recommended.monthlyCost;

export const VIEWS: Record<string, [string, (w: Summary) => boolean]> = {
  all: ["All", () => true],
  savings: ["Savings", (w) => net(w) > 0.5],
  reliability: ["Reliability", (w) => signals(w).length > 0],
  under: ["Under-provisioned", (w) => w.under || w.noRequests],
  pending: ["To apply", (w) => applicable(w) && !w.autoDeploy],
  auto: ["Auto-deploy", (w) => w.autoDeploy],
};

export const atRisk = (w: Summary) => w.under || w.noRequests || signals(w).length > 0;

export const applicable = (w: Summary) => !!w.hash && w.pending && (w.phase === "Ready" || w.phase === "Learning");

const COLS: Record<string, (w: Summary) => number | string> = {
  name: (w) => refKey(w),
  status: (w) => w.phase,
  reliability: (w) => severity(w).rank,
  score: (w) => w.score?.overall ?? -1,
  cpu: (w) => w.current.cpu - w.recommended.cpu,
  memory: (w) => w.current.memory - w.recommended.memory,
  net,
};

/** Reliability signals of a workload at a glance, with the list in a tooltip. */
function RelBadge({ w }: { w: Summary }) {
  const s = severity(w);
  const [at, setAt] = useState<{ left: number; top?: number; bottom?: number } | null>(null);
  if (s.level === "healthy") return <span className="rz-rel healthy">healthy</span>;
  const Icon = s.level === "critical" ? Flame : s.level === "fixed" ? ShieldCheck : TriangleAlert;
  // Fixed to the viewport, so the table's scroll box does not clip it; above the badge near the bottom edge.
  const show = (e: React.SyntheticEvent<HTMLElement>) => {
    const r = e.currentTarget.getBoundingClientRect();
    setAt(
      r.bottom + 140 > window.innerHeight
        ? { left: r.left, bottom: window.innerHeight - r.top + 6 }
        : { left: r.left, top: r.bottom + 6 },
    );
  };
  return (
    <span
      className={`rz-rel ${s.level}`}
      tabIndex={0}
      onMouseEnter={show}
      onFocus={show}
      onMouseLeave={() => setAt(null)}
      onBlur={() => setAt(null)}
    >
      <Icon size={12} /> {s.label}
      <span className="rz-reltip" role="tooltip" style={at ? { display: "flex", ...at } : undefined}>
        {s.signals.map((x) => (
          <span key={x.title} className={x.fixed ? "fixed" : x.tone}>
            <i />
            {x.title}
            {x.fixed && <em> · fixed by the recommendation</em>}
          </span>
        ))}
      </span>
    </span>
  );
}

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

const PAGE_SIZES = [25, 50, 100];

/** Page numbers around the current one, with the first and last always shown; 0 marks a gap. */
export function pageList(page: number, pages: number): number[] {
  const out: number[] = [];
  for (let p = 1; p <= pages; p++) {
    if (p === 1 || p === pages || Math.abs(p - page) <= 1) out.push(p);
    else if (out[out.length - 1] !== 0) out.push(0);
  }
  return out;
}

function Pager({
  page,
  pages,
  size,
  total,
  onPage,
  onSize,
}: {
  page: number;
  pages: number;
  size: number;
  total: number;
  onPage: (p: number) => void;
  onSize: (n: number) => void;
}) {
  if (total <= PAGE_SIZES[0]) return null;
  return (
    <nav className="rz-pager" aria-label="Pages">
      <span className="rz-dim">
        {(page - 1) * size + 1}–{Math.min(page * size, total)} of {total}
      </span>
      <label className="rz-dim">
        Rows
        <select className="rz-select small" value={size} onChange={(e) => onSize(Number(e.target.value))}>
          {PAGE_SIZES.map((n) => (
            <option key={n}>{n}</option>
          ))}
        </select>
      </label>
      <span className="rz-grow" />
      <button aria-label="Previous page" disabled={page === 1} onClick={() => onPage(page - 1)}>
        <ChevronLeft size={15} />
      </button>
      {pageList(page, pages).map((p, i) =>
        p ? (
          <button key={p} aria-current={p === page ? "page" : undefined} onClick={() => onPage(p)}>
            {p}
          </button>
        ) : (
          <span key={`gap${i}`} className="rz-dim">
            …
          </span>
        ),
      )}
      <button aria-label="Next page" disabled={page === pages} onClick={() => onPage(page + 1)}>
        <ChevronRight size={15} />
      </button>
    </nav>
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
  const [page, setPage] = useState(1);
  const [size, setSize] = useState(() => Number(localStorage.getItem("resourcerer.pageSize")) || 25);
  // Any change to what is listed starts over at the first page.
  const reset =
    <T,>(set: (v: T) => void) =>
    (v: T) => {
      set(v);
      setPage(1);
    };
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
  const pages = Math.max(1, Math.ceil(shown.length / size));
  const at = Math.min(page, pages);
  const pageRows = shown.slice((at - 1) * size, at * size);
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
        onClick={() => {
          setSort((s) => ({ col, dir: s.col === col ? -s.dir : col === "name" ? 1 : -1 }));
          setPage(1);
        }}
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
              <button key={id} role="tab" aria-selected={view === id} onClick={() => reset(setView)(id)}>
                {label} <span className="n">{scoped.filter(f).length}</span>
              </button>
            ))}
          </div>
          <div className="rz-tools">
            <label className="rz-input">
              <Search size={14} />
              <input
                placeholder="Search"
                value={q}
                onChange={(e) => reset(setQ)(e.target.value)}
                aria-label="Search workloads"
              />
            </label>
            <select
              className="rz-select"
              value={ns}
              onChange={(e) => reset(setNs)(e.target.value)}
              aria-label="Namespace"
            >
              <option value="">Namespace</option>
              {namespaces.map((n) => (
                <option key={n}>{n}</option>
              ))}
            </select>
            <select
              className="rz-select"
              value={status}
              onChange={(e) => reset(setStatus)(e.target.value as Phase)}
              aria-label="Status"
            >
              <option value="">Status</option>
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
                {head("reliability", "Reliability")}
                {head("score", "Score")}
                {head("cpu", "CPU requests", true)}
                {head("memory", "Memory requests", true)}
                {head("net", "Net impact /mo", true)}
                <th>Auto</th>
              </tr>
            </thead>
            <tbody>
              {pageRows.map((w) => (
                <tr
                  key={refKey(w)}
                  className={`rz-click rel-${severity(w).level}`}
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
                  </td>
                  <td>
                    <RelBadge w={w} />
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
        <Pager
          page={at}
          pages={pages}
          size={size}
          total={shown.length}
          onPage={setPage}
          onSize={(n) => {
            localStorage.setItem("resourcerer.pageSize", String(n));
            reset(setSize)(n);
          }}
        />
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
