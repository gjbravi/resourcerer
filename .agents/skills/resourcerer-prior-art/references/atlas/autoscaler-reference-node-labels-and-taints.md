---
updatedAt: 2026-10-09T11:52:46.000Z
---

# Autoscaler Node Labels and Taints

Reference of all labels and taints that Atlas automatically applies to nodes it manages.

This document lists all labels and taints that Atlas automatically applies to nodes it manages.

> 🚧 Important: Reserved Labels and Taints
>
> Atlas reserves the labels and taints listed in this document. If you configure any of these in a **Node Template**, Atlas will **overwrite your values** with the ones
> it determines based on the node's instance type, cloud provider, and feature configuration.
>
> **Do not configure these reserved labels or taints in Node Templates.** Doing so will not produce the expected result, and may cause pods to not be scheduled correctly.

# Labels

## Node Identity & Management

| Label                                          | Value       | When Applied                                                                                                                      |
| ---------------------------------------------- | ----------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `provisioner.atlas.example/node-id`                  | UUID        | Set at node creation. Unique identifier for the node assigned by Atlas.                                                         |
| `provisioner.atlas.example/managed-by`               | `atlas.example`   | Set at node creation. Marks the node as managed by Atlas. Used by controllers to filter Atlas-owned nodes.                    |
| `provisioner.atlas.example/node-configuration-name`  | config name | Set at node creation. References the node configuration name used to provision the node.                                          |
| `provisioner.atlas.example/node-configuration-id`    | config UUID | Set at node creation. References the node configuration ID used to provision the node.                                            |
| `provisioner.atlas.example/hyper-threading-disabled` | `true`      | Set when the node configuration disables hyper-threading. Used to distinguish nodes where HT has been turned off at the OS level. |
| `charts.atlas.example/managed`                       | varies      | Set on nodes managed via Helm chart deployments.                                                                                  |

## Lifecycle / Instance Type

| Label                                 | Value         | When Applied                                                                                                    |
| ------------------------------------- | ------------- | --------------------------------------------------------------------------------------------------------------- |
| `scheduling.atlas.example/spot`             | `true`        | Set when the node is a spot/preemptible instance. Used for spot-aware scheduling.                               |
| `scheduling.atlas.example/on-demand`        | `true`        | Set when the node is an on-demand instance. Mutually exclusive with `scheduling.atlas.example/spot`.                  |
| `scheduling.atlas.example/spot-fallback`    | `true`        | Set when the node is a spot-fallback instance (on-demand used as fallback for a spot configuration).            |
| `scheduling.atlas.example/spot-reliability` | score (0–100) | Reliability score of the spot instance type, used by the scheduler to prefer more stable spot instances.        |
| `scheduling.atlas.example/interrupted`      | `true`        | Set when a spot node receives an interruption notice. Signals the node is being drained due to spot preemption. |

For details on configuring workloads for spot instances, including fallback and diversity settings, see [Spot Instances](spot.md).

## Node Templates

| Label                                      | Value          | When Applied                                                                                           |
| ------------------------------------------ | -------------- | ------------------------------------------------------------------------------------------------------ |
| `scheduling.atlas.example/node-template`         | template name  | Set when the node is provisioned using a specific node template.                                       |
| `scheduling.atlas.example/node-template-version` | version string | Set when a node template is applied or updated. Tracks the version of the node template configuration. |

For information on creating and configuring node templates, including custom labels and workload targeting, see [Node Templates](node-templates.md).

## Compute / Storage Optimization

| Label                                    | Value                                                 | When Applied                                                                                                                                                                                                                                            |
| ---------------------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scheduling.atlas.example/compute-optimized`   | `true`                                                | Set when the node is compute-optimized (high CPU-to-memory ratio). Used to schedule compute-intensive workloads.                                                                                                                                        |
| `scheduling.atlas.example/storage-optimized`   | `true`                                                | Set when the node is storage-optimized (locally attached NVMe/SSD disks). See [Storage-optimized nodes](storage-optimized-nodes.md).                                                                                                                   |
| `scheduling.atlas.example/cpu-manufacturer`    | `amd`, `intel`, `aws`, `ampere`, `apple`, or `google` | Set when the instance type has exactly one known CPU vendor. Not set on instances with multiple or unknown manufacturers. Use in a workload's `nodeSelector` or `nodeAffinity` to target a specific CPU vendor without creating a custom node template. |
| `scheduling.atlas.example/instance-cpu`        | vCPU count (integer string)                           | Set to the number of vCPUs on the node. Use with `nodeAffinity` `Gt`/`Lt` operators (strict: `Gt: "15"` matches 16+) to target nodes above or below a core count, or with `nodeSelector` for an exact match.                                            |
| `scheduling.atlas.example/instance-family`     | e.g. `c6a`, `m5` (AWS); `n2` (GCP)                    | Set to the instance family identifier (the part of the instance type name before the first separator). Use with `nodeAffinity` `In` to target specific families, or `NotIn` to keep a workload off them, or with `nodeSelector` for an exact match.     |
| `scheduling.atlas.example/instance-generation` | e.g. `5`, `6` (integer string)                        | **AWS only.** Set to the instance generation number. Use in a workload's `nodeSelector` to target a specific generation within a family.                                                                                                                |
| `scheduling.atlas.example/instance-hypervisor` | `nitro` or `xen`                                      | **AWS only.** Set to the hypervisor type of the instance. Use in a workload's `nodeSelector` to require a specific hypervisor.                                                                                                                          |
| `scheduling.atlas.example/instance-memory`     | Memory in MiB (integer string)                        | Set to the node's total memory in MiB. Use with `nodeAffinity` `Gt`/`Lt` operators (strict: `Gt: "65536"` matches nodes with more than 64 GiB) to target nodes above or below a memory threshold, or with `nodeSelector` for an exact amount.           |
| `scheduling.atlas.example/premium-storage`     | `true`                                                | Set on AKS nodes that support premium storage (Premium\_LRS, PremiumV2\_LRS, Premium\_ZRS).                                                                                                                                                             |

## GPU / Accelerators

| Label                                                | Value                          | When Applied                                                                                                                                 |
| ---------------------------------------------------- | ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `nvidia.com/gpu`                                     | `true`                         | Set by the autoscaler during GPU node provisioning. Applied when the instance type has an NVIDIA GPU device.                                 |
| `nvidia.com/gpu.present`                             | `true`                         | Set by the autoscaler alongside `nvidia.com/gpu` during GPU node provisioning. Acts as an alternative indicator of GPU presence.             |
| `nvidia.com/gpu.name`                                | GPU model                      | GPU model name (e.g. `A100`). Set by the autoscaler based on the instance type's GPU device metadata.                                        |
| `nvidia.com/gpu.count`                               | integer                        | Number of physical GPUs on the node. Set by the autoscaler based on instance type GPU count.                                                 |
| `nvidia.com/gpu.memory`                              | MiB                            | Memory per single GPU in MiB. Set by the autoscaler based on instance type GPU memory.                                                       |
| `nvidia.com/gpu.total-memory`                        | MiB                            | Total GPU memory across all GPUs on the node.                                                                                                |
| `nvidia.com/gpu.mig`                                 | `true`                         | Set when MIG (Multi-Instance GPU) partitioning is enabled.                                                                                   |
| `nvidia.com/gpu.mig-partition-{size}`                | `true`                         | Set per MIG partition size (e.g. `nvidia.com/gpu.mig-partition-1g.10gb`). Indicates available MIG slice profiles.                            |
| `nvidia.com/gpu.dra`                                 | `true`                         | Set when the node uses the NVIDIA DRA (Dynamic Resource Allocation) driver instead of the NVIDIA device plugin.                              |
| `scheduling.atlas.example/gpu.count`                       | integer                        | Atlas internal GPU count used during scheduling simulation. Set alongside NVIDIA labels.                                                   |
| `scheduling.atlas.example/gpu-shared`                      | integer                        | Set when GPU time-sharing or MPS is configured. Value is the number of max shared clients per GPU.                                           |
| `scheduling.atlas.example/gpu-sharing-strategy`            | `time-sharing` or `mps`        | Set alongside `scheduling.atlas.example/gpu-shared` to indicate the GPU sharing strategy configured for the node.                                  |
| `scheduling.atlas.example/nvidia-device-plugin-static-pod` | `true`                         | Set on AWS AL2023 nodes with GPU sharing enabled. Signals that the NVIDIA device plugin is deployed as a static pod rather than a DaemonSet. |
| `scheduling.atlas.example/gpu-partition-size`              | partition size (e.g. `1g.5gb`) | Set on nodes where MIG is configured in single mode. Indicates the active MIG partition size.                                                |
| `scheduling.atlas.example/bottlerocket-gpu-partition-size` | partition size (e.g. `1g.5gb`) | Set on AWS Bottlerocket nodes with MIG configured in single mode.                                                                            |
| `scheduling.atlas.example/preinstalled-nvidia-driver`      | `true`                         | Set on nodes that come with a pre-installed NVIDIA driver (e.g. Bottlerocket).                                                               |
| `aws.amazon.com/neuron`                              | `true`                         | Set on AWS instances with Inferentia/Trainium (Neuron) accelerators. Used together with `TaintNeuron`.                                       |
| `cloud.google.com/gke-accelerator`                   | GPU type                       | Set on GKE GPU nodes by the autoscaler. Indicates the accelerator type (e.g. `nvidia-tesla-t4`).                                             |
| `cloud.google.com/gke-gpu-driver-version`            | version                        | Set on GKE nodes when using the default NVIDIA device plugin. Specifies the GPU driver version to install.                                   |
| `cloud.google.com/gke-gpu-sharing-strategy`          | `time-sharing` or `mps`        | Set on GKE nodes when GPU sharing is configured. GKE-specific counterpart to `scheduling.atlas.example/gpu-sharing-strategy`.                      |
| `cloud.google.com/gke-max-shared-clients-per-gpu`    | integer                        | Set on GKE nodes when GPU sharing is configured. GKE-specific counterpart to `scheduling.atlas.example/gpu-shared`.                                |
| `gke-no-default-nvidia-gpu-device-plugin`            | `true`                         | Set on GKE nodes when Atlas manages GPU plugin deployment instead of the default GKE NVIDIA DaemonSet.                                     |
| `cloud.google.com/gke-gpu-partition-size`            | partition size                 | Set on GKE nodes with MIG enabled. Indicates the MIG partition profile for the node.                                                         |

For GPU provisioning setup, workload configuration examples, and sharing strategies, see [GPU Instances](gpu.md). For MIG partitioning, see [GPU sharing with MIG](gpu-sharing-mig.md). For time-slicing, see [GPU sharing with time-slicing](gpu-sharing-time-slicing.md).

## Rebalancing

| Label                              | Value         | When Applied                                                                                                                                                                                                                        |
| ---------------------------------- | ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `rebalancing.atlas.example/plan-id`      | UUID          | Set on green nodes created as part of a rebalancing plan. Used to associate the node with its rebalancing operation.                                                                                                                |
| `rebalancing.atlas.example/operation-id` | UUID          | Set on green nodes to identify the specific operation within a rebalancing plan.                                                                                                                                                    |
| `scheduling.atlas.example/delete-reason` | plugin name   | Set when the autoscaler marks a node for deletion.                                                                                                                                                                                  |
| `autoscaling.atlas.example/draining`     | reason string | Set on nodes being drained by the rebalancer or evictor. Possible values: `rebalancing`, `aws-rebalance-recommendation`, `spot-prediction`, `spot-fallback`, `spot-interruption`, `evictor`. Applied alongside `TaintNodeDraining`. |

For information on rebalancing operations and how to prepare workloads, see [Rebalancing](rebalancing.md) and [Workload preparation](preparation.md). To understand how Evictor and Rebalancer work together, see [Evictor vs. Rebalancer](node-autoscaling-evictor-vs-rebalancer.md).

## Removal Control

| Label                                         | Value                    | When Applied                                                                                                                                     |
| --------------------------------------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `autoscaling.atlas.example/removal-disabled`        | `true`                   | Set on a node or pod to prevent the rebalancer from evicting pods from that node.                                                                |
| `autoscaling.atlas.example/removal-disabled-until`  | Unix timestamp (seconds) | Set to temporarily prevent removal until the given timestamp. Applied during green node initialization to protect it from premature rebalancing. |
| `autoscaling.atlas.example/live-migration-disabled` | `true`                   | Pod-level label/annotation. Disables live migration for pods that cannot tolerate it.                                                            |

<Accordion title="kubectl: Query removal-protected nodes and pods" icon="fa-terminal">
  ```shell
  # List nodes protected from removal
  kubectl get nodes -l autoscaling.atlas.example/removal-disabled=true

  # List pods protected from removal
  kubectl get pods -A -l autoscaling.atlas.example/removal-disabled=true

  # List pods opted out of live migration
  kubectl get pods -A -l autoscaling.atlas.example/live-migration-disabled=true
  ```
</Accordion>

For details on Evictor override rules and advanced configuration, see [Evictor](evictor.md). For Container Live Migration opt-out behavior, see [CLM Labels, Annotations, and Events](clm-reference-labels-and-annotations.md).

## Live Migration (CLM / LIVE)

| Label                            | Value  | When Applied                                                                                                                                                           |
| -------------------------------- | ------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `live.atlas.example/install`           | varies | Set by Atlas to indicate CLM (Cluster Live Migration) component installation status on the node.                                                                     |
| `live.atlas.example/migration-enabled` | `true` | Set by the LIVE daemonset once all LIVE components are installed and operational on the node. Indicates the node is eligible as both migration source and destination. |

<Accordion title="kubectl: Check CLM status" icon="fa-terminal">
  ```shell
  # List nodes with Container Live Migration enabled
  kubectl get nodes -l live.atlas.example/migration-enabled=true

  # List pods eligible for live migration
  kubectl get pods -A -l live.atlas.example/migration-enabled=true

  # Monitor active migrations
  kubectl get migrations -A -w
  ```
</Accordion>

For an overview of Container Live Migration, see [Container Live Migration](clm-overview.md). For setup instructions, see [Getting started with CLM](clm-getting-started.md). For the full CLM labels and annotations reference, see [CLM Labels, Annotations, and Events](clm-reference-labels-and-annotations.md).

## Predictions / ML

| Label                             | Value   | When Applied                                                                                                                                                                       |
| --------------------------------- | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `predictions.atlas.example/ttl-minutes` | integer | Set on nodes that were rebalanced due to an ML-predicted spot interruption. Specifies how many minutes the replacement node should be kept alive after the original was evacuated. |

## Volume Support

| Label                                     | Value  | When Applied                                                                 |
| ----------------------------------------- | ------ | ---------------------------------------------------------------------------- |
| `volume.scheduling.atlas.example/{volume-name}` | `true` | Set per storage volume/class to indicate the node supports that volume type. |

## Topology

| Label                                | Value                 | When Applied                                                                                     |
| ------------------------------------ | --------------------- | ------------------------------------------------------------------------------------------------ |
| `topology.atlas.example/csp`               | `aws`, `gcp`, `azure` | Set at node creation. Identifies the cloud service provider.                                     |
| `topology.atlas.example/subnet-id`         | subnet ID             | Set at node creation. Identifies the subnet the node was provisioned in.                         |
| `topology.atlas.example/pod-subnet-id`     | subnet ID             | Set on AKS nodes. Identifies the subnet used for pod IP allocation.                              |
| `topology.atlas.example/resource-group`    | resource group        | Set on AKS nodes. Azure resource group containing the node.                                      |
| `topology.atlas.example/virtual-network`   | vnet name             | Set on AKS nodes. Azure virtual network name.                                                    |
| `topology.atlas.example/subscription-id`   | subscription ID       | Set on AKS nodes. Azure subscription ID.                                                         |
| `topology.disk.csi.azure.com/zone`   | AZ name               | Set on Azure nodes. Availability zone for Azure CSI disk topology.                               |
| `topology.ebs.csi.aws.com/zone`      | AZ name               | Set on AWS nodes. Availability zone for EBS CSI disk topology.                                   |
| `network-tag.gcp.atlas.example/{tag-name}` | `true`                | Set on GCP nodes for each network tag associated with the node. Prefix-based, one label per tag. |
| `topology.kubernetes.io/region`      | region                | Standard Kubernetes label. Set at node creation with the cloud region.                           |
| `topology.kubernetes.io/zone`        | AZ name               | Standard Kubernetes label. Set at node creation with the availability zone.                      |
| `topology.gke.io/zone`               | AZ name               | GKE-specific zone label.                                                                         |

For information on configuring pod placement by topology, see [Pod placement](pod-placement.md). For subnet configuration, see [Subnets](subnets.md).

## Cloud Provider Specific

### AKS (Azure)

| Label                               | Value              | When Applied                                                                                  |
| ----------------------------------- | ------------------ | --------------------------------------------------------------------------------------------- |
| `kubernetes.azure.com/agentpool`    | pool name          | Set to `atlas` for Atlas provisioned nodes. Required by AKS for node pool membership.      |
| `kubernetes.azure.com/cluster`      | cluster name       | Set to identify the AKS cluster.                                                              |
| `kubernetes.azure.com/mode`         | `system` or `user` | Set to `system` on system node pools.                                                         |
| `kubernetes.azure.com/role`         | `agent`            | Set on all AKS agent nodes.                                                                   |
| `agentpool` *(deprecated)*          | pool name          | Deprecated AKS agent pool label (pre k8s 1.24). Replaced by `kubernetes.azure.com/agentpool`. |
| `kubernetes.io/role` *(deprecated)* | role               | Deprecated AKS role label (pre k8s 1.24).                                                     |

#### EKS (AWS)

| Label                            | Value     | When Applied                                                           |
| -------------------------------- | --------- | ---------------------------------------------------------------------- |
| `eks.amazonaws.com/compute-type` | `fargate` | Set on Fargate nodes. Used to distinguish Fargate from EC2 node types. |

## Standard Kubernetes Labels (set by autoscaler)

| Label                              | Value              | When Applied                                                                            |
| ---------------------------------- | ------------------ | --------------------------------------------------------------------------------------- |
| `kubernetes.io/arch`               | `amd64`, `arm64`   | Set at node creation based on instance architecture.                                    |
| `kubernetes.io/os`                 | `linux`, `windows` | Set at node creation based on instance OS.                                              |
| `beta.kubernetes.io/arch`          | `amd64`, `arm64`   | Legacy beta arch label, set alongside `kubernetes.io/arch` for backwards compatibility. |
| `beta.kubernetes.io/os`            | `linux`, `windows` | Legacy beta OS label.                                                                   |
| `node.kubernetes.io/instance-type` | instance type name | Standard label set to the cloud provider instance type (e.g. `m5.xlarge`).              |
| `kubernetes.io/hostname`           | hostname           | Standard hostname label.                                                                |

## Resource Offering

| Label                                               | Value         | When Applied                                                                     |
| --------------------------------------------------- | ------------- | -------------------------------------------------------------------------------- |
| `autoscaling.atlas.example/provisioned-resource-offering` | offering type | Set at node creation to record the resource offering type used for provisioning. |

## OMNI Edge

| Label                                   | Value  | When Applied                                                                 |
| --------------------------------------- | ------ | ---------------------------------------------------------------------------- |
| `virtual-node.omni.atlas.example/not-allowed` | `true` | Set on nodes that are not allowed to run workloads in the OMNI edge context. |

# Taints

## Rebalancing Taints

| Key                                        | Value    | Effect       | When Applied                                                                                                                                                                                                                                                                                         |
| ------------------------------------------ | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `rebalancing.atlas.example/preparing`            | *(none)* | `NoSchedule` | Applied to green (replacement) nodes during rebalancing plan execution. Prevents new pods from being scheduled until the node is fully prepared. Removed once the node is ready to receive workloads.                                                                                                |
| `scheduling.atlas.example/pod-pinning-preparing` | *(none)* | `NoSchedule` | Applied during rebalancing when pod pinning is enabled. Prevents scheduling until pod pinning setup is complete.                                                                                                                                                                                     |
| `autoscaling.atlas.example/draining`             | `true`   | `NoSchedule` | Applied to blue (old) nodes when they are being drained during rebalancing or spot interruption handling. Applied alongside the `autoscaling.atlas.example/draining` label. Also present on nodes that fail to drain when graceful eviction is enabled. Removed automatically when the node is uncordoned. |
| `provisioner.atlas.example/uninitialized`        | *(none)* | `NoSchedule` | Applied to newly provisioned nodes before initialization is complete. Prevents pod scheduling until the node is fully set up.                                                                                                                                                                        |

For details on how rebalancing operates and the blue/green node replacement process, see [Rebalancing](rebalancing.md). For paused drain configuration, see [Paused drain configuration](rebalancing-paused-drain-configuration.md).

## Lifecycle / Spot Taint

| Key                       | Value    | Effect       | When Applied                                                                                                                                                                                                                 |
| ------------------------- | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scheduling.atlas.example/spot` | *(none)* | `NoSchedule` | Applied to spot nodes when the lifecycle taint feature is enabled in node template configuration. Requires pods to explicitly tolerate spot instances. Only applied when both spot and on-demand nodes exist in the cluster. |

For workload configuration patterns using spot tolerations and node selectors, see [Spot Instances](spot.md).

## Node Template Taint

| Key                                | Value                                    | Effect       | When Applied                                          |
| ---------------------------------- | ---------------------------------------- | ------------ | ----------------------------------------------------- |
| `scheduling.atlas.example/node-template` | template name (e.g. `default-by-atlas`) | `NoSchedule` | Applied to all nodes provisioned via a node template. |

## Scoped Autoscaler Taint

| Key                                    | Value    | Effect       | When Applied                                                                                                                                                |
| -------------------------------------- | -------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scheduling.atlas.example/scoped-autoscaler` | *(none)* | `NoSchedule` | Applied when the scoped autoscaler feature is enabled for a node template. Restricts scheduling to workloads explicitly intended for the scoped autoscaler. |

## Storage Optimization Taint

| Key                                    | Value    | Effect       | When Applied                                                                                                                                                           |
| -------------------------------------- | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scheduling.atlas.example/storage-optimized` | *(none)* | `NoSchedule` | Applied to storage-optimized nodes. Requires workloads to explicitly tolerate storage-optimized instances. See [Storage-optimized nodes](storage-optimized-nodes.md). |

## GPU / Accelerators Taints

| Key                     | Value  | Effect       | When Applied                                                                                                                                              |
| ----------------------- | ------ | ------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `nvidia.com/gpu`        | `true` | `NoSchedule` | Applied by the autoscaler during GPU node provisioning. Restricts scheduling to GPU-tolerant workloads, preventing non-GPU pods from consuming GPU nodes. |
| `nvidia.com/gpu.mig`    | `true` | `NoSchedule` | Applied by the autoscaler when provisioning nodes with MIG (Multi-Instance GPU) partitioning enabled. Restricts to MIG-compatible workloads.              |
| `aws.amazon.com/neuron` | `true` | `NoSchedule` | Applied to AWS Inferentia/Trainium nodes. Restricts scheduling to workloads that explicitly require Neuron accelerators.                                  |

For GPU workload configuration examples and toleration patterns, see [GPU Instances](gpu.md).

## Architecture Taint

| Key                  | Value   | Effect       | When Applied                                                                                                                        |
| -------------------- | ------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| `kubernetes.io/arch` | `arm64` | `NoSchedule` | Applied to ARM64 nodes. Requires pods to tolerate ARM64 architecture, preventing x86-only images from being scheduled on ARM nodes. |

## Eviction Taints (applied by Evictor)

| Key                        | Value    | Effect                            | When Applied                                                                                                                                                                                                                                                                                                                 |
| -------------------------- | -------- | --------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `evictor.atlas.example/evicting` | *(none)* | `NoSchedule`                      | Applied by Evictor when it begins draining a node. Blocks new pod scheduling while the node is being evacuated.                                                                                                                                                                                                              |
| `evictor.atlas.example/evicted`  | *(none)* | `NoSchedule` / `PreferNoSchedule` | Applied by Evictor after the node has been fully drained. With the default hard cordon flow, the effect is `NoSchedule`. When [soft tainting](evictor.md#soft-tainting) is enabled, the effect is `PreferNoSchedule`, allowing pods to schedule onto the node if no other capacity is available, until the node is deleted. |

For Evictor operating modes, override rules, and advanced configuration, see [Evictor](evictor.md).

## OMNI Edge Taint

| Key                                     | Value  | Effect      | When Applied                                                                                                                          |
| --------------------------------------- | ------ | ----------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `virtual-node.omni.atlas.example/not-allowed` | `true` | `NoExecute` | Applied to OMNI virtual nodes that are not allowed to run workloads. Evicts any existing pods in addition to blocking new scheduling. |

## Container Live Migration Taint (applied by CLM daemon)

| Key                      | Value    | Effect       | When Applied                                                                                                                                                                                                                                                                                                                                                           |
| ------------------------ | -------- | ------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `live.atlas.example/not-ready` | *(none)* | `NoSchedule` | Applied to CLM-enabled nodes during provisioning via a mutating admission webhook. Prevents scheduling until the CLM daemon has finished installing and initializing its components. Removed once all components report ready. The taint is removed even in degraded mode, where the node cannot perform live migrations but remains available for regular scheduling. |

For the CLM component lifecycle and the full list of CLM labels, annotations, and events, see [Labels, Annotations, Taints, and Events](clm-reference-labels-and-annotations.md).

# Annotations (related, node-level)

These annotations are not labels or taints but are closely related and set on nodes:

| Annotation                                    | Value                                        | Purpose                                                                                                                                                                                                                                                                                                                        |
| --------------------------------------------- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `autoscaling.atlas.example/removal-delay-seconds`   | integer                                      | Sets a removal delay in seconds for a node before it can be deleted.                                                                                                                                                                                                                                                           |
| `autoscaling.atlas.example/paused-draining-until`   | timestamp                                    | Pauses the draining process on a node until the given timestamp.                                                                                                                                                                                                                                                               |
| `rebalancing.atlas.example/status`                  | `drain-failed`                               | Set on a node when a drain operation during rebalancing has failed.                                                                                                                                                                                                                                                            |
| `rebalancing.atlas.example/uncordon-after`          | timestamp                                    | Set on drain-failed nodes when graceful eviction is enabled. Records the time after which the node will be automatically uncordoned. Default: 3 hours after the drain failure.                                                                                                                                                 |
| `evictor.atlas.example/eviction-status`             | varies; `done` when soft tainting is enabled | Tracks the current eviction status of a node. When [soft tainting](evictor.md#soft-tainting) is enabled, Evictor sets this to `done` after draining the node. The Atlas Autoscaler uses this value to identify nodes ready for deletion.                                                                                    |
| `predictions.atlas.example/remove-after`            | timestamp                                    | Set on nodes rebalanced due to ML spot interruption predictions. Indicates when the node should be removed.                                                                                                                                                                                                                    |
| `rebalancing.atlas.example/prevented-drain-pod-ids` | comma-separated UIDs                         | Set on a node when a drain operation fails during rebalancing. Contains the UIDs of normal pods that were on the node at the time of failure. These pods are treated as problematic for subsequent rebalancing plans. Uses UIDs (not names) so that recreated StatefulSet pods with the same name are not incorrectly flagged. |

<Accordion title="kubectl: Query nodes by annotation" icon="fa-terminal">
  ```shell
  # Find nodes with a removal delay configured
  kubectl get nodes -o json | jq -r '.items[] | select(.metadata.annotations["autoscaling.atlas.example/removal-delay-seconds"] != null) | .metadata.name'

  # Find nodes with paused draining
  kubectl get nodes -o json | jq -r '.items[] | select(.metadata.annotations["autoscaling.atlas.example/paused-draining-until"] != null) | "\(.metadata.name) paused until \(.metadata.annotations["autoscaling.atlas.example/paused-draining-until"])"'

  # Find nodes with failed drain operations
  kubectl get nodes -o json | jq -r '.items[] | select(.metadata.annotations["rebalancing.atlas.example/status"] == "drain-failed") | .metadata.name'
  ```
</Accordion>

For configuring paused drain behavior during rebalancing, see [Paused drain configuration](rebalancing-paused-drain-configuration.md).

## Zone spread and provisioning

This section details how the scheduler counts zone spread and how the autoscaler chooses where to provision nodes for spread-constrained pods. For the concepts these rules use (spread domain, eligible node, provisioning zone), see [Topology spread constraints and node provisioning](autoscaler.md#topology-spread-constraints-and-node-provisioning). To configure the constraints on a workload, see [Pod placement](pod-placement.md#highly-available-pod-scheduling).

### How the scheduler counts spread

1. The scheduler takes the nodes that exist and carry the `topology.kubernetes.io/zone` label.
2. It keeps only the nodes eligible for this pod, on top of the pod's own scheduling requirements. A zone with no eligible nodes is not counted at all.
3. In each zone, the scheduler counts the pods from the same namespace as the constraint-owning pod that match the constraint's `labelSelector`. The difference between the fullest and emptiest zone count must stay within `maxSkew`.
4. With `whenUnsatisfiable: DoNotSchedule`, zones that would break the limit are excluded, and the pod cannot run there. With `ScheduleAnyway`, the pod can still be placed there.

Two policies on the constraint decide which nodes the count includes:

* `nodeAffinityPolicy: Honor` (the default) counts only nodes matching the pod's `nodeSelector` or required `nodeAffinity`. `Ignore` counts all nodes.
* `nodeTaintsPolicy: Honor` counts only nodes whose taints the pod tolerates. `Ignore` (the default) counts all nodes.

For example, in a cluster with nodes in `us-east-1a`, `us-east-1b`, and `us-east-1c`, a pod without a node selector has three spread domains. A pod whose selector matches only nodes in `us-east-1a` and `us-east-1b` has two. For the full field reference, see [Pod topology spread constraints](https://kubernetes.io/docs/concepts/scheduling-eviction/topology-spread-constraints/#spread-constraint-definition) in the Kubernetes documentation.

### How the autoscaler picks a provisioning zone

1. Start from every zone configured on the cluster, then keep the zones that match the pod's zone `nodeSelector` or required `nodeAffinity`.
2. If the [Node Template](node-templates.md) restricts provisioning to specific AZs, only the overlap survives. If nothing overlaps, the pod is blocked. If the restriction removes zones, the autoscaler verifies that spreading into what remains can still satisfy `maxSkew`.
3. Zones bound to the pod's PVCs must remain allowed. A pod stays with its data.
4. On AWS clusters (EKS or self-managed Kubernetes on EC2), a node can be created only in a zone where the [Node Configuration](node-configuration.md) has a subnet. Zones without subnets are removed from the pod's target zones. On other providers, subnets do not restrict spreading.
5. The autoscaler counts same-namespace pods per zone on eligible nodes, assigns the pending pod to the least-populated allowed zone, and verifies the result against `maxSkew`.
6. The pod's zone node selector is set, and the autoscaler creates a node in that zone from the Node Template. The new node carries the template's labels and taints, so the pod schedules onto it.

Only zones that hold at least one node the pod could run on require subnet coverage. A zone whose existing nodes do not match the pod can never host it, so it does not demand a subnet and never blocks the pod. But because the Node Template cannot provision there, the autoscaler still removes it from the pod's target zones.

### Example

A cluster in `us-east-1` has five zones, and the existing nodes are labeled by the workload they run:

| Zone         | Existing nodes                      |
| ------------ | ----------------------------------- |
| `us-east-1a` | 3 nodes labeled `node_type=workers` |
| `us-east-1b` | 1 node labeled `node_type=workers`  |
| `us-east-1c` | nodes labeled `node_type=other`     |
| `us-east-1d` | nodes labeled `node_type=other`     |
| `us-east-1e` | nodes labeled `node_type=other`     |

A pending pod requires nodes labeled `node_type=workers` (required node affinity, with the default `nodeAffinityPolicy: Honor`). It also sets a hard zone spread constraint. The pod's Node Template is `workers-template`, and its Node Configuration has subnets in `us-east-1a`, `us-east-1b`, and `us-east-1c`. Two different zone sets come out, and both are correct:

| Zone         | Nodes the pod can run on today? | Counted by scheduler | Autoscaler may create a node here |
| ------------ | ------------------------------- | -------------------- | --------------------------------- |
| `us-east-1a` | yes                             | yes                  | **yes**                           |
| `us-east-1b` | yes                             | yes                  | **yes**                           |
| `us-east-1c` | no                              | no (zero-count)      | **yes**                           |
| `us-east-1d` | no                              | no                   | no                                |
| `us-east-1e` | no                              | no                   | no                                |

The scheduler counts two domains (`us-east-1a` and `us-east-1b`), the zones holding nodes the pod could run on today. The autoscaler may provision in three zones (`us-east-1a`, `us-east-1b`, and `us-east-1c`), the zones the Node Configuration covers. A node created by the autoscaler carries the Node Template's labels (`node_type=workers`), so the pod can run on it. Creating a node in `us-east-1c` is useful even though no such node exists there today: after the node exists, the zone is eligible and counted.

Because the zones holding runnable nodes (`us-east-1a` and `us-east-1b`) are fully covered by the Node Configuration's subnets, the pod is not blocked. The autoscaler places it in the least-populated allowed zone and creates a node there from the Node Template. If the pod's affinity had instead matched nodes in `us-east-1d`, a zone without a subnet in the Node Configuration, the pod would stay `Pending`; for the resulting event and resolution, see [FailedUpscaling: topology spread blocks node creation](autoscaler-troubleshooting.md#failedupscaling-topology-spread-blocks-node-creation).