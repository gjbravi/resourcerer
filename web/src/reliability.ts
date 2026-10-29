import type { Check, ReliabilityConfig, ReliabilityReport, Summary } from "./api";

export interface Signal {
  tone: "bad" | "warn";
  title: string;
  detail: string;
  /** The live check that measures the same thing. */
  check: { id: Check["id"]; container?: string; signal?: string };
  /** The recommendation itself fixes it, by raising the memory or the CPU limit. */
  fixed?: boolean;
}

/** Same threshold as the engine: from here the recommendation raises the CPU limit. */
export const THROTTLED_RAISE = 25;

/** Warning levels of the built-in signals, in percent; the chart values can change them. */
const levels = { throttling: 10, memoryHeadroom: 90 };

/** Applies the reliability configuration from /api/v1/info. */
export function configureSignals(cfg?: ReliabilityConfig) {
  const warn = (id: string, def: number) => {
    const w = cfg?.builtin.find((b) => b.id === id)?.warn;
    return w == null ? def : Math.round(w * 100);
  };
  levels.throttling = warn("throttling", 10);
  levels.memoryHeadroom = warn("memoryHeadroom", 90);
}

/** Reliability signals of a workload, worst first. */
export function signals(w: Summary): Signal[] {
  const r = w.reliability;
  const out: Signal[] = [];
  if (w.oomKills > 0)
    out.push({
      tone: "bad",
      title: `${w.oomKills} OOM ${w.oomKills === 1 ? "kill" : "kills"}`,
      detail: "The memory recommendation is raised above the limit that killed it.",
      check: { id: "oom" },
      fixed: true,
    });
  if (r?.crashLooping?.length)
    out.push({
      tone: "bad",
      title: `CrashLoopBackOff · ${r.crashLooping.join(", ")}`,
      detail: "The container keeps crashing; fix it before changing its resources.",
      check: { id: "crashloop" },
    });
  for (const c of r?.containers ?? []) {
    const t = c.cpuThrottledPercent ?? 0;
    if (t >= levels.throttling)
      out.push({
        tone: t >= THROTTLED_RAISE ? "bad" : "warn",
        title: `CPU throttled ${t}% of the time · ${c.name}`,
        detail:
          t >= THROTTLED_RAISE
            ? "The CPU limit caps it in bursts; the recommendation raises the limit."
            : "The CPU limit caps it in short bursts.",
        check: { id: "throttling", container: c.name },
        fixed: t >= THROTTLED_RAISE,
      });
  }
  if (w.oomKills === 0)
    for (const c of r?.containers ?? []) {
      const m = c.memoryPeakPercent ?? 0;
      if (m >= levels.memoryHeadroom)
        out.push({
          tone: "warn",
          title: `Memory peaked at ${m}% of the limit · ${c.name}`,
          detail: "One larger spike would be an OOM kill.",
          check: { id: "memory", container: c.name },
        });
    }
  if (r?.restarts)
    out.push({
      tone: "warn",
      title: `${r.restarts} ${r.restarts === 1 ? "restart" : "restarts"}`,
      detail: "Containers exited with errors other than OOM kills.",
      check: { id: "restarts" },
    });
  if (r?.hpaAtMax)
    out.push({
      tone: "warn",
      title: "HPA at its maximum replicas",
      detail: "It cannot absorb more load; lower CPU requests raise its utilization.",
      check: { id: "hpa" },
    });
  if (r?.bestEffortPods)
    out.push({
      tone: "warn",
      title: `${r.bestEffortPods} BestEffort ${r.bestEffortPods === 1 ? "pod" : "pods"}`,
      detail: "Pods without requests are the first evicted under node pressure.",
      check: { id: "qos" },
    });
  for (const x of r?.signals ?? [])
    if (x.status !== "ok")
      out.push({
        tone: x.status === "critical" ? "bad" : "warn",
        title: `${x.name} ${x.display}${x.container ? ` · ${x.container}` : ""}`,
        detail: "A signal declared in the reliability configuration.",
        check: { id: "signal", signal: x.id, container: x.container },
      });
  return out.sort((a, b) => (a.tone === b.tone ? 0 : a.tone === "bad" ? -1 : 1));
}

/** How a workload's reliability reads in a list: what applying fixes counts apart from what it does not. */
export interface Severity {
  level: "critical" | "warning" | "fixed" | "healthy";
  /** Sorts the worst first. */
  rank: number;
  label: string;
  signals: Signal[];
}

export function severity(w: Summary): Severity {
  const sig = signals(w);
  const open = sig.filter((s) => !s.fixed);
  const bad = open.filter((s) => s.tone === "bad").length;
  const n = (k: number, one: string, many: string) => `${k} ${k === 1 ? one : many}`;
  if (bad) return { level: "critical", rank: 3000 + sig.length, label: n(bad, "critical", "critical"), signals: sig };
  if (open.length)
    return { level: "warning", rank: 2000 + sig.length, label: n(open.length, "warning", "warnings"), signals: sig };
  if (sig.length)
    return { level: "fixed", rank: 1000 + sig.length, label: `${sig.length} fixed by apply`, signals: sig };
  return { level: "healthy", rank: 0, label: "healthy", signals: sig };
}

/** Whether a signal from the learning period no longer shows in a recent report. */
export function resolved(sig: Signal, recent?: ReliabilityReport): boolean {
  const c = recent?.checks.find(
    (x) =>
      x.id === sig.check.id &&
      (!sig.check.signal || x.signal === sig.check.signal) &&
      (!sig.check.container || x.container === sig.check.container),
  );
  return c?.status === "ok";
}
