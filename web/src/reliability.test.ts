import type { Check, ReliabilityReport, Summary } from "./api";
import { resolved, signals } from "./reliability";

const base: Summary = {
  namespace: "demo",
  kind: "Deployment",
  name: "api",
  replicas: 1,
  avgReplicas: 1,
  phase: "Ready",
  current: { cpu: 0.05, memory: 2 ** 27, monthlyCost: 1 },
  recommended: { cpu: 0.05, memory: 2 ** 27, monthlyCost: 1 },
  pending: false,
  autoDeploy: false,
  oomKills: 0,
  noRequests: false,
  under: false,
  namespaceAnnotated: false,
  reliability: { containers: [{ name: "burner", cpuThrottledPercent: 71 }], restarts: 2 },
};

const report = (checks: Partial<Check>[]): ReliabilityReport => ({
  window: "15m",
  from: 0,
  to: 900,
  step: 15,
  generatedAt: "",
  checks: checks as Check[],
});

describe("reliability signals", () => {
  it("lists the worst first and links each to its live check", () => {
    const s = signals(base);
    expect(s.map((x) => x.check)).toEqual([{ id: "throttling", container: "burner" }, { id: "restarts" }]);
  });

  it("counts a signal as resolved only when its live check passes", () => {
    const [throttling, restarts] = signals(base);
    const recent = report([
      { id: "throttling", container: "burner", status: "ok" },
      { id: "restarts", status: "warn" },
    ]);
    expect(resolved(throttling, recent)).toBe(true);
    expect(resolved(restarts, recent)).toBe(false);
    expect(resolved(throttling, undefined)).toBe(false);
    expect(resolved(throttling, report([{ id: "throttling", container: "other", status: "ok" }]))).toBe(false);
  });
});
