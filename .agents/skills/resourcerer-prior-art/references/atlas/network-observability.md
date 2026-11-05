---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Network intelligence

This guide explains how to use network intelligence features.

## Netflow list

Netflow is a network-related metric that tracks sent and received traffic in bytes. Unlike VPC flow logs, it contains rich Kubernetes context.

In the Netflow list, you can find various Kubernetes cluster network metrics.

The combination of dynamic grouping and advanced [CEL](https://github.com/google/cel-spec)-based search makes it easy to find the top cross-zone, cross-namespace, traffic to the Internet, and unexpected spikes.

![](https://files.readme.io/17aea8929b85ba13d8a4d44673362bb39614f6cb9ca5c3803b36347719486aca-image.png)

### Search examples

Find cross-zone traffic within the Kubernetes cluster. You can also group sources and destinations by zone only.

```c
zone != "" && dst_zone != "" && zone != dst_zone
```

Find traffic to the internet.

```c
dst_workload_kind=="internet"
```

Find any traffic from source namespaces.

```c
ns in ["metrics", "kube-system"]
```

Find any traffic from source to destination namespaces.

```c
ns == "metrics" && dst_ns == "kube-system"
```

Find GitLab runners containers that process traffic to the internet. For this search, you can group sources by process only. It can also be helpful to sort by received traffic to see the possible NAT processing impact.

```c
ns == "gitlab" && dst_workload_kind == "internet"
```

Find traffic from the source namespace and pod name prefix.

```c
ns == "gitlab" && pod_name.startsWith("runner-")
```

### All available search fields

CEL search can combine multiple available fields for Netflow metadata.

| Property            | Type   | Description                                                                                                                           |
| ------------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------------- |
| protocol            | Enum   | Available values for the protocol are `udp` and `tcp`                                                                                 |
| process             | string | Source container Linux process name.                                                                                                  |
| container\_name     | string | Source Kubernetes container name.                                                                                                     |
| pod\_name           | string | Source Kubernetes pod name.                                                                                                           |
| ns                  | string | Source Kubernetes namespace.                                                                                                          |
| zone                | string | Source node zone.                                                                                                                     |
| workload\_name      | string | Source Kubernetes workload name.                                                                                                      |
| workload\_kind      | string | Source Kubernetes workload kind.                                                                                                      |
| node\_name          | string | Source Kubernetes node name.                                                                                                          |
| addr                | string | Source IP address                                                                                                                     |
| dst\_domain         | string | Destination DNS domain. This field is the best effort and may not always be present.                                                  |
| dst\_pod\_name      | string | Destination Kubernetes pod name.                                                                                                      |
| dst\_ns             | string | Destination Kubernetes namespace.                                                                                                     |
| dst\_workload\_name | string | Destination Kubernetes workload name                                                                                                  |
| dst\_workload\_kind | string | Destination Kubernetes workload kind or special kinds such as `private` for any unknown private IPs and `internet` for any public IP. |
| dst\_node\_name     | string | Destination Kubernetes node name.                                                                                                     |

### How to enable Netflow collection

The Netflow collection is implemented with [Atlas Sentinel agent](sentinel.md) and can be installed via the Helm chart.

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.netflow-enabled=true
```

> 📘 Note
>
> If you have `egressd` running it should be uninstalled.

<Cards columns={2}>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure network traffic monitoring and eBPF settings.
  </Card>
  <Card title="Network cost" href="https://docs.atlas.example/docs/network-cost" icon="fa-network-wired">
    Analyze network traffic distribution and associated costs.
  </Card>
</Cards>