import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { Boxes, LayoutDashboard, Moon, Server, Settings as SettingsIcon, Sun, Unlock } from "lucide-react";
import { StrictMode, useState } from "react";
import { createRoot } from "react-dom/client";
import { HatLogo } from "./art";
import { ErrorBox, Toasts } from "./components";
import { useHash, useInfo, useWorkloads } from "./hooks";
import { configureSignals } from "./reliability";
import { Details } from "./pages/Details";
import { Overview } from "./pages/Overview";
import { SettingsPage } from "./pages/SettingsPage";
import { Workloads } from "./pages/Workloads";
import "./styles.css";

const queryClient = new QueryClient({ defaultOptions: { queries: { staleTime: 5_000, retry: 1 } } });

function App() {
  const [dark, setDark] = useState(() => localStorage.getItem("resourcerer.theme") === "dark");
  const hash = useHash();
  const info = useInfo();
  configureSignals(info.data?.reliability);
  const workloads = useWorkloads();
  const [path, query = ""] = hash.slice(1).split("?");
  const parts = path.split("/").filter(Boolean).map(decodeURIComponent);
  const page = parts[0] === "w" ? "workloads" : (parts[0] ?? "overview");
  const rows = workloads.data ?? [];
  const cluster = info.data?.cluster ?? "";
  const readOnly = info.data?.readOnly ?? false;
  const nav = [
    ["#/overview", "overview", "Overview", LayoutDashboard],
    ["#/workloads", "workloads", "Workloads", Boxes],
    ["#/settings", "settings", "Settings", SettingsIcon],
  ] as const;

  let body;
  if (parts[0] === "w" && parts.length >= 4) {
    body = (
      <Details
        key={parts.slice(1, 4).join("/")}
        wref={{ namespace: parts[1], kind: parts[2], name: parts[3] }}
        tab={parts[4] ?? "recommendation"}
        readOnly={readOnly}
      />
    );
  } else if (page === "workloads") {
    body = <Workloads key={query} rows={rows} query={query} readOnly={readOnly} />;
  } else if (page === "settings") {
    body = info.data ? <SettingsPage info={info.data} /> : <div className="rz-loading">Loading…</div>;
  } else {
    body = <Overview rows={rows} cluster={cluster} />;
  }

  return (
    <div className={`rz${dark ? " dark" : ""}`}>
      <aside className="rz-side">
        <a className="rz-logo" href="#/overview">
          <HatLogo /> <span>resourcerer</span>
          <span className="rz-beta" title="Beta: APIs, annotations and the CRD may change">
            Beta
          </span>
        </a>
        <div className="rz-cluster">
          <Server size={14} />
          <div>
            <div className="rz-dim rz-xs">Cluster</div>
            <b>{cluster || "…"}</b>
          </div>
        </div>
        <nav>
          {nav.map(([href, id, label, Icon]) => (
            <a key={id} href={href} className={page === id ? "on" : ""} aria-current={page === id ? "page" : undefined}>
              <Icon size={17} /> {label}
            </a>
          ))}
        </nav>
        <div className="rz-sidefoot">
          <div className="rz-dim rz-xs">{info.data?.version}</div>
          <div className="rz-dim rz-xs">Prometheus · {info.data?.prometheus.url.replace(/^https?:\/\//, "")}</div>
        </div>
      </aside>
      <div className="rz-main">
        <div className="rz-surface">
          <header className="rz-top">
            <div className="rz-crumbs rz-dim">
              {cluster} <span>/</span>
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
            <div className="rz-topright">
              <span
                className="rz-anon"
                title="Authentication strategy: anonymous. Anyone who reaches this page can use it."
              >
                <Unlock size={12} /> anonymous{readOnly ? " · read-only" : ""}
              </span>
              <button
                className="rz-icon"
                onClick={() => setDark((x) => (localStorage.setItem("resourcerer.theme", x ? "light" : "dark"), !x))}
                aria-label={dark ? "Light theme" : "Dark theme"}
              >
                {dark ? <Sun size={16} /> : <Moon size={16} />}
              </button>
            </div>
          </header>
          {(info.error || workloads.error) && (
            <div className="rz-page" style={{ paddingBottom: 0 }}>
              <ErrorBox error={info.error ?? workloads.error} />
            </div>
          )}
          {body}
        </div>
      </div>
      <Toasts />
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
