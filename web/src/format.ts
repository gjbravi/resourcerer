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

export const fmtMoney = (v: number) =>
  `${v < 0 ? "-" : ""}$${Math.abs(v).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;

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
