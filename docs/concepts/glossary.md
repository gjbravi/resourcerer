---
title: Glossary
description: The words the UI and these docs use with a specific meaning.
---

| Term | Meaning |
|---|---|
| **Apply policy** | A `ClusterApplyPolicy` or `ApplyPolicy` deciding when an apply may start: windows, freezes and a rate. Every policy that selects a workload must allow the apply. See [Apply policies](../guides/apply-policies.md). |
| **Auto-deploy** | The applier applying `Ready`, pending recommendations without a click, once per hash. See [Auto-deploy](../guides/auto-deploy.md). |
| **Basis** | The per-value explanation shown in the UI, such as `p95 + 10% over 7d`. |
| **Busiest pod** | Usage is taken per container from the pod using the most at each step, not the average: the pod that would be throttled or OOM-killed first. |
| **Cluster factor** | Calibrated rates over the configured rates: what capacity costs on this cluster relative to `cost.rates`. |
| **Coverage** | The share of a workload's requests that ran on nodes with a known price. |
| **Decision model** | A model asked, when a user clicks Evaluate, whether a recommendation is safe to apply: a second opinion beside the verdict, with a probability per answer. It never applies anything. See [Decision models](../guides/decision-models.md). |
| **Evidence** | The replay of usage against the recommendation (`status.evidence`): peak and time above it, p95 usage priced. See [Verdict and evidence](verdict-and-evidence.md). |
| **Freeze** | A one-off period in which nothing is applied, both kinds of apply included. |
| **Goal** | `Savings`, `Balanced` or `Reliability`: the percentile and margin applied to usage, per resource. |
| **Hash** | Identity of a set of recommended values; an apply carries the hash the user reviewed. |
| **Hold** | An apply a policy refuses now: it is deferred, not cancelled, and the decision carries the rule and the instant it ends. |
| **In progress** | An apply whose rollout has not finished. A rate limit counts those, and a workload is never applied on top of a rollout, ours or not. |
| **Learning period** | The history a recommendation is computed from (`learning-period`, default `P7D`). |
| **Pending** (phase) | Not enough samples for any recommendation yet. See [Phases](phases.md). |
| **Pending** (status field) | The recommendation differs from the current requests by more than `auto-deploy.threshold`: there is something to apply. |
| **Phase** | How far a recommendation can be trusted: `Pending`, `Learning`, `Ready`, `Warning`, `Unsupported`, `Error`. |
| **Placement** | Where a workload's requests ran over the learning period: capacity types and instance types, with what each cost. |
| **Score** | 0–100, how closely the current requests match the recommendation, weighted by cost. 100 is right-sized. |
| **Signal** | A reliability check, built-in (throttling, memory headroom, restarts…) or declared as a PromQL template. "Fixed by apply" when the recommendation itself addresses it. |
| **Verdict** | The workload page's answer on whether to apply, from phase, evidence and reliability signals. |
| **Window** | A recurring period, in a time zone, in which an apply may start. |
| **Workload factor** | The cluster factor for one workload's nodes, under `cost.allocation: node`. |
