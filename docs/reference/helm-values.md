---
title: Helm values
description: Every value of the resourcerer chart, with its default.
---

The chart's `values.schema.json` rejects unknown keys. The source of truth, with comments, is
[`charts/resourcerer/values.yaml`](../../charts/resourcerer/values.yaml).

## General

| Value | Default | Description |
|---|---|---|
| `clusterName` | `kubernetes` | Cluster name shown in the UI. |
| `selection` | `all` | `all`: every supported workload outside `excludedNamespaces`; opt out with `resourcerer.io/enabled=false`. `label`: only workloads with `resourcerer.io/enabled=true` on the workload or its namespace. |
| `excludedNamespaces` | `[kube-system, kube-public, kube-node-lease]` | Namespaces never managed. The release namespace is always added. |
| `defaults` | `{}` | Cluster defaults for [settings](settings.md), keyed by annotation name without `resourcerer.io/`. |
| `readOnly` | `false` | Refuse every change from the UI and API; drops the server's `patch` permissions. |
| `auth.strategy` | `anonymous` | Only `anonymous` for now. See [Exposing the UI](../guides/exposing-the-ui.md). |

## Prometheus

| Value | Default | Description |
|---|---|---|
| `prometheus.url` | `http://prometheus-server.monitoring.svc` | Prometheus-compatible API with cAdvisor and kube-state-metrics data. |
| `prometheus.rateWindow` | `5m` | `rate()` window for CPU usage; at least 4× the scrape interval. |
| `prometheus.minStep` | `30s` | Finest query resolution, about the scrape interval. |

## Cost

| Value | Default | Description |
|---|---|---|
| `cost.source` | `auto` | `static`, `karpenter` or `auto`. See [Cost estimates](../configuration/cost-estimates.md). |
| `cost.allocation` | `cluster` | `cluster` or `node` (needs Karpenter prices). |
| `cost.rates.cpuCoreHour` | `0.0316` | USD per CPU core-hour; with Karpenter, the CPU to memory ratio. |
| `cost.rates.memoryGiBHour` | `0.0042` | USD per GiB-hour. |
| `cost.karpenter.refreshInterval` | `10m` | How often rates are recalibrated; at least `30s`. |

## Reliability

| Value | Default | Description |
|---|---|---|
| `reliability.builtin.throttling` | `{enabled: true, warn: 0.10}` | Share of CFS periods throttled. Critical from 0.25. |
| `reliability.builtin.memoryHeadroom` | `{enabled: true, warn: 0.90}` | Peak working set as a share of the memory limit. |
| `reliability.builtin.restarts` | `{enabled: true}` | Restarts other than OOM kills. |
| `reliability.builtin.crashLoop` | `{enabled: true}` | Containers in CrashLoopBackOff. |
| `reliability.builtin.hpaAtMax` | `{enabled: true}` | HPA at its maximum replicas. |
| `reliability.builtin.bestEffort` | `{enabled: true}` | Pods without requests or limits. |
| `reliability.signals` | `[]` | Signals declared as PromQL templates, at most 20. See [Reliability signals](../guides/reliability-signals.md#declared-signals). |

## Apply policies

| Value | Default | Description |
|---|---|---|
| `applyPolicies.cluster` | `[]` | `ClusterApplyPolicy` objects created with the release: `name` and `spec`. See [Apply policies](../guides/apply-policies.md). |

## Components

| Value | Default | Description |
|---|---|---|
| `controller.replicas` | `1` | Above 1, leader election: one replica works. |
| `controller.resources` | requests `20m`/`64Mi`, limit `256Mi` | |
| `applier.enabled` | `true` | The only component that writes workloads. `false` for a recommend-only install. |
| `applier.replicas` | `1` | Above 1, leader election. |
| `applier.resources` | requests `10m`/`48Mi`, limit `128Mi` | |
| `server.replicas` | `1` | Every replica serves. |
| `server.ui` | `true` | Serve the embedded UI; the API is served either way. |
| `server.resources` | requests `10m`/`48Mi`, limit `192Mi` | |
| `server.service.type`, `server.service.port` | `ClusterIP`, `80` | |
| `server.ingress.enabled` | `false` | An Ingress for the server. It has no authentication: see [Exposing the UI](../guides/exposing-the-ui.md). |
| `server.ingress.className`, `annotations`, `hosts`, `tls` | | Standard Ingress fields. |
| `<component>.extraArgs` | `[]` | Extra command-line arguments. |
| `<component>.podLabels`, `<component>.podAnnotations` | `{}` | On top of the global `podLabels` and `podAnnotations`. |

## Metrics

| Value | Default | Description |
|---|---|---|
| `metrics.scrapeAnnotations` | `true` | `prometheus.io/scrape` annotations on the controller's and the applier's metrics Services. |
| `metrics.serviceMonitor.enabled` | `false` | A Prometheus Operator `ServiceMonitor` for both instead. |
| `metrics.serviceMonitor.interval` | `30s` | |
| `metrics.serviceMonitor.labels` | `{}` | Labels your Prometheus selects ServiceMonitors by. |

## Pods

| Value | Default | Description |
|---|---|---|
| `image.repository` | `ghcr.io/gjbravi/resourcerer` | |
| `image.tag` | the chart's `appVersion` | |
| `image.pullPolicy` | `IfNotPresent` | |
| `imagePullSecrets` | `[]` | |
| `nameOverride`, `fullnameOverride` | `""` | |
| `rbac.create` | `true` | ServiceAccounts, ClusterRoles and bindings. See [Permissions](permissions.md). |
| `podLabels`, `podAnnotations` | `{}` | For every component's pods. |
| `podSecurityContext` | `runAsNonRoot: true`, `seccompProfile: RuntimeDefault` | |
| `securityContext` | no privilege escalation, read-only root filesystem, all capabilities dropped | |
| `nodeSelector`, `tolerations`, `affinity` | empty | |
| `priorityClassName` | `""` | |
