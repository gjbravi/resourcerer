import type { Mode, Status } from "./api";
import type { Sizing } from "./format";

/** All user-facing copy lives here. */
export const t = {
  app: "resourcerer",
  tagline: "VPA recommendations for your workloads",
  theme: { light: "Switch to light theme", dark: "Switch to dark theme" },
  readOnly: "Read-only",
  mode: {
    label: { Off: "Recommend", Auto: "Auto", Disabled: "Disabled" } satisfies Record<Mode, string>,
    help: {
      Off: "VPA computes recommendations only. Pods keep their current resources.",
      Auto: "VPA updates pod resources, in place when possible, otherwise by recreating pods. The workload spec is not changed.",
      Disabled: "No VPA for this workload.",
    } satisfies Record<Mode, string>,
    confirmTitle: (m: Mode) => `Switch to ${t.mode.label[m]}?`,
    confirmBody: {
      Off: "VPA stops changing pods. Pods keep the resources they currently have until they are recreated.",
      Auto: "VPA will start changing resources of running pods. Pods may be restarted when an in-place resize is not possible.",
      Disabled: "The VPA for this workload is deleted, together with its usage history.",
    } satisfies Record<Mode, string>,
    source: (s: string) => `Set by ${s}`,
    confirm: "Switch",
    gitops: (label: string, annotation: string) =>
      `Stored on the workload as label ${label} and annotation ${annotation}, so you can also set it in Git.`,
  },
  status: {
    label: {
      Pending: "Pending",
      Ready: "Ready",
      LowConfidence: "Learning",
      Conflict: "Conflict",
      Error: "Error",
      None: "–",
    } satisfies Record<Status, string>,
    help: {
      Pending: "Waiting for the VPA recommender to produce a first recommendation.",
      Ready: "VPA has a recommendation.",
      LowConfidence: "VPA has a recommendation but little usage history; it may still change.",
      Conflict: "A VPA with the same name exists and is not managed by resourcerer.",
      Error: "VPA cannot produce a recommendation.",
      None: "",
    } satisfies Record<Status, string>,
  },
  sizing: {
    label: { over: "Over-provisioned", under: "Under-provisioned", fit: "Right-sized", unknown: "–" } satisfies Record<
      Sizing,
      string
    >,
    help: {
      over: "Requests are above the VPA target.",
      under: "Requests are below the VPA target: risk of CPU throttling or OOM kills.",
      fit: "Requests are within 10% of the VPA target.",
      unknown: "",
    } satisfies Record<Sizing, string>,
  },
  list: {
    title: "Workloads",
    subtitle: (selection: string) =>
      selection === "label"
        ? "Workloads opted in with the resourcerer.io/enabled=true label on the workload or its namespace."
        : "All workloads outside excluded namespaces. Opt out with the resourcerer.io/enabled=false label.",
    search: "Search",
    namespace: "Namespace",
    allNamespaces: "All namespaces",
    empty: "No workloads match these filters.",
    filters: {
      all: "All",
      under: "Under-provisioned",
      over: "Over-provisioned",
      learning: "Learning",
      disabled: "Disabled",
    },
    columns: {
      workload: "Workload",
      mode: "Mode",
      status: "Status",
      sizing: "Sizing",
      cpu: "CPU request → target",
      memory: "Memory request → target",
    },
    stats: { total: "Workloads", managed: "With VPA", under: "Under-provisioned", over: "Over-provisioned" },
  },
  detail: {
    back: "Workloads",
    notFound: "Workload not found.",
    replicas: (n: number) => `${n} ${n === 1 ? "replica" : "replicas"}`,
    sections: { mode: "Mode", recommendation: "VPA recommendation", pods: "Pods", events: "Events" },
    recommendationHelp:
      "Target is what VPA recommends for requests. Lower and upper bounds are the range VPA considers safe; Auto mode only acts outside it.",
    container: "Container",
    request: "Request",
    target: "Target",
    range: "Lower – upper",
    limit: "Limit",
    noRecommendation: "No recommendation yet.",
    podsHelp: "Live pod resources. In Auto mode they follow the VPA target instead of the workload spec.",
    pod: "Pod",
    eventsEmpty: "No recent events.",
    disabled: "No VPA for this workload. Choose Recommend or Auto to create one.",
  },
  common: { cancel: "Cancel", retry: "Retry" },
};
