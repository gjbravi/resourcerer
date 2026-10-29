import { describe, expect, it } from "vitest";
import type { Placement } from "./api";
import { capacitySlices, fmtWindow, nodeMonthly, placementSummary, typeRows, workloadRates } from "./placement";

const base: Placement = {
  available: true,
  allocation: "node",
  factor: 0.66,
  clusterFactor: 0.55,
  coverage: 0.92,
  capacity: { "on-demand": { requests: 0.38, cost: 0.61 }, spot: { requests: 0.54, cost: 0.31 } },
  unpriced: { requests: 0.08, cost: 0.08 },
  types: [
    { name: "m7g.2xlarge", capacity: "on-demand", factor: 0.86, requests: 0.3, cost: 0.48 },
    { name: "c6i.xlarge", capacity: "spot", factor: 0.31, requests: 0.22, cost: 0.13 },
    { name: "m6a.xlarge", capacity: "spot", factor: 0.29, requests: 0.18, cost: 0.1 },
    { name: "c5.2xlarge", capacity: "spot", factor: 0.33, requests: 0.09, cost: 0.06 },
    { name: "m7g.xlarge", capacity: "on-demand", factor: 0.84, requests: 0.08, cost: 0.13 },
    { name: "c6a.xlarge", capacity: "spot", factor: 0.3, requests: 0.03, cost: 0.01 },
    { name: "t3.xlarge", capacity: "spot", factor: 0.26, requests: 0.02, cost: 0.01 },
  ],
};

describe("placement", () => {
  it("orders capacity types and puts unpriced last", () => {
    expect(capacitySlices(base, "cost").map((s) => s.key)).toEqual(["on-demand", "spot", "unpriced"]);
  });
  it("folds instance types beyond the top into one row", () => {
    const { rows, others } = typeRows(base);
    expect(others).toHaveLength(2);
    expect(rows[5]).toMatchObject({ name: "2 other types", capacity: "spot", count: 2 });
    expect(rows[5].requests).toBeCloseTo(0.05);
    expect(rows.at(-1)?.name).toBe("Nodes without a price");
    expect(rows[0].relative).toBeCloseTo(0.86 / 0.55);
  });
  it("names what drives the cost", () => {
    expect(placementSummary(base)).toBe("38% of requests ran on on-demand and account for 61% of the cost.");
    expect(
      placementSummary({ ...base, capacity: { spot: { requests: 1, cost: 1 } }, unpriced: { requests: 0, cost: 0 } }),
    ).toBe("All requests ran on spot nodes.");
  });
  it("prices at the nodes when the cost uses the cluster average", () => {
    expect(nodeMonthly(base, 100)).toBe(100);
    expect(nodeMonthly({ ...base, allocation: "cluster" }, 100)).toBeCloseTo(120);
  });
  it("formats windows", () => {
    expect([fmtWindow(900), fmtWindow(6 * 3600), fmtWindow(7 * 86400)]).toEqual(["15m", "6h", "7d"]);
  });
});

describe("workloadRates", () => {
  const rates = { cpuCoreHour: 0.02, memoryGiBHour: 0.002 };
  it("scales the cluster rates to the workload's nodes", () => {
    const r = workloadRates(rates, { allocation: "node", factor: "0.660", clusterFactor: "0.550" });
    expect(r.relative).toBeCloseTo(1.2);
    expect(r.cpuCoreHour).toBeCloseTo(0.024);
  });
  it("keeps the cluster rates otherwise", () => {
    expect(workloadRates(rates, { allocation: "cluster", factor: "0.5", clusterFactor: "0.5" }).relative).toBe(1);
    expect(workloadRates(rates).cpuCoreHour).toBe(0.02);
  });
});
