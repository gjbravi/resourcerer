---
updatedAt: 2026-05-12T14:12:37.000Z
---

# Managing DaemonSets with Atlas

Learn about the implications of making changes to DaemonSets and its effects on existing nodes.

# Background

A DaemonSet in Kubernetes ensures that a specific pod runs on all or selected nodes (using Node Selectors and Node Affinity) in a cluster. It's typically used for background tasks like logging, monitoring, or networking.

Generally, Atlas aims to bin-pack pods as tightly as possible into as few nodes as possible, which can present challenges when increasing DaemonSet requests or adding new DaemonSets.

# The problem

When you change a DaemonSet's container requests, the DaemonSet controller starts a rollout. Here's an example flow:

1. **Node identified**: The DaemonSet controller identifies a node that needs an updated pod.
2. **Delete old pod**: The existing pod on that node is deleted.
3. **Create a new pod**: With the updated container requests, a new pod is created on the same node (using node affinity to ensure it is scheduled on the correct node).
4. **Repeat for each node**: This process is repeated sequentially for all nodes where the DaemonSet is running.

Imagine your nodes have 99% CPU or memory utilization. There's a high chance that when you increase the requests, the new DaemonSet pods won't fit and will stay in the Pending state. If your DaemonSets are providing critical functionality, you might experience downtime.

The same applies to new DaemonSets. New pods might not fit into existing nodes if their resource utilization is high.

# Prerequisites

* Basic understanding of [Kubernetes DaemonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/) and resource management
* Familiarity with Atlas's [rebalancing](https://docs.atlas.example/docs/how-rebalancing-works) feature
* Access to modify cluster resources and Atlas settings

# Solution 1: Rebalancing

One possible solution is to rebalance your cluster or just the nodes where the DaemonSets don't fit. Atlas considers DaemonSet requests and will create the right-sized nodes to accommodate the new or changed DaemonSet pods.

This solution is viable if you're dealing with a new DaemonSet, the DaemonSet isn't critical, and you can tolerate some pods being unavailable temporarily.

# Solution 2: Using priority classes

Another solution is a little more complex, but it suitable for situations where you can't afford your DaemonSet pods going down - adding a `system-cluster-critical` priority class to your DaemonSet. If the DaemonSet pods don't fit when recreated, the Scheduler will evict lower-priority class pods to make room for the DaemonSet.

First, you have to define a `ResourceQuota` that allows your pods to utilize a priority class:

```yaml yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: critical-daemonsets
  namespace: your-namespace
spec:
  scopeSelector:
    matchExpressions:
      - operator: In
        scopeName: PriorityClass
        values:
          - system-cluster-critical
```

Then, you can add it to your DaemonSet:

```yaml yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: critical-daemonset
  namespace: your-namespace
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: critical-daemonset
  template:
    metadata:
      labels:
        app.kubernetes.io/name: critical-daemonset
    spec:
      priorityClassName: system-cluster-critical
      containers:
        - image: nginx
          name: nginx
          resources:
            limits:
              cpu: 500m
              memory: 128Mi
            requests:
              cpu: 500m
              memory: 128Mi
```

# How Atlas Autoscaler handles DaemonSet resources

When making autoscaling decisions, Atlas's Autoscaler intelligently examines both:

* The DaemonSet specifications (what's defined in your YAML)
* The actual resource requests of running DaemonSet pods in the cluster

This approach provides accurate capacity planning even when tools like Atlas's Workload Autoscaler or third-party solutions modify resource requests. The Autoscaler specifically:

* Identifies the newest running DaemonSet pods in the cluster
* Compares their resource requests with what's defined in the DaemonSet specification
* Uses the higher value for capacity planning when creating new nodes
* Accounts for any Workload Autoscaler recommendations applied to DaemonSet pods

This prevents scaling loops where nodes are continuously added and removed without successfully scheduling workloads, especially when actual DaemonSet resource requests differ from specifications.

# Conclusion

Managing DaemonSet resources in a Atlas-optimized cluster requires considering the impact on Node utilization and overall cluster efficiency. Whether you choose to rebalance your cluster or use priority classes, monitoring the effects of these changes and adjusting your strategy as needed is crucial.

Whichever solution you choose, adding new DaemonSets or changing the resources of existing ones can lead to cluster inefficiencies. Rebalancing the cluster after such changes is always recommended to ensure that your nodes are right-sized.