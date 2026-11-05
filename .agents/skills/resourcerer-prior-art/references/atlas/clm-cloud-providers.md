---
updatedAt: 2026-09-03T07:15:53.000Z
---

# Cloud providers

For per-cloud setup details, CNI configuration, and step-by-step instructions, see the dedicated pages:

* [AWS EKS](clm-cloud-providers-eks.md)
* [GCP GKE](clm-cloud-providers-gke.md)
* [Azure AKS](clm-cloud-providers-aks.md)

## How to enable CLM

The recommended way to enable CLM is through the Atlas console. When you enable CLM on a node template, the Atlas platform automatically installs the CLM components and labels nodes with `live.atlas.example/install: "true"`. The `live.atlas.example/migration-enabled` label is set internally by the daemon when each node is ready — do not set it manually.

Node-level configuration (image family, AMI, instance type, etc.) is managed through Atlas node configurations, not by running raw cloud CLI commands.

***

## TCP preservation paths by cloud

Each cloud supports different paths for preserving TCP connections during migration:

| Path                                                 | EKS               | GKE                 | AKS                                            | Pod IP after migration |
| ---------------------------------------------------- | ----------------- | ------------------- | ---------------------------------------------- | ---------------------- |
| **VPC CNI** (IP preserved)                           | ✅ Default         | ❌ Not available     | ❌ Not available                                | Same IP                |
| **Traffic Control (TC)** (packet rewriting)          | ✅ Kernel 6.6+     | ✅ Kernel 6.6+ (COS) | ✅ Kernel 6.6+ (Ubuntu 24.04 / Azure Linux 3.0) | New IP                 |
| **Calico as full CNI** (VXLAN overlay, IP preserved) | ✅ Replace VPC CNI | ❌ Not available     | ✅ Via BYO CNI                                  | Same IP                |

> 🚧 Important
>
> Calico as a network policy engine on top of the cloud's native CNI is NOT the same as Calico as the full CNI with VXLAN. IP preservation only works when Calico is the full CNI. See each cloud's dedicated page for installation instructions.

***

## Supported node images

| Cloud | Node image                           | Typical kernel                                 | TC support (6.6+)   | ARM64        | Supported by CLM          |
| ----- | ------------------------------------ | ---------------------------------------------- | ------------------- | ------------ | ------------------------- |
| EKS   | AL2023 standard                      | 6.1 (K8s 1.30–1.32), 6.12 (1.33+), 6.18 (1.36) | ✅ on K8s 1.33+      | ✅ Graviton   | ✅                         |
| EKS   | AL2023 NVIDIA/Neuron                 | 6.12 (K8s 1.33+)                               | ✅ Yes               | ✅ Graviton   | ✅                         |
| EKS   | Bottlerocket                         | 6.1 or 6.12                                    | ✅ on 6.12           | ✅ Graviton   | ❌ Not supported (planned) |
| GKE   | Container-Optimized OS (COS)         | 6.6 (m121), 6.12 (m125+)                       | ✅ Yes               | ✅ Axion C4A  | ✅                         |
| GKE   | Ubuntu                               | 5.15 or 6.8                                    | ❌ on 5.15, ✅ on 6.8 | ✅ Axion C4A  | ✅                         |
| AKS   | Ubuntu 22.04 (default K8s 1.25–1.34) | 5.15                                           | ❌ No                | ✅ Cobalt 100 | ✅                         |
| AKS   | Ubuntu 24.04 (default K8s 1.35+)     | 6.8                                            | ✅ Yes               | ✅ Cobalt 100 | ✅                         |
| AKS   | Azure Linux 3.0                      | 6.6                                            | ✅ Yes               | ✅ Cobalt 100 | ✅                         |

> Always verify the actual kernel with `uname -r` — image kernels change over time.

***

## ARM64 support

| Cloud | Processor            | Instance families    | Status  | Iterative pre-dumps             |
| ----- | -------------------- | -------------------- | ------- | ------------------------------- |
| EKS   | AWS Graviton         | m7g, c7g, r7g, t4g   | GA      | ❌ Disabled (freeze = full dump) |
| GKE   | Google Axion         | C4A                  | GA      | ❌ Disabled (freeze = full dump) |
| GKE   | Google Axion         | N4A                  | Preview | ❌ Disabled (freeze = full dump) |
| AKS   | Microsoft Cobalt 100 | Dpsv6, Dplsv6, Epsv6 | GA      | ❌ Disabled (freeze = full dump) |

**Iterative pre-dumps** reduce migration freeze time by taking successive memory snapshots while the pod continues running. Only the final dump freezes the process. On ARM64, this feature is disabled because ARM64 kernels lack the soft-dirty bit support required for incremental memory dumps. The migration still works correctly, but the freeze time equals the full memory dump time — which is longer for memory-hungry workloads.

***

## PVC and zone support (all clouds)

The system supports both zonal and regional persistent volumes on all clouds:

* **Zonal PVs** (single-zone): source and destination must be in the same zone.
* **Regional PVs** (multi-zone): destination can be in any zone listed in the PV's `NodeAffinity`.
* If the PV has no zone affinity, the migration is allowed regardless of zone.
* The system checks `topology.kubernetes.io/zone`, `topology.gke.io/zone`, and `failure-domain.beta.kubernetes.io/zone` labels.

***

## Comparison table

| Feature                              | EKS (AWS)                                | GKE (GCP)                                              | AKS (Azure)                              |
| ------------------------------------ | ---------------------------------------- | ------------------------------------------------------ | ---------------------------------------- |
| **Supported node OS**                | Amazon Linux 2023                        | COS, Ubuntu                                            | Ubuntu 22.04/24.04, Azure Linux 3.0      |
| **Bottlerocket**                     | ❌ Not supported (planned)                | N/A                                                    | N/A                                      |
| **Default CNI**                      | AWS VPC CNI                              | GKE VPC-native CNI (Cilium with Dataplane V2 optional) | Azure CNI                                |
| **TCP — VPC CNI path**               | ✅ Yes (IP preserved)                     | ❌ No                                                   | ❌ No                                     |
| **TCP — TC path**                    | ✅ on kernel 6.6+                         | ✅ on COS (kernel 6.6+)                                 | ✅ on Ubuntu 24.04 / Azure Linux 3.0      |
| **TCP — Calico full CNI (VXLAN)**    | ✅ Replace VPC CNI                        | ❌ Not available                                        | ✅ Via BYO CNI                            |
| **TC on default node image**         | ✅ (AL2023 K8s 1.33+ = 6.12)              | ✅ (COS = 6.6+)                                         | ❌ (Ubuntu 22.04 = 5.15)                  |
| **ARM64**                            | ✅ Graviton (no pre-dumps)                | ✅ Axion C4A (no pre-dumps)                             | ✅ Cobalt 100 (no pre-dumps)              |
| **Subnet constraint (VPC CNI path)** | ✅ Same subnet required                   | ❌ No VPC CNI path                                      | ❌ No VPC CNI path                        |
| **PVC zone constraint (all modes)**  | ✅ Zonal: same zone; Regional: multi-zone | ✅ Zonal: same zone; Regional: multi-zone               | ✅ Zonal: same zone; Regional: multi-zone |
| **GPU workloads**                    | ❌ Not supported                          | ❌ Not supported                                        | ❌ Not supported                          |
| **Container runtime**                | containerd                               | containerd                                             | containerd                               |

***

## Unsupported on all clouds

* **GPU workloads** — not currently supported. Pods requesting GPU resources (`nvidia.com/gpu`, `amd.com/gpu`, `intel.com/gpu`) cannot be migrated. GPU support is planned, with no ETA.
* **DaemonSets** — one pod per node; no destination to migrate to.
* **Cross-architecture migration** — amd64 to arm64 or vice versa is forbidden.
* **Windows nodes** — Linux only.
* **Non-containerd runtimes** — the CRI proxy requires containerd.

## Cloud-specific unsupported images

* **Bottlerocket (EKS)** — not detected by the daemon installer. Support is planned but has no ETA.
* **Amazon Linux 2 (EKS)** — no longer supported. End of support was November 2025.

# See also

<Cards columns={2}>
  <Card title="TCP Migration Modes" href="https://docs.atlas.example/docs/clm-reference-tcp-migration-modes" icon="fa-network-wired">
    Conceptual overview of the TCP preservation paths compared on this page.
  </Card>
  <Card title="Requirements and Limitations" href="https://docs.atlas.example/docs/clm-requirements-and-limitations" icon="fa-list-check">
    Infrastructure requirements and current limitations for container live migration.
  </Card>
</Cards>