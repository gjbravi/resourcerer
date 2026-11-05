---
updatedAt: 2026-10-14T14:00:51.000Z
---

# TCP migration modes

Overview of the three TCP preservation paths available during container live migration: AWS VPC CNI, Traffic Control, and Calico as full CNI.

Container Live Migration supports several ways to handle a pod's network identity during migration. Each path has different requirements and trade-offs.

| Mode                     | Pod IP after migration | What happens to established connections                                           | When to choose                                                                                              |
| ------------------------ | ---------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| **No TCP preservation**  | New IP                 | Existing TCP connections close; peers reconnect.                                  | Fastest migration; applications tolerate reconnects; you do not need TCP state kept.                        |
| **Traffic Control (TC)** | New IP                 | Peer nodes rewrite IPv4 TCP packets so connections survive the IP change.         | You need IPv4 TCP preserved and can run a privileged DaemonSet with kernel 6.6+.                            |
| **AWS VPC CNI**          | Same IPv4 IP           | Connections keep working because the pod keeps its original IPv4 IP.              | EKS with AWS VPC CNI running IPv4; you need IPv4 traffic preserved or cannot run the TC DaemonSet.          |
| **Calico as full CNI**   | Same IP                | Connections keep working because the pod keeps its original IP via VXLAN overlay. | EKS (replace VPC CNI) or AKS (via BYO CNI); you need protocol-agnostic IP preservation without kernel 6.6+. |

For details on each path, see:

* [Traffic Control](clm-reference-traffic-control.md)
* [AWS VPC CNI](clm-reference-aws-vpc-cni.md)
* [AWS EKS — Calico as full CNI](clm-cloud-providers-eks.md)
* [Azure AKS — Calico via BYO CNI](clm-cloud-providers-aks.md)

## No TCP preservation

The simplest mode. The pod is checkpointed and restored on a new node with a new IP, but its TCP sockets are **not** preserved. Peers see the connection drop and must reconnect.

Enable or disable it per migration, per pod, or cluster-wide via the `features.migrationOptions.default.withTCP` Helm value (see [Quick configuration](#quick-configuration) below). This is the default when neither the VPC CNI path nor the TC DaemonSet is available.

## Traffic Control (TC)

The pod gets a new IP, but a privileged eBPF-based DaemonSet on peer nodes rewrites IPv4 TCP packets so that established connections survive. The translation hides the IP change from both ends of the connection.

* **Protocols:** IPv4 TCP only.
* **Requirements:** Linux kernel with TCX (6.6+), eBPF enabled.
* **Trade-off:** Lower freeze time than VPC CNI, but requires privileged DaemonSet on every node.
* **Limitations:** Preserves in-cluster IPv4 TCP connections only. Connections to external endpoints outside the cluster do not survive migration. UDP, SCTP, and IPv6 are not preserved.

See [Traffic Control](clm-reference-traffic-control.md) for full details on external connection behavior and limitations.

## AWS VPC CNI

The pod keeps its original IPv4 IP after migration because the chart installs or mutates a managed copy of the AWS VPC CNI `aws-node` DaemonSet on live-migration nodes. AWS networking re-assigns the same secondary IPv4 IP on the destination node.

* **Protocols:** IPv4 protocol-agnostic — any IPv4 traffic keeps working because the IPv4 IP does not change.
* **Requirements:** EKS with AWS VPC CNI; must not use prefix delegation, IPv6, or multi-NIC.
* **Trade-off:** No eBPF privilege required, but freeze time can be longer while AWS assigns the same IP.

See [AWS VPC CNI](clm-reference-aws-vpc-cni.md).

## Calico as full CNI

Calico can be installed as the **full CNI** (not just network policy) with VXLAN overlay mode on EKS and AKS. In this mode, the pod keeps its original IP after migration — the system pins the IP via a Calico annotation on the restored pod.

This is different from installing Calico as a network policy engine on top of the cloud's native CNI. IP preservation only works when Calico is the full CNI with VXLAN.

* **Protocols:** IPv4 protocol-agnostic — any IPv4 traffic keeps working because the IP does not change.
* **Requirements:** EKS (replace VPC CNI with Calico) or AKS (use BYO CNI feature). Not available on GKE.
* **Trade-off:** No kernel 6.6+ requirement (works on any kernel), no privileged DaemonSet needed, but requires managing Calico yourself. No subnet or AZ constraint (overlay IPs are not from the VPC).
* **Freeze time:** Similar to VPC CNI path — the migration waits for the CNI to assign the same IP on the destination.

> 🚧 Important
>
> Calico as a network policy engine on top of the cloud's native CNI is NOT the same as Calico as the full CNI with VXLAN. IP preservation only works when Calico is the full CNI. GKE does not support Calico as the full CNI.

For installation instructions, see [AWS EKS](clm-cloud-providers-eks.md) and [Azure AKS](clm-cloud-providers-aks.md).

## How to pick

| Situation                                                                                | Recommended mode    |
| ---------------------------------------------------------------------------------------- | ------------------- |
| EKS with AWS VPC CNI running IPv4, need IPv4 UDP or other non-TCP IPv4 traffic preserved | AWS VPC CNI         |
| Need IPv4 TCP only and want lower freeze time                                            | TC                  |
| EKS or AKS, need protocol-agnostic IP preservation without kernel 6.6+                   | Calico as full CNI  |
| Applications tolerate reconnects and you want simplest setup                             | No TCP preservation |

## Quick configuration

Cluster-wide defaults live under `features.migrationOptions.default` in the chart values:

```yaml
features:
  migrationOptions:
    default:
      withTCP: true        # true = try to preserve TCP; false = no TCP preservation
```

Apply the change without editing a values file:

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set features.migrationOptions.default.withTCP=true
```

If you manage configuration in a values file, use `-f values.yaml` instead of `--set`.

Per-pod overrides:

```yaml
annotations:
  live.atlas.example/migration-default-with_tcp: "true"
```

The chart autodiscovers the VPC CNI path and defaults `withTCP` to `true` when it is available. To use the TC path, deploy the TC DaemonSet (`tc.enabled: true`) and set the annotations or defaults accordingly. To force no TCP preservation, set `withTCP: false`:

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set features.migrationOptions.default.withTCP=false
```

Or, if you use a values file:

```yaml
features:
  migrationOptions:
    default:
      withTCP: false
```

```bash
helm upgrade -f values.yaml <release-name> <chart>
```

# See also

<Cards columns={2}>
  <Card title="Traffic Control" href="https://docs.atlas.example/docs/clm-reference-traffic-control" icon="fa-network-wired">
    Reference for the TC TCP preservation path.
  </Card>
  <Card title="AWS VPC CNI" href="https://docs.atlas.example/docs/clm-reference-aws-vpc-cni" icon="fa-server">
    Reference for the AWS VPC CNI TCP preservation path.
  </Card>
</Cards>