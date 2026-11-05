---
updatedAt: 2026-10-30T09:15:14.000Z
---

# Annotations reference

All Workload Autoscaler settings are available by adding annotations to the workload controller. Annotations are the only way to override vertical scaling settings for individual workloads — the Atlas console does not support per-workload setting overrides. When the `workloads.atlas.example/configuration` annotation is detected on a workload, it will be considered as configured by annotations. This allows for flexible configuration, combining annotations and scaling policies.

**Changes to the settings via the API/UI are no longer permitted for workloads with annotations**. The default or scaling policy value is used when a workload does not have an annotation for a specific setting.

Annotation values take precedence over what is defined in a scaling policy. This means that if a scaling policy is defined in the workload configuration under annotations, all of the individual configuration options defined under the annotation will override the respective policy values. Those that are not defined under the annotation will use system defaults or what is defined in the scaling policy.

The following example shows a workload configured with the `workloads.atlas.example/configuration` annotation:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    app: my-app
  annotations:
    workloads.atlas.example/configuration: |
      scalingPolicyName: custom
      vertical:
        optimization: on
        applyType: immediate
        anomalyDetection:
          cpuPressure:
            cpuStallThresholdPercentage: 30.0
            minPressuredPodPercentage: 20.0
        excludedContainers:
          - istio-proxy
        antiAffinity:
          considerAntiAffinity: false
        startup:
          period: 5m
        confidence:
          threshold: 0.5
        cpu:
          target: p81
          lookBackPeriod: 25h
          min: 1000m
          max: 2500m
          applyThresholdStrategy:
            type: defaultAdaptive
          overhead: 0.15
          limit:
            type: multiplier
            multiplier: 2.0
        memory:
          target: max
          lookBackPeriod: 30h
          min: 2Gi
          max: 10Gi
          applyThresholdStrategy:
            type: defaultAdaptive
          overhead: 0.35
          limit:
            type: noLimit
        downscaling:
          applyType: immediate
        memoryEvent:
          applyType: immediate
        containers:
          {container_name}:
            cpu:
              min: 10m
              max: 1000m
            memory:
              min: 10Mi
              max: 2048Mi
        jvm:
          enabled: true
          autoInstrument: true
      hpaConverters:
        - type: AverageValueFromOriginalRequests
      rolloutBehavior:
        type: NoDisruption
        preferOneByOne: true
        delaySeconds: 120
      horizontal:
        optimization: on
        minReplicas: 5
        maxReplicas: 10
        scaleDown:
          stabilizationWindow: 5m
      containersGrouping:
        - key: name
          operator: contains
          values: ["data-processor"]
          into: data-processor
```

## Labels

In addition to the `workloads.atlas.example/configuration` annotation, Workload Autoscaler recognizes the following label on workload controllers and pods:

| Label                                 | Values   | Where to apply                                                   | Description                                                                                                                                                                                                                                                                         |
| ------------------------------------- | -------- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `workload-autoscaler.atlas.example/enabled` | `"true"` | Controller `metadata.labels` and `spec.template.metadata.labels` | Required when [whitelisting mode](workload-autoscaling-configuration.md#whitelisting-mode) is enabled. Only workloads with this label receive applied recommendations. The label must be present on both the controller (Deployment, StatefulSet, DaemonSet) and the pod template. |

## Runtime annotation

In addition to the `workloads.atlas.example/configuration` annotation, Workload Autoscaler recognizes the following annotation, which declares the workload's runtime instead of configuring scaling settings:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
spec:
  template:
    metadata:
      annotations:
        workloads.atlas.example/runtime: "jvm"
```

| Annotation                  | Values                   | Where to apply                                    | Description                                                                                                                                                                                                                                                  |
| --------------------------- | ------------------------ | ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `workloads.atlas.example/runtime` | `jvm` (case-insensitive) | Pod template `spec.template.metadata.annotations` | Marks the workload as a JVM workload. Use it when the automatic signals don't match your workload, such as a custom base image or a launcher script that doesn't call `java` directly. Any other value has no effect. Applies to single-container pods only. |

Detected workloads get the **JVM** badge in the console and become eligible for [JVM workload optimization](woop-jvm-optimization.md). The annotation doesn't enable JVM optimization by itself. Enable it through a scaling policy or the `workloads.atlas.example/configuration` annotation.

## Configuration structure

Below is a configuration structure reference for setting up a workload to be controlled by annotations.

> 📘 Note
>
> `workloads.atlas.example/configuration` has to be a valid YAML string. In cases where the annotation contains an invalid YAML string, the entire configuration will be ignored.

In the field tables, `*Yes` in the Required column marks a conditionally required field, and `-` in the Default column means no default.

### scalingPolicyName

If not set, the system assigns the workload to system default policies based on workload type: `resiliency` for StatefulSets and `balanced` for all other workloads.

| Field               | Type   | Required | Default                    | Description                                                                                                                                                 |
| :------------------ | :----- | :------- | :------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scalingPolicyName` | string | No       | System default (see below) | Specifies the scaling policy name to use. When set, this annotation allows the workload to be managed by both annotations and the specified scaling policy. |

```yaml
scalingPolicyName: custom-policy
```

**Fallback for invalid policy names**

If the `scalingPolicyName` value does not match any existing scaling policy, the workload falls back to system default policies:

* **StatefulSets**: `resiliency`
* **All other workloads**: `balanced`

When this fallback occurs, the Atlas Console displays a warning status on the workload in the Optimization page, such as: `Could not find scaling policy by name '<policy-name>'`. To resolve this, update the annotation to reference a valid scaling policy name.

### vertical

| Field      | Type   | Required | Default | Description                     |
| :--------- | :----- | :------- | :------ | :------------------------------ |
| `vertical` | object | No       | -       | Vertical scaling configuration. |

```yaml
vertical:
  optimization: on
  applyType: immediate
  antiAffinity:
    considerAntiAffinity: false
  startup:
    period: 5m
  confidence:
    threshold: 0.5
  excludedContainers:
    - istio-proxy
```

#### vertical.optimization

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `optimization`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Enable vertical scaling ("on"/"off").

        If using the `vertical` configuration option, this field becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  optimization: on
```

#### vertical.applyType

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `applyType`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        "immediate"
      </td>

      <td>
        Allows configuring the autoscaler operating mode to apply the recommendations.
        Use `immediate` to apply recommendations as soon as the thresholds are passed.

        **Note**: `immediate` mode can cause pod restarts. Use `deferred` to avoid forced restarts — on Kubernetes v1.33+ clusters recommendations may still be applied via in-place resizing without a restart.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  applyType: immediate
```

#### vertical.antiAffinity

| Field          | Type   | Required | Default | Description                                                          |
| :------------- | :----- | :------- | :------ | :------------------------------------------------------------------- |
| `antiAffinity` | object | No       | -       | Configuration for handling pod anti-affinity scheduling constraints. |

```yaml
vertical:
  antiAffinity:
    considerAntiAffinity: false
```

##### vertical.antiAffinity.considerAntiAffinity

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `considerAntiAffinity`
      </td>

      <td>
        boolean
      </td>

      <td>
        *Yes
      </td>

      <td>
        `false`
      </td>

      <td>
        When `true`, workload autoscaler will respect pod anti-affinity rules on hostname or host port and, as a result, issue recommendations for these pods in a deferred manner.
        When `false` (default), recommendations for pods containing one of the constraints above will be applied immediately instead.

        *If using the `vertical.antiAffinity` configuration option, this field becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  antiAffinity:
    considerAntiAffinity: false
```

#### vertical.startup

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `startup`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Configuration for how Workload Autoscaler handles resource usage observed during the startup period. Supports three modes: include startup metrics in calculations (default), apply original CPU requests during startup then switch to the optimized recommendation (startup recommendations), or exclude startup metrics from calculations (legacy ignore mode).

        See [Startup metrics](woop-configuration-settings.md#startup-metrics).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 5m
```

##### vertical.startup.period

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `period`
      </td>

      <td>
        duration
      </td>

      <td>
        *Yes
      </td>

      <td>
        "2m"
      </td>

      <td>
        Duration of the startup period. Applies to all startup metrics modes: determines how long to apply original CPU requests (startup recommendations mode), or how long to exclude metrics from calculations (ignore mode). Valid values range from `2m` to `60m`. Set to `0s` to disable. Required when `vertical.startup` is set.

        *If using the `vertical.startup` configuration option, this field becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 5m # Example: ignore first 5 minutes of metrics
```

```yaml
vertical:
  startup:
    period: 0s # Disable startup metrics ignore period
```

##### vertical.startup.twoPhaseRecommendations

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `twoPhaseRecommendations`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Configure two-phase startup recommendations. When enabled, newly created pods receive the workload's original resource requests during the startup period, then transition to the optimized recommendation via in-place resize once the startup period ends.

        See [Pod startup recommendations](woop-startup-recommendations.md).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 30s
    twoPhaseRecommendations:
      enabled: true
```

###### vertical.startup.twoPhaseRecommendations.enabled

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `enabled`
      </td>

      <td>
        boolean
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Set to `true` to enable two-phase startup recommendations for this workload.

        *Required when using the `vertical.startup.twoPhaseRecommendations` configuration option.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 30s
    twoPhaseRecommendations:
      enabled: true
```

###### vertical.startup.twoPhaseRecommendations.requestsOnStartup

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `requestsOnStartup`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Override the resource requests used during the startup phase. Each field (`cpu`, `memory`) is optional and falls back to the workload's original pod request for that resource when omitted. Set this when the original requests are unknown or should be different from the workload spec.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 30s
    twoPhaseRecommendations:
      enabled: true
      requestsOnStartup:
        cpu: 500m
        memory: 256Mi
```

###### vertical.startup.twoPhaseRecommendations.requestsOnStartup.cpu

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `cpu`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        Original pod request
      </td>

      <td>
        CPU request to use during the startup phase. Uses Kubernetes CPU notation (for example, `500m`, `2`). When omitted, the workload's original CPU request is used.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 30s
    twoPhaseRecommendations:
      enabled: true
      requestsOnStartup:
        cpu: 500m # memory falls back to original pod request
```

###### vertical.startup.twoPhaseRecommendations.requestsOnStartup.memory

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `memory`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        Original pod request
      </td>

      <td>
        Memory request to use during the startup phase. Uses Kubernetes memory notation (for example, `256Mi`, `1Gi`). When omitted, the workload's original memory request is used.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  startup:
    period: 30s
    twoPhaseRecommendations:
      enabled: true
      requestsOnStartup:
        memory: 256Mi # cpu falls back to original pod request
```

#### vertical.confidence

| Field        | Type   | Required | Default | Description                                             |
| :----------- | :----- | :------- | :------ | :------------------------------------------------------ |
| `confidence` | object | No       | -       | Configuration for recommendation confidence thresholds. |

```yaml
vertical:
  confidence:
    threshold: 0.5
    required: false
```

##### vertical.confidence.required

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `required`
      </td>

      <td>
        bool
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        When set to:

        * `true`: The workload will **require** confidence metrics before optimization, even when directly enabled via annotations. This prevents immediate optimization of workloads with highly variable load patterns until sufficient historical data is available.

        * `false`: When enabled via annotations, the workload will be optimized immediately, bypassing the confidence check. **Note**: This option is not recommended.

        *If using the `vertical.confidence` configuration option, at least one confidence field is required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  confidence:
    required: true
```

##### vertical.confidence.threshold

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `threshold`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        0.9
      </td>

      <td>
        Minimum confidence score required to apply recommendations (0.0-1.0). Higher values require more data points for recommendations.

        *If using the `vertical.confidence` configuration option, at least one confidence field is required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  confidence:
    threshold: 0.5
```

#### vertical.anomalyDetection

| Field              | Type   | Required | Default | Description                                                                                                                |
| :----------------- | :----- | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------- |
| `anomalyDetection` | object | No       | -       | Configuration for anomaly detection features, like CPU pressure-based stall detection and unbound memory growth detection. |

```yaml
vertical:
  anomalyDetection:
    cpuPressure:
      cpuStallThresholdPercentage: 10.0
      minPressuredPodPercentage: 50.0
```

##### vertical.anomalyDetection.cpuPressure

| Field         | Type   | Required | Default | Description                                                             |
| :------------ | :----- | :------- | :------ | :---------------------------------------------------------------------- |
| `cpuPressure` | object | No       | -       | Configuration for CPU pressure-based stall detection using PSI metrics. |

```yaml
vertical:
  anomalyDetection:
    cpuPressure:
      cpuStallThresholdPercentage: 10.0
      minPressuredPodPercentage: 50.0
```

###### vertical.anomalyDetection.cpuPressure.cpuStallThresholdPercentage

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `cpuStallThresholdPercentage`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        Varies by policy
      </td>

      <td>
        The CPU stall percentage that triggers resource increases. When workloads can't get CPU time and stall exceeds this threshold, they'll be scaled up to reduce contention. Measured over a 5-minute window. Value range: 1.0–100.0.

        See [Stall detection](woop-configuration-settings.md#stall-detection).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  anomalyDetection:
    cpuPressure:
      cpuStallThresholdPercentage: 30.0 # 30% stall threshold
```

###### vertical.anomalyDetection.cpuPressure.minPressuredPodPercentage

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `minPressuredPodPercentage`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        Varies by policy
      </td>

      <td>
        The percentage of pods that must exceed the stall threshold before scaling triggers. This prevents scaling the entire workload when only a single pod or small subset is experiencing pressure. Value range: 1.0–100.0.

        See [Stall detection](woop-configuration-settings.md#stall-detection).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  anomalyDetection:
    cpuPressure:
      minPressuredPodPercentage: 20.0 # Scale when 20% of pods are stalling
```

##### vertical.anomalyDetection.infiniteMemoryScaling

| Field                   | Type   | Required | Default | Description                                                            |
| :---------------------- | :----- | :------- | :------ | :--------------------------------------------------------------------- |
| `infiniteMemoryScaling` | object | No       | -       | Configuration for unbound memory growth detection. Enabled by default. |

```yaml
vertical:
  anomalyDetection:
    infiniteMemoryScaling:
      enabled: true
```

###### vertical.anomalyDetection.infiniteMemoryScaling.enabled

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `enabled`
      </td>

      <td>
        boolean
      </td>

      <td>
        No
      </td>

      <td>
        `true`
      </td>

      <td>
        Keep unbound memory growth detection enabled for this workload. The detector watches for memory usage and recommendations that keep growing without stabilizing.

        When detected, the Workload Autoscaler caps memory for the affected containers to stop the growth. If it can't establish a reliable baseline, it sets the workload's automation to read-only for 24 hours and notifies you. A baseline is unreliable for auto-instrumented containers and custom workloads with a modified pod spec.

        Set to `false` to exclude the workload from detection.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  anomalyDetection:
    infiniteMemoryScaling:
      enabled: false # Exclude this workload from unbound memory growth detection
```

#### vertical.resourceStrategy

| Field              | Type   | Required | Default | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| :----------------- | :----- | :------- | :------ | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `resourceStrategy` | object | No       | -       | Sizes the workload using a strategy other than the standard resource recommendation. At pod admission, Workload Autoscaler resolves it per container and writes the resulting requests and limits to the pod, overriding the standard recommendation for that resource. You can also set it per container via [`vertical.containers.{container_name}.resourceStrategy`](#verticalcontainerscontainer_nameresourcestrategy), which overrides the workload-level setting. See [Node-aware DaemonSet sizing](woop-node-aware-daemonset-sizing.md). |

```yaml
vertical:
  resourceStrategy:
    type: nodeAllocatablePercentage
    nodeAllocatablePercentage:
      cpuPercent: 5
      memoryPercent: 2
```

At this (workload) level, the strategy applies to every container in the pod that already has CPU or memory requests in its spec. To size a container that has no requests defined, or to use different percentages per container, set it at the [container level](#verticalcontainerscontainer_nameresourcestrategy).

##### vertical.resourceStrategy.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Which sizing strategy to apply to the workload. Supported values:

        * `nodeAllocatablePercentage` — size requests as a percentage of the node's allocatable CPU and memory, so each pod gets requests proportional to the node it runs on. See [Node-aware DaemonSet sizing](woop-node-aware-daemonset-sizing.md).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  resourceStrategy:
    type: nodeAllocatablePercentage
```

##### vertical.resourceStrategy.nodeAllocatablePercentage

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `nodeAllocatablePercentage`
      </td>

      <td>
        object
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Percentages used by the `nodeAllocatablePercentage` strategy. Set `cpuPercent`, `memoryPercent`, or both — the resource you don't configure here keeps the standard Workload Autoscaler recommendation.

        Cannot be combined with the `keepLimits` resource limit option on the same resource. Workload Autoscaler rejects the configuration if both are set on the same resource (CPU or memory).
      </td>
    </tr>

  </tbody>
</Table>

\*Required when `type` is `nodeAllocatablePercentage`.

```yaml
vertical:
  resourceStrategy:
    type: nodeAllocatablePercentage
    nodeAllocatablePercentage:
      cpuPercent: 5
      memoryPercent: 2
```

###### vertical.resourceStrategy.nodeAllocatablePercentage.cpuPercent

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `cpuPercent`
      </td>

      <td>
        float
      </td>

      <td>
        One of `cpuPercent` or `memoryPercent` is required
      </td>

      <td>
        \-
      </td>

      <td>
        Percentage of the node's allocatable CPU to request, resolved **per container** (each container gets its own request of `node allocatable CPU × cpuPercent / 100`, not a shared pod budget). Valid range: `(0, 100]`.

        At pod admission time, Workload Autoscaler computes the CPU request as `node allocatable CPU × cpuPercent / 100`, clamps it to the workload's `vertical.cpu.min` / `vertical.cpu.max` constraints (if set), then derives the CPU limit from the clamped request using the workload's [CPU limit strategy](#verticalcpulimit).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  resourceStrategy:
    type: nodeAllocatablePercentage
    nodeAllocatablePercentage:
      cpuPercent: 5
```

###### vertical.resourceStrategy.nodeAllocatablePercentage.memoryPercent

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `memoryPercent`
      </td>

      <td>
        float
      </td>

      <td>
        One of `cpuPercent` or `memoryPercent` is required
      </td>

      <td>
        \-
      </td>

      <td>
        Percentage of the node's allocatable memory to request, resolved **per container** (each container gets its own request of `node allocatable memory × memoryPercent / 100`, not a shared pod budget). Valid range: `(0, 100]`.

        At pod admission time, Workload Autoscaler computes the memory request as `node allocatable memory × memoryPercent / 100`, clamps it to the workload's `vertical.memory.min` / `vertical.memory.max` constraints (if set), then derives the memory limit from the clamped request using the workload's [memory limit strategy](#verticalmemorylimit).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  resourceStrategy:
    type: nodeAllocatablePercentage
    nodeAllocatablePercentage:
      memoryPercent: 2
```

#### vertical.cpu

| Field | Type   | Required | Default | Description                         |
| :---- | :----- | :------- | :------ | :---------------------------------- |
| `cpu` | object | No       | -       | CPU-specific scaling configuration. |

```yaml
vertical:
  cpu:
    target: p80
    lookBackPeriod: 24h
    min: 100m
    max: 1000m
    applyThresholdStrategy:
    	type: defaultAdaptive
    overhead: 0.1
```

##### vertical.cpu.target

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `target`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        "p80"
      </td>

      <td>
        Resource usage target:

        * `max` - Use maximum observed usage
        * `p{0-99.9}` - Use percentile (e.g., `p80` for 80th percentile).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    target: p80
```

##### vertical.cpu.customMetric

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `customMetric`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Use a custom metric as the recommendation source for CPU instead of container usage samples. The metric must come from a connected custom metrics data source. If the metric has no data, the Workload Autoscaler reports a metric-missing error for the workload and keeps the current recommendation.

        See [Custom metrics](woop-custom-metrics.md).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    customMetric:
      name: my_cpu_metric
```

##### vertical.cpu.lookBackPeriod

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `lookBackPeriod`
      </td>

      <td>
        duration
      </td>

      <td>
        No
      </td>

      <td>
        "24h"
      </td>

      <td>
        Historical resource usage data window to consider for recommendations (3h-168h).

        See <Anchor label="Look-back Period" target="_blank" href="doc:workload-autoscaling-configuration#look-back-period">Look-back Period</Anchor>.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    lookBackPeriod: 24h
```

##### vertical.cpu.constraints

Defines minimum and maximum CPU bounds for the autoscaler. Two formats are supported:

* **`constraints` object** (required when using `percentageOfOriginal`): specifies a typed constraint with `type` and `value` fields.
* **Legacy flat format** (`min`/`max` as strings): still supported for constant values and not planned for deprecation.

| Field         | Type   | Required | Default | Description                                                                                       |
| :------------ | :----- | :------- | :------ | :------------------------------------------------------------------------------------------------ |
| `constraints` | object | No       | -       | Container for `min` and `max` CPU constraint objects. Required when using `percentageOfOriginal`. |

###### vertical.cpu.constraints.min

| Field   | Type             | Required | Description                                                                                                               |
| :------ | :--------------- | :------- | :------------------------------------------------------------------------------------------------------------------------ |
| `type`  | string           | Yes      | `constant` for an absolute value; `percentageOfOriginal` for a percentage of the workload's original request.             |
| `value` | string \| number | Yes      | For `constant`: Kubernetes CPU notation (e.g., `"100m"`, `"2"`). For `percentageOfOriginal`: a number (e.g., `90` = 90%). |

###### vertical.cpu.constraints.max

Same fields as `vertical.cpu.constraints.min`.

<Tabs>
  <Tab title="constraints object">

```yaml
vertical:
  cpu:
    constraints:
      min:
        type: percentageOfOriginal
        value: 90
      max:
        type: constant
        value: 2000m
```

  </Tab>
  <Tab title="Legacy flat format (still supported)">

```yaml
vertical:
  cpu:
    min: 100m
    max: 2000m
```

  </Tab>
</Tabs>

> 📘 Note
>
> `percentageOfOriginal` requires the workload to have defined resource requests. Not supported for custom workloads or injected containers. If requests are not defined, the constraint is treated as having no limit set.

##### vertical.cpu.applyThreshold

> 🚧 Deprecation Notice
>
> The `applyThreshold` configuration option is deprecated but still supported for backward compatibility. We strongly recommend migrating to the new `applyThresholdStrategy` configuration format for future compatibility and access to the latest features. See [applyThresholdStrategy](workload-autoscaling-configuration.md#verticalcpuapplythresholdstrategy).

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `applyThreshold`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        0.1
      </td>

      <td>
        The relative difference required between current and recommended resource values to apply a change immediately:

        * for upscaling, the difference is calculated relative to current resource requests;
        * for downscaling, it's calculated relative to the new recommended value. For example, with a threshold of 0.1 (10%), an upscale from 100m to 120m CPU would be applied immediately (20% increase relative to current 100m), while an upscale from 110m to 120m would not be applied immediately (8% increase relative to new 120m).Value range: 0.01-2.5.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThreshold: 0.1
```

##### vertical.cpu.applyThresholdStrategy

> 🚧 Warning
>
> `applyThreshold` and `applyThresholdStrategy` cannot be used simultaneously in a configuration as that will result in an error. `applyThresholdStrategy` is the latest and recommended configuration option.

| Field                    | Type   | Required | Default | Description                                                                                                                                                                                     |
| :----------------------- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `applyThresholdStrategy` | object | No       | -       | Configuration for the strategy used to determine when recommendations should be applied. The strategy determines how the threshold percentage is calculated based on current resource requests. |

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: defaultAdaptive
```

###### vertical.cpu.applyThresholdStrategy.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        "defaultAdaptive"
      </td>

      <td>
        The type of threshold strategy to use:

        * `defaultAdaptive` - Automatically adjusts thresholds based on workload size.
        * `percentage` - Uses a fixed percentage threshold.
        * `customAdaptive` - Allows custom configuration of the adaptive threshold formula. **Recommended for power users only**. It works in the same way as the Default Adaptive Threshold, but it allows tweaking the parameters of the adaptive threshold formula.The `defaultAdaptive` threshold option uses the following values:
          `numerator` = `0.5`
          `denominator` = `1`
          `exponent` = `1` (same effect as not used, i.e., no influence on the calculation)*Required when using `applyThresholdStrategy`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: defaultAdaptive # Using default adaptive threshold
```

###### vertical.cpu.applyThresholdStrategy.percentage

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `percentage`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        The fixed percentage threshold to use. Value range: 0.01-2.5.

        *Required when type is `percentage`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: percentage # Using fixed percentage threshold
      percentage: 0.3 # 30% threshold
```

###### vertical.cpu.applyThresholdStrategy.numerator

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `numerator`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        0.5
      </td>

      <td>
        Affects the vertical stretch of the threshold function. Lower values create smaller thresholds.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: customAdaptive # Using custom adaptive threshold
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

###### vertical.cpu.applyThresholdStrategy.denominator

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `denominator`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        1
      </td>

      <td>
        Affects threshold sensitivity for small workloads. Values close to `0` result in larger thresholds for small workloads. For example, when `numerator` is `1`, `exponent` is `1` and `denominator` is `0` the threshold for `0.5` req. CPU will be 200%.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: customAdaptive
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

###### vertical.cpu.applyThresholdStrategy.exponent

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `exponent`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        1
      </td>

      <td>
        Controls how quickly the threshold decreases for larger workloads. Lower values prevent extremely small thresholds for large resources.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    applyThresholdStrategy:
      type: customAdaptive
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

##### vertical.cpu.overhead

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `overhead`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        0.1
      </td>

      <td>
        Additional resource buffer when applying recommendations (0.0-2.5, e.g., 0.1 = 10%).

        If a 10% buffer is configured, the issued recommendation will have +10% added to it, so that the workload can handle further increased resource demand.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    overhead: 0.1
```

##### vertical.cpu.limit

> 📘 Relationship to UI options
>
> The annotation fields in this section correspond to the **Workload resource limits** options in the Atlas console. See [Workload resource limits](woop-configuration-settings.md#workload-resource-limits) for the UI-level description of each mode, including how the **Semi-automatic** (CPU) and **Automatic** (memory) shortcuts map to specific `type`, `multiplier`, `onlyIfOriginalExist`, and `onlyIfOriginalLower` combinations.

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `limit`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Configuration for container CPU limit scaling.

        Default behaviour when not specified:

        * If `spec.containers[].resources.limits.cpu` is _not_ defined on the workload, no limit is set by the workload autoscaler.
        * If `spec.containers[].resources.limits.cpu` _is_ defined on the workload, it is removed by Workload Autoscaler.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    limit:
      type: multiplier
      multiplier: 2.0
```

###### vertical.cpu.limit.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Type of CPU limit scaling to apply:

        * `noLimit` - Remove the resource limit from the workload definition entirely.

        * `multiplier` - Set limit as a multiple of requests using the following formula: `resources.requests.cpu * {multiplier}`. The behavior can be further controlled with optional flags `onlyIfOriginalExist` and `onlyIfOriginalLower`.

        * `keepLimits` - Preserve the existing limits as defined in the workload manifest without modification.

        * `maintainRatio` - Scale limits proportionally with recommended requests, preserving the original limit-to-request ratio from the workload manifest. If the original limit is 2× the request, the updated limit stays 2× the new recommended request. If no CPU limit exists on the workload, falls back to the default behavior (limits are removed).

        *If using the `vertical.cpu.limit` configuration option, this field becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    limit:
      type: keepLimits
```

###### vertical.cpu.limit.multiplier

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `multiplier`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Value to multiply the requests by to set the limit. The calculation is: `resources.requests.cpu * {multiplier}`. The value must be greater than or equal to `1`.

        *Required when type is set to `multiplier`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    limit:
      type: multiplier
      multiplier: 2.0
```

###### vertical.cpu.limit.onlyIfOriginalExist

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `onlyIfOriginalExist`
      </td>

      <td>
        boolean
      </td>

      <td>
        No
      </td>

      <td>
        false
      </td>

      <td>
        When set to `true`, CPU limits will only be set if the workload originally had CPU limits defined in its manifest. If the original workload has no CPU limits specified, no limits will be added.

        This flag allows conditional limit management based on the original workload configuration.

        Only applicable when the type is set to `multiplier`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    limit:
      type: multiplier
      multiplier: 2.0
      onlyIfOriginalExist: true
```

###### vertical.cpu.limit.onlyIfOriginalLower

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `onlyIfOriginalLower`
      </td>

      <td>
        boolean
      </td>

      <td>
        No
      </td>

      <td>
        false
      </td>

      <td>
        When set to `true`, CPU limits will only be updated if the original limits are lower than the calculated value (requests × multiplier). If the original limits are already higher than the calculated value, they remain unchanged.

        This flag prevents reducing existing limits and ensures limits only increase when beneficial.

        Only applicable when the type is set to `multiplier`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  cpu:
    limit:
      type: multiplier
      multiplier: 2.0
      onlyIfOriginalLower: true
```

**Combining both flags:**

When both `onlyIfOriginalExist` and `onlyIfOriginalLower` are set to `true`, the behavior matches Memory's Automatic mode: limits are only set when the workload originally had limits defined AND only when those original limits are lower than the calculated value.

```yaml
vertical:
  cpu:
    limit:
      type: multiplier
      multiplier: 1.5
      onlyIfOriginalExist: true
      onlyIfOriginalLower: true
```

##### vertical.cpu.optimization

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `optimization`
      </td>

      <td>
        string
      </td>

      <td>
        no
      </td>

      <td>
        \-
      </td>

      <td>
        This configuration option can be used to disable CPU management for workloads that benefit from memory management only. The workload will then use CPU requests/limits configured in the template.

        The only allowed option is off. If not set, the resource will inherit the workload management option as normal. Minimum required `workload-autoscaler` component version to use this feature is `v0.23.1`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml YAML
vertical:
  cpu:
    optimization: off
```

#### vertical.memory

| Field    | Type   | Required | Default | Description                            |
| :------- | :----- | :------- | :------ | :------------------------------------- |
| `memory` | object | No       | -       | Memory-specific scaling configuration. |

```yaml
vertical:
  memory:
    target: max
    lookBackPeriod: 24h
    min: 128Mi
    max: 2Gi
    applyThresholdStrategy:
    	type: defaultAdaptive
    overhead: 0.1
```

##### vertical.memory.target

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `target`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        "max"
      </td>

      <td>
        Resource usage target:

        * `max` - Use maximum observed usage.
        * `p{0-99.9}` - Use percentile (e.g., `p80` for 80th percentile).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    target: max
```

##### vertical.memory.customMetric

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `customMetric`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Use a custom metric as the recommendation source for memory instead of container usage samples. The metric must come from a connected custom metrics data source. If the metric has no data, the Workload Autoscaler reports a metric-missing error for the workload and keeps the current recommendation.

        See [Custom metrics](woop-custom-metrics.md).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    customMetric:
      name: my_memory_metric
```

##### vertical.memory.lookBackPeriod

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `lookBackPeriod`
      </td>

      <td>
        duration
      </td>

      <td>
        No
      </td>

      <td>
        "24h"
      </td>

      <td>
        Historical resource usage data window to consider for recommendations (3h-168h).

        See [Look-back Period](workload-autoscaling-configuration.md#look-back-period).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    lookBackPeriod: 24h
```

##### vertical.memory.constraints

Defines minimum and maximum memory bounds for the autoscaler. Two formats are supported:

* **`constraints` object** (required when using `percentageOfOriginal`): specifies a typed constraint with `type` and `value` fields.
* **Legacy flat format** (`min`/`max` as strings): still supported for constant values and not planned for deprecation.

| Field         | Type   | Required | Default | Description                                                                                          |
| :------------ | :----- | :------- | :------ | :--------------------------------------------------------------------------------------------------- |
| `constraints` | object | No       | -       | Container for `min` and `max` memory constraint objects. Required when using `percentageOfOriginal`. |

###### vertical.memory.constraints.min

| Field   | Type             | Required | Description                                                                                                                                                                                                  |
| :------ | :--------------- | :------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`  | string           | Yes      | `constant` for an absolute value; `percentageOfOriginal` for a percentage of the workload's original request.                                                                                                |
| `value` | string \| number | Yes      | For `constant`: Kubernetes memory notation (e.g., `"128Mi"`, `"2Gi"`). For `percentageOfOriginal`: a number (e.g., `90` = 90%). To control memory limits, see [vertical.memory.limit](#verticalmemorylimit). |

###### vertical.memory.constraints.max

Same fields as `vertical.memory.constraints.min`. **Note**: This does not cap the memory limit directly. To control memory limits, see [vertical.memory.limit](#verticalmemorylimit).

<Tabs>
  <Tab title="constraints object">

```yaml
vertical:
  memory:
    constraints:
      min:
        type: constant
        value: 128Mi
      max:
        type: percentageOfOriginal
        value: 150
```

  </Tab>
  <Tab title="Legacy flat format (still supported)">

```yaml
vertical:
  memory:
    min: 128Mi
    max: 2Gi
```

  </Tab>
</Tabs>

> 📘 Note
>
> `percentageOfOriginal` requires the workload to have defined resource requests. Not supported for custom workloads or injected containers. If requests are not defined, the constraint is treated as having no limit set.

##### vertical.memory.applyThreshold

> 🚧 Deprecation Notice
>
> The `applyThreshold` configuration option is deprecated but still supported for backward compatibility. We strongly recommend migrating to the new `applyThresholdStrategy` configuration format for future compatibility and access to the latest features. See [applyThresholdStrategy](workload-autoscaling-configuration.md#verticalmemoryapplythresholdstrategy).

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `applyThreshold`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        0.1
      </td>

      <td>
        The relative difference required between current and recommended resource values to apply a change immediately:

        * for upscaling, the difference is calculated relative to current resource requests;
        * for downscaling, it's calculated relative to the new recommended value. For example, with a threshold of 0.1 (10%), an upscale from 100MiB to 120MiB of memory would be applied immediately (20% increase relative to current 100MiB), while an upscale from 110MiB to 120MiB would not be applied immediately (8% increase relative to new 120MiB). Value range: 0.01-2.5.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThreshold: 0.1
```

##### vertical.memory.applyThresholdStrategy

> 🚧 Warning
>
> `applyThreshold` and `applyThresholdStrategy` cannot be used simultaneously in a configuration as that will result in an error. `applyThresholdStrategy` is the latest and recommended configuration option.

| Field                    | Type   | Required | Default | Description                                                                                                                                                                                     |
| :----------------------- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `applyThresholdStrategy` | object | No       | -       | Configuration for the strategy used to determine when recommendations should be applied. The strategy determines how the threshold percentage is calculated based on current resource requests. |

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: customAdaptive # Using custom adaptive threshold
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

###### vertical.memory.applyThresholdStrategy.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        "defaultAdaptive"
      </td>

      <td>
        The type of threshold strategy to use:

        * `defaultAdaptive` - Automatically adjusts thresholds based on workload size.
        * `percentage` - Uses a fixed percentage threshold.
        * `customAdaptive` - Allows custom configuration of the adaptive threshold formula. **Recommended for power users only**. It works in the same way as the Default Adaptive Threshold, but it allows tweaking the parameters of the adaptive threshold formula. The `defaultAdaptive` threshold option uses the following values:
          `numerator` = `0.5`
          `denominator` = `1`
          `exponent` = `1` (same effect as not used, i.e., no influence on the calculation)*Required when using `applyThresholdStrategy`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: defaultAdaptive # Using default adaptive threshold
```

###### vertical.memory.applyThresholdStrategy.percentage

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `percentage`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        The fixed percentage threshold to use. Value range: 0.01-2.5.

        *Required when type is `percentage`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: percentage # Using fixed percentage threshold
      percentage: 0.3 # 30% threshold
```

###### vertical.memory.applyThresholdStrategy.numerator

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `numerator`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        0.5
      </td>

      <td>
        Affects the vertical stretch of the threshold function. Lower values create smaller thresholds.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: customAdaptive # Using custom adaptive threshold
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

###### vertical.memory.applyThresholdStrategy.denominator

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `denominator`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        1
      </td>

      <td>
        Affects threshold sensitivity for small workloads. Values close to `0` result in larger thresholds for small workloads. For example, when `numerator` is `1`, `exponent` is `1` and `denominator` is `0` the threshold for `0.5` req. Memory will be 200%.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: customAdaptive
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

###### vertical.memory.applyThresholdStrategy.exponent

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `exponent`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        1
      </td>

      <td>
        Controls how quickly the threshold decreases for larger workloads. Lower values prevent extremely small thresholds for large resources.

        *Required when type is `customAdaptive`
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    applyThresholdStrategy:
      type: customAdaptive
      numerator: 0.1
      exponent: 0.1
      denominator: 2
```

##### vertical.memory.overhead

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `overhead`
      </td>

      <td>
        float
      </td>

      <td>
        No
      </td>

      <td>
        0.1
      </td>

      <td>
        Additional resource buffer when applying recommendations (0.0-2.5, e.g., 0.1 = 10%).

        If a 10% buffer is configured, the issued recommendation will have +10% added to it so that the workload can handle further increased resource demand.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    overhead: 0.1
```

##### vertical.memory.limit

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `limit`
      </td>

      <td>
        object
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Configuration for container memory limit scaling.

        The default behavior when not specified:

        * If `spec.containers[].resources.limits.memory` is _not_ defined on the workload, no limit is set by Workload Autoscaler.

        * If `spec.containers[].resources.limits.memory` _is_ defined on the workload, Workload Autoscaler calculates the limit using the following formula: `max(max(resources.requests.memory * 1.5, 128MiB), current.resources.limit)` (it will only increase the limit, but never lower it, and will enforce a minimum of 128MiB).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    limit:
      type: multiplier
      multiplier: 1.5
```

###### vertical.memory.limit.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Type of limit scaling to apply:

        * `noLimit` - Remove the resource limit from the workload definition entirely.

        * `multiplier` - Set limit as a multiple of requests using the following formula: `resources.requests.memory * {multiplier}`.

        * `keepLimits` - Preserve the existing limits as defined in the workload manifest without modification.

        * `maintainRatio` - Scale limits proportionally with recommended requests, preserving the original limit-to-request ratio from the workload manifest. If the original limit is 2× the request, the updated limit stays 2× the new recommended request. If the workload has no existing memory limit, a default 1.5× multiplier is applied to the recommended request. The 128MiB minimum memory limit applies.

        *If using the `vertical.memory.limit` configuration option, this field (type) becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    limit:
      type: keepLimits
```

###### vertical.memory.limit.multiplier

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `multiplier`
      </td>

      <td>
        float
      </td>

      <td>
        *Yes
      </td>

      <td>
        \-
      </td>

      <td>
        Value to multiply the requests by to set the limit on the workload. The calculation is: `max(resources.requests.memory * {multiplier}, 128MiB)`. The value must be greater than or equal to `1`. Note that the recommended limit will never be less than 128MiB.

        *Required when type is set to `multiplier`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memory:
    limit:
      type: multiplier
      multiplier: 1.5
```

##### vertical.memory.optimization

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `optimization`
      </td>

      <td>
        string
      </td>

      <td>
        no
      </td>

      <td>
        \-
      </td>

      <td>
        This configuration option can be used to disable memory management for workloads that benefit from CPU management only (e.g., Java workloads with a fixed heap size). The workload will then use memory requests/limits configured in the template.

        The only allowed option is off. If not set, the resource will inherit the workload management option as normal. Minimum required `workload-autoscaler` component version to use this feature is `v0.23.1`.
      </td>
    </tr>

  </tbody>
</Table>

```yaml YAML
vertical:
  memory:
    optimization: off
```

#### vertical.excludedContainers

Specifies container names to exclude from vertical autoscaling optimization. Excluded containers retain their current resource settings and are not scaled by the Workload Autoscaler. Recommendations are still generated and visible in event logs and the workload detail view.

| Field                | Type  | Required | Default | Description                                                                                                                                                   |
| -------------------- | ----- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `excludedContainers` | array | No       | -       | List of container names to exclude from optimization. Names must match exactly as defined in the pod specification. Regex or glob patterns are not supported. |

```yaml
vertical:
  excludedContainers:
    - istio-proxy
    - logging-agent
```

> 📘 Note
>
> `excludedContainers` targets containers explicitly defined in your pod spec, such as application containers or native Kubernetes sidecar containers.

#### vertical.predictiveScaling

| Field               | Type   | Required | Default | Description                               |
| :------------------ | :----- | :------- | :------ | :---------------------------------------- |
| `predictiveScaling` | object | No       | -       | Predictive scaling configuration for CPU. |

```yaml
vertical:
  predictiveScaling:
    cpu:
      enabled: true
```

##### vertical.predictiveScaling.cpu

| Field | Type   | Required | Default | Description                               |
| :---- | :----- | :------- | :------ | :---------------------------------------- |
| `cpu` | object | No       | -       | CPU-specific predictive scaling settings. |

```yaml
vertical:
  predictiveScaling:
    cpu:
      enabled: true
```

###### vertical.predictiveScaling.cpu.enabled

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `enabled`
      </td>

      <td>
        boolean
      </td>

      <td>
        No
      </td>

      <td>
        false
      </td>

      <td>
        Enable predictive scaling for CPU resources. When enabled, the system forecasts CPU usage based on historical patterns and generates proactive recommendations. Requires `vertical.optimization` to be set to `on`.

        Predictive scaling can be enabled on any workload via annotations, even if it is not currently eligible for scaling in this manner. The system will automatically activate predictive scaling once the workload becomes predictable and will seamlessly revert to standard scaling if the patterns are lost. This allows preemptive enablement without monitoring for eligibility.

        See <Anchor label="Predictive workload scaling" target="_blank" href="doc:workload-autoscaling-configuration#predictive-workload-scaling">Predictive workload scaling</Anchor> for more information.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  predictiveScaling:
    cpu:
      enabled: true
```

#### vertical.downscaling

| Field         | Type   | Required | Default | Description                    |
| :------------ | :----- | :------- | :------ | :----------------------------- |
| `downscaling` | object | No       | -       | Downscaling behavior override. |

```yaml
vertical:
  downscaling:
    applyType: immediate
```

##### vertical.downscaling.applyType

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `applyType`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        Default is taken from the vertical scaling policy controlling the workload.
      </td>

      <td>
        Override application mode:

        * `immediate` - Apply changes immediately
        * `deferred` - Apply without forcing restarts (in-place resizing on Kubernetes v1.33+, otherwise on natural pod lifecycle events)
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  downscaling:
    applyType: immediate
```

#### vertical.memoryEvent

| Field         | Type   | Required | Default | Description                     |
| :------------ | :----- | :------- | :------ | :------------------------------ |
| `memoryEvent` | object | No       | -       | Memory event behavior override. |

```yaml
vertical:
  memoryEvent:
    applyType: immediate
```

##### vertical.memoryEvent.applyType

This configuration option is fully compatible with other `applyType` options and is meant to be used in combination with them. This allows for fine-grained control over both upscaling and downscaling. Here's how they interact:

1. If both configuration options are set to the same value (both `immediate` or both `deferred`), the behavior remains unchanged.
2. If `vertical.downscaling.applyType` is set to `deferred` and `vertical.memoryEvent.applyType` is set to `immediate`:
   * Upscaling operations will be applied immediately.
   * Downscaling operations will be deferred to natural pod restarts.
3. If `vertical.downscaling.applyType` is set to `immediate` and `vertical.memoryEvent.applyType` is set to `deferred`:
   * Upscaling operations will be deferred to natural pod restarts.
   * Downscaling operations will be applied immediately.

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `applyType`
      </td>

      <td>
        string
      </td>

      <td>
        *Yes
      </td>

      <td>
        Default is taken from the vertical scaling policy controlling the workload.
      </td>

      <td>
        Override application mode for memory-related events (OOM kills, pressure):

        * `immediate` - Apply changes immediately
        * `deferred` - Apply without forcing restarts (in-place resizing on Kubernetes v1.33+, otherwise on natural pod lifecycle events)*If using the `vertical.memoryEvent` configuration option, this field becomes required.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  memoryEvent:
    applyType: immediate
```

#### vertical.containers.\{container\_name}

Configuration object that contains resource constraints for a specific container. Replace `{container_name}` with the name of your container.

| Field              | Type   | Required | Default | Description                                                                                                            |
| ------------------ | ------ | -------- | ------- | ---------------------------------------------------------------------------------------------------------------------- |
| `containers`       | object | No       | -       | Container configuration mapping.                                                                                       |
| \{container\_name} | object | No       | -       | Container resource configuration object. Has to be the name of the container for which resources are being configured. |

```yaml
vertical:
  containers:
    ? { container_name }
```

> 📘 Note
>
> Container constraints apply to application containers (`spec.containers[]`) and native sidecar containers (`spec.initContainers[]` with `restartPolicy: Always`). Traditional initContainers that run during pod startup are not optimized by Workload Autoscaler.

##### vertical.containers.\{container\_name}.resourceStrategy

Container-level sizing strategy. Sets a resource strategy for a specific container, overriding [vertical.resourceStrategy](#verticalresourcestrategy) for that container. Supports the same `type` and `nodeAllocatablePercentage` fields as the workload-level strategy. See [Node-aware DaemonSet sizing](woop-node-aware-daemonset-sizing.md#per-container-sizing).

Setting the strategy at the container level sizes only that container. It can also add requests to a container that has none in its pod spec. The workload-level strategy, by contrast, only applies to containers that already have requests.

| Field           | Type   | Required                                           | Default | Description                                                                                          |
| --------------- | ------ | -------------------------------------------------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `type`          | string | Yes                                                | -       | `nodeAllocatablePercentage`: size requests as a percentage of the node's allocatable CPU and memory. |
| `cpuPercent`    | float  | One of `cpuPercent` or `memoryPercent` is required | -       | Percentage of node-allocatable CPU to request. Valid range: `(0, 100]`.                              |
| `memoryPercent` | float  | One of `cpuPercent` or `memoryPercent` is required | -       | Percentage of node-allocatable memory to request. Valid range: `(0, 100]`.                           |

```yaml
vertical:
  containers:
    app:
      resourceStrategy:
        type: nodeAllocatablePercentage
        nodeAllocatablePercentage:
          cpuPercent: 30
          memoryPercent: 40
```

> 📘 Note
>
> Container `min`/`max` constraints and the resource strategy are per container, but limits and `lookBackPeriod` are workload-level only. Per-container `limit` and `lookBackPeriod` settings are not supported and are silently ignored.

##### vertical.containers.\{container\_name}.cpu

Container CPU constraints. Set minimum and maximum CPU limits for a specific container to define the workload autoscaler's scaling range. With the [`nodeAllocatablePercentage`](#verticalcontainerscontainer_nameresourcestrategy) strategy, these constrain (clamp) the resolved percentage for this container. The `percentageOfOriginal` type is not yet supported at container level — only `constant` is. The `constraints.min` and `constraints.max` blocks are optional, but if specified, both `type` and `value` are required.

| Field                   | Type   | Required | Default | Description                                                                                |
| ----------------------- | ------ | -------- | ------- | ------------------------------------------------------------------------------------------ |
| `constraints.min.type`  | string | Yes      | -       | Must be `constant`.                                                                        |
| `constraints.min.value` | string | Yes      | -       | Kubernetes CPU notation (e.g., `"10m"`, `"1"`). Min cannot be greater than max.            |
| `constraints.max.type`  | string | Yes      | -       | Must be `constant`.                                                                        |
| `constraints.max.value` | string | Yes      | -       | Kubernetes CPU notation (e.g., `"1000m"`, `"2"`). Recommendations won't exceed this value. |

```yaml
vertical:
  containers:
    { container_name }:
      cpu:
        constraints:
          min:
            type: constant
            value: 10m
          max:
            type: constant
            value: 1000m
```

##### vertical.containers.\{container\_name}.memory

Container memory request constraints. Set minimum and maximum memory requests for a specific container to define the workload autoscaler's scaling range. With the [`nodeAllocatablePercentage`](#verticalcontainerscontainer_nameresourcestrategy) strategy, these constrain (clamp) the resolved percentage for this container. These values control requests only; memory limits are managed separately via [vertical.memory.limit](#verticalmemorylimit). The `percentageOfOriginal` type is not yet supported at container level — only `constant` is. The `constraints.min` and `constraints.max` blocks are optional, but if specified, both `type` and `value` are required.

| Field                   | Type   | Required | Default | Description                                                                                         |
| ----------------------- | ------ | -------- | ------- | --------------------------------------------------------------------------------------------------- |
| `constraints.min.type`  | string | Yes      | -       | Must be `constant`.                                                                                 |
| `constraints.min.value` | string | Yes      | -       | Kubernetes memory notation (e.g., `"10Mi"`, `"2Gi"`). Recommendations will not go below this value. |
| `constraints.max.type`  | string | Yes      | -       | Must be `constant`.                                                                                 |
| `constraints.max.value` | string | Yes      | -       | Kubernetes memory notation (e.g., `"2048Mi"`, `"4Gi"`). Recommendations will not exceed this value. |

```yaml
vertical:
  containers:
    { container_name }:
      memory:
        constraints:
          min:
            type: constant
            value: 10Mi
          max:
            type: constant
            value: 2048Mi
```

##### vertical.containers.\{container\_name}.jvm

Container-level JVM optimization settings. These override the workload-level `vertical.jvm` settings for a specific container, allowing you to enable or disable JVM optimization and auto-instrumentation on individual containers within a multi-container pod. For details on JVM optimization, see [JVM workload optimization](woop-jvm-optimization.md).

| Field | Type   | Required | Default | Description                                                                   |
| :---- | :----- | :------- | :------ | :---------------------------------------------------------------------------- |
| `jvm` | object | No       | -       | Container-level JVM heap optimization and auto-instrumentation configuration. |

```yaml
vertical:
  containers:
    { container_name }:
      jvm:
        enabled: true
        autoInstrument: true
```

###### vertical.containers.\{container\_name}.jvm.enabled

| Field     | Type    | Required | Default | Description                                                                                                                                                 |
| :-------- | :------ | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled` | boolean | No       | `false` | Enable JVM heap-based memory optimization for this specific container. Overrides the workload-level `vertical.jvm.enabled` setting for the named container. |

```yaml
vertical:
  containers:
    { container_name }:
      jvm:
        enabled: true
```

###### vertical.containers.\{container\_name}.jvm.autoInstrument

| Field            | Type    | Required | Default | Description                                                                                                                                                                                |
| :--------------- | :------ | :------- | :------ | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `autoInstrument` | boolean | No       | `false` | Enable automatic JMX agent injection for this specific container. Overrides the workload-level `vertical.jvm.autoInstrument` setting. Pods must be restarted for injection to take effect. |

```yaml
vertical:
  containers:
    { container_name }:
      jvm:
        autoInstrument: true
```

###### vertical.containers.\{container\_name}.jvm.optionsEnvVar

| Field           | Type   | Required | Default        | Description                                                                                                                                                                                                                            |
| :-------------- | :----- | :------- | :------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `optionsEnvVar` | string | No       | Auto-detection | The environment variable that receives the injected `-Xms`/`-Xmx` flags for this specific container. Overrides the workload-level `vertical.jvm.optionsEnvVar` setting. Without it, the Workload Autoscaler auto-detects the variable. |

```yaml
vertical:
  containers:
    { container_name }:
      jvm:
        optionsEnvVar: JAVA_OPTS
```

#### vertical.pod

Configuration object that contains resource constraints for the pod. Limits the **total** CPU and memory recommended requests for the whole pod. These totals are evaluated after per-container recommendations have been calculated. When the recommended pod total request would violate a pod-level constraint, the recommendation is not generated and the last applied recommendation (if any) remains in effect.

| Field | Type   | Required | Default | Description                              |
| ----- | ------ | -------- | ------- | ---------------------------------------- |
| `pod` | object | No       | -       | Pod-level vertical configuration object. |

> 📘 Note
>
> Pod-level recommendation totals are summed recommended requests for application containers (`spec.containers[]`) and native sidecar containers (`spec.initContainers[]` with `restartPolicy: Always`). Traditional initContainers that run during pod startup are not taken into consideration.

> ⚠️ Warning
>
> While a pod-level constraint is being violated, no new recommendation will be generated for the affected workload. The previously applied recommendation (if any) remains unchanged until the conflict is resolved. The scaling policy or the `pod` constraints may need to be adjusted to unblock generation of recommendations.

```yaml
vertical:
  pod:
    cpu:
      constraints:
        min:
          type: constant
          value: 500m
        max:
          type: constant
          value: 4000m
    memory:
      constraints:
        min:
          type: constant
          value: 128Mi
        max:
          type: constant
          value: 4Gi
```

##### vertical.pod.cpu

Pod CPU constraints. Set minimum and maximum CPU requests that can be recommended for this pod. If these constraints are violated, no new recommendation will be generated and the last applied recommendation will remain in effect. The `percentageOfOriginal` type is not yet supported at the pod level — only `constant` is. At least one of `constraints.min` and `constraints.max` blocks must be specified, and if specified, both `type` and `value` are required.

| Field                   | Type   | Required | Default | Description                                                                                    |
| ----------------------- | ------ | -------- | ------- | ---------------------------------------------------------------------------------------------- |
| `constraints.min.type`  | string | Yes      | -       | Must be `constant`.                                                                            |
| `constraints.min.value` | string | Yes      | -       | Kubernetes CPU notation (e.g., `"500m"`, `"1"`). Recommendations will not go below this value. |
| `constraints.max.type`  | string | Yes      | -       | Must be `constant`.                                                                            |
| `constraints.max.value` | string | Yes      | -       | Kubernetes CPU notation (e.g., `"4000m"`, `"4"`). Recommendations will not exceed this value.  |

```yaml
vertical:
  pod:
    cpu:
      constraints:
        min:
          type: constant
          value: 500m
        max:
          type: constant
          value: 4000m
```

##### vertical.pod.memory

Pod memory constraints. Set minimum and maximum memory requests that can be recommended for this pod. If these constraints are violated, no new recommendation will be generated and the last applied recommendation will remain in effect. The `percentageOfOriginal` type is not yet supported at the pod level — only `constant` is. At least one of `constraints.min` and `constraints.max` blocks must be specified, and if specified, both `type` and `value` are required.

| Field                   | Type   | Required | Default | Description                                                                                          |
| ----------------------- | ------ | -------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `constraints.min.type`  | string | Yes      | -       | Must be `constant`.                                                                                  |
| `constraints.min.value` | string | Yes      | -       | Kubernetes memory notation (e.g., `"128Mi"`, `"1Gi"`). Recommendations will not go below this value. |
| `constraints.max.type`  | string | Yes      | -       | Must be `constant`.                                                                                  |
| `constraints.max.value` | string | Yes      | -       | Kubernetes memory notation (e.g., `"4Gi"`). Recommendations will not exceed this value.              |

```yaml
vertical:
  pod:
    memory:
      constraints:
        min:
          type: constant
          value: 128Mi
        max:
          type: constant
          value: 4Gi
```

#### vertical.jvm

JVM-specific optimization settings for Java workloads. When enabled, the Workload Autoscaler uses actual JVM runtime data (heap usage, GC behavior, thread count) to generate memory recommendations instead of relying on standard Kubernetes container memory metrics. For a full explanation of JVM optimization, auto-instrumentation, prerequisites, and troubleshooting, see [JVM workload optimization](woop-jvm-optimization.md).

| Field | Type   | Required | Default | Description                                                  |
| :---- | :----- | :------- | :------ | :----------------------------------------------------------- |
| `jvm` | object | No       | -       | JVM heap-based memory optimization and auto-instrumentation. |

```yaml
vertical:
  jvm:
    enabled: true
    autoInstrument: true
```

##### vertical.jvm.enabled

| Field     | Type    | Required | Default | Description                                                                                                                                                                                                                                                                                                                                                          |
| :-------- | :------ | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `enabled` | boolean | No       | `false` | Enable JVM heap-based memory optimization for this workload. When enabled, the Workload Autoscaler uses JVM heap metrics to compute memory recommendations and injects `-Xms`/`-Xmx` flags. Requires JVM metrics to be available — either through `autoInstrument` or an existing Prometheus pipeline. See [prerequisites](woop-jvm-optimization.md#prerequisites). |

```yaml
vertical:
  jvm:
    enabled: true
```

##### vertical.jvm.autoInstrument

| Field            | Type    | Required | Default | Description                                                                                                                                                                                                                                                                                                                |
| :--------------- | :------ | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `autoInstrument` | boolean | No       | `false` | Enable automatic JMX agent injection for this workload. When enabled, the Workload Autoscaler injects a JMX Prometheus Java agent into Java containers at pod admission time. Pods must be restarted for injection to take effect. See [auto-instrumentation](woop-jvm-optimization.md#auto-instrumentation) for details. |

```yaml
vertical:
  jvm:
    autoInstrument: true
```

##### vertical.jvm.optionsEnvVar

| Field           | Type   | Required | Default        | Description                                                                                                                                                                                                                                                                                |
| :-------------- | :----- | :------- | :------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `optionsEnvVar` | string | No       | Auto-detection | The environment variable that receives the injected `-Xms`/`-Xmx` flags. By default, the Workload Autoscaler uses the highest-priority JVM options variable already present on the container and falls back to `JDK_JAVA_OPTIONS`. Set this field to pin the flags to a specific variable. |

```yaml
vertical:
  jvm:
    optionsEnvVar: JAVA_OPTS
```

#### vertical.scope

Scopes the workload's recommendations by a node label. When set, the Workload Autoscaler computes a separate recommendation for each group of pods on nodes that share the same label value. Without it, the workload gets a single recommendation. Currently supported for DaemonSet workloads.

| Field   | Type   | Required | Default | Description                                                                                                                                                |
| :------ | :----- | :------- | :------ | :--------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scope` | string | No       | -       | The node label key that scopes recommendations. Pods on nodes that share the same label value receive one shared recommendation. Supported for DaemonSets. |

```yaml
vertical:
  scope: topology.kubernetes.io/zone
```

#### vertical.hpaConverters

| Field           | Type  | Required | Default | Description                                                                                                                                                                                                                                                                      |
| :-------------- | :---- | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `hpaConverters` | array | No       | -       | List of HPA conversion strategies to apply when vertical optimization changes resource requests. Prevents percentage-based HPA utilization targets from drifting after Atlas adjusts requests. See [HPA converter](woop-configuration-settings.md#hpa-converter) for details. |

```yaml
vertical:
  hpaConverters:
    - type: AverageValueFromOriginalRequests
```

##### vertical.hpaConverters\[].type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        Yes
      </td>

      <td>
        \-
      </td>

      <td>
        The conversion strategy to use. Supported values:

        * `AverageValueFromOriginalRequests` — Converts `Utilization`-based HPA targets to `AverageValue` targets using the formula `original_request × (averageUtilization / 100)`. Applies to `Resource` and `ContainerResource` metrics for CPU or memory. External, Object, Pods, and already-absolute metrics are left untouched.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
vertical:
  hpaConverters:
    - type: AverageValueFromOriginalRequests
```

> 📘 Note
>
> The HPA converter applies to workloads that have their own HPA while Atlas handles only vertical optimization. If Atlas's horizontal autoscaling is also enabled for the workload, Atlas already manages the HPA directly and the converter is not needed. Requires Workload Autoscaler version v1.0.11 or later.

### rolloutBehavior

| Field             | Type   | Required | Default | Description                                                       |
| :---------------- | :----- | :------- | :------ | :---------------------------------------------------------------- |
| `rolloutBehavior` | object | No       | -       | Configuration for controlling how recommendations are rolled out. |

```yaml
rolloutBehavior:
  type: NoDisruption
  preferOneByOne: true
  delaySeconds: 120
```

#### rolloutBehavior.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        `-`
      </td>

      <td>
        Controls how recommendation updates are rolled out.

        `NoDisruption` Ensures zero-downtime updates for single-replica workloads by temporarily scaling to two replicas during updates. This prevents service interruptions when applying new resource recommendations.<br /><br />**Note**: Requires `workload-autoscaler` component version v0.35.3 or higher. This setting is incompatible with the `deferred` apply type - if you change policy to `deferred`, you must remove the `rolloutBehavior` setting. This setting is applicable only to Deployment resources with a single replica, whose rollout strategy allows for downtime.

        For comprehensive workload requirements, see [Configuring zero-downtime updates](woop-configuration-settings.md#/prerequisites).
      </td>
    </tr>

  </tbody>
</Table>

```yaml
rolloutBehavior:
  type: NoDisruption
```

#### rolloutBehavior.preferOneByOne

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `preferOneByOne`
      </td>

      <td>
        boolean
      </td>

      <td>
        No
      </td>

      <td>
        `false`
      </td>

      <td>
        When `true`, enables a one-by-one pod restart strategy for immediate mode recommendations. This applies recommendations sequentially, waiting for each pod to become healthy before proceeding to the next. This reduces disruption for workloads without Pod Disruption Budgets or with aggressive rollout strategies. When `false` or unset, Workload Autoscaler may evict multiple pods in parallel based on workload availability settings.<br /><br />
        **Note**: Requires `workload-autoscaler` component version v0.57.0 or higher. This setting applies only to immediate mode and has no effect when using the `deferred` apply type. Supported for replicated workloads (Deployments, StatefulSets, ReplicaSets, ArgoCD Rollouts).

        See <Anchor label="Rollout behavior" target="_blank" href="https://docs.atlas.example/docs/workload-autoscaling-configuration#/rollout-behavior">Rollout behavior</Anchor> for more information.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
rolloutBehavior:
  preferOneByOne: true
```

#### rolloutBehavior.delaySeconds

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `delaySeconds`
      </td>

      <td>
        integer
      </td>

      <td>
        No
      </td>

      <td>
        `-`
      </td>

      <td>
        Minimum number of seconds Workload Autoscaler waits between evictions when rolling out a recommendation. This delay is enforced regardless of whether `preferOneByOne` is enabled. Accepted range: 0–3600.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
rolloutBehavior:
  delaySeconds: 120
```

### horizontal

| Field        | Type   | Required | Default | Description                           |
| :----------- | :----- | :------- | :------ | :------------------------------------ |
| `horizontal` | object | No       | -       | Horizontal autoscaling configuration. |

When horizontal autoscaling is configured via annotations, Atlas creates and manages a native Kubernetes `HorizontalPodAutoscaler` (autoscaling/v2) resource in the cluster. For concepts, supported workload types, and compatibility details, see <Anchor label="Horizontal autoscaling" target="_blank" href="doc:horizontal-pod-autoscaling">Horizontal autoscaling</Anchor>.

```yaml
horizontal:
  optimization: true
  useNative: true
  minReplicas: 2
  maxReplicas: 10
  metrics:
    - type: resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
```

#### horizontal.optimization

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `optimization`
      </td>

      <td>
        boolean
      </td>

      <td>
        Yes*
      </td>

      <td>
        \-
      </td>

      <td>
        Controls whether Atlas actively manages horizontal scaling for this workload. Set to `true` for managed scaling or `false` for read-only mode (Atlas monitors but does not scale).

        *Required when the `horizontal` configuration block is present.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
horizontal:
  optimization: true
```

#### horizontal.useNative

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `useNative`
      </td>

      <td>
        boolean
      </td>

      <td>
        Yes*
      </td>

      <td>
        false
      </td>

      <td>
        Enables native Kubernetes HPA mode. When `true`, Atlas creates and manages a native `HorizontalPodAutoscaler` resource in the cluster.

        *Required to enable horizontal autoscaling.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
horizontal:
  useNative: true
```

#### horizontal.takeOwnership

| Field           | Type    | Required | Default | Description                                                                                                                                                                                                                                                                                              |
| :-------------- | :------ | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `takeOwnership` | boolean | No       | false   | When `true`, Atlas takes ownership of an existing native HPA on the workload and replaces its configuration with Atlas-managed settings. The existing HPA must not be managed by a third party such as KEDA. Only HPAs with CPU or memory utilization triggers are currently eligible for ownership. |

```yaml
horizontal:
  takeOwnership: true
```

#### horizontal.minReplicas

| Field         | Type    | Required | Default | Description                                             |
| :------------ | :------ | :------- | :------ | :------------------------------------------------------ |
| `minReplicas` | integer | Yes\*    | -       | Minimum number of pod replicas. Must be greater than 0. |

```yaml
horizontal:
  minReplicas: 2
```

#### horizontal.maxReplicas

| Field         | Type    | Required | Default | Description                                                                 |
| :------------ | :------ | :------- | :------ | :-------------------------------------------------------------------------- |
| `maxReplicas` | integer | Yes\*    | -       | Maximum number of pod replicas. Must be greater than 0 and ≥ `minReplicas`. |

```yaml
horizontal:
  maxReplicas: 10
```

#### horizontal.metrics

| Field     | Type  | Required | Default | Description                                                                                                                                                      |
| :-------- | :---- | :------- | :------ | :--------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `metrics` | array | Yes\*    | -       | Array of metric objects that define scaling triggers. At least one metric is required. Only `resource` metric type (CPU or memory) is supported via annotations. |

```yaml
horizontal:
  metrics:
    - type: resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
```

##### metrics\[].type

| Field  | Type   | Required | Default | Description                                                           |
| :----- | :----- | :------- | :------ | :-------------------------------------------------------------------- |
| `type` | string | Yes      | -       | The metric source type. Only `resource` is supported via annotations. |

##### metrics\[].resource.name

| Field  | Type   | Required | Default | Description                                 |
| :----- | :----- | :------- | :------ | :------------------------------------------ |
| `name` | string | Yes      | -       | The resource to monitor: `cpu` or `memory`. |

##### metrics\[].resource.target.type

| Field  | Type   | Required | Default | Description                                                                                                                                                                                                 |
| :----- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type` | string | Yes      | -       | How the target value is interpreted. Accepted values: `Utilization` (percentage of resource requests averaged across all pods), `AverageValue` (average raw value across all pods), or `Value` (raw value). |

##### metrics\[].resource.target.averageUtilization

| Field                | Type    | Required    | Default | Description                                                                                          |
| :------------------- | :------ | :---------- | :------ | :--------------------------------------------------------------------------------------------------- |
| `averageUtilization` | integer | Conditional | -       | Target average utilization percentage across all pods. Required when `target.type` is `Utilization`. |

```yaml
metrics:
  - type: resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

##### metrics\[].resource.target.averageValue

| Field          | Type   | Required    | Default | Description                                                                                                           |
| :------------- | :----- | :---------- | :------ | :-------------------------------------------------------------------------------------------------------------------- |
| `averageValue` | string | Conditional | -       | Target average value as a Kubernetes quantity string (e.g., `"100m"`). Required when `target.type` is `AverageValue`. |

##### metrics\[].resource.target.value

| Field   | Type   | Required    | Default | Description                                                                                            |
| :------ | :----- | :---------- | :------ | :----------------------------------------------------------------------------------------------------- |
| `value` | string | Conditional | -       | Target value as a Kubernetes quantity string (e.g., `"100m"`). Required when `target.type` is `Value`. |

#### horizontal.behavior

| Field      | Type   | Required | Default | Description                                                                                                             |
| :--------- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------- |
| `behavior` | object | No       | -       | Custom scaling behavior for scale-up and scale-down directions. Maps directly to the Kubernetes HPA v2 `behavior` spec. |

```yaml
horizontal:
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      selectPolicy: Max
      tolerance: "0.1"
      policies:
        - type: Pods
          value: 4
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      selectPolicy: Min
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
```

##### horizontal.behavior.scaleUp / horizontal.behavior.scaleDown

| Field       | Type   | Required | Default | Description                        |
| :---------- | :----- | :------- | :------ | :--------------------------------- |
| `scaleUp`   | object | No       | -       | Scale-up behavior configuration.   |
| `scaleDown` | object | No       | -       | Scale-down behavior configuration. |

The `scaleUp` and `scaleDown` objects share the same structure. The following fields apply to both.

###### behavior.direction.stabilizationWindowSeconds

| Field                        | Type    | Required | Default | Description                                                                                                                                                                |
| :--------------------------- | :------ | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `stabilizationWindowSeconds` | integer | No       | -       | The number of seconds the autoscaler looks back at previous scaling recommendations before acting. This prevents rapid fluctuations in replica count. Valid range: 0–3600. |

```yaml
behavior:
  scaleDown:
    stabilizationWindowSeconds: 300
```

###### behavior.direction.selectPolicy

| Field          | Type   | Required | Default | Description                                                                                                                                                                                                                  |
| :------------- | :----- | :------- | :------ | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `selectPolicy` | string | No       | -       | When multiple scaling policies are defined, determines which one to use. `Max` selects the policy allowing the most change, `Min` selects the most conservative, and `Disabled` prevents scaling in this direction entirely. |

###### behavior.direction.tolerance

| Field       | Type   | Required | Default | Description                                                                                                             |
| :---------- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------------------- |
| `tolerance` | string | No       | -       | A non-negative quantity value. Scaling is triggered only when the metric deviation exceeds this tolerance. Must be ≥ 0. |

```yaml
behavior:
  scaleUp:
    tolerance: "0.1"
```

###### behavior.direction.policies

| Field      | Type  | Required | Default | Description                                                    |
| :--------- | :---- | :------- | :------ | :------------------------------------------------------------- |
| `policies` | array | No       | -       | Array of scaling policy rules that control the rate of change. |

###### policies\[].type

| Field  | Type   | Required | Default | Description                                                                                                 |
| :----- | :----- | :------- | :------ | :---------------------------------------------------------------------------------------------------------- |
| `type` | string | Yes      | -       | The unit for the scaling rate limit: `Pods` (absolute count) or `Percent` (percentage of current replicas). |

###### policies\[].value

| Field   | Type    | Required | Default | Description                                               |
| :------ | :------ | :------- | :------ | :-------------------------------------------------------- |
| `value` | integer | Yes      | -       | Maximum amount of change allowed per period. Must be > 0. |

###### policies\[].periodSeconds

| Field           | Type    | Required | Default | Description                                                             |
| :-------------- | :------ | :------- | :------ | :---------------------------------------------------------------------- |
| `periodSeconds` | integer | Yes      | -       | Time window in seconds for the scaling rate limit. Valid range: 1–1800. |

```yaml
policies:
  - type: Pods
    value: 4
    periodSeconds: 60
  - type: Percent
    value: 100
    periodSeconds: 60
```

### (Deprecated) legacy horizontal fields

> 🚧 Deprecated
>
> The following fields apply only to <Anchor label="legacy horizontal scaling" target="_blank" href="https://docs.atlas.example/update/docs/workload-autoscaler-legacy-hpa-v1">legacy horizontal scaling</Anchor>. They are not compatible with native HPA mode (`useNative: true`). New configurations should use the fields documented in the [horizontal](#horizontal) section above. To migrate existing workloads, see <Anchor label="Migrate from legacy horizontal scaling" target="_blank" href="https://docs.atlas.example/update/docs/woop-hpa-how-to-migrate">Migrate from legacy horizontal scaling</Anchor>.

When using legacy Atlas horizontal scaling (without `useNative: true`), the `horizontal` block uses the following structure:

```yaml
horizontal:
  optimization: on
  minReplicas: 1
  maxReplicas: 10
  scaleDown:
    stabilizationWindow: 5m
  shortAverage: 3m
```

The `minReplicas` and `maxReplicas` fields work identically to the native HPA configuration documented above. The fields below are specific to legacy mode.

#### horizontal.optimization (legacy)

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `optimization`
      </td>

      <td>
        string
      </td>

      <td>
        Yes*
      </td>

      <td>
        \-
      </td>

      <td>
        Enable legacy horizontal scaling. Set to `"on"` to enable or `"off"` to disable.

        In native HPA mode, this field is a boolean (`true`/`false`) instead. See [horizontal.optimization](#horizontaloptimization).

        *Required when the `horizontal` configuration block is present.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
horizontal:
  optimization: on
```

#### horizontal.scaleDown.stabilizationWindow (legacy)

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `stabilizationWindow`
      </td>

      <td>
        duration
      </td>

      <td>
        *Yes
      </td>

      <td>
        "5m"
      </td>

      <td>
        Cooldown period between scale-downs. The duration needs to be [parsable](https://pkg.go.dev/time#ParseDuration) (e.g., `"3m"`, `"10m"`).

        *Required if the `horizontal.scaleDown` block is present.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
horizontal:
  scaleDown:
    stabilizationWindow: 5m
```

#### horizontal.shortAverage (legacy)

| Field          | Type     | Required | Default | Description                                                                                                                                                                           |
| :------------- | :------- | :------- | :------ | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `shortAverage` | duration | No       | "3m"    | Time period to average CPU metrics over before making scaling decisions. Valid range: 1–10 minutes. Not applicable to native HPA mode — this field is ignored when `useNative: true`. |

```yaml
horizontal:
  shortAverage: 3m
```

### containersGrouping

| Field                | Type  | Required | Default | Description                                                                       |
| :------------------- | :---- | :------- | :------ | :-------------------------------------------------------------------------------- |
| `containersGrouping` | array | No       | -       | Rules for grouping dynamically generated containers with similar naming patterns. |

```yaml
containersGrouping:
  - key: name
    operator: contains
    values: ["data-processor"]
    into: data-processor
```

#### containersGrouping.\[].key

| Field | Type   | Required | Default | Description                                                                                                          |
| :---- | :----- | :------- | :------ | :------------------------------------------------------------------------------------------------------------------- |
| `key` | string | \*Yes    | -       | The attribute used to match containers. Currently, only supports `name` which refers to the container name property. |

#### containersGrouping.\[].operator

| Field      | Type   | Required | Default | Description                                                                                      |
| :--------- | :----- | :------- | :------ | :----------------------------------------------------------------------------------------------- |
| `operator` | string | \*Yes    | -       | Defines how the `key` is evaluated against the `values` list. Currently, only supports contains. |

#### containersGrouping.\[].values

| Field    | Type  | Required | Default | Description                                                                                                              |
| :------- | :---- | :------- | :------ | :----------------------------------------------------------------------------------------------------------------------- |
| `values` | array | \*Yes    | -       | A list of string values used for matching against the `key` with the specified operator. Must contain at least one item. |

#### containersGrouping.\[].into

| Field  | Type   | Required | Default | Description                                                                 |
| :----- | :----- | :------- | :------ | :-------------------------------------------------------------------------- |
| `into` | string | \*Yes    | -       | The target container name into which matching containers should be grouped. |

**Example usage**

See [Container Grouping for Dynamic Containers](workload-autoscaling-configuration.md#container-grouping-for-dynamic-containers).

\*Required if the parent object is present in the configuration.

### schedule

| Field      | Type   | Required | Default | Description                                                                                                                              |
| :--------- | :----- | :------- | :------ | :--------------------------------------------------------------------------------------------------------------------------------------- |
| `schedule` | object | No       | -       | Controls whether a custom workload is treated as job-like (sporadic execution) or continuous (always running) for optimization purposes. |

```yaml
schedule:
  type: jobLike
```

#### schedule.type

<Table align={["left","left","left","left","left"]}>
  <thead>
    <tr>
      <th>
        Field
      </th>

      <th>
        Type
      </th>

      <th>
        Required
      </th>

      <th>
        Default
      </th>

      <th>
        Description
      </th>
    </tr>

  </thead>

  <tbody>
    <tr>
      <td>
        `type`
      </td>

      <td>
        string
      </td>

      <td>
        No
      </td>

      <td>
        \-
      </td>

      <td>
        Explicitly sets whether the custom workload should be treated as job-like or continuous:

        * `jobLike`– Treats the workload as a job with sporadic execution patterns. For confidence, it requires only 3 runs per configured look-back period and keeps recommendations in the cluster even when no pods are running.
        * `continuous`– Treats the workload as a long-running application. It requires more metric density for confidence.**Note**: This configuration is only applicable to custom workloads (those with the `workloads.atlas.example/custom-workload` label). It does not affect native Kubernetes Jobs or other standard workload types.
      </td>
    </tr>

  </tbody>
</Table>

```yaml
schedule:
  type: jobLike # For sporadic workloads like batch jobs
```

> 🚧 Important
>
> The `schedule.type` configuration only applies to custom job-like workloads identified by the `workloads.atlas.example/custom-workload` label. Native Kubernetes Jobs with this label are always treated as job-like, and standard workload types (Deployments, StatefulSets, etc.) are always treated as continuous.

> 🚧 Legacy Annotation Support
>
> For documentation on the legacy annotation format, which is now deprecated, see the [Legacy Annotations Reference](https://docs.atlas.example/atlas/docs/workload-autoscaler-legacy-annotations-reference) page .

## Migration guide

> 📘 Note
>
> The annotations V2 structure cannot be combined with deprecated annotations V1. When the annotation `workloads.atlas.example/configuration` is detected, the workload is considered to be configured by using that annotation and all other annotations starting with `workloads.atlas.example` will be ignored.

To migrate from v1 to v2 annotations:

1. Remove all individual legacy `workloads.atlas.example/*` annotations
2. Add the new `workloads.atlas.example/configuration` annotation
3. Move all settings into the YAML structure under the new annotation

For example, these v1 annotations:

```yaml
workloads.atlas.example/vertical-autoscaling: "on"
workloads.atlas.example/cpu-target: "p80"
workloads.atlas.example/memory-max: "2Gi"
```

Would become:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    optimization: on
    cpu:
      target: p80
    memory:
      max: 2Gi
```