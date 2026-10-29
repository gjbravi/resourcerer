import type { Check, ReliabilityReport, Summary } from "./api";
import { configureSignals, resolved, severity, signals } from "./reliability";

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

describe("declared signals and severity", () => {
  const custom = (status: "ok" | "warn" | "critical", container?: string) => ({
    ...base,
    reliability: {
      signals: [{ id: "http-5xx", name: "HTTP 5xx rate", value: "0.062", display: "6.2%", status, container }],
    },
  });

  it("lists declared signals that are not ok, critical as bad", () => {
    expect(signals(custom("ok"))).toEqual([]);
    const [s] = signals(custom("critical", "app"));
    expect(s).toMatchObject({
      tone: "bad",
      title: "HTTP 5xx rate 6.2% · app",
      check: { id: "signal", signal: "http-5xx" },
    });
    expect(s.fixed).toBeUndefined();
  });

  it("matches a declared signal to its live check", () => {
    const [s] = signals(custom("warn"));
    expect(resolved(s, report([{ id: "signal", signal: "http-5xx", status: "ok" }]))).toBe(true);
    expect(resolved(s, report([{ id: "signal", signal: "latency", status: "ok" }]))).toBe(false);
  });

  it("ranks what applying does not fix above what it does", () => {
    // Throttling past 25% is critical but fixed by the recommendation; the restarts are what is left.
    expect(severity(base)).toMatchObject({ level: "warning", label: "1 warning" });
    const fixedOnly = { ...base, reliability: { containers: [{ name: "burner", cpuThrottledPercent: 71 }] } };
    expect(severity(fixedOnly)).toMatchObject({ level: "fixed", label: "1 fixed by apply" });
    expect(severity(custom("warn")).level).toBe("warning");
    expect(severity({ ...base, reliability: {} }).level).toBe("healthy");
    expect(severity(custom("critical")).rank).toBeGreaterThan(severity(custom("warn")).rank);
  });

  it("uses the configured warning levels of built-in signals", () => {
    const w = { ...base, reliability: { containers: [{ name: "app", cpuThrottledPercent: 15 }] } };
    expect(signals(w)).toHaveLength(1);
    configureSignals({ builtin: [{ id: "throttling", name: "", enabled: true, warn: 0.2 }], signals: [] });
    expect(signals(w)).toHaveLength(0);
    configureSignals(undefined);
    expect(signals(w)).toHaveLength(1);
  });
});
