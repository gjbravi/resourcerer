---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Installation & upgrading

This guide walks you through the process of installing and upgrading Sentinel in your Kubernetes clusters. Sentinel can be installed using various methods, including the Atlas console UI, Terraform, or Helm.

## Prerequisites

Before installing Sentinel, ensure you have:

* A Kubernetes cluster connected to Atlas
* Appropriate permissions to create resources in your cluster
* Required tools for your chosen installation method (kubectl, helm, etc.)

## Installation methods

### Atlas console

The simplest way to install Sentinel is through the Atlas console:

1. Log in to the Atlas console
2. Select a cluster and navigate to **Settings**
3. Sentinel is installed as part of the Atlas agent stack. Verify it is running in your cluster

### Terraform

If you manage your infrastructure as code, you can use our Terraform modules for GKE, EKS, and AKS to install Sentinel. Set the `install_security_agent=true` module input variable:

```toml Terraform
module "atlas-eks-cluster" {
  // ...
  install_security_agent = true
}
```

For complete examples, see:

* [EKS module](https://registry.terraform.io/modules/atlas/eks-cluster/atlas/latest#input_install_security_agent), [Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/eks/eks_cluster_with_security/atlas.tf)
* [AKS module](https://registry.terraform.io/modules/atlas/aks/atlas/latest#input_install_security_agent), [Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/aks/aks_cluster_with_security/atlas.tf)
* [GKE module](https://registry.terraform.io/modules/atlas/gke-cluster/atlas/latest#input_install_security_agent), [Terraform example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/gke/gke_cluster_with_security/atlas.tf)

### Helm

To install Sentinel using Helm, follow these steps:

1. Add the Atlas Helm repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

2. (Optional) List available components and versions:

```shell
helm search repo atlas-helm
```

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

3. Install Sentinel:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-linter-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-bench-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=<aks|eks|gke>
```

Replace the placeholders with your actual values:

* `<your-api-token>`: Your Atlas API token
* `<your-cluster-id>`: Your cluster's ID in Atlas
* `<aks|eks|gke>`: Your Kubernetes provider (leave empty if not using these platforms)

## Upgrading Sentinel

### Atlas console

The easiest way to keep Sentinel up to date is through the Component Control dashboard:

1. In the Atlas console, select **Manage Organization** in the top right
2. Navigate to **Component control** in the left menu
3. Find Sentinel in the component list
4. For any cluster showing a warning status, click on the component to see details
5. Click the **Update** button to get the upgrade script
6. Copy and run the provided command in your terminal

![Component Control Dashboard](https://files.readme.io/350fccd19a16026145fca209fdf514b73ed2108a7973f6df4c247ef4628ddf8c-image.png)

This approach ensures you're always running the latest version with all security updates and feature enhancements.

### Helm

If you installed Sentinel with Helm, you can upgrade it with:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

The `--reset-then-reuse-values` flag preserves your existing configuration while applying the latest updates.

#### Upgrading from v0.x.x to v1.x.x

The Sentinel v1 Helm chart contains breaking changes requiring a fresh installation. Follow these steps:

1. Get your current Sentinel values:

```shell
helm get values atlas -n atlas-agent
```

2. Note your `clusterID`, `apiKey`, and kube-bench provider for use in the next step.

3. Reinstall Sentinel with your values:

```shell
helm repo update atlas-helm

helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-sentinel.atlas.apiKey=<your-api-token> \
  --set autoscaler.atlas-sentinel.atlas.clusterID=<your-cluster-id> \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-linter-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.image-scan-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.kube-bench-enabled=true \
  --set autoscaler.atlas-sentinel.controller.extraArgs.cloud-provider=<aks|eks|gke>
```

## Troubleshooting

If you encounter issues with Sentinel installation or upgrades, try these troubleshooting steps:

### Check the Sentinel controller logs

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-controller -n atlas-agent
```

### Check the Sentinel agent logs

```shell
kubectl logs -l app.kubernetes.io/name=atlas-sentinel-agent -n atlas-agent
```

### Verify your configuration

```shell
helm get values atlas -n atlas-agent
```

### Check component status in the console

You can also verify the status of Sentinel in the Atlas console by navigating to **Security** > **Settings** and checking the status indicators next to your cluster.

For additional assistance, contact Atlas support or visit our [community Slack channel](https://atlas-community.slack.com/).

## Next steps

Once you have Sentinel installed, you can configure various security features:

* [Private Image Scanning](image-scanning.md): Configure access to private or managed container image registries
* [Configuring Sentinel Features](sec-configuring-sentinel.md): Learn how to configure network monitoring, resource stats, and other features

<Cards columns={2}>
  <Card title="Configuring Sentinel" href="https://docs.atlas.example/docs/sec-configuring-sentinel" icon="fa-gear">
    Configure scanning, monitoring, and resource settings.
  </Card>
  <Card title="Sentinel" href="https://docs.atlas.example/docs/sentinel" icon="fa-shield-halved">
    Architecture, capabilities, and permissions.
  </Card>
</Cards>