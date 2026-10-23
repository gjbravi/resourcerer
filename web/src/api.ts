/** Types mirror internal/server/types.go. */
export type Mode = "Off" | "Auto" | "Disabled";
export type Status = "Pending" | "Ready" | "LowConfidence" | "Conflict" | "Error" | "None";
export type ResourceName = "cpu" | "memory";

export interface Resources {
  cpu?: string;
  memory?: string;
}

export interface Container {
  name: string;
  requests: Resources;
  limits: Resources;
  target?: Resources;
  lowerBound?: Resources;
  upperBound?: Resources;
  uncappedTarget?: Resources;
}

export interface Workload {
  namespace: string;
  kind: string;
  name: string;
  replicas: number;
  mode: Mode;
  status: Status;
  message?: string;
  vpa?: string;
  modeSource: string;
  containers: Container[];
}

export interface WorkloadDetail extends Workload {
  pods: { name: string; containers: Container[] }[];
  events: { type: string; reason: string; message: string; time: string }[];
}

export interface Info {
  version: string;
  selection: "all" | "label";
  defaultMode: Mode;
  readOnly: boolean;
}

export interface WorkloadRef {
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

export const workloadPath = (w: WorkloadRef) =>
  `/workloads/${encodeURIComponent(w.namespace)}/${w.kind}/${encodeURIComponent(w.name)}`;

export const api = {
  info: () => call<Info>("GET", "/info"),
  list: () => call<Workload[]>("GET", "/workloads"),
  get: (w: WorkloadRef) => call<WorkloadDetail>("GET", workloadPath(w)),
  setMode: (w: WorkloadRef, mode: Mode) => call<WorkloadDetail>("PUT", `${workloadPath(w)}/mode`, { mode }),
};
