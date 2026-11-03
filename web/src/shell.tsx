import { Boxes, LayoutDashboard, Settings as SettingsIcon, Server } from "lucide-react";
import { HatLogo } from "./art";
import type { ShellLabels } from "./session";

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
