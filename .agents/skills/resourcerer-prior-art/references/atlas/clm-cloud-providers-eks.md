---
updatedAt: 2026-09-03T07:02:11.000Z
---

# AWS EKS

For the cross-cloud comparison and what GKE/AKS support, see [Cloud providers](clm-cloud-providers.md). For TCP preservation in general, see [TCP migration modes](clm-reference-tcp-migration-modes.md). For deeper detail on each TCP path, see [AWS VPC CNI](clm-reference-aws-vpc-cni.md) and [Traffic Control](clm-reference-traffic-control.md).

## Supported node images

CLM detects the node OS by its image identifier. On EKS, the only supported node image today is **Amazon Linux 2023 (AL2023)**. Other AMIs are detected but blocked.

| Node image                        | Supported | Notes                                                                                      |
| --------------------------------- | --------- | ------------------------------------------------------------------------------------------ |
| Amazon Linux 2023 standard        | Yes       | Kernel varies by K8s version — see below                                                   |
| Amazon Linux 2023 NVIDIA / Neuron | Yes       | K8s 1.33+ only. NVIDIA and Neuron are distinct AMIs with the same K8s version requirement. |
| Bottlerocket                      | No        | Support planned, no ETA                                                                    |
| Amazon Linux 2                    | No        | No longer supported (end of support: November 2025)                                        |

### AL2023 kernel by Kubernetes version

The kernel that ships with AL2023 depends on the Kubernetes version selected for the node:

| K8s version | AL2023 standard kernel | AL2023 NVIDIA/Neuron kernel |
| ----------- | ---------------------- | --------------------------- |
| 1.30 – 1.32 | 6.1                    | —                           |
| 1.33 – 1.35 | 6.12                   | 6.12                        |
| 1.36+       | 6.18                   | 6.18                        |

> 📘 Note
>
> Always verify the actual kernel on the node with `uname -r`. AWS updates AMI kernels over time, and the table above reflects what shipped at the time of writing. The kernel matters because the Traffic Control (TC) path requires kernel 6.6 or later — see [TCP preservation](#tcp-preservation) below.

## TCP preservation

EKS has the most extensive TCP connection preservation options. The choice matters because each path has different kernel requirements, different subnet/AZ constraints, and different freeze-time characteristics.

### The three paths

| Path                              | Pod IP after migration            | Kernel requirement | Subnet / AZ constraint                                                        | Typical freeze time                              |
| --------------------------------- | --------------------------------- | ------------------ | ----------------------------------------------------------------------------- | ------------------------------------------------ |
| **AWS VPC CNI** (default)         | Same IP (secondary IP reassigned) | None               | Source and destination in the same subnet (or same AZ with custom networking) | Longer — waits for AWS networking to reassign IP |
| **Traffic Control (TC)**          | New IP (peers rewrite packets)    | 6.6 or later       | None                                                                          | Shorter                                          |
| **Calico as full CNI with VXLAN** | Same IP (annotation pinned)       | None               | None                                                                          | Similar to VPC CNI path                          |

> 💡 Decision guide
>
> If you do not change anything, the chart autodetects the AWS VPC CNI and uses the VPC CNI path — that is the right default for most EKS clusters. Switch to **TC** when you want shorter freeze times and only need IPv4 TCP preserved. Switch to **Calico as full CNI** when you want VPC-CNI-style IP preservation without the subnet constraint and are willing to operate Calico yourself.

The following subsections summarize each path. The configuration reference, autodiscovery rules, and known limitations live in the linked docs.

### AWS VPC CNI path (default)

The pod keeps its **original IP** after migration. AWS VPC CNI re-assigns the same secondary IP from the subnet on the destination node, so peers continue sending traffic to the same address. Because the IP does not change, packet rewriting is not needed.

**Why use it:**

* Works on any kernel — no eBPF requirement
* Preserves more than just TCP, because the real IPv4 pod IP is unchanged
* No privileged DaemonSet is required

**Why you might avoid it:**

* The pod freeze time can be longer than TC, because the migration waits for AWS networking to assign the same IP
* Requires the source and destination nodes to reach the same subnet
* Incompatible with prefix delegation, IPv6, and multi-NIC — if any of these are enabled on the existing AWS VPC CNI, the managed fork is not installed and the cluster falls back to TC (if enabled) or to no TCP preservation

**Defaults when this path is active:** the chart autodiscovers the existing `aws-node` DaemonSet in `kube-system` and installs a managed fork on live-migration nodes; `features.migrationOptions.default.withTCP` defaults to `true`.

**External connections:** Because the pod IP is preserved, connections from resources in the same VPC (e.g. RDS databases, EC2 instances) continue to work after migration. However, outbound connections to external endpoints that go through Source NAT (SNAT) to the node IP usually break, because the SNAT state is tied to the source node. Inbound external connections via NodePort also typically break.

For subnet/AZ rules, warm-pool tuning, configuration values, and the comparison with TC, see [AWS VPC CNI](clm-reference-aws-vpc-cni.md).

#### Subnet labeling on AWS

The compatibility checker reads the `topology.atlas.example/subnet-id` node label to verify that source and destination are in the same subnet. AWS nodes may already carry this label depending on how they were provisioned. If they do not, enable `daemon.labelNodeSubnet: true` in the chart — the daemon will query the instance metadata service and label each node with its subnet ID. This setting is AWS-only and has no effect on GKE or AKS.

### Traffic Control (TC) path

The pod gets a **new IP** on the destination node. To hide the change from peers, a privileged DaemonSet on every peer node rewrites packets for the affected connections using eBPF.

**Why use it:**

* Shorter freeze time than the VPC CNI path
* No subnet or AZ constraint
* Available on AL2023 standard with K8s 1.33+ (kernel 6.12) and on the NVIDIA/Neuron AMIs

**Why you might avoid it:**

* Requires Linux kernel 6.6 or later with eBPF enabled, on the migration target node and every node that hosts a peer
* Preserves IPv4 TCP only — UDP, SCTP, and IPv6 are not handled
* Requires a privileged DaemonSet on every node
* External connections (outside the cluster) do not survive — see [Traffic Control](clm-reference-traffic-control.md) for the egress/inbound guidance

**How to enable:** set `tc.enabled: true` and `features.migrationOptions.default.withTCP: true` in the chart. See [Traffic Control](clm-reference-traffic-control.md) for the full configuration reference and the comparison with the VPC CNI path.

### Calico as a full CNI replacement

> 🚧 Important
>
> This section covers **Calico as the full CNI** with VXLAN overlay (replacing AWS VPC CNI), not Calico's network policy engine running on top of AWS VPC CNI — where IP preservation does **not** apply. For the full distinction, see [Calico as full CNI on the hub](clm-cloud-providers.md#tcp-preservation-paths-by-cloud).

When Calico is the full CNI with VXLAN overlay enabled, pod IPs come from the overlay rather than from the VPC. The system sets an annotation on the restored pod that pins it to the same IP it had on the source node, so the pod effectively keeps its original IP. Because the IP is preserved and the TC DaemonSet is not used, the `withTCP=true` / `tc.enabled=false` configuration behaves like the VPC CNI path — but without the same-subnet or same-AZ constraint.

**Why use it:**

* Same IP preservation as the VPC CNI path
* No subnet or AZ constraint (overlay IPs are not allocated from the VPC)
* Useful when the VPC CNI path is unavailable (prefix delegation, IPv6, multi-NIC) but you still want TCP preservation

**Why you might avoid it:**

* AWS officially supports only the AWS VPC CNI plugin on EKS EC2 nodes. Installing Calico as the full CNI is done at your own risk and is not covered by AWS support.

**External connections:** Because the pod IP is preserved via Calico's VXLAN overlay, in-cluster connections survive migration. External connections face the same SNAT and NodePort limitations as the VPC CNI path — see [AWS VPC CNI](clm-reference-aws-vpc-cni.md) for details on external connection behavior.

#### Install Calico as the full CNI on EKS

1. Install Calico with VXLAN overlay mode. See the [official documentation](https://docs.tigera.io/calico/latest/getting-started/kubernetes/managed-public-cloud/eks#install-eks-with-calico-networking) for details.

2. In the CLM Helm chart, leave `tc.enabled: false` and ensure `features.migrationOptions.default.withTCP: true` is set. With the AWS VPC CNI removed, the chart will not detect it and these values become the effective defaults.

## ARM64 (Graviton)

ARM64 is supported on EKS via Graviton instance families (m7g, c7g, r7g, t4g, and others). Two caveats apply:

* **Iterative pre-dumps are disabled** (the iteration count is forced to 0) because ARM64 kernels lack the soft-dirty bit support needed for incremental memory dumps. The pod still migrates correctly, but the freeze time equals the full dump time. For memory-hungry workloads on Graviton, expect a noticeably longer freeze window than on AMD64.
* **Same-architecture migration only.** Migration between an AMD64 node and an ARM64 node is not allowed.

## PVC and zone support

Persistent volume behavior follows the standard CLM rules — see the PVC and zone section of [Cloud providers](clm-cloud-providers.md) for the cross-cloud details. On EKS specifically:

* **Zonal PVs** (single-zone) require source and destination nodes to be in the **same zone**.
* **Regional PVs** (multi-zone) can be migrated to any zone in the PV's allowed list.
* The system reads `topology.kubernetes.io/zone` and the legacy `failure-domain.beta.kubernetes.io/zone` label on the destination node.

## GPU workloads

GPU workloads are not currently supported. Any container requesting `nvidia.com/gpu`, `amd.com/gpu`, or `intel.com/gpu` is blocked from migration. GPU support is planned, with no ETA.

## How to enable CLM on EKS

CLM is enabled on the node template in the Atlas console — see [How to enable CLM](clm-cloud-providers.md#how-to-enable-clm) for the canonical flow. On EKS, the platform additionally:

* Provisions AL2023 nodes with the required image
* Configures the chart values needed for the chosen TCP path (VPC CNI by default; Traffic Control or Calico if selected)

Node-level configuration (image family, AMI, instance type, kubelet arguments, and so on) is managed through **Atlas node configurations**, not by editing raw EC2 instances or running AWS CLI commands.

# See also

<Cards columns={2}>
  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Cross-cloud comparison of node images, TCP preservation paths, ARM64, and PVC support.
  </Card>
  <Card title="AWS VPC CNI" href="https://docs.atlas.example/docs/clm-reference-aws-vpc-cni" icon="fa-network-wired">
    Configuration reference for the AWS VPC CNI TCP preservation path on EKS.
  </Card>
</Cards>