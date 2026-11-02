import { fmtDate } from "./format";

/** The rules a policy author sees, keyed by the field path the applier records. */
const RULES: Record<string, string> = {
  "when.freezes": "a freeze",
  "when.windows": "the apply windows",
  "limits.cooldown": "the cooldown after the last change",
};

/**
 * What a `blockedBy` field path means. The applier records a path such as `when.windows` so the UI can name the
 * rule without parsing a message; a path from a newer version still reads as a policy.
 */
export function blockedRule(blockedBy?: string): string {
  return (blockedBy && RULES[blockedBy]) || "an apply policy";
}

/** When something happens: the time on the clock and how far away it is. */
export function until(iso?: string, now = Date.now()): string {
  if (!iso) return "";
  const at = Date.parse(iso);
  if (Number.isNaN(at)) return "";
  const s = Math.round((at - now) / 1000);
  if (s <= 60) return `${fmtDate(iso)} (any moment)`;
  const rel =
    s < 3600 ? `${Math.round(s / 60)}m` : s < 86400 ? `${+(s / 3600).toFixed(1)}h` : `${+(s / 86400).toFixed(1)}d`;
  return `${fmtDate(iso)} (in ${rel})`;
}
