---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Storage cost

Atlas provides visibility into storage costs as part of its comprehensive cost-monitoring capabilities for Kubernetes clusters. This dimension lets you track and analyze the costs associated with persistent volumes (PVs) and ephemeral storage in your clusters.

![](https://files.readme.io/9b9f4c56c132aa62c98956f60b9a1031f6f395a8b43d36b053c7e94e763e2250-image.png)

## Overview

Storage costs can be a significant component of overall Kubernetes cluster expenses. Atlas's storage cost monitoring helps you:

* Track and visualize storage provisioning, requested capacity, and costs across your clusters.
* Identify storage over-provisioning and optimization opportunities.
* Break down storage costs by cluster, namespace, and workload.

## Supported storage

Atlas's storage cost monitoring feature is designed for the storage types described in the following sections. While other cloud providers or Kubernetes distributions are not currently supported, Atlas is committed to expanding support to other major cloud providers and storage types in future releases.

### GKE persistent volumes and ephemeral storage

Atlas supports storage cost monitoring features for:

* Google Kubernetes Engine (GKE) clusters running on Google Cloud Platform (GCP)
* GCP persistent disk types used as Kubernetes persistent volumes (PVs)
* Ephemeral storage provisioned on cluster nodes

> 📘 Note
>
> While ephemeral storage is included in totals, detailed cost and requested capacity analysis is only available for persistent volumes.

#### Supported persistent disk types

The following Google Cloud persistent disk types are supported:

| String        | Name                      | Category | Description                                                                                                            |
| :------------ | :------------------------ | :------- | :--------------------------------------------------------------------------------------------------------------------- |
| `pd-standard` | Standard persistent disk  | HDD      | Backed by standard hard disk drives (HDD).                                                                             |
| `pd-balanced` | Balanced persistent disks | SSD      | Backed by solid-state drives (SSD). They are an alternative to SSD persistent disks that balance performance and cost. |
| `pd-ssd`      | SSD persistent disk       | SSD      | Backed by solid-state drives (SSD).                                                                                    |

### Amazon AWS Elastic Block Store (EBS)

Default storage type when creating an Amazon EKS cluster. Atlas supports reporting data for EBS volumes when used as Kubernetes persistent volumes (PVs) in Amazon EKS clusters.

> 📘 Important limitations
>
> * Only provisioned capacity costs are currently monitored and reported.
> * Additional pricing components (IOPS and throughput) are not included in cost calculations.
> * Only persistent volumes are supported; ephemeral storage monitoring is not yet available for EKS clusters.

#### Supported persistent disk types

The following Amazon EBS volume types are supported for capacity cost reporting:

| Volume Type       | Name                                   | Category | Description                    |
| ----------------- | -------------------------------------- | :------- | ------------------------------ |
| io2 Block Express | Amazon EBS Provisioned IOPS SSD volume | SSD      | Highest performance SSD volume |
| io1               | Amazon EBS Provisioned IOPS SSD volume | SSD      | High-performance SSD           |
| gp3               | Amazon EBS General Purpose SSD volume  | SSD      | General purpose SSD            |
| gp2               | Amazon EBS General Purpose SSD volume  | SSD      | General purpose SSD            |
| st1               | Throughput Optimized HDD volume        | HDD      | Throughput-optimized HDD       |
| sc1               | Cold HDD volume                        | HDD      | Cold storage HDD               |

### Azure AKS managed disks

Atlas supports storage cost monitoring for Azure Managed Disks used as Kubernetes persistent volumes (PVs) in Azure Kubernetes Service (AKS) clusters.

Both in-tree (`kubernetes.io/azure-disk`) and CSI (`disk.csi.azure.com`) provisioners are supported.

> 📘 Important limitations
>
> * Only Azure Managed Disks are supported. Azure Files (`file.csi.azure.com`) are not currently supported.
> * Only provisioned capacity costs are monitored and reported.
> * Standard HDD (S-series) and Standard SSD (E-series) disks carry per-transaction charges (per 10K transactions) that are not included in cost calculations.
> * Premium SSD (P-series) pricing is exact since IOPS and throughput are bundled into the tier price.
> * Only persistent volumes are supported; ephemeral storage monitoring is not yet available for AKS clusters.

#### Supported managed disk types

The following Azure Managed Disk types are supported for capacity cost reporting:

| Storage Type      | Series   | Category | Description                                                                                                |
| :---------------- | :------- | :------- | :--------------------------------------------------------------------------------------------------------- |
| `Standard_LRS`    | S-series | HDD      | Standard HDD locally redundant storage. Cost-effective option for dev/test and infrequently accessed data. |
| `StandardSSD_LRS` | E-series | SSD      | Standard SSD locally redundant storage. Balanced price and performance for light workloads.                |
| `Premium_LRS`     | P-series | SSD      | Premium SSD locally redundant storage. High-performance, low-latency disks for production workloads.       |

Azure Managed Disks use a tiered pricing model where each disk is billed at the next tier that fits its provisioned size. For example, a 50 GiB `Standard_LRS` disk is billed at the S6 tier (64 GiB). Atlas automatically resolves the correct tier for cost calculations.

## Metrics tracked

Atlas collects and reports on the following storage metrics, with availability varying by cloud provider:

| Cloud provider | Provisioned persistent storage capacity | Claimed persistent storage capacity | Ephemeral storage | Storage costs          |
| :------------- | :-------------------------------------- | :---------------------------------- | :---------------- | :--------------------- |
| GCP (GKE)      | ✓                                       | ✓                                   | ✓                 | Persistent + Ephemeral |
| AWS (EKS)      | ✓                                       | ✓                                   | ✓                 | Persistent + Ephemeral |
| Azure (AKS)    | ✓                                       | ✓                                   | ✓                 | Persistent + Ephemeral |

> 📘 Note
>
> For AWS EKS and Azure AKS clusters, only provisioned capacity costs are included in storage cost calculations. IOPS, throughput, and transaction costs are not currently included (except for Azure Premium SSD, where these are bundled into the tier price).

## Understanding storage metrics

* **Provisioned storage**: Total storage capacity provisioned in the cluster (persistent + ephemeral).
* **Claimed storage**: Persistent storage capacity requested and bound to workloads.
* **Over-provisioning**: Difference between provisioned and claimed persistent storage.

Storage costs are calculated based on the provisioned capacity and pricing for the specific disk type.

## Viewing storage metrics

Storage metrics are integrated throughout the Atlas console reporting suite and available via the Atlas public APIs.

> 📘 Storage utilization prerequisite
>
> To view storage utilization data per node in the **Node list** page, install Sentinel with the `storage-stats-enabled` flag and the controller enabled. For configuration instructions, see [Configuring Sentinel features](sec-configuring-sentinel.md#kubernetes-storage-metrics).

### Cluster-level reports

#### Cluster dashboard

The cluster dashboard now includes a storage gauge showing the following:

* Total provisioned storage capacity (persistent + ephemeral)
* Amount of persistent storage requested by workloads

<Image align="center" alt="Storage widget in the Cluster Dashboard" caption="Storage widget in the Cluster Dashboard" src="https://files.readme.io/7cb21e410e9264ff3dff86e902460f2ea1980b25980a6716939d52eec2be577a-image.png" width="50%" />

* Storage utilization metrics (based on persistent storage claims vs. total provisioned storage)

<Image align="center" alt="Storage utilization widget in the Cluster Dashboard" caption="Storage utilization widget in the Cluster Dashboard" src="https://files.readme.io/ca9b1288a61b0991f8e476dbc47481af60397418e8b210eb478588c4339db1b7-image.png" />

#### Cluster cost

The Cluster compute cost report includes average storage costs per GiB, reflecting both persistent and ephemeral storage:

<Image align="center" alt="Average daily cost per resource, including storage" caption="Average daily cost per resource, including storage" src="https://files.readme.io/d62adea11fa5f5bb91a2cbf1002a500628172ec90ac7b332a4547935628c2cad-image.png" width="40%" />

Total cost metrics in the Cluster cost report now account for storage costs, which are included in the totals:

<Image align="center" alt="Daily compute spending details" caption="Daily compute spending details" src="https://files.readme.io/827384c0a896f04c2f69037f5b80c6949d14266ccbd9684c4c8ac73cbb089456-image.png" />

#### Cluster efficiency

The Efficiency tab has an at-a-glance Storage widget showing storage over-provisioning:

![](https://files.readme.io/39ffeae4e9383968239c055ea30dc65b41c5e82dca52e040718be26389d6567d-image.png)

As well as a dedicated Storage section showing:

* Graphs of provisioned vs. claimed storage capacity and costs over time.

![](https://files.readme.io/4fad6a56b91e4d39ef0f56ae250d5e4a963e36124b6d88e639ac8871197237d3-Frame_10380.jpg)

* A table of daily storage efficiency details, like cost per provisioned and requested storage, and overprovisioning information.

![](https://files.readme.io/0a8643284f0d65e8e4c9b23137b671f42dab1b3d61c14db08c727e7a29b09aee-image.png)

#### Workloads

Workload-level reports include:

* Storage claimed by each workload
* Associated storage costs

Detailed storage cost and requested capacity information at the workload level allow you to analyze storage requests for individual applications or services within your clusters.

The **Workloads** compute cost page provides a high-level summary of storage costs across all workloads:

![](https://files.readme.io/239159626010d6f06be0d027845246169e33c44c5c80d792bacb4b845b8d8d18-Group_3703.jpg)

**Key features:**

* Storage cost and capacity (GiB) for each workload
* Ability to filter and sort workloads based on storage requests

##### Individual workload view

Clicking on a specific workload provides a detailed view of its storage requests and costs:

1. **Cost overview**

   * Average daily cost per resource type, including storage:

   <Image align="center" alt="Average daily cost per resource, including storage" caption="Average daily cost per resource, including storage" src="https://files.readme.io/d62adea11fa5f5bb91a2cbf1002a500628172ec90ac7b332a4547935628c2cad-image.png" width="40%" />

2. **Provisioned storage cost**

   * Chart showing provisioned storage cost over time

   <Image align="center" alt="Provisioned storage cost over time" caption="Provisioned storage cost over time" src="https://files.readme.io/04575267d1395a636d6a9f289859c74fc326f4944771e75eaa5c3a717d052d24-image.png" width="50%" />

3. **Daily compute cost details**
   * Detailed table showing daily storage costs
   * Includes storage capacity (GiB) provisioned each day
   * Cost per GiB of storage

<Image align="center" alt="Daily compute cost details" caption="Daily compute cost details" src="https://files.readme.io/13f8b32a44b6d0cc97b72e0b19365412ed5136b5e112fd275d5138154ac4b0ca-Group_3704.jpg" />

#### Namespaces

Detailed storage cost and requested capacity information at the namespace level allow you to analyze storage requests across different groupings within your clusters.

The Namespaces overview page provides a summary of storage costs across all namespaces:

* Storage cost and requested capacity (GiB) for each namespace
* Detailed list of namespaces with storage metrics
* Ability to filter and sort namespaces based on storage requests

<Image align="center" alt="Cost breakdown for namespaces on a cluster" caption="Cost breakdown for namespaces on a cluster" src="https://files.readme.io/a36fa18684ee0f1e168f5166dc7ccc33811ba8009136cdef15f53329045d4f27-image.png" />

##### Individual namespace view

Clicking on a specific namespace provides a detailed view of its storage requests and costs.

### Organization-level reports

Organization-wide views allow you to analyze storage costs and capacity requests across all clusters in your organization. These reports offer a higher-level perspective on your Kubernetes infrastructure costs.

The organization cluster cost report gives you a comprehensive view of storage costs across all your clusters:

**Key features:**

* At-a-glance total storage cost for all clusters in the organization
* Storage capacity (GiB) provisioned across all clusters

<Image align="center" alt="Total storage cost for all clusters in the right-most widget" caption="Total storage cost for all clusters in the right-most widget" src="https://files.readme.io/55080d18b6e63433e09264e6efe34ba0c3692102174d027800fe170bd04d2a87-image.png" />

* Storage capacity and cost breakdown by cluster

![](https://files.readme.io/9f443fa80fae280caa493583f0edc394a6c9f7e6a8865768d0a70d5cc8a8e894-image.png)

This breakdown helps you identify which clusters consume the most storage resources and incur the highest costs, allowing for targeted optimization efforts.

## Recommendations and best practices

To optimize storage costs:

1. Right-size your persistent volume claims (PVCs) based on actual requested and claimed capacity observed by leveraging Atlas's reports.
2. Use storage classes with the appropriate performance characteristics for your use case.
3. Clean up unused PVs and PVCs.

By leveraging Atlas's storage cost monitoring, you can gain deeper insights into your Kubernetes storage expenses and identify opportunities for optimization across both persistent and ephemeral storage resources.

<Cards columns={2}>
  <Card title="Network cost" href="https://docs.atlas.example/docs/network-cost" icon="fa-network-wired">
    Monitor and attribute Kubernetes network costs.
  </Card>
  <Card title="GPU utilization" href="https://docs.atlas.example/docs/gpu-utilization" icon="fa-microchip">
    Track GPU utilization metrics across your workloads.
  </Card>
</Cards>