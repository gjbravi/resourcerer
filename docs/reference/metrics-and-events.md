---
title: Metrics and events
description: The Prometheus metrics and Kubernetes events resourcerer emits, and what to alert on.
---

## Metrics

The controller and the applier serve metrics on port `8081`, behind the Services `<release>-metrics` and
`<release>-applier-metrics`. The chart annotates both for scraping (`metrics.scrapeAnnotations`), or creates a
`ServiceMonitor` (`metrics.serviceMonitor.enabled`). Both also expose the standard controller-runtime and Go metrics.

### Controller

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_recommendation` | gauge | `namespace`, `workload_kind`, `workload`, `container`, `resource`, `type` | Recommended requests and limits per container, in cores and bytes. The UI charts recommendation history from it. |
| `resourcerer_optimization_score` | gauge | `namespace`, `workload_kind`, `workload` | How closely current requests match the recommendation, 0–100. |

### Applier

| Metric | Type | Labels | Description |
|---|---|---|---|
| `resourcerer_apply_total` | counter | `result`: `Succeeded`, `Failed`, `Rejected` | Applies and reverts by result. |
| `resourcerer_apply_deferred_total` | counter | `reason`: `window`, `freeze`, `cooldown`, `concurrency`, `namespace`, `rate`, `spread`, `rollout`, `policy` | Automatic applies held by an apply policy, by the rule that held them. |
| `resourcerer_apply_reverted_total` | counter | `reason`: `manual` | Reverts. |
| `resourcerer_applypolicy_slots_available` | gauge | `policy` | Applies a policy's `maxConcurrent` still allows. |

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

```bash
kubectl -n payments events --for deployment/checkout
kubectl get events -A --field-selector reason=ApplyFailed
```

## Logs

Each component logs in JSON to stdout. Useful lines: `read usage history` errors from the controller (Prometheus
problems), and the applier's message at start when the apply policy CRDs are missing.
