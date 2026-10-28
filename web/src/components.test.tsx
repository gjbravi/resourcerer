import { render, screen } from "@testing-library/react";
import { PiggyBank } from "lucide-react";
import { CompareBars, Figure, ImpactTile } from "./components";
import { fmtCores } from "./format";

const fills = (c: HTMLElement) => Array.from(c.querySelectorAll<HTMLElement>(".rz-cmp-fill"));

describe("CompareBars", () => {
  it("scales both bars to the larger value", () => {
    const { container } = render(<CompareBars current={0.9} recommended={0.675} format={fmtCores} />);
    const [cur, rec] = fills(container);
    expect(cur.style.width).toBe("100%");
    expect(rec.style.width).toBe("75%");
    expect(rec).not.toHaveClass("up");
    expect(screen.getByRole("img")).toHaveAttribute("aria-label", "Current 0.900 Cores, recommended 0.675 Cores");
  });

  it("marks an increase with the bad tone", () => {
    const { container } = render(<CompareBars current={1} recommended={2} format={fmtCores} />);
    const [cur, rec] = fills(container);
    expect(cur.style.width).toBe("50%");
    expect(rec.style.width).toBe("100%");
    expect(rec).toHaveClass("up");
  });

  it("keeps widths valid when values are zero or missing", () => {
    const { container } = render(<CompareBars current={0} recommended={undefined} format={fmtCores} />);
    for (const f of fills(container)) expect(f.style.width).toBe("0%");
    expect(container.innerHTML).not.toContain("NaN");
  });
});

describe("Figure", () => {
  it("splits the unit from the number", () => {
    const { container } = render(<Figure text="0.675 Cores" />);
    expect(container.querySelector(".unit")).toHaveTextContent("Cores");
    expect(container.textContent).toBe("0.675Cores");
  });
});

describe("ImpactTile", () => {
  it("applies the tone to the icon chip", () => {
    const { container, rerender } = render(
      <ImpactTile icon={PiggyBank} label="Monthly cost">
        $1
      </ImpactTile>,
    );
    expect(container.querySelector(".rz-chip")).toHaveClass("t-save");
    rerender(
      <ImpactTile icon={PiggyBank} label="Monthly cost" tone="bad">
        $2
      </ImpactTile>,
    );
    expect(container.querySelector(".rz-chip")).toHaveClass("t-bad");
  });
});
