import { ago, fmtCpu, fmtMem, pct, quantity, costSource } from "./format";

describe("format", () => {
  it("keeps CPU in millicores and memory in MiB/GiB", () => {
    expect(fmtCpu(0.126)).toBe("126m");
    expect(fmtCpu(2)).toBe("2000m");
    expect(fmtCpu(undefined)).toBe("–");
    expect(fmtMem(74 * 2 ** 20)).toBe("74MiB");
    expect(fmtMem(1.5 * 2 ** 30)).toBe("1.50GiB");
  });

  it("writes Kubernetes quantities", () => {
    expect(quantity("cpu", 0.105)).toBe("105m");
    expect(quantity("memory", 115 * 2 ** 20)).toBe("115Mi");
    expect(quantity("memory", undefined)).toBeUndefined();
  });

  it("computes percent change and relative time", () => {
    expect(pct(1, 0.126)).toBe(-87);
    expect(pct(0, 1)).toBeNull();
    expect(ago(new Date(Date.now() - 90_000).toISOString())).toBe("2m ago");
  });
});

describe("costSource", () => {
  it("names the source and says when it is degraded", () => {
    const k = { source: "karpenter", spot: 26, onDemand: 1 };
    expect(costSource({ ...k, state: "ok" })).toBe("Karpenter node prices · 26 spot, 1 on-demand nodes");
    expect(costSource({ ...k, state: "stale" })).toBe(
      "Karpenter node prices · 26 spot, 1 on-demand nodes · last calibration, refresh failing",
    );
    expect(costSource({ source: "static", state: "ok" })).toBe("configured rates");
    expect(costSource({ source: "static", state: "fallback" })).toBe("configured rates (no Karpenter in this cluster)");
    expect(costSource({ source: "static", state: "error" })).toBe("configured rates · Karpenter prices unavailable");
  });
});
