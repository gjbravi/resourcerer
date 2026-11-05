---
updatedAt: 2026-10-12T09:17:12.000Z
---

# Available settings

This page provides a reference for configuration options available in the Atlas console when optimizing workloads with Atlas's Workload Autoscaler. These settings control how the autoscaler analyzes resource usage patterns, generates resource request recommendations, and applies optimizations to your workloads.

For comprehensive, advanced configuration options, see our [Annotations reference](workload-autoscaler-annotations-reference.md) to learn how to configure and manage workload optimization on Atlas via Kubernetes annotations.

# Configuration scope and hierarchy

The settings documented here can be applied at multiple levels, creating a flexible configuration hierarchy.

## Policy-level configuration

Settings defined in scaling policies serve as defaults for all workloads assigned to that policy. This approach enables consistent optimization strategies across groups of similar workloads while reducing individual configuration overhead. Atlas strongly recommends leveraging scaling policies with automated assignment rules to reduce the amount of manual work needed to manage and optimize your workloads.

## Workload-level configuration

Individual workloads can override policy-level settings using [Kubernetes annotations](workload-autoscaler-annotations-reference.md) when specific requirements differ from the policy defaults. Annotation-based settings take precedence over policy-level configurations.

> 📘 Note
>
> The Atlas console does not support editing per-workload vertical scaling settings. You can still toggle optimization on or off for individual workloads via the console, but all other workload-level overrides must be configured through annotations.

## Configuration hierarchy

When multiple configuration sources are present, they are applied in the following order of precedence:

1. **Workload-level annotations** (highest priority)
2. **Policy-level settings** (lowest priority)

This hierarchy allows for centralized policy management while maintaining flexibility for specific workload requirements through annotations.

# Configuration methods

You can configure these settings through multiple interfaces, each suited for different use cases and workflows.

## Atlas Console

The web-based interface provides user-friendly controls for configuring scaling policies and toggling optimization on or off for individual workloads. This method is ideal for:

* Interactive policy creation and management
* Enabling or disabling optimization on individual workloads

See our guide on [creating scaling policies](woop-scaling-policies-create.md) or the [available settings reference](#settings-reference) below for instructions on policy settings available in the console.

## API integration

The [Workload Optimization API](https://docs.atlas.example/reference/workloadoptimizationapi_createworkloadscalingpolicy) enables programmatic configuration of scaling policies and workload settings.

### Terraform provider

The [Atlas Terraform provider](https://github.com/atlas/terraform-provider-atlas/blob/master/docs/resources/workload_scaling_policy.md) allows you to define optimization settings as infrastructure code.

## Kubernetes annotations

Workload-specific settings can be applied directly to Kubernetes resources using annotations. This method provides:

* Fine-grained control at the workload level
* Integration with existing Kubernetes manifests

For detailed information about annotation syntax and examples, see the [Configuration via annotations](workload-autoscaling-configuration.md#configuration-via-annotations) section in the Workload Autoscaling Configuration documentation.

# Settings reference

The following sections provide detailed information about each available configuration option, including their purpose, impact on workload optimization, and configuration examples across different interfaces.

You can configure the following settings in your custom scaling policies. To override these settings for individual workloads, use [annotations](workload-autoscaler-annotations-reference.md).

## Automatically optimize workloads

Specify whether resource request recommendations should be automatically applied to all workloads associated with the scaling policy. This feature enables automation only when enough data is available to make informed recommendations.

![](https://files.readme.io/815f41b03f6f89824d4ee2df0c2f20a9238b8e5586c7258b030ea910960ad632-image.png)

## Resource-specific optimization

When configuring vertical scaling, you can enable or disable CPU and memory optimization independently while still receiving recommendations for both resources. Even when optimization is disabled for a resource, Workload Autoscaler continues to generate recommendations but won't apply them automatically. This setting can be configured at the **vertical policy level**. To override it for individual workloads, use [annotations](workload-autoscaler-annotations-reference.md).

<Image src="https://files.readme.io/9b42788baf7835445fcfc86022d6ab9e6805c8f81561e67c9ef613b071b470c2-image.png" alt="Selective resource optimization controls in the vertical scaling policy settings" align="center" caption="Selective resource optimization controls in the vertical scaling policy settings" />

> 📘 Note
>
> At least one resource type must remain enabled – you cannot disable both CPU and memory optimization simultaneously.

### Version requirements

The minimum required `workload-autoscaler` component version to use this feature is `v0.23.1`.

### Configuration options

At the policy level, you can configure resource-specific optimization through:

* The Atlas console UI using the resource checkboxes
* The Workload Autoscaler API or the Terraform module

To override this setting for individual workloads, use annotations:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    memory:
      optimization: off
```

For detailed reference information on Workload Autoscaler annotations, see [Configuration via annotations](workload-autoscaler-annotations-reference.md).

### Container exclusions

You can exclude specific containers from automatic optimization within a scaling policy. Excluded containers retain their current resource settings and are not scaled by the Workload Autoscaler. Recommendations are still generated and visible in the event log and the workload detail view.

This setting is configured under the **Selective optimization** section of a vertical scaling policy, alongside the resource selection (CPU/Memory) described above.

<Image src="https://files.readme.io/48cc165e25e3db733b61b86897a07e959bbcef56bd0683bc4eefe91a36cfb064-image.png" align="center" caption="Container exclusion settings under Selective optimization" />

> 📘 Note
>
> `excludedContainers` targets containers explicitly defined in your pod spec, such as application containers or native Kubernetes sidecar containers. For information on dynamically injected containers, see [Sidecar container support](workload-autoscaling-configuration.md#sidecar-container-support).

#### Configuring container exclusions

To exclude containers from optimization in the Atlas console:

1. Navigate to **Workload Autoscaler** → **Scaling policies**.
2. Edit an existing policy or create a new one.
3. Under the **Vertical Rightsizing** tab, locate the **Selective optimization** section.
4. In the **Container** subsection, enable **Selectively exclude containers**.
5. Enter the exact container names to exclude, separated by commas.
6. Save the policy.

Container exclusions can also be configured via annotations:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    excludedContainers:
      - istio-proxy
      - logging-agent
```

For the full annotation reference, see [vertical.excludedContainers](workload-autoscaler-annotations-reference.md#verticalexcludedcontainers).

## When to apply changes

| Setting Name | Description                                                                         | Possible Values         | Default Value |
| ------------ | ----------------------------------------------------------------------------------- | ----------------------- | ------------- |
| Apply type   | Controls how and when the Workload Autoscaler applies recommendations to workloads. | `immediate`, `deferred` | `immediate`   |

### Immediate mode

When set to `immediate`, the Workload Autoscaler proactively implements resource optimization. The system monitors your workloads and applies new resource recommendations when they exceed the configured thresholds. This approach prioritizes rapid resource optimization, automatically triggering pod restarts to implement the new allocations.

* Apply recommendations as soon as they exceed the configured thresholds
* Trigger pod restarts to implement the new resource allocations immediately

For a detailed explanation of how immediate mode applies recommendations (eviction mechanism, PDB handling, single-replica behavior), see [How immediate mode applies recommendations](workload-autoscaling-configuration.md#how-immediate-mode-applies-recommendations).

#### Rollout behavior

When using immediate mode, you can control how Workload Autoscaler rolls out recommendations to your pods. Under **Advanced settings**, the **Rollout behavior** section groups all three controls: **Zero-downtime updates** for single-replica workloads, **Evict pods one by one** to minimize disruption during resource adjustments, and **Eviction delay** to pace how often Workload Autoscaler evaluates the next eviction.

![](https://files.readme.io/ada5106ec73bbb359423ebd15b34762c3bd0c2b1478e3c197049795530226b00-image.png)

##### Zero-downtime updates

The Workload Autoscaler provides a zero-downtime update mechanism for single-replica workloads. This feature enables resource adjustments without service interruptions by temporarily scaling to two replicas during updates.

**How it works**

When enabled, this setting:

1. Temporarily scales single-replica deployments to two replicas
2. Applies resource adjustments to the new pods
3. Waits for the new pod to become healthy and ready
4. Scales back to one replica, removing the original pod

Thus, it maintains continuous service availability throughout the process. This setting is especially valuable for applications where even brief interruptions could cause failed requests or other issues.

**Configuring zero-downtime updates**

You can enable this feature in your vertical scaling policy settings:

1. Navigate to the **Workload Autoscaler** → **Scaling policies** section
2. Edit an existing policy or create a new one
3. Under **Advanced settings**, locate the **Rollout behavior** section
4. Check **Zero-downtime updates**

**Prerequisites**

For zero-downtime updates to work effectively, your workloads must meet these requirements:

* This feature is only applicable to Deployment resources that support running as multi-replica
* Running with a single replica (replica count = 1)
* Deployment's rollout strategy allows for downtime
* It works with the immediate apply type and is not needed for deferred mode
* Must not use PersistentVolumeClaims (PVCs)
* It requires `workload-autoscaler` component version v0.35.3 or higher

For workloads where brief interruptions are unacceptable, this setting provides a way to achieve continuous availability without permanently increasing replica counts and associated costs.

##### Evict pods one by one

When **Evict pods one by one** is enabled, immediate mode applies recommendations as follows:

1. **Sequential processing**: The system applies the recommendation to a single pod at a time
2. **Health verification**: Waits for each pod to become healthy and ready before proceeding
3. **Gradual rollout**: Moves to the next pod only after the previous one is successfully running
4. **Automatic recovery**: If a pod fails to start, the rollout pauses until the workload recovers

This ensures gradual rollouts and maintains service availability throughout the resource adjustment process.

You can enable this feature through the Atlas console:

1. Navigate to **Workload Autoscaler** → **Scaling policies**
2. Select or create a policy with **When to apply changes** set to **Immediately**
3. Under **Advanced settings**, locate **Rollout behavior**
4. Check **Evict pods one by one**
5. Save the policy

You can also configure this setting using annotations:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        optimization: on
        applyType: immediate
      rolloutBehavior:
        type: NoDisruption
        preferOneByOne: true
        delaySeconds: 120
spec:
  # deployment spec...
```

For more information on configuring workloads via annotations, see <Anchor target="_blank" href="doc:workload-autoscaler-annotations-reference">Annotations reference</Anchor>.

> 📘 Note
>
> The one-by-one restart feature is currently opt-in. Based on customer feedback and continued improvements, it may become the default behavior in system policies in future releases.

##### Eviction delay

**Eviction delay** sets a minimum wait time between evictions while Workload Autoscaler rolls out a recommendation. The delay applies whether or not **Evict pods one by one** is enabled — it controls how often Workload Autoscaler evaluates the next eviction during any rollout.

You can enable this feature through the Atlas console:

1. Navigate to **Workload Autoscaler** → **Scaling policies**
2. Select or create a policy with **When to apply changes** set to **Immediately**
3. Under **Advanced settings**, locate **Rollout behavior**
4. Check **Eviction delay** and select a preset (10 sec, 60 sec, 5 min, 60 min) or enter a custom value

You can also configure this setting using the `rolloutBehavior.delaySeconds` annotation. See <Anchor target="_blank" href="doc:workload-autoscaler-annotations-reference">Annotations reference</Anchor> for details.

##### Supported workload types

The one-by-one restart feature works with the following workload types:

* Deployments
* StatefulSets
* ReplicaSets
* ArgoCD Rollouts

**Not supported**: Bare pods, CronJobs, Jobs, and other non-replicated workloads.

##### Version requirements

To use the one-by-one restart feature, ensure you have the following minimum versions:

| Component                  | Minimum Version |
| :------------------------- | :-------------- |
| atlas-workload-autoscaler | v0.57.0         |
| Helm chart                 | 0.1.127         |

##### Considerations and limitations

When using one-by-one restart, be aware of these considerations:

**Rollout duration**: For workloads with many replicas (100+), the gradual rollout process may take longer depending on pod startup time. This is expected behavior as each pod must become healthy before the next one is processed.

**Mid-rollout recommendations**: If a new recommendation is generated while a rollout is in progress, it takes precedence. For example, if half the pods have been updated with an immediate recommendation and a deferred recommendation is then issued, the rollout will stop. The new deferred recommendation will be applied to the remaining pods only during their natural restarts.

**Failed pod scenarios**: If a pod fails to start during the rollout (for example, due to image pull errors or application crashes), the entire rollout pauses until the workload recovers. This prevents further disruption but requires manual intervention to resolve the underlying issue.

**OOM kills and surge handling**: When enabled, OOM kill responses and surge-triggered adjustments also follow the one-by-one pattern, which may delay the system's response to these events.

### Deferred mode

When set to `deferred`, the Workload Autoscaler takes a non-disruptive approach to resource optimization. Rather than forcing pod restarts, the system stores recommendations and applies them without disrupting your workloads.

* Store recommendations without forcing pod restarts
* On Kubernetes v1.33+ clusters, attempt to apply recommendations via [in-place resizing](woop-in-place-resizing.md) without waiting for a pod restart
* When in-place resizing is unavailable or not supported, apply recommendations during the next natural pod lifecycle event (e.g., application deployments, scaling events, or node maintenance)

### Recommendation annotations in different scaling modes

When the Workload Autoscaler applies recommendations to your workloads, it adds annotations to track when and which recommendations have been applied.

| Annotation                                         | Description                                                                                                                                                   |
| -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `autoscaling.atlas.example/vertical-recommendation-hash` | A hash value representing the applied recommendation. This annotation appears on all workloads with applied recommendations, regardless of the scaling mode.  |
| `autoscaling.atlas.example/recommendation-applied-at`    | A timestamp indicating when a recommendation was applied. Updated each time a recommendation is applied — whether through a pod restart or in-place resizing. |

#### Scaling mode behavior

* **Immediate mode**: Both annotations will be present. The `recommendation-applied-at` annotation is updated each time a recommendation is applied, including when applied via in-place resizing without a pod restart.

* **Deferred mode**: The `vertical-recommendation-hash` annotation is always present. The `recommendation-applied-at` annotation is also present when in-place resizing is enabled (the default on Kubernetes v1.33+), because recommendations are applied to running pods without a restart. When in-place resizing is unavailable or disabled, `recommendation-applied-at` is not added in deferred mode — recommendations are applied at the next natural pod lifecycle event, and you can determine timing from the pod's creation timestamp.

> 📘 High-churn annotation
>
> When [in-place resizing](woop-in-place-resizing.md) is enabled (the default on Kubernetes v1.33+), `autoscaling.atlas.example/recommendation-applied-at` updates on running pods each time a recommendation is applied — potentially every 30 minutes. Observability tools that index pod annotations as time-series labels (such as Prometheus, Grafana, Dash0, and similar collectors) may experience high label cardinality or broken time-series continuity as a result. To avoid this, either configure your collector to exclude this annotation from label collection, or [disable in-place resizing](woop-in-place-resizing.md#configuration) if annotation stability is a hard requirement.

### Change Sensitivity

Change sensitivity determines when the Workload Autoscaler applies resource recommendation changes to your workloads. The sensitivity represents the minimum percentage difference between current resource requests and new recommendations that triggers an update.

Workload Autoscaler offers two sensitivity options in the UI:

* **Percentage**: A fixed threshold value that applies equally to all workloads
* **Dynamic**: An adaptive threshold that automatically adjusts based on workload size

<Image src="https://files.readme.io/f7466340e42b6a6f14b3f48d7603c7f20be75135a114274e0fb8a75bbd92fbeb-image.png" align="center" width="75% " />

#### Dynamic vs Percentage Sensitivity

##### Percentage Sensitivity

A percentage sensitivity applies the same fixed percentage to all workloads regardless of their size. For example, with a 10% sensitivity:

* A workload requesting 100m CPU will only be scaled if the new recommendation differs by at least 10m
* A workload requesting 10 CPUs will only be scaled if the new recommendation differs by at least 1 CPU

While this is straightforward, it can be less optimal for workloads of varying sizes.

##### Dynamic Sensitivity (Recommended)

The dynamic sensitivity automatically adjusts based on workload size, providing more appropriate scaling behavior:

* **For small workloads**: Uses a higher threshold percentage to prevent frequent, insignificant updates
* **For large workloads**: Uses a lower threshold percentage to enable meaningful optimizations

This helps prevent unnecessary pod restarts for small workloads while ensuring larger workloads are also efficiently optimized.

##### Recommendations

For most users, we recommend using the **Dynamic** sensitivity setting, as it provides appropriate thresholds across workloads of all sizes and requires no manual tuning or maintenance.

The **Percentage** sensitivity setting is best suited for workloads of similar sizes and behavior when they are managed by the same scaling policy.

#### Dynamic Sensitivity Simulation

The policy/workload configuration interface includes a sensitivity simulation graph that shows how the dynamic threshold changes based on workload size. This visualization helps you understand how the sensitivity percentage varies as resource requests increase.

Key elements of the simulation:

* The X-axis represents resource requests (CPU or memory)
* The Y-axis represents the threshold percentage
* The curve shows how the threshold decreases as resource size increases

You can toggle between CPU and memory views using the dropdown.

![](https://files.readme.io/e779775a2685c62b082a4257d23abe3a43539b33fba2e5afa88960cc84f710cb-image.png)

The threshold value indicator shows the exact sensitivity percentage that would apply to a specific resource request amount. For example, in the screenshot above, a request of 0.1 CPU would have a threshold of 45.45%. Meaning it would need to change by that amount to trigger a change.

#### Advanced Configuration Options

While the UI offers simplified access to dynamic sensitivity settings, power users can access additional customization options through annotations. These advanced options include:

* Custom adaptive thresholds with configurable parameters
* Independent sensitivity settings for CPU and memory
* Fine-tuning of the adaptive algorithm formula

For details on these advanced options, refer to our [Annotations reference](workload-autoscaling-configuration.md#verticalcpuapplythresholdstrategy) documentation.

## Resource overhead

Resource overhead allows you to add a buffer to the recommendations generated by the Workload Autoscaler. This buffer provides extra capacity for your workloads to handle unexpected load increases without immediately triggering scaling events.

Resource overhead is a percentage value that is added to the recommended resource requests. For example, if the recommended CPU usage for a workload is 100m and you set a CPU overhead of 10%, the final recommendation will be 110m.

<Image src="https://files.readme.io/5296cf6b56d55a05a0752a269a52b4d4801b832dfe86ba1a03d99e37a1ce4e7c-image.png" align="center" width="50% " />

### Configuring resource overhead

You can configure different overhead values for CPU and memory:

* **CPU Overhead**: Typically set between 10% (default) and 20%.
* **Memory Overhead**: Typically set between 10% (default) and 30%.

Properly configured overhead helps prevent out-of-memory (OOM) events and CPU throttling while maintaining cost efficiency.

### Limitations

* Overhead cannot exceed 250% (2.5), as extremely high values could lead to significant resource waste.
* When both Vertical and Horizontal workload scaling are enabled, memory overhead can still be configured as normal, but CPU overhead settings are ignored as the system automatically balances vertical and horizontal scaling.

## Recommendation percentile

The recommendation percentile setting determines how conservatively the Workload Autoscaler allocates resources based on the workload's historical usage patterns. It defines how close recommendations should be to observed workload resource usage, with higher percentiles leading to more generous resource allocations.

The recommendation percentile represents the statistical threshold used when analyzing workload resource usage. For example:

* A **p80** (80th percentile) setting for CPU means the recommendation will ensure resources are sufficient to handle 80% of all observed load scenarios.
* A **max** (100th percentile) memory setting means recommendations will account for the absolute highest observed memory usage.

<Image src="https://files.readme.io/461e992e2411b2be5154d00299a40cd77102b4d7b142f6826e6b30935f1d1559-image.png" align="center" width="50% " />

#### Configuring percentiles

You can configure different percentile values for CPU and memory independently:

* **CPU Percentile**: Typically set at p80 (default), can range from p50 to max
* **Memory Percentile**: Typically set at max (default) or p99, rarely lower

The recommendation is calculated using the average target percentile across all pods spanning the recommendation period. When you set the percentile to "max" (100%), the system will use the maximum observed value over the period instead of the average across pods.

## Workload resource limits

Workload resource limits allow you to configure how container resource limits are managed relative to the resource requests that Workload Autoscaler optimizes. This feature provides fine-grained control over the relationship between requests and limits, ensuring your workloads have the right balance of resource guarantees and constraints.

In Kubernetes, each container can specify both resource requests (guaranteed resources) and resource limits (maximum allowed resources). The Workload Autoscaler primarily optimizes resource requests based on actual usage patterns, but the workload resource limits setting determines how container limits are handled during this optimization.

![](https://files.readme.io/ecb0c93e959a126d76e57ec0dc70736541f5ce25d5832c832158ee793877ed73-image.png)

### Configuration options

You can configure resource limits separately for CPU and memory.

#### CPU limit options

* **Remove limits**: Removes any existing resource limits from containers, including those specified in your workload manifest.
* **Keep existing resource limits**: Preserves the existing limits as defined in the workload manifest without modification.
* **Keep original limit-to-request ratio**: Scales limits proportionally with recommended requests, preserving the original limit-to-request ratio from the workload manifest. For example, if the original limit is 2× the request, the updated limit stays 2× the new request. If the workload has no existing CPU limit, this option falls back to the default behavior (no limit).
* **Semi-automatic**: Sets CPU limits as a multiple of requests when workloads already have CPU limits defined and those existing limits are lower than the calculated result. Otherwise, limits remain unchanged.
* **Custom multiplier**: Sets CPU limits as a multiple of requests for all workloads (e.g., `2.0` means limits = requests × 2).

#### Memory limit options

* **Automatic** (default): Automatically sets memory limits to 1.5x requests (with a minimum of 128MiB) when limits are lower than this value. If existing limits are higher, they remain unchanged.
* **Remove limits**: Removes any existing resource limits from containers, including those specified in your workload manifest.
* **Keep existing resource limits**: Preserves existing resource limits as defined in your workload manifest without modification.
* **Keep original limit-to-request ratio**: Scales limits proportionally with recommended requests, preserving the original limit-to-request ratio from the workload manifest. For example, if the original limit is 2× the request, the updated limit stays 2× the new request. If the workload has no existing memory limit, a default 1.5× multiplier is applied to the recommended request. The 128MiB minimum memory limit applies.
* **Custom multiplier**: Set memory limits as a multiple of the requests (e.g., `2.0` means limits = requests × 2).

> 📘 Configuring limits via Terraform or API
>
> **Semi-automatic** and **Automatic** are UI labels — they do not exist as named values in the API or Terraform. When configuring the same behavior through Terraform or the API, you set the `limit` block explicitly using `type`, `multiplier`, `only_if_original_exist`, and `only_if_original_lower` fields.
>
> See the [Terraform provider reference](https://registry.terraform.io/providers/atlas/atlas/latest/docs/resources/workload_scaling_policy#resource-limit-mapping) for the complete mapping between UI options and Terraform configuration.

### Impact on resource management

The relationship between requests and limits affects how Kubernetes schedules and manages your workloads:

* **Scheduling**: Pods are scheduled based on requests, not limits.
* **CPU Throttling**: Containers using more CPU than their limit will be throttled.
* **OOM Kills**: Containers exceeding memory limits will be terminated (OOM killed), as they cannot burst past the limit.

Even though Atlas does not recommend using resource limits on workloads actively managed by Workload Autoscaler, properly configured limits can prevent problems from neighboring workloads and provide predictable performance.

### Limitations

* The multiplier value must be greater than or equal to `1.0`
* When setting a custom multiplier, Workload Autoscaler will never reduce existing limits below the calculated value (limits = requests × multiplier)
* Workload Autoscaler enforces a minimum memory limit of 128MiB whenever it calculates a memory limit (Automatic, Custom multiplier, Semi-automatic, and Keep original limit-to-request ratio modes). This floor prevents unreasonably low memory allocations.

## Workload Autoscaler constraints

Workload Autoscaler constraints allow you to set minimum and maximum resource limits that the Workload Autoscaler will respect when generating and applying recommendations. These constraints act as guardrails to prevent resources from being scaled too low or too high, ensuring your workloads maintain appropriate resource boundaries.

These constraints define the allowed resource scaling range for each container managed by the autoscaler. By setting these constraints, you can:

* Prevent resources from being scaled below a functional minimum
* Limit maximum resource allocation to control costs
* Customize scaling boundaries for each container individually

<Image src="https://files.readme.io/94b9a3bd11ff6fdba893ea8e357beca053f1de62647f2ce24b2105d3d126ddcc-woop-constraints-percentage.png" align="center" width="80%" />

### Constraint types

Each constraint (min or max) can be set using one of two modes, selectable via the **Value** or **% of request** toggle in the UI:

| UI mode          | Annotation `type`      | Description                                                                              |
| ---------------- | ---------------------- | ---------------------------------------------------------------------------------------- |
| **Value**        | `constant`             | An absolute Kubernetes resource value (e.g., `500m`, `2Gi`).                             |
| **% of request** | `percentageOfOriginal` | A percentage of the workload's original resource request (e.g., `90` = 90% of original). |

> 📘 Note
>
> **% of request** requires the workload to have defined resource requests. If a workload does not have defined requests, the constraint will be treated as having no limit set. This mode is not supported for custom workloads or injected containers.

### Configuration options

Constraints can be configured at three levels.

#### Policy-level constraints (global)

Policy-level constraints apply to all workloads associated with a scaling policy. These constraints serve as default guardrails for all containers managed by that policy.
You can set:

* Global minimum CPU and memory limits
* Global maximum CPU and memory limits

#### Container-level constraints

For more granular control, you can specify constraints for individual containers within a workload using [annotations](workload-autoscaler-annotations-reference.md). Container-specific constraints override policy-level constraints when both are defined.
Container constraints include:

* Minimum and maximum CPU resources
* Minimum and maximum memory resources

##### Annotations example

For workloads with multiple containers with different resource profiles, you can set specific constraints for each using annotations:

<Tabs>
  <Tab title="Constant value">

```yaml
workloads.atlas.example/configuration: |
  vertical:
    containers:
      app-server:
        cpu:
          constraints:
            min:
              type: constant
              value: 500m
            max:
              type: constant
              value: 2000m
        memory:
          constraints:
            min:
              type: constant
              value: 1Gi
            max:
              type: constant
              value: 4Gi
      metrics-sidecar:
        cpu:
          constraints:
            min:
              type: constant
              value: 10m
            max:
              type: constant
              value: 100m
        memory:
          constraints:
            min:
              type: constant
              value: 64Mi
            max:
              type: constant
              value: 256Mi
```

  </Tab>
  <Tab title="% of original request">

```yaml
workloads.atlas.example/configuration: |
  vertical:
    cpu:
      constraints:
        min:
          type: percentageOfOriginal
          value: 90
        max:
          type: percentageOfOriginal
          value: 200
    memory:
      constraints:
        min:
          type: constant
          value: 128Mi
        max:
          type: percentageOfOriginal
          value: 150
```

> 📘 Note
>
> `percentageOfOriginal` is available at the global (`vertical.cpu` / `vertical.memory`) level only. Container-level constraints (`vertical.containers`) do not yet support `percentageOfOriginal`.

  </Tab>
</Tabs>

This allows for precise control over scaling boundaries for each container within your pod.

> 📘 Note
>
> When configuring container-specific constraints, the container name must match exactly what's defined in the pod specification.

#### Pod-level constraints

You can also specify constraints for the entire workload pod using [annotations](workload-autoscaler-annotations-reference.md). Pod-specific constraints are evaluated after policy-level and container-level constraints have been evaluated.
Pod constraints include:

* Minimum and maximum CPU resources
* Minimum and maximum memory resources

### Use cases

Workload Autoscaler constraints are valuable in multiple scenarios.

Setting minimum CPU/memory resources ensures business-critical services never scale below functional thresholds. Some applications (like JVM-based services) need a certain amount of memory to function, so setting a minimum for them will prevent performance issues on cold starts. For JVM-specific optimization that uses actual heap metrics, see [JVM workload optimization](woop-jvm-optimization.md).

Setting maximum CPU/memory resources is typically done for reasons such as:

* **Cost Control**: Limiting maximum resources prevents unexpected cost increases
* **Compliance**: Enforcing organizational policies about resource consumption limits

## Look-back period

The look-back period defines the timeframe the Workload Autoscaler uses to observe CPU and memory usage when calculating scaling recommendations. This feature allows you to customize the historical data window used for generating recommendations, which can be particularly useful for workloads with atypical resource usage patterns.

You can configure the look-back period under the Advanced Settings of a vertical scaling policy:

<Image src="https://files.readme.io/bd2b0a8a6c2310de019469e9425f1612751ba823c0cbfce59bab9daf4b95b286-image.png" alt="Look-back period in Advanced Settings" align="center" width="100% " caption="Look-back period in Advanced Settings" />

1. Set the look-back period for CPU and memory separately.
2. Specify the duration in days (d) and hours (h). The minimum allowed period is 3 hours, and the maximum is 7 days.

This feature allows you to:

* Adjust the recommendation window based on your workload's specific resource usage patterns.
* Account for longer-term trends or cyclical resource usage in your applications.

You can configure this setting at different levels:

* **Policy level**: Apply the setting to all workloads assigned to a specific scaling policy.
* **Individual workload level**: Configure the setting for a specific workload using [annotations](workload-autoscaler-annotations-reference.md) to override policy-level settings.

The look-back period can also be configured via Annotations, the [API](https://docs.atlas.example/reference/workloadoptimizationapi_createworkloadscalingpolicy), or Terraform.

### Choosing the right look-back period

The optimal look-back period largely depends on your workload's resource usage patterns. Most applications benefit from a shorter look-back period of 1-2 days. This approach works particularly well for standard web applications, capturing daily usage patterns while maintaining high responsiveness to changes. Shorter periods enable more aggressive optimization and often lead to higher savings.

Some workloads, however, require longer observation periods of 3-7 days. Applications with significant differences between weekday and weekend usage patterns benefit from a 7-day period to capture these weekly variations. Batch processing jobs that run every few days need a look-back period that covers at least one full job cycle to prevent potential out-of-memory (OOM) situations.

Common use cases and recommended periods:

* **High-frequency trading or real-time applications**: 3-6 hours for rapid scaling response
* **Standard web applications**: 1-2 days captures daily patterns while maintaining responsiveness to changes
* **Batch processing jobs**: Set to cover at least one full job cycle to account for periodic resource spikes
* **Weekend-sensitive workloads**: 7 days to capture both weekday and weekend patterns
* **Variable workloads**: Start with 1-2 days and adjust based on observed scaling behavior

> 💡 Tip
>
> For workloads with variable or uncertain patterns, start with a shorter period and adjust based on observed behavior. The key is to match the look-back period to your application's actual resource usage patterns – whether that's daily consistency, weekly cycles, or periodic processing jobs.

## Startup metrics

Some workloads, notably Java and .NET applications, consume more CPU during initialization than at steady state. Without configuration, Workload Autoscaler treats startup usage like any other sample, which can inflate recommendations and cause HPA spikes. The startup metrics setting controls how Workload Autoscaler handles resource usage observed during the startup period.

You can configure this setting in the Atlas console under the Advanced Settings of a vertical scaling policy:

<Image src="https://files.readme.io/c7dd15fd57d6d386a35eb2a155f367fc237ec050befea53129945de265a8741f-2026-07-14_13-45-27.png" alt="Startup metrics at the policy level" align="center" caption="Startup metrics at the vertical scaling policy level" />

### Modes

Choose one of three modes and set the startup period duration (2–60 minutes, or a custom value):

| Mode                                      | Behavior                                                                                                                                                           | When to use                                                                                     |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------- |
| **Regular recommendations**               | Includes startup metrics in recommendation calculations like any other period. May inflate values for workloads with high startup usage.                           | Workloads whose startup resource usage reflects their steady-state usage                        |
| **Startup recommendations** (recommended) | Applies the workload's original CPU requests during the startup period, then transitions to the optimized recommendation via in-place resize once the period ends. | Workloads that need extra CPU to start quickly or that trigger HPA spikes during initialization |
| **Ignore startup metrics** (legacy)       | Excludes startup metrics from recommendation calculations, preventing inflated values.                                                                             | Workloads where startup recommendations are not suitable or not available                       |

> 📘 Note
>
> The startup metrics setting only affects vertical scaling recommendation calculations. Horizontal scaling (HPA) responds normally to resource usage during startup regardless of this setting.

For Java workloads, also consider enabling [JVM heap-based optimization](woop-jvm-optimization.md), which uses actual JVM runtime metrics for more accurate memory recommendations.

You can also configure this setting via the [API](https://docs.atlas.example/reference/workloadoptimizationapi_createworkloadscalingpolicy) or Terraform.

For more information on startup recommendations, see [Pod startup recommendations](woop-startup-recommendations.md).

## Recommendation confidence

The recommendation confidence setting determines the minimum confidence level required before Workload Autoscaler automatically applies resource recommendations to workloads assigned to this policy.

### How confidence works

Workload Autoscaler calculates confidence based on the quality and quantity of historical usage data available for each workload. Confidence is expressed as a percentage that represents the **ratio of collected metric data points to the expected number of data points** within the configured look-back period.

For example, with a 7-day look-back period:

* **100% confidence** means the workload has complete metric coverage for all 7 days
* **90% confidence** (default threshold) means at least 90% of expected metric data points have been collected
* **40% confidence** represents the minimum metric density needed beyond the initial 24-hour period

The confidence score reflects both:

* **Time coverage**: How long the workload has been monitored
* **Data completeness**: The density and continuity of collected metrics within that period

**Confidence levels:**

* **Full confidence**: The workload has sufficient historical data for reliable, unrestricted recommendations. Optimization proceeds based on comprehensive usage analysis without artificial limits.
* **Low confidence**: The workload lacks sufficient historical data for reliable analysis. The autoscaler applies different strategies based on how optimization was enabled:
  * **Policy-optimized workloads**: Receive gradual scaling with conservative limits during the confidence-building period
  * **Manually enabled workloads**: Receive unrestricted recommendations immediately, bypassing confidence checks

### Setting the confidence threshold

Set the minimum confidence level required before automatically applying resource recommendations to workloads assigned to this policy. The default value is 90%.

![](https://files.readme.io/7fe3fe7195530cd15a2bf9f6edc9aaf2bfe943223272a0aafe95912ea39c2d8f-image.png)

**Confidence (default: 90%)**

* **Lower values** (e.g., 50-80%): Allow earlier optimization with less historical data. Useful for faster time-to-value but may be less reliable for workloads with variable usage patterns.
* **Higher values** (e.g., 95-100%): Wait for more complete historical data before applying recommendations. Provides more reliable optimization but delays scaling actions.

### Gradual scaling for new clusters

For newly onboarded clusters (within the first 24 hours), Workload Autoscaler provides immediate optimization benefits through gradual scaling, even for workloads with low confidence. This allows you to see optimization results within hours rather than waiting days for complete data collection.

| Workload Data | Max Resource Request Adjustment | Behavior After Threshold      |
| ------------- | ------------------------------- | ----------------------------- |
| < 2.4 hours   | 10%                             | -                             |
| < 6 hours     | 25%                             | -                             |
| < 24 hours    | 35%                             | -                             |
| > 24 hours    | -                               | Confidence-based optimization |

During this period:

* **Policy-optimized workloads** automatically receive conservative resource adjustments with progressive limits
* **Manually enabled workloads** (via annotations or direct UI toggle) bypass gradual scaling and receive unrestricted recommendations immediately

### Building sufficient confidence

For the Workload Autoscaler to consider confidence sufficient for unrestricted optimization:

1. **Minimum monitoring period**: The workload must be monitored for at least 24 hours, even for existing workloads in newly connected clusters.

2. **Adequate metric density**: Beyond the minimum 24-hour requirement, workloads with longer [look-back periods](#look-back-period) need metrics for approximately 40% of the configured period. For example, with a 7-day lookback period, confidence continues to build over approximately 3 days of continuous operation.

> ⚠️ Cluster hibernation impact
>
> If your cluster uses hibernation or has scheduled downtime periods, confidence calculations may be affected. During hibernation, no metrics are collected, which reduces the metric density used to calculate confidence scores. Workloads in hibernating clusters may require longer active periods to reach the configured confidence threshold before automatic optimization begins.

## Stall detection

Stall detection uses Pressure Stall Information (PSI) metrics to identify CPU contention that standard usage metrics miss. When workloads experience CPU stall above configured thresholds, Workload Autoscaler automatically increases resource recommendations to reduce contention.

Resource contention can cause workloads to underperform even when CPU usage appears normal. This happens when requests are set too low or the node lacks effective headroom. Stall detection captures this hidden contention and responds by adjusting resources, providing a more accurate signal than traditional CPU utilization alone.

![](https://files.readme.io/7e952769df5167c047b0f2cf4afbea17e192541971bd7f599e8dc730e4410f94-image.png)

### CPU stall threshold

The CPU stall percentage that triggers resource increases. When workloads can't get CPU time and stall exceeds this threshold, they'll be scaled up to reduce contention.

The stall percentage is measured over a 5-minute window. For example, a 50% threshold means workloads stalling for more than 2.5 minutes within each 5-minute period will trigger scaling.

| Setting             | Description                                                     | Value range | Default                 |
| ------------------- | --------------------------------------------------------------- | ----------- | ----------------------- |
| CPU Stall Threshold | Percentage of time spent waiting for CPU that triggers scaling. | 1–100%      | Varies by system policy |

System policies have the following default CPU stall thresholds:

| Policy               | CPU Stall Threshold | Description                                   |
| -------------------- | ------------------- | --------------------------------------------- |
| cost-saving          | 30%                 | Moderate contention acceptable                |
| burstable            | 50%                 | Higher contention acceptable                  |
| balanced             | 10%                 | Light contention acceptable                   |
| stability/resiliency | 5%                  | Minimal contention; prioritize responsiveness |

Custom policies inherit stall detection settings from the `balanced` policy by default.

### Pod threshold

The percentage of pods that must exceed the stall threshold before scaling. This prevents scaling the entire workload when only a single pod or a small subset is experiencing pressure.

| Setting       | Description                                       | Value range | Default |
| ------------- | ------------------------------------------------- | ----------- | ------- |
| Pod Threshold | Percentage of pods required above stall threshold | 1–100%      | 50%     |

For example, with a 50% pod threshold on a workload with 10 pods, at least 5 pods must experience CPU stall above the configured threshold before scaling triggers. All the pods need to report CPU stall above the threshold, it is not an average across affected pods.

### How stall detection works

The Workload Autoscaler collects CPU stall metrics per workload container through the [Workload Autoscaler Exporter](helm-charts.md#workload-autoscaler). The system checks each workload every 5 minutes, calculating the percentage of time it was waiting for CPU over that period. If the calculated percentage exceeds the threshold, a new recommendation is generated with adjustments designed to relieve contention. The recommendation is applied according to the scaling mode configured in the policy (immediate or deferred).

When CPU stall is detected, the system adjusts recommendations by:

* **Increasing recommendation percentile**: Higher percentiles capture more of the workload's resource demand
* **Shortening look-back window**: Focuses on recent usage patterns rather than historical data
* **Increasing overhead**: Surfaces "unseen" demand that traditional metrics miss

After issuing a stall-aware recommendation, the system waits 15 minutes before re-evaluating that workload. This stabilization window gives the updated recommendation time to take effect.

These adjustments appear in the [Event log](event-log.md) with a dedicated "CPU stall detected" event type.

### Detection constraints

Stall detection only activates when all of the following conditions are met:

* Workload kind is **not** DaemonSet
* Vertical autoscaling is enabled for the workload
* If the workload has HPA configured, HPA must be at maximum replicas already
* Latest recommendation was created on the cluster
* Latest recommendation is older than 1 minute
* The CPU stall threshold (derived from the assigned scaling policy) is exceeded
* At least the configured percentage of pods (Pod Threshold) is above the stall threshold
* If the last recommendation was stall-related, at least 15 minutes have passed since it was applied to the workload

### Prerequisites

Stall detection requires:

* `atlas-workload-autoscaler-exporter` (CME) installed in the cluster. This component is installed by default when installing Workload Autoscaler for new clusters.
* `atlas-workload-autoscaler-exporter` version v0.67.0 or above
* Kubernetes cluster version 1.34 or above

Once the Workload Autoscaler Exporter is installed on a compatible cluster, PSI metrics are automatically scraped.

### Configuring stall detection

You can configure stall detection thresholds when creating or editing a scaling policy:

1. Navigate to **Workload Autoscaler** → **Scaling policies**
2. Edit an existing policy or click **Create scaling policy**
3. Under the **Vertical optimization** tab, locate **Stall detection**
4. Set the **CPU Stall Threshold** percentage
5. Set the **Pod Threshold** percentage
6. Save the policy

The thresholds you configure apply to all workloads assigned to that scaling policy.

Stall detection can also be configured via annotations. See [vertical.anomalyDetection.cpuPressure](workload-autoscaler-annotations-reference.md#verticalanomalydetection) for details.

## HPA converter

When Atlas optimizes a workload's CPU or memory requests, any existing Horizontal Pod Autoscaler that uses percentage-based utilization targets can drift out of alignment. For example, an HPA targeting 50% CPU utilization effectively targets 500m when the request is 1000m — but only 100m after Atlas reduces the request to 200m. This mismatch can cause the HPA to scale to maximum replicas unnecessarily.

The HPA converter prevents this by converting utilization-based targets into fixed absolute values calculated from the workload's original (pre-optimization) requests. The conversion preserves the original scaling intent: the same real CPU or memory threshold triggers scaling, regardless of how Atlas adjusts the request.

> 📘 When does the converter apply?
>
> The HPA converter is designed for workloads that have their own HPA (one you created, or managed by a third-party controller) while Atlas handles only vertical optimization. If Atlas's horizontal autoscaling is also enabled for the workload, Atlas already manages the HPA directly and the converter is not needed.

### What gets converted

The converter rewrites `Resource` and `ContainerResource` metrics that use `target.type: Utilization` for CPU or memory. The formula is:

```text
averageValue = original_request × (averageUtilization / 100)
```

For pod-level `Resource` metrics on multi-container workloads, the total of all container requests is used as the baseline. For `ContainerResource` metrics, only the named container's request is used.

The converter requires all metrics in the HPA to be CPU or memory `Utilization` percentages. If the HPA contains any of the following, the converter skips the entire HPA and no metrics are converted:

* **External metrics** — metrics from outside the cluster (for example, cloud provider queue depth)

<Accordion title="Example: external metric target">

```yaml
spec:
  metrics:
  - external:
      metric:
        name: external_metric
      target:
        averageValue: "100"
        type: AverageValue
    type: External
```

</Accordion>

* **Object metrics** — metrics from a specific Kubernetes object

<Accordion title="Example: object metric target">

```yaml
spec:
  metrics:
  - object:
      metric:
        name: requests-per-second
      describedObject:
        apiVersion: networking.k8s.io/v1
        kind: Ingress
        name: main-route
      target:
        type: Value
        value: "10k"
    type: Object
```

</Accordion>

* **Pods metrics** — custom metrics reported per pod

<Accordion title="Example: pods metric target">

```yaml
spec:
  metrics:
  - pods:
      metric:
        name: packets-per-second
      target:
        type: AverageValue
        averageValue: "1k"
    type: Pods
```

</Accordion>

* **Already absolute targets** — CPU or memory metrics already expressed as `AverageValue`

<Accordion title="Example: absolute-value target only">

```yaml
spec:
  metrics:
  - resource:
      name: cpu
      target:
        averageValue: 500m
        type: AverageValue
    type: Resource
```

</Accordion>

* **Mixed utilization and absolute-value targets** — for example, CPU as `Utilization` and memory as `AverageValue`

<Accordion title="Example: mixed utilization and absolute-value targets">

```yaml
spec:
  metrics:
  - resource:
      name: cpu
      target:
        averageValue: 500m
        type: AverageValue
    type: Resource
  - resource:
      name: memory
      target:
        averageUtilization: 80
        type: Utilization
    type: Resource
```

</Accordion>

### How it behaves

* **Dynamic recalculation.** The converter allows all HPA updates and recalculates the conversion whenever the HPA changes — for example, when you update an existing CPU utilization target or add a memory utilization percentage to an HPA that previously only had CPU.
* **Only converted fields change.** The target type and value are rewritten. Replicas, behavior, labels, and annotations are untouched.
* **Original source remains the owner.** Continue managing your HPA through its original source, such as Helm, GitOps, KEDA, or your CI/CD pipeline.
  * Updates through the original source are fully supported. Kubernetes sends Atlas the complete desired HPA specification, allowing the converter to recalculate the converted fields while preserving your intent.
  * **Do not edit or patch the converted HPA directly.** Kubernetes sends Atlas only the final resolved HPA specification, which already contains the converter's `AverageValue` metrics. Because Atlas cannot distinguish the original HPA configuration from user changes, it treats the entire resolved specification as the new source of truth. As a result, the converter may step aside and stop managing the HPA until it is reapplied from the original configuration.
* **Automatic revert.** Atlas reverts the converted fields back to their pre-conversion values when any of the following occur:
  * You disable the converter or delete the HPA.
  * You remove resource requests from the workload, leaving the converter without a calculation baseline.
  * You introduce external metrics or otherwise change the HPA configuration to one the converter does not support.
  * You enable Atlas's horizontal autoscaling for a workload that previously used only vertical optimization with the converter. Horizontal autoscaling takes priority and the converter steps aside automatically.
* **Automatic refresh.** The converter recalculates immediately whenever resource requests change. There is no waiting period.

### Supported workload types and version requirements

The converter works with HPAs targeting Deployments, ReplicaSets, StatefulSets, and Argo Rollouts.

| Component                    | Minimum version |
| :--------------------------- | :-------------- |
| `atlas-workload-autoscaler` | v1.0.11         |

### Enabling the HPA converter

To enable the HPA converter in a scaling policy:

1. Navigate to **Workload Autoscaler** → **Scaling policies**.
2. Edit an existing policy or click **Create scaling policy**.
3. Under the **Vertical Rightsizing** tab, scroll to **HPA converter**.
4. Enable the **Enable HPA converter** toggle.
5. Save the policy.

The converter can also be enabled via workload annotations:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    optimization: true
    hpaConverters:
      - type: AverageValueFromOriginalRequests
```

For the full annotation reference, see [vertical.hpaConverters](workload-autoscaler-annotations-reference.md#verticalhpaconverters).

### Monitoring conversion health

After enabling the converter, watch for two conditions that can reduce its effectiveness:

* **HPA maxing out after conversion** — if average usage already exceeds the converted absolute target, the HPA will scale to its maximum replica count. This typically means the original utilization percentage was set too low relative to actual usage. Review the HPA's `averageValue` target against current pod metrics using `kubectl get hpa` and consider adjusting the utilization target or workload requests.
* **Converted target becoming unreachable** — if Atlas's recommended request grows larger than the converted target value, the HPA target can never be reached and scaling will not trigger. This can happen when actual usage rises significantly after conversion. Check that the relationship between the original request, utilization percentage, and current recommendation still makes sense.