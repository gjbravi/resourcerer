---
updatedAt: 2026-09-24T08:16:37.000Z
---

# Terraform provider

 Register your cluster and provision cloud IAM resources with the Atlas Terraform provider, then install Atlas components using the Helm chart.

Atlas publishes a Terraform provider on the <Anchor label="Terraform Registry" target="_blank" href="https://registry.terraform.io/namespaces/atlas">Terraform Registry</Anchor> for registering clusters and provisioning the cloud IAM resources that Atlas needs to manage nodes on your behalf. Source code and release notes are available in the <Anchor label="GitHub repository" target="_blank" href="https://github.com/atlas/terraform-provider-atlas">GitHub repository</Anchor>.

## The umbrella Helm chart

Atlas provides an umbrella Helm chart (`atlas-helm/atlas`) that installs all required components through a single Helm release. Instead of managing individual per-component charts, you select an operating mode using Helm tags and upgrade between modes with a single `helm upgrade` command.

The umbrella chart supports the following modes:

| Tag                             | Use case                                                     |
| ------------------------------- | ------------------------------------------------------------ |
| `tags.readonly=true`            | Cost Monitoring. No changes to workloads or nodes.           |
| `tags.workload-autoscaler=true` | Right-size workload CPU/memory requests automatically.       |
| `tags.node-autoscaler=true`     | Autoscale nodes (provision, bin-pack, evict workloads).      |
| `tags.full=true`                | Node autoscaling + Workload Autoscaler + additional features |

Only one tag should be `true` at a time. Component overrides use the stable `autoscaler.*` prefix regardless of which mode is active, so customizations persist across mode upgrades without re-applying them.

> 📘 Note
>
> The exact components installed by each mode vary by cloud provider. See the provider-specific guides below for details.

## When does Terraform enter the picture?

The umbrella chart GitOps flow is incremental. Read-only and Workload Autoscaler modes require only a Atlas API key and Helm. Terraform is introduced when you need node autoscaling, because that mode requires cloud IAM resources (such as a service account with node provisioning permissions) that Terraform provisions for you.

| Mode                   | Terraform in this flow? | What Terraform does                                                        |
| ---------------------- | ----------------------- | -------------------------------------------------------------------------- |
| Read-only              | No                      | —                                                                          |
| Workload Autoscaler    | No                      | —                                                                          |
| Node Autoscaler / Full | Yes                     | Creates a cloud service account with IAM permissions for node provisioning |

> 📘 Note
>
> Terraform is not the only way to onboard a cluster or provision IAM resources. The Atlas console onboarding script handles this automatically. The guides below are for teams that manage infrastructure as code with Terraform and deploy Atlas components via GitOps tools

## Setting up API access

For long-term use with Terraform-managed clusters, create a **Full Access API Key**:

1. Log in to the <Anchor label="Atlas console" target="_blank" href="https://console.atlas.example">Atlas console</Anchor>.
2. Navigate to **API Access Keys** under your user.
3. Create a new Full Access API Key.

For detailed instructions, refer to the [authentication documentation](https://docs.atlas.example/atlas/docs/authentication#obtaining-api-access-key).

## Migrating module-managed clusters to the umbrella chart

If your cluster was onboarded with the Atlas Terraform cluster modules (`atlas-eks-cluster`, `atlas-gke-cluster`, or `atlas-aks`) using standalone Helm releases, migrate to the umbrella chart through the module itself rather than atlasctl. Set `umbrella_enabled=true` and apply, following the <Anchor label="module umbrella migration guide" target="_blank" href="https://github.com/atlas/terraform-provider-atlas/blob/master/docs/umbrella-migration/README.md">module umbrella migration guide</Anchor>. The migration is not automatic: it only runs when you set the variable.

This requires the module major version that introduced umbrella support: `~> 15.0` for EKS, `~> 11.0` for GKE, `~> 12.0` for AKS. Do not use a newer major version. A future major version will remove standalone chart support, so migrate while on these versions.

## Provider-specific guides

Each guide walks through the full onboarding flow, from read-only mode (Helm-only) through node autoscaling (Terraform + Helm), using a single chart.

<Cards>
  <Card title="GKE + Atlas GitOps with Helm chart" href="https://docs.atlas.example/docs/terraform-provider-gke" icon="fa-cloud">
    Walk through the full onboarding flow for GKE, from read-only mode through node autoscaling.
  </Card>

  <Card title="EKS + Atlas GitOps" icon="fa-cloud">
    Coming soon
  </Card>

  <Card title="AKS + Atlas GitOps" icon="fa-cloud">
    Coming soon
  </Card>
</Cards>

<br />