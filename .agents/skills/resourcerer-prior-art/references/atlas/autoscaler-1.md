---
updatedAt: 2026-09-25T08:58:21.000Z
---

# Autoscaler

### What is the impact of Atlas on availability?

The Atlas autoscaler continuously monitors the cluster and upscales or downscales it based on the real-time demand.

<br />

***

### Does the Atlas balancing function call the Kubernetes autoscaler function?

No, Atlas has its own autoscaler, which replaces the default Kubernetes autoscaler.

<br />

***

### When a cluster is connected to Atlas, does it control autoscaling for the entire cluster or select node pools? For example, if I created a new node pool, is there way I can keep it out of the Atlas autoscaler?

Atlas will look for optimization opportunities everywhere by default. However, if you'd like to exclude certain nodes or node pools, you can do that easily by adding this label to them:

```yaml
autoscaling.atlas.example/removal-disabled: "true
```

<br />

***

### Is there a way to provision several or many nodes at the same time?

Atlas provisions compute instances based on workload demands, which may result in our autoscaler automatically provisioning several nodes. The process of node provisioning takes 1 to 2 minutes.

<br />

***

### Should we remove all the existing node groups if we are going for Atlas's full cluster autoscaler?

Atlas supports node groups. However, physical isolation may create waste. So we recommend removing existing node groups if they do not add business value.

<br />

***

### Should we remove the AWS autoscaler deployment?

Example scenario:

In our shared-staging cluster, we still have a deployment called aws-cluster-autoscaler-shared-staging-aws-cluster-autoscaler. If it is going to be overridden/ignored, I'd rather leave it there just in case we need to roll back the Atlas optimization deployment

Please scale down to zero and let Atlas take over the autoscaler part. If you'd like to keep it up, you can. However, the Atlas and cluster autoscaler both will add capacity (this extra capacity will be removed later on).

<br />

***

### Can you please share some guidance on cluster headroom? I would like to add some buffer room so that pods have a place to run when nodes go down.

We recommend using placeholder pods; sleeping pods with the same requests as your critical pods. If you'd like to have room for 3 pods, set up 3 sleeping pods with a low `priorityClass` (negative number).

That way, your high-priority workloads will evict the low-priority pods during a spot interrupt event. The evicted pods will trigger the autoscaler, and the capacity will be replaced.

<br />

***

### Under what condition will the Atlas controller stop scaling up new node? Can we see such an event in the Atlas controller logs?

Atlas will continue scaling as long as the autoscaler is active.

However:

* If a newly added node remains in a non-ready state for 15 minutes, it will be replaced.
* If the kubelet on a node fails to send a heartbeat for 15 minutes, indicating a dead node, it will be replaced.

Currently the [cluster audit log](audit-log.md) is the right place for alerting. You can also make use of [audit events API](https://docs.atlas.example/reference/auditapi_listauditentries).

<br />

***

### What is the implication for our service if Atlas is down?

If Atlas experiences an issue, this wouldn't impact anything running in the cluster. However, it would prevent Atlas from scaling new instances within the cluster.

You can set up Cluster Autoscaler as a delayed secondary autoscaler if desired. For example, if Atlas doesn't schedule a pod within 3 minutes, the Cluster Autoscaler would be a fallback. Atlas is characterized by high availability, and so far, none of our customers have had to use this fallback method.

<br />

***

### What is the logic behind the 70/30 CPU/MEM cost split?

The CPU/MEM split is based on the pricing of different tiers of instances that run the same CPU but have different levels of memory. For instance, if you have 1:2, 1:4, and 1:8 CPU-to-memory ratios, we can abstract out what the cost is for each additional gig of memory and decide what the correct allocation is across different shapes.

<br />

***

### How are costs counted for EC2? Do they include only instances, or traffic, volumes, and other cost factors as well?

Atlas only calculates, for instance, (on-demand or spot) pricing based on the provider's public pricing, updated hourly on our side.

<br />

***

### Does Atlas support GCP Anthos or other on-prem solutions for Kubernetes?

Atlas supports on-premises and alternative Kubernetes environments through **Atlas Anywhere**.

Atlas Anywhere brings our core optimization capabilities—workload autoscaling and resource consolidation—to any Kubernetes cluster. It works alongside your existing infrastructure and autoscaling setup while requiring minimal permissions.

**Learn more:**

* [Atlas Anywhere Overview](atlas-anywhere-overview.md) – understand how it works and key benefits
* [Getting Started with Atlas Anywhere](atlas-anywhere-getting-started.md) – connect your cluster and start optimizing

If you have questions about whether Atlas Anywhere fits your specific environment, reach out to our team or join our [Slack community](https://join.slack.com/t/atlas-community/shared_invite/zt-nwtbsbqq-dWOBiUgLtIGMsUSup7Xksw).

***

### Does Atlas help to reduce costs for GKE?

Yes, Atlas helps reduce costs in your Google Kubernetes Engine clusters. The amount of savings will vary based on the cluster configuration you choose. We have consistently helped our customers save anywhere between 30% and 50% on their GKE costs.

<br />

***

### Can I set the autoscaler to give nodes a buffer zone so when resources hit 80% in either CPU or RAM, Atlas can spin up a new node?

We typically recommend running a dummy pod for this kind of scenario. You can create a deployment or pod with the desired resource requests and give it a low-priority class to create room in the cluster. The low-priority class will allow the dummy pod to be preempted when you introduce a new workload that needs to be scheduled.

Here's an example you can adjust or use as a guide:

Create priority classes:

```shell
cat \<\<EOF | kubectl apply -f -  
apiVersion: scheduling.k8s.io/v1  
kind: PriorityClass  
metadata:  
  name: low-priority  
value: -1  
globalDefault: false  
description: "This priority class should be used for dummy service pods only."  
EOF
```

Low-priority Workloads example:

```shell
cat \<\<EOF | kubectl apply -f -  
apiVersion: apps/v1  
kind: Deployment  
metadata:  
  name: spark-app-dummy  
  namespace: default  
  labels:  
    app: overprovisioning  
spec:  
  replicas: 5  
  selector:  
    matchLabels:  
      app: overprovisioning  
  template:  
    metadata:  
      labels:  
        app: overprovisioning  
    spec:  
      containers:  
      - name: pause  
        image: registry.k8s.io/pause  
        resources:  
          requests:  
            cpu: "4"  
            memory: "5Gi"  
            ephemeral-storage: "150Gi"  
      priorityClassName: low-priority  
      nodeSelector:  
        scheduling.atlas.example/node-template: "atlas-storage-optimized"  
      tolerations:  
      - key: "scheduling.atlas.example/node-template"  
        value: "atlas-storage-optimized"  
        operator: "Equal"  
        effect: "NoSchedule"  
EOF
```

<br />

***

### Can I make sure that there are always at least two on-demand instances (regardless of type) when autoscaling with Atlas?

This use case makes sense if you have some essential services that you'd like to host on on-demand instances instead of Spot Instances to eliminate the chance of failure.

There are a few ways in which you can accomplish this in Atlas:

1. You can do it via pod mutations (<Anchor label="Spot only cluster" target="_blank" href="doc:pod-mutations-quickstart">Spot only cluster</Anchor>).
2. You could also set up a dedicated node template with `lifecycle=on-demand` and workload pinned to this node template via the `nodeSelector`.
3. Finally, if you'd like to hold two on-demand nodes open and keep them from getting bin-packed, you can launch a busy box workload with 2 replicas and a `podAntiAffinity` to itself. Place this on your on-demand `nodeTemplate` with a minimum node size set to the minimum on-demand capacity you want. For instance, if you need 2-4 core on-demand boxes, set the `min-cpu` to 4 cores and your `busybox` placeholder to 2 replicas:

```yaml
apiVersion: apps/v1  
kind: Deployment  
metadata:  
  name: busybox-deployment  
spec:  
  replicas: 2  
  selector:  
    matchLabels:  
      app: busybox  
  template:  
    metadata:  
      labels:  
        app: busybox  
    spec:  
      affinity:  
        podAntiAffinity:  
          requiredDuringSchedulingIgnoredDuringExecution:  
            - labelSelector:  
                matchExpressions:  
                  - key: app  
                    operator: In  
                    values:  
                      - busybox  
              topologyKey: "kubernetes.io/hostname"  
      containers:  
        - name: busybox  
          image: busybox  
          command:  
            - "sleep"  
            - "3600"
```

<br />