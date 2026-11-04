import { fireEvent, render } from "@testing-library/react";
import { renderToStaticMarkup } from "react-dom/server";
import { beforeEach, describe, expect, it } from "vitest";
import type { Session } from "./api";
import { readSignOutIntent, type ShellLabels } from "./session";
import { Crumbs, Sidebar, Who } from "./shell";

/**
 * The shell holds the labels the boot views render without: `GET /api/v1/info` is answered only to a signed-in caller,
 * so every one of those screens shows this with nothing in it. The assertions say both what each label renders as and
 * which parts are allowed to be absent, so a screen that has been told nothing still shows the way between pages.
 */
const everyLabel = { cluster: "cluster-east", version: "0.7.0", prometheus: "prom.metrics.svc:9090" };
const none = (over: Partial<ShellLabels> = {}): ShellLabels => over;

describe("Sidebar", () => {
  it("is the shell a signed-in caller sees", () => {
    const html = renderToStaticMarkup(<Sidebar page="workloads" shell={everyLabel} />);
    expect(html).toContain("rz-side");
    expect(html).toContain("resourcerer");
    // Each of the three labels `info` supplies, where the page reads it.
    expect(html).toContain('<div class="rz-dim rz-xs">Cluster</div>');
    expect(html).toContain("<b>cluster-east</b>");
    expect(html).toContain('<div class="rz-dim rz-xs">0.7.0</div>');
    expect(html).toContain("Prometheus · prom.metrics.svc:9090");
    // And the way between the pages, with the current one marked.
    expect(html).toContain('<a href="#/workloads" class="on" aria-current="page">');
  });
  it("says so when the sign-in configuration could not be reread", () => {
    expect(
      renderToStaticMarkup(<Sidebar page="settings" shell={none()} configStale configError="expired" />),
    ).toContain('title="expired"');
    expect(renderToStaticMarkup(<Sidebar page="settings" shell={none()} configStale />)).toContain(
      "The sign-in configuration could not be reread.",
    );
  });
  it("renders no cluster block and no footer at all before `info` answers", () => {
    for (const shell of [none(), none({ cluster: "", version: "", prometheus: "" })]) {
      const html = renderToStaticMarkup(<Sidebar page="overview" shell={shell} />);
      expect(html).not.toContain("rz-cluster");
      expect(html).not.toContain("rz-sidefoot");
      expect(html).not.toContain("Cluster");
      expect(html).not.toContain("Prometheus ·");
      // What is left is the point of the screen: the name and the way between the three pages.
      expect(html).toContain("resourcerer");
      expect(html).toContain('href="#/overview"');
      expect(html).toContain('href="#/workloads"');
      expect(html).toContain('href="#/settings"');
    }
  });
  it("keeps the footer for the notice alone, with nothing else to say", () => {
    const html = renderToStaticMarkup(<Sidebar page="overview" shell={none()} configStale />);
    expect(html).toContain("rz-sidefoot");
    expect(html).not.toContain("Prometheus ·");
    expect(html).not.toContain("0.7.0");
  });
  it("marks the page the caller is on, and only that one", () => {
    const html = renderToStaticMarkup(<Sidebar page="settings" shell={everyLabel} />);
    expect(html).toContain('<a href="#/settings" class="on" aria-current="page">');
    expect(html.match(/aria-current="page"/g)).toHaveLength(1);
    expect(html).toContain('<a href="#/workloads" class="">');
  });
});

describe("Crumbs", () => {
  it("names the cluster the page belongs to", () => {
    const parts = ["w", "demo", "Deployment", "api"];
    const html = renderToStaticMarkup(<Crumbs page="workloads" parts={parts} shell={everyLabel} />);
    expect(html).toBe(
      '<div class="rz-crumbs rz-dim">cluster-east <span>/</span><a href="#/workloads">Workloads</a> <span>/</span> <b>demo/api</b></div>',
    );
  });
  it("leaves the cluster out of a page's crumbs until `info` says it", () => {
    const html = renderToStaticMarkup(<Crumbs page="overview" parts={["overview"]} shell={none()} />);
    expect(html).toBe('<div class="rz-crumbs rz-dim"><b>Overview</b></div>');
  });
  it("reads a workload as a workload with no cluster to put it in", () => {
    const html = renderToStaticMarkup(
      <Crumbs page="workloads" parts={["w", "demo", "Deployment", "api"]} shell={none()} />,
    );
    expect(html).toContain('<a href="#/workloads">Workloads</a>');
    expect(html).toContain("<b>demo/api</b>");
    expect(html).not.toContain("cluster-east");
  });
  it("capitalises a page name the address spells in lower case", () => {
    expect(renderToStaticMarkup(<Crumbs page="settings" parts={["settings"]} shell={everyLabel} />)).toContain(
      "<b>Settings</b>",
    );
  });
});

/**
 * Who the header says is calling, and the one control that ends it. Rendering this proves what it shows; the promise it
 * carries is about a moment after the click, when the cookie is gone and every other tab of this browser is still
 * polling. So the click is the test.
 */
describe("Who", () => {
  const editor: Session = {
    authenticated: true,
    user: "ada",
    groups: ["sre"],
    role: "editor",
    strategy: "oidc",
    impersonating: false,
    configStale: false,
  };
  const anonymous: Session = {
    authenticated: true,
    user: "anonymous",
    groups: [],
    role: "editor",
    strategy: "anonymous",
    impersonating: false,
    configStale: false,
  };
  const nobody: Session = { authenticated: false, strategy: "oidc" };

  beforeEach(() => localStorage.clear());

  it("names the caller, their role, and what the install lets them do", () => {
    const html = renderToStaticMarkup(<Who session={editor} readOnly />);
    expect(html).toContain("ada · editor · read-only");
    expect(html).toContain("Signed in as ada (editor) · sre");
    expect(html).toContain('href="/auth/logout"');
  });
  // The tab that was signed out from is not the one that reads the answer: between the click and the page the sign-out
  // lands on, the other tabs take the 401 this caused and would spend it signing the whole browser back in. Nothing
  // separates that window but the order of two statements in one handler, so this is what says the order is the point.
  it("writes down the sign-out on the way out of the click", () => {
    const { getByTitle } = render(<Who session={editor} readOnly={false} />);
    // jsdom cannot follow the link and would complain about it; holding the default action back changes nothing about
    // the handler under test, which is the part that runs before the browser goes anywhere.
    const hold = (e: Event) => e.preventDefault();
    document.addEventListener("click", hold, true);
    try {
      fireEvent.click(getByTitle("Sign out"));
      expect(readSignOutIntent()).toBe(true);
    } finally {
      document.removeEventListener("click", hold, true);
    }
  });
  it("names an anonymous install as one, with no session to end", () => {
    const html = renderToStaticMarkup(<Who session={anonymous} readOnly />);
    expect(html).toContain("anonymous · read-only");
    expect(html).not.toContain("/auth/logout");
  });
  it("says nothing about the caller before `/me` has answered", () => {
    for (const session of [undefined, nobody]) {
      expect(renderToStaticMarkup(<Who session={session} readOnly={false} />)).toBe("");
    }
  });
});
