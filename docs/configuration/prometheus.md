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

## Check the connection

The UI's **Settings** page shows whether Prometheus answers, its version and its retention. From the command line,
a workload in the `Error` phase carries the query error:

```bash
kubectl get recommendations -A | grep Error
kubectl -n <namespace> get recommendation <name> -o jsonpath='{.status.message}'
```

## Recommendation history

The UI's charts show how the recommended values changed over time from the controller's own metrics,
`resourcerer_recommendation` and `resourcerer_optimization_score`, so Prometheus must scrape the controller:

- with annotation-based discovery, keep `metrics.scrapeAnnotations: true` (the default);
- with the Prometheus Operator, set `metrics.serviceMonitor.enabled: true`, and its `labels` to what your Prometheus
  selects.
