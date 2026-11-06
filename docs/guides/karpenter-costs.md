---
title: Karpenter costs
description: Calibrate costs to Karpenter's node prices and price each workload at the nodes it ran on, spot or on-demand.
---

On a cluster run by Karpenter, resourcerer can price capacity at what the nodes really cost instead of fixed rates, and
price each workload at the nodes it actually ran on. A workload on spot then shows smaller savings than the same
requests on on-demand, which is closer to what an apply changes on the bill.

## What it needs

Prometheus must scrape Karpenter's controller metrics, which include:

- `karpenter_nodes_allocatable`: each node Karpenter manages, with its instance type, zone and capacity type;
- `karpenter_cloudprovider_instance_type_offering_price_estimate`: the price estimate of each instance type offering.

Check that they are there:

```promql
count(karpenter_nodes_allocatable{resource_type="cpu"})
count(karpenter_cloudprovider_instance_type_offering_price_estimate)
```

## Calibrate the rates

```yaml title="values.yaml"
cost:
  source: karpenter
```

`auto`, the default, does the same when the metrics are found, and falls back to `cost.rates` quietly when they are
not. `karpenter` makes the fallback an error on the Overview and Settings pages, so a scrape that broke is noticed.

The calibration prices every Karpenter node, divides by what its capacity costs at `cost.rates`, and scales the rates
by that **cluster factor**, keeping their CPU to memory ratio. It runs every `cost.karpenter.refreshInterval` (`10m`).
The **Settings** page shows the configured and the calibrated rates; a failed refresh keeps the last good calibration,
and with `karpenter` three failures in a row turn it into an error.

## Price each workload at its nodes

```yaml title="values.yaml"
cost:
  source: karpenter
  allocation: node
```

Each workload is then priced at the nodes its pods ran on over its learning period, weighted by time and requests,
at each node's current price: what those nodes cost today, the estimate the savings are about. The
**How the cost is calculated** section of the Impact tab shows:

- the **workload factor** next to the cluster factor;
- the **coverage**: the share of its requests on nodes with a known price (the rest is priced at the cluster factor);
- the **placement**: the share on spot, on-demand and reserved capacity, and the instance types that carried most of
  it.

Reserved capacity is priced as on-demand: Karpenter does not know what a reservation costs you.

## Costs of the queries

Karpenter exports a price for every offering of the region, often tens of thousands of series. resourcerer reads node
prices only for the instance types the nodes had in the window, and runs the placement queries per workload at its
regeneration, cached by the server for 5 minutes when you open a workload page.

## Without Karpenter

Use `static` and set `cost.rates` to your prices: see [Cost estimates](../configuration/cost-estimates.md#choosing-the-rates).
Savings Plans, Reserved Instances and negotiated discounts are not modelled yet; scale the rates by your average
discount in the meantime.
