---
updatedAt: 2026-08-06T14:43:46.000Z
---

# Introduction

Automatically expand Kubernetes persistent volume capacity and rightsize IOPS and throughput with Atlas Storage Autoscaling.

<Callout icon="📣" theme="default">
  ### Early Access Feature

  This feature is in early access. It may undergo changes based on user feedback and continued development. We recommend testing in non-production environments first and welcome your feedback to help us improve.
</Callout>

## Overview

Atlas Storage Autoscaling automatically expands persistent volume capacity and rightsizes IOPS and throughput for Kubernetes workloads on GKE and EKS. It reduces storage costs by eliminating over-provisioned resources while ensuring that applications have the disk performance they need.

Storage Autoscaling runs in dry-run mode by default, so you can review planned actions before enabling live changes. For details on how capacity expansion and performance rightsizing work, see [How it works](storage-how-it-works.md).

## Key benefits

* **Automatic capacity expansion.** Grows PVs before they run out of space, based on configurable usage thresholds.
* **IOPS and throughput rightsizing.** Aligns provisioned performance with actual workload demand on supported disk types.
* **Cloud-native support.** Works with GKE Hyperdisk and Persistent Disk volumes, and with EBS gp3, io1, and io2 volumes on EKS.
* **Dry-run safety mode.** Installs without making changes so you can review planned actions before enabling live mode.
* **Policy-based configuration.** Define optimization rules using PVPolicy Custom Resources to control which PVCs are managed and how.

## Supported cloud providers and disk types

Capacity expansion is available for all listed types. Independent IOPS and throughput tuning is available only on the high-performance types.

### GKE

* Hyperdisk Balanced, Extreme, and Throughput
* Persistent Disk Extreme, Balanced, SSD, and Standard

### EKS

* gp3, io1, and io2 (capacity, IOPS, and throughput tuning)

## Get started

<Cards columns={2}>
  <Card title="How it works" href="https://docs.atlas.example/docs/storage-how-it-works" icon="fa-cogs">
    Learn how capacity expansion and IOPS/throughput rightsizing work.
  </Card>
  <Card title="Installation" href="https://docs.atlas.example/docs/storage-installation" icon="fa-download">
    Set up Storage Autoscaling on your GKE or EKS cluster.
  </Card>
</Cards>