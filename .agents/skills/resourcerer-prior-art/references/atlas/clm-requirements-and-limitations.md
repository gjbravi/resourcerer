---
updatedAt: 2026-10-15T07:13:49.000Z
---

# Requirements and limitations

Before enabling container live migration in your cluster, ensure your infrastructure meets the specific requirements for successful pod migration between nodes. This capability requires particular node configurations, compatible instance types, and specific Kubernetes versions to function reliably.

# System requirements

## Cloud platform support

| **Platform** | **Support status** | **Details**                                                                                                           |
| ------------ | ------------------ | --------------------------------------------------------------------------------------------------------------------- |
| AWS EKS      | Supported          | TCP preservation via AWS VPC CNI, Traffic Control, or Calico as full CNI. See [AWS EKS](clm-cloud-providers-eks.md). |
| GCP GKE      | Supported          | TCP preservation via Traffic Control. See [GCP GKE](clm-cloud-providers-gke.md).                                     |
| Azure AKS    | Supported          | TCP preservation via Traffic Control or Calico as full CNI (BYO CNI). See [Azure AKS](clm-cloud-providers-aks.md).   |

For a cross-cloud comparison of node images, TCP preservation paths, ARM64 support, and PVC zone rules, see [Cloud providers](clm-cloud-providers.md).

## Kubernetes version

Container live migration requires **Kubernetes 1.30 or later**. This minimum version ensures compatibility with the container runtime enhancements and custom resource definitions that enable live migration functionality.

Earlier Kubernetes versions lack the necessary API stability and container runtime features required for reliable checkpoint and restore operations. In Kubernetes 1.30 (2024), the checkpoint/restore support graduated to the Beta phase.

## Node infrastructure requirements

**Atlas management**: Both source and destination nodes must be managed by Atlas. Live migration cannot occur between Atlas-managed nodes and nodes managed by other provisioners or cloud provider native node groups.

**Node image requirements**: Supported node images vary by cloud provider. Node images are configured in the cluster's [node configuration](node-configuration.md).

| Cloud | Supported node images                          | Notes                                             |
| ----- | ---------------------------------------------- | ------------------------------------------------- |
| EKS   | Amazon Linux 2023 (standard and NVIDIA/Neuron) | Bottlerocket and Amazon Linux 2 are not supported |
| GKE   | Container-Optimized OS (COS), Ubuntu           | COS recommended for TC (kernel 6.6+)              |
| AKS   | Ubuntu 22.04, Ubuntu 24.04, Azure Linux 3.0    | Ubuntu 24.04 or Azure Linux 3.0 needed for TC     |

For kernel requirements and TCP preservation path availability per image, see the [cloud provider guide](clm-cloud-providers.md).

**Container runtime**: Nodes must use `containerd` v2+ as the container runtime engine. Other runtimes are not supported due to specific integration requirements for checkpoint and restore functionality.

**Network topology**: Networking requirements depend on the cloud provider and the TCP preservation path in use. The most constrained path is the AWS VPC CNI path on EKS, which requires source and destination nodes to be in the same subnet. The Traffic Control (TC) path has no subnet or AZ constraint on any cloud. For full details, see [Cloud providers](clm-cloud-providers.md).

When using the AWS VPC CNI path on EKS, the following networking constraints apply:

**Standard configuration** (`ENABLE_PREFIX_DELEGATION=false`):

* Single subnet per Node and Pod
* No cross-subnet migrations (validated by checking node `topology.atlas.example/subnet-id` label)

**Dedicated subnet for Pods**:

* ENIConfig CRD must be configured in the cluster
* ENIConfig is per availability zone only
* No cross-AZ migration (validated by `topology.kubernetes.io/zone`)
* Source node must have `live.atlas.example/custom-network` label set

**Subnet discovery must not be configured**:

* `kubernetes.io/role/cni` tag must not be configured on any subnets
* **Subnet IP capacity**: Migrations occur within the same subnet. If a subnet's available IP addresses are exhausted, migrations may fail or become stuck waiting for IP allocation. Monitor subnet IP utilization in clusters with frequent migration activity, particularly in smaller subnets or during large-scale rebalancing operations, where many migrations would be triggered at once.

The TC path and the Calico-as-full-CNI path have no subnet or AZ constraint. See [AWS EKS](clm-cloud-providers-eks.md) for the full comparison of TCP preservation paths.

## Instance type and CPU compatibility

**CPU architecture consistency**: CPU architecture consistency is critical for successful live migration. Node templates used for live migration must be configured for a single processor architecture—either AMD64 or ARM64, but not both (any). Attempting to migrate workloads between nodes with different architectures (e.g., from AMD64 to ARM64) will fail. Cloud providers may also provision different CPU models within the same instance type, creating compatibility issues with migration.

**Node template configuration requirements**: Instance families in your [node templates](node-templates.md) must belong to the same generation set for migration compatibility:

* **Compatible sets**: m3 with c3, or c5 with r5, m5, etc.
* **Incompatible**: c3 with c5, or mixing different generation families

Configure your node templates to include only instance types from the same generation family via instance constraints. Migration between nodes with different CPU architectures or generation sets will fail.

**Architecture support**: Container live migration supports both AMD64 and ARM64 architectures, but node templates must be configured for a single architecture. You cannot mix architectures within a live migration-enabled node template.

When configuring your node template:

* Set processor architecture to **either** AMD64 **or** ARM64
* Do not select "Any"

All nodes provisioned from the template will use the selected architecture

There are also functional differences between architectures:

* **AMD64**: Full incremental memory transfer
* **ARM64**: Non-iterative memory transfer

The architectural difference on ARM64 affects the approach to memory dumping during the checkpoint process. This limitation cannot be overcome.

## Container Network Interface (CNI)

TCP preservation during migration is handled differently depending on the cloud provider and CNI in use. There are three TCP preservation paths:

| Path                                   | Pod IP after migration         | Available on  | Kernel requirement |
| -------------------------------------- | ------------------------------ | ------------- | ------------------ |
| **VPC CNI** (forked AWS VPC CNI)       | Same IP (preserved)            | EKS only      | None               |
| **Traffic Control (TC)**               | New IP (peers rewrite packets) | EKS, GKE, AKS | Linux kernel 6.6+  |
| **Calico as full CNI** (VXLAN overlay) | Same IP (annotation pinned)    | EKS, AKS      | None               |

On EKS, a **forked version of the AWS VPC CNI** is automatically installed on Atlas-managed, live-enabled nodes when the VPC CNI path is active. It enables IP address preservation across nodes and TCP session continuity during migration. This fork maintains compatibility with standard AWS VPC networking while adding the necessary features for live migration.

On GKE and AKS (and optionally on EKS), the **Traffic Control (TC)** path uses a privileged DaemonSet that rewrites packets on peer nodes via eBPF so established TCP connections survive migration even though the pod gets a new IP. TC requires Linux kernel 6.6 or later.

On EKS and AKS, **Calico as the full CNI** with VXLAN overlay preserves the pod IP via annotation pinning, without the subnet constraint of the VPC CNI path and without the kernel requirement of TC.

For per-cloud setup instructions and CNI configuration details, see:

* [AWS EKS](clm-cloud-providers-eks.md)
* [GCP GKE](clm-cloud-providers-gke.md)
* [Azure AKS](clm-cloud-providers-aks.md)

For detailed configuration reference for each TCP preservation path, see:

* [Traffic Control](clm-reference-traffic-control.md)
* [AWS VPC CNI](clm-reference-aws-vpc-cni.md)

# Supported workload types

Container live migration supports the following Kubernetes workload types:

| **Workload type**      | **Support status** | **Notes**                                           |
| ---------------------- | ------------------ | --------------------------------------------------- |
| **StatefulSets**       | Supported          |                                                     |
| **Deployments**        | Supported          |                                                     |
| **Bare pods**          | Supported          |                                                     |
| **Jobs**               | Supported          |                                                     |
| **CronJobs**           | Supported          |                                                     |
| **Custom controllers** | Limited support    | Compatibility assessment required. Contact Atlas. |
| **DaemonSets**         | Not supported      | Cannot be migrated by design.                       |

**Multi-container support**: Pods with multiple containers are fully supported. All containers within a pod are migrated together, **except Init containers**: Init containers are skipped during migration by default and will not be rerun on the restored pod.

# Storage compatibility

## Supported storage types

| **Storage type**                    | **Support scope**                                                                       |
| :---------------------------------- | :-------------------------------------------------------------------------------------- |
| **Persistent Volume Claims (PVCs)** | Zonal PVs: same zone required; Regional PVs: any zone in PV's allowed list (all clouds) |
| **Network File System (NFS)**       | Cross-node                                                                              |
| **EmptyDir volumes**                | Supported                                                                               |
| **ConfigMap volumes**               | Supported                                                                               |
| **Secret volumes**                  | Supported                                                                               |
| **Host path volumes**               | Experimental support only. Contact Atlas for specific use cases.                      |

**Host path volumes**: Host path volumes present unique challenges because they access node-local storage. Limited support exists for some use cases.

**PVC reattachment timing**: When migrating workloads with Persistent Volume Claims, the PVC must be detached from the source node and reattached to the destination node. On EKS, the average reattachment duration is approximately 9 seconds, though this varies by volume type, size, and current cloud API latency. Factor this additional time into migration planning for PVC-backed workloads and test in your environment to establish baseline expectations.

# Current limitations

## Hardware and performance constraints

**GPU workloads**: Not currently supported due to the complexity of checkpointing and restoring GPU memory and compute.

For maximizing the efficiency of GPU workloads on a given node, Atlas offers sophisticated alternative approaches in the forms of [MIG and GPU time-slicing](gpu-sharing-mig.md).

**Memory transfer performance**: Migration time increases with workload memory usage. Large memory footprints may experience naturally longer migration times.

**Spot Instance compatibility**: When used with Spot Instances, container live migration cannot guarantee successful migration before the CSP terminates the node. Spot interruption notices provide limited time windows that may be insufficient for completing the migration.

**Before enabling container live migration on Node templates configured for Spot Instances, contact Atlas support to assess your workloads and determine the appropriate configuration**.

## Problematic workloads

Certain workload types may have an increased rate of failure during migration, even when labeled as eligible. The live controller cannot detect all incompatible configurations, so test these workload types thoroughly before production use.

| Workload type                 | Issue                                                           | Mitigation                                                             |
| :---------------------------- | :-------------------------------------------------------------- | :--------------------------------------------------------------------- |
| Applications using `io_uring` | The `io_uring` async I/O interface state cannot be checkpointed | Disable at kernel level via node configuration init script (see below) |
| Applications using MPTCP      | Multipath TCP connection state cannot be checkpointed           | Disable MPTCP in your application (see below)                          |
| eBPF programs                 | eBPF program state is not checkpointed                          | Not currently supported for live migration                             |
| Privileged containers         | Multiple checkpoint incompatibilities                           | Avoid privileged containers for CLM-enabled workloads                  |

**Disabling&#x20;**`io_uring`**&#x20;for CLM compatibility:**

Add the following to your [node configuration](node-configuration.md) init script to disable `io_uring` at the kernel level:

```shell
sudo sysctl -w kernel.io_uring_disabled=2
```

This setting prevents applications from using the `io_uring` interface, which cannot be checkpointed by CRIU.

**Disabling MPTCP for CLM compatibility:**

For Go applications, disable MPTCP by setting the following environment variable:

```shell
GODEBUG=multipathtcp=0
```

Add this to your container's environment variables in the pod spec. Other languages and frameworks may have similar configuration options to disable MPTCP.

## Scheduling constraints set by admission webhooks

Live migration plans the destination from the scheduling constraints (`nodeSelector`, affinity, tolerations) that are part of the pod's spec, including constraints an admission webhook injected when the pod was created.

Admission policies whose scheduling outcome is decided when each pod is created are not supported. The decision is made again for the pod created by the migration and can differ from the source pod, so the migrated pod can be unschedulable on the planned destination.

<Callout icon="📘" theme="info">
  ### Eligibility vs. success

  The `live.atlas.example/migration-enabled=true` label indicates a workload meets basic eligibility criteria. It does not guarantee successful migration. Always validate specific workload types in a test environment before production rollout.
</Callout>

# Automatic workload assessment

Atlas automatically evaluates workloads for migration eligibility:

**Automatic labeling**: The live controller scans your cluster and applies migration-eligible labels (`live.atlas.example/migration-enabled=true`) to workloads that meet all requirements.

**Continuous assessment**: As workloads change (storage additions, security context modifications, etc.), the controller updates eligibility labels accordingly.

## Checking workload eligibility

The simplest way to check if workloads are eligible for live migration is to look for the migration label:

```shell
# Check all pods with live migration enabled
kubectl get pods -A -l live.atlas.example/migration-enabled=true
```

Do note that this will only work once container live migration is enabled and the live controller is operating in the cluster.

# See also

<Cards>
  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Per-cloud setup guides for EKS, GKE, and AKS — node images, TCP preservation, ARM64, and PVC support.
  </Card>

  <Card title="Probe and Lifecycle Behavior" href="https://docs.atlas.example/docs/clm-concept-probe-lifecycle-behavior" icon="fa-heartbeat">
    How Kubernetes probes and lifecycle hooks behave during migration.
  </Card>

  <Card title="Labels, Annotations, and Events" href="https://docs.atlas.example/docs/clm-reference-labels-and-annotations" icon="fa-tags">
    Reference for all CLM labels, annotations, and migration events.
  </Card>
</Cards>

<br />