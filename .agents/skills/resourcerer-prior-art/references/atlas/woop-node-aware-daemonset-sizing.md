---
updatedAt: 2026-09-29T09:00:29.000Z
---

# Node-aware DaemonSet sizing

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

Some workloads — most notably DaemonSets, but also any pod whose footprint is naturally proportional to the node it runs on (log shippers, metric agents, CSI drivers, service meshes) — are hard to size with a single static request. A request that fits a small node wastes capacity on a large one, and a request that fits a large node gets evicted on a small one.

The **node allocatable percentage** resource strategy lets you express requests as a percentage of the node's allocatable CPU and memory, resolved at pod admission time. Workload Autoscaler reads the percentage from the workload's resource recommendation, looks up the node the pod is bound to, and rewrites the pod's requests and limits to a concrete value before the pod is admitted.

The strategy is currently only configurable via workload annotations. Scaling policy and UI support will follow.

# How it works

When you enable the strategy on a workload, Workload Autoscaler:

1. Picks up the configuration on the next reconciliation and stores the percentages alongside the workload's existing recommendation, together with a static fallback request and limit derived from observed usage.
2. As each pod is admitted, reads the node assigned to the pod and computes the request as a percentage of that node's allocatable CPU and memory:
   * `cpu request   = node allocatable CPU × cpuPercent / 100`
   * `memory request = node allocatable memory × memoryPercent / 100`
3. Clamps the resolved request to the workload's `min` and `max` [constraints](woop-configuration-settings.md#workload-autoscaler-constraints) (if set), then derives the limit from the clamped request using the workload's existing limit strategy (`multiplier`, `noLimit`, or `maintainRatio` in annotations).
4. Stamps the resolved request and limit into the pod spec.

> 📘 Note
>
> The percentage is resolved **per container**, not split across the pod. Each container in the pod gets its own request of `node allocatable × percent / 100`; it is not a shared budget divided between containers. See [Per-container sizing](#per-container-sizing).

If the target node cannot be determined at admission — for example, when an unbound pod that doesn't pin a node via `nodeAffinity` is admitted before scheduling — Workload Autoscaler applies the static fallback request and limit instead. Pods are never rejected.

The strategy fully overrides the requests and limits that Workload Autoscaler would otherwise apply for the configured resource.

> ⚠️ Not compatible with `keepLimits`
>
> If a workload's policy keeps existing limits (`keepLimits` in annotations) for CPU or memory, the percentage strategy is **not applied** to that resource — the workload falls back to the recommender's request and the manifest's original limit. Use `multiplier`, `noLimit`, or `maintainRatio` instead.

# Compatibility

Minimum component versions required for node-aware DaemonSet sizing:

| Component                    | Minimum version |
| :--------------------------- | :-------------- |
| `atlas-workload-autoscaler` | v0.105.0        |
| `atlas-agent`               | v0.123.1        |

> 📘 Note
>
> The percentage is resolved at pod admission. Existing pods keep whatever requests and limits they were admitted with — they pick up new values the next time they're recreated.

> 📘 Note
>
> **Pod-level resources (Kubernetes v1.34):** Workload Autoscaler converts [pod-level resource requests and limits](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#pod-level-resource-specification) to container-level ones and sizes each container individually. The `nodeAllocatablePercentage` strategy operates at container level, so it is unaffected by whether a pod uses pod-level or container-level resources.

# Configuration

Configure the strategy via the `workloads.atlas.example/configuration` annotation on the workload (Deployment, StatefulSet, DaemonSet, etc.):

```yaml
metadata:
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        resourceStrategy:
          type: nodeAllocatablePercentage
          nodeAllocatablePercentage:
            cpuPercent: 5
            memoryPercent: 2
```

You can set `cpuPercent` only, `memoryPercent` only, or both. The resource that is **not** configured falls back to whatever the rest of the Workload Autoscaler policy dictates (target utilization, recommender output, and so on).

Limits are derived from the resolved request using the workload's existing CPU and memory limit strategy. See [Annotations reference](workload-autoscaler-annotations-reference.md) for the available limit options.

## Per-container sizing

Workload Autoscaler resolves the percentage independently for **each container in the pod**. It does not split it across the pod as a shared budget. For example, on a node with 8 allocatable CPU, a workload configured with `cpuPercent: 10` gives **every** container a CPU request of `800m`. A pod with two containers gets `800m` per container (1.6 CPU total), not `800m` shared between them.

You can configure the strategy at three scopes:

* **Workload level**: set `vertical.resourceStrategy`. This applies to every container in the pod that has CPU or memory requests in its spec. Autoscaler leaves a container with no requests defined alone.
* **Container level**: set `vertical.containers.{container_name}.resourceStrategy`. This sizes only that container, and can use a different percentage per container. Because the strategy targets that container explicitly, it can also add requests to a container that has none.
* **Both**: a container-level `resourceStrategy` **overrides** the workload-level strategy for that container. All other containers fall back to the workload-level strategy.

A container can opt out entirely with `vertical.excludedContainers`. An excluded container keeps its original (static) resources while Autoscaler sizes the rest of the pod by percentage. See [Annotations reference](workload-autoscaler-annotations-reference.md#verticalexcludedcontainers).

### Example: different percentages per container

A DaemonSet with two containers. The workload defaults to 80% of node-allocatable CPU, but you set the `app` container explicitly to 30% CPU and 40% memory:

```yaml
metadata:
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        resourceStrategy:
          type: nodeAllocatablePercentage
          nodeAllocatablePercentage:
            cpuPercent: 80
        containers:
          app:
            resourceStrategy:
              type: nodeAllocatablePercentage
              nodeAllocatablePercentage:
                cpuPercent: 30
                memoryPercent: 40
```

On a node with **8 allocatable CPU / 32 GiB allocatable memory**, each pod resolves to:

| Container | CPU request | Memory request | Why                                            |
| :-------- | :---------- | :------------- | :--------------------------------------------- |
| `app`     | 2400m       | \~12.8 GiB     | Container-level override (30% CPU, 40% memory) |
| (others)  | 6400m       | unchanged      | Workload-level `cpuPercent: 80` applies        |

The `app` container uses its container-level override for both resources. Every other container in the pod uses the workload-level strategy. Each such container gets a CPU request of `6400m` (80%), while its memory stays untouched because the workload level sets no `memoryPercent`.

## Settings reference

The same fields exist at the workload level (`vertical.resourceStrategy`) and the container level (`vertical.containers.{container_name}.resourceStrategy`). Container-level settings override the workload-level strategy for that container.

| Field                                                      | Type   | Required                                        | Range                       | Description                                                                  |
| :--------------------------------------------------------- | :----- | :---------------------------------------------- | :-------------------------- | :--------------------------------------------------------------------------- |
| `resourceStrategy.type`                                    | string | Yes                                             | `nodeAllocatablePercentage` | Selects the strategy.                                                        |
| `resourceStrategy.nodeAllocatablePercentage.cpuPercent`    | float  | One of `cpuPercent`/`memoryPercent` is required | `(0, 100]`                  | Percentage of node-allocatable CPU to request. Resolved per container.       |
| `resourceStrategy.nodeAllocatablePercentage.memoryPercent` | float  | One of `cpuPercent`/`memoryPercent` is required | `(0, 100]`                  | Percentage of node-allocatable memory to request. Resolved per container.    |
| `containers.{container_name}.resourceStrategy`             | object | No                                              | -                           | Per-container override. Same shape as the workload-level `resourceStrategy`. |

# Example

Consider a DaemonSet annotated to request 5% of node-allocatable CPU and 2% of node-allocatable memory, with limits set to 1.5× the resolved request:

```yaml
metadata:
  annotations:
    workloads.atlas.example/configuration: |
      vertical:
        resourceStrategy:
          type: nodeAllocatablePercentage
          nodeAllocatablePercentage:
            cpuPercent: 5
            memoryPercent: 2
        cpu:
          limit:
            type: multiplier
            multiplier: 1.5
        memory:
          limit:
            type: multiplier
            multiplier: 1.5
```

On a **16 vCPU / 64 GiB** node, each pod is sized to:

| Resource | Request    | Limit      |
| :------- | :--------- | :--------- |
| CPU      | 800m       | 1200m      |
| Memory   | \~1.28 GiB | \~1.92 GiB |

On a **4 vCPU / 16 GiB** node, each pod is sized to:

| Resource | Request   | Limit     |
| :------- | :-------- | :-------- |
| CPU      | 200m      | 300m      |
| Memory   | \~327 MiB | \~491 MiB |

The same workload definition produces different requests on different nodes — no extra configuration, no per-node policies.

This example sets the strategy at the workload level, so the same percentages apply to every container in the pod that has CPU and memory requests in its spec. To use different percentages per container, or to size a container that has no requests defined, use [per-container sizing](#per-container-sizing) instead.

# Limitations

* **CPU and memory only.** Ephemeral storage is not supported.
* **Limits and `lookBackPeriod` are workload-level only.** Limits are configured at the workload level (`vertical.cpu.limit` / `vertical.memory.limit`) and apply uniformly to all containers; different multipliers per container are not supported. Per-container `limit` and `lookBackPeriod` settings are not supported and are silently ignored, so set them on `vertical` instead. (Per-container `min`/`max` constraints are supported.)
* **Not compatible with `keepLimits` for the same resource.** Workload Autoscaler rejects configurations that combine the percentage strategy with `keepLimits` on the same resource. Use `multiplier`, `noLimit`, or `maintainRatio` instead.
* **Target node must be known at admission.** Workload Autoscaler only resolves the percentage when the pod has `spec.nodeName` set, or when a `nodeAffinity` rule pins it to a single node by `metadata.name`. DaemonSet pods always meet this requirement — the DaemonSet controller fills in `spec.nodeName` before the pod is admitted. For Deployments, StatefulSets, and other workloads where the pod is admitted *before* scheduling, the pod gets the static fallback request and limit unless it pins a node via `nodeAffinity`.
* **Resolved per pod, not retroactively.** Resizing a node in place does not resize the pods already running on it — they pick up new values the next time they're recreated.
* **Annotations only.** There is no UI or scaling policy support yet; configure the strategy on each workload directly.

# Troubleshooting

* **Pod has the static fallback request, not the percentage-derived one.** Workload Autoscaler couldn't determine the target node at admission. Confirm the pod has `spec.nodeName` set or a `nodeAffinity` rule that pins it to a single node by `metadata.name`. Check the Workload Autoscaler logs for fallback messages.
* **Every pod gets the same request regardless of node.** Autoscaler first computes the percentage from the node, then clamps it to the workload's `min`/`max` constraints. On an 8 vCPU node, `cpuPercent: 5` resolves to \~`400m`, but a `max` of `10m` clamps it back to `10m`. Every node then produces the same value. Check your effective constraints and raise `max`, or lower a `min` that's pinning values up.
* **A container in a multi-container pod isn't being sized.** Two common causes: (1) you set only a workload-level strategy and that container has no CPU/memory requests in its spec. The workload-level strategy only applies to containers that already have requests, so set a [container-level strategy](#per-container-sizing) to add them. (2) The container is in `vertical.excludedContainers`, which keeps its original resources by design.
* **Per-container `limit` or `lookBackPeriod` seems to be ignored.** Those settings are workload-level only (see [Limitations](#limitations)). Move them to `vertical.cpu.limit` / `vertical.memory.limit`.
* **The percentage strategy isn't being applied at all.** Check that the workload doesn't combine the strategy with `keepLimits` on the same resource — if it does, Workload Autoscaler falls back to the standard recommendation for that resource. Either remove `keepLimits` or remove the percentage for that resource.
* **Limits look higher than expected.** Limits are derived from the **resolved** request — that is, after `min`/`max` constraints are applied. If you set a `min` of `500m` and a limit `multiplier` of `1.5`, a node where 5% would be `200m` still produces a request of `500m` and a limit of `750m`.

# See also

<Cards>
  <Card title="Workload Autoscaler configuration" href="https://docs.atlas.example/docs/workload-autoscaling-configuration" icon="fa-gear">
    Configure how Workload Autoscaler sizes your workloads — apply modes, percentiles, overhead, and resource limits.
  </Card>

  <Card title="Annotations reference" href="https://docs.atlas.example/docs/workload-autoscaler-annotations-reference" icon="fa-tag">
    Full reference for the `workloads.atlas.example/configuration` annotation, including limit strategies and per-container overrides.
  </Card>
</Cards>