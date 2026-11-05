---
updatedAt: 2026-10-07T11:40:06.000Z
---

# Savings calculations

Baseline methodology, calculation formulas, and pricing logic behind every number in the Savings Report.

This page is the technical reference for the [Realized savings](savings-report.md) report. It covers the baseline methodology (how Atlas establishes a pre-optimization reference point) and the exact formulas that turn that baseline into projected cost and savings figures.

## Baseline overview

To say "Atlas saved you money," Atlas needs a reference point: what would this cluster have cost, and how would it have looked, without Atlas? That reference point is the **baseline**, and it's the foundation every number in the report is measured against.

The baseline isn't a single dollar figure. It's a small model of your cluster's behavior before Atlas's optimizations took effect, made up of four numbers plus a time window:

1. **CPU overprovisioning factor**: how much more CPU was provisioned than workloads actually requested.
2. **RAM overprovisioning factor**: the same, for memory.
3. **CPU unit cost**: the effective price you paid per CPU core, per hour.
4. **RAM unit cost**: the effective price you paid per GiB of RAM, per hour.

These four numbers are measured over a **baseline period**, a historical window representing your cluster before Atlas optimization. Once captured, the model stays fixed and is reapplied going forward to reconstruct what your projected (no-Atlas-AI) cost would be on any given day.

> 📘 Note
>
> Why a model, and not just an old bill? Clusters grow and shrink over time, so comparing today's bill to a fixed dollar amount from months ago wouldn't be fair. Instead, the baseline captures *ratios and rates* (how much you tended to overprovision, and at what price) and reapplies those ratios to today's actual workload demand. That keeps the comparison fair as your cluster scales up or down.

### What the numbers mean in practice

| Parameter                   | What it captures                                                             | Example                                                                     |
| --------------------------- | ---------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| CPU overprovisioning factor | Average ratio of provisioned CPU to requested CPU during the baseline period | `1.8` means 80% more CPU was sitting on nodes than workloads ever asked for |
| RAM overprovisioning factor | Average ratio of provisioned RAM to requested RAM                            | `1.4` means 40% more memory was provisioned than requested                  |
| CPU unit cost               | Average cost per CPU core, per hour, during the baseline period              | `$0.031` per core-hour                                                      |
| RAM unit cost               | Average cost per GiB of RAM, per hour, during the baseline period            | `$0.004` per GiB-hour                                                       |

These four values are averaged across your entire baseline period and applied uniformly. They don't currently distinguish between on-demand, spot, and fallback capacity.

## How your baseline is determined

Atlas computes baselines automatically, once per day, for every eligible cluster. It tries three methods in order, from most to least specific to your cluster, and uses the first one that has enough data:

```text
1. Your cluster's own history      →  most reliable
2. Other clusters in your org      →  fallback
3. Atlas's broader fleet average →  last resort
```

A cluster only moves to the next method if the previous one didn't have enough data to compute a baseline. Once a cluster gets a baseline from an earlier method, it's never re-evaluated against a later one.

### Your cluster's own history

This is the highest-quality baseline, because it's built entirely from your cluster's own pre-optimization behavior. Atlas looks at the period between when your cluster first connected and when Atlas took over active node management (specifically, the point at which more than 20% of your nodes became Atlas-managed), and calculates the four parameters from your cluster's actual metrics during that window.

Your cluster qualifies for this method only if there's **at least 7 days** of history between cluster creation and the point Atlas took over. Without a week of pre-optimization data, the cluster falls through to method 2.

In the report and API, this baseline type is labeled **Cluster history**.

### Other clusters in your organization

If your cluster doesn't have enough of its own history, Atlas looks for other clusters **in your organization** that already have a Cluster history baseline, and averages their overprovisioning factors and unit costs. The baseline window becomes the combined span of those peer clusters' own baseline windows.

The assumption here is that clusters within the same organization tend to share infrastructure patterns and pricing, making a sibling cluster's measured behavior a reasonable stand-in for yours.

In the report and API, this baseline type is labeled **Peer clusters**.

### Industry average

If your cluster has no usable history of its own, and no eligible peers in your organization, Atlas falls back to averaging overprovisioning and cost across **all clusters observed on the fleet**, on the single day your cluster connected to Atlas's active management. This is a genuine cross-fleet average, not limited to your organization.

In the report and API, this baseline type is labeled **Industry average**. Think of it as "no data more specific to this cluster was available for comparison."

> 📘 Note
>
> An industry-average baseline is the least specific of the three, since it isn't based on your own workloads or infrastructure. If your Savings Report is using an industry-average baseline and the numbers look off, contact your Atlas representative about a manual baseline review.

### When a baseline can't be computed yet

A cluster needs to be at least **14 days old**, and Atlas needs to be actively managing it, before a baseline is computed at all. If your cluster is younger than that, or hasn't yet reached active management, it simply won't appear in the Savings Report yet. It won't show zero savings; it will be absent entirely. This is intentional: it avoids showing noisy, unstable savings numbers for clusters that just connected.

Once your cluster clears both thresholds, Atlas computes a baseline on its next daily run, and the cluster starts appearing in your report.

### Manual baseline adjustments

Sometimes the modeled baseline doesn't reflect reality. For example, if you have out-of-band knowledge of your actual pre-Atlas-AI costs, or the automatic calculation landed on a window that wasn't representative. Two adjustments are available through your Atlas representative:

* **Override the baseline**: hand-set any of the four parameters directly. Fields you don't specify keep their current values. Once a baseline is overridden, Atlas's automatic daily calculation will never touch that cluster again, so the override sticks until it's changed manually.
* **Recalculate over a custom date range**: re-run one of the three calculation methods above, but over a date range you choose instead of the one Atlas selected automatically.

Both baseline types are labeled **Overridden** in the report once applied, so you can always tell when a baseline reflects a manual adjustment rather than the automatic calculation.

### Baseline assignment summary

| Your cluster's situation                                                                 | Baseline you'll get                                                          |
| ---------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| At least 7 days of history before Atlas took over node management                      | Cluster history                                                              |
| Not enough own history, but your org has another cluster with a Cluster history baseline | Peer clusters                                                                |
| No own history and no eligible peers in your org                                         | Industry average                                                             |
| Younger than 14 days, or Atlas hasn't yet taken over management                        | No baseline yet                                                              |
| Atlas representative manually set the baseline                                         | Overridden (see [Manual baseline adjustments](#manual-baseline-adjustments)) |

## Calculation formulas

### Actual cost

Actual cost is the straightforward sum of real dollars spent on provisioned CPU and RAM over the period you're viewing, with no modeling involved.

```text
actual cost = actual CPU cost + actual RAM cost
```

### Projected cost

Projected cost is the modeled estimate of what you'd be paying without Atlas. It's the number every savings figure is derived from, so it's worth understanding piece by piece.

```text
projected CPU cost =
  if (day is on/after the baseline period ends):
      greatest(original requested core-hours, current requested core-hours)
        × CPU overprovisioning factor
        × greatest(current effective $/core-hour, adjusted baseline $/core-hour)
  else:
      actual CPU cost
```

RAM is calculated the same way, using the RAM overprovisioning factor and RAM unit cost. The RAM baseline unit cost is adjusted to your commitment tier identically.

Reading it piece by piece:

* **Before your baseline period ends, projected cost equals actual cost.** Atlas hadn't optimized the cluster yet during the baseline window, so there's nothing to attribute savings to. Savings only start accruing once the baseline period closes. See [Realized savings](savings-report.md#when-savings-start-appearing).
* **Demand is the larger of original or current requested core-hours.** If a workload's requests were rightsized down, projection uses the larger, pre-rightsizing original request. If a workload's requests grew above its original level, projection uses the larger current request. Either way, the model never assumes you'd be running leaner than you actually run today.
* **Demand is scaled up by the overprovisioning factor**, the ratio your cluster historically ran at without Atlas's node-level optimization.
* **The result is priced at whichever is higher: your current effective rate or your adjusted baseline unit cost.** This ensures the counterfactual is never underpriced. If Atlas lowered your unit cost (for example, through spot adoption), the projection reflects the higher pre-optimization rate.
* **The baseline rate is adjusted to your current commitment tier** before this comparison, so a change in your commitment tier doesn't distort the result (see [Commitment tier changes](#commitment-tier-changes)).

### Realized savings

```text
realized savings = projected cost − actual cost
```

This is zero for any cluster without a baseline, and zero for any day inside the baseline period, since projected cost equals actual cost during that window (see above).

### Workload autoscaler savings

```text
workload autoscaler savings =
  (original requested core-hours − current requested core-hours)
    × greatest(current effective $/core-hour, adjusted baseline $/core-hour)
```

counted only on days when workload autoscaling (VPA) was actually active on the cluster.

> 📘 Note
>
> This calculation runs independently of the baseline period. Unlike realized savings, workload autoscaler savings can accrue even on days that fall inside your baseline window, since rightsizing isn't tied to when Atlas took over node management.

### Reported cost fields

The report combines the CPU and RAM figures above into the fields you see in the UI and API:

| Field                       | Value                                                                                                                                   |
| --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| Actual cost                 | Actual CPU cost + actual RAM cost                                                                                                       |
| Projected cost              | Projected CPU cost + projected RAM cost                                                                                                 |
| Autoscaler savings          | Autoscaler CPU savings + autoscaler RAM savings                                                                                         |
| Workload autoscaler savings | Workload autoscaler CPU savings + workload autoscaler RAM savings                                                                       |
| Total savings               | Currently equals autoscaler savings only, since the workload autoscaler impact is already included by decreasing the CPU and RAM demand |

## Listing prices vs. discounted prices

The report can show actual cost and current unit price using either your **discounted** (effective, negotiated) rates or **listing** (on-demand list) prices, ignoring any discounts. Toggling this changes what "actual cost" means, but not your baseline:

* **Discounted (default):** reflects what you actually pay: negotiated rates, real spot prices. To ensure these reflect your actual cloud spend, configure [Price adjustments](discount-engine-overview.md).
* **Listing:** uses on-demand list prices, useful for showing savings against undiscounted, rack-rate pricing.

Two things to know about this toggle:

* **Listing prices aren't available before August 6, 2025.** If you select a time range that starts before this date, the report automatically falls back to discounted prices regardless of your selection.
* **Your baseline unit cost is always based on discounted pricing**, since that's what was available when your baseline was calculated.
* **In Discounted mode, that stored baseline rate is rescaled to your current commitment tier** (see [Commitment tier changes](#commitment-tier-changes)).
* **If you switch to listing prices**, actual and current costs rise to list-price levels while the baseline floor stays at your discounted historical rate. The "whichever is higher" pricing rule described in [Demand and price selection](#demand-and-price-selection) still applies on top of that.

## Commitment tier changes

Your baseline unit costs capture the discounted rates in effect during the baseline period, after whatever cloud commitments you already had. Your commitments are, for example, an enterprise agreement, committed-use discounts, Savings Plans, or your configured [Price adjustments](discount-engine-overview.md). If those commitments change afterwards, your **commitment tier** moves with them, and the stored baseline rate no longer reflects what that capacity would cost you today. Savings figures then misattribute the change: overstated when your discounts deepened, understated when they shrank.

To keep the comparison fair, the report rescales the baseline unit cost to your current commitment tier before using it in any savings calculation:

```text
adjusted baseline unit cost = baseline unit cost × (current discount level ÷ baseline-window discount level)
```

The discount level is the ratio of your effective (discounted) cost to the listing cost of the same capacity, measured per cluster and separately for CPU and RAM. For example, if your effective CPU price is $0.021 per core-hour and the listing price for that same capacity is $0.030, your CPU discount level for that window is 0.7. The current discount level is measured over the report's selected time range. The baseline-window discount level is measured over the cluster's baseline period.

In practice:

* **Discounts deepened since the baseline**: the baseline rate is adjusted down, so savings no longer credit a price reduction Atlas didn't cause.
* **Discounts shrank since the baseline**: the baseline rate is adjusted up accordingly.
* **Commitment unchanged**: the ratio is 1 and the stored baseline rate is used as-is.
* **Manually overridden baselines are never rescaled**: the value set by your Atlas representative is used exactly as provided (see [Manual baseline adjustments](#manual-baseline-adjustments)).

Two boundaries to be aware of:

* The adjustment applies only in the default **Discounted** pricing mode. In Listing mode the current side is priced at list, and the stored baseline is used without rescaling.
* Discount levels are measured against listing-price data; see [Listing prices vs. discounted prices](#listing-prices-vs-discounted-prices) for the data-availability boundary. For a baseline window that falls entirely before listing-price data exists, the baseline-window discount level is treated as 1 (undiscounted).

## Historical data quality

A small number of data-quality safeguards apply to historical data, mainly affecting workloads' *original* (pre-rightsizing) requests:

* Some workload types had a data-collection gap that produced zero values for original requests before Atlas resolved it in early 2026; the report substitutes current requests as a reasonable proxy for those older data points.
* Values that are implausibly large (well beyond what any real pod would request) are excluded and replaced with a sane fallback, to guard against occasional data-collection errors.
* For time ranges that start before October 2025, the report uses the earliest reliable snapshot of original requests it can find, rather than assuming zero.

These safeguards only affect historical accuracy for older data. They don't change how current and future savings are calculated.

## Scenario reference

### Per-day savings

| Day is...                      | Workload autoscaling active? | Realized savings that day | Workload autoscaler savings that day |
| ------------------------------ | ---------------------------- | ------------------------- | ------------------------------------ |
| Inside the baseline period     | No                           | None (projected = actual) | None                                 |
| Inside the baseline period     | Yes                          | None (projected = actual) | Accrues                              |
| After the baseline period ends | No                           | Accrues                   | None                                 |
| After the baseline period ends | Yes                          | Accrues                   | Accrues                              |

### Demand and price selection

The projected-cost formula always picks the higher option for both demand and price. If Atlas lowered your unit cost (for example, through spot adoption), the projection reflects the higher pre-optimization rate:

* **Demand** uses whichever is larger: original (pre-rightsizing) requests or current requests.
* **Price** uses whichever is higher: your current effective unit cost or your adjusted baseline unit cost.

## Known limitations

* **Clusters without a baseline are absent, not zeroed.** A cluster that's too new, not yet under active management, or that produced no usable data in any of the three baseline methods simply doesn't appear in either report view. It isn't shown with zero savings.
* **Clusters created before node-level monitoring became available (February 2025) may have a less precise baseline window.** For these clusters, the baseline period ends when Atlas took over credentials rather than when it took over active node management, which can make the overprovisioning and cost factors slightly less representative.
* **An industry-average baseline draws from Atlas's entire fleet, not just your organization.** This is intentional (it's meant to be a genuine cross-fleet reference point), but it means an industry-average baseline for your cluster is influenced by data from unrelated organizations.
* **Savings only start after your baseline period ends.** A report whose entire time range falls inside the baseline window will show projected cost equal to actual cost and zero realized savings (workload autoscaler savings may still appear, since that calculation isn't tied to the baseline window).
* **Baseline factors don't distinguish between on-demand, spot, and fallback capacity.** All four baseline parameters are single averaged values applied uniformly across lifecycles, so spot-versus-on-demand nuance isn't captured in the model.

<Cards columns={2}>
  <Card title="Realized savings" href="https://docs.atlas.example/docs/savings-report" icon="fa-piggy-bank">
    What the Savings Report shows and the two sources of savings it tracks.
  </Card>
  <Card title="Available savings" href="https://docs.atlas.example/docs/available-savings" icon="fa-chart-line">
    See how much you could save by optimizing your cluster.
  </Card>
</Cards>