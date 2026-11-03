---
title: Using the UI
description: A tour of the Overview, Workloads, workload and Settings pages, and what to look at on each.
---

The UI is served by the `server` component on the `resourcerer` Service. It has a light and a dark theme, switched
from the header, and refreshes itself: the workload page every few seconds, lists and charts less often.

## Overview

The cluster at a glance:

- **Cost and net impact**: the monthly cost of all managed workloads' requests, and what applying every recommendation
  would save (or cost). The cost source is shown next to it: configured rates or calibrated to Karpenter.
- **Namespaces**: cost and savings per namespace, marked when the namespace has `resourcerer.io` annotations.
- **Top savings**: the workloads with the most to save, linking to the **Savings** view of the list.
- **Reliability risks**: workloads with open reliability signals, linking to the **Reliability** view.
- **Recommendation status**: how many workloads are in each phase.
- **Cluster timeline**: requests, usage and recommended requests of the whole cluster over time.

## Workloads

Every managed workload, with its namespace, phase, score, current and recommended cost, and reliability. Views across
the top filter the list:

| View | Shows |
|---|---|
| **All** | Every managed workload. |
| **Savings** | Workloads whose recommendation saves money. |
| **Reliability** | Workloads with an open reliability signal. |
| **Under-provisioned** | Workloads whose requests are below their usage, or that have no requests. |
| **To apply** | Pending recommendations that can be applied and are not on auto-deploy. |
| **Auto-deploy** | Workloads with auto-deploy enabled. |

Search by name, sort by any column, and select several workloads to **apply them in bulk**; "Select all ready to apply"
picks those whose recommendation can be applied now.

## A workload

The header shows the workload's phase, the effect of applying (`saves $42.10/mo`), the **verdict** and the **Apply**
button, an **auto-deploy** switch for the workload, and warnings such as an invalid annotation. When an apply policy
holds the apply, the button says why and until when. The tabs:

### Impact

Cost and requests before and after, and **How the cost is calculated**: the rates, the replica average, and with
`cost.allocation: node` where the workload ran (capacity types and instance types). The verdict's reasons are
detailed here. See [Verdict and evidence](../concepts/verdict-and-evidence.md).

### Recommendation

Per container: current and recommended requests and limits, each with its **basis** (`p95 + 10% over 7d`). Usage charts
show the busiest pod's usage percentiles with the current and recommended requests over time, opening on the learning
period so the chart matches the numbers; zoom from 1 hour to 1 month. The HPA target change, when there is one.

**Regenerate** recomputes the recommendation now instead of at its next slot.

### Reliability

Every reliability signal for the workload, checked live over a window you pick (15 minutes to the learning period):
what it measures, its value against its levels, a chart, and the PromQL behind it. Signals the recommendation fixes are
marked as such. See [Reliability signals](reliability-signals.md).

### Patches

The recommendation as a change you can commit instead of applying from the UI:

- a **Kustomize patch**;
- **Helm values**, in the common `resources:` layout;
- a **`kubectl`** command.

See [GitOps](gitops.md).

### Settings

Every effective setting with its value and source (`cluster`, `namespace`, `workload`). Change one for the workload or
its namespace from here; this writes the `resourcerer.io/<key>` annotation. Resetting removes it, so the value comes
from the level above again.

## Settings

The installation's configuration, read-only:

- **Metrics source**: whether Prometheus answers, its version and retention.
- **Cost rates**: configured and, with Karpenter, calibrated rates and the cluster factor.
- **Scope**: the selection mode and excluded namespaces.
- **Cluster defaults**: every setting's cluster-level value.
- **Reliability signals**: built-in and declared signals, with a test run of each declared query against Prometheus.

## Read-only mode

With `readOnly: true`, every action is disabled and the apply button reads **Read-only**: no apply, revert,
regenerate, auto-deploy switch or settings changes. The server refuses them too, whatever calls it. Use it for an
install that only reports.
