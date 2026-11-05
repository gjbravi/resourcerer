---
updatedAt: 2026-09-29T11:02:53.000Z
---

# Storage-optimized nodes

Use storage-optimized nodes with locally attached NVMe/SSD disks for workloads that need high-throughput, low-latency local storage.

Storage-optimized nodes have locally attached NVMe/SSD disks that provide significantly higher throughput and lower latency than network-attached volumes (such as EBS, Persistent Disk, or Azure Managed Disks). They are ideal for workloads that require fast local I/O: caching layers, temporary data processing, log aggregation, or any application that benefits from high-speed ephemeral storage.

## How it works

When Atlas provisions a storage-optimized node, the following happens:

1. An instance type with local NVMe/SSD disks is selected.
2. The local disks are pooled together (using LVM or managed natively by the cloud provider) into a single volume.
3. Kubernetes data directories (`kubelet`, container runtime data) are placed on the fast local storage.
4. The boot/OS disk size is reduced (typically to 100 GB), since bulk data lives on the local disks.
5. The node is labeled with `scheduling.atlas.example/storage-optimized=true`.
6. The node is tainted with `scheduling.atlas.example/storage-optimized:NoSchedule`.

Because of the taint, only pods that explicitly tolerate it will be scheduled on storage-optimized nodes. This prevents regular workloads from accidentally landing on these specialized (and typically more expensive) instances.

> ⚠️ Warning
>
> Local storage is **ephemeral**. All data on locally attached disks is permanently lost when the node is terminated, replaced, or stopped. Do not rely on local storage for persistent data. Use Persistent Volumes backed by network-attached storage for any data that must survive node lifecycle events.

## Supported providers

| Provider  | Supported |
| --------- | --------- |
| AWS EKS   | Yes       |
| GCP GKE   | Yes       |
| Azure AKS | Yes       |

## Cloud provider differences

While the user-facing behavior is the same across clouds (label, taint, pod scheduling), the underlying mechanisms differ.

<Tabs>
  <Tab title="AWS (EKS)">

**Local storage source:** NVMe instance storage built into the instance type (for example, `i3`, `c5d`, `c6id`, `m5d` families).

**Volume management:** Atlas uses LVM to pool all local NVMe disks into a single volume. The kubelet, containerd, and Docker data directories are placed on this volume.

**Boot disk optimization:** When local NVMe storage capacity meets or exceeds the requested volume size, Atlas uses the AMI's default (smaller) block device mapping instead of provisioning a full-size EBS volume. This reduces cost.

**Instance selection:** Only instances with both an NVMe storage driver **and** SSD storage devices are eligible. Instances without local NVMe disks cannot be storage-optimized on AWS.

  </Tab>
  <Tab title="GCP (GKE)">

**Local storage source:** Local SSD disks attached to the instance at creation time. Each local SSD is exactly **375 GB**. Some instance families (names containing `-lssd`) have pre-attached local SSDs; others (such as `n1`, `n2`, `c2`) support attaching SSDs on demand.

**Volume management:** GKE manages the local SSD mounts natively; Atlas attaches the correct number of disks and configures the appropriate kube-env variables.

**Boot disk optimization:** The boot disk is reduced to **100 GB** when local SSDs are attached.

**Instance selection:** Instances with `-lssd` in the name are always storage-optimized. Other NVMe-capable families can be storage-optimized on demand.

> 📘 GCP local SSD granularity
>
> Because each local SSD is exactly 375 GB, the total local storage must be an exact multiple of 375 GB. Atlas automatically rounds up your requested storage to the nearest valid boundary. For example, requesting 500 GB results in 750 GB (2 x 375 GB) of local storage.

  </Tab>
  <Tab title="Azure (AKS)">

**Local storage source:** NVMe disks on instance types that support them (such as `Lsv3`, `Lasv3` families). If the instance type has no NVMe disks but has Azure's standard ephemeral/temp disk (mounted at `/mnt`), Atlas uses that instead.

**Volume management:** When NVMe disks are present, Atlas uses LVM to pool them. The kubelet and Docker data directories are placed on this volume. When no NVMe disks are found, the Azure ephemeral disk at `/mnt` is used.

**Boot disk optimization:** The OS disk is reduced to **100 GB** for storage-optimized nodes (unless ephemeral OS disks are also configured — see below).

**Instance selection:** AKS considers **all** instance types with attached SSDs as storage-optimized, not just those in Azure's "Storage Optimized VMs" family. This means you may get instances from different VM families than expected.

For nodes with local disks but without the storage-optimized label and taint, use an [ephemeral OS disk](node-configuration.md#ephemeral-os-disk-for-aks) with a resource disk or NVMe disk placement. That filters instance types to local-storage types without applying the taint.

> 📘 AKS ephemeral OS disks
>
> Azure AKS also supports **ephemeral OS disks**, which is a separate feature from storage-optimized nodes. With ephemeral OS disks, the OS disk itself is placed on local storage (cache disk, resource disk, or NVMe disk) instead of a managed disk, resulting in faster boot times and no OS disk cost. This is configured at the **node configuration** level, not via pod scheduling. The two features can be combined: a node can have both an ephemeral OS disk and local NVMe data disks for storage optimization. See the [Azure ephemeral OS disks documentation](https://learn.microsoft.com/en-us/azure/virtual-machines/ephemeral-os-disks) for more details.

  </Tab>
</Tabs>

## Scheduling workloads on storage-optimized nodes

There are two ways to get storage-optimized nodes: per-pod scheduling (where individual pods request storage-optimized nodes) and template-wide constraints (where all nodes from a template are storage-optimized).

### Per-pod scheduling

Add a `nodeSelector` and a toleration for the `scheduling.atlas.example/storage-optimized` label to your pod spec. When the Atlas Autoscaler detects this pod as unschedulable, it provisions a storage-optimized node.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: io-intensive-app
spec:
  nodeSelector:
    scheduling.atlas.example/storage-optimized: "true"
  tolerations:
    - key: scheduling.atlas.example/storage-optimized
      operator: Exists
      effect: NoSchedule
  containers:
    - name: app
      image: my-app:latest
      resources:
        requests:
          ephemeral-storage: "50Gi"
      volumeMounts:
        - name: scratch
          mountPath: "/data"
  volumes:
    - name: scratch
      emptyDir: {}
```

You can also use a `requiredDuringSchedulingIgnoredDuringExecution` node affinity instead of a `nodeSelector`:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
        - matchExpressions:
            - key: scheduling.atlas.example/storage-optimized
              operator: In
              values:
                - "true"
```

### Node template constraints

To make **all** nodes provisioned by a specific node template storage-optimized, enable the `storageOptimized` constraint on the template. This is useful when you have a dedicated template for storage-intensive workloads.

When this constraint is enabled, Atlas only selects instance types that support local storage for that template. You don't need to add storage-optimized node selectors or tolerations to every pod — instead, use the template's own label and toleration to target it:

```yaml
nodeSelector:
  scheduling.atlas.example/node-template: "my-storage-template"
tolerations:
  - key: scheduling.atlas.example/node-template
    value: "my-storage-template"
    operator: Equal
    effect: NoSchedule
```

For more information on configuring node templates, see [Node templates](node-templates.md).

### Controlling the storage amount

Pods can control the amount of local storage provisioned by specifying `ephemeral-storage` in their resource requests:

```yaml
resources:
  requests:
    ephemeral-storage: "100Gi"
```

Atlas uses this request to select an instance type with sufficient local storage capacity. If you don't specify an `ephemeral-storage` request, Atlas provisions a storage-optimized node with the smallest available local storage for the selected instance type.

## Important considerations

### Data volatility

Local storage is **ephemeral**. Data on locally attached disks is lost when the node is terminated, stopped, preempted (spot), or replaced during operations like rebalancing or upgrades. Always use Persistent Volumes backed by network-attached storage (EBS, Persistent Disk, Azure Managed Disks) for data that must persist.

### Rebalancing and node draining

Pods using local persistent volumes on storage-optimized nodes can block node draining during [rebalancing](scheduled-rebalancing.md). If you use local persistent volumes on storage-optimized nodes, consider enabling the `ignoreLocalPersistentVolumes` option in your rebalancing configuration to allow these nodes to be drained. Be aware that this **will result in data loss** for any data stored in local persistent volumes.

### GCP local SSD size granularity

On GCP, local SSD capacity must be an exact multiple of 375 GB. If your workload requests an amount that isn't a multiple (for example, 500 GB), Atlas rounds up to the nearest valid value (750 GB). Plan your `ephemeral-storage` requests accordingly to avoid over-provisioning.

### Instance type availability

Not all instance types support local storage. When you request storage-optimized nodes, Atlas automatically filters out instance types that don't have locally attached disks. If your node template has strict constraints (for example, specific instance families), ensure that at least some of the allowed families support local storage.

### Cost implications

Storage-optimized nodes can be more cost-effective for I/O-heavy workloads because:

* The boot/OS disk is smaller (100 GB instead of a larger network-attached volume), reducing storage costs.
* On AWS, when local NVMe capacity is sufficient, Atlas skips provisioning an additional EBS volume entirely.
* Local SSDs offer higher throughput and lower latency than network-attached storage at no additional per-GB cost (the cost is included in the instance price).

However, instance types with local storage are typically more expensive than their non-storage counterparts. Evaluate whether the I/O performance benefits justify the instance cost for your workload.