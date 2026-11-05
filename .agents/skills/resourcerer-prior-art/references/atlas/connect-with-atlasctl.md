---
updatedAt: 2026-09-23T08:06:35.000Z
---

# Connect using the atlasctl CLI

Connect or disconnect a Kubernetes cluster from Atlas using the atlasctl command-line tool.

`atlasctl` is the recommended way to connect Kubernetes clusters to Atlas. It auto-detects your cloud provider, cluster name, and region from your kubeconfig, walks you through feature selection, and handles cluster registration, IAM setup, and the Helm install in a single command.

You can also connect clusters from the [Atlas console](connect-via-console.md) or with the [Terraform provider](terraform-provider.md).

## What atlasctl supports

* **Cloud providers**: Amazon EKS, Google GKE, Microsoft AKS, and [Atlas Anywhere](atlas-anywhere-overview.md) (on-premises, bare metal, and other clouds).
* **Features**: Cost Monitoring (always enabled), [Node Autoscaling](autoscaler.md), and [Workload Autoscaling](workload-autoscaling-overview.md).
* **Karpenter**: If Karpenter is already installed in the cluster, `atlasctl` offers to keep it as your node provisioner via [Atlas for Karpenter](karpenter-enterprise.md).

> 📘 Plan requirements
>
> Node Autoscaling and Workload Autoscaling are Premium features. Cost Monitoring is available on every plan.

## Before you start

Before you connect a cluster with `atlasctl`, make sure you have:

* A Atlas account ([sign up](https://console.atlas.example/)).
* A Kubernetes cluster on EKS, GKE, AKS, or a [Atlas Anywhere](atlas-anywhere-overview.md) supported environment.
* `kubectl` configured for that cluster with admin access (`kubectl auth can-i '*' '*'`).
* Cloud-side permissions equivalent to the console connection script's least-privileged policy. See [Cloud permissions](cloud-permissions.md).
* `metrics-server` installed in the cluster, if you plan to enable Workload Autoscaling.

`atlasctl` bundles its own Helm logic, so you do **not** need `helm`, `jq`, or `yq` installed locally.

## Install atlasctl

<Tabs>
  <Tab title="Homebrew (macOS / Linux)">
    ```shell
    brew tap atlas/tap
    brew trust --cask atlas/tap/atlasctl
    brew install atlasctl
    ```

    To upgrade to the latest version:

    ```shell
    brew update && brew upgrade atlasctl
    ```
  </Tab>

  <Tab title="Install script (macOS / Linux)">
    ```shell
    curl -fsSL https://get.atlas.example/linux | bash
    ```

    The script installs `atlasctl` to `~/.local/bin` by default and verifies the SHA256 checksum.

    Pin a specific version or change the install location:

    ```shell
    # Install to /usr/local/bin
    curl -fsSL https://get.atlas.example/linux \
      | bash
    ```
  </Tab>

  <Tab title="Windows (PowerShell)">
    ```powershell
    iwr -useb https://get.atlas.example/windows | iex
    ```
  </Tab>
</Tabs>

Verify the installation:

```shell
atlasctl version
```

## Log in

Authenticate `atlasctl` with Atlas using your browser:

1. Run the login command:

   ```shell
   atlasctl auth login
   ```

2. `atlasctl` opens your default browser and asks to sign in to your Atlas account.

3. After authentication, control returns to your terminal. If you belong to more than one organization, pick one from the interactive list. If you belong to a single organization, it's selected automatically.

4. `atlasctl` saves your token and organization ID to `~/.config/atlasctl/config.yaml` and prints a confirmation message.

> 📘 Switching account or organization
>
> Run `atlasctl auth login` again. It overwrites the stored credentials.

## Connect your cluster

1. Start the interactive connection:

   ```shell
   atlasctl cluster connect
   ```

   `atlasctl` detects the cloud provider, cluster name, and region from your kubeconfig and node labels, then prints the resolved configuration for you to review.

2. Select the features you want to enable.

3. Confirm the **Proceed with connecting?** prompt. `atlasctl` then:

   1. Registers the cluster with Atlas.
   2. Installs the `atlas` Helm chart in the `atlas-agent` namespace.
   3. Creates the IAM resources required in your cloud account (AWS / GCP / Azure).

4. On success, `atlasctl` prints the cluster's Atlas console URL. Open it to verify everything is connected.

### Preview without making changes

Resolve the configuration and print it as YAML without registering the cluster or modifying anything:

```shell
atlasctl cluster connect --dry-run
```

Useful for validating that auto-detection picked the right provider, region, and cluster name.

### Non-interactive mode (CI/scripts)

In CI pipelines or scripts, disable prompts with `--non-interactive`. The process can be customized using flags passed to the binary.

```shell
atlasctl cluster connect \
  --non-interactive \
  --cluster-optimization=false \
  --workload-autoscaler=true
```

For the complete list of flags supported by the current `atlasctl`, including feature toggles, GKE credential method, and Helm values overrides, run:

```shell
atlasctl cluster connect --help
```

### Custom umbrella chart values

During cluster onboarding, `atlasctl` installs the Atlas components using an umbrella Helm chart. You can customize the chart values (such as tolerations, node selectors, or resource limits) by providing a YAML values file with the `--umbrella-values-file` flag:

```shell
atlasctl cluster connect --umbrella-values-file /path/to/custom-values.yaml
```

This flag accepts a YAML file containing Helm values that are merged with the configuration `atlasctl` generates based on your feature selections. This approach provides flexibility to configure values that `atlasctl` doesn't explicitly manage through command-line flags.

Example `custom-values.yaml`:

```yaml
global:
  tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "atlas"
      effect: "NoSchedule"
```

For a complete reference of available values in the umbrella chart, see the [atlas-umbrella Helm chart documentation](https://github.com/atlas/helm-charts/tree/main/charts/atlas-umbrella).

## Connect a Atlas Anywhere cluster

For clusters running outside EKS, GKE, or AKS, such as on-premises, bare metal, or other cloud providers, use the `--provider anywhere` flag:

```shell
atlasctl cluster connect --provider anywhere
```

Click here to learn more about [Atlas Anywhere](atlas-anywhere-overview.md).

## After connecting

* **Import your existing node configuration.** This is the recommended way to prepare a newly-connected cluster: `atlasctl cluster import-config` discovers your existing node groups and creates matching node configurations and node templates automatically, so Atlas provisions nodes that match what you already run. See [Import and apply cluster configuration with atlasctl](atlasctl-import-cluster-configuration.md).
* **Review the Node Autoscaler checklist.** Whether or not you import existing config, check the [Node Autoscaler preparation checklist](autoscaler-checklist.md) for the workload-level requirements (PodDisruptionBudgets, startup probes) automated import can't set for you.
* **Migrating from standalone Helm charts?** If the cluster was previously connected using individual Atlas Helm releases, `atlasctl cluster migrate` consolidates them into the unified umbrella chart. This causes brief downtime: Node Autoscaling and Workload Autoscaling pause while the migration runs. See [Migrate standalone Helm charts with atlasctl](atlasctl-migrate-standalone-helm-charts.md). If those releases are managed by ArgoCD, Flux, or Terraform, see [Migrate GitOps and IaC installations with atlasctl](atlasctl-migrate-gitops.md) instead. It is a phased flow with no downtime.

For the complete list of atlasctl commands and flags, see the [atlasctl command reference](atlasctl-reference.md).

## Disconnect a cluster

> ⚠️ This removes Atlas completely
>
> `atlasctl cluster disconnect` performs the equivalent of both **Disconnect cluster** and **Remove cluster** in the Atlas console: it uninstalls the Atlas Helm releases, deletes the IAM resources created during the connection, and removes the cluster record from Atlas.

> 🚧 Not for Terraform-managed clusters
>
> Do **not** run `atlasctl cluster disconnect` against a cluster that was connected with the [Atlas Terraform provider](terraform-provider.md). The CLI does not coordinate with Terraform state and will leave your Terraform configuration out of sync. For Terraform-managed clusters, run `terraform destroy` instead. See [Disconnect your cluster](disconnect-your-cluster.md#terraform-managed-clusters).
>
> Clusters connected via the Atlas console or `atlasctl` are supported.

1. Switch your kubeconfig to the cluster you want to disconnect.

2. Run the disconnect command:

   ```shell
   atlasctl cluster disconnect
   ```

   `atlasctl` reads the Atlas cluster ID from the in-cluster configuration, so you usually do not need to provide it.

3. Review the summary box. `atlasctl` prints every cloud resource it removed, any resources that were already absent, and any warnings to follow up on.

If a step fails, `atlasctl` prints a failure box noting which step failed and which cloud resources were already cleaned up. Fix the underlying issue and re-run the command. Partial progress is preserved.

For the full list of disconnect options (including flags to skip cloud cleanup or the Kubernetes uninstall), run:

```shell
atlasctl cluster disconnect --help
```

## See also

<Cards columns={2}>
  <Card title="Import and apply cluster configuration with atlasctl" icon="fa-download" href="https://docs.atlas.example/docs/atlasctl-import-cluster-configuration">Match Atlas-provisioned nodes to your existing node groups right after connecting.</Card>
  <Card title="Migrate standalone Helm charts with atlasctl" icon="fa-arrows-rotate" href="https://docs.atlas.example/docs/atlasctl-migrate-standalone-helm-charts">For clusters already connected with individual Atlas Helm releases.</Card>
</Cards>