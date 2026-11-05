---
updatedAt: 2026-10-07T08:55:53.000Z
---

# Feature reference

Karpenter Enterprise suite brings Atlas's optimization capabilities to clusters running open-source Karpenter. This page provides an overview of available features and how they integrate with your existing Karpenter setup.

For a conceptual introduction to Karpenter Enterprise suite, see [Karpenter Enterprise suite overview](karpenter-enterprise.md).

# Feature availability

The following features are available for Karpenter-managed clusters:

| Feature                                                   | Description                                                                                        | Karpenter integration                     |
| --------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ----------------------------------------- |
| [Continuous rebalancing](#continuous-rebalancing)         | Ongoing workload consolidation with container live migration capabilities                          | Replaces Karpenter's native consolidation |
| [Rebalancer](#rebalancer)                                 | Cluster-wide cost optimization through Node selection, replacement, and capacity failure avoidance | Coordinates with Karpenter provisioning   |
| [Spot intelligence](#spot-intelligence)                   | Interruption prediction and observed interruption reporting                                        | Enhances Karpenter's Spot handling        |
| [Capacity failure detection](#capacity-failure-detection) | Avoids recently failing instance type, zone, and capacity type combinations                        | Improves capacity intelligence            |
| [Workload Autoscaler](#workload-autoscaler)               | Continuous workload rightsizing                                                                    | Feeds optimized requests to Karpenter     |
| [Pod mutations](#pod-mutations)                           | Automated Pod spec adjustments                                                                     | Simplifies workload configuration         |
| [Cost reporting](#cost-reporting)                         | Savings analysis and cost monitoring                                                               | Read-only analysis of Karpenter clusters  |
| [Commitments](#commitments)                               | Uses AWS Reserved Instances and Savings Plans through weighted shadow NodePools                    | Steers Karpenter provisioning             |

# How Atlas features work with Karpenter

Atlas features are designed to *extend* Karpenter rather than *replace* it. The integration follows these principles:

**Karpenter remains the provisioner**<br />Node creation and deletion continue to flow through Karpenter. Atlas influences decisions by modifying Karpenter CRDs and providing optimization signals, but Karpenter executes the actual infrastructure changes.

**CRD-native configuration**<br />Where possible, Atlas stores configuration in Kubernetes-native formats. Your existing [NodePools](https://karpenter.sh/docs/concepts/nodepools/) and [EC2NodeClasses](https://karpenter.sh/docs/concepts/nodeclasses/) remain the source of truth for provisioning constraints.

**Incremental enablement**<br />Each feature can be enabled independently. You can start with cost reporting only, then gradually enable optimization features as you build confidence.

# Feature details

## Continuous rebalancing

Keeper's [Continuous Rebalancing](karpenter-enterprise-continuous-rebalancing.md) monitors your cluster on a recurring cycle and consolidates underutilized nodes. When enabled, it takes over consolidation from Karpenter entirely — Karpenter continues to handle drift detection, but cost-driven consolidation is managed exclusively by Keeper.

**What it adds to Karpenter:**

* Coordination with [Workload Autoscaler](workload-autoscaling-overview.md) to consolidate Pods based on actual resource usage, including pending rightsizing recommendations not yet applied
* [Container Live Migration](clm-overview.md) for eligible workloads, preserving Pod state and TCP connections (graceful fallback to traditional eviction included)
* Configurable modes — `delete-empty`, `drain-only`, and `full` — for progressive or aggressive consolidation
* Savings thresholds that ensure rebalancing only runs when projected gains are worthwhile
* Fine-grained per-Pod and per-Node eviction policies via eviction config

### Protecting Nodes from consolidation

To exclude specific Nodes from Continuous Rebalancing, apply the `autoscaling.atlas.example/removal-disabled` label:

```bash
kubectl label node <node-name> autoscaling.atlas.example/removal-disabled=true
```

In the default opt-out mode, you can exclude an entire NodePool from all rebalancing (continuous, scheduled, and drift replacement) by annotating it with `keeper.atlas.example/consolidation: "exclude"`. The top-level `nodePoolFilterMode` Helm value controls how Keeper applies the annotation.

For full configuration details, see [Continuous rebalancing](karpenter-enterprise-continuous-rebalancing.md).

## Rebalancer

The [Rebalancer](rebalancing.md) optimizes your entire cluster by identifying Nodes that could be replaced with more cost-effective alternatives for your workloads.

**What it adds to Karpenter:**

* Cross-NodePool optimization that Karpenter doesn't perform natively
* Awareness of Reserved Instances and Savings Plans
* Coordinated replacements that maintain workload stability
* Integration with [Workload Autoscaler](workload-autoscaling-overview.md) to rebalance based on optimized resource requirements
* Avoidance of recently failed instance type, zone, and capacity type combinations observed by Keeper or Atlas

**How it differs from standard Atlas:**

| Aspect               | With Karpenter                                              | Standard Atlas                                |
| -------------------- | ----------------------------------------------------------- | ----------------------------------------------- |
| Node replacement     | Rebalancer cordons Nodes; Karpenter provisions replacements | Atlas handles both cordoning and provisioning |
| Instance selection   | Influences Karpenter via CRD modifications                  | Atlas selects instances directly              |
| Commitment awareness | No native commitments integration                           | Native commitments integration                  |
| Drain controls       | Limited by Karpenter's drain behavior                       | Full control over drain timing and aggression   |

### Rebalancing scope

Karpenter Enterprise suite only rebalances nodes that were created and are managed by Karpenter. The following nodes are excluded from rebalancing:

* Legacy nodes created before Karpenter was installed
* EKS-managed node group nodes
* Nodes created by other provisioners (Cluster Autoscaler, etc.)

This limitation exists by design: Karpenter Enterprise suite operates through Karpenter's CRDs and permissions, and does not have cloud-side control over non-Karpenter nodes. If you need to optimize non-Karpenter nodes, consider migrating them to Karpenter management or using Atlas's standard Autoscaler.

Within Karpenter-managed nodes, you can restrict which NodePools participate in rebalancing with the `keeper.atlas.example/consolidation` annotation. See [Node exclusions](karpenter-enterprise-continuous-rebalancing.md#node-exclusions).

## Spot intelligence

Atlas improves Karpenter's Spot Instance handling with predictive capabilities and reliability improvements.

**What it adds to Karpenter:**

* **Interruption prediction** — Identifies at-risk Nodes before AWS announces interruptions
* **Observed interruptions** — Reports observed Spot interruptions so Atlas can improve capacity intelligence
* **Spot reliability model** (coming in future releases) — Steers toward historically stable Spot pools
* **Spot fallback recovery** (coming in future releases) — Automatically returns to Spot when capacity becomes available again

### How interruption prediction works

Atlas's Spot interruption prediction provides up to 30 minutes of advance warning before interruptions occur—significantly extending AWS's standard 2-minute notice. The prediction mechanism differs based on whether Container Live Migration is enabled:

| CLM Status      | Prediction Response                                                                                                                                                                        |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Without CLM** | Keeper signals Karpenter's interruption queue, triggering Karpenter's standard node replacement workflow with extended lead time                                                       |
| **With CLM**    | Keeper uses Continuous Rebalancing to consolidate the at-risk node, using Container Live Migration to move workloads to stable nodes with zero downtime before the interruption occurs |

In both cases, the extended prediction window provides significantly more time for graceful workload migration compared to waiting for AWS's native interruption signal.

**How it differs from standard Atlas:**

| Aspect              | With Karpenter                                   | Standard Atlas               |
| ------------------- | ------------------------------------------------ | ------------------------------ |
| Pool selection      | Influences Karpenter's instance type priorities  | Atlas selects pools directly |
| Fallback handling   | Monitors Karpenter's fallback Nodes for recovery | Native fallback and recovery   |
| Prediction response | Signals Karpenter to replace at-risk Nodes       | Direct Node replacement        |

For Spot handling documentation, see [Spot Instances](spot.md) and [Spot Handler](spot-handler.md).

## Capacity failure detection

Keeper records an `UnavailableOffering` when it learns that a specific instance type, zone ID, and capacity type combination should be avoided temporarily. This can happen after an insufficient capacity error during NodeClaim creation, after an observed Spot interruption, or when Spot interruption prediction marks a Spot offering as risky.

Keeper uses the record locally when it builds new plans. It reports only observed insufficient capacity errors (ICE) and observed Spot interruptions to Atlas, enabling Atlas to improve capacity intelligence for Karpenter-backed clusters.

Prediction-only records are not treated as observed interruptions. For the CRD reference and troubleshooting commands, see [Keeper](karpenter-enterprise-keeper.md#capacity-failure-observations).

## Workload Autoscaler

[Workload Autoscaler](workload-autoscaling-overview.md) continuously rightsizes workloads based on actual resource usage.

**What it adds to Karpenter:**

* Automatic adjustment of CPU and memory requests to match actual usage
* Tighter bin-packing as rightsized workloads require less capacity
* Integration with Continuous Rebalancing and Rebalancer for coordinated optimization

**How it differs from standard Atlas:**

| Aspect           | With Karpenter                                                        | Standard Atlas                                              |
| ---------------- | --------------------------------------------------------------------- | ------------------------------------------------------------- |
| Request updates  | Workload Autoscaler updates requests; Karpenter sees new requirements | Same behavior                                                 |
| Node impact      | Karpenter may consolidate as requests decrease                        | Atlas coordinates this directly with Continuous Rebalancing |
| Scaling policies | Applied identically                                                   | Applied identically                                           |

Workload Autoscaler behavior is largely identical whether you're using Karpenter Enterprise suite or Atlas's Autoscaler—it operates at the workload level independently of Node provisioning.

For Workload Autoscaler documentation, see [Workload Autoscaling](workload-autoscaling-overview.md).

## Pod mutations

[Pod mutations](pod-mutations-quickstart.md) automate Pod spec adjustments to simplify workload configuration and reduce manual efforts by teams.

**What it adds to Karpenter:**

* Automatic application of labels, tolerations, and NodeSelectors
* Simplified onboarding without modifying Deployment manifests
* Consistent Pod configuration across workloads

**How it differs from standard Atlas:**

Pod mutations work identically with Karpenter and standard Atlas. The mutations apply to Pod specs before creation, independent of which autoscaler provisions Nodes.

For Pod mutations documentation, see [Pod mutations](pod-mutations-quickstart.md).

## Cost reporting

The savings report and other cost monitoring capabilities provide visibility into your cluster's optimization potential without making any changes.

**What it provides:**

* Current vs. optimized cost comparison
* Node utilization and bin-packing analysis
* Commitment utilization tracking
* Spot adoption opportunities
* Workload rightsizing recommendations

**How it differs from standard Atlas:**

Cost reporting works identically for Karpenter clusters. The analysis examines your current state and models what Atlas optimization could achieve.

For general cost monitoring, see [Cost Monitoring](cost-management.md).

## Commitments

Atlas steers Karpenter toward AWS Reserved Instances and Savings Plans by creating weighted shadow NodePools that Karpenter prefers over regular NodePools. Commitments are imported and assigned through the Atlas [Commitments](commitments.md) page, and Atlas automatically syncs them to the cluster as `CommitmentBudget` resources. Keeper then creates, rebalances, and cleans up the shadow NodePools.

**What it adds to Karpenter:**

* Commitment-aware instance selection without Karpenter having native commitment awareness
* Automatic limit rebalancing across matched NodePools based on observed usage
* Support for zonal RIs, regional RIs, Compute Savings Plans, and EC2 Instance Savings Plans

For full configuration details, see [Commitments in Karpenter Enterprise](karpenter-enterprise-commitments.md).

# Labels reference

Karpenter Enterprise suite uses specific labels to coordinate optimization activities with Karpenter. Understanding these labels helps with troubleshooting and protecting critical nodes.

## Node labels

| Label                                  | Purpose                                                      | Applied by                                              |
| -------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------- |
| `karpenter.sh/do-not-disrupt`          | Prevents Karpenter from consolidating or disrupting the node | Keeper (automatically managed during consolidation) |
| `autoscaling.atlas.example/removal-disabled` | Prevents Continuous Rebalancing from selecting the node      | Customer (manual protection)                            |

## Workload labels

| Label                                  | Purpose                                                         | Applied by                             |
| -------------------------------------- | --------------------------------------------------------------- | -------------------------------------- |
| `live.atlas.example/migration-enabled=true`  | Indicates the workload is eligible for Container Live Migration | Live Controller (automatic assessment) |
| `live.atlas.example/migration-enabled=false` | Indicates the workload cannot be live-migrated                  | Live Controller (automatic assessment) |

## Protecting nodes from optimization

**To exclude a node from Continuous Rebalancing:**

```bash
kubectl label node <node-name> autoscaling.atlas.example/removal-disabled=true
```

**To also prevent Karpenter from disrupting the node (including drift-triggered replacements):**

```bash
kubectl label node <node-name> karpenter.sh/do-not-disrupt=true
```

# Features not available with Karpenter

Some Atlas capabilities require tighter integration with Node scheduling than the Karpenter-layered approach allows:

| Feature                                          | Why it's not available                                              | Alternative                                                                                                                     |
| ------------------------------------------------ | ------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| Pod Pinner                                       | Requires Atlas Autoscaler's scheduling integration                | Use Karpenter's native Pod affinity                                                                                             |
| Cluster hibernation                              | Requires direct control over Node lifecycle                         | Use Karpenter's NodePool weight=0 for manual scaling                                                                            |
| Commitment-aware instance selection (Rebalancer) | Rebalancer cannot directly influence Karpenter's instance selection | Use [Commitments](karpenter-enterprise-commitments.md) to steer Karpenter toward committed instance types via shadow NodePools |

Additionally, the following Rebalancer capabilities are not available when using Karpenter (see [Rebalancing scope](#rebalancing-scope) for details):

* Aggressive mode
* Graceful eviction controls
* Paused drain configuration
* Progress bar in UI

To benefit from these capabilities, consider migrating to Atlas Autoscaler.

# Related resources

* [Karpenter Enterprise suite overview](karpenter-enterprise.md) — Conceptual introduction
* [Getting started](karpenter-enterprise-getting-started.md) — Connect your cluster

<br />