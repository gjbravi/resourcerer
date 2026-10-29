import { describe, expect, it } from "vitest";
import type { Detail } from "./api";
import type { Signal } from "./reliability";
import { verdict } from "./verdict";

const base = {
  phase: "Ready",
  pending: true,
  current: { cpu: 2.4, memory: 6 * 2 ** 30, monthlyCost: 412 },
  recommended: { cpu: 1.1, memory: 3.5 * 2 ** 30, monthlyCost: 186 },
  evidence: { historySeconds: 604800, learningSeconds: 604800, cpuAbovePercent: 4, memoryPeakPercent: 87 },
} as unknown as Detail;

const oom: Signal = { tone: "bad", title: "2 OOM kills", detail: "", check: { id: "oom" }, fixed: true };
const restarts: Signal = { tone: "warn", title: "3 restarts", detail: "", check: { id: "restarts" } };

describe("verdict", () => {
  it("is safe with full history and headroom", () => {
    const v = verdict(base, []);
    expect(v).toMatchObject({ tone: "ok", title: "Safe to apply", effect: "saves $226.00/mo (55%)" });
    expect(v.reasons.map((r) => r.text)).toEqual([
      "7d of history",
      "memory peak at 87% of the recommendation",
      "CPU above the recommendation 4% of the time",
    ]);
  });
  it("asks for a review on preliminary history or busy CPU", () => {
    expect(verdict({ ...base, evidence: { ...base.evidence!, historySeconds: 86400 } }, []).title).toBe(
      "Review before applying",
    );
    expect(verdict({ ...base, evidence: { ...base.evidence!, cpuAbovePercent: 20 } }, []).tone).toBe("warn");
  });
  it("counts what the recommendation fixes for applying, and other signals against it", () => {
    expect(verdict(base, [oom])).toMatchObject({ tone: "ok", title: "Safe to apply" });
    expect(verdict(base, [oom]).reasons.at(-1)).toEqual({ tone: "ok", text: "fixes 2 OOM kills" });
    expect(verdict(base, [restarts]).title).toBe("Review before applying");
  });
  it("frames an increase as a reliability fix", () => {
    const up = { ...base, recommended: { ...base.recommended, monthlyCost: 500 } };
    expect(verdict(up, [oom])).toMatchObject({
      title: "Apply to fix reliability",
      effect: "costs $88.00/mo more (21%)",
    });
  });
  it("says when nothing is left to do", () => {
    expect(verdict({ ...base, pending: false }, []).title).toBe("Already right-sized");
    expect(verdict({ ...base, phase: "Pending", message: "Waiting" }, []).tone).toBe("wait");
  });
});
