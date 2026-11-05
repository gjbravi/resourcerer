---
updatedAt: 2026-07-08T10:44:06.000Z
---

# AWS capacity reservations

Learn how to use AWS On-Demand Capacity Reservations (ODCRs) and EC2 Capacity Blocks for ML with Atlas to guarantee instance availability for critical workloads.

AWS capacity reservations guarantee EC2 instance availability in a specific Availability Zone and instance type. Unlike [Reserved Instances and Savings Plans](commitments.md), which reduce costs through long-term pricing commitments, capacity reservations ensure that the instances you need are available when you need them, regardless of overall AWS capacity constraints.

Atlas treats capacity reservations as first-class commitments. Once imported, they appear on the **Commitments** page under the **Capacity reservations** tab, where you can assign them to clusters, track their utilization, and configure the autoscaler to provision nodes into reserved capacity.

Atlas supports two types of AWS capacity reservations:

**On-Demand Capacity Reservations (ODCRs)** reserve instance capacity in a specific AZ for as long as needed. ODCRs are a good fit for planned migrations, high-availability setups, and any scenario where you need guaranteed capacity. Atlas supports both non-interruptible and interruptible ODCRs, as well as open and targeted reservation types.

**EC2 Capacity Blocks for ML** reserve GPU instance capacity for a fixed time window in the future. Capacity Blocks are designed for ML training jobs and AI workloads that require scarce GPU instances (such as p5, p5en, or p4d families). You purchase a block in advance, specifying the instance type, quantity, start date, and duration. Atlas supports instance-level Capacity Blocks. UltraServer Capacity Blocks are not currently supported.

[AWS documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/capacity-reservation-overview.html).

> 📘 Key difference from other commitments
>
> Most capacity reservations require explicit targeting through [Node Templates](node-templates.md). The node autoscaler does not automatically consume these capacity reservations the way it does with Reserved Instances or Savings Plans. You must configure a Node Template to target specific reservations before the autoscaler will provision nodes based on them. An exception to this treatment is [Open ODCRs](#open-vs-targeted-odcrs), which are treated by Atlas in the same way as Reserved Instances in terms of consumption.

## How capacity reservations work in Atlas

The overall workflow is:

1. **Purchase capacity in AWS.** Create ODCRs or Capacity Blocks through AWS.
2. **Import into Atlas.** Run the commitments import script or wait for an active Cloud Connect integration to sync. Atlas discovers capacity reservations by calling the AWS `DescribeCapacityReservations` API.
3. **Assign to clusters.** On the **Commitments** page, assign the imported reservation to the clusters that should use it.
4. **Configure Node Templates.** For each assigned cluster, select which Node Templates should target the reservation. The autoscaler only provisions nodes into a capacity reservation when a Node Template explicitly targets it.
5. **Run workloads.** When pending pods match the constraints of a Node Template targeting a capacity reservation, the autoscaler provisions nodes into the reserved capacity first.

> 📘 Two ways to link reservations and Node Templates
>
> You can configure the relationship between capacity reservations and Node Templates from either direction: in the **commitment detail drawer**, assign one reservation to multiple Node Templates across clusters, or in the **Node Template editor**, target multiple reservations from a single template.

### Provisioning order

When a Node Template targets one or more capacity reservations, the autoscaler follows this priority order:

1. **Capacity reservations** — Provisions into the targeted reservation first, up to the reserved instance count. Atlas is also aware of and will take into account any external usage of a reservation by systems outside of Atlas-managed clusters.
2. **Other commitment-covered On-Demand** — If the reservation is fully utilized, it falls back to RI/SP-covered capacity if available.
3. **Regular On-Demand/Spot** — Standard On-Demand/Spot Instances if no commitment capacity remains.

If the Node Template lacks fallback options and the reservation is fully utilized, pods remain pending until capacity becomes available.

### Reservation statuses

Each capacity reservation has a lifecycle status visible on the **Commitments** page:

**Scheduled** — The reservation exists but has not yet reached its start time. This status applies primarily to Capacity Blocks, which are purchased in advance for a future time window. The autoscaler does not provision into scheduled reservations.

**Active** — The reservation is currently available for use. The autoscaler can provision nodes into active reservations when matching pending pods exist.

**Cancelled** — The reservation has been cancelled in AWS. Atlas reflects this status after the next Cloud Connect sync (up to 1 hour). The autoscaler ignores cancelled reservations.

**Expired** — The reservation's time window has ended. This status applies to Capacity Blocks that have reached their end time. The autoscaler ignores expired reservations.

## Prerequisites

Before importing capacity reservations, ensure the following:

**AWS permissions.** Your Cloud Connect role must include the `ec2:DescribeCapacityReservations` permission. See [Cloud permissions](cloud-permissions.md) for the full policy. The script for uploading commitments requires the same permission.

**Existing integration.** You need either Cloud Connect configured for your AWS or use the commitments import script. Capacity reservations are discovered through the same import mechanism used for Reserved Instances and Savings Plans.

**Cluster connected.** At least one AWS cluster must be connected to Atlas in the same region as the capacity reservation for them to be visible in the user interface after importing.

## Importing capacity reservations

Capacity reservations are imported using the same workflow as other AWS commitments.

If you already have a commitments import or Cloud Connect set up, newly created capacity reservations are discovered automatically during the next sync cycle.

To import manually:

1. Navigate to **Optimization** > **Commitments** (Classic Console) or **Settings** > **System** > **Commitments** (Enhanced Console) in the left sidebar.
2. Click **Upload commitments** in the top right corner.
3. Run the script. Atlas discovers all commitment types, including capacity reservations.

Once imported, capacity reservations appear under the **Capacity reservations** tab on the Commitments page:

<Image align="center" caption="The Capacity reservations tab on the Commitments page showing imported reservations" src="https://files.readme.io/e423c26b287655738c3ae683636c71c76199be42de261511d813f0db989c3194-image.png" />

## Assigning capacity reservations to clusters

<Columns layout="auto">
  <Column>
    After importing, assign the reservation to the clusters that should use it:

    1. Click on a reservation name to open the detail drawer.
    2. Under **Assign clusters to the commitment**, select the clusters that should have access to this reservation.
    3. For each assigned cluster, select at least one **Node Template** from the dropdown. This is a mandatory step. The autoscaler requires explicit Node Template targeting to use capacity reservations.
    4. Click **Save**.
  </Column>

  <Column>
    ![](https://files.readme.io/9c12a4da271a6aa53def3dd2e5bd869c5cb617e76e9e5f563577ad02d9a6750d-image.png)
  </Column>
</Columns>

The detail panel also allows you to configure:

**Use Commitment for cluster autoscaling** — When enabled, Atlas uses this reservation when scaling up assigned clusters. When disabled, the reservation is not used in autoscaling decisions, but its utilization is still tracked.

**Set maximum usage limit** — Optionally restrict Atlas to use only a percentage of the reservation's capacity, reserving the remainder for workloads outside Atlas-managed clusters.

> 📘 Node Template selection is mandatory
>
> Unlike RIs and Savings Plans, capacity reservations require you to select Node Templates for each assigned cluster. If no Node Template is selected, the autoscaler cannot provision into the reservation.

## Configuring Node Templates for Capacity Reservations

To target a capacity reservation from a Node Template, use the **Target capacity reservations** field in the Node Template configuration:

1. Navigate to your cluster's **Configuration** > **Node autoscaler** page, then choose the **Node templates** tab.
   <Accordion title="I'm using Classic Console" icon="fa-info-circle">
     Navigate to your cluster's **Autoscaler** > **Node templates** page.
   </Accordion>
2. Create a new Node Template or edit an existing one.
3. In the **Target capacity reservations** section, select one or more capacity reservations from the dropdown.
4. Save the template.

When you select a reservation, Atlas automatically populates and locks certain constraints inherited from the reservation:

**Instance type** — Locked to the reservation's instance type.

**Availability Zone** — Locked to the reservation's AZ.

These inherited constraints cannot be removed unless you remove the reservation targeting. You can still configure additional settings on the Node Template.

The Node Template also automatically applies a `reserved=only:NoSchedule` taint to nodes provisioned from a capacity reservation. Your workloads must include a matching toleration to be scheduled on these nodes.

<Image align="center" caption="Node Template configuration showing the field for targetting capacity reservations" src="https://files.readme.io/38ce3082c789f09ed42e8a17a6b293cdd28fb643a0b45630a931aefca40ffc83-image.png" />

For full Node Template documentation, see [Node Templates](node-templates.md).

## Capacity Blocks

Capacity Blocks reserve GPU instance capacity for a fixed time window, making them the primary mechanism for ensuring availability of scarce GPU instances for ML training jobs.

### Setting up Capacity Blocks

1. **Purchase a Capacity Block through AWS.**

2. **Import into Atlas.** Navigate to **Optimization** > **Commitments** and import commitments, or wait for your existing integration to sync. The Capacity Block appears under the **Capacity reservations** tab with a **Scheduled** status until its start time.

3. **Assign to your ML cluster.** Click the Capacity Block and assign it to the cluster running your ML workloads. Select the Node Template(s) that should target this block.

4. **Configure the Node Template.** Create or edit a Node Template and select the Capacity Block under **Target capacity reservations**. Atlas locks the instance type and AZ to match the block. Add taints and tolerations to ensure only your ML workloads are scheduled on these nodes.

5. **Submit ML workloads.** When the block's start time arrives, and you submit training jobs, the autoscaler detects pending GPU pods and provisions nodes into the Capacity Block. Atlas provisions nodes up to the reserved instance count.

6. **Monitor utilization.** Track utilization on the **Commitments** page. Since you pay for the entire block duration regardless of usage, maximizing utilization reduces wasted spend.

### Capacity Block expiration

When a Capacity Block's end time approaches, AWS terminates the instances running within it. Atlas automatically handles node drain before AWS begins termination, giving your workloads time to be evicted more gracefully. After expiration, Atlas marks the block as **Expired** and the autoscaler stops considering it for scheduling.

If your workload is still running when the block expires, the affected pods go pending. The autoscaler then attempts to schedule them on other available capacity, depending on your Node Template configuration.

## On-Demand Capacity Reservations (ODCRs)

On-Demand Capacity Reservations let you reserve instance capacity in a specific AZ without a fixed end date. They are ideal for migrations, recovery, and other scenarios requiring guaranteed instance availability.

### Open vs. targeted ODCRs

AWS supports two types of ODCRs, and Atlas handles each differently:

**Open ODCRs** are automatically applied to any node being provisioned matching the reservation's instance type, AZ, and platform. Note that other systems in your AWS account may also consume open ODCR capacity.

**Targeted ODCRs** are only applied when you explicitly provision nodes using a specific Node Template that is tied to the ODCR in the Commitments interface.

### Interruptible vs. non-interruptible ODCRs

Atlas supports both non-interruptible and interruptible ODCRs:

**Non-interruptible ODCRs** provide guaranteed, uninterrupted capacity for the duration of the reservation. Once an instance is running in a non-interruptible ODCR, AWS will not reclaim it.

**Interruptible ODCRs** provide capacity that AWS may reclaim under certain conditions. When setting up these types of interruptible commitments, users should be aware of the risk of the instance being reclaimed and the effects that this may have on their workloads.

### Setting up ODCRs

1. **Create ODCRs in AWS.** Create capacity reservations for the instance type and AZ you need.

2. **Import into Atlas.** Import commitments or wait for Cloud Connect to sync. ODCRs appear under the **Capacity reservations** tab with their current status.

3. **Assign to your cluster.** Select the cluster and Node Templates that should use this ODCR.

4. **Configure the Node Template.** Select the ODCR under **Target capacity reservations**.

5. **Deploy workloads.** The autoscaler provisions nodes into the ODCR when matching pods are pending, up to the reserved instance count. If the reservation is fully utilized, the autoscaler falls back to other options configured on the Node Template.

## Interaction with Rebalancer

The [Rebalancer](rebalancing.md) respects capacity reservations when evaluating node replacements. Nodes backed by active capacity reservations are not replaced with non-reserved alternatives during their active windows. For Capacity Blocks specifically, the Rebalancer can help re-pack workloads onto reserved nodes to maximize utilization during the block window.

## Troubleshooting

### Capacity reservation not being utilized

If the autoscaler is not provisioning into an available capacity reservation, verify:

**Node Template targeting.** The reservation must be selected in the **Target capacity reservations** field of at least one enabled Node Template assigned to the cluster.

**Cluster assignment.** The reservation must be assigned to the cluster, and Node Templates must be selected for that cluster.

**Reservation status.** The reservation must be **Active**. Scheduled reservations (Capacity Blocks that haven't reached their start time) are not used for provisioning. Expired reservations are ignored.

**Matching workloads.** Pending pods must match the Node Template's constraints (labels, tolerations, resource requests) and the reservation's inherited constraints (instance type, AZ).

**Autoscaling toggle.** The **Use Commitment for cluster autoscaling** toggle must be enabled on the reservation.

**Toleration.** Workloads must include a `reserved=only:NoSchedule` toleration to be scheduled on reservation-backed nodes.

## Known limitations

**AWS only.** Capacity reservations are supported only for AWS clusters.

**UltraServer Capacity Blocks.** Atlas supports instance-level Capacity Blocks. UltraServer Capacity Blocks are not currently supported.

**Shared reservation tracking.** If multiple clusters or Node Templates target the same reservation, Atlas tracks utilization per Reservation ID across all of them. AWS enforces capacity limits, so over-assignment may result in some workloads not receiving reserved capacity.

## See also

* [Commitments](commitments.md) — Overview of discount-based commitments (RIs, Savings Plans, CUDs)
* [Node Templates](node-templates.md) — Configure Node Templates including capacity reservation targeting
* [Cloud permissions](cloud-permissions.md) — Required AWS permissions for importing capacity reservations
* [AWS On-Demand Capacity Reservations](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-capacity-reservations.html) — AWS documentation
* [Amazon EC2 Capacity Blocks for ML](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-capacity-blocks.html) — AWS documentation