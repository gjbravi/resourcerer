---
updatedAt: 2026-10-07T08:55:53.000Z
---

# Scheduled rebalancing for Karpenter clusters

Atlas provides a Kubernetes-native approach to scheduled rebalancing for clusters managed by Karpenter. Instead of configuring schedules through the Atlas console or API, you create `RebalancePlanSchedule` custom resources directly in your cluster. Keeper watches these resources and automatically triggers rebalancing operations on a cron-based schedule.

> 📘 Note
>
> This feature applies to clusters using Karpenter for node provisioning. For standard Atlas clusters, see [Scheduled rebalancing](scheduled-rebalancing.md).

# How it works

Kubernetes-native scheduled rebalancing uses two custom resource definitions (CRDs):

* **`RebalancePlanSchedule`** — defines when rebalancing should run, including the cron schedule and rebalancing configuration. This is a cluster-scoped resource (shortname: `rps`).
* **`RebalancePlanClaim`** — represents a single rebalancing execution request. The schedule controller creates these automatically at each scheduled time. This is a cluster-scoped resource (shortname: `rpc`).

When a schedule fires, Keeper creates a `RebalancePlanClaim`, submits it to Atlas for plan generation, and executes the resulting plan. Only one claim runs at a time — if a claim from a previous schedule is still active, the new execution is skipped.

## Lifecycle states

A `RebalancePlanClaim` progresses through the following states:

| State        | Description                                               |
| ------------ | --------------------------------------------------------- |
| `Pending`    | Claim created, not yet submitted to Atlas               |
| `Generating` | Atlas is computing the rebalancing plan                 |
| `Ready`      | Plan generated, waiting for execution to begin            |
| `Executing`  | Rebalancing is actively running (nodes being replaced)    |
| `Completed`  | Rebalancing completed successfully                        |
| `Failed`     | Rebalancing failed; see `status.errorMessage` for details |

# Create a rebalancing schedule

## Basic example

The following schedule runs every 30 minutes and rebalances all nodes in the cluster:

```yaml
apiVersion: autoscaling.atlas.example/v1alpha
kind: RebalancePlanSchedule
metadata:
  name: rebalance-every-30min
spec:
  schedule: "*/30 * * * *"
  rebalancePlanClaimTemplate:
    metadata: {}
    spec:
      autoExecute: true
```

Apply the schedule to your cluster:

```shell
kubectl apply -f rebalance-schedule.yaml
```

## Example: Rebalance spot nodes with savings threshold

This schedule runs nightly and only executes if projected savings are at least 15%:

```yaml
apiVersion: autoscaling.atlas.example/v1alpha
kind: RebalancePlanSchedule
metadata:
  name: nightly-spot-rebalance
spec:
  schedule: "0 2 * * *"
  rebalancePlanClaimTemplate:
    metadata: {}
    spec:
      autoExecute: true
      minSavingsPercentage: 15
      nodeConstraints:
        minAgeSeconds: 300
        maxNodes: 10
      scope:
        nodeSelector:
          nodeSelectorTerms:
            - matchExpressions:
                - key: karpenter.sh/capacity-type
                  operator: In
                  values:
                    - spot
```

## Example: Weekend full-cluster rebalance

This schedule runs every Saturday at midnight and rebalances the entire cluster, keeping at least 3 nodes:

```yaml
apiVersion: autoscaling.atlas.example/v1alpha
kind: RebalancePlanSchedule
metadata:
  name: weekend-rebalance
spec:
  schedule: "0 0 * * 6"
  rebalancePlanClaimTemplate:
    metadata: {}
    spec:
      autoExecute: true
      nodeConstraints:
        minClusterSize: 3
```

# Configuration reference

## `RebalancePlanSchedule` spec

| Field                        | Type      | Required | Description                                                                                                                                               |
| ---------------------------- | --------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `schedule`                   | `string`  | Yes      | Cron expression (UTC timezone). Standard 5-field format: `minute hour day month weekday`.                                                                 |
| `suspend`                    | `boolean` | No       | When `true`, pauses the schedule without deleting it. Defaults to `false`.                                                                                |
| `startingDeadlineSeconds`    | `integer` | No       | Maximum seconds after a scheduled time to still start a missed execution. If exceeded, the execution is skipped and recorded as missed.                   |
| `successfulJobsHistoryLimit` | `integer` | No       | Number of successful `RebalancePlanClaim` objects to retain. Defaults to `3`.                                                                             |
| `failedJobsHistoryLimit`     | `integer` | No       | Number of failed `RebalancePlanClaim` objects to retain. Defaults to `3`.                                                                                 |
| `preserveHistory`            | `boolean` | No       | When `false` (default), all claims created by this schedule are deleted when the schedule is deleted. Set to `true` to keep claim history after deletion. |
| `rebalancePlanClaimTemplate` | `object`  | Yes      | Template for the `RebalancePlanClaim` created at each schedule trigger.                                                                                   |

## `rebalancePlanClaimTemplate.spec` fields

### Scope

Use `scope` to limit rebalancing to specific nodes. If omitted, the entire cluster is considered.

**Target specific nodes by name:**

```yaml
spec:
  scope:
    nodeNames:
      - node-abc
      - node-def
```

**Target nodes by label selector:**

```yaml
spec:
  scope:
    nodeSelector:
      nodeSelectorTerms:
        - matchExpressions:
            - key: karpenter.sh/capacity-type
              operator: In
              values:
                - spot
```

`nodeNames` and `nodeSelector` are mutually exclusive — use one or the other, not both.

The `nodeSelector` follows standard Kubernetes node selector semantics:

* Multiple `nodeSelectorTerms` are evaluated with **OR** logic
* Multiple `matchExpressions` within a term are evaluated with **AND** logic

<Callout icon="📘" theme="info">
  ### Note

  Excluded NodePools override any scope you define. In `opt-out` mode, Keeper never rebalances nodes from pools annotated `exclude`. In `opt-in` mode, it rebalances only pools annotated `include`. See [Node exclusions](karpenter-enterprise-continuous-rebalancing.md#node-exclusions).
</Callout>

### Node constraints

`nodeConstraints` limits which nodes are eligible and how many are rebalanced at once:

| Field            | Type      | Description                                                                                                                                                     |
| ---------------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `minAgeSeconds`  | `integer` | Minimum node age in seconds before considering for rebalancing. Prevents rebalancing freshly created nodes. Example: `300` = skip nodes younger than 5 minutes. |
| `maxNodes`       | `integer` | Maximum number of nodes to rebalance in a single operation. Limits blast radius. Minimum: `1`.                                                                  |
| `minClusterSize` | `integer` | Minimum number of nodes to keep in the cluster. Safety guard to prevent the cluster from becoming too small.                                                    |

### Savings threshold

`minSavingsPercentage` sets the minimum projected cost savings percentage required before executing a plan. Valid range: `0`–`100`.

* `0` — always rebalance regardless of savings (useful for spot rotation or rolling nodes)
* `>0` — only execute if projected savings meet or exceed this percentage

### Execution policy

`executionPolicy` adds a safeguard that validates savings are being achieved during execution:

```yaml
spec:
  executionPolicy:
    achievedSavingsPercentageThreshold: 80
```

| Field                                | Description                                                                                                                                                                               |
| ------------------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `achievedSavingsPercentageThreshold` | Percentage of predicted savings that must be realized during execution. Range: `0`–`100`. `0` means no validation (always proceed). `80` means 80% of predicted savings must be achieved. |

### Aggressive mode

`aggressiveModeConfig` allows rebalancing to include pods that are normally skipped due to safety constraints:

```yaml
spec:
  aggressiveModeConfig:
    ignoreLocalPersistentVolumes: false
    ignoreProblemRemovalDisabledPods: false
    ignoreProblemJobPods: false
    ignoreProblemPodsWithoutController: false
    ignoreInstanceCriteria: false
    aggressiveEviction: false
    drainTimeout: 30m
```

| Field                                | Description                                                                                  |
| ------------------------------------ | -------------------------------------------------------------------------------------------- |
| `ignoreLocalPersistentVolumes`       | Allow rebalancing nodes with local-path-provisioner PVs.                                     |
| `ignoreProblemRemovalDisabledPods`   | Allow rebalancing pods with the `removal-disabled` annotation.                               |
| `ignoreProblemJobPods`               | Allow rebalancing Job and CronJob pods.                                                      |
| `ignoreProblemPodsWithoutController` | Allow rebalancing bare pods without a controller.                                            |
| `ignoreInstanceCriteria`             | Remove instance type constraints from NodePools, allowing broader instance type selection.   |
| `aggressiveEviction`                 | When `true`, sets a drain timeout on NodeClaims to force eviction.                           |
| `drainTimeout`                       | Drain timeout when `aggressiveEviction` is enabled. Duration string (e.g., `"30m"`, `"1h"`). |

> ⚠️ Warning
>
> Enabling `ignoreLocalPersistentVolumes` may cause data loss for workloads using local PVs. Enabling `ignoreProblemPodsWithoutController` affects bare pods that will not be rescheduled automatically.

### Timeouts

```yaml
spec:
  timeouts:
    planGenerationTimeout: 5m
    rebalanceExecutionTimeout: 1h
```

| Field                       | Default | Description                                                        |
| --------------------------- | ------- | ------------------------------------------------------------------ |
| `planGenerationTimeout`     | `5m`    | Maximum time to wait for the rebalancing plan to be generated.     |
| `rebalanceExecutionTimeout` | `1h`    | Maximum time to wait for the rebalancing plan to finish executing. |

# Manage schedules

## View schedules

To list all rebalancing schedules:

```shell
kubectl get rebalanceplanschedules
# or using the short name:
kubectl get rps
```

Example output:

```text
NAME                    SCHEDULE       SUSPEND   ACTIVE   LAST SCHEDULE   AGE
nightly-spot-rebalance  0 2 * * *      false     0        2h              5d
weekend-rebalance       0 0 * * 6      false     0        6d              10d
```

## View claim history

To list all rebalancing plan claims:

```shell
kubectl get rebalanceplanclaims
# or using the short name:
kubectl get rpc
```

Example output:

```text
NAME                                      STATE       PLANID   NODES   SAVINGS%   REBALANCE   AGE
nightly-spot-rebalance-1735776000-123     Completed   abc123   5       12.4       rb-xyz      2h
nightly-spot-rebalance-1735689600-456     Completed   def456   3       8.1        rb-uvw      1d
```

## Check schedule status

To view the full status of a specific schedule:

```shell
kubectl describe rebalanceplanschedule nightly-spot-rebalance
```

The status section shows:

* **`active`** — currently running claims
* **`activeCount`** — number of active claims
* **`lastScheduleTime`** — when the schedule last fired
* **`lastSuccessfulTime`** — when the last claim completed successfully
* **`nextScheduleTime`** — when the schedule will fire next
* **`lastExecution`** — summary of the most recently triggered execution (name, plan ID, state)
* **`conditions`** — `Ready` and `Scheduled` conditions

## Suspend a schedule

To pause a schedule without deleting it, set `suspend: true`:

```shell
kubectl patch rebalanceplanschedule nightly-spot-rebalance \
  --type=merge -p '{"spec":{"suspend":true}}'
```

To resume:

```shell
kubectl patch rebalanceplanschedule nightly-spot-rebalance \
  --type=merge -p '{"spec":{"suspend":false}}'
```

## Delete a schedule

To delete a schedule and its associated claims:

```shell
kubectl delete rebalanceplanschedule nightly-spot-rebalance
```

By default, all `RebalancePlanClaim` objects created by the schedule are also deleted. To retain claim history after deletion, set `preserveHistory: true` in the schedule spec before deleting.

# Schedule behavior

## Concurrency

Only one `RebalancePlanClaim` per schedule can be active at a time. If a claim from the previous execution is still running when the next scheduled time fires, the new execution is skipped and a `ScheduleSkipped` event is recorded.

## Missed schedules

If the controller was unavailable when a schedule was due to fire, the execution is processed when the controller recovers. If `startingDeadlineSeconds` is configured and the missed time exceeded the deadline, the execution is skipped and a `ScheduleMissed` warning event is recorded.

## History cleanup

The controller automatically deletes old claims based on `successfulJobsHistoryLimit` and `failedJobsHistoryLimit`. The oldest claims are removed first. Defaults are `3` for both successful and failed claims.

# Events

The schedule controller emits Kubernetes events that you can view with:

```shell
kubectl describe rebalanceplanschedule <schedule-name>
```

| Event                  | Type    | Description                                                      |
| ---------------------- | ------- | ---------------------------------------------------------------- |
| `ScheduleCreated`      | Normal  | Schedule was first created                                       |
| `ScheduleSuspended`    | Normal  | Schedule was suspended                                           |
| `ScheduleResumed`      | Normal  | Schedule was resumed                                             |
| `ClaimCreated`         | Normal  | A new `RebalancePlanClaim` was created                           |
| `ClaimCreationFailed`  | Warning | Failed to create a `RebalancePlanClaim`                          |
| `ScheduleMissed`       | Warning | A scheduled time was missed (exceeded `startingDeadlineSeconds`) |
| `ScheduleSkipped`      | Normal  | Skipped because a previous claim is still active                 |
| `ScheduleInvalid`      | Warning | The cron expression is invalid                                   |
| `HistoryCleanup`       | Normal  | Old claims were deleted per history limits                       |
| `HistoryCleanupFailed` | Warning | Failed to clean up old claims                                    |

# Troubleshooting

## Schedule not firing

1. Check schedule conditions:
   ```shell
   kubectl describe rps <schedule-name>
   ```
   Look for `ScheduleInvalid` warnings and verify the `Ready` condition is `True`.

2. Verify the rebalancer feature is enabled in Atlas. If disabled, the schedule controller skips all reconciliations.

3. Check if the schedule is suspended (`spec.suspend: true`).

## Claims stuck in Pending or Generating

1. Check the claim status:
   ```shell
   kubectl describe rpc <claim-name>
   ```
   Review `status.conditions` and `status.errorMessage`.

2. Verify the Atlas agent is connected and communicating with Atlas.

3. Check if `minSavingsPercentage` is set too high — if projected savings don't meet the threshold, the plan will not execute and the claim will fail.

## Old claims not being cleaned up

Verify `successfulJobsHistoryLimit` and `failedJobsHistoryLimit` are set to the expected values. The cleanup runs after each schedule reconciliation, so a brief delay after execution is normal.

# Related resources

<Cards columns={2}>
  <Card title="Keeper" href="https://docs.atlas.example/docs/karpenter-enterprise-keeper" icon="fa-cogs">
    The in-cluster controller that executes rebalancing plans and coordinates with Karpenter.
  </Card>
  <Card title="Karpenter Enterprise suite overview" href="https://docs.atlas.example/docs/karpenter-enterprise" icon="fa-cubes">
    How Atlas extends Karpenter with optimization capabilities.
  </Card>
  <Card title="Feature reference" href="https://docs.atlas.example/docs/karpenter-enterprise-features" icon="fa-list">
    Available optimization features for Karpenter-managed clusters.
  </Card>
  <Card title="Scheduled rebalancing (standard)" href="https://docs.atlas.example/docs/scheduled-rebalancing" icon="fa-calendar">
    Console-based scheduled rebalancing for standard Atlas clusters.
  </Card>
</Cards>