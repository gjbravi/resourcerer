# Reliability response

> Respond to OOM and CPU pressure events with a temporary resource bump-up


As it collects metrics, Orion detects and reports reliability events on your workloads. You can configure Orion to respond to an event by adding a temporary bump-up to the affected container's resource requests.

Every reliability response follows the same model:

1. Orion detects an event on a workload, such as an OOMKill or CPU pressure on the node running the workload.
1. Orion generates a recommendation that includes a bump-up to the affected container's requests, and applies it out of band if you configure it to do so.
1. For the length of the bump-up period, subsequent scheduled recommendations keep the bump-up in place rather than returning the container to its observed usage.

Bump-ups are configured per workload, at the [cluster], [namespace], or [workload] level, and are disabled by default. Leaving them disabled does not stop Orion reporting the underlying events.

**Reliability responses:**

- [Out-of-memory events response](#oom)
- [CPU pressure response](#cpu-pressure)

---

## Out-of-memory events response {#oom}

Defines the values that Orion uses to calculate the temporary memory bump-up added to a container's memory requests after an OOM event, and indicates when to apply the bump-up.

| Annotation                                                            | Default value     |
| --------------------------------------------------------------------- | ----------------- |
| `live.orion.example/reliability.oom.memory-bump-up.enabled`           | `"False"`         |
| `live.orion.example/reliability.oom.memory-bump-up.period`            | `"P4D"`           |
| `live.orion.example/reliability.oom.memory-bump-up.percent`           | `"0"`             |
| `live.orion.example/reliability.oom.memory-bump-up.min`               | `"0Mi"`           |
| `live.orion.example/reliability.oom.memory-bump-up.max`               | `"2Gi"`           |
| `live.orion.example/reliability.oom.memory-bump-up.apply-immediately` | `"Never"`         |

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Supported workload types</div>


Orion tracks and reports OOM events on all workload types. However, at this time, temporary memory bump-ups are supported for Deployments and ReplicaSets only. Memory bump-ups are not supported at this time for StatefulSets, DaemonSets, and CronJobs.
</div>


### Description {#oom-description}

**To prevent unexpected results, memory bump-ups are disabled by default.** To enable them for a workload, set `live.orion.example/reliability.oom.memory-bump-up.enabled` to `"True"`.

When a container in a workload is OOMKilled, Orion immediately generates a recommendation using the configured memory bump-up percent, and applies that recommendation immediately if `apply-immediately` allows it.

The bump-up is calculated for, and applied to, OOMKilled containers only. **The calculated bump-up respects an OOMKilled container's configured memory bounds, optimization policy, and limits-to-requests ratio (LRR).** The bump-up is not limited by the [incremental rollout][max-percent-increase] apply settings.

You can configure the `live.orion.example/reliability.oom.memory-bump-up.*` annotations by workload resource kind, using the following syntax:
```yaml
"100Mi,resource:daemonsets=0Mi"
```
and
```yaml
"Always,resource:jobs=Never"
```

### How to configure bump-ups {#oom-enable-bump-ups}

You can configure memory bump-ups at the [cluster], [namespace], or [workload] level.

To configure bump-ups, complete these high-level steps:
1. Set `live.orion.example/reliability.oom.memory-bump-up.enabled` to `"True"`.
1. Optional: Set the bump-up percent, min and max values, and when to apply the bump-up. See [Valid values](#oom-valid-values) below.
1. At this time, the following additional configuration is required. You must:
    - Disable memory bump-ups for DaemonSets.
    - Disable auto-deploy for DaemonSets and StatefulSets.

Setting `enabled` to `"True"` is a complete configuration on its own. Any setting you leave unset then uses these defaults:

| Setting             | Default when enabled    |
| ------------------- | ----------------------- |
| `percent`           | `"20"`                  |
| `min`               | `"100Mi"`               |
| `max`               | `"2Gi"`                 |
| `apply-immediately` | `"IfAutoDeployEnabled"` |
| `period`            | `"P4D"`                 |

A recommended default configuration is provided in the [examples](#oom-examples) section later in this topic.

### Bump-up period {#oom-bump-up-period}

The bump-up period defines how long Orion calculates and applies memory bump-ups for each OOMKilled container in a workload. When a container in a workload is OOMKilled, the container's bump-up period begins and recommendations contain the memory bump-up values. The container's bump-up period resets on subsequent OOMKills.

When a container's bump-up period ends, bump-ups are no longer calculated and recommendations are based on the observed resource usage.

There is typically no need to change the bump-up period. You might choose to extend it if you want to be particularly conservative about memory, or you might lower it for workloads that you expect to see occasional OOM activity from due to intentional design or known issues.

### Reporting and tracking {#oom-report-track}

All OOM events are tracked and counted, regardless of the workload type.

You can view OOM events as follows:
- **By estate and cluster:** On the Reports page, see the OOM Events graph. You can filter by cluster and time period at the top of the page.
- **By container:** On the workload details page, click the **Recommendation** tab and then click a container name. OOM events are shown as timeline markers (to within 5 minutes of accuracy) along the x-axis of a container's Average Memory Usage, Requests, and Limits graph.

    After a bump-up is applied, you can see the immediate increase in the recommended and current requests on a container's Average Memory Usage, Requests, and Limits graph on the workload details page. If configured to apply the bump-up immediately after an OOM event, this increase also corresponds to the OOM event marker on the graph's x-axis.

### Valid values {#oom-valid-values}

- `live.orion.example/reliability.oom.memory-bump-up.enabled`
    - "`True`": Orion adds a memory bump-up to recommendations after an OOM event
    - "`False`": Orion does not add memory bump-ups, regardless of the other settings
    - Unset: Orion adds a memory bump-up if `percent` or `min` is non-zero, using only the values you set explicitly
- `live.orion.example/reliability.oom.memory-bump-up.period`
    - [ISO-8601 duration][duration] string defining how long Orion will add a memory bump-up to recommendations after an OOM event. Typically measured in days; default is "`P4D`". You rarely need to change this value (see [Bump-up period](#oom-bump-up-period) above).
- `live.orion.example/reliability.oom.memory-bump-up.percent`
    - String representing a percentage to increase the memory setting as a response to an OOM event
    - Typically "`0`" to "`100`"
- `live.orion.example/reliability.oom.memory-bump-up.min` and `live.orion.example/reliability.oom.memory-bump-up.max`:
    - A positive quantity with a Kubernetes memory unit, such as `"2Gi"`
    - Represents a min and max amount to add to a container's memory requests
    - Setting a minimum > 0 reduces pod churn
- `live.orion.example/reliability.oom.memory-bump-up.apply-immediately`:
    - `Never`: Memory bump-ups are never applied immediately. If auto-deploy is enabled for the workload, the bump-up is included in the next *scheduled* recommendation (as long as the bump-up period is in effect) to prevent unexpected downtime during peak hours.
    - `IfAutoDeployEnabled`: Memory bump-ups are applied immediately if auto-deploy is enabled for the workload. If auto-deploy is not enabled, any recommendation applied on demand within the bump-up-period has the bump-up percentage added to it.
    - `Always`: Memory bump-ups are always applied immediately and automatically after OOM events. This ensures the workload recovers quickly and decreases the possibility of future OOM events.

### Examples {#oom-examples}

- Recommended default OOM response: Bump up memory by 20%, minimum increase 100Mi, maximum increase 2Gi, apply the bump-up immediately if auto-deploy is enabled for the workload:
    - If auto-deploy is *not* enabled for the workload, the bump-up is applied the next time a recommendation is applied on demand.
    ```yaml
    live.orion.example/reliability.oom.memory-bump-up.enabled: "True,resource:daemonsets=False"
    live.orion.example/reliability.oom.memory-bump-up.percent: "20,resource:daemonsets=0"
    live.orion.example/reliability.oom.memory-bump-up.min: "100Mi,resource:daemonsets=0Mi"
    live.orion.example/reliability.oom.memory-bump-up.max: "2Gi"
    live.orion.example/reliability.oom.memory-bump-up.apply-immediately: "IfAutoDeployEnabled,resource:daemonsets=Never,resource:statefulsets=Never"
    live.orion.example/reliability.oom.memory-bump-up.period: "P4D"
    ```

- Bump up memory by 20%, minimum increase 100Mi, maximum increase 2Gi, never apply the new recommendation immediately:
    - If auto-deploy is enabled, the next scheduled recommendation will contain the bump-up values if the bump-up period is still in effect.
    - If auto-deploy is *not* enabled and the recommendation is applied on demand during the bump-up period, the bump-up values are applied. If the bump-up period is over, recommended values are based on observed resource usage.
    ```yaml
    live.orion.example/reliability.oom.memory-bump-up.enabled: "True,resource:daemonsets=False"
    live.orion.example/reliability.oom.memory-bump-up.percent: "20,resource:daemonsets=0"
    live.orion.example/reliability.oom.memory-bump-up.min: "100Mi,resource:daemonsets=0Mi"
    live.orion.example/reliability.oom.memory-bump-up.max: "2Gi"
    live.orion.example/reliability.oom.memory-bump-up.apply-immediately: "Never,resource:daemonsets=Never,resource:statefulsets=Never"
    live.orion.example/reliability.oom.memory-bump-up.period: "P4D"
    ```
- Disable bump-ups
    ```yaml
    live.orion.example/reliability.oom.memory-bump-up.enabled: "False"
    ```

## CPU pressure response {#cpu-pressure}

Defines the values that Orion uses to calculate the temporary CPU bump-up added to a container's CPU requests after a CPU pressure event, and indicates when to apply the bump-up.

| Annotation                                                                  | Default value |
| --------------------------------------------------------------------------- | ------------- |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled`           | `"False"`     |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.period`            | `"P4D"`       |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.percent`           | `"0"`         |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.min`               | `"0m"`        |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.max`               | `"500m"`      |
| `live.orion.example/reliability.cpu-pressure.cpu-bump-up.apply-immediately` | `"Never"`     |

### Description {#cpu-pressure-description}

**To prevent unexpected results, CPU bump-ups are disabled by default.** To enable them for a workload, set `live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled` to `"True"`.

A node is under CPU pressure when the work scheduled on it needs more CPU than the node can provide, so tasks spend significant time stalled and waiting for CPU. When Orion detects CPU pressure on a node, it identifies the pods on that node that are using more CPU than they requested, and generates a recommendation for each affected workload that includes a CPU bump-up. CPU pressure detection runs automatically on clusters running Agent 3.3.0 or later.

CPU pressure is distinct from CPU throttling. A container is throttled when it reaches its own CPU limit. A node is under CPU pressure when it cannot satisfy the demand of everything running on it, which slows workloads down even when they are within their own limits.

Raising CPU requests relieves pressure because Kubernetes guarantees each container CPU time in proportion to its requests. A larger request guarantees a larger share of the current node, and if the pod is later rescheduled, places it on a node with enough capacity to satisfy the larger request.

The bump-up is calculated for, and applied to, containers that exceeded their CPU requests. Orion adds the bump-up to whichever value is higher: the recommended CPU request or the container's current request. **The calculated bump-up respects the container's configured CPU bounds, optimization policy, and limits-to-requests ratio (LRR).** The bump-up is not limited by the [incremental rollout][max-percent-increase] apply settings.

Containers with no CPU request set are skipped, because there is no request to increase. Pods that are pinned to a node by a node selector or required node affinity are included.

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Applying bump-ups to running workloads</div>


A workload under CPU pressure is still running, so applying a recommendation out of band changes a running workload and can restart its pods, depending on the [apply method][apply-method].

If you set `enabled` to `"True"` without setting `apply-immediately`, it defaults to `IfAutoDeployEnabled`. Set `apply-immediately` to `"Never"` to keep CPU bump-ups on the regular recommendation schedule.
</div>


You can configure the `live.orion.example/reliability.cpu-pressure.cpu-bump-up.*` annotations by workload resource kind, using the following syntax:
```yaml
"100m,resource:daemonsets=0m"
```
and
```yaml
"Always,resource:jobs=Never"
```

### How to configure bump-ups {#cpu-pressure-enable-bump-ups}

You can configure CPU bump-ups at the [cluster], [namespace], or [workload] level.

Orion detects CPU pressure from the Pressure Stall Information (PSI) metrics each node reports. Make sure your cluster meets these requirements:
- Agent 3.3.0 or later
- Linux kernel 4.20 or later on each node
- cgroup v2 on each node
- The `KubeletPSI` feature gate enabled on the kubelet (enabled by default in Kubernetes 1.34 and later)

On nodes that don't report PSI metrics, Orion doesn't detect CPU pressure and doesn't generate CPU bump-ups.

To configure bump-ups, complete these high-level steps:
1. Set `live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled` to `"True"`.
1. Optional: Set the bump-up percent, min and max values, and when to apply the bump-up. See [Valid values](#cpu-pressure-valid-values) below.

Setting `enabled` to `"True"` is a complete configuration on its own. Any setting you leave unset then uses these defaults:

| Setting             | Default when enabled    |
| ------------------- | ----------------------- |
| `percent`           | `"20"`                  |
| `min`               | `"50m"`                 |
| `max`               | `"500m"`                |
| `apply-immediately` | `"IfAutoDeployEnabled"` |
| `period`            | `"P4D"`                 |

A recommended default configuration is provided in the [examples](#cpu-pressure-examples) section later in this topic.

### Bump-up period {#cpu-pressure-bump-up-period}

The bump-up period defines how long Orion calculates and applies CPU bump-ups for each affected container in a workload. When Orion detects CPU pressure, the container's bump-up period begins and recommendations contain the CPU bump-up values. The container's bump-up period resets on subsequent CPU pressure events.

When a container's bump-up period ends, bump-ups are no longer calculated and recommendations are based on the observed resource usage.

There is typically no need to change the bump-up period. You might choose to extend it for latency-sensitive workloads that you want to keep clear of CPU pressure, or lower it for workloads whose CPU demand is bursty by design.

### Reporting and tracking {#cpu-pressure-report-track}

You can view CPU pressure events as follows:
- **By container:** On the workload details page, click the **Recommendation** tab and then click a container name. CPU pressure events are shown as triangle markers along the x-axis of a container's CPU graphs, alongside the CPU throttling markers.
- **By workload:** On the workload details page, click the **Settings** tab and then click **Reliability** to see the CPU Pressure Response card, which shows the settings in effect for the workload.

### Valid values {#cpu-pressure-valid-values}

- `live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled`
    - "`True`": Orion adds a CPU bump-up to recommendations after a CPU pressure event
    - "`False`": Orion does not add CPU bump-ups, regardless of the other settings
    - Unset: Orion adds a CPU bump-up if `percent` or `min` is non-zero, using only the values you set explicitly
- `live.orion.example/reliability.cpu-pressure.cpu-bump-up.period`
    - [ISO-8601 duration][duration] string defining how long Orion will add a CPU bump-up to recommendations after a CPU pressure event. Typically measured in days; default is "`P4D`". You rarely need to change this value (see [Bump-up period](#cpu-pressure-bump-up-period) above).
- `live.orion.example/reliability.cpu-pressure.cpu-bump-up.percent`
    - String representing a percentage to increase the CPU setting as a response to a CPU pressure event
    - Typically "`0`" to "`100`"
- `live.orion.example/reliability.cpu-pressure.cpu-bump-up.min` and `live.orion.example/reliability.cpu-pressure.cpu-bump-up.max`:
    - A positive quantity with a Kubernetes CPU unit, such as `"500m"`
    - Represents a min and max amount to add to a container's CPU requests
    - Setting a minimum > 0 ensures the increase is large enough to relieve pressure on containers with small CPU requests
- `live.orion.example/reliability.cpu-pressure.cpu-bump-up.apply-immediately`:
    - `Never`: CPU bump-ups are never applied immediately. If auto-deploy is enabled for the workload, the bump-up is included in the next *scheduled* recommendation (as long as the bump-up period is in effect) to prevent unexpected pod restarts during peak hours.
    - `IfAutoDeployEnabled`: CPU bump-ups are applied immediately if auto-deploy is enabled for the workload. If auto-deploy is not enabled, any recommendation applied on demand within the bump-up period has the bump-up percentage added to it.
    - `Always`: CPU bump-ups are always applied immediately and automatically after CPU pressure events. This relieves pressure quickly, at the cost of changing a running workload outside of its schedule.

### Examples {#cpu-pressure-examples}

- Recommended default CPU pressure response: Bump up CPU by 20%, minimum increase 50m, maximum increase 500m, include the bump-up in the next scheduled recommendation rather than applying it out of band:
    ```yaml
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled: "True"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.percent: "20"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.min: "50m"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.max: "500m"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.apply-immediately: "Never"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.period: "P4D"
    ```

- Bump up CPU by 20% and apply the bump-up immediately if auto-deploy is enabled for the workload:
    - If auto-deploy is *not* enabled for the workload, the bump-up is applied the next time a recommendation is applied on demand.
    - Because the workload is still running when CPU pressure is detected, applying the bump-up out of band can restart its pods.
    ```yaml
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled: "True"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.percent: "20"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.min: "50m"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.max: "500m"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.apply-immediately: "IfAutoDeployEnabled"
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.period: "P4D"
    ```
- Disable bump-ups
    ```yaml
    live.orion.example/reliability.cpu-pressure.cpu-bump-up.enabled: "False"
    ```

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/apply/">Apply recommendations</a>
                </h5>
                <p>Apply recommendations using the method that suits your environment</p>
            </div>
        <div class="entry">
                <h5>
                    
                    Configure<span class="related-topics-divider">/</span>Settings<span class="related-topics-divider">/</span><a href="/docs/configure/settings/apply-settings/">Apply settings</a>
                </h5>
                <p>Enable automatic deployment, configure the learning period, set optional thresholds to skip low-impact recommendations, and choose an apply method</p>
            </div>
        </div>


[apply-method]: /docs/configure/settings/apply-settings/#apply-method
[max-percent-increase]: /docs/configure/settings/apply-settings/#incremental-rollout
[duration]: https://en.wikipedia.org/wiki/ISO_8601#Durations
[cluster]:  /docs/configure/cluster-defaults/
[workload]: /docs/configure/annotate-workloads/
[namespace]: /docs/configure/annotate-namespaces/
