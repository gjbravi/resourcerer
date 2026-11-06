---
title: Helm values
description: Every value of the resourcerer chart, with its default.
---

The chart's `values.schema.json` rejects unknown keys. The source of truth, with comments, is
[`charts/resourcerer/values.yaml`](https://github.com/gjbravi/resourcerer/blob/main/charts/resourcerer/values.yaml).

The CRDs are installed by a chart of their own, `resourcerer-crds`, which has no values and must be installed before
this one. See
[the CRDs are a release of their own](../getting-started/installation.md#the-crds-are-a-release-of-their-own).

## General

| Value | Default | Description |
|---|---|---|
| `clusterName` | `kubernetes` | Cluster name shown in the UI. |
| `selection` | `all` | `all`: every supported workload outside `excludedNamespaces`; opt out with `resourcerer.io/enabled=false`. `label`: only workloads with `resourcerer.io/enabled=true` on the workload or its namespace. |
| `excludedNamespaces` | `[kube-system, kube-public, kube-node-lease]` | Namespaces never managed. The release namespace is always added. |
| `defaults` | `{}` | Cluster defaults for [settings](settings.md), keyed by annotation name without `resourcerer.io/`. |
| `readOnly` | `false` | Refuse every change from the UI and API; drops the server's `patch` permissions. |

## Authentication

See [Exposing the UI](../guides/exposing-the-ui.md) for the two ways to put an identity in front of the server.

| Value | Default | Description |
|---|---|---|
| `auth.strategy` | `anonymous` | `anonymous`: anyone who reaches the server may use it, with the server's own permissions. `oidc`: users sign in at an OpenID Connect provider, and their groups decide the role. |
| `auth.cookieSecure` | `auto` | `Secure` on the session cookie: `auto` (when the request is HTTPS), `true` (behind a TLS-terminating proxy), `false` (plain HTTP only). |
| `auth.oidc.issuer`, `auth.oidc.clientID` | unset | The provider (an `http(s)` URL) and the confidential client resourcerer presents there. Required with `oidc`. |
| `auth.oidc.clientSecretSecretRef`, `auth.session.keySecretRef` | unset | Secrets with the client secret and the session key (standard base64 of 32 bytes: `openssl rand -base64 32`). Required with `oidc`; the chart renders references, never the material. |
| `auth.session.keyPreviousSecretRef` | unset | The key being retired, still accepted while a rotation is in flight. |
| `auth.oidc.allowedGroups` | `[]` | Which groups may use resourcerer. Empty gives read access to every user the provider authenticates. |
| `auth.oidc.roles.editorGroups` | `[]` | Which groups may apply, revert and change settings. Empty means nobody can. |
| `auth.oidc.roles.anyAuthenticatedIsViewer` | `true` | Whether an admitted user who is not an editor may read. |
| `auth.oidc.scopes`, `auth.oidc.usernameClaim`, `auth.oidc.groupsClaim` | `openid profile email`, `preferred_username`, `groups` | What to ask the provider for, and which claims name the user and the groups. |
| `auth.oidc.redirectURL` | derived | Absolute callback URL, for a proxy that rewrites the Host the server sees. Otherwise derived from the request, so every way of reaching the server needs its own callback registered at the provider. |
| `auth.oidc.sessionTTL` | `12h` | How long a session is valid without visiting the provider again: whole minutes (`5m` to `10080m`) or hours (`1h` to `168h`). Nothing revokes it sooner. |

## Prometheus

| Value | Default | Description |
|---|---|---|
| `prometheus.url` | `http://prometheus-server.monitoring.svc` | Prometheus-compatible API with cAdvisor and kube-state-metrics data. |
| `prometheus.rateWindow` | `5m` | `rate()` window for CPU usage; at least 4× the scrape interval. |
| `prometheus.minStep` | `30s` | Finest query resolution, about the scrape interval. |
| `prometheus.timeout` | `2m` | Longest a query may take, response included; `0s` means no limit. |

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
| `server.ingress.enabled` | `false` | An Ingress for the server. Under `oidc` the server authenticates users itself; under `anonymous` put an authenticating proxy in front: see [Exposing the UI](../guides/exposing-the-ui.md). |
| `server.ingress.className`, `annotations`, `hosts`, `tls` | | Standard Ingress fields. |
| `<component>.extraArgs` | `[]` | Extra command-line arguments. |
| `<component>.podLabels`, `<component>.podAnnotations` | `{}` | On top of the global `podLabels` and `podAnnotations`. |

## Metrics

| Value | Default | Description |
|---|---|---|
| `metrics.scrapeAnnotations` | `true` | `prometheus.io/scrape` annotations on the controller's, the server's and the applier's metrics Services. |
| `metrics.serviceMonitor.enabled` | `false` | A Prometheus Operator `ServiceMonitor` for all of them instead. |
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
