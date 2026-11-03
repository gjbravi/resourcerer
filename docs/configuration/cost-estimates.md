---
title: Cost estimates
description: Configure the rates, Karpenter calibration and per-node allocation behind every cost and saving.
---

```yaml title="values.yaml"
cost:
  source: auto          # static | karpenter | auto
  allocation: cluster   # cluster | node
  rates:
    cpuCoreHour: 0.0316
    memoryGiBHour: 0.0042
  karpenter:
    refreshInterval: 10m
```

| Value | Default | Description |
|---|---|---|
| `cost.source` | `auto` | `static`: `cost.rates` as they are. `karpenter`: rates calibrated to Karpenter's price estimates; an error in the UI when they are missing. `auto`: `karpenter` when its metrics are in Prometheus, `static` otherwise. |
| `cost.allocation` | `cluster` | `cluster`: every workload at the cluster's average rates. `node`: each workload at the rates of the nodes it ran on. Needs Karpenter prices. |
| `cost.rates.cpuCoreHour` | `0.0316` | USD per CPU core-hour. With Karpenter, only its ratio to the memory rate matters, plus it is used until the first calibration. |
| `cost.rates.memoryGiBHour` | `0.0042` | USD per GiB-hour of memory. |
| `cost.karpenter.refreshInterval` | `10m` | How often the rates are recalibrated (Go duration, at least `30s`). Workloads pick up new rates at their next regeneration. |

How these are used: [Costs](../concepts/costs.md).

## Choosing the rates

Without Karpenter, the rates are all there is, so set them from your provider's prices. A simple way is to take the
on-demand price of your most common instance type and split it between CPU and memory. For a node with 4 vCPU and
16 GiB at $0.192/hour, with the default 7.5:1 CPU-to-GiB price ratio:

```text
0.192 = 4 × cpu + 16 × mem,   cpu = 7.5 × mem
mem = 0.192 / (4 × 7.5 + 16) ≈ 0.0042 per GiB-hour
cpu ≈ 0.0313 per core-hour
```

If most of your capacity is spot or discounted, scale both rates down by the average discount. The point is a fair
comparison between before and after, not a bill.

## With Karpenter

`auto` picks Karpenter's prices up by itself when Prometheus scrapes Karpenter. Set `karpenter` instead to be told,
with an error on the Overview and Settings pages, when the prices go missing. Then `node` allocation prices spot and
on-demand workloads differently. See [Karpenter costs](../guides/karpenter-costs.md).
