---
updatedAt: 2026-08-25T13:38:10.000Z
---

# Commitments

The commitments feature set is Atlas's approach to utilizing cloud provider commitments in autoscaling. Commitments cover two categories:

**Cost optimization commitments** — [Reserved Instances](https://aws.amazon.com/ec2/pricing/reserved-instances/) (AWS, Azure), [Savings Plans](https://aws.amazon.com/savingsplans/) (AWS, Azure), and [Committed Use Discounts](https://cloud.google.com/compute/docs/instances/signing-up-committed-use-discounts) (GCP) reduce costs through long-term pricing agreements.

**Capacity reservations** — [AWS On-Demand Capacity Reservations](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-capacity-reservations.html) (ODCRs) and [EC2 Capacity Blocks for ML](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-capacity-blocks.html) guarantee instance availability in a specific Availability Zone and instance type. [GCP Compute Engine on-demand reservations](https://cloud.google.com/compute/docs/instances/reservations-overview) guarantee VM availability in a specific zone and machine configuration. For full documentation, see [AWS capacity reservations](aws-capacity-reservations.md) and [GCP reservations](gcp-reservations.md).

Atlas imports both categories and surfaces them on the Commitments page, where you can assign them to clusters and track utilization.

AWS Savings Plans are a flexible pricing model that provides significant savings on compute usage. Unlike Reserved Instances that are tied to specific instance types and availability zones, Savings Plans offer greater flexibility by applying to any instance family, size, operating system, tenancy, or region within the commitment scope. Atlas supports both Compute Savings Plans (covering all AWS compute services) and EC2 Instance Savings Plans (for specific instance families in particular regions).

GCP Committed Use Discounts come in two types. Resource-based CUDs commit to a specific amount of vCPU and memory in a region. Compute flexible CUDs (Flex CUDs) are spend-based commitments that cover overall compute spend, measured as a single utilization rate instead of CPU and memory splits. Flex CUDs work the same way as AWS Savings Plans — they apply to any machine type in the committed region and provide flexibility across instance families. Atlas supports both types for utilization tracking and node autoscaling. On the Commitments page, each type appears under its own tab: **Resource-based** and **Flexible CUDs**.

## Supported commitment types

| Provider         | Commitment type      | Supported for node autoscaling | Tracking of commitment utilization outside of Atlas-onboarded clusters |
| ---------------- | -------------------- | ------------------------------ | ------------------------------------------------------------------------ |
| AWS              | Reserved instances   | +                              | +                                                                        |
|                  | Savings plan         | +                              | +                                                                        |
| GCP              | Resource-based CUD   | +                              | -                                                                        |
|                  | Compute flexible CUD | +                              | -                                                                        |
| Azure            | Reserved instances   | +                              | -                                                                        |
|                  | Savings plan         | +                              | -                                                                        |
| Atlas Anywhere |                      | -                              | -                                                                        |

> 📘 Note
>
> Capacity reservations are supported for AWS (ODCRs and Capacity Blocks) and GCP (Compute Engine on-demand reservations). For details, see [AWS capacity reservations](aws-capacity-reservations.md) and [GCP reservations](gcp-reservations.md).

# How it works

Commitments are imported at the organization level. For AWS and GCP, you can import commitments using either Cloud Connect or the commitments import script. For Azure, use the commitments import script. One CUD, Reserved Instance, or Savings Plan in the user's cloud account equals one commitment in the Atlas platform.

AWS capacity reservations follow the same import process but behave differently from cost optimization commitments. While RIs and Savings Plans are consumed automatically when the autoscaler provisions matching instance types, capacity reservations require explicit targeting through [Node Templates](node-templates.md). You must configure a Node Template to target a specific reservation before the autoscaler will provision nodes into it. For the full setup workflow, see [AWS capacity reservations](aws-capacity-reservations.md).

For AWS, Azure, and GCP, the import workflow includes an auto-enablement option (selected by default) that automatically enables newly imported commitments for autoscaler usage and assigns them to all clusters. This ensures immediate utilization without requiring manual activation. You can disable this option if you prefer to manually enable and assign commitments after import.

> 📘 Note
>
> For AWS and GCP, commitment data imported through Cloud Connect is synchronized every hour.

Once imported, commitments can be managed in the following way:

* **Enabled**—When a commitment is assigned to a cluster but is disabled, it will not be actively utilized in the autoscaling.
  * Atlas can be restricted to use only a specified percentage of an uploaded commitment by setting the allowed percentage of utilization in the commitment settings.
* **Assigned** to one or multiple clusters - if a commitment is uploaded to Atlas but is not assigned to a cluster and is not enabled, it will not be used in the autoscaling; however, its utilization is still tracked.
  * Atlas supports assigning multiple commitments to all existing and newly added clusters as long as they match by region.
* **Deleted**—The commitment can be deleted from Atlas inventory and will no longer be tracked or utilized. This action does not affect the commitment record in the cloud account.

When autoscaling and tracking commitment usage, Atlas supports flex sizing, meaning an instance of any size can be provisioned as long as it is within the same instance family covered by the commitment.

# Setting up commitments

## Accessing the commitments page

Navigate to the **Commitments** page through the left sidebar:

1. In the left navigation panel, expand the **Optimization** section
2. Select **Commitments** from the submenu

![](https://files.readme.io/493c109517dc27333ff7b045e117f9606bbe92fe092b03a0aeaa55720493ed07-image.png)

## Uploading commitments

To upload commitments to Atlas:

1. **Click Upload commitments:** On the Commitments page, click **Upload commitments** in the top right corner

2. **Select your cloud provider:** If you have connected clusters from more than one Cloud Service Provider (CSP) in your Atlas organization, choose AWS, Azure, or GCP from the available options in the upload flow. For single-CSP organizations, the CSP is detected automatically.

3. **Configure auto-enablement:** By default, the **Enable commitments for autoscaler usage and assign to all clusters** option is selected. When enabled, newly imported commitments are automatically enabled for autoscaler usage and assigned to all clusters, ensuring immediate utilization without requiring manual activation. You can disable this option if you prefer to manually enable and assign commitments after import.

4. **Run the integration script:** Copy and execute the provided script to establish the connection between Atlas and your cloud account.\
   Atlas automatically pulls commitment data, including Reserved Instances, Savings Plans (AWS, Azure), Azure Reservations, or GCP Committed Use Discounts (both resource-based and flexible).

<Image align="center" width="50% " src="https://files.readme.io/19adc172054fcadfc5ed3e126b348a59070318accc75106e8ddea6fee68c124d-image.png" />

Once the upload is complete, your commitments will appear in the table where you can manage their assignments and utilization settings.

> 📘 GCP Cloud Connect
>
> If you have Cloud Connect configured for GCP, Atlas automatically discovers and imports both resource-based CUDs and Flex CUDs during each sync cycle (every 60 minutes). The auto-enablement setting applies to Cloud Connect imports as well.

## Managing commitment assignments

After uploading commitments, you can control how they're utilized:

**Enable and disable commitments:** Toggle individual commitments on or off to control their use in autoscaling decisions.

**Assign to clusters:** Allocate commitments to specific clusters or enable auto-assignment for newly added compatible clusters.

**Set utilization scope:** Restrict Atlas to use only a specified percentage of a commitment to accommodate services outside of Kubernetes that also use the commitment.

![](https://files.readme.io/142f0160430c317bbe255a5964c2b8c3aa0d5f6b832bce07308e4c8327bf3862-image.png)

# Upscaling the cluster using commitments

When Atlas scales up your cluster, it follows specific rules to maximize your commitment utilization. Here's how the process works:

**Regional matching:** Commitments must be in the same region as your cluster to be used by the Autoscaler.

**Priority order (AWS):** When you have both Reserved Instances and Savings Plans, Atlas uses Reserved Instances first (they provide the highest savings), then moves to Savings Plans once RIs are fully utilized.

**Priority order (GCP):** When you have both resource-based CUDs and Flex CUDs, Atlas maximizes resource-based CUD utilization first, then moves to Flex CUDs, and finally falls back to the next cheapest Node Template-compatible option.

**Capacity reservation priority:** When a Node Template targets one or more capacity reservations ([AWS](aws-capacity-reservations.md) or [GCP](gcp-reservations.md)), the autoscaler provisions into reserved capacity first, before considering other commitment-covered On-Demand, regular On-Demand, or Spot options. For AWS ODCRs and GCP specifically targeted reservations, this priority applies only through Node Templates that explicitly target the reservation. GCP automatically consumed reservations are an exception — any matching VM in the project consumes them without explicit Node Template targeting.

**Commitment scope:** The Autoscaler uses up to 100% of each assigned commitment. When a commitment reaches full capacity, scaling continues with regular On-Demand or Spot Instances.

**Organization-wide tracking:** Commitment usage is tracked across your entire organization. If any cluster uses an instance type covered by a commitment, it counts toward that commitment's utilization - even if the cluster isn't directly assigned to that commitment.

**Multi-cluster assignments:** You can assign a single commitment to multiple clusters. During simultaneous scaling events, this may temporarily exceed the commitment capacity. Use scheduled rebalancing to optimize resource allocation during off-peak hours.

**Pricing behavior:** Atlas shows accurate discounted prices for commitment-covered instances in Cost Monitoring and Rebalancer. The Autoscaler prioritizes these instances when scheduling conditions are met, unless [instance family prioritization](node-templates.md#set-up-instance-family-prioritization) is configured on the node template (see [Using commitments with instance family prioritization](#using-commitments-with-instance-family-prioritization)).

**Commitment lifecycle:** The Autoscaler stops using commitments when they expire or reach their maximum CPU capacity.

## Using commitments with Spot Instances

When no [instance family prioritization](node-templates.md#set-up-instance-family-prioritization) is configured, Atlas prioritizes commitment-covered instances when scaling up your cluster. The Autoscaler treats commitment-covered capacity as On-Demand instances and provisions them first to maximize commitment utilization.

> 📘 **Maximizing commitment utilization with Spot Instance fallback**
>
> Atlas automatically prioritizes commitment-covered capacity when scaling your cluster. To set up workloads that maximize commitment usage and fall back to Spot Instances when commitments are exhausted:
>
> **Requirements:**
>
> * Your node template must include both On-Demand and Spot Instance offerings
> * Workloads must have Spot tolerations (applied directly in manifests or via [Pod Mutations](https://docs.atlas.example/docs/pod-mutations#spot-instance-configuration))
>
> **Scaling behavior with these settings:**
>
> 1. Atlas provisions commitment-covered instances first (treated as On-Demand capacity)
> 2. Once commitments reach full utilization, new workloads with Spot tolerations are scheduled on Spot Instances
> 3. If Spot capacity is unavailable, workloads fall back to regular On-Demand instances
>
> **Note:** If your node template is configured for Spot-only (without On-Demand offering enabled), the Autoscaler will exclusively use Spot Instances regardless of commitment availability.

### Using commitments with instance family prioritization

If your node template has [instance family prioritization](node-templates.md#set-up-instance-family-prioritization) configured, the priority order takes precedence over commitment-based instance selection. The Autoscaler selects instances from the highest available priority tier first, even if a different instance family is covered by a commitment.

For example, if you have a resource-based commitment for the N2 family but your node template sets C2 as the first priority tier, the Autoscaler provisions C2 instances first. Once all priority tiers are exhausted or unavailable, the Autoscaler falls back to its default behavior, which includes prioritizing commitment-covered instances.

> 📘 **Tip**
>
> To get the best of both features, place commitment-covered instance families in your highest priority tier. This way, the Autoscaler provisions commitment-covered instances first while still respecting your priority preferences for remaining capacity.

### Prioritized utilization

The `Prioritized Utilization` feature enables allocating commitment-covered resources among different clusters based on each cluster's business priority. This feature allows high-priority clusters, such as production environments, to have first access to capacity, ensuring they primarily run on these discounted resources during peak times for maximum stability and cost efficiency. Lower-priority clusters, like non-production environments, can be set up to utilize Spot Instances when commitment-covered resources are unavailable. During off-peak hours, when higher-priority clusters scale down, lower-priority clusters can access any freed-up capacity.

The feature allows users to set priority levels for each cluster assigned to the commitment. As clusters upscale or downscale, the system dynamically adjusts resource allocation based on these priorities, ensuring that the highest priority clusters maximize commitments. When upscaling, a top-priority cluster's Autoscaler ignores the commitment count assigned to lower-priority clusters and considers the commitment fully available (only counting its own utilization). During the upscale of the second priority cluster, the commitment calculation considers the utilization of the top priority cluster; however, it ignores the utilization of all lower priority clusters, and so on. This behavior could lead to over-utilization of the assigned commitments and should be addressed by running a Rebalancer on the lower-priority clusters.

## Reporting utilization of commitments

Currently, reporting only provides a snapshot of the present situation without the ability to review historical data.

# Troubleshooting

## Commitments not being utilized

If the Autoscaler is provisioning Spot or On-Demand instances when you have available commitment capacity, **verify**:

* Your node template includes the On-Demand offering (commitments are treated as On-Demand capacity)
* The commitment is enabled and assigned to your cluster
* The commitment matches your cluster's region
* The commitment has available capacity (not fully utilized by other clusters or workloads)
* Your node template does not have [instance family prioritization](node-templates.md#set-up-instance-family-prioritization) configured with a non-commitment-covered family in a higher priority tier. Instance family prioritization takes precedence over commitments. If a higher-priority family is available, the Autoscaler selects it even when commitment capacity exists for a different family.

# Key formulas

The following formulas are applied when it comes to commitment management:

* `Commitment utilization` = (`Provisioned CPU` ÷ `Total CPU in the commitment`) × 0.5 + (`Provisioned MEM`÷ `Total MEM in the commitment`) × 0.5

> 📘 Note
>
> For spend-based commitments (AWS Savings Plans and GCP Flex CUDs), utilization is measured as a single rate covering overall compute spend rather than separate CPU and memory splits.

# Known limitations

The current implementation of commitment management has the following limitations:

* Commitment utilization calculations are not updated frequently. They will be recalculated with each node addition or deletion and whenever clusters are onboarded or removed.
* If you upload commitments but haven't connected any clusters from that cloud provider to Atlas, the commitments won't appear in the user interface.
* Commitments are not supported for Atlas Anywhere deployments.
* GCP commitment utilization by cloud assets outside of Atlas-onboarded clusters is not currently tracked. This is on the roadmap.

# Enterprise commitments

If your organization uses Atlas's Enterprise account structure, you can import a commitment once at the Enterprise level and share it automatically across all child organizations. See [Enterprise commitments](enterprise-commitments.md).