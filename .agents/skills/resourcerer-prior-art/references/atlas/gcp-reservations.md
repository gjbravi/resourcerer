---
updatedAt: 2026-08-26T09:24:06.000Z
---

# GCP capacity reservations

Learn how to use Google Cloud Compute Engine on-demand reservations with Atlas to guarantee VM availability for critical workloads.

GCP reservations guarantee Compute Engine VM availability in a specific zone for a specific machine configuration. Unlike [Committed Use Discounts (CUDs)](commitments.md), which reduce costs through long-term spend or resource commitments, reservations ensure that the VMs you need are available when you need them, regardless of overall zonal capacity constraints.

Atlas treats reservations as first-class commitments. Once imported, they appear on the **Commitments** page under the **Reservations** tab, where you can assign them to clusters, track their utilization, and configure the autoscaler to provision nodes into reserved capacity.

Atlas supports **Compute Engine on-demand reservations**, in both consumption modes:

**Automatically consumed reservations** are consumed by any VM in the project whose properties match the reservation (machine type, zone, GPUs, Local SSDs). This is the default consumption type in GCP and is the closest equivalent to how Atlas treats CUDs — no explicit targeting is required.

**Specifically targeted reservations** are consumed only by VMs that explicitly reference the reservation by name. In Atlas, this requires targeting the reservation from a Node Template.

Both single-project and shared (cross-project) reservations are supported. [Google Cloud documentation](https://cloud.google.com/compute/docs/instances/reservations-overview).

> 📘 Key difference from CUDs
>
> Specifically targeted reservations require explicit targeting through [Node Templates](node-templates.md). The node autoscaler does not automatically consume them the way it does with CUDs. You must configure a Node Template to target the reservation before the autoscaler provisions nodes into it. Automatically consumed reservations are an exception — Atlas handles them the same way it handles CUDs in terms of consumption.

## How reservations work in Atlas

The overall workflow is:

1. **Create capacity in GCP.** Create on-demand reservations in the zones and machine types you need.
2. **Import into Atlas.** Run the commitments import script or wait for an active Cloud Connect integration to sync. Atlas discovers reservations by calling the Compute Engine `reservations.aggregatedList` API.
3. **Assign to clusters.** On the **Commitments** page, assign the imported reservation to the clusters that should use it.
4. **Configure Node Templates.** For specifically targeted reservations, select which Node Templates should target the reservation for each assigned cluster. The autoscaler only provisions nodes into a specifically targeted reservation when a Node Template explicitly targets it.
5. **Run workloads.** When pending pods match the constraints of a Node Template targeting the reservation, the autoscaler provisions nodes into the reserved capacity first.

> 📘 Two ways to link reservations and Node Templates
>
> You can configure the relationship between reservations and Node Templates from either direction: in the **commitment detail drawer**, assign one reservation to multiple Node Templates across clusters, or in the **Node Template editor**, target multiple reservations from a single template.

### Provisioning order

When a Node Template targets one or more reservations, the autoscaler follows this priority order:

1. **Reservations** — Provisions into the targeted reservation first, up to the reserved VM count. Atlas is also aware of and will take into account any external usage of a reservation by systems outside of Atlas-managed clusters.
2. **Other commitment-covered on-demand** — If the reservation is fully utilized, it falls back to CUD-covered capacity if available.
3. **Regular on-demand/Spot** — Standard on-demand or Spot VMs if no commitment capacity remains.

If the Node Template lacks fallback options and the reservation is fully utilized, pods remain pending until capacity becomes available.

### Reservation statuses

Each reservation has a lifecycle status visible on the **Commitments** page:

**Active** — The reservation is ready and available for use. The autoscaler can provision nodes into active reservations when matching pending pods exist.

**Pending** — The reservation is still being created or updated in GCP. The autoscaler does not provision into it until it becomes active.

**Deleted** — The reservation no longer exists in GCP, either because it was deleted manually or because its auto-delete time was reached. Atlas reflects this status after the next Cloud Connect sync (up to 1 hour). The autoscaler ignores deleted reservations.

> 📘 Auto-delete
>
> GCP reservations can be created with an auto-delete time, after which Compute Engine removes them automatically. Atlas does not manage the auto-delete setting — it reflects the reservation's state after it is applied. Nodes running when a reservation is deleted keep running and are billed at the standard rate.

## Prerequisites

Before importing reservations, ensure the following:

**GCP permissions.** The service account used by Cloud Connect must be able to read reservations in the project — the `compute.reservations.get` and `compute.reservations.list` permissions, included in the `roles/compute.viewer` role. See [Cloud permissions](cloud-permissions.md) for the full set of permissions. The script for uploading commitments requires the same permissions.

**Existing integration.** You need either Cloud Connect configured for your GCP project or use the commitments import script. Reservations are discovered through the same import mechanism used for CUDs.

**Cluster connected.** At least one GCP cluster must be connected to Atlas in the same region as the reservation for it to be visible in the user interface after importing.

**Shared reservations.** For a shared reservation to be discovered, Atlas needs read access to the **owner project** — the project where the reservation was created. Granting access only to a consumer project is not sufficient, as GCP lists each shared reservation only in the project that created it.

## Importing reservations

Reservations are imported using the same workflow as other GCP commitments.

If you already have a commitments import or Cloud Connect set up, newly created reservations are discovered automatically during the next sync cycle.

To import manually:

1. Navigate to **Optimization** > **Commitments** (Classic Console) or **Settings** > **System** > **Commitments** (Enhanced Console) in the left sidebar.
2. Click **Upload commitments** in the top right corner.
3. Run the script. Atlas discovers all commitment types, including reservations.

Once imported, reservations appear under the **Reservations** tab on the Commitments page:

<Image align="center" caption="The Reservations tab on the Commitments page showing imported reservations" src="https://files.readme.io/5c6fc3cd54b50c83355294843557645d3a34f270aa8ffc67d74dd1836930edcd-gcp-commitments-reservations-tab.png" />

## Assigning reservations to clusters

<Columns layout="auto">
  <Column>
    After importing, assign the reservation to the clusters that should use it:

    1. Click on a reservation name to open the detail drawer.
    2. Under **Assign clusters to the commitment**, select the clusters that should have access to this reservation.
    3. For each assigned cluster, select at least one **Node Template** from the dropdown. This is mandatory for specifically targeted reservations — the autoscaler requires explicit Node Template targeting to consume them.
    4. Click **Save**.
  </Column>

  <Column>
    <Image align="center" caption="Assigning clusters and Node Templates to a reservation in the detail drawer" src="https://files.readme.io/b8c37ac3337f0d4f85f7f8e55a289071477a840bff91ec284a670acacc77c136-gcp-reservation-assign-clusters-drawer.png" />
  </Column>
</Columns>

The detail panel also allows you to configure:

**Use Commitment for cluster autoscaling** — When enabled, Atlas uses this reservation when scaling up assigned clusters. When disabled, the reservation is not used in autoscaling decisions, but its utilization is still tracked.

**Set maximum usage limit** — Optionally restrict Atlas to use only a percentage of the reservation's capacity, reserving the remainder for workloads outside Atlas-managed clusters.

> 📘 Consumption type determines what is required
>
> For **specifically targeted** reservations, Node Template selection is mandatory. If no Node Template is selected, the autoscaler cannot provision into the reservation. For **automatically consumed** reservations, matching nodes consume the reservation without Node Template targeting, though you can still target them explicitly to control which workloads land on reserved capacity.

## Configuring Node Templates for reservations

To target a reservation from a Node Template, use the **Target capacity reservations** field in the Node Template configuration:

1. Navigate to your cluster's **Configuration** > **Node autoscaler** page, then choose the **Node templates** tab.
   <Accordion title="I'm using Classic Console" icon="fa-info-circle">
     Navigate to your cluster's **Autoscaler** > **Node templates** page.
   </Accordion>
2. Create a new Node Template or edit an existing one.
3. In the **Target capacity reservations** section, select one or more reservations from the dropdown.
4. Save the template.

GCP consumes a reservation only when the VM's properties match the reservation's properties exactly, so any mismatch means the node is created outside the reservation. Atlas matches nodes to a reservation on machine type, zone, and GPU (accelerator type and count). Atlas does **not** currently match on the reservation's minimum CPU platform or Local SSD configuration, so a reservation that specifies either of those is not used by the autoscaler. See [Known limitations](#known-limitations).

For specifically targeted reservations, Atlas provisions nodes with a GCP reservation affinity (`consumeReservationType: SPECIFIC_RESERVATION`, key `compute.googleapis.com/reservation-name`) that references the reservation by name, so the VM consumes only that reservation. Automatically consumed reservations do not require reservation affinity — a matching VM consumes them without explicit targeting.

<Image align="center" caption="Node Template configuration showing the field for targeting reservations" src="https://files.readme.io/3f7c0489b5a51dded508075577b8761ebbe7a61244845c91ccb0a34ecea0e99f-gcp-node-template-target-reservations.png" />

For full Node Template documentation, see [Node Templates](node-templates.md).

## Automatically consumed vs. specifically targeted reservations

GCP reservations have a consumption type that determines which VMs can use them, and Atlas handles each differently:

**Automatically consumed reservations** (the GCP default) are applied to any VM in the project matching the reservation's machine type, zone, and optional resources. Atlas-provisioned nodes consume this capacity as soon as they match, without any Node Template targeting. Note that other systems in your project may also consume the same capacity.

**Specifically targeted reservations** are consumed only by VMs that reference the reservation by name. Atlas provisions nodes with the corresponding reservation affinity only when a Node Template is explicitly tied to the reservation in the Commitments interface.

If a project has VMs that can automatically consume multiple reservations, GCP consumes all single-project reservations before any shared ones.

## Shared reservations

A shared reservation can be consumed by VMs in the project that created it (the *owner project*) and in any project it is shared with (*consumer projects*), as configured through the reservation's share settings in GCP.

Atlas imports a shared reservation from the owner project and makes it available to clusters in any of the projects that are connected to your Atlas organization and included in the reservation's share settings. Utilization is tracked per reservation across all consuming clusters and projects.

Keep the following in mind:

* Cloud Connect must have read access to the owner project for the reservation to be discovered at all.
* If a cluster runs in a project that is not listed in the reservation's share settings, its nodes cannot consume the reservation even if the reservation is assigned to that cluster in Atlas.
* Sharing settings themselves are managed in GCP, not in Atlas. Changes are reflected after the next sync.

## Setting up a reservation

1. **Create the reservation in GCP.** Create an on-demand reservation for the machine type, zone, and VM count you need. If you plan to use specific targeting, create it as a specifically targeted reservation.

2. **Import into Atlas.** Import commitments or wait for Cloud Connect to sync. Reservations appear under the **Reservations** tab with their current status.

3. **Assign to your cluster.** Select the cluster and, for specifically targeted reservations, the Node Templates that should use this reservation.

4. **Configure the Node Template.** Select the reservation under **Target capacity reservations**. Ensure the Node Template's constraints (machine type, zone, and GPU configuration) match the reservation's properties — GCP requires an exact match for the reservation to be consumed. For specifically targeted reservations, Atlas provisions nodes with a GCP reservation affinity that references the reservation by name so the VM consumes only that reservation.

5. **Deploy workloads.** The autoscaler provisions nodes into the reservation when matching pods are pending, up to the reserved VM count. If the reservation is fully utilized, the autoscaler falls back to other options configured on the Node Template.

6. **Monitor utilization.** Track utilization on the **Commitments** page. Because a reservation is billed for as long as it exists regardless of whether it is consumed, maximizing utilization directly reduces wasted spend.

## Interaction with Rebalancer

The [Rebalancer](rebalancing.md) respects reservations when evaluating node replacements. Nodes backed by active reservations are not replaced with non-reserved alternatives. The Rebalancer can also help re-pack workloads onto reserved nodes to maximize utilization of a reservation.

## Troubleshooting

### Reservation not being utilized

If the autoscaler is not provisioning into an available reservation, verify:

**Node Template targeting.** For specifically targeted reservations, the reservation must be selected in the **Target capacity reservations** field of at least one enabled Node Template assigned to the cluster.

**Cluster assignment.** The reservation must be assigned to the cluster, and Node Templates must be selected for that cluster.

**Reservation status.** The reservation must be **Active**. Pending reservations are not used for provisioning, and deleted reservations are ignored.

**Property match.** GCP requires an exact match between the VM and the reservation. Atlas matches on machine type, zone, and GPU type and count. Atlas does not match on the reservation's minimum CPU platform or Local SSD configuration, so a reservation that specifies either is not used by the autoscaler (see [Known limitations](#known-limitations)). A reservation with a compact placement policy can only be consumed by VMs using the same policy.

**Project and share settings.** The cluster's project must be either the owner project or one of the consumer projects listed in the reservation's share settings.

**Matching workloads.** Pending pods must match the Node Template's constraints (labels, tolerations, resource requests). The provisioned VM's properties (machine type, zone, GPU) must also match the reservation's properties exactly for GCP to consume the reservation.

**Autoscaling toggle.** The **Use Commitment for cluster autoscaling** toggle must be enabled on the reservation.

**Quota.** Reserved resources consume project quota. If the project is out of quota for the reserved machine type, node creation fails even though reserved capacity exists.

## Known limitations

**On-demand reservations only.** Atlas supports Compute Engine on-demand reservations. Future reservations, including future reservations in calendar mode, are not currently supported.

**Zonal scope.** Reservations are zonal resources. A reservation can only be consumed by nodes in the same zone, which constrains any Node Template that targets it to a single zone.

**Minimum CPU platform and Local SSD are not matched.** Atlas matches nodes to a reservation on machine type, zone, and GPU accelerators (type and count). It does not currently match on the reservation's minimum CPU platform or Local SSD configuration. If a reservation specifies a minimum CPU platform or Local SSDs, the autoscaler will not provision nodes into it. GPU accelerators are supported.

**Shared reservation discovery.** Shared reservations are only discovered through the owner project. If Cloud Connect has access only to a consumer project, the reservation will not appear in Atlas.

**Shared reservation tracking.** If multiple clusters or Node Templates target the same reservation, Atlas tracks utilization per reservation across all of them. GCP enforces capacity limits, so over-assignment may result in some workloads not receiving reserved capacity.

## See also

* [Commitments](commitments.md) — Overview of discount-based commitments (CUDs, RIs, Savings Plans)
* [Node Templates](node-templates.md) — Configure Node Templates including reservation targeting
* [Cloud permissions](cloud-permissions.md) — Required GCP permissions for importing reservations
* [About reservations](https://cloud.google.com/compute/docs/instances/reservations-overview) — Google Cloud documentation
* [Consume a reservation](https://cloud.google.com/compute/docs/instances/reservations-consume) — Google Cloud documentation