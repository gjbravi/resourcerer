---
updatedAt: 2026-10-25T20:35:32.000Z
---

# How it works

Understand how Atlas Storage Autoscaling expands persistent volume capacity and rightsizes IOPS and throughput performance.

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

## Overview

Atlas Storage Autoscaling manages two dimensions of persistent volume optimization: capacity and performance (IOPS/throughput). You deploy it to your cluster using the `atlas-storage-autoscaler` Helm chart, and it manages PVCs according to PVPolicy Custom Resources that define which PVCs to manage and how.

> ⚠️ Warning
> Cloud providers enforce cooldowns between disk modifications. Plan iteration accordingly, because rapid testing will hit these limits.

## Capacity expansion

Storage Autoscaling monitors disk usage and expands persistent volumes when usage exceeds a threshold for a sustained period.

* When storage usage exceeds the configured threshold for a sustained period, the PV is expanded.
* Expansion size is calculated as a percentage of the current volume size, with a minimum step to ensure meaningful growth.
* A cooldown period prevents consecutive expansions on the same PV.

You configure the threshold, growth percentage, cooldown, and other settings in the `spec.config.storage` section of a PVPolicy. For the full field reference, see [Configuration](storage-configuration.md).

## IOPS and throughput rightsizing

Storage Autoscaling periodically evaluates whether provisioned IOPS and throughput match actual workload demand. When a sustained gap is detected between observed usage and provisioned capacity, it adjusts the provisioned values to better match your workload by adding a configurable buffer, so there's headroom above observed peak usage.

A cooldown period prevents rapid consecutive modifications, giving the cloud provider time to stabilize the new values.

Only certain disk types support independent IOPS and throughput tuning. You configure these settings in the `spec.config.performance` section of a PVPolicy. For the full field reference, see [Configuration](storage-configuration.md).

## Configuration model

Storage Autoscaling uses three layers of configuration, applied in order:

1. **Helm values** set global defaults for all PVCs.
2. **PVPolicy Custom Resources** override those defaults for specific namespaces, labels, StorageClasses, or individual PVCs.
3. **PVC annotations** override the matched policy on a per-PVC basis.

### Default PVPolicy

The Helm chart automatically creates a `default` PVPolicy in the `atlas-agent` namespace when `defaultPVPolicy.enabled: true` (the default). This ensures every PVC is managed from install, in dry-run mode. In dry-run mode, planned actions are logged but no changes are made, so you can review behavior before going live. To disable the automatic creation, set `defaultPVPolicy.enabled: false` in Helm values.

For the full default policy manifest, see [Example 1 in Configuration](storage-configuration.md#example-1-default-policy-automatically-created-by-helm).

### Policy resolution

When multiple policies match a PVC, the one with the highest `spec.priority` takes precedence. Assign distinct priorities to your policies to keep resolution predictable.

For the full PVPolicy field reference, selector syntax, annotation list, and examples, see [Configuration](storage-configuration.md).

## Next steps

<Cards columns={2}>
  <Card title="Installation" href="https://docs.atlas.example/docs/storage-installation" icon="fa-download">
    Set up Storage Autoscaling on your GKE or EKS cluster.
  </Card>
  <Card title="Configuration" href="https://docs.atlas.example/docs/storage-configuration" icon="fa-sliders">
    Configure PVPolicy Custom Resources, PVC annotations, and Helm values.
  </Card>
</Cards>