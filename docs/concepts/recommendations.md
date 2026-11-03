---
title: Recommendations
description: How requests and limits are computed from usage, per container and resource, and what the basis of each value means.
---

A recommendation gives each container of a workload a CPU and a memory **request**, and a **limit** where the limits
policy calls for one. Every value comes with its **basis**, a short explanation such as `p95 + 10% over 7d`, shown next
to it in the UI and stored in the `Recommendation`'s status.

There is no model: the rules below are the whole method.

## Usage of the busiest pod

For each container, resourcerer reads CPU (`rate()` of `container_cpu_usage_seconds_total`) and memory (working set)
over the **learning period**, and at each step keeps the **busiest pod**: the replica using the most.

The busiest pod, not the average, because it is the one that gets throttled or OOM-killed first. A workload whose
traffic is uneven across replicas would be undersized by an average.

## Requests: a percentile plus a margin

The request is a percentile of that usage plus a margin. The percentile and margin come from the **optimization goal**,
set per resource:

| Goal | CPU request | Memory request | Use it for |
|---|---|---|---|
| `Savings` | p85 + 5% | p99 + 10% | Non-production, batch, anything that tolerates throttling and the odd restart |
| `Balanced` (default) | p95 + 10% | max + 15% | Most production workloads |
| `Reliability` | p99 + 20% | max + 30% | Critical paths, memory-hungry runtimes, spiky workloads |

Memory is sized from the maximum under `Balanced` and `Reliability` because running out of memory kills the container,
while running out of CPU only slows it down. The goals are independent: `Savings` for CPU and `Reliability` for memory
is a common pair.

Then:

1. **Minimums** apply: `cpu.requests.min` (default `10m`) and `memory.requests.min` (default `16Mi`).
2. **Rounding up**: CPU to 1m below 100m and to 5m above, memory to 1Mi.

## After an OOM kill

When the kernel OOM-kills a container, its memory usage was capped by the limit at that moment, so the history
underestimates what it needed. If a container was OOM-killed during the learning period, its memory request is raised
to **at least 1.2× the memory limit that killed it**, and the basis says `raised after OOM kill`.

## Limits

What happens to limits is the **limits policy**, set per resource:

| Policy | Requests | Limits |
|---|---|---|
| `RequestsRaiseLimitsIfNeeded` (default) | Recommended | Kept as they are. Raised to request × ratio only if the new request would pass them. No limit is added where there was none. |
| `RequestsOnly` | Recommended | Never changed. |
| `RequestsAndLimits` | Recommended | Always set to request × ratio. |
| `DoNotOptimize` | Kept | Kept. resourcerer leaves the resource alone. |

The ratios are `cpu.limit-request-ratio` (default `2.0`) and `memory.limit-request-ratio` (default `1.5`).

### CPU throttling

CPU limits are enforced in 100ms periods, so a container can be throttled while its sampled usage stays well under the
limit. When a container was throttled in **at least 25% of its CFS periods** over the learning period, its CPU limit is
raised to **1.5× the current limit** (unless the policy is `RequestsOnly` or `DoNotOptimize`), and the basis says
`limit raised: throttled N% of the time`.

## Containers left alone

A container keeps its current values, and the basis says why, when:

- its name is in `containers.skip` (default `istio-proxy,linkerd-proxy`: service mesh sidecars are sized by the mesh);
- the policy for that resource is `DoNotOptimize`;
- there is not enough data for it yet;
- its workload has an HPA scaling on CPU and `hpa.preserve-scale-point` is `false` (CPU only).

## HPAs

Changing CPU requests changes what an HPA's utilization target means: 70% of a smaller request is less CPU. When an HPA
scales the workload on **CPU utilization**, resourcerer also recommends a new **target utilization** that keeps the
HPA scaling at the same absolute CPU per pod, clamped to 50–90%. The applier patches both together.

When an HPA scales on **both CPU and memory utilization**, any change to requests would change replica counts in ways
that cannot be compensated, so the workload is `Unsupported` and keeps its values. HPAs on other metrics (requests per
second, queue length) do not interfere.

## When a recommendation counts as a change

A recommendation is **pending** (worth applying) when at least one request moves by more than
`auto-deploy.threshold` (default `10%`) **and** by at least 10m of CPU or 16Mi of memory. Smaller drifts are ignored, so
a workload does not roll out for a few millicores. The workload list's **To apply** view and auto-deploy both use it.

## The hash

Each set of recommended values has a **hash**. An apply carries the hash of the recommendation you reviewed; if the
recommendation changed in between, the applier refuses it and asks you to review again. Auto-deploy applies each hash
at most once, and never re-applies a hash that was reverted.

## Score

The **score** (0–100) says how closely the current requests match the recommendation: for each resource, the sum over
containers of the smaller of current and recommended over the larger, weighted by what CPU and memory cost. 100 means
already right-sized; a low score means a large change, in either direction.

## Settings involved

| Setting | Default |
|---|---|
| `cpu.optimization-goal`, `memory.optimization-goal` | `Balanced` |
| `learning-period` | `P7D` |
| `cpu.limits-policy`, `memory.limits-policy` | `RequestsRaiseLimitsIfNeeded` |
| `cpu.limit-request-ratio`, `memory.limit-request-ratio` | `2.0`, `1.5` |
| `cpu.requests.min`, `memory.requests.min` | `10m`, `16Mi` |
| `containers.skip` | `istio-proxy,linkerd-proxy` |
| `hpa.preserve-scale-point` | `true` |

How to set them: [Configuration](../configuration/overview.md). All of them: [Settings reference](../reference/settings.md).
