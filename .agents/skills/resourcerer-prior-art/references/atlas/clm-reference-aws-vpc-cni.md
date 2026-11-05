---
updatedAt: 2026-09-03T07:02:11.000Z
---

# AWS VPC CNI

Reference for the AWS VPC CNI TCP preservation path — autodiscovery, configuration, external connections, limitations, and comparison with Traffic Control.

For the TC alternative, see [Traffic Control](clm-reference-traffic-control.md).

## What the VPC-CNI-based path does

When a pod is live-migrated to a new node, its pod IP usually changes. Established connections to peers on other nodes would normally break because those peers still send traffic to the old IP.

The VPC-CNI-based path solves this by keeping the **original pod IP** after migration. The destination node assigns the migrated pod the same secondary IP that it had on the source node. Because the IP does not change, peers continue talking to the same address and no packet rewriting is needed.

This path is an **alternative** to the TC path. You choose one or the other, not both.

## When to choose the VPC-CNI-based path

Choose this path when:

* You are running on **AWS EKS** with the AWS VPC CNI plugin installed.
* You need to preserve IPv4 traffic other than TCP — for example IPv4 UDP — because the real IPv4 pod IP stays unchanged and the path is IPv4 protocol-agnostic.
* You cannot run the privileged node-level TC DaemonSet.
* You are okay with a potentially longer freeze time during migration while AWS networking re-assigns the same IP.

Choose the TC path when:

* You want lower freeze times and only need IPv4 TCP preserved.
* You are not using the Atlas-managed AWS VPC CNI fork for IP preservation.
* Your kernel supports eBPF TCX (6.6+) and you can run a privileged DaemonSet.

> 📘 Note
>
> The VPC-CNI-based path can have a longer freeze time than the TC path because it waits for AWS networking to assign the same IP on the destination node. Unlike the TC path, it does not require a privileged DaemonSet. Because the pod IP is preserved, connections from resources in the same VPC but outside the cluster (e.g. RDS databases, EC2 instances) continue to work after migration — the TC path cannot preserve these because external resources do not run the TC daemon.

## How it works

1. **Autodiscovery.** During installation the chart looks up the existing `aws-node` DaemonSet in `kube-system`. If it exists, has an `aws-node` container, and does not use an unsupported configuration, the chart decides to install the managed VPC CNI fork.

2. **Separate the original CNI from live-migration nodes.** The original `aws-node` DaemonSet is mutated so it does **not** run on nodes labeled `live.atlas.example/install: "true"`. This is done both by a Helm pre-install/pre-upgrade Job and by a mutating admission webhook that watches changes to the `kube-system/aws-node` DaemonSet.

3. **Install the Atlas-managed fork on live-migration nodes.** The chart deploys a second `aws-node` DaemonSet (named `atlas-aws-node` by default) that uses the Atlas fork image and a node affinity requiring `live.atlas.example/install: "true"`. It inherits key environment variables from the original DaemonSet so it uses the same VPC, subnet, and warm-pool settings.

4. **Preserve the IP during migration.** When the destination pod is created, the managed VPC CNI fork assigns it the same secondary IP that the source pod had. Because the IP does not change, existing connections resume without packet rewriting.

5. **Enable TCP preservation by default.** When the VPC CNI path is active, `features.migrationOptions.default.withTCP` defaults to `true`, so migrations automatically try to preserve established TCP sockets.

## Configuring the chart

The VPC CNI path is controlled by the `atlas-aws-vpc-cni` section in the chart values and by autodiscovery logic.

### Enable or disable autodiscovery

```yaml
atlas-aws-vpc-cni:
  enabled: true   # allow autodiscovery/installation (default)
  # enabled: false  # unconditionally skip the VPC CNI path
```

Setting `enabled: false` disables the path entirely, regardless of whether `aws-node` is detected. You might do this when you want to use the TC path exclusively, or when you are running Calico as the full CNI and do not need the managed VPC CNI fork.

### Autodiscovery rules

The chart installs the managed fork only when **all** of the following are true:

1. `atlas-aws-vpc-cni.enabled` is not explicitly `false`.
2. A DaemonSet named `aws-node` exists in `kube-system`.
3. That DaemonSet has a container named `aws-node` (confirming it is the AWS VPC CNI plugin).
4. The existing DaemonSet does **not** have any of these unsupported settings:
   * `ENABLE_PREFIX_DELEGATION=true`
   * `ENABLE_IPv6=true`
   * `ENABLE_MULTI_NIC=true`

If any unsupported setting is detected, the managed fork is **not** installed. The cluster then falls back to the TC path (if `tc.enabled: true`) or to no TCP preservation. This check runs at install time — after installation, the active path is determined by which components were successfully deployed (VPC CNI fork, TC DaemonSet, or neither).

### Core values

| Value                                 | Default                            | Description                                                                                                                                                                                                                                                                                                                                                          |
| ------------------------------------- | ---------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `atlas-aws-vpc-cni.enabled`          | `true`                             | Allow the chart to autodiscover and install the managed VPC CNI fork. Set to `false` to skip the path unconditionally.                                                                                                                                                                                                                                               |
| `atlas-aws-vpc-cni.fullnameOverride` | `atlas-aws-node`                  | Name of the managed DaemonSet deployed on live-migration nodes.                                                                                                                                                                                                                                                                                                      |
| `atlas-aws-vpc-cni.image.override`   | Defined in chart `values.yaml`     | Atlas fork image used on live-migration nodes. The exact tag is set by the chart and changes with each release.                                                                                                                                                                                                                                                    |
| `atlas-aws-vpc-cni.envOverrides`     | `{}`                               | Map of environment variables that always override anything else on the managed fork, including autodiscovered values from `kube-system/aws-node` and values set under `atlas-aws-vpc-cni.env`. Use this, not `env`, to force a value for any of the 8 autodiscovered keys. See [Autodiscovered environment variables](#autodiscovered-environment-variables) below. |
| `atlas-aws-vpc-cni.affinity`         | `live.atlas.example/install In ["true"]` | Affinity that restricts the managed fork to live-migration nodes.                                                                                                                                                                                                                                                                                                    |

The controller is informed about VPC CNI through command-line flags generated from the chart. When autodiscovery succeeds, the controller defaults `withTCP` to `true`.

### Autodiscovered environment variables

The chart reads these 8 variables from the existing `aws-node` DaemonSet and merges them into the managed fork:

* `ENI_CONFIG_LABEL_DEF`
* `WARM_IP_TARGET`
* `AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG`
* `ADDITIONAL_ENI_TAGS`
* `CLUSTER_ENDPOINT`
* `CLUSTER_NAME`
* `VPC_ID`
* `ENI_CONFIG_ANNOTATION_DEF`

This ensures the Atlas-managed fork uses the same VPC, subnets, custom network configuration, and warm-pool behavior as the original plugin.

**Precedence** (lowest to highest) for these 8 keys on the managed fork:

1. Subchart defaults from `helm/dependencies/aws-vpc-cni/values.yaml`.
2. Autodiscovered values read from the existing `kube-system/aws-node` DaemonSet, which **always** override both the subchart default and anything set under `atlas-aws-vpc-cni.env`.
3. `atlas-aws-vpc-cni.envOverrides.<KEY>`, which **always** overrides any key, autodiscovered or not.

> ⚠️ Important
>
> Autodiscovery overrides `atlas-aws-vpc-cni.env` in the precedence order. Setting `atlas-aws-vpc-cni.env.AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG` or `env.ADDITIONAL_ENI_TAGS` does **not** override what was autodiscovered from `aws-node`. To force a value for any of the 8 keys, use `envOverrides`. Two of the 8 autodiscovered keys, `AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG` and `ADDITIONAL_ENI_TAGS`, have non-empty subchart defaults.

Example (values file):

```yaml
atlas-aws-vpc-cni:
  envOverrides:
    AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG: "true"
```

Equivalent `--set` form:

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set atlas-aws-vpc-cni.envOverrides.AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG=true
```

`envOverrides` may also be used for any other env var, not only the 8 autodiscovered ones — it is the single override path for the managed fork's environment.

### Mutation webhook and install hook

Two mechanisms keep the original `aws-node` DaemonSet off live-migration nodes:

* **Pre-install / pre-upgrade Job:** Patches the existing `kube-system/aws-node` DaemonSet to add the node affinity `live.atlas.example/install NotIn ["true"]`.
* **Mutating admission webhook:** Watches for updates to `kube-system/aws-node` and re-applies the same affinity if something removes it.

On uninstall, a pre-delete Job removes the affinity expression from the original DaemonSet.

### Migration options

When the VPC CNI path is active, `features.migrationOptions.default.withTCP` defaults to `true`. You can override it explicitly if needed:

```yaml
features:
  migrationOptions:
    default:
      withTCP: true
```

Apply without a values file:

```bash
helm upgrade --reuse-values <release-name> <chart> \
  --set features.migrationOptions.default.withTCP=true
```

If you manage configuration in a values file, use `-f values.yaml` instead of `--set`.

## External connections

The VPC-CNI-based path keeps the pod's original IPv4 IP, so intra-VPC traffic and any connection that is truly bound to the pod IP can survive migration. However, by default AWS VPC CNI Source-NATs outbound traffic to destinations outside the VPC to the node IP. External connections therefore usually depend on the node IP, not only the pod IP.

### Outbound connections to external endpoints

By default, `AWS_VPC_K8S_CNI_EXTERNALSNAT` is `false` in the chart. With this default, traffic from a pod to an IP address outside the VPC is SNATed to the primary private IP of the node's primary ENI. The external endpoint sees the **node IP** as the source, and the conntrack entry is tied to the original node.

Because the migrated pod runs on a different node, that SNAT/conntrack state is lost. The connection usually breaks even though the pod IP itself did not change.

If you set `AWS_VPC_K8S_CNI_EXTERNALSNAT=true` and route outbound traffic through an external NAT gateway (for example, an AWS NAT Gateway with a stable IP), the external endpoint sees the NAT gateway IP instead of the node IP. In that setup, the pod IP is preserved by CLM and the NAT gateway IP is stable, so outbound connections are more likely to survive.

### Inbound connections from external endpoints

Inbound external traffic generally reaches the pod through one of these paths:

* **Direct pod IP targeting** (for example, a Network Load Balancer with pod IP targets, or direct intra-VPC traffic): can survive because the destination pod IP does not change.
* **NodePort or Load Balancer targeting node IPs**: usually breaks because the connection state is tied to the original node. When the pod moves, the node-level conntrack and routing state do not move with it.

### Mitigation with ingress/egress proxies

If your workload needs long-lived external connections and you cannot rely on pod-IP-only routing, the usual mitigation is to terminate external connections in a proxy or load balancer that is independent of the migrating pod:

* **Ingress:** Put a layer-7 or layer-4 load balancer / reverse proxy in front of the pods. The external client talks to the proxy, and the proxy opens a separate connection to the pod. When the pod migrates, the proxy reconnects to the new instance.
* **Egress:** Send outbound traffic through an egress proxy or NAT gateway with a stable IP. The pod talks to the proxy, and the proxy opens the external connection. Migration only affects the pod-to-proxy hop, which can be re-established.

This is the same pattern used for any Kubernetes workload that needs stable external identities: keep the external identity on the proxy or gateway, not on the individual pod.

## Limitations and real-world setup guidance

### Incompatible VPC CNI configurations

The managed fork is **not** installed, and therefore the VPC-CNI-based path is unavailable, if the existing `aws-node` uses any of these settings:

| Setting                         | Why it is incompatible                                                                                                                         |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `ENABLE_PREFIX_DELEGATION=true` | Prefix delegation allocates `/28` prefixes rather than individual secondary IPs. The current path requires individual secondary IP assignment. |
| `ENABLE_IPv6=true`              | Only IPv4 is supported by the managed fork path.                                                                                               |
| `ENABLE_MULTI_NIC=true`         | Multi-NIC mode is not supported.                                                                                                               |

If any of these are enabled, use the TC path or disable TCP preservation.

### Warm pool and small subnets

AWS VPC CNI maintains a warm pool of secondary IPs or prefixes on each node. The size of that pool is controlled by `WARM_IP_TARGET`, `WARM_PREFIX_TARGET`, and `MINIMUM_IP_TARGET`.

A common source of failures in small subnets looks like this:

* A subnet has only 32 available IPs.
* `MINIMUM_IP_TARGET` is set to 16 and `WARM_IP_TARGET` is 2.
* A node holds a full `/28` prefix (16 IPs) to satisfy the warm target, even if it is only running a few pods.
* When a pod is migrated to that node, VPC CNI may report that no IP is available because the warm pool has consumed the subnet, even though the node is not actually using all of those IPs.

This is especially problematic when:

* Subnets are small or highly fragmented.
* Many nodes are cordoned or draining but still hold warm IPs.
* The cluster has high pod churn.

**Best practice:** For small subnets, tune `WARM_IP_TARGET` and `MINIMUM_IP_TARGET` to values slightly above the expected pod density per node, rather than using the default `WARM_ENI_TARGET=1` which allocates a full ENI worth of IPs. AWS documentation recommends `MINIMUM_IP_TARGET` slightly higher than the expected pod count and a small `WARM_IP_TARGET`.

> See the AWS VPC CNI docs on [prefix and IP targets](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/prefix-and-ip-target.md) and [ENI and IP targets](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/eni-and-ip-target.md) for details.

### Same-VPC and same-subnet considerations

The VPC-CNI-based path works best when:

* Source and destination nodes are in the **same VPC** and can reach the same subnet.
* The destination node has an available secondary IP in the same subnet as the source pod.
* No custom networking configuration prevents IP reuse across availability zones.

Setups that can cause problems include:

* **Multiple small subnets per AZ with custom networking.** If `AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG=true` maps different subnets to different nodes or AZs, the destination node may not have an IP available in the source pod's subnet, and migration will fail.
* **Random ENI or subnet assignment.** If the CNI can place the pod in any of several subnets, the destination pod may receive a different subnet and therefore cannot reuse the source IP.
* **Security Groups per Pod.** The managed fork does not enable `ENABLE_POD_ENI`; if you rely on per-pod security groups, verify that behavior is compatible with your setup.

**Best practice:** Use a single, large subnet per AZ for live-migration nodes, avoid custom networking unless you can guarantee the same subnet on both source and destination, and ensure `WARM_IP_TARGET`/`MINIMUM_IP_TARGET` leave enough free IPs for migrations.

### Freeze time

Because the migration waits for the AWS API to assign the same secondary IP on the destination node, the pod freeze time can be longer than with the TC path. The exact duration depends on:

* Whether the destination node already has a free secondary IP in its warm pool.
* Whether VPC CNI needs to attach a new ENI (can take up to \~10 seconds).
* EC2 API latency and throttling.

For a side-by-side comparison of the VPC CNI and Traffic Control paths across all dimensions, see the [Cloud providers comparison table](clm-cloud-providers.md#comparison-table).

## External references

* AWS EKS documentation: [Prefix Mode for Linux](https://docs.aws.amazon.com/eks/latest/best-practices/prefix-mode-linux.html) and [Amazon VPC CNI](https://docs.aws.amazon.com/eks/latest/best-practices/vpc-cni.html).
* amazon-vpc-cni-k8s project: [prefix and IP targets](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/prefix-and-ip-target.md), [ENI and IP targets](https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/eni-and-ip-target.md), and [VPC CNI configuration variables](https://github.com/aws/amazon-vpc-cni-k8s#cni-configuration-variables).

# See also

<Cards columns={2}>
  <Card title="Traffic Control" href="https://docs.atlas.example/docs/clm-reference-traffic-control" icon="fa-network-wired">
    The alternative TCP preservation path compared throughout this page.
  </Card>
  <Card title="AWS EKS" href="https://docs.atlas.example/docs/clm-cloud-providers-eks" icon="fa-cloud">
    Per-cloud setup for EKS, where the VPC CNI path is the default.
  </Card>
</Cards>