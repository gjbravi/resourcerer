---
updatedAt: 2026-07-20T08:40:14.000Z
---

# Enterprise commitments

> 🚧 Limited Availability Feature
>
> This feature is currently available through feature flags. Contact us to enable access for your organization.

Enterprise commitments let a central team import a Reserved Instance, Savings Plan, or Committed Use Discount once — at the Enterprise organization level — and have every child organization automatically consume it. There is no need to re-import the same commitment in each child organization or maintain separate assignments per team.

This page explains what Enterprise commitments are and how utilization is tracked and displayed across organizations.

## What changes with Enterprise commitments

Without Enterprise commitments, a commitment purchased at the cloud-provider level must be imported separately into each Atlas child organization. This causes the same commitment to appear multiple times in the platform, each copy tracking only its own organization's usage. The result is that every child organization sees artificially low utilization — because none of them can see what the others are consuming.

Enterprise commitments fix this by storing the commitment once, at the Enterprise (parent) organization level. All child organizations draw from the same capacity pool, so the utilization figure each organization sees reflects the actual cross-organization consumption.

## Supported commitment types

Enterprise commitments support the same types as standard commitments: AWS Reserved Instances, AWS Savings Plans, GCP resource-based CUDs, GCP Compute Flexible CUDs, Azure Reserved Instances, and Azure Savings Plans.

> 📘 Note
>
> AWS On-Demand Capacity Reservations are not shared across organizations. They remain scoped to the organization that imports them. For details, see [AWS capacity reservations](aws-capacity-reservations.md).

## How utilization is displayed

Because a single commitment is now shared across multiple child organizations, the Commitments page shows utilization broken down by organization.

**Enterprise organization view** — shows the total utilization across all child organizations, with a per-organization breakdown: each child org's share of CPU, memory, and spend consumed.

**Child organization view** — shows that organization's own utilization alongside an **Other organizations** figure representing the combined consumption from all sibling organizations. This gives each team full visibility into how much of the shared commitment is available to them.

> 📘 Note
>
> Child organization admins can view commitment priorities but cannot edit them. Priority changes must be made at the Enterprise level.

## Known limitations

* **Priority is read-only for child organizations.** Only Enterprise admins can adjust commitment priorities.
* **Capacity reservations are not shared.** AWS On-Demand Capacity Reservations remain scoped to the importing organization.
* **Region matching is required.** A child organization only consumes an Enterprise commitment if its clusters run in a matching cloud provider region.

<Cards columns={2}>
  <Card title="Commitments" icon="fa-bookmark" href="https://docs.atlas.example/docs/commitments">Single-org commitment lifecycle, supported types, and the Commitments page</Card>
</Cards>