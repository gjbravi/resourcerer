/** Types mirror internal/server/types.go. CPU is in cores and memory in bytes. */
export type Resource = "cpu" | "memory";
export type Resources = Partial<Record<Resource, number>>;
export type Phase = "Pending" | "Learning" | "Ready" | "Warning" | "Unsupported" | "Error";

export interface SettingInfo {
  key: string;
  value: string;
  source: "cluster" | "namespace" | "workload";
  annotation: string;
  options?: string[];
}

export interface PrometheusStatus {
  url: string;
  connected: boolean;
  error?: string;
  version?: string;
  retention?: string;
  oldest?: string;
  checks: { name: string; metric: string; ok: boolean }[] | null;
  missing?: string[];
}

export interface Info {
  version: string;
  cluster: string;
  auth: string;
  readOnly: boolean;
  selection: "all" | "label";
  excludedNamespaces: string[];
  defaults: SettingInfo[];
  cost: CostInfo;
  prometheus: PrometheusStatus;
}

export interface CostInfo {
  cpuCoreHour: number;
  memoryGiBHour: number;
  /** The configured cost.source. */
  mode: "auto" | "karpenter" | "static";
  /** The source in use: rates calibrated to Karpenter node prices, or the configured rates. */
  source: "karpenter" | "static";
  /** How the rates are split between workloads: the cluster's average, or the nodes each workload ran on. */
  allocation: "cluster" | "node";
  state: "ok" | "stale" | "fallback" | "error";
  reason?: string;
  nodes?: number;
  spot?: number;
  onDemand?: number;
  hourlyCost?: number;
  onDemandHourlyCost?: number;
  calibratedAt?: string;
  /** How often calibrated rates are recomputed, as a Go duration. */
  refresh?: string;
}

/** How a workload's costs were priced; factors multiply the configured rates and are decimal strings. */
export interface Pricing {
  allocation: "cluster" | "node";
  factor: string;
  clusterFactor: string;
  coveragePercent?: number;
  capacity?: Record<string, number>;
  instanceTypes?: { name: string; percent: number }[];
  /** Why node allocation fell back to the cluster's rates. */
  reason?: string;
}

export interface Totals {
  cpu: number;
  memory: number;
  monthlyCost: number;
}

export interface Summary {
  namespace: string;
  kind: string;
  name: string;
  replicas: number;
  avgReplicas: number;
  phase: Phase;
  message?: string;
  score?: { overall: number; cpu: number; memory: number };
  current: Totals;
  recommended: Totals;
  pending: boolean;
  autoDeploy: boolean;
  oomKills: number;
  reliability?: Reliability;
  noRequests: boolean;
  under: boolean;
  hpa?: string[];
  generatedAt?: string;
  lastAppliedAt?: string;
  hash?: string;
  namespaceAnnotated: boolean;
}

export interface Reliability {
  containers?: { name: string; cpuThrottledPercent?: number; memoryPeakPercent?: number }[];
  restarts?: number;
  crashLooping?: string[];
  bestEffortPods?: number;
  hpaAtMax?: boolean;
}

export interface Check {
  id: "throttling" | "memory" | "oom" | "restarts" | "crashloop" | "hpa" | "qos";
  name: string;
  container?: string;
  pod?: string;
  status: "ok" | "warn" | "bad";
  unit: "ratio" | "count" | "text";
  value: number;
  display: string;
  threshold?: number;
  detail: string;
  source: string;
  query?: string;
  series?: Point[];
}

export interface ReliabilityReport {
  window: string;
  from: number;
  to: number;
  step: number;
  generatedAt: string;
  checks: Check[];
}

export type ReliabilityWindow = "learning" | "1d" | "6h" | "1h" | "15m";

export interface Container {
  name: string;
  skipped: boolean;
  currentRequests: Resources;
  currentLimits: Resources;
  recommendedRequests: Resources;
  recommendedLimits: Resources;
  basis?: Partial<Record<Resource, string>>;
}

export interface Change {
  name: string;
  fromRequests: Resources;
  toRequests: Resources;
  fromLimits: Resources;
  toLimits: Resources;
}

export interface Detail extends Summary {
  /** Pods per capacity type of their Karpenter node. */
  capacity?: Record<string, number>;
  pricing?: Pricing;
  nextScheduleAt?: string;
  containers: Container[] | null;
  hpaRecommendation?: { name: string; current: number; recommended: number; clamped: boolean };
  settings: SettingInfo[] | null;
  warnings?: string[];
  history: { time: string; action: "Apply" | "Revert"; by: string; hash?: string; containers: Change[] | null }[];
  events: { type: string; reason: string; message: string; time: string }[];
  lastApplyResult?: { time: string; result: "Succeeded" | "Failed" | "Rejected"; message?: string };
  applyMethod: string;
  templateKind?: string;
  templateName?: string;
}

export interface Point {
  t: number;
  v: number;
}

export interface Series {
  from: number;
  to: number;
  bands: { t: number; p50: number; p75: number; p95: number }[];
  lines: Record<string, Point[]>;
}

export type Range = "1h" | "1d" | "1w" | "2w" | "1m";

export interface Ref {
  namespace: string;
  kind: string;
  name: string;
}

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}

async function call<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(`/api/v1${path}`, {
    method,
    headers: body === undefined ? undefined : { "Content-Type": "application/json" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new ApiError(res.status, (data as { error?: string }).error ?? res.statusText);
  return data as T;
}

export const refKey = (w: Ref) => `${w.namespace}/${w.kind}/${w.name}`;
const path = (w: Ref) => `/workloads/${encodeURIComponent(w.namespace)}/${w.kind}/${encodeURIComponent(w.name)}`;

export const api = {
  info: () => call<Info>("GET", "/info"),
  list: () => call<Summary[]>("GET", "/workloads"),
  get: (w: Ref) => call<Detail>("GET", path(w)),
  series: (w: Ref, container: string, resource: Resource, range: Range) =>
    call<Series>(
      "GET",
      `${path(w)}/series?container=${encodeURIComponent(container)}&resource=${resource}&range=${range}`,
    ),
  impact: (w: Ref, resource: Resource, range: Range) =>
    call<Series>("GET", `${path(w)}/impact?resource=${resource}&range=${range}`),
  reliability: (w: Ref, window: ReliabilityWindow) =>
    call<ReliabilityReport>("GET", `${path(w)}/reliability?window=${window}`),
  clusterSeries: (resource: Resource, range: Range) =>
    call<Series>("GET", `/cluster/series?resource=${resource}&range=${range}`),
  apply: (w: Ref, hash: string) => call<unknown>("POST", `${path(w)}/apply`, { hash }),
  revert: (w: Ref) => call<unknown>("POST", `${path(w)}/revert`),
  regenerate: (w: Ref) => call<unknown>("POST", `${path(w)}/regenerate`),
  setWorkloadSetting: (w: Ref, key: string, value: string | null) =>
    call<unknown>("PUT", `${path(w)}/settings`, { key, value }),
  setNamespaceSetting: (ns: string, key: string, value: string | null) =>
    call<unknown>("PUT", `/namespaces/${encodeURIComponent(ns)}/settings`, { key, value }),
};
