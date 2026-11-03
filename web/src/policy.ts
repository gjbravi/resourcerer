import type { ApplyDecision, Detail } from "./api";
import { fmtDate } from "./format";

/** A rule of an apply policy that holds an apply now, ready to be shown. */
export interface Hold {
  blocked: boolean;
  /** The policy that holds it, e.g. "ClusterApplyPolicy/production". */
  policy?: string;
  /** The rule, as the spec field path: windows, freezes, rate.cooldown, … */
  rule?: string;
  /** The same in the policy's words. */
  reason?: string;
  /** When the hold ends, when the policy can promise an instant. */
  nextAllowed?: string;
  /** What Audit-mode policies would have blocked: they record, they do not hold. */
  wouldBlock: string[];
}

/**
 * hold reads one decision. A decision that named an instant only holds until that instant, and the status is
 * written when it changes rather than every minute, so the instant has to be read here too.
 */
export function hold(d: ApplyDecision | undefined, now = Date.now()): Hold {
  const wouldBlock = d?.wouldBlock ?? [];
  if (d?.state !== "Blocked") return { blocked: false, wouldBlock };
  if (d.nextAllowed != null && Date.parse(d.nextAllowed) <= now) return { blocked: false, wouldBlock };
  return {
    blocked: true,
    policy: d.policy,
    rule: d.blockedBy,
    reason: d.reason,
    nextAllowed: d.nextAllowed,
    wouldBlock,
  };
}

/** What holds an apply a person asks for. */
export function manualHold(d: Detail, now = Date.now()): Hold {
  return hold(d.applyPolicy?.manual, now);
}

/** What holds an automatic apply. */
export function autoHold(d: Detail, now = Date.now()): Hold {
  return hold(d.applyPolicy?.auto, now);
}

/** What a rule is about, in the words of someone deciding whether to apply. */
export function ruleLabel(rule?: string): string {
  switch (rule) {
    case "windows":
      return "outside the apply windows";
    case "freezes":
      return "frozen";
    case "rate.cooldown":
      return "cooling down";
    case "rate.maxConcurrent":
      return "waiting for a slot";
    case "rate.maxPerHour":
      return "hourly limit reached";
    case "rate.perNamespace":
      return "another rollout in the namespace";
    case "rate.spread":
      return "waiting for its offset";
    case "rollout":
      return "its own rollout is in progress";
    case "policy":
      return "the policy cannot be read";
  }
  return "held by a policy";
}

/** One sentence for a hold: the reason, and when it ends. */
export function holdText(h: Hold): string {
  if (!h.blocked) return "";
  const base = h.reason ?? ruleLabel(h.rule);
  return h.nextAllowed ? `${base}; next at ${fmtDate(h.nextAllowed)}` : base;
}
