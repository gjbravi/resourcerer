# HPA

> Configure how Orion interacts with HPA-managed workloads


Orion integrates with [Horizontal Pod Autoscalers](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/) (HPA) and can:

- Generate vertical rightsizing recommendations for HPA-managed workloads
- Preserve HPA scaling behavior when adjusting resource requests

## How Orion integrates with HPA {#how-it-works}

Orion integrates with the HPA by adjusting HPA target utilization when Pod resource requests change.

Because HPA scaling is based on a percentage of resource requests, changing requests without adjusting target utilization changes how a workload scales.

Orion balances vertical sizing changes with corresponding adjustments to HPA target utilization so the workload continues to scale at the same overall load level.

**The HPA scale point is the average Pod resource usage level at which the HPA begins adding or removing replicas.**

The scale point determines how many replicas the HPA deploys to handle a given load.



Preserving the scale point is usually the desired behavior because it maintains a workload's horizontal scaling characteristics.

The settings in this section control how strictly Orion preserves the scale point, and whether it may change it for idle or saturated workloads.

## HPA settings

Use these settings to control how Orion interacts with HPA-managed workloads.

- [Minimum and maximum bounds for HPA target utilization](#target-min-max)
- [Preserve horizontal scale point](#preserve-scale-point)
- [Ignore horizontally-scaled resources](#ignore-hpa-resources)
- [Ignore the HPA](#ignore-hpa)

## Minimum and maximum bounds for HPA target utilization {#target-min-max}

Specifies the lower and upper bounds within which Orion can set HPA target utilization.

| Annotation                                             | Default value                                                   |
| ------------------------------------------------------ | --------------------------------------------------------------- |
| `live.orion.example/hpa.cpu.target-utilization.min`    | `"50"`    |
| `live.orion.example/hpa.cpu.target-utilization.max`    | `"90"`    |
| `live.orion.example/hpa.memory.target-utilization.min` | `"50"` |
| `live.orion.example/hpa.memory.target-utilization.max` | `"90"` |

### Description {#target-min-max-description}

These constraints prevent Orion from recommending HPA target utilization values outside the defined bounds. CPU and memory constraints are configured separately.

These settings take effect only when:
- An HPA is enabled for a workload
- The HPA scales on CPU or memory

### Valid values {#target-min-max-valid-values}

- Integer strings greater than or equal to zero, for example, `"0"` or `"100"`

## Preserve horizontal scale point {#preserve-scale-point}

Controls whether Orion constrains resource request recommendations to maintain a stable HPA scale point.

| Annotation                                                 | Default value                                                       |
| ---------------------------------------------------------- | ------------------------------------------------------------------- |
| `live.orion.example/hpa.preserve-horizontal-scale-point` | `"False"` |

### Description {#preserve-scale-point-description}

When Orion adjusts resource requests for workloads with an HPA, the HPA scale point may change.

This setting controls whether, and under what circumstances, Orion may recommend values that change the scale point, in combination with other configuration constraints such as [min/max bounds for HPA target utilization](#target-min-max).

### Valid values {#preserve-scale-point-valid-values}

#### `True` {#preserve-scale-point-true}

Recommendations _never_ cause the HPA scale point to change.

Vertical sizing recommendations are constrained to values that can be balanced by corresponding adjustments to HPA target utilization without violating the configured [min/max bounds for HPA target utilization](#target-min-max).

Use this setting to preserve HPA scaling behavior in both directions. This ensures the workload scales up and down the same way it did before Orion was enabled, but may limit some potential cost savings.

#### `False` (default) {#preserve-scale-point-false}

Recommendations _may_ cause the HPA scale point to change.

Orion attempts to preserve the scale point for actively scaling workloads, but may change it for:

* An idle HPA that runs with `minReplicas` most or all of the time
* A saturated HPA that runs with `maxReplicas` most or all of the time
* A workload whose observed target utilization falls outside the configured [min/max bounds for HPA target utilization](#target-min-max)

Use this setting when you want Orion to preserve HPA scaling behavior where possible while allowing scale point changes that improve cost efficiency or reliability.

#### `UnlessUnderProvisioned` {#preserve-scale-point-unless-under-provisioned}

Recommendations never _reduce_ resource requests in a way that causes the HPA scale point to change.

Orion may recommend increasing resource requests to correct underprovisioning, even if doing so changes the scale point.

Use this setting when you want to preserve scaling behavior in most cases, but allow scale point changes to resolve underprovisioned workloads with a saturated HPA (running at `maxReplicas`).

## Ignore horizontally-scaled resources {#ignore-hpa-resources}

Controls whether Orion recommends changes to resource requests for resources scaled by an HPA.

| Annotation                                                    | Default value                                                          |
| ------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `live.orion.example/hpa.ignore-horizontally-scaled-resources` | `"False"` |

### Description {#ignore-hpa-resources-description}

In some environments, you may want to disable automatic rightsizing for resources scaled by an HPA.

For example, you may want to preserve manual CPU request and HPA configuration while still allowing Orion to rightsize memory.

When this setting is enabled, Orion doesn't recommend changes to resource requests for resources scaled by an HPA.

### Valid values {#ignore-hpa-resources-valid-values}

#### `True` {#ignore-hpa-resources-true}

Orion doesn't make recommendations for resources scaled by an HPA.

#### `False` (default) {#ignore-hpa-resources-false}

Orion may make recommendations for resources scaled by an HPA.

## Ignore the HPA {#ignore-hpa}

Controls whether Orion considers a workload's HPA configuration at all when generating recommendations.

| Annotation                      | Default value                            |
| ------------------------------- | ---------------------------------------- |
| `live.orion.example/hpa.ignore` | `"False"` |

### Description {#ignore-hpa-description}

Some HPAs exist only as an emergency backstop and aren't intended to scale the workload under normal conditions. For example, an HPA with target utilization thresholds far above normal operating levels, such as 500% CPU or 500% memory, adds replicas only during extreme resource spikes.

For workloads like these, preserving the HPA scale point isn't meaningful, and HPA-based constraints can prevent otherwise beneficial rightsizing.

When this setting is enabled, Orion generates recommendations as if the workload has no HPA:

- Vertical rightsizing recommendations aren't constrained by the HPA configuration.
- Orion doesn't recommend changes to HPA target utilization.
- The other settings in this section have no effect.

Unlike [ignore horizontally-scaled resources](#ignore-hpa-resources), which prevents Orion from changing resource requests for resources the HPA scales on, this setting removes the HPA from consideration entirely: Orion rightsizes all resources and leaves the HPA configuration unchanged.

### Valid values {#ignore-hpa-valid-values}

#### `True` {#ignore-hpa-true}

Orion ignores the HPA when generating recommendations for the workload.

#### `False` (default) {#ignore-hpa-false}

Orion considers the HPA when generating recommendations for the workload.

## KEDA-scaled workloads {#keda}

[Kubernetes Event-driven Autoscaling](https://keda.sh/) (KEDA) scales workloads on external event sources, such as queue depth or message lag, as well as on CPU and memory. KEDA creates and manages a standard Kubernetes HPA for each `ScaledObject`, so Orion detects KEDA-scaled workloads automatically and applies the [HPA settings](#hpa-settings) to them like any other HPA-managed workload.

Because KEDA owns the HPA it generates and reverts direct changes to it, Orion writes target utilization changes to the CPU or memory trigger on the `ScaledObject` instead. This requires no additional configuration, because the Applier's ClusterRole includes `get`, `list`, `watch`, and `patch` permissions on the `keda.sh` API group's `scaledobjects` resource by default.

[Minimum and maximum bounds for HPA target utilization](#target-min-max) and [preserve horizontal scale point](#preserve-scale-point) apply only where the HPA scales on CPU or memory. For a `ScaledObject` that scales solely on an external metric, Orion still generates CPU and memory rightsizing recommendations, but these two settings have no effect, since no CPU or memory target utilization drives the replica count.
