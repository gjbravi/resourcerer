---
updatedAt: 2026-10-30T12:30:58.000Z
---

# Edge configuration

An **Edge Configuration** tells Atlas how to build the underlying compute instance for an edge node in a specific edge location. It covers settings that live below the Kubernetes layer — the boot disk, the machine image, the container runtime socket, and cloud instance tags — while [node templates](node-templates.md) continue to control Kubernetes-level placement (constraints, labels, taints, and which edge locations are eligible).

This is conceptually the edge equivalent of a [Node Configuration](node-configuration.md), scoped to a single edge location.

## Before you begin

* OMNI must be enabled on your cluster. See [Getting started with OMNI](omni-getting-started.md).
* You need at least one edge location already created. See [Step 2: Create and onboard an edge location](omni-getting-started.md#step-2-create-and-onboard-an-edge-location).

## Edge Configurations and edge locations

Each Edge Configuration belongs to a single edge location, and Atlas creates a default configuration automatically with each location. You can create additional configurations, for example one with a hardened custom image for production workloads and one with a stock distro image for testing. You mark one configuration as the **default**, and Atlas uses it whenever a node request does not name a specific configuration.

| Concept            | Scope                                   | Configured by                                                  |
| ------------------ | --------------------------------------- | -------------------------------------------------------------- |
| Edge location      | Cluster + region/provider               | You, during onboarding                                         |
| Edge Configuration | A single edge location                  | You, as described below                                        |
| Node template      | Cluster-wide, references edge locations | You, per [Getting started with OMNI](omni-getting-started.md) |

## Create an Edge Configuration

1. In the Atlas console, open the edge location you want to configure and select **Create edge node configuration**.
2. Fill in the fields in the drawer:

| Field                     | Description                                                                                                                                                                                                                                                                                                                    |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Configuration name**    | A unique, descriptive name (for example, `compute-optimized`). **This cannot be changed after saving**, so name it deliberately.                                                                                                                                                                                               |
| **Edge location**         | The edge location this configuration applies to. To configure additional edge locations, select them here and adjust their settings.                                                                                                                                                                                           |
| **Instance tags**         | Key/value pairs applied to every Atlas–provisioned instance built from this configuration. Use **+ Add** to add more than one pair. These map to cloud-provider tags/labels (AWS tags, GCP labels, etc.) on the underlying instance.                                                                                         |
| **Boot disk → Disk size** | The boot disk size, in GiB, for instances built from this configuration.                                                                                                                                                                                                                                                       |
| **Image**                 | The machine image Atlas should use when building the instance — for example, a GCP image path like `projects/debian-cloud/global/images/debian-12-bookworm-v20240515`, or the equivalent AWS AMI ID / OCI image OCID for other providers. Leave blank to use Atlas's default image for the edge location's cloud provider. |
| **CRI socket path**       | Path to the Container Runtime Interface socket on the node (for example, `/run/containerd/containerd.sock`). Only change this if your image uses a non-default container runtime or socket location.                                                                                                                           |

3. Save the configuration.

> 📘 Note
>
> Configuration names are permanent once saved. If you need to change other settings later, you can edit the existing configuration's other fields, or create a new configuration and switch your node template or manually created edge nodes to reference it instead.

## Using an Edge Configuration

Once saved, an Edge Configuration is used automatically:

* If it's marked as the **default** for its edge location, it's used for any edge node Atlas provisions there — whether through the Autoscaler or through [manual edge provisioning](omni-manual-edge-provisioning.md) — unless another configuration is explicitly requested.
* If it's not the default, it's used only when explicitly selected — for example, when [manually provisioning an edge node](omni-manual-edge-provisioning.md) and choosing a specific configuration.

## Managing multiple configurations

To see all configurations for an edge location, open the edge location's detail page in the console. From there you can:

* Review each configuration's settings.
* Change which configuration is marked as **default**.
* Create additional configurations for different use cases (for example, GPU-optimized vs. general-purpose instances, or different images for production vs. testing).

## Related documentation

* [OMNI Overview](omni-overview.md)
* [Getting started with OMNI](omni-getting-started.md)
* [Custom edge locations](omni-custom-edge-location.md)
* [Node configuration](node-configuration.md)