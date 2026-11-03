---
title: Learning period and schedule
description: How much history a recommendation uses, and how often it is regenerated.
---

## Learning period

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/learning-period` | `P7D` | An ISO 8601 duration: `P<n>D`, `P<n>W`, `PT<n>H`, `PT<n>M`, or combined such as `P1DT12H` |

The learning period is the window of usage history a recommendation is computed from. It decides:

- **What the recommendation has seen.** A peak older than the learning period does not count. Cover the longest cycle
  that matters: a week covers weekday and weekend traffic; a monthly batch needs `P30D` or more.
- **When it is `Ready`.** A recommendation is `Ready` once the history covers 95% of the learning period. A new workload
  stays `Learning` until then.
- **What Prometheus must keep.** Retention (or your long-term store, such as Thanos or Mimir) must cover the longest
  learning period you set; otherwise recommendations never leave `Learning`.
- **How heavy the queries are.** Longer periods read more samples; resourcerer caps the number of points per query, so
  long periods use coarser steps rather than larger responses.

Shorter is more reactive: a workload that changed its footprint last week is right-sized sooner. Longer is more
conservative. Use a short period where usage changes with each release and a long one for steady, seasonal services:

```bash
kubectl annotate namespace reporting resourcerer.io/learning-period=P30D
```

## Schedule

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/schedule` | `@daily` | `@hourly`, `@daily`, `@weekly`, or `@every <duration>` with a Go duration of at least `1m` |

The schedule is how often a recommendation is regenerated. Each workload gets a **slot** within the period, derived
from its name, so a cluster's workloads are spread over the day instead of querying Prometheus at once. The next slot
is in `status.nextScheduleAt`.

A recommendation is also regenerated outside its slot when:

- its settings change (at any level);
- the workload's current requests change, after a deploy or an apply;
- the last run failed (it retries);
- someone clicks **Regenerate** on the workload page.

With auto-deploy, the schedule bounds how often a workload can change: at most once per regeneration that produces a
new hash. `@daily` or `@weekly` keeps rollouts rare; use apply policies to decide *when* in the day they happen.
