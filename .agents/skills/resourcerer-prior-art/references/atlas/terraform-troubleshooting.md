---
updatedAt: 2026-09-22T08:46:05.000Z
---

# Terraform troubleshooting

Solutions for common Terraform issues when setting up and managing Kubernetes clusters with Atlas.

## Resource conflicts in state file

* **Why it happens:** Multiple Terraform executions or manual changes to cloud resources can cause the state file to become out of sync with the actual infrastructure.
* **Solution:** Use Terraform state commands to adjust the state file manually. For example, the command below removes a resource from the state file, resolving conflicts.

```toml Terraform
terraform state rm [resource_name]
```

***

## Cluster onboarding phase mismatch error

* **Error message:**

```json
Error: expected status code 200, received: status=400 body={"message":"cluster is not onboarded to phase 2", "fieldViolations":[]}
with module.eks_cluster.module.atlas[0].atlas_rebalancing_job.spots[0],
on .terraform/modules/eks_cluster.atlas/atlas.tf line 190, in resource "atlas_rebalancing_job" "spots":
190: resource "atlas_rebalancing_job" "spots" {
```

* **Why it happened:** The error indicates that a 400 Bad Request HTTP response was returned because the operation expected the cluster to be in Phase 2 of the onboarding process (connected to and managed by the Atlas platform). However, the cluster was still in Phase 1 (connected to the Atlas platform). The customer used a repository or configuration intended for clusters already onboarded to Phase 2, which caused the mismatch and subsequent error.
* **Solution Summary:**
  * Ensure you're using the Phase 1 configuration to onboard a cluster in Phase 1/read-only mode (follow our repository and, depending on which cloud provider, use the read-only option of the example repos).
  * Follow the steps in the README to onboard your cluster in Phase 1.
  * Follow the official onboarding process to move the cluster to Phase 2 if needed.

***

## Node template creation error

**Error message:**

```json
Error: expected status code 200, received: status=404 body={"message":"node template not found", "fieldViolations":\[]}

with module.server_eks.module.atlas-eks-cluster.atlas_node_template.this["default_by_atlas"],  
on modules/server-eks/terraform-atlas-eks-cluster/main.tf line 41, in resource "atlas_node_template" "this":  
41: resource "atlas_node_template" "this" {
```

**Issue:** This error typically indicates that Terraform attempted to configure a node template that was expected to exist but was not found in its state management. This situation arises when a node template already exists outside of Terraform's knowledge and needs to be imported into Terraform's state for proper management.

**Solution:** To fix this, you must import the node template into Terraform's state if it already exists but is not recognized by Terraform. Before importing, remove any existing references from the state to avoid conflicts.

1. **Remove the Node Template from Terraform State:** To prevent Terraform from attempting to create a new resource that conflicts with the existing one, use the following command:

```toml Terraform
terraform state rm 'atlas_node_template.this'
```

2. **Import the Existing Node Template into Terraform State:** After removing the resource from the state, you can then import the existing node template so that Terraform can manage it:

```toml Terraform
terraform import 'atlas_node_template.this' \<NODE_TEMPLATE_ID>
```

Replace `\<NODE_TEMPLATE_ID>` with the actual identifier of the node template.

This process will synchronize Terraform's state with the actual infrastructure, allowing further resource management through Terraform.

***

## Importing resources not created by Terraform

* **Why it happens**: Sometimes, resources created outside Terraform need to be managed by Terraform without causing recreation.
* **Solution:** Use the below command to bring externally created resources under Terraform management.

```toml Terraform
terraform import [resource_name] [resource_id]
```

***

## Improper variable reference or undefined variables

* **Why it happens:** Errors in variable references or using undefined variables can lead to configuration failures.
* **Solution:** Define all variables in the `variables.tf` file and reference them correctly using `var.variable_name`. Ensure that all used variables are declared and have assigned values.

```
variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t2.micro"
}
```

***

## Provider configuration not present

* **Example:** "Error: Provider configuration not present."
* **Why it happens:** The required provider, such as Atlas, isn't defined or initialized in your Terraform setup.
* **Solution:** Define the Atlas provider in your configuration and run Terraform init.

```toml Terraform
provider "atlas" { 
	# Configuration attributes
}
```

***

## Error 401 (Unauthorized)

* **Why**: Incorrect or missing authentication credentials.
* **Solution:** Ensure your provider configuration includes valid credentials. For example:

```toml Terraform
provider "aws" {  
  access_key = "YOUR_ACCESS_KEY"  
  secret_key = "YOUR_SECRET_KEY"  
  region     = "us-west-2"  
}
```

***

## Authentication errors for Azure

* **Why:** Occurs when the Terraform provider cannot authenticate with the cloud service.
* **Solution:** Double-check your provider's authentication configuration. For example:

```toml Terraform
provider "azurerm" {  
  features {}  
}
```

Ensure environment variables `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_SUBSCRIPTION_ID`, and `ARM_TENANT_ID` are set correctly.

***

## Capturing Terraform debug logs

* **Example:** Error `context deadline exceeded` when applying Terraform configurations
* **Why it happens:** Terraform operations may time out or fail without providing enough information about the root cause
* **Solution:** Enable debug logging to get detailed information about the Terraform execution process, which will help to identify and troubleshoot the issue

```toml Terraform
TF_LOG=DEBUG TF_LOG_PATH=tf.log terraform plan
```

For more detailed instructions on using the Terraform debugging and logging features, refer to [Terraform debugging documentation](https://developer.hashicorp.com/terraform/internals/debugging).