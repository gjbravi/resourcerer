---
updatedAt: 2026-10-15T07:13:49.000Z
---

# Workload eligibility

How the CLM controller determines which pods are eligible for live migration — pod owner types, GPU and CPU constraints, PVC zones, pod state, and refusal reasons.

Before Container Live Migration moves a pod, the live controller checks whether the pod is eligible. This page explains each check, why it exists, and what happens when a pod fails one.

The checks confirm that the workload, the source node, and the destination node are all compatible. The controller evaluates these conditions up front and refuses the migration with a clear reason rather than letting the migration proceed and fail in the middle. Knowing what the system checks — and how to read those refusal reasons — saves a lot of troubleshooting time.

This is the second step in the reader journey: after understanding [what live migration is](clm-overview.md), the question is "should I even try migrating this pod?"

## Why eligibility checks matter

Live migration captures a running container's state and restores it on another node. That only works when the source and destination are similar enough that the process can resume. Different CPU architectures, different CPU features, different storage zones, or different network subnets can all cause a restored container to crash, hang, or lose data.

### Platform awareness

The Atlas platform components — autoscaler, evictor, and rebalancer — are aware of the constraints declared on the workload, which is what this document enumerates. When deciding which nodes to drain, consolidate, or rebalance, these components automatically select workloads that are eligible for live migration and skip those that are not. You do not need to manually verify each constraint before triggering a migration; the platform does this for you.

The platform accounts for scheduling constraints that are part of the pod's spec, including ones an admission webhook injected when the pod was created. It cannot account for admission policies whose scheduling outcome is decided per pod at creation time. See [Requirements and limitations](clm-requirements-and-limitations.md#scheduling-constraints-set-by-admission-webhooks) for this limitation.

Use this document to understand **why** a migration was refused or why the platform chose one workload over another. If a workload was skipped during node consolidation, the reason is almost always one of the constraints described here.

## Quick eligibility checklist

A workload is considered eligible for migration when its pods and nodes meet all of the following criteria:

* The pod is owned by a supported controller, or has no controller (see [Pod owner types](#pod-owner-types)).
* The pod does **not** request GPU resources (see [GPU workloads](#gpu-workloads)).
* The pod is in the `Running` phase (see [Pod state](#pod-state)).
* The source and destination nodes have the same CPU architecture (see [CPU architecture](#cpu-architecture)).
* For amd64 nodes, the source and destination expose compatible CPU features (see [CPU capability checks](#cpu-capability-checks)).
* Both nodes carry the required labels (see [Node requirements](#node-requirements)).
* The destination node is in a zone that is allowed for the pod's persistent volumes, if any (see [PVC and zone constraints](#pvc-and-zone-constraints)).
* For TCP preservation on AWS, the source and destination are in compatible subnets or zones (see [TCP and subnet constraints](#tcp-and-subnet-constraints)).

If any of these fail, the migration is refused before any checkpoint is taken, so there is no cleanup burden on the cluster.

## Pod owner types

The controller only migrates pods that are owned by a known controller, or that have no controller at all. The owner determines how the system handles the pod after migration.

### Supported owner kinds

| Owner kind      | Migratable? | Notes                                                                                                 |
| --------------- | ----------- | ----------------------------------------------------------------------------------------------------- |
| None (bare Pod) | Yes         | The pod is migrated as-is.                                                                            |
| `Job`           | Yes         |                                                                                                       |
| `ReplicaSet`    | Yes         | Pods belonging to a ReplicaSet are supported directly.                                                |
| `StatefulSet`   | Yes         | The system coordinates with the StatefulSet's ordinal so the replacement pod keeps the same identity. |

Custom controllers are not supported by default.

Our team keeps expanding support for custom controllers of popular frameworks.

### Why Deployments work even though Deployment is not listed

A Deployment does not own pods directly. It creates a ReplicaSet, and the ReplicaSet owns the pods. Because pods are always owned by a ReplicaSet in a Deployment-managed workload, they are supported via that ReplicaSet path. This is the normal Kubernetes relationship, not a special case.

> **Note:** If you ever see a pod owned directly by a `Deployment` (without a ReplicaSet in between), that pod cannot be migrated. In practice this should not happen because the Deployment controller always creates a ReplicaSet.

### Unsupported owner kinds

| Owner kind                                            | Why not                                                                                                                                    |
| ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| `Deployment` (direct ownership, bypassing ReplicaSet) | Does not happen in normal operation; supported via the ReplicaSet it creates.                                                              |
| `DaemonSet`                                           | One pod per node by definition — there is nothing to migrate to.                                                                           |
| `CronJob`                                             | Not recognized as a direct pod controller. Migrate the `Job` that the CronJob creates instead.                                             |
| `ReplicationController`                               | Legacy controller; not supported.                                                                                                          |
| `Rollout` (Argo Rollouts)                             | Not recognized as a direct pod controller. Pods owned via a ReplicaSet under a Rollout are supported (the ReplicaSet is the direct owner). |
| Custom controllers / CRDs                             | Unknown ownership semantics; refused by default.                                                                                           |

## GPU workloads

Pods that request GPU resources are **not eligible** for live migration. Live migration cannot move the GPU device context — the memory on the GPU card, the kernel state, and the CUDA allocations — across nodes. Even if the destination node has a GPU, the running workload would have to be reinitialized on the new GPU, which defeats the purpose of live migration.

A pod is treated as a GPU workload if **any** container in the pod (including init containers) requests or limits a resource whose name contains a vendor prefix (`nvidia.com`, `amd.com`, or `intel.com`) and the substring `gpu`.

The system refuses the migration with a clear "GPU workloads are currently not supported" message. GPU support for live migration is planned for a future release.

**What to do instead:**

* Drain GPU nodes using standard Kubernetes eviction. The application restarts on a new GPU node; for many GPU workloads this is acceptable because warm-up is dominated by model loading anyway.
* For long-running GPU training jobs that need to survive node drains, consider checkpointing at the application level (for example, periodic checkpoints to object storage) and resuming on a new node.

## CPU architecture

The source and destination nodes must run the same CPU architecture. A pod compiled for, or already running on, `amd64` cannot be restored on an `arm64` node (and vice versa). The architecture is read from the standard `kubernetes.io/arch` label on each node.

If either node is missing the architecture label, the migration is refused.

| Source arch | Destination arch | Outcome                            |
| ----------- | ---------------- | ---------------------------------- |
| `amd64`     | `amd64`          | Allowed (subject to other checks). |
| `arm64`     | `arm64`          | Allowed (subject to other checks). |
| `amd64`     | `arm64`          | Refused.                           |
| `arm64`     | `amd64`          | Refused.                           |
| Any         | Missing label    | Refused.                           |

## CPU capability checks

On `amd64` nodes, the system goes further than just checking the architecture: it compares the **CPU features** the kernel exposes on each node. Two amd64 nodes can still be incompatible at the instruction-set level — for example, one may have AVX-512 and the other may not. A workload that uses AVX-512 at runtime cannot be restored on a node that lacks it.

CPU feature checks are skipped on `arm64` nodes. Currently it relies on appropriate instance constraints set on the node template via the compatibility helper.

### How the check works

Each node carries an annotation, `live.atlas.example/cpuinfo`, that summarizes the CPU features visible to the kernel. The migration agent compares the source's annotation to the destination's annotation using a strict comparison: capability version, every capability bit, the extended-state (`xfeatures`) mask, the `xsave` size, and the `xsave` maximum size must all match exactly.

If any of these differ, the migration is refused — even a single missing CPU feature is enough to block it.

### Missing annotation

| Situation                               | Outcome                                                                                                                                                  |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Source node missing the annotation      | Refused. The source must report its CPU features.                                                                                                        |
| Destination node missing the annotation | Allowed. The destination may not exist yet (for example, an autoscaling group that has not yet booted a node), so the absence is treated as a soft fail. |

### When CPU mismatches happen

CPU mismatches are most common in clusters that mix instance families (for example, older Intel nodes without AVX-512 alongside newer nodes with AVX-512). Use the annotation to see what each node reports, and either move workloads to compatible nodes or schedule them with node selectors / affinity to keep them on a compatible pool.

## PVC and zone constraints

If the pod uses a persistent volume (PV), the destination node must be in a zone that the PV is allowed to be attached to. This protects against silent data unavailability after migration.

### How the constraint is decided

1. The system looks up the PV's `NodeAffinity`, if set. This is the standard Kubernetes way to declare which zones a PV can be attached to.
2. If the PV declares allowed zones, the destination node's zone must be in that list. Multi-zone and regional volumes work — the destination can be in any of the listed zones.
3. If the PV does not declare zone affinity, the constraint is skipped and the migration is allowed regardless of zone.

This logic supports regional PVs (which list multiple allowed zones) without special configuration.

The system checks zone compatibility in two modes:

* **PV lookup mode (default):** The system reads the PV's `NodeAffinity` to find the allowed zones. Zonal PVs require source and destination in the same zone. Regional PVs allow migration to any zone in the PV's allowed list. If the PV has no zone affinity, the migration is allowed regardless of zone.

* **Legacy mode:** When PV lookup is disabled by configuration, the system falls back to a stricter check that requires source and destination zones to match exactly. Multi-zone PVs would be blocked unless source and destination happen to share a zone. This mode exists for backward compatibility and is not recommended for new deployments.

## TCP and subnet constraints

When TCP preservation is in use, additional networking constraints apply so that connections to and from the migrated pod can be correctly rebuilt on the destination. These checks are only meaningful when the pod is being migrated with TCP preservation enabled — if TCP preservation is off, the pod simply gets a new IP and connections break anyway.

The exact constraints depend on whether the cluster is running on AWS and whether the pod uses custom networking. On non-AWS providers there are no TCP-specific constraints beyond the zone/PVC checks above.

### Custom-networked pods on AWS

Some pods are scheduled with a `live.atlas.example/custom-network` label that puts them on a specific custom network. For these pods, the source and destination must be in the **same availability zone** so the custom network reaches both sides.

### Standard TCP pods on AWS

For pods that are not custom-networked, the source and destination must share the same **subnet**. The system uses the `topology.atlas.example/subnet-id` label on each node to compare subnets.

Subnets are usually tied to a single availability zone, so this is effectively a same-zone check in most clusters. The subnet label is set by enabling `daemon.labelNodeSubnet: true` in the CLM Helm chart — the daemon queries the instance metadata service and labels each node with its subnet ID. This setting is AWS-only and has no effect on GKE or AKS.

### Decision summary

| Environment | Custom network? | Constraint                                                    |
| ----------- | --------------- | ------------------------------------------------------------- |
| AWS         | Yes (label set) | Source and destination must be in the same availability zone. |
| AWS         | No              | Source and destination must share the same subnet.            |
| Non-AWS     | Any             | No TCP-specific constraint.                                   |

## Node requirements

Both the source and destination nodes must carry specific labels that opt them into live migration.

### Required on both nodes

```
live.atlas.example/install: "true"
```

This label triggers installation of the live DaemonSet and runtime. The controller only considers nodes with this label as valid source or destination nodes. This is handled automatically by the Atlas autoscaler.

```
live.atlas.example/migration-enabled: "true"
```

When the live DaemonSet and runtime are installed, the node is labeled with this label, indicating that it is a valid node for live migration.

## Pod state

The pod must be in the `Running` phase to be migrated. Migration captures a live process, and there is no live process to capture if the pod has not started yet or has already terminated.

### Pending pods with running init containers

There is one exception: if the pod is `Pending` but has one or more init containers that are already running (not sidecar init containers), the migration waits for the pod to reach `Running`. This is treated as a transient condition and retried rather than refused.

In practice this means a pod that is mid-startup is not refused outright — the migration controller re-checks until the pod finishes initialization or the timeout is hit.

### Other states

| Pod phase                                         | Outcome                                                                      |
| ------------------------------------------------- | ---------------------------------------------------------------------------- |
| `Running`                                         | Eligible (subject to other checks).                                          |
| `Pending` with non-sidecar init container running | Retried until the pod reaches `Running` or the retry budget is exhausted.    |
| `Pending` (no running init containers)            | Retried, then fails with `PodNotRunning` if the pod never reaches `Running`. |
| `Succeeded`, `Failed`, or any terminating state   | Refused.                                                                     |

## Reading refusal reasons

When a migration is refused, the reason is recorded in the Migration object's status and surfaced as an event. The reasons map directly to the sections in this document:

| Refusal reason                         | Section                                                   |
| -------------------------------------- | --------------------------------------------------------- |
| GPU workload detected                  | [GPU workloads](#gpu-workloads)                           |
| Unsupported owner kind                 | [Pod owner types](#pod-owner-types)                       |
| Architecture mismatch or missing label | [CPU architecture](#cpu-architecture)                     |
| CPU capability mismatch                | [CPU capability checks](#cpu-capability-checks)           |
| Destination zone not allowed for PV    | [PVC and zone constraints](#pvc-and-zone-constraints)     |
| Subnet / zone mismatch for TCP         | [TCP and subnet constraints](#tcp-and-subnet-constraints) |
| Required node label missing            | [Node requirements](#node-requirements)                   |
| Pod not in Running phase               | [Pod state](#pod-state)                                   |

Use this table to jump to the relevant section when triaging a refused migration.

## Limitations summary

Here are the workload categories that should make you pause and reconsider before attempting a live migration:

* **GPU workloads** — never eligible.
* **DaemonSets** — by definition one pod per node; nothing to migrate.
* **Custom controllers / CRDs** — refused by default unless you have specifically verified ownership semantics.
* **Mixing node pools with different CPU features on amd64** — works within a pool but migrations across pools may be refused.
* **Mixing amd64 and arm64 nodes** — never works.

# See also

<Cards columns={2}>
  <Card title="Overview" href="https://docs.atlas.example/docs/clm-overview" icon="fa-info-circle">
    What container live migration is and how it works at a high level.
  </Card>
  <Card title="Labels, Annotations, and Events" href="https://docs.atlas.example/docs/clm-reference-labels-and-annotations" icon="fa-tags">
    Reference for the labels and annotations referenced by the eligibility checks.
  </Card>
</Cards>