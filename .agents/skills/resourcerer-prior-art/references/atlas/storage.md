---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Storage

### Does ATLAS have a way to configure root volume size when using a node template?

Root volume is currently only accessible in node configurations. Currently, this would require a dedicated node configuration and node template.

We tried to simulate the scenario where Template A linked to Node config A. By default Node config A has 100 + 10 CPU per core disk configuration. The pending pod A has a selector for Template A and requests just CPU/MEM - it gets 2 CPU node with 120 GiB disk. The pending pod B has a selector for Template A and in the requests also has a request for ephemeral storage:

```
resources:  
          requests:  
            cpu: '1'  
            ephemeral-storage: "100Gi"
```

This pod gets a 2 CPU node with 213 GiB disk.

<br />

***

### Is it possible to define zones and disk types when creating a node template? Is there an API option available for this purpose?

Node templates don’t offer this option at the moment. However, you can create a custom node config and link it to the template. The node config can have a subset of the cluster’s zones.

A side option would be to add a zone selector for the workloads:

```
nodeSelector:  
   topology.kubernetes.io/zone: "some-zone"
```

Regarding disk types, if you'd like to use storage optimized resources (local SSDs), you can select storage optimized instances only via instance constraints in the template. For more details, see [Storage-optimized nodes](storage-optimized-nodes.md).

<Image align="center" src="https://files.readme.io/419d819-Screenshot_2023-12-21_at_4.43.08_PM.png" />

<br />

***

### How does Atlas manage local disks on storage-optimized nodes?

Atlas uses LVM to pool locally attached NVMe/SSD disks into a single volume on AWS (EKS) and Azure (AKS). On GCP (GKE), the local SSD mounts are managed natively by GKE. On AWS, RAID-0 is also supported for specific configurations. Kubernetes data directories (kubelet, container runtime) are placed on the local storage volume for improved I/O performance.

For a detailed breakdown of how each cloud provider handles local storage, see [Storage-optimized nodes](storage-optimized-nodes.md).

<br />

***

### How does Atlas enforce "ephemeral storage in order to limit users storage on nodes" when consolidating?

Atlas respects Kubernetes storage requests:

```
ephemeral-storage: "100Gi"
```

This means that Atlas will find a node with enough local storage to fit pods that have requirements defined. Read more about it in [Storage-optimized nodes](storage-optimized-nodes.md) and [Pod placement](pod-placement.md#scheduling-on-nodes-with-locally-attached-ssd).

<br />

***

### What is the default volume size for EBS?

The default volume size is 100GiB + CPU to Storage ratio, to increase it further for every CPU provisioned on the node.