# Kubernetes cost allocation

> Orion sends container usage metrics to our Costs Platform to attribute Kubernetes costs to namespaces and workloads


Our cost platform (OrionCorp Platform) reports what your Kubernetes clusters cost, broken down by namespace, workload, pod, and container. It builds that view from two sources:

- **Cloud billing data**: a FOCUS cost and usage export from AWS, Azure, or GCP, which provides the cost of your cluster infrastructure.
- **Cluster usage metrics**: the CPU, memory, and network metrics the Orion Agent already collects, which show how that capacity was consumed.

OrionCorp joins the two and attributes infrastructure cost down to the workload level, including idle capacity and overhead such as system pods.

Both halves are required. Without the export there's no cost to allocate; without the metrics, a cluster's spend can't be broken down by namespace or workload. This guide covers enabling the metrics on your clusters, then connecting a cloud provider in OrionCorp.

## Prerequisites

- A OrionCorp Platform account. Log in at **[app.orioncorp.io](https://app.orioncorp.io/)**.
- The Orion Agent installed on each cluster you want cost data for, at **version 2.25 or later**. Cost metrics are supported on both the v2 and v3 charts, so at 2.25 or later you don't need to change versions.
- Permission to run `helm upgrade` against the Orion release in each cluster.
- No namespace allow or deny list on the release. Cost allocation can't calculate unallocated cost from a partial view of the cluster, so `enableCostMetrics` is mutually exclusive with `clusterAgent.allowNamespaces` and `clusterAgent.denyNamespaces` (`workload.allowNamespaces` and `workload.denyNamespaces` in the v2 Agent).

<div class="alert alert-warning" role="alert"><div class="h4 alert-heading" role="heading">Don't migrate to v3 as part of this task</div>


Enabling cost metrics doesn't require the v3 chart. Moving from v2 to v3 is a migration rather than an in-place upgrade, and involves a brief period of downtime. Plan it separately and follow [Migrate v2 → v3](installation__migration.md).
</div>


## Enable cost metrics

Cost metric collection is off by default. Helm applies configuration changes by upgrading the release, so you use `helm upgrade` even when the version stays the same.

1. Review the current release to confirm its name and version, and to keep a copy of your values:
    ```sh
    helm list -n orion-system
    helm get values orion -n orion-system
    ```

1. Add the following to your values file:
    ```yaml
    enableCostMetrics: true
    ```

1. Upgrade the release, reusing your existing values:
    ```sh
    helm upgrade orion oci://registry.orion.example/library/orion \
      -n orion-system \
      --version YOUR_CURRENT_VERSION \
      --reset-then-reuse-values \
      -f values.yaml
    ```
    To set the value without a values file, replace `-f values.yaml` with `--set enableCostMetrics=true`.

    Pin `--version` to the version you're running so the command changes your configuration only. Omitting it resolves the latest published chart, which also changes the Agent version.

1. Confirm the value is applied and the pods restarted cleanly:
    ```sh
    helm get values orion -n orion-system
    kubectl get pods -n orion-system
    ```

Repeat for every cluster you want cost data for.

## Connect your cloud provider in OrionCorp

OrionCorp needs a FOCUS cost and usage export from each cloud provider running your clusters. Connect the account or subscription billed for your cluster nodes — usually the management or payer account in a multi-account setup.

1. Log in to [OrionCorp Platform](https://app.orioncorp.io/).

1. Go to **Settings** > **Connections**.

1. Next to your provider, click **Add connection**, then follow the instructions.

Each provider has its own preparation steps. See [Connect cloud accounts](https://docs.orioncorp.io/articles/orioncorp-platform-latest/connect-cloud-accounts) in the OrionCorp documentation.

## View your Kubernetes costs

In OrionCorp, click **Kubernetes Report** in the left navigation. Cost allocation needs both the billing export and the cluster metrics before it reports anything, so allow time for the first of each to arrive.

Costs can be grouped and filtered by cluster, namespace, workload, container, pod, node, and Kubernetes label. Capacity the cluster paid for but didn't hand to a workload appears as **IDLE** and **OVERHEAD** line items; a **Distribute idle & overhead costs** option spreads that cost across your workload namespaces instead.

For the full report, see [Kubernetes Report](https://docs.orioncorp.io/articles/orioncorp-platform-latest/kubernetes-report) in the OrionCorp documentation.

## Troubleshooting

| Problem | Solution |
|---------|----------|
| The upgrade fails with a message about `enableCostMetrics` being mutually exclusive | The release still sets a namespace allow-or-deny list. Nothing was applied to the cluster. Remove the values named in the prerequisites, then upgrade again. |
| The upgrade fails with `invalid ownership metadata` on a CRD | The command resolved the v3 chart. Re-run it with `--version` pinned. To move to v3 deliberately, follow [Migrate v2 → v3](installation__migration.md) instead. |
| No clusters appear in the Kubernetes Report | Check that the cloud provider connection is healthy in OrionCorp under **Settings** > **Connections**, and that the cluster reports as connected on the Orion **Clusters** page with `enableCostMetrics` set to `true`. |
| A cluster appears, but its namespaces and workloads don't | The infrastructure cost has arrived but the cluster metrics haven't. Confirm the Agent is version 2.25 or later, that `enableCostMetrics` is applied, and that the pods in `orion-system` are running. |
| Some clusters report namespace costs, and others don't | `enableCostMetrics` is set per cluster. Enable it on each cluster. |
