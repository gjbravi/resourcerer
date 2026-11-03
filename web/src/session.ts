import type { Info, Me, Session } from "./api";

/**
 * The header that marks a request as resourcerer's own. The server requires it on mutations, where it is the one
 * thing a page from another origin cannot attach: a custom header makes the browser ask the server to allow that
 * origin first, and the API grants none.
 */
export const ajaxHeader = "X-Resourcerer-Request";
export const ajaxValue = "1";

/** Where the browser is, as far as a redirect back here is concerned. The page lives in the hash, so it travels too. */
export interface Page {
  pathname: string;
  search: string;
  hash: string;
  assign(url: string): void;
}

export const pagePath = (page: Page) => `${page.pathname}${page.search}${page.hash}`;

/** Whether the browser sits on a path the server serves itself: sign in, sign out, or the callback in between. */
export const onAuthPath = (pathname: string) => pathname.startsWith("/auth/");

/**
 * Where to send a caller who is not signed in, or nothing when the browser is already on a path the server owns:
 * handing one of those to sign in only brings it back to the page that would send it again.
 */
export function signInURL(page: Page): string | null {
  if (onAuthPath(page.pathname)) return null;
  return `/auth/login?redirect=${encodeURIComponent(pagePath(page))}`;
}

/**
 * How long a sign-in attempt stays current. Coming back from the provider without a session cookie inside that window
 * means the browser will not keep it, and every further turn of `/me` → provider → callback → `/me` is another full
 * authorization for the same nothing.
 */
export const loopWindow = 30_000;

/** This tab's sign-in record: when it went to the provider, and when a round trip came back with no session. */
export interface SignRecord {
  at?: number;
  stopped?: number;
}

/** The part of `sessionStorage` the record needs. */
export interface Store {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

const attemptKey = "resourcerer.sign-in";
const stoppedKey = "resourcerer.sign-in-stopped";

/**
 * `sessionStorage`, or an empty one where the browser will not hand it over. Losing the record costs the loop breaker
 * and nothing else, which is what a browser that refuses storage already costs the session cookie.
 */
export const signStore: Store = {
  getItem: (key) => guarded(() => sessionStorage.getItem(key), null),
  setItem: (key, value) => guarded(() => sessionStorage.setItem(key, value), undefined),
  removeItem: (key) => guarded(() => sessionStorage.removeItem(key), undefined),
};

function guarded<T>(run: () => T, ifBlocked: T): T {
  try {
    return run();
  } catch {
    return ifBlocked;
  }
}

export function readSignRecord(store: Store = signStore): SignRecord {
  return { at: when(store.getItem(attemptKey)), stopped: when(store.getItem(stoppedKey)) };
}

function when(stored: string | null): number | undefined {
  const at = Number(stored);
  return stored !== null && stored !== "" && Number.isFinite(at) ? at : undefined;
}

export function clearSignRecord(store: Store = signStore) {
  store.removeItem(attemptKey);
  store.removeItem(stoppedKey);
}

/** Whether to hand the browser to the provider, or to stop because the last attempt is what put us back here. */
export function signMove(now: number, record: SignRecord): "sign-in" | "wait" {
  return record.at !== undefined && now - record.at < loopWindow ? "wait" : "sign-in";
}

/**
 * The marker sign-out leaves on the address the browser lands on. Sitting there with no session is a state the user
 * asked for, so it is the one place an unauthenticated `/me` must not hand the browser to the provider: the provider
 * has its own session and would sign the user straight back in.
 */
export const signedOutMarker = "signed-out";
export const signedOut = (search: string) => new URLSearchParams(search).has(signedOutMarker);

/**
 * Hand the browser to the provider, keeping the page it was on so the way back is the same page. An attempt inside
 * `loopWindow` is not repeated: the message on screen says what a second one would not fix.
 */
export function sendToSignIn(page: Page = window.location, store: Store = signStore, now = Date.now()) {
  if (signedOut(page.search)) return;
  const to = signInURL(page);
  if (!to) return;
  if (signMove(now, readSignRecord(store)) === "wait") {
    store.setItem(stoppedKey, String(now));
    return;
  }
  store.setItem(attemptKey, String(now));
  store.removeItem(stoppedKey);
  page.assign(to);
}

/**
 * What the answer from `GET /api/v1/me` means for this tab: a session retires the sign-in record, and no session
 * starts the sign-in that the record may decide to hold back.
 */
export function afterMe(session: Session, page: Page = window.location, store: Store = signStore, now = Date.now()) {
  if (!session.authenticated) {
    sendToSignIn(page, store, now);
    return;
  }
  // An answer inside the window is not evidence that the last attempt worked: a session cookie the server refuses
  // reads as signed in for one call and as a 401 on the next, and clearing the record here would start the round trip
  // over with a clean slate every time. Retire it once the window has passed, which is a session that stayed.
  const record = readSignRecord(store);
  if (record.at === undefined || now - record.at >= loopWindow) clearSignRecord(store);
}

/** Whether the caller may change anything, and what to say where a control is disabled instead. */
export interface Lock {
  write: boolean;
  /** Short enough for a tooltip; absent when there is no reason to explain, or none to explain it to. */
  reason?: string;
  /** The install refuses changes from anyone, which is why some controls hide rather than disable. */
  readOnly: boolean;
}

/** A read-only install comes first: it is the operator's decision, taken over whatever role the caller has. */
export function writeLock(session: Session | undefined, readOnly = false): Lock {
  if (readOnly) return { write: false, reason: "Read-only", readOnly: true };
  if (session?.authenticated !== true) return { write: false, readOnly: false };
  if (session.role === "editor") return { write: true, readOnly: false };
  if (session.role === "viewer") return { write: false, reason: "View-only access", readOnly: false };
  return { write: false, readOnly: false };
}

/** What to show before there is a page to show. Every one of these is an answer `/me` gave, or the lack of one. */
export type Boot =
  | { view: "loading" }
  | { view: "error"; error: unknown }
  | { view: "signing-in" }
  | { view: "signed-out" }
  | { view: "cookie-dropped" }
  | { view: "no-access"; me: Me }
  | { view: "app"; me: Me };

/**
 * The boot decision. An answer outranks an error: when `/me` fails after having answered once, the page stays up with
 * the error in the header, which is what an install did before there was a session to ask about. Only a `/me` that
 * never answered leaves the app with nothing to render, and that has to say so rather than load forever. Of the three
 * ways to have no session, the one the user asked for comes first: `search` carries the marker sign-out left there.
 */
export function bootState(
  session: { data: Session | undefined; error: unknown },
  record: SignRecord = {},
  search = "",
): Boot {
  const me = session.data;
  if (!me) return session.error ? { view: "error", error: session.error } : { view: "loading" };
  if (!me.authenticated) {
    if (signedOut(search)) return { view: "signed-out" };
    return record.stopped === undefined ? { view: "signing-in" } : { view: "cookie-dropped" };
  }
  return me.role === "none" ? { view: "no-access", me } : { view: "app", me };
}

/**
 * The labels the shell carries around whatever page it holds: the cluster being looked at, the build serving it, and
 * the Prometheus behind that. All three come from `GET /api/v1/info`, which is answered only to a signed-in caller, so
 * every boot view renders the shell without them — and those screens are where a new install is first seen, where
 * `Cluster …` and `Prometheus ·` read as a broken install rather than as nobody signed in yet. A label the shell does
 * not have is therefore absent, never placeholder text, and a value the server left empty counts as not having one.
 */
export interface ShellLabels {
  cluster?: string;
  version?: string;
  prometheus?: string;
}

export function shellLabels(info: Info | undefined): ShellLabels {
  if (!info) return {};
  const labels: ShellLabels = {};
  if (info.cluster) labels.cluster = info.cluster;
  if (info.version) labels.version = info.version;
  const url = info.prometheus.url.replace(/^https?:\/\//, "");
  if (url) labels.prometheus = url;
  return labels;
}
