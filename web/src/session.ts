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
  /** When a call this tab believed it was entitled to make was refused, and how many refusals in a row. */
  refused?: number;
  refusals?: number;
  /** When `/me` last said this tab has a session, which is what makes a later refusal a second answer about a cookie
   *  the server has just accepted. See `countsAsDisagreement`. */
  confirmed?: number;
}

/** The part of `sessionStorage` the record needs. */
export interface Store {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
}

const attemptKey = "resourcerer.sign-in";
const stoppedKey = "resourcerer.sign-in-stopped";
const confirmedKey = "resourcerer.sign-confirmed";
const refusedKey = "resourcerer.refused";
const refusalsKey = "resourcerer.refusals";
/**
 * Where the sign-out intent is kept. Named because a tab that did not sign anyone out has to notice: `localStorage`
 * announces the write to every other tab of this browser, and that announcement is what stops them sitting on a session
 * the browser no longer has.
 */
export const signOutKey = "resourcerer.signed-out";

/**
 * `sessionStorage`, or an empty one where the browser will not hand it over. Losing the record costs the loop breaker
 * and nothing else, which is what a browser that refuses storage already costs the session cookie.
 */
export const signStore: Store = {
  getItem: (key) => guarded(() => sessionStorage.getItem(key), null),
  setItem: (key, value) => guarded(() => sessionStorage.setItem(key, value), undefined),
  removeItem: (key) => guarded(() => sessionStorage.removeItem(key), undefined),
};

/**
 * `localStorage`, for the one fact that is not this tab's own. Same deal as above where the browser refuses it: the
 * intent is simply not there, and everything falls back to what one tab's address can say.
 */
export const intentStore: Store = {
  getItem: (key) => guarded(() => localStorage.getItem(key), null),
  setItem: (key, value) => guarded(() => localStorage.setItem(key, value), undefined),
  removeItem: (key) => guarded(() => localStorage.removeItem(key), undefined),
};

function guarded<T>(run: () => T, ifBlocked: T): T {
  try {
    return run();
  } catch {
    return ifBlocked;
  }
}

export function readSignRecord(store: Store = signStore): SignRecord {
  return {
    at: when(store.getItem(attemptKey)),
    stopped: when(store.getItem(stoppedKey)),
    confirmed: when(store.getItem(confirmedKey)),
    refused: when(store.getItem(refusedKey)),
    refusals: when(store.getItem(refusalsKey)),
  };
}

function when(stored: string | null): number | undefined {
  const at = Number(stored);
  return stored !== null && stored !== "" && Number.isFinite(at) ? at : undefined;
}

/**
 * The attempt itself: the trip to the provider and the stop that followed one that came back with no cookie. A session
 * that stayed retires these and nothing else — see `afterMe`.
 */
export function clearSignAttempt(store: Store = signStore) {
  store.removeItem(attemptKey);
  store.removeItem(stoppedKey);
}

/** Everything this tab remembers about signing in and being refused, which is what a deliberate retry starts over. */
export function clearSignRecord(store: Store = signStore) {
  clearSignAttempt(store);
  store.removeItem(confirmedKey);
  store.removeItem(refusedKey);
  store.removeItem(refusalsKey);
}

/** Whether to hand the browser to the provider, or to stop because the last attempt is what put us back here. */
export function signMove(now: number, record: SignRecord): "sign-in" | "wait" {
  return record.at !== undefined && now - record.at < loopWindow ? "wait" : "sign-in";
}

/**
 * Whether the refusals this tab has taken add up to a disagreement rather than a session that ended. A run holds until
 * whoever reads the explanation acts on it: the card's own retry clears the count, because `/me` — the half of the
 * disagreement that keeps answering — has nothing to admit, and a page that forgot the count on its own would spend the
 * explanation it just gave.
 */
export const refusalLimit = 2;
export const refusedTwice = (record: SignRecord) => (record.refusals ?? 0) >= refusalLimit;

/**
 * The next refusal of a call this tab believed it was entitled to make. Refusals count while they keep arriving; one
 * after the window has passed starts a fresh count instead of lengthening a run nobody remembers.
 */
export function countRefusal(now: number, record: SignRecord): number {
  return record.refused !== undefined && now - record.refused < loopWindow ? (record.refusals ?? 0) + 1 : 1;
}

/**
 * Whether a 401 is evidence of a disagreement at all, or the tail of a sign-in this tab has not seen confirmed. The
 * counter is about one cookie answered two ways, which needs an answer from `/me` to disagree with: until the trip this
 * tab started comes back with a session, a refusal only says the provider has not finished answering. Everything in
 * flight when a session ends lands in that tail — the polls and their retries arrive within milliseconds of each other,
 * which is how an ordinary expiry came to accuse the server of dropping a cookie it had never been shown.
 */
export const countsAsDisagreement = (record: SignRecord) =>
  record.at === undefined || (record.confirmed ?? 0) > record.at;

export function noteRefusal(store: Store = signStore, now = Date.now()) {
  const record = readSignRecord(store);
  if (!countsAsDisagreement(record)) return;
  store.setItem(refusedKey, String(now));
  store.setItem(refusalsKey, String(countRefusal(now, record)));
}

/**
 * The marker sign-out leaves on the address the browser lands on. Sitting there with no session is a state the user
 * asked for, so it is the one place an unauthenticated `/me` must not hand the browser to the provider: the provider
 * has its own session and would sign the user straight back in.
 */
export const signedOutMarker = "signed-out";
export const signedOut = (search: string) => new URLSearchParams(search).has(signedOutMarker);

/**
 * Whether this browser was signed out on purpose and has not been asked back. The marker describes one tab's address,
 * while the cookie every tab shares is what has to be held back: a second tab left on the app polls, takes a 401, and
 * would sign the whole browser in again behind the screen promising the opposite. So the intent is written where every
 * tab reads it, and lifted only by whoever asks for a session again.
 *
 * It is not a lock and it outlives the tab. The server signs in anyone who asks at `/auth/login`, and so does any
 * browser this one cannot reach; the intent is this browser declining to ask on its own. It also has no expiry, which
 * is the same promise taken at its word — a window short enough to never annoy anyone is one short enough to let the
 * next morning's visit sign itself in — and it costs the person who signed out one click on the card that says so.
 */
export const readSignOutIntent = (store: Store = intentStore) => store.getItem(signOutKey) !== null;

export function setSignOutIntent(store: Store = intentStore, now = Date.now()) {
  store.setItem(signOutKey, String(now));
}

export function clearSignOutIntent(store: Store = intentStore) {
  store.removeItem(signOutKey);
}

/**
 * Whether this browser was told to stay signed out: by this tab's own address, or by the tab that took the sign-out and
 * wrote it down for the rest. Two things follow from that answer and not from the cookie — an unauthenticated `/me` is
 * what was asked for rather than a reason to visit the provider, and a 401 is the answer a signed-out browser expects,
 * so it is not the two answers about one cookie that the refusal counter counts.
 */
export const holdingAtSignIn = (search: string, intent: Store = intentStore) =>
  signedOut(search) || readSignOutIntent(intent);

/**
 * The marker says this tab is where a sign-out left it. Once `/me` reports a session it describes a state the tab does
 * not have, and the next time that session ends the page would call an expiry a sign-out nobody asked for.
 */
export const staleMarker = (session: Session | undefined, search: string) =>
  session?.authenticated === true && signedOut(search);

/** The tab's address with the marker taken out of the query, and nothing else about it changed. */
export function withoutSignedOut(page: { pathname: string; search: string; hash: string }): string {
  const kept = new URLSearchParams(page.search);
  kept.delete(signedOutMarker);
  const query = kept.toString();
  return `${page.pathname}${query ? `?${query}` : ""}${page.hash}`;
}

/**
 * Hand the browser to the provider, keeping the page it was on so the way back is the same page. An attempt inside
 * `loopWindow` is not repeated: the message on screen says what a second one would not fix. A run of refusals stops it
 * outright, window or no window — that is the server answering itself two ways about one cookie, and another
 * authorization at the provider cannot change its mind. The card that appears is the one with the retry in it.
 */
export function sendToSignIn(
  page: Page = window.location,
  store: Store = signStore,
  now = Date.now(),
  intent: Store = intentStore,
) {
  // Either half of the same answer: what this tab's address says, and what every tab of this browser was told. The
  // first is all there is where the browser will not hold the second.
  if (holdingAtSignIn(page.search, intent)) return;
  const to = signInURL(page);
  if (!to) return;
  const record = readSignRecord(store);
  if (signMove(now, record) === "wait" || refusedTwice(record)) {
    store.setItem(stoppedKey, String(now));
    return;
  }
  store.setItem(attemptKey, String(now));
  store.removeItem(stoppedKey);
  // A run of refusals is a statement about the cookie that came back from the provider, so it ends where a new round
  // trip starts. What follows this address is the browser asking again, and the answer to that is worth reading on its
  // own rather than as the third refusal in a row nobody remembers.
  store.removeItem(refusedKey);
  store.removeItem(refusalsKey);
  page.assign(to);
}

/**
 * What the answer from `GET /api/v1/me` means for this tab: a session retires the sign-in record, and no session
 * starts the sign-in that the record may decide to hold back.
 */
export function afterMe(
  session: Session,
  page: Page = window.location,
  store: Store = signStore,
  now = Date.now(),
  intent: Store = intentStore,
) {
  if (!session.authenticated) {
    // Landing on the marker is the sign-out completing, so this is where the browser-wide intent gets written down:
    // the tab that was told has to be the one that tells the rest.
    if (signedOut(page.search)) setSignOutIntent(intent, now);
    sendToSignIn(page, store, now, intent);
    return;
  }
  // The session this page can now spend is the answer a later refusal disagrees with, so this is what makes one count:
  // until the trip this tab started comes back, a 401 only says the provider has not finished answering.
  store.setItem(confirmedKey, String(now));
  // An answer inside the window is not evidence that the last attempt worked: a session cookie the server refuses
  // reads as signed in for one call and as a 401 on the next, and clearing the record here would start the round trip
  // over with a clean slate every time. Retire it once the window has passed, which is a session that stayed. A run of
  // refusals is retired by neither: `/me` is the half of that disagreement that keeps answering, so letting it wipe the
  // count would hide the explanation the card exists to give until the next 401 rebuilt it. Whoever reads the card ends
  // the run.
  const record = readSignRecord(store);
  if (record.at === undefined || now - record.at >= loopWindow) clearSignAttempt(store);
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

/**
 * What to show before there is a page to show. Each is an answer `/me` gave, the lack of one, or — for the cookie
 * `/me` reads and the calls after it do not — an answer the rest of the server will not stand behind.
 */
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
 * never answered leaves the app with nothing to render, and that has to say so rather than load forever. Of the ways
 * to end up with no session, the one the caller asked for comes first, whether it is this tab's own address or the
 * intent every tab of this browser shares. A session can still be a dead end: `/me` reading the cookie while the calls
 * that follow it are refused is the same cookie answered twice by replicas that disagree, and that card explains it.
 */
export function bootState(
  session: { data: Session | undefined; error: unknown },
  record: SignRecord = {},
  search = "",
  browserSignedOut = false,
): Boot {
  const me = session.data;
  if (!me) return session.error ? { view: "error", error: session.error } : { view: "loading" };
  if (!me.authenticated) {
    if (signedOut(search) || browserSignedOut) return { view: "signed-out" };
    return record.stopped === undefined ? { view: "signing-in" } : { view: "cookie-dropped" };
  }
  if (me.role === "none") return { view: "no-access", me };
  if (refusedTwice(record)) return { view: "cookie-dropped" };
  return { view: "app", me };
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
