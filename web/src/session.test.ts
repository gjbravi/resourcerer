import { describe, expect, it, vi } from "vitest";
import type { Session } from "./api";
import {
  afterMe,
  bootState,
  clearSignRecord,
  loopWindow,
  onAuthPath,
  pagePath,
  readSignRecord,
  sendToSignIn,
  signInURL,
  signMove,
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
  it("drops the record when a session did come back, so a later sign-out is not held against it", () => {
    const s = store();
    afterMe(nobody, page("/").page, s, 1_000);
    afterMe(signedIn, page("/").page, s, 2_000);
    expect(readSignRecord(s)).toEqual({ at: undefined, stopped: undefined });
    const again = page("/");
    afterMe(nobody, again.page, s, 2_500);
    expect(again.assign).toHaveBeenCalledTimes(1);
  });
});

describe("bootState", () => {
  const editor = session({ authenticated: true, role: "editor", strategy: "oidc", user: "ada", groups: ["sre"] });
  const viewer = session({ authenticated: true, role: "viewer", strategy: "oidc", user: "vera", groups: [] });
  const noRole = session({ authenticated: true, role: "none", strategy: "oidc", user: "dana", groups: ["dev"] });
  const anonymous = session({ authenticated: true, role: "editor", strategy: "anonymous", user: "anonymous" });
  const signedOut = session({ authenticated: false, strategy: "oidc" });
  const down = new Error("Failed to fetch");

  for (const [name, session_, record, view] of [
    ["nothing has answered yet", { data: undefined, error: null }, {}, "loading"],
    ["the answer is an editor", { data: editor, error: null }, {}, "app"],
    ["the answer is the anonymous editor", { data: anonymous, error: null }, {}, "app"],
    ["the answer is a viewer", { data: viewer, error: null }, {}, "app"],
    ["signed in without a role", { data: noRole, error: null }, {}, "no-access"],
    ["nobody is signed in", { data: signedOut, error: null }, {}, "signing-in"],
    [
      "nobody is signed in and the last attempt found no cookie",
      { data: signedOut, error: null },
      { at: 9, stopped: 9 },
      "cookie-dropped",
    ],
    ["a session that outlived a stopped attempt", { data: editor, error: null }, { at: 9, stopped: 9 }, "app"],
    // I3: without this row the page sits on "Loading…" forever, on every install including anonymous ones.
    ["the only answer was an error", { data: undefined, error: down }, {}, "error"],
    ["an error after a good answer keeps the page", { data: editor, error: down }, {}, "app"],
    ["an error while the browser is with the provider", { data: signedOut, error: down }, {}, "signing-in"],
  ] as [string, { data: Session | undefined; error: unknown }, ReturnType<typeof readSignRecord>, string][]) {
    it(name, () => expect(bootState(session_, record)).toMatchObject({ view }));
  }

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
