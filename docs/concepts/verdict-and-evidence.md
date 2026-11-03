---
title: Verdict and evidence
description: How the workload page decides whether to apply, from history coverage, a replay of usage and reliability signals.
---

Under the **Apply** button, the workload page answers the question the page exists for: *should I apply this?* The
answer is the **verdict**, with the reasons behind it listed underneath, each marked as reassuring or worth a look.

## Verdicts

| Verdict | When |
|---|---|
| **Safe to apply** | `Ready`, a pending change, and nothing worth a look: full history, usage stayed under the recommendation, no open signal. |
| **Apply to fix reliability** | The recommendation costs more, and it fixes an open signal: OOM kills, or CPU throttling from 25%. Applying is the fix. |
| **Review before applying** | Something is worth a look: preliminary history (`Learning`), memory above the recommendation at some point, CPU above it more than 10% of the time, or a reliability signal the recommendation does not fix. |
| **Already right-sized** | The current requests match the recommendation within the threshold. Nothing to apply. |
| **Cannot apply now** | `Warning`, such as a workload scaled to zero. |
| **Collecting usage history** | `Pending`. |
| **Not optimized** | `Unsupported`. |
| **Metrics unavailable** | `Error`. |

Next to the verdict, the effect in money: `saves $42.10/mo (38%)` or `costs $6.20/mo more (12%)`.

"Review" is not "don't": it means a reason is listed that a person should read. A recommendation that raises memory
after OOM kills, for instance, is "Apply to fix reliability", while one that cuts memory on a workload with a high error
rate is "Review before applying", since the error rate may have nothing to do with resources or everything.

## Evidence: replaying the history

At each regeneration, resourcerer replays the usage history it computed the recommendation from against the
recommended requests, and stores the result in `status.evidence`:

| Evidence | Meaning | In the verdict |
|---|---|---|
| History | How much history the recommendation saw, against the learning period. | "7d of history", or "2d of 7d history: preliminary" below 95% |
| Memory above | Share of samples in which the busiest pod used more memory than the recommended request (worst container). | Any memory above the request is worth a look |
| Memory peak | The busiest pod's highest memory, as a percent of the recommended request. | Reassuring when below 100% |
| CPU above | Share of samples in which the busiest pod used more CPU than the recommended request. | Worth a look above 10% |
| Used | p95 usage × average replicas, priced like the requests: what the workload actually needs. | Shown in the cost breakdown |

CPU above the request some of the time is expected: a `p95` request is, by definition, passed 5% of the time, and CPU
above the request only borrows idle CPU from the node. Memory above the request is different: under memory pressure
the node evicts the pods using most above their requests first.

## Reliability signals

Reliability signals flag workloads whose requests or limits may be too tight, or that are unhealthy for other reasons:
CPU throttling, memory headroom, OOM kills, restarts, crash loops, an HPA at its maximum, BestEffort pods, and any
signal you declare as PromQL, such as error rates or latency.

A signal is **fixed by the recommendation** when the recommendation itself addresses it: OOM kills (memory goes above
the limit that killed the container) and CPU throttling from 25% (the limit is raised). These count *for* applying;
any other open signal asks for a review. See [Reliability signals](../guides/reliability-signals.md).

## Where the verdict comes from

The verdict is computed in the UI from the `Recommendation`'s status (phase, pending, evidence, impact) and the open
signals. It is not stored, so it always reflects the latest status and the latest live checks.
