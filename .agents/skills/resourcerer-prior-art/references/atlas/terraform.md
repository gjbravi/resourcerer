---
updatedAt: 2026-05-12T13:39:45.000Z
---

# Terraform

### Can one import the nodeTemplate state to the current Terraform state?

This question may concern you if you made some updates via the Atlas console UI, and when you tried to redeploy, you see them get changed or destroyed.

You need to change the Terraform script and add those fields. The Terraform state import just connects the Terraform resource to the node template—by importing, it just knows what object to change when applying changes (not how to apply them).

Our recommendation is to add the changes to a given module, run the Terraform plan, and repeat until there are no changes in the plan.

Alternatively, you can remove the module from the script and from the Terraform state. This will cause Terraform to stop tracking the node template, and it can be managed only from the UI.

<br />

***

### What is the limitation on the API token used for Terraform onboarding? In my cluster managed via Terraform, I've used both a full-access API key and the agent's API key for the cluster controller and didn't encounter any errors.

We created the token in the UI (onboarding script) for a specific cluster, bound to a specific onboarding; this token is tied to that cluster. The API key (created via UI) is not bound to a cluster. We don’t recommend sharing it between clusters, but it can be done.

If you reuse to token from the UI, you will face issues. The token from the UI expires after 4 hours and is also deleted when you delete a cluster.

For long-term use of clusters managed via Terraform, you need to use an API Token with Full Access created in the Atlas UI.

<Image align="center" src="https://files.readme.io/850275b-image-20230821-202228.png" />

<br />

***

### How do we disconnect a cluster from Atlas if we onboarded it using Terraform?

You can destroy Atlas resources using Terraform, as this is how the cluster was initially deployed. Also please ensure `delete_nodes_on_disconnect-variable` is set to true if you want to delete Atlas-created resources on deletion.

<br />

***

### What is your recommendation for moving from version 3 to 5? I would like to know the impact each component will have.

We recommend that you follow the outlined steps for migrating from version 3.x.x to 5.x.x:

1. Follow [these steps](https://github.com/atlas/terraform-provider-atlas#migrating-from-3xx-to-4xx) to migrate from version 3.x.x to 4.x.x.
2. Once you've completed the above step, proceed to follow the migration steps from version 4.x.x to 5.x.x.- [here's a detailed instruction](https://github.com/atlas/terraform-provider-atlas#migrating-from-4xx-to-5xx).
3. Lastly, make sure to add the parameters `spot_interruption_predictions_enabled` and `spot_interruption_predictions_type to the atlas_node_template`to enable Spot interruption predictions.

The following are moved from `autoscaler_policies_json` property with source and targets:

```
nodeConstraints.minCpuCores -> atlas_node_template (min_cpu)
nodeConstraints.maxCpuCores -> atlas_node_template (max_cpu)
nodeConstraints.minRamMib  -> atlas_node_template (min_memory)
nodeConstraints.maxRamMib -> atlas_node_template (max_memory)
customInstancesEnabled -> atlas_node_template (custom_instances_enabled)
spotInstances.Enabled -> atlas_node_template (spot)
spotInstances.spotDiversityEnabled -> atlas_node_template (enable_spot_diversity) 
spotInstances.spotDiversityPriceIncreaseLimitPercent -> atlas_node_template (spot_diversity_price_increase_limit_percent)
spotInstances.spotInterruptionPredictions -> Moved to atlas_node_template spotInstances.spot_interruption_predictions_enabled)
```

<br />

***

### Why do I get this error when running Terraform for a new cluster with the same scheduled rebalancing and job?

Error:

```
│ Error: expected status code 200, received: status=404 body={"message":"Job not found","fieldViolations":\[]}  
│  
│   with module.atlas-terraform.atlas_rebalancing_job.spots,  
│   on .terraform/modules/atlas-terraform/atlas.tf line 117, in resource "atlas_rebalancing_job" "spots":  
│  117: resource "atlas_rebalancing_job" "spots" {  
│  
╵Error: Error returned by 'terraform plan': Failed to run terraform: exit status 1
```

Currently, a cluster must be created with its own rebalance schedule and rebalance job ID. Atlas doesn't support this via Terraform at the moment.

<Image align="center" src="https://files.readme.io/d611d0b-Screenshot_2024-02-15_at_4.19.02_PM.png" />

<br />

***

### What steps should I follow to sequentially migrate my ATLAS installation to Terraform (IaC)?

1. Disable autoscaling in Atlas.
2. Activate non-Atlas node pools.
3. Provision nodes from the cloud provider for these non-Atlas node pools.
4. Sequentially delete all nodes from each node template, handling one node template at a time.
5. Verify that all Atlas nodes are deleted, ensuring only the cloud provider nodes remain operational.
6. Finally, disconnect Atlas from the cluster.