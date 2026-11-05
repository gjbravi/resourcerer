# CPU and memory

> Optimization policy, limit-to-request ratio, and bounds for CPU and memory


CPU and memory settings are applied at the container level. The configuration used can vary from container to container within a workload.

**CPU and memory settings:**

- [Optimization policy](#opt-policy)
- [Limit-to-request ratio](#lrr)
- [Min and max for requests](#requests-min-max)
- [Min and max for limits](#limits-min-max)
- [Patch paths](#patch-path)

For details about how to set these values using annotations, see one of these topics: [Configure clusters](configure__cluster-defaults.md), [Configure namespaces](configure__annotate-namespaces.md), or [Configure workloads](configure__annotate-workloads.md).

## Container-level settings syntax

When specifying configuration values for CPU and memory settings, you can specify a default value for all containers, one or more values for specific named containers, or a comma-separated list.

- **Default** (`"RequestsRaiseLimitsIfNeeded"`): default values for all containers are simple strings with a valid value for the setting. The default value is used for any containers that do not have a named value.
- **Named** (`"istio-proxy=DoNotOptimize"`): named container settings apply the value to containers with the given name. The name is specified first, followed by an `=` sign, then the value to apply to that container.
- **Comma-separated** (`"RequestsRaiseLimitsIfNeeded,istio-proxy=DoNotOptimize"`): an optional default, and one or more named values, can be specified as a comma-separated list. When specifying multiple values, start with the default (if you are including one), then list each named container value.

### Annotation examples

```yaml
live.orion.example/containers.cpu.requests.min: "10m"
live.orion.example/containers.cpu.requests.max: "16000m,jenkins=4000m,logshipper=20m"
live.orion.example/containers.memory.requests.min: "128Mi,jenkins=1Gi"
live.orion.example/containers.memory.requests.max: "logshipper=512Mi"
```

## CPU and memory settings {#cpu-and-memory-settings}

The following settings specify per-container values.

---

### Optimization policy {#opt-policy}

Indicates what to optimize for a container, for each resource (CPU and memory).

| Annotation                                                 | Default value    |
| ---------------------------------------------------------- | ---------------- |
| `live.orion.example/containers.cpu.optimization-policy`    | `"RequestsRaiseLimitsIfNeeded"` |
| `live.orion.example/containers.memory.optimization-policy` | `"RequestsRaiseLimitsIfNeeded"` |

#### Description {#opt-policy-description}

The configured optimization policy for a resource dictates how (and whether) Orion will recommend and make changes to a resource (CPU or memory). There are several ways organizations might like to manage limits, and there are use cases for choosing not to optimize a workload or container at all. Selecting an appropriate value for this setting should enable organizations to implement the resource management style of their choice.

#### Valid values {#opt-policy-valid-values}

- `RequestsRaiseLimitsIfNeeded` (default): Always optimize requests. Optimize limits only if they already exist:
  - If limits are set and are high enough already, leave them at their existing values.
  - If limits are lower than the recommended values, raise the limits to the recommended values.
  - If limits are higher than what Orion recommends, do not modify them. Orion will never lower a limit when using this policy.
  - If no limits are set, don't set them.

    Orion calculates its recommended limits as follows:\
    `min(max-limit, recommended-requests × limit-request-ratio)`
- `RequestsAndLimits`: Optimize both requests and limits. If limits don't exist, Orion will add them.
- `RequestsOnly`: Optimize requests only and do not modify limits. If limits are set, the existing limits will act as a constraint on the maximum value that Orion can recommend for requests.
- `DoNotOptimize`: Optimize neither requests nor limits. Orion will not suggest modifications to any existing settings for the resource.

#### Key points {#opt-policy-key-points}

- Optimization policy defines how and whether requests and limits will be optimized.
- Injected containers cannot be optimized without extra configuration. One way to handle this is to set both the CPU and memory policies for injected sidecar containers to `DoNotOptimize`. For example:

    ```yaml
    live.orion.example/containers.cpu.optimization-policy: "istio-proxy=DoNotOptimize"
    live.orion.example/containers.memory.optimization-policy: "istio-proxy=DoNotOptimize"
    ```

---

### Limit-to-request ratio (LRR) {#lrr}

This value is used to calculate limits when the optimization policy is set to `RequestsRaiseLimitsIfNeeded` or `RequestsAndLimits`.

| Annotation                                                        | Default value |
| ----------------------------------------------------------------- | ------------- |
| `live.orion.example/containers.cpu.limits.limit-request-ratio`    | `"2.0"` |
| `live.orion.example/containers.memory.limits.limit-request-ratio` | `"2.0"` |

#### Description {#lrr-description}

When Orion is configured to set limits, the limit it sets is computed as the recommended request value multiplied by the LRR. This ratio can be adjusted separately for each resource, CPU or memory.

#### Valid values {#lrr-valid-values}

- `"2.0"` (default)
- `"1.0"` (Guaranteed QoS)
- Any value >= `"1.0"` to a maximum of two decimal places

#### Key points {#lrr-key-points}

- If the `RequestsOnly` or `DoNotOptimize` policy is in use, the LRR setting has no effect.
- Using an LRR that permits bursting, such as `"2.0"`, lets workloads occasionally exceed their request values without being throttled or OOM-killed. This is advisable when automating request settings, because it gives Orion time to come back and adjust resource requests as necessary, rather than the workload being immediately OOM-killed or unnecessarily throttled.

---

### Minimum and maximum bounds for requests {#requests-min-max}

Values indicate the lower bound and upper bound of the range for CPU or memory requests.

| Annotation                                          | Default value |
| --------------------------------------------------- | ------------- |
| `live.orion.example/containers.cpu.requests.min`    | `"10m"` |
| `live.orion.example/containers.cpu.requests.max`    | None          |
| `live.orion.example/containers.memory.requests.min` | `"32Mi"` |
| `live.orion.example/containers.memory.requests.max` | None          |

#### Description {#requests-min-max-description}

The minimum and maximum **request** bounds for each resource act as a constraint on what request values Orion will recommend, regardless of the observed usage for the resource. These bounds can be used to set a floor below which requests will never drop and a ceiling above which Orion will never raise them.

When used as defaults for the cluster, request bounds can be used to implement default automation policies such as "never allocate less than 10m of CPU" or "never allocate more than 16Gi of memory without a specific exception". Exceptions can then be implemented by overriding the cluster default settings on a per-namespace or per-workload basis.

#### Valid values {#requests-min-max-valid-values}

Any positive Kubernetes quantity, appropriate to the resource. Examples:

- `"10m"` (for CPU)
- `"32Gi"` (for memory)

---

### Minimum and maximum bounds for limits {#limits-min-max}

Values indicate the lower bound and upper bound of the range for CPU or memory limits.

| Annotation                                        | Default value |
| ------------------------------------------------- | ------------- |
| `live.orion.example/containers.cpu.limits.min`    | None          |
| `live.orion.example/containers.cpu.limits.max`    | None          |
| `live.orion.example/containers.memory.limits.min` | None          |
| `live.orion.example/containers.memory.limits.max` | None          |

#### Description {#limits-min-max-description}

The minimum and maximum **limit** bounds for each resource act as a constraint on what limit values Orion will set, regardless of the request value for the resource. These bounds can be used to set a floor below which limits will never drop and a ceiling above which Orion will never raise limits.

When used as defaults for the cluster, limit bounds can be used to implement default automation policies such as "never set a CPU limit of less than 1 full core" or "never limit memory to less than 512Mi".

#### Valid values {#limits-min-max-valid-values}

Any positive Kubernetes quantity, appropriate to the resource. Examples:

- `"1000m"` (for CPU)
- `"512Mi"` (for memory)

---

### Patch paths {#patch-path}

A patchPath maps a container resource value from a recommendation to where it belongs in the patchTarget.

| Annotation                                                 | Default value |
| ---------------------------------------------------------- | ------------- |
| `live.orion.example/containers.cpu.requests.patch-path`    | None          |
| `live.orion.example/containers.cpu.limits.patch-path`      | None          |
| `live.orion.example/containers.memory.requests.patch-path` | None          |
| `live.orion.example/containers.memory.limits.patch-path`   | None          |

#### Description {#patch-path-description}

For built-in Kubernetes workload types, Orion knows which fields correspond to request and limit values and can automatically generate patches. When generating patches for workloads owned by custom resource types (CRDs), the patch path settings let you specify which CRD fields Orion will update with recommended request or limit values.

#### Valid values {#patch-path-values}

String. Examples:

- `"/spec/resources/requests/cpu"`
- `"/spec/resources/limits/cpu"`
- `"/spec/resources/requests/memory"`
- `"/spec/resources/limits/memory"`
- `"/spec/resources/limits/cpu,fluentd=/spec/fluentd/resources/limits/cpu"`

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/requests-and-limits/">Requests and limits</a>
                </h5>
                <p>Recommendations for CPU and memory resource settings</p>
            </div>
        </div>
