---
updatedAt: 2026-10-12T09:11:33.000Z
---

# Workload Autoscaler configuration

Workload Autoscaling can be configured in different ways to suit your specific needs. You can configure scaling policies using the Atlas console, API, or Terraform. To override settings for individual workloads, use [workload annotations](workload-autoscaler-annotations-reference.md).

# Components

Workload Autoscaler consists of two Helm charts that work together to provide optimization capabilities:

* **atlas-workload-autoscaler** – The main component that generates and applies resource recommendations to your workloads
* **atlas-workload-autoscaler-exporter** – Collects workload metrics from your cluster to support recommendation generation, including [custom metrics](woop-custom-metrics.md) for JVM heap scaling, PSI pressure, and startup failure detection

Both components are installed automatically during cluster onboarding when Workload Autoscaler is selected. The exporter currently collects probe failure metrics by default.

# Upgrading

<HelmUpgradeMethodCallout />

<Callout icon="📘" theme="info">
  ### Note

  Some of the Helm commands below use the `--reset-then-reuse-values` flag, which requires Helm v3.14.0 or higher.
</Callout>

To upgrade all Atlas components, including Workload Autoscaler and its exporter:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent --reset-then-reuse-values
```

## Upgrade with Helm test enabled

To verify the deployment's functionality during an upgrade, you can enable Helm's built-in testing mechanism. This ensures both the reconciliation loop and admission webhook are working correctly:

<Callout icon="📘" theme="info">
  ### Configuration-only changes

  The commands in this section upgrade all components in the umbrella chart, not just Workload Autoscaler. To change configuration without upgrading component versions, pin the chart to your current version by adding `--version <your-current-chart-version>` to the command. Run `helm list -n atlas-agent` to check your current version.
</Callout>

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.test.enabled=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters (bare metal, on-prem, edge, GKE Autopilot), component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.test.enabled=true
```

  </Tab>
</Tabs>

## Upgrade with memory limits & Helm test enabled

For clusters with high workload counts or complex scaling requirements, you may need to adjust memory limits to prevent OOM kills. This command combines memory limit configuration with Helm testing:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.resources.limits.memory=5Gi \
  --set autoscaler.atlas-workload-autoscaler.resources.requests.memory=5G \
  --set autoscaler.atlas-workload-autoscaler.test.enabled=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.resources.limits.memory=5Gi \
  --set autoscaler-anywhere.atlas-workload-autoscaler.resources.requests.memory=5G \
  --set autoscaler-anywhere.atlas-workload-autoscaler.test.enabled=true
```

  </Tab>
</Tabs>

## Enable whitelisting mode

Whitelisting mode provides fine-grained control over which workloads are managed by Workload Autoscaler. When enabled, only workloads explicitly labeled for optimization will have recommendations applied.

To enable whitelisting mode:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.workloadWhitelistingEnabled=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm repo update atlas-helm
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.workloadWhitelistingEnabled=true
```

  </Tab>
</Tabs>

With whitelisting enabled, workloads require the `workload-autoscaler.atlas.example/enabled: "true"` label to have recommendations applied. See [Whitelisting mode](#whitelisting-mode) for details.

## Rollback to a previous version

If you encounter issues after an upgrade, you can roll back to a previous version using the following steps:

### Check release history

Before rolling back, check the current release history to identify available versions:

```shell
helm history atlas -n atlas-agent
```

### Rollback to a specific version

Once you've identified the desired version from the history, execute the rollback. Replace `[VERSION]` with the desired revision number:

```shell
helm rollback atlas [VERSION] -n atlas-agent
```

This will restore the Atlas release to the specified version.

### Verify the deployment

After an **upgrade** or **rollback**, confirm that the deployment is running correctly:

```shell
kubectl get pods -n atlas-agent
kubectl logs -l app=atlas-workload-autoscaler -n atlas-agent
```

### ArgoCD compatibility

If you are using ArgoCD and managing each component as a separate Helm release, you may need to disable Helm test hooks to prevent sync issues. The Workload Autoscaler includes a test hook that verifies its basic functionality, but this can cause problems with ArgoCD, as it does not fully support Helm test hooks. To disable test hook creation, set `test.enabled=false`.

For new installations:

```shell
helm install atlas-workload-autoscaler atlas-helm/atlas-workload-autoscaler \
  -n atlas-agent \
  --set test.enabled=false
```

For upgrades:

```shell
helm repo update atlas-helm
helm upgrade -i atlas-workload-autoscaler atlas-helm/atlas-workload-autoscaler \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set test.enabled=false
```

This option is available from version 0.1.75 (`appVersion` v0.27.3) onwards.

# Sidecar container support

Workload Autoscaler supports optimization of [Kubernetes native sidecar containers](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/), which are init containers with `restartPolicy: Always`. This enables optimization of containers that run alongside your main application containers, such as logging agents or other monitoring tools.

## Version requirements for sidecar support

To optimize sidecar containers, ensure you have the following minimum component versions:

| Component                  | Minimum Version                       |
| :------------------------- | :------------------------------------ |
| atlas-workload-autoscaler | v0.56.0 (helm chart 0.1.126 or later) |
| atlas-evictor             | helm chart 0.33.61 or later           |
| Kubernetes                 | v1.29+ (native sidecar support beta)  |

> :blue\_book: Note
>
> Kubernetes native sidecar containers became stable in v1.33 and are available in beta since v1.29. The SidecarContainers feature gate is enabled by default in Kubernetes v1.29+.

## What are sidecar containers?

Sidecar containers are special init containers that have `restartPolicy: Always` set:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp
  labels:
    app: myapp
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        optimization: on
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
        - name: myapp
          image: alpine:latest
          command: ['sh', '-c', 'while true; do echo "app running" >> /opt/data.txt; sleep 5; done']
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
          volumeMounts:
            - name: data
              mountPath: /opt
      initContainers:
        - name: sidecar
          image: alpine:latest
          restartPolicy: Always  # This makes the init container into a sidecar
          command: ['sh', '-c', 'tail -F /opt/data.txt']
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
          volumeMounts:
            - name: data
              mountPath: /opt
      volumes:
        - name: data
          emptyDir: {}
```

Common examples include Istio service mesh proxies (`istio-proxy` containers), logging agents that collect and forward application logs, or monitoring sidecars that collect metrics.

## How sidecar optimization works

This is how Workload Autoscaler detects sidecar containers in your workloads and optimizes them:

1. Workload Autoscaler automatically identifies containers in the `initContainers` section with `restartPolicy: Always`
2. Collects CPU and memory usage metrics for sidecar containers alongside main containers
3. Generates resource recommendations for both sidecar and main containers
4. Applies resource optimizations either during natural Pod restarts or immediately (including in-place resizing) based on your scaling policy
5. Works with Atlas Evictor and Rebalancer - when pods are moved or rebalanced, sidecar containers receive updated resource requests based on the latest recommendations

## Integration with service meshes

### Dynamically injected containers

By default, Workload Autoscaler optimizes containers that are injected during runtime via mutating admission webhooks (such as those injected by Istio's mutating webhook), even if they are not explicitly defined in the original pod specification.

Unlike native Kubernetes sidecar containers (defined in the pod spec as initContainers with `restartPolicy: Always`), dynamically injected containers are added by third-party mutating admission webhooks at pod creation time. This is why the default `atlas-workload-autoscaler` helm chart includes `webhook.reinvocationPolicy=IfNeeded` — it ensures Atlas's webhook runs after injection webhooks to properly handle the injected containers in their final state.

<Callout icon="💡" theme="default">
  ### Tip

  If you want to exclude specific application containers or native sidecar containers from optimization, you can use [container exclusions](woop-configuration-settings.md#container-exclusions) in your scaling policy to prevent named containers from being scaled. If you need to disable injected container optimization entirely, contact Atlas support.
</Callout>

### Istio integration

Starting with Istio v1.27, the `istio-proxy` sidecar uses native sidecar containers by default (init containers with `restartPolicy: Always`).

### Other service meshes

Atlas supports sidecar optimization for any service mesh or sidecar implementation that uses Kubernetes native sidecar containers with `restartPolicy: Always`.

## Predictive workload scaling

Predictive workload scaling adds another layer of automation to the Workload Autoscaler by using Machine Learning (ML) to forecast future resource needs based on historical usage patterns. Instead of reacting to usage changes after they occur, predictive scaling proactively adjusts resources before anticipated traffic increases or decreases.

### Benefits and limitations

Predictive scaling is designed for workloads with consistent, repeatable patterns. The system forecasts CPU usage up to 3 days in advance and updates predictions every 30 minutes.

**Key benefits:**

* Proactive scaling before traffic spikes
* Reduced over-provisioning during predictable low-traffic periods
* Automatic adaptation when workload patterns change
* 15% safety buffer to prevent under-provisioning

**Limitations:**

* **CPU only** – Memory and custom metrics are not yet supported
* **No HPA support** – Workloads with Horizontal Pod Autoscaling are excluded
* **Fixed safety buffer** – 15% overhead is non-configurable
* **Savings not yet tracked** – Predictive workload savings are not reflected in dashboard metrics

### How it works

Traditional vertical autoscaling operates reactively, adjusting resources based on historical usage patterns. While this approach works well for stable workloads, it presents challenges for workloads with predictable, time-based traffic patterns. Resources remain allocated during predictable usage dips, the system cannot preemptively scale before known traffic increases, and each predictable dip represents potential savings that reactive systems cannot capture.

<Image src="https://files.readme.io/eccbfebc004b10254fa2cc1803ea0880ae199293b2657d0179f8a984d73fff08-image.png" align="center" caption="Graph showing missed optimization opportunities" />

Predictive scaling addresses these issues through pattern-based forecasting. The system analyzes workload usage over time (with a minimum of 5 days of data required) to identify repeatable patterns, then uses these patterns to forecast CPU usage. Recommendations are generated based on predicted future usage rather than waiting for changes to occur, with a 15% safety buffer automatically applied.

When predictive scaling is enabled, CPU recommendations shift from reactive to proactive, while memory recommendations continue using standard logic. The system continuously evaluates whether workloads remain predictable and automatically reverts to standard behavior if patterns are lost, requiring no manual intervention.

### Prerequisites

Your workload must meet these requirements to be eligible for predictive scaling:

* **Predictable resource usage** – Must exhibit clear, repeatable patterns (consistent daily or weekly cycles)
* **Historical data** – Workload must run for at least 5 days to gather sufficient pattern data
* **Vertical scaling enabled** – Predictive scaling requires vertical scaling to be enabled
* **No horizontal scaling** – HPA must be disabled or have equal min/max replicas

While Kubernetes 1.33+ is not a hard requirement, it is highly recommended to utilize in-place resizing in conjunction with predictive scaling, which reduces the number of workload restarts in the cluster.

### Identifying eligible workloads

Predictive scaling is disabled by default and requires explicit opt-in. To identify workloads suitable for predictive scaling:

1. Navigate to **Workload autoscaler** > **Optimization**
2. Use the **Predictive scaling** quick filter to view eligible workloads

![](https://files.readme.io/a64bdada7e698d4708a9dc4d883585174c5b197beea81b1fc20993920c747678-image.png)

The Optimization overview includes a **Predictive scaling** widget showing:

* Total number of predictable workloads detected
* Count of workloads with predictive scaling enabled versus those eligible
* Quick action to enable predictive scaling for all eligible workloads

<Image src="https://files.readme.io/ba27a01afe1ec7813bd4632d9223fd58d01e98757484581095e302acc8a5f20a-image.png" align="center" width="75% " />

### Enabling predictive scaling

#### Via annotations

Predictive scaling can be enabled for any workload using annotations, even if it's not currently eligible. When enabled on a non-eligible workload, the system continues to use standard vertical scaling until the workload meets eligibility requirements. Once the workload becomes predictable, predictive scaling activates automatically without additional configuration. This allows you to enable predictive scaling preemptively rather than monitoring workloads for eligibility.

The system continuously monitors enabled workloads and automatically switches between predictive and standard scaling as patterns appear and disappear, requiring no manual intervention.

Add the predictive scaling configuration to your workload annotations:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-workload
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        optimization: on
        predictiveScaling:
          cpu:
            enabled: true
spec:
  # ... rest of deployment spec
```

<Callout icon="📘" theme="info">
  ### Important

  The `vertical.optimization: on` setting is required – predictive scaling cannot work without vertical optimization enabled.
</Callout>

#### Via Console UI

There are multiple ways to enable predictive scaling for a workload or multiple workloads in the Console. Here is an example:

1. Navigate to the workload page in Workload Autoscaler
2. Review the prediction graph showing forecasted versus current recommendations
3. Click **Enable** to turn on predictive scaling for the workload

![](https://files.readme.io/331dbd08e6e3c44902204602db76058a95e2fd0e2e9a56a6bc915d7b9565c119-image.png)

To enable predictive scaling in bulk:

1. Filter for predictable workloads using the **Predictable workloads** quick filter
2. Click **Enable for eligible workloads** in the predictive scaling widget to enable it for all eligible workloads, or alternatively, select the specific workloads you wish to enable it for post-filtering and click **Enable predictive scaling** above the table:

![](https://files.readme.io/36b1964c26be92241d205232181d05d1623f34b3bcbc3d8920dcb9d9ecf20593-image.png)

# Advanced configuration

Advanced configuration options provide fine-tuned control over Workload Autoscaler behavior for specific requirements. These settings are configured at the cluster level via Helm values and affect all workloads in the cluster.

## Whitelisting mode

Whitelisting mode allows you to explicitly control which workloads receive automated optimization. When enabled at the cluster level, Workload Autoscaler continues to generate recommendations for all eligible workloads but only applies them to workloads with the appropriate label.

### How it works

1. **Enable cluster-wide**: Set `workloadWhitelistingEnabled: true` in Helm values
2. **Label workloads**: Add `workload-autoscaler.atlas.example/enabled: "true"` to workload pods (`spec.template.metadata.labels`) and controllers (`metadata.labels`)

Recommendations are generated for all workloads, but only applied to labeled ones.

### Configuration

**Enable via Helm:**

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.workloadWhitelistingEnabled=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.workloadWhitelistingEnabled=true
```

  </Tab>
</Tabs>

**Label workloads for optimization:**

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  labels:
    workload-autoscaler.atlas.example/enabled: "true"
spec:
  template:
    metadata:
      labels:
        workload-autoscaler.atlas.example/enabled: "true"
    spec:
      containers:
        - name: app
          # container spec...
```

<Callout icon="📘" theme="info">
  ### Note

  Future releases will also respect this label for recommendation generation.
</Callout>

### When to use whitelisting mode

Whitelisting mode is recommended when you need fine-grained control over which workloads Workload Autoscaler optimizes. Common use cases include:

* **Gradual rollout** — Enable whitelisting cluster-wide and selectively label workloads as you validate optimization behavior, rather than optimizing all workloads at once.
* **Excluding specific workloads** — If most workloads should be optimized but a few should not (for example, workloads with strict resource requirements managed by another controller), enable whitelisting and label only the workloads you want optimized.
* **Testing in production** — Label a subset of workloads to evaluate Workload Autoscaler recommendations before expanding to the full cluster.

### Label workloads with kubectl

To label an existing workload for optimization without editing the manifest, use `kubectl label` on the controller:

```shell
kubectl label deployment my-app workload-autoscaler.atlas.example/enabled=true
```

You must also add the label to the pod template so that newly created pods are intercepted by the admission webhook. Edit the Deployment to add the label under `spec.template.metadata.labels`:

```shell
kubectl edit deployment my-app
```

### Verify whitelisting is active

To confirm that whitelisting mode is in effect, inspect the `MutatingWebhookConfiguration` and check for the whitelist label in the `objectSelector`:

```shell
kubectl get mutatingwebhookconfiguration atlas-workload-autoscaler -o yaml | grep -A3 "workload-autoscaler.atlas.example/enabled"
```

If whitelisting is active, the output shows the label with `operator: In` and `values: ["true"]`.

### Disable whitelisting mode

To disable whitelisting and resume optimization for all eligible workloads:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.workloadWhitelistingEnabled=false
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.workloadWhitelistingEnabled=false
```

  </Tab>
</Tabs>

When whitelisting is disabled, the `workload-autoscaler.atlas.example/enabled` label has no effect. All eligible workloads receive applied recommendations regardless of whether they carry the label.

<Callout icon="⚠️" theme="warn">
  ### Warning

  The label `workload-autoscaler.atlas.example/ignore` may appear in the `MutatingWebhookConfiguration` selectors, but it is not a supported configuration option. When present on the pod template (`spec.template.metadata.labels`) or on a namespace, it prevents the admission webhook from mutating pod resources. It has no effect when set only on the controller's own `metadata.labels`. If you need to control which workloads are optimized, use whitelisting mode instead. Remove any `workload-autoscaler.atlas.example/ignore` labels from your pod templates and namespaces, and use `workload-autoscaler.atlas.example/enabled` with whitelisting mode enabled.
</Callout>

### Version requirements

Minimum required component versions:

* `atlas-workload-autoscaler`: v0.58.3 (Helm chart 0.1.131 or later)

## ResourceQuota-aware scaling

ResourceQuota-aware scaling allows Workload Autoscaler to upscale workloads within the available Kubernetes ResourceQuota capacity in their namespace. When enabled, the autoscaler checks quota headroom before applying recommendations and uses an admission webhook as a secondary safeguard to prevent quota violations at pod creation time.

For details on how ResourceQuota constraints affect recommendations, see [ResourceQuota integration](#resourcequota-integration).

### Configuration

Enable ResourceQuota-aware scaling via Helm values:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.resourceQuotaAwareScaling.enabled=true \
  --set autoscaler.atlas-workload-autoscaler.resourceQuotaAwareScaling.headroomPercent=10
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.resourceQuotaAwareScaling.enabled=true \
  --set autoscaler-anywhere.atlas-workload-autoscaler.resourceQuotaAwareScaling.headroomPercent=10
```

  </Tab>
</Tabs>

| Value                                       | Type    | Default | Description                                                                                                                                                                                                                                                                                     |
| :------------------------------------------ | :------ | :------ | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `resourceQuotaAwareScaling.enabled`         | boolean | `false` | Enables ResourceQuota-aware scaling for the cluster. When `true`, the autoscaler checks namespace quota capacity before applying upscaling recommendations.                                                                                                                                     |
| `resourceQuotaAwareScaling.headroomPercent` | integer | `10`    | Percentage of the ResourceQuota hard limit to reserve as a buffer. The autoscaler will not recommend resource values that would use more than `100 - headroomPercent` percent of the quota. For example, the default of `10` means the autoscaler will not use more than 90% of the hard limit. |

### Tuning headroom

Start with the default headroom of 10%. If you observe quota violations in namespaces with many concurrently-scaling workloads, increase the headroom to provide a larger buffer. For example, a namespace with 20+ workloads scaling simultaneously may benefit from a headroom of 20–30%.

### Version requirements

Minimum required component version:

* `atlas-workload-autoscaler`: v0.82.0 (Helm chart version 0.1.170)

# Available workload settings

The following settings are available in Atlas scaling policies. To override these settings for individual workloads, use [annotations](workload-autoscaler-annotations-reference.md).

* **Automation** - on/off marks whether Atlas should apply or just generate recommendations. This is the only setting that can also be toggled per-workload via the Atlas console.
* **Scaling policy**- allows for the selection of policy names. It must be one of the policies available for a cluster.
* **Recommendation Percentile** - which percentile Atlas will recommend, looking at the last day of the usage. The recommendation will be the average target percentile across all pods spanning the recommendation period. Setting the percentile to 100% will no longer use the average of all pods but the maximum observed value over the period.
* **Overhead** - marks how many extra resources should be added to the recommendation. By default, it's set to 10% for memory and 10% for CPU.
* **Optimization Threshold** - When automation is enabled, how much of a difference should there be between the current pod requests and the new recommendation so that the recommendation can be applied immediately? Defaults to 10% for both memory and CPU.
* **Workload Autoscaler constraints** - sets the minimum and maximum values for resources, which will dictate that Workload Autoscaler cannot scale CPU/Memory above the max or below the minimum limits. The limit is set for all containers.
* **Startup metrics** - controls how the startup period is handled in vertical scaling recommendations: include startup usage (regular), apply original CPU requests during startup then switch to the optimized value (startup recommendations), or exclude startup metrics from calculations (legacy ignore mode). Only affects vertical autoscaling decisions.
* **Look-back period** - defines a custom timeframe (between 3 hours and 7 days) the Workload Autoscaler uses to observe CPU and memory usage when calculating scaling recommendations. It can be set separately for CPU and memory.

<Callout icon="📘" theme="info">
  ### Note

  It is recommended to wait for a week before enabling Workload Autoscaling for "all workloads", so that the system has understanding how the resource consumption varies on weekdays and weekends.
</Callout>

## Recommendation apply type

| Setting Name | Description                                                                         | Possible Values         | Default Value |
| ------------ | ----------------------------------------------------------------------------------- | ----------------------- | ------------- |
| Apply type   | Controls how and when the Workload Autoscaler applies recommendations to workloads. | `immediate`, `deferred` | `immediate`   |

### Immediate mode

When set to `immediate`, the Workload Autoscaler proactively implements resource optimization. The system monitors your workloads and applies new resource recommendations when they exceed the configured thresholds. This approach prioritizes rapid resource optimization, automatically triggering pod restarts to implement the new allocations.

* Apply recommendations as soon as they exceed the configured thresholds
* Trigger pod restarts to implement the new resource allocations immediately

#### How immediate mode applies recommendations

When immediate mode applies a recommendation, the Workload Autoscaler **evicts pods** using the Kubernetes Eviction API. The pod's controller then creates a replacement pod, and the Atlas mutating admission webhook applies the updated resource requests to the replacement at creation time.

Because eviction is the mechanism:

* **Pod Disruption Budgets are honored**: The Kubernetes Eviction API enforces PDBs, so eviction is blocked when disrupting a pod would violate the budget.
* **Unhealthy pods are recovered without PDB checks**: Pods that are OOMKilled or in CrashLoopBackOff are already disrupted. Workload Autoscaler bypasses the Eviction API for these and deletes them directly, so recovery is not blocked by a PDB.
* **Single-replica workloads**: A single-replica Deployment will briefly have zero running pods between eviction and replacement. Enable **Zero-downtime updates** if this is unacceptable — it temporarily scales to two replicas before evicting the original.

#### Rollout behavior

When using immediate mode, you can control how Workload Autoscaler rolls out recommendations to your pods. Under **Advanced settings**, the **Rollout behavior** section groups all three controls: **Zero-downtime updates** for single-replica workloads, **Evict pods one by one** to minimize disruption during resource adjustments, and **Eviction delay** to pace how often Workload Autoscaler evaluates the next eviction.

<Image src="https://files.readme.io/3b54852f6216abab1c7b25cbaa4036672f58427023779be31f9aa3b30d8ea002-image.png" align="center" width="65%" />

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
5. Save the policy

**Prerequisites**

For zero-downtime updates to work effectively, your workloads must meet these requirements:

* This feature is only applicable to Deployment resources that support running as multi-replica
* Running with a single replica (replica count = 1)
* Deployment's rollout strategy allows for downtime
* It works with the immediate apply type and is not needed for deferred mode
* Must not use PersistentVolumeClaims (PVCs)
* It requires `workload-autoscaler` component version v0.35.3 or higher

For workloads with unacceptable interruptions, this setting provides a way to achieve continuous availability without permanently increasing replica counts and associated costs.

**When zero-downtime updates are unnecessary**

The zero-downtime updates feature temporarily scales your workload to two replicas during updates. However, this may not be necessary if your Deployment already uses `strategy.type: RollingUpdate` with default settings.

With Kubernetes defaults:

* `maxSurge: 25%` rounds **up** to 1 for single replicas
* `maxUnavailable: 25%` rounds **down** to 0 for single replicas

This configuration naturally creates a "surge-first" update pattern where Kubernetes launches a replacement pod before terminating the original.

Enable zero-downtime updates if:

* Your Deployment uses `strategy.type: Recreate`
* You have explicitly set `maxUnavailable` to a value that allows termination before replacement (e.g., `maxUnavailable: 1`)
* Your rollout strategy configuration causes downtime during updates

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

To use the one-by-one eviction feature, ensure you have the following minimum versions:

| Component                  | Minimum Version |
| :------------------------- | :-------------- |
| atlas-workload-autoscaler | v0.57.0         |
| Helm chart                 | 0.1.127         |

##### Considerations and limitations

When using one-by-one eviction, be aware of these considerations:

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

<Callout icon="📘" theme="info">
  ### High-churn annotation

  When [in-place resizing](woop-in-place-resizing.md) is enabled (the default on Kubernetes v1.33+), `autoscaling.atlas.example/recommendation-applied-at` updates on running pods each time a recommendation is applied — potentially every 30 minutes. Observability tools that index pod annotations as time-series labels (such as Prometheus, Grafana, Dash0, and similar collectors) may experience high label cardinality or broken time-series continuity as a result. To avoid this, either configure your collector to exclude this annotation from label collection, or [disable in-place resizing](woop-in-place-resizing.md#configuration) if annotation stability is a hard requirement.
</Callout>

### In-place resizing

<Callout icon="📘" theme="info">
  ### Kubernetes v1.33+ Enhancement

  On clusters supporting in-place pod resizing, both immediate and deferred modes can leverage this capability to reduce disruption during resource optimization.
</Callout>

When running on Kubernetes v1.33+ with in-place resizing enabled:

**Enhanced immediate mode:**

* Attempts in-place resizing first for compatible pods
* Minimizes pod restarts when resize policies and resource constraints allow
* Falls back to pod restart if in-place resizing fails or enters a pending state
* Reduces disruption during optimization when possible

**Enhanced deferred mode:**

* Attempts in-place resizing first, applying recommendations without waiting for a pod restart
* Skips containers requiring restarts for specific resources, falling back to natural pod lifecycle events for those

The apply mode setting works the same way, but the underlying implementation automatically uses the most appropriate method (in-place resizing or pod restart) based on cluster capabilities and pod configuration.

## Recommendation percentile

The recommendation percentile setting determines how conservatively the Workload Autoscaler allocates resources based on the workload's historical usage patterns. It defines how close recommendations should be to observed workload resource usage, with higher percentiles leading to more generous resource allocations.

The recommendation percentile represents the statistical threshold used when analyzing workload resource usage. For example:

* A **p80** (80th percentile) setting for CPU means the recommendation will ensure sufficient resources to handle 80% of all observed load scenarios.
* A **max** (100th percentile) memory setting means recommendations will account for the absolute highest observed memory usage.

<Image src="https://files.readme.io/461e992e2411b2be5154d00299a40cd77102b4d7b142f6826e6b30935f1d1559-image.png" align="center" width="50% " />

### Configuring percentiles

You can configure different percentile values for CPU and memory independently:

* **CPU Percentile**: Typically set at p80 (default), can range from p50 to max
* **Memory Percentile**: Typically set at max (default) or p99, rarely lower

The recommendation is calculated using the average target percentile across all pods spanning the recommendation period. When you set the percentile to "max" (100%), the system will use the maximum observed value over the period instead of the average across pods.

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
* When both Vertical Autoscaling and Horizontal Autoscaling are enabled, memory overhead can still be configured as usual, but CPU overhead settings are ignored as the system automatically balances vertical and horizontal scaling.

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

### LimitRange integration

The Workload Autoscaler respects Kubernetes LimitRanges when generating resource recommendations. This ensures that optimized workloads comply with namespace-level resource policies and remain schedulable.

**Supported LimitRange types:**

* **Container**: Fully supported for all constraint types (min, max, default, ratios)
* **Pod**: Not supported (coming in future releases)

When LimitRanges are present, the autoscaler may generate non-optimal recommendations to maintain compliance with the defined constraints. This trade-off prioritizes workload schedulability over absolute optimization.

<Callout icon="📘" theme="info">
  ### Note

  LimitRange constraints take precedence over Workload Autoscaler settings when conflicts arise. The most restrictive constraint is always applied.
</Callout>

### ResourceQuota integration

The Workload Autoscaler respects Kubernetes [ResourceQuotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/) when generating and applying resource recommendations. When ResourceQuota-aware scaling is enabled, and at least one ResourceQuota exists in a workload's namespace, the autoscaler constrains its recommendations to stay within quota limits. This prevents pod admission failures caused by recommendations that would exceed the namespace's available quota capacity.

ResourceQuota-aware scaling uses a best-effort approach with a configurable headroom buffer. The autoscaler checks available quota capacity before applying recommendations and reserves a percentage of the quota's hard limit as headroom. This headroom absorbs potential race conditions when multiple workloads in the same namespace scale concurrently.

#### How to enable

To enable ResourceQuota-aware scaling, see [ResourceQuota-aware scaling](#resourcequota-aware-scaling) in the Advanced configuration section.

#### How requests are handled

When a ResourceQuota defines `requests.cpu` or `requests.memory` constraints, the autoscaler allows upscaling requests up to the available quota headroom threshold. Downscaling requests below the current values is always allowed, regardless of the ResourceQuota configuration.

If a recommendation would push the namespace's total request usage beyond the headroom threshold, Workload Autoscaler caps the recommendation to stay within the available capacity.

#### How limits are handled

When a ResourceQuota defines `limits.cpu` or `limits.memory` constraints, Workload Autoscaler preserves or restores original limits to prevent pod admission failures. If a recommendation removes a limit (for example, due to a scaling policy configured with `noLimit`), the original limit from the workload spec is restored. If a recommendation increases a limit above its original value, the limit is capped at the original value. Downscaling to a value below the original value is always allowed.

#### Scope matching

The system evaluates ResourceQuota scopes and scope selectors to determine which quotas apply to a given workload. A ResourceQuota only constrains a workload when both are in the same namespace, and all scope conditions match. The following scopes are evaluated:

* **NotBestEffort / BestEffort**: Workloads managed by Workload Autoscaler always have resource requests set, so they match `NotBestEffort` quotas and do not match `BestEffort` quotas.
* **Terminating / NotTerminating**: Evaluated based on whether the workload has an active deadline set. Standard Deployments and StatefulSets match `NotTerminating`.
* **PriorityClass**: Evaluated against the workload's `priorityClassName` using the scope selector's operator (`In`, `NotIn`, `Exists`, `DoesNotExist`).

When a ResourceQuota uses multiple scope selectors, all selectors must match for the quota to apply (`AND` logic).

#### Visibility

When ResourceQuota corrections are applied to a recommendation, the details are recorded in the Workload Autoscaler Event Log. Open the event details to see which resources were adjusted and why in the recommendation. Each correction entry includes the recommended value, the original spec value, and the final applied value.

```json
{
  "message": "Resource quota corrections applied",
  "args": {
    "cpuRequestReason": "capped_to_original_due_to_resource_quota",
    "cpuRecommendedRequest": "200m",
    "cpuOriginalSpecRequest": "100m",
    "cpuFinalRequest": "100m",
    "memoryLimitReason": "preserved_due_to_resource_quota",
    "memoryRecommendedLimit": "not set",
    "memoryOriginalSpecLimit": "500Mi",
    "memoryFinalLimit": "500Mi"
  }
}
```

#### Sync cadence

ResourceQuotas are synced every 20 minutes per cluster, with near-immediate reactivity for quota changes within a namespace. Recommendations are regenerated every 30 minutes. Thus, when a ResourceQuota is added or modified, recommendations stabilize within approximately 50 minutes.

#### Limitations

* **Race conditions**: Multiple workloads in the same namespace can receive recommendations concurrently. The headroom percentage absorbs this in most cases, but namespaces with many concurrently-scaling workloads may need a higher headroom value.
* **HPA interaction**: The system does not apply corrections when HPA creates new pods. If a ResourceQuota is tight, new HPA-created pods may be rejected. However, because Workload Autoscaler respects the headroom threshold, quota capacity intended for horizontal scaling is preserved.
* **DaemonSets**: DaemonSets are excluded from ResourceQuota-aware scaling.

<Callout icon="📘" theme="info">
  ### Note

  For full optimization, Atlas recommends removing ResourceQuotas from namespaces where Workload Autoscaler is actively managing workloads. ResourceQuotas constrain the autoscaler's ability to right-size workloads for both cost and performance.
</Callout>

## Startup metrics

Some workloads, notably Java and .NET applications, consume more CPU during initialization than at steady state. Without configuration, Workload Autoscaler treats startup usage like any other sample, which can inflate recommendations and cause HPA spikes. The startup metrics setting controls how Workload Autoscaler handles resource usage observed during the startup period.

You can configure this setting in the Atlas console under the Advanced Settings of a vertical scaling policy:

<Image src="https://files.readme.io/c7dd15fd57d6d386a35eb2a155f367fc237ec050befea53129945de265a8741f-2026-07-14_13-45-27.png" alt="Startup metrics at the policy level" align="center" caption="Startup metrics at the vertical scaling policy level" />

Choose one of three modes and set the startup period duration (2–60 minutes, or a custom value):

| Mode                                      | Behavior                                                                                                                                                           |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Regular recommendations**               | Includes startup metrics in recommendation calculations like any other period. May inflate values for workloads with high startup usage.                           |
| **Startup recommendations** (recommended) | Applies the workload's original CPU requests during the startup period, then transitions to the optimized recommendation via in-place resize once the period ends. |
| **Ignore startup metrics** (legacy)       | Excludes startup metrics from recommendation calculations, preventing inflated values.                                                                             |

The startup metrics setting only affects vertical scaling recommendation calculations. Horizontal autoscaling (HPA) responds normally to resource usage during startup regardless of this setting.

You can also configure this setting via the [API](https://docs.atlas.example/reference/workloadoptimizationapi_createworkloadscalingpolicy) or Terraform.

For more information on startup recommendations, see [Pod startup recommendations](woop-startup-recommendations.md).

### Change sensitivity

Change sensitivity determines when the Workload Autoscaler applies resource recommendation changes to your workloads. The sensitivity represents the minimum percentage difference between current resource requests and new recommendations that triggers an update.

Workload Autoscaler offers two sensitivity options in the UI:

* **Percentage**: A fixed threshold value that applies equally to all workloads
* **Dynamic**: An adaptive threshold that automatically adjusts based on workload size

<Image src="https://files.readme.io/b2c079f2e1959be0c2b27e62b36ce07f6de9aa809bf5d19b0e9f3037fa73d023-image.png" align="center" width="70% " />

#### Dynamic vs percentage sensitivity

##### Percentage sensitivity

A percentage sensitivity applies the same fixed percentage to all workloads regardless of their size. For example, with a 10% sensitivity:

* A workload requesting 100m CPU will only be scaled if the new recommendation differs by at least 10m
* A workload requesting 10 CPUs will only be scaled if the new recommendation differs by at least 1 CPU

While this is straightforward, it can be less optimal for workloads of varying sizes.

##### Dynamic sensitivity (recommended)

The dynamic sensitivity automatically adjusts based on workload size, providing more appropriate scaling behavior:

* **For small workloads**: Uses a higher threshold percentage to prevent frequent, insignificant updates
* **For large workloads**: Uses a lower threshold percentage to enable meaningful optimizations

This helps prevent unnecessary pod restarts for small workloads while ensuring larger workloads are also efficiently optimized.

##### Recommendations

For most users, we recommend using the **Dynamic** sensitivity setting, as it provides appropriate thresholds across workloads of all sizes and requires no manual tuning or maintenance.

The **Percentage** sensitivity setting is best suited for workloads of similar sizes and behavior when they are managed by the same scaling policy.

#### Configuring change sensitivity in policies

You can configure change sensitivity through the scaling policy settings:

1. Navigate to **Workload Autoscaler** → **Scaling Policies**
2. Select an existing policy or create a new one
3. Click **When to apply changes** → **Advanced Settings** to expand additional options
4. Under **Change sensitivity**, select your preferred strategy:
   * **Percentage**: Enter a fixed percentage value
   * **Dynamic**: Use the adaptive algorithm (recommended option)

#### Dynamic Sensitivity Simulation

The UI includes a sensitivity simulation graph that shows how the dynamic threshold changes based on workload size. This visualization helps you understand how the sensitivity percentage varies as resource requests increase.

Key elements of the simulation:

* The X-axis represents resource requests (CPU or memory)
* The Y-axis represents the threshold percentage
* The curve shows how the threshold decreases as resource size increases

You can toggle between CPU and memory views using the dropdown.

![](https://files.readme.io/e779775a2685c62b082a4257d23abe3a43539b33fba2e5afa88960cc84f710cb-image.png)

The threshold value indicator shows the exact sensitivity percentage that would apply to a specific resource request amount. For example, in the screenshot above, a request of 0.1 CPU would have a threshold of 45.45%. Meaning it would need to change by that amount to trigger a change.

#### Advanced configuration options

While the UI offers simplified access to dynamic sensitivity settings, power users can access additional customization options through annotations. These advanced options include:

* Custom adaptive thresholds with configurable parameters
* Independent sensitivity settings for CPU and memory
* Fine-tuning of the adaptive algorithm formula

For details on these advanced options, refer to our [Annotations reference](workload-autoscaling-configuration.md#verticalcpuapplythresholdstrategy) documentation.

## Workload Autoscaler constraints

Workload Autoscaler constraints allow you to set minimum and maximum resource limits that the Workload Autoscaler will respect when generating and applying recommendations. These constraints act as guardrails to prevent resources from being scaled too low or too high, ensuring your workloads maintain appropriate resource boundaries.

These constraints work in conjunction with Kubernetes LimitRanges. When both are present, the most restrictive constraint is applied. LimitRange constraints take precedence over Workload Autoscaler constraints when conflicts arise.

Workload Autoscaler constraints define the allowed resource scaling range for each container managed by the autoscaler. By setting these constraints, you can:

* Prevent resources from being scaled below a functional minimum
* Limit maximum resource allocation to control costs
* Customize scaling boundaries for each container individually

<Image src="https://files.readme.io/94b9a3bd11ff6fdba893ea8e357beca053f1de62647f2ce24b2105d3d126ddcc-woop-constraints-percentage.png" align="center" width="80%" />

### Constraint types

Each constraint (min or max) can be set using one of two modes, selectable via the **Value** or **% of request** toggle in the UI:

| UI mode          | Annotation `type`      | Description                                                                              |
| ---------------- | ---------------------- | ---------------------------------------------------------------------------------------- |
| **Value**        | `constant`             | An absolute Kubernetes resource value (e.g., `0.5`).                                     |
| **% of request** | `percentageOfOriginal` | A percentage of the workload's original resource request (e.g., `90` = 90% of original). |

<Callout icon="📘" theme="info">
  ### Note

  **% of request** requires the workload to have defined resource requests. If a workload does not have defined requests, the constraint will be treated as having no limit set. This mode is not supported for custom workloads or injected containers.
</Callout>

### Configuration options

Constraints can be configured at three levels:

1. **Policy-level constraints**
   Policy-level constraints apply to all workloads associated with a scaling policy. These constraints serve as default guardrails for all containers managed by that policy.
   You can set:
   * Global minimum CPU and memory limits
   * Global maximum CPU and memory limits
2. **Container-level constraints**
   For more granular control, you can specify constraints for individual containers within a workload using [annotations](workload-autoscaler-annotations-reference.md). Container-specific constraints override policy-level constraints when both are defined.
   Container constraints include:
   * Minimum and maximum CPU resources
   * Minimum and maximum memory resources
3. **Pod-level constraints**
   You can also specify constraints for the entire workload pod using [annotations](workload-autoscaler-annotations-reference.md). Pod-specific constraints are evaluated after policy-level and container-level constraints have been evaluated.
   Pod constraints include:
   * Minimum and maximum CPU resources
   * Minimum and maximum memory resources

### Use cases

Workload Autoscaler constraints are valuable in multiple scenarios.

Setting minimum CPU/memory resources ensures business-critical services never scale below functional thresholds. Some applications (like JVM-based services) need a certain amount of memory to function, so setting a minimum for them will prevent performance issues on cold starts.

Setting maximum CPU/memory resources is typically done for reasons such as:

* **Cost control**: Limiting maximum resources prevents unexpected cost increases
* **Compliance**: Enforcing organizational policies about resource consumption limits

### Limitations

When configuring container-specific constraints, the container name must match exactly what's defined in the pod specification.

### Example

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

    <Callout icon="📘" theme="info">
      ### Note

      `percentageOfOriginal` is available at the global (`vertical.cpu` / `vertical.memory`) level only. Container-level constraints (`vertical.containers`) do not yet support `percentageOfOriginal`.
    </Callout>
  </Tab>
</Tabs>

This allows for precise control over scaling boundaries for each container within your pod.

<Callout icon="📘" theme="info">
  ### Note

  Container constraints apply to application containers (`spec.containers[]`) and native sidecar containers (`spec.initContainers[]` with `restartPolicy: Always`). Traditional initContainers that run during pod startup are not optimized by Workload Autoscaler.
</Callout>

## Look-back period

The look-back period defines the timeframe the Workload Autoscaler uses to observe CPU and memory usage when calculating scaling recommendations. This feature allows you to customize the historical data window used for generating recommendations, which can be particularly useful for workloads with atypical resource usage patterns.

You can configure the look-back period in the Atlas console under Advanced Settings of a vertical scaling policy:

<Image src="https://files.readme.io/bd2b0a8a6c2310de019469e9425f1612751ba823c0cbfce59bab9daf4b95b286-image.png" alt="Look-back period in Advanced Settings" align="center" width="100% " caption="Look-back period in Advanced Settings" />

1. Set the look-back period for CPU and memory separately.
2. Specify the duration in days (d) and hours (h). The minimum allowed period is 3 hours, and the maximum is 7 days.

This feature allows you to:

* Adjust the recommendation window based on your workload's specific resource usage patterns.
* Account for longer-term trends or cyclical resource usage in your applications.

You can configure this setting at different levels:

* **Policy-level**: Apply the setting to all workloads assigned to a specific scaling policy.
* **Individual workload-level**: Configure the setting for a specific workload using [annotations](workload-autoscaler-annotations-reference.md) to override policy-level settings.

The look-back period can also be configured via Annotations, the [API](https://docs.atlas.example/reference/workloadoptimizationapi_createworkloadscalingpolicy), or Terraform.

For sporadically running [job-like custom workloads](workload-autoscaling-configuration.md#job-like-custom-workloads) that do not generate enough usage data within 7 days, Atlas can apply an extended 30-day look-back period instead. See [Extended look-back period for sporadic workloads](workload-autoscaling-configuration.md#extended-look-back-period-for-sporadic-workloads).

### Choosing the correct look-back period

The optimal look-back period largely depends on your workload's resource usage patterns. Most applications benefit from a shorter look-back period of 1-2 days. This approach works particularly well for standard web applications, capturing daily usage patterns while maintaining high responsiveness to changes. Shorter periods enable more aggressive optimization and often lead to higher savings.

Some workloads, however, require longer observation periods of 3-7 days. Applications with significant differences between weekday and weekend usage patterns benefit from a 7-day period to capture these weekly variations. Batch processing jobs that run every few days need a look-back period that covers at least one full job cycle to prevent potential out-of-memory (OOM) situations.

Common use cases and recommended periods:

* **High-frequency trading or real-time applications**: 3-6 hours for rapid scaling response
* **Standard web applications**: 1-2 days capture daily patterns while maintaining responsiveness to changes
* **Batch processing jobs**: Set to cover at least one full job cycle to account for periodic resource spikes
* **Weekend-sensitive workloads**: 7 days to capture both weekday and weekend patterns
* **Variable workloads**: Start with 1-2 days and adjust based on observed scaling behavior

<Callout icon="💡" theme="default">
  ### Tip

  For workloads with variable or uncertain patterns, start with a shorter period and adjust based on observed behavior. The key is to match the look-back period to your application's actual resource usage patterns – whether that's daily consistency, weekly cycles, or periodic processing jobs.
</Callout>

## Resource-specific optimization

When configuring vertical scaling, you can enable or disable CPU and memory optimization independently while still receiving recommendations for both resources. Even when optimization is disabled for a resource, Workload autoscaler continues to generate recommendations but won't apply them automatically. This setting can be configured at the **vertical policy level**. To override it for individual workloads, use [annotations](workload-autoscaler-annotations-reference.md).

<Image src="https://files.readme.io/9b42788baf7835445fcfc86022d6ab9e6805c8f81561e67c9ef613b071b470c2-image.png" alt="Selective resource optimization controls in the vertical scaling policy settings" align="center" caption="Selective resource optimization controls in the vertical scaling policy settings" />

<Callout icon="📘" theme="info">
  ### Note

  At least one resource type must remain enabled - you cannot disable both CPU and memory optimization simultaneously.
</Callout>

### Version requirements

The minimum required `workload-autoscaler` component version to use this feature is `v0.23.1`.

### Configuration options

At the policy level, you can configure resource-specific optimization through:

* The Atlas console UI using the resource checkboxes
* The Atlas API or Terraform

To override this setting for individual workloads, use annotations:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    memory:
      optimization: off
```

For detailed reference information on Workload Autoscaler annotations, see [Configuration via annotations](workload-autoscaling-configuration.md#configuration-via-annotations).

### Container exclusions

You can exclude specific containers from automatic optimization within a scaling policy. Excluded containers retain their current resource settings and are not scaled by the Workload Autoscaler. Recommendations are still generated and visible in the event log and the workload detail view.

This setting is configured under the **Selective optimization** section of a vertical scaling policy, alongside the resource selection (CPU/Memory) described above.

<Image src="https://files.readme.io/48cc165e25e3db733b61b86897a07e959bbcef56bd0683bc4eefe91a36cfb064-image.png" align="center" caption="Container exclusion settings under Selective optimization" />

<Callout icon="📘" theme="info">
  ### Note

  `excludedContainers` targets containers explicitly defined in your pod spec, such as application containers or native Kubernetes sidecar containers. For information on dynamically injected containers, see [Sidecar container support](workload-autoscaling-configuration.md#sidecar-container-support).
</Callout>

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

## HPA converter

When Atlas adjusts a workload's CPU or memory requests, any existing HPA with percentage-based utilization targets can drift — a "50% CPU" target changes meaning when the underlying request changes. The HPA converter prevents this by converting utilization targets into fixed absolute values derived from the workload's original requests.

The converter is designed for workloads that have their own HPA while Atlas handles only vertical optimization. If Atlas's horizontal autoscaling is also enabled for the workload, Atlas already manages the HPA directly and the converter is not needed.

For a detailed explanation of the conversion formula, supported metric types, behavior, and monitoring, see [HPA converter](woop-configuration-settings.md#hpa-converter) in the available settings reference.

### Enabling the HPA converter

To enable via scaling policy:

1. Navigate to **Workload Autoscaler** → **Scaling policies**.
2. Edit an existing policy or click **Create scaling policy**.
3. Under the **Vertical Rightsizing** tab, scroll to **HPA converter**.
4. Enable the **Enable HPA converter** toggle.
5. Save the policy.

To enable via workload annotation:

```yaml
workloads.atlas.example/configuration: |
  vertical:
    optimization: true
    hpaConverters:
      - type: AverageValueFromOriginalRequests
```

For the full annotation reference, see [vertical.hpaConverters](workload-autoscaler-annotations-reference.md#verticalhpaconverters).

### Version requirements

| Component                    | Minimum version |
| :--------------------------- | :-------------- |
| `atlas-workload-autoscaler` | v1.0.11         |

# In-Place Pod Resizing

<Callout icon="📘" theme="info">
  ### Kubernetes v1.33+ Feature

  This feature requires Kubernetes v1.33+ with Atlas Workload Autoscaler v0.53.0 or later.
</Callout>

Atlas Workload Autoscaler leverages Kubernetes in-place pod resizing when available, minimizing disruption during resource adjustments. This feature reduces pod restarts in many scenarios while maintaining the same level of resource efficiency. When in-place resizing cannot be applied, the system gracefully falls back to the configured apply mode.

## Automatic enablement

When enabled, in-place resizing is automatically used when all conditions are met:

* Kubernetes cluster v1.33 or later
* Atlas Workload Autoscaler v0.53.0 or later
* Pods and containers support in-place resizing
* In-place resizing is enabled via Helm configuration

## Apply mode compatibility

In-place resizing works with both immediate and deferred apply modes.

### Immediate mode with in-place resizing

When using immediate mode on clusters that support in-place resizing:

1. **In-place attempt**: Workload Autoscaler first attempts to resize the pod in-place
2. **Restart when necessary**: Container restarts may occur when resize policies require it, resource constraints prevent immediate resizing, or changes cannot be applied in-place
3. **Graceful fallback**: If in-place resizing fails or enters a pending state, the system falls back to the traditional pod restart approach through controller updates
4. **Memory protection**: Memory limits are never decreased during in-place operations to prevent application issues

### Deferred mode with in-place resizing

For deferred mode, Atlas can optionally apply recommendations through in-place resizing during natural pod lifecycle events:

1. **Natural restart detection**: When pods restart for other reasons (deployments, scaling, etc.)
2. **In-place opportunity**: If the pod supports in-place resizing, Atlas attempts to apply recommendations without additional restart
3. **Restart scenarios**: Containers may still restart if their resize policy requires it or if the resource change cannot be applied in-place
4. **Additional limitations**: Containers with `RestartRequired` resize policy are skipped during deferred in-place operations to avoid unnecessary restarts

## Configuration options

Control in-place resizing behavior through Helm values:

<Tabs>
  <Tab title="Standard (EKS, AKS, GKE)">

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler.atlas-workload-autoscaler.inPlaceResizeEnabled=true \
  --set autoscaler.atlas-workload-autoscaler.inPlaceResizeDeferredEnabled=true
```

  </Tab>
  <Tab title="Anywhere clusters">

On Anywhere clusters, component configuration is passed through the `autoscaler-anywhere` prefix instead of `autoscaler`:

```shell
helm upgrade atlas atlas-helm/atlas \
  -n atlas-agent \
  --reset-then-reuse-values \
  --set autoscaler-anywhere.atlas-workload-autoscaler.inPlaceResizeEnabled=true \
  --set autoscaler-anywhere.atlas-workload-autoscaler.inPlaceResizeDeferredEnabled=true
```

  </Tab>
</Tabs>

## Compatibility and limitations

In-place resizing respects all Kubernetes limitations:

* **QoS class preservation**: Pods maintain their original Quality of Service class
* **Memory decrease restrictions**: Memory limits cannot be decreased without container restart
* **Container type restrictions**: Only regular containers and sidecar containers support in-place resizing
* **Platform limitations**: Windows pods, pods with swap memory, and pods with static resource policies are excluded

When in-place resizing cannot be used, Atlas Workload Autoscaler automatically falls back to the configured apply mode behavior (immediate or deferred pod restart).

# Custom workload support

## When a natively supported workload is skipped

Workload Autoscaler only mutates workloads when it is confident the change won't conflict with another controller. To stay safe, it ignores any Deployment or StatefulSet whose `metadata.ownerReferences` field is set, because the referenced owner may reconcile the object with custom logic (for example, an operator that resizes volumes, repairs nodes, or rewrites the spec on its own).

This commonly affects workloads managed by operators, such as:

* Elasticsearch (StatefulSets created by the ECK operator)
* Cassandra / k8ssandra (StatefulSets created by the K8ssandra operator)
* RabbitMQ, Kafka, and other stateful services driven by a custom controller
* Any in-house controller that owns its Deployments or StatefulSets

If a workload shows up in **Cost Monitoring → Workloads** but not in the Workload Autoscaler view, an owner reference is the most likely cause. Cost Monitoring only reads resource usage and is not affected by this rule; Workload Autoscaler needs to write to the object and avoids workloads whose parent might overwrite the change.

### Opting an owner-controlled workload in

If you've confirmed that autoscaling won't conflict with the parent controller — for example, the controller reconciles but doesn't touch `spec.template.spec.containers[*].resources`, or you've coordinated with the owning team — add the `workloads.atlas.example/custom-workload` label to the pod template to opt the workload into autoscaling. The full procedure is described under [Label-based workload selection](#label-based-workload-selection) below.

<Callout icon="📘" theme="info">
  ### The supported apply modes are still determined by the workload's controller kind

  Adding the custom workload label tells Workload Autoscaler to manage the workload, but the supported apply modes (Immediate vs. Deferred) are determined by the workload's controller kind and the agent version. See [Apply modes for custom workloads](#apply-modes-for-custom-workloads).
</Callout>

***

Workload Autoscaler supports the scaling of custom workloads through label-based selection. This allows autoscaling for:

* Bare pods (pods without controllers)
* Pods created programmatically (as Spark Executors or Airflow Workers).
* Jobs (**Note**: CronJobs are natively supported as of `workload-autoscaler` v0.44.3)
* Workloads with custom controllers not natively supported by Atlas
* Groups of related workloads that should be scaled together

Workload Autoscaler will track and scale these workloads based on resource usage patterns, applying the same autoscaling policies and recommendations as standard workloads.

## Label-based workload selection

To enable autoscaling for custom workloads, add the `workloads.atlas.example/custom-workload` label to the Pod template specification. This is crucial - the label must be present in the **Pod template**, not just on the controller.

The label value must conform to RFC 1123 DNS subdomain and label name restrictions [imposed by Kubernetes](https://kubernetes.io/docs/concepts/overview/working-with-objects/names/), meaning it must:

* Contain only lowercase alphanumeric characters or '-'
* contain at most 63 characters
* Start and end with an alphanumeric character
* Not contain underscores (\_) or other special characters

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    workloads.atlas.example/custom-workload: "my-custom-workload"
spec:
  containers:
    - name: app
```

Workloads with the same label value will be treated as a single workload for autoscaling purposes. The value acts as a unique identifier for the workload group.

Workloads are uniquely identified by their:

* Namespace
* Label value
* Controller kind

## Apply modes for custom workloads

Custom workloads support different recommendation apply modes depending on their controller type and `workload-autoscaler` version:

#### Immediate mode support

**Requirements:**

* `workload-autoscaler` version >= 0.41.2
* The Pod's direct parent controller must be either a ReplicaSet or a StatefulSet

Custom workloads support **immediate** apply mode when the Pod's direct parent controller is a ReplicaSet or StatefulSet, allowing recommendations to be applied immediately rather than waiting for natural pod restarts.

This supports controller hierarchies such as:

* Pod/s → ReplicaSet → Deployment → Custom resource
* Pod/s → StatefulSet → Custom resource

<br />

<Callout icon="📘" theme="info">
  ### In-place resizing support

  With Kubernetes 1.33+, clusters that have in-place resizing enabled can apply resource recommendations to custom workloads with reduced pod disruption, even when using immediate apply mode. This allows the Workload Autoscaler to attempt modifying resource requests on running pods directly, providing faster optimization when resize policies and resource constraints allow.
</Callout>

### Deferred mode fallback

Custom workloads automatically fall back to **deferred** mode when:

* The Pod's direct parent controller is not a ReplicaSet or a StatefulSet (e.g., bare Pods, Jobs, CronJobs)
* The `workload-autoscaler` version is older than v0.41.2

## Configuring custom workloads via annotations

Regarding custom workload discovery, Atlas configuration annotations must also be specified either in the Pod template specification or at the top level of the manifest.

The placement of the configuration annotations for each `Kind`:

* **For deployments**: Place annotations in the `spec.template.metadata.annotations`

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: custom-managed-app
spec:
  template:
    metadata:
      annotations:
        workloads.atlas.example/configuration: | # Configuration annotation in pod template spec
          scalingPolicyName: custom
          vertical:
            optimization: on
      labels:
        workloads.atlas.example/custom-workload: "custom-controlled-app"
    spec:
      containers:
        - name: app
          # Container spec...
```

* **For jobs**: Place both the custom workload label and all configuration annotations in the Pod template specification (`spec.template.metadata`):

```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: batch-job
spec:
  template:
    metadata:
      labels:
        workloads.atlas.example/custom-workload: "batch-processor"
      annotations:
        workloads.atlas.example/configuration: | # Configuration annotation in pod template spec
          scalingPolicyName: custom
          vertical:
            optimization: on
    spec:
      containers:
        # Container spec...
```

### Configuring autoscaling behavior

Both labels and annotations used to configure autoscaling behavior must be specified in the Pod template specification, not on the controller or running Pod.

Key points about label-based workload configuration:

* Workloads are grouped per controller kind (deployments and StatefulSets with the same label will be treated as separate workloads)
* For grouped workloads, the newest/latest matching controller's pod template configuration is used as the workload specification
* Only workloads with the `workloads.atlas.example/custom-workload` label will be discovered for custom workload autoscaling
* The label value must be unique for each distinct workload or group of workloads you want to scale together
* All configuration labels and annotations must be specified in the Pod template specification

#### Examples

Scale a bare pod:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    workloads.atlas.example/custom-workload: "standalone-pod"
  annotations:
    workloads.atlas.example/configuration: | # Configuration annotation in pod template spec
      scalingPolicyName: custom
      vertical:
        optimization: on
spec:
  containers:
    - name: app
      # Container spec...
```

Group related jobs:

```yaml
apiVersion: batch/v1
kind: Job
spec:
  template:
    metadata:
      labels:
        workloads.atlas.example/custom-workload: "batch-processors"
      annotations:
        workloads.atlas.example/configuration: | # Configuration annotation in pod template spec
          scalingPolicyName: custom
          vertical:
            optimization: on
    spec:
      containers:
        - name: processor
          # Container spec...
```

Scale workloads with custom controllers:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: custom-managed-app
  ownerReferences: # Custom controller resource 
    - apiVersion: customcontroller.example.com/v1alpha1
      kind: CustomResourceType
      name: custom-resource
      uid: abc123
      controller: true
spec:
  template:
    metadata:
      annotations:
        workloads.atlas.example/configuration: | # Configuration annotation in pod template spec
          scalingPolicyName: custom
          vertical:
            optimization: on
      labels:
        workloads.atlas.example/custom-workload: "custom-controlled-app"
    spec:
      containers:
        - name: app
          # Container spec...
```

The Workload Autoscaler will track and scale workloads with custom controllers based on resource usage patterns, applying the same autoscaling policies and recommendations as standard workloads, except:

* These workloads are only scaled vertically using Vertical Pod Autoscaling (VPA)
* Only the deferred recommendation mode is supported

### Container grouping for dynamic containers

Starting with `workload-autoscaler` v0.33.0, Atlas Workload Autoscaler supports grouping dynamically created containers with similar names into a single logical container for resource recommendation purposes. This is meant to be done for workloads that generate containers with auto-generated names containing a common pattern (typically in the form of a suffix), such as batch jobs or Tekton pipelines.

For upgrading your `workload-autoscaler` version to leverage this feature, see [upgrading instructions](workload-autoscaling-configuration.md#upgrading).

#### Container grouping annotation

To group containers with similar naming patterns, you can add the `containersGrouping` section to your workload configuration annotation:

```yaml
workloads.atlas.example/configuration: |
  containersGrouping:
    - key: name
      operator: contains
      values: ["data-processor"]
      into: data-processor
```

Once configured, the workload usage will appear stable in the Atlas console under one container name, making it easier to track resource usage patterns and apply consistent optimization recommendations.

##### Configuration options

For the most up-to-date information on workload autoscaling annotation configuration options, always refer to [Configuration via Annotations](workload-autoscaling-configuration.md#configuration-via-annotations).

| Field      | Type   | Required | Description                                                                                                              |
| ---------- | ------ | -------- | ------------------------------------------------------------------------------------------------------------------------ |
| `key`      | string | Yes      | The attribute used to match containers. Currently only supports `name` which refers to the container name property.      |
| `operator` | string | Yes      | Defines how the `key` is evaluated against the `values` list. Currently only supports `contains`.                        |
| `values`   | array  | Yes      | A list of string values used for matching against the `key` with the specified operator. Must contain at least one item. |
| `into`     | string | Yes      | The target container name into which matching containers should be grouped for workload optimization purposes.           |

#### Use case example

This feature is meant to be used for workloads that create containers with auto-generated names based on a common pattern. For example, a TaskRun in Tekton might create containers with names like:

* `step-data-processor-167588-0-0-m9cun-0`
* `step-data-processor-167589-0-0-m9cun-0`
* `step-data-processor-167590-0-0-m9cun-0`

<Callout icon="🚧" theme="warn">
  ### Important

  Without container grouping, the workload autoscaler would treat each of these as completely separate, thus interfering with metrics collection and making it difficult to establish meaningful resource usage patterns and generate accurate recommendations for scaling.
</Callout>

By using the container grouping feature, you can consolidate all these containers under a single logical name (e.g., `data-processor`), allowing the Workload Autoscaler to provide more accurate and stable resource recommendations.

#### Complete example

Here's a complete example showing how to configure container grouping for a Tekton TaskRun:

```yaml
apiVersion: tekton.dev/v1beta1
kind: TaskRun
metadata:
  name: example-taskrun
spec:
  # TaskRun specifications
  podTemplate:
    metadata:
      labels:
        workloads.atlas.example/custom-workload: "team-neptune-data-processing-batch-job"
      annotations:
        workloads.atlas.example/configuration: |
          containersGrouping:
            - key: name
              operator: contains
              values: ["data-processing-batch-job"]
              into: data-processing-batch-job
    # Other pod template specifications
```

In this example:

* The workload is identified by the custom workload label `team-neptune-data-processing-batch-job`.
* Any container with a name containing the string `data-processing-batch-job` will be grouped into a logical container named `data-processing-batch-job`.
* The Workload Autoscaler tracks resource usage and generates recommendations for this logical container group rather than for each individual container.

### Scaling job workloads

<Callout icon="📘" theme="info">
  ### Custom job-like workloads

  If you have custom workloads that behave like Jobs but use different Kubernetes kinds (e.g., Spark executors, GitLab runners), see [Job-like custom workloads](workload-autoscaling-configuration.md#job-like-custom-workloads) for configuration options.
</Callout>

The Workload Autoscaler can automatically manage resources for Job workloads when they're labeled as custom workloads. CronJobs, on the other hand, are supported and optimized natively by Workload Autoscaler.

To enable automatic scaling for your Jobs that is uniquely optimized for those types of workloads, add the custom workload label to your Job specifications:

```yaml Job
# For Jobs (custom workload configuration required)
apiVersion: batch/v1
kind: Job
spec:
  template:
    metadata:
      labels:
        workloads.atlas.example/custom-workload: "batch-processor"
      annotations:
        workloads.atlas.example/configuration: |
          scalingPolicyName: custom
          vertical:
            optimization: on
    spec:
      containers:
        - name: processor
          # Container spec...
```
```yaml CronJob
# For CronJobs (native support as of v0.44.3)
apiVersion: batch/v1
kind: CronJob
spec:
  schedule: "*/10 * * * *"
  jobTemplate:
    spec:
      template:
        metadata:
          annotations:
            workloads.atlas.example/configuration: |
              scalingPolicyName: custom
              vertical:
                optimization: on
        spec:
          containers:
            - name: processor
              # Container spec...
```

Once labeled, the Workload Autoscaler will uniquely recognize these workloads, begin tracking your Jobs' resource usage patterns, and generate scaling recommendations different from those for other workloads.

### Creating scaling policies for Jobs

Different types of Jobs benefit from different scaling configurations. For best results, set up separate vertical scaling policies to match your Job frequency patterns.

1. **Group similar Jobs together**
   Create dedicated scaling policies for Jobs with similar patterns. For example, put all your hourly Jobs under one policy and daily Jobs under another. That is because the optimal settings for scaling jobs of different frequencies will differ (e.g., have different look-back periods configured).

2. **Consider Job duration**
   Very short-lived Jobs (running for just seconds) might be better served with fixed resource allocations rather than automatic scaling.

<Callout icon="📘" theme="info">
  ### Note

  To reach full recommendation confidence, a job needs to run at least 3 times within the configured look-back period.
</Callout>

## Job-like custom workloads

Some custom workloads behave like Kubernetes Jobs even though they're created as Pods or other resource types by custom controllers. These workloads run sporadically, often for just a few minutes, then disappear from the cluster. Examples include Spark executors, GitLab runners, and workflow engines.

Atlas's Workload Autoscaler recognizes these patterns and treats job-like custom workloads differently from continuous workloads:

**Key differences for job-like workloads:**

* **Confidence building**: Only requires 3 runs within the look-back period
* **CRD retention**: Recommendation CRDs stay in the cluster for the full look-back period even when no pods are running

#### Auto-discovered job-like workloads

The Workload Autoscaler automatically identifies and treats the following custom workloads as job-like:

* **Tekton TaskRuns** - Identified by `tekton.dev/v1beta1.TaskRun` kind
* **Argo Workflows** - Identified by `argoproj.io/v1alpha1.Workflow` kind
* **GitLab runner pods** - Identified by the `job.runner.gitlab.com/id` annotation
* **Apache Spark workloads** - Identified by `spark-role=driver` or `spark-role=executor` labels

<Callout icon="📘" theme="info">
  ### Note

  Auto-discovered job-like workloads still require the `workloads.atlas.example/custom-workload` label to be managed by the Workload Autoscaler.
</Callout>

### Manually marking workloads as job-like

For workloads not automatically detected, you can explicitly mark them as job-like using the `schedule.type` configuration:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    workloads.atlas.example/custom-workload: "job-like-pod"
  annotations:
    workloads.atlas.example/configuration: |
      schedule:
        type: jobLike
      vertical:
        optimization: on
spec:
  containers:
    - name: app
      # Container spec...
```

You can also explicitly mark auto-discovered workloads as continuous if needed:

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    workloads.atlas.example/custom-workload: "continous-gitlab-runner"
  annotations:
    job.runner.gitlab.com/id: "123456"  # Would normally be auto-discovered as job-like
    workloads.atlas.example/configuration: |
      schedule:
        type: continuous  # Override auto-discovery
      vertical:
        optimization: on
spec:
  containers:
    - name: app
      # Container spec...
```

For more information on configuring job-like custom workloads via annotations, see the [Annotations reference](workload-autoscaler-annotations-reference.md#schedule).

### Extended look-back period for sporadic workloads

Job-like workloads that run sporadically — for example, daily or weekly batch jobs, Spark jobs, or GitLab pipelines — may not generate enough usage data within the standard configured look-back period (maximum 7 days) for a reliable recommendation. For these workloads, Atlas can analyze CPU and memory usage over an **extended 30-day look-back period** instead. The same extended window is also applied when checking job history, so sporadic jobs retain their recommendations between runs.

The extended look-back period:

* Is applied automatically to job-like workloads once enabled — both auto-discovered and those marked with `schedule.type: jobLike`. It does not apply to continuous workloads.
* Overrides the configured look-back period for both CPU and memory, and requires no other configuration changes.
* Is enabled per organization or cluster and is off by default. To enable it for your organization, contact Atlas support.

## GitLab CI Runner Integration

When using Atlas's workload autoscaling with GitLab CI runners, you must properly configure custom workload labels to enable optimization. Here's how to set it up.

### Configuring GitLab Runner

Add the custom workload label to your GitLab runner configuration in the `config.toml` file:

```toml
[[runners]]
  [runners.kubernetes]
    namespace = "{{.Release.Namespace}}"
    image = "alpine"
    [runners.kubernetes.pod_labels]
      # Use CI_JOB_NAME_SLUG and CI_PROJECT_ID for unique workload identification
      ["workloads.atlas.example/custom-workload"] = "${CI_JOB_NAME_SLUG}-${CI_PROJECT_ID}"
```

<Callout icon="📘" theme="info">
  ### Automatic job-like detection

  GitLab runner pods are automatically detected and treated as job-like workloads when they contain the `job.runner.gitlab.com/id` annotation and `workloads.atlas.example/custom-workload` label. This means they'll only need 3 runs within the configured look-back period to gain full confidence, making them easier to optimize.
</Callout>

### Important Considerations

These are some of the considerations around best practices when implementing custom workload logic with GitLab runners that should help avoid potential issues:

1. **Label naming restrictions**: The custom workload label value must conform to Kubernetes naming conventions (RFC 1123).

2. **Using GitLab CI variables**: When constructing the custom workload label, use GitLab's predefined variables that are already sanitized:
   * `CI_JOB_NAME_SLUG`: A lowercase, sanitized version of `CI_JOB_NAME`
   * `CI_PROJECT_ID`: The unique ID of the project

3. **Shared runners**: For environments with shared runners across multiple projects, using a combination of `CI_JOB_NAME_SLUG` and `CI_PROJECT_ID` helps ensure unique workload identification per project and job type.

### Example configuration

Here's a complete example of a GitLab runner configuration with workload autoscaling enabled:

```toml
[[runners]]
  [runners.kubernetes]
    namespace = "{{.Release.Namespace}}"
    image = "alpine"
    [runners.kubernetes.pod_annotations]
      "workloads.atlas.example/configuration" = "vertical:\n  optimization: on\n"
    [runners.kubernetes.pod_labels]
      "workloads.atlas.example/custom-workload" = "${CI_JOB_NAME_SLUG}-${CI_PROJECT_ID}"
```

This configuration ensures the following key requirements:

* Each job type per project gets its own workload optimization profile
* Label values comply with Kubernetes naming requirements

For more information about available GitLab CI variables, refer to the [GitLab CI/CD predefined variables documentation](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html).

# Configuration via API/UI

The Atlas console provides an intuitive interface for configuring workload optimization settings at the scaling policy level. Through the web-based UI, you can adjust autoscaling behavior, set resource constraints, configure recommendation thresholds, and monitor optimization performance in real-time. You can also toggle optimization on or off for individual workloads.

For a comprehensive overview of all configurable parameters and their impact on workload optimization, see [Available settings](woop-configuration-settings.md).

![](https://files.readme.io/74326d7-image.png)

# Configuration via Annotations

Atlas Workload Autoscaler supports comprehensive configuration through Kubernetes annotations, enabling fine-grained control at the workload level. This method allows you to define optimization settings directly in your Kubernetes manifests, providing seamless integration with GitOps workflows and infrastructure-as-code practices.

When the `workloads.atlas.example/configuration` annotation is detected on a workload, it takes precedence over scaling policy settings for that specific workload. This approach enables centralized policy management while maintaining flexibility for workloads with unique requirements.

For complete annotation syntax, configuration examples, and detailed field references, see the [Workload Autoscaler Annotations Reference](workload-autoscaler-annotations-reference.md).