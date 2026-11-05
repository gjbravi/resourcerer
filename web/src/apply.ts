import type { Detail } from "./api";
import { changed } from "./format";

export interface ApplyChoices {
  /** Containers the recommendation optimizes, in the workload's order, and whether it changes them. */
  containers: { name: string; changed: boolean }[];
  /** Whether a subset may be applied: more than one container changes and no HPA target depends on all of them. */
  partial: boolean;
  /** Why only the whole recommendation can be applied, when more than one container changes. */
  reason?: string;
}

/** What the apply dialog may offer: the containers to choose from, and whether choosing is allowed. */
export function applyChoices(d: Pick<Detail, "containers" | "hpaRecommendation">): ApplyChoices {
  const containers = (d.containers ?? [])
    .filter((c) => !c.skipped)
    .map((c) => ({
      name: c.name,
      changed: changed(c.currentRequests, c.recommendedRequests) || changed(c.currentLimits, c.recommendedLimits),
    }));
  const changing = containers.filter((c) => c.changed).length;
  if (changing < 2) return { containers, partial: false };
  if (d.hpaRecommendation)
    return {
      containers,
      partial: false,
      reason: `HPA ${d.hpaRecommendation.name}'s new target needs every container's CPU, so all containers are applied together.`,
    };
  return { containers, partial: true };
}

/** The containers to send with an apply: undefined when the selection is every changing container. */
export function applySelection(choices: ApplyChoices, selected: ReadonlySet<string>): string[] | undefined {
  const changing = choices.containers.filter((c) => c.changed).map((c) => c.name);
  if (!choices.partial || changing.every((n) => selected.has(n))) return undefined;
  return changing.filter((n) => selected.has(n));
}
