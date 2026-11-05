---
updatedAt: 2026-09-22T08:52:42.000Z
---

# Overview

Atlas Anywhere extends our cost optimization and workload automation capabilities beyond traditional cloud environments, enabling you to optimize Kubernetes clusters running on SUSE Rancher RKE2, Oracle Cloud Infrastructure, on-premises infrastructure, bare metal, private data centers, and other Kubernetes distributions.

# What is Atlas Anywhere?

Atlas Anywhere is a specialized deployment model that brings Atlas's core optimization features to any Kubernetes environment while requiring minimal permissions. It focuses on workload optimization and efficient resource utilization through two main capabilities:

* **Workload optimization** - automatically adjusts resource requests and limits based on actual usage patterns
* **Resource consolidation** - actively bin-packs workloads to reduce cluster size and eliminate waste

Unlike traditional Atlas deployments, which directly manage node provisioning, Atlas Anywhere works in conjunction with your existing autoscaling solution, delivering significant cost savings through advanced workload optimization.

# How does it compare to standard Atlas?

| Feature                                | Standard Atlas        | Atlas Anywhere                                                                            |
| -------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------- |
| **Node autoscaling and provisioning**  | Managed by Atlas      | Managed by your cloud provider or configured on-premises                                    |
| **Workload optimization**              | ✓                       | ✓                                                                                           |
| **Resource consolidation via Evictor** | ✓                       | ✓                                                                                           |
| **Kubernetes security**                | ✓                       | ✗                                                                                           |
| **Cost monitoring**                    | ✓                       | ✓\*                                                                                         |
| **Supported environments**             | AWS, GCP, Azure, Oracle | SUSE Rancher RKE2, VMware Tanzu, on-premises, bare metal, and other Kubernetes environments |
| **Spot Instance management**           | ✓                       | \*Depends on the environment                                                                |

# How does it work?

Atlas Anywhere consists of the following deployable components:

1. **Atlas Agent** - collects cluster metrics and state information
2. **Cluster Controller** - manages CRD creation and handles optimization logic
3. **Workload Autoscaler** - adjusts workload resources based on usage patterns
4. **Evictor** - handles workload consolidation and bin-packing

These components work together with your existing node autoscaler. To learn more about each component and its functions, refer to [Hosted components](hosted-components.md).

# Key benefits

## For platform teams

### Reduced operational overhead with automated workload optimization

Managing Kubernetes clusters across different environments creates significant operational complexity. Platform teams often struggle with the varying APIs, permissions, and infrastructure requirements of different cloud providers. Atlas Anywhere streamlines this by providing automated workload optimization that works seamlessly with your existing setup, regardless of where your clusters run.

### Works with existing infrastructure automation

Atlas Anywhere works alongside your existing infrastructure automation tools, whether you're using Terraform, Crossplane, or other solutions. This means you can maintain your current workflows while gaining advanced optimization capabilities.

### Unified optimization across different environments

Atlas Anywhere provides a unified approach to optimization across all your environments. Whether you're running SUSE Rancher RKE2 clusters, workloads in Oracle Cloud, on bare metal, or in private data centers, you get consistent optimization behavior and reliable results. This standardization helps reduce complexity and makes it easier to manage clusters at scale.

## For security teams

### Minimal required permissions

Security teams often face challenges when external tools require elevated cloud permissions or direct access to infrastructure. Atlas Anywhere addresses these concerns by operating with minimal permissions. Since it doesn't need to provision or manage infrastructure directly, you maintain full control over your environment's security boundaries.

### Audit-friendly deployment model

Security teams can easily verify and monitor the tool's operation with clear permission scopes and documented data flows. This transparency helps streamline security reviews and ensures compliance with the strictest internal policies.

## For finance teams

### Reduced Kubernetes costs through efficient resource utilization

Kubernetes infrastructure costs can quickly spiral out of control without proper optimization. Atlas Anywhere helps finance teams get these costs under control by automatically identifying and eliminating resource waste. By ensuring workloads use only the resources they need, you can significantly reduce Kubernetes spending across all environments.

### ROI visibility through detailed reporting

The platform offers comprehensive cost monitoring across all environments through a single, intuitive interface. This helps finance teams track spending trends, identify cost optimization opportunities, and make data-driven decisions about resource allocation. With detailed reporting capabilities, you can easily demonstrate the ROI of your optimization efforts and justify technology investments to stakeholders.

# Next steps

To get started with Atlas Anywhere, follow our <Anchor label="Getting Started guide" target="_blank" href="doc:atlas-anywhere-getting-started">Getting Started guide</Anchor>.