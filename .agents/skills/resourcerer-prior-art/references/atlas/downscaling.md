---
updatedAt: 2026-05-12T14:01:40.000Z
---

# Downscaling

The Atlas Autoscaler removes nodes when it detects that they have not been utilized for a certain period of time. Removing empty nodes saves costs and reduces waste.

The Autoscaler only removes **empty** nodes. Nodes end up empty because all pods running on the node have been deleted, possibly due to `ReplicaSet` scale down or `Job` completion.

## Downscaling options

You have two levels of downscaling available at your disposal:

* **Node deletion policy** - this policy just removes nodes that are empty and no longer running in any capacity. For example, if a job you're running goes past its run, a node may become empty, and Atlas will automatically remove it to avoid waste.
* **[Evictor](evictor.md)** - it continuously compacts pods into fewer nodes, creating empty nodes that can be removed following the Node deletion policy (if you choose to enable it). Evictor actively bin packs your cluster state and moves pods around to achieve higher node utilization. Nodes that have been freed up are removed in accordance with the **Node deletion policy**.

![](https://files.readme.io/7f7b501-image.png)

## Bin-packing flow

Atlas implemented the Evictor component to solve the [bin-packing problem](https://en.wikipedia.org/wiki/Bin_packing_problem).

1. Evictor continuously scans your cluster every 60 seconds (default cycle time).
2. It identifies nodes that have been active for more than 5 minutes (default `NODE_GRACE_PERIOD_MINUTES`) as potential candidates for eviction.
3. Evictor assesses if all current workloads on a candidate node can be rescheduled on existing capacity in the cluster, considering node selectors, affinity rules, and other constraints.
4. If a suitable node is found, Evictor cordons the node, drains it, and moves the workloads to other nodes, optimizing resource utilization.
5. Once a node becomes empty after draining, Evictor immediately triggers node deletion through the Node deletion policy (which should be enabled). The configured TTL does not apply to nodes evicted through bin-packing.

> 📘 Note: Node TTL and Evictor
>
> Node TTL settings apply only to nodes that become empty naturally (through job completion or replica scale-down). When Evictor actively drains a node during bin-packing, the node is deleted immediately after draining completes, regardless of the configured TTL value.

> 📘 Note
>
> Evictor does not use percentage-based utilization criteria. It considers overall cluster capacity and workload distribution when making eviction decisions.

<Image align="center" alt="Evictor explained with train senario" border={false} caption="Evictor explained on a train scenario." src="https://files.readme.io/24d1889-image.png" />