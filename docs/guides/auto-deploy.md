---
title: Roll out auto-deploy
description: Let resourcerer apply recommendations without a click, and turn it on gradually, within apply policies.
---

Auto-deploy is what turns recommendations into savings across a cluster without someone clicking through every
workload. It is off by default, everywhere.

## What auto-deploy applies

With `auto-deploy: Enabled`, the applier applies a recommendation without a click when all of these hold:

- the phase is **`Ready`**: the history covers the learning period (`Learning` recommendations are never auto-applied);
- it is **pending**: a request moves by more than `auto-deploy.threshold` (default `10%`) and by at least 10m of CPU or
  16Mi of memory;
- its **hash was never applied or reverted** before: each set of values is auto-applied at most once, and a revert is
  respected;
- **no rollout is in progress** on the workload, ours or anyone's;
- **every apply policy** that selects the workload allows an automatic apply now.

It is recorded as an apply by `auto-deploy` in the history and the events, and can be reverted like any other.

| Annotation | Default | Values |
|---|---|---|
| `resourcerer.io/auto-deploy` | `Disabled` | `Enabled`, `Disabled` |
| `resourcerer.io/auto-deploy.threshold` | `10%` | A percentage, `0%` to `100%` |
| `resourcerer.io/apply-method` | `PatchImmediateRollout` | `PatchImmediateRollout`: patch the pod template, which starts a rollout at once |

## A gradual rollout

1. **Watch.** Install with auto-deploy off. Let recommendations reach `Ready` and read a few verdicts.
2. **Apply by hand.** Apply a handful of workloads in a non-critical namespace from the UI; watch them for a few days
   (latency, restarts, throttling: the Reliability tab shows them).
3. **Add a policy before anything is automatic.** A window in working hours, one rollout at a time per namespace, a
   cooldown. In `Audit` mode first, to see what it would hold:

    ```yaml
    applyPolicies:
      cluster:
        - name: business-hours
          spec:
            mode: Audit
            appliesTo: [Auto]
            windows:
              - { days: [Mon, Tue, Wed, Thu], from: "09:00", to: "16:00", timeZone: Europe/Lisbon }
            rate:
              maxConcurrent: 2
              perNamespace: { maxConcurrent: 1 }
              cooldown: 24h
              spread: 30m
    ```

4. **Enable a first namespace.** Development or staging:

    ```bash
    kubectl annotate namespace staging resourcerer.io/auto-deploy=Enabled
    ```

5. **Enforce the policy** (`mode: Enforce`, or remove `mode`) once its `wouldBlock` records look right.
6. **Extend namespace by namespace**, leaving out the workloads that should stay manual:

    ```bash
    kubectl annotate namespace payments resourcerer.io/auto-deploy=Enabled
    kubectl -n payments annotate deployment ledger resourcerer.io/auto-deploy=Disabled
    ```

Or the other way around once you trust it: `auto-deploy: Enabled` in the chart's `defaults`, and `Disabled` on the
namespaces that opt out.

## How often a workload changes

At most once per regeneration that produces a new hash, so the [schedule](../configuration/learning-period-and-schedule.md#schedule)
bounds it: `@daily` means at most one automatic apply a day per workload. A larger `auto-deploy.threshold` makes
changes rarer and larger; `rate.cooldown` in an apply policy sets a hard minimum between two applies of the same
workload.

## Watch it

- The **Auto-deploy** view of the workload list shows the workloads with it on.
- `resourcerer_apply_total{result}` counts applies and failures; `resourcerer_apply_deferred_total{reason}` counts the
  ones a policy held, and why. See [Metrics and events](../reference/metrics-and-events.md).
- The `ApplyDeferred` event on a workload says which policy held it.

## Turn it off

Remove the annotation or set it to `Disabled`, at the level where you enabled it. Nothing is reverted: the values
already applied stay. To stop every apply at once without touching annotations, add a freeze. It holds manual applies
too; reverts stay possible.

```yaml
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: stop
spec:
  freezes:
    - { from: "2026-01-01T00:00:00Z", to: "2099-01-01T00:00:00Z", reason: auto-deploy paused }
```
