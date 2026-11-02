import type { Detail } from "./api";
import { blockedRule, until } from "./policy";

export type AutoTone = "on" | "off" | "wait" | "done";

export interface AutoDeploy {
  on: boolean;
  tone: AutoTone;
  /** What auto-deploy does with the recommendation on screen. */
  text: string;
  /** Where the setting comes from: cluster, namespace or workload. */
  source: string;
  threshold: string;
}

const AUTO_USER = "auto-deploy";

/**
 * What auto-deploy does with this recommendation, mirroring the applier: it applies Ready recommendations that
 * moved more than the threshold, once per hash, and never one that was already applied or reverted.
 */
export function autoDeploy(d: Detail): AutoDeploy {
  const setting = d.settings?.find((s) => s.key === "auto-deploy");
  const threshold = d.settings?.find((s) => s.key === "auto-deploy.threshold")?.value ?? "10%";
  const base = { on: d.autoDeploy, source: setting?.source ?? "cluster", threshold };
  if (!d.autoDeploy)
    return {
      ...base,
      tone: "off",
      text: `Applies Ready recommendations that move requests more than ${threshold}, without a click.`,
    };
  const seen = d.hash ? d.history.find((h) => h.hash === d.hash) : undefined;
  if (seen?.action === "Revert") return { ...base, tone: "wait", text: "Skips this recommendation: it was reverted." };
  if (seen) return { ...base, tone: "done", text: seen.by === AUTO_USER ? "Applied it automatically." : "Applied." };
  if (!d.pending) return { ...base, tone: "done", text: `Nothing to apply: requests are within ${threshold}.` };
  const held = d.applyPolicy?.auto;
  if (held?.state === "Blocked") {
    const at = until(held.nextAllowed);
    const why = held.reason || `blocked by ${blockedRule(held.blockedBy)}`;
    return {
      ...base,
      tone: "wait",
      text:
        [`Held by ${held.policy ?? "an apply policy"}: ${why}`, at && `may apply ${at}`].filter(Boolean).join(" · ") +
        ".",
    };
  }
  if (d.phase === "Ready") return { ...base, tone: "on", text: "Applies this recommendation in a moment." };
  if (d.phase === "Learning" || d.phase === "Pending")
    return { ...base, tone: "wait", text: "Applies it once the recommendation is Ready." };
  return { ...base, tone: "wait", text: `Paused while the recommendation is ${d.phase}.` };
}
