---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Available savings

Use this report to find out how much exactly you can save thanks to optimizing your cluster.

The available savings report provides insights into potential cost savings achievable through Atlas's autoscaling features. This guide explains how to interpret the report and use its features effectively.

## Overview

The report content varies based on your cluster management status:

1. Read-only clusters
2. Atlas-managed clusters

Each type offers different levels of detail and actionable insights.

> 📘 Important requirement
>
> To view the available savings report, your cluster must have at least one supported node type. If your cluster consists only of nodes that are not supported for savings calculations, the report will not load.

## Read-only cluster report

### Key features

* Estimated total savings potential
* Current cluster costs
* Potential savings from recommendations

<Image alt="Overview of savings estimates" border={false} src="https://files.readme.io/81946da-availablesavings1.png" />

### Report cards

1. **Workload rightsizing**: Estimates savings from optimizing workloads.

2. **Spot Instances**: Calculates potential savings from Spot Instances automation.
   * Toggle between all workloads or spot-friendly workloads only

3. **ARM support**: Projects savings from using ARM instances.
   * Adjust the percentage of ARM CPUs to see different scenarios

<Image alt="Report cards for Workload Rightsizing, Spot Instances, and ARM Support" border={false} src="https://files.readme.io/89c1ef0-availablesavings2.png" />

4. **Configuration comparison**: Analyzes resource usage and potential savings over time.

#### Current vs. optimized configuration

<Image alt="Configuration comparison" border={false} src="https://files.readme.io/de1d77a-availablesavings3.png" />

This comparison shows:

* Current cluster configuration: details of existing instances, including quantity, name, CPU cost, hourly cost, and total monthly cost
* Optimized cluster configuration: a summary of the proposed optimal setup, with total cost savings
* Resource usage: CPU, memory, and GPU usage breakdown for current and optimized configurations

#### Comparison over time

<Image alt="Comparison over time" border={false} src="https://files.readme.io/675c29c-availablesavings4.png" />

This detailed chart provides:

* The average available savings percentage
* Current vs. optimal cost comparison
* Daily CPU cost graph over the selected time period
* Cluster resource analysis, including:
  * Average CPU waste percentage
  * Current vs. optimal CPU count
  * Total current vs. optimal CPU provisioned over time

You can adjust the time range (e.g., 24 hours, 7 days) and cost rate display (hourly, daily, monthly) to suit your analysis needs.

> 📘 Note
>
> Detailed configuration comparisons are only visible to Atlas managed service users.

## Atlas-managed cluster report

For clusters managed by Atlas, the report offers additional features:

1. **Progress indicator**: Shows how close your cluster is to the recommended setup.

<Image alt="Progress bar towards optimal configuration" border={false} src="https://files.readme.io/e1b3c35-availablesavings5.png" />

2. **Detailed configuration comparison**: Provides in-depth analysis of:
   * Optimized cost projections
   * Efficient configuration details

<Image alt="Detailed configuration comparison" border={false} src="https://files.readme.io/90541e4-availablesavings6.png" />

## How to use this report

1. Review the estimated total savings to understand the overall optimization potential.
2. Examine each report card to identify specific areas for improvement:
   * Workload rightsizing opportunities
   * Potential for Spot Instances usage
   * Benefits of incorporating ARM instances
3. Use the configuration comparison to track progress over time and identify persistent resource waste.
4. For Atlas-managed clusters, use the progress indicator to gauge optimization efforts and refer to the detailed comparison for actionable insights.

<Cards columns={2}>
  <Card title="Realized savings" href="https://docs.atlas.example/docs/savings-report" icon="fa-piggy-bank">
    Track realized savings once Atlas is managing your cluster.
  </Card>
  <Card title="Savings calculations" href="https://docs.atlas.example/docs/savings-baseline" icon="fa-scale-balanced">
    Baseline methodology and calculation formulas behind the report.
  </Card>
</Cards>