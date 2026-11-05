---
updatedAt: 2026-10-27T09:40:57.000Z
---

# Configuring Sentinel features

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

This guide explains how to configure Sentinel's monitoring, telemetry, and security insight features. Sentinel provides several specialized capabilities that can be enabled and customized according to your needs.

## Configuration overview

Sentinel supports multiple configuration options that can be set via Helm during installation or upgrade. The basic format for enabling or modifying features is:

> 📘 Configuration-only changes
>
> This command upgrades all components in the umbrella chart, not just Sentinel. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.[configuration-option]=[value]
```

The Sentinel chart has two kinds of settings. Chart values cover the agent and controller images, resources, and feature switches such as `agent.gpu.enabled`. Feature settings pass through the `controller.extraArgs` and `agent.extraArgs` values, one flag per setting. This page lists the settings for each feature.

To view the chart values for the chart version deployed in your cluster, run:

```shell
helm show values atlas-helm/atlas-sentinel
```

## Scanning frequencies and intervals

Atlas performs different types of security scans at various intervals. You can customize these frequencies to meet your specific requirements:

### Image vulnerability scanning

**Default behavior:**

* **Detection interval**: Checks for new images every 30 seconds
* **Concurrent scans**: 1 image scanned at a time
* **Trigger**: Scans start automatically when new running images are detected
* **Scan timeout**: 10 minutes per image
* **Scan job resources**: CPU request 1m, limit 2; memory request 1Mi, limit 2Gi

#### Configure image scan detection interval

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-interval=60s
```

#### Configure concurrent image scans

To increase scanning performance for environments with many images:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-concurrent-scans=6
```

#### Set the scan timeout

A single image scan runs for at most 10 minutes by default. To change the limit:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-timeout=15m
```

#### Configure the scan job resources

Each image scan runs as a Kubernetes Job with its own CPU and memory settings. The default CPU request is 1m with a limit of 2. The default memory request is 1Mi with a limit of 2Gi.

To change the resources for image scans:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-cpu-request=500m \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-cpu-limit=2 \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-memory-request=512Mi \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-memory-limit=4Gi
```

#### Include secret detection in image scans

By default, image scans do not look for exposed secrets inside container images. This keeps scans fast. To include secret detection:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-disabled-analyzers=""
```

### Compliance checking

Compliance runs two kinds of checks. Configuration checks look for misconfigurations and RBAC issues in your cluster. Node benchmark checks validate kubelet and system components against the Atlas security benchmark.

**Default behavior:**

* **Configuration checks**: Every 60 seconds
* **Node benchmark checks**: Every 5 minutes

#### Configure how often configuration checks run

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-linter-scan-interval=120s
```

#### Configure how often node benchmark checks run

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-bench-scan-interval=10m
```

### Summary of scan frequencies

| Security Feature                 | Default Frequency | Configuration Parameter     |
| -------------------------------- | ----------------- | --------------------------- |
| **Image vulnerability scanning** | Every 30 seconds  | `image-scan-interval`       |
| **Configuration checks**         | Every 60 seconds  | `kube-linter-scan-interval` |
| **Node benchmark checks**        | Every 5 minutes   | `kube-bench-scan-interval`  |

### Use a custom image

You can configure Sentinel to use a custom image from your private registry:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.image.repository=my-sentinel-repository \
  --set autoscaler.atlas-sentinel.image.tag=my-tag \
  --set 'autoscaler.atlas-sentinel.imagePullSecrets[0].name=my-pull-secret'
```

### Private image scanning

For detailed instructions on configuring private image scanning, refer to the dedicated [Private Image Scanning](image-scanning.md) documentation.

### Excluding namespaces from scanning

You may want to exclude specific namespaces from image scanning, particularly for third-party monitoring tools or systems that manage their own security. To configure namespace exclusions:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-ignored-namespaces=namespace1,namespace2
```

For example, to exclude the Dynatrace namespace from scanning:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-ignored-namespaces=dynatrace
```

You can specify multiple namespaces by separating them with commas.

## Network traffic monitoring

Sentinel can collect Kubernetes network flows using eBPF. This data powers [network cost monitoring](network-cost.md) and gives you visibility into pod-to-pod and pod-to-external traffic patterns.

To enable network traffic monitoring:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.netflow-enabled=true
```

Sentinel batches the collected flows and exports them to Atlas every 15 seconds by default. To change the export interval:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.extraArgs.netflow-export-interval=30s
```

> 📘 Note
>
> If you have the `egressd` component running, it should be uninstalled before enabling network traffic monitoring in Sentinel:
>
> ```shell
> helm uninstall egressd -n atlas-agent
> ```

> 📘 Enrich network flows with cloud context
>
> To enrich netflow data with VPC, subnet, and region information from your cloud provider, see [Cloud network context](cloud-network-context.md).

## Resource statistics monitoring

Sentinel can collect resource usage statistics from containers and nodes. These monitoring capabilities are controlled by two independent flags, each enabling a distinct category of metrics.

### Kubernetes storage metrics

Sentinel collects storage utilization data for both ephemeral storage and persistent volumes across your cluster nodes. This data surfaces per-node storage usage in the **Node list** of the Atlas console.

Storage metrics collection requires both the Sentinel controller and agent to be enabled.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.enabled=true \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.storage-stats-enabled=true
```

Besides storage utilization directly in the node, Sentinel is also able to collect metrics on cloud volumes that are
being used. Collection happens inside the controller component and requires cloud access permissions, in order to fetch
volume related information. For now, only AWS and GCP are supported.

#### AWS

If IMDS (Instance Metadata Service) access is enabled for the Sentinel controller pod, no additional configuration is
required.  Sentinel will automatically use the node's instance role to authenticate with AWS.

If IMDS access is not available (e.g. it has been restricted at the instance or pod level), you must set up
[IRSA (IAM Roles for Service Accounts)](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
and annotate the Sentinel controller service account with the appropriate IAM role ARN. The IAM role must have
the following permission:

* `ec2:DescribeVolumes`

Once authentication is configured, enable cloud volume metrics collection:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.enabled=true \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.storage-stats-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=aws \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-aws-region=<your-aws-region> \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-storage-sync-enabled=true
```

Replace `<your-aws-region>` with the AWS region your cluster is running in.

#### GCP

Cloud volume metrics collection on GCP requires [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
to be configured so that the Sentinel controller service account can authenticate with the GCP API. The bound GCP
service account must have the following permissions in your project:

* `compute.disks.list`
* `compute.instances.list`

Once Workload Identity is configured, enable cloud volume metrics collection:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.enabled=true \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.storage-stats-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=gcp \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-gcp-project-id=<your-gcp-project-id> \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider-storage-sync-enabled=true
```

Replace `<your-gcp-project-id>` with your GCP project ID.

### Pressure Stall Information (PSI) metrics

Sentinel can collect Pressure Stall Information (PSI) metrics from containers and nodes, along with CPU, memory, and I/O usage statistics. PSI metrics indicate how often workloads experience delays due to contention for CPU, memory, or I/O resources, helping you identify performance bottlenecks before they affect application behavior.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.stats-enabled=true
```

Sentinel collects resource statistics every 60 seconds by default. To change the collection interval:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.extraArgs.stats-scrape-interval=120s
```

### GPU metrics

Sentinel collects GPU metrics from NVIDIA GPUs in your cluster using DCGM Exporter. The GPU export pipeline and `dcgm-exporter` run only on nodes with GPUs attached, regardless of cloud provider. Sentinel itself continues to run on all cluster nodes.

To enable GPU metrics collection:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.agent.gpu.enabled=true
```

#### Configure the GPU metrics interval

Sentinel collects GPU metrics every 15 seconds by default. To change the interval:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.gpu.exportInterval=30s
```

#### Use an existing DCGM exporter

By default, Sentinel adds a DCGM exporter sidecar to its agent pods on GPU nodes. If your cluster already runs its own DCGM exporter, you can use it instead of the sidecar.

> ⚠️ Important
>
> Only disable the sidecar when every GPU node in your cluster runs a DCGM exporter. Sentinel reads the exporter on each node and reports no GPU metrics for nodes without one. The exporter must expose its metrics on port 9400. If your exporter uses a different port, set `agent.gpu.dcgmExporter.port` to match.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.gpu.dcgmExporter.enabled=false
```

#### Show per-pod GPU usage with GPU sharing

On clusters that use [GPU sharing](gpu-sharing.md), such as MIG partitions or time slicing, each pod can use part of a GPU. Enable per-pod GPU metrics to see the usage of each pod:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.gpu.dcgmExporter.virtualGPUs.enabled=true
```

#### Support DRA clusters

On clusters that use dynamic resource allocation (DRA) for GPUs, enable DRA support:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.gpu.dra.enabled=true
```

DRA support requires Kubernetes 1.32 or later with the DRA feature gates enabled and the NVIDIA DRA driver installed. See [DRA](dra.md).

#### Attribute GPU usage to custom workloads

If a pod label identifies your GPU workloads, you can attribute GPU metrics to the label value:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.extraArgs.gpu-workload-label-keys=workloads.atlas.example/custom-workload
```

The label value overrides the workload name that Sentinel resolves for the pod.

#### Migrating from gpu-metrics-exporter

If you are currently using the standalone `gpu-metrics-exporter` component, follow these steps to migrate:

1. Upgrade Sentinel and enable GPU metrics:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.agent.gpu.enabled=true
```

2. Uninstall the old component:

```shell
helm uninstall gpu-metrics-exporter -n atlas-agent
```

> ⚠️ Important
>
> Sentinel detects if `gpu-metrics-exporter` is present in the cluster and will not report GPU metrics while it remains installed. Uninstall the old component for GPU metrics to start flowing.

## Resource configuration

For large clusters or environments with intensive monitoring requirements, you may need to increase the resources allocated to Sentinel:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.resources.requests.cpu=100m \
  --set autoscaler.atlas-sentinel.controller.resources.requests.memory=2Gi \
  --set autoscaler.atlas-sentinel.controller.resources.limits.memory=2Gi
```

The agent runs as a DaemonSet on every node, with a default memory request of 64Mi and a limit of 512Mi. To change the agent resources:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.resources.requests.memory=128Mi \
  --set autoscaler.atlas-sentinel.agent.resources.limits.memory=1Gi
```

Adjust these values based on your specific cluster size and monitoring needs.

## Cluster proxy

The cluster proxy lets Atlas query the Kubernetes API of your cluster on demand. For example, [AtlasPilot Assistant](atlaspilot-assistant.md) uses the proxy to read container logs during troubleshooting. The proxy runs with read-only access. For its permissions, see [Sentinel Cluster Proxy permissions](kubernetes-permissions.md).

To enable the cluster proxy:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cluster-proxy-enabled=true
```

The chart creates the ServiceAccount and the read-only role for the proxy automatically.

Proxy access uses a token that expires after 900 seconds by default. To change the token lifetime:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cluster-proxy-token-expiration=1800
```

## Container runtime settings

Sentinel reads the container runtime socket to collect container details such as image digests. Clusters that use containerd with the default socket path need no configuration.

If your runtime socket is at a different path on your nodes, set the socket path:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.containerdSocketPath=<your-socket-path>
```

On clusters that use CRI-O, also disable the containerd-specific features:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.agent.disableContainerD=true
```

Sentinel then does not collect process trees and does not resolve image digests.

## Combine multiple features

You can enable multiple features and configure scan intervals in a single Helm command:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.enabled=true \
  --set autoscaler.atlas-sentinel.agent.enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.netflow-enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.stats-enabled=true \
  --set autoscaler.atlas-sentinel.agent.extraArgs.storage-stats-enabled=true \
  --set autoscaler.atlas-sentinel.agent.gpu.enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-concurrent-scans=6 \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-interval=30s \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-linter-scan-interval=60s
```

### Verify your configuration

To verify that your configuration changes have been applied correctly:

```shell
helm get values atlas -n atlas-agent
```

This command displays the current configuration values for Sentinel.

## Troubleshooting

If you encounter issues after changing Sentinel's configuration:

### Check controller logs

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent
```

### Check agent logs

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-agent -n atlas-agent
```

### Restart the pods

If configuration changes don't seem to be taking effect, you can restart the Sentinel pods:

```shell
kubectl rollout restart deployment atlas-sentinel-controller -n atlas-agent
kubectl rollout restart daemonset atlas-sentinel-agent -n atlas-agent
```

<Cards columns={2}>
  <Card title="Vulnerabilities" href="https://docs.atlas.example/docs/image-security" icon="fa-shield-halved">
    Identify container image security issues across your cluster.
  </Card>
  <Card title="Compliance" href="https://docs.atlas.example/docs/best-practices" icon="fa-clipboard-check">
    Assess your cluster against security best practices.
  </Card>
</Cards>