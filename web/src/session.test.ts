import { describe, expect, it, vi } from "vitest";
import type { Info, Session } from "./api";
import {
  afterMe,
  bootState,
  clearSignOutIntent,
  clearSignRecord,
  countRefusal,
  holdingAtSignIn,
  loopWindow,
  noteRefusal,
  onAuthPath,
  pagePath,
  readSignOutIntent,
  readSignRecord,
  refusedTwice,
  sendToSignIn,
  setSignOutIntent,
  shellLabels,
  signInURL,
  signedOut,
  signMove,
  staleMarker,
  withoutSignedOut,
  writeLock,
  type Store,
} from "./session";

const page = (pathname: string, search = "", hash = ""): { assign: ReturnType<typeof vi.fn>; page: PageLike } => {
  const assign = vi.fn();
  return { assign, page: { pathname, search, hash, assign } };
};
type PageLike = { pathname: string; search: string; hash: string; assign(url: string): void };

const session = (over: Partial<Session>): Session => ({ ...over }) as Session;

/** `sessionStorage` without the browser, keeping its contents where a test can read them back. */
function store(): Store & { values: Record<string, string> } {
  const values: Record<string, string> = {};
  return {
    values,
    getItem: (key) => values[key] ?? null,
    setItem: (key, value) => {
      values[key] = value;
    },
    removeItem: (key) => {
      delete values[key];
    },
  };
}

describe("signInURL", () => {
  it("keeps the page the caller was on, hash included", () => {
    expect(signInURL({ pathname: "/", search: "", hash: "#/w/default/Deployment/api/settings", assign: vi.fn() })).toBe(
      "/auth/login?redirect=%2F%23%2Fw%2Fdefault%2FDeployment%2Fapi%2Fsettings",
    );
  });
  it("keeps the query string too", () => {
    expect(signInURL({ pathname: "/x", search: "?a=1", hash: "", assign: vi.fn() })).toBe(
      "/auth/login?redirect=%2Fx%3Fa%3D1",
    );
  });
  it("sends nothing that already belongs to the sign-in flow", () => {
    for (const pathname of ["/auth/login", "/auth/callback", "/auth/logout"]) {
      expect(signInURL({ pathname, search: "", hash: "", assign: vi.fn() })).toBeNull();
    }
  });
});

describe("pagePath", () => {
  it("is the whole address a redirect has to bring back", () => {
    expect(pagePath({ pathname: "/a", search: "?b=2", hash: "#/c", assign: vi.fn() })).toBe("/a?b=2#/c");
  });
});

describe("onAuthPath", () => {
  it("only matches paths the server serves itself", () => {
    expect(onAuthPath("/auth/login")).toBe(true);
    expect(onAuthPath("/authorised")).toBe(false);
    expect(onAuthPath("/api/v1/me")).toBe(false);
    expect(onAuthPath("/")).toBe(false);
  });
});

describe("sendToSignIn", () => {
  it("hands the browser to the provider", () => {
    const { assign, page: p } = page("/", "", "#/workloads");
    sendToSignIn(p, store());
    expect(assign).toHaveBeenCalledWith("/auth/login?redirect=%2F%23%2Fworkloads");
  });
  it("does nothing where that would loop", () => {
    const { assign, page: p } = page("/auth/callback", "?error=access_denied");
    const s = store();
    sendToSignIn(p, s);
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
  });

  // The browser can undo a sign-in: cookies blocked for this site but not for the provider, or a cookie the browser
  // drops for size. Every turn of that loop is a full authorization at the provider, so the second one is refused.
  it("records the attempt it makes", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 1_000);
    expect(readSignRecord(s)).toEqual({ at: 1_000, stopped: undefined });
  });
  it("refuses the second attempt inside the window and marks why", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 1_000);
    const second = page("/");
    sendToSignIn(second.page, s, 1_000 + loopWindow - 1);
    expect(second.assign).not.toHaveBeenCalled();
    expect(readSignRecord(s)).toEqual({ at: 1_000, stopped: 1_000 + loopWindow - 1 });
  });
  // The run of refusals is a different animal from a round trip that came back empty: it is one cookie answered two
  // ways by the server, and once it adds up another authorization cannot fix it, so the window ageing out does not
  // earn one. The card with the retry in it is where this ends.
  it("stops for good once the refusals add up, window or no window", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 1_000, store());
    noteRefusal(s, 1_200);
    noteRefusal(s, 1_400);
    const late = page("/");
    sendToSignIn(late.page, s, 1_000 + loopWindow, store());
    expect(late.assign).not.toHaveBeenCalled();
    expect(readSignRecord(s)).toMatchObject({ at: 1_000, stopped: 1_000 + loopWindow, refusals: 2 });
  });
  it("tries again once the window has passed, and forgets it stopped", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 1_000);
    sendToSignIn(page("/").page, s, 1_000 + 1_000);
    const later = page("/");
    sendToSignIn(later.page, s, 1_000 + loopWindow);
    expect(later.assign).toHaveBeenCalledTimes(1);
    expect(readSignRecord(s)).toEqual({ at: 1_000 + loopWindow, stopped: undefined });
  });
  it("still signs in when the browser will not give it a record", () => {
    vi.stubGlobal("sessionStorage", blockedStorage);
    try {
      const { assign, page: p } = page("/");
      sendToSignIn(p);
      expect(assign).toHaveBeenCalledTimes(1);
      expect(readSignRecord()).toEqual({ at: undefined, stopped: undefined });
      expect(() => clearSignRecord()).not.toThrow();
    } finally {
      vi.unstubAllGlobals();
    }
  });
  // Sign-out leaves the browser on an address that says it asked for this, and an unauthenticated `/me` there is the
  // point of the page rather than a reason to visit the provider again.
  it("leaves the signed-out screen signed out", () => {
    const s = store();
    const { assign, page: p } = page("/", "?signed-out");
    sendToSignIn(p, s, 1_000);
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
  });
});

describe("signedOut", () => {
  for (const [search, want] of [
    ["?signed-out", true],
    ["?signed-out=1", true],
    ["?a=1&signed-out", true],
    ["", false],
    ["?x=signed-out", false],
    ["?signedout", false],
  ] as [string, boolean][]) {
    it(`${search || "no query"}: ${want}`, () => expect(signedOut(search)).toBe(want));
  }
});

/** What a browser with storage turned off for the site hands back: a SecurityError on every touch. */
const blockedStorage = {
  getItem: () => {
    throw new Error("storage is blocked");
  },
  setItem: () => {
    throw new Error("storage is blocked");
  },
  removeItem: () => {
    throw new Error("storage is blocked");
  },
};

/**
 * One browser, several tabs, one cookie. The marker belongs to a single tab's address, so it cannot be what keeps the
 * others from signing the whole browser back in; these tests share one store the way tabs share `localStorage`.
 */
describe("the sign-out intent", () => {
  const signedIn = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const nobody = session({ authenticated: false, strategy: "oidc" });

  it("is left by the tab a sign-out landed on", () => {
    const browser = store();
    afterMe(nobody, page("/", "?signed-out").page, store(), 1_000, browser);
    expect(readSignOutIntent(browser)).toBe(true);
  });
  it("is not left by a session, nor by a page nobody signed out", () => {
    const browser = store();
    afterMe(signedIn, page("/", "?signed-out").page, store(), 1_000, browser);
    afterMe(nobody, page("/").page, store(), 1_001, browser);
    expect(readSignOutIntent(browser)).toBe(false);
  });
  // The bug the marker alone leaves behind: the other tab polls, takes the same 401, and its address says nothing about
  // a sign-out, so it hands the browser to a provider that still has its own session open.
  it("holds a tab that carries no marker of its own", () => {
    const browser = store();
    afterMe(nobody, page("/", "?signed-out").page, store(), 1_000, browser);
    const other = page("/", "", "#/workloads");
    const record = store();
    sendToSignIn(other.page, record, 2_000, browser);
    expect(other.assign).not.toHaveBeenCalled();
    expect(readSignRecord(record).at).toBeUndefined();
    expect(bootState({ data: nobody, error: null }, {}, "", readSignOutIntent(browser))).toEqual({
      view: "signed-out",
    });
  });
  it("lifts only when someone asks for a session again", () => {
    const browser = store();
    afterMe(nobody, page("/", "?signed-out").page, store(), 1_000, browser);
    sendToSignIn(page("/").page, store(), 2_000, browser);
    clearSignOutIntent(browser);
    expect(readSignOutIntent(browser)).toBe(false);
    const back = page("/");
    sendToSignIn(back.page, store(), 3_000, browser);
    expect(back.assign).toHaveBeenCalledTimes(1);
  });
  it("is absent rather than fatal where the browser will not hold it", () => {
    vi.stubGlobal("localStorage", blockedStorage);
    try {
      expect(readSignOutIntent()).toBe(false);
      expect(() => setSignOutIntent()).not.toThrow();
      expect(() => clearSignOutIntent()).not.toThrow();
      // With nothing shared to read, the tab's own address is still enough to keep it signed out.
      const { assign, page: p } = page("/", "?signed-out");
      sendToSignIn(p, store());
      expect(assign).not.toHaveBeenCalled();
    } finally {
      vi.unstubAllGlobals();
    }
  });
});

describe("holdingAtSignIn", () => {
  it("reads either half of the answer, and neither once the intent is lifted", () => {
    const browser = store();
    expect(holdingAtSignIn("", browser)).toBe(false);
    expect(holdingAtSignIn("?signed-out", browser)).toBe(true);
    setSignOutIntent(browser, 1_000);
    expect(holdingAtSignIn("", browser)).toBe(true);
    expect(holdingAtSignIn("?signed-out", browser)).toBe(true);
    clearSignOutIntent(browser);
    expect(holdingAtSignIn("", browser)).toBe(false);
    expect(holdingAtSignIn("?x=signed-out", browser)).toBe(false);
  });
});

describe("countRefusal", () => {
  for (const [name, now, record, want] of [
    ["a first refusal", 1_000, {}, 1],
    ["a second inside the window", 1_000 + loopWindow - 1, { refused: 1_000, refusals: 1 }, 2],
    ["a third, still inside it", 1_000 + loopWindow - 1, { refused: 1_000, refusals: 2 }, 3],
    ["one after the window starts the run over", 1_000 + loopWindow, { refused: 1_000, refusals: 2 }, 1],
    ["a time with no count behind it", 1_000, { refused: 900 }, 1],
    ["a count with no time behind it", 1_000, { refusals: 4 }, 1],
  ] as [string, number, ReturnType<typeof readSignRecord>, number][]) {
    it(`${name}: ${want}`, () => expect(countRefusal(now, record)).toBe(want));
  }
});

// A caller `/me` signed in and the API refused is the same cookie answered two ways. One is worth another request;
// a run of them is a disagreement about who this browser is, and the page has to say so rather than keep asking.
describe("noteRefusal", () => {
  it("keeps the time and the count of a run", () => {
    const s = store();
    noteRefusal(s, 1_000);
    expect(readSignRecord(s)).toMatchObject({ refused: 1_000, refusals: 1 });
    expect(refusedTwice(readSignRecord(s))).toBe(false);
    noteRefusal(s, 1_500);
    expect(readSignRecord(s)).toMatchObject({ refused: 1_500, refusals: 2 });
    expect(refusedTwice(readSignRecord(s))).toBe(true);
  });
  it("is over when the card's own retry clears the record, which is the action that ends a run", () => {
    const s = store();
    noteRefusal(s, 1_000);
    noteRefusal(s, 1_500);
    clearSignRecord(s);
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
    expect(refusedTwice(readSignRecord(s))).toBe(false);
    noteRefusal(s, 2_000);
    expect(readSignRecord(s)).toMatchObject({ refusals: 1 });
  });
  it("counts nothing where the browser keeps no record", () => {
    vi.stubGlobal("sessionStorage", blockedStorage);
    try {
      expect(() => noteRefusal()).not.toThrow();
      expect(refusedTwice(readSignRecord())).toBe(false);
    } finally {
      vi.unstubAllGlobals();
    }
  });
});

describe("signMove", () => {
  for (const [name, now, record, want] of [
    ["never sent", 1_000, {}, "sign-in"],
    ["back inside the window", 1_000 + loopWindow - 1, { at: 1_000 }, "wait"],
    ["back after the window", 1_000 + loopWindow, { at: 1_000 }, "sign-in"],
    ["a stop with nothing behind it", 1_000, { stopped: 900 }, "sign-in"],
    ["a record written by hand is no attempt", 1_000, { at: undefined }, "sign-in"],
  ] as [string, number, ReturnType<typeof readSignRecord>, "sign-in" | "wait"][]) {
    it(`${name}: ${want}`, () => expect(signMove(now, record)).toBe(want));
  }
});

describe("readSignRecord", () => {
  it("reads back what sendToSignIn left", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 7);
    sendToSignIn(page("/").page, s, 9);
    expect(readSignRecord(s)).toEqual({ at: 7, stopped: 9 });
  });
  it("ignores a record it cannot read", () => {
    const s = store();
    s.setItem("resourcerer.sign-in", "yesterday");
    expect(readSignRecord(s).at).toBeUndefined();
  });
  it("clears on request", () => {
    const s = store();
    sendToSignIn(page("/").page, s, 7);
    clearSignRecord(s);
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
  });
});

describe("afterMe", () => {
  const signedIn = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const nobody = session({ authenticated: false, strategy: "oidc" });

  it("starts the sign-in for a caller who has none, which is what useSession asks it to do", () => {
    const s = store();
    const { assign, page: p } = page("/", "", "#/w/demo/Deployment/api/impact");
    afterMe(nobody, p, s, 1_000);
    expect(assign).toHaveBeenCalledWith("/auth/login?redirect=%2F%23%2Fw%2Fdemo%2FDeployment%2Fapi%2Fimpact");
    expect(readSignRecord(s).at).toBe(1_000);
  });
  it("stops once the round trip came back with no session cookie", () => {
    const s = store();
    afterMe(nobody, page("/").page, s, 1_000);
    const second = page("/");
    afterMe(nobody, second.page, s, 2_000);
    expect(second.assign).not.toHaveBeenCalled();
    expect(readSignRecord(s).stopped).toBe(2_000);
  });
  it("holds the record when a session answers inside the window, so the 401 after it is the second attempt", () => {
    const s = store();
    afterMe(nobody, page("/").page, s, 1_000);
    afterMe(signedIn, page("/").page, s, 2_000);
    expect(readSignRecord(s).at).toBe(1_000);
    const bounce = page("/");
    afterMe(nobody, bounce.page, s, 2_500);
    expect(bounce.assign).not.toHaveBeenCalled();
    expect(readSignRecord(s).stopped).toBe(2_500);
  });
  it("retires the record once the window has passed, so a session that stayed is not held against it", () => {
    const s = store();
    afterMe(nobody, page("/").page, s, 1_000);
    afterMe(signedIn, page("/").page, s, 1_000 + loopWindow);
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
    const again = page("/");
    afterMe(nobody, again.page, s, 1_000 + loopWindow + 1);
    expect(again.assign).toHaveBeenCalledTimes(1);
  });
  // `/me` is the half of a cookie disagreement that keeps answering, so it retires the sign-in attempt its answer
  // explains and leaves the count the card reads alone. Wiping it here would hide the explanation until the next 401.
  it("retires the attempt a session settled, but not a run of refusals", () => {
    const s = store();
    noteRefusal(s, 900);
    noteRefusal(s, 1_000);
    sendToSignIn(page("/").page, s, 1_500, store());
    afterMe(signedIn, page("/").page, s, 1_500 + loopWindow);
    expect(readSignRecord(s)).toMatchObject({ at: undefined, stopped: undefined, refused: 1_000, refusals: 2 });
    expect(bootState({ data: signedIn, error: null }, readSignRecord(s))).toMatchObject({ view: "cookie-dropped" });
  });
  it("leaves the signed-out screen to the page that explains it", () => {
    const s = store();
    const { assign, page: p } = page("/", "?signed-out");
    afterMe(nobody, p, s, 1_000);
    expect(assign).not.toHaveBeenCalled();
  });
});

describe("bootState", () => {
  const editor = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const viewer = session({ authenticated: true, role: "viewer", strategy: "oidc", user: "vera", groups: [] });
  const noRole = session({ authenticated: true, role: "none", strategy: "oidc", user: "dana", groups: ["dev"] });
  const anonymous = session({ authenticated: true, role: "editor", strategy: "anonymous", user: "anonymous" });
  const noSession = session({ authenticated: false, strategy: "oidc" });
  const down = new Error("Failed to fetch");

  for (const [name, session_, record, view] of [
    ["nothing has answered yet", { data: undefined, error: null }, {}, "loading"],
    ["the answer is an editor", { data: editor, error: null }, {}, "app"],
    ["the answer is the anonymous editor", { data: anonymous, error: null }, {}, "app"],
    ["the answer is a viewer", { data: viewer, error: null }, {}, "app"],
    ["signed in without a role", { data: noRole, error: null }, {}, "no-access"],
    ["nobody is signed in", { data: noSession, error: null }, {}, "signing-in"],
    [
      "nobody is signed in and the last attempt found no cookie",
      { data: noSession, error: null },
      { at: 9, stopped: 9 },
      "cookie-dropped",
    ],
    ["a session that outlived a stopped attempt", { data: editor, error: null }, { at: 9, stopped: 9 }, "app"],
    // Without this row the page sits on "Loading…" forever, on every install including anonymous ones.
    ["the only answer was an error", { data: undefined, error: down }, {}, "error"],
    ["an error after a good answer keeps the page", { data: editor, error: down }, {}, "app"],
    ["an error while the browser is with the provider", { data: noSession, error: down }, {}, "signing-in"],
  ] as [string, { data: Session | undefined; error: unknown }, ReturnType<typeof readSignRecord>, string][]) {
    it(name, () => expect(bootState(session_, record)).toMatchObject({ view }));
  }

  // The address sign-out leaves the browser on is the user's own answer to "who is signed in here", so it outranks
  // both ways of explaining the missing session, and loses to a session that came back anyway.
  it("shows the signed-out screen rather than signing straight back in", () => {
    expect(bootState({ data: noSession, error: null }, {}, "?signed-out")).toEqual({ view: "signed-out" });
  });
  it("shows it even where the last attempt had found no cookie", () => {
    expect(bootState({ data: noSession, error: null }, { at: 9, stopped: 9 }, "?signed-out")).toEqual({
      view: "signed-out",
    });
  });
  it("shows the app to a session that came back with the marker still in the address", () => {
    expect(bootState({ data: editor, error: null }, {}, "?signed-out")).toEqual({ view: "app", me: editor });
  });
  it("reads no marker in a query that only looks like one", () => {
    for (const search of ["", "?signed-off", "?x=signed-out", "?signedout"]) {
      expect(bootState({ data: noSession, error: null }, {}, search)).toMatchObject({ view: "signing-in" });
    }
  });

  it("hands the error to the page to name", () => {
    expect(bootState({ data: undefined, error: down })).toEqual({ view: "error", error: down });
  });
  it("keeps the caller to explain on the no-access page", () => {
    expect(bootState({ data: noRole, error: null })).toEqual({ view: "no-access", me: noRole });
  });
  it("hands the page the session it renders", () => {
    expect(bootState({ data: viewer, error: null })).toEqual({ view: "app", me: viewer });
  });
});

// Two decisions the boot screen makes with more than `/me` in front of it: what the rest of the browser was told, and
// whether the API has been refusing the session `/me` accepted.
describe("bootState with what the browser knows, and what the API answered", () => {
  const editor = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const noRole = session({ authenticated: true, role: "none", strategy: "oidc", user: "dana", groups: ["dev"] });
  const noSession = session({ authenticated: false, strategy: "oidc" });

  for (const [name, me, record, search, browserSignedOut, view] of [
    ["a tab held by the browser's answer", noSession, {}, "", true, "signed-out"],
    [
      "a tab held by it, even where its own last attempt found no cookie",
      noSession,
      { at: 9, stopped: 9 },
      "",
      true,
      "signed-out",
    ],
    ["a tab the browser never signed out", noSession, {}, "", false, "signing-in"],
    ["a session that came back anyway", editor, {}, "", true, "app"],
    ["a session the API keeps refusing", editor, { refused: 9, refusals: 2 }, "", false, "cookie-dropped"],
    ["one refusal, which is worth another request", editor, { refused: 9, refusals: 1 }, "", false, "app"],
    ["a refusal run under a session with no role", noRole, { refused: 9, refusals: 2 }, "", false, "no-access"],
    [
      "no session and a run of refusals",
      noSession,
      { refused: 9, refusals: 2, stopped: 9 },
      "",
      false,
      "cookie-dropped",
    ],
  ] as [string, Session, ReturnType<typeof readSignRecord>, string, boolean, string][]) {
    it(name, () =>
      expect(bootState({ data: me, error: null }, record, search, browserSignedOut)).toMatchObject({ view }),
    );
  }
});

describe("staleMarker", () => {
  const signedIn = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const nobody = session({ authenticated: false, strategy: "oidc" });

  for (const [name, me, search, want] of [
    ["a session with the marker left in the address", signedIn, "?signed-out", true],
    ["no session to make it stale", nobody, "?signed-out", false],
    ["a session and no marker", signedIn, "", false],
    ["nothing has answered yet", undefined, "?signed-out", false],
    ["a query that only looks like one", signedIn, "?x=signed-out", false],
  ] as [string, Session | undefined, string, boolean][]) {
    it(`${name}: ${want}`, () => expect(staleMarker(me, search)).toBe(want));
  }
});

describe("withoutSignedOut", () => {
  for (const [search, want] of [
    ["?signed-out", "/"],
    ["?signed-out=1", "/"],
    ["?a=1&signed-out", "/?a=1"],
    ["?signed-out&a=1&b=2", "/?a=1&b=2"],
    ["?x=signed-out", "/?x=signed-out"],
    ["", "/"],
  ] as [string, string][]) {
    it(`${search || "no query"} → ${want}`, () =>
      expect(withoutSignedOut({ pathname: "/", search, hash: "#/workloads" })).toBe(`${want}#/workloads`));
  }
  it("changes nothing else about the address", () => {
    expect(withoutSignedOut({ pathname: "/settings", search: "?signed-out", hash: "" })).toBe("/settings");
  });
});

describe("writeLock", () => {
  it("lets an editor write, with nothing to explain", () => {
    expect(writeLock(session({ authenticated: true, role: "editor", strategy: "oidc" }))).toEqual({
      write: true,
      readOnly: false,
    });
  });
  it("says why a viewer cannot", () => {
    expect(writeLock(session({ authenticated: true, role: "viewer", strategy: "oidc" }))).toMatchObject({
      write: false,
      reason: "View-only access",
    });
  });
  it("has nothing to say to someone who is not signed in yet", () => {
    expect(writeLock(session({ authenticated: false, strategy: "oidc" }))).toEqual({ write: false, readOnly: false });
    expect(writeLock(undefined)).toEqual({ write: false, readOnly: false });
  });
  it("says nothing to a caller signed in without a role: the page explains that one", () => {
    expect(writeLock(session({ authenticated: true, role: "none", strategy: "oidc" }))).toEqual({
      write: false,
      readOnly: false,
    });
  });
  it("keeps the anonymous editor it had before, which is what the UI was built around", () => {
    expect(writeLock(session({ authenticated: true, role: "editor", strategy: "anonymous" }))).toEqual({
      write: true,
      readOnly: false,
    });
  });
  it("comes after the install's own read-only mode, whoever the caller is", () => {
    const editor = session({ authenticated: true, role: "editor", strategy: "anonymous" });
    expect(writeLock(editor, true)).toEqual({ write: false, reason: "Read-only", readOnly: true });
    expect(writeLock(session({ authenticated: true, role: "viewer", strategy: "oidc" }), true)).toMatchObject({
      readOnly: true,
      reason: "Read-only",
    });
  });
});

describe("shellLabels", () => {
  const info = (over: { cluster?: string; version?: string; prometheus?: { url: string } }): Info =>
    ({ cluster: "", version: "", prometheus: { url: "" }, ...over }) as Info;

  for (const [name, given, labels] of [
    ["nothing has been answered for the shell yet", undefined, {}],
    [
      "everything the server filled in",
      info({ cluster: "cluster-east", version: "0.7.0", prometheus: { url: "http://127.0.0.1:9090" } }),
      { cluster: "cluster-east", version: "0.7.0", prometheus: "127.0.0.1:9090" },
    ],
    [
      "the scheme, which is not part of the label",
      info({ prometheus: { url: "https://prom.internal:9090" } }),
      {
        prometheus: "prom.internal:9090",
      },
    ],
    [
      "a host with a path keeps it",
      info({ cluster: "cluster-east", prometheus: { url: "http://prom.metrics.svc/prom" } }),
      { cluster: "cluster-east", prometheus: "prom.metrics.svc/prom" },
    ],
    ["a value the server left empty is not a label", info({ cluster: "", version: "", prometheus: { url: "" } }), {}],
    [
      "one of them missing leaves the other two",
      info({ cluster: "cluster-east", version: "", prometheus: { url: "http://prom:9090" } }),
      { cluster: "cluster-east", prometheus: "prom:9090" },
    ],
  ] as [string, Info | undefined, Record<string, string>][]) {
    it(name, () => expect(shellLabels(given)).toEqual(labels));
  }

  // Every one of these is an answer the shell gets only once someone is signed in, and the screens before that are
  // the first thing a new install shows: an ellipsis where the cluster belongs reads as a broken install.
  it("gives the shell nothing to render rather than a placeholder", () => {
    for (const given of [undefined, info({ cluster: "", version: "", prometheus: { url: "" } })]) {
      expect(Object.values(shellLabels(given))).toEqual([]);
    }
  });
});
