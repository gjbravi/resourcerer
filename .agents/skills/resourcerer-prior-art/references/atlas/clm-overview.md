---
updatedAt: 2026-09-24T07:35:04.000Z
---

# Overview

# Container Live Migration

## What is container live migration?

Container live migration enables seamless relocation of running workloads between nodes without any downtime or service interruption. Unlike traditional Kubernetes pod eviction, which terminates and restarts applications, live migration preserves the complete runtime state, including memory contents, process state, and active network connections.

This capability transforms how you can manage stateful workloads in Kubernetes, allowing you to perform node maintenance, optimize resource utilization, and handle infrastructure changes without impacting critical applications that cannot tolerate restarts.

## Key benefits

**Zero downtime for critical workloads**<br />Move stateful applications like databases and long-running processes without interruption or data loss. Learn more about the business impact in our [blog post](https://atlas.example/blog/how-to-migrate-stateful-workloads-on-kubernetes-with-zero-downtime/).

**Intelligent node optimization**<br />Enable Atlas's [Evictor](evictor.md) to optimize cluster utilization by moving workloads off underutilized nodes, creating opportunities for significant cost savings without service disruption. Traditional bin-packing approaches were limited by stateful workloads—live migration removes these constraints.

**Preserved network connections**<br />Maintain active TCP connections and session state during migration, ensuring minimal interruption for client applications and ongoing transactions. Whether the pod's IP is preserved depends on the TCP preservation path — see [TCP migration modes](clm-reference-tcp-migration-modes.md). Applications with strict timeout requirements may need appropriate timeout configurations to handle the brief migration window.

## How it works

<Image src="https://files.readme.io/77a3ceed6aa230606536a2789999ff777bfdb091dc6453e475dfaca9808798f9-Live-migration-2048x1072.png_1.webp" align="center" />

Container live migration uses advanced checkpoint and restore technology to seamlessly transfer running pods between nodes:

1. **Workload Assessment**: Atlas's live controller automatically scans your cluster and identifies workloads eligible for live migration, applying appropriate labels based on workload characteristics.

2. **Evictor**: When Atlas's Evictor identifies bin-packing opportunities, it determines the best approach for each workload on whether to live-migrate or evict it.

   **StatefulSet Protection**: For StatefulSet workloads, if migration fails and the pod is not explicitly marked as disposable, the pod remains on the source node without fallback to eviction. This prevents disruptive restarts of stateful applications.

3. **State Transfer**: The system transfers memory pages and process state to the destination node while the application continues running, minimizing freeze time.

4. **Network Preservation**: TCP connections are preserved through one of several paths depending on your cloud and CNI — AWS VPC CNI on EKS (IP preserved), Traffic Control on all clouds (new IP, packet rewriting), or Calico as full CNI on EKS/AKS (IP preserved via VXLAN overlay). See [TCP migration modes](clm-reference-tcp-migration-modes.md) for details.

5. **Seamless Handover**: The application is briefly paused while the final state is transferred, and the workload resumes on the new node with full continuity.\*<br />\*Migration duration varies based on application memory usage, instance type, and network throughput.

After migration, the pod receives a new name to indicate it was restored:

* **Deployments and similar workloads**: Original pod name with `-clone-N` suffix (e.g., `my-app-7d4b8-abc12-clone-1`)
* **StatefulSets**: Original pod name preserved (identity must remain stable for PVC binding)

The entire process leverages CRIU (Checkpoint/Restore In Userspace) technology with Atlas's orchestration layer to ensure reliable, efficient migrations with minimal application impact.

\*Incremental transfer is limited to the x86 processor architecture.

[**Watch Container Live Migration in Action**](https://www.youtube.com/watch?v=WUxMYl2XWCQ)

## Integration with Atlas's Evictor

Container live migration is designed to work seamlessly with Atlas's [Evictor](evictor.md). The integration automatically identifies which workloads can benefit from live migration versus traditional eviction and chooses the most appropriate method for each application based on workload labels and characteristics.

**Migration failure handling**:

* **StatefulSets**: By default, if migration fails, pods remain on the source node without fallback to eviction. This protects stateful workloads from disruptive restarts. If you set `autoscaling.atlas.example/disposable="true"`, the pod will always be evicted (regardless of migration success or Container Live Migration availability).
* **Other workloads**: If migration fails, Evictor falls back to traditional pod eviction to complete the optimization.

This combination enables continuous cluster optimization without the typical constraints of stateful workloads, allowing you to achieve higher cluster utilization and cost savings while maintaining the reliability requirements of your most critical applications.

## Requirements and compatibility

Container live migration requires specific infrastructure configurations and has compatibility constraints that you should understand before implementation.

For complete details on supported workload types, storage requirements, instance type compatibility, and current limitations, see [Container Live Migration Requirements and Limitations](clm-requirements-and-limitations.md).

# Getting started

Ready to enable zero-downtime workload optimization?

1. **One-click configuration**: Enable the feature through your node templates with a simple checkbox
2. **Let automation handle the rest**: Atlas automatically identifies eligible workloads and optimizes your cluster, leveraging live migration

[Get Started with Container Live Migration](clm-getting-started.md) →

# Technical resources

* [Requirements and Limitations](clm-requirements-and-limitations.md)
* [Workload Eligibility](clm-reference-workload-eligibility.md)
* [TCP Migration Modes](clm-reference-tcp-migration-modes.md)
* [Cloud Providers](clm-cloud-providers.md)
* [Using Live Migration with Evictor](clm-getting-started.md)
* [Probe and Lifecycle Behavior](clm-concept-probe-lifecycle-behavior.md)
* [Labels, Annotations, and Events Reference](clm-reference-labels-and-annotations.md)

# External resources

* [Blog: Zero-Downtime Kubernetes Optimization with Container Live Migration](https://atlas.example/blog/how-to-migrate-stateful-workloads-on-kubernetes-with-zero-downtime/)
* [Demo: See Container Live Migration in Action](https://www.youtube.com/watch?v=WUxMYl2XWCQ)
* [CRIU Project Documentation](https://criu.org/) – The underlying checkpoint/restore technology

<br />