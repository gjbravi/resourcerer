---
updatedAt: 2026-05-12T14:12:41.000Z
---

# Overview

# Price adjustments

Use Price adjustments to apply custom pricing to your cloud resources in Atlas. This feature enables the platform to use customer-specific prices instead of public list prices, helping align Atlas's cost reporting and optimization decisions with your real cloud costs.

## How it works

Atlas allows application of price adjustments to specific instances or resources based on parameters like:

* Instance type names and families
* Regions and availability zones
* Operating systems
* Resource lifecycle (spot/on-demand)
* Specific clusters

You can configure multiple discounts to affect the same VM shape, but the behavior will differ depending on the discount target. CPU, memory, and instance are three independent discount targets. In cases where multiple discounts affect the same target, Atlas will apply the one that results in the best value. For example, if you have two discount records for CPU and both apply to the same VM shape, Atlas will choose the better deal.

On the other hand, if multiple discounts affect the same VM because one is configured to apply to CPU cost while another targets instance cost, these are treated as two separate discount targets, and both will be applied.

Lastly, surcharges are always applied last—on top of previously applied price reductions. If multiple surcharges affect the same target, all of them will be applied.

## Supported discount types

There are three price adjustment types currently supported in the platform

* **Custom pricing** - replaces the public list price for the target resource with the one specified
* **Discounts** - reduces the public list price for the target resource by applying a discount. There are two supported discount options:
  * **Percentage discount**: Reduces the public list price by a specified percentage
  * **Flat discount**: Reduces the public list price by a specified $ value
* **Surcharges** - increases the public list price by a specified $ value

## Configurable resources

You can apply discounts to the following resources:

* CPU
* Memory
* Extended memory (for GCP custom instances)
* Whole instances (e.g., -10% on a particular instance type)

> 📘 Note
>
> GPU discounts are not currently supported.

## Current limitations

Currently, Price adjustments feature has the following limitations:

* Only single-tier discounts are supported. Discounts cannot target specific resource levels (e.g., a discount on a specific amount of memory)
* When applying a price adjustment to an instance family (e.g., 10% off N1 family on GCP), it applies to all eligible instances in that family regardless of their memory usage
* There may be up to a one-hour delay for new adjustments to be reflected in reporting values due to caching

## Impact on Atlas features

When you enable price adjustments, they affect:

* Cost reporting and monitoring
* Autoscaling decisions
* Rebalancing recommendations
* Node cost calculations

> ❗ Important
>
> Price adjustments are not applied retroactively to historical cost data. They only affect prices from the time they are enabled forward.

## Example use cases

Here are some common price adjustment scenarios:

1. **Global enterprise discount**
   * 10% off all compute resources
   * Applies across all clusters
   * No time limitation
2. **Regional spot instance discount**
   * 15% off memory for spot instances
   * Specific to us-east-1 region
   * Applies to all instance families
3. **Instance family discount**
   * Fixed $1/GB extended memory price
   * For N1 and E2 custom instances on GCP
   * Limited to specific clusters
4. **Instance cost increase to account for 3rd party observability agent license cost**
   * Flat $0.15 surcharge
   * For all instances
   * No time or cluster scope limitation