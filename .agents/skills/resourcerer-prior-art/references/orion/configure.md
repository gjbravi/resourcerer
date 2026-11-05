# Configure optimization settings

> Define the optimization policy for requests and limits at the cluster, namespace, or workload level


Policy-based optimization ensures consistency across workloads and gives teams the control they need over their work - for example:

- Platform teams can set organization-wide default optimization values  
- Development teams can fine-tune settings using annotations on namespaces, Deployments, StatefulSets, and other workload resources

This topic explains how to configure optimization settings at the cluster, namespace, or workload levels using annotation syntax. 

For a list and description of what you can configure, see the [Optimization settings and descriptions](configure__settings.md) topic.

## Key points

- Think of settings at the **cluster**, **namespace**, and **workload** levels as a hierarchy of configuration.

  - **Cluster-level defaults** apply to all workloads in a cluster. The cluster administrator defines cluster-wide defaults that are used when a more specific policy is not defined. Default values set at this level have the **lowest precedence and can be overridden** by values set at the namespace and workload levels.

  - **Namespace-level configuration** sets default values for all workloads in a namespace, overriding existing cluster defaults. Namespace owners can adjust these policies, or define their own, by using annotations on the namespace.

  - **Workload-level configuration** customizes the optimization behavior for individual workloads. Workload owners can further adjust or make exceptions to optimization policies, by using annotations on the actual workload objects (Deployments, StatefulSets, and others).

## Quickstart: Copy a template {#quickstart-templates}


**Choose a template**

Each tab provides a template with annotations for each optimization setting.

- Copy the **Cluster** template to set cluster-level defaults
- Copy the **Namespace** template to set namespace-level defaults
- Copy the **Workload** template to set workload-specific defaults

**Cluster**

To configure default optimization settings for **all workloads in a cluster**, copy annotations as needed from the sample file below, then pass the file to `kubectl apply -f FILENAME`.

Review the example values and adjust for your environment before applying. All of these settings are optional, so remove any you don’t need.

[Go to the steps for using this template.](configure__cluster-defaults.md#configure-defaults)

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-defaults
  namespace: orion-system
data:
  cluster-defaults.yaml: |

    # WORKLOAD SETTINGS
    # Specify single values for the whole workload.
    live.orion.example/schedule: "@daily"
    live.orion.example/cpu.optimization-goal: "Balanced"
    live.orion.example/memory.optimization-goal: "Balanced"

    # Main configuration control for auto-deploy
    live.orion.example/auto-deploy: "Disabled"

    # When auto-deploy is Disabled, the following subsettings have no effect
    live.orion.example/auto-deploy.thresholds.cpu.percent: "5"    # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.percent: "5" # 2/2.
    live.orion.example/auto-deploy.thresholds.cpu.unit: "10m"     # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.unit: "20Mi" # 2/2.

    # The following settings have an effect only if an HPA is enabled and the HPA scales on CPU or memory.
    live.orion.example/hpa.cpu.target-utilization.min: "50"
    live.orion.example/hpa.cpu.target-utilization.max: "90"
    live.orion.example/hpa.memory.target-utilization.min: "50"
    live.orion.example/hpa.memory.target-utilization.max: "90"

    # CONTAINER SETTINGS
    # Specify defaults and/or per-container exceptions.
    # For example: "RequestsRaiseLimitsIfNeeded,istio-proxy=DoNotOptimize"
    live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.cpu.requests.min: "10m"
    live.orion.example/containers.cpu.limits.min: "2000m"
    live.orion.example/containers.cpu.limits.limit-request-ratio: "2.0"
    live.orion.example/containers.memory.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.memory.requests.min: "32Mi"
    live.orion.example/containers.memory.limits.min: "512Mi"
    live.orion.example/containers.memory.limits.limit-request-ratio: "2.0"
    # Maximum bounds: align with your organization's maximums
    # live.orion.example/containers.cpu.requests.max: "16000m"
    # live.orion.example/containers.cpu.limits.max: "16000m"
    # live.orion.example/containers.memory.requests.max: "32Gi"
    # live.orion.example/containers.memory.limits.max: "64Gi"
```

**Namespace**

To configure default optimization settings for **all workloads in a namespace**, copy annotations as needed from the template below into the `metadata.annotations` section of a namespace.

Review the example values and adjust for your environment before applying. All of these settings are optional, so remove any you don’t need.

[Go to the steps for using this template.](configure__annotate-namespaces.md#annot-ns)

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: example-namespace-name
  annotations:

    # WORKLOAD SETTINGS
    # Specify single values for the whole workload.
    live.orion.example/schedule: "@daily"
    live.orion.example/cpu.optimization-goal: "Balanced"
    live.orion.example/memory.optimization-goal: "Balanced"

    # Main configuration control for auto-deploy
    live.orion.example/auto-deploy: "Disabled"

    # When auto-deploy is Disabled, the following subsettings have no effect
    live.orion.example/auto-deploy.thresholds.cpu.percent: "5"    # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.percent: "5" # 2/2.
    live.orion.example/auto-deploy.thresholds.cpu.unit: "10m"     # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.unit: "20Mi" # 2/2.

    # The following settings have an effect only if an HPA is enabled and the HPA scales on CPU or memory.
    live.orion.example/hpa.cpu.target-utilization.min: "50"
    live.orion.example/hpa.cpu.target-utilization.max: "90"
    live.orion.example/hpa.memory.target-utilization.min: "50"
    live.orion.example/hpa.memory.target-utilization.max: "90"

    # CONTAINER SETTINGS
    # Specify defaults and/or per-container exceptions.
    # For example: "RequestsRaiseLimitsIfNeeded,istio-proxy=DoNotOptimize"
    live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.cpu.requests.min: "10m"
    live.orion.example/containers.cpu.limits.min: "2000m"
    live.orion.example/containers.cpu.limits.limit-request-ratio: "2.0"
    live.orion.example/containers.memory.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.memory.requests.min: "32Mi"
    live.orion.example/containers.memory.limits.min: "512Mi"
    live.orion.example/containers.memory.limits.limit-request-ratio: "2.0"
    # Maximum bounds: align with your organization's maximums
    # live.orion.example/containers.cpu.requests.max: "16000m"
    # live.orion.example/containers.cpu.limits.max: "16000m"
    # live.orion.example/containers.memory.requests.max: "32Gi"
    # live.orion.example/containers.memory.limits.max: "64Gi"
```

**Workload**

To configure optimization settings for a workload, copy annotations as needed from the template below into the `metadata.annotations` section of the workload resource.

The template shows annotations defined on a Deployment workload. The same pattern applies for all supported workload types.

Review the example values and adjust for your environment before applying. All of these settings are optional, so remove any you don’t need.

[Go to the steps for using this template.](configure__annotate-workloads.md#annot-workload)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-workload-name
  annotations:

    # WORKLOAD SETTINGS
    # Specify single values for the whole workload.
    live.orion.example/schedule: "@daily"
    live.orion.example/cpu.optimization-goal: "Balanced"
    live.orion.example/memory.optimization-goal: "Balanced"

    # Main configuration control for auto-deploy
    live.orion.example/auto-deploy: "Disabled"

    # When auto-deploy is Disabled, the following subsettings have no effect
    live.orion.example/auto-deploy.thresholds.cpu.percent: "5"    # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.percent: "5" # 2/2.
    live.orion.example/auto-deploy.thresholds.cpu.unit: "10m"     # 1/2 Paired parameters; set both or neither.
    live.orion.example/auto-deploy.thresholds.memory.unit: "20Mi" # 2/2.

    # The following settings have an effect only if an HPA is enabled and the HPA scales on CPU or memory.
    live.orion.example/hpa.cpu.target-utilization.min: "50"
    live.orion.example/hpa.cpu.target-utilization.max: "90"
    live.orion.example/hpa.memory.target-utilization.min: "50"
    live.orion.example/hpa.memory.target-utilization.max: "90"

    # CONTAINER SETTINGS
    # Specify defaults and/or per-container exceptions.
    # For example: "RequestsRaiseLimitsIfNeeded,istio-proxy=DoNotOptimize"
    live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.cpu.requests.min: "10m"
    live.orion.example/containers.cpu.limits.min: "2000m"
    live.orion.example/containers.cpu.limits.limit-request-ratio: "2.0"
    live.orion.example/containers.memory.optimization-policy: "RequestsRaiseLimitsIfNeeded"
    live.orion.example/containers.memory.requests.min: "32Mi"
    live.orion.example/containers.memory.limits.min: "512Mi"
    live.orion.example/containers.memory.limits.limit-request-ratio: "2.0"
    # Maximum bounds: align with your organization's maximums
    # live.orion.example/containers.cpu.requests.max: "16000m"
    # live.orion.example/containers.cpu.limits.max: "16000m"
    # live.orion.example/containers.memory.requests.max: "32Gi"
    # live.orion.example/containers.memory.limits.max: "64Gi"

spec:
  template:
    metadata:
```



**When editing annotations:**

- Enclose all values in double quotation marks (`"..."`).
- **Best practice:** When editing _container-specific_ resources, specify the value to apply to _all or most containers_ first, followed by container-specific exceptions, as in these examples:
  - When annotating a workload or namespace:\
 `live.orion.example/containers.cpu.optimization-policy: "RequestsRaiseLimitsIfNeeded,istio-proxy=DoNotOptimize,logwriter=RequestsAndLimits"`

---

Section pages:

- [Getting started with configuring Orion](configure__configure-opt-settings.md): The five most important optimization settings and a starter template
- [Configure optimization defaults for the cluster](configure__cluster-defaults.md): Configure default optimization settings for the cluster
- [Configure optimization defaults for namespaces](configure__annotate-namespaces.md): Default optimization settings for every workload in a namespace, set with annotations
- [Configure optimization for individual workloads](configure__annotate-workloads.md): Optimization settings for a single workload, set with annotations
- [Configure optimization using CRDs](configure__optimization-configurations.md): Specify optimization configurations at the cluster or namespace level using Custom Resource Definitions
- [Optimization settings and descriptions](configure__settings.md): Orion optimization settings, by category
