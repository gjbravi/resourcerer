---
updatedAt: 2026-10-09T11:52:46.000Z
---

# Autoscaling

# Introduction

The Atlas Autoscaler is a tool designed to scale Kubernetes clusters with cost efficiency as the primary objective. Its goal is to dynamically adjust the number of nodes by adding new right-sized nodes and removing underutilized nodes when:

* There are pods that are unschedulable due to insufficient resources in the cluster.
* There are empty nodes that have not been utilized for a certain period.

The Autoscaler works through the following process:

1. **Watching** your cluster state using the Atlas Kubernetes Agent, which synchronizes cluster state changes to Atlas.
2. **Calculating** the required extra cluster capacity and **detecting** empty nodes.
3. **Evaluating** the extra capacity and converting it into **optimal** node instance types.
4. **Provisioning** new optimal nodes while **removing** underutilized empty nodes.

<Image alt="Autoscaler high-level diagram" border={false} src="https://files.readme.io/837da56-autoscaler-high-level.png" title="Autoscaler high-level diagram" />

# Key Concepts

To fully understand the Autoscaler's functionality, it's important to be familiar with the following concepts:

* **Unschedulable pods**: Kubernetes pods in the `Pending` phase with their `PodScheduled` condition set to `False` and reason `Unschedulable`.

* **Pod constraints**: Kubernetes pods can be declared with node selectors, node affinity, and pod anti/affinity properties. These reduce the number of node types and topologies (i.e., the hardware and zonal placement) that could satisfy the pod's requirements.

* **Kubernetes scheduler**: The control plane process responsible for assigning pods to nodes.

* **Empty node**: A node without any running pods. The Autoscaler ignores exited, static, or `DaemonSet` pods when determining if a node is empty.

# Cluster State Monitoring

The Autoscaler relies on accurate cluster state information to optimize the cluster effectively. The Atlas Kubernetes Agent collects all cluster state changes and sends them to the Autoscaler. It uses [Kubernetes informers](https://pkg.go.dev/k8s.io/client-go/informers) to watch Kubernetes resources and synchronizes all changes to Atlas at 15-second intervals.

The Kubernetes Agent monitors the following resources to enable correct autoscaling decisions:

* `Node`
* `Pod`
* `PersistentVolume`
* `PersistentVolumeClaim`
* `ReplicationController`
* `Service`
* `Deployment`
* `ReplicaSet`
* `DaemonSet`
* `StatefulSet`
* `StorageClasse`
* `Job`
* `CSINode`
* `HorizontalPodAutoscaler`

## DaemonSet Resource Calculation

When making autoscaling decisions, the Autoscaler examines both the DaemonSet specifications and the actual running DaemonSet pods in the cluster. This ensures accurate capacity planning even when tools like the Workload Autoscaler modify resource requests.

This prevents scaling loops where nodes are continuously added and removed without successfully scheduling workloads, especially in scenarios where actual DaemonSet resource usage differs from specifications.

# Cluster Upscaling

The Autoscaler adds new nodes when it detects unschedulable pods. Pods become unschedulable when the Kubernetes scheduler fails to place them on any existing node. This can occur for various reasons, such as insufficient resources on existing nodes or pods not matching node constraints.

### Bin-packing Algorithm

To determine the required extra capacity, the Autoscaler employs bin-packing algorithms to find the optimal pod groupings. This process takes into account that some pods cannot be placed on the same node due to their declared constraints.

Bin-packing is an [NP-hard problem](https://en.wikipedia.org/wiki/NP-hardness), meaning its complexity increases exponentially with the number of items to pack. As the number of unschedulable pods grows, the Autoscaler implements various approximation strategies to maintain consistent algorithm performance.

<Image alt="Bin-packing algorithm visualization" border={false} src="https://files.readme.io/69e09fc-tarjans-algorithm-animation.gif" title="Bin-packing algorithm visualization" />

## Topology spread constraints and node provisioning

Topology spread constraints run through two components that look at the cluster from two different angles:

* The Kubernetes scheduler enforces spread on the nodes that exist right now.
* The autoscaler decides where new nodes may be created for pending pods.

A zone with no nodes the pod can run on is invisible to the scheduler's spread check, but it can still be a valid target for a new node.

| Term               | Meaning                                                                                                                                                   |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Spread domain      | A zone that holds at least one node eligible for the pod. Zones without eligible nodes do not form domains and are not counted.                           |
| Eligible node      | A node the pod could run on, considering the pod's `nodeSelector` or required `nodeAffinity` and, when `nodeTaintsPolicy: Honor` is set, its tolerations. |
| Cluster zones      | All availability zones configured on the cluster, whether or not nodes exist in them yet.                                                                 |
| Provisioning zones | The zones where the autoscaler may create a node for the pod, after the Node Template and Node Configuration restrictions are applied.                    |
| Bound zones        | The zones of the pod's persistent volume claims. A pod stays with its data.                                                                               |

The zone-counting rules, the autoscaler's zone-picking order, and a worked example are in the [zone spread reference](autoscaler-reference-node-labels-and-taints.md#zone-spread-and-provisioning). To place pods with these constraints, see [Pod placement](pod-placement.md#highly-available-pod-scheduling). If upscaling fails, see [FailedUpscaling: topology spread blocks node creation](autoscaler-troubleshooting.md#failedupscaling-topology-spread-blocks-node-creation).

## Instance Type Selection

After bin-packing pod groups, the Autoscaler converts them into appropriate instance types. The goal is to find the most cost-effective instance type that meets all constraints. This process involves:

1. Filtering the list of all available instance types to identify viable options.
2. Sorting the viable instance types by price.
3. Selecting the cheapest option that satisfies all requirements.

This approach consistently results in the most optimal node selection.

Atlas regularly updates instance type pricing and availability information by querying cloud service provider APIs. This eliminates the need for manual pricing tracking and analysis of viable instance-type offerings.

When pods have no specific constraints, the Autoscaler can choose from a wide range of instance types. However, pods often require defining constraints for high availability or specific hardware requirements. For more information on controlling pod placement, refer to the [Pod placement](pod-placement.md) section.

#### Spot Instances

Atlas also fully supports Spot Instance scaling. To learn how to leverage Spot Instances for cost savings, visit the [Spot Instances](spot.md) section.

# Cluster Downscaling

The Autoscaler removes nodes that have remained unutilized for a specified period. This process aims to reduce costs and minimize waste. It's essential to note that the Autoscaler only removes completely empty nodes.

Nodes typically become empty when all running pods on the node have been deleted. This can occur due to:

* `ReplicaSet` scale-down operations
* `Job` completions

To further optimize cluster resources, Atlas provides an additional component called the [Evictor](evictor.md). The Evictor helps consolidate running pods into fewer nodes through intelligent bin-packing strategies.

By combining upscaling and downscaling processes, the Atlas Autoscaler ensures your Kubernetes cluster maintains optimal resource utilization and cost efficiency.