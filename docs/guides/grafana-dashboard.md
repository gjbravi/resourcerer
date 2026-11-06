---
title: Grafana dashboard
description: "An example Grafana dashboard of how resourcerer is doing across the cluster: a starting point to import and adapt."
---

The UI answers whether to apply one workload's recommendation. A dashboard over resourcerer's
[metrics](../reference/metrics-and-events.md) answers how it is doing as a whole, over time: what it would save and
what was applied, how far along the workloads are, how much it asks of Prometheus, and whether it is healthy.

The repository publishes an example: a starting point to import and adapt to your Grafana, your Prometheus and what
your team watches, not part of the chart. Prometheus must scrape resourcerer first (see
[Prometheus](../configuration/prometheus.md#recommendation-history)).

## Import it

The dashboard is `docs/assets/grafana/resourcerer-dashboard.json` in the repository. Take it from the tag of the
version you run, so it matches its metrics:

```bash
RESOURCERER_VERSION=1.0.0
curl -fsSLO https://raw.githubusercontent.com/gjbravi/resourcerer/v${RESOURCERER_VERSION}/docs/assets/grafana/resourcerer-dashboard.json
```

Then, in Grafana, **Dashboards → New → Import** and pick your Prometheus data source. With a Grafana that loads
dashboards from labelled ConfigMaps (the sidecar of the Grafana chart and of kube-prometheus-stack), put the file in
one:

```bash
kubectl -n monitoring create configmap grafana-dashboard-resourcerer --from-file=resourcerer-dashboard.json
kubectl -n monitoring label configmap grafana-dashboard-resourcerer grafana_dashboard=1
```

## Try it with a Grafana of its own

To look at it without touching your Grafana, install one next to your Prometheus with the Grafana Helm chart and
[these values](https://github.com/gjbravi/resourcerer/blob/main/docs/assets/grafana/grafana-values.yaml), which load
the dashboard from the repository. They have no persistence: for trying, not for production.

```bash
helm repo add grafana https://grafana.github.io/helm-charts
curl -fsSL https://raw.githubusercontent.com/gjbravi/resourcerer/v${RESOURCERER_VERSION}/docs/assets/grafana/grafana-values.yaml \
  | envsubst > grafana-values.yaml
helm install grafana grafana/grafana --namespace monitoring --values grafana-values.yaml
kubectl -n monitoring port-forward svc/grafana 3000:80
```

The `admin` password is in the `grafana` Secret:
`kubectl -n monitoring get secret grafana -o jsonpath='{.data.admin-password}' | base64 -d`.

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

Panels that need to know which process a series comes from match it by `instance` to `resourcerer_build_info`, so
they work whatever your release and scrape job are called. A Prometheus shared by several clusters needs a cluster
filter added to the queries.

## Change the example

The JSON is generated: edit `hack/grafana-dashboard.py` and run `make dashboard`; `make lint` fails when the two
differ. `make kind-grafana` installs Grafana on the kind cluster with the dashboard, and loads it again after a change:

```bash
make kind-grafana
kubectl --context kind-resourcerer -n monitoring port-forward svc/grafana 3000:80
# http://localhost:3000/d/resourcerer
```
