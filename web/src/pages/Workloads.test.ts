import type { Summary } from "../api";
import { applicable, VIEWS } from "./Workloads";

const base: Summary = {
  namespace: "demo",
  kind: "Deployment",
  name: "api",
  replicas: 2,
  avgReplicas: 2,
  phase: "Ready",
  current: { cpu: 2, memory: 2 * 2 ** 30, monthlyCost: 52 },
  recommended: { cpu: 0.25, memory: 2 ** 27, monthlyCost: 6 },
  pending: true,
  autoDeploy: false,
  oomKills: 0,
  noRequests: false,
  under: false,
  hash: "abc",
  namespaceAnnotated: false,
};

describe("workload views", () => {
  it("only applies pending Ready or Learning recommendations with a hash", () => {
    expect(applicable(base)).toBe(true);
    expect(applicable({ ...base, phase: "Unsupported" })).toBe(false);
    expect(applicable({ ...base, pending: false })).toBe(false);
    expect(applicable({ ...base, hash: undefined })).toBe(false);
  });

  it("separates savings from reliability", () => {
    const risky = { ...base, name: "ledger", oomKills: 3, current: base.recommended, recommended: base.current };
    const rows = [base, risky];
    expect(rows.filter(VIEWS.savings[1]).map((w) => w.name)).toEqual(["api"]);
    expect(rows.filter(VIEWS.reliability[1]).map((w) => w.name)).toEqual(["ledger"]);
    expect(rows.filter(VIEWS.pending[1])).toHaveLength(2);
  });
});
