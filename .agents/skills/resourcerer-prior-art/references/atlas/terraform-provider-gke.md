---
updatedAt: 2026-08-04T08:19:56.000Z
---

# GKE via GitOps

Onboard a GKE cluster to Atlas using the umbrella Helm chart and Terraform. Start in read-only mode with Helm alone and upgrade to node autoscaling when ready.

This guide covers onboarding a GKE cluster to Atlas using a GitOps approach with the umbrella Helm chart (`atlas-helm/atlas`).

The umbrella chart uses **Helm tags** to control which Atlas components are installed. Each tag activates a different operating mode. For example, `--set tags.readonly=true` installs only observability components, while `--set tags.full=true` installs the complete suite for GKE. You switch between modes by flipping tags in a single `helm upgrade` command. For more background on the umbrella chart, see [Terraform provider](terraform-provider.md).

Choose the mode that fits your needs. If you start with a lighter mode, you can upgrade later without reinstalling (see [Upgrading between modes](#upgrading-between-modes)).

## Umbrella chart modes on GKE

The table below shows which components each mode installs on GKE.

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
* `kubectl` configured for your target GKE cluster
* The `atlas-helm` Helm repository added:

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
```

**Node autoscaler and full modes** additionally require:

* `terraform` v1.3.2 or higher
* `gcloud` configured with permissions to create IAM resources in your GCP project

## Installing in read-only mode (Helm only)

Read-only mode installs observability components that let you monitor your cluster on Atlas without changing workloads or nodes. No Terraform is needed.

```shell
helm upgrade -i atlas atlas-helm/atlas -n atlas-agent --create-namespace \
  --set global.atlas.apiKey="<your-atlas-api-key>" \
  --set global.atlas.provider="gke" \
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
  --set global.atlas.provider="gke" \
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

Node autoscaler and full modes enable node provisioning, bin-packing, Spot Instance handling, and workload eviction. These modes require a GCP service account with IAM permissions for node management. Terraform creates it.

* **`tags.node-autoscaler=true`** for Node autoscaling only (includes Container Live Migration).
* **`tags.full=true`** for Node autoscaling + Workload Autoscaler (includes Container Live Migration).

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
  }
}
```

**`providers.tf`**

```toml
provider "atlas" {
  api_url   = var.atlas_api_url
  api_token = var.atlas_api_token
}
```

**`variables.tf`**

```toml
variable "project_id" {
  type        = string
  description = "GCP project ID in which GKE cluster is located."
}

variable "cluster_name" {
  type        = string
  description = "GKE cluster name in GCP project."
}

variable "cluster_region" {
  type        = string
  description = "Region of the cluster to be connected to Atlas."
}

variable "subnets" {
  type        = list(string)
  description = "Subnet IDs used by Atlas to provision nodes."
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
  default     = "https://api.atlas.example" # See API access doc for regional endpoints
}
```

**`atlas.tf`**

```toml
resource "atlas_gke_cluster" "this" {
  project_id                 = var.project_id
  location                   = var.cluster_region
  name                       = var.cluster_name
  delete_nodes_on_disconnect = var.delete_nodes_on_disconnect
  credentials_json           = module.atlas-gke-iam.private_key
}

module "atlas-gke-iam" {
  source  = "atlas/gke-iam/atlas"
  version = "~> 0.5"
  project_id       = var.project_id
  gke_cluster_name = var.cluster_name
}

resource "atlas_node_configuration" "default" {
  cluster_id     = atlas_gke_cluster.this.id
  name           = "default"
  disk_cpu_ratio = 0
  min_disk_size  = 100
  subnets        = var.subnets
}

resource "atlas_node_configuration_default" "this" {
  cluster_id       = atlas_gke_cluster.this.id
  configuration_id = atlas_node_configuration.default.id
}
```

**`outputs.tf`**

```toml
output "cluster_id" {
  value       = atlas_gke_cluster.this.id
  description = "Atlas cluster ID."
}

output "cluster_token" {
  value       = atlas_gke_cluster.this.cluster_token
  description = "Atlas cluster token used by Atlasware to authenticate to SaaS."
  sensitive   = true
}
```

**`tf.vars.example`**

```toml
atlas_api_token = "PLACEHOLDER"
project_id       = "PLACEHOLDER"
cluster_region   = "PLACEHOLDER"   # e.g. "us-central1" or "us-central1-a"
cluster_name     = "PLACEHOLDER"
subnets          = ["PLACEHOLDER"] # e.g. ["default"]
```

### Run Terraform

Copy the example variables file, fill in your values, and apply:

```shell
cp tf.vars.example terraform.tfvars
# Edit terraform.tfvars with your GCP project details
```

```shell
terraform init
terraform apply
```

Terraform registers the cluster with Atlas and creates the GCP service account. Capture the outputs. You'll need them to configure the Helm release

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
  --set global.atlas.provider="gke" \
  --set tags.full=true
```

Replace `tags.full=true` with `tags.node-autoscaler=true` if you do not need Workload Autoscaler.

<Accordion title="Verify the installation">
  ```shell
  kubectl get pods -n atlas-agent
  ```

  In full mode, you should see all components from the [mode table](#umbrella-chart-modes-on-gke) running, including `atlas-live` and both Workload Autoscaler components in addition to the shared base.

  For node-autoscaler mode, you should see the shared base components plus `atlas-cluster-controller`, `atlas-evictor`, `atlas-pod-mutator`, and `atlas-live`.
</Accordion>

## Upgrading between modes

If you installed a lighter mode and later want to move to a more capable one, flip the relevant tags and pass `--reset-then-reuse-values`. Any component overrides you previously set carry forward automatically as well.

> 📘 Note
>
> Moving to `node-autoscaler` or `full` from a Helm-only mode requires completing the [Terraform configuration](#create-the-terraform-configuration) first, since those modes need GCP IAM resources.

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