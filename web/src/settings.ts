/** Labels, help and suggested values for each setting; the server validates values. */
export interface SettingMeta {
  group: string;
  label: string;
  help: (v: string) => string;
  suggestions?: string[];
}

const goal = (r: "cpu" | "memory") => (v: string) => {
  const table: Record<string, Record<string, string>> = {
    Savings: { cpu: "p85 + 5%", memory: "p99 + 10%" },
    Balanced: { cpu: "p95 + 10%", memory: "max + 15%" },
    Reliability: { cpu: "p99 + 20%", memory: "max + 30%" },
  };
  return `${r === "cpu" ? "CPU" : "Memory"} request = ${table[v]?.[r] ?? "?"} of usage over the learning period.`;
};

const limits: Record<string, string> = {
  RequestsOnly: "Change requests; keep limits as they are.",
  RequestsRaiseLimitsIfNeeded: "Change requests; raise the limit only when the new request would exceed it.",
  RequestsAndLimits: "Change requests and set limits to request × ratio.",
  DoNotOptimize: "Keep requests and limits as they are.",
};

export const SETTINGS: Record<string, SettingMeta> = {
  "cpu.optimization-goal": { group: "Goal", label: "CPU optimization goal", help: goal("cpu") },
  "memory.optimization-goal": { group: "Goal", label: "Memory optimization goal", help: goal("memory") },
  "learning-period": {
    group: "Schedule",
    label: "Learning period",
    help: () =>
      "History used for a recommendation (ISO 8601 duration). Recommendations are preliminary until it is covered.",
    suggestions: ["PT15M", "PT1H", "P1D", "P7D", "P14D"],
  },
  schedule: {
    group: "Schedule",
    label: "Recommendation schedule",
    help: () =>
      "How often recommendations are regenerated. Each workload gets a stable, spread-out time in the period.",
    suggestions: ["@every 2m", "@hourly", "@daily", "@weekly"],
  },
  "auto-deploy": {
    group: "Apply",
    label: "Automatic deployment",
    help: (v) =>
      v === "Enabled"
        ? "Ready recommendations are applied when they pass the threshold."
        : "Recommendations wait for Apply Now.",
  },
  "auto-deploy.threshold": {
    group: "Apply",
    label: "Auto-deploy threshold",
    help: (v) => `Only apply when a request changes by at least ${v} and by 10m CPU or 16Mi memory.`,
    suggestions: ["5%", "10%", "20%"],
  },
  "apply-method": {
    group: "Apply",
    label: "Apply method",
    help: () => "Server-side apply on the workload (field manager resourcerer), followed by a rolling update.",
  },
  "cpu.limits-policy": { group: "Limits", label: "CPU limits policy", help: (v) => limits[v] ?? "" },
  "memory.limits-policy": { group: "Limits", label: "Memory limits policy", help: (v) => limits[v] ?? "" },
  "cpu.limit-request-ratio": {
    group: "Limits",
    label: "CPU limit / request ratio",
    help: () => "Used when a policy sets or raises the limit.",
    suggestions: ["1.0", "1.5", "2.0", "4.0"],
  },
  "memory.limit-request-ratio": {
    group: "Limits",
    label: "Memory limit / request ratio",
    help: () => "Used when a policy sets or raises the limit.",
    suggestions: ["1.0", "1.25", "1.5", "2.0"],
  },
  "cpu.requests.min": {
    group: "Bounds",
    label: "CPU request minimum",
    help: () => "Recommendations never go below this.",
    suggestions: ["5m", "10m", "25m", "50m"],
  },
  "memory.requests.min": {
    group: "Bounds",
    label: "Memory request minimum",
    help: () => "Recommendations never go below this.",
    suggestions: ["16Mi", "32Mi", "64Mi", "128Mi"],
  },
  "containers.skip": {
    group: "Bounds",
    label: "Containers not optimized",
    help: () => "Comma-separated container names that keep their resources, such as sidecars managed elsewhere.",
    suggestions: ["istio-proxy,linkerd-proxy", "none"],
  },
  "hpa.preserve-scale-point": {
    group: "HPA",
    label: "Preserve HPA scale point",
    help: (v) =>
      v === "true"
        ? "When CPU requests change, the HPA target utilization changes too, so scaling happens at the same absolute CPU."
        : "CPU is not optimized for workloads whose HPA scales on CPU utilization.",
  },
};

export const meta = (key: string): SettingMeta => SETTINGS[key] ?? { group: "Other", label: key, help: () => "" };
