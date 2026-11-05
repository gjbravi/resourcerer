---
updatedAt: 2026-10-28T06:18:50.000Z
---

# Platform usage

Track billable usage, credits, and consumption across Atlas products in your organization.

Platform usage gives you visibility into how your organization consumes Atlas's billable products. It tracks usage metrics, credit allocations, and per-cluster consumption so you can understand what you're being charged for and how your usage trends over time.

> 🚧 Note
>
> The metrics, products, and credit terms shown on your Platform usage page reflect your organization's agreement with Atlas. For questions about your specific contract terms, contact your account representative.

<Tabs>
  <Tab title="Enhanced console">
    Navigate to **Settings → Platform usage** in the left sidebar.

    ![](https://files.readme.io/d85cbc2ab8ac591ea8108262cb9ca45368a4a693a00f53c0b885fa54ac08a377-image.png)
  </Tab>

  <Tab title="Classic console">
    Navigate to **Manage Organization → Platform usage** in the left sidebar.

    ![](https://files.readme.io/edc39563b10d5fea40f9e8988ae740955d8b108018d2dad983716f88a6dd0d25-image.png)
  </Tab>
</Tabs>

# What Platform usage shows

The landing page lists your active Marketplace subscriptions and shows a card for each billable product — Node Autoscaler, Workload Autoscaler, and/or Database Optimization. Each card displays the product's current month usage, credit consumption, and credit reset cadence at a glance. Select a product card to open its detailed view.
The subscription banner at the top shows which cloud providers have active Marketplace subscriptions and how many days remain on each.

## Product detail view

When you select a product card, the detail view opens with:

* **Current month usage** — total billable usage for the current month: billable CPUs across all clusters for Node Autoscaler and Workload Autoscaler, or billable database instances for Database Optimization.
* **Credits** — how much of your credit allocation has been consumed, shown as a percentage and as an absolute count (for example, 8,000 / 10,000 CPU used). Credits are pre-purchased usage allocations included in your contract.
* **Period** — the active credit period (for example, 03/31/2025 to 03/31/2026).
* **Credit reset** — whether your credits renew monthly (on the 1st of each month) or are fixed for the full period.

## Usage over time

Below the summary widgets, a set of category chips breaks down the current month's usage by cluster name or workload category, or — for Database Optimization — by database instance. These chips provide a quick read on where consumption is concentrated.

You can switch between a line chart and a bar chart using the toggle in the top-right corner of the chart area.

![](https://files.readme.io/a16285adbaa09602adb803b6db768603c37150b77f6a79e4982176d864522a84-image.png)

# Key concepts

## Billable CPU

Atlas measures usage in billable CPUs. One billable CPU equals one provisioned CPU running continuously for the full month. For example, a CPU running 15 days in a 30-day month counts as 0.5 billable CPU.

Usage data is collected at regular intervals and normalized against the total hours in the month to produce the billable figure shown in the console.

## Billable database instances

Database Optimization usage is measured in billable database instances. One billable database instance equals one database instance optimized by Atlas for a full calendar month. Usage is prorated for partial months: an instance optimized for 15 days in a 30-day month counts as 0.5 billable database instances.

Each optimized database instance is tracked individually, and the month's total is the sum across all instances. For example, two instances optimized for the entire month plus one instance optimized for half of the month count as 2.5 billable database instances.

Enabling or disabling optimization on an instance starts or stops billing immediately. If optimization on the same instance is stopped and re-enabled within the same month, its active time is combined.

## Credits

Credits are pre-purchased usage allocations defined in your Atlas contract. Each credit corresponds to a specific billable metric (e.g., CPUs) and a specific product (e.g., Workload Autoscaler).

Credits have a defined period and a reset cadence:

* **Monthly reset** — credit allocation renews on the 1st of each month. If you have 10,000 CPU credits with a monthly reset, you get 10,000 CPUs to use each month, regardless of what you consumed the previous month.
* **Custom reset** — credit allocation is fixed for the entire period. Once consumed, no additional credits are available until the next period begins.

The credits widget shows how much of your allocation has been used and how much remains.

## Billable products

Platform usage tracks the following Atlas products:

| Product               | Billable metric    |
| --------------------- | ------------------ |
| Node Autoscaler       | CPUs per cluster   |
| Workload Autoscaler   | CPUs per cluster   |
| Database Optimization | Database instances |

Additional billable metrics (for example, GPUs) may appear as your organization's product usage evolves.

# Enterprise organizations

If your organization is set up as an Enterprise with child organizations, Platform usage shows a consolidated view. The Enterprise-level page aggregates usage across all child organizations, allowing you to see total consumption.

Each child organization's usage is broken out individually, so you can identify which organizations are driving consumption.

# Related resources

<Cards>
  <Card title="Commitments" href="https://docs.atlas.example/docs/commitments" icon="fa-handshake">
    Manage cloud provider commitments (Reserved Instances, Savings Plans, CUDs) used by the Node Autoscaler.
  </Card>

  <Card title="Organization-level cost reports" href="https://docs.atlas.example/docs/organizational-level-reports" icon="fa-chart-bar">
    View compute cost and efficiency data across all clusters.
  </Card>

  <Card title="Price adjustments" href="https://docs.atlas.example/docs/discount-engine-overview" icon="fa-tags">
    Apply custom pricing to align Atlas's cost reporting with your actual cloud costs.
  </Card>
</Cards>