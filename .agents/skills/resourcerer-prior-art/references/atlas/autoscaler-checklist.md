---
updatedAt: 2026-09-22T11:21:49.000Z
---

# Node Autoscaler preparation checklist

The recommended way to prepare a newly-connected cluster for Node Autoscaling is to run `atlasctl cluster import-config`. It discovers your existing node group configuration from your cloud provider and creates matching [node configurations](node-configuration.md) and [node templates](node-templates.md) automatically, so Atlas provisions nodes that match what you already run. This is the fastest, least error-prone way to get a working setup. See [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md) for the full guide.

```shell
atlasctl cluster import-config
```

Automated import handles node-level configuration. It doesn't, and can't, change how your workloads are deployed, so review the requirements below regardless of which path you take.

## Before enabling Node Autoscaling in production

Atlas's Node Autoscaler and Evictor work by provisioning and removing nodes, and by evicting and rescheduling pods. Workloads that don't follow Kubernetes best practices can experience unexpected downtime when automation is active.

Before enabling Node Autoscaling on a production cluster, verify that your workloads meet the following requirements:

| Requirement                                                           | Why it matters                                                                                                                                                                                                                          |
| --------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **PodDisruptionBudgets (PDBs) configured** for all critical workloads | Without a PDB, Evictor may evict all pods of a deployment simultaneously. A PDB with `minAvailable: 1` is the minimum recommended setting. See [PodDisruptionBudgets](https://kubernetes.io/docs/tasks/run-application/configure-pdb/). |
| **`imagePullPolicy` is not `Always`** unless required                 | `Always` introduces delays when pods are rescheduled to new nodes, and will fail entirely if the registry is temporarily unreachable.                                                                                                   |
| **Startup probes configured** for services with slow initialization   | JVM-based services and other slow-starting workloads should have startup probes to prevent traffic routing before the app is ready. This is a best practice rather than a hard requirement.                                             |
| **Stateful workloads reviewed individually**                          | Databases, caches, and message queues must have data replication in place and be able to tolerate pod restarts.                                                                                                                         |
| **Pod startup times are compatible with `nodeGracePeriodMinutes`**    | The default grace period before Evictor targets a new node is 5 minutes. Workloads with slow startup times may need this value increased.                                                                                               |

> 📘 Tip
>
> Enable and validate Node Autoscaling in a staging or development environment with representative workloads before rolling it out to production. See the [Evictor documentation](evictor.md) for more details on production readiness.

## Goals

* Upscale the cluster during peak hours.
* Binpack or downscale the cluster when excess capacity is no longer required.
* Use Spot Instances to reduce your infrastructure costs, but have the safety of on-demand instances when needed.

## General guidelines: manual node configuration

`atlasctl cluster import-config` sets up the policies below for you automatically, based on your cluster's existing configuration. The rest of this section explains what it configures and why, for reference. Use it if you're configuring node templates manually, or if you want to understand what automated import sets up on your behalf.

### Unscheduled pods policy

To upscale a cluster, Atlas needs to react to unschedulable pods. You can achieve this by turning on the Unscheduled Pods policy and configuring the [Default Node template](node-templates.md).

> 📘 What is an unschedulable pod?
>
> This term refers to a pod stuck in a pending state, meaning that it cannot be scheduled onto a node. Generally, this is because insufficient resources of one type or another prevent scheduling.

<Image align="center" alt="Unscheduled pods policy" border={false} caption="Unscheduled pods policy with Default Node template" src="https://files.readme.io/5ac09d6-Screenshot_2023-08-07_at_21.30.47.png" />

**Why?**

* Automatically adds the required capacity to the cluster.
* Enables **[Spot instances](spot.md)**, so Atlas can handle spot instances and their interruptions for you.
* Enables **[Spot Fallbacks](spot.md#spot-fallback)** to automatically switch back and forth to on-demand capacity when spot isn't available in your cloud environment.

### Node deletion policy / Evictor

Atlas can constantly binpack the cluster and remove any excess capacity. To achieve this, we recommend the following initial setup:

<Image align="center" alt="Node deletion policy recommended settings" border={false} caption="Node deletion policy recommended settings" src="https://files.readme.io/62a4686-image.png" />

**Why?**

* Ensures that empty nodes don't keep running in the cluster longer than the configured duration.
* Enables **[Evictor](evictor.md)** for higher node resource utilization and less waste. Evictor continuously simulates scenarios where it tries to eliminate underutilized nodes by checking whether the pods could be scheduled in the remaining capacity. Simulation respects [PDBs](https://kubernetes.io/docs/tasks/run-application/configure-pdb/) and all other Kubernetes restrictions. **Configure PDBs for all critical workloads before enabling Evictor.** Without a PDB, there's no guarantee that at least one replica remains available during bin-packing.

<Callout icon="🎯" theme="default">
  ### What is Evictor's aggressive mode?

  When Evictor runs in aggressive mode, it considers workloads with a single replica as potential targets for binpacking. This might cause some disruption in single-replica workloads.
</Callout>

## Validate the setup

After enabling Node Autoscaling, whether through automated import or manual configuration, run a simple upscale/downscale test to confirm the Node Autoscaler reacts as expected:

1. Deploy a test workload that requires Spot capacity:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/atlas/examples/main/evictor-demo-pods/test_pod_spot.yaml
   ```
   ```yaml YAML source
   # https://github.com/atlas/examples/blob/main/evictor-demo-pods/test_pod_spot.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: atlas-test-spot
     namespace: atlas-agent
     labels:
       app: atlas-test-spot
   spec:
     replicas: 10
     selector:
       matchLabels:
         app: atlas-test-spot
     template:
       metadata:
         labels:
           app: atlas-test-spot
       spec:
         tolerations:
           - key: scheduling.atlas.example/spot
             operator: Exists
         nodeSelector:
           scheduling.atlas.example/spot: "true"
         containers:
           - name: nginx
             image: nginx:latest
             ports:
               - containerPort: 80
             resources:
               requests:
                 cpu: 4
   ```

   This confirms Atlas has what it needs to upscale the cluster automatically.

2. Once capacity is added, check that your DaemonSet pod count matches the node count:

   ```shell
   # Get DaemonSets in all namespaces
   kubectl get ds -A

   # Get node count
   kubectl get nodes | grep -v NAME | wc -l
   ```

3. Confirm the deployed pod is `Running`, then scale it back down:

   ```shell
   kubectl scale deployment/atlas-test-spot --replicas=0
   ```

   Verify that Atlas removes the now-empty nodes within the configured time interval.

## Troubleshooting

### Partially managed by Atlas

<Image align="center" alt="Not fully managed by Atlas" border={false} caption="One node not managed by Atlas" src="https://files.readme.io/9d91e67-small-Screenshot_2023-04-28_at_10.47.56.png" width="30% " />

When you connect a cluster, it may not be immediately fully managed by Atlas: some workloads can still run on existing legacy node pools or Autoscaling groups outside Atlas's control.

* **Recommended:** run `atlasctl cluster import-config` to translate those legacy node pools into matching Atlas node configurations and node templates, bringing them under management instead of excluding them. See [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md).
* **Alternative:** add `autoscaling.atlas.example/removal-disabled="true"` on such node pools/Autoscaling groups so Atlas excludes those nodes from Evictor and Rebalancer instead of managing them.

## See also

<Cards columns={2}>
  <Card title="Import and apply cluster configuration with atlasctl" icon="fa-download" href="https://docs.atlas.example/docs/atlasctl-import-cluster-configuration">The recommended way to prepare a connected cluster for Node Autoscaling.</Card>
  <Card title="Evictor" icon="fa-broom" href="https://docs.atlas.example/docs/evictor">How Atlas removes underutilized nodes.</Card>
</Cards>