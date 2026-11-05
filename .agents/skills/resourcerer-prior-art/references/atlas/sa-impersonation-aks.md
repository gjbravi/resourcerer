---
updatedAt: 2026-09-22T08:46:05.000Z
---

# AKS workload identity impersonation

> 🚧 Limited Availability Feature
>
> This feature is currently available through feature flags. Contact us to enable access for your organization.

> 📘 Note
>
> The feature flag must be enabled before running `terraform apply`. `terraform plan` will succeed without it, but `terraform apply` will fail with a 403 error.

Atlas supports Azure Workload Identity as an alternative to App Registration for AKS clusters onboarded with impersonation. Workload Identity uses federated credentials instead of client secrets, eliminating secret rotation and allowing Atlas's permissions to be scoped to a specific resource group rather than the whole subscription.

The UI script method also supports it. `atlasctl` does not currently support it.

## How it works

Instead of creating an Azure App Registration and client secret, Atlas uses a User-Assigned Managed Identity with a federated credential. The federated credential allows Atlas's own identity to impersonate the managed identity — no secrets are stored or rotated.

Terraform creates the following Azure resources:

* A User-Assigned Managed Identity in your cluster's resource group
* A federated identity credential on that managed identity
* A `AtlasAKSRole` custom role definition
* Role assignments scoped to the cluster resource group, the node resource group (`MC_*`), and — when the cluster's VNet or subnet is in a different resource group than the cluster — the VNet resource group

> 📘 Note
>
> To cover a VNet that is in a different resource group than the cluster, add that resource group via the `additional_resource_groups` variable of the [Atlas Terraform module for AKS](https://registry.terraform.io/modules/atlas/aks/atlas/latest). See [Resource groups that `AtlasAKSRole` is assigned to](cloud-permissions.md#resource-groups-that-atlasaksrole-is-assigned-to) for details.

## Terraform onboarding with workload identity

Use the [aks\_cluster\_workload\_identity example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/aks/aks_cluster_workload_identity) from the Atlas Terraform provider repository. Copy it as your starting point — `authentication_method = "workload_identity"` is already set. The default is `"client_secret"` (App Registration); the module handles managed identity and federated credential creation automatically.

## Migrating an existing cluster

Clusters already onboarded via App Registration (Terraform or UI script) can migrate to workload identity without disconnecting from Atlas.

**Via Terraform:**

Follow the [aks\_cluster\_workload\_identity example](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/aks/aks_cluster_workload_identity) in the Atlas Terraform provider repository. Running `terraform apply` with workload identity configured destroys the App Registration, service principal, and application password, then creates the managed identity and federated credential as replacements. The role assignments are also replaced since they move from the service principal to the managed identity. Atlas will be unable to authenticate during this window — this typically resolves within a few minutes once the new credentials propagate.

**Via the UI script:**

For clusters originally connected through the Atlas console:

1. Go to the **Cluster List** in the Atlas console.
2. Click the **⋯** (three-dot) menu on the cluster's row.
3. Select **Atlas features** from the dropdown.
4. Copy and run the script shown in the modal.

## atlasctl limitation

`atlasctl` does not currently support AKS onboarding with workload identity impersonation. Use Terraform or the UI script method instead.

## See also

<Cards columns={3}>
  <Card title="AKS GitOps onboarding" href="https://docs.atlas.example/docs/terraform-provider-aks" icon="fa-code-branch">
    Terraform-based onboarding guide for AKS.
  </Card>
  <Card title="Cloud permissions" href="https://docs.atlas.example/docs/cloud-permissions#azure" icon="fa-shield">
    Full list of Azure permissions used by Atlas.
  </Card>
  <Card title="GKE service account impersonation" href="https://docs.atlas.example/docs/sa-impersonation-gke" icon="fa-key">
    Equivalent feature for GKE clusters.
  </Card>
</Cards>