import type { Resource, Resources } from "./api";

export const Mi = 2 ** 20;
export const Gi = 2 ** 30;

export function fmtCpu(c?: number): string {
  if (c == null) return "–";
  if (c === 0) return "0";
  return `${Math.max(1, Math.round(c * 1000))}m`;
}

export const fmtCores = (c: number) => `${c.toFixed(c >= 10 ? 1 : 3)} Cores`;

export function fmtMem(b?: number): string {
  if (b == null) return "–";
  if (b === 0) return "0";
  if (b >= Gi) return `${(b / Gi).toFixed(2)}GiB`;
  return `${Math.max(1, Math.round(b / Mi))}MiB`;
}

export const fmt = (r: Resource, v?: number) => (r === "cpu" ? fmtCpu(v) : fmtMem(v));

export const fmtMoney = (v: number, digits = 2) =>
  `${v < 0 ? "-" : ""}$${Math.abs(v).toLocaleString(undefined, { minimumFractionDigits: digits, maximumFractionDigits: digits })}`;

export const fmtDate = (iso?: string | number) =>
  iso == null
    ? "–"
    : new Date(iso).toLocaleString([], { month: "short", day: "numeric", hour: "numeric", minute: "2-digit" });

export function ago(iso?: string, now = Date.now()): string {
  if (!iso) return "–";
  const s = Math.round((now - new Date(iso).getTime()) / 1000);
  const a = Math.abs(s);
  const v =
    a < 60
      ? `${a}s`
      : a < 3600
        ? `${Math.round(a / 60)}m`
        : a < 86400
          ? `${Math.round(a / 3600)}h`
          : `${Math.round(a / 86400)}d`;
  return s >= 0 ? `${v} ago` : `in ${v}`;
}

/** Kubernetes quantity in the units people write in manifests. */
export function quantity(r: Resource, v?: number): string | undefined {
  if (v == null) return undefined;
  return r === "cpu" ? `${Math.round(v * 1000)}m` : `${Math.ceil(v / Mi)}Mi`;
}

export function changed(a: Resources, b: Resources): boolean {
  return (["cpu", "memory"] as const).some((r) => a[r] !== b[r]);
}

/** Percent change from a to b, rounded; null when a is unset or zero. */
export function pct(a?: number, b?: number): number | null {
  if (!a || b == null) return null;
  return Math.round(((b - a) / a) * 100);
}

/** One line on where cost rates come from, for captions and tooltips. */
export function costSource(c: { source: string; state?: string; spot?: number; onDemand?: number }): string {
  const nodes = `Karpenter node prices · ${c.spot ?? 0} spot, ${c.onDemand ?? 0} on-demand nodes`;
  switch (c.state) {
    case "stale":
      return `${nodes} · last calibration, refresh failing`;
    case "error":
      return c.source === "karpenter"
        ? `${nodes} · Karpenter prices failing`
        : "configured rates · Karpenter prices unavailable";
    case "fallback":
      return "configured rates (no Karpenter in this cluster)";
  }
  return c.source === "karpenter" ? nodes : "configured rates";
}

/** Seconds in an ISO 8601 duration as the settings write it (P7D, PT15M, P1W); 0 when it does not parse. */
export function isoSeconds(v?: string): number {
  const m = /^P(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?)?$/.exec(v ?? "");
  if (!m || v === "P" || v?.endsWith("T")) return 0;
  const [w, d, h, min, s] = m.slice(1).map((x) => Number(x ?? 0));
  return (((w * 7 + d) * 24 + h) * 60 + min) * 60 + s;
}
