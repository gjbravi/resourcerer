---
title: Custom resources
description: The fields of Recommendation, ApplyPolicy and ClusterApplyPolicy (resourcerer.io/v1alpha1).
---

All three are in the API group `resourcerer.io`, version `v1alpha1`. The Go types, with every field documented, are in
[`api/v1alpha1`](https://github.com/gjbravi/resourcerer/blob/main/api/v1alpha1/recommendation_types.go).

## Recommendation

Namespaced, short names `rec` and `recs`. One per managed workload, named `<kind>-<name>` in lowercase, owned by the
workload. The controller creates and deletes them; you do not write them by hand.

```bash
kubectl get recommendations -A
```

```text
NAMESPACE   NAME                  KIND         TARGET     PHASE   SCORE   COST/MO   RECOMMENDED/MO   AGE
payments    deployment-checkout   Deployment   checkout   Ready   45      96.30     51.90            9d
```

### spec

Written by the controller (`targetRef`) and the server (requests from the UI).

| Field | Description |
|---|---|
| `targetRef` | `apiVersion`, `kind`, `name` of the workload. |
| `applyRequest` | A request for the applier: `id` (unique), `action` (`Apply` or `Revert`), `hash` of the reviewed recommendation, `requestedBy`, `requestedAt`. Handled once per `id`. |
| `regenerateRequestedAt` | Regenerate now instead of at the next slot, when newer than `status.generatedAt`. |

### status

Written by the controller, except `history`, `lastApplyResult`, `lastHandledRequest*` and `applyPolicy`, written by the
applier.

| Field | Description |
|---|---|
| `phase`, `message` | `Pending`, `Learning`, `Ready`, `Warning`, `Unsupported` or `Error`, and why. See [Phases](../concepts/phases.md). |
| `templateRef` | For a Rollout with `workloadRef`: the Deployment whose template an apply patches. |
| `generatedAt`, `nextScheduleAt` | When the recommendation was last generated, and its next slot. |
| `hash` | Identity of the recommended values. |
| `pending` | `true` when the recommendation differs enough from the current requests to apply. |
| `settings` | Every effective setting: `value` and `source` (`cluster`, `namespace`, `workload`). |
| `warnings` | Invalid settings that were ignored, and similar. |
| `containers[]` | Per container: `name`, `current` and `recommended` resources, `basis` per resource, `skipped`. |
| `hpa` | `name`, `current` and `recommended` CPU target utilization, `clamped` when the 50–90% range cut it. |
| `score` | `overall`, `cpu`, `memory`, 0–100. |
| `impact` | `avgReplicas`; `current` and `recommended` totals: `cpu`, `memory`, `monthlyCost`. |
| `pricing` | `allocation`, `factor`, `clusterFactor`, `coveragePercent`, `capacity` shares, top `instanceTypes`, and `reason` when node allocation fell back. |
| `evidence` | `historySeconds`, `learningSeconds`, `cpuAbovePercent`, `memoryAbovePercent`, `cpuPeakPercent`, `memoryPeakPercent`, `used` totals. See [Verdict and evidence](../concepts/verdict-and-evidence.md). |
| `oomKills` | OOM kills seen in the learning period. |
| `reliability` | Built-in signals per container (`cpuThrottledPercent`, `memoryPeakPercent`), `restarts`, `crashLooping`, `bestEffortPods`, `hpaAtMax`, and declared `signals[]` with `value`, `display`, `status` (`ok`, `warn`, `critical`). |
| `targetLabels` | The workload's and its pod template's labels, for apply policy selectors. |
| `history[]` | The last 10 applies and reverts: `time`, `action`, `by`, `hash`, and per container the resources `from` and `to`. |
| `lastHandledRequestID`, `lastHandledRequestAt` | The last `applyRequest` the applier handled. |
| `lastApplyResult` | `time`, `result` (`Succeeded`, `Failed`, `Rejected`), `message`. |
| `applyPolicy` | The apply policy decision for `auto` and `manual` applies: `state`, `policy`, `blockedBy`, `reason`, `nextAllowed`, `wouldBlock`; and `evaluatedAt`. See [Apply policies](../guides/apply-policies.md#the-decision). |

## ApplyPolicy and ClusterApplyPolicy

`ApplyPolicy` is namespaced (short name `ap`) and governs the workloads of its namespace. `ClusterApplyPolicy` is
cluster-scoped (short name `cap`) and governs the namespaces it selects. Every policy that selects a workload must allow
an apply. How they are evaluated: [Apply policies](../guides/apply-policies.md).

```bash
kubectl get clusterapplypolicies
kubectl -n payments get applypolicies
```

### spec

| Field | Type | Description |
|---|---|---|
| `namespaceSelector` | label selector | `ClusterApplyPolicy` only. The namespaces it governs; empty means every managed namespace. |
| `workloadSelector` | label selector | The workloads it governs, matched against the workload's labels and its pod template's. Empty means all in scope. |
| `appliesTo` | `[Auto]`, `[Manual]`, `[Auto, Manual]` | The kinds of apply it governs. Empty means `Auto`. |
| `mode` | `Enforce` (default), `Audit` | `Audit` records what it would block in `wouldBlock` and lets the apply through. |
| `windows[]` | up to 32 | When an apply may start; a union. Empty means any time. |
| `windows[].days` | `Mon` … `Sun` | Empty means every day. |
| `windows[].from`, `windows[].to` | `HH:MM` | Local times; `to` is exclusive. `to` before `from` crosses midnight. |
| `windows[].timeZone` | IANA name | Empty means UTC. |
| `freezes[]` | up to 64 | When nothing is applied, windows or not. |
| `freezes[].from`, `freezes[].to` | RFC 3339 time | `to` is exclusive and must be after `from`. |
| `freezes[].reason` | string, up to 256 | Shown in the decision, the UI and the events. |
| `rate.maxConcurrent` | integer or percentage | Applies in progress at once among the selected workloads; a percentage of those with auto-deploy enabled, rounded up. `0` blocks automatic applies. |
| `rate.maxPerHour` | integer | Applies in any hour among the selected workloads. `0` blocks automatic applies. |
| `rate.perNamespace.maxConcurrent` | integer or percentage | Rollouts in progress in the workload's namespace, including those resourcerer did not start. |
| `rate.cooldown` | Go duration | Minimum time between two changes (apply or revert) to the same workload. Holds manual applies too. |
| `rate.spread` | Go duration | Delays each held apply to a stable offset in `[0, spread)` after the window opens. |

### status

| Field | Description |
|---|---|
| `workloads` | How many workloads the policy selects. |
| `conditions` | `Valid`: `True` when the policy is applied to the workloads it selects; `False` with reason `InvalidPolicy` and the error when it cannot be interpreted (it then holds the applies it governs). |
