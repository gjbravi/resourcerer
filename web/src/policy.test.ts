import { describe, expect, it } from "vitest";
import type { Detail } from "./api";
import { autoHold, hold, holdText, manualHold, ruleLabel } from "./policy";

const NOW = Date.parse("2026-11-03T12:00:00Z");

const detail = (applyPolicy?: Detail["applyPolicy"]) =>
  ({ history: [], settings: [], applyPolicy }) as unknown as Detail;

describe("hold", () => {
  it("reports nothing for an allowed apply", () => {
    expect(hold({ state: "Allowed" }, NOW)).toEqual({ blocked: false, wouldBlock: [] });
    expect(hold(undefined, NOW).blocked).toBe(false);
  });
  it("carries the reason and the instant", () => {
    const h = hold(
      {
        state: "Blocked",
        policy: "ClusterApplyPolicy/production",
        blockedBy: "windows",
        reason: "outside the apply windows",
        nextAllowed: "2026-11-03T13:00:00Z",
      },
      NOW,
    );
    expect(h).toMatchObject({ blocked: true, rule: "windows", nextAllowed: "2026-11-03T13:00:00Z" });
    // The reason, then the instant it ends, in the reader's locale.
    expect(holdText(h).startsWith("outside the apply windows; next at ")).toBe(true);
  });
  it("stops holding once the instant it named has passed", () => {
    const h = hold({ state: "Blocked", blockedBy: "windows", nextAllowed: "2026-11-03T11:00:00Z" }, NOW);
    expect(h.blocked).toBe(false);
  });
  it("keeps a hold that waits for a slot, which has no instant", () => {
    const h = hold({ state: "Blocked", blockedBy: "rate.maxConcurrent", reason: "waiting for a slot: 1 of 1" }, NOW);
    expect(h.blocked).toBe(true);
    expect(holdText(h)).toBe("waiting for a slot: 1 of 1");
  });
  it("shows what an audit policy would have blocked", () => {
    const h = hold({ state: "Allowed", wouldBlock: ["ClusterApplyPolicy/quiet: outside the apply windows"] }, NOW);
    expect(h.blocked).toBe(false);
    expect(h.wouldBlock).toHaveLength(1);
  });
});

describe("the two kinds of apply", () => {
  const d = detail({
    auto: { state: "Blocked", blockedBy: "rate.cooldown", reason: "cooling down until 2026-11-04T10:00:00Z" },
    manual: { state: "Allowed" },
  });
  it("holds the automatic apply and leaves the person free", () => {
    expect(autoHold(d, NOW).blocked).toBe(true);
    expect(manualHold(d, NOW).blocked).toBe(false);
  });
});

describe("ruleLabel", () => {
  it("names each rule for a person", () => {
    expect(ruleLabel("windows")).toBe("outside the apply windows");
    expect(ruleLabel("policy")).toBe("the policy cannot be read");
    expect(ruleLabel("something-new")).toBe("held by a policy");
  });
});
