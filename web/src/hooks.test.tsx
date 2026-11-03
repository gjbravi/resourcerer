import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { renderHook, waitFor } from "@testing-library/react";
import type { ReactNode } from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { Me, Unauthenticated } from "./api";
import { useSession } from "./hooks";
import { loopWindow, readSignOutIntent, readSignRecord } from "./session";

const signedIn: Me = {
  authenticated: true,
  user: "eddie",
  groups: ["platform"],
  role: "editor",
  strategy: "oidc",
  impersonating: false,
  configStale: false,
};
const nobody: Unauthenticated = { authenticated: false, strategy: "oidc" };

/** One client per render, with no retries: these tests are about the first answer from `/me`. */
function Wrapper({ children }: { children: ReactNode }) {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return <QueryClientProvider client={client}>{children}</QueryClientProvider>;
}

function answer(body: unknown) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => ({ ok: true, status: 200, json: async () => body })),
  );
}

function location(pathname: string, search = "") {
  const assign = vi.fn();
  vi.stubGlobal("location", { pathname, search, hash: "", assign });
  return assign;
}

beforeEach(() => {
  sessionStorage.clear();
  // One browser per test: the sign-out intent is the one thing these tests leave behind for the next tab.
  localStorage.clear();
  answer(nobody);
  location("/");
});

// The page asks who is calling before it shows anything, so this is where an unauthenticated visitor is handed to
// the provider: the hook has to make the call `afterMe` decides on, not only own the query.
describe("useSession", () => {
  it("sends a caller `/me` does not know to the provider, keeping the page", async () => {
    const assign = location("/workloads/");
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(assign).toHaveBeenCalledWith("/auth/login?redirect=%2Fworkloads%2F");
    expect(readSignRecord().at).toBeTypeOf("number");
  });

  it("holds back the second hand-off a round trip of its own caused", async () => {
    const assign = location("/");
    sessionStorage.setItem("resourcerer.sign-in", String(Date.now()));
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(readSignRecord().stopped).toBeTypeOf("number");
    expect(assign).not.toHaveBeenCalled();
  });

  it("keeps a signed-in caller where they are, and stops holding a sign-in that aged out against the loop", async () => {
    const assign = location("/");
    sessionStorage.setItem("resourcerer.sign-in", String(Date.now() - loopWindow));
    answer(signedIn);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(true));
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord().at).toBeUndefined();
    expect(readSignRecord().stopped).toBeUndefined();
  });

  // A round trip that came back with a session and loses it a second later is the loop this breaker exists for, so
  // the answer that arrived inside the window does not retire the attempt that made it.
  it("holds on to an attempt that only just came back, even though `/me` says it worked", async () => {
    const assign = location("/");
    sessionStorage.setItem("resourcerer.sign-in", String(Date.now()));
    answer(signedIn);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(true));
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord().at).toBeTypeOf("number");
  });

  it("leaves the signed-out screen alone, which is where the sign-out link lands", async () => {
    const assign = location("/", "?signed-out");
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord()).toEqual({ at: undefined, stopped: undefined });
  });

  // What the marker cannot reach on its own: a second tab whose address says nothing about a sign-out, polling in the
  // same browser that was just told to stay out.
  it("holds a tab whose address carries no marker, once the browser was signed out", async () => {
    localStorage.setItem("resourcerer.signed-out", String(Date.now()));
    const assign = location("/", "");
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord().at).toBeUndefined();
  });

  // The tab the sign-out landed on is the one that tells the rest of the browser, so this is where that gets written.
  it("tells the rest of the browser when this tab is where a sign-out left it", async () => {
    location("/", "?signed-out");
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(readSignOutIntent()).toBe(true);
  });
});

// The marker is a claim about the tab's session, and a tab that has one is no longer where the sign-out left anyone:
// left in the address, the next expiry reads as a sign-out nobody asked for. `/me` is what settles it.
describe("the address `/me` settles", () => {
  it("loses the marker as soon as a session answers", async () => {
    const assign = location("/", "?signed-out");
    const replace = vi.spyOn(window.history, "replaceState");
    answer(signedIn);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(true));
    expect(replace).toHaveBeenCalledWith("", "", "/");
    expect(assign).not.toHaveBeenCalled();
    replace.mockRestore();
  });
  it("keeps the rest of the address while it is at it", async () => {
    location("/", "?a=1&signed-out");
    const replace = vi.spyOn(window.history, "replaceState");
    answer(signedIn);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(true));
    expect(replace).toHaveBeenCalledWith("", "", "/?a=1");
    replace.mockRestore();
  });
  it("leaves the address of a browser that is not signed in exactly as it is", async () => {
    location("/", "?signed-out");
    const replace = vi.spyOn(window.history, "replaceState");
    answer(nobody);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(false));
    expect(replace).not.toHaveBeenCalled();
    replace.mockRestore();
  });
  it("has nothing to take out when the address never carried the marker", async () => {
    location("/workloads/", "");
    const replace = vi.spyOn(window.history, "replaceState");
    answer(signedIn);
    const { result } = renderHook(() => useSession(), { wrapper: Wrapper });
    await waitFor(() => expect(result.current.data?.authenticated).toBe(true));
    expect(replace).not.toHaveBeenCalled();
    replace.mockRestore();
  });
});
