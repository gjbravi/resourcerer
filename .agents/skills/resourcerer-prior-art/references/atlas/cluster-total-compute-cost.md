---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Cluster cost report

Cluster cost report provides detailed data on your total compute spend costs for a given cluster. You can use it to analyze current expenses, get forecasts, and see exact daily costs per resource type. The report also lets you switch to a cumulative cost view.

To view costs across all clusters in your organization, see the [Organization cost report](organizational-cluster-cost-report.md).

> ⚠️ Important
>
> The cluster report presents data calculations based on public cloud inventory data. To reflect your organization's custom pricing models (CUDs, Savings Plans, EDP discounts, and private pricing), configure [Price adjustments](discount-engine-overview.md).

## What you can find in the cluster report

You can browse the report using different pre-set time ranges:

* This month
* Last month
* Last 7, 14, and 30 days

<Image align="center" width="400px" src="https://files.readme.io/ca7c93d-Screenshot_2023-05-31_at_16.23.00.png" />

The report will yield different results depending on the selected time range, and some metrics may change accordingly.

**The cluster report consists of three main parts:**

1. Cluster cost overview
2. Compute spend
3. Daily compute spend

## What’s inside each report section

### Cluster cost overview

The top report section summarizes your cluster expenses for the selected period. It includes a monthly forecast, average daily cost, and average daily cost per resource type (CPU and memory).

<Image align="center" src="https://files.readme.io/84e269f-Screenshot_2023-09-19_at_12.28.55.png" />

### Compute spend

This section presents your total compute costs. You can further break down expenses by on-demand, fallback, and spot instances. You can switch between a daily and cumulative view.

<Image align="center" src="https://files.readme.io/d445425-Screenshot_2023-09-19_at_12.29.18.png" />

This part of the report also presents the cost of provisioned CPU and memory, which you can further analyze by normalized, on-demand, fallback, and spot instances. You can switch between hourly, daily, and monthly cost rates.

<Image align="center" src="https://files.readme.io/8eb2b2f-Screenshot_2023-09-19_at_12.29.37.png" />

### Daily compute spend details

In this part, you can find the daily costs of provisioned resources broken down by CPU and memory and the total daily cost per resource.

<Image align="center" src="https://files.readme.io/61343f5-Screenshot_2023-09-19_at_12.30.01.png" />

By pressing the arrows next to the CPU and GiB, you can identify the day with the highest/lowest cost per resource or provisioned resource in this cluster. The arrows by the total cost column let you sort days by your cluster cost: from the lowest to the highest or the other way round.

## How to understand metrics in the cluster report

Some of the metrics you see differ depending on whether you select the current or past period. Here's how to understand what you see in your report.

### The current period

<Image align="center" src="https://files.readme.io/84e269f-Screenshot_2023-09-19_at_12.28.55.png" />

* **Current month spend** is the sum of costs generated on all previous days and the sum of projected today's spend.
* **Monthly forecast** considers all complete days and forecasts of spending until the end of the month using a 7-day moving average. That means that tomorrow's average spend stems from dividing the spend of today and the previous six days by 7. The day after tomorrow is calculated similarly, including tomorrow's projection.
* **Avg. daily cost** is the total cluster compute cost in the selected period divided by the number of days.

### The past period

<Image align="center" src="https://files.readme.io/9269e68-Screenshot_2023-09-19_at_12.30.48.png" />

* **Compute spend** is the actual cluster compute spend in the selected period.
* **Avg. monthly cost** is calculated as follows:

```text
Avg. monthly cost = daily average x 30 days
```

<Image align="center" src="https://files.readme.io/813f4ef-Screenshot_2023-09-19_at_12.31.11.png" />

* **Total compute spend** is the cluster cost in the selected period broken down by on-demand, fallback, and spot.

### Other metrics

<Image align="center" src="https://files.readme.io/ec00f3e-Screenshot_2023-09-19_at_12.32.44.png" />

* **Provisioned CPU** cost is the total compute spend divided by the average number of provisioned CPUs in the selected period.
* **Normalized cost** per CPU is calculated in the following way:

```text
Hourly  = average hourly CPU cost / average hourly provisioned CPUs
Daily   = average hourly CPU cost / average hourly provisioned CPUs x 24
Monthly = average hourly CPU cost / average hourly provisioned CPUs x 730
```

<Image align="center" src="https://files.readme.io/bfacde7-Screenshot_2023-09-19_at_12.33.11.png" />

* **Daily compute spend** details are calculated similarly:

```text
On-demand = total day's on-demand CPU cost / day's provisioned on-demand CPUs
Spot      = total day's spot CPU cost / day's provisioned spot CPUs
Fallback  = total day's fallback CPU cost / day's provisioned fallback CPUs
```

<Cards columns={2}>
  <Card title="Workload costs" href="https://docs.atlas.example/docs/workloads" icon="fa-cubes">
    Break down cluster costs by individual workload.
  </Card>
  <Card title="Namespace costs" href="https://docs.atlas.example/docs/namespaces" icon="fa-folder-tree">
    Break down cluster costs by namespace.
  </Card>
</Cards>