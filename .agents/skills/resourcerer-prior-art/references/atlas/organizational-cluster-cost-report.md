---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Organization cost report

The organization-level cost report provides a centralized view of costs and resource utilization across all your Kubernetes clusters, giving you critical insights into your cloud infrastructure spending. This report combines data from all clusters connected to Atlas in your organization, helping you understand total expenses, track provisioned resources, and identify opportunities for optimization.

To drill into a single cluster's cost and efficiency data, see the [Cluster cost report](cluster-total-compute-cost.md).

> 📘 Price adjustments
>
> Cost data is based on public cloud list prices by default. To reflect your organization's custom pricing (CUDs, Savings Plans, EDP discounts, and private pricing), configure [Price adjustments](discount-engine-overview.md).

![](https://files.readme.io/394271a14940bcc773467b35d2d66c046407adcf796abbe47fb82c72a1185a2c-image.png)

## Finding your way around the report

You'll find the organization-level cost report in the Atlas console under **Cost monitoring**. The report interface offers two distinct analysis paths through dedicated tabs at the top of the screen:

* The **Compute cost** tab shows spending across clusters and resources, breaking down expenses by CPU, memory, storage, and instance types.
* The **Efficiency** tab reveals how effectively resources are being used, highlighting areas where you might be over-provisioning or could optimize further.

![](https://files.readme.io/9c7b72d9ea015c86a0da3356074e59730cc5ad2114d9f97e0b9d2553dd0356b7-image.png)

## Compute cost

### Understanding your organization's spending

The report's compute cost overview presents your total organizational spending with context that helps you assess your Kubernetes costs:

* Month-to-date spending compared to previous periods helps identify trends
* Cost distribution between resource types reveals your main cost drivers
* The top 5 most expensive clusters are highlighted for you to review, bringing focus on high-impact areas

![](https://files.readme.io/a179c8913e724f13655e9c4c2793c98e62f2f269dee9af60f306b800c439ffaf-image.png)

### Meaningful cost comparisons

When analyzing spending trends, focus on:

* **Sudden cost increases**: These often indicate scaling events or potential issues requiring investigation. Make sure to consider any configuration changes you have made across the organization or growth that may be the root cause of costs being up compared to previous periods.
* **Weekend vs. weekday patterns**: Different usage patterns can suggest opportunities for scheduled scaling. If your applications have drastic differences between weekend vs. weekday or night vs. day traffic, consider this an opportunity to leverage for sizeable savings, which can be delivered by Atlas automation offerings. See [Workload autoscaler](workload-autoscaling-overview.md) for more information on how Atlas can automatically detect and optimize in line with these patterns.
* **Month-over-month changes**: Gradual increases might reveal resource creep that needs addressing. Alternatively, the relationship between the organization's growth and cost trends should be evaluated. Costs going up may not necessarily indicate a problem but, in fact, reveal that the proportional increase in costs related to organizational growth can indicate successful optimization attempts, depending on the ratio. For example, a 10% growth of your organization's infrastructure with only a 5% uptick in costs is commendable.

### Instance lifecycle impact on costs

Understanding how different instance types affect your costs is crucial for optimization:

* **On-demand instances** provide baseline capacity but at premium prices.
* **Spot instances** can significantly reduce costs but require proper handling of potential interruptions.
* **Fallback instances** temporarily protect workloads during spot unavailability.

High fallback costs often indicate that your spot instance configuration needs adjustment, either in terms of instance types selected or availability zones used.

![](https://files.readme.io/af930e5f74a708fff3b563499ab2d32350cb68a62a97b10b0b12a68c21c188d3-image.png)

### Most expensive clusters

The compute cost report highlights up to five clusters with the highest costs in your organization. This section helps you:

* Identify your main cost drivers
* Track cost trends over time between clusters
* Focus optimization efforts where they'll have the most impact

![](https://files.readme.io/912a3da205298f1c0cfab9ce84bdab3ead7102195362106c5cc2fc6e53d03e2c-image.png)

### Detailed cluster list

This section includes a table with detailed information on each cluster. Each cluster entry shows:

* Total cost
* Cost distribution visualization
* Comparison with other high-cost clusters
* Direct link to detailed cluster analysis

![](https://files.readme.io/753f4c2aa13b713a870fff25ef76df2093efeabccc2177ab2107a66b62a7e9ac-image.png)

The cluster list accommodates organizations with a high cluster count for improved convenience. It provides ample filtering and sorting options to narrow your selection to show only what matters to *you*.

Use the search functionality to isolate a specific cluster you wish to analyze.

* If you do this often with the same parameters, use suggested search criteria or leverage one of the recent searches.

![](https://files.readme.io/0dda940c0f9f1030239e3cb0cf509a90d4c24714b8dad276eca141621ff26f6f-image.png)

For organizations that have their clusters spread out across multiple cloud service providers (CSPs), the report offers the ability to filter by CSP:

![](https://files.readme.io/0618afc9f7cbe07dc941317791265dd1e7743026de6af050a87a7e2a8cb22ada-image.png)

## Resource efficiency

Switch to the Efficiency tab to understand what you're spending and how effectively your resources are being used. The efficiency view provides insights into resource utilization patterns and highlights potential optimization opportunities across your organization.

### Current overprovisioned resources

At the top of the efficiency view, you'll find percentages showing your current overprovisioned resources:

* **CPU overprovisioning**: The percentage of CPU resources that are provisioned but not used.
* **Memory overprovisioning**: The percentage of memory that is allocated but not utilized.
* **Storage overprovisioning**: The percentage of storage space that is provisioned but remains unused (available for GKE clusters).

![](https://files.readme.io/f2e23df5d4c1312d6a9c18e8af68f71ab6ffe728c67c04a076978e217d7021b5-image.png)

These percentages help you quickly assess whether your current resource allocation aligns with actual needs. High percentages often indicate opportunities for significant cost savings.

### Resource-specific analysis

The efficiency view provides dedicated tabs for CPU, memory, and storage analysis. For each resource type, you can track:

* **Provisioned resources**: Total capacity available across your clusters
* **Requested resources**: Capacity requested by your workloads
* **Used resources**: Actual consumption by your applications
* **Cost implications**: Financial impact of your resource allocation decisions

![](https://files.readme.io/d0fdd05a6b1b15eeb39367f3458cf6f4f3a8a905bb600b8ce73b896a30d4aa04-image.png)

When analyzing these metrics, focus on:

* Gaps between provisioned and used resources that indicate overprovisioning
* Consistent underutilization patterns suggesting rightsizing opportunities
* Cyclical usage patterns that could benefit from dynamic scaling

The relationship between these metrics often reveals optimization opportunities. For instance, if your resources are consistently well below the requested amount, you might benefit from adjusting your workload resource requests using Atlas's [workload autoscaler](workload-autoscaling-overview.md).

> 📘 Note
>
> Storage efficiency metrics are currently available only for GKE clusters.

### Resource costs and lifecycle

The efficiency view also breaks down resource costs by instance lifecycle:

* **Normalized view**: Overall resource efficiency
* **On-demand resources**: Efficiency of regular instances
* **Spot resources**: Utilization of spot instances
* **Fallback resources**: Usage patterns during spot unavailability

![](https://files.readme.io/d621c74b15e9dc8f20c9cfe7bdaea24817dc54df88ddab3815dac8657b1ec7f6-image.png)

Understanding these patterns helps optimize your instance lifecycle strategy for better cost efficiency.

## Cost-actionable data

The organization-level cost report is not just about viewing numbers; it is about identifying opportunities and taking action to optimize your Kubernetes infrastructure.

### Monitor spending trends

Start by monitoring trends in your month-over-month data. Gradual cost increases often signal resource creep, where additional resources are provisioned but never decommissioned. Watch for sudden spikes, too - these usually indicate an unexpected scaling event or a potential issue that needs investigation. Pay attention to the differences between weekend and weekday patterns, as these can help you implement more efficient scheduling strategies.

### Focus on high-impact targets

When deciding where to focus your optimization efforts, prioritize high-cost and low-utilization clusters. These clusters often provide the quickest wins in terms of cost reduction. Pay attention to patterns of consistent overprovisioning and inefficient distribution of instance lifecycles, as addressing these issues can lead to significant savings without compromising performance.

<Cards columns={2}>
  <Card title="Cluster cost report" href="https://docs.atlas.example/docs/cluster-total-compute-cost" icon="fa-server">
    Detailed compute spend breakdown for a single cluster.
  </Card>
  <Card title="Cluster efficiency" href="https://docs.atlas.example/docs/cluster-efficiency-report" icon="fa-gauge-high">
    Measure and track cluster efficiency metrics.
  </Card>
</Cards>