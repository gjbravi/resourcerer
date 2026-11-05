---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Overview

Tracking and monitoring expenses is an essential part of cloud cost management.

Built with cloud-native teams in mind, Atlas’s monitoring presents all expenses in one place and enables breaking them down by K8s concepts like cluster, workload, and namespaces. The feature lets you analyze the efficiency and resources used by your cluster and workload and demonstrates cost fluctuations over time.

This guide describes the insights you will find in Atlas’s cost monitoring and how to use it to understand your cloud expenses better.

## How to use cost monitoring in Atlas

Cost monitoring becomes available in your product console right after you [connect the cluster](getting-started.md).

There is no need to install anything else if [the Atlas agent](about-the-read-only-agent.md) already runs in your cluster. For workload efficiency, you also need to have a [Kubernetes metrics server](https://github.com/kubernetes-sigs/metrics-server#installation).

### Available insights

* Your most expensive workload, its efficiency, and how many of your provisioned resources get wasted
* Your daily and monthly cluster spend and its efficiency
* Your compute spend per on-demand, spot, and fallback nodes
* Cost per CPU and memory
* Monthly spend forecasts and the prediction of your end-of-month bill

### Key reasons to use Atlas’s cost monitoring

<Table align={["left","left","left"]}>
  <thead>
    <tr>
      <th>

      </th>

      <th>

      </th>

      <th>

      </th>
    </tr>
  </thead>

  <tbody>
    <tr>
      <td>
        **Free for unlimited clusters**
        There are no limits on the number or size of your clusters.
      </td>

      <td>
        **Available immediately**\
        You get cost insights right after connecting your cluster.
      </td>

      <td>
        **Real-time**\
        Atlas refreshes the underlying data every 60 seconds.
      </td>
    </tr>

    <tr>
      <td>
        **Assess your K8s efficiency**\
        Clearly see the difference between the provisioned and requested resources.
      </td>

      <td>
        **Billing access not required**\
        Atlas uses public pricing, so you don’t have to share billing details.
      </td>

      <td>
        **Unlimited access to historical data**\
        You can analyze months of past data for free.
      </td>
    </tr>
  </tbody>
</Table>

## Main reports

Atlas’s cost monitoring includes five main sections providing different levels of granularity.

### Report details

* **Cluster** gets you an overview of cluster expenses: compute spend, cost per provisioned resources, average daily cost, and daily compute spend details, including cost per CPU and memory. You also get a forecast of your final monthly bill and the overall change compared to the previous month.
* **Workloads** report presents the compute cost for each workload, with additional information on their controller type and namespace and the total cost per CPU and memory. You can further filter your results by labels and namespaces. Additionally, Workload Efficiency highlights the difference between the requested and used resources for each workload, helping to put a number on wasted resources.
* **Namespaces** report provides data on the compute cost for each namespace, including average CPU and memory requirement per hour and the total cost per resource.
* **Allocation Groups** report provides insights into the allocation groups you add to your cluster. These custom workload groups allow you to allocate costs by grouping workloads by namespaces or labels.
* **Cost comparison** lets you compare the requested CPUs’ cost between different periods to understand the level of delivered savings.

## Cost monitoring concepts

This section outlines the key concepts you need to understand the cost report:

* **Cluster** is a set of nodes that run containerized applications.
* **Workload** refers to an application running on Kubernetes.
* **Namespace** provides a mechanism for isolating groups of resources within a cluster. Resource names must be unique within a given namespace but not across all namespaces.
* **Cluster compute cost** is the total monthly cost of compute resources provisioned in a cluster.
* **Node resource offering** refers to on-demand, spot, and spot fallback nodes, where the last one involves temporarily using on-demand nodes when spot instances become unavailable.
* **Normalized cost per CPU** is `total cluster compute cost / total number of CPUs provisioned`. Atlas also calculates subtotals of this value for spot, on-demand, and fallback instances.
* **Price per provisioned resource** indicates your average cost per resource unit (CPU, memory). This value comes from `total cost / number of resources in your cluster` and mostly depends on your VM type and resource offering.
* **Price per requested resource** shows the cost per resource depending on your workload needs. The value results from `total cost of resources / requested units`. You can use it to assess the efficiency of autoscaling, because when you overprovision, the cost per requested resource increases.

<Cards columns={2}>
  <Card title="Realized savings" href="https://docs.atlas.example/docs/savings-report" icon="fa-piggy-bank">
    Track how much Atlas has saved you, and how.
  </Card>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Detailed compute spend breakdown for a single cluster.
  </Card>
</Cards>