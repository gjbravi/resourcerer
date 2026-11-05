---
updatedAt: 2026-08-25T13:13:30.000Z
---

# Idle resources

Identify and eliminate cloud waste by tracking unused resources across your organization.

## Overview

The Idle resources report provides visibility into cloud resources, such as disks, that remain unattached, unused, or forgotten after their associated clusters are deleted. This organization-level report helps you identify cost optimization opportunities by surfacing resources that accumulate charges without providing value.

### Cloud provider support

| Resource Type | GCP     | AWS     | Azure   |
| ------------- | ------- | ------- | ------- |
| Idle Disks    | ✓       | Planned | Planned |
| Idle VMs      | Planned | Planned | Planned |

### What are idle resources?

Idle resources in the context of this report are considered cloud infrastructure components that persist after the Kubernetes clusters they were created for have been deleted. Common examples include:

* **Unattached disks**: Storage volumes no longer connected to any virtual machine
* **Idle virtual machines**: Compute instances with minimal to no CPU utilization or network activity
  These resources often result from default cloud provider behaviors. For example, when a GKE cluster is deleted, associated persistent volumes may remain based on their reclaim policy (`Retain` or `Delete`), quietly accumulating costs over time.

## Access the report

Navigate to **Idle resources** in the left sidebar under the **Cost monitoring** section.

<Image border={false} src="https://files.readme.io/e0ba37719118c1469b043a1942d2bd74819556fcb32c2f156fce7a52426b193f-image.png" />

### Prerequisites

The Idle Resources report requires that the Cloud Connect script be run for your organization. Cloud Connect provides the permissions necessary to sync resource data from your cloud providers.

If you don't see any idle resources in the report, verify that:

* The Cloud Connect script has been run successfully.
* Your cloud provider credentials have the necessary read permissions for disk resources.

## Current capabilities

### Idle disks (GCP)

The report currently identifies unattached disks in Google Cloud Platform environments.

**Detection criteria:** A disk is marked as idle when it has no users, meaning it is not currently attached to any virtual machine. The system syncs disk data from your GCP projects and checks the attachment status during each sync cycle. Once a disk appears in the report, it represents an ongoing cost with no active compute workload.

**Displayed information:**

* Name
* Project
* Last detached date
* Availability zone (AZ)
* Size
* Status (Unattached)
* Estimated monthly cost

#### Take action

Look for patterns in resource creation. Multiple idle disks in a single project may indicate a systemic issue with cleanup processes or reclaim policies.

To take action on a specific resource:

1. Examine the disk details in the table.
2. Locate and view the resource in your cloud provider console.
3. Verify the disk is no longer needed by checking with the application or infrastructure owner.
4. Delete the resource through your cloud provider console.

The report updates regularly as new data syncs from your cloud providers, in line with typical Cloud Connect sync times, so you can track the impact of your cleanup efforts over time.

## What's coming

The Idle resources report will expand over time to cover more cloud providers and resource types. Planned improvements include:

* **Cost summary widgets**: Estimated monthly waste totals and resource counts at the top of the report
* **AWS support**: Detection of idle disks in Amazon Web Services
* **Azure support**: Detection of idle disks in Microsoft Azure
* **Idle VMs**: Identification of underutilized virtual machines across all providers
* **Additional resource types**: IP addresses, snapshots, images, and load balancers

<Cards columns={2}>
  <Card title="Cluster insights" href="https://docs.atlas.example/docs/cluster-insights" icon="fa-chart-line">
    Dashboard overview of your cluster's key data and status.
  </Card>
  <Card title="Cluster score" href="https://docs.atlas.example/docs/cluster-score" icon="fa-gauge-high">
    Comprehensive rating of your cluster's optimization and efficiency.
  </Card>
</Cards>