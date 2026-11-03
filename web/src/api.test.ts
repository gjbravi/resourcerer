import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { ApiError, api } from "./api";
import { ajaxHeader, ajaxValue, loopWindow, readSignRecord } from "./session";

/** What the browser was asked to send, as `call` built it. */
interface Sent {
  url: string;
  method: string;
  headers: Record<string, string>;
  body: string | undefined;
}

const sent: Sent[] = [];

function reply(status: number, body: unknown = {}) {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init: RequestInit) => {
      sent.push({
        url: String(url),
        method: String(init.method ?? "GET"),
        headers: (init.headers ?? {}) as Record<string, string>,
        body: init.body === undefined ? undefined : String(init.body),
      });
      return {
        ok: status >= 200 && status < 300,
        status,
        statusText: "Server Error",
        json: async () => body,
      };
    }),
  );
}

const last = () => sent[sent.length - 1];

const location = (pathname: string, hash = "", search = "") => {
  const assign = vi.fn();
  vi.stubGlobal("location", { pathname, search, hash, assign });
  return assign;
};

beforeEach(() => {
  sent.length = 0;
  sessionStorage.clear();
  // One browser per test, including the one thing it keeps across tabs.
  localStorage.clear();
  reply(200, { ok: true });
  location("/");
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the AJAX header", () => {
  it("rides on every request, so the server can tell them from a form someone else posted", async () => {
    await api.me();
    await api.list();
    await api.apply({ namespace: "demo", kind: "Deployment", name: "api" }, "hash");
    await api.revert({ namespace: "demo", kind: "Deployment", name: "api" });
    await api.setNamespaceSetting("demo", "learning-period", "P7D");
    expect(sent).toHaveLength(5);
    for (const s of sent) expect(s.headers[ajaxHeader]).toBe(ajaxValue);
  });

  it("sends the content type only where there is a body to type", async () => {
    await api.get({ namespace: "demo", kind: "Deployment", name: "api" });
    expect(last().headers["Content-Type"]).toBeUndefined();

    await api.apply({ namespace: "demo", kind: "Deployment", name: "api" }, "hash");
    expect(last().headers["Content-Type"]).toBe("application/json");
    expect(last().body).toBe(JSON.stringify({ hash: "hash" }));

    // The mutations without a body carry the header alone.
    await api.regenerate({ namespace: "demo", kind: "Deployment", name: "api" });
    expect(last().method).toBe("POST");
    expect(last().body).toBeUndefined();
    expect(last().headers["Content-Type"]).toBeUndefined();
    expect(last().headers[ajaxHeader]).toBe(ajaxValue);
  });

  it("escapes the workload it names", async () => {
    await api.get({ namespace: "a b", kind: "Deployment", name: "x/y" });
    expect(last().url).toBe("/api/v1/workloads/a%20b/Deployment/x%2Fy");
  });
});

describe("a refused request", () => {
  it("keeps the reason the server gave, which is what the UI answers with", async () => {
    reply(403, { error: "you need one of the editor groups", reason: "editor-required" });
    const err = await api
      .apply({ namespace: "demo", kind: "Deployment", name: "api" }, "hash")
      .catch((e: unknown) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 403, reason: "editor-required" });
    expect(String(err)).toContain("you need one of the editor groups");
  });

  it("reads a server that answers without JSON", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => ({
        ok: false,
        status: 502,
        statusText: "Bad Gateway",
        // What a proxy's error page does to res.json(): it rejects, and the status text is all there is.
        json: async () => {
          throw new SyntaxError("<html> is not JSON");
        },
      })),
    );
    const err = (await api.info().catch((e: unknown) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 502, reason: undefined });
    expect(err.message).toBe("Bad Gateway");
  });

  it("signs in again when the session ended, coming back to the page it was on", async () => {
    const assign = location("/deep/", "#/w/demo/Deployment/api/impact");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    expect(assign).toHaveBeenCalledWith("/auth/login?redirect=%2Fdeep%2F%23%2Fw%2Fdemo%2FDeployment%2Fapi%2Fimpact");
  });

  it("does not chase a sign-in from a page the server owns", async () => {
    const assign = location("/auth/callback");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    expect(assign).not.toHaveBeenCalled();
  });

  it("leaves every other 401 alone: the caller is signed in, this request was not accepted", async () => {
    const assign = location("/");
    reply(401, { error: "nope" });
    await api.info().catch(() => undefined);
    expect(assign).not.toHaveBeenCalled();
  });

  it("reports a viewer's refusal without signing anyone out", async () => {
    const assign = location("/");
    reply(403, { error: "view-only", reason: "forbidden" });
    const err = (await api.info().catch((e: unknown) => e)) as ApiError;
    expect(err.status).toBe(403);
    expect(assign).not.toHaveBeenCalled();
  });

  it("does not circle the provider when a sign-in brings back no cookie", async () => {
    const assign = location("/");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    await api.info().catch(() => undefined);
    expect(assign).toHaveBeenCalledTimes(1);
    // The mark that tells the page to stop and explain rather than send the browser off again.
    expect(readSignRecord().stopped).toEqual(expect.any(Number));
  });

  // The counter behind that explanation: only a page that believes it has a session gets this far, so a second refusal
  // of the same cookie is the server answering itself two ways, and the page says so rather than keep asking.
  it("counts the refusals of a session the page believed in", async () => {
    location("/");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    expect(readSignRecord()).toMatchObject({ refusals: 1 });
    await api.info().catch(() => undefined);
    expect(readSignRecord()).toMatchObject({ refused: expect.any(Number), refusals: 2 });
  });
  it("counts nothing when the refusal was not about the session", async () => {
    location("/");
    reply(403, { error: "view-only", reason: "forbidden" });
    await api.info().catch(() => undefined);
    await api.info().catch(() => undefined);
    expect(readSignRecord().refusals).toBeUndefined();
  });
  // A browser someone signed out expects a 401: counting those would hand the next session a card about a disagreement
  // that never happened.
  it("counts no refusal from a browser that was signed out on purpose", async () => {
    const assign = location("/", "", "?signed-out");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    await api.info().catch(() => undefined);
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord().refusals).toBeUndefined();
  });
  it("holds that answer for a tab that carries no marker of its own", async () => {
    localStorage.setItem("resourcerer.signed-out", String(Date.now()));
    const assign = location("/deep/", "#/workloads");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord()).toEqual({ at: undefined, stopped: undefined });
  });
  // The shape of a disagreement between replicas over one cookie: an attempt old enough to be retried, and a run of
  // refusals that says a retry is the wrong answer. The page stops here rather than authorize again every window.
  it("asks the provider no more once the refusals add up, window or no window", async () => {
    const now = Date.now();
    sessionStorage.setItem("resourcerer.sign-in", String(now - loopWindow - 1));
    sessionStorage.setItem("resourcerer.refused", String(now - 1_000));
    sessionStorage.setItem("resourcerer.refusals", "2");
    const assign = location("/");
    reply(401, { error: "sign in", reason: "unauthenticated" });
    await api.info().catch(() => undefined);
    expect(assign).not.toHaveBeenCalled();
    expect(readSignRecord().stopped).toEqual(expect.any(Number));
  });
});
