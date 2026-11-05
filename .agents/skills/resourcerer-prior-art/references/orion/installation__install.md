# Basic installation

> Install Orion on a cluster


## Prerequisites

- A Kubernetes cluster (from any Kubernetes distribution, such as EKS, AKS, GKE, minikube, or kind)
- A Orion account — sign up for free at **[app.orion.example/signup](https://app.orion.example/signup)**
- In an enterprise environment, make sure a system administrator has addressed [additional prerequisites (such as firewall rules and proxies)](installation__install-adv.md#prerequisites)

## Install Orion {#install-orion}

1. Log in to **[app.orion.example](https://app.orion.example/)**.

    If you are setting up Orion for the first time, you will be automatically taken to the **Add Cluster** page — continue to step 3.

1. Click **Clusters** in the left navigation; then, on the **Clusters** page, click **Add Cluster**.

1. Give your cluster a name and click **Continue**.

    The name can't exceed 63 characters, can't start or end with a dash (`-`), and must follow the [RFC 1123 DNS label standard][rfc-1123].

1. Install Orion: Copy and run the generated `helm install` command, which looks similar to the following (where `my-cluster` is the name you chose for this cluster, and `CLIENT_ID` and `CLIENT_SECRET` are generated alphanumeric strings).

    This installs the Agent, Applier, and Forwarder into the `orion-system` namespace. To collect GPU metrics, enable the [Node Agent](installation__node-agent-config.md).

    ```shell
    helm install orion oci://registry.orion.example/library/orion \
      --create-namespace --namespace orion-system \
      --set clusterName="my-cluster" \
      --set authorization.clientID="CLIENT_ID" \
      --set authorization.clientSecret="CLIENT_SECRET"
    ```
    To install without the Applier, add `--set applier.enabled=false` to the command.

### Validate the installation

1. Return to the Orion UI and click **Check Connection**. You'll be redirected to the Overview page where you can see a Welcome message and confirmation that the Agent is connected.

1. After about 10 minutes, you can see your cluster and its workloads on the Clusters and Workloads pages.

1. In the left navigation, click **Workloads** and search for your cluster name. Make sure you see the `orion-cluster-agent` workload and any workloads that were already on your cluster.

## What happens after you install?

The Orion Agent starts collecting workload metrics as soon as you install.

### First 7 days: Metrics collection and preliminary recommendations

**Ten minutes** after installation, once initial metrics are collected, Orion provides its first _preliminary_ recommendation to give you a preview of potential optimization impact.

Preliminary recommendations are then generated:
- **Hourly**, for the first 24 hours after installing
- **Once daily**, on days 2 to 7 after installing

These can be viewed as a preview, but we recommend waiting for a complete recommendation before applying.

It typically takes about 7 days to collect enough metrics to generate an optimal recommendation (referred to as a _complete_ recommendation) for you to apply.

For this reason, we recommend **waiting 7 days before applying recommendations**.

### After 7 days: View and apply recommendations

Orion has collected sufficient metrics and now generates recommendations once daily for each workload. You can [change the schedule](configure__settings__schedule.md) at any time.

Review the projected impact of recommendations:
- At the estate level on the **Overview** page.
- At the cluster level on the **Clusters** page — drill down to view namespaces, then workloads.
- At the workload level on the **Workloads** page.
    - Click **View Summary** for a breakdown of workload statuses (useful for troubleshooting and checking workload health).
    - Click a workload name to view workload details and recommendation impact.

## Orion Applier {#orion-applier}

The Applier is installed and enabled by default. By default, **recommendations aren't applied automatically**, giving you control over how and when they're applied. The Applier allows you to:

- **Apply recommendations on demand** from the Orion UI.
- **Apply them automatically on a schedule**. For this method, you must also [enable automatic deployment](configure__settings__apply-settings.md).

For more information about how the Applier works, see [Orion Applier][concepts-applier].

### Continuous reconciliation {#applier-cont-reconciliation-desc}

You can control whether the Applier continuously reconciles workload resource settings to the recommended values. For details, see [Continuous reconciliation](installation__applier-config.md#continuous-rec) in the Applier configuration topic.

### Disabling the Applier

If you plan to deploy recommendations as part of a GitOps workflow or export recommendations as patches, you can disable the Applier at install time. Recommendations will continue to be generated; their application is up to you.

```shell
helm install orion oci://registry.orion.example/library/orion \
  --create-namespace --namespace orion-system \
  --set clusterName="my-cluster" \
  --set authorization.clientID="CLIENT_ID" \
  --set authorization.clientSecret="CLIENT_SECRET" \
  --set applier.enabled=false
```

## Next steps

**Review the optimization settings**

Review the settings Orion uses to generate recommendations for a workload: In the Orion UI, click **Workloads**; then click a workload name and go to the **Settings** tab.

By default, these settings apply to all workloads in the cluster.

**Change the optimization settings (optional)**

If you've enabled automatic deployment and are satisfied with the default cluster settings, no further configuration is needed.

If you want to adjust the defaults, we recommend setting them at the cluster level, then defining exceptions at the namespace and workload levels. See [Configure optimization settings](configure.md) for details.

## Helm chart and troubleshooting

To learn more about the Orion Helm chart and troubleshooting steps:

```shell
helm show readme oci://registry.orion.example/library/orion
```

To see all available configuration parameters:

```shell
helm show values oci://registry.orion.example/library/orion
```

[rfc-1123]: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#dns-label-names
[concepts-applier]: /docs/concepts/#applier
