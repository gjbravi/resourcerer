# resourcerer

[![CI](https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml/badge.svg)](https://github.com/gjbravi/resourcerer/actions/workflows/ci.yaml)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)

resourcerer recommends CPU and memory requests and limits for your Kubernetes workloads from their usage history
in Prometheus, shows the projected impact in a UI, and applies recommendations on request or automatically.

It runs entirely in your cluster. There is no machine learning: every recommendation is a documented percentile
of usage plus headroom, so you can see exactly why a value was chosen.

> Status: proof of concept. APIs, annotations and the CRD may change.

## Features

- A recommendation per Deployment, StatefulSet and DaemonSet, regenerated on a schedule spread across workloads.
- Optimization goals per resource (Savings, Balanced, Reliability), limit policies, minimums, skipped sidecars.
- Reliability response: after an OOM kill, memory is raised above the limit that killed the container.
- HPA-aware: when the HPA scales on CPU utilization, its target is adjusted with the requests to keep the scale point.
- Settings cascade: Helm defaults → namespace annotations → workload annotations, all GitOps-friendly.
- UI: cluster overview, workload list with bulk apply, per-workload impact, usage percentile charts with current and
  recommended requests over time, Kustomize/Helm/kubectl patches, and effective settings with their source.
- Apply Now, Revert and auto-deploy with thresholds, through server-side apply with field manager `resourcerer`.
- No database: state lives in a `Recommendation` CRD, history in Prometheus.

## Architecture

One image, three components, each with its own ServiceAccount:

| Component | Does | Writes |
|---|---|---|
| `controller` | Discovers workloads, resolves settings, queries Prometheus, computes recommendations, exports them as metrics | `Recommendation` objects, Events |
| `applier` | Applies or reverts on request, and auto-deploys | Workload resources, HPA targets |
| `server` | REST API and UI | Apply requests on `Recommendation`, `resourcerer.io/*` annotations |

The applier is the only component that changes workloads; disable it (`applier.enabled=false`) for a recommend-only
install, or set `readOnly=true` to make the UI read-only.

Authentication is `anonymous` for now, like Kiali's anonymous strategy: anyone who reaches the UI can use it. Keep the
Service internal (port-forward) or put an authenticating proxy in front.

## Requirements

- Kubernetes 1.33+
- A Prometheus-compatible API (Prometheus, Thanos, Mimir, VictoriaMetrics) with cAdvisor and kube-state-metrics,
  retaining at least the longest learning period (7 days by default), and scraping the controller's metrics Service
  (annotated with `prometheus.io/scrape` by default) for recommendation history.

## Install

```bash
helm install resourcerer ./charts/resourcerer -n resourcerer --create-namespace \
  --set prometheus.url=http://prometheus-server.monitoring.svc
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

See [charts/resourcerer/values.yaml](charts/resourcerer/values.yaml) for all options.

## How recommendations are computed

For each container, over the learning period, using the busiest pod at each sample:

| Goal | CPU request | Memory request |
|---|---|---|
| Savings | p85 + 5% | p99 + 10% |
| Balanced (default) | p95 + 10% | max + 15% |
| Reliability | p99 + 20% | max + 30% |

Then the minimum is applied and values are rounded (CPU to 1m below 100m and 5m above, memory to 1Mi). Limits
follow the limits policy:

| Policy | Limits |
|---|---|
| `RequestsRaiseLimitsIfNeeded` (default) | Kept, unless the new request exceeds them: then request × ratio |
| `RequestsOnly` | Kept |
| `RequestsAndLimits` | request × ratio |
| `DoNotOptimize` | Requests and limits kept |

Phases: **Pending** (collecting history), **Learning** (preliminary, learning period not covered), **Ready**,
**Warning** (e.g. scaled to zero), **Unsupported** (HPA scales on CPU and memory utilization), **Error** (metrics
unavailable). Auto-deploy only applies Ready recommendations.

The score (0–100) is how closely current requests match the recommendation, weighted by cost; net impact is requests
× average replicas, priced with `cost.cpuCoreHour` and `cost.memoryGiBHour`.

## Settings

Set cluster defaults in the Helm values under `defaults`; override them per namespace or workload with annotations:

```bash
kubectl annotate namespace payments resourcerer.io/auto-deploy=Enabled
kubectl -n payments annotate deployment ledger resourcerer.io/memory.optimization-goal=Reliability
```

| Annotation (`resourcerer.io/…`) | Default | Values |
|---|---|---|
| `cpu.optimization-goal`, `memory.optimization-goal` | `Balanced` | `Savings`, `Balanced`, `Reliability` |
| `learning-period` | `P7D` | ISO 8601 duration, e.g. `PT1H`, `P1D`, `P14D` |
| `schedule` | `@daily` | `@hourly`, `@daily`, `@weekly`, `@every <duration>` |
| `auto-deploy` | `Disabled` | `Enabled`, `Disabled` |
| `auto-deploy.threshold` | `10%` | Minimum relative change (also at least 10m CPU or 16Mi memory) |
| `apply-method` | `PatchImmediateRollout` | `PatchImmediateRollout` |
| `cpu.limits-policy`, `memory.limits-policy` | `RequestsRaiseLimitsIfNeeded` | See above |
| `cpu.limit-request-ratio`, `memory.limit-request-ratio` | `2.0`, `1.5` | Number ≥ 1 |
| `cpu.requests.min`, `memory.requests.min` | `10m`, `16Mi` | Quantity |
| `containers.skip` | `istio-proxy,linkerd-proxy` | Comma-separated names, or `none` |
| `hpa.preserve-scale-point` | `true` | `true`, `false` |

Which workloads get a recommendation: `selection: all` (default; opt out with the label
`resourcerer.io/enabled=false` on the workload or namespace) or `selection: label` (opt in with
`resourcerer.io/enabled=true`).

### GitOps

Applied values are owned by the field manager `resourcerer`. With Argo CD, ignore them so they are not reverted:

```yaml
ignoreDifferences:
  - group: apps
    kind: Deployment
    managedFieldsManagers: [resourcerer]
```

Or copy the patch from the workload's **Patches** tab into Git instead of applying from the UI.

## Development

```bash
make kind-up        # kind cluster with Prometheus, kube-state-metrics, metrics-server and demo workloads
make kind-deploy    # build the image, load it, install the chart with hack/resourcerer-values.yaml
make e2e            # end-to-end checks
kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80
```

The kind values use a 15-minute learning period and a 2-minute schedule, so recommendations show up within minutes.
Behind a TLS-intercepting proxy, pass its root CA: `EXTRA_CA_CERT=/path/ca.pem make kind-up`.

To run the components locally against the kind cluster: `make prom-forward` in one terminal, `make run` in another,
and `make web-dev` for the UI with hot reload.

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache-2.0](LICENSE)
