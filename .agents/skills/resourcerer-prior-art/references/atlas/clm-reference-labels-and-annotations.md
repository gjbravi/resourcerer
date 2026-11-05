---
updatedAt: 2026-09-04T11:55:05.000Z
---

# Labels, Annotations, Taints, and Events

Reference documentation for all labels, annotations, taints, and Kubernetes events used by Container Live Migration.

This reference documents all labels, annotations, taints, and Kubernetes events used by Container Live Migration. Use this page to configure workloads, troubleshoot migrations, or monitor migration status.

# Labels

Labels control or show how Atlas's live controller and Evictor handle workloads during migration operations.

## Node labels

These labels are applied by Atlas to nodes to indicate live migration capability.

| Label                            | Values    | Applied by     | Description                                                                                                         |
| :------------------------------- | :-------- | :------------- | :------------------------------------------------------------------------------------------------------------------ |
| `live.atlas.example/migration-enabled` | `true`    | Atlas        | Indicates the node supports container live migration. Applied to nodes provisioned from CLM-enabled node templates. |
| `live.atlas.example/custom-network`    | `true`    | Atlas        | Indicates the node uses ENIConfig for dedicated Pod subnets. Required for custom networking configurations.         |
| `live.atlas.example/install`           | `true`    | Atlas        | Triggers installation of the migration daemon and runtime on the node.                                              |
| `topology.atlas.example/subnet-id`     | subnet ID | Atlas/daemon | Identifies the node's subnet for VPC CNI path compatibility checks (EKS only).                                      |
| `live.atlas.example/cpuinfo`           | CPU info  | Atlas        | CPU capability information used for migration compatibility checks.                                                 |

## Pod labels

These labels are applied to pods either automatically by the live controller or manually to control migration behavior.

| Label                                         | Values                  | Applied by      | Description                                                                                                                                                                                        |
| :-------------------------------------------- | :---------------------- | :-------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `live.atlas.example/migration-enabled`              | `true`                  | Live controller | Indicates the pod is eligible for live migration. Applied automatically when a workload meets all eligibility requirements.                                                                        |
| `autoscaling.atlas.example/live-migration-disabled` | `true`                  | User            | Opt-out label. Prevents live migration for this pod. The pod will be evicted using traditional eviction instead.                                                                                   |
| `autoscaling.atlas.example/removal-disabled`        | `true`                  | User            | Prevents pod removal during optimization. If migration fails, the pod is recovered on the original node rather than evicted. Can be combined with `live-migration-disabled`.                       |
| `autoscaling.atlas.example/migration-requested-by`  | `rebalancer`, `evictor` | Atlas         | Identifies which Atlas component created the live migration. Set automatically on migration resources to distinguish migrations initiated by the rebalancer from those initiated by the evictor. |

### Opting out of live migration

To disable live migration for specific workloads, add the opt-out label to your pod spec:

```yaml
metadata:
  labels:
    autoscaling.atlas.example/live-migration-disabled: "true"
```

To prevent removal entirely (allowing only migration, with recovery on failure):

```yaml
metadata:
  labels:
    autoscaling.atlas.example/removal-disabled: "true"
```

These labels can be combined. When both are set, the workload is excluded from both live migration and traditional eviction.

# Taints

Atlas applies the following taint to CLM-enabled nodes during provisioning. It prevents workloads from scheduling on the node before the live migration components are fully installed and operational.

## Node taints

| Taint                    | Effect       | Description                                                                                                                                                                                                                                                            |
| :----------------------- | :----------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `live.atlas.example/not-ready` | `NoSchedule` | Prevents workloads from scheduling on the node until the CLM daemon has finished installing and initializing the CRI proxy, mTLS proxy, page server, and file server. Once all components report ready, the daemon removes the taint and the node becomes schedulable. |

### Lifecycle

1. A new node is created with the `live.atlas.example/install=true` label. A mutating admission webhook adds the `live.atlas.example/not-ready` taint with `NoSchedule` effect.
2. The CLM daemon starts on the node and installs its components — CRI proxy, mTLS proxy, page server, and file server.
3. Once all components report ready, the daemon removes the taint. The node becomes schedulable and can receive pods, including live-migrated ones.

<Callout icon="📘" theme="info">
  The taint is removed even when the node operates in degraded mode — for example, if the CRI proxy version check fails. In degraded mode, the node cannot perform live migrations, but it remains available for regular workload scheduling.
</Callout>

For the full list of taints applied across all Atlas components, see [Autoscaler Node Labels and Taints](autoscaler-reference-node-labels-and-taints.md).

# Annotations

Annotations control specific behaviors during the migration restore phase. These are applied to the source pod before migration.

## Restore behavior annotations

By default, init containers, PostStart hooks, and startup probes are skipped on the restored pod because the container state is fully preserved. In rare cases, you may need to override this behavior.

These annotations on the source pod control which init containers, startup probes, and postStart lifecycle hooks re-run when the pod is restored from the checkpoint. By default, the restored process resumes as-is — init containers do not re-run and lifecycle hooks are skipped. You opt in to re-running specific ones with these annotations.

All three accept the same value format:

| Value               | Meaning                             |
| ------------------- | ----------------------------------- |
| `"ALL"`             | Apply to all containers.            |
| `"name1,name2"`     | Apply only to the named containers. |
| `""` (empty/absent) | Apply to none (default behavior).   |

| Annotation                            | Value format                              | Description                                                                                                                                                                         |
| :------------------------------------ | :---------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `atlas.example/restore-run-init-containers` | `"ALL"`, `"name1,name2"`, or empty/absent | Controls which init containers run fresh during restore. Init containers that are not in this list are skipped — their state is already captured in the checkpoint.                 |
| `atlas.example/restore-run-startup-probe`   | `"ALL"`, `"name1,name2"`, or empty/absent | Controls which containers' startup probes re-execute on restore. By default, all startup probes are skipped — the restored container is already running and past the startup phase. |
| `atlas.example/restore-run-post-start-exec` | `"ALL"`, `"name1,name2"`, or empty/absent | Controls which containers' `lifecycle.postStart` hooks re-execute on restore. By default, all postStart hooks are skipped. Only affects exec-based hooks.                           |

### When to use each annotation

#### `atlas.example/restore-run-init-containers`

* **When to use:** Your pod has init containers that must re-run on restore (e.g. config bootstrap, iptables setup, filesystem preparation). Alternatively, keep it absent if your init container is not idempotent and you want to skip it on restore (the default behavior).
* **When it does NOT make sense:** Your init containers only do image-pulling or one-time setup already reflected in the checkpointed filesystem. Re-running them is wasteful or harmful.

#### `atlas.example/restore-run-startup-probe`

* **When to use:** Your startup probe performs a check that must re-validate after restore (e.g. verifying external dependencies are reachable from the new node).
* **When it does NOT make sense (default):** The startup probe only gates initial container readiness. The restored container is already past startup. Re-running it can cause a false failure if the probe target is transiently unavailable.

#### `atlas.example/restore-run-post-start-exec`

* **When to use:** Your postStart hook performs registration or initialization that must happen on the new node (e.g. re-registering with a service mesh, re-opening a sidecar channel).
* **When it does NOT make sense (default):** The postStart hook does one-time setup already reflected in the checkpoint. Re-running it may conflict with the already-running process state.

### Example: Forcing startup probe execution

```yaml
metadata:
  annotations:
    atlas.example/restore-run-startup-probe: "container1,container2"
```

### Example: Forcing init container execution

```yaml
metadata:
  annotations:
    atlas.example/restore-run-init-containers: "init-config,init-deps"
```

<Callout icon="⚠️" theme="warn">
  ### Use with caution

  These annotations should only be used when necessary:

  - Re-running PostStart hooks can cause duplicate side effects (duplicate service registrations, duplicate cache entries)
  - Re-running startup probes delays pod readiness
  - Re-running init containers increases migration time and can cause errors

  For most applications, the default behavior (skipping these components) is correct and considered best practice.
</Callout>

# Migration options

These options control individual migration behavior — TCP preservation, memory dump tuning, pod replacement, and restore hooks. They can be set per-migration, per-pod, or as cluster-wide defaults.

For cluster-wide Helm chart values that configure components, scheduling, and deployment, see the chart's `values.yaml`.

## How options are resolved

Most migration options can be set at three levels. The system checks them in order — the first value set wins:

1. **Migration CR** — options on the `Migration` object created for a specific migration. This is the per-migration override.
2. **Pod annotation** — annotations on the source pod being migrated. This is the per-pod override; it only takes effect when the Migration CR does not already specify the option.
3. **Helm chart default** — cluster-wide defaults in the Helm values. This is the fallback when neither the CR nor the pod annotation sets the option.

Once resolved, the effective values are written back into the Migration object and a `MigrationReconfigured` event is emitted so you can audit what was applied.

<Callout icon="⚠️" theme="warn">
  Pod annotations only override the cluster default for specific pods. If a Migration CR explicitly sets an option, the pod annotation is ignored.
</Callout>

## TCP/IP stack options

These control whether and how established TCP connections survive migration. For the conceptual overview, see [TCP migration modes](clm-reference-tcp-migration-modes.md). For the TC and VPC CNI implementation details, see [Traffic Control](clm-reference-traffic-control.md) and [AWS VPC CNI](clm-reference-aws-vpc-cni.md).

For detailed configuration reference for each TCP preservation path, see [Traffic Control](clm-reference-traffic-control.md) and [AWS VPC CNI](clm-reference-aws-vpc-cni.md).

### `withTCP`

| Where      | Key                                         | Default                                                            |
| ---------- | ------------------------------------------- | ------------------------------------------------------------------ |
| Helm       | `features.migrationOptions.default.withTCP` | Autodiscovered: `true` if AWS VPC CNI is detected, otherwise empty |
| Annotation | `live.atlas.example/migration-default-with_tcp`   | —                                                                  |
| CR         | `spec.options.with_tcp`                     | —                                                                  |

Controls whether the migration attempts to preserve established TCP sockets.

**When it makes sense:**

* Your workload has long-lived TCP connections (databases, message queues, gRPC streams) that must survive migration.
* You are running on EKS with AWS VPC CNI (the chart autodiscovers this and defaults to `true`), or you have the TC DaemonSet enabled.

**When it does NOT make sense:**

* Your application tolerates reconnects and you want the simplest, fastest migration. Set `withTCP: false`.
* You are on a cluster where neither the VPC CNI path nor the TC DaemonSet is available. TCP state will be checkpointed but cannot be meaningfully restored because the pod IP changes and peers have no translation mechanism.

### `withTCPDnatRewrite`

| Where      | Key                                                    | Default                                        |
| ---------- | ------------------------------------------------------ | ---------------------------------------------- |
| Helm       | `features.migrationOptions.default.withTCPDnatRewrite` | `true` (only applied when `withTCP` is `true`) |
| Annotation | `live.atlas.example/migration-default-with_tcp_dnat_rewrite` | —                                              |
| CR         | `spec.options.with_tcp_dnat_rewrite`                   | —                                              |

When `true`, Kubernetes Service destination IPs in the checkpoint are rewritten to the real backend pod IPs so connections survive the migration. Without this, connections to Service IPs may route differently after the pod moves.

**When it makes sense:**

* Your pod talks to other pods through Kubernetes `ClusterIP` Services. Enabled by default for this reason.

**When it does NOT make sense:**

* Your pod connects directly to pod IPs (headless Services, direct pod-to-pod). The rewrite is harmless but unnecessary. There is rarely a reason to disable it.

### `withTCPSnat`

| Where      | Key                                             | Default   |
| ---------- | ----------------------------------------------- | --------- |
| Helm       | `features.migrationOptions.default.withTCPSnat` | `"close"` |
| Annotation | `live.atlas.example/migration-default-with_tcp_snat`  | —         |
| CR         | `spec.options.with_tcp_snat`                    | —         |

Controls what happens to connections created through Source NAT (SNAT) — typically outbound connections to external endpoints.

* `"close"` — SNATed connections are closed before restore. The application must reconnect after migration.
* Any other value (e.g. `"keep"`) — SNATed connections are kept in the checkpoint as-is.

**When `"close"` makes sense (default):**

* Almost always. SNATed connections tie back to the source node's IP. After migration, the source node changes and the SNAT state is no longer valid. Keeping them usually results in broken connections that the application cannot detect cleanly.

**When `"keep"` makes sense:**

* Rarely. Only if you have verified that the source IP and SNAT state are stable across migration (e.g. an egress proxy with a stable IP). Keeping stale SNATed connections will cause silent connection failures.

## Iterative dump tuning

Live migration uses iterative pre-dumps to reduce the freeze time: the pod keeps running while successive memory snapshots are taken, and only the final dump freezes the process. These options control how many iterations are done and when to stop.

### `iterations`

| Where      | Key                                            | Default                                          |
| ---------- | ---------------------------------------------- | ------------------------------------------------ |
| Helm       | `features.migrationOptions.default.iterations` | `10` (if set to `0`, the chart defaults to `10`) |
| Annotation | — (not available)                              | —                                                |
| CR         | `spec.options.iterations`                      | —                                                |

The maximum number of pre-dump iterations before the final freeze-and-dump. The actual number may be lower if the dirty-pages threshold is reached first.

<Callout icon="⚠️" theme="warn">
  **Special case — ARM nodes:** On `arm64` nodes, iterations are forced to `0` (single dump, no pre-dump) because ARM64 kernels historically lack the soft-dirty bit support required for incremental memory dumps. This cannot be overridden.
</Callout>

**When a higher value makes sense:**

* The pod has a large memory footprint with moderate write churn. More iterations give more chances to converge on a small set of dirty pages, reducing freeze time.

**When a lower value makes sense:**

* The pod has very high memory churn (it dirties pages faster than they can be tracked). Additional iterations will not converge and just add total migration time. A lower value or `optimizeFor: DumpTime` is better.
* The pod is small or low-traffic. The default 10 is usually fine, but you can lower it to reduce total migration time.

### `dirtyPagesThreshold`

| Where      | Key                                                     | Default |
| ---------- | ------------------------------------------------------- | ------- |
| Helm       | `features.migrationOptions.default.dirtyPagesThreshold` | `12000` |
| Annotation | — (not available)                                       | —       |
| CR         | `spec.options.dirty_pages_threshold`                    | —       |

The threshold (in 4 KB pages) below which the system decides the remaining dirty pages are small enough to do the final freeze-and-dump. This acts as the early-exit condition for the iterative dump strategy.

When `optimizeFor` is `FreezeTime` (the default), the system first tries to compute a statistical threshold from historical dirty-page stats. This Helm/CR value serves as the fallback if the statistical calculation fails (e.g. no stats available yet).

**When to adjust:**

* **Lower** if you want stricter convergence (fewer dirty pages at final dump → shorter freeze, but more iterations and longer total time).
* **Higher** if you want to finish faster at the cost of a slightly longer freeze. The freeze time is roughly proportional to the number of dirty pages that must be written during the final dump.

**When it does NOT matter:**

* When `optimizeFor` is `DumpTime`, this threshold is unused — the fast-dump strategy does a single dump with no iteration.

### `optimizeFor`

| Where      | Key                                           | Default                                         |
| ---------- | --------------------------------------------- | ----------------------------------------------- |
| Helm       | — (not configurable via Helm)                 | —                                               |
| Annotation | `live.atlas.example/migration-default-optimize_for` | —                                               |
| CR         | `spec.options.optimize_for`                   | `"FreezeTime"` (applied automatically if unset) |

Selects the dump strategy. Two values:

| Value                  | Behavior                                                                                                                                       |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `FreezeTime` (default) | Iterative pre-dumps with early stopping based on dirty-page threshold. Minimizes the freeze window at the cost of longer total migration time. |
| `DumpTime`             | Single dump, no iterations. Minimizes total migration time but the freeze window equals the full dump time.                                    |

**When `FreezeTime` makes sense (default):**

* The application is latency-sensitive and must minimize the freeze window. This is the common case for user-facing services.

**When `DumpTime` makes sense:**

* The application tolerates a longer freeze (e.g. batch jobs, background workers) and you want to minimize total migration time and network transfer.
* The pod has very high memory churn where iterative pre-dumps never converge.

<Callout icon="📘" theme="info">
  `optimizeFor` is settable via the Migration CR or the pod annotation `live.atlas.example/migration-default-optimize_for` (values: `FreezeTime`, `DumpTime`). There is no Helm value for it — this is a per-workload tradeoff, not a cluster-wide default. An invalid annotation value is logged as a warning and treated as absent; resolution falls through to the CR, then to `FreezeTime`.
</Callout>

## Requested freeze time

### `requestedFreezeTime`

| Where      | Key                                                                   | Default       |
| ---------- | --------------------------------------------------------------------- | ------------- |
| Helm       | `features.migrationOptions.default.requestedFreezeTimeDefaultSeconds` | `5` (seconds) |
| Annotation | `live.atlas.example/migration-requested-freeze-time`                        | —             |
| CR         | `spec.options.requested_freeze_time`                                  | —             |

The expected freeze time for the migration. This is not a hard limit on the actual freeze — it is an estimate used by the scheduler to decide whether a migration can proceed given the available network bandwidth.

The scheduler estimates the migration data transfer size, divides it by the requested freeze time to get the required bandwidth, and checks whether the source and destination nodes have enough available bandwidth. If not, the migration is held until capacity frees up.

The annotation accepts a Go duration string (e.g. `"5s"`, `"10s"`, `"500ms"`).

**When a lower value makes sense:**

* The workload is latency-sensitive and you want the scheduler to be conservative — it will only start migrations it estimates can finish within a tight freeze window. Be aware that this may cause migrations to be deferred if bandwidth is limited.

**When a higher value makes sense:**

* The workload tolerates a longer freeze and you want migrations to proceed more readily (the scheduler will accept migrations with larger estimated transfer sizes). This is appropriate for batch or background workloads.

**When the default (5s) is fine:**

* For most general-purpose workloads. Only adjust if the scheduler is either too aggressively deferring or too aggressively starting migrations.

## Pod replacement policy

### `podReplacementPolicy`

| Where      | Key                                                      | Default                               |
| ---------- | -------------------------------------------------------- | ------------------------------------- |
| Helm       | `features.migrationOptions.default.podReplacementPolicy` | `""` (empty = auto-resolve per owner) |
| Annotation | `live.atlas.example/migration-default-pod_replacement_policy`  | —                                     |
| CR         | `spec.options.pod_replacement_policy`                    | —                                     |

Controls how the source pod is replaced by the restored clone. Three values:

| Policy       | Clone name           | Source deleted        | Recoverable? |
| ------------ | -------------------- | --------------------- | ------------ |
| `SameName`   | Same as source       | Before clone creation | ❌            |
| `EvictFirst` | `<pod-name>-clone-N` | Before clone creation | ❌            |
| `KeepSource` | `<pod-name>-clone-N` | After clone is ready  | ✅            |

**Resolution:** first non-empty wins, highest precedence first:

1. Pod annotation `live.atlas.example/migration-default-pod_replacement_policy` (values: `SameName`, `EvictFirst`, `KeepSource`).
2. CR `spec.options.pod_replacement_policy` (per-migration override).
3. Helm default `features.migrationOptions.default.podReplacementPolicy`.
4. Per-owner default: `StatefulSet` → `SameName`; all others → `KeepSource`.

An invalid annotation value is logged as a warning and treated as absent; resolution falls through to the next tier.

<Callout icon="⚠️" theme="warn">
  `StatefulSet` → always coerced to `SameName` (StatefulSet cannot share identity with a `-clone-N` sibling). This coercion runs after the tiers above, so an annotation on a StatefulSet pod is overridden.
</Callout>

For the full migration path matrix (PVC impact, pre-creation, controller blocking), see [Pod replacement policy](#pod-replacement-policy) above.

**When `SameName` makes sense:**

* StatefulSets (enforced). The clone takes the source pod's name so the StatefulSet controller recognizes it as the same ordinal.
* Any workload that depends on the pod name (e.g. DNS-based discovery via pod name).

**When `EvictFirst` makes sense:**

* You want the source gone before the clone starts to free resources immediately. This is the fastest cleanup path but is not recoverable if the clone fails.

**When `KeepSource` makes sense (default for non-StatefulSet):**

* You want safety: the source pod stays alive until the clone is confirmed ready. If the clone fails, the original is recovered. This is the safest option for production workloads.

## System annotations

These annotations are applied by Atlas during the migration process. Do not modify these manually.

| Annotation        | Applied by      | Description                                                                                                                     |
| :---------------- | :-------------- | :------------------------------------------------------------------------------------------------------------------------------ |
| `atlas.example/restore` | Live controller | JSON restore request with TCP config, container mappings, and peer IPs. Set by the controller, removed after restore completes. |

If you are unfamiliar with **clone pods** or **destination nodes**, please refer to the [Overview](clm-overview.md) document.

## Internal annotations and labels (do not set manually)

The following are set and managed internally by the controller during migration.

<Callout icon="⚠️" theme="warn">
  Do not set these manually — they carry per-migration state and setting them incorrectly can break migration or recovery. They are listed here for awareness and debugging.
</Callout>

### On the restored (clone) pod

| Annotation                                | Purpose                                                                                                                         |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `atlas.example/restore`                         | JSON restore request with TCP config, container mappings, and peer IPs. Set by the controller, removed after restore completes. |
| `atlas.example/migration_reference`             | Name of the Migration object that created this pod.                                                                             |
| `live.atlas.example/ip`                         | The pod IP to restore (set when `withTCP` is enabled).                                                                          |
| `live.atlas.example/move-out-ip`                | Marker to move the IP out of the source (set when `withTCP` is enabled).                                                        |
| `live.atlas.example/restored-pod-name`          | Name of the restored clone. Presence indicates the pod was created by a migration.                                              |
| `live.atlas.example/pod-pre-created-in-running` | Indicates the clone was pre-created during the final dump phase.                                                                |

### On the Migration object

| Annotation / Label                          | Purpose                                                                                               |
| ------------------------------------------- | ----------------------------------------------------------------------------------------------------- |
| `live.atlas.example/original-pod-state`           | JSON snapshot of the original pod spec, used for recovery if the clone fails.                         |
| `live.atlas.example/block-pod-controller-actions` | Blocks the pod's controller from issuing new pods during migration. Removed when migration completes. |
| `live.atlas.example/capacity-pod-template`        | Records the original pod name for capacity-probe cleanup.                                             |

### Other internal labels

| Label                                  | Purpose                                                                                                         |
| -------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `live.atlas.example/capacity-pod`            | Identifies pods created as capacity probes during the `WaitingForCapacity` state.                               |
| `live.atlas.example/tracking_id`             | Internal tracing metadata for test and benchmark runs. Not used in production.                                  |
| `autoscaling.atlas.example/removal-disabled` | Marks a pod as protected from removal by the Atlas autoscaler. The controller respects this during migration. |

## Quick reference

### Per-migration options (CR → annotation → Helm default)

| Option                 | CR field                            | Pod annotation                                          | Helm default                                                          | Default value  |
| ---------------------- | ----------------------------------- | ------------------------------------------------------- | --------------------------------------------------------------------- | -------------- |
| With TCP               | `with_tcp`                          | `live.atlas.example/migration-default-with_tcp`               | `features.migrationOptions.default.withTCP`                           | Autodiscovered |
| TCP DNAT rewrite       | `with_tcp_dnat_rewrite`             | `live.atlas.example/migration-default-with_tcp_dnat_rewrite`  | `features.migrationOptions.default.withTCPDnatRewrite`                | `true`         |
| TCP SNAT handling      | `with_tcp_snat`                     | `live.atlas.example/migration-default-with_tcp_snat`          | `features.migrationOptions.default.withTCPSnat`                       | `"close"`      |
| Iterations             | `iterations`                        | —                                                       | `features.migrationOptions.default.iterations`                        | `10`           |
| Dirty pages threshold  | `dirty_pages_threshold`             | —                                                       | `features.migrationOptions.default.dirtyPagesThreshold`               | `12000`        |
| Optimize for           | `optimize_for`                      | `live.atlas.example/migration-default-optimize_for`           | — (CR only)                                                           | `"FreezeTime"` |
| Requested freeze time  | `requested_freeze_time`             | `live.atlas.example/migration-requested-freeze-time`          | `features.migrationOptions.default.requestedFreezeTimeDefaultSeconds` | `5s`           |
| Pod replacement policy | `pod_replacement_policy`            | `live.atlas.example/migration-default-pod_replacement_policy` | `features.migrationOptions.default.podReplacementPolicy`              | `""` (auto)    |
| Fail on Istio inbound  | `fail_on_inbound_istio_connections` | —                                                       | — (auto-resolved)                                                     | Auto           |

### Restore control annotations (source pod only)

| Annotation                            | Value format             | Default            |
| ------------------------------------- | ------------------------ | ------------------ |
| `atlas.example/restore-run-init-containers` | `"ALL"`, `"name1,name2"` | `""` (none)        |
| `atlas.example/restore-run-startup-probe`   | `"ALL"`, `"name1,name2"` | `""` (all skipped) |
| `atlas.example/restore-run-post-start-exec` | `"ALL"`, `"name1,name2"` | `""` (all skipped) |

# Migration events

Live migration progress is tracked through Kubernetes events on the migration custom resource. Use these events to monitor migration status and troubleshoot failures.

## Viewing migration events

```shell
# List all migrations
kubectl get migrations -A

# Watch migrations in real-time
kubectl get migrations -A -w

# Get detailed events for a specific migration
kubectl describe migration <migration-name> -n <namespace>
```

## Event reference

| Event                   | Description                                                                                             | Phase    |
| :---------------------- | :------------------------------------------------------------------------------------------------------ | :------- |
| `PreDumpSkipped`        | Memory pre-dump iteration was skipped.                                                                  | Pre-dump |
| `PreDumpFinished`       | Memory pre-dump completed successfully. Dirty pages captured for incremental transfer.                  | Pre-dump |
| `PreDumpFailed`         | Memory pre-dump operation failed. Check logs for details.                                               | Pre-dump |
| `MigrationReconfigured` | Migration configuration was updated during the operation.                                               | Any      |
| `PodCreateFinished`     | Pod successfully recreated on the destination node.                                                     | Restore  |
| `PodCreateFailed`       | Failed to create pod on destination node. Check destination node capacity and permissions.              | Restore  |
| `MigrationFinished`     | Pod successfully migrated to the new node. Migration complete.                                          | Complete |
| `MigrationFailed`       | Migration failed. Check logs for details. Evictor will fall back to traditional eviction if configured. | Failed   |

## Migration phases

Events occur during specific phases of the migration process:

1. **Pre-dump phase**: Iterative memory snapshots while the container continues running. `PreDumpFinished` events indicate successful iterations. Multiple pre-dumps reduce final checkpoint time.

2. **Checkpoint phase**: Final memory dump and state capture. No specific event due to the briefness of this phase.

3. **Restore phase**: Container state restoration on the destination node. `PodCreateFinished` indicates the pod was created successfully.

4. **Completion**: `MigrationFinished` indicates the entire migration succeeded and the pod is running on the new node.

# Pod naming

After migration, pods receive modified names to indicate they are restored clones:

| Workload type   | Naming pattern              | Example                                                 |
| :-------------- | :-------------------------- | :------------------------------------------------------ |
| Deployments     | `<original-name>-clone-<N>` | `my-app-7d4b8c9-abc12` → `my-app-7d4b8c9-abc12-clone-1` |
| StatefulSets    | Original name preserved     | `my-db-0` → `my-db-0`                                   |
| Other workloads | `<original-name>-clone-<N>` | Follows deployment pattern                              |

StatefulSets retain their original pod names because their identity (and associated PVCs) must remain stable.

# Common kubectl commands

This section provides commonly used commands for working with Container Live Migration. Each command helps you verify configuration, check eligibility, or monitor migration activity.

## Check if a node supports live migration

To verify which nodes in your cluster are configured for Container Live Migration, query for the migration-enabled label. Only nodes provisioned from CLM-enabled node templates will have this label applied.

```shell
kubectl get nodes -l live.atlas.example/migration-enabled=true
```

If no nodes are returned, either CLM is not enabled in your node templates, or you need to trigger a rebalancing operation to replace existing nodes with CLM-enabled ones.

## Check if a pod is eligible for live migration

The live controller automatically labels pods that meet all eligibility requirements. Use this command to see which pods in your cluster are currently eligible for live migration.

```shell
kubectl get pods -A -l live.atlas.example/migration-enabled=true
```

Pods that do not appear in this list either fail to meet the technical requirements or are running on nodes where the live controller has not yet completed its assessment.

## Opt a workload out of live migration

If you need to exclude a specific workload from live migration while still allowing traditional eviction, add the opt-out label to your pod specification. This is useful for workloads that have known compatibility issues or require additional testing before enabling migration.

```yaml
metadata:
  labels:
    autoscaling.atlas.example/live-migration-disabled: "true"
```

## Monitor active migrations

During a rebalancing operation or when Evictor triggers migrations, you can watch migration resources in real-time to observe progress and catch any failures as they occur.

```shell
kubectl get migrations -A -w
```

The `-w` flag keeps the command running and displays updates as migration status changes. Look for `MigrationFinished` events to confirm successful migrations, or `MigrationFailed` events that may require investigation.

# See also

<Cards>
  <Card title="Container Live Migration Overview" href="https://docs.atlas.example/docs/clm-overview" icon="fa-info-circle">
    Introduction to CLM concepts, benefits, and how it integrates with Evictor.
  </Card>

  <Card title="Requirements and Limitations" href="https://docs.atlas.example/docs/clm-requirements-and-limitations" icon="fa-list-check">
    Infrastructure requirements, supported workloads, and known constraints.
  </Card>

  <Card title="Getting Started with CLM" href="https://docs.atlas.example/docs/clm-getting-started" icon="fa-rocket">
    Step-by-step guide to enabling and testing live migration in your cluster.
  </Card>

  <Card title="Probe and Lifecycle Behavior" href="https://docs.atlas.example/docs/clm-concept-probe-lifecycle-behavior" icon="fa-heartbeat">
    How Kubernetes probes and lifecycle hooks behave during migration.
  </Card>

  <Card title="Cloud Providers" href="https://docs.atlas.example/docs/clm-cloud-providers" icon="fa-cloud">
    Per-cloud setup guides for EKS, GKE, and AKS.
  </Card>
</Cards>

<br />