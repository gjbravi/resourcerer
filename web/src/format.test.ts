import { ago, fmtCpu, fmtMem, pct, quantity } from "./format";

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
