import { CircleCheck, CircleX, Database, PiggyBank } from "lucide-react";
import type { Info } from "../api";
import { SettingsTable } from "./Details";

export function SettingsPage({ info }: { info: Info }) {
  const p = info.prometheus;
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
            <dt>CPU per core-hour</dt>
            <dd>${info.cost.cpuCoreHour}</dd>
            <dt>Memory per GiB-hour</dt>
            <dd>${info.cost.memoryGiBHour}</dd>
          </dl>
          <p className="rz-foot rz-dim rz-xs">
            Estimates only; monthly = hourly × 730. Set with <code>cost.cpuCoreHour</code> and{" "}
            <code>cost.memoryGiBHour</code>.
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
    </div>
  );
}
