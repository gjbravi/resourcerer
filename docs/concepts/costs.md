---
title: Costs
description: How monthly costs and savings are estimated, from configured rates or Karpenter's node prices, per cluster or per node.
---

Every recommendation is priced twice, at the current requests and at the recommended ones; the difference is the
savings (or the extra cost) shown on the workload page, the workload list and the overview.

## The formula

```text
monthly cost = (CPU requests × CPU rate + memory requests × memory rate) × average replicas × 730 hours
```

- **Requests**, not usage: a request is capacity reserved on a node whether it is used or not, which is what you pay
  for. (The "Used" figure on the workload page prices p95 usage the same way, to show what the workload needs.)
- **Average replicas** over the learning period, so a workload scaled by an HPA is priced at the replicas it actually
  ran.
- **Rates** per CPU core-hour and per GiB-hour of memory, in USD.

Costs are estimates for comparing before and after, not a bill: they leave out the share of node capacity no pod
requests, DaemonSets' overhead on each node, storage and network.

## Where the rates come from

`cost.source` decides:

| Source | Rates |
|---|---|
| `static` | `cost.rates` as configured (default `0.0316` per core-hour, `0.0042` per GiB-hour). |
| `karpenter` | `cost.rates` scaled so the Karpenter nodes cost what Karpenter's price estimates say, spot or on-demand. An error in the UI when the prices are missing. |
| `auto` (default) | `karpenter` when its metrics are in Prometheus, `static` otherwise. |

With Karpenter, the **calibration** keeps the configured CPU to memory ratio (a node's price does not say how much of it
is CPU) and scales both rates by one **cluster factor**: what capacity costs on this cluster relative to `cost.rates`.
It is refreshed every `cost.karpenter.refreshInterval` (default `10m`); a failed refresh keeps the last good
calibration. The **Settings** page shows the configured and calibrated rates.

## Per cluster or per node

`cost.allocation` decides how the rates are split between workloads:

| Allocation | Each workload is priced at |
|---|---|
| `cluster` (default) | The cluster's average rates. Two workloads with the same requests cost the same. |
| `node` | The rates of the nodes its pods ran on, weighted by time and requests over the learning period. A workload on on-demand nodes costs more than the same requests on spot. |

`node` needs Karpenter prices. For each node, its **node factor** is its price divided by its capacity at the configured
rates; reserved capacity is priced as on-demand. A workload's **workload factor** is the request-weighted average of the
factors of the nodes it ran on. Requests on nodes without a known price count at the cluster factor; the share of
requests that ran on priced nodes is the **coverage**.

The same factor prices both the current and the recommended requests, so savings stay comparable: resourcerer does not
assume a smaller pod moves to cheaper nodes.

## Placement

With `node` allocation, the **How the cost is calculated** section of the Impact tab shows where the workload ran over
its learning period: the share of its requests on spot, on-demand and reserved capacity, the instance types that
carried most of them, and what each cost. It explains why two similar workloads have different costs.

## What to read next

- [Karpenter costs](../guides/karpenter-costs.md): setting up `karpenter` and `node` allocation.
- [Cost settings](../configuration/cost-estimates.md): every value under `cost`.
