# Node Agent configuration

> Configure the Orion Node Agent component


The Node Agent is an optional DaemonSet that runs on your GPU nodes and collects GPU metrics from them. It is disabled by default.

This topic covers the Node Agent configuration parameters. For what the metrics are used for, see [GPU optimization](recommendations__gpu.md).

## Applying these settings

These are Helm values for the `orion` chart. Add them to a YAML values file (for example, `values.yaml`) and pass it to your Helm command.

At install time:

```shell
helm install orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  -f values.yaml
```

To apply changes after installation:

```shell
helm upgrade orion oci://registry.orion.example/library/orion \
  --namespace orion-system \
  --reset-then-reuse-values \
  -f values.yaml
```

---


**Node Agent parameters:**

- [Enable the Node Agent](#node-agent-enabled)
- [Namespace scope](#watch-namespace)
- [Node selection](#node-selection)
- [Pod priority and scheduling](#pod-priority)
- [Resource requests and limits](#resources)
- [Metrics port](#metrics-port)
- [Scrape interval](#scrape-interval)
- [Private container registry](#private-registry)
- [Permissions and host access](#permissions)

---

## Enable the Node Agent {#node-agent-enabled}

Controls whether Orion deploys the Node Agent.

| Helm parameter       | Default value |
| -------------------- | ------------- |
| `nodeAgent.enabled`  | `false`       |

### Description {#node-agent-enabled-description}

Set to `true` to deploy a DaemonSet named `orion-node-agent` that collects GPU metrics from the nodes it runs on. Enabling it also creates a ServiceAccount and a ClusterRole of the same name.

The Node Agent component is available in Orion Agent version 3.2.0 or later.

### Valid values {#node-agent-enabled-valid-values}

- `true`
- `false` (default)

### Example {#node-agent-enabled-example}

```yaml
nodeAgent:
  enabled: true
```

---

## Namespace scope {#watch-namespace}

Limits the namespaces whose pods the Node Agent watches.

| Helm parameter             | Default value |
| -------------------------- | ------------- |
| `nodeAgent.watchNamespace` | `""`          |

### Description {#watch-namespace-description}

The Node Agent watches pods to map each GPU process back to the container that owns it. By default it watches pods in all namespaces. Set `nodeAgent.watchNamespace` to a single namespace to narrow the scope, limiting GPU attribution to workloads in that namespace.

### Valid values {#watch-namespace-valid-values}

A namespace name, or `""` to watch all namespaces.

---

## Node selection {#node-selection}

Controls which nodes the Node Agent runs on.

| Helm parameter             | Default value                      |
| -------------------------- | ---------------------------------- |
| `nodeAgent.nodeSelector`   | `{nvidia.com/gpu.present: "true"}` |
| `nodeAgent.tolerations`    | `[{operator: Exists}]`             |

### Description {#node-selection-description}

Because the Node Agent only needs to run on GPU nodes, it ships with its own scheduling defaults, unlike the other Orion components. It selects nodes labeled `nvidia.com/gpu.present: "true"`, which the [NVIDIA GPU Feature Discovery][gfd] component sets, and tolerates every taint so that it can be scheduled onto tainted GPU node pools.

GPU nodes are not labeled the same way in every cluster. Cloud-provider accelerator labels, Karpenter node pools, and custom labels are all common. If your GPU nodes do not carry `nvidia.com/gpu.present`, override `nodeAgent.nodeSelector`.

`nodeAgent.nodeSelector` and `nodeAgent.tolerations` behave differently when you override them:

- `nodeAgent.tolerations` is a list, so the value you set **replaces** the default.
- `nodeAgent.nodeSelector` is a map, so the value you set **merges with** the default. Any label you add applies in addition to `nvidia.com/gpu.present`, and a node must carry both to be selected. To remove the default, set it to `null`.

### Valid values {#node-selection-valid-values}

For `nodeSelector` and `tolerations` format, see the [Kubernetes documentation on assigning pods to nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/).

### Example {#node-selection-example}

To select GPU nodes by a cloud-provider accelerator label instead of the default, and tolerate only the GPU taint:

```yaml
nodeAgent:
  nodeSelector:
    nvidia.com/gpu.present: null
    cloud.google.com/gke-accelerator: nvidia-tesla-t4
  tolerations:
  - key: nvidia.com/gpu
    operator: Exists
    effect: NoSchedule
```

To keep the default and narrow it further, add your label without the `null`. A node must then carry both labels:

```yaml
nodeAgent:
  nodeSelector:
    cloud.google.com/gke-accelerator: nvidia-tesla-t4
```
---

## Pod priority and scheduling {#pod-priority}

Enables priority configuration on Orion Node Agent pods.

| Helm parameter                   | Default value |
| -------------------------------- | ------------- |
| `priorityClassName`              | None          |
| `nodeAgent.priorityClassName`    | None          |
| `nodeAgent.affinity`             | None (inherits the top-level `affinity`) |

### Description {#pod-priority-description}

The top-level `priorityClassName` applies to all Orion components. Set `nodeAgent.priorityClassName` to override it for the Node Agent only. A non-empty per-component value **replaces** the top-level default.

`nodeAgent.priorityClassName` assigns an existing Pod `PriorityClass` to the `orion-node-agent` pods.

### Valid values {#pod-priority-valid-values}

A string that matches the `.metadata.name` of a `PriorityClass` defined in your cluster.

### Example {#pod-priority-example}

```yaml
nodeAgent:
  priorityClassName: "high-priority"
```

---

## Resource requests and limits {#resources}

Sets the Node Agent container's own resource settings.

| Helm parameter         | Default value                                                        |
| ---------------------- | -------------------------------------------------------------------- |
| `nodeAgent.resources`  | requests `cpu: 50m`, `memory: 128Mi`; limits `memory: 512Mi`          |

### Description {#resources-description}

The Node Agent's memory use grows with the number of GPUs on the node. The default limit suits dense multi-GPU nodes; raise it if the Node Agent is OOMKilled.

### Valid values {#resources-valid-values}

A Kubernetes [ResourceRequirements](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) object.

### Example {#resources-example}

```yaml
nodeAgent:
  resources:
    requests:
      cpu: 50m
      memory: 128Mi
    limits:
      memory: 1Gi
```

---

## Metrics port {#metrics-port}

Sets the port the Node Agent serves metrics and health checks on.

| Helm parameter          | Default value |
| ----------------------- | ------------- |
| `nodeAgent.metricsPort` | `9401`        |

### Description {#metrics-port-description}

The Node Agent serves its metrics at `/metrics` and its readiness check at `/healthz` on this port. The Forwarder scrapes the Node Agent on this port, and changing it updates both sides. Change it if `9401` conflicts with another workload on your GPU nodes, or if a network policy requires a specific port.

---

## Scrape interval {#scrape-interval}

Sets how often the Forwarder scrapes the Node Agent.

| Helm parameter             | Default value                              |
| -------------------------- | ------------------------------------------ |
| `nodeAgent.scrapeInterval` | `""` (inherits `forwarder.scrapeInterval`) |

### Description {#scrape-interval-description}

Leave empty to use the Forwarder's [scrape interval][forwarder-scrape-interval]. Set it to collect GPU metrics at a different frequency from the rest of the metrics.

### Valid values {#scrape-interval-valid-values}

A [Prometheus duration string](https://prometheus.io/docs/prometheus/latest/querying/basics/#float-literals-and-time-durations) (for example, `30s`, `1m`).

---

## Private container registry {#private-registry}

Override the Node Agent image repository for environments that use a private container registry.

| Helm parameter                 | Default value                                          |
| ------------------------------ | ------------------------------------------------------ |
| `nodeAgent.image.repository`   | `registry.orion.example/optimize/orion-node-agent` |
| `nodeAgent.image.tag`          | `""` (defaults to the chart appVersion)                 |
| `nodeAgent.image.pullPolicy`   | `IfNotPresent`                                          |

### Description {#private-registry-description}

The Node Agent uses its own image, separate from the image the Agent, Applier, and webhook share. When using a private registry, push this image to your registry as well and set `nodeAgent.image.repository` accordingly.

For the full setup procedure, see [If you have a private container registry](installation__install-adv.md#if-you-have-a-private-container-registry).

### Example {#private-registry-example}

```yaml
nodeAgent:
  image:
    repository: my-registry.example.com/optimize/orion-node-agent
```

---

## Permissions and host access {#permissions}

The Node Agent needs more node access than the other Orion components.

| Helm parameter                | Default value                                                                                     |
| ----------------------------- | ------------------------------------------------------------------------------------------------- |
| `nodeAgent.securityContext`   | `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, capabilities drop `ALL`, add `SYS_ADMIN` |

### Description {#permissions-description}

Reading GPU state per process requires access to the NVIDIA driver and to process information on the node, so the DaemonSet:

- Runs in the host PID namespace. The driver reports host process IDs, and the Node Agent reads each process's cgroup to map it to a pod. This is not configurable.
- Holds the `SYS_ADMIN` capability, and drops all others. On some GPU node images, including Bottlerocket's NVIDIA variant, the NVIDIA container runtime injects the driver libraries only into containers that hold this capability. Without it, the Node Agent runs without collecting GPU metrics and logs `nvml init: ERROR_LIBRARY_NOT_FOUND`.
- Mounts the kubelet's pod resources socket directory, `/var/lib/kubelet/pod-resources`, read-only, to learn which containers hold which GPU devices.
- Sets `readOnlyRootFilesystem: true` and `allowPrivilegeEscalation: false`.

On node images that inject the driver libraries unconditionally, such as Amazon Linux 2023, you can override `nodeAgent.securityContext` to drop `SYS_ADMIN` where your Pod Security or OPA policy forbids it. Confirm that the Node Agent starts afterward. The host PID namespace is required either way.

The Node Agent's ClusterRole grants read-only access (`get`, `list`, and `watch`) on pods and on `resourceslices.resource.k8s.io`.

The Node Agent does not mount the container runtime socket or the host root filesystem.

---

## Related topics

- [GPU optimization](recommendations__gpu.md)
- [Orion Node Agent](concepts.md#node-agent)
- Kubernetes official documentation:
  - [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
  - [Assigning Pods to Nodes](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/)

[gfd]: https://github.com/NVIDIA/gpu-feature-discovery
[forwarder-scrape-interval]: /docs/installation/forwarder-config/#scrape-interval
