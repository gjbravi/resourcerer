---
updatedAt: 2026-08-25T13:13:30.000Z
---

# GPU utilization

> 🚧 Limited Availability Feature
>
> This feature is currently available through feature flags. Contact us to enable access for your organization.

GPU utilization monitoring helps you track and optimize the usage of GPU resources across your Kubernetes workloads. This feature, which is part of Atlas's Cost Monitoring suite, provides detailed insights into GPU usage patterns, memory waste, and associated costs at both cluster and workload levels.

## Prerequisites

Before you can access GPU utilization metrics, ensure that:

1. Your cluster has GPU-enabled nodes
2. You have workloads running on GPU nodes
3. The Sentinel agent is installed with GPU metrics enabled (`--set agent.gpu.enabled=true`). See [GPU metrics collection](sec-configuring-sentinel.md#gpu-metrics-collection) for setup instructions.

## Accessing GPU metrics

Once the prerequisites are met, GPU utilization metrics become available in the Atlas console:

1. The cluster dashboard shows aggregate GPU metrics

<Image border={false} src="https://files.readme.io/2f846a8247a7794b6a61883389d71b3802214d627208878aa47ac2d0e2fb5e4b-image.png" />

2. The workloads report under cost monitoring provides detailed per-workload GPU metrics

<Image border={false} src="https://files.readme.io/38e4a73f5586b21542f482a377bb66a051ac96aebc7c0dc85aae570941774729-image.png" />

### Cluster dashboard metrics

The cluster dashboard displays key aggregate GPU metrics:

<Image align="center" border={false} width="35% " src="https://files.readme.io/8b471740ea17d152ec4b1ff7a93a871b25fab25c36c683a57cb9405098fa3d6e-image.png" />

The dashboard distinguishes between physical GPUs and virtual GPU allocations (vGPUs) to provide accurate resource tracking.

* **Provisioned GPUs**: Count of physical GPU devices in your cluster
* **Allocatable GPUs**: Total number of vGPUs available to Kubernetes based on your sharing configuration
* **Requested GPUs**: Number of vGPUs requested by workloads
* **Used GPUs**: Active vGPU utilization across the cluster

> 📘 Note
>
> With GPU sharing enabled, the **Requested** and **Allocatable** values may exceed the **Provisioned** count, as multiple workloads can share a single physical GPU. This is by design and reflects the resource multiplexing your GPU sharing configuration enabled.

In addition to the above, you will find a GPU utilization graph in the dashboard, which highlights average GPU and GPU Memory utilization across your selected time window:

<Image border={false} src="https://files.readme.io/669e4792da7954b18686c8815bf973f27e473a459af4be1d9d5d67c86100cbea-image.png" />

### Workload-level metrics

When you select a specific workload from the workloads report, you'll see detailed GPU metrics organized into three tabs: utilization, memory waste, and cost waste. Each view provides different insights into how your workload uses its GPU resources.

#### Utilization

The utilization view shows how effectively your workload uses its allocated GPU compute capacity. An interactive graph tracks the average GPU utilization over time, helping you identify patterns in resource usage.

<Image align="center" alt="Utilization tab" border={false} caption="Utilization tab" src="https://files.readme.io/07303c62148a2fc2402098a69de4396495142c6c508ab7133f901e02ce077b21-image.png" />

#### Memory waste

Switch to the memory waste tab to understand how your workload manages GPU memory. The view breaks down available versus used memory and visualizes memory usage trends. This information helps identify if you're over-provisioning GPU memory, which is directly impacting your costs. The graph shows both used and available memory over time, making it easy to spot opportunities for optimization.

<Image align="center" alt="Memory waste tab" border={false} caption="Memory waste tab" src="https://files.readme.io/e8593c4e926ce6edcb7c886119e359d8b70cab23c0b686faf64a0dd85defb640-image.png" />

#### Cost waste

The cost waste tab translates your GPU usage into financial terms. It breaks down costs by instance lifecycle: on-demand, spot, and fallback instances, and shows how these costs accumulate over time.

<Image align="center" alt="Cost waste tab" border={false} caption="Cost waste tab" src="https://files.readme.io/6647248b63e8a57b3e4605d218eec3a6e865298743a0c9daf1495f02f18910d7-image.png" />

<Cards columns={2}>
  <Card title="Workload costs" href="https://docs.atlas.example/docs/workloads" icon="fa-cubes">
    Break down cluster costs by individual workload.
  </Card>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure Sentinel for GPU metrics collection.
  </Card>
</Cards>