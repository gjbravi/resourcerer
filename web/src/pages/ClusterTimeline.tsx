import { LineChart } from "lucide-react";
import { useState } from "react";
import type { Point, Range, Resource, Series } from "../api";
import { at, TimeChart, type Line } from "../chart";
import { ErrorBox } from "../components";
import { fmtCores, fmtMem, fmtMoney, Gi } from "../format";
import { useClusterSeries, useInfo } from "../hooks";
import { Zoom } from "./Details";

type Mode = Resource | "cost";

const MODES: [Mode, string][] = [
  ["cpu", "CPU"],
  ["memory", "Memory"],
  ["cost", "Cost"],
];

/** Hourly cost of a cpu and a memory line sampled at the same steps. */
function costLine(cpu: Point[], mem: Point[], rates: { cpuCoreHour: number; memoryGiBHour: number }): Point[] {
  return cpu.map((p) => ({ t: p.t, v: p.v * rates.cpuCoreHour + ((at(mem, p.t)?.v ?? 0) / Gi) * rates.memoryGiBHour }));
}

const last = (pts?: Point[]) => (pts?.length ? pts[pts.length - 1].v : 0);

export function ClusterTimeline() {
  const [range, setRange] = useState<Range>("1d");
  const [mode, setMode] = useState<Mode>("cpu");
  const cpu = useClusterSeries("cpu", range);
  const mem = useClusterSeries("memory", range);
  const rates = useInfo().data?.cost ?? { cpuCoreHour: 0, memoryGiBHour: 0 };
  const cur = mode === "memory" ? mem : cpu;
  const l = (s: Series | undefined, k: string) => s?.lines[k] ?? [];

  let series: Series | undefined = cur.data;
  let lines: Line[];
  let fmtY: (v: number) => string;
  let unit: string;
  if (mode === "cost") {
    const requested = costLine(l(cpu.data, "requested"), l(mem.data, "requested"), rates);
    const withRecs = costLine(l(cpu.data, "withRecommendations"), l(mem.data, "withRecommendations"), rates);
    const usage = costLine(l(cpu.data, "usage"), l(mem.data, "usage"), rates);
    series = cpu.data && mem.data ? { ...cpu.data, bands: [], lines: {} } : undefined;
    lines = [
      { name: "Requested cost", color: "var(--c-cur)", points: requested, shape: "smooth", width: 2.6 },
      {
        name: "With recommendations",
        color: "var(--save)",
        points: withRecs,
        shape: "smooth",
        fill: "Requested cost",
      },
      { name: "Usage cost", color: "var(--band0)", points: usage, shape: "smooth", fill: true, width: 1.6 },
    ];
    fmtY = (v) => `$${v < 10 ? v.toFixed(2) : v.toFixed(0)}`;
    unit = "USD / hour";
  } else {
    const k = mode === "cpu" ? 1 : Gi;
    lines = [
      {
        name: "Allocatable",
        color: "var(--dim)",
        points: l(cur.data, "allocatable"),
        shape: "smooth",
        fill: true,
        width: 1.4,
      },
      { name: "Requested", color: "var(--c-cur)", points: l(cur.data, "requested"), shape: "smooth", width: 2.6 },
      {
        name: "With recommendations",
        color: "var(--save)",
        points: l(cur.data, "withRecommendations"),
        shape: "smooth",
        dash: "6 4",
        fill: "Requested",
      },
      { name: "Usage", color: "var(--band0)", points: l(cur.data, "usage"), shape: "smooth", fill: true },
    ];
    fmtY = (v) => (v / k).toFixed(v / k < 10 ? 1 : 0);
    unit = mode === "cpu" ? "Cores" : "GiB";
  }

  const alloc = last(l(cur.data, "allocatable"));
  const used = last(l(cur.data, "usage"));
  const req = last(l(cur.data, "requested"));
  const fmt = mode === "memory" ? fmtMem : fmtCores;
  const hourly = (k: string) => last(costLine(l(cpu.data, k), l(mem.data, k), rates).slice(-1)) * 730;

  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <h2>
          <LineChart size={17} className="t-accent" /> Cluster timeline
        </h2>
        <div className="rz-ctl">
          <Zoom value={range} onChange={setRange} />
          <span className="rz-seg">
            {MODES.map(([m, label]) => (
              <button key={m} aria-pressed={mode === m} onClick={() => setMode(m)}>
                {label}
              </button>
            ))}
          </span>
        </div>
      </div>
      <div className="rz-cluster-kpis">
        {mode === "cost" ? (
          <>
            <div>
              <span>Requested / mo</span>
              <b>{fmtMoney(hourly("requested"))}</b>
            </div>
            <div>
              <span>With recommendations / mo</span>
              <b className="t-save">{fmtMoney(hourly("withRecommendations"))}</b>
            </div>
            <div>
              <span>Used / mo</span>
              <b>{fmtMoney(hourly("usage"))}</b>
            </div>
          </>
        ) : (
          <div className="rz-eff">
            <span>
              {mode === "cpu" ? "CPU" : "Memory"} efficiency · {fmt(used)} used, {fmt(req)} requested of {fmt(alloc)}{" "}
              allocatable
            </span>
            <div className="bar">
              <i className="u" style={{ width: `${alloc ? (used / alloc) * 100 : 0}%` }} />
              <i className="r" style={{ width: `${alloc ? (Math.max(0, req - used) / alloc) * 100 : 0}%` }} />
            </div>
            <div className="lbl">
              <span>{alloc ? Math.round((used / alloc) * 100) : 0}% used</span>
              <span>{alloc ? Math.round((req / alloc) * 100) : 0}% requested</span>
            </div>
          </div>
        )}
      </div>
      <div className="rz-cluster-chart">
        <ErrorBox error={cur.error ?? (mode === "cost" ? mem.error : null)} />
        <TimeChart
          key={mode}
          title={
            mode === "cost"
              ? "Hourly cost of requests, with every recommendation applied, and of actual usage"
              : "Capacity, requests and usage across the cluster"
          }
          unit={unit}
          height={280}
          series={series}
          loading={cur.isLoading}
          lines={lines}
          fmtY={fmtY}
        />
      </div>
    </section>
  );
}
