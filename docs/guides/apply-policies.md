---
title: Apply policies
description: Decide when an apply may start, with windows, freezes, rate limits and cooldowns, as ClusterApplyPolicy and ApplyPolicy objects.
---

An apply policy decides **when** an apply may start. Without one, an apply that is due starts at once: the applier
patches the pod template and the pods restart. A policy lets a release run automatic applies in namespaces where
that has to be more careful:

- only during working hours, and never during a sale or an incident (windows and freezes);
- not more than one rollout at a time, in the cluster or in a namespace (rate limits);
- not twice in an hour for the same workload (cooldown);
- and, when several applies are held by a closed window, not all at the same second (spread).

How much one apply may change — `apply.max-decrease`, `apply.max-increase` — is a setting, not a policy field, so
it follows the settings cascade and the namespace annotations a team already uses. It is not implemented yet.

Two custom resources carry policies:

| | Scope | Written by |
|---|---|---|
| `ClusterApplyPolicy` | cluster; selects namespaces by label | the platform team |
| `ApplyPolicy` | one namespace | the team that owns it |

Every policy that selects a workload has to allow the apply, so a namespace can be **stricter than the platform
and never looser**. A namespaced policy cannot remove a cluster freeze, and a cluster window cannot reopen one a
namespace closed.

```yaml
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: business-hours
spec:
  namespaceSelector:                        # absent: every managed namespace
    matchLabels:
      resourcerer.io/environment: production
  workloadSelector:                         # absent: every workload in scope
    matchExpressions:
      - { key: tier, operator: In, values: [frontend, backend] }
  appliesTo: [Auto]                         # default Auto: a person is never held unless the policy says so
  mode: Enforce                             # or Audit, which records without holding
  windows:
    - days: [Mon, Tue, Wed, Thu, Fri]
      from: "09:00"
      to: "17:30"
      timeZone: Europe/Lisbon
  freezes:
    - from: 2026-11-27T00:00:00Z
      to: 2026-12-01T00:00:00Z
      reason: Black Friday
  rate:
    maxConcurrent: 25%                      # of the auto-deploy workloads it selects, rounded up
    maxPerHour: 4
    perNamespace: { maxConcurrent: 1 }
    cooldown: 24h
    spread: 15m
```

`workloadSelector` matches the workload's labels **and its pod template's**, since manifests often label only the
pod.

## Two kinds of apply

A policy names the kinds it governs with `appliesTo`. `Auto` is auto-deploy; `Manual` is a person clicking Apply.

| Rule | Automatic applies | Manual applies |
|---|---|---|
| `windows` | holds | holds |
| `freezes` | holds | holds |
| `rate.cooldown` | holds | holds |
| `rate.maxConcurrent` | holds | never holds, counts as in progress |
| `rate.maxPerHour` | holds | never holds, counts as in progress |
| `rate.perNamespace` | holds | never holds, counts as in progress |
| `rate.spread` | holds | never holds, counts as in progress |
| the workload's own rollout | holds | never holds |

A person is not a queue: someone who reviewed a recommendation and clicked Apply is not told to come back in an
hour. Their apply still counts against the limits, and a manual apply that starts while one is in progress waits
for that rollout, like any other.

**Reverts are never held.** A revert is how a bad apply is undone, and a freeze or the cooldown that follows the
apply it undoes is exactly when it has to stay possible.

### Applying outside a window

A person who owns the application is not the audience of a window: a window says "this is when the least traffic
runs", not "nobody may act". Which one you get is one field:

- **Advisory, the default.** Leave `Manual` out of `appliesTo`. The window holds auto-deploy and nothing else, so a
  click at 14:00 applies. The decision in the status still names the window, so the UI and `kubectl` can show that
  the apply is outside it.
- **Committed.** List `Manual`. The click is refused with `409` and the reason, and in this version there is no way
  past it in the product.

That second option decides for the person because resourcerer cannot yet tell who is clicking: authentication is
`anonymous`, so `history[].by` records what the client claimed, not who acted, and an override with no caller
behind it records nothing worth keeping. [#33](https://github.com/gjbravi/resourcerer/issues/33) brings the
identity, and with it the override: a caller allowed to bypass the advisory rules (`windows`, `rate.cooldown`)
applies with a reason the history keeps, while freezes, rate limits and a rollout in progress still hold.

So list `Manual` where the rule is a promise to somebody else — a freeze for a sales event, a code freeze — and
leave a window that only reflects traffic on `Auto`.

## Windows

A window is open while the local clock shows a time in `[from, to)` on a listed day. Windows are a union: inside
any of them an apply may start. `from` and `to` are `HH:MM` in `timeZone` (an IANA name; empty means UTC). A window
whose `to` is earlier than its `from` crosses midnight and closes the next day, so `22:00`–`02:00` on `Mon` is open
until 02:00 on Tuesday and does not reopen on Tuesday evening.

Daylight saving follows the clock, not `time.Date`:

- a **time the clock skips** is the instant the clock resumes. A window `02:00`–`05:00` in `Europe/Lisbon` opens at
  03:00 on the spring-forward day, not an hour early;
- a **time the clock repeats** is its first occurrence, so a window closes the first time it reaches `to` and never
  reopens for the repeated hour.

Only the US shifts by an hour: a 30-minute shift (`Australia/Lord_Howe`) is handled the same way.

## Freezes

A freeze holds every apply of both kinds, in or out of a window, until `to`. It takes effect within seconds: the
applier reads the policies uncached right before it decides, and a policy change wakes every workload it governs.
Give the reason: it is what a person sees on the Apply button and in the event.

## Rate limits

Each policy counts the workloads **it selects**, wherever they run, and an apply starts only when every policy
that selects it has a slot.

| Field | Counts | Notes |
|---|---|---|
| `maxConcurrent` | applies of ours in progress: the rollout has not finished | a whole number or a percentage of the workloads the policy selects with auto-deploy enabled, rounded up; `0` blocks every automatic apply |
| `maxPerHour` | applies of ours in the last hour | `0` blocks every automatic apply |
| `perNamespace.maxConcurrent` | rollouts in progress in the workload's namespace | counts rollouts resourcerer did not start, such as a deploy from CI |
| `cooldown` | time since the last apply or revert of the same workload | holds manual applies too |
| `spread` | — | delays each held apply to a stable offset in `[0, spread)` after the window opens |

`maxConcurrent: 1` with `perNamespace: { maxConcurrent: 1 }` is what stops two rollouts in one namespace: the
first is the policy's own slot, the second is the namespace's.

`spread` is capped at the window's length, and its offset is computed from the workload's identity, so the same
workload always starts at the same moment and different workloads spread out. It starts counting when the window
opens, or when a freeze that ran into the window ends.

An apply is **in progress until its rollout finishes**. The applier tracks the applies it started, and a workload
whose rollout is in progress — anyone's — is never applied on top of.

## Audit

`mode: Audit` records what the policy would have blocked in `status.applyPolicy.wouldBlock` and lets the apply
through. Use it to write a window or a rate limit and watch it for a week before enforcing it. An Audit policy
does not count in the instant an apply becomes possible, since it blocks nothing.

## The decision

The applier decides for every managed workload and records it in the `Recommendation`:

```yaml
status:
  applyPolicy:
    auto:
      state: Blocked
      policy: ClusterApplyPolicy/business-hours
      blockedBy: windows
      reason: outside the apply windows (Mon,Tue,Wed,Thu,Fri 09:00-17:30 Europe/Lisbon)
      nextAllowed: "2026-11-04T09:00:00Z"
    manual:
      state: Allowed
    evaluatedAt: "2026-11-03T21:14:02Z"
```

`blockedBy` is the rule as a spec field path, so the UI can point at it instead of parsing a message: `windows`,
`freezes`, `rate.cooldown`, `rate.maxConcurrent`, `rate.maxPerHour`, `rate.perNamespace`, `rate.spread`, `rollout`
(its own rollout has not finished), or `policy` (a policy could not be read).

`nextAllowed` is the first instant every policy allows an apply. It is absent while waiting for a slot, which
depends on other rollouts finishing; the applier looks again within 30 seconds then, and at most an hour later in
any case. `evaluatedAt` is when the decision last **changed**, not when it was last computed.

The server refuses a manual apply the `manual` decision holds with `409 Conflict` and the reason, so the click gets
its answer at once. The applier decides again on the live policies when the request arrives — the recorded decision
may be a minute old — and rejects the request with the same reason if it is still held.

## When the rules cannot be read

A policy resourcerer cannot interpret — an unknown time zone, a malformed selector, a percentage that is not one —
holds the applies it governs: a policy that cannot be read never opens the door. The decision says
`blockedBy: policy` and carries the error, the policy's status carries `Valid: False` with the same message, and
the applier tries again every minute.

A policy that **cannot be read from the API server at all** (RBAC, an API outage) holds the same way. The one
exception is the CRDs being missing: an installation upgraded without applying `charts/resourcerer/crds` runs
ungated and says so in the log, rather than refusing every apply.

## Metrics

| Metric | Labels | Use |
|---|---|---|
| `resourcerer_apply_total` | `result` = `Succeeded`, `Failed`, `Rejected` | how often applies happen and fail |
| `resourcerer_apply_deferred_total` | `reason` = `window`, `freeze`, `cooldown`, `concurrency`, `namespace`, `rate`, `spread`, `rollout`, `policy` | how much auto-deploy is being held, and why |
| `resourcerer_apply_reverted_total` | `reason` = `manual` | how often an apply is undone; every revert in this version is one a person asked for |
| `resourcerer_applypolicy_slots_available` | `policy` | how far a `maxConcurrent` policy is from its limit, like Karpenter's allowed disruptions |

The applier serves them on `:8081`, as the controller does its own; the chart exposes both Services
(`<release>-metrics` and `<release>-applier-metrics`) and scrapes them when `metrics.scrapeAnnotations` is on.

Watch `resourcerer_apply_deferred_total` for the first hours of a policy in `Enforce`: a window that was meant to
allow a nightly batch and a window that never opens look the same without it.

## Shipping policies with the release

`applyPolicies.cluster` in the chart values renders `ClusterApplyPolicy` objects with the release:

```yaml
applyPolicies:
  cluster:
    - name: business-hours
      spec:
        appliesTo: [Auto]
        windows:
          - { days: [Mon, Tue, Wed, Thu, Fri], from: "09:00", to: "17:30", timeZone: Europe/Lisbon }
```

A policy is a Kubernetes object, so a team can also write its own `ApplyPolicy` in its namespace, and both apply.

## Not here yet

Later phases of the design, in the open issue, add:
- a **disruption budget** against the workload's PodDisruptionBudget and rollout strategy, and a verdict the apply
  must be confident enough for (`disruption`, `require`);
- **freezes declared by a query**: a freeze while a metric is above a threshold (`freezeWhen`);
- **verification with automatic revert**: watch the pods after an apply and put the previous values back when they
  do not settle (`verify`);
- **change limits as settings** (`apply.max-decrease`, `apply.max-increase`) and steps over several runs.

## See also

- [Architecture](../architecture.md): where the applier decides and writes.
- [Reliability signals](reliability-signals.md): the signals behind "Review before applying".
- The `Recommendation` CRD in `api/v1alpha1/recommendation_types.go` for the status field.
