---
title: Introduction
description: What resourcerer is, what it does and does not do, and how it fits next to the tools you already run.
---

resourcerer recommends CPU and memory **requests and limits** for Kubernetes workloads from their usage history in
Prometheus, shows what each recommendation changes in a web UI, and applies it on request or automatically.

It is built around one decision: **should I apply this recommendation?** Every workload page answers it with three
things side by side:

- **What it saves**: the monthly cost of the current requests against the recommended ones, priced at configured rates
  or at what your nodes really cost.
- **What it risks**: a replay of the usage history against the recommendation (how often the busiest pod would have
  used more than it), and the reliability signals of the workload (throttling, OOM kills, restarts, your own PromQL).
- **Why the numbers can be trusted**: each value comes with its basis, such as `p95 + 10% over 7d`, and how much of the
  learning period the history covers.

## What it does

- Keeps one [`Recommendation`](../reference/custom-resources.md) per Deployment, StatefulSet, DaemonSet and Argo
  Rollout, regenerated on a schedule spread across workloads.
- Computes requests from the **busiest pod** of each container: the pod that would be throttled or OOM-killed first,
  not the average.
- Lets you choose an **optimization goal** per resource (`Savings`, `Balanced`, `Reliability`), a limits policy,
  minimums, and containers to skip, as cluster defaults or per namespace and workload with annotations.
- Responds to **OOM kills and CPU throttling**: memory goes above the limit that killed the container, a throttling
  CPU limit is raised.
- Keeps **HPAs** scaling at the same point: when an HPA scales on CPU utilization, its target moves with the requests.
- **Applies** with server-side apply, records a history, and **reverts** with one click.
- **Auto-deploys** `Ready` recommendations that changed enough, gated by **apply policies** (windows, freezes, rate
  limits) your teams write themselves.

## What it does not do

- **No machine learning.** Recommendations are percentiles plus margins. That makes them predictable and explainable,
  and it means seasonal patterns longer than the learning period are not anticipated: choose a learning period that
  covers your cycle.
- **No horizontal scaling.** resourcerer sizes pods; it does not change replica counts or HPA bounds. Workloads whose
  HPA scales on both CPU and memory utilization are left alone, since changing requests would change their replicas.
- **No node autoscaling.** It works next to Karpenter or the Cluster Autoscaler, and reads Karpenter's prices to cost
  workloads, but does not provision nodes.
- **No data leaves the cluster.** There is no SaaS backend; the UI talks to the server in your cluster, which talks to
  your Prometheus and the Kubernetes API.

## How it compares to the Vertical Pod Autoscaler

The VPA also recommends requests, and can apply them by evicting pods. resourcerer differs in what it is for:

| | Vertical Pod Autoscaler | resourcerer |
|---|---|---|
| Data | Its own in-memory histograms, from metrics-server | The history already in Prometheus, from day one |
| Method | Decaying histogram, target percentile | Percentile + margin per goal, shown per value |
| Review | `kubectl describe vpa` | A UI with cost, risk, verdict and patches |
| Apply | Evicts pods, or in-place resize | Patches the pod template (a normal rollout), on click or automatically, within apply policies |
| Limits | Keeps the ratio | Four limits policies per resource |
| HPA on CPU | Conflicts | Supported: the HPA target is adjusted |

Do not let the VPA and resourcerer manage the same workload: both would write its requests.

## Next steps

- [Quickstart](quickstart.md): install it and see your first recommendations.
- [How it works](../concepts/how-it-works.md): the components and the path from usage to an applied change.
