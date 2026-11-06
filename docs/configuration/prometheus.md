---
title: Prometheus
description: Connect resourcerer to Prometheus, Thanos, Mimir or VictoriaMetrics, and tune the query window and resolution.
---

```yaml title="values.yaml"
prometheus:
  url: http://prometheus-server.monitoring.svc
  rateWindow: 5m
  minStep: 30s
```

| Value | Default | Description |
|---|---|---|
| `prometheus.url` | `http://prometheus-server.monitoring.svc` | Base URL of a Prometheus-compatible API, without `/api/v1`. |
| `prometheus.rateWindow` | `5m` | The `rate()` window for CPU usage. At least 4× the scrape interval, so every window holds enough samples. |
| `prometheus.minStep` | `30s` | The finest query resolution, about the scrape interval. Longer ranges use coarser steps. |

The data resourcerer reads is listed in [Requirements](../getting-started/requirements.md#prometheus).

## Compatible backends

| Backend | URL to use |
|---|---|
| Prometheus | The server's Service, such as `http://prometheus-server.monitoring.svc` or `http://prometheus-operated.monitoring.svc:9090` |
| Thanos | The Querier (or Query Frontend), such as `http://thanos-query.monitoring.svc:9090` |
| Mimir / Cortex | The query path including the Prometheus prefix, such as `http://mimir-query-frontend.mimir.svc:8080/prometheus` |
| VictoriaMetrics | `vmselect` with its tenant path, or single-node `http://victoria-metrics.monitoring.svc:8428` |

Where the history is split, use the endpoint that sees all of it: a Prometheus with a short retention behind Thanos
should be queried through Thanos, or recommendations stay `Learning`.

### When a query fails

A failed usage query puts the workload in the `Error` phase ("Metrics unavailable: …") and is retried a minute later;
the last recommended values stay in the status until a run succeeds.

Thanos can answer with a **partial response** when a store is down, flagged by a warning. The node-cost queries behind
`cost.allocation: node` refuse those, so a workload is never priced from part of its placement; it falls back to the
cluster's rates and says why. Set your Querier to fail on partial responses (`--no-query.partial-response`) if you want
the same for usage history.

### Authentication

`prometheus.url` is called without credentials. For an API that needs them, put a proxy that adds them in front, in
the cluster, or expose a read-only endpoint to the `resourcerer` namespace with a NetworkPolicy.

## Load on Prometheus

resourcerer queries Prometheus in two places. The controller computes each recommendation on its schedule (`@daily`
by default, spread over the day by workload); a recommendation that ends `Pending` or `Error` is tried again after a
minute, then two, four and so on up to the schedule, so workloads scaled to zero or a Prometheus that is down do not
turn into a query every minute. The server runs live queries when someone opens a workload page, behind a cache that
shares identical queries and keeps answers for 30 seconds or more.

Every query is bounded by `prometheus.timeout` (`2m`) and counted by what it was for in
`resourcerer_prometheus_requests_total`, `resourcerer_prometheus_samples_total` and the other
[Prometheus query metrics](../reference/metrics-and-events.md#every-component-prometheus-queries). Queries carry the
`User-Agent` `resourcerer/<version> (<component>)`, so they can be told apart in Prometheus's, Thanos's or Mimir's
own query logs.

```promql title="Queries per second resourcerer sends, by purpose"
sum by (purpose) (rate(resourcerer_prometheus_requests_total[5m]))
```

## Check the connection

The UI's **Settings** page shows whether Prometheus answers, its version and its retention. From the command line,
a workload in the `Error` phase carries the query error:

```bash
kubectl get recommendations -A | grep Error
kubectl -n <namespace> get recommendation <name> -o jsonpath='{.status.message}'
```

## Recommendation history

The UI's charts show how the recommended values changed over time from the controller's own metrics,
`resourcerer_recommendation` and `resourcerer_optimization_score`, so Prometheus must scrape the controller. The
other [metrics](../reference/metrics-and-events.md) come with the same scrape:

- with annotation-based discovery, keep `metrics.scrapeAnnotations: true` (the default);
- with the Prometheus Operator, set `metrics.serviceMonitor.enabled: true`, and its `labels` to what your Prometheus
  selects.

Either way the scrape must keep the metrics' own `namespace` and `container` labels, which name the workload, not
resourcerer's pods: the chart's `ServiceMonitor` sets `honorLabels: true`, and an annotation-based job needs
`honor_labels: true`. Without it Prometheus renames them `exported_namespace` and `exported_container`, and the
charts show no recommended lines.
