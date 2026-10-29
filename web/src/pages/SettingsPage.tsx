import { CircleCheck, CircleX, Database, PiggyBank, ShieldCheck } from "lucide-react";
import { Fragment, useState } from "react";
import type { Info } from "../api";
import { CopyButton } from "../components";
import { ago } from "../format";
import { SettingsTable } from "./Details";

const STATE_TONE: Record<Info["cost"]["state"], string> = { ok: "ok", stale: "warn", fallback: "info", error: "bad" };

export function SettingsPage({ info }: { info: Info }) {
  const p = info.prometheus;
  const c = info.cost;
  return (
    <div className="rz-page">
      <div className="rz-pagehd">
        <div>
          <h1>Settings</h1>
          <p className="rz-dim">
            Cluster-wide defaults come from the Helm values. Namespaces and workloads override them with annotations.
          </p>
        </div>
      </div>
      <div className="rz-grid2">
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <Database size={16} /> Metrics source
            </h2>
            {p.connected ? (
              <span className="rz-badge t-ok">
                <CircleCheck size={13} /> Connected
              </span>
            ) : (
              <span className="rz-badge t-bad">
                <CircleX size={13} /> Unreachable
              </span>
            )}
          </div>
          <dl className="rz-dl">
            <dt>Prometheus API</dt>
            <dd className="rz-mono">{p.url}</dd>
            {p.error && (
              <>
                <dt>Error</dt>
                <dd className="t-bad">{p.error}</dd>
              </>
            )}
            {p.version && (
              <>
                <dt>Version</dt>
                <dd>{p.version}</dd>
              </>
            )}
            <dt>Retention</dt>
            <dd>
              {p.retention ?? "unknown"}
              {p.oldest && <span className="rz-dim"> · oldest sample {new Date(p.oldest).toLocaleString()}</span>}
            </dd>
            {(p.checks ?? []).map((c) => (
              <div key={c.metric} style={{ display: "contents" }}>
                <dt>{c.name}</dt>
                <dd>
                  {c.ok ? <CircleCheck size={13} className="t-accent" /> : <CircleX size={13} className="t-bad" />}{" "}
                  <code>{c.metric}</code>
                </dd>
              </div>
            ))}
          </dl>
          <p className="rz-foot rz-dim rz-xs">
            Set with the Helm value <code>prometheus.url</code>. Any Prometheus-compatible API works (Thanos, Mimir,
            VictoriaMetrics).
          </p>
        </section>
        <section className="rz-card">
          <div className="rz-cardhd">
            <h2>
              <PiggyBank size={16} /> Cost rates
            </h2>
          </div>
          <dl className="rz-dl">
            <dt>Source</dt>
            <dd>
              <span className={`rz-badge t-${STATE_TONE[c.state]}`}>
                {c.source === "karpenter" ? "Karpenter node prices" : "Configured rates"}
                {c.state !== "ok" && ` · ${c.state}`}
              </span>{" "}
              <span className="rz-dim">
                <code>cost.source: {c.mode}</code>
              </span>
            </dd>
            {c.reason && (
              <>
                <dt>{c.state === "fallback" ? "Why" : "Problem"}</dt>
                <dd className={c.state === "fallback" ? "rz-dim" : `t-${STATE_TONE[c.state]}`}>{c.reason}</dd>
              </>
            )}
            <dt>CPU per core-hour</dt>
            <dd>${c.cpuCoreHour.toFixed(4)}</dd>
            <dt>Memory per GiB-hour</dt>
            <dd>${c.memoryGiBHour.toFixed(4)}</dd>
            {c.source === "karpenter" && (
              <>
                <dt>Nodes</dt>
                <dd>
                  {c.nodes} ({c.spot} spot, {c.onDemand} on-demand)
                </dd>
                <dt>Node cost</dt>
                <dd>
                  ${(c.hourlyCost ?? 0).toFixed(2)}/h · ${((c.hourlyCost ?? 0) * 730).toFixed(0)}/mo
                  {!!c.onDemandHourlyCost && (
                    <span className="rz-dim"> (on-demand: ${(c.onDemandHourlyCost * 730).toFixed(0)}/mo)</span>
                  )}
                </dd>
                {c.calibratedAt && (
                  <>
                    <dt>Calibrated</dt>
                    <dd>
                      {ago(c.calibratedAt)}
                      {c.refresh && <span className="rz-dim"> · every {c.refresh}</span>}
                    </dd>
                  </>
                )}
              </>
            )}
          </dl>
          <p className="rz-foot rz-dim rz-xs">
            {c.source === "karpenter" ? (
              <>
                The configured rates (<code>cost.cpuCoreHour</code>, <code>cost.memoryGiBHour</code>) are scaled so the
                Karpenter nodes cost what Karpenter's price estimates say, spot or on-demand, refreshed every 10
                minutes. Node prices do not split CPU from memory, so the configured ratio is kept.
              </>
            ) : (
              <>
                Set with <code>cost.rates</code>. With Karpenter, <code>cost.source: auto</code> calibrates them to node
                prices.
              </>
            )}{" "}
            Monthly = hourly × 730, from requests × average replicas.
          </p>
          <div className="rz-cardhd">
            <h2>Scope</h2>
          </div>
          <dl className="rz-dl">
            <dt>Selection</dt>
            <dd>
              <code>{info.selection}</code>{" "}
              <span className="rz-dim rz-xs">
                {info.selection === "all"
                  ? "opt out with resourcerer.io/enabled=false"
                  : "opt in with resourcerer.io/enabled=true"}
              </span>
            </dd>
            <dt>Excluded namespaces</dt>
            <dd className="rz-chips">
              {info.excludedNamespaces.map((n) => (
                <code key={n}>{n}</code>
              ))}
            </dd>
            <dt>Authentication</dt>
            <dd>
              <code>{info.auth}</code> {info.readOnly && <span className="rz-dim rz-xs">read-only</span>}
            </dd>
          </dl>
        </section>
      </div>
      <section className="rz-card">
        <div className="rz-cardhd">
          <div>
            <h2>Cluster defaults</h2>
            <p className="rz-dim rz-xs">
              Change them in the Helm values under <code>defaults</code>, for example{" "}
              <code>defaults.learning-period: P14D</code>.
            </p>
          </div>
        </div>
        <SettingsTable values={info.defaults} showSource={false} />
      </section>
      <ReliabilitySignals cfg={info.reliability} />
    </div>
  );
}

const levelText = (l?: { warn?: number; critical?: number }, unit?: string) =>
  [l?.warn != null && `warn ${fmtLevel(l.warn, unit)}`, l?.critical != null && `critical ${fmtLevel(l.critical, unit)}`]
    .filter(Boolean)
    .join(" · ");

const fmtLevel = (v: number, unit?: string) => (unit === "ratio" ? `${+(v * 100).toFixed(2)}%` : String(v));

/** The reliability signals in use: built-in ones and those declared in the chart values, each test-run. */
function ReliabilitySignals({ cfg }: { cfg: Info["reliability"] }) {
  const [open, setOpen] = useState<string | null>(null);
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <div>
          <h2>
            <ShieldCheck size={16} /> Reliability signals
          </h2>
          <p className="rz-dim rz-xs">
            Set in the Helm values under <code>reliability</code>; see <code>docs/reliability-signals.md</code>. Turn
            one off for a namespace or workload with <code>resourcerer.io/reliability.disabled</code>.
          </p>
        </div>
      </div>
      {!!cfg.errors?.length && (
        <div className="rz-alert t-bad rz-sigerr">
          <CircleX size={16} />
          <div>
            {cfg.errors.map((e) => (
              <div key={e}>{e}</div>
            ))}
          </div>
        </div>
      )}
      <div className="rz-tablewrap">
        <table className="rz-table compact">
          <thead>
            <tr>
              <th>Signal</th>
              <th>Type</th>
              <th>Levels</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            {cfg.builtin.map((b) => (
              <tr key={b.id}>
                <td>
                  {b.name} <code className="rz-dim">{b.id}</code>
                </td>
                <td className="rz-dim">built-in</td>
                <td>{b.warn != null ? `warn ${Math.round(b.warn * 100)}%` : "—"}</td>
                <td>
                  {b.enabled ? (
                    <span className="rz-badge t-ok">
                      <CircleCheck size={13} /> On
                    </span>
                  ) : (
                    <span className="rz-badge">Off</span>
                  )}
                </td>
              </tr>
            ))}
            {cfg.signals.map((x) => (
              <Fragment key={x.id}>
                <tr className="rz-click" onClick={() => setOpen(open === x.id ? null : x.id)}>
                  <td>
                    {x.name} <code className="rz-dim">{x.id}</code>
                    {x.description && <div className="rz-dim rz-xs">{x.description}</div>}
                  </td>
                  <td className="rz-dim">declared · {x.unit}</td>
                  <td>
                    {x.above ? "above" : "below"}: {levelText(x.above ?? x.below, x.unit)}
                  </td>
                  <td>
                    {x.error ? (
                      <span className="rz-badge t-bad" title={x.error}>
                        <CircleX size={13} /> Query fails
                      </span>
                    ) : (
                      <span className="rz-badge t-ok">
                        <CircleCheck size={13} /> Valid
                      </span>
                    )}
                  </td>
                </tr>
                {open === x.id && (
                  <tr>
                    <td colSpan={4}>
                      {x.error && <p className="t-bad rz-xs">{x.error}</p>}
                      <div className="rz-query">
                        <pre>{x.query}</pre>
                        <CopyButton text={x.query} />
                      </div>
                      {x.runbook && (
                        <a className="rz-link small" href={x.runbook} target="_blank" rel="noreferrer">
                          Runbook
                        </a>
                      )}
                    </td>
                  </tr>
                )}
              </Fragment>
            ))}
          </tbody>
        </table>
      </div>
      {!cfg.signals.length && (
        <p className="rz-dim rz-xs rz-sigempty">
          No declared signals. Add PromQL templates under <code>reliability.signals</code> to flag what matters to your
          workloads, such as error rates or latency.
        </p>
      )}
    </section>
  );
}
