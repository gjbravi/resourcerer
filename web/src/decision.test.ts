import { describe, expect, it } from "vitest";
import type { DecisionReport } from "./api";
import { fmtProb, modelView } from "./decision";

const evaluated: DecisionReport = {
  configured: true,
  model: "decider-4b",
  evaluation: {
    version: "decider-4b-v2.1",
    evaluatedAt: "2026-01-05T10:00:00Z",
    hash: "abc",
    choice: "review",
    options: [
      { name: "apply", probability: 0.3 },
      { name: "review", probability: 0.6 },
      { name: "wait", probability: 0.1 },
    ],
  },
};

describe("modelView", () => {
  it("is absent without a model or with nothing to apply", () => {
    expect(modelView(undefined, true)).toBeUndefined();
    expect(modelView({ configured: false, reason: "No decision model is configured." }, true)).toBeUndefined();
    expect(modelView({ configured: true, model: "m", reason: "Nothing to apply." }, true)).toBeUndefined();
    expect(modelView(evaluated, false)).toBeUndefined();
  });
  it("offers an evaluation before the first one", () => {
    expect(modelView({ configured: true, model: "decider-4b" }, true)).toEqual({ model: "decider-4b" });
  });
  it("shows the answer and every option in order", () => {
    const e = modelView(evaluated, true)!.evaluation!;
    expect(e.choice).toMatchObject({ label: "Review before applying", probability: 0.6 });
    expect(e.tone).toBe("warn");
    expect(e.options.map((o) => o.label)).toEqual(["Safe to apply", "Review before applying", "Wait"]);
  });
  it("carries the reason the model could not answer", () => {
    expect(modelView({ configured: true, model: "m", reason: "connection refused" }, true)).toEqual({
      model: "m",
      error: "connection refused",
    });
  });
});

describe("fmtProb", () => {
  it("rounds to a whole percent", () => {
    expect(fmtProb(0.508)).toBe("51%");
  });
});
