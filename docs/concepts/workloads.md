---
title: Workloads
description: Which workloads get a recommendation, how Argo Rollouts and HPAs are handled, and how to opt a workload in or out.
---

## Supported kinds

| Kind | Notes |
|---|---|
| `Deployment` | |
| `StatefulSet` | Applying rolls the pods out one by one, following its update strategy. |
| `DaemonSet` | One recommendation for every node's pod; the busiest pod is usually on the busiest node. |
| `Rollout` (Argo Rollouts) | Picked up automatically when the `rollouts.argoproj.io` CRD is installed. See [Argo Rollouts](../guides/argo-rollouts.md). |

Jobs, CronJobs and bare pods are not managed.

## Which workloads are managed

`selection` in the chart values decides:

- **`all`** (default): every workload of a supported kind, outside the excluded namespaces. Opt one out with the label
  `resourcerer.io/enabled=false` on the workload or its namespace.
- **`label`**: only workloads labeled `resourcerer.io/enabled=true`, on the workload or its namespace. Use it to roll
  resourcerer out namespace by namespace.

```bash
# selection: all — leave a namespace out
kubectl label namespace batch resourcerer.io/enabled=false

# selection: label — bring a namespace in, except one workload
kubectl label namespace payments resourcerer.io/enabled=true
kubectl -n payments label deployment legacy-ledger resourcerer.io/enabled=false
```

`excludedNamespaces` (default `kube-system`, `kube-public`, `kube-node-lease`) are never managed, whatever their labels,
and neither is the namespace resourcerer is installed in.

Taking a workload out deletes its `Recommendation`. Values already applied stay on the workload.

## One recommendation per workload

The `Recommendation` is named `<kind>-<name>`, in lowercase, in the workload's namespace, and owned by the workload, so
it is deleted with it:

```bash
kubectl -n payments get recommendation deployment-checkout -o yaml
```

## Containers

Every container of the pod template gets values, except those listed in `containers.skip` (by default the Istio and
Linkerd proxies, which their mesh sizes). Init containers are not changed.

## HPAs

An HPA targeting the workload is read on every regeneration:

- **Scaling on CPU utilization**: the recommendation includes a new target utilization that keeps the HPA scaling at the
  same absolute CPU per pod (`hpa.preserve-scale-point: true`, the default), clamped to 50–90%. With `false`, CPU is
  left alone and only memory is recommended.
- **Scaling on memory utilization only**: requests are recommended and the HPA keeps its target. Memory utilization is
  relative to the request, so a lower memory request makes the HPA scale out sooner; review the replica count after
  applying, or set `memory.limits-policy: DoNotOptimize` on the workload to leave memory alone.
- **Scaling on both CPU and memory utilization**: the workload is `Unsupported`.
- **Scaling on other metrics** (custom, external): no interaction.
- **At its maximum replicas**: the `hpaAtMax` reliability signal flags it, since the HPA cannot absorb more load.

Costs use the average replica count over the learning period, so the savings of an HPA-scaled workload reflect the
replicas it actually ran.
