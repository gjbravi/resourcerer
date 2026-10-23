import type { Container, WorkloadDetail } from "../api";

export function container(over: Partial<Container> = {}): Container {
  return {
    name: "api",
    requests: { cpu: "1", memory: "1Gi" },
    limits: { cpu: "2", memory: "2Gi" },
    target: { cpu: "126m", memory: "74Mi" },
    lowerBound: { cpu: "50m", memory: "60Mi" },
    upperBound: { cpu: "1500m", memory: "1Gi" },
    ...over,
  };
}

export function workload(over: Partial<WorkloadDetail> = {}): WorkloadDetail {
  return {
    namespace: "shop",
    kind: "Deployment",
    name: "api",
    replicas: 2,
    mode: "Off",
    status: "Ready",
    vpa: "deployment-api",
    modeSource: "default",
    containers: [container()],
    pods: [],
    events: [],
    ...over,
  };
}
