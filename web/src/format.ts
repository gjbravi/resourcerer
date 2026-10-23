import type { Container, ResourceName, Workload } from "./api";

const binary: Record<string, number> = { Ki: 2 ** 10, Mi: 2 ** 20, Gi: 2 ** 30, Ti: 2 ** 40 };
const decimal: Record<string, number> = { n: 1e-9, u: 1e-6, m: 1e-3, k: 1e3, M: 1e6, G: 1e9, T: 1e12 };

/** Parses a Kubernetes quantity into base units: cores for CPU, bytes for memory. */
export function parseQuantity(q?: string): number | undefined {
  if (!q) return undefined;
  const m = /^([+-]?[0-9.]+(?:[eE][+-]?[0-9]+)?)([A-Za-z]*)$/.exec(q.trim());
  if (!m) return undefined;
  const n = parseFloat(m[1]);
  const suffix = m[2];
  if (!suffix) return n;
  if (suffix in binary) return n * binary[suffix];
  if (suffix in decimal) return n * decimal[suffix];
  return undefined;
}

/** CPU always in millicores and memory in Mi/Gi, so columns compare at a glance. */
export function format(resource: ResourceName, q?: string): string {
  const v = parseQuantity(q);
  if (v === undefined) return "–";
  if (resource === "cpu") return `${Math.round(v * 1000)}m`;
  if (Math.abs(v) >= 2 ** 30) {
    const gi = v / 2 ** 30;
    return `${Number.isInteger(gi) ? gi : gi.toFixed(1)}Gi`;
  }
  return `${Math.ceil(v / 2 ** 20)}Mi`;
}

/** Relative change from current request to VPA target, rounded to whole percent. */
export function change(current?: string, target?: string): number | undefined {
  const c = parseQuantity(current);
  const t = parseQuantity(target);
  if (c === undefined || t === undefined || c === 0) return undefined;
  return Math.round(((t - c) / c) * 100);
}

export type Sizing = "over" | "under" | "fit" | "unset" | "unknown";

/** Within ±10% of the VPA target counts as right-sized. */
const TOLERANCE = 10;

export function containerSizing(c: Container): Sizing {
  if (!c.target) return "unknown";
  if (!c.requests.cpu && !c.requests.memory) return "unset";
  let over = false;
  let under = false;
  for (const r of ["cpu", "memory"] as const) {
    if (!c.target[r]) continue;
    if (!c.requests[r]) {
      under = true;
      continue;
    }
    const pct = change(c.requests[r], c.target[r])!;
    if (pct < -TOLERANCE) over = true;
    if (pct > TOLERANCE) under = true;
  }
  // Under-provisioning risks throttling or OOM kills, so it wins over savings.
  if (under) return "under";
  if (over) return "over";
  return "fit";
}

export function sizing(w: Workload): Sizing {
  const s = w.containers.map(containerSizing).filter((x) => x !== "unknown");
  if (s.length === 0) return "unknown";
  if (s.includes("under")) return "under";
  if (s.includes("over")) return "over";
  if (s.includes("unset")) return "unset";
  return "fit";
}

export function relativeTime(iso: string, now = Date.now()): string {
  const s = Math.round((now - new Date(iso).getTime()) / 1000);
  const abs = Math.abs(s);
  const [n, u] =
    abs < 60
      ? [abs, "s"]
      : abs < 3600
        ? [Math.round(abs / 60), "m"]
        : abs < 86400
          ? [Math.round(abs / 3600), "h"]
          : [Math.round(abs / 86400), "d"];
  return s >= 0 ? `${n}${u} ago` : `in ${n}${u}`;
}
