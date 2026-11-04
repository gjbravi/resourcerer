import { Boxes, LayoutDashboard, LogOut, Server, Settings as SettingsIcon, Unlock, UserRound } from "lucide-react";
import type { Session } from "./api";
import { HatLogo } from "./art";
import { setSignOutIntent, type ShellLabels } from "./session";

const nav = [
  ["#/overview", "overview", "Overview", LayoutDashboard],
  ["#/workloads", "workloads", "Workloads", Boxes],
  ["#/settings", "settings", "Settings", SettingsIcon],
] as const;

/**
 * What this install is, which cluster it is looking at, and the way between its three pages. Every label comes from
 * `GET /api/v1/info`, which is answered only to a signed-in caller, so the boot views render this with none of them —
 * and a label the shell does not have is absent rather than placeholder text.
 */
export function Sidebar({
  page,
  shell,
  configStale,
  configError,
}: {
  page: string;
  shell: ShellLabels;
  configStale?: boolean;
  configError?: string;
}) {
  return (
    <aside className="rz-side">
      <a className="rz-logo" href="#/overview">
        <HatLogo /> <span>resourcerer</span>
        <span className="rz-beta" title="Beta: APIs, annotations and the CRD may change">
          Beta
        </span>
      </a>
      {shell.cluster && (
        <div className="rz-cluster">
          <Server size={14} />
          <div>
            <div className="rz-dim rz-xs">Cluster</div>
            <b>{shell.cluster}</b>
          </div>
        </div>
      )}
      <nav>
        {nav.map(([href, id, label, Icon]) => (
          <a key={id} href={href} className={page === id ? "on" : ""} aria-current={page === id ? "page" : undefined}>
            <Icon size={17} /> {label}
          </a>
        ))}
      </nav>
      {(shell.version || shell.prometheus || configStale) && (
        <div className="rz-sidefoot">
          {shell.version && <div className="rz-dim rz-xs">{shell.version}</div>}
          {shell.prometheus && <div className="rz-dim rz-xs">Prometheus · {shell.prometheus}</div>}
          {configStale && (
            <div className="rz-xs t-warn" title={configError || "The sign-in configuration could not be reread."}>
              Sign-in configuration is out of date
            </div>
          )}
        </div>
      )}
    </aside>
  );
}

/** Where the caller is, next to the cluster they are in — which the crumbs do not claim to know before `info` says. */
export function Crumbs({ page, parts, shell }: { page: string; parts: string[]; shell: ShellLabels }) {
  return (
    <div className="rz-crumbs rz-dim">
      {shell.cluster && (
        <>
          {shell.cluster} <span>/</span>
        </>
      )}
      {parts[0] === "w" ? (
        <>
          <a href="#/workloads">Workloads</a> <span>/</span>{" "}
          <b>
            {parts[1]}/{parts[3]}
          </b>
        </>
      ) : (
        <b>{page[0].toUpperCase() + page.slice(1)}</b>
      )}
    </div>
  );
}

/**
 * Who is calling, where the install says who may do what. The strategy decides which chip this is, so "anyone who
 * reaches this page can use it" appears only under `anonymous` and never while an `oidc` install is still finding out
 * who is calling. A caller under `anonymous` is named the way this header always named them: there is no one behind
 * that session, and signing out of it would achieve nothing.
 *
 * Signing out is recorded as the click happens, not when the page it lands on finds out that it happened. In between,
 * the cookie is gone and the answer every other tab reads by is not written yet, while those tabs keep polling: each
 * takes the 401 this click caused and answers it the only way a page blind to the click can — by handing the browser
 * to a provider that still has its own session open, which brings back the session the click asked to end. A click's
 * handler and a `localStorage` write are both synchronous, so nothing slips between them. Landing on the marker writes
 * it too, which covers a sign-out started anywhere else: an address typed in, or a link followed from outside the app.
 */
export function Who({ session, readOnly }: { session: Session | undefined; readOnly: boolean }) {
  if (session?.strategy === "anonymous") {
    return (
      <span className="rz-anon" title="Authentication strategy: anonymous. Anyone who reaches this page can use it.">
        <Unlock size={12} /> anonymous{readOnly ? " · read-only" : ""}
      </span>
    );
  }
  if (!session || !session.authenticated) return null;
  const me = session;
  const groups = me.groups.length ? ` · ${me.groups.join(", ")}` : "";
  return (
    <>
      <span className="rz-anon" title={`Signed in as ${me.user} (${me.role})${groups}`}>
        <UserRound size={12} /> {me.user} · {me.role}
        {readOnly ? " · read-only" : ""}
      </span>
      <a className="rz-icon" href="/auth/logout" title="Sign out" onClick={() => setSignOutIntent()}>
        <LogOut size={16} />
      </a>
    </>
  );
}
