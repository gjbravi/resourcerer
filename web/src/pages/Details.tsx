import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  Boxes,
  ChevronLeft,
  Cpu,
  Hourglass,
  Info,
  MemoryStick,
  Pencil,
  RefreshCw,
  RotateCcw,
  ShieldCheck,
  Sparkles,
  TriangleAlert,
  Zap,
  ZapOff,
} from "lucide-react";
import { useState } from "react";
import { api, type Detail, type Range, type Ref, type Resource, type SettingInfo } from "../api";
import { Art, NoteArt, WandArt } from "../art";
import { TimeChart, type ChartEvent } from "../chart";
import {
  Change,
  CopyButton,
  Delta,
  ErrorBox,
  FitBar,
  Metric,
  Modal,
  Money,
  SourceTag,
  StatusBadge,
  phaseTone,
} from "../components";
import { ago, fmt, fmtCores, fmtDate, fmtMem, fmtMoney, Gi, Mi, quantity } from "../format";
import { toast, useAction, useImpact, useSeries, useWorkload, workloadHash } from "../hooks";
import { meta } from "../settings";

const TABS = [
  ["impact", "Impact"],
  ["recommendation", "Recommendation"],
  ["patches", "Patches"],
  ["settings", "Settings"],
] as const;
const RANGES: Range[] = ["1h", "1d", "1w", "2w", "1m"];
const fmtCpuAxis = (v: number) => `${Math.round(v * 1000)}m`;
const fmtMemAxis = (v: number) => (v >= Gi ? `${(v / Gi).toFixed(1)}GiB` : `${Math.round(v / Mi)}MiB`);

export function Details({ wref, tab, readOnly }: { wref: Ref; tab: string; readOnly: boolean }) {
  const { data: d, error, isLoading } = useWorkload(wref);
  const [confirm, setConfirm] = useState(false);
  const apply = useAction(wref, (hash: string) => api.apply(wref, hash));
  const revert = useAction(wref, () => api.revert(wref));
  const regenerate = useAction(wref, () => api.regenerate(wref));

  if (isLoading) return <div className="rz-loading">Loading…</div>;
  if (error || !d)
    return (
      <div className="rz-page">
        <a className="rz-back" href="#/workloads">
          <ChevronLeft size={15} /> Workloads
        </a>
        <ErrorBox error={error ?? "Workload not found"} />
      </div>
    );

  const containers = d.containers ?? [];
  const canApply = !readOnly && !!d.hash && d.pending && (d.phase === "Ready" || d.phase === "Learning");
  const lastApply = d.history.find((h) => h.action === "Apply");
  const net = d.current.monthlyCost - d.recommended.monthlyCost;
  const doRevert = () =>
    revert.mutate(undefined, {
      onSuccess: () =>
        toast({
          kind: "info",
          title: `Revert requested for ${d.name}`,
          body: "The applier restores the previous requests and limits.",
        }),
      onError: (e) => toast({ kind: "error", title: "Revert failed", body: String(e) }),
    });

  return (
    <div className="rz-page">
      <a className="rz-back" href="#/workloads">
        <ChevronLeft size={15} /> Workloads
      </a>
      <div className="rz-titlerow">
        <h1>{d.name}</h1>
        <span className="rz-dim">
          {d.namespace} · {d.kind}
        </span>
        <span className="rz-grow" />
        <div className="rz-tabs" role="tablist">
          {TABS.map(([id, label]) => (
            <a key={id} role="tab" aria-selected={tab === id} href={workloadHash(d, id)}>
              {label}
            </a>
          ))}
        </div>
      </div>

      <section className="rz-metrics">
        <Metric
          label={
            <>
              {net >= 0 ? "Estimated savings" : "Extra capacity needed"}{" "}
              <Info size={15} aria-label="Requests × average replicas, priced with the cost rates in Settings" />
            </>
          }
          art={d.oomKills > 0 || net < -0.005 ? <Art kind="shield" size={128} /> : <NoteArt />}
          trend={
            <>
              <Delta from={d.current.monthlyCost} to={d.recommended.monthlyCost} money />{" "}
              <span className="rz-dim">
                {fmtMoney(d.current.monthlyCost)} → {fmtMoney(d.recommended.monthlyCost)} per month
              </span>
            </>
          }
        >
          <Money value={Math.abs(net)} tone={net >= 0 ? "save" : "cost"} />
        </Metric>
        <Metric
          label="Fit score"
          art={<WandArt />}
          trend={
            <span className="rz-dim">
              {d.score ? `CPU ${d.score.cpu}% · Memory ${d.score.memory}%` : "not scored yet"}
            </span>
          }
        >
          <div className="rz-money">
            {d.score ? d.score.overall : "–"}
            <span className="dec">/100</span>
          </div>
          <FitBar value={d.score?.overall} />
        </Metric>
      </section>

      <div className="rz-actionsrow">
        <div className="rz-recactions" style={{ marginTop: 0 }}>
          <button
            className="rz-btn primary"
            disabled={!canApply}
            onClick={() => setConfirm(true)}
            title={
              readOnly ? "Read-only" : !d.pending ? "Current requests already match the recommendation" : d.message
            }
          >
            <Sparkles size={15} /> Apply recommendation
          </button>
          <button
            className="rz-btn ghost"
            disabled={readOnly || d.phase === "Unsupported" || regenerate.isPending}
            onClick={() =>
              regenerate.mutate(undefined, {
                onSuccess: () =>
                  toast({
                    kind: "info",
                    title: "Regeneration requested",
                    body: "The controller recomputes from the latest history.",
                  }),
              })
            }
          >
            <RefreshCw size={14} /> Regenerate
          </button>
        </div>
        <div className="rz-recmeta" style={{ marginTop: 0 }}>
          <StatusBadge phase={d.phase} title={d.message} />
          <span>
            Generated <b>{ago(d.generatedAt)}</b>
          </span>
          <span>
            Next <b>{ago(d.nextScheduleAt)}</b>
          </span>
          <span>
            {d.autoDeploy ? <Zap size={12} className="t-accent" /> : <ZapOff size={12} />} Auto-deploy{" "}
            <b>{d.autoDeploy ? "on" : "off"}</b>
          </span>
          {lastApply && (
            <span>
              Applied <b>{ago(lastApply.time)}</b> by {lastApply.by}
            </span>
          )}
        </div>
      </div>

      <section className="rz-card rz-kpis">
        <div className="rz-kpi">
          <span>
            <Cpu size={15} /> CPU requested
          </span>
          <b>
            {fmtCores(d.recommended.cpu)} <Delta from={d.current.cpu} to={d.recommended.cpu} r="cpu" />
          </b>
          <span className="rz-dim rz-xs">from {fmtCores(d.current.cpu)}</span>
        </div>
        <div className="rz-kpi">
          <span>
            <MemoryStick size={15} /> Memory requested
          </span>
          <b>
            {fmtMem(d.recommended.memory)} <Delta from={d.current.memory} to={d.recommended.memory} r="memory" />
          </b>
          <span className="rz-dim rz-xs">from {fmtMem(d.current.memory)}</span>
        </div>
        <div className="rz-kpi">
          <span>
            <ShieldCheck size={15} /> Reliability
          </span>
          <b className={d.oomKills ? "t-bad" : ""}>{d.oomKills ? `${d.oomKills} OOM kills` : "Healthy"}</b>
          <span className="rz-dim rz-xs">
            {d.oomKills ? "memory is raised above the limit that killed it" : "no OOM kills in the learning period"}
          </span>
        </div>
        <div className="rz-kpi">
          <span>
            <Boxes size={15} /> Replicas
          </span>
          <b>
            {d.replicas}
            <span className="rz-dim rz-xs">now · {d.avgReplicas.toFixed(1)} avg</span>
          </b>
          <span className="rz-dim rz-xs">
            {d.hpaRecommendation ? `HPA ${d.hpaRecommendation.name} at ${d.hpaRecommendation.current}% CPU` : "no HPA"}
          </span>
        </div>
      </section>

      {d.phase !== "Ready" && d.message && (
        <div className={`rz-alert t-${phaseTone(d.phase)}`}>
          <Info size={16} />
          {d.message}
        </div>
      )}
      {d.noRequests && (
        <div className="rz-alert t-warn">
          <TriangleAlert size={16} />
          No requests are set, so the scheduler treats these pods as best-effort and evicts them first. The
          recommendation is a good starting point.
        </div>
      )}
      {d.lastApplyResult && d.lastApplyResult.result !== "Succeeded" && (
        <div className="rz-alert t-bad">
          <TriangleAlert size={16} />
          Last apply {d.lastApplyResult.result.toLowerCase()} {ago(d.lastApplyResult.time)}: {d.lastApplyResult.message}
        </div>
      )}
      {d.warnings?.map((w) => (
        <div key={w} className="rz-alert t-warn">
          <TriangleAlert size={16} />
          {w}
        </div>
      ))}

      {tab === "impact" && <ImpactTab d={d} />}
      {tab === "recommendation" && <RecommendationTab d={d} />}
      {tab === "patches" && <PatchesTab d={d} />}
      {tab === "settings" && <SettingsTab d={d} readOnly={readOnly} />}

      <section className="rz-card">
        <div className="rz-cardhd">
          <h2>Events</h2>
        </div>
        <ul className="rz-events">
          {d.events.slice(0, 10).map((e, i) => (
            <li key={i}>
              <span
                className={`dot ${e.type === "Warning" ? "bad" : e.reason.startsWith("Recommendation") ? "acc" : ""}`}
              />
              <div>
                <b className={e.type === "Warning" ? "t-bad" : ""}>{e.reason}</b>{" "}
                <span className="rz-dim rz-xs">{ago(e.time)}</span>
                <div className="rz-dim rz-xs">{e.message}</div>
              </div>
              {i === d.events.findIndex((x) => x.reason === "RecommendationApplied") && lastApply && !readOnly && (
                <button className="rz-btn ghost small" onClick={doRevert} disabled={revert.isPending}>
                  <RotateCcw size={13} /> Revert
                </button>
              )}
            </li>
          ))}
          {!d.events.length && <li className="rz-dim">No recent events. Kubernetes keeps events for about an hour.</li>}
        </ul>
      </section>

      {confirm && (
        <Modal
          title={`Apply recommendation to ${d.name}?`}
          onClose={() => setConfirm(false)}
          actions={
            <>
              <button className="rz-btn ghost" onClick={() => setConfirm(false)}>
                Cancel
              </button>
              <button
                className="rz-btn primary"
                disabled={apply.isPending}
                onClick={() =>
                  apply.mutate(d.hash!, {
                    onSuccess: () => {
                      setConfirm(false);
                      toast({
                        kind: "ok",
                        title: `Apply requested for ${d.name}`,
                        body: "The applier patches the workload and a rolling update starts.",
                        action: { label: "Undo", run: doRevert },
                      });
                    },
                    onError: (e) =>
                      toast({ kind: "error", title: "Apply failed", body: e instanceof Error ? e.message : String(e) }),
                  })
                }
              >
                Apply
              </button>
            </>
          }
        >
          <p>{applyDescription(d)}</p>
          <table className="rz-table compact">
            <thead>
              <tr>
                <th>Container</th>
                <th>CPU req</th>
                <th>CPU lim</th>
                <th>Mem req</th>
                <th>Mem lim</th>
              </tr>
            </thead>
            <tbody>
              {containers
                .filter((c) => !c.skipped)
                .map((c) => (
                  <tr key={c.name}>
                    <td>
                      <b>{c.name}</b>
                    </td>
                    {(["cpu", "memory"] as const).flatMap((r) => [
                      <td key={r}>
                        <Change r={r} a={c.currentRequests[r]} b={c.recommendedRequests[r]} />
                      </td>,
                      <td key={`${r}l`}>
                        <Change r={r} a={c.currentLimits[r]} b={c.recommendedLimits[r]} />
                      </td>,
                    ])}
                  </tr>
                ))}
            </tbody>
          </table>
          {containers.some(
            (c) => c.currentRequests.memory && c.recommendedRequests.memory! < c.currentRequests.memory * 0.5,
          ) && (
            <p className="rz-alert t-warn small">
              <TriangleAlert size={14} />
              Memory drops by more than half for at least one container. Check that the learning period covers peak
              traffic.
            </p>
          )}
          {d.hpaRecommendation && (
            <p className="rz-xs">
              HPA {d.hpaRecommendation.name} target utilization changes {d.hpaRecommendation.current}% →{" "}
              {d.hpaRecommendation.recommended}% to keep the scale point.
            </p>
          )}
          {d.phase === "Learning" && (
            <p className="rz-alert t-info small">
              <Hourglass size={14} />
              This recommendation is preliminary: the learning period is not covered yet.
            </p>
          )}
        </Modal>
      )}
    </div>
  );
}

/** Applies and reverts from the recommendation's history, as chart markers. */
const historyEvents = (d: Detail): ChartEvent[] =>
  d.history.map((h) => ({ t: Date.parse(h.time) / 1000, label: h.action === "Apply" ? "Applied" : "Reverted" }));

/** Where an apply lands and what it triggers, in the words of the workload kind. */
export function applyDescription(d: Pick<Detail, "kind" | "name" | "templateKind" | "templateName">): string {
  if (d.kind === "Rollout") {
    const where = d.templateName
      ? `the ${d.templateKind} ${d.templateName} it references (spec.workloadRef)`
      : "the Rollout's pod template";
    return `Server-side apply on ${where} with field manager resourcerer. Argo Rollouts then starts a progressive rollout following the Rollout's strategy (canary or blue-green), not an immediate restart.`;
  }
  return `Server-side apply on the ${d.kind} with field manager resourcerer, followed by a rolling update.`;
}

export function Zoom({ value, onChange }: { value: Range; onChange: (r: Range) => void }) {
  return (
    <div className="rz-zoom">
      <span className="rz-dim rz-xs">Zoom:</span>
      {RANGES.map((z) => (
        <button key={z} aria-pressed={value === z} onClick={() => onChange(z)}>
          {z.toUpperCase()}
        </button>
      ))}
    </div>
  );
}

const Arrow = ({ a, b }: { a?: number; b?: number }) =>
  a == null || b == null || Math.abs(b - a) / (a || 1) < 0.02 ? null : b < a ? (
    <ArrowDown size={13} className="t-save" />
  ) : (
    <ArrowUp size={13} className="t-bad" />
  );

function ContainerChart({
  d,
  container,
  resource,
  range,
}: {
  d: Detail;
  container: string;
  resource: Resource;
  range: Range;
}) {
  const { data, isLoading, error } = useSeries(d, container, resource, range);
  const l = data?.lines ?? {};
  return (
    <>
      <ErrorBox error={error} />
      <TimeChart
        title={
          <>
            {resource === "cpu" ? "CPU" : "Memory"} · <span className="t-accent">{container}</span>
          </>
        }
        unit={resource === "cpu" ? "Millicores" : "MiB"}
        series={data}
        loading={isLoading}
        events={historyEvents(d)}
        fmtY={resource === "cpu" ? fmtCpuAxis : fmtMemAxis}
        lines={[
          { name: "Current requests", color: "var(--c-cur)", points: l.currentRequests ?? [] },
          { name: "Recommended requests", color: "var(--c-rec)", points: l.recommendedRequests ?? [] },
          { name: "Current limits", color: "var(--c-curlim)", points: l.currentLimits ?? [], dash: "6 4" },
          { name: "Recommended limits", color: "var(--c-reclim)", points: l.recommendedLimits ?? [], dash: "2 3" },
        ]}
      />
    </>
  );
}

function RecommendationTab({ d }: { d: Detail }) {
  const containers = d.containers ?? [];
  const [ci, setCi] = useState(() =>
    Math.max(
      0,
      containers.findIndex((c) => !c.skipped),
    ),
  );
  const [range, setRange] = useState<Range>("1h");
  const sel = containers[ci];
  return (
    <>
      <div className="rz-sechd">
        <h2>Containers</h2>
        <Zoom value={range} onChange={setRange} />
      </div>
      <div className="rz-recgrid">
        <div className="rz-conts">
          {containers.map((c, i) => (
            <div key={c.name} className={`rz-cont${i === ci ? " on" : ""}`}>
              <div className="rz-conthd">
                <b>{c.name}</b>
                <button className={`rz-btn small ${i === ci ? "primary" : "ghost"}`} onClick={() => setCi(i)}>
                  {i === ci ? "Shown" : "Show"} <ArrowRight size={13} />
                </button>
              </div>
              {c.skipped ? (
                <p className="rz-dim rz-xs">
                  {c.basis?.cpu ?? "Not optimized"}. Keeps {fmt("cpu", c.currentRequests.cpu)} /{" "}
                  {fmt("memory", c.currentRequests.memory)}.
                </p>
              ) : (
                (["cpu", "memory"] as const).map((r) => (
                  <div key={r} className="rz-contres">
                    <div className="rz-contrh">
                      {r === "cpu" ? "CPU" : "Memory"} <span className="rz-dim rz-xs">{c.basis?.[r] ?? ""}</span>
                    </div>
                    <div className="rz-controw">
                      <span className="rz-dim">Req:</span>
                      <span>{c.currentRequests[r] == null ? "none" : fmt(r, c.currentRequests[r])}</span>
                      <ArrowRight size={13} className="rz-dim" />
                      <b>{c.recommendedRequests[r] == null ? "none" : fmt(r, c.recommendedRequests[r])}</b>
                      <Arrow a={c.currentRequests[r]} b={c.recommendedRequests[r]} />
                    </div>
                    <div className="rz-controw">
                      <span className="rz-dim">Lim:</span>
                      <span>{c.currentLimits[r] == null ? "none" : fmt(r, c.currentLimits[r])}</span>
                      <ArrowRight size={13} className="rz-dim" />
                      <b>{c.recommendedLimits[r] == null ? "none" : fmt(r, c.recommendedLimits[r])}</b>
                      <Arrow a={c.currentLimits[r]} b={c.recommendedLimits[r]} />
                    </div>
                  </div>
                ))
              )}
            </div>
          ))}
          {d.hpaRecommendation && (
            <div className="rz-cont">
              <div className="rz-conthd">
                <b>HPA {d.hpaRecommendation.name}</b>
              </div>
              <div className="rz-controw">
                <span className="rz-dim">CPU:</span>
                <span>{d.hpaRecommendation.current}%</span>
                <ArrowRight size={13} className="rz-dim" />
                <b>{d.hpaRecommendation.recommended}%</b>
              </div>
              <p className="rz-dim rz-xs">
                Keeps scaling at the same absolute CPU per pod after requests change
                {d.hpaRecommendation.clamped ? " (clamped to 50–90%)" : ""}.
              </p>
            </div>
          )}
          {!containers.length && <p className="rz-empty">No recommendation yet.</p>}
        </div>
        <div className="rz-charts">
          {sel &&
            (sel.skipped ? (
              <p className="rz-empty">{sel.name} is not optimized.</p>
            ) : (
              (["cpu", "memory"] as const).map((r) => (
                <ContainerChart key={`${sel.name}-${r}`} d={d} container={sel.name} resource={r} range={range} />
              ))
            ))}
        </div>
      </div>
    </>
  );
}

function ImpactChart({ d, resource, range }: { d: Detail; resource: Resource; range: Range }) {
  const { data, isLoading, error } = useImpact(d, resource, range);
  return (
    <>
      <ErrorBox error={error} />
      <TimeChart
        title={`Total ${resource === "cpu" ? "CPU" : "memory"} across replicas`}
        unit={resource === "cpu" ? "Cores" : "GiB"}
        height={240}
        series={data}
        loading={isLoading}
        events={historyEvents(d)}
        bandsLabel="Usage"
        fmtY={resource === "cpu" ? (v) => v.toFixed(v < 1 ? 2 : 1) : (v) => (v / Gi).toFixed(2)}
        lines={[
          { name: "Requested", color: "var(--c-cur)", points: data?.lines.requested ?? [] },
          {
            name: "With recommendation",
            color: "var(--c-rec)",
            points: data?.lines.withRecommendation ?? [],
            fill: "Requested",
          },
        ]}
      />
    </>
  );
}

function ImpactTab({ d }: { d: Detail }) {
  const [range, setRange] = useState<Range>("1h");
  return (
    <>
      <div className="rz-sechd">
        <h2>Projected impact</h2>
        <Zoom value={range} onChange={setRange} />
      </div>
      <div className="rz-kpis">
        <div className="rz-kpi">
          <span>Monthly cost</span>
          <b>
            ${d.current.monthlyCost.toFixed(2)} → ${d.recommended.monthlyCost.toFixed(2)}
          </b>
        </div>
        <div className="rz-kpi">
          <span>CPU requested</span>
          <b>
            {fmtCores(d.current.cpu)} → {fmtCores(d.recommended.cpu)}
          </b>
        </div>
        <div className="rz-kpi">
          <span>Memory requested</span>
          <b>
            {fmtMem(d.current.memory)} → {fmtMem(d.recommended.memory)}
          </b>
        </div>
        <div className="rz-kpi">
          <span>Reliability</span>
          <b className={d.oomKills ? "t-bad" : ""}>{d.oomKills ? `${d.oomKills} OOM kills` : "No OOM kills"}</b>
          <span className="rz-dim rz-xs">
            {d.oomKills
              ? "The memory recommendation is raised above the limit that killed it."
              : "Usage stayed under limits."}
          </span>
        </div>
      </div>
      <div className="rz-card pad">
        <ImpactChart d={d} resource="cpu" range={range} />
        <ImpactChart d={d} resource="memory" range={range} />
      </div>
    </>
  );
}

function PatchesTab({ d }: { d: Detail }) {
  const [format, setFormat] = useState<"kustomize" | "helm" | "kubectl">("kustomize");
  const recs = (d.containers ?? []).filter((c) => !c.skipped);
  if (!d.hash || !recs.length)
    return (
      <div className="rz-card">
        <p className="rz-empty">No recommendation to export yet.</p>
      </div>
    );
  const resources = (c: (typeof recs)[number]) => {
    const req = Object.fromEntries(
      (["cpu", "memory"] as const).map((r) => [r, quantity(r, c.recommendedRequests[r])]).filter(([, v]) => v),
    );
    const lim = Object.fromEntries(
      (["cpu", "memory"] as const).map((r) => [r, quantity(r, c.recommendedLimits[r])]).filter(([, v]) => v),
    );
    return { requests: req, ...(Object.keys(lim).length ? { limits: lim } : {}) };
  };
  const yaml = (c: (typeof recs)[number], ind: string) => {
    const r = resources(c);
    const lines = [
      `${ind}resources:`,
      `${ind}  requests:`,
      ...Object.entries(r.requests).map(([k, v]) => `${ind}    ${k}: ${v}`),
    ];
    if (r.limits) lines.push(`${ind}  limits:`, ...Object.entries(r.limits).map(([k, v]) => `${ind}    ${k}: ${v}`));
    return lines.join("\n");
  };
  // A Rollout with spec.workloadRef keeps its template in a Deployment; that is what to patch.
  const target = d.templateName
    ? { apiVersion: "apps/v1", kind: d.templateKind ?? "Deployment", name: d.templateName }
    : { apiVersion: d.kind === "Rollout" ? "argoproj.io/v1alpha1" : "apps/v1", kind: d.kind, name: d.name };
  const text = {
    kustomize: [
      `apiVersion: ${target.apiVersion}`,
      `kind: ${target.kind}`,
      "metadata:",
      `  name: ${target.name}`,
      `  namespace: ${d.namespace}`,
      "spec:",
      "  template:",
      "    spec:",
      "      containers:",
      ...recs.map((c) => `        - name: ${c.name}\n${yaml(c, "          ")}`),
    ].join("\n"),
    helm:
      recs.length === 1
        ? `# values.yaml\n${yaml(recs[0], "")}`
        : recs.map((c) => `# values.yaml — ${c.name}\n${c.name}:\n${yaml(c, "  ")}`).join("\n\n"),
    kubectl:
      target.kind === "Rollout"
        ? // CRDs do not support strategic merge; a JSON patch per container index keeps the rest of the template.
          `kubectl -n ${d.namespace} patch rollout ${target.name} --type=json -p '${JSON.stringify(
            recs.map((c) => ({
              op: "replace",
              path: `/spec/template/spec/containers/${(d.containers ?? []).findIndex((x) => x.name === c.name)}/resources`,
              value: resources(c),
            })),
          )}'`
        : `kubectl -n ${d.namespace} patch ${target.kind.toLowerCase()} ${target.name} --type=strategic -p '${JSON.stringify(
            {
              spec: {
                template: { spec: { containers: recs.map((c) => ({ name: c.name, resources: resources(c) })) } },
              },
            },
          )}'`,
  }[format];
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <div>
          <h2>
            <Art kind="scroll" size={44} /> Patches
          </h2>
          <p className="rz-dim rz-xs">
            Put the recommendation in Git instead of letting the applier patch the cluster. Generated{" "}
            {fmtDate(d.generatedAt)}.
          </p>
        </div>
        <div className="rz-zoom">
          {(
            [
              ["kustomize", "Kustomize patch"],
              ["helm", "Helm values"],
              ["kubectl", "kubectl"],
            ] as const
          ).map(([k, l]) => (
            <button key={k} aria-pressed={format === k} onClick={() => setFormat(k)}>
              {l}
            </button>
          ))}
        </div>
      </div>
      <div className="rz-code">
        <CopyButton text={text} />
        <pre>{text}</pre>
      </div>
      <p className="rz-foot rz-dim rz-xs">
        With GitOps, add <code>managedFieldsManagers: [resourcerer]</code> to the Argo CD ignoreDifferences of this
        workload so applied values are not reverted.
      </p>
    </section>
  );
}

export function SettingsTable({
  values,
  onEdit,
  showSource = true,
}: {
  values: SettingInfo[];
  onEdit?: (s: SettingInfo) => void;
  showSource?: boolean;
}) {
  const groups = [...new Set(values.map((s) => meta(s.key).group))];
  return (
    <table className="rz-table settings">
      <thead>
        <tr>
          <th>Setting</th>
          <th>Value</th>
          {showSource && <th>Source</th>}
          <th>Annotation</th>
          <th />
        </tr>
      </thead>
      <tbody>
        {groups.map((g) => [
          <tr key={g} className="grp">
            <td colSpan={5}>{g}</td>
          </tr>,
          ...values
            .filter((s) => meta(s.key).group === g)
            .map((s) => (
              <tr key={s.key}>
                <td>
                  <b>{meta(s.key).label}</b>
                  <div className="rz-dim rz-xs">{meta(s.key).help(s.value)}</div>
                </td>
                <td>
                  <code>{s.value}</code>
                </td>
                {showSource && (
                  <td>
                    <SourceTag source={s.source} />
                  </td>
                )}
                <td className="rz-dim rz-xs rz-mono">{s.annotation}</td>
                <td>
                  {onEdit && (
                    <button className="rz-icon" aria-label={`Edit ${meta(s.key).label}`} onClick={() => onEdit(s)}>
                      <Pencil size={14} />
                    </button>
                  )}
                </td>
              </tr>
            )),
        ])}
      </tbody>
    </table>
  );
}

function SettingsTab({ d, readOnly }: { d: Detail; readOnly: boolean }) {
  const [edit, setEdit] = useState<SettingInfo | null>(null);
  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <div>
          <h2>Settings</h2>
          <p className="rz-dim rz-xs">
            Effective values for {d.name}: cluster defaults, overridden by namespace annotations, overridden by workload
            annotations.
          </p>
        </div>
      </div>
      <SettingsTable values={d.settings ?? []} onEdit={readOnly ? undefined : setEdit} />
      {edit && <EditSetting d={d} setting={edit} onClose={() => setEdit(null)} />}
    </section>
  );
}

function EditSetting({ d, setting, onClose }: { d: Detail; setting: SettingInfo; onClose: () => void }) {
  const m = meta(setting.key);
  const choices = setting.options ?? m.suggestions ?? [];
  const [scope, setScope] = useState<"workload" | "namespace">(
    setting.source === "namespace" ? "namespace" : "workload",
  );
  const [value, setValue] = useState(setting.value);
  const save = useAction(d, (v: string | null) =>
    scope === "workload"
      ? api.setWorkloadSetting(d, setting.key, v)
      : api.setNamespaceSetting(d.namespace, setting.key, v),
  );
  const run = (v: string | null) =>
    save.mutate(v, {
      onSuccess: () => {
        toast({
          kind: "ok",
          title: v == null ? "Setting reset" : "Setting saved",
          body: `${setting.annotation}${v == null ? " removed" : `: "${v}"`} on ${scope === "workload" ? `${d.kind} ${d.name}` : `namespace ${d.namespace}`}. The recommendation is recomputed.`,
        });
        onClose();
      },
    });
  return (
    <Modal
      title={m.label}
      onClose={onClose}
      actions={
        <>
          {setting.source === scope && (
            <button className="rz-btn ghost" onClick={() => run(null)} disabled={save.isPending}>
              Reset to inherited
            </button>
          )}
          <span className="rz-grow" />
          <button className="rz-btn ghost" onClick={onClose}>
            Cancel
          </button>
          <button
            className="rz-btn primary"
            disabled={save.isPending || (value === setting.value && scope === setting.source)}
            onClick={() => run(value)}
          >
            Save
          </button>
        </>
      }
    >
      <div className="rz-scope rz-zoom" role="radiogroup" aria-label="Scope">
        <span className="rz-dim rz-xs">Set on:</span>
        <button aria-pressed={scope === "workload"} onClick={() => setScope("workload")}>
          This {d.kind}
        </button>
        <button aria-pressed={scope === "namespace"} onClick={() => setScope("namespace")}>
          Namespace {d.namespace}
        </button>
      </div>
      <div className="rz-radios" role="radiogroup">
        {choices.map((o) => (
          <label key={o} className={`rz-radio${value === o ? " on" : ""}`}>
            <input type="radio" name="opt" checked={value === o} onChange={() => setValue(o)} />
            <div>
              <b>{o}</b>
              <div className="rz-dim rz-xs">{m.help(o)}</div>
            </div>
          </label>
        ))}
      </div>
      {!setting.options && (
        <label className="rz-dim rz-xs">
          Or type a value
          <input className="rz-textin" value={value} onChange={(e) => setValue(e.target.value)} />
        </label>
      )}
      <ErrorBox error={save.error} />
      <p className="rz-dim rz-xs">Saved as an annotation, so the same setting can live in Git:</p>
      <pre className="rz-snippet">{`# ${scope === "workload" ? `${d.kind} ${d.namespace}/${d.name}` : `Namespace ${d.namespace}`}\nmetadata:\n  annotations:\n    ${setting.annotation}: "${value}"`}</pre>
    </Modal>
  );
}
