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
  it("waits for a Ready recommendation", () => {
    expect(autoDeploy({ ...base, phase: "Learning" }).tone).toBe("wait");
    expect(autoDeploy({ ...base, phase: "Warning" }).text).toBe("Paused while the recommendation is Warning.");
  });
  it("names the containers it is limited to", () => {
    const scoped = { ...base, autoDeployScope: { containers: ["istio-proxy"], pending: true } };
    expect(autoDeploy(scoped).text).toBe("Applies this recommendation to istio-proxy in a moment.");
    expect(autoDeploy({ ...scoped, pending: false }).tone).toBe("on");
    expect(autoDeploy({ ...scoped, autoDeployScope: { containers: ["istio-proxy"] } }).text).toBe(
      "Nothing to apply in istio-proxy: requests are within 15%.",
    );
    expect(
      autoDeploy({
        ...scoped,
        autoDeployScope: { reason: "none of the containers in auto-deploy.containers is optimized" },
      }),
    ).toMatchObject({
      tone: "wait",
      text: "Applies nothing: none of the containers in auto-deploy.containers is optimized.",
    });
  });
  it("does nothing within the threshold", () => {
    expect(autoDeploy({ ...base, pending: false }).text).toBe("Nothing to apply: requests are within 15%.");
  });
  it("says a policy holds it, and leaves the rest to that policy", () => {
    const held = autoDeploy({
      ...base,
      applyPolicy: {
        auto: {
          state: "Blocked",
          policy: "ClusterApplyPolicy/production",
          blockedBy: "windows",
          reason: "outside the apply windows (Mon,Tue 10:00-16:00 UTC)",
          nextAllowed: new Date(Date.now() + 3600_000).toISOString(),
        },
        manual: { state: "Allowed" },
      },
    } as unknown as Detail);
    expect(held).toMatchObject({ on: true, tone: "hold", heldBy: "ClusterApplyPolicy/production" });
    expect(held.text).toContain("outside the apply windows");
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
