import { container, workload } from "../test/fixtures";
import { matches } from "./WorkloadsPage";

describe("filters", () => {
  const under = workload({
    name: "worker",
    containers: [container({ requests: { cpu: "50m" }, target: { cpu: "517m" } })],
  });
  const learning = workload({ name: "new", status: "LowConfidence" });
  const disabled = workload({
    name: "off",
    mode: "Disabled",
    status: "None",
    containers: [container({ target: undefined })],
  });
  const rows = [workload(), under, learning, disabled];
  const names = (f: Parameters<typeof matches>[1]) => rows.filter((w) => matches(w, f)).map((w) => w.name);

  it("selects by sizing, status and mode", () => {
    expect(names("under")).toEqual(["worker"]);
    expect(names("over")).toEqual(["api", "new"]);
    expect(names("learning")).toEqual(["new"]);
    expect(names("disabled")).toEqual(["off"]);
    expect(names("all")).toHaveLength(4);
  });
});
