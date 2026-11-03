import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { LogOut, Moon, Sun, Unlock, UserRound } from "lucide-react";
import { StrictMode, useState } from "react";
import { createRoot } from "react-dom/client";
import type { Me, Session } from "./api";
import { ErrorBox, Toasts } from "./components";
import { useHash, useInfo, useSession, useWorkloads } from "./hooks";
import { Details } from "./pages/Details";
import { Overview } from "./pages/Overview";
import { SettingsPage } from "./pages/SettingsPage";
import { Workloads } from "./pages/Workloads";
import { configureSignals } from "./reliability";
import {
  bootState,
  clearSignOutIntent,
  clearSignRecord,
  readSignOutIntent,
  readSignRecord,
  shellLabels,
  writeLock,
} from "./session";
import { Crumbs, Sidebar } from "./shell";
import "./styles.css";

const queryClient = new QueryClient({ defaultOptions: { queries: { staleTime: 5_000, retry: 1 } } });

function App() {
  const [dark, setDark] = useState(() => localStorage.getItem("resourcerer.theme") === "dark");
  const hash = useHash();
  // The session is the one call the server answers without a signed-in caller, so it comes first and everything else
  // waits for it: asking as nobody would answer every request with 401.
  const session = useSession();
  const boot = bootState(session, readSignRecord(), window.location.search, readSignOutIntent());
  const me = session.data?.authenticated ? session.data : undefined;
  const asks = boot.view === "app";
  const info = useInfo(asks);
  configureSignals(info.data?.reliability);
  const workloads = useWorkloads(asks);
  const [path, query = ""] = hash.slice(1).split("?");
  const parts = path.split("/").filter(Boolean).map(decodeURIComponent);
  const page = parts[0] === "w" ? "workloads" : (parts[0] ?? "overview");
  const rows = workloads.data ?? [];
  const shell = shellLabels(info.data);
  const cluster = shell.cluster ?? "";
  const readOnly = info.data?.readOnly ?? false;
  const lock = writeLock(session.data, readOnly);

  let body;
  if (boot.view === "loading") body = <div className="rz-loading">Loading…</div>;
  else if (boot.view === "error")
    body = (
      <div className="rz-page">
        <ErrorBox error={boot.error} />
      </div>
    );
  else if (boot.view === "signing-in") body = <div className="rz-loading">Signing in…</div>;
  else if (boot.view === "signed-out") body = <SignedOut />;
  else if (boot.view === "cookie-dropped") body = <CookieDropped />;
  else if (boot.view === "no-access") body = <NoAccess me={boot.me} />;
  else if (parts[0] === "w" && parts.length >= 4) {
    body = (
      <Details
        key={parts.slice(1, 4).join("/")}
        wref={{ namespace: parts[1], kind: parts[2], name: parts[3] }}
        tab={parts[4] ?? "recommendation"}
        lock={lock}
      />
    );
  } else if (page === "workloads") {
    body = <Workloads key={query} rows={rows} query={query} lock={lock} />;
  } else if (page === "settings") {
    body = info.data ? <SettingsPage info={info.data} /> : <div className="rz-loading">Loading…</div>;
  } else {
    body = <Overview rows={rows} cluster={cluster} />;
  }

  return (
    <div className={`rz${dark ? " dark" : ""}`}>
      <Sidebar page={page} shell={shell} configStale={me?.configStale} configError={me?.configError} />
      <div className="rz-main">
        <div className="rz-surface">
          <header className="rz-top">
            <Crumbs page={page} parts={parts} shell={shell} />
            <div className="rz-topright">
              <Who session={session.data} readOnly={readOnly} />
              <button
                className="rz-icon"
                onClick={() => setDark((x) => (localStorage.setItem("resourcerer.theme", x ? "light" : "dark"), !x))}
                aria-label={dark ? "Light theme" : "Dark theme"}
              >
                {dark ? <Sun size={16} /> : <Moon size={16} />}
              </button>
            </div>
          </header>
          {/* A strip of its own only while there is a page under it: with no page at all the boot view carries the
              failure, and saying it twice would only make it look like two problems. */}
          {boot.view === "app" && (session.error || info.error || workloads.error) && (
            <div className="rz-page" style={{ paddingBottom: 0 }}>
              <ErrorBox error={session.error ?? info.error ?? workloads.error} />
            </div>
          )}
          {body}
        </div>
      </div>
      <Toasts />
    </div>
  );
}

/**
 * Who is calling, where the install says who may do what. The strategy decides which chip this is, so "anyone who
 * reaches this page can use it" appears only under `anonymous` and never while an `oidc` install is still finding out
 * who is calling. A caller under `anonymous` is named the way this header always named them: there is no one behind
 * that session, and signing out of it would achieve nothing.
 */
function Who({ session, readOnly }: { session: Session | undefined; readOnly: boolean }) {
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
      <a className="rz-icon" href="/auth/logout" title="Sign out">
        <LogOut size={16} />
      </a>
    </>
  );
}

/**
 * Signed in, but with no role here. An empty page would read as a broken install, so this says what is missing. Which
 * of the two rules refused the caller is the install's own business, so both are named and the user's own groups are
 * the only list put on screen.
 */
function NoAccess({ me }: { me: Me }) {
  return (
    <div className="rz-page">
      <section className="rz-card">
        <div className="rz-cardhd">
          <div>
            <h2>Signed in without access</h2>
            <p className="rz-dim rz-xs">
              <b>{me.user}</b> signed in with {me.strategy}, but none of the groups below grants access here. Either
              none of them is in <code>auth.oidc.allowedGroups</code>, or this install reads for editors only (
              <code>auth.oidc.roles.anyAuthenticatedIsViewer: false</code>) and none of them is in{" "}
              <code>auth.oidc.roles.editorGroups</code>. Ask whoever owns this install to check those two.
            </p>
            <p className="rz-dim rz-xs">
              Signing out and in again will not change it: a first sign-in is refused before it reaches this page, so
              you are here because the groups or the configuration changed since you signed in.
            </p>
          </div>
        </div>
        <div className="rz-cardft">
          <div className="rz-chips">
            {me.groups.length ? (
              me.groups.map((g) => <code key={g}>{g}</code>)
            ) : (
              <span className="rz-dim">The provider presented no groups.</span>
            )}
          </div>
        </div>
      </section>
    </div>
  );
}

/**
 * Sign-out ended this browser's session and left it on the address that says so. Nothing happens here on its own: the
 * provider still has a session of its own, so the round trip the SPA would otherwise make ends with the user signed
 * back in and the screen that told them otherwise gone. Holding that back is the browser's own business, not this
 * tab's — every tab shares one cookie — so the intent is kept where all of them read it, and this card's button is the
 * only thing that lifts it.
 */
function SignedOut() {
  return (
    <div className="rz-page">
      <section className="rz-card">
        <div className="rz-cardhd">
          <div>
            <h2>Signed out</h2>
            <p className="rz-dim rz-xs">
              This browser holds no resourcerer session any more, and no tab of it will start one on its own — not on a
              reload, not after the browser is closed and opened again — until someone presses Sign in again here. The
              identity provider was not told: its own session stays open, so in another browser or profile on this
              machine, which this answer does not reach, opening this address signs straight in with no click and no
              password. Signing out of resourcerer is not signing out of the provider: whoever needs this machine to
              forget them signs out there too.
            </p>
          </div>
        </div>
        <div className="rz-cardft">
          <a className="rz-btn" href="/auth/login" onClick={() => clearSignOutIntent()}>
            Sign in again
          </a>
        </div>
      </section>
    </div>
  );
}

/**
 * The session cookie this page carries is not one the server will accept, so the SPA stopped sending the browser to
 * the provider. Only the browser or the install's configuration can fix that, so this names them instead of looping.
 */
function CookieDropped() {
  return (
    <div className="rz-page">
      <section className="rz-card">
        <div className="rz-cardhd">
          <div>
            <h2>The session cookie was not kept</h2>
            <p className="rz-dim rz-xs">
              Either the cookie coming back from the provider held no session resourcerer could read, or this page reads
              it fine while the calls that follow it are refused — one cookie, two answers. Either way the page stopped
              asking rather than loop through it. A session is sealed with the key at{" "}
              <code>auth.session.keySecretRef</code> and names the issuer and client it was minted for, so a cookie
              written before either changed, one sealed with a key a rollout has since retired, one from another
              install, or one a proxy stripped on the way here all read the same way. Check that the server's replicas
              all mount the same session key, including <code>auth.session.keyPreviousSecretRef</code> while a rotation
              is in flight, and that nothing between the browser and the server drops cookies.
            </p>
          </div>
        </div>
        <div className="rz-cardft">
          <button
            className="rz-btn"
            onClick={() => {
              // Retrying is the user's decision: it is the one action that ends a run of refusals, and the one that
              // earns another round trip to the provider.
              clearSignRecord();
              window.location.reload();
            }}
          >
            Try again
          </button>
        </div>
      </section>
    </div>
  );
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>
  </StrictMode>,
);
