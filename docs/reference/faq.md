---
title: FAQ
description: Short answers to the questions people ask before installing and before applying.
---

### Does installing resourcerer change anything in my workloads?

No. It only reads until someone clicks **Apply** or enables auto-deploy. For a guarantee, install with
`applier.enabled: false` and `readOnly: true`.

### Does applying restart my pods?

Yes. Requests and limits are part of the pod template, so applying starts a rollout, exactly like a deploy, following
the workload's strategy (`maxSurge`, `maxUnavailable`, a PodDisruptionBudget, an Argo Rollouts canary). Apply policies
decide when it may happen.

### Why the busiest pod and not the average?

Because resources are set per pod and every replica gets the same values. The replica with the most traffic is the
one that would be throttled or OOM-killed with a request sized for the average.

### Why no machine learning?

So every number can be explained and checked: `p95 + 10% over 7d` can be verified against the chart next to it, and
when a value is off, the reason is visible and the fix is a setting. The trade-off is that patterns longer than the
learning period are not anticipated; choose a period that covers your traffic cycle.

### Can I use it with the Vertical Pod Autoscaler?

Not on the same workload: both write requests. Take the workloads the VPA manages out with `resourcerer.io/enabled=false`,
or remove their VPA. See [Introduction](../getting-started/introduction.md#how-it-compares-to-the-vertical-pod-autoscaler).

### Does it work with HPAs?

Yes, with HPAs on CPU (the target is adjusted), on memory, or on custom metrics. Not with HPAs scaling on both CPU and
memory utilization. See [Workloads](../concepts/workloads.md#hpas).

### Does any data leave the cluster?

No. resourcerer has no backend outside your cluster and makes no outbound calls besides Prometheus.

### What Prometheus retention do I need?

At least your longest learning period, 7 days by default. See [Requirements](../getting-started/requirements.md).

### How accurate are the costs?

They are estimates to compare before and after: requests × average replicas × hourly rates × 730. With Karpenter, the
rates follow what your nodes cost. They leave out unrequested capacity, storage and network. See [Costs](../concepts/costs.md).

### Can a team configure its own namespace?

Yes: settings are namespace and workload annotations, and `ApplyPolicy` objects live in the team's namespace. A team's
policy can only be stricter than the platform's `ClusterApplyPolicy`.

### What happens to applied values if I uninstall?

They stay. See [Upgrade and uninstall](../getting-started/upgrade-and-uninstall.md#uninstall).

### Is it production-ready?

It is in beta: it runs on real clusters, and APIs, annotations and the CRDs may still change between minor versions,
always noted in the changelog. Start in recommend-only mode, then apply by hand, then auto-deploy with apply policies.
