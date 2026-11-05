---
updatedAt: 2026-08-26T11:20:13.000Z
---

# Namespace costs

<Image align="center" src="https://files.readme.io/d5d0ae0-Screenshot_2023-09-19_at_12.38.57.png" />

This report allows you to see your cluster costs broken down by namespaces. You can use this data to analyze compute expenses per namespace or a group of selected namespaces, identify inefficiencies, and discover potential savings opportunities.

At the moment, the Namespaces report provides two views:

* Daily namespace cost chart
* List of namespaces with cost details per selected period

## Daily namespace cost chart

<Image align="center" src="https://files.readme.io/3015acf-Screenshot_2023-09-19_at_12.39.48.png" />

This view lets you see the top five namespaces and their daily costs in one place. You can check all the namespace cost details by hovering over a selected date.

## List of namespaces with cost details per selected period

<Image align="center" src="https://files.readme.io/533d323-Screenshot_2023-09-19_at_12.40.14.png" />

This view allows you to dive into the details of each namespace. For each of the listed namespaces, you can check:

* The number of workloads in the namespace
* The average requested CPU and GiB per hour
* Total cost per resource (CPU and GiB)
* The total cost of the namespace

You can select namespaces by ticking the box on the left-hand side. If you do that, you will see a summary of the cost details at the bottom of the screen:

<Image align="center" src="https://files.readme.io/350ab80-Screenshot_2023-09-19_at_12.40.46.png" />

## Include Idle Cost

By default, the Namespaces report shows only what your workloads **request**. Any CPU or memory on a node that no workload requested is idle capacity and stays at the cluster level.

When you enable **Include Idle Cost**, Atlas splits that idle cost among the namespaces on each node, based on how much each namespace's workloads requested.

For a detailed breakdown of the calculation, see [How idle cost is calculated](workloads.md#include-idle-cost).

<Cards columns={2}>
  <Card title="Workload costs" href="https://docs.atlas.example/docs/workloads" icon="fa-cubes">
    Break down cluster costs by individual workload.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Overview of total compute spend for your cluster.
  </Card>
</Cards>