import type { Detail } from "./api";
import { fmtMoney } from "./format";
import type { Signal } from "./reliability";

export type Tone = "ok" | "warn" | "bad" | "wait";

export interface Reason {
  tone: "ok" | "warn" | "info";
  text: string;
}

export interface Verdict {
  tone: Tone;
  title: string;
  /** What applying changes in money, when there is a recommendation. */
  effect?: string;
  reasons: Reason[];
}

/** CPU above the request this share of the time is expected: the request targets a percentile, not the peak. */
const CPU_ABOVE_WARN = 10;

const span = (s: number) =>
  s >= 86400 ? `${+(s / 86400).toFixed(1)}d` : s >= 3600 ? `${+(s / 3600).toFixed(1)}h` : `${Math.round(s / 60)}m`;

/** Whether to apply the recommendation, and why, from the status and the open reliability signals. */
export function verdict(d: Detail, open: Signal[]): Verdict {
  switch (d.phase) {
    case "Pending":
      return { tone: "wait", title: "Collecting usage history", reasons: [{ tone: "info", text: d.message ?? "" }] };
    case "Error":
      return { tone: "bad", title: "Metrics unavailable", reasons: [{ tone: "warn", text: d.message ?? "" }] };
    case "Unsupported":
      return { tone: "wait", title: "Not optimized", reasons: [{ tone: "info", text: d.message ?? "" }] };
  }
  const saved = d.current.monthlyCost - d.recommended.monthlyCost;
  const pct = d.current.monthlyCost > 0 ? Math.round((Math.abs(saved) / d.current.monthlyCost) * 100) : 0;
  const effect =
    Math.abs(saved) < 0.005
      ? undefined
      : saved > 0
        ? `saves ${fmtMoney(saved)}/mo (${pct}%)`
        : `costs ${fmtMoney(-saved)}/mo more (${pct}%)`;
  if (!d.pending)
    return {
      tone: "ok",
      title: "Already right-sized",
      reasons: [{ tone: "ok", text: "Current requests match the recommendation." }],
    };

  const reasons: Reason[] = [];
  const ev = d.evidence;
  if (ev) {
    const covered = ev.learningSeconds > 0 ? ev.historySeconds / ev.learningSeconds : 1;
    reasons.push(
      covered >= 0.95
        ? { tone: "ok", text: `${span(ev.learningSeconds)} of history` }
        : { tone: "warn", text: `${span(ev.historySeconds)} of ${span(ev.learningSeconds)} history: preliminary` },
    );
    if ((ev.memoryAbovePercent ?? 0) > 0)
      reasons.push({ tone: "warn", text: `memory passed the recommendation ${ev.memoryAbovePercent}% of the time` });
    else if (ev.memoryPeakPercent)
      reasons.push({ tone: "ok", text: `memory peak at ${ev.memoryPeakPercent}% of the recommendation` });
    const cpu = ev.cpuAbovePercent ?? 0;
    reasons.push(
      cpu > CPU_ABOVE_WARN
        ? { tone: "warn", text: `CPU above the recommendation ${cpu}% of the time` }
        : { tone: "ok", text: `CPU above the recommendation ${cpu}% of the time` },
    );
  } else if (d.phase === "Learning") {
    reasons.push({ tone: "warn", text: "preliminary: the learning period is not covered yet" });
  }
  for (const s of open)
    reasons.push(s.fixed ? { tone: "ok", text: `fixes ${s.title}` } : { tone: "warn", text: s.title });
  if (d.hpaRecommendation)
    reasons.push({
      tone: "info",
      text: `HPA target ${d.hpaRecommendation.current}% → ${d.hpaRecommendation.recommended}%`,
    });

  const warned = reasons.some((r) => r.tone === "warn");
  if (d.phase === "Warning") return { tone: "warn", title: "Cannot apply now", effect, reasons };
  if (saved < 0 && open.some((s) => s.fixed)) return { tone: "ok", title: "Apply to fix reliability", effect, reasons };
  return warned || d.phase === "Learning"
    ? { tone: "warn", title: "Review before applying", effect, reasons }
    : { tone: "ok", title: "Safe to apply", effect, reasons };
}
