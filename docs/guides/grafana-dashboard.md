---
title: Grafana dashboard
description: "See how resourcerer is doing across the cluster in Grafana: savings, workloads, applies, load on Prometheus and health."
---

The UI answers whether to apply one workload's recommendation. The Grafana dashboard answers how resourcerer is doing
as a whole, over time: what it would save and what was applied, how far along the workloads are, how much it asks of
Prometheus, and whether it is healthy. It reads the [metrics](../reference/metrics-and-events.md) resourcerer
exports, so Prometheus must scrape them first (see [Prometheus](../configuration/prometheus.md#recommendation-history)).

## Install it

The chart ships the dashboard in a ConfigMap for a Grafana that loads dashboards from labelled ConfigMaps: the sidecar
of the Grafana chart, which kube-prometheus-stack enables by default.

```yaml title="values.yaml"
metrics:
  grafanaDashboard:
    enabled: true
    # The label your sidecar looks for; grafana_dashboard: "1" is the default of both charts.
    labels:
      grafana_dashboard: "1"
    # When the sidecar only watches some namespaces.
    namespace: monitoring
```

Without the sidecar, import `charts/resourcerer/dashboards/resourcerer.json` from the
[repository](https://github.com/gjbravi/resourcerer/blob/main/charts/resourcerer/dashboards/resourcerer.json) in
Grafana (**Dashboards → New → Import**) and pick your Prometheus data source.

## What it shows

The **Namespace** and **Kind** filters at the top narrow the workload panels.

| Row | Answers |
|---|---|
| **Summary** | What applying every recommendation would save each month, what is reclaimable, how many workloads are managed, `Ready`, pending an apply and on auto-deploy. |
| **Savings** | Monthly cost now and with the recommendations over time; reclaimable by namespace and by workload; the workloads that need more capacity, usually to fix throttling or OOM kills. |
| **Capacity** | CPU and memory requested now and with the recommendations, against the cluster's allocatable CPU; the cost rates in use, to tell a price change from a rightsizing. |
| **Workloads** | Workloads by kind and [phase](../concepts/phases.md), the backlog pending an apply, auto-deploy adoption and drifted workloads. |
| **Applies** | Applies and reverts by trigger and result, applies an [apply policy](apply-policies.md) held, and applies something else changed back (see [GitOps](gitops.md#notice-an-apply-that-was-undone)). |
| **Load on Prometheus** | Queries, samples and bytes per second by purpose, query duration, failures and queries in flight: what resourcerer costs your Prometheus. |
| **Server** | Requests and latency by route, errors, how much the cache absorbs, and the [decision models](decision-models.md) by model. |
| **Health** | Components up, the oldest recommendation, the least history, cost calibration, signal and reconcile errors, recommendations computed by phase and how long they take, work queues and versions. |

The costs are the ones the Overview shows: requests × average replicas × rates × 730 for the managed workloads, not
the node bill. See [Costs](../concepts/costs.md).

## Change it

The JSON is generated: edit `hack/grafana-dashboard.py` and run `make dashboard`; `make lint` fails when the two
differ. To try a change, `make kind-up` runs Grafana next to Prometheus, and `make kind-deploy` installs the
dashboard there:

```bash
kubectl --context kind-resourcerer -n monitoring port-forward svc/grafana 3000:80
# http://localhost:3000/d/resourcerer
```
