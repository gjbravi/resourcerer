---
updatedAt: 2026-10-16T11:58:49.000Z
---

# Traffic Control

Reference for the Traffic Control (TC) TCP preservation path — configuration, kernel requirements, how it works, limitations, and comparison with the VPC CNI path.

For the AWS VPC CNI alternative, see [AWS VPC CNI](clm-reference-aws-vpc-cni.md).

## What TC does

When a pod is live-migrated to a new node, its pod IP usually changes. Established TCP connections to peers on other nodes would normally break because those peers still send traffic to the old IP.

The TC path solves this by installing temporary packet-level translations on the peer nodes. After the migrated pod is restored with its new IP, the TC daemon on each peer node rewrites the source/destination addresses of the relevant TCP flows so that peers continue talking to the old IP while the kernel actually delivers packets to the new IP. The translations are removed automatically when the sockets close.

TC is an **alternative** to the VPC-CNI-based path. You choose one or the other, not both.

## When to choose TC

Choose the TC path when:

* You are **not** using the Atlas-managed AWS VPC CNI subchart for IP preservation.
* You want to preserve established IPv4 TCP connections during migration.
* You can run a privileged node-level DaemonSet and your nodes run Linux kernel 6.6 or later.

Choose the VPC-CNI-based path when:

* You are running on EKS and can rely on the Atlas-managed AWS VPC CNI to assign the pod its original IP after migration.
* You need to preserve IPv4 traffic other than TCP (for example UDP), because the VPC-CNI path keeps the real IPv4 pod IP unchanged and is IPv4 protocol-agnostic.

> 📘 Note
>
> The VPC-CNI-based path can have a longer freeze time during the transition than the TC path because it waits for AWS networking to re-assign the same IP.

## Configuring the chart

The TC subsystem is controlled by the `tc` section in the chart values.

### Enable the TC DaemonSet

```yaml
tc:
  enabled: true
```

When `tc.enabled` is `true`, the chart deploys a DaemonSet named `atlas-live-tc` (default; configurable via `tc.name`) on every node that has the `live.atlas.example/install: "true"` label.

### Core TC values

| Value                  | Default                               | Description                                                                                                                     |
| ---------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `tc.enabled`           | `false`                               | Deploy the TC DaemonSet and wire the controller to use it.                                                                      |
| `tc.name`              | `atlas-live-tc`                      | Name used for the TC DaemonSet and its ServiceAccount.                                                                          |
| `tc.grpcServerPort`    | `5010`                                | TCP port on which the in-pod proxy listens for controller requests. The actual host-level gRPC server listens on a Unix socket. |
| `tc.healthzServerPort` | `8083`                                | Port for the `/readyz` health endpoint.                                                                                         |
| `tc.storagePath`       | host path under `/var/lib/tc-daemon/` | Host path where active translations are persisted so they survive a TC daemon restart.                                          |
| `tc.daemonSocket`      | Unix socket under `/run/atlas/`      | Unix socket path used by the host-level `tc-daemon` service.                                                                    |
| `tc.image.name`        | `tc`                                  | Image name for the TC container. Tag defaults to the chart's `AppVersion`.                                                      |
| `tc.tolerations`       | `NoSchedule`/`NoExecute` `Exists`     | Tolerations for the TC DaemonSet pods.                                                                                          |

The controller is informed about TC through command-line flags generated from the chart:

* `--tc-enabled={{ .Values.tc.enabled }}`
* `--tc-daemon-set-name={{ .Values.tc.name }}`

## Kernel requirements and supported node images

The TC daemon uses the **TCX** eBPF attach API. TCX was merged in Linux **6.6**, so every node that runs the TC DaemonSet or hosts a peer of a migrated pod must have kernel **6.6 or later** with eBPF enabled. Kernels older than 6.6 do not expose the TCX ingress hook and will reject the attachment.

> 🚧 Important
>
> The migration target node and every node that hosts a peer of the migrated pod must satisfy this requirement. If a peer lives on an older kernel, TC translations cannot be installed there and connections to that peer will break.

### Supported managed Kubernetes node images

The table below lists common managed Kubernetes node images and whether they ship a kernel that supports TCX. Image kernels change over time, so verify the actual kernel on your nodes with `uname -r` before enabling TC.

| Cloud   | Node image / OS                             | Typical kernel | TCX support | CLM support               |
| ------- | ------------------------------------------- | -------------- | ----------- | ------------------------- |
| **EKS** | Amazon Linux 2023 standard (K8s 1.30–1.32)  | 6.1            | ❌ No        | ✅ Yes                     |
| **EKS** | Amazon Linux 2023 standard (K8s 1.33–1.35)  | 6.12           | ✅ Yes       | ✅ Yes                     |
| **EKS** | Amazon Linux 2023 standard (K8s 1.36+)      | 6.18           | ✅ Yes       | ✅ Yes                     |
| **EKS** | Amazon Linux 2023 NVIDIA/Neuron (K8s 1.33+) | 6.12           | ✅ Yes       | ✅ Yes                     |
| **EKS** | Bottlerocket aws-k8s-1.33+                  | 6.12           | ✅ Yes       | ❌ Not supported (planned) |
| **GKE** | Container-Optimized OS (COS) 121 LTS        | 6.6            | ✅ Yes       | ✅ Yes                     |
| **GKE** | Container-Optimized OS (COS) 125/129 LTS    | 6.12           | ✅ Yes       | ✅ Yes                     |
| **GKE** | Ubuntu 24.04 containerd (linux-gke rolling) | 6.8+           | ✅ Yes       | ✅ Yes                     |
| **AKS** | Ubuntu 22.04 containerd                     | 5.15           | ❌ No        | ✅ Yes                     |
| **AKS** | Ubuntu 24.04 containerd                     | 6.8            | ✅ Yes       | ✅ Yes                     |
| **AKS** | Azure Linux 2.0                             | 5.15           | ❌ No        | ❌ Not recommended         |
| **AKS** | Azure Linux 3.0                             | 6.6            | ✅ Yes       | ✅ Yes                     |

If your current node image is in the "No" column for TCX or CLM support, switch to a supported image before enabling the TC DaemonSet. For EKS, use AL2023 with K8s 1.33 or later (standard, NVIDIA, or Neuron AMIs). For AKS, use Ubuntu 24.04 or Azure Linux 3.0. Bottlerocket is not supported by CLM despite shipping a compatible kernel — see [AWS EKS node images](clm-cloud-providers-eks.md) for details.

### Migration options that control TCP preservation

Independent of the TC DaemonSet, TCP preservation is controlled by migration options. These can be set cluster-wide in the chart values or overridden per pod via annotations (see below).

```yaml
features:
  migrationOptions:
    default:
      withTCP:                        # defaults to autodiscovered; true if VPC CNI is detected
      withTCPDnatRewrite: true        # rewrite service IPs to backend IPs before restore
      withTCPSnat: close              # what to do with SNATed connections: "close" or "keep"
```

Apply without a values file:

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set features.migrationOptions.default.withTCP=true \
  --set features.migrationOptions.default.withTCPDnatRewrite=true \
  --set features.migrationOptions.default.withTCPSnat=close
```

If you manage configuration in a values file, use `-f values.yaml` instead of `--set`.

| Option               | Default        | Description                                                                                                                               |
| -------------------- | -------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `withTCP`            | autodiscovered | Whether to try to preserve established TCP sockets during migration. When `tc.enabled` is `true`, this must be enabled for TC to be used. |
| `withTCPDnatRewrite` | `true`         | Rewrite Kubernetes Service destination IPs in the checkpoint to the real backend pod IPs so they survive the migration.                   |
| `withTCPSnat`        | `close`        | What to do with connections that were created through Source NAT (`close` = close them on restore; other values keep them as-is).         |

If you enable the TC DaemonSet, you typically also want to set:

```yaml
features:
  migrationOptions:
    default:
      withTCP: true
```

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set features.migrationOptions.default.withTCP=true
```

### VPC CNI autodiscovery and defaults

The chart can autodetect an existing `aws-node` DaemonSet in `kube-system`. When the VPC CNI path is available, `withTCP` defaults to `true` even if `tc.enabled` is `false`. This means the cluster will use the VPC-CNI-based path unless you explicitly enable the TC DaemonSet.

If both paths are theoretically available, remember that they are **alternatives**. Enable `tc.enabled` to use TC; leave it `false` and rely on VPC CNI autodiscovery to use the VPC-CNI-based path.

### Per-pod overrides

You can override the cluster defaults for a specific pod with annotations:

```yaml
annotations:
  live.atlas.example/migration-default-with_tcp: "true"
  live.atlas.example/migration-default-with_tcp_dnat_rewrite: "true"
  live.atlas.example/migration-default-with_tcp_snat: "close"
```

## How it works

1. **DaemonSet on every live-migration node.** The `atlas-live-tc` pod runs in privileged mode with `hostPID: true`. It installs a `tc-daemon` binary and a systemd service onto the host, then runs a gRPC proxy inside the pod that forwards requests to the host-level daemon over a Unix socket (`/run/atlas/tc-daemon.sock`).

2. **Peer discovery during migration.** The controller tracks which node each migration peer lives on and maps peer pod IPs to the IP of the `atlas-live-tc` pod running on the same node. That mapping is included in the restore request.

3. **IP replacement in the checkpoint.** Before the pod is restored, the migration agent replaces occurrences of the old pod IP with the new pod IP in the checkpoint files. It also rewrites conntrack entries so that injected flows carry the new IP.

4. **Translation installation on peers.** For each peer connection, the migration agent calls the peer node's TC daemon and asks it to install translations for the 4-tuples that were found in the checkpoint. The daemon resolves the correct host and veth interfaces for the peer and the new IP, then loads an eBPF program on those interfaces.

5. **Packet rewriting.** The eBPF program, attached to the TCX ingress hook, rewrites:
   * **Source addresses** of packets that leave the migrated pod with its new IP, so they look like they still come from the old IP.
   * **Destination addresses** of packets that peers send to the old IP, so they are delivered to the new IP.
   * It then redirects the packet either directly into the peer's veth (same-node peers) or through the host interface (remote peers).

6. **Automatic cleanup.** A kernel tracepoint watcher listens for TCP socket state changes. When a tracked connection closes, the watcher removes the matching translations from eBPF maps and from persistent storage.

7. **Persistence.** Active translations are persisted in a JSON file on the host. If the TC daemon restarts, it can restore the in-kernel state from that file.

## External connections

TC translations are installed only on **peer nodes** — nodes that host another pod communicating with the migrated pod. External endpoints outside the cluster do not run the TC daemon, so packets cannot be rewritten on the remote side.

### Outbound connections to external endpoints

When a pod connects to an external endpoint, the connection usually goes through Source NAT (SNAT) on the source node. The conntrack entry ties the connection to the source node IP.

* **`withTCPSnat: close` (default):** SNATed connections are closed in the CRIU checkpoint before restore. The application must reconnect after migration.
* **`withTCPSnat: keep`:** SNATed connections are kept in the checkpoint, but they often break because the source node changed and the conntrack/SNAT state is no longer valid.

Use `close` unless you have verified that the source IP and SNAT state are stable across migration.

### Inbound connections from external endpoints

Traffic from outside the cluster (for example, through a LoadBalancer or NodePort) reaches the pod via the node network. Because the migrated pod gets a new IP, external senders continue to use the old IP or node routing. There is no TC translation on the external path, so these connections usually do not survive migration.

If you need external inbound connections to survive, consider the VPC-CNI-based path instead, which keeps the pod's original IP.

### Mitigation with ingress/egress proxies

If your workload needs long-lived external connections and neither TC nor VPC CNI can preserve them, terminate the external connection in a proxy or load balancer that is independent of the migrating pod:

* **Ingress:** Put a layer-7 or layer-4 load balancer / reverse proxy in front of the pods. The external client talks to the proxy, and the proxy opens a separate connection to the pod. When the pod migrates, the proxy reconnects to the new instance.
* **Egress:** Send outbound traffic through an egress proxy or NAT gateway with a stable IP. The pod talks to the proxy, and the proxy opens the external connection. Migration only affects the pod-to-proxy hop, which can be re-established.

This is the standard pattern for keeping a stable external identity while allowing individual pods to migrate.

## Limitations

* **Linux only.** The eBPF TCX program is loaded only on Linux nodes. Non-Linux builds of the TC daemon are no-ops.
* **IPv4 only.** The eBPF program matches `ETH_P_IP` and stores addresses as 32-bit values. IPv6 traffic is not translated.
* **TCP only.** The program inspects `IPPROTO_TCP`. UDP, SCTP, and other L4 protocols are not handled.
* **Privileged node access.** The TC DaemonSet requires a privileged container, `hostPID: true`, access to `/sys/kernel/debug`, the host root filesystem, and the systemd private socket.
* **Kernel must support TCX/eBPF.** The daemon attaches its program to the TCX ingress hook. TCX requires Linux kernel 6.6 or later; older kernels do not expose the hook and will reject the attachment.
* **One daemon per peer node.** A translation must be installed on every node that hosts a peer of the migrated pod. If the TC daemon is not running on a peer node, connections to peers on that node will not survive.
* **Cannot be combined with VPC CNI path.** TC and the VPC-CNI-based IP preservation are alternatives. Enabling the TC DaemonSet switches the cluster to the TC path.
* **Active development.** The TC subsystem is still evolving. Configuration details and supported scenarios may change between releases.

For a side-by-side comparison of the TC and VPC CNI paths across all dimensions, see the [Cloud providers comparison table](clm-cloud-providers.md#comparison-table).

## External references

* AWS EKS documentation: [Prefix Mode for Linux](https://docs.aws.amazon.com/eks/latest/best-practices/prefix-mode-linux.html) and [Amazon VPC CNI](https://docs.aws.amazon.com/eks/latest/best-practices/vpc-cni.html).
* amazon-vpc-cni-k8s project: [ENABLE\_PREFIX\_DELEGATION](https://github.com/aws/amazon-vpc-cni-k8s) and [VPC CNI configuration variables](https://github.com/aws/amazon-vpc-cni-k8s#cni-configuration-variables).

# See also

<Cards columns={2}>
  <Card title="AWS VPC CNI" href="https://docs.atlas.example/docs/clm-reference-aws-vpc-cni" icon="fa-server">
    The alternative TCP preservation path compared throughout this page.
  </Card>
  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Cross-cloud comparison of where the TC path is available and its kernel requirements.
  </Card>
</Cards>