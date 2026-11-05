# Pod scheduling

> Scheduling optimization policy for node affinity recommendations


- [Affinity type](#affinity-type)
- [Affinity weight](#affinity-weight)
- [Scheduling optimization policy](#sched-opt-policy)
- [Patch path](#patch-path)
- [Patch format](#patch-format)

## Scheduling optimization policy {#sched-opt-policy}

Whether or not node affinity is included in recommendation patches depends on the optimization policy selected.

| Annotation                                              | Default value  |
| --------------------------------------------------------| -------------- |
| `live.orion.example/pod-scheduling.optimization-policy` | `DoNotOptimize`|

### Description {#sched-opt-policy-description}

Orion recommends an instance category for pod scheduling by default, but that recommendation will not be included in patches unless an optimization policy is configured.

Instance category node affinity recommendations can be added to patches by setting the optimization policy to `InstanceCategories`.

### Valid values {#sched-opt-policy-valid-values}

- `DoNotOptimize`
- `InstanceCategories`

---

## Affinity type {#affinity-type}

| Annotation                                                                | Default value |
| --------------------------------------------------------------------------| ------------- |
|`live.orion.example/pod-scheduling.instance-categories.node-affinity.type` | `preferred`   |

### Description {#affinity-type-description}

Specifies which node affinity type to use when patching in node affinity settings.

Instance category recommendations will only be included in Orion patches if you have configured a specific node affinity type. By default, no node affinity type is configured.

The available settings align with [Kubernetes node affinity][k8s-node-affinity] types.

### Valid values {#affinity-type-valid-values}

- `preferred` – alias for `PreferredDuringSchedulingIgnoredDuringExecution`
- `required` – alias for `RequiredDuringSchedulingIgnoredDuringExecution`

---

## Affinity weight {#affinity-weight}

| Annotation                                                                        | Default value |
| ----------------------------------------------------------------------------------| ------------- |
| `live.orion.example/pod-scheduling.instance-categories.node-affinity.weight`      | `"70"`        |

### Description {#affinity-weight-description}

Specifies a weight for a `preferred` affinity.

This setting has no effect when [affinity type](#affinity-type) is set to `required`.

### Valid values {#affinity-weight-valid-values}

- Integer value between 1 and 100 (as a string)

---

## Patch path {#patch-path}

| Annotation                                                                        | Default value |
| ----------------------------------------------------------------------------------| ------------- |
| `live.orion.example/pod-scheduling.instance-categories.node-affinity.patch-path`  | None          |

### Description {#patch-path-description}

Indicates the absolute JSON pointer path to the field where the value is stored on the patched resource.

This can be used to fine-tune Orion patching behavior when optimizing custom CRD and operator-owned workload types.

### Examples {#patch-path-examples}

- `/spec/template/spec/affinity/nodeAffinity`

---

## Patch format {#patch-format}

| Annotation                                                                        | Default value |
| ----------------------------------------------------------------------------------| ------------- |
| `live.orion.example/pod-scheduling.instance-categories.node-affinity.patch-format`| None          |

### Description {#patch-format-description}

Defines the Go template that formats the value for the patched resource.

This can be used to fine-tune Orion patching behavior when optimizing custom CRD and operator-owned workload types.

### Valid values {#patch-format-valid-values}

- String

<div class="section-index related-topics">
    
    <hr class="panel-line">
        <h4>Related topics</h4>
        <div class="entry">
                <h5>
                    
                    Recommendations<span class="related-topics-divider">/</span><a href="/docs/recommendations/pod-scheduling/">Pod scheduling</a>
                </h5>
                <p>Automated node affinity for proportional resource scheduling</p>
            </div>
        </div>


[k8s-node-affinity]: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#node-affinity
