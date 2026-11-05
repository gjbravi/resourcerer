---
updatedAt: 2026-05-12T13:39:45.000Z
---

# egressd, network and VPC

### Does the egressd image on the default repository of the chart support multiple architectures? Is there a way to only collect data from amd64 nodes?

There's not currently a way to limit the collection to a specific architecture. The team has an [existing task](https://github.com/atlas/egressd/issues/25) to make egressd compile on Arm.

<br />

***

### I want to run egressd-exporter as standalone pod and run the collector only on particular nodes, will it work or not?

If your goal is to set the collector to the node with api-platform-read and api-platform-write pods to see the networking costs of these services.

The collector will collect metrics from nodes where it runs and then the exporter will pull it from existing collectors, so the scenario should work fine.

<br />

***

### Are egress and ingress calculations based on each individual node host or are they computed from nodes that have an ingress serving internet traffic?

Do these network costs reflect current-day traffic?

Are they related to our existing load balancers that expose these services, or are they solely about inter-node usage?

Currently, we provide data only on internal traffic and not public internet traffic.

The daemon runs on each node and collects data on the workload level (per pod).

Yes, the data includes the current day (with a delay of around one minute).

<br />

***

### Do you have any recommended tuning guidelines around which SKUs to use in AWS that have reasonable max network interface and ips per interface?

You can implement constraints within the node template. Including only those specific families with max network cards would be beneficial in addressing this kind of problem. Compute-optimized instances might be a good option in such cases.

Here's a reference from AWS: [Elastic network interfaces](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-eni.html)

<br />

***

### Are there any penalties for using multi-zone topology like latency, increased cost, or something similar? Do charges for zone egress traffic apply?

You will face an increased cost because, in that case, egressd will select the cheapest node within that zone rather than choosing the cheapest node overall.

If the workloads are communicative and spread across various zones, they will inevitably accumulate network expenses. It represents a delicate balance between expenditure and dependability.

egressd currently offers an egress cost segment that presents an overview of egress costs. In the upcoming update, we plan to introduce a more comprehensive report featuring finer details for your convenience.

<br />

***

### Is egressd supported on Windows nodes?

No, egressd is not supported on Windows nodes. It is only supported on Linux nodes