import type { Placement } from "./api";

export type Measure = "requests" | "cost";

export const UNPRICED = "unpriced";

/** Capacity types in a fixed order, so colors and stacking stay put between workloads. */
export const CAPACITIES = ["on-demand", "reserved", "spot", UNPRICED] as const;

export const capacityLabel = (c: string) => (c === UNPRICED ? "no price" : c);

export interface Slice {
  key: string;
  label: string;
  share: number;
}

/** Shares per capacity type, unpriced last, leaving out empty ones. */
export function capacitySlices(p: Placement, m: Measure): Slice[] {
  const all: Record<string, number> = {
    ...Object.fromEntries(Object.entries(p.capacity ?? {}).map(([k, u]) => [k, u[m]])),
  };
  all[UNPRICED] = p.unpriced[m];
  const known = CAPACITIES.filter((c) => all[c] > 0.0005);
  const other = Object.keys(all).filter((c) => !(CAPACITIES as readonly string[]).includes(c) && all[c] > 0.0005);
  return [...known, ...other].map((k) => ({ key: k, label: capacityLabel(k), share: all[k] }));
}

export interface TypeRow {
  name: string;
  capacity: string;
  requests: number;
  cost: number;
  /** Price relative to the cluster average; 1 for unpriced requests, which are counted at the average. */
  relative: number;
  /** Instance types folded into an "others" row. */
  count?: number;
}

/** The top instance types by share of requests, the rest folded into one row, and unpriced requests last. */
export function typeRows(p: Placement, top = 5): { rows: TypeRow[]; others: TypeRow[] } {
  const k = p.clusterFactor || 1;
  const all: TypeRow[] = (p.types ?? []).map((t) => ({
    name: t.name,
    capacity: t.capacity,
    requests: t.requests,
    cost: t.cost,
    relative: t.factor / k,
  }));
  const rows = all.slice(0, top);
  const others = all.slice(top);
  if (others.length) {
    const requests = others.reduce((a, t) => a + t.requests, 0);
    const cost = others.reduce((a, t) => a + t.cost, 0);
    const caps = new Set(others.map((t) => t.capacity));
    rows.push({
      name: `${others.length} other ${others.length === 1 ? "type" : "types"}`,
      capacity: caps.size === 1 ? [...caps][0] : "mixed",
      requests,
      cost,
      relative: requests > 0 ? others.reduce((a, t) => a + t.requests * t.relative, 0) / requests : 1,
      count: others.length,
    });
  }
  if (p.unpriced.requests > 0.0005)
    rows.push({ name: "Nodes without a price", capacity: UNPRICED, ...p.unpriced, relative: 1 });
  return { rows, others };
}

/** Monthly cost at the prices of the nodes the workload ran on, from its current cost. */
export function nodeMonthly(p: Placement, monthly: number): number {
  if (p.allocation === "node" || !p.factor || !p.clusterFactor) return monthly;
  return (monthly * p.factor) / p.clusterFactor;
}

/** One sentence on what drives the workload's cost. */
export function placementSummary(p: Placement): string {
  const req = capacitySlices(p, "requests");
  const cost = capacitySlices(p, "cost");
  if (!req.length) return "No requests on nodes in this window.";
  if (req.length === 1)
    return `All requests ran on ${req[0].key === UNPRICED ? "nodes without a price" : `${req[0].label} nodes`}.`;
  const top = [...cost].filter((s) => s.key !== UNPRICED).sort((a, b) => b.share - a.share)[0] ?? cost[0];
  const r = req.find((s) => s.key === top.key)?.share ?? 0;
  const pc = (v: number) => `${Math.round(v * 100)}%`;
  return `${pc(r)} of requests ran on ${top.label} and account for ${pc(top.share)} of the cost.`;
}

/** Window length for labels: 15m, 6h, 7d. */
export function fmtWindow(seconds: number): string {
  if (seconds >= 86400 * 0.95) return `${Math.round(seconds / 86400)}d`;
  if (seconds >= 3600 * 0.95) return `${Math.round(seconds / 3600)}h`;
  return `${Math.max(1, Math.round(seconds / 60))}m`;
}

/** The hourly rates a workload's cost uses: the cluster's, scaled to its nodes under node allocation. */
export function workloadRates(
  rates: { cpuCoreHour: number; memoryGiBHour: number },
  pricing?: { allocation: string; factor: string; clusterFactor: string },
): { cpuCoreHour: number; memoryGiBHour: number; relative: number } {
  const f = Number(pricing?.factor);
  const k = Number(pricing?.clusterFactor);
  const relative = pricing?.allocation === "node" && f > 0 && k > 0 ? f / k : 1;
  return { cpuCoreHour: rates.cpuCoreHour * relative, memoryGiBHour: rates.memoryGiBHour * relative, relative };
}
