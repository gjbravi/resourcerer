---
updatedAt: 2026-10-07T09:08:02.000Z
---

# Continuous rebalancing

Continuous rebalancing is a Keeper feature that autonomously optimizes your cluster on a recurring cycle without any manual intervention. Rather than waiting for a scheduled trigger or a user-initiated rebalancing plan, Keeper continuously monitors your cluster and replaces inefficient nodes with cheaper alternatives whenever it identifies a worthwhile opportunity.

This is distinct from [scheduled rebalancing](karpenter-enterprise-scheduled-rebalancing.md), which fires on a cron expression you define. Continuous rebalancing runs on its own polling interval and reacts to the cluster's current state at each cycle.

<Callout icon="📘" theme="info">
  ### Note

  Continuous rebalancing operates on Karpenter-managed nodes only. Nodes not managed by Karpenter are excluded. See [rebalancing scope](karpenter-enterprise-features.md#rebalancing-scope) for details.
</Callout>

# How it works

On each cycle, Keeper collects the current state of your cluster — nodes, pods, NodePools, NodeClaims, and PodDisruptionBudgets — and intelligently evaluates which nodes are candidates for replacement or removal. Only nodes that have been running for at least the configured minimum age are considered, which prevents churn on recently provisioned nodes.

After evaluating the cluster, Keeper generates a `RebalancePlan` only if projected savings meet both configured thresholds (percentage and absolute monthly amount). The plan is executed automatically — no human approval required. Only one plan can be active at a time; if a plan is already running, the cycle waits until it completes.

If a cycle fails to produce a useful plan, Keeper backs off automatically to avoid generating noise. Repeated failures result in longer backoff intervals before the next attempt.

When a replacement NodeClaim fails because the cloud provider has no capacity for a specific instance type, zone ID, and capacity type, Keeper records that combination as an `UnavailableOffering`. Future cycles avoid the same offering until the record expires. Keeper uses the same mechanism for observed Spot interruptions during rebalancing.

## Modes

Continuous rebalancing supports three modes that control what Keeper is allowed to do each cycle. Modes form a hierarchy — more aggressive modes include the behaviors of simpler ones.

### delete-empty

Removes nodes that have no running workloads. This is the most conservative mode and carries no risk of workload disruption. Keeper deletes the NodeClaim for any empty node whose NodePool has consolidation enabled and that Karpenter has marked as consolidatable.

### drain-only

Bin-packs pods from underutilized nodes onto other nodes with available capacity. Once a node is emptied, it is deleted. Replacement with new nodes does not occur — only existing capacity is used. If a pod cannot be rescheduled on existing nodes, the cycle produces no plan for that run.

This mode includes `delete-empty` behavior.

### full

Replaces underutilized or overpriced nodes with new nodes running cheaper instance types. Keeper creates replacement NodeClaims, drains the old nodes, and deletes them. This is the most aggressive mode and can achieve the greatest cost savings.

This mode includes both `drain-only` and `delete-empty` behaviors.

# Enable Continuous Rebalancing

How you configure Continuous Rebalancing depends on which chart you installed.

## Using the umbrella chart (kent mode)

Most Karpenter installations use the `atlas-umbrella` chart with the `kent` profile, which bundles Keeper as a subchart. Pass Keeper values under the `atlas-keeper` key:

```yaml
atlas-keeper:
  atlas:
    continuousRebalancing:
      enabled: true
      mode: full
```

Apply the change:

```shell
helm upgrade atlas atlas/atlas-umbrella \
  --reset-then-reuse-values \
  -f values.yaml
```

## Using the standalone Keeper chart

If you installed `atlas-keeper` directly, set values at the top level:

```yaml
atlas:
  continuousRebalancing:
    enabled: true
    mode: full
```

Apply the change:

```shell
helm upgrade atlas-keeper atlas/atlas-keeper \
  --reset-then-reuse-values \
  -f values.yaml
```

# Configuration reference

## Helm values

The following values are available under `atlas.continuousRebalancing` (or `atlas-keeper.atlas.continuousRebalancing` when using the umbrella chart).

| Value                        | Default | Description                                                                                                                    |
| ---------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `enabled`                    | `false` | Enable or disable Continuous Rebalancing.                                                                                      |
| `mode`                       | `full`  | Rebalancing mode: `delete-empty`, `drain-only`, or `full`.                                                                     |
| `cycleIntervalSeconds`       | `60`    | How often (in seconds) the rebalancing cycle runs. Lower values increase responsiveness but also churn.                        |
| `savingsThresholdPercentage` | `15`    | Minimum projected savings percentage required before executing a plan. Applies to `full` mode only.                            |
| `minNodeAgeSeconds`          | `300`   | Minimum node age in seconds before a node is considered a candidate. Prevents rebalancing freshly provisioned nodes.           |
| `evictionConfig`             | `[]`    | Ordered list of selector + settings pairs controlling pod and node eviction behavior. See [Eviction config](#eviction-config). |

## Advanced environment variables

For fine-grained control, you can set additional environment variables on the Keeper deployment. These are typically configured at installation time.

| Variable                                                | Default   | Description                                                                                                                                                                         |
| ------------------------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CONTINUOUS_REBALANCING_MIN_NODES_TO_CONSIDER`          | `1`       | Minimum number of eligible nodes required to run a cycle.                                                                                                                           |
| `CONTINUOUS_REBALANCING_MAX_NODES_PER_ITERATION`        | `10`      | Maximum number of candidate nodes evaluated per cycle.                                                                                                                              |
| `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_COST_MONTHLY` | `50.0`    | Minimum absolute monthly savings in USD required before executing a plan.                                                                                                           |
| `CONTINUOUS_REBALANCING_FAILURE_BACKOFF_DURATION`       | `30m`     | Base backoff duration after a failed cycle.                                                                                                                                         |
| `NODE_POOL_FILTER_MODE`                                 | `opt-out` | NodePool consolidation filter mode: `opt-out` (all NodePools unless annotated `exclude`) or `opt-in` (only NodePools annotated `include`). See [Node exclusions](#node-exclusions). |

# Node exclusions

## Protecting individual nodes

To prevent a specific node from being selected by Continuous Rebalancing, apply the `autoscaling.atlas.example/removal-disabled` label:

```shell
kubectl label node <node-name> autoscaling.atlas.example/removal-disabled=true
```

Remove the label to make the node eligible again:

```shell
kubectl label node <node-name> autoscaling.atlas.example/removal-disabled-
```

## NodePool-level exclusions

Continuous rebalancing respects Karpenter's consolidation policies at the NodePool level:

* **Static NodePools** (fixed replica count) — nodes from these pools are never selected for rebalancing.
* `WhenEmpty`**&#x20;consolidation policy** — only empty nodes are selected from NodePools with `consolidationPolicy: WhenEmpty`. Nodes with running workloads in such pools are excluded from `drain-only` and `full` modes.

```yaml
spec:
  disruption:
    consolidationPolicy: WhenEmpty
    consolidateAfter: 30s
```

You can also exclude NodePools explicitly with the `keeper.atlas.example/consolidation` annotation. See [Excluding NodePools with the consolidation annotation](#excluding-nodepools-with-the-consolidation-annotation).

## NodePool disruption budgets

Continuous rebalancing respects the `disruption.budgets` configuration on your NodePools. A rebalancing plan can include as many nodes as needed to achieve cost savings, but the actual deletions are gated by each NodePool's budget at execution time.

During the deletion phase, Keeper checks how many nodes in a given NodePool are already being disrupted (draining or NotReady) and compares that against the budget limit. If the budget is exhausted, those deletions are deferred. Keeper re-checks every 30 seconds — as draining nodes complete and free up budget slots, the remaining deletions proceed automatically.

This means a single plan can safely replace a large number of nodes while still honoring the disruption rate limits you've configured in Karpenter. You don't need to size your rebalancing plans conservatively to avoid overwhelming a NodePool.

## Excluding NodePools with the consolidation annotation

The `keeper.atlas.example/consolidation` NodePool annotation controls which NodePools Keeper may rebalance, independent of Karpenter's disruption policies. The top-level `nodePoolFilterMode` Helm value selects how Keeper applies the annotation:

| Value               | Behavior                                                            |
| ------------------- | ------------------------------------------------------------------- |
| `opt-out` (default) | All NodePools participate in rebalancing unless annotated `exclude` |
| `opt-in`            | Only NodePools annotated `include` participate in rebalancing       |

Annotate the NodePool with the value matching your filter mode:

<Tabs>
  <Tab title="Exclude (opt-out mode)">
    ```yaml
    apiVersion: karpenter.sh/v1
    kind: NodePool
    metadata:
      name: example-nodepool
      annotations:
        keeper.atlas.example/consolidation: "exclude"
    ```
  </Tab>

  <Tab title="Include (opt-in mode)">
    ```yaml
    apiVersion: karpenter.sh/v1
    kind: NodePool
    metadata:
      name: example-nodepool
      annotations:
        keeper.atlas.example/consolidation: "include"
    ```
  </Tab>
</Tabs>

`nodePoolFilterMode` is a top-level value, not nested under `atlas.continuousRebalancing`, because it also governs scheduled and API-triggered rebalancing:

<Tabs>
  <Tab title="Umbrella chart (kent mode)">
    ```yaml
    atlas-keeper:
      nodePoolFilterMode: opt-in
    ```
  </Tab>

  <Tab title="Standalone Keeper chart">
    ```yaml
    nodePoolFilterMode: opt-in
    ```
  </Tab>
</Tabs>

The filter mode is a static, environment-based setting. Changing it requires a Keeper pod restart.

<Callout icon="⚠️" theme="warn">
  ### Exclusion blocks every rebalancing path

  Keeper skips nodes of excluded NodePools for all rebalancing activity, not just cost-driven consolidation:

  - Continuous rebalancing in every mode, including `delete-empty`: even fully empty nodes are never removed.
  - Drift replacement: a drifted NodeClaim in an excluded NodePool is never replaced.
  - Scheduled rebalancing and API-triggered `RebalancePlanClaims`: excluded nodes produce no candidates regardless of the claim's scope.

  Aggressive mode and eviction config cannot override exclusion.

  Switching to `opt-in` mode without annotating any NodePools with `include` excludes every pool and stops all rebalancing. Annotate the pools you want included before enabling the mode.
</Callout>

The filter controls which existing nodes Keeper may disrupt, not where it places new capacity. Rebalancing plans may still create new nodes in an excluded NodePool when it offers the best capacity for a replacement. These new nodes belong to the excluded pool: Keeper will not consolidate them, delete them when empty, or replace them later.

# Eviction config

Eviction config gives you fine-grained control over how Keeper handles individual pods and nodes during rebalancing. It is an ordered list of selector + settings pairs — each entry matches pods, nodes, or both and applies an eviction policy to them.

By default, no eviction config is set and all pods are subject to standard eviction rules: PodDisruptionBudgets are respected, bare pods block node drain, and so on.

## Behaviors

Each selector entry can enable one or more of the following settings:

`removalDisabled` — Prevents Keeper from evicting or draining matching pods or nodes. Use this for critical workloads or dedicated nodes that must never be disrupted.

`aggressive` — Allows evicting pods that would ordinarily block node drain: single-replica pods and pods protected by a PodDisruptionBudget. Local persistent volumes are still respected.

`disposable` — Removes all eviction guards. Matching pods can be evicted unconditionally regardless of PDBs, replica count, or local storage. Suitable for batch jobs and ephemeral workloads.

The strictness order is `removalDisabled` > `disposable` > `aggressive`. This order applies both when multiple behaviors are enabled on a single entry and when multiple entries match the same pod or node — the most restrictive matching setting always wins.

## Selectors

Each entry targets pods, nodes, or both:

`podSelector` — Matches pods by one or more criteria:

| Field           | Description                                                                                                               |
| --------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `namespace`     | Namespace to match. Empty matches all namespaces.                                                                         |
| `kind`          | Owner kind to match, e.g. `Deployment`, `StatefulSet`, `Job`. Empty matches any owner.                                    |
| `replicasMin`   | Minimum desired replica count of the owning controller. `0` disables the filter. Does not apply to DaemonSet or Job pods. |
| `labelSelector` | Standard Kubernetes label selector (`matchLabels` / `matchExpressions`).                                                  |

`nodeSelector` — Matches nodes by:

| Field           | Description                         |
| --------------- | ----------------------------------- |
| `labelSelector` | Standard Kubernetes label selector. |

Each entry must use either `podSelector` or `nodeSelector`, not both.

## Configuration

Add the `evictionConfig` list under `atlas.continuousRebalancing`. The following example prevents Keeper from touching GPU nodes, and treats batch Jobs as fully disposable:

```yaml
atlas:
  continuousRebalancing:
    enabled: true
    mode: full
    evictionConfig:
      - nodeSelector:
          labelSelector:
            matchLabels:
              dedicated: gpu
        settings:
          removalDisabled:
            enabled: true
      - podSelector:
          kind: Job
          namespace: batch
        settings:
          disposable:
            enabled: true
```

Use `matchExpressions` for more flexible matching. The following example aggressively evicts pods belonging to a Deployment with at least 3 replicas, as long as they are not labeled `environment: production`:

```yaml
evictionConfig:
  - podSelector:
      kind: Deployment
      replicasMin: 3
      labelSelector:
        matchExpressions:
          - key: environment
            operator: NotIn
            values:
              - production
    settings:
      aggressive:
        enabled: true
```

Apply the change:

```shell
helm upgrade atlas-keeper atlas/atlas-keeper \
  --reset-then-reuse-values \
  -f values.yaml
```

For the umbrella chart, nest the values under `atlas-keeper.atlas.continuousRebalancing.evictionConfig`.

# Savings thresholds

A rebalancing plan is only executed if projected savings meet **both** of the following thresholds:

* **Percentage threshold** — projected savings must be at or above `savingsThresholdPercentage` (default: 15%)
* **Absolute monthly threshold** — projected savings in USD must be at or above `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_COST_MONTHLY` (default: $50/month)

If either threshold is not met, no plan is created and the cycle waits for the next polling interval.

# Troubleshooting

## Continuous rebalancing is enabled but no plans are created

1. Verify the feature is enabled and the mode is set. For the umbrella chart:
   ```shell
   helm get values atlas | grep -A5 continuousRebalancing
   ```
   For the standalone chart:
   ```shell
   helm get values atlas-keeper | grep -A5 continuousRebalancing
   ```

2. Check that enough candidate nodes exist. If fewer nodes are available than `CONTINUOUS_REBALANCING_MIN_NODES_TO_CONSIDER`, the cycle exits early without creating a plan.

3. Verify the savings thresholds are reachable. If the cluster is already well-optimized, plans may be discarded because projected savings fall below `CONTINUOUS_REBALANCING_SAVINGS_THRESHOLD_COST_MONTHLY`.

4. Check the NodePool consolidation filter. In `opt-in` mode, Keeper rebalances only NodePools annotated `keeper.atlas.example/consolidation: "include"`; if no pools carry it, rebalancing stops entirely. Pools annotated `exclude` are skipped in every mode. See [Node exclusions](#node-exclusions).

5. Check Keeper logs for cycle output:
   ```shell
   kubectl logs -n atlas-agent -l app=atlas-keeper | grep "continuous-rebalancing"
   ```

## A node is not being selected despite being underutilized

Check whether the node has the `autoscaling.atlas.example/removal-disabled=true` label:

```shell
kubectl get node <node-name> --show-labels
```

Also verify the node's NodePool does not use a `WhenEmpty` consolidation policy and that the node is older than `CONTINUOUS_REBALANCING_MIN_NODE_AGE`.

The `keeper.atlas.example/consolidation` annotation can also exclude the node's entire NodePool: in `opt-in` mode, Keeper rebalances only pools annotated `include`. To list every NodePool's annotation:

```shell
kubectl get nodepool -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.annotations.keeper\.atlas\.example/consolidation}{"\n"}{end}'
```

The node may also be excluded due to problematic pods running on it. Common causes include:

* A PodDisruptionBudget that allows zero disruptions, blocking eviction of any pod on the node
* Pods without a controller (bare pods) that cannot be rescheduled
* Pods with local persistent volumes

To check for restrictive PDBs:

```shell
kubectl get pdb -A
```

Look for any PDB where `ALLOWED DISRUPTIONS` is `0` and whose selector matches pods on the node in question.

## Continuous rebalancing is backing off

If the controller has encountered repeated plan failures, it applies exponential backoff. Check the logs for backoff messages:

```shell
kubectl logs -n atlas-agent -l app=atlas-keeper | grep -i "backoff\|consecutive"
```

Backoff resets after a successful cycle.

# Related resources

<Cards columns={2}>
  <Card title="Keeper" href="https://docs.atlas.example/docs/karpenter-enterprise-keeper" icon="fa-cogs">
    The in-cluster controller that runs Continuous Rebalancing and coordinates with Karpenter.
  </Card>
  <Card title="Scheduled rebalancing for Karpenter" href="https://docs.atlas.example/docs/karpenter-enterprise-scheduled-rebalancing" icon="fa-clock">
    Cron-based rebalancing using Kubernetes-native CRDs.
  </Card>
  <Card title="Feature reference" href="https://docs.atlas.example/docs/karpenter-enterprise-features" icon="fa-list">
    All optimization features available for Karpenter-managed clusters.
  </Card>
  <Card title="Karpenter Enterprise suite overview" href="https://docs.atlas.example/docs/karpenter-enterprise" icon="fa-cubes">
    How Atlas extends Karpenter with optimization capabilities.
  </Card>
</Cards>

<br />