---
updatedAt: 2026-09-04T13:00:29.000Z
---

# Keeper

Keeper is the in-cluster control-plane component that coordinates Atlas's node rebalancing and automation features with Karpenter. It runs inside your cluster as a Kubernetes controller, watches Atlas custom resources, and drives optimization workflows by interacting directly with Karpenter's NodeClaims and NodePools.

# What Keeper does

Keeper is responsible for:

* Generating and reconciling rebalancing plans using CRDs in the `autoscaling.atlas.example/v1alpha` API group
* Handling claim- and schedule-driven rebalancing flows
* Running continuous in-cluster rebalancing to find cost-saving node replacements
* Replacing spot instances proactively based on interruption predictions
* Recording unavailable offerings when local capacity failures occur
* Coordinating node consolidation with Karpenter NodeClaim deletion
* Managing commitment budgets that steer Karpenter toward Reserved Instances and Savings Plans

# How it works

## Component interactions

Keeper interacts with the following systems:

| System                   | Purpose                                                                              |
| ------------------------ | ------------------------------------------------------------------------------------ |
| Kubernetes API server    | Watches and updates Nodes, Pods, NodeClaims, NodePools, and Atlas CRDs             |
| Karpenter                | Reads NodeClaims and NodePools; delegates node and instance termination to Karpenter |
| Atlas API              | Plan coordination, configuration, capacity failure reporting, and audit logging      |
| Atlas ML API           | Spot interruption prediction model queries                                           |
| AWS EC2 and Pricing APIs | Instance inventory and cost-aware rebalancing decisions                              |

Keeper maintains a persistent connection to the Atlas backend for real-time plan coordination. This connection is automatically re-established if it drops.

## CRDs managed by Keeper

Keeper defines and reconciles the following Custom Resource Definitions (CRDs) in the `autoscaling.atlas.example/v1alpha` API group:

| CRD                      | Short name | Scope   | Description                                                                                                                          |
| ------------------------ | ---------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `RebalancePlanSchedule`  | `rps`      | Cluster | Defines when rebalancing should run on a cron schedule                                                                               |
| `RebalancePlanClaim`     | `rpc`      | Cluster | Represents a single rebalancing execution request                                                                                    |
| `RebalancePlan`          | `rp`       | Cluster | Holds the concrete plan: which NodeClaims to add, remove, and which pods to migrate                                                  |
| `RebalanceMigrationPlan` | —          | Cluster | Coordinates container live migrations as part of a rebalancing plan                                                                  |
| `UnavailableOffering`    | `uo`       | Cluster | Records an instance type, zone ID, and capacity type combination that Keeper should avoid temporarily                            |
| `CommitmentBudget`       | `cb`       | Cluster | Defines an AWS Reserved Instance or Savings Plan commitment to steer Karpenter NodePool provisioning toward committed instance types |

## Feature gating

All Keeper features are enabled and disabled dynamically through the Atlas console without restarting the controller.

## Capacity failure observations

Keeper records an `UnavailableOffering` when it learns that a specific offering should be avoided temporarily. This can happen after an observed capacity failure, such as Karpenter being unable to create a NodeClaim because the cloud provider reports an insufficient capacity error (ICE), or when Keeper observes a Spot interruption. It can also happen when Spot interruption prediction marks a Spot offering as risky.

An unavailable offering is a specific combination of instance type, zone ID, and capacity type. Keeper uses this record locally so later plans avoid retrying the same failing combination until the record expires.

Keeper also reports observed ICE failures and observed Spot interruptions to Atlas over its existing connection. Atlas uses these observations, along with other capacity data, to improve capacity intelligence for Karpenter-backed clusters.

Keeper reports only local observations. It does not report prediction-only `UnavailableOffering` records as observed interruptions. Reporting an observation does not change local `UnavailableOffering` expiration, backoff, reason, or cleanup behavior.

# Rebalancing workflows

Keeper supports three ways to initiate rebalancing:

## Schedule-driven rebalancing

When a `RebalancePlanSchedule` fires, Keeper creates a `RebalancePlanClaim` from the schedule template and submits it to Atlas for plan generation. For details, see [Scheduled rebalancing for Karpenter clusters](karpenter-enterprise-scheduled-rebalancing.md).

## Claim-driven rebalancing

A `RebalancePlanClaim` can also be created manually without a schedule. This triggers a single on-demand rebalancing operation with the configuration specified in the claim's `spec`.

## Continuous in-cluster rebalancing

The Continuous Rebalancing controller runs periodically inside the cluster and optimizes nodes without calling the Atlas backend for plan generation. It works entirely from local cluster state and AWS pricing data:

1. **Resource collection** — Keeper reads all Nodes, NodeClaims, NodePools, Pods, DaemonSets, PodDisruptionBudgets, and EC2NodeClasses from the cluster.
2. **Node analysis** — Nodes are classified as candidates for replacement based on age, workload compatibility, and NodePool constraints. Nodes with blocking workloads (such as local PVs or pods without a controller) are excluded unless aggressive mode is configured.
3. **Optimization search** — Keeper evaluates candidate node sets to find the largest subset of nodes that can be replaced for a net cost saving.
4. **Savings validation** — A plan is only executed if it meets both a savings percentage threshold (default: 0%) and an absolute monthly savings threshold (default: $50/month).
5. **Plan creation** — A `RebalancePlan` is created directly in the cluster without a `RebalancePlanClaim`. Keeper then executes the plan by provisioning replacement NodeClaims and deleting the originals.

The cycle repeats on a configurable polling period (default: 10 seconds). Only one active plan can exist at a time — if a plan is already running, the cycle skips until it completes.

## Spot interruption prediction

The spot interruption prediction service polls the Atlas ML API at a configurable interval (default: every 1 minute) to identify spot nodes at risk of interruption. When a node is predicted to be interrupted:

1. Keeper labels the node with `autoscaling.atlas.example/predicted-interruption=true`
2. A `RebalancePlanClaim` is created with the at-risk node in its scope, triggering proactive replacement
3. A `SpotInterruptionPredicted` event is emitted on the node
4. After replacement, a cooldown period (default: 15 minutes) prevents the same node from being replaced again immediately

This gives workloads significantly more lead time than AWS's standard two-minute interruption warning.

# Rebalance execution

## Plan lifecycle

Once a `RebalancePlanClaim` is created, it progresses through the following states:

| State        | Description                                               |
| ------------ | --------------------------------------------------------- |
| `Pending`    | Claim created, not yet submitted to Atlas               |
| `Generating` | Atlas is computing the rebalancing plan                 |
| `Ready`      | Plan generated, waiting for execution to begin            |
| `Executing`  | Rebalancing is actively running (nodes being replaced)    |
| `Completed`  | Rebalancing completed successfully                        |
| `Failed`     | Rebalancing failed; see `status.errorMessage` for details |

## Execution phases

When a `RebalancePlan` runs, Keeper proceeds in two phases:

**Creation phase** — New NodeClaims are provisioned through Karpenter. If a provisioning attempt fails due to insufficient capacity, Keeper retries with the next available instance type from the NodePool's requirements. Each attempt is recorded in `status.nodesCreation`:

| Field           | Description                                               |
| --------------- | --------------------------------------------------------- |
| `instanceType`  | The instance type attempted                               |
| `nodeClaimName` | NodeClaim name, including `-attempt-N` suffix for retries |
| `status`        | `InProgress`, `Success`, or `Failed`                      |
| `description`   | Details, such as `InsufficientCapacityError`              |

**Deletion phase** — Old NodeClaims are drained and deleted. If container live migration is available, Keeper creates a `RebalanceMigrationPlan` to move workloads without restarts before deleting the node. Deletion events are recorded in `status.nodesDeletion`:

| Status                    | Description                                            |
| ------------------------- | ------------------------------------------------------ |
| `LiveMigrationCreated`    | A `RebalanceMigrationPlan` was created for this node   |
| `LiveMigrationInProgress` | Live migration is running                              |
| `LiveMigrationSucceeded`  | Live migration completed; node is safe to delete       |
| `LiveMigrationFailed`     | Live migration failed; falls back to standard eviction |
| `InProgress`              | NodeClaim deletion in progress                         |
| `Success`                 | NodeClaim deleted successfully                         |
| `Failed`                  | NodeClaim deletion failed                              |

## Savings validation

If `executionPolicy.achievedSavingsPercentageThreshold` is set on the claim, Keeper validates actual savings after new nodes are provisioned but before old nodes are deleted. If the realized savings do not meet the threshold, the plan is aborted with `FailureReasonInsufficientSavings`.

Validated savings are stored in `status.validatedSavings` on the `RebalancePlan`. If a fallback instance type was used due to insufficient capacity, the originally planned price is used as a conservative estimate.

## Failure recovery

When a plan fails, Keeper records which phase caused the failure in `status.failurePhase` to ensure correct cleanup after a controller restart:

| Phase      | Meaning                                     | Cleanup action                                           |
| ---------- | ------------------------------------------- | -------------------------------------------------------- |
| `Creation` | Failed while provisioning new NodeClaims    | Roll back by deleting newly created NodeClaims           |
| `Deletion` | Failed while draining or deleting old nodes | Uncordon new nodes so Karpenter can manage them normally |

## Savings reporting

After plan generation, Atlas populates `status.savings` on the `RebalancePlan`:

| Field                                                       | Description                                             |
| ----------------------------------------------------------- | ------------------------------------------------------- |
| `projectedSavingsPercent`                                   | Percentage savings relative to current node spend       |
| `projectedSavingsCostMonthly`                               | Estimated monthly savings                               |
| `currentMonthlyCost` / `projectedMonthlyCost`               | Before and after monthly cost for the rebalanced nodes  |
| `currentClusterMonthlyCost` / `projectedClusterMonthlyCost` | Before and after monthly cost for the entire cluster    |
| `blueNodes` / `greenNodes`                                  | Per-node cost details for nodes being removed and added |

# Node consolidation

The node consolidation controller watches `Node` objects and deletes the corresponding Karpenter `NodeClaim` when consolidation conditions are met. Deleting the `NodeClaim` delegates final node and cloud instance termination to Karpenter.

Consolidation is triggered through three paths:

| Path           | Trigger condition                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------------- |
| **Expiration** | `NodeClaim.spec.expireAfter` has elapsed                                                                             |
| **CRB drain**  | Keeper's Continuous Rebalancing has drained the node                                                             |
| **Empty node** | Node is empty, NodePool has consolidation enabled, and the node has `Consolidatable=True` condition set by Karpenter |

Node consolidation is feature-gated and can be enabled or disabled dynamically without restarting Keeper.

# Configuration

## Required environment variables

| Variable     | Description                                 |
| ------------ | ------------------------------------------- |
| `CLUSTER_ID` | Atlas cluster identifier                  |
| `API_KEY`    | Atlas API key for authentication          |
| `API_URL`    | Atlas REST API endpoint                   |
| `GRPC_URL`   | Atlas gRPC endpoint for plan coordination |

<Accordion title="Optional environment variables" icon="fa-cog">

| Variable | Default | Description |
|----------|---------|-------------|
| `GRPC_DISABLE_TLS` | `false` | Disables TLS for gRPC connections |
| `FEATURES_WATCH_NAMESPACES` | `atlas-agent` | Namespaces to watch for feature state ConfigMaps |
| `DYNAMIC_CONFIG_MAP_NAME` | — | ConfigMap name for dynamic configuration |
| `NODE_CONSOLIDATION_MAX_CONCURRENT_RECONCILES` | `10` | Parallel node consolidation operations |
| `REBALANCER_MAX_CONCURRENT_RECONCILES` | `10` | Parallel rebalancing plan operations |

</Accordion>

<Accordion title="Continuous rebalancing variables" icon="fa-sync">

| Variable | Default | Description |
|----------|---------|-------------|
| `CONTINUOUS_REBALANCING_MIN_NODE_AGE` | `5m` | Minimum node age before a node is considered a rebalancing candidate |
| `CONTINUOUS_REBALANCING_MIN_NODES_TO_CONSIDER` | `1` | Minimum number of candidate nodes required to run a cycle |
| `CONTINUOUS_REBALANCING_MAX_NODES_PER_ITERATION` | `10` | Maximum candidates evaluated per run |
| `CONTINUOUS_REBALANCING_MAX_BINARY_SEARCH_DEPTH` | `20` | Maximum search depth |
| `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_PERCENT` | `0.0` | Minimum savings percentage relative to replaced nodes |
| `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_COST_MONTHLY` | `50.0` | Minimum absolute monthly savings in USD |
| `CONTINUOUS_REBALANCING_FAILURE_BACKOFF_DURATION` | `30m` | Backoff after a failed cycle |
| `CONTINUOUS_REBALANCING_POLLING_PERIOD` | `10s` | How often the Continuous Rebalancing cycle runs |

</Accordion>

<Accordion title="Spot interruption prediction variables" icon="fa-bolt">

| Variable | Default | Description |
|----------|---------|-------------|
| `SPOT_INTERRUPTION_PREDICTION_MODEL_NAME` | — | ML model name to use for predictions |
| `SPOT_INTERRUPTION_PREDICTION_POLL_INTERVAL` | `1m` | How often to poll the Atlas ML API |
| `SPOT_INTERRUPTION_PREDICTION_SPOT_REPLACEMENT_COOLDOWN` | `15m` | Cooldown period before the same node can be replaced again |
| `SPOT_INTERRUPTION_PREDICTION_INTERRUPTION_THRESHOLD` | — | Probability threshold (0–1) above which a node is considered at risk |

</Accordion>

# View resources

## List UnavailableOfferings

To list unavailable offerings recorded by Keeper:

```shell
kubectl get unavailableofferings
# or using the short name:
kubectl get uo
```

Example output:

```text
NAME                     INSTANCE TYPE   ZONE ID    CAPACITY   REASON                         EXPIRES   ATTEMPTS   AGE
m5-large-use1-az1-spot   m5.large        use1-az1   spot       SpotInterruptionObserved       25m       1          5m
c6i-xlarge-use1-az2-od   c6i.xlarge      use1-az2   on-demand  InsufficientCapacityError      18m       2          12m
```

To inspect one record:

```shell
kubectl describe unavailableoffering <name>
```

Check `spec.reason`, `spec.expiresAt`, and `spec.attempts` to understand why Keeper is avoiding that offering and when it can be considered again.

## List RebalancePlans

To list all rebalancing plans in the cluster:

```shell
kubectl get rebalanceplans
# or using the short name:
kubectl get rp
```

Example output:

```text
NAME                          BEFORE     AFTER      SAVINGS%   STATE   AUTO-EXECUTED   AGE
nightly-rebalance-1735776000  1234.56    1111.11    10.0       Done    true            2h
```

## List RebalancePlanClaims

To list all rebalancing plan claims:

```shell
kubectl get rebalanceplanclaims
# or using the short name:
kubectl get rpc
```

Example output:

```text
NAME                                  STATE       PLANID   NODES   SAVINGS%   REBALANCE   AGE
nightly-spot-rebalance-1735776000     Completed   abc123   5       12.4       rp-xyz      2h
```

## Inspect a RebalancePlan

To view the full details of a specific plan:

```shell
kubectl describe rebalanceplan <plan-name>
```

The status section shows:

* `state` — current execution state: `Pending`, `Running`, `Done`, `DoneWithWarnings`, or `Failed`
* `savings` — cost details for current and projected node configurations
* `nodesCreation` — per-node creation attempt history
* `nodesDeletion` — per-node deletion event history
* `failurePhase` — if failed, which phase caused the failure (`Creation` or `Deletion`)

# Troubleshooting

## RebalancePlanClaim stuck in Pending or Generating

1. Check the claim status:
   ```shell
   kubectl describe rpc <claim-name>
   ```
   Review `status.conditions` and `status.errorMessage`.

2. Verify the Atlas agent is connected and communicating with Atlas.

3. Check if `minSavingsPercentage` is set too high — if projected savings don't meet the threshold, the claim will fail without executing.

## RebalancePlan stuck in Running

1. Check for failed node creation attempts:
   ```shell
   kubectl describe rp <plan-name>
   ```
   Look at `status.nodesCreation` for `InsufficientCapacityError` on all instance types. If all attempts are exhausted, the plan transitions to `Failed` with `FailureReasonNodeCreateFailed`.

2. Check `status.nodesDeletion` for stuck live migrations. If a `RebalanceMigrationPlan` is stuck, inspect it directly:
   ```shell
   kubectl get rebalancemigrationplans
   kubectl describe rebalancemigrationplan <name>
   ```

## Continuous rebalancing not running

1. Verify the Continuous Rebalancing feature is enabled in Atlas.

2. Check that enough candidate nodes exist — `CONTINUOUS_REBALANCING_MIN_NODES_TO_CONSIDER` must be met.

3. Verify the savings thresholds are reachable. If the cluster is already well-optimized, plans may be generated but discarded because projected savings fall below `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_COST_MONTHLY`.

4. Check Keeper logs:
   ```shell
   kubectl logs -n atlas-agent -l app=atlas-keeper | grep "continuous-rebalancing"
   ```

## Node consolidation not triggering

1. Verify the node consolidation feature is enabled in Atlas.

2. For the empty node path, check that the NodePool has consolidation enabled and that Karpenter is setting the `Consolidatable=True` condition on empty nodes.

3. Check Keeper logs:
   ```shell
   kubectl logs -n atlas-agent -l app=atlas-keeper | grep "empty-node-deleter"
   ```

## Spot interruption replacements not happening

1. Verify the spot interruption prediction feature is enabled in Atlas.

2. Check that `SPOT_INTERRUPTION_PREDICTION_MODEL_NAME` is set and that Keeper has connectivity to the Atlas ML API.

3. Look for `SpotInterruptionPredicted` events on spot nodes:
   ```shell
   kubectl get events --field-selector reason=SpotInterruptionPredicted
   ```

## Keeper keeps avoiding an instance offering

1. Check whether Keeper recorded an `UnavailableOffering`:
   ```shell
   kubectl get uo
   kubectl describe uo <name>
   ```

2. Review `spec.reason` to see whether the offering was blocked because of an insufficient capacity error, an observed Spot interruption, or a predicted Spot interruption.

3. Check `spec.expiresAt`. Keeper ignores expired entries, and the garbage collector removes them later.

4. If the same offering is marked unavailable repeatedly, check `spec.attempts`. Repeated failures extend the local backoff window.

# Related resources

<Cards columns={2}>
  <Card title="Scheduled rebalancing for Karpenter" href="https://docs.atlas.example/docs/karpenter-enterprise-scheduled-rebalancing" icon="fa-clock">
    Set up cron-based rebalancing using Kubernetes-native CRDs.
  </Card>
  <Card title="Karpenter Enterprise suite overview" href="https://docs.atlas.example/docs/karpenter-enterprise" icon="fa-cubes">
    How Atlas extends Karpenter with optimization capabilities.
  </Card>
  <Card title="Feature reference" href="https://docs.atlas.example/docs/karpenter-enterprise-features" icon="fa-list">
    Available optimization features for Karpenter-managed clusters.
  </Card>
  <Card title="Hosted components" href="https://docs.atlas.example/docs/hosted-components" icon="fa-server">
    All Atlas in-cluster components and their roles.
  </Card>
</Cards>

<br />