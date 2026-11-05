---
updatedAt: 2026-10-07T12:52:11.000Z
---

# In-Place Pod Resizing

Atlas Workload Autoscaler supports Kubernetes in-place resource resizing for clusters running Kubernetes v1.33+, allowing resource adjustments that minimize pod disruption. In-place resizing attempts to apply resource changes without restarting pods, significantly reducing optimization overhead compared to traditional pod recreation. This feature allows workloads to have their CPU and memory requests dynamically adjusted while maintaining continuous operation in many scenarios.

> 📘 Kubernetes v1.33+ Feature
>
> In-place pod resizing requires Kubernetes v1.33 or later with the `InPlacePodVerticalScaling` feature gate enabled (enabled by default in v1.33+).

# Key capabilities

* Resource requests are adjusted with minimal disruption—often without pod restarts—maintaining application availability during optimization
* When in-place resizing cannot be applied immediately, the system gracefully falls back to the configured apply mode (immediate or deferred pod restart)
* Automatic protection against memory limit reduction during downscaling to prevent potential application issues, following Kubernetes best practices
* In-place resizing complements existing immediate and deferred scaling modes with reduced disruption
* Leverages the standard Kubernetes in-place resizing API for compatibility across different cluster configurations

This is particularly valuable for production workloads where minimizing disruption is critical, while still achieving the cost and performance benefits of continuous resource optimization.

Learn more about the benefits of in-place rightsizing on our blog [How In-Place Pod Resizing Works in Kubernetes and Why Atlas Makes It Better](https://atlas.example/blog/in-place-pod-resizing-with-atlas/).

# Prerequisites

For in-place resizing to work with Workload Autoscaler:

* **Kubernetes version**: v1.33 or later
* **Operating system**: Linux nodes only (Windows pods not supported)
* **Quality of Service**: Pods must maintain their original QoS class (Guaranteed, Burstable, or BestEffort)
* **Workload Autoscaler version**: v0.53.0 (Helm chart v0.1.122) or later for full support

# How it works

When Workload Autoscaler generates a recommendation for a pod that supports in-place resizing:

1. **Compatibility check**: The system verifies that the pod, containers, and cluster support in-place resizing
2. **In-place attempt**: If compatible, attempts to adjust CPU and memory requests directly on the running pod
3. **Restart scenarios**: Container restarts may occur when:
   * The container's resize policy requires a restart for the resource being changed (typically memory)
   * The resize enters a pending state due to resource constraints
   * The recommendation includes changes that cannot be applied in-place (such as removing CPU limits)
4. **Graceful fallback**: If in-place resizing fails or isn't supported, the system falls back to the configured apply mode (immediate or deferred pod restart)
5. **Memory protection**: Memory limits are not decreased during in-place resizing to prevent application issues

<Embed url="https://www.youtube.com/embed/7lndbsa07SU" href="https://www.youtube.com/embed/7lndbsa07SU" typeOfEmbed="iframe" height="360px" width="640px" iframe="true" />

## Pod statuses during resizing

During in-place resizing operations, you may observe certain Pod statuses that appear concerning but are part of normal Kubernetes behavior. These statuses occur due to the asynchronous nature of in-place resizing and standard Kubernetes scheduling.

### Normal operating statuses

#### `PodResizePending`

The `PodResizePending` status indicates that a resize request has been submitted to Kubernetes but hasn't yet been applied. This is expected behavior for in-place resizing, which operates asynchronously.

When Workload Autoscaler submits a resize request:

1. The resize request enters a pending state while Kubernetes evaluates Node capacity
2. If the Node has sufficient resources, the resize completes and the Pod continues running
3. If the resize remains pending due to resource constraints, Workload Autoscaler takes action based on your configured [mode for applying changes](workload-autoscaling-configuration.md#/recommendation-apply-type):
   * **Immediate mode**: Triggers Pod eviction to complete the resize through a restart
   * **Deferred mode**: Leaves the pending resize until the next natural Pod restart

A resize can also stay stuck when the node lacks the unallocated resources for it. Stuck pod resize reconciliation, an opt-in setting on your [node templates](node-templates.md), resolves this. Atlas detects the stuck resize and makes room for it. It provisions a new node and evicts other pods onto it, using live migration when possible (see [Container Live Migration](clm-overview.md)). The resize can then apply. Atlas evicts the stuck pod itself when nothing else can make room.

> ⚠️ Warning
>
> Do not enable both this feature and the Kubernetes `InPlacePodVerticalScalingSchedulerPreemption` alpha gate (v1.37+). Both react to the same stuck resize and the Kubernetes gate frees space by evicting lower-priority pods, so running them together causes additional evictions.

Workload Autoscaler doesn't check Node capacity before submitting resize requests. Instead, it submits the resize and handles the result. Node capacity can change between submission and application, so this approach allows resizes to succeed when resources become available.

#### `OutOfCPU` and `OutOfMemory`

`OutOfCPU` and `OutOfMemory` statuses occur during Pod admission when `kubelet` rejects a Pod due to insufficient Node resources. These statuses indicate a timing issue during Kubernetes scheduling, not a problem with your workload configuration.

These statuses typically appear when:

* Multiple Pods scale up simultaneously
* The Kubernetes scheduler makes parallel scheduling decisions
* Those scheduling decisions conflict, resulting in more Pods assigned to a Node than can fit

This is a normal race condition in Kubernetes scheduling. The scheduler includes optimizations for handling large Pod bursts, but occasionally these parallel decisions result in temporary admission failures.

**Important:** These are `Failed` Pod statuses that Kubernetes automatically handles through retry mechanisms. Provided your workloads have appropriate <Anchor label="Pod Disruption Budgets (PDBs)" target="_blank" href="https://kubernetes.io/docs/tasks/run-application/configure-pdb/">Pod Disruption Budgets (PDBs)</Anchor> configured, these transient failures won't impact application availability. Kubernetes will reschedule the affected Pods to suitable Nodes.

> 📘 Note
>
> `OutOfCPU` and `OutofMemory` occur during admission for newly scheduled Pods, not for Pods already running on a Node during in-place resize operations.

For more information on Pod statuses during resize operations, see <Anchor label="Kubernetes documentation" target="_blank" href="https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/#pod-resize-status">Kubernetes documentation</Anchor>.

# Configuration

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

In-place resizing is automatically enabled for eligible workloads when running on Kubernetes v1.33+ clusters. No additional configuration is required at the workload level.

For clusters that support in-place resizing, you can control this behavior through Helm values:

> 📘 Configuration-only changes
>
> These commands upgrade all components in the umbrella chart, not just Workload Autoscaler. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.

`inPlaceResizeEnabled` defaults to `true`; set it to `false` to disable in-place resizing entirely. When in-place resizing is enabled, `inPlaceResizeDeferredEnabled` also defaults to `true`; set it to `false` to keep in-place resizing active for immediate mode while disabling it for deferred mode.

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
# Disable in-place resizing entirely
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.inPlaceResizeEnabled=false

# Disable deferred in-place resizing only (keep in-place resizing active for immediate mode)
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.inPlaceResizeDeferredEnabled=false
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
# Disable in-place resizing entirely
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.inPlaceResizeEnabled=false

# Disable deferred in-place resizing only (keep in-place resizing active for immediate mode)
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.inPlaceResizeDeferredEnabled=false
```

  </Tab>
</Tabs>

## Enable stuck pod resize reconciliation

Atlas makes room for resizes stuck because the node lacks enough unallocated resources. It is off by default. Enable it when creating or updating a node template:

<Tabs>
  <Tab title="API">

Set `stuckPodResizeReconciliation.enabled` through the <Anchor label="Node Templates API" target="_blank" href="ref:nodetemplatesapi_updatenodetemplate">Node Templates API</Anchor>:

```json
{
  "stuckPodResizeReconciliation": {
    "enabled": true
  }
}
```

  </Tab>
  <Tab title="Terraform">

Add the `stuck_pod_resize_reconciliation` block to your node template in the <Anchor label="Atlas Terraform provider" target="_blank" href="https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/node_template">Atlas Terraform provider</Anchor>:

```toml Terraform
resource "atlas_node_template" "example" {
  # ... other configuration

  stuck_pod_resize_reconciliation {
    enabled = true
  }
}
```

  </Tab>
</Tabs>

# Observability considerations

When in-place resizing is active, the `autoscaling.atlas.example/recommendation-applied-at` annotation is updated on running pods each time a recommendation is applied — without a pod restart. The Workload Autoscaler generates recommendations approximately every 30 minutes, so this annotation can update at that frequency for any managed workload.

Observability tools that index pod annotations as time-series labels — such as Prometheus, Grafana, Dash0, and similar collectors — may experience:

* **High label cardinality**: each annotation value change is treated as a new label set
* **Broken time-series continuity**: existing series are abandoned and new ones started on each update

To mitigate this, choose one of the following approaches:

* **Exclude the annotation from label collection**: configure your collector to drop `autoscaling.atlas.example/recommendation-applied-at` (and optionally `autoscaling.atlas.example/vertical-recommendation-hash`) from the annotations it indexes. This is the lowest-disruption option.
* **Disable in-place resizing**: set `inPlaceResizeEnabled=false` as shown above. Recommendations will be applied through pod restarts only, and the annotation will no longer change on running pods.

# Limitations

In-place resizing has several limitations that Workload Autoscaler respects:

* **Resource types**: Only CPU and memory can be resized in-place
* **Memory limits**: Cannot be decreased without container restart
* **QoS class preservation**: Original QoS class must be maintained
* **Container types**: Non-restartable init containers and ephemeral containers cannot be resized in-place
* **Swap memory**: Pods using swap cannot resize memory without a restart
* **Static policies**: Pods with static CPU/memory manager policies cannot be resized in-place

For detailed information about Kubernetes in-place resizing limitations, see the [official Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/#limitations).

## When restarts occur

While in-place resizing reduces disruption, container restarts can still occur in the following scenarios:

* **Resize policy configuration**: Containers with `RestartContainer` resize policy for specific resources will restart when those resources change. By default, Kubernetes applies `NotRequired` for both CPU and memory, but certain configurations or resource changes may require restarts.
* **Resource constraints**: When the node cannot immediately accommodate the resize request, the request enters a `Pending` state. In immediate mode, if the resize remains stuck, Workload Autoscaler falls back to evicting the pod so the controller can reschedule it with the updated resources. In deferred mode, no eviction occurs — the pod retains its current resources until a natural restart opportunity arises. To make room when a resize is stuck because the node lacks enough unallocated resources, see [stuck pod resize reconciliation](#podresizepending).
* **Incompatible changes**: Some resource changes cannot be applied in-place due to Kubernetes limitations. For example, removing or significantly modifying CPU limits during the first recommendation application may require a restart.

These restart scenarios are handled gracefully by Workload Autoscaler, which respects Pod Disruption Budgets and follows the configured apply mode's rollout strategy.