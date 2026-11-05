---
updatedAt: 2026-07-14T12:55:42.000Z
---

# Evictor

### How aggressively will Evictor behave in a scenario where a microservice with multiple replicas needs to have one replica ready at all times?

Evictor will end up bin-packing a single replica as well. That is why the aggressive mode is recommended for more fault-tolerant systems.

You can annotate the workload using the following command to prevent Evictor from evicting a node running an annotated pod (this can also be applied to a node):

```shell
kubectl annotate pods <pod-name> autoscaling.atlas.example/removal-disabled="true"
```

Note that this solution has one limitation: it will lock other replicas from moving and prevent rebalancing.

<br />

***

### Is it possible we are doing something that's not releasing resources and forces us to perform rebalancing every few days?

Aggressive Evictor might work if you have a lot of single replicas as long as they aren't mission-critical and can be moved around.

<br />

***

### Does Atlas redeploy Evictor, overwriting parameters?

Policies in the API override the parameters, you can update this via the API.

The first step is to get the current policy configurations: [Gets policies configuration for the target cluster](https://docs.atlas.example/reference/policiesv2api_getclusterpolicies)

Then update the configuration accordingly via PATCH: [Update cluster policies](https://docs.atlas.example/reference/policiesv2api_updateclusterpolicies)

<br />

***

### Does ATLAS support HA mode for atlas-evictor Helm chart? Is leaderElection the flag?

All Atlas components can run in Active/Passive mode, and Evictor supports it.

<br />

***

### Does Evictor take into consideration pods under ReplicaSet only?

Evictor in non-aggressive mode will only consider ReplicaSets (usually Deployments), so elements like ReplicaSet controllers should be treated the same - they will get evicted if there are 2+ replicas.

<br />

***

### How does Evictor address parallel jobs in Kubernetes? Is there a possible interruption?

The latest version of Evictor in non-aggressive mode will not touch parallel jobs.

<br />

***

### Within scopedMode, what exact labels does Evictor follow?

ScopedMode means that Evictor will only consider evicting pods from nodes that were created by Atlas (`atlas.example/managedby=atlas` label on node).

<br />

***

### What are the limits of Evictor for calculating bin-packing? How many pods/nodes can Evictor handle?

Evictor has been shown to run fine in a 30,000 CPU cluster where 100 node batches are being removed every minute. You can specify batch size, and how many nodes Evictor can evict in a single batch. With a sleep of 10 seconds, one node in a batch is usually enough for sub 10,000 CPU clusters.

<br />

***

### What permissions does Evictor need?

Check out [this page](kubernetes-permissions.md#atlas-evictor-phase-2) to learn more about the permissions required by Evictor.

<br />

***

### Does Evictor equal rebalancing, or are these two different components of Atlas?

Rebalancing and Evictor are two different Atlas features.

You can find detailed information on how each of them works at the following links:

* [Evictor](evictor.md)
* [Rebalancing](https://docs.atlas.example/docs/how-rebalancing-works)

<br />

***

### We observed an incident where bin-packing is killing all the pods of a stateless ReplicaSet at once. This results  in service unavailability for a couple of minutes. Can we do something about it to ensure at least one pod always running?

We recommend using Pod Disruption Budgets to control the drain rate of critical applications, here's a helpful resource: [Specifying a disruption budget for your application](https://kubernetes.io/docs/tasks/run-application/configure-pdb/).

Setting minAvailable: 1 in a Pod Disruption Budget will ensure that at least one pod is always available.

<br />

***

### What is Evictor's default threshold for MEM and/or CPU it will look at to delete a node?

Evictor doesn't work on the basis of thresholds but runs a simulation to detect where pods would go if a node was deleted. If simulations result in a graceful (and non-impactful) migration of all pods to other nodes, Evictor will carry on with the eviction.

<br />

***

### Which version of the Atlas agent is needed for advanced configuration on Evictor?

The Managed Evictor Advanced Configuration feature only works if the atlas-agent version is at least v0.49.1. You can manually change the Evictor Advanced Configuration configmap with older versions. If you experience any issues with Evictor Advanced Configuration, please check the atlas-agent version first.

<br />

***

### Even when Evictor is not set to aggressive mode, will it ignore deployment configuration like maxSurge and maxUnavailable?

Yes, Evictor will ignore them as they are used by the Deployment level for rolling updates and Evictor works on pod scope. However, Evictor takes the pod distribution budget into account.

<br />

***

### Does Atlas evict a pod if it's in the init state? Is there any downtime?

Yes, Atlas does that.

Init containers are the containers that run until completion, and all the same rules apply to them.

Kubernetes will execute the preStop hooks for any running containers, if configured. Next, it will send the SIGTERM signal to all running containers, which all apps must listen to when running in K8s. Then it will wait for 30 seconds, by default, or the configured terminationGracePeriodSeconds time for all containers to exit. If the containers are still running after that time, it sends the SIGKILL signal, which stops the container abruptly.

If any of those mechanisms aren't respected by the app (e.g., it's not listening to SIGTERM), downtime is possible.

<br />

***

### What does "Eliminated" mean within the Evictor logs?

```
message:time="2023-12-20T08:50:59Z"  
level=debug msg="#569 node - ip-10-16-126-241.ec2.internal eliminated by PodsEvictable, reason: pod disruption budget" aggressive_mode=true level_int=5 scoped_mode=true
```

"Eliminated" means that the pod was spared from eviction to avoid breaking Pod Disruption Budgets (PDBs).

<br />

***

### Can the Evictor be set to wait a specific amount of time before shutting down a node?

The NODE\_GRACE\_PERIOD determines the time we wait before considering a new node for eviction. While the timeout for a force-delete can be set in the API call, there might not be a specific setting to forcefully terminate a node after a certain time.

All nodes currently have a roughly 20-minute drain timeout. If a pod doesn't drain gracefully within this timeframe, the node is returned to service. This means if you have a deployment set to sleep for 20 minutes before terminating, it may not be fully honored if it exceeds the autoscaler's set timeout.

<br />

***

### Why am I seeing high memory utilization on atlas nodes? What does this indicate?

This high utilization indicates the effective packing of pods onto the nodes, suggesting that the atlas is performing well in terms of packing. We are effectively managing resource allocation, and it's expected that if additional workloads arrive requiring more memory, the atlas will provision accordingly.

<br />

***

### How does Container Live Migration affect StatefulSet pods?

When Container Live Migration is enabled in your cluster, StatefulSet pods that meet Container Live Migration eligibility criteria are automatically considered for live migration without requiring the `autoscaling.atlas.example/disposable="true"` flag.

**Key behavioral differences**:

* **Without disposable flag**: If migration fails, the pod remains on the source node. Evictor will NOT fallback to traditional eviction, protecting the stateful workload from disruptive restarts.
* **With disposable flag**: The pod will always be evicted. With Container Live Migration, migration is attempted first, but the pod is evicted if migration fails. Without Container Live Migration, the pod is evicted directly.

This change ensures stateful workloads benefit from optimization opportunities through live migration while maintaining protection against disruptive evictions. See the [Evictor documentation](evictor.md#eviction-of-statefulset-pods) for more details.

<br />

***

### Why are my evictor pods restarting every few minutes even though they are not crashing?

If evictor pods are being replaced repeatedly — new ReplicaSets appearing, old pods terminated — but pod logs show a clean shutdown rather than a panic or OOMKill, the cause is likely CPVPA recalculations.

`atlas-evictor-cpvpa` adjusts evictor memory every \~5 minutes based on the current node count. Each recalculation that produces a new value modifies the Deployment spec, which Kubernetes treats as a rolling update trigger — terminating the current pod and creating a replacement.

To confirm: run `kubectl describe deployment atlas-evictor -n atlas-agent` and check the Events section for `ScalingReplicaSet` entries. If these appear on a \~5-minute cadence, CPVPA is the source.

Options:

* Reduce the `step` value in the `atlas-evictor-autoscaler` ConfigMap so CPVPA produces the same value across more poll cycles, reducing how often the Deployment spec changes.
* Temporarily scale CPVPA to 0 to freeze memory values while the cluster stabilizes: `kubectl scale deployment atlas-evictor-cpvpa -n atlas-agent --replicas=0`. Restore it once stable: `kubectl scale deployment atlas-evictor-cpvpa -n atlas-agent --replicas=1`.

See [Resource management for large clusters](evictor.md#resource-management-for-large-clusters) for configuration details and recommended values by cluster size.

<br />

***