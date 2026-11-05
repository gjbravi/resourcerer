---
updatedAt: 2026-09-18T07:24:49.000Z
---

# AKS via GitOps

Onboard an AKS cluster to Atlas using the umbrella Helm chart and Terraform. Choose the operating mode that fits your needs and install directly into it.

This guide covers onboarding an AKS cluster to Atlas using a GitOps approach with the umbrella Helm chart (`atlas-helm/atlas`).

The umbrella chart uses **Helm tags** to control which Atlas components are installed. Each tag activates a different operating mode — for example, `--set tags.readonly=true` installs only observability components, while `--set tags.full=true` installs the complete suite for AKS. You switch between modes by flipping tags in a single `helm upgrade` command. For more background on the umbrella chart, see [Terraform provider](terraform-provider.md).

Choose the mode that fits your needs. If you start with a lighter mode, you can upgrade later without reinstalling (see [Upgrading between modes](#upgrading-between-modes)).

## Umbrella chart modes on AKS

The table below shows which components each mode installs on AKS.

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

## Prerequisites

**All modes** require:

* A Atlas account and an **organization-level API key** from <Anchor label="console.atlas.example → Service Accounts" target="_blank" href="https://console.atlas.example/organization/management/access-control/service-accounts">console.atlas.example → Service Accounts</Anchor>
* `helm` v3.14.0 or higher (required for the `--reset-then-reuse-values` flag)
* The `atlas-helm` Helm repository added:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

**Node autoscaler and full modes** additionally require:

* `terraform` v1.3.2 or higher
* Azure CLI configured with permissions to create role assignments in your Azure subscription

## Installing in read-only mode (Helm only)

Read-only mode installs observability components that let Atlas monitor your cluster without making any changes to workloads or nodes. No Terraform is needed.

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="aks" \
  --set tags.readonly=true
```

After the pods become ready, your cluster appears as **Read only** in the Atlas console.

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  You should see pods for `atlas-agent`, `atlas-spot-handler`, and `atlas-sentinel` in a Running state.
</Accordion>

## Installing in Workload Autoscaler mode (Helm only)

Workload Autoscaler mode for automatically right-sizing workload CPU/memory requests based on actual usage. No Terraform is needed.

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="aks" \
  --set tags.workload-autoscaler=true
```

Configure scaling policies and workload-level settings from the Atlas console or via [annotations](workload-autoscaler-annotations-reference.md).

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  You should see the shared base components (`atlas-agent`, `atlas-spot-handler`, `atlas-sentinel`) plus `atlas-cluster-controller`, `atlas-evictor`, `atlas-pod-mutator`, `atlas-workload-autoscaler`, and `atlas-workload-autoscaler-exporter`.
</Accordion>

## Installing in node autoscaler or full mode (Terraform + Helm)

Node autoscaler and full modes enable node provisioning, bin-packing, Spot Instance handling, and workload eviction. These modes require Azure role assignments that grant Atlas permissions for node management. Terraform creates them.

> 📘 Note
>
> If your cluster's VNet or subnet is in a *different* resource group than the cluster, Atlas also needs permissions on that networking resource group. When you use the [Atlas Terraform module for AKS](https://registry.terraform.io/modules/atlas/aks/atlas/latest), add the VNet's resource group via the `additional_resource_groups` variable. See [Resource groups that `AtlasAKSRole` is assigned to](cloud-permissions.md#resource-groups-that-atlasaksrole-is-assigned-to).

* **`tags.node-autoscaler=true`** for Node autoscaling only (includes Container Live Migration).
* **`tags.full=true`** for Node autoscaling + Workload Autoscaler (includes Container Live Migration).

### Authentication method

The configuration below uses the **App Registration** (client secret) authentication method, which is the default. If you prefer to use **Workload Identity** — which uses federated credentials instead of client secrets and does not require secret rotation — see [AKS workload identity impersonation](sa-impersonation-aks.md).

For maintained, copy-ready Terraform examples see the [atlas/terraform-provider-atlas examples](https://github.com/atlas/terraform-provider-atlas/tree/master/examples/aks).

### Create the Terraform configuration

Create the following files in your Terraform project directory.

**`versions.tf`**

```toml
terraform {
  required_version = ">= 1.3.2"
  required_providers {
    atlas = {
      source  = "atlas/atlas"
      version = ">= 3.11.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 3.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = ">= 2.0"
    }
  }
}
```

**`providers.tf`**

```toml
provider "atlas" {
  api_url   = var.atlas_api_url
  api_token = var.atlas_api_token
}

provider "azurerm" {
  subscription_id = var.azure_subscription_id
  features {}
}

provider "azuread" {
  tenant_id = data.azurerm_subscription.current.tenant_id
}
```

**`variables.tf`**

```toml
variable "azure_subscription_id" {
  type        = string
  description = "Azure subscription ID where the AKS cluster is located."
}

variable "aks_cluster_name" {
  type        = string
  description = "AKS cluster name."
}

variable "aks_cluster_region" {
  type        = string
  description = "Region of the AKS cluster to be connected to Atlas."
}

variable "aks_resource_group" {
  type        = string
  description = "Resource group containing the AKS cluster."
}

variable "aks_node_resource_group" {
  type        = string
  description = "Resource group containing the AKS node pool resources."
}

variable "delete_nodes_on_disconnect" {
  type        = bool
  description = "Optionally delete Atlas created nodes when the cluster is destroyed."
  default     = false
}

variable "atlas_api_token" {
  type        = string
  description = "Atlas API token created in console.atlas.example API Access keys section."
}

variable "atlas_api_url" {
  type        = string
  description = "Atlas API URL."
  default     = "https://api.atlas.example"
}

variable "subnets" {
  type        = list(string)
  description = "Subnet IDs used by Atlas to provision nodes."
}
```

**`atlas.tf`**

```toml
resource "atlas_aks_cluster" "this" {
  name                       = var.aks_cluster_name
  region                     = var.aks_cluster_region
  subscription_id            = var.azure_subscription_id
  tenant_id                  = data.azurerm_subscription.current.tenant_id
  client_id                  = azuread_application.atlas.client_id
  client_secret              = azuread_application_password.atlas.value
  node_resource_group        = var.aks_node_resource_group
  delete_nodes_on_disconnect = var.delete_nodes_on_disconnect
}

data "azurerm_subscription" "current" {}

resource "azuread_application" "atlas" {
  display_name = "atlas-${var.aks_cluster_name}"
}

resource "azuread_application_password" "atlas" {
  application_id = azuread_application.atlas.id
}

resource "azuread_service_principal" "atlas" {
  client_id = azuread_application.atlas.client_id
}

resource "azurerm_role_assignment" "atlas_contributor" {
  scope                = "/subscriptions/${var.azure_subscription_id}/resourceGroups/${var.aks_node_resource_group}"
  role_definition_name = "Contributor"
  principal_id         = azuread_service_principal.atlas.object_id
}

resource "atlas_node_configuration" "default" {
  cluster_id     = atlas_aks_cluster.this.id
  name           = "default"
  disk_cpu_ratio = 0
  min_disk_size  = 100
  subnets        = var.subnets

  aks {
    max_pods_per_node = 40
  }
}

resource "atlas_node_configuration_default" "this" {
  cluster_id       = atlas_aks_cluster.this.id
  configuration_id = atlas_node_configuration.default.id
}
```

**`outputs.tf`**

```toml
output "cluster_id" {
  value       = atlas_aks_cluster.this.id
  description = "Atlas cluster ID."
}

output "cluster_token" {
  value       = atlas_aks_cluster.this.cluster_token
  description = "Atlas cluster token used by Atlasware to authenticate to SaaS."
  sensitive   = true
}
```

**`tf.vars.example`**

```toml
atlas_api_token         = "PLACEHOLDER"
azure_subscription_id    = "PLACEHOLDER"   # e.g. "12345678-1234-1234-1234-123456789012"
aks_cluster_name         = "PLACEHOLDER"
aks_cluster_region       = "PLACEHOLDER"   # e.g. "eastus"
aks_resource_group       = "PLACEHOLDER"
aks_node_resource_group  = "PLACEHOLDER"   # e.g. "MC_mygroup_mycluster_eastus"
subnets                  = ["PLACEHOLDER"] # e.g. ["/subscriptions/.../subnets/default"]
```

### Run Terraform

Copy the example variables file, fill in your values, and apply:

```shell
cp tf.vars.example terraform.tfvars
# Edit terraform.tfvars with your Azure subscription details
```

```shell
terraform init
terraform apply
```

Terraform registers the cluster with Atlas and creates the Azure AD application and role assignment. Capture the outputs. You'll need them for the Helm release:

```shell
terraform output cluster_id
terraform output -raw cluster_token
```

> ⚠️ Warning
>
> The `cluster_token` expires after a few hours if no Atlas component connects. Run the Helm install promptly after this step.

### Install the Helm release

For a fresh install, use `helm upgrade -i` with the mode tag that matches your needs:

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="aks" \
  --set tags.full=true
```

Replace `tags.full=true` with `tags.node-autoscaler=true` if you do not need Workload Autoscaler.

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  In full mode, you should see all components from the [mode table](#umbrella-chart-modes-on-aks) running, including `atlas-live` and both Workload Autoscaler components in addition to the shared base.

  For node-autoscaler mode, you should see the shared base components plus `atlas-cluster-controller`, `atlas-evictor`, `atlas-pod-mutator`, and `atlas-live`.
</Accordion>

## Custom CA certificate for MITM proxies

When your network routes Atlas node traffic through a TLS-intercepting (MITM) corporate proxy, the proxy re-signs certificates with its own CA. Atlas configures node-level components to trust that CA by default, or TLS verification fails and node bootstrap breaks.

Set the `ca_certs` variable on the `atlas/aks/atlas` Terraform module. It accepts a **list of base64-encoded PEM certificates** and generates a `ca_cert_config` block on the `atlas_aks_cluster` resource.

> Requires `atlas/aks/atlas` module **v11.2.0** or later.

```hcl
module "atlas_aks_cluster" {
  source = "atlas/aks/atlas"

  # ... other required variables ...

  # Proxy settings (required when routing node traffic through a corporate proxy).
  http_proxy  = var.http_proxy
  https_proxy = var.https_proxy
  no_proxy    = var.no_proxy

  # Read the PEM file and base64-encode its contents.
  ca_certs = [base64encode(file("${path.module}/ca.crt"))]
}
```

Pass multiple entries for a chain (for example, root plus intermediate):

```hcl
  ca_certs = [
    base64encode(file("${path.module}/root-ca.crt")),
    base64encode(file("${path.module}/intermediate-ca.crt")),
  ]
```

The module configures each node to trust the CA by default during bootstrap. No further configuration is needed on the nodes.

> 📘 Pod-level vs. node-level CA certificates
>
> This section covers **node-level** CA certificates, applied to Atlas components that run directly on VMs. For **pod-level** CA certificates — consumed by in-cluster components such as `atlas-agent` and `atlas-cluster-controller` via a Kubernetes secret and `trustedCACertSecretRef` — see [Helm charts](helm-charts.md).

## Upgrading between modes

If you installed a lighter mode and later want to move to a more capable one, flip the relevant tags and pass `--reset-then-reuse-values`. Any component overrides you previously set carry forward automatically.

> 📘 Note
>
> Moving to `node-autoscaler` or `full` from a Helm-only mode requires completing the [Terraform configuration](#create-the-terraform-configuration) first, since those modes need Azure IAM resources.

For example, upgrading from read-only to full:

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set tags.readonly=false \
  --set tags.full=true
```

The shared components that were already running from the previous mode are patched in place with no pod restarts. Only the new components required by the target mode are created.

## Next steps

<Cards columns={2}>
  <Card title="Scaling policies" href="https://docs.atlas.example/docs/woop-scaling-policies" icon="fa-sliders">
    Configure Workload Autoscaler scaling policies for your workloads.
  </Card>
  <Card title="Node templates" href="https://docs.atlas.example/docs/node-templates" icon="fa-layer-group">
    Review node templates and instance constraints for node autoscaling.
  </Card>
  <Card title="Container Live Migration" href="https://docs.atlas.example/docs/clm-overview" icon="fa-arrow-right-arrow-left">
    Set up zero-downtime node replacements with live migration.
  </Card>
  <Card title="Pod mutations" href="https://docs.atlas.example/docs/pod-mutations-overview" icon="fa-wand-magic-sparkles">
    Automate workload configuration with pod mutation policies.
  </Card>
</Cards>