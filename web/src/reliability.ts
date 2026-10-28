import type { Summary } from "./api";

export interface Signal {
  tone: "bad" | "warn";
  title: string;
  detail: string;
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
    });
  if (r?.crashLooping?.length)
    out.push({
      tone: "bad",
      title: `CrashLoopBackOff · ${r.crashLooping.join(", ")}`,
      detail: "The container keeps crashing; fix it before changing its resources.",
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
        });
    }
  if (r?.restarts)
    out.push({
      tone: "warn",
      title: `${r.restarts} ${r.restarts === 1 ? "restart" : "restarts"}`,
      detail: "Containers exited with errors other than OOM kills.",
    });
  if (r?.hpaAtMax)
    out.push({
      tone: "warn",
      title: "HPA at its maximum replicas",
      detail: "It cannot absorb more load; lower CPU requests raise its utilization.",
    });
  if (r?.bestEffortPods)
    out.push({
      tone: "warn",
      title: `${r.bestEffortPods} BestEffort ${r.bestEffortPods === 1 ? "pod" : "pods"}`,
      detail: "Pods without requests are the first evicted under node pressure.",
    });
  return out.sort((a, b) => (a.tone === b.tone ? 0 : a.tone === "bad" ? -1 : 1));
}
