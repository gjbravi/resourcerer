---
title: Troubleshooting
description: Common problems, from workloads with no recommendation to applies that do not happen, and how to diagnose them.
---

Start with the `Recommendation`: its phase, message, warnings and last apply result answer most questions.

```bash
kubectl -n <namespace> get recommendation <kind>-<name> -o yaml
```

## A workload has no recommendation

- **Excluded.** Its namespace is in `excludedNamespaces`, or it is the release namespace.
- **Not selected.** With `selection: all`, look for `resourcerer.io/enabled=false` on the workload or its namespace.
  With `selection: label`, one of them needs `resourcerer.io/enabled=true`.
- **Unsupported kind.** Jobs, CronJobs and bare pods are not managed. A Deployment referenced by an Argo Rollout's
  `workloadRef` has none of its own: the Rollout carries it.
- **Argo Rollouts** are picked up only if the `rollouts.argoproj.io` CRD existed when the controller started; restart it
  after installing Argo Rollouts.

## A workload stays Pending

The controller has fewer than 10 samples per container. Check that Prometheus has usage for the workload's pods:

```promql
count by (pod) (container_memory_working_set_bytes{namespace="payments", pod=~"checkout-.*", container!=""})
count by (pod) (kube_pod_owner{namespace="payments", pod=~"checkout-.*"})
```

- No cAdvisor series: the kubelet's `/metrics/cadvisor` endpoint is not scraped, or the `container` label is dropped by
  relabelling.
- No `kube_pod_owner`: kube-state-metrics is missing or not scraped. resourcerer joins pods to their workload through
  it, and through `kube_replicaset_owner` for Deployments.

## A workload stays Learning

The history does not cover 95% of the learning period. Either it is new (wait), or Prometheus keeps less than the
learning period: compare its retention on the UI's **Settings** page with `learning-period`. Query through Thanos or
Mimir if they keep the longer history, or shorten the learning period.

## Metrics unavailable (Error)

The message carries the Prometheus error. Common causes: a wrong `prometheus.url` (it is the base URL, without
`/api/v1`), a NetworkPolicy blocking the `resourcerer` namespace, a timeout on a very long learning period, or an API
that needs authentication.

## The recommendation looks wrong

- **Too low for a known peak**: the peak is older than the learning period, or it is in the top percent of CPU samples
  the goal's percentile leaves out. Use a longer learning period or the `Reliability` goal.
- **Too high**: look at the basis. `raised after OOM kill` means memory follows the OOM rule; `limit raised: throttled`
  means the CPU limit follows the throttling rule. The usage charts on the **Recommendation** tab show which samples
  drove the percentile.
- **A sidecar is changed**: add it to `containers.skip`, keeping the default proxies in the list.
- **Settings ignored**: an invalid annotation value is skipped; the workload page and `status.warnings` say which.

## Apply does nothing

Check `status.lastApplyResult` and the events on the workload:

| What you see | Cause |
|---|---|
| No result at all | The applier is not running (`applier.enabled: false`, or its pod is failing). |
| `Rejected: the recommendation changed since it was reviewed` | Reload the page and review the new values. |
| `Rejected: cannot apply a … recommendation` | Only `Ready` and `Learning` can be applied. |
| `Rejected` with a policy's reason | An apply policy holds manual applies now; the reason says until when. |
| `Failed` with an API error | An admission webhook, a LimitRange or a ResourceQuota refused the new values. |
| `Succeeded`, but the values come back | A GitOps tool or `helm upgrade` re-applied the old values. See [GitOps](../guides/gitops.md). |

## Auto-deploy does not apply

In order:

1. `auto-deploy` is `Enabled` in the **Settings** tab, with the source you expect.
2. The phase is `Ready`, not `Learning`.
3. The recommendation is pending: it is listed in the **To apply** view, or `status.pending` is `true`.
4. Its hash was not already applied or reverted (`status.history`).
5. No rollout is in progress on the workload.
6. `status.applyPolicy.auto` is `Allowed`. If `Blocked`, `blockedBy` names the rule and `nextAllowed` when it lifts;
   an `ApplyDeferred` event says the same.

## Apply policies

- **Everything is held with `blockedBy: policy`**: a policy cannot be interpreted (an unknown time zone, a bad selector).
  `kubectl get clusterapplypolicies,applypolicies -A -o wide` and the `Valid` condition name it.
- **A window never opens**: check `timeZone` and that `to` is after `from` unless you meant to cross midnight. Run the
  policy in `Audit` mode and watch `resourcerer_apply_deferred_total`.
- **Applies run with no policy at all after an upgrade**: the `resourcerer-crds` release was not installed or upgraded,
  so the policy CRDs are missing. See [Upgrade](../getting-started/upgrade-and-uninstall.md#upgrade).

## Costs look off

- **All rates are the defaults**: `cost.source` is `auto` and Karpenter's metrics are not in Prometheus. Set
  `karpenter` to get an error instead, then fix the scrape.
- **Node allocation falls back to cluster rates**: `status.pricing.reason` says why; low coverage means many of the
  workload's requests ran on nodes without a price. `price the Karpenter nodes: … incomplete result` means Prometheus
  answered only in part, usually a Thanos store that cannot serve the window: the reason names the store, and
  resourcerer prices at the cluster rates rather than with part of the nodes.

## Getting help

Open an issue on [GitHub](https://github.com/gjbravi/resourcerer/issues) with the resourcerer version, the
`Recommendation`'s status, and the components' logs. Remove anything specific to your environment first.
