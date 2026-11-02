import { describe, expect, it, vi } from "vitest";
import type { Session } from "./api";
import { onAuthPath, pagePath, sendToSignIn, signInURL, writeLock } from "./session";

const page = (pathname: string, search = "", hash = ""): { assign: ReturnType<typeof vi.fn>; page: PageLike } => {
  const assign = vi.fn();
  return { assign, page: { pathname, search, hash, assign } };
};
type PageLike = { pathname: string; search: string; hash: string; assign(url: string): void };

const session = (over: Partial<Session>): Session => ({ ...over }) as Session;

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
    sendToSignIn(p);
    expect(assign).toHaveBeenCalledWith("/auth/login?redirect=%2F%23%2Fworkloads");
  });
  it("does nothing where that would loop", () => {
    const { assign, page: p } = page("/auth/callback", "?error=access_denied");
    sendToSignIn(p);
    expect(assign).not.toHaveBeenCalled();
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
