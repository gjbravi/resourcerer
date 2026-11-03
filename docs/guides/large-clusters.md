---
title: Large clusters
description: Keep resourcerer and your Prometheus comfortable with thousands of workloads.
---

resourcerer is built to run on clusters with thousands of workloads without a large footprint or a heavy load on
Prometheus. These are the levers when it grows.

## Load on Prometheus

- **Spread schedules.** Each workload regenerates at its own slot within its schedule period, so a cluster on `@daily`
  sends its queries spread over the day, not at midnight. Avoid short schedules (`@every 5m`) on many workloads.
- **Learning period.** Longer periods read more data per query. resourcerer caps the number of points per query, so a
  `P30D` period uses coarser steps rather than larger responses, but the backend still scans more blocks.
- **Declared signals** run one instant query per signal and workload at each regeneration, with a 10s timeout. Keep the
  list to signals that change a decision.
- **Live checks** for the workload page (reliability windows, placement, charts) run when someone opens it, and are
  cached by the server; identical concurrent requests share one query.
- With **Thanos**, point `prometheus.url` at a Query Frontend to benefit from its splitting and caching.

## The controller's memory

The controller caches the workloads, their pods and HPAs, namespaces and Recommendations. Managed fields are stripped
from cached objects, and only the defaults ConfigMap's namespace is cached for ConfigMaps. Its memory grows with the
number of pods; raise `controller.resources.limits.memory` on large clusters and watch its usage after install.

## Narrow the scope

- `selection: label` to manage only labeled namespaces or workloads;
- `excludedNamespaces` for system and tooling namespaces;
- `resourcerer.io/enabled=false` on namespaces full of short-lived or generated workloads.

## The UI

Lists and workload details come from the cached `Recommendation`s, with no Prometheus query, so the workload list stays
fast with thousands of rows. Only the charts and live checks of one workload query Prometheus.

## Auto-deploy at scale

Turning auto-deploy on for many workloads at once can start many rollouts at once. Apply policies keep it in check:
`rate.maxConcurrent` (a number, or a percentage of the selected workloads), `rate.maxPerHour`, one rollout per
namespace with `perNamespace`, and `spread` so the workloads held by a closed window do not all start the second it
opens. See [Apply policies](apply-policies.md#rate-limits).
