import type { Check, ReliabilityReport, Summary } from "./api";

export interface Signal {
  tone: "bad" | "warn";
  title: string;
  detail: string;
  /** The live check that measures the same thing. */
  check: { id: Check["id"]; container?: string };
}

/** Same threshold as the engine: from here the recommendation raises the CPU limit. */
export const THROTTLED_RAISE = 25;
const THROTTLED_SHOW = 10;
const MEMORY_NEAR_LIMIT = 90;

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
    if (t >= THROTTLED_SHOW)
      out.push({
        tone: t >= THROTTLED_RAISE ? "bad" : "warn",
        title: `CPU throttled ${t}% of the time · ${c.name}`,
        detail:
          t >= THROTTLED_RAISE
            ? "The CPU limit caps it in bursts; the recommendation raises the limit."
            : "The CPU limit caps it in short bursts.",
        check: { id: "throttling", container: c.name },
      });
  }
  if (w.oomKills === 0)
    for (const c of r?.containers ?? []) {
      const m = c.memoryPeakPercent ?? 0;
      if (m >= MEMORY_NEAR_LIMIT)
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
  return out.sort((a, b) => (a.tone === b.tone ? 0 : a.tone === "bad" ? -1 : 1));
}

/** Whether a signal from the learning period no longer shows in a recent report. */
export function resolved(sig: Signal, recent?: ReliabilityReport): boolean {
  const c = recent?.checks.find(
    (x) => x.id === sig.check.id && (!sig.check.container || x.container === sig.check.container),
  );
  return c?.status === "ok";
}
