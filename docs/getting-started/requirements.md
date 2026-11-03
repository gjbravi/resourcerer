---
title: Requirements
description: The Kubernetes version, the Prometheus data and the permissions resourcerer needs.
---

## Kubernetes

- **Kubernetes 1.33 or later.** The chart refuses older versions.
- Workloads of kind **Deployment**, **StatefulSet**, **DaemonSet** and, when its CRD is installed, **Argo Rollout**.
  Jobs, CronJobs and bare pods are not managed.

## Prometheus

resourcerer keeps no metrics of its own: it reads the history your Prometheus already has. Any API compatible with
Prometheus' HTTP API works: **Prometheus**, **Thanos**, **Mimir**, **VictoriaMetrics**.

| Data | From | Used for |
|---|---|---|
| `container_cpu_usage_seconds_total`, `container_memory_working_set_bytes` | cAdvisor (kubelet) | Usage per container, the base of every recommendation |
| `container_cpu_cfs_throttled_periods_total`, `container_cpu_cfs_periods_total`, `container_spec_memory_limit_bytes` | cAdvisor | CPU throttling and memory headroom signals |
| `kube_pod_owner`, `kube_replicaset_owner`, `kube_pod_info`, `kube_pod_status_phase` | kube-state-metrics | Which pods belong to which workload, and where they ran |
| `kube_pod_container_resource_requests`, `kube_pod_container_status_restarts_total`, `kube_node_status_allocatable` | kube-state-metrics | Requests over time, restarts, node capacity |
| `karpenter_nodes_allocatable`, `karpenter_cloudprovider_instance_type_offering_price_estimate` | Karpenter (optional) | [Costs calibrated to node prices](../guides/karpenter-costs.md) |
| `resourcerer_recommendation`, `resourcerer_optimization_score` | resourcerer's controller | Recommendation history in the UI charts |

Make sure that:

- **Retention covers the learning period.** With the default `P7D`, Prometheus (or Thanos, or Mimir) must keep at least
  7 days. With less, recommendations stay `Learning`.
- **The controller's metrics are scraped.** The chart annotates its metrics Services with `prometheus.io/scrape`, or
  creates a `ServiceMonitor` with `metrics.serviceMonitor.enabled=true` for the Prometheus Operator.
- **`prometheus.rateWindow` is at least 4× the scrape interval** (default `5m`), and `prometheus.minStep` about the
  scrape interval (default `30s`). See [Prometheus](../configuration/prometheus.md).

## Resources

The three components are small. The chart's defaults:

| Component | CPU request | Memory request | Memory limit |
|---|---|---|---|
| controller | 20m | 64Mi | 256Mi |
| applier | 10m | 48Mi | 128Mi |
| server | 10m | 48Mi | 192Mi |

The controller's memory grows with the number of workloads it caches; raise its limit on clusters with thousands of
workloads. The release namespace is always excluded, so resourcerer never rightsizes itself.

## Permissions

The chart creates one ServiceAccount and ClusterRole per component, and needs permission to install CRDs and
cluster-wide RBAC. What each component may do is listed in [Permissions](../reference/permissions.md). In short, only
the **applier** writes workloads, and only their pod template resources and HPA targets.

## Network

- The controller and the server reach the Prometheus URL.
- Nothing else: no outbound traffic to the internet, no inbound traffic except to the server's Service.
