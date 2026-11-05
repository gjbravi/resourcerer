import { describe, expect, it } from "vitest";
import type { Container, Detail } from "./api";
import { applyChoices, applySelection } from "./apply";

const c = (name: string, from: number, to: number, skipped = false): Container => ({
  name,
  skipped,
  currentRequests: { cpu: from },
  currentLimits: {},
  recommendedRequests: { cpu: to },
  recommendedLimits: {},
});

const pod = (...containers: Container[]) => ({ containers }) as Pick<Detail, "containers" | "hpaRecommendation">;

describe("applyChoices", () => {
  it("offers a choice when more than one container changes", () => {
    const ch = applyChoices(pod(c("app", 1, 0.2), c("worker", 0.05, 0.1), c("istio-proxy", 0.1, 0.1, true)));
    expect(ch).toEqual({
      containers: [
        { name: "app", changed: true },
        { name: "worker", changed: true },
      ],
      partial: true,
    });
  });
  it("offers none for a single changing container", () => {
    expect(applyChoices(pod(c("app", 1, 0.2), c("worker", 0.1, 0.1))).partial).toBe(false);
  });
  it("applies everything together when an HPA target depends on it", () => {
    const ch = applyChoices({
      ...pod(c("app", 1, 0.2), c("worker", 0.05, 0.1)),
      hpaRecommendation: { name: "web", current: 70, recommended: 80, clamped: false },
    });
    expect(ch.partial).toBe(false);
    expect(ch.reason).toContain("HPA web's new target needs every container's CPU");
  });
});

describe("applySelection", () => {
  const ch = applyChoices(pod(c("app", 1, 0.2), c("worker", 0.05, 0.1), c("log", 0.1, 0.1)));
  it("sends nothing when every changing container is selected", () => {
    expect(applySelection(ch, new Set(["app", "worker"]))).toBeUndefined();
  });
  it("sends the chosen containers in the workload's order", () => {
    expect(applySelection(ch, new Set(["worker"]))).toEqual(["worker"]);
    expect(applySelection(ch, new Set(["worker", "log"]))).toEqual(["worker"]);
  });
  it("sends nothing when choosing is not allowed", () => {
    expect(applySelection({ ...ch, partial: false }, new Set(["worker"]))).toBeUndefined();
  });
});
