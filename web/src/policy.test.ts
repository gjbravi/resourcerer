import { describe, expect, it } from "vitest";
import { blockedRule, until } from "./policy";

const now = Date.parse("2026-11-02T18:00:00Z");

describe("blockedRule", () => {
  it("names the rule behind the field path", () => {
    expect(blockedRule("when.freezes")).toBe("a freeze");
    expect(blockedRule("when.windows")).toBe("the apply windows");
    expect(blockedRule("limits.cooldown")).toBe("the cooldown after the last change");
  });
  it("still reads as a policy when the path is unknown or missing", () => {
    expect(blockedRule("when.somethingNew")).toBe("an apply policy");
    expect(blockedRule(undefined)).toBe("an apply policy");
  });
});

describe("until", () => {
  it("gives the clock time and how far away it is", () => {
    expect(until("2026-11-02T18:30:00Z", now)).toContain("(in 30m)");
    expect(until("2026-11-04T18:00:00Z", now)).toContain("(in 2d)");
    expect(until("2026-11-02T18:00:00Z", now)).toContain("(any moment)");
  });
  it("is empty when there is no time", () => {
    expect(until(undefined, now)).toBe("");
    expect(until("not a time", now)).toBe("");
  });
});
