---
updatedAt: 2026-09-10T08:10:32.000Z
---

# Helm charts

Install and manage Atlas components in your Kubernetes cluster using the umbrella Helm chart or individual charts.

This page covers how to install and manage Atlas components using Helm. The recommended approach is the **umbrella Helm chart** (`atlas-helm/atlas`), which installs all required components for your chosen operating mode in a single release. Individual chart installs are available for advanced use cases.

> 📘 Simplest option: atlasctl
>
> For the fastest way to connect a cluster, use the [atlasctl CLI](connect-with-atlasctl.md). It auto-detects your cloud provider, handles IAM setup, and runs the Helm install in a single command. The steps below are for users who prefer direct Helm control or need GitOps-compatible manifests.

## Prerequisites

* `helm` v3.14.0 or higher (required for the `--reset-then-reuse-values` flag)
* `kubectl` configured for your target cluster
* A Atlas **organization-level API key** from <Anchor label="console.atlas.example" target="_blank" href="https://console.atlas.example/organization/management/access-control/service-accounts">console.atlas.example → Service Accounts</Anchor>

Add the Atlas Helm repository:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

## Umbrella chart

The umbrella chart (`atlas-helm/atlas`) uses **Helm tags** to control which components are installed. Each tag activates a different operating mode. You switch between modes by flipping tags in a single `helm upgrade` command.

For provider-specific walkthroughs that include the Terraform configuration needed for node autoscaling and full modes, see the GitOps onboarding guides:

<Cards columns={3}>
  <Card title="EKS via GitOps" href="https://docs.atlas.example/docs/terraform-provider-eks" icon="fa-brands fa-aws">
    Onboard an EKS cluster with the umbrella chart and Terraform.
  </Card>
  <Card title="GKE via GitOps" href="https://docs.atlas.example/docs/terraform-provider-gke" icon="fa-brands fa-google">
    Onboard a GKE cluster with the umbrella chart and Terraform.
  </Card>
  <Card title="AKS via GitOps" href="https://docs.atlas.example/docs/terraform-provider-aks" icon="fa-brands fa-microsoft">
    Onboard an AKS cluster with the umbrella chart and Terraform.
  </Card>
</Cards>

### Operating modes

Choose the mode that matches the level of automation you need:

* **`readonly`** — Cost visibility and security telemetry only. No changes to workloads or nodes. Start here if you want to evaluate Atlas before enabling automation.
* **`workload-autoscaler`** — Automatically right-sizes workload CPU and memory requests based on actual usage. No cloud IAM setup or Terraform required.
* **`node-autoscaler`** — Automated node provisioning, bin-packing, and Spot Instance management. Requires Terraform for cloud IAM. Includes Container Live Migration.
* **`full`** — Combines node autoscaling with Workload Autoscaler. The most complete mode.

| Component                              | `readonly` | `workload-autoscaler` | `node-autoscaler` | `full` |
| :------------------------------------- | :--------- | :-------------------- | :---------------- | :----- |
| atlas-agent                           | Yes        | Yes                   | Yes               | Yes    |
| atlas-spot-handler                    | Yes        | Yes                   | Yes               | Yes    |
| atlas-sentinel                          | Yes        | Yes                   | Yes               | Yes    |
| atlas-cluster-controller              | —          | Yes                   | Yes               | Yes    |
| atlas-evictor                         | —          | Yes                   | Yes               | Yes    |
| atlas-pod-mutator                     | —          | Yes                   | Yes               | Yes    |
| atlas-workload-autoscaler             | —          | Yes                   | —                 | Yes    |
| atlas-workload-autoscaler-exporter    | —          | Yes                   | —                 | Yes    |
| atlas-pod-pinner                      | —          | —                     | Yes               | Yes    |
| atlas-live (Container Live Migration) | —          | —                     | Yes               | Yes    |

> 📘 Note
>
> Only one mode tag should be `true` at a time. The `full` mode combines node autoscaling with Workload Autoscaler.

### Install in read-only mode

Read-only mode installs observability components. No Terraform or cloud IAM setup is needed.

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="<eks|gke|aks>" \
  --set tags.readonly=true
```

Replace `<eks|gke|aks>` with your cloud provider.

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  You should see pods for `atlas-agent`, `atlas-spot-handler`, and `atlas-sentinel` in a Running state.
</Accordion>

### Install in Workload Autoscaler mode

Workload Autoscaler mode automatically right-sizes workload CPU and memory requests based on actual usage. No Terraform or cloud IAM setup is needed.

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="<eks|gke|aks>" \
  --set tags.workload-autoscaler=true
```

Configure scaling policies and workload-level settings from the Atlas console or via [annotations](workload-autoscaler-annotations-reference.md).

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  You should see the shared base components (`atlas-agent`, `atlas-spot-handler`, `atlas-sentinel`) plus `atlas-cluster-controller`, `atlas-evictor`, `atlas-pod-mutator`, `atlas-workload-autoscaler`, and `atlas-workload-autoscaler-exporter`.
</Accordion>

### Install in node autoscaler or full mode

Node autoscaler and full modes require cloud IAM resources for node management. Use Terraform to create them, then install the Helm release. See the provider-specific GitOps guides linked above for the complete Terraform configuration:

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="<eks|gke|aks>" \
  --set tags.full=true
```

Replace `tags.full=true` with `tags.node-autoscaler=true` if you do not need Workload Autoscaler.

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  In full mode, you should see all components from the [mode table](#operating-modes) running, including `atlas-live` and both Workload Autoscaler components in addition to the shared base.

  For node-autoscaler mode, you should see the shared base components plus `atlas-cluster-controller`, `atlas-evictor`, `atlas-pod-mutator`, and `atlas-live`.
</Accordion>

### Upgrade between modes

Flip the relevant tags and pass `--reset-then-reuse-values`. The chart enforces that exactly one mode tag is `true` at a time — if multiple tags are enabled, `helm upgrade` fails with a validation error. Always explicitly disable the previous mode when enabling the new one. Existing component overrides carry forward automatically.

For example, upgrading from read-only to full:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set tags.readonly=false \
  --set tags.full=true
```

Or from workload-autoscaler to full:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set tags.workload-autoscaler=false \
  --set tags.full=true
```

> 📘 Note
>
> Moving to `node-autoscaler` or `full` from a Helm-only mode requires completing the Terraform configuration first, since those modes need cloud IAM resources. See the provider-specific GitOps guides above.

### Upgrade to a newer chart version

To upgrade all components to the latest chart version while preserving your existing values:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

### Customize chart values

The umbrella chart passes values through to each sub-component. Create a `values.yaml` file and use the `-f` flag:

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="<eks|gke|aks>" \
  --set tags.readonly=true \
  -f custom-values.yaml
```

Example `custom-values.yaml`:

```yaml
global:
  tolerations:
    - key: "dedicated"
      operator: "Equal"
      value: "atlas"
      effect: "NoSchedule"
```

You can also override values for a specific component using the component's key as a prefix. For example, to set a custom resource limit on the evictor:

```yaml
autoscaler:
  atlas-evictor:
    resources:
      limits:
        memory: 512Mi
```

### Inject a custom CA certificate

If your environment requires components to trust a private CA — for example, to pull images from a private registry over TLS — you can provide a CA certificate via a Kubernetes secret and reference it with `trustedCACertSecretRef` on each component.

> ⚠️ Key name inconsistency
>
> Components expect the certificate under different secret keys:
>
> | Secret key         | Components                                                         |
> | :----------------- | :----------------------------------------------------------------- |
> | `TLS_CA_CERT_FILE` | `atlas-agent`, `atlas-cluster-controller`, `atlas-spot-handler` |
> | `TLS_CA_CERT`      | `atlas-workload-autoscaler`                                       |
>
> Until this is standardized, populate the secret with **both** key names pointing to the same certificate file.
>
> `atlas-sentinel` does not support `trustedCACertSecretRef`. `atlas-live` mounts the secret as a volume and does not read an environment variable key, so no special key handling is required for it.

Create the secret with both keys:

```shell
kubectl create secret generic atlas-ca-cert \
  -n atlas-agent \
  --from-file=TLS_CA_CERT_FILE=ca.crt \
  --from-file=TLS_CA_CERT=ca.crt
```

Then reference the secret across all supporting components when upgrading the umbrella chart:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-agent.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler.atlas-cluster-controller.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler.atlas-spot-handler.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler.atlas-workload-autoscaler.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler.atlas-live.trustedCACertSecretRef="atlas-ca-cert"
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`. The Anywhere bundle includes a different set of components, so reference the secret on the components your cluster runs:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-agent.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler-anywhere.atlas-cluster-controller.trustedCACertSecretRef="atlas-ca-cert" \
  --set autoscaler-anywhere.atlas-workload-autoscaler.trustedCACertSecretRef="atlas-ca-cert"
```

  </Tab>
</Tabs>

## Individual chart installs (advanced)

In most cases, the umbrella chart is the recommended approach. Individual chart installs are available for environments where you need to manage each component's lifecycle independently, such as when using a GitOps tool that requires separate Helm releases per component.

> ⚠️ Warning
>
> When installing components individually, you are responsible for ensuring that all required components for your desired functionality are installed and kept compatible. The umbrella chart handles this automatically.

Create the `atlas-agent` namespace before installing the first component if it does not already exist:

```shell
kubectl create namespace atlas-agent --dry-run=client -o yaml | kubectl apply -f -
```

### Cluster controller

```shell
helm upgrade -i cluster-controller atlas-helm/atlas-cluster-controller -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>"
```

For **AKS** clusters, add `--set aks.enabled=true`.

### Evictor

```shell
helm upgrade -i atlas-evictor atlas-helm/atlas-evictor -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>" \
  --set dryRun=false
```

Learn more about [bin-packing and Evictor](evictor.md).

### Spot Handler

```shell
helm upgrade -i atlas-spot-handler atlas-helm/atlas-spot-handler -n atlas-agent --create-namespace \
  --set atlas.clusterID="<your-cluster-id>" \
  --set atlas.provider="<aws|gcp|azure>"
```

Learn more about [Spot Handler](spot-handler.md).

### Sentinel

```shell
helm upgrade -i atlas-sentinel atlas-helm/atlas-sentinel -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>" \
  --set controller.extraArgs.cloud-provider="<eks|gke|aks>" \
  --set atlas.apiURL="https://api.atlas.example"
```

> 📘 Note
>
> If your cluster is not in the US region, set `atlas.apiURL` to your regional endpoint. See [Communication requirements](communication-requirements.md) for the full list.

### Workload Autoscaler

Workload Autoscaler consists of two charts that work together:

```shell
helm upgrade -i atlas-workload-autoscaler atlas-helm/atlas-workload-autoscaler -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>"
```

```shell
helm upgrade -i atlas-workload-autoscaler-exporter atlas-helm/atlas-workload-autoscaler-exporter -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>"
```

When upgrading, update both charts together:

```shell
helm repo update atlas-helm
helm upgrade -i atlas-workload-autoscaler atlas-helm/atlas-workload-autoscaler -n atlas-agent --reset-then-reuse-values
helm upgrade -i atlas-workload-autoscaler-exporter atlas-helm/atlas-workload-autoscaler-exporter -n atlas-agent --reset-then-reuse-values
```

### Container Live Migration controller

```shell
helm upgrade -i atlas-live atlas-helm/atlas-live -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>"
```

For Container Live Migration requirements and supported providers, see [Container Live Migration](clm-overview.md).

### Keeper (Karpenter integration)

```shell
helm upgrade -i atlas-keeper atlas-helm/atlas-keeper -n atlas-agent --create-namespace \
  --set atlas.apiKey="<your-api-key>" \
  --set atlas.clusterID="<your-cluster-id>"
```

See [Karpenter Enterprise suite](karpenter-enterprise.md) for details.

## Next steps

<Cards columns={2}>
  <Card title="Hosted components" href="https://docs.atlas.example/docs/hosted-components" icon="fa-cubes">
    Review the components installed in your cluster and how they are upgraded.
  </Card>
  <Card title="Component control" href="https://docs.atlas.example/docs/component-control" icon="fa-sliders">
    Manage component versions and automatic updates from the Atlas console.
  </Card>
  <Card title="Workload Autoscaler" href="https://docs.atlas.example/docs/workload-autoscaling-configuration" icon="fa-chart-line">
    Configure scaling policies and workload-level settings.
  </Card>
  <Card title="Node templates" href="https://docs.atlas.example/docs/node-templates" icon="fa-layer-group">
    Review instance constraints and node configuration for autoscaling.
  </Card>
</Cards>