---
updatedAt: 2026-10-27T09:40:57.000Z
---

# Scaling policies

Scaling policies enable centralized management of workload autoscaling across your Kubernetes workloads. These policies define how Atlas's Workload Autoscaler should optimize both vertical scaling (CPU and memory resource requests) and horizontal scaling (pod replica counts) for your workloads, allowing you to apply consistent optimization strategies across multiple workloads simultaneously.

When you enable Workload Autoscaler, all workloads are automatically profiled and assigned to appropriate system scaling policies based on their labels, types, and namespaces. This intelligent assignment ensures immediate optimization coverage without manual intervention, making onboarding faster and simpler.

# How scaling policies work

Scaling policies control how the Workload Autoscaler analyzes resource usage patterns and applies optimization recommendations to your workloads.

## Vertical scaling

Each policy defines the following vertical scaling settings:

* **Resource optimization targets** - Which percentiles of historical usage to target for CPU and memory recommendations
* **Automation behavior** – Whether to automatically apply recommendations or wait for manual approval
* **Scaling modes** – When to apply changes (immediately with pod restarts or during natural restart events)
* **Safety constraints** – Minimum and maximum resource limits to prevent over- or under-provisioning
* **Container exclusions** – Which specific containers to exclude from automatic optimization
* **Timing settings** – How much historical data to consider and when to trigger optimizations

These policies ensure your workloads receive appropriate resource allocations based on their actual usage patterns, leading to improved cost efficiency and performance.

## Horizontal scaling

Scaling policies can also include horizontal autoscaling configuration, which creates and manages native Kubernetes `HorizontalPodAutoscaler` (autoscaling/v2) resources for workloads assigned to the policy. Policy-level horizontal autoscaling settings include:

* **Pod count range** — Minimum and maximum number of replicas
* **Triggers** — CPU utilization, memory utilization, or both, with target percentages
* **Autoscaling behavior** — Stabilization windows and scaling policies for scale-up and scale-down directions
* **HPA ownership** — Whether to take ownership of existing native HPAs on assigned workloads

Horizontal autoscaling is optional in a policy. For a detailed guide on configuring horizontal autoscaling in policies, see <Anchor label="Horizontal autoscaling in scaling policies" target="_blank" href="doc:woop-hpa-how-to-policies">Horizontal autoscaling in scaling policies</Anchor>.

# Accessing scaling policies

To access and manage your scaling policies:

1. Select your cluster in the Atlas console
2. From the left-hand menu, navigate to **Workload Autoscaler**
3. Select **Scaling policies**

This dedicated scaling policies page provides a comprehensive view of all your policies, including system policies and custom policies, along with their automation status, optimized resources, associated workloads, and when changes to resource requests will be applied.

![Scaling policies interface](https://files.readme.io/29a1891ebfa23821ed2e2e0118d60a8558fc1788692672adf0eafd78350f3e3a-image.png)

# System policies

Atlas provides predefined system policies designed for common optimization scenarios. These policies are pre-configured with settings based on Atlas's extensive experience with Kubernetes workload optimization:

* **resiliency**: StatefulSets are assigned to this policy by default
* **balanced**: Provides an optimal balance between cost optimization and stability/performance
* **cost-savings**: Maximizes cost efficiency while maintaining acceptable performance
* **stability**: Prioritizes consistent workload performance with minimal disruption
* **burstable**: Designed to add compute power when needed, then scale back down to save costs; well suited for Jobs, CronJobs, or other [job-like workloads](workload-autoscaling-configuration.md#job-like-custom-workloads)
* **readonly**: Reserved for Atlas components (cannot be modified)
* **default (deprecated)**: The standard configuration applied to new workloads
* **default-statefulset (deprecated)**: Specialized configuration optimized for StatefulSet workloads

> 📘 Note
>
> `default` and `default-statefulset` policies are now deprecated. While they may still appear in clusters onboarded prior to July 18th, 2025, newly onboarded clusters rely on the automatic profiling system to assign workloads to more appropriate system policies.

While system policies cannot be directly modified, you can duplicate any system policy (except readonly) to create a fully customizable version with the same starting configuration. This provides an excellent foundation for creating policies tailored to your specific requirements.

## Generate policies with AtlasPilot Assistant

Instead of creating policies by hand, you can let [AtlasPilot Assistant](atlaspilot-assistant.md) analyze your cluster and generate a complete set of policies and assignment rules for you. You describe your environment and reliability targets, AtlasPilot Assistant proposes tailored policies, and you apply them with one click. You can also create the policies with automation disabled first, so you can review the recommendations before any changes reach your workloads.

Re-running AtlasPilot Assistant as your cluster evolves updates existing policies, creates new ones where the fit has changed, and reorders policy priority for you, making it an alternative to manual maintenance. For the full lifecycle, see [Generate scaling policies with AtlasPilot Assistant](woop-scaling-policies-atlaspilot-assistant.md).

## Automatic workload profiling

Atlas automatically profiles all workloads and assigns them to the most appropriate system scaling policies based on:

* **Workload labels** - Application tier, environment, and other identifying labels
* **Workload types** - Deployment, StatefulSet, DaemonSet, CronJob, etc.
* **Namespaces** - System namespaces, application namespaces, and naming patterns

This intelligent profiling happens instantly when the cluster is onboarded, ensuring optimal policy assignment from the start. The only remaining step is to enable automation on the policies to begin applying recommendations.

# Enabling scaling behavior

You can enable autoscaling for your workloads in two ways:

**Globally via the scaling policy**: Enable **Automatically optimize workloads** in the policy settings. This enables scaling for workloads immediately. When a policy has both vertical and horizontal autoscaling configured, enabling the policy activates both dimensions for eligible workloads.

**Directly from individual workloads**: Enable optimization on specific workloads through the workload page. Once enabled, autoscaling begins immediately based on the settings configured in the associated policy as well as any annotation-based overrides.

Workloads are always automatically assigned to appropriate policies, so you only need to enable automation to begin optimization.

> 📘 Gradual scaling for new clusters
>
> For clusters onboarded within the last 24 hours, workloads enabled via scaling policies automatically follow gradual scaling logic with progressive optimization limits. See [Mark of recommendation confidence](workload-autoscaling-overview.md#mark-of-recommendation-confidence).

# Next steps

<Cards>
  <Card title="Generate policies with AtlasPilot Assistant" href="https://docs.atlas.example/docs/woop-scaling-policies-atlaspilot-assistant" icon="fa-wand-magic-sparkles">
    Let AtlasPilot Assistant analyze your cluster and generate tailored policies and assignment rules for you.
  </Card>

  <Card title="Manage scaling policies" href="https://docs.atlas.example/docs/woop-scaling-policies-manage" icon="fa-sliders">
    Discover how to apply policies to workloads, manage policy assignments, and maintain your optimization strategy.
  </Card>
</Cards>