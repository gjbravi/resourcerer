import { describe, expect, it } from "vitest";
import type { Detail } from "./api";
import { autoDeploy } from "./autodeploy";

const base = {
  phase: "Ready",
  pending: true,
  hash: "abc",
  autoDeploy: true,
  history: [],
  settings: [
    { key: "auto-deploy", value: "Enabled", source: "namespace", annotation: "resourcerer.io/auto-deploy" },
    {
      key: "auto-deploy.threshold",
      value: "15%",
      source: "cluster",
      annotation: "resourcerer.io/auto-deploy.threshold",
    },
  ],
} as unknown as Detail;

describe("autoDeploy", () => {
  it("explains what turning it on does", () => {
    expect(autoDeploy({ ...base, autoDeploy: false })).toMatchObject({ on: false, tone: "off" });
    expect(autoDeploy({ ...base, autoDeploy: false }).text).toContain("more than 15%");
  });
  it("applies a Ready, pending recommendation", () => {
    expect(autoDeploy(base)).toMatchObject({ on: true, tone: "on", source: "namespace", threshold: "15%" });
  });
  it("explains what holds it back when a policy blocks it", () => {
    const frozen = {
      ...base,
      applyPolicy: {
        auto: { state: "Blocked", policy: "ApplyPolicy/demo/freeze", blockedBy: "when.freezes", reason: "frozen" },
        manual: { state: "Allowed" },
      },
    } as unknown as Detail;
    const a = autoDeploy(frozen);
    expect(a).toMatchObject({ on: true, tone: "wait" });
    expect(a.text).toContain("Held by ApplyPolicy/demo/freeze");
    expect(a.text).toContain("frozen");
  });
  it("waits for a Ready recommendation", () => {
    expect(autoDeploy({ ...base, phase: "Learning" }).tone).toBe("wait");
    expect(autoDeploy({ ...base, phase: "Warning" }).text).toBe("Paused while the recommendation is Warning.");
  });
  it("does nothing within the threshold", () => {
    expect(autoDeploy({ ...base, pending: false }).text).toBe("Nothing to apply: requests are within 15%.");
  });
  it("applies a hash once and never one that was reverted", () => {
    const applied = { time: "", action: "Apply" as const, by: "auto-deploy", hash: "abc", containers: null };
    expect(autoDeploy({ ...base, history: [applied] })).toMatchObject({
      tone: "done",
      text: "Applied it automatically.",
    });
    expect(autoDeploy({ ...base, history: [{ ...applied, action: "Revert", by: "alice" }, applied] }).tone).toBe(
      "wait",
    );
  });
});
