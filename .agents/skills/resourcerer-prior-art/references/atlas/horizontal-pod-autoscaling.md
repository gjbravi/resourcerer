---
updatedAt: 2026-10-21T11:49:08.000Z
---

# Horizontal Pod Autoscaling

Atlas's Workload Autoscaler manages native Kubernetes `HorizontalPodAutoscaler` (autoscaling/v2) resources on your behalf, providing full Kubernetes HPA capability integrated with Atlas's workload optimization. This gives you a managed horizontal scaling solution that works in harmony with Atlas's vertical scaling to optimize your workloads across both dimensions.

Horizontal autoscaling automatically adjusts the number of pod replicas based on resource utilization metrics. By scaling the number of pods up or down in response to demand, it helps maintain application responsiveness during traffic spikes while preventing over-provisioning during quieter periods.

# How it works

When you enable horizontal autoscaling for a workload, Atlas creates a native Kubernetes `HorizontalPodAutoscaler` resource in your cluster. Configuration is managed through scaling policies, workload-level settings in the Atlas console, the REST API, or Kubernetes annotations. The `atlas-workload-autoscaler` component running in your cluster reconciles your configuration into a native HPA object, which Kubernetes then uses to make scaling decisions.

Because these are standard Kubernetes HPA objects, they are fully visible through `kubectl get hpa` and integrate with your existing monitoring and tooling.

Atlas's horizontal autoscaling provides:

* Full Kubernetes HPA v2 API support, including CPU and memory utilization targets
* Custom scale-up and scale-down behavior with stabilization windows, scaling policies, and select policy configuration
* Tolerance configuration per scaling direction
* The ability to take ownership of existing native HPAs and manage them through Atlas
* Automatic coordination between vertical and horizontal scaling through Atlas's <Anchor label="VPA and HPA optimization algorithms" target="_blank" href="doc:workload-autoscaler-hpa-vpa">VPA and HPA optimization algorithms</Anchor>
* Horizontal autoscaling can also be configured through the <Anchor label="Workload Optimization API" target="_blank" href="https://docs.atlas.example/reference/workloadoptimizationapi_updateworkloadv2">Workload Optimization API</Anchor>.

# System requirements

To use horizontal autoscaling, your cluster needs the following minimum component versions:

| Component                  | Version          |
| :------------------------- | :--------------- |
| atlas-workload-autoscaler | v0.44.4 or later |
| atlas-agent               | v0.60.0 or later |

# Supported workload types

| Workload kind  | Supported |
| :------------- | :-------- |
| Deployment     | Yes       |
| StatefulSet    | Yes       |
| ReplicaSet     | Yes       |
| Rollout (Argo) | Yes       |

# Configuration

Horizontal autoscaling can be configured at two levels, following the same hierarchy as vertical scaling in Atlas.

**Scaling policies** define horizontal autoscaling settings that apply to all workloads assigned to the policy. This is the recommended approach for managing horizontal scaling at scale, as it provides consistent behavior across groups of workloads. Policies also support a take-ownership option that lets Atlas assume management of existing native HPAs on workloads assigned to the policy.

**Workload-level settings** allow you to override policy defaults for individual workloads. Any change to horizontal autoscaling settings at the workload level creates a full configuration override, meaning the workload's HPA is no longer inherited from its policy. You can reset overrides to return to the policy-level configuration.

> 📘 Overrides
>
> Horizontal autoscaling uses full-object overrides. When you modify any HPA setting at the workload level, the entire HPA configuration is overridden. This differs from vertical scaling, which supports field-level overrides. Workload-level overrides persist when a workload is moved between policies.

# Taking ownership of existing HPAs

If your workloads already have native Kubernetes HPAs that are not managed by a third-party controller (such as KEDA), Workload Autoscaler can take ownership of them to bring them under scaling policies. When you allow the transfer of ownership, Workload Autoscaler replaces the existing HPA's configuration with your Atlas-managed settings, making the scaling policy or workload configuration the single source of truth.

Ownership configuration is available at both the policy (Console UI, API, Annotations) and individual workload levels (API, Annotations). At the policy level, it applies to all eligible workloads assigned to that policy.

**Eligibility for taking ownership:**

* The existing HPA must not be managed by a third-party controller (such as KEDA). If a third-party owner is detected, Workload Autoscaler reports the incompatibility and skips the workload.
* Currently, only HPAs with CPU or memory utilization triggers are eligible.
* HPAs with external metrics or container-level targets are not supported for ownership in the current release.
* For workloads with third-party managed HPAs, see <Anchor label="KEDA compatibility" target="_blank" href="doc:woop-hpa-third-party-support">KEDA compatibility</Anchor> for information on how Workload Autoscaler coordinates with these controllers.

# Compatibility

## Vertical and horizontal scaling coordination

When both vertical and horizontal scaling are enabled for a workload, Workload Autoscaler automatically coordinates between the two to prevent conflicts and optimize resource allocation.

This includes replica target corrections for stable workloads, seasonality-aware optimization for workloads with cyclical usage patterns, and automatic CPU overhead calculations that maintain appropriate headroom below HPA scaling thresholds.

These optimizations work the same way regardless of whether horizontal scaling is configured through Atlas's managed HPA or an existing native Kubernetes HPA. For a detailed explanation of these algorithms, see <Anchor label="Vertical & horizontal workload autoscaling" target="_blank" href="doc:workload-autoscaler-hpa-vpa">Vertical & horizontal workload autoscaling</Anchor>.

## HPA converter

If your workload has its own HPA with percentage-based utilization targets and Atlas handles only vertical optimization, the HPA converter can convert those targets into fixed absolute values to prevent scaling drift after Atlas adjusts resource requests. See [HPA converter](woop-configuration-settings.md#hpa-converter) for details.

## Predictive scaling

Predictive scaling uses historical patterns to pre-emptively adjust vertical resource allocations. It is not compatible with horizontal autoscaling. When horizontal autoscaling is enabled for a workload, predictive scaling is automatically disabled for that workload.

## KEDA

Workload Autoscaler is compatible with KEDA `ScaledObjects`. When KEDA manages horizontal scaling, Workload Autoscaler still applies vertical scaling optimization and coordinates between the two dimensions. The level of optimization depends on which metrics are configured in the KEDA `ScaledObject`.

See <Anchor label="KEDA compatibility" target="_blank" href="doc:woop-hpa-third-party-support">KEDA compatibility</Anchor> for details.

# Migrating from legacy horizontal scaling

If you are using Atlas's legacy horizontal scaling (the previous proprietary implementation that does not create native Kubernetes HPA objects), you should migrate to the new native HPA-based implementation.

Migration preserves your existing `minReplicas` and `maxReplicas` settings and creates a native `HorizontalPodAutoscaler` resource for each workload.

For migration instructions, see <Anchor label="Migrate from legacy horizontal scaling" target="_blank" href="doc:woop-hpa-how-to-migrate">Migrate from legacy horizontal scaling</Anchor>.

> 📘 Legacy documentation
>
> Documentation for the previous horizontal scaling implementation is available at <Anchor label="Legacy horizontal scaling (deprecated)" target="_blank" href="doc:workload-autoscaler-legacy-hpa-v1">Legacy horizontal scaling (deprecated)</Anchor>.

# Next steps

<Cards>
  <Card title="Configure horizontal autoscaling" href="https://docs.atlas.example/docs/woop-hpa-how-to-configure" icon="fa-sliders">
    Set up horizontal autoscaling on individual workloads via the console, annotations, or API.
  </Card>

  <Card title="Horizontal autoscaling in scaling policies" href="https://docs.atlas.example/docs/woop-hpa-how-to-policies" icon="fa-layer-group">
    Configure horizontal autoscaling at the policy level with take-ownership support.
  </Card>

  <Card title="Annotations reference" href="https://docs.atlas.example/docs/workload-autoscaler-annotations-reference#horizontal" icon="fa-book">
    Complete field reference for all horizontal autoscaling annotation fields.
  </Card>
</Cards>