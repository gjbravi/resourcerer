---
updatedAt: 2026-09-22T11:35:10.000Z
---

# Pod Pinner

Pod Pinner coordinates between the Atlas Autoscaler and the Kubernetes scheduler to ensure the optimal pod placement in your Kubernetes cluster. Without Pod Pinner, the Kubernetes scheduler might place pods on unintended nodes, resulting in suboptimal resource utilization and higher costs.

Pod Pinner integrates Atlas Autoscaler decisions into your cluster, allowing it to override the Kubernetes scheduler's decisions when necessary. For example, Pod Pinner assigns pending pods to a placeholder node during node rebalancing, ensuring they are directed to the correct node once it’s ready.

Installing Pod Pinner can directly enhance cluster savings. Pod Pinner is a Atlas in-cluster component similar to the Atlas agent, cluster controller, and others. There is no risk of downtime if Pod Pinner fails; the default Kubernetes scheduler will automatically take over pod placement decisions for rebalancing and autoscaling.

# How Pod Pinner works

Pod Pinner works alongside the Atlas Autoscaler to ensure pods are scheduled onto the specific nodes chosen by the Autoscaler. Here's how this process works:

1. When the Atlas Autoscaler determines new nodes are needed, it sends scheduling instructions to Pod Pinner through a continuous data stream.

2. Pod Pinner creates a placeholder Node resource in Kubernetes using the API. This placeholder reserves the space for the incoming node before it physically joins the cluster, allowing Pod Pinner to coordinate scheduling decisions ahead of time.

3. For pods that need scheduling, Pod Pinner uses the Kubernetes `/pods/binding` API to explicitly assign them to specific nodes by setting `spec.nodeName`. This direct binding bypasses the default Kubernetes scheduling process.

4. When the actual node joins the cluster, it automatically connects to its corresponding placeholder node, ensuring continuity in the scheduling process.

This coordination prevents scheduling conflicts between Atlas and Kubernetes by ensuring pods land on their intended nodes rather than letting the Kubernetes scheduler place them based on its own scoring mechanisms. This is especially important because the Kubernetes scheduler isn't aware of nodes that are about to join the cluster and may make placement decisions that are different from what the Atlas Autoscaler intended.

# Limitations

> 🚧 Mutating admission webhooks
>
> Pod Pinner is not compatible with mutating admission webhooks. It pins pods based on the original pod spec before any webhook mutations are applied, which can cause incorrect scheduling or resource oversubscription.

## Scheduling constraint mutations

Pod Pinner is not compatible with mutating admission webhooks that modify scheduling-related pod specs. This includes Atlas [Pod Mutations](pod-mutations-overview.md), Atlas Spot-webhook, and Kyverno.

The Autoscaler and Rebalancer make placement decisions based on the original pod spec. If a webhook adds `nodeSelector`, tolerations, or affinity rules after pinning occurs, those constraints are ignored because `nodeName` takes precedence over `nodeSelector`. For example, if a mutation adds `scheduling.atlas.example/spot: "true"` to a pod that was already pinned to an on-demand node, the pod will fail with a `NodeAffinity` mismatch.

If you use mutating webhooks that affect scheduling, disable Pod Pinner or define scheduling constraints directly in your workload specs. We are working on improving compatibility between Pod Pinner and Pod Mutations.

## Runtime sidecar injection

Pod Pinner calculates node capacity based on the original pod spec at pin time. If a mutating webhook, such as Istio's or Linkerd's sidecar injector, adds containers after pinning occurs, their resource requests are not accounted for. During large-scale or burst scheduling events, this can cause nodes to become oversubscribed, resulting in pods failing with `OutOfcpu` or `OutOfmemory` status.

If your cluster uses runtime sidecar injection, disable Pod Pinner to avoid these scheduling failures.

## Custom resources

Pod Pinner cannot pin pods that use custom resources (such as GPUs and ENIs). This is because custom resources like GPUs are initialized after a node joins the cluster. When a node first boots up, custom resources (like GPUs) are not immediately available - they require initialization by their respective drivers.

# Installation and version upgrade

When you onboard a cluster using the umbrella Helm chart in `node-autoscaler` or `full` mode, the `atlas-pod-pinner` Deployment is created in the `atlas-agent` namespace, but with **0 replicas** by default. No Pod Pinner pods run until you explicitly activate it. See [Enabling/Disabling Pod Pinner](#enablingdisabling-pod-pinner).

1. Check whether the `atlas-pod-pinner` deployment exists and how many replicas are running:

```shell
$ kubectl get deployments.apps -n atlas-agent
NAME                        READY   UP-TO-DATE   AVAILABLE   AGE
atlas-agent                1/1     1            1           15m
atlas-agent-cpvpa          1/1     1            1           15m
atlas-cluster-controller   2/2     2            2           15m
atlas-evictor              0/0     0            0           15m
atlas-sentinel               1/1     1            1           15m
atlas-pod-pinner           0/0     0            0           15m
```

If the replica count shows `0/0`, Pod Pinner is deployed but not active.

## Prerequisites

Before installing Pod Pinner, ensure your cluster meets the following requirements:

* Kubernetes version 1.27 or later

## Helm

<Callout icon="📘" theme="info">
  ### Which upgrade method to use

  The Helm commands on this page use the **umbrella chart** (`atlas-helm/atlas`) by default. If you need to use a different method:

  - **[atlasctl](connect-with-atlasctl.md)**: To upgrade all Atlas components at once without managing Helm flags:
    ```shell
    atlasctl atlasware upgrade
    ```
    This preserves your existing configuration. See the [atlasctl documentation](connect-with-atlasctl.md) for installation and authentication instructions.
  - **Individual charts**: If you installed each component as a separate Helm release (e.g., for ArgoCD or custom GitOps), replace the release name and chart reference with the component-specific ones (e.g., `atlas-workload-autoscaler` and `atlas-helm/atlas-workload-autoscaler`) and remove the `autoscaler.atlas-workload-autoscaler.` value prefix.

  Not sure which method you used? Run `helm list -n atlas-agent`. A single release named `atlas` means umbrella chart; separate releases like `atlas-workload-autoscaler` mean individual charts.
</Callout>

<br />

### Option 1: Atlas-Managed (default)

By default, Atlas manages Pod Pinner: the Atlas platform syncs Pod Pinner configuration from the console, and Pod Pinner's `charts.atlas.example/managed=true` label keeps it under Atlas management.

1. Check the currently installed Pod Pinner chart version. If it's >= `1.0.0`, an upgrade is not needed.\
   You can check the version with the following command:

```shell
$ helm list -n atlas-agent --filter atlas-pod-pinner
NAME             	NAMESPACE   	REVISION	UPDATED                                	STATUS  	CHART                  	APP VERSION
atlas-pod-pinner	atlas-agent	11      	2024-09-26 11:40:00.245427517 +0000 UTC	deployed	atlas-pod-pinner-1.0.2	v1.0.0
```

2. If the version is < `1.0.0` run the following commands to install or upgrade Pod Pinner to the latest version.

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

When you enable Pod Pinner from the Atlas console, it runs with 2 replicas and is managed by Atlas, as indicated by the `charts.atlas.example/managed=true` label applied to the pods of the `atlas-pod-pinner` deployment.

### Option 2: Self-Managed

To control the Pod Pinner version yourself:

> 📘 Configuration-only changes
>
> This command upgrades all components in the umbrella chart, not just Pod Pinner. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

```shell
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-pod-pinner.managedByATLAS=false
```

This prevents Atlas from automatically managing and upgrading Pod Pinner.

You can also scale the deployment manually, but in that case the deployment is not automatically synced from the Atlas console:

```shell
kubectl scale deploy atlas-pod-pinner -n atlas-agent --replicas=2
```

Upgrading using Pod Pinner when in Self-Managed mode:

If the version is < `1.0.0` run the following commands to upgrade Pod Pinner to the latest version.

```shell
helm repo add atlas-helm https://atlas.github.io/helm-charts
helm repo update
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-pod-pinner.managedByATLAS=false
```

## Re-running Onboarding Script

You can also install Pod Pinner by re-running the phase 2 onboarding script. For more information, see the [cluster onboarding documentation](legacy-onboarding-script.md#actions-the-onboarding-script-performs-1).

## Terraform Users

For Terraform users, you can manage Pod Pinner installation and configuration through your Terraform scripts. This allows for version control and infrastructure-as-code management of Pod Pinner settings.

# Enabling/Disabling Pod Pinner

Pod Pinner deploys with 0 replicas by default. To activate it, use the Atlas console or manually scale the deployment.

When enabled, the Atlas platform syncs Pod Pinner configuration from the console and scales the deployment to 2 replicas. When disabled, the deployment is scaled back down to 0 replicas and configuration sync stops.

### Autoscaler Settings UI

To enable/disable Pod Pinner:

1. Navigate to Autoscaler settings in the Atlas console.
2. You'll find the Pod Pinner option under the "Unscheduled pods policy" section.
3. Check/uncheck the "Enable pod pinning" box to activate/deactivate the Pod Pinner. Disabling Pod Pinner will scale down the deployment to 0 replicas and stop syncing its configuration from the Atlas console.

![](https://files.readme.io/2a6af2207e75f1af27ac6e3ebc8b68821c917b57e07a7874852827a4e6af48a2-image.png)

> 📘 Note
>
> When Pod Pinner is disabled through the console and the `charts.atlas.example/managed=true` label is present, Atlas will scale down the deployment to 0 replicas no matter what. To manually control Pod Pinner while keeping it active, use the self-managed installation option mentioned above.

# Ensuring stability

It is suggested that you keep the Pod Pinner pod as stable as possible, especially during rebalancing. You can do so by applying the same approach you are using for `atlas-agent`.

For instance, you can add the `autoscaling.atlas.example/removal-disabled: "true"` label/annotation to the pod. If the Pod Pinner pod restarts during rebalancing, the pods won't get pinned to the nodes as expected by the Rebalancer. It may result in suboptimal pod placement as the Kubernetes cluster scheduler will schedule the pods.

> 📘 Note
>
> You can scale down the `atlas-pod-pinner` deployment anytime. This will result in normal behavior and will not impact the cluster negatively other than the Kubernetes scheduler taking over pod scheduling.

# Logs

You can access logs in the Pod Pinner pod to see what decisions are being made. Here is a list of the most important log entries:

| Example                                                | Meaning                                                                                                                                                        |
| :----------------------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `node placeholder created`                             | A node placeholder has been created. The real node will use this placeholder when it joins the cluster.                                                        |
| `pod pinned`                                           | A pod has been successfully bound to a node. Such logs always appear after the node placeholder is created.                                                    |
| `node placeholder not found`                           | This log appears when Pod Pinner tries to bind a pod to a non-existing node. This may occur if  Pod Pinner fails to create the node placeholder.               |
| `pinning pod`                                          | This log occurs when the Pod Pinner's webhook intercepts a pod creation and binds it to a node. This happens during rebalancing.                               |
| `node placeholder deleted`                             | A node placeholder has been deleted. This happens when a node fails to be created in the cloud, and Pod Pinner must clean up the placeholder that was created. |
| `failed streaming pod pinning actions, restarting...`  | The connection between the Pod Pinner pod and Atlas has been reset. This is expected to happen occasionally and will not negatively impact your cluster.     |
| `http: TLS handshake error from 10.0.1.135:48024: EOF` | This log appears as part of the certificate rotation performed by the webhook. This is a non-issue log and will not negatively impact the cluster.             |

# Troubleshooting

## Failed pod status reason: `OutOf{resource}`

`OutOfcpu`, `OutOfmemory`, `OutOf{resource}` pod statuses happen when the scheduler schedules a pod on a node, but the `kubelet` rejects it due to a lack of some sort of resource. These are `Failed` pods that Atlas and the Kubernetes `control-plane` know how to ignore.

This happens when many pods are upscaled at the same time. The scheduler has various optimizations to deal with large bursts of pods, so it makes scheduling decisions in parallel. Sometimes, those decisions conflict, resulting in pods scheduled on nodes where they don't fit. This happens especially in GKE. If you see this status, don't be afraid. The `control-plane` will eventually clean those pods up after a few days.

Pods might get this status when the Kubernetes scheduler takes over scheduling decisions due to a blip in Pod Pinner's availability. However, this does not negatively impact the cluster as Kubernetes recreates the pods.

## Failed pinning reason: pod requires unsupported resources

Pod Pinner `>=1.6.3` checks resources requested by pods. If you see this error, it means the pod requests (or limits) some resource that is not yet available on the node. The Kubernetes scheduler will still schedule these pods when there is a suitable node available. This is common for delayed initialization resources like GPUs. If scheduled too early on a node, such pods often crash-loop until these resources become available.

If crash-looping is desired (as opposed to not getting pinned), you can annotate the pod with `autoscaling.atlas.example/pin-ignore-missing-resources=true` to bypass this validation. This annotation is supported by Pod Pinner `>=1.7.0`.