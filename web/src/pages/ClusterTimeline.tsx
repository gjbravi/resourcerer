import { LineChart } from "lucide-react";
import { useState } from "react";
import type { Point, Range, Resource, Series } from "../api";
import { at, TimeChart, type Line } from "../chart";
import { ErrorBox } from "../components";
import { fmtMem, Gi, isoSeconds } from "../format";
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

export function ClusterTimeline() {
  const [range, setRange] = useState<Range>("learning");
  const [mode, setMode] = useState<Mode>("cpu");
  const cpu = useClusterSeries("cpu", range);
  const mem = useClusterSeries("memory", range);
  const info = useInfo().data;
  const rates = info?.cost ?? { cpuCoreHour: 0, memoryGiBHour: 0 };
  const learning = isoSeconds(info?.defaults.find((x) => x.key === "learning-period")?.value) || 7 * 86400;
  const cur = mode === "memory" ? mem : cpu;
  const l = (s: Series | undefined, k: string) => s?.lines[k] ?? [];

  let series: Series | undefined = cur.data;
  let lines: Line[];
  let fmtY: (v: number) => string;
  let fmtTip: (v: number) => string;
  let unit: string;
  if (mode === "cost") {
    const requested = costLine(l(cpu.data, "requested"), l(mem.data, "requested"), rates);
    const withRecs = costLine(l(cpu.data, "withRecommendations"), l(mem.data, "withRecommendations"), rates);
    const usage = costLine(l(cpu.data, "usage"), l(mem.data, "usage"), rates);
    series = cpu.data && mem.data ? { ...cpu.data, bands: [], lines: {} } : undefined;
    lines = [
      { name: "Requested cost", color: "var(--c-cur)", points: requested, shape: "smooth", width: 2 },
      {
        name: "With recommendations",
        color: "var(--c-rec)",
        points: withRecs,
        shape: "smooth",
      },
      { name: "Usage cost", color: "var(--c-usage)", points: usage, shape: "smooth", width: 2 },
    ];
    fmtY = (v) => `$${v < 10 ? v.toFixed(2) : v.toFixed(0)}`;
    fmtTip = (v) => `$${v.toFixed(2)}/h`;
    unit = "USD / hour";
  } else {
    const k = mode === "cpu" ? 1 : Gi;
    lines = [
      { name: "Allocatable", color: "var(--c-alloc)", points: l(cur.data, "allocatable"), shape: "smooth", width: 2 },
      { name: "Requested", color: "var(--c-cur)", points: l(cur.data, "requested"), shape: "smooth", width: 2 },
      {
        name: "With recommendations",
        color: "var(--c-rec)",
        points: l(cur.data, "withRecommendations"),
        shape: "smooth",
        dash: "6 4",
      },
      { name: "Usage", color: "var(--c-usage)", points: l(cur.data, "usage"), shape: "smooth", width: 2 },
    ];
    fmtY = (v) => (v / k).toFixed(v / k < 10 ? 1 : 0);
    fmtTip = mode === "cpu" ? (v) => `${v.toFixed(2)} cores` : fmtMem;
    unit = mode === "cpu" ? "Cores" : "GiB";
  }

  return (
    <section className="rz-card">
      <div className="rz-cardhd">
        <h2>
          <LineChart size={17} className="t-accent" /> Cluster timeline
        </h2>
        <div className="rz-ctl">
          <Zoom value={range} onChange={setRange} learning={learning} />
          <span className="rz-seg">
            {MODES.map(([m, label]) => (
              <button key={m} aria-pressed={mode === m} onClick={() => setMode(m)}>
                {label}
              </button>
            ))}
          </span>
        </div>
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
          fmtTip={fmtTip}
        />
      </div>
    </section>
  );
}
