---
updatedAt: 2026-09-03T07:02:11.000Z
---

# Azure AKS

For a side-by-side comparison with EKS and GKE, see [Cloud providers](clm-cloud-providers.md). For TCP preservation details, see [TCP migration modes](clm-reference-tcp-migration-modes.md) and [Traffic Control](clm-reference-traffic-control.md). For general prerequisites, see [Overview](clm-overview.md).

## What is supported on AKS

CLM supports the following on AKS:

* **Node images** — Ubuntu 22.04, Ubuntu 24.04, and Azure Linux 3.0. The default AKS image depends on the Kubernetes version (see [Supported node images](#supported-node-images)).
* **TCP preservation** — two paths: the TC path and the Calico-as-full-CNI path via AKS Bring Your Own CNI. TC is the default when the node image supports kernel 6.6+ (Ubuntu 24.04 or Azure Linux 3.0); on Ubuntu 22.04 (kernel 5.15), TC is not available and the cluster falls back to no TCP preservation unless Calico BYO CNI is configured.
* **ARM64** — Microsoft Cobalt 100 processors (Dpsv6, Dplsv6, Epsv6 series).
* **PVC migration** — Azure Disk CSI driver (`disk.csi.azure.com`), both zonal and regional PVs.

What is **not** supported on AKS:

* **GPU workloads** — not currently supported. GPU support is planned, with no ETA.
* **Windows nodes** — Linux only.
* **Azure Linux 2.0 (CBL-Mariner)** — the daemon installer does not reliably detect this image. Use Azure Linux 3.0 instead.

## How to enable CLM

CLM is enabled on the node template in the Atlas console — see [How to enable CLM](clm-cloud-providers.md#how-to-enable-clm) for the canonical flow. On AKS, node configuration (image family, instance type, etc.) is managed through Atlas node configurations. To change the node image after a cluster is created, update the image family on the node configuration. Available AKS image families in Atlas are `ubuntu`, `ubuntu2204`, `ubuntu2404`, and `azure-linux`.

You can also install CLM manually via Helm on nodes you label yourself, but most users should use the Atlas console for the full platform experience.

## Supported node images

The table below lists AKS node images that CLM supports. Image kernels change over time, so always verify the actual kernel on your nodes with `uname -r`.

| Node image                    | Default for K8s version | Kernel | TC support (requires kernel 6.6+) | ARM64             |
| ----------------------------- | ----------------------- | ------ | --------------------------------- | ----------------- |
| Ubuntu 22.04                  | K8s 1.25–1.34           | 5.15   | ❌ No                              | ✅ Cobalt 100      |
| Ubuntu 24.04                  | K8s 1.35+               | 6.8    | ✅ Yes                             | ✅ Cobalt 100      |
| Azure Linux 3.0               | available on K8s 1.30+  | 6.6    | ✅ Yes                             | ✅ Cobalt 100      |
| Azure Linux 2.0 (CBL-Mariner) | legacy                  | 5.15   | ❌ No                              | ❌ Not recommended |

> 🚧 Important
>
> Ubuntu 22.04 is the most common AKS image — it ships kernel 5.15, which does **not** support TC. If you need TCP preservation, switch node pools to Ubuntu 24.04 or Azure Linux 3.0. Azure Linux 2.0 is not recommended because only Azure Linux 3.0 is reliably detected by the installer.

> 📘 Note
>
> Cobalt 100 on Ubuntu 22.04: Ubuntu 22.04's kernel 5.15 includes backported errata patches for Azure Cobalt 100 processors, so ARM64 Cobalt 100 VMs are fully supported on this image despite the older kernel version.

To change the node image, update the image family in the Atlas node configuration. After new nodes are provisioned, verify the kernel with `uname -r` on a node.

## TCP preservation

AKS does not have a built-in equivalent to the AWS VPC CNI for IP preservation. There are two TCP preservation paths on AKS:

* **TC path** — the default option. Requires kernel 6.6+ (Ubuntu 24.04 or Azure Linux 3.0).
* **Calico-as-full-CNI path** — uses Calico as the cluster CNI with VXLAN overlay. Works on any AKS node image because there is no kernel 6.6+ requirement.

Both paths are explained below.

### TC path (default)

The TC path installs a privileged DaemonSet that translates packets on peer nodes so that established TCP connections survive migration. For full details, see [Traffic Control](clm-reference-traffic-control.md).

On AKS:

* Requires `tc.enabled: true` and kernel 6.6+ nodes (Ubuntu 24.04 or Azure Linux 3.0).
* No subnet or availability zone constraint.
* IPv4 TCP only.

On Ubuntu 22.04 nodes (kernel 5.15), TC is not available. Migrated pods get a new IP and existing TCP connections drop.

### Calico-as-full-CNI path via AKS Bring Your Own CNI

AKS supports a **Bring Your Own CNI** (BYO CNI) mode where you create a cluster with no CNI preinstalled and then install Calico as the full CNI (not just a network policy plugin). When Calico is the cluster CNI with VXLAN overlay, it preserves pod IP across migration — the pod keeps its IP after the move.

This path is useful when:

* Your AKS node pools use Ubuntu 22.04 (kernel 5.15) and you cannot upgrade them to Ubuntu 24.04.
* You want to preserve IPv4 traffic other than TCP, because the real pod IP is unchanged.
* You want to preserve connections without the TC kernel 6.6+ requirement.

> 🚧 Important
>
> AKS can run Azure CNI with Calico as a **network policy engine only** — in that setup Azure CNI still assigns pod IPs and IP preservation does **not** apply. IP preservation works only when Calico is the **full CNI** with VXLAN via the BYO CNI flow. For the full distinction, see [Calico as full CNI on the hub](clm-cloud-providers.md#tcp-preservation-paths-by-cloud).

#### Installing Calico as the full CNI on AKS

1. Follow the [official documentation](https://docs.tigera.io/calico/latest/getting-started/kubernetes/managed-public-cloud/aks#install-aks-with-self-managed-calico-for-networking-and-network-policy) to create a cluster with no CNI preinstalled.

2. In the CLM Helm chart, enable TCP preservation with the IP-preservation path:

```yaml
tc:
  enabled: false
features:
  migrationOptions:
    default:
      withTCP: true
```

> 📘 Note
>
> Operational note: With BYO CNI, IP address management, routing, and scaling are determined by the Calico CNI plugin, not by AKS. The lifecycle of the CNI is managed outside the AKS cluster lifecycle.

The combination of `withTCP=true` with `tc.enabled=false` selects the IP-preservation path, but actual pod IP preservation requires a compatible CNI that supports IP pinning — such as Calico as the full CNI with VXLAN via AKS BYO CNI. On a default AKS cluster using Azure CNI without Calico, this configuration does **not** preserve pod IP or TCP connections. Calico with VXLAN works on any AKS node image because there is no kernel 6.6+ requirement — Calico handles IP preservation without the TC/eBPF subsystem.

## ARM64 (Cobalt 100)

AKS supports ARM64 through Microsoft's Cobalt 100 processors. Cobalt 100 VMs are available in the Dpsv6, Dplsv6, and Epsv6 series and have been generally available since October 2024.

ARM64 nodes work with CLM with one caveat:

* **Iterative pre-dumps are disabled.** ARM64 kernels lack the soft-dirty bit support that iterative pre-dumps rely on. Migration still works correctly, but the freeze time equals the full dump time instead of converging toward a small final dump.

Migration between same-architecture nodes only is enforced (amd64 to arm64 and vice versa is forbidden). This applies across all clouds.

## PVC and zone support

CLM supports both zonal and regional persistent volumes on AKS.

* **Zonal PVs** — source and destination nodes must be in the same zone.
* **Regional (multi-zone) PVs** — the destination node can be in any of the zones listed in the PV's allowed zones.
* **PVs with no zone affinity** — the migration is allowed regardless of zone.

The Azure Disk CSI driver (`disk.csi.azure.com`) is supported. During migration, the system detaches the disk from the source node and attaches it to the destination node.

The system checks the standard `topology.kubernetes.io/zone` label and the legacy `failure-domain.beta.kubernetes.io/zone` label on nodes.

## GPU workloads

GPU workloads are not currently supported. Pods requesting GPU resources (`nvidia.com/gpu`, `amd.com/gpu`, `intel.com/gpu`) cannot be migrated. GPU support is planned, with no ETA.

## Comparing TCP preservation options on AKS

| Dimension                   | TC path                                                                    | Calico-as-full-CNI path                                                                                                  |
| --------------------------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Node image requirement**  | Ubuntu 24.04 or Azure Linux 3.0 (kernel 6.6+)                              | Any supported AKS node image                                                                                             |
| **Kernel requirement**      | Linux kernel 6.6+ with eBPF / TCX                                          | None beyond standard CNI requirements                                                                                    |
| **How pod IP is preserved** | Pod gets a new IP; peer nodes rewrite packets                              | Pod keeps its original IP because Calico/CNI re-assigns it                                                               |
| **Protocols**               | IPv4 TCP only                                                              | IPv4 protocol-agnostic (because the real pod IP is unchanged)                                                            |
| **Setup complexity**        | Set `tc.enabled: true` and `withTCP: true`; deploys a privileged DaemonSet | Create AKS with `--network-plugin none`; install Tigera Operator and Calico; set `withTCP: true` and `tc.enabled: false` |
| **Subnet / AZ constraint**  | None                                                                       | None                                                                                                                     |
| **Best for**                | Lower freeze times; only need IPv4 TCP preserved                           | Ubuntu 22.04 node pools; want protocol-agnostic preservation without the kernel 6.6+ requirement                         |

### Quick decision guide

* **Use the TC path** if your node pools already run Ubuntu 24.04 or Azure Linux 3.0, you only need IPv4 TCP preserved, and you want lower freeze times.
* **Use the Calico-as-full-CNI path** if your node pools run Ubuntu 22.04 (kernel 5.15) and you cannot upgrade, or if you need to preserve non-TCP IPv4 traffic.

## Caveats and gotchas

* **Default AKS image is Ubuntu 22.04.** New AKS clusters typically ship Ubuntu 22.04 with kernel 5.15, which does not support TC. Either switch the image family to `ubuntu2404` or `azure-linux` for new node pools, or use the Calico-as-full-CNI path if you need TCP preservation on Ubuntu 22.04.
* **Azure Linux 2.0 detection.** The CLM installer does not reliably detect Azure Linux 2.0 (CBL-Mariner 2.0). Use Azure Linux 3.0 for new node pools.
* **No VPC CNI equivalent for IP preservation.** AKS does not ship a built-in CNI path that preserves the pod IP across migration. You must choose between TC and Calico-as-full-CNI.
* **GPU workloads** are not currently supported. GPU support is planned, with no ETA.
* **PVC zone constraints still apply.** The TC path has no subnet or AZ constraint, but PVC zone constraints (same zone for zonal PVs, allowed zones for regional PVs) still apply. See [PVC and zone support](#pvc-and-zone-support) above.
* **ARM64 iterative pre-dumps are disabled** — see [ARM64 (Cobalt 100)](#arm64-cobalt-100).

## External references

* Microsoft Azure documentation on [Bring Your Own CNI](https://learn.microsoft.com/en-us/azure/aks/use-byo-cni) for AKS.
* Tigera Calico documentation on [installing Calico on AKS](https://docs.tigera.io/calico/latest/getting-started/kubernetes/managed-public-cloud/aks#install-aks-with-self-managed-calico-for-networking-and-network-policy).

# See also

<Cards columns={2}>
  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Cross-cloud comparison of node images, TCP preservation paths, ARM64, and PVC support.
  </Card>
  <Card title="Traffic Control" href="https://docs.atlas.example/docs/clm-reference-traffic-control" icon="fa-network-wired">
    Reference for the TC TCP preservation path, also available on AKS.
  </Card>
</Cards>