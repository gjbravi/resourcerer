import { container, workload } from "./test/fixtures";
import { change, containerSizing, format, parseQuantity, relativeTime, sizing } from "./format";

describe("parseQuantity", () => {
  it.each([
    ["1", 1],
    ["126m", 0.126],
    ["74Mi", 74 * 2 ** 20],
    ["2Gi", 2 * 2 ** 30],
    ["77594624", 77594624],
    ["100M", 100e6],
  ])("%s", (q, want) => expect(parseQuantity(q)).toBeCloseTo(want));

  it("rejects invalid input", () => {
    expect(parseQuantity("abc")).toBeUndefined();
    expect(parseQuantity(undefined)).toBeUndefined();
  });
});

describe("format", () => {
  it("shows CPU in millicores and memory in Mi or Gi", () => {
    expect(format("cpu", "1")).toBe("1000m");
    expect(format("memory", "77594624")).toBe("74Mi");
    expect(format("memory", "1536Mi")).toBe("1.5Gi");
    expect(format("cpu", undefined)).toBe("–");
  });
});

describe("sizing", () => {
  it("over-provisioned when requests exceed the target", () => {
    expect(sizing(workload())).toBe("over");
    expect(change("1", "126m")).toBe(-87);
  });

  it("under-provisioning wins", () => {
    const c = container({ requests: { cpu: "50m", memory: "1Gi" }, target: { cpu: "517m", memory: "132Mi" } });
    expect(containerSizing(c)).toBe("under");
  });

  it("missing requests count as under-provisioned", () => {
    expect(containerSizing(container({ requests: {}, target: { cpu: "100m" } }))).toBe("under");
  });

  it("within 10% is right-sized, no target is unknown", () => {
    expect(
      containerSizing(
        container({ requests: { cpu: "200m", memory: "128Mi" }, target: { cpu: "190m", memory: "135Mi" } }),
      ),
    ).toBe("fit");
    expect(sizing(workload({ containers: [container({ target: undefined })] }))).toBe("unknown");
  });
});

describe("relativeTime", () => {
  it("formats past and future", () => {
    const now = Date.parse("2026-10-23T12:00:00Z");
    expect(relativeTime("2026-10-23T11:59:30Z", now)).toBe("30s ago");
    expect(relativeTime("2026-10-23T14:00:00Z", now)).toBe("in 2h");
  });
});
