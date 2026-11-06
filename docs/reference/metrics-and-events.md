---
title: Metrics and events
description: The Prometheus metrics and Kubernetes events resourcerer emits, and what to alert on.
---

## Metrics

The controller, the server and the applier serve metrics on port `8081`, behind the Services `<release>-metrics`,
`<release>-server-metrics` and `<release>-applier-metrics`. The chart annotates them for scraping
(`metrics.scrapeAnnotations`), or creates a `ServiceMonitor` (`metrics.serviceMonitor.enabled`). All of them also
expose the standard controller-runtime and Go metrics, and `resourcerer_build_info{version, component}`.

The metrics about workloads carry the workload's own `namespace` and, for `resourcerer_recommendation`, `container`
labels: the scrape must keep them (`honorLabels`, see [Prometheus](../configuration/prometheus.md#recommendation-history)).

### Controller: workloads and savings

Read from the Recommendations when Prometheus scrapes, by the leading replica only.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_recommendation` | gauge | `namespace`, `workload_kind`, `workload`, `container`, `resource`, `type` | Recommended requests and limits per container, in cores and bytes. The UI charts recommendation history from it. |
| `resourcerer_optimization_score` | gauge | `namespace`, `workload_kind`, `workload` | How closely current requests match the recommendation, 0–100. |
| `resourcerer_workloads` | gauge | `workload_kind`, `phase`, `pending`, `auto_deploy` | Managed workloads by kind and [phase](../concepts/phases.md); `pending` is whether the recommendation differs from the requests by more than `auto-deploy.threshold`. |
| `resourcerer_workload_monthly_cost` | gauge | `namespace`, `workload_kind`, `workload`, `state`: `current`, `recommended` | Monthly cost of the requests now and with the recommendation, as the Overview computes it. |
| `resourcerer_workload_requests` | gauge | `namespace`, `workload_kind`, `workload`, `resource`, `state` | Requests × average replicas now and with the recommendation, in cores and bytes. |
| `resourcerer_workload_drifted` | gauge | `namespace`, `workload_kind`, `workload` | Containers whose resources no longer match the last apply: something else changed them since. Only drifted workloads. |
| `resourcerer_recommendation_generated_timestamp_seconds` | gauge | `namespace`, `workload_kind`, `workload` | When the recommendation was last computed. |
| `resourcerer_recommendation_history_coverage_ratio` | gauge | `namespace`, `workload_kind`, `workload` | Usage history the recommendation saw over its learning period, 0 to 1. |
| `resourcerer_settings_warnings` | gauge | `namespace`, `workload_kind`, `workload` | Settings that could not be used and fell back to their default. Only workloads with some. |

### Controller: computing recommendations

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_recommendation_generations_total` | counter | `workload_kind`, `phase` | Recommendations computed, by the phase they ended in; `Error` when Prometheus could not be read. A reconcile that finds nothing due is not counted. |
| `resourcerer_recommendation_generation_duration_seconds` | histogram | `workload_kind` | Time to compute one, Prometheus queries included. |
| `resourcerer_signal_errors_total` | counter | `signal` | Reliability signal queries that failed; the signal is left out of the workload's status. |
| `resourcerer_apply_overwritten_total` | counter | `manager` | Applied recommendations whose resources another field manager changed afterwards, such as a GitOps sync. |
| `resourcerer_cost_calibration` | gauge | `source`, `state` | 1 for where cost rates come from (`static`, `karpenter`) and the state of the last calibration (`ok`, `stale`, `fallback`, `error`); see [Costs](../concepts/costs.md). |
| `resourcerer_cost_rate` | gauge | `resource` | Rates in use, per core-hour for `cpu` and per GiB-hour for `memory`. |
| `resourcerer_cost_cluster_factor` | gauge | | Rates in use over `cost.rates`. |

A recommendation that ends `Pending` or `Error` is computed again after a minute, then two, four and so on, up to its
schedule: a workload scaled to zero, or a Prometheus that is down, is not queried every minute for as long as it lasts.

### Every component: Prometheus queries

What resourcerer asks of Prometheus, so its load can be watched as the cluster grows. The component is the scrape
target's own labels.

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_prometheus_requests_total` | counter | `purpose`, `type`: `range`, `instant`, `code` | Queries sent. `code` is the HTTP status, `error` without a response, `timeout` past `prometheus.timeout`. |
| `resourcerer_prometheus_request_duration_seconds` | histogram | `purpose`, `type` | Time to the whole response. |
| `resourcerer_prometheus_requests_in_flight` | gauge | | Queries not answered yet. |
| `resourcerer_prometheus_response_bytes_total` | counter | `purpose` | Bytes read from responses. |
| `resourcerer_prometheus_samples_total` | counter | `purpose` | Samples the queries returned: what they cost Prometheus to evaluate and send. |

`purpose` is what the query was for: `usage`, `signals`, `allocation` and `pricing` from the controller; the route
that sent it from the server (`workload/series`, `workload/reliability`, `cluster/series`…); `status` for the
Settings page's checks. Queries also carry the `User-Agent` `resourcerer/<version> (<component>)`, for Prometheus's
own logs.

### Server

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_http_requests_total` | counter | `route`, `method`, `code` | API and UI requests. `route` is the API path after `/api/v1/`, with `workload` for `workloads/{ns}/{kind}/{name}`; `ui` for the UI. |
| `resourcerer_http_request_duration_seconds` | histogram | `route`, `method` | Time to answer. |
| `resourcerer_server_cache_requests_total` | counter | `query`, `result`: `hit`, `shared`, `miss` | Lookups of the cache in front of the live queries; `shared` waited for the same query already running. |
| `resourcerer_decision_requests_total` | counter | `model`, `result`: `ok`, `error`, `invalid` | Calls to a [decision model](../guides/decision-models.md); `model` is its `spec.model`. |
| `resourcerer_decision_request_duration_seconds` | histogram | `model` | Time a decision model took to answer. |

### Applier

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_apply_total` | counter | `action`: `Apply`, `Revert`; `trigger`: `manual`, `auto`; `result`: `Succeeded`, `Failed`, `Rejected` | Applies and reverts. `auto` is auto-deploy. |
| `resourcerer_apply_deferred_total` | counter | `reason`: `window`, `freeze`, `cooldown`, `concurrency`, `namespace`, `rate`, `spread`, `rollout`, `policy` | Automatic applies held by an apply policy, by the rule that held them. |
| `resourcerer_apply_reverted_total` | counter | `reason`: `manual` | Reverts. |
| `resourcerer_applypolicy_slots_available` | gauge | `policy` | Applies a policy's `maxConcurrent` still allows. |

The [Grafana dashboard](../guides/grafana-dashboard.md) the chart ships charts most of them.

### Queries worth having

```promql title="Estimated savings per month, over time"
sum(resourcerer_workload_monthly_cost{state="current"}) - sum(resourcerer_workload_monthly_cost{state="recommended"})
```

```promql title="Managed workloads by kind"
sum by (workload_kind) (resourcerer_workloads)
```

```promql title="Prometheus queries per second, by purpose"
sum by (purpose) (rate(resourcerer_prometheus_requests_total[5m]))
```

```promql title="Samples per second resourcerer pulls from Prometheus"
sum(rate(resourcerer_prometheus_samples_total[5m]))
```

### Alerts worth having

```yaml
groups:
  - name: resourcerer
    rules:
      - alert: ResourcererApplyFailures
        expr: increase(resourcerer_apply_total{result="Failed"}[1h]) > 0
        labels: { severity: warning }
        annotations:
          summary: resourcerer failed to apply a recommendation; see the ApplyFailed events.
      - alert: ResourcererPolicyHoldsEverything
        expr: increase(resourcerer_apply_deferred_total{reason="policy"}[1h]) > 0
        labels: { severity: warning }
        annotations:
          summary: An apply policy cannot be read and holds every apply it governs.
      - alert: ResourcererPrometheusErrors
        expr: sum(rate(resourcerer_prometheus_requests_total{code!~"2.."}[15m])) > 0
        for: 30m
        labels: { severity: warning }
        annotations:
          summary: resourcerer's queries to Prometheus are failing; recommendations end in the Error phase.
      - alert: ResourcererStaleRecommendations
        expr: count(time() - resourcerer_recommendation_generated_timestamp_seconds > 3 * 86400) > 0
        labels: { severity: warning }
        annotations:
          summary: Some recommendations have not been computed for three days; adjust 86400 to your schedule.
      - alert: ResourcererCostsUncalibrated
        expr: max(resourcerer_cost_calibration{state="error"}) == 1
        for: 1h
        labels: { severity: warning }
        annotations:
          summary: Cost rates cannot be calibrated from Karpenter prices; costs use the configured rates.
      - alert: ResourcererAppliesOverwritten
        expr: increase(resourcerer_apply_overwritten_total[1d]) > 0
        labels: { severity: info }
        annotations:
          summary: Something changed applied recommendations back; see the RecommendationOverwritten events.
      - alert: ResourcererDown
        expr: absent(up{job=~".*resourcerer.*"} == 1)
        for: 15m
        labels: { severity: warning }
        annotations:
          summary: No resourcerer metrics endpoint is up; recommendations are not being refreshed.
```

Adapt the `job` matcher to how your Prometheus names the scrape job.

## Events

Events are recorded on the **workload**, related to its `Recommendation`:

| Reason | Type | When |
|---|---|---|
| `RecommendationGenerated` | Normal | A new recommendation (a new hash), with its phase and score. |
| `RecommendationApplied` | Normal | An apply succeeded: by whom, the hash and the changes. |
| `RecommendationReverted` | Normal | A revert succeeded. |
| `ApplyDeferred` | Normal | An automatic apply was held by an apply policy, with the policy and the reason. |
| `ApplyFailed` | Warning | An apply or revert failed or was rejected, with the message. |
| `RecommendationOverwritten` | Warning | Another field manager changed the resources an apply set, such as a GitOps sync: who, the hash and the containers. |

```bash
kubectl -n payments events --for deployment/checkout
kubectl get events -A --field-selector reason=ApplyFailed
```

## Logs

Each component logs in JSON to stdout. Useful lines: `read usage history` errors from the controller (Prometheus
problems), and the applier's message at start when the apply policy CRDs are missing.
