import type { DecisionReport } from "./api";
import type { Tone } from "./verdict";

export interface ModelOption {
  name: string;
  label: string;
  probability: number;
  /** The model's choice. */
  chosen: boolean;
}

export interface ModelView {
  model: string;
  /** Why there is no evaluation, when the model could not answer. */
  error?: string;
  evaluation?: {
    version?: string;
    evaluatedAt: string;
    choice: ModelOption;
    tone: Tone;
    /** Apply, review and wait, in that order. */
    options: ModelOption[];
  };
}

const labels: Record<string, [string, Tone]> = {
  apply: ["Safe to apply", "ok"],
  review: ["Review before applying", "warn"],
  wait: ["Wait", "wait"],
};

/** Reasons the server gives when there is nothing to ask, rather than a failure to show. */
const quiet = ["No decision model is configured.", "Nothing to apply."];

/**
 * The decision model block: absent without a model or with nothing to apply; otherwise the last evaluation, or the
 * reason the model could not answer.
 */
export function modelView(r: DecisionReport | undefined, pending: boolean): ModelView | undefined {
  if (!r?.configured || !pending || (r.reason && quiet.includes(r.reason))) return undefined;
  const view: ModelView = { model: r.model ?? "", error: r.reason };
  const e = r.evaluation;
  if (!e) return view;
  const options = e.options.map((o) => ({
    name: o.name,
    label: labels[o.name]?.[0] ?? o.name,
    probability: o.probability,
    chosen: o.name === e.choice,
  }));
  const choice = options.find((o) => o.chosen);
  if (!choice) return view;
  return {
    ...view,
    evaluation: {
      version: e.version,
      evaluatedAt: e.evaluatedAt,
      choice,
      tone: labels[e.choice]?.[1] ?? "wait",
      options,
    },
  };
}

/** A probability as a whole percent. */
export const fmtProb = (p: number) => `${Math.round(p * 100)}%`;
