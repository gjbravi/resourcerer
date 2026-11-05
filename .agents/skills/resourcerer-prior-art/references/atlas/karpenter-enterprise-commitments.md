---
updatedAt: 2026-10-16T09:23:19.000Z
---

# Commitments in Karpenter Enterprise

Atlas automatically puts AWS commitments — Reserved Instances and Savings Plans — to work in Karpenter Enterprise clusters. Karpenter has no native understanding of commitments, so it may provision full-price on-demand or spot capacity while discounted committed capacity sits unused. Atlas closes this gap by creating high-weight *shadow NodePools* that Karpenter prefers over regular NodePools, steering provisioning toward committed instance types.

### Commitment support by cloud provider:

| Cloud provider                             | Commitments support                             |
| ------------------------------------------ | ----------------------------------------------- |
| AWS (Amazon EKS)                           | Supported: Reserved Instances and Savings Plans |
| Azure (AKS, self-managed Karpenter or NAP) | Not yet supported                               |

The feature is fully managed: assign commitments to a Karpenter cluster in the Atlas console, and Atlas handles the rest — syncing, matching, limit rebalancing, and cleanup.

> 📘 Note
>
> This feature is specific to Karpenter Enterprise clusters managed by [Keeper](karpenter-enterprise-keeper.md). With the standard Atlas Autoscaler, commitments are consumed automatically — see [Commitments](commitments.md).

# Prerequisites

Commitment budgets need two things in place:

1. **Keeper v0.1.162 or later** — the minimum version with Savings Plan support. Older Keeper versions receive only Reserved Instance budgets; Savings Plan budgets are omitted until the upgrade.
2. **The CommitmentBudgets controller enabled** — on by default in the Keeper Helm chart (`atlas.commitmentBudgets.enabled: true`), so no action is needed unless it was explicitly turned off.
3. **Commitments must be assigned to the cluster** through the Atlas [Commitments](commitments.md) page. Until at least one commitment is assigned, no `CommitmentBudget` resources are synced and no shadow NodePools are created.

To enable (or re-enable) the controller in Helm, most Karpenter installations use the `atlas-umbrella` chart with the `kent` profile, which bundles Keeper as a subchart. Keeper values are passed under the `atlas-keeper` key:

```yaml
atlas-keeper:
  atlas:
    commitmentBudgets:
      enabled: true
```

If you installed the standalone `atlas-keeper` chart, set the value at the top level:

```yaml
atlas:
  commitmentBudgets:
    enabled: true
```

# How it works

The flow has two stages: Atlas syncs the assigned commitments into the cluster, and Keeper turns them into shadow NodePools.

## 1. Assign commitments in the Atlas platform

1. Atlas reads the cluster's assigned commitments.
2. Assigned commitments are converted to `CommitmentBudget` CRDs with the appropriate type (RI, Compute SP, or EC2 Instance SP). Multiple commitments may be merged into a single budget.
3. For Savings Plans, Atlas fetches per-family CPU pricing from the AWS pricing API and embeds it in the CRD spec.
4. CRDs are synced to the cluster automatically — created, updated, or deleted as commitments change.
5. Atlas resyncs periodically, and immediately when a commitment is added, removed, or reassigned.

## 2. Keeper creates shadow NodePools

Keeper watches `CommitmentBudget` CRDs and reconciles them into shadow NodePools:

1. **Matches budgets to NodePools** — For each synced budget, which may represent one or more merged commitments, Keeper finds regular NodePools whose scheduling requirements are compatible with the budget's instance type or family constraints. Compatibility is validated against AWS instance inventory to ensure at least one real instance type satisfies both the NodePool's and the budget's constraints.
2. **Creates shadow NodePools** — For each matching budget–NodePool pair, Keeper creates a dedicated shadow NodePool with `weight: 100` and a `zzz-` name prefix. Karpenter's equal-weight tiebreaker picks the lexicographically highest name, so these pools are preferred even over NodePools that also set `weight: 100`.
3. **Rebalances limits** — Every 2 minutes, Keeper redistributes the budget's limits across matched NodePools proportionally to their observed usage. For zonal RIs, usage is measured in node count; for regional RIs and Savings Plans, usage is measured in CPU millicores.
4. **Drains and deletes on removal** — When a budget is removed because its commitments were unassigned or deleted in the platform, or the feature is disabled, Keeper drains the shadow NodePools by setting all limits to 0, waits for nodes to reach 0, then deletes the NodePools.

## Why only spot-allowing NodePools are matched

Shadow NodePools always run on-demand capacity. They add value by replacing spot instances with committed on-demand instances that cost less. A NodePool that only allows on-demand already provisions committed instances through normal Karpenter scheduling, so creating a shadow pool for it would add no benefit.

# Supported commitment types

Atlas supports all three AWS commitment types — Reserved Instances (in zonal and regional variants), Compute Savings Plans, and EC2 Instance Savings Plans:

| Type                       | `spec.type` value           | Scope                               | Limit unit      | CRD spec field                         |
| -------------------------- | --------------------------- | ----------------------------------- | --------------- | -------------------------------------- |
| Zonal Reserved Instance    | `reserved-instance`         | Specific instance type + AZ         | Node count      | `spec.ri` (with `availabilityZone`)    |
| Regional Reserved Instance | `reserved-instance`         | Instance family, region-wide        | CPU + memory    | `spec.ri` (without `availabilityZone`) |
| Compute Savings Plan       | `compute-savings-plan`      | Region-wide (all instance families) | $/hr commitment | `spec.computeSP`                       |
| EC2 Instance Savings Plan  | `ec2-instance-savings-plan` | Single instance family              | $/hr commitment | `spec.ec2InstanceSP`                   |

## How monetary commitments become CPU limits

For Compute Savings Plans and EC2 Instance Savings Plans, Keeper converts the hourly dollar commitment into CPU millicores using the per-vCPU price:

1. The total available CPU (in millicores) = `floor(commitmentAmount / pricePerCPU × 1000)` — the result is rounded down to a whole millicore
2. This total is distributed across matched NodePools proportionally to their CPU usage, but every matched pool is guaranteed at least 1000 millicores (1 vCPU) so it can still accept workloads. A commitment too small to grant 1 vCPU to any pool is reported as `InsufficientCommitment`
3. Each pool's CPU allocation is set as a `ResourceCPU` limit on the shadow NodePool

For EC2 Instance Savings Plans, the `pricePerCPU` is embedded in the CRD spec by Atlas. For Compute Savings Plans, Atlas embeds per-family pricing in the CRD, and Keeper selects the most expensive valid `pricePerCPU` among the instance families supported by each NodePool, ensuring the allocation is conservative.

# NodePool filtering

Not all NodePools need commitment coverage. Keeper supports two filtering modes controlled by Helm configuration:

## Opt-out mode (default)

All NodePools are included unless explicitly excluded. To exclude a NodePool:

```yaml
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: my-nodepool
  annotations:
    keeper.atlas.example/commitment-nodepool-clone: "exclude"
```

## Opt-in mode

Only NodePools explicitly annotated as included are considered:

```yaml
apiVersion: karpenter.sh/v1
kind: NodePool
metadata:
  name: my-nodepool
  annotations:
    keeper.atlas.example/commitment-nodepool-clone: "include"
```

# Shadow NodePool behavior

Shadow NodePools created by Keeper have the following characteristics:

| Property               | Value                                           | Reason                                                                         |
| ---------------------- | ----------------------------------------------- | ------------------------------------------------------------------------------ |
| Name format            | `zzz-atlas-budget-{budgetName}-{sourceNPName}` | Sorts last alphabetically; Karpenter's tiebreaker prefers it                   |
| Compute SP name format | `zzz-atlas-sp-{sourceNPName}`                  | All Compute Savings Plans aggregate into one pool per source NodePool          |
| Name length limit      | 63 bytes                                        | Karpenter uses the NodePool name as a label value on NodeClaims                |
| Weight                 | `100`                                           | Matches the maximum weight Karpenter allows; `zzz-` prefix wins the tiebreaker |
| Capacity type          | On-demand only                                  | Shadow pools replace spot with committed on-demand                             |
| Taints                 | Inherited from source NodePool                  | Workloads scheduling onto shadow pools must tolerate the same taints           |
| NodeClass              | Inherited from source NodePool                  | Same EC2NodeClass as the source                                                |
| Disruption             | Inherited from source NodePool                  | Consolidation policy and budgets copied from source                            |
| ConsolidateAfter       | `5m` default if source doesn't specify          | Prevents "Never" (zero-value) from disabling consolidation                     |

## Labels on shadow NodePools

| Label                                | Example value       | Purpose                                                                 |
| ------------------------------------ | ------------------- | ----------------------------------------------------------------------- |
| `keeper.atlas.example/managed-by`      | `commitmentbudgets` | Identifies the NodePool as managed by the commitment budgets controller |
| `keeper.atlas.example/budget-id`       | `ri-zonal`          | Links the NodePool to its source CommitmentBudget                       |
| `keeper.atlas.example/source-nodepool` | `my-app-pool`       | Links to the source NodePool this pool was derived from                 |
| `keeper.atlas.example/budget-type`     | `reserved-instance` | The commitment type that created this pool                              |

# Enable or disable the feature

Commitment support is controlled at two levels:

* **Helm value** (`atlas.commitmentBudgets.enabled`): set at install or upgrade time. To disable the controller, set it to `false` and redeploy — Keeper then performs a one-time cleanup of any existing shadow NodePools and stays inert. See [Prerequisites](#prerequisites) for the Helm value.
* **Cluster feature flag**: toggled dynamically by Atlas through the cluster's feature configuration, without restarting or redeploying Keeper. When it is turned off, Keeper drains and deletes existing shadow NodePools, then goes idle. Contact Atlas support to have this flag toggled on a cluster.

# Inspecting commitments on the cluster

Since Atlas syncs commitments automatically, the resulting `CommitmentBudget` resources can be inspected with `kubectl`:

```shell
kubectl get commitmentbudgets
```

```
NAME          TYPE                   INSTANCE TYPE   RI COUNT   RI FAMILY   COMMITMENT   MATCHED NODEPOOLS   COMPATIBLE   AGE
ri-zonal      reserved-instance      m5.xlarge       3                                   2                   Compatible   5m
ri-regional   reserved-instance                                 m5                       2                   Compatible   5m
sp-compute    compute-savings-plan                                          1.50         3                   Compatible   5m
```

Each `CommitmentBudget` reports its compatibility state through the `CompatibleNodePools` condition and the `matchedNodePools` status field. Use the condition reason to troubleshoot commitments that don't produce shadow NodePools:

| Condition status | Reason                   | Meaning                                                                             |
| ---------------- | ------------------------ | ----------------------------------------------------------------------------------- |
| `True`           | `Compatible`             | Matched N regular NodePool(s); shadow NodePools created                             |
| `False`          | `Incompatible`           | No regular NodePool matches this budget's instance type/family constraints          |
| `False`          | `InsufficientCommitment` | Commitment amount is too small to allocate any CPU above the minimum vCPU threshold |
| `False`          | `InvalidPricing`         | `commitmentAmount` or `pricePerCPU` is missing, zero, negative, or non-numeric      |
| `False`          | `NoCustomerNodePools`    | No regular NodePools exist in the cluster                                           |

# CommitmentBudget CRD reference

`CommitmentBudget` resources are synced to the cluster by Atlas. The reference below documents the CRD structure for observability and troubleshooting.

| Field                                 | Type         | Description                                                                       |
| ------------------------------------- | ------------ | --------------------------------------------------------------------------------- |
| `spec.budgetID`                       | string       | Unique identifier for the budget; one budget may span multiple merged commitments |
| `spec.type`                           | enum         | One of `reserved-instance`, `compute-savings-plan`, `ec2-instance-savings-plan`   |
| `spec.ri.availabilityZone`            | string       | AZ for zonal RI; omitted for regional RI                                          |
| `spec.ri.instanceType`                | string       | Exact instance type for zonal RI (e.g. `m5.xlarge`)                               |
| `spec.ri.count`                       | integer      | Number of reserved nodes (zonal RI)                                               |
| `spec.ri.instanceFamily`              | string       | Instance family for regional RI (e.g. `m5`)                                       |
| `spec.ri.cpu`                         | Quantity     | Total CPU for the commitment (e.g. `"40"`)                                        |
| `spec.ri.memory`                      | Quantity     | Total memory (e.g. `"96Gi"`)                                                      |
| `spec.computeSP.offeringId`           | string       | AWS offering ID                                                                   |
| `spec.computeSP.commitmentAmount`     | string       | Hourly commitment in USD (e.g. `"1.50"`)                                          |
| `spec.computeSP.familyCPUPrices`      | map          | Per-family pricing: family → `pricePerCPU` + `instanceType`                       |
| `spec.ec2InstanceSP.instanceFamily`   | string       | Instance family locked by this SP                                                 |
| `spec.ec2InstanceSP.offeringId`       | string       | AWS offering ID                                                                   |
| `spec.ec2InstanceSP.commitmentAmount` | string       | Hourly commitment in USD                                                          |
| `spec.ec2InstanceSP.pricePerCPU`      | string       | Cheapest per-vCPU price for the family                                            |
| `status.matchedNodePools`             | integer      | Number of regular NodePools matched                                               |
| `status.conditions`                   | \[]Condition | `CompatibleNodePools` condition with status and reason                            |

**Resource details:** `CommitmentBudget` (short name: `cb`), cluster-scoped, API group `autoscaling.atlas.example/v1alpha`.

# Related resources

* [Keeper](karpenter-enterprise-keeper.md) — The in-cluster component that creates and manages shadow NodePools
* [Karpenter Enterprise suite features](karpenter-enterprise-features.md) — Overview of all Karpenter Enterprise capabilities
* [Commitments](commitments.md) — Importing and managing commitments in the Atlas platform