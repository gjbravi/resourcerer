---
updatedAt: 2026-09-03T07:02:11.000Z
---

# GCP GKE

For a cross-cloud overview and comparison, see [Cloud providers](clm-cloud-providers.md). For TCP preservation modes in general, see [TCP migration modes](clm-reference-tcp-migration-modes.md). For details on the Traffic Control path, see [Traffic Control](clm-reference-traffic-control.md).

## What you need to know up front

GKE does not have a VPC-CNI equivalent for IP preservation. Unlike EKS, where the pod's original IP can be kept on the destination node, GKE's default CNI does not preserve pod IPs across migration. The only TCP preservation option on GKE is the **Traffic Control (TC)** path, which requires kernel 6.6+.

If TC is not enabled, the pod gets a new IP on the destination node and existing TCP connections drop.

> 🚧 Important
>
> Calico as a full CNI replacement is not available on GKE. Tigera's documentation for GKE only supports Calico as a network policy engine on top of GKE's native CNI, not as a full CNI replacement. GKE does not offer a "bring your own CNI" feature like AKS does. If you need IP preservation without TC, consider EKS with VPC CNI or Calico, or AKS with BYO CNI and Calico. See [AWS EKS](clm-cloud-providers-eks.md) and [Azure AKS](clm-cloud-providers-aks.md) for details.

***

## Supported node images

GKE supports two node image families: **Container-Optimized OS (COS)** and **Ubuntu**. Node images are configured through Atlas node configurations.

| Node image                   | Typical kernel                             | TC support (kernel 6.6+)                     | ARM64       | Supported by CLM |
| ---------------------------- | ------------------------------------------ | -------------------------------------------- | ----------- | ---------------- |
| Container-Optimized OS (COS) | 6.6 (milestone 121), 6.12 (milestone 125+) | Yes                                          | Yes (Axion) | Yes              |
| Ubuntu (GKE)                 | 5.15 on older versions                     | No on 5.15; 6.8+ is supported when available | Yes (Axion) | Yes              |

> 📘 Note
>
> GKE does not publish a simple mapping of GKE version to Ubuntu kernel version. If you use Ubuntu, verify the actual kernel on your nodes with `uname -r`. The safest path for TCP preservation is COS, which ships kernel 6.6+ on current GKE versions.

### Choosing the node image in Atlas

Node images are managed through **Atlas node configurations**. In the Atlas console, edit the node configuration linked to your node template and set the image family to COS or Ubuntu. New nodes provisioned by Atlas will use the selected image. Existing nodes are not re-imaged automatically — they are replaced as new nodes are provisioned.

***

## TCP preservation: Traffic Control (TC) path

TC is the only TCP preservation path available on GKE. It rewrites packets on peer nodes so the migrated pod's new IP is hidden from peers while the connection is still active. The pod gets a new IP; TCP traffic continues without interruption.

To use the TC path on GKE:

1. Make sure your nodes run kernel 6.6 or later. COS ships kernel 6.6+ on current versions. Ubuntu requires verification — check with `uname -r` on each node.

2. Enable the TC DaemonSet in the chart:

   ```yaml
   tc:
     enabled: true
   ```

3. Enable TCP preservation cluster-wide:

   ```yaml
   features:
     migrationOptions:
       default:
         withTCP: true
   ```

For full details on what TC does, how it works, and its limitations, see [Traffic Control](clm-reference-traffic-control.md).

> 📘 Note
>
> TC has no subnet or AZ constraint on GKE. Source and destination nodes can be in different subnets or availability zones. PVC zone constraints still apply — see [PVC and zone support](#pvc-and-zone-support).

### Dataplane compatibility

GKE clusters can run with two different dataplanes. The TC path works on both:

| Dataplane                 | Description                                                                            | TC path     |
| ------------------------- | -------------------------------------------------------------------------------------- | ----------- |
| **Dataplane V1** (legacy) | Uses GKE's built-in VPC-native CNI based on iptables.                                  | ✅ Supported |
| **Dataplane V2** (Cilium) | Uses Cilium with eBPF for networking and network policy. Default for new GKE clusters. | ✅ Supported |

The TC DaemonSet and Cilium (Dataplane V2) can coexist on the same nodes without conflict. The TC eBPF programs attach to the TCX ingress hook on network interfaces, while Cilium's eBPF programs operate at different layers — they do not interfere with each other.

GKE does not support IP preservation (pod keeping its original IP after migration) on either dataplane. The only TCP preservation option on GKE is the TC path.

***

## ARM64 (Axion)

GKE supports ARM64 via Google's Axion processors. C4A instances are generally available since October 2024. N4A instances are in preview. Both are supported by CLM.

| Axion instance | Status  | Supported by CLM |
| -------------- | ------- | ---------------- |
| C4A            | GA      | Yes              |
| N4A            | Preview | Yes              |

### Caveat: no iterative pre-dumps

ARM64 kernels lack the soft-dirty bit support required for incremental memory dumps. Iterative pre-dumps are disabled on ARM64, so the freeze time equals the full dump time. Migration still works correctly, but the freeze window is longer than on AMD64 nodes for the same workload.

### Cross-architecture restriction

A pod can only migrate between nodes of the same architecture. A pod running on an AMD64 node cannot migrate to an ARM64 node and vice versa. Use instance type constraints in your Atlas node template to keep workloads on a single architecture pool.

***

## PVC and zone support

CLM supports both zonal and regional persistent volumes on GKE.

* **Zonal PVs** — source and destination nodes must be in the same zone.
* **Regional (multi-zone) PVs** — destination can be any zone listed in the PV's allowed zones.

The compatibility check reads the standard `topology.kubernetes.io/zone` label, the GKE-specific `topology.gke.io/zone` label, and the legacy `failure-domain.beta.kubernetes.io/zone` label.

***

## GPU

GPU workloads are not currently supported. Pods that request GPU resources cannot be migrated. GPU support is planned, with no ETA.

***

## How to enable CLM

CLM is enabled on the node template in the Atlas console — see [How to enable CLM](clm-cloud-providers.md#how-to-enable-clm) for the canonical flow. On GKE, node-level configuration — image family (COS vs Ubuntu), instance types, architecture constraints — is managed through Atlas node configurations and node templates, not through the GKE console or `gcloud` commands.

***

## Reference: TCP preservation options on GKE

| Option              | Pod IP after migration | Setup                               | Protocols     | Kernel requirement |
| ------------------- | ---------------------- | ----------------------------------- | ------------- | ------------------ |
| TC path             | New IP                 | `tc.enabled: true`, `withTCP: true` | IPv4 TCP only | 6.6+               |
| No TCP preservation | New IP                 | Default                             | None          | None               |

### Quick decision guide

* **Use TC** if you need IPv4 TCP preservation and your nodes run kernel 6.6+ (COS works out of the box).
* **Skip TCP preservation** if your workloads are stateless and tolerate reconnects after migration.
* If you need IP preservation (pod keeps its original IP) and cannot use TC, GKE does not support this. Consider EKS or AKS with Calico as the full CNI — see [AWS EKS](clm-cloud-providers-eks.md) and [Azure AKS](clm-cloud-providers-aks.md).

# See also

<Cards columns={2}>
  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Cross-cloud comparison of node images, TCP preservation paths, ARM64, and PVC support.
  </Card>
  <Card title="Traffic Control" href="https://docs.atlas.example/docs/clm-reference-traffic-control" icon="fa-network-wired">
    Reference for the TC TCP preservation path used on GKE.
  </Card>
</Cards>