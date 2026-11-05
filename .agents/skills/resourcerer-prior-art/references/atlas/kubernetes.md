---
updatedAt: 2026-07-08T12:32:41.000Z
---

# Kubernetes

### Why are pods getting this error and being evicted: The node was low on resource: memory. Threshold quantity: 750Mi, available: 590728Ki.

This error isn't related to Atlas and is a common issue with Kubernetes workload configuration.

Regarding the eviction; this is an expected behavior of the Kubernetes Scheduler when requests != limits for memory, and all of the containers defined for this pod have that setup. Learn more about [Kubernetes pod priority and preemption](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/).

It's a Kubernetes best practice to set requests = limits for memory and no limits for the CPU. The idea is to get the requested memory as close as usage to get better resilience. Limits are there to protect us from bursting.

<br />

***

### Can we safely remove CPU limits?

Our general recommendation is not to set CPU limits as it's a costly functionality and the burstable nature of workloads ensures better hardware utilization.

Add CPU limits only in case you have abusers like miners, stress-testers, etc. that may eat all CPU for long periods.

Note: Your workloads will lose Guaranteed QoS

<br />

***

### Can a single replica deployment wait for another replica to be up before terminating? In Atlas, can I ensure the app is healthy during a spot interruption without running multiple replicas continuously?

With a single replica deployment, it's not possible to automatically ensure that another replica is up and healthy before terminating the existing one.

The recommended solution is always to have at least two replicas running for high availability and fault tolerance to avoid service interruptions and downtime. However, there are measures you can take to mitigate downtime during spot interruptions or maintenance activities without increasing the replica count.

Suggested solutions:

1. **Implement readiness probes** - By implementing a readiness probe in your application's container specification, you can ensure that the replacement pod is ready before terminating the current one. The readiness probe signals to Kubernetes when a pod is ready to serve requests.

But to ensure high availability during spot interruptions, it's recommended to deploy multiple replicas, which is the best solution. Here's what you can do:

2. **Deploy multiple replicas** - Set up and configure your application to run multiple replicas concurrently. This way, the interruption of one replica does not impact the availability of your application, as other replicas can take over the workload.

<br />

***

### How can we manage OOM issues in our low-memory service during bursts, especially on nodes with insufficient memory?

To address this issue, it's common when there's a significant difference between requests and limits. Kubernetes scheduler and autoscaler consider requests, not limits. The recommendation is to avoid large deltas between them. A workaround is to create dummy pods with affinity to the problematic pod or set higher requests to prevent this scenario.