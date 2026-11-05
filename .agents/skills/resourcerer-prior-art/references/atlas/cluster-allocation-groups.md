---
updatedAt: 2026-08-26T11:20:13.000Z
---

# Cluster allocation groups

The Allocation groups report is a custom-made report that lets you allocate the costs of your cluster per team, application, or other criteria by selecting relevant labels and namespaces. You can use the data to analyze compute costs per created allocation.

To view allocation groups across multiple clusters, see [Organizational allocation groups](organizational-allocation-groups.md).

## How do you create allocation groups?

<Image align="center" src="https://files.readme.io/f8aad5a-Screenshot_2023-09-19_at_12.41.50.png" />

Navigate to the Allocation groups report in the side panel menu. Create an allocation group by clicking on the **Create allocation group** button in the upper right corner of the screen.

Enter your allocation group name and select the relevant namespaces and labels from the list.

> 📘 Note
>
> Make sure to create a name that reflects your group by entering the application or team name, or anything else that would help understand the contents of the group in the future.

Once everything is selected, save your allocation group. You’ll see this group's data in the report.

<Image align="center" width="30%" src="https://files.readme.io/9df4486-Screenshot_2023-07-03_at_12.10.53.png" />

To manage your allocation groups (edit or delete them), select the allocation group you want to manage, click on the menu near it, and select the action.

![](https://files.readme.io/d137e63-Screenshot_2023-07-03_at_12.09.43.png)

## Allocation groups report

<Image align="center" src="https://files.readme.io/85433e5-Screenshot_2023-09-19_at_12.42.36.png" />

Once you create your allocation groups, this report lets you see your cluster costs broken down by allocation groups. You can use this data to analyze compute expenses per allocation group, identify inefficiencies, and discover potential savings opportunities.

At the moment, the Allocation groups report provides three views:

* Daily allocation groups cost chart
* List of allocation groups with cost details per selected period
* List of workloads in the allocation group (once you click on the workload link)

### Daily allocation group cost chart

<Image align="center" src="https://files.readme.io/845af71-Screenshot_2023-09-19_at_12.43.13.png" />

This view lets you see the top 5 most expensive allocation groups and their daily costs in one place. You can check allocation group cost details by hovering over a selected date.

### List of allocation groups with cost details per selected period

<Image align="center" src="https://files.readme.io/87bad4d-Screenshot_2023-09-19_at_12.43.35.png" />

In the detailed view, you can dive into the details of each allocation group. For each of the listed allocation groups, you can check:

* The number of workloads in the allocation group
* The average requested CPU and GiB per hour
* The total cost per resource (CPU and GiB)
* The total cost of the allocation group split by resource offering (On-demand, Fallback, Spot)

By analyzing the data in the report, you can find out which team or application generates the largest costs and plan changes accordingly.

## Alternative creation methods

### Terraform provider

Use the Atlas Terraform provider to define allocation groups as infrastructure-as-code. This approach is ideal for managing allocation groups consistently across multiple clusters or environments.

For complete documentation and examples, see the [Atlas Terraform provider documentation](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/resources/allocation_group.md).

> 📘 Note
>
> Full support for managing allocation groups is available in Atlas Terraform provider v7.60.1 and later.

## Include Idle Cost

By default, the Allocation groups report shows only what your workloads **request**. Any CPU or memory on a node that no workload requested is idle capacity and stays at the cluster level.

When you enable **Include Idle Cost**, Atlas splits that idle cost among the allocation groups on each node, based on how much each group's workloads requested.

For a detailed breakdown of the calculation, see [How idle cost is calculated](workloads.md#include-idle-cost).

<Cards columns={2}>
  <Card title="Organizational allocation groups" href="https://docs.atlas.example/docs/organizational-allocation-groups" icon="fa-building">
    Track costs across multiple clusters.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Detailed compute spend breakdown for your cluster.
  </Card>
</Cards>