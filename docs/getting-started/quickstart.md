---
title: Quickstart
description: Install resourcerer with Helm, open the UI and review your first recommendations.
---

This quickstart installs resourcerer with one Helm command. Preliminary recommendations show up within minutes; complete
ones once the learning period (7 days by default) is covered by history, which is usually right away if your
Prometheus already keeps a week of data.

Nothing is changed in your cluster until you click **Apply** or turn on auto-deploy.

## What you need

- A Kubernetes cluster, 1.33 or later, and `kubectl` and `helm` pointed at it.
- A Prometheus-compatible API with **cAdvisor** and **kube-state-metrics** data, reachable from the cluster: Prometheus,
  Thanos, Mimir or VictoriaMetrics. See [Requirements](requirements.md) for the details.

## 1. Install

Pick a version from the [releases page](https://github.com/gjbravi/resourcerer/releases) and point
`prometheus.url` at your Prometheus API:

```bash
helm install resourcerer oci://ghcr.io/gjbravi/charts/resourcerer --version <version> \
  --namespace resourcerer --create-namespace \
  --set prometheus.url=http://prometheus-server.monitoring.svc \
  --set clusterName=staging
```

## 2. Check that it runs

```bash
kubectl -n resourcerer get pods
```

```text
NAME                                      READY   STATUS    RESTARTS   AGE
resourcerer-applier-6d8f9c7b5d-x2kqp      1/1     Running   0          1m
resourcerer-controller-7c9b8d6f4c-m7wzt   1/1     Running   0          1m
resourcerer-server-5f7d9c8b6a-q4hvn       1/1     Running   0          1m
```

The controller creates one `Recommendation` per workload as it reaches them:

```bash
kubectl get recommendations -A
```

```text
NAMESPACE   NAME                  KIND         TARGET     PHASE      SCORE   COST/MO   RECOMMENDED/MO   AGE
demo        deployment-api        Deployment   api        Ready      62      41.20     18.75            2m
demo        statefulset-cache     StatefulSet  cache      Learning   88      12.10     10.40            2m
payments    deployment-checkout   Deployment   checkout   Ready      45      96.30     51.90            1m
```

## 3. Open the UI

The chart installs with `auth.strategy: anonymous`, so there is no login and the port-forward is what decides who gets
in:

```bash
kubectl -n resourcerer port-forward svc/resourcerer 8080:80
```

To share the UI with a team, see [Exposing the UI](../guides/exposing-the-ui.md): OIDC sign-in, or an authenticating
proxy.

Open [http://localhost:8080](http://localhost:8080). The **Settings** page shows whether Prometheus answers and how
much history it keeps; the **Overview** shows the cluster's cost and savings, and **Workloads** lists every
recommendation.

## What happens next

- **Within minutes**: workloads move from `Pending` to `Learning` or `Ready`. A `Learning` recommendation is
  preliminary: its history does not cover the learning period yet.
- **Over the learning period**: recommendations are regenerated on their schedule (`@daily` by default, at a time
  spread per workload) as history accumulates.
- **When `Ready`**: the history covers the learning period. This is what auto-deploy waits for.

See [Phases](../concepts/phases.md) for every state a recommendation can be in.

## 4. Apply your first recommendation

1. On **Workloads**, pick a workload with savings in a non-critical namespace.
2. Open it. The **Impact** tab shows the cost before and after and the verdict under the **Apply** button.
3. Check the **Recommendation** tab: each value with its basis, and the usage charts with the current and recommended
   requests.
4. Click **Apply**. The applier patches the pod template and the workload rolls out as on any deploy. **Revert** puts
   the previous values back.

Prefer Git? The **Patches** tab has the same change as a Kustomize patch, Helm values or a `kubectl` command.

## Next steps

- [Configure resourcerer](../configuration/overview.md): the five settings worth deciding first.
- [Apply and revert](../guides/apply-and-revert.md): what an apply does, step by step.
- [Roll out auto-deploy](../guides/auto-deploy.md): when you trust the recommendations.
- Try it without a cluster of your own: [resourcerer on kind](../guides/try-on-kind.md).
