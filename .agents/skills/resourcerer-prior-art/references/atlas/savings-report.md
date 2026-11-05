---
updatedAt: 2026-10-07T11:40:06.000Z
---

# Realized savings

Understand how much money you have saved thanks to optimizing your clusters.

The Savings Report answers two questions: how much money has Atlas saved you, and how? For every cluster, and for your organization as a whole over time, it compares what you're actually paying today against a modeled estimate of what you'd be paying without Atlas.

## Key concepts

A few terms in this report have specific meanings. Understanding them makes the rest of the report much easier to read.

**Provisioned** is the amount of a resource (CPU cores or RAM) that's actually allocated on your nodes. This is the capacity you're paying for.

**Requested** is the amount of a resource your workloads ask for, via Kubernetes pod resource requests, after Atlas's workload autoscaler has rightsized them. **Original requested** is what workloads asked for *before* rightsizing. The difference between the two is the effect of rightsizing.

**Lifecycle** refers to how a node is purchased: on-demand, spot, or fallback (a substitute instance type used when a preferred one isn't available). The report combines all three into single totals for provisioned resources and cost. It doesn't break savings down by lifecycle.

**Realized savings** are savings realized primarily through the node autoscaler: the difference between projected and actual cost, driven by bin-packing workloads more efficiently, adopting spot instances, and choosing cheaper node types. When both the node autoscaler and the workload autoscaler are enabled, the impact of the workload autoscaler is included in the realized savings.

**Workload autoscaler savings** come from rightsizing: reducing workloads' CPU and RAM requests to match what they actually use, so fewer nodes are needed to serve the same workloads. These are modeled estimates. If the node autoscaler is enabled, they are already included in realized savings. Otherwise, they are potential savings not reflected in actual spend.

**Baseline** is the reference point Atlas uses to model what your cluster would have cost without its optimizations. It's covered in detail in [Savings reference](savings-baseline.md).

## The two sources of savings

Atlas saves you money in two structurally different ways, and the report attributes each separately so you can see which is doing the work.

### Realized savings (node-level)

The cluster autoscaler reduces the cost of the capacity you provision: it packs pods onto fewer nodes, adopts spot instances where appropriate, and picks cheaper node types for the same workloads. The report calculates this as:

```text
realized savings = projected cost − actual cost
```

where projected cost reconstructs what you'd be paying if you were still provisioning at your old overprovisioning ratio, at your old unit price adjusted to your current commitment tier (see [Savings reference](savings-baseline.md#commitment-tier-changes)).

### Workload autoscaler savings (rightsizing)

The workload autoscaler reduces demand itself: it lowers CPU and RAM requests that were set too high to begin with, so workloads ask for closer to what they actually use. Smaller requests mean fewer nodes are needed. The report calculates this as:

```text
workload autoscaler savings = (original requests − current requests) × unit cost
```

counted only on days when workload autoscaling (VPA) was actually active on the cluster.

> 📘 Note
>
> These two numbers are not redundant. Autoscaler savings come from satisfying demand more efficiently. Workload autoscaler savings come from reducing demand itself. A cluster can show strong savings from one and little from the other, depending on how it's configured.

## When savings start appearing

A cluster needs a baseline before it can show any savings at all, and savings only start accruing once the baseline period ends. If you're looking at a very new cluster, or a time range that falls entirely inside a cluster's baseline period, you may see projected cost equal actual cost and zero autoscaler savings. That's expected behavior. See [Savings reference](savings-baseline.md) for the exact conditions.

## Report views

The report is available at two levels, each reflecting the same underlying calculations grouped differently. Both views include the same set of summary cards and charts, scoped to either your entire organization or a single cluster.

### Shared elements

#### Overview cards

* **Baseline spend**: the modeled cost of running your cluster(s) without Atlas.
* **Realized savings**: savings through the node autoscaler, including workload autoscaler impact when both are enabled. Appears only when the node autoscaler manages at least 20% of nodes (on at least one cluster in the organization view, or on the specific cluster in the cluster view).
* **Workload autoscaler savings**: modeled savings from the workload autoscaler. Appears when the workload autoscaler manages at least 20% of workloads in VPA mode.
* **Actual spend**: real spending over the selected period.

> 📘 Note
>
> Realized savings and workload autoscaler savings appear independently. If neither autoscaler meets the adoption threshold, the corresponding card is hidden.

#### Savings over time

Shows realized savings and workload autoscaler savings trends across the selected period. Which lines appear depends on which autoscalers are adopted.

#### Workload autoscaler adoption

Shows the percentage of workloads managed by the workload autoscaler over time.

#### Workload utilization rate

Shows CPU and memory utilization (switchable between the two), helping you assess whether resources are being used efficiently.

### Organization view

The organization view aggregates data across all clusters in your Atlas organization. In addition to the shared elements above, it includes a per-cluster breakdown.

#### Breakdown by cluster

A table with one row per cluster, aggregated over the selected time range. For each cluster, you can see:

* Which optimizations are turned on (node autoscaler, workload autoscaler)
* Baseline source: how the baseline was determined for this cluster (see [Savings reference](savings-baseline.md))
* Baseline cost: the modeled cost without Atlas
* Actual cost: real spending for this cluster
* Savings: either realized or workload autoscaler savings, depending on which autoscalers are active

Click any cluster to drill into the cluster view.

### Cluster view

The cluster view shows how much Atlas has saved on a single cluster and where those savings come from. It adapts its sections based on which optimization components are active.

You can reach the cluster view by clicking a cluster in the organization-level breakdown table.

#### Node autoscaler impact

This section appears when the node autoscaler is active on the cluster, managing at least 20% of the nodes.

**Provisioned capacity**: shows baseline, actual, and reduction in provisioned resources, for both CPU and memory. Baseline is what the cluster would have provisioned without Atlas; actual is what's currently provisioned; reduction is the difference.

**Average node count**: compares the average number of nodes the cluster would have run without Atlas against the actual average node count, with the resulting savings.

**Average cost per node**: compares the average cost per node without Atlas against the actual average cost per node, with the resulting savings.

#### Workload autoscaler impact

This section appears when the workload autoscaler is active on the cluster, managing at least 20% of workloads in VPA mode.

**Workload resource requests**: shows average original requests, average current requests, and resources freed over time, for both CPU and memory. This visualizes how rightsizing has reduced resource demand.

**Workload optimization breakdown**: a table listing every workload in the cluster. For each workload, you can see:

* Namespace
* Which optimizations are turned on (VPA, HPA)
* Workload autoscaler savings
* Percentage share of total savings

<Cards columns={2}>
  <Card title="Available savings" href="https://docs.atlas.example/docs/available-savings" icon="fa-chart-line">
    See how much you could save by optimizing your cluster.
  </Card>
  <Card title="Savings calculations" href="https://docs.atlas.example/docs/savings-baseline" icon="fa-scale-balanced">
    Baseline methodology and calculation formulas behind the report.
  </Card>
</Cards>