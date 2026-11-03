---
title: Settings and annotations
description: Every resourcerer.io annotation and label, with its default and accepted values.
---

Settings resolve per key from the built-in default, then the chart's `defaults` (keyed without the `resourcerer.io/`
prefix), then namespace annotations, then workload annotations. An invalid value is ignored with a warning and the
previous level wins. See [Configure resourcerer](../configuration/overview.md).

## Optimization settings

| Annotation (`resourcerer.io/…`) | Default | Values | Details |
|---|---|---|---|
| `cpu.optimization-goal` | `Balanced` | `Savings`, `Balanced`, `Reliability` | [Goals](../configuration/goals-and-limits.md#optimization-goal) |
| `memory.optimization-goal` | `Balanced` | `Savings`, `Balanced`, `Reliability` | [Goals](../configuration/goals-and-limits.md#optimization-goal) |
| `learning-period` | `P7D` | ISO 8601 duration: `P<n>D`, `P<n>W`, `PT<n>H`, `PT<n>M`, combined like `P1DT12H` | [Learning period](../configuration/learning-period-and-schedule.md#learning-period) |
| `schedule` | `@daily` | `@hourly`, `@daily`, `@weekly`, `@every <duration>` (at least `1m`) | [Schedule](../configuration/learning-period-and-schedule.md#schedule) |
| `auto-deploy` | `Disabled` | `Enabled`, `Disabled` | [Auto-deploy](../guides/auto-deploy.md) |
| `auto-deploy.threshold` | `10%` | A percentage from `0%` to `100%`; a change also needs at least 10m CPU or 16Mi memory | [Auto-deploy](../guides/auto-deploy.md) |
| `apply-method` | `PatchImmediateRollout` | `PatchImmediateRollout` | [Auto-deploy](../guides/auto-deploy.md) |
| `cpu.limits-policy` | `RequestsRaiseLimitsIfNeeded` | `RequestsRaiseLimitsIfNeeded`, `RequestsOnly`, `RequestsAndLimits`, `DoNotOptimize` | [Limits](../configuration/goals-and-limits.md#limits-policy) |
| `memory.limits-policy` | `RequestsRaiseLimitsIfNeeded` | same | [Limits](../configuration/goals-and-limits.md#limits-policy) |
| `cpu.limit-request-ratio` | `2.0` | A number ≥ 1 | [Limits](../configuration/goals-and-limits.md#limits-policy) |
| `memory.limit-request-ratio` | `1.5` | A number ≥ 1 | [Limits](../configuration/goals-and-limits.md#limits-policy) |
| `cpu.requests.min` | `10m` | A Kubernetes quantity | [Minimums](../configuration/goals-and-limits.md#minimum-requests) |
| `memory.requests.min` | `16Mi` | A Kubernetes quantity | [Minimums](../configuration/goals-and-limits.md#minimum-requests) |
| `containers.skip` | `istio-proxy,linkerd-proxy` | Comma-separated container names, or `none` | [Skipped containers](../configuration/goals-and-limits.md#containers-to-skip) |
| `hpa.preserve-scale-point` | `true` | `true`, `false` | [HPAs](../concepts/workloads.md#hpas) |

## Reliability signal overrides

Set on a namespace or a workload; the workload's win over the namespace's. They are not part of the cascade above:
the signals themselves are declared in the chart values.

| Annotation | Value | Effect |
|---|---|---|
| `resourcerer.io/reliability.disabled` | Comma-separated signal ids, built-in or declared | Those signals do not apply to the namespace or workload. |
| `resourcerer.io/reliability.<id>.warn` | A number | The warning level of a declared signal. |
| `resourcerer.io/reliability.<id>.critical` | A number | The critical level of a declared signal. |

See [Reliability signals](../guides/reliability-signals.md#overrides-per-namespace-or-workload).

## Labels

| Label | On | Effect |
|---|---|---|
| `resourcerer.io/enabled=false` | A namespace or workload | With `selection: all`, takes it out. |
| `resourcerer.io/enabled=true` | A namespace or workload | With `selection: label`, brings it in. |

See [Workloads](../concepts/workloads.md#which-workloads-are-managed).

## Setting them

```bash
kubectl annotate namespace payments resourcerer.io/auto-deploy=Enabled
kubectl -n payments annotate deployment checkout resourcerer.io/memory.optimization-goal=Reliability
kubectl -n payments annotate deployment checkout resourcerer.io/memory.optimization-goal-   # remove
```

```yaml title="values.yaml"
defaults:
  auto-deploy: Disabled
  learning-period: P14D
```

The effective values and their sources are in the workload page's **Settings** tab and in `status.settings`:

```bash
kubectl -n payments get recommendation deployment-checkout -o jsonpath='{.status.settings}' | jq
```
