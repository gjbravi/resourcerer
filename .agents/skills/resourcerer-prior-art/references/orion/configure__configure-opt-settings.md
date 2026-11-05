# Getting started with configuring Orion

> The five most important optimization settings and a starter template


## Getting started

Orion performs best when it is properly configured with goals, constraints, and requirements appropriate for the workloads it is optimizing. One of the first steps you should take when deploying Orion is to **review the available configuration options, and adjust settings as needed to suit your organization’s requirements.**

Optimization settings can be applied as cluster-scoped defaults, namespace-scoped defaults, or directly to individual workloads. Making sure your cluster-scoped default values are aligned with your organization’s requirements first provides the most return for your effort. You can carve out exceptions to these settings for particular namespaces or workloads later.

Sometimes the simplest way to identify adjustments that should be made is to just start applying recommendations, then examine the results and make configuration adjustments until any undesirable behavior has been resolved. This guide, however, aims to minimize the time spent in that process by giving you **guidance on how to craft an informed initial optimization configuration suited to your organization’s goals, and tailored for workload types relevant to you.**

### The 5 most important optimization settings

Some optimization settings are more important than others. These are the top five options to learn about and consider setting cluster default values for when you’re first getting started with Orion.
- [Optimization goal](#optimization-goal)
- [Recommendation schedule](#rec-schedule)
- [Minimum and maximum bounds](#min-max-bounds)
- [Limit handling](#limit-handling)
- [Auto-deploy](#auto-deploy)

How each of these options are set can have a big impact on Orion’s behavior.

#### Optimization goal {#optimization-goal}

Orion’s goal setting, **which you can configure separately for CPU and memory**, is a hint to the recommendation engine about how you would like it to balance the competing objectives of reliability and cost savings when generating resource request recommendations.

The default goal is `Balanced`, which is a middle ground between `Savings` and `Reliability`. Orion always strives to ensure workloads have sufficient resource allocation for their observed actual usage. The `Reliability` goal puts more weight on the maximum observed usage when forecasting and recommending requests, while the `Savings` goal results in recommendations more tolerant of possible occasional bursting.

It's a common practice to select the `Savings` optimization goal for non-prod environments, the `Balanced` goal for production, and the `Reliability` goal for specifically identified mission-critical namespaces and workloads.

#### Recommendation schedule {#rec-schedule}

The `schedule` setting dictates how frequently Orion evaluates workloads and updates their resource recommendations.

Orion uses machine learning to forecast resource usage and create request recommendations that satisfy predicted resource usage for the upcoming _schedule period_ – the length of time between a recommendation being generated and the time the following recommendation is scheduled.

Pick a schedule that reflects how often you want to, or are willing to, automatically update resource requests for your workloads.
A `@daily` or `@weekly` schedule produces good results for most organizations while incurring minimal resizing overhead. Resizing overhead can be further reduced by configuring [auto-deploy thresholds][thresholds].

#### Minimum and maximum bounds {#min-max-bounds}

By default, Orion is not configured with any minimum or maximum bounds. Unless your organization prefers to set minimum or maximum bounds, recommendations are made based only on the usage observed for each workload, for each resource.

Sizing bounds let you define the upper and lower values for CPU and memory that are always allocated to even the smallest, nearly-idle workload, and ceiling values above which CPU and memory requests are never raised (at least not automatically).

Choosing a minimum bound ensures that you never receive recommendations with values lower than you’re comfortable with. Be aware that setting minimums which are too high may hinder Orion’s ability to reduce waste.

As a general starting point, we suggest trying out the following minimum bound values.
- Minimum CPU request: `10m`
- Minimum memory request: `32Mi`

It’s not critical to set maximums, but if you prefer to, we suggest setting them based on the largest node or workload size permitted in your organization. For example:
- Maximum CPU request: `15000m` (or `15`), if the largest node type in the cluster has 16 cores
- Maximum memory request: `54Gi`, if the largest node type in the cluster has 64Gi of memory

#### Limit handling {#limit-handling}

Different organizations have different policies about how to (or not) configure limits. As such, Orion can be configured to match whichever limits practices your organization adheres to.

Orion offers several different options, called Optimization Policies, for what to do with limits when it makes adjustments to request settings:
- `RequestsRaiseLimitsIfNeeded` (default): This policy uses an adaptive behavior that works well for most organizations. Orion adjusts its behavior based on how limits are configured on the workload already.
  - If no limits are set, Orion won’t set them.
  - If limits are set but they are high enough already, Orion will leave them at their existing values.
  - If limits are set but Orion wants to increase requests, it will raise the limit if needed to accommodate the higher requests.
  
  Notably, Orion will never lower limits when using this option.

  This policy and the following alternative options are described in more detail in the [Optimization policy][opt-policy] section of the Containers topic.

- `RequestsAndLimits`: Always set limits, according to limits configuration.
- `RequestsOnly`: Never set or change limits.
- `DoNotOptimize`: Indicates to Orion to leave a particular resource completely alone.

#### Auto-deploy {#auto-deploy}

Whether or not to automatically deploy recommendations, or apply recommended settings, is the linchpin of realizing Orion’s value.

When you first roll out Orion, auto-deploy is `Disabled` on all workloads. As you validate that the recommended settings look good, you can manually apply recommendations on canary workloads, then start enabling auto-deploy for workloads or namespaces one at a time until you’ve built up confidence in the tool.
If you would like, you could alternatively set `auto-deploy` to `Enabled` as a cluster default, and opt out individual namespaces or workloads instead of opting them in.

The rollout strategy for enabling the auto-deploy setting in particular is likely to be organization dependent. For now, the important thing to know is that this is effectively the “on” switch for Orion’s automated management mode. Be thoughtful in deciding when and where to turn it on first.

## Starter template {#cluster-defaults-template}

The following template covers the five key settings discussed in this guide. Review the example values and adjust for your environment before applying.

For goal-based templates (Balanced, Savings, Reliability), see [Cluster defaults templates][cluster-defaults-templates].

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |

    # Optimization goal
    live.orion.example/cpu.optimization-goal: "Balanced"
    live.orion.example/memory.optimization-goal: "Balanced"

    # Schedule
    live.orion.example/schedule: "@daily"

    # Minimum bounds
    live.orion.example/containers.cpu.requests.min: "10m"
    live.orion.example/containers.memory.requests.min: "32Mi"

    # Maximum bounds: align with your organization's maximums
    # live.orion.example/containers.cpu.requests.max: "16000m"
    # live.orion.example/containers.memory.requests.max: "32Gi"

    # Limit handling
    live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.memory.optimization-policy: "RequestsRaiseLimitsIfNeeded"

    # Auto-deploy
    live.orion.example/auto-deploy: "Disabled"
```

For steps to apply this template, see [Configure optimization defaults for the cluster][configure-clusters].

## Container-specific settings

In this guide so far, we’ve talked about default settings very generally. As you begin to refine your configuration, note that appropriate settings can vary container-to-container inside a Pod. 

Orion’s configuration permits you to adjust many optimization settings on a container-by-container basis. See the topics under [Configure optimization][config-topics] for more details.

[opt-policy]: /docs/configure/settings/cpu-and-memory/#opt-policy
[configure-clusters]: /docs/configure/cluster-defaults/
[cluster-defaults-templates]: /docs/installation/cluster-defaults-templates/
[thresholds]: /docs/configure/settings/apply-settings/#thresholds
[config-topics]: /docs/configure/
