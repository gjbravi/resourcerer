---
updatedAt: 2026-08-26T11:20:13.000Z
---

# Organizational allocation groups

The organizational level report lets you view allocation groups from multiple clusters, find common namespaces and workloads, and see their costs. As a result, you can better understand your Kubernetes costs across organization-wide groups and initiatives.

To view allocation groups for a single cluster, see [Cluster allocation groups](cluster-allocation-groups.md).

<Image align="center" src="https://files.readme.io/e61fb2e-alloc-1.png" />

You can find it in the main Atlas menu under the cost monitoring tab, where you can also create new organizational allocation groups.

## What’s inside the organizational allocation group report

The organizational allocation group presents the selected clusters, workloads, and namespaces with summaries of their requested resources, costs per resource, and total costs.

When you click on the workload count in the summary, you can see details of all workloads in the group with information on the clusters they belong to.

The choice of workloads and clusters depends entirely on your needs. When creating an organizational allocation group, you select the clusters to include and filters like namespaces and labels.

> ⚠️ Important
>
> Due to current limitations, the report needs to stay within the limit of 20,000 workloads. Exceeding this number can result in performance slowdowns.

## How to create an organizational allocation group

Setting up an organizational allocation group is similar to what it looks like at the cluster level.

1. Navigate to **Allocation Group** in the main menu and click **Create allocation group**.

<Image align="center" src="https://files.readme.io/21b5f6c-alloc-2.png" />

2. Select the clusters you wish to include in your allocation group.

<Image align="center" width="400px" src="https://files.readme.io/c9f41ef-alloc-3.png" />

3. Choose at least one filter: namespaces or workload labels.\
   Once done, Atlas creates an allocation group that includes all the workloads from the clusters that match your filter.
4. When selecting the workload labels, choose their key and specify the values you want to include in the report.

<Image align="center" src="https://files.readme.io/3aed998-alloc-4.png" />

<Image align="center" src="https://files.readme.io/9e9175c-alloc-5.png" />

5. To add a new key and value pair, click **Add another label** and repeat the action.

<Image align="center" src="https://files.readme.io/bd91f07-alloc-6.png" />

## Organizational and cluster-level allocation groups

The organizational allocation group also appears in the cluster-level report, showing relevant workloads with summaries.

Such multi-cluster allocation groups appear in cluster reports as any other allocation groups and show data relevant to the cluster you are viewing at the moment.

Cluster-level allocation groups are also visible in the organizational view but only include one cluster. You can edit your allocation groups by adding more clusters at any time.

## Include Idle Cost

By default, the organizational allocation groups report shows only what your workloads **request**. Any CPU or memory on a node that no workload requested is idle capacity and stays at the cluster level.

When you enable **Include Idle Cost**, Atlas splits that idle cost among the allocation groups on each node, based on how much each group's workloads requested.

For a detailed breakdown of the calculation, see [How idle cost is calculated](workloads.md#include-idle-cost).

<Cards columns={2}>
  <Card title="Cluster allocation groups" href="https://docs.atlas.example/docs/cluster-allocation-groups" icon="fa-tags">
    Attribute costs within a single cluster.
  </Card>
  <Card title="Namespace costs" href="https://docs.atlas.example/docs/namespaces" icon="fa-folder-tree">
    Break down cluster costs by namespace.
  </Card>
</Cards>