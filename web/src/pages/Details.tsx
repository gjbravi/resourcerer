import {
  ArrowDown,
  ArrowRight,
  ArrowUp,
  ChevronLeft,
  CircleCheck,
  CircleDollarSign,
  CircleX,
  Cpu,
  Flame,
  Hourglass,
  Info,
  Lock as LockIcon,
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
import { useState, type ReactNode } from "react";
import { api, type Detail, type Range, type Ref, type Resource, type SettingInfo } from "../api";
import { Art, NoteArt, WandArt } from "../art";
import { TimeChart, type ChartEvent } from "../chart";
import {
  Change,
  CopyButton,
  Delta,
  ErrorBox,
  FitBar,
  InfoTip,
  Metric,
  Modal,
  Money,
  SourceTag,
  StatusBadge,
  phaseTone,
} from "../components";
import { ago, fmt, fmtCores, fmtDate, fmtMem, fmtMoney, Gi, isoSeconds, Mi, quantity } from "../format";
import { autoDeploy } from "../autodeploy";
import { holdText, manualHold } from "../policy";
import { fmtWindow } from "../placement";
import { toast, useAction, useImpact, useReliability, useSeries, useWorkload, workloadHash } from "../hooks";
import { resolved, signals, type Signal } from "../reliability";
import { verdict } from "../verdict";
import type { Lock } from "../session";
import { CostSection } from "./Placement";
import { ReliabilityTab } from "./ReliabilityTab";
import { meta } from "../settings";

const TABS = [
  ["impact", "Impact"],
  ["recommendation", "Recommendation"],
  ["reliability", "Reliability"],
  ["patches", "Patches"],
  ["settings", "Settings"],
] as const;
const RANGES: Range[] = ["1h", "1d", "1w", "2w", "1m"];
const RANGE_SECONDS: Record<Range, number> = {
  learning: 0,
  "1h": 3600,
  "1d": 86400,
  "1w": 7 * 86400,
  "2w": 14 * 86400,
  "1m": 30 * 86400,
};
const learningSeconds = (d: Detail) =>
  isoSeconds(d.settings?.find((x) => x.key === "learning-period")?.value) || d.evidence?.learningSeconds || 7 * 86400;
const fmtCpuAxis = (v: number) => `${Math.round(v * 1000)}m`;
const fmtMemAxis = (v: number) => (v >= Gi ? `${(v / Gi).toFixed(1)}GiB` : `${Math.round(v / Mi)}MiB`);

export function Details({ wref, tab, lock }: { wref: Ref; tab: string; lock: Lock }) {
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
  const held = manualHold(d);
  const canApply =
    lock.write && !!d.hash && d.pending && (d.phase === "Ready" || d.phase === "Learning") && !held.blocked;
  const lastApply = d.history.find((h) => h.action === "Apply");
  const net = d.current.monthlyCost - d.recommended.monthlyCost;
  const sig = signals(d);
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
        <span className="rz-dim" title="Costs are priced at the average replicas over the learning period">
          {d.namespace} · {d.kind} · {d.replicas} {d.replicas === 1 ? "replica" : "replicas"} (
          {d.avgReplicas.toFixed(1)} avg) ·{" "}
          {d.hpaRecommendation ? `HPA ${d.hpaRecommendation.name} at ${d.hpaRecommendation.current}% CPU` : "no HPA"}
        </span>
      </div>

      <ReliabilityAlert d={d} sig={sig} />

      <section className="rz-metrics">
        <Metric
          label={
            <>
              {net >= 0 ? "Estimated savings" : "Extra capacity needed"}{" "}
              <InfoTip label="How the savings are estimated">
                <b>
                  What the recommendation {net >= 0 ? "saves" : "adds"} each month, from what the requests reserve, not
                  from usage.
                </b>
                <span>
                  Requests summed over containers × {d.avgReplicas.toFixed(1)} average replicas × the hourly rates × 730
                  h, now and with the recommendation; the difference is this number.
                </span>
                <span>
                  Rates:{" "}
                  {d.pricing?.allocation === "node"
                    ? "the prices of the nodes it ran on over the learning period."
                    : "the cluster's average rates (Settings)."}
                </span>
                <span className="rz-dim">
                  The full calculation is in How the cost is calculated, in the Impact tab.
                </span>
              </InfoTip>
            </>
          }
          art={<NoteArt />}
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
          label={
            <>
              Rightsizing score{" "}
              <InfoTip label="How the rightsizing score works">
                <b>How close the current requests are to the recommendation, from 0 to 100.</b>
                <span>
                  For CPU and for memory: the smaller of current and recommended requests over the larger, so requests
                  that are too high and too low both lower it.
                </span>
                <span>The score weights CPU and memory by what they cost. 100 means the requests already match.</span>
              </InfoTip>
            </>
          }
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
        <div className="rz-actleft">
          <div className="rz-recactions" style={{ marginTop: 0 }}>
            <button
              className="rz-btn primary"
              disabled={!canApply}
              onClick={() => setConfirm(true)}
              title={
                !lock.write
                  ? lock.reason
                  : !d.pending
                    ? "Current requests already match the recommendation"
                    : held.blocked
                      ? holdText(held)
                      : d.message
              }
            >
              {held.blocked ? <LockIcon size={15} /> : <Sparkles size={15} />} Apply recommendation
            </button>
            <button
              className="rz-btn ghost"
              disabled={!lock.write || d.phase === "Unsupported" || regenerate.isPending}
              title={lock.reason}
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
            <StatusBadge phase={d.phase} title={d.message} />
          </div>
          {held.blocked && (
            <span className="rz-recmeta" style={{ marginTop: 0 }} title={held.policy}>
              <LockIcon size={13} /> {holdText(held)}
            </span>
          )}
          {lastApply && (
            <span className="rz-recmeta" style={{ marginTop: 0 }}>
              <span>
                Last applied <b>{ago(lastApply.time)}</b> by {lastApply.by}
              </span>
            </span>
          )}
        </div>
        <AutoDeployCard d={d} lock={lock} />
      </div>

      <Verdict d={d} />

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

      <nav className="rz-tabbar" role="tablist">
        {TABS.map(([id, label]) => (
          <a key={id} role="tab" aria-selected={tab === id} href={workloadHash(d, id)}>
            {label}
            {id === "reliability" && sig.length > 0 && <em className={sig[0].tone}>{sig.length}</em>}
          </a>
        ))}
      </nav>

      {tab === "impact" && <ImpactTab d={d} />}
      {tab === "recommendation" && <RecommendationTab d={d} />}
      {tab === "patches" && <PatchesTab d={d} />}
      {tab === "reliability" && <ReliabilityTab d={d} />}
      {tab === "settings" && <SettingsTab d={d} lock={lock} />}

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
              {i === d.events.findIndex((x) => x.reason === "RecommendationApplied") && lastApply && !lock.readOnly && (
                <button
                  className="rz-btn ghost small"
                  onClick={doRevert}
                  disabled={!lock.write || revert.isPending}
                  title={lock.reason}
                >
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

/** Chart ranges; with learning (seconds), the learning period comes first, labeled with its length. */
export function Zoom({ value, onChange, learning }: { value: Range; onChange: (r: Range) => void; learning?: number }) {
  // The learning period takes its place by length, and replaces a fixed range of the same length.
  const zooms = [
    ...RANGES.filter((z) => RANGE_SECONDS[z] !== learning).map((z) => ({ z, s: RANGE_SECONDS[z], label: z })),
    ...(learning ? [{ z: "learning" as Range, s: learning, label: fmtWindow(learning) }] : []),
  ].sort((a, b) => a.s - b.s);
  return (
    <div className="rz-seg" role="group" aria-label="Time range">
      {zooms.map(({ z, label }) => (
        <button
          key={z}
          aria-pressed={value === z}
          onClick={() => onChange(z)}
          title={
            z === "learning" ? "The learning period: the history the recommendation and the costs come from" : undefined
          }
        >
          {label.toUpperCase()}
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
  const [range, setRange] = useState<Range>("learning");
  const sel = containers[ci];
  return (
    <>
      <div className="rz-sechd">
        <h2>Containers</h2>
        <Zoom value={range} onChange={setRange} learning={learningSeconds(d)} />
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
          },
        ]}
      />
    </>
  );
}

/**
 * Signals from the learning period, checked against the last 15 minutes: a fix shows up as resolved right
 * away instead of when its old samples leave the learning period.
 */
function ReliabilityAlert({ d, sig }: { d: Detail; sig: Signal[] }) {
  const recent = useReliability(d, "15m", true).data;
  if (!sig.length) return null;
  const open = sig.filter((x) => !resolved(x, recent));
  const fixed = sig.length - open.length;
  const bad = open.filter((x) => x.tone === "bad").length;
  const warn = open.length - bad;
  const top = open[0];
  return (
    <a className={`rz-relalert ${top ? top.tone : "ok"}`} href={workloadHash(d, "reliability")}>
      <span className="ic">
        {!top ? <ShieldCheck size={16} /> : top.tone === "bad" ? <Flame size={16} /> : <TriangleAlert size={16} />}
      </span>
      <span className="main">
        <b>{top ? top.title : `${fixed} reliability ${fixed === 1 ? "issue" : "issues"} resolved in the last 15m`}</b>
        <span>
          {top
            ? open.length > 1
              ? `and ${open.length - 1} more reliability ${open.length === 2 ? "issue" : "issues"}`
              : top.detail
            : "Still counted in the learning period until the old samples age out."}
        </span>
      </span>
      <span className="cnt">
        {bad > 0 && <span className="b">{bad} critical</span>}
        {warn > 0 && (
          <span className="w">
            {warn} {warn === 1 ? "warning" : "warnings"}
          </span>
        )}
        {fixed > 0 && top && <span className="o">{fixed} resolved</span>}
      </span>
      <span className="go">Review →</span>
    </a>
  );
}

function Verdict({ d }: { d: Detail }) {
  const recent = useReliability(d, "15m", true).data;
  const v = verdict(
    d,
    signals(d).filter((x) => !resolved(x, recent)),
  );
  return (
    <section className={`rz-card rz-verdict t-${v.tone}`}>
      <span className="ic">
        {v.tone === "ok" ? (
          <CircleCheck size={20} />
        ) : v.tone === "wait" ? (
          <Hourglass size={20} />
        ) : v.tone === "bad" ? (
          <CircleX size={20} />
        ) : (
          <TriangleAlert size={20} />
        )}
      </span>
      <div>
        <h2>
          {v.title}
          {v.effect && <span className="eff"> · {v.effect}</span>}
        </h2>
        <ul className="rz-reasons">
          {v.reasons
            .filter((r) => r.text)
            .map((r) => (
              <li key={r.text} className={r.tone}>
                {r.text}
              </li>
            ))}
        </ul>
      </div>
      <span
        className="rz-soon"
        title="Statistical models that estimate the chance of peaks beyond the usage history are on the way."
      >
        <Sparkles size={12} /> Decision models coming soon
      </span>
    </section>
  );
}

function ImpactTab({ d }: { d: Detail }) {
  const [range, setRange] = useState<Range>("learning");
  const [resource, setResource] = useState<Resource>("cpu");
  const used = d.evidence?.used;
  const usedCost = used ? Number(used.monthlyCost) : undefined;
  const idle = usedCost != null ? d.current.monthlyCost - usedCost : undefined;
  const ev = d.evidence;
  return (
    <>
      <div className="rz-ba-grid">
        <BeforeAfter
          icon={<CircleDollarSign size={15} />}
          label="Monthly cost"
          from={d.current.monthlyCost}
          to={d.recommended.monthlyCost}
          fmt={fmtMoney}
          delta={<Delta from={d.current.monthlyCost} to={d.recommended.monthlyCost} money />}
          note={
            idle != null && idle > 0.005 ? `${fmtMoney(idle)} of it pays for capacity the pods do not use` : undefined
          }
        />
        <BeforeAfter
          icon={<Cpu size={15} />}
          label="CPU, all replicas"
          from={d.current.cpu}
          to={d.recommended.cpu}
          fmt={fmtCores}
          delta={<Delta from={d.current.cpu} to={d.recommended.cpu} r="cpu" />}
          note={used ? `uses ${fmtCores(Number(used.cpu))} at p95` : undefined}
        />
        <BeforeAfter
          icon={<MemoryStick size={15} />}
          label="Memory, all replicas"
          from={d.current.memory}
          to={d.recommended.memory}
          fmt={fmtMem}
          delta={<Delta from={d.current.memory} to={d.recommended.memory} r="memory" />}
          note={used ? `uses ${fmtMem(Number(used.memory))} at p95` : undefined}
        />
      </div>

      <section className="rz-card pad">
        <div className="rz-sechd">
          <h2>Usage against the recommendation</h2>
          <div className="rz-ba-ctl">
            <div className="rz-seg" role="group" aria-label="Resource">
              {(["cpu", "memory"] as const).map((r) => (
                <button key={r} aria-pressed={resource === r} onClick={() => setResource(r)}>
                  {r === "cpu" ? "CPU" : "Memory"}
                </button>
              ))}
            </div>
            <Zoom value={range} onChange={setRange} learning={learningSeconds(d)} />
          </div>
        </div>
        {ev && (
          <p className="rz-dim rz-xs rz-headroom">
            {resource === "cpu"
              ? `During the learning period, the busiest pod peaked at ${ev.cpuPeakPercent ?? 0}% of the recommended CPU and used more than it ${ev.cpuAbovePercent ?? 0}% of the time.`
              : `During the learning period, the busiest pod peaked at ${ev.memoryPeakPercent ?? 0}% of the recommended memory.`}
          </p>
        )}
        <ImpactChart d={d} resource={resource} range={range} />
      </section>

      <CostSection d={d} />
    </>
  );
}

/** One figure now and with the recommendation, on a shared scale. */
function BeforeAfter({
  icon,
  label,
  from,
  to,
  fmt,
  delta,
  note,
}: {
  icon: ReactNode;
  label: string;
  from: number;
  to: number;
  fmt: (v: number) => string;
  delta: ReactNode;
  note?: string;
}) {
  const top = Math.max(from, to) || 1;
  return (
    <div className="rz-card rz-bacard">
      <span className="lbl">
        {icon} {label}
      </span>
      <div className="vals">
        <b>{fmt(to)}</b>
        {delta}
      </div>
      <div className="bars">
        <span className="k">Now</span>
        <span className="track">
          <i className="cur" style={{ width: `${(from / top) * 100}%` }} />
        </span>
        <span className="v">{fmt(from)}</span>
        <span className="k">After</span>
        <span className="track">
          <i className="rec" style={{ width: `${(to / top) * 100}%` }} />
        </span>
        <span className="v">{fmt(to)}</span>
      </div>
      {note && <span className="note rz-dim rz-xs">{note}</span>}
    </div>
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
  lock,
}: {
  values: SettingInfo[];
  onEdit?: (s: SettingInfo) => void;
  showSource?: boolean;
  /** When the caller can read but not write, the edit action stays visible and explains itself. */
  lock?: Lock;
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
                    <button
                      className="rz-icon"
                      aria-label={`Edit ${meta(s.key).label}`}
                      disabled={!!lock && !lock.write}
                      title={lock?.reason}
                      onClick={() => onEdit(s)}
                    >
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

function SettingsTab({ d, lock }: { d: Detail; lock: Lock }) {
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
      <SettingsTable
        values={d.settings ?? []}
        onEdit={lock.readOnly ? undefined : setEdit}
        lock={lock.write ? undefined : lock}
      />
      {edit && <EditSetting d={d} setting={edit} onClose={() => setEdit(null)} />}
    </section>
  );
}

/** Auto-deploy at a glance: on or off, where that is set, what it does with this recommendation, and a switch. */
function AutoDeployCard({ d, lock }: { d: Detail; lock: Lock }) {
  const [edit, setEdit] = useState(false);
  const a = autoDeploy(d);
  const setting = d.settings?.find((s) => s.key === "auto-deploy");
  const from =
    a.source === "workload"
      ? "set on this workload"
      : a.source === "namespace"
        ? "set on the namespace"
        : "cluster default";
  return (
    <div className={`rz-autocard ${a.tone}`}>
      <span className="ic">{a.on ? <Zap size={16} /> : <ZapOff size={16} />}</span>
      <span className="main">
        <span>
          <b>Auto-deploy {a.on ? "on" : "off"}</b> <span className="rz-dim rz-xs">· {from}</span>
        </span>
        <span className="rz-dim rz-xs">{a.text}</span>
      </span>
      <button
        className="rz-switch"
        role="switch"
        aria-checked={a.on}
        aria-label="Auto-deploy"
        disabled={!lock.write || !setting}
        title={lock.reason ?? (a.on ? "Turn auto-deploy off" : "Turn auto-deploy on")}
        onClick={() => setEdit(true)}
      >
        <i />
      </button>
      {edit && setting && (
        <EditSetting
          d={d}
          setting={setting}
          initial={a.on ? "Disabled" : "Enabled"}
          workloadOnly
          onClose={() => setEdit(false)}
        />
      )}
    </div>
  );
}

function EditSetting({
  d,
  setting,
  initial,
  workloadOnly,
  onClose,
}: {
  d: Detail;
  setting: SettingInfo;
  initial?: string;
  /** Set the value on this workload only, without offering the namespace. */
  workloadOnly?: boolean;
  onClose: () => void;
}) {
  const m = meta(setting.key);
  const choices = setting.options ?? m.suggestions ?? [];
  const [scope, setScope] = useState<"workload" | "namespace">(
    setting.source === "namespace" && !workloadOnly ? "namespace" : "workload",
  );
  const [value, setValue] = useState(initial ?? setting.value);
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
      {workloadOnly && (
        <p className="rz-dim rz-xs">
          On this {d.kind} only. To set it for the whole namespace, use the{" "}
          <a href={workloadHash(d, "settings")} onClick={onClose}>
            Settings tab
          </a>
          .
        </p>
      )}
      {!workloadOnly && (
        <div className="rz-scope rz-zoom" role="radiogroup" aria-label="Scope">
          <span className="rz-dim rz-xs">Set on:</span>
          <button aria-pressed={scope === "workload"} onClick={() => setScope("workload")}>
            This {d.kind}
          </button>
          <button aria-pressed={scope === "namespace"} onClick={() => setScope("namespace")}>
            Namespace {d.namespace}
          </button>
        </div>
      )}
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
