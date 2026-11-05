---
updatedAt: 2026-09-18T14:38:37.000Z
---

# Configuration

Configure PVPolicy Custom Resources and PVC annotations for storage capacity, IOPS, and throughput autoscaling.

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

## Overview

This page documents the PVPolicy Custom Resource fields, PVC annotations, and Helm values that control Storage Autoscaling behavior. For background on how the configuration layers interact, including the default PVPolicy and policy resolution, see [How it works](storage-how-it-works.md#configuration-model).

PVC annotations override the matched policy on a per-PVC basis.

## PVPolicy Custom Resource

PVPolicy is a namespaced Custom Resource Definition in the `storage.atlas.example/v1alpha1` API group.

| Attribute   | Value                      |
| ----------- | -------------------------- |
| API version | `storage.atlas.example/v1alpha1` |
| Kind        | `PVPolicy`                 |
| Short name  | `pvpol`                    |
| Scope       | Namespaced                 |

Create PVPolicy CRs in the `atlas-agent` namespace using `kubectl apply`.

### PVPolicy spec

| Field           | Type    | Required | Default | Description                                                                                                                            |
| --------------- | ------- | -------- | ------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `spec.priority` | integer | Yes      | —       | Priority for conflict resolution. Higher values take precedence. Minimum `1`.                                                          |
| `spec.dryRun`   | boolean | Yes      | —       | When `true`, planned actions are logged but no changes are made for PVCs matched by this policy.                                       |
| `spec.selector` | object  | Yes      | —       | Defines which PVCs this policy applies to. See [PVC selector](#pvc-selector).                                                          |
| `spec.config`   | object  | Yes      | —       | Autoscaling configuration. Omitted fields inherit global defaults from Helm values. See [Policy configuration](#policy-configuration). |

## PVC selector

The `spec.selector` field determines which PVCs a policy applies to. It has two active subfields: `namespaces` and `pvc`.

### Namespace filter

`spec.selector.namespaces` is a list of namespace names. When empty, all namespaces are eligible. When non-empty, only PVCs in the listed namespaces are considered.

### PVC matcher

`spec.selector.pvc` defines which PVCs match by name, label, or StorageClass. An empty object (`pvc: {}`) matches all PVCs in the selected namespaces. When omitted entirely, no PVCs are matched through this selector.

| Field            | Type   | Default | Description                                                                                                                             |
| ---------------- | ------ | ------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `names`          | array  | `[]`    | List of PVC names. A PVC must be in this list to match. Empty matches any name.                                                         |
| `labels`         | object | —       | Kubernetes label selector. Use `matchLabels` for exact key-value pairs and `matchExpressions` for `In`/`NotIn`/`Exists`/`DoesNotExist`. |
| `storageClasses` | array  | `[]`    | List of StorageClass names. The PVC's `spec.storageClassName` must be in this list. Empty matches any StorageClass.                     |

All non-empty fields are AND-ed: a PVC must satisfy every specified condition.

> 📘 Note
> PVCs with no `storageClassName` set do not match any `storageClasses` filter. If you need to manage these PVCs, use the `names` or `labels` fields instead.

> 📘 Note
> Workload-based matching (`spec.selector.workload`) is planned for a future release. If set, it is currently ignored.

## Policy configuration

The `spec.config` field contains the autoscaling settings applied to matched PVCs. The `dimensions` block is required. The `storage` and `performance` sub-objects are optional. When you omit a field, it inherits the corresponding global default from Helm values.

### Dimensions

`spec.config.dimensions` controls which autoscaling dimensions are enabled for matched PVCs.

| Field               | Type    | Default | Description                                           |
| ------------------- | ------- | ------- | ----------------------------------------------------- |
| `storageEnabled`    | boolean | —       | Enable automatic capacity expansion for matched PVCs. |
| `iopsEnabled`       | boolean | —       | Enable IOPS rightsizing for matched PVCs.             |
| `throughputEnabled` | boolean | —       | Enable throughput rightsizing for matched PVCs.       |

### Storage settings

`spec.config.storage` configures capacity expansion behavior.

| Field              | Type    | Default | Description                                                                                                                    |
| ------------------ | ------- | ------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `thresholdPercent` | integer | `80`    | Storage usage percentage that triggers expansion (0–100).                                                                      |
| `thresholdPeriod`  | string  | `"5m"`  | How long the threshold must be exceeded before acting. Prevents temporary spikes from triggering unnecessary resizes.          |
| `growthPercent`    | integer | `20`    | How much to grow the PVC by, as a percentage of the current size.                                                              |
| `minGrowthStepGiB` | integer | `5`     | Minimum growth increment in GiB. Applied when the percentage-based increment would be smaller.                                 |
| `cooldownPeriod`   | string  | `"30m"` | Minimum time between expansions on a single PVC. Keep at 10 minutes or more to allow time for filesystem resizing to complete. |

### Performance settings

`spec.config.performance` configures IOPS and throughput rightsizing behavior.

| Field             | Type   | Default | Description                                                                                                                                                                              |
| ----------------- | ------ | ------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `target`          | string | `"max"` | Which observation to use for the target computation. `"max"` uses the peak value over the lookback window. Also accepts percentile strings such as `"p50"`, `"p95"`, or `"p99.9"`.       |
| `overhead`        | number | `0.2`   | Buffer added on top of the observed value: `target = observed × (1 + overhead)`. Valid range is 0.0–2.5.                                                                                 |
| `lookbackPeriod`  | string | `"6h"`  | How far back to look at cloud metrics when computing the target. Valid range is 1h–24h.                                                                                                  |
| `minWarmupWindow` | string | `"30m"` | Minimum metric data that must accumulate before acting on a disk. Valid range is 15m–24h. Set to `"0"` to disable.                                                                       |
| `changeCooldown`  | string | `"15m"` | Minimum time between consecutive performance modifications on the same PVC per dimension. Stacks on top of cloud-enforced cooldowns. Set to `"0"` to rely only on cloud-enforced limits. |

## Helm values (global defaults)

Helm values provide global defaults that PVPolicy CRs inherit when their config fields are omitted. You can pass a values file with `-f values.yaml` or use `--set` for individual overrides.

PVPolicy fields mirror the storage and performance settings documented above. When you set these values in Helm, they serve as the baseline that policies inherit from. For tuning specific PVCs or namespaces, prefer creating PVPolicy CRs over overriding Helm values.

### Controller settings

These settings are specific to Helm and do not appear in PVPolicy CRs.

| Key                                    | Description                                                                                                                                                                                    | Default |
| -------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| `config.dryRun`                        | Sets the `dryRun` value of the default PVPolicy on first creation. After the policy exists, this value has no effect. Manage `dryRun` directly on the PVPolicy using `kubectl patch`.          | `true`  |
| `config.allowUnknownCloudProvider`     | Allow the chart to run on clusters where the cloud provider cannot be detected (for example, kind or k3s). Capacity expansion works; IOPS rightsizing is disabled. For local development only. | `false` |
| `config.manageVACPVCPerformance`       | When false, PVCs referencing a VolumeAttributesClass are skipped for performance management. Enable only when you have confirmed no conflict with the VAC controller.                          | `false` |
| `config.storage.checkInterval`         | How often disk usage metrics are polled.                                                                                                                                                       | `"1m"`  |
| `config.storage.cloudDiskSyncCooldown` | Minimum time between periodic cloud disk state refreshes. Set to `"0"` to disable periodic sync.                                                                                               | `"15m"` |

### Namespace and label filtering

These Helm-level filters apply globally. PVPolicy selectors are the preferred way to scope management.

| Key                          | Description                                                             | Default |
| ---------------------------- | ----------------------------------------------------------------------- | ------- |
| `config.namespaces.watchAll` | Monitor all namespaces. Set to `false` to scope to specific namespaces. | `true`  |
| `config.namespaces.watch`    | List of namespaces to watch when `watchAll` is `false`.                 | `[]`    |
| `config.pvc.includeLabels`   | Only manage PVCs with all of these labels.                              | `{}`    |
| `config.pvc.excludeLabels`   | Exclude PVCs with any of these labels.                                  | `{}`    |

### Global PVC defaults

These values apply when a PVPolicy's dimension fields are omitted. They can also be overridden per-PVC via annotations.

| Key                                    | Description                                                  | Default |
| -------------------------------------- | ------------------------------------------------------------ | ------- |
| `config.pvcDefaults.storageEnabled`    | Enable automatic capacity expansion by default for all PVCs. | `true`  |
| `config.pvcDefaults.iopsEnabled`       | Enable IOPS rightsizing by default for all PVCs.             | `true`  |
| `config.pvcDefaults.throughputEnabled` | Enable throughput rightsizing by default for all PVCs.       | `true`  |

## PVPolicy examples

### Example 1: Default policy (automatically created by Helm)

The `default` PVPolicy is created automatically on controller pod start, if it does not yet exist. It is **not** updated later by the controller, so any configuration changes diverging from the default profile should be done in the PVPolicy resource.

```yaml
apiVersion: storage.atlas.example/v1alpha1
kind: PVPolicy
metadata:
  name: default
  namespace: atlas-agent
spec:
  priority: 1
  dryRun: true
  selector:
    namespaces: []
    pvc: {}
  config:
    dimensions:
      storageEnabled: true
      iopsEnabled: true
      throughputEnabled: true
    storage:
      thresholdPercent: 80
      thresholdPeriod: "5m"
      growthPercent: 20
      minGrowthStepGiB: 5
      cooldownPeriod: "30m"
    performance:
      target: "max"
      overhead: 0.2
      lookbackPeriod: "6h"
      minWarmupWindow: "30m"
      changeCooldown: "15m"
```

### Example 2: Scope to a specific namespace with custom thresholds

This policy matches PVCs in the `production` namespace, uses a 90% threshold and 25% growth, and runs in live mode. Because it has a higher priority than the default policy, it takes precedence for production PVCs.

```yaml
apiVersion: storage.atlas.example/v1alpha1
kind: PVPolicy
metadata:
  name: production
  namespace: atlas-agent
spec:
  priority: 2
  dryRun: false
  selector:
    namespaces:
      - production
    pvc: {}
  config:
    dimensions:
      storageEnabled: true
      iopsEnabled: true
      throughputEnabled: true
    storage:
      thresholdPercent: 90
      growthPercent: 25
```

### Example 3: Match by label and StorageClass

This policy matches PVCs with the label `app=database` and the StorageClass `gp3`, and disables IOPS rightsizing for them.

```yaml
apiVersion: storage.atlas.example/v1alpha1
kind: PVPolicy
metadata:
  name: database-gp3
  namespace: atlas-agent
spec:
  priority: 3
  dryRun: false
  selector:
    namespaces: []
    pvc:
      labels:
        matchLabels:
          app: database
      storageClasses:
        - gp3
  config:
    dimensions:
      storageEnabled: true
      iopsEnabled: false
      throughputEnabled: true
```

## PVC annotations

### Overview

Annotations override the matched policy's settings on a per-PVC basis. The applicable PVPolicy is resolved first, and then annotation overrides are applied on top.

Set a `*-enabled` annotation to `"true"` only if you have disabled the feature in the matched policy and want to re-enable it for a specific PVC.

### Annotations you set

| Annotation                                  | What it does                                                                                     | Example value |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------ | ------------- |
| `storage.atlas.example/storage-enabled`           | Enable or disable capacity monitoring for this PVC.                                              | `"false"`     |
| `storage.atlas.example/storage-threshold-percent` | Override the policy threshold percent.                                                           | `"90"`        |
| `storage.atlas.example/storage-growth-percent`    | Override the policy growth percent.                                                              | `"30"`        |
| `storage.atlas.example/storage-size-max`          | Maximum capacity this PVC may be grown to.                                                       | `"500Gi"`     |
| `storage.atlas.example/storage-cooldown`          | Override the policy `cooldownPeriod` for this PVC.                                               | `"1h"`        |
| `storage.atlas.example/desired-capacity`          | Expand the PVC to at least this size.                                                            | `"500Gi"`     |
| `storage.atlas.example/iops-enabled`              | Enable or disable IOPS management for this PVC.                                                  | `"false"`     |
| `storage.atlas.example/iops-max`                  | Cap the maximum IOPS that may be provisioned.                                                    | `"80000"`     |
| `storage.atlas.example/desired-iops`              | Pin IOPS to a specific value, overriding automatic scaling.                                      | `"10000"`     |
| `storage.atlas.example/throughput-enabled`        | Enable or disable throughput management for this PVC.                                            | `"false"`     |
| `storage.atlas.example/throughput-max-mibps`      | Cap the maximum throughput that may be provisioned.                                              | `"1200"`      |
| `storage.atlas.example/desired-throughput-mibps`  | Pin throughput to a specific value, overriding automatic scaling.                                | `"500"`       |
| `storage.atlas.example/performance-cooldown`      | Override the policy `changeCooldown` for this PVC. Stacks on top of the cloud-enforced cooldown. | `"30m"`       |

### Status annotations

Storage Autoscaling writes status annotations back to PVCs so you can inspect current state and troubleshoot issues.

| Annotation                                     | What it shows                                                                                                                                 |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `storage.atlas.example/matched-pvpolicy`             | Name of the PVPolicy that matched this PVC.                                                                                                   |
| `storage.atlas.example/dry-run`                      | Whether the matched policy is in dry-run mode (`"true"` or `"false"`).                                                                        |
| `storage.atlas.example/disk-type`                    | Detected cloud disk type (for example, `hyperdisk-balanced` or `gp3`).                                                                        |
| `storage.atlas.example/current-iops`                 | Last known provisioned IOPS from the cloud API.                                                                                               |
| `storage.atlas.example/current-throughput-mibps`     | Last known provisioned throughput in MiB/s.                                                                                                   |
| `storage.atlas.example/current-capacity-bytes`       | Last known provisioned capacity in bytes.                                                                                                     |
| `storage.atlas.example/target-iops`                  | IOPS value currently being targeted.                                                                                                          |
| `storage.atlas.example/target-throughput-mibps`      | Throughput value currently being targeted.                                                                                                    |
| `storage.atlas.example/target-capacity-bytes`        | Capacity value currently being targeted.                                                                                                      |
| `storage.atlas.example/operation-status`             | Status of an in-flight operation: `pending`, `running`, `done`, or `failed`.                                                                  |
| `storage.atlas.example/last-capacity-modification`   | Timestamp of the last successful capacity change.                                                                                             |
| `storage.atlas.example/last-iops-modification`       | Timestamp of the last successful IOPS change.                                                                                                 |
| `storage.atlas.example/last-throughput-modification` | Timestamp of the last successful throughput change.                                                                                           |
| `storage.atlas.example/modifications-history`        | (AWS only) JSON array of recent ModifyVolume timestamps. Shows how many of the 4 operations in the rolling 24-hour window have been consumed. |
| `storage.atlas.example/last-error`                   | Last error message if an operation failed.                                                                                                    |

### PVC annotation examples

#### GCP / Hyperdisk Balanced: opt out of IOPS/throughput tuning

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-database-pvc
  annotations:
    storage.atlas.example/storage-threshold-percent: "85"
    storage.atlas.example/storage-growth-percent: "25"
    storage.atlas.example/iops-enabled: "false"
    storage.atlas.example/throughput-enabled: "false"
spec:
  storageClassName: hyperdisk-balanced
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
```

#### AWS / gp3: cap IOPS/throughput ceilings

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: my-database-pvc-aws
  annotations:
    storage.atlas.example/iops-max: "16000"
    storage.atlas.example/throughput-max-mibps: "500"
spec:
  storageClassName: gp3
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 100Gi
```

## Related resources

<Cards columns={2}>
  <Card title="Troubleshooting" href="https://docs.atlas.example/docs/storage-troubleshooting" icon="fa-wrench">
    Resolve common issues with capacity expansion and performance rightsizing.
  </Card>
  <Card title="Installation" href="https://docs.atlas.example/docs/storage-installation" icon="fa-download">
    Set up Storage Autoscaling on your GKE or EKS cluster.
  </Card>
</Cards>