---
title: Goals and limits
description: Optimization goals, limits policies and ratios, minimum requests, skipped containers and HPA handling.
---

These settings decide the recommended values. They resolve per key from the chart's `defaults`, then namespace
annotations, then workload annotations (`resourcerer.io/<key>`). How each is used in the computation:
[Recommendations](../concepts/recommendations.md).

## Optimization goal

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/cpu.optimization-goal` | `Balanced` | `Savings`, `Balanced`, `Reliability` |
| `resourcerer.io/memory.optimization-goal` | `Balanced` | `Savings`, `Balanced`, `Reliability` |

| Goal | CPU request | Memory request |
|---|---|---|
| `Savings` | p85 + 5% | p99 + 10% |
| `Balanced` | p95 + 10% | max + 15% |
| `Reliability` | p99 + 20% | max + 30% |

Choosing:

- **CPU**: going above the request borrows idle CPU from the node; under contention, the container gets its share by
  request. `Savings` suits workloads that tolerate latency under load; `Reliability` latency-sensitive ones whose
  traffic spikes.
- **Memory**: going above the request makes the pod a candidate for eviction under node memory pressure, and above the
  limit kills it. `Savings` sizes from p99, so it accepts that the top 1% of samples pass the request; keep it for
  workloads that restart cheaply.

```bash
kubectl annotate namespace dev resourcerer.io/cpu.optimization-goal=Savings resourcerer.io/memory.optimization-goal=Savings
kubectl -n search annotate deployment indexer resourcerer.io/memory.optimization-goal=Reliability
```

## Limits policy

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/cpu.limits-policy` | `RequestsRaiseLimitsIfNeeded` | `RequestsRaiseLimitsIfNeeded`, `RequestsOnly`, `RequestsAndLimits`, `DoNotOptimize` |
| `resourcerer.io/memory.limits-policy` | `RequestsRaiseLimitsIfNeeded` | same |
| `resourcerer.io/cpu.limit-request-ratio` | `2.0` | a number ≥ 1 |
| `resourcerer.io/memory.limit-request-ratio` | `1.5` | a number ≥ 1 |

| Policy | Limits after an apply |
|---|---|
| `RequestsRaiseLimitsIfNeeded` | Unchanged, unless the new request would be above the limit: then request × ratio. A container without a limit gets none. |
| `RequestsOnly` | Unchanged, always. |
| `RequestsAndLimits` | Request × ratio, always, adding limits where there were none. |
| `DoNotOptimize` | Requests and limits unchanged: resourcerer leaves this resource of the workload alone. |

Two exceptions raise a limit regardless of the request:

- a CPU limit throttling the container in 25% or more of its CFS periods is raised to 1.5× (except under `RequestsOnly`
  and `DoNotOptimize`);
- after an OOM kill, the memory request goes to at least 1.2× the limit that killed the container, so under
  `RequestsRaiseLimitsIfNeeded` the memory limit follows at request × ratio.

Common practices:

| Practice | Settings |
|---|---|
| No CPU limits, memory limit equal to the request | `cpu.limits-policy: RequestsOnly` (and no CPU limits in the manifests), `memory.limits-policy: RequestsAndLimits`, `memory.limit-request-ratio: 1.0` |
| Keep whatever the teams set, only fix requests | `RequestsOnly` for both |
| Limits at a fixed ratio everywhere | `RequestsAndLimits` for both, with your ratios |

## Minimum requests

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/cpu.requests.min` | `10m` | a Kubernetes quantity |
| `resourcerer.io/memory.requests.min` | `16Mi` | a Kubernetes quantity |

A recommendation is never below the minimum. Raise the memory minimum for runtimes with a large fixed baseline, and the
CPU minimum where a container needs CPU to start in time for its probes.

## Containers to skip

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/containers.skip` | `istio-proxy,linkerd-proxy` | comma-separated container names, or `none` |

Skipped containers keep their values, and are left out of the reliability signals: applying never changes them, so
a throttled sidecar does not weigh on the verdict. Setting the annotation replaces the list, so include the defaults if you still
want them skipped:

```bash
kubectl -n payments annotate deployment checkout resourcerer.io/containers.skip=istio-proxy,linkerd-proxy,log-shipper
```

## HPA scale point

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/hpa.preserve-scale-point` | `true` | `true`, `false` |

With `true`, a workload whose HPA scales on CPU utilization gets a new HPA target that keeps it scaling at the same
CPU per pod. With `false`, its CPU is left alone. See [Workloads](../concepts/workloads.md#hpas).
