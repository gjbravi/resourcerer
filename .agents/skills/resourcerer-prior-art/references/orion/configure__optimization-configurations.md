# Configure optimization using CRDs

> Specify optimization configurations at the cluster or namespace level using Custom Resource Definitions


Orion's optimization configuration options ensure consistent right-sizing across workloads and give teams control over their work.

* **Platform teams** can set organization-wide default optimization values
    
* **Platform teams** can provide different default optimization values for all workloads matching a type or label selector
    
* **Development teams** can fine-tune settings further for workloads in their namespaces, or for those matching specific label selectors
    
Optimization configurations are Custom Resource Definitions (CRDs) consisting of [rules](#rules) evaluated and applied to workloads in definition order.

## Types

Define optimization configurations using the following CRD types:

* **ClusterOptimizationConfiguration**
    * A cluster-scoped resource intended for use by platform teams to define organization-wide optimization policy.

    * **ClusterOptimizationConfiguration rules are always evaluated before any OptimizationConfiguration rules.**

* **OptimizationConfiguration**
    * A namespace-scoped resource intended for use by application or development teams that own namespace resources.

    * **Apply only to workloads in the defined namespace and are evaluated after ClusterOptimizationConfiguration rules.**

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Optimization configuration ordering</div>


If multiple ClusterOptimizationConfigurations exist, or if multiple OptimizationConfigurations exist for a namespace, Orion orders each type by `spec.order` ascending, then `metadata.name` ascending, and applies in that order. Take a look at [Ordering](#ordering) for more information.
</div>


### ClusterOptimizationConfiguration

Defines a list of cluster-scoped rules. Each rule can include `matchConditions`, `selector`, and/or `namespaceSelector` fields. An `order` field enables specifying the relative evaluation order of a ClusterOptimizationConfiguration relative to other ClusterOptimizationConfigurations.

```yaml
apiVersion: optimize.orion.example/v1
kind: ClusterOptimizationConfiguration
metadata:
  name: cluster-defaults
spec:
  order: 10
  rules:
  - name: NAME
    matchConditions: []
    selector: {}
    namespaceSelector: {}
    settings: []
```

### OptimizationConfiguration

Defines a list of namespace-scoped rules. Each rule can include `matchConditions` and `selector`. An `order` field enables specifying the relative evaluation order of an OptimizationConfiguration relative to other OptimizationConfigurations.

```yaml
apiVersion: optimize.orion.example/v1
kind: OptimizationConfiguration
metadata:
  name: namespace-defaults
  namespace: default
spec:
  order: 10
  rules:
  - name: NAME
    matchConditions: []
    selector: {}
    settings: []
```

## Ordering

When multiple optimization configurations of the same type exist, `spec.order` controls their relative evaluation order. The order field follows the **Higher = Stronger** pattern, where higher numbers are evaluated later, and can override settings specified in lower-order configurations. Configurations with the same `spec.order` are ordered alphabetically by name.

```yaml
spec:
  order: 10
```

<div class="alert alert-info" role="alert"><div class="h4 alert-heading" role="heading">Order values</div>


Order values must be in the range `[−1000, 1000]`, inclusive. If not specified, order defaults to 0.
</div>


## Rules

Define cascading rules—or configuration steps—as `spec.rules`. Each rule applies a new configuration or modifies the previous configuration to all workloads or workloads matching specific criteria.

```yaml
spec:
  rules:
  - name: NAME

    # CEL expressions enable non-label filtering. Expressions must be truthy.
    matchConditions:
    - name: NAME
      expression: 'object.kind == "Deployment"'

    # Selects workloads based on workload labels. Should be a Kube `metav1.LabelSelector`.
    selector:
      matchLabels:
        "app.kubernetes.io/name": "foobar"

    # Selects workloads based on namespace labels. Should be a Kube `metav1.LabelSelector`.
    # When the CRD is namespace-scoped, the namespace selector is effectively implied.
    namespaceSelector:
      matchLabels:
        "kubernetes.io/metadata.name": "default"

    # A Kube `listMap` keyed on `name`. Any setting string from:
    # https://docs.orion.example/configure/settings/
    settings:
    - name: "schedule"
      value: "P2D"
    - name: "memory.optimization-goal"
      value: "Balanced"
```

### Match conditions and selectors

Use `matchConditions` and `selectors` to constrain which workloads a rule applies to. For `matchConditions`, Common Expression Language (CEL) expressions are used to match non-label workload characteristics such as annotations, workload type, or Orion patch-target type. Namespace and workload selectors can match namespaces and workloads, respectively, by label.

Empty `selectors` or `matchConditions` match all workloads.

For a rule to match, a workload must satisfy all of `matchConditions`, `selector`, and `namespaceSelector`.

#### Common Expression Language variables

CEL is used to evaluate `matchConditions`. The following variables are available in the evaluation context:

* **`object:`** The workload object.

    * `apiVersion`: The workload's GroupVersion string (for example, `apps/v1`).

    * `kind`: The workload kind (for example, `Deployment`).

    * `metadata.name`: The workload name.

    * `metadata.namespace`: The workload namespace.

    * `metadata.labels`: A map of the workload's labels.

    * `metadata.annotations`: A map of the workload's annotations.

*   **`namespaceObject`:** The namespace object containing the workload; ClusterOptimizationConfiguration only.

    * `metadata.name`: The namespace name.

    * `metadata.labels`: A map of the namespace's labels.

    * `metadata.annotations`: A map of the namespace's annotations.

### Settings

Settings is a `listMap` keyed on `name`, consisting of `name` and `value` (or list of values), where `name` must be the name of a [Orion configuration setting](../settings), like `schedule` or `containers.cpu.requests.max`, and `value` must be a string.

If all `matchConditions` and `selectors` match a workload, settings are applied. Settings defined in later rules can override those from earlier rules.
