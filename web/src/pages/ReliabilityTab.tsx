import { Check as CheckIcon, Code, ExternalLink, Flame, RefreshCw, TriangleAlert } from "lucide-react";
import { useState, type ReactNode } from "react";
import type { Check, Detail, ReliabilityReport, ReliabilityWindow } from "../api";
import { TimeChart } from "../chart";
import { CopyButton, ErrorBox } from "../components";
import { fmtCpu, fmtMem } from "../format";
import { useNow, useReliability } from "../hooks";

const WINDOWS: [ReliabilityWindow, string][] = [
  ["learning", "Learning period"],
  ["1d", "1d"],
  ["6h", "6h"],
  ["1h", "1h"],
  ["15m", "15m"],
];

const key = (c: Check) => `${c.id}/${c.signal ?? ""}/${c.container ?? ""}`;
const pct = (v: number) => `${Math.round(v * 100)}%`;

/** A check's value in its unit; built-in charts are all ratios. */
function fmtCheck(c: Check, v: number): string {
  if (c.id !== "signal" || c.unit === "ratio") return pct(v);
  if (c.unit === "seconds") return v < 1 ? `${Math.round(v * 1000)}ms` : `${+v.toFixed(2)}s`;
  if (c.unit === "bytes") return fmtMem(v);
  return `${+v.toFixed(2)}`;
}

function StatusIcon({ s }: { s: Check["status"] }) {
  return (
    <span className={`rz-st ${s}`}>
      {s === "ok" ? <CheckIcon size={13} /> : s === "bad" ? <Flame size={13} /> : <TriangleAlert size={13} />}
    </span>
  );
}

function Gauge({ c }: { c: Check }) {
  if (c.unit !== "ratio" || !c.threshold || c.threshold > 1 || c.display === "no limit") return null;
  return (
    <div className="rz-gauge">
      <div className="track">
        <i className={c.status} style={{ width: `${Math.min(100, c.value * 100)}%` }} />
        <s style={{ left: `${c.threshold * 100}%` }} />
      </div>
      <div className="lbl">
        <span>0%</span>
        <span>threshold {pct(c.threshold)}</span>
        <span>{c.id === "memory" ? "limit" : "100%"}</span>
      </div>
    </div>
  );
}

/** A value change the recommendation already makes, or null when it leaves the value alone. */
function Change({ label, from, to, fmt }: { label: string; from?: number; to?: number; fmt: (v?: number) => string }) {
  if (to == null || (from != null && to <= from)) return null;
  return (
    <div className="fix">
      {label}{" "}
      <span className="chg">
        <del>{from == null ? "none" : fmt(from)}</del> → {fmt(to)}
      </span>
      <span className="rz-tag ok">in the recommendation</span>
    </div>
  );
}

function Adjust({ c, d }: { c: Check; d: Detail }) {
  const ct = d.containers?.find((x) => x.name === c.container);
  const logs = c.pod && (
    <div className="fix">
      <code>
        kubectl logs -n {d.namespace} {c.pod}
        {c.container && !c.container.includes(",") ? ` -c ${c.container}` : ""} --previous
      </code>
      <CopyButton
        text={`kubectl logs -n ${d.namespace} ${c.pod}${c.container && !c.container.includes(",") ? ` -c ${c.container}` : ""} --previous`}
      />
    </div>
  );
  const items: Record<Check["id"], ReactNode> = {
    throttling: (
      <>
        {Change({ label: "CPU limit", from: ct?.currentLimits.cpu, to: ct?.recommendedLimits.cpu, fmt: fmtCpu }) ?? (
          <span>Raise the CPU limit above the bursts.</span>
        )}
        <span className="rz-dim">
          The kernel enforces the limit every 100ms, so bursts get paused even when average usage is lower. You can also
          stop setting CPU limits with <code>cpu.limits-policy: RequestsOnly</code> in Settings.
        </span>
      </>
    ),
    memory: (
      <>
        {Change({
          label: "Memory limit",
          from: ct?.currentLimits.memory,
          to: ct?.recommendedLimits.memory,
          fmt: fmtMem,
        }) ?? <span>Raise the memory limit well above the peak.</span>}
        <Change
          label="Memory request"
          from={ct?.currentRequests.memory}
          to={ct?.recommendedRequests.memory}
          fmt={fmtMem}
        />
      </>
    ),
    oom: (
      <>
        <span>The memory recommendation is raised above the limit that killed it.</span>
        {d.containers
          ?.filter((x) => !x.skipped)
          .map((x) => (
            <Change
              key={x.name}
              label={`Memory limit · ${x.name}`}
              from={x.currentLimits.memory}
              to={x.recommendedLimits.memory}
              fmt={fmtMem}
            />
          ))}
      </>
    ),
    restarts: (
      <>
        <span>Find out why the container exits; the last exit code is in the detail above.</span>
        {logs}
      </>
    ),
    crashloop: (
      <>
        <span>Fix the crash first: new requests or limits do not stop a container that exits on its own.</span>
        {logs}
      </>
    ),
    hpa: (
      <span>
        Raise <code>maxReplicas</code> on HPA {c.container}, or lower the load per replica. Lowering CPU requests raises
        its utilization and makes it scale out sooner.
      </span>
    ),
    signal: (
      <>
        <span>{c.detail || "A signal declared in the reliability configuration."}</span>
        {c.runbook && (
          <a className="rz-link small" href={c.runbook} target="_blank" rel="noreferrer">
            <ExternalLink size={13} /> Runbook
          </a>
        )}
        <span className="rz-dim">
          Not changed by the recommendation. Levels can be tuned per namespace or workload with{" "}
          <code>resourcerer.io/reliability.{c.signal}.warn</code> and <code>.critical</code> annotations.
        </span>
      </>
    ),
    qos: d.containers?.some((x) => !x.skipped && Object.keys(x.recommendedRequests).length) ? (
      <div className="fix">
        Requests for every container <span className="rz-tag ok">in the recommendation</span>
      </div>
    ) : (
      <span>Set CPU and memory requests on every container.</span>
    ),
  };
  return items[c.id];
}

function Measured({ c }: { c: Check }) {
  const [open, setOpen] = useState(false);
  return (
    <>
      <span className="rz-dim">{c.source}</span>
      {c.query && (
        <>
          <button className="rz-link small" onClick={() => setOpen(!open)}>
            <Code size={13} /> {open ? "Hide" : "Show"} exact query
          </button>
          {open && (
            <div className="rz-query">
              <pre>{c.query}</pre>
              <CopyButton text={c.query} />
            </div>
          )}
        </>
      )}
    </>
  );
}

function CheckCard({ c, d, rep, before }: { c: Check; d: Detail; rep: ReliabilityReport; before?: Check }) {
  return (
    <div className={`rz-check ${c.status}`}>
      <div className="hd">
        <StatusIcon s={c.status} />
        <div className="nm">
          <b>
            {c.name}
            {c.container && ` · ${c.container}`}
          </b>
          <span>{c.detail}</span>
        </div>
        <Gauge c={c} />
        <div className={`val t-${c.status}`}>
          {c.display}
          {before && before.display !== c.display && <small>learning period: {before.display}</small>}
        </div>
      </div>
      <div className="body">
        <div className="box">
          <h5>What to adjust</h5>
          <Adjust c={c} d={d} />
        </div>
        <div className="box">
          <h5>How it is measured</h5>
          <Measured c={c} />
        </div>
      </div>
      {!!c.series?.length && (
        <div className="chart">
          <TimeChart
            title={
              c.id === "signal" ? c.name : c.id === "memory" ? "Memory working set / limit" : "Throttled CFS periods"
            }
            unit={c.id === "signal" ? c.unit : c.id === "memory" ? "% of limit" : "% of periods"}
            maxY={c.id === "signal" ? undefined : 1}
            height={170}
            series={{ from: rep.from, to: rep.to, bands: [], lines: {} }}
            lines={[
              {
                name: c.id === "signal" ? c.name : c.id === "memory" ? "Peak / limit" : "Throttled",
                color: c.status === "ok" ? "var(--c-usage)" : `var(--${c.status})`,
                points: c.series,
                shape: "smooth",
              },
            ]}
            thresholds={c.threshold ? [{ v: c.threshold, label: `threshold ${fmtCheck(c, c.threshold)}` }] : []}
            fmtY={(v) => fmtCheck(c, v)}
          />
        </div>
      )}
    </div>
  );
}

export function ReliabilityTab({ d }: { d: Detail }) {
  const [window, setWindow] = useState<ReliabilityWindow>("learning");
  const [live, setLive] = useState(false);
  const learning = useReliability(d, "learning", false);
  const current = useReliability(d, window, live);
  const now = useNow();
  const rep = current.data;
  const base = learning.data;
  const comparing = window !== "learning" && base && rep;
  const was = new Map(base?.checks.map((c) => [key(c), c]));

  const failing = rep?.checks.filter((c) => c.status !== "ok") ?? [];
  const resolved = comparing
    ? rep.checks.filter((c) => c.status === "ok" && (was.get(key(c))?.status ?? "ok") !== "ok")
    : [];
  const passed = rep?.checks.filter((c) => c.status === "ok" && !resolved.includes(c)) ?? [];
  const fixable = failing.filter((c) => {
    const ct = d.containers?.find((x) => x.name === c.container);
    if (c.id === "throttling") return (ct?.recommendedLimits.cpu ?? 0) > (ct?.currentLimits.cpu ?? Infinity);
    if (c.id === "memory") return (ct?.recommendedLimits.memory ?? 0) > (ct?.currentLimits.memory ?? Infinity);
    return c.id === "oom";
  });
  const status = failing.some((c) => c.status === "bad") ? "bad" : failing.length ? "warn" : "ok";
  const ago = current.dataUpdatedAt ? Math.max(0, Math.round((now - current.dataUpdatedAt) / 1000)) : null;
  const span = rep ? rep.to - rep.from : 0;

  return (
    <>
      <div className="rz-sechd">
        <h2>Reliability</h2>
        <div className="rz-ctl">
          <span className="rz-seg">
            {WINDOWS.map(([w, label]) => (
              <button key={w} aria-pressed={window === w} onClick={() => setWindow(w)}>
                {label}
              </button>
            ))}
          </span>
          <button
            className={`rz-live${live ? " on" : ""}`}
            aria-pressed={live}
            onClick={() => {
              setLive(!live);
              if (!live && window === "learning") setWindow("15m");
            }}
          >
            <i /> Live
          </button>
          <button
            className="rz-btn ghost small"
            onClick={() => {
              void current.refetch();
              void learning.refetch();
            }}
            disabled={current.isFetching}
          >
            <RefreshCw size={14} className={current.isFetching ? "rz-spin" : ""} /> Refresh
          </button>
        </div>
      </div>
      <ErrorBox error={current.error} />
      {rep && (
        <>
          <div className="rz-relsum">
            <div>
              <span>Status</span>
              <b className={`t-${status}`}>{status === "ok" ? "Healthy" : status === "bad" ? "At risk" : "Warning"}</b>
            </div>
            <div>
              <span>Checks passed</span>
              <b>
                {rep.checks.length - failing.length} of {rep.checks.length}
              </b>
            </div>
            <div>
              <span>Window</span>
              <b>{window === "learning" ? `Learning period · ${fmtSpan(span)}` : `Last ${window}`}</b>
              <span>
                {ago == null ? "" : `updated ${ago}s ago`}
                {live && " · refreshes every 30s"}
              </span>
            </div>
            <div>
              <span>Fixed by the recommendation</span>
              <b>{failing.length ? `${fixable.length} of ${failing.length}` : "—"}</b>
            </div>
          </div>
          {comparing && (
            <p className="rz-dim rz-xs rz-relnote">
              Comparing the last {window} with the learning period ({fmtSpan(base.to - base.from)}), which the
              recommendation and the alert above are based on. A fix shows up here first; the learning period keeps the
              old samples until they age out.
            </p>
          )}
          {failing.map((c) => (
            <CheckCard key={key(c)} c={c} d={d} rep={rep} before={comparing ? was.get(key(c)) : undefined} />
          ))}
          {resolved.length > 0 && (
            <div className="rz-passed resolved">
              <h5>Resolved in the last {window}</h5>
              {resolved.map((c) => (
                <div key={key(c)}>
                  <StatusIcon s="ok" />
                  <span>
                    {c.name}
                    {c.container && ` · ${c.container}`}
                  </span>
                  <span className="m">
                    <span className={`t-${was.get(key(c))?.status}`}>{was.get(key(c))?.display}</span> → {c.display}
                  </span>
                </div>
              ))}
            </div>
          )}
          {passed.length > 0 && (
            <div className="rz-passed">
              <h5>Passed</h5>
              {passed.map((c) => (
                <div key={key(c)} title={c.source}>
                  <StatusIcon s="ok" />
                  <span>
                    {c.name}
                    {c.container && ` · ${c.container}`}
                  </span>
                  <span className="m">{c.display}</span>
                </div>
              ))}
            </div>
          )}
        </>
      )}
      {!rep && current.isLoading && <div className="rz-loading">Checking…</div>}
    </>
  );
}

function fmtSpan(s: number) {
  if (s >= 86400) return `${Math.round(s / 86400)}d`;
  if (s >= 3600) return `${Math.round(s / 3600)}h`;
  return `${Math.max(1, Math.round(s / 60))}m`;
}
