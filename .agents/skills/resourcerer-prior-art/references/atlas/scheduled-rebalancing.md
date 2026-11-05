---
updatedAt: 2026-07-20T07:37:57.000Z
---

# Scheduled rebalancing

# How it works

Scheduled rebalancing enables full or partial [rebalancing process](rebalancing.md) on a user-defined schedule, scope, and trigger to automate various use cases. Most commonly scheduled rebalancing is used in the following scenarios:

* Identify the most expensive Spot Instances and replace them with cheaper alternatives
* Perform full rebalancing of clusters on weekends
* Roll old nodes
* Periodically target and replace nodes with specific labels

A rebalancing schedule is an organization-wide object that can be assigned to one or multiple clusters. Once triggered, it creates a rebalancing plan scoped by parameters provided in the node selection preferences, while the [problematic workloads](preparation.md#problematic-workloads) are excluded automatically.

To set up scheduled rebalancing, navigate to the organizational-level Optimization menu or access the Rebalancer view within the cluster.

![](https://files.readme.io/db293fe-image.png)

# Setup

There are two main components in the scheduled rebalancing functionality

* A rebalancing job is a job that triggers a rebalancing schedule on an associated cluster.
* Rebalancing schedule - an organization-wide rebalancing schedule that can be used on multiple rebalancing jobs.

The Rebalancing schedule consists of four parts:

* **The schedule** describes when to run the scheduling (periodically, within the maintenance window, etc.)
* **Node selection preferences** contain rules for picking specific nodes, how many to target, and similar rules that mimic decision-making when manually rebalancing a cluster.
* **Trigger requirements** are decision rules for executing the rebalancing plan. They include rules like the “Savings threshold,” determining whether the generated plan should be executed.
* **Execution safeguard**: Atlas stops rebalancing before draining the original nodes if it finds that the user-specified minimum level of savings can't be achieved. This additional protection layer is required as planned nodes might temporarily be unavailable in the cloud provider's inventory during the rebalancing.

# Settings

The following settings can be adjusted when setting up the scheduled rebalancing:

<Table align={["left","left"]}>
  <thead>
    <tr>
      <th>
        Setting
      </th>

      <th>
        Description
      </th>
    </tr>
  </thead>

  <tbody>
    <tr>
      <td>
        Specify resource offering
      </td>

      <td>
        Spot, On-Demand, or Any
      </td>
    </tr>

    <tr>
      <td>
        Target using labels
      </td>

      <td>
        Key-value pairs are provided as `nodeSelector` terms. In the UI, when multiple terms are provided, the values are handled using `AND` logic, i.e., only nodes that satisfy all listed selector terms will be targeted.

        For `OR` logic between label values, use the Atlas API endpoint `POST /v1/rebalancing-schedules`. In the API request, use the `In` operator and specify multiple values for the same label key in the `matchExpressions` section. This API-only feature allows targeting nodes matching any specified label values. Note that the UI may not display all values correctly when using this API method.
      </td>
    </tr>

    <tr>
      <td>
        Minimum node age
      </td>

      <td>
        Amount of time since the node creation before the node can be considered for rebalancing. `0` - means a node of any age can be considered.
      </td>
    </tr>

    <tr>
      <td>
        Evict nodes gracefully
      </td>

      <td>
        Defines whether nodes that fail to drain within a [configurable timeout](rebalancing.md#pod-disruption-budget-pdb-handling) (default: 20 minutes, maximum: 180 minutes) will be kept with a `rebalancing.atlas.example/status=drain-failed` annotation instead of being forcefully drained.

        When **disabled** (default): Nodes are forcefully drained after the timeout, which may override Pod Disruption Budgets (PDBs) to complete the rebalancing operation.

        When **enabled**: Nodes are annotated, cordoned, and kept in the cluster. PDBs are strictly respected. By default, Atlas automatically uncordons these Nodes after 3 hours. To disable automatic uncordoning, set `drainFailureConfig.disableUncordon=true` via the API.
      </td>
    </tr>

    <tr>
      <td>
        Drain timeout
      </td>

      <td>
        Maximum time to wait for each node to drain during rebalancing. Default: 20 minutes. Maximum: 180 minutes (10800 seconds). Default: 20 minutes. Configure this in **Advanced settings** when setting up a rebalancing schedule.

        After this timeout expires, the node is either forcefully drained (default) or annotated and kept in the cluster (when **Evict nodes gracefully** is enabled).
      </td>
    </tr>

    <tr>
      <td>
        Maximum batch size
      </td>

      <td>
        Maximum number of nodes that will be selected for rebalancing. '0' indicates that all nodes in the cluster can be selected
      </td>
    </tr>

    <tr>
      <td>
        Sort selected nodes
      </td>

      <td>
        The algorithm used to sort selected nodes:

        - _Highest normalized CPU price_\* - sorts by the most expensive nodes based on the price of normalized CPU (node cost / CPU count).
        - _Highest requested resource price (CPU + RAM)_\* - sorts by the most expensive nodes based on the requested resource price, considering both CPU and RAM in a 1:7 ratio. This helps optimize costs for memory-focused workloads and node templates.
        - _Least utilized_\* - sorts by the least utilized nodes first, regardless of price.
      </td>
    </tr>

    <tr>
      <td>
        Aggressive mode
      </td>

      <td>
        Rebalance problematic pods, those without a controller, job pods, and pods with the `removal-disabled` annotation
      </td>
    </tr>

    <tr>
      <td>
        Savings threshold
      </td>

      <td>
        - _Target savings_\* (`savings_percentage`)<br />The minimum projected savings percentage required before executing a rebalancing plan. Plans that don't meet this threshold will be skipped.\\

        - _Guaranteed minimum savings_\* (`achieved_savings_percentage`)<br />An execution safeguard that prevents rebalancing when minimum savings cannot be achieved. When planned nodes become unavailable during execution, alternative nodes are created. If alternatives result in higher costs that don't meet the guaranteed threshold, new nodes are deleted, and rebalancing is aborted.
      </td>
    </tr>

    <tr>
      <td>
        Execution times
      </td>

      <td>
        Execution time can be adjusted by providing a timezone and a crontab expression
      </td>
    </tr>
  </tbody>
</Table>

## Node targeting

When configuring node targeting for scheduled rebalancing, you can choose between two methods: **Labels** or **Fields**. This selection determines how you'll identify nodes for the rebalancing operation.

### Using labels

To configure label-based targeting:

1. Select "Labels" under "Target using"
2. Configure your label-matching rules:
   * **Key**: Enter the Kubernetes label key (e.g., `scheduling.atlas.example`)
   * **Operator**: Choose from:
     * `Exists`: Match nodes where the label exists
     * `DoesNotExist`: Match nodes where the label is absent
     * `In`: Match nodes where the label value matches one in a list
     * `NotIn`: Match nodes where the label value doesn't match any in a list
   * **Value**: Enter the label value to match (required for `In` and `NotIn` operators)
3. Add additional label-matching expressions using the "+ Add" button if needed

<Image src="https://files.readme.io/d9f65376f7220ceadb918f46eb3b8719cef8c681573b33e8a360429b4a24af36-Section.jpg" align="center" />

**Multiple node label matching expressions are combined using the&#x20;**`AND`**&#x20;logic** – nodes must match all specified conditions to be selected for rebalancing.

For example, if you configure:

* `field1` `In` `spot_tmp`
* `field2` `In` `wrestle`

Only nodes with **both** `field1=spot_tmp` **AND** `field2=wrestle` will be selected.

<Image src="https://files.readme.io/a73a3bc75ec36701eedc7fe2008d94119fd8a794ed733f20e8361402721777c2-image.png" align="center" width="75% " />

#### Technical details

Under the hood, the Console UI creates a single `nodeSelectorTerm` with multiple `matchExpressions`. According to Kubernetes selector logic:

* All expressions within a `nodeSelectorTerm` are **ANDed** together
* Multiple `nodeSelectorTerms` are **ORed** together (API-only, see below)

### Using fields

Field-based targeting provides an alternative method when label-based selection isn't sufficient. When selecting **Fields**, you can target nodes based on their Kubernetes field values rather than labels.

#### Example

An example use case for field targeting is rebalancing a specific node by name. This is useful when you've identified a particular node that needs replacement. For example, a node running an outdated configuration or one flagged for maintenance.

1. Select **Fields** under **Target using**
2. Configure the field selector:
   * **Key**: `metadata.name`
   * **Operator**: `=`
   * **Value**: the name of the node to target (e.g., `ip-10-0-1-42.ec2.internal`)

![](https://files.readme.io/95dc0312150fe5a795bec8914ee9ed8d3d8f49499f3706fe97b22eb51b423fa8-image.png)

This configuration targets a single node by its Kubernetes node name, allowing you to schedule its replacement without affecting other nodes in the cluster.

For more information on using field selectors, refer to <Anchor target="_blank" href="https://kubernetes.io/docs/concepts/overview/working-with-objects/field-selectors/">Kubernetes documentation</Anchor>.

> ⚠️ **Known limitation: cordoned-node targeting may cause schedules to be skipped**
>
> Targeting cordoned Nodes with `spec.unschedulable=true` can cause the scheduled rebalancing operation to be skipped entirely rather than rebalancing the targeted Nodes. Drain-failed Nodes left cordoned after a graceful eviction timeout are automatically uncordoned after 3 hours by default, but will still be skipped by scheduled rebalancing while cordoned.
>
> For recovery options, see [Handling failed Node drains](rebalancing.md#handling-failed-node-drains).

### Using `OR` conditions for node labels in scheduled rebalancing

The console UI supports `AND` logic only. To use `OR` conditions between node label values, use the Atlas API:

1. Use the [Atlas API endpoint](https://docs.atlas.example/reference/scheduledrebalancingapi_createrebalancingschedule) for creating rebalancing schedules: `POST /v1/rebalancing-schedules`
2. Specify multiple `nodeSelectorTerms` in your request to create `OR` logic between terms
3. Within each term, use `matchExpressions` with the `In` operator to target multiple values for the same label key

#### Example: OR logic for multiple values of the same label

```json
"nodeSelectorTerms": [
  {
    "matchExpressions": [
      {
        "key": "nodetemplate",
        "operator": "In",
        "values": ["customerA", "customerB", "customerC"]
      }
    ]
  }
]
```

This configuration targets nodes with the `nodetemplate` label matching any of the specified values (`customerA` OR `customerB` OR `customerC`).

#### Example: OR logic between different label conditions

To create `OR` logic between multiple distinct label conditions, add multiple objects to the `nodeSelectorTerms` array:

```json
"nodeSelectorTerms": [
  {
    "matchExpressions": [
      {
        "key": "nodetemplate1",
        "operator": "In",
        "values": ["customerA"]
      }
    ]
  },
  {
    "matchExpressions": [
      {
        "key": "nodetemplate2",
        "operator": "In",
        "values": ["customberB"]
      }
    ]
  }
]
```

This selects nodes with **either** `nodetemplate=customerA` **OR** `nodetemplate2=customerB`.

> 🚧 Warning
>
> When using multiple `nodeSelectorTerms` via the API, the UI may not display all values correctly. The UI is limited to displaying a single term with AND logic only.

For complete details, refer to the <Anchor target="_blank" href="https://docs.atlas.example/reference/scheduledrebalancingapi_createrebalancingschedule">Atlas API reference documentation for creating rebalancing schedules</Anchor>.

## Sort selected nodes

When configuring scheduled rebalancing, the **Sort selected nodes** setting determines how nodes are prioritized for rebalancing. Each algorithm serves different optimization goals.

![](https://files.readme.io/cd559e907c7971fce69205f7846af89f5d1e9ffc2ac08f5f22812c79d5fa037a-image.png)

**Highest normalized CPU price**

* Sorts nodes by dividing the node's total cost by its CPU count, which prioritizes removing expensive nodes based on their provisioned (total available) CPUs.
* **Best for**: Spot Instance clusters when you want to remove the most expensive instances or a partial rebalancing where you just want to remove the most expensive nodes

> 📘 Note
>
> It doesn't consider actual resource utilization, only the raw cost per CPU.

**Highest requested resource price (CPU + RAM)**

* Sorts nodes by calculating the cost of actually requested (used) resources, considering both CPU and memory in a 1:7 ratio. For example, a $5/hour node using 1 CPU and 7GB RAM has a resource price of $2.50/unit, while the same node using 4 CPU and 28GB RAM has a resource price of $0.63/unit.
* **Best for**: Most scenarios, especially mixed workloads with varying memory and CPU needs, as it provides the best overall cost optimization by targeting nodes with the highest wasted resource cost.

**Least utilized**

* Sorts nodes purely based on CPU utilization percentage, regardless of node size or cost. For example, a 2-core node with 1 core used (50% utilized) would be prioritized over a 64-core node with 20 cores used (31% utilized), even though the larger node has more wasted resources
* **Best for**: Specific use cases where you want to consolidate many small, underutilized nodes, regardless of cost impact

> 📘 Note
>
> It may miss opportunities to remove larger, more expensive nodes that have higher percentage utilization but more wasted resources as a sum.

For most cost optimization scenarios, the **Highest requested resource price (CPU + RAM)** sorting algorithm provides the best balance for removing expensive and underutilized nodes.

## Aggressive mode configuration

When using aggressive mode in scheduled rebalancing, you can configure additional options to control how problematic workloads are handled.

### Local Persistent Volumes configuration

By default, even in aggressive mode, nodes with Local Persistent Volumes (LPVs) are marked as `"Not ready"` for rebalancing. This is because LPVs are tied to specific nodes through node affinity rules (typically using the `kubernetes.io/hostname` topology label) and cannot be migrated between nodes.

You can configure aggressive mode to ignore these local volume constraints.

#### Configuration via API

When creating or updating a rebalancing schedule through the API, include the `aggressiveModeConfig` in your request:

```json
{
  "launchConfiguration": {
    "rebalancingOptions": {
      "aggressiveMode": true,
      "aggressiveModeConfig": {
        "ignoreLocalPersistentVolumes": true
      }
    }
  }
}
```

This configuration:

* Requires `aggressiveMode` to be set to `true`
* Allows rebalancing to include nodes with Local Persistent Volumes

> 🚧 Warning
>
> Setting `ignoreLocalPersistentVolumes` to `true` will allow rebalancing to evict pods with local storage. This **will result in data loss** for applications using Local Persistent Volumes. Only use this option when data persistence is not critical.

For complete API documentation and the latest schema, refer to our [Scheduled Rebalancing API Reference](https://docs.atlas.example/reference/scheduledrebalancingapi_createrebalancingschedule).

#### Configuration via Terraform

In Terraform, use the `aggressive_mode_config` block to configure Local Persistent Volumes handling:

```toml Terraform
resource "atlas_rebalancing_schedule" "example" {
  # Required configuration elements...
  
  launch_configuration {
    # Enable aggressive mode
    aggressive_mode = true
    
    # Configure aggressive mode options
    aggressive_mode_config {
      ignore_local_persistent_volumes = true
    }
    
    # Other launch configuration settings...
  }
}
```

> 🚧 Warning
>
> Setting `ignore_local_persistent_volumes` to `true` will allow rebalancing to evict pods with local storage. This **will result in data loss** for applications using Local Persistent Volumes. Only use this option when data persistence is not critical.

For more information about Terraform configuration options, refer to the [Atlas Terraform Provider documentation](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/resources/rebalancing_schedule.md).

# Configuration examples

## Using the API

Create a rebalancing schedule using the [API](https://docs.atlas.example/reference/scheduledrebalancingapi_createrebalancingschedule):

```json
{
  "schedule": {
    "cron": "*/30 * * * *"
  },
  "launchConfiguration": {
    "selector": {
      "nodeSelectorTerms": [
        {
          "matchExpressions": [
            {
              "key": "scheduling.atlas.example/spot",
              "operator": "Exists"
            }
          ]
        }
      ]
    },
    "rebalancingOptions": {
      "executionConditions": {
        "enabled": true,
        "achievedSavingsPercentage": 10
      },
      "minNodes": 2,
      "keepDrainTimeoutNodes": false
    },
    "targetNodeSelectionAlgorithm": "TargetNodeSelectionAlgorithmNormalizedPrice",
    "numTargetedNodes": 3,
    "nodeTtlSeconds": 300
  },
  "triggerConditions": {
    "savingsPercentage": 20
  },
  "name": "rebalance spots at every 30th minute"
}
```

## Using Terraform

Define a rebalancing schedule using the [Terraform provider](https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/rebalancing_schedule):

```json Terraform
resource "atlas_rebalancing_schedule" "spots" {
  name = "rebalance spots at every 30th minute"
  schedule {
    cron = "*/30 * * * *"
  }
  trigger_conditions {
    savings_percentage = 20
  }
  launch_configuration {
    # only consider instances older than 5 minutes
    node_ttl_seconds         = 300
    num_targeted_nodes       = 3
    rebalancing_min_nodes    = 2
    keep_drain_timeout_nodes = false
    selector = jsonencode({
      nodeSelectorTerms = [{
        matchExpressions = [
          {
            key      = "scheduling.atlas.example/spot"
            operator = "Exists"
          }
        ]
      }]
    })
    execution_conditions {
      enabled                     = true
      achieved_savings_percentage = 10
    }
  }
}
```

<br />