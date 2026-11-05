---
updatedAt: 2026-07-14T11:14:14.000Z
---

# Pod startup recommendations

Workload Autoscaler rightsizes CPU requests to a steady-state optimum. For workloads that need more CPU to initialize than to run at steady state, this creates two problems:

* **Slow startup** — the pod cannot get the CPU it needs to start quickly, because the lowered request limits what the scheduler reserves for it.
* **HPA spikes** — reduced CPU requests mean higher measured utilization. If an HPA is targeting CPU utilization, the spike during startup causes it to scale out replicas until the workload stabilizes.

Startup recommendations address both problems by generating a two-phase resource profile. A newly created pod receives the workload's original CPU requests during the startup period, then Workload Autoscaler transitions it to the optimized recommendation via [in-place pod resize](woop-in-place-resizing.md) once the startup period ends.

# How it works

A startup recommendation contains two phases:

* **Startup phase**: The pod receives the workload's original CPU requests — the values from the workload spec before optimization. This ensures the workload has sufficient CPU to initialize reliably.
* **Post-startup phase**: After the configured startup period elapses, Workload Autoscaler applies the optimized recommendation via in-place resize, without restarting the pod.

The startup period duration comes from the workload's configured [startup period](woop-configuration-settings.md#startup-metrics). For example, if a workload has a 5-minute startup period configured, the startup phase resources are applied for 5 minutes before the in-place resize to the optimized recommendation.

<Callout icon="📘" theme="info">
  ### CPU only

  Startup recommendations apply to CPU only. Memory requests remain the same across both phases.
</Callout>

When a startup recommendation is active, the `Recommendation` object in the cluster reflects both phases:

```yaml
apiVersion: autoscaling.atlas.example/v2
kind: Recommendation
metadata:
  name: nginx-deployment
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: nginx
  ...

  # startup phase: original workload CPU requests
  recommendation:
    - containerName: nginx
      requests:
        cpu: 300m
        memory: 300Mi

  startup:
    duration: "5m"

  postStartup:
    verticalHash: abc1e2323asda
    recommendation:
      - containerName: nginx
        requests:
          cpu: 150m
          memory: 150Mi
```

# Prerequisites

For startup recommendations to be generated:

* **Workload Autoscaler**: v0.63.1 or later
* **Kubernetes**: v1.33 or later — the cluster must support [in-place pod resizing](woop-in-place-resizing.md)
* **[Startup period](woop-configuration-settings.md#startup-metrics)**: The workload must have a startup period configured and the **Startup recommendations** mode selected. Without a startup period, no startup phase is generated.

To configure this in a scaling policy, navigate to **Advanced settings** and select **Startup recommendations** under **Startup metrics**:

<Image src="https://files.readme.io/c7dd15fd57d6d386a35eb2a155f367fc237ec050befea53129945de265a8741f-2026-07-14_13-45-27.png" alt="Startup metrics setting with Startup recommendations selected" align="center" caption="Select Startup recommendations in the vertical scaling policy settings" />

# Considerations

Enabling startup recommendations can reduce cluster bin-packing efficiency. Because newly scheduled pods temporarily hold their original (higher) CPU requests during the startup period, the cluster must reserve capacity for those values at scheduling time. After the startup period passes and the in-place resize completes, the reserved headroom is no longer needed — but it cannot be immediately reclaimed for other pods. This results in nodes that are slightly over-provisioned for the duration of each startup period.

# Behavior details

Startup recommendations are generated selectively. A startup phase is only added when it meaningfully differs from the post-startup recommendation:

| Scenario                                              | Behavior                                                                                                                                                               |
| ----------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Optimized CPU is higher than the original spec        | A regular single-phase recommendation is generated                                                                                                                     |
| Resources not defined in the original spec            | A regular single-phase recommendation is generated                                                                                                                     |
| Startup and post-startup CPU differ by less than 100m | Regular recommendation is generated — the difference is too small to warrant two phases                                                                                |
| OOMKill event                                         | Memory is unchanged between phases, so OOMKill handling is unaffected                                                                                                  |
| Surge detected                                        | The startup phase is added only if the startup CPU exceeds the post-startup recommendation by at least 100m                                                            |
| Mid-rollout change from two-phase to regular          | Startup-phase resources are updated to the regular recommendation values via standard reconciliation                                                                   |
| Mid-rollout change from regular to two-phase          | Pods still in the startup receive startup resources; pods past the startup period receive post-startup resources                                                       |
| In-place resize not possible                          | Falls back to a standard single-phase recommendation. This can occur when [JVM heap rightsizing](woop-jvm-optimization.md) is used alongside startup recommendations. |

<br />