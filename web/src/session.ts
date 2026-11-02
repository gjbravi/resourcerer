import type { Session } from "./api";

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

/** Hand the browser to the provider, keeping the page it was on so the way back is the same page. */
export function sendToSignIn(page: Page = window.location) {
  const to = signInURL(page);
  if (to) page.assign(to);
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
