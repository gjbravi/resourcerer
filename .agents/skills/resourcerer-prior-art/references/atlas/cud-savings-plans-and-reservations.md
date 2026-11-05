---
updatedAt: 2026-07-15T13:22:15.000Z
---

# CUD, Savings Plans, and reservations

### Are Committed Use Discounts (CUDs) considered when calculating saving costs on Atlas console?

Currently, the calculation doesn't include CUD prices. However, the Atlas autoscaler can be configured to first pick reserved families to utilize CUDs.

<br />

***

### How much can we save on the cloud with Atlas if we're using CUDs?

The primary role of the read-only script is to estimate the potential savings realized with Atlas.

The read-only script won't make any changes to your cluster configuration without explicit permission. It allows us to read basic K8s primitives to give you an estimation of how much savings is possible in a specific cluster.  You can read more about it [here](about-the-read-only-agent.md#what-data-can-the-atlas-read-only-agent-access).

<br />

***

### Does the Atlas autoscaler support Reserved Instances?

Yes, Reserved Instances are supported for autoscaling.

<br />

***

### Do I need to complete all the fields in the CSV Inventory Template?

Zone fields are optional since reservations can be zone-based or region-based. However, note that other providers may include different formats and fields that could be cloud-specific.

<br />

***

### Do Reservations support regions that don't have availability zones?

Yes, you can get region-based reserved capacity.

<br />

***

### When considering autoscaling, is the duration of the contract's start date and end date taken into account?

Only currently active reservations are taken into account by the Atlas autoscaler.

<br />

***

### During the generation of a rebalancing plan, I noticed that Atlas attempted to replace a node with an instance type not specified in the inventory. Was this an expected behavior?

Using instances that aren't part of the reservation inventory might happen if:

* reservations are fully utilized,
* the cluster region/zone is not the same,
* the workload requires spot instances,
* the workload requires some special features that aren't available in reserved instances: storage-optimized, compute-optimized, GPU, etc.

<br />

***

### What does the 'Price' column represent in the RI inventory template?

The price column displays the initial purchase price for a particular family of reserved instances. Since Reserved Instances (RIs) are prepaid and their cost remains constant, whether you use the virtual machines (VMs) or not, their price is effectively 0 in this instance of time.

As a result, Autoscaler doesn't consider its price when making decisions. Instead, it prioritizes RIs based on other criteria, with the price you input being treated as text without influencing its behavior.

<br />

***

### Does the Atlas CUD balancer consider instance(s) that aren't managed by Atlas (GKE-provisioned ones)?

Yes, Atlas takes into account all instances from clusters onboarded to Atlas (readonly and no-readonly mode). It doesn't matter if its Atlas-managed or not.