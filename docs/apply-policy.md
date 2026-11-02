# Apply policies

An apply policy decides **when** an apply may start. It says nothing about how much to change: that is a setting
(`auto-deploy.threshold`, the optimization goals, the limits policy). Auto-deploy used to be a switch that applied a
`Ready` recommendation the moment it appeared; a policy is how a production namespace turns it on.

Two kinds, with the same rules:

| Kind | Scope | Selects |
|---|---|---|
| `ClusterApplyPolicy` | cluster | namespaces (`namespaceSelector`) and workloads (`workloadSelector`) |
| `ApplyPolicy` | one namespace | workloads (`workloadSelector`, matched against the workload's labels and its pod template's) |

```yaml
apiVersion: resourcerer.io/v1alpha1
kind: ClusterApplyPolicy
metadata:
  name: production
spec:
  namespaceSelector: { matchLabels: { env: production } }
  appliesTo: [Auto]                 # Auto | Manual
  mode: Enforce                     # Enforce | Audit
  when:
    windows:
      - days: [Mon, Tue, Wed, Thu, Fri]
        from: "10:00"
        to: "16:00"
        timeZone: Europe/Lisbon
    freezes:
      - from: "2026-11-20T00:00:00Z"
        to: "2026-12-02T00:00:00Z"
        reason: Sales event
  limits:
    cooldown: 24h
```

A team can add its own policy in its namespace, with rules aligned to its on-call hours:

```yaml
apiVersion: resourcerer.io/v1alpha1
kind: ApplyPolicy
metadata:
  name: checkout
  namespace: payments
spec:
  workloadSelector: { matchLabels: { app.kubernetes.io/part-of: checkout } }
  appliesTo: [Auto, Manual]
  mode: Enforce
  when:
    windows:
      - days: [Tue, Wed, Thu]
        from: "14:00"
        to: "17:00"
        timeZone: Europe/Lisbon
    freezes:
      - { from: "2026-11-10T00:00:00Z", to: "2026-11-12T00:00:00Z", reason: Checkout v5 launch }
  limits:
    cooldown: 72h
```

## `appliesTo`

Which applies a policy governs:

| `appliesTo` | Automatic applies (auto-deploy) | The Apply button and the API |
|---|---|---|
| empty, or `[Auto]` | governed | free |
| `[Manual]` | free | governed |
| `[Auto, Manual]` | governed | governed |

Both decisions are always computed, so `[Auto, Manual]` is the whole story: one policy, both kinds. A policy that lists
only `Auto` is not consulted for a manual apply, so it neither blocks the button nor shows up in that decision. Empty
means `Auto` on purpose: a policy written to hold back the automatic apply must not lock a person out of it by accident.

## Rules

- **Windows** are the periods an apply may start in. They are a union: inside any of them an apply may start. `to` is
  exclusive, `days` empty means every day, and a window whose `to` is earlier than its `from` crosses midnight and
  belongs to the day it opens on. `timeZone` is an IANA name; empty means UTC.
- **Freezes** are periods nothing is applied in, whatever the windows say. `to` is exclusive. An apply that is due
  waits for the freeze to end, and then still has to be inside a window.
- **Cooldown** is the minimum time between two changes to the same workload, counted from the last apply or revert.
- **`appliesTo`** says which applies the policy governs; see above.
- **Every policy that selects a workload has to allow the apply.** A namespaced policy can therefore be stricter than
  the cluster's, never looser: with the platform allowing `10:00–16:00` and the team `14:00–17:00`, applies happen
  between 14:00 and 16:00.
- **`mode: Audit`** lets the apply through and records what the policy would have blocked. Use it to try a policy on
  real traffic before enforcing it.
- A policy resourcerer cannot interpret, such as an unknown time zone, blocks and the decision says so
  (`blockedBy: policy` and the error as the reason). If the policies cannot even be read (a malformed object, a
  duration the API server accepted under an older CRD), no apply of either kind goes through until the problem is
  fixed: the reason is recorded on every recommendation and the applier retries every minute instead of stopping.

## How it works

The applier is the only component that changes workloads, so it is also the only one that asks the policies:

1. On every reconcile, and immediately when a policy is created, changed or deleted, it lists the
   `ClusterApplyPolicy`s whose `namespaceSelector` matches the workload's namespace and every `ApplyPolicy` in that
   namespace. The reads bypass the informer cache: a policy created a second ago has to be seen before the next apply,
   and a cached list would let that apply through.
2. It keeps the rules that select the workload: labels of the workload itself and of its pod template, since a
   manifest often labels only the pod.
3. It evaluates them twice, once for `Auto` and once for `Manual`; a rule takes part only in the kind it lists. The
   evaluation is a pure function over the rules and the clock, so the same input always gives the same decision.
4. If anything blocks, the apply does not happen. The applier records the decision, asks to be woken at `nextAllowed`
   (at most an hour away, so a changed clock or a new policy is noticed) and waits. A policy change wakes every
   recommendation at once, which is why a window opening does not wait for the next regeneration of the
   recommendation; the fan-out is cheap because deciding is a pure function over a handful of rules.
5. If the policies cannot be read at all, nothing is applied and the reason is recorded on every recommendation; the
   applier retries every minute. It never treats "could not read the rules" as "no rules".

The manual path gets the answer twice. The server refuses an apply or a revert that the manual decision holds, with
`409` and the reason, so the click fails at once; the applier checks again when the request reaches it, because the
status the server read may be a moment old.

## Fields

`ClusterApplyPolicy.spec` and `ApplyPolicy.spec` carry the same rules; `ClusterApplyPolicy` adds `namespaceSelector`.

| Field | Type | Default | Meaning |
|---|---|---|---|
| `namespaceSelector` | label selector | every namespace | Namespaces the policy covers, matched against the namespace's labels. |
| `workloadSelector` | label selector | every workload | Workloads the policy covers, matched against the workload's labels and its pod template's. |
| `priority` | integer | `0` | Which policy is named when more than one blocks: the highest, ties by name. It never makes a policy stronger or weaker. |
| `appliesTo` | `Auto`, `Manual` | `[Auto]` | Kinds of apply the policy governs. See above. |
| `mode` | `Enforce`, `Audit` | `Enforce` | `Audit` lets the apply through and records what it would have blocked. |
| `when.windows` | list | any time | The periods an apply may start in. A union: one is enough. |
| `when.windows[].days` | `Mon`…`Sun` | every day | Days the window opens on. An overnight window belongs to the day it opens. |
| `when.windows[].from`, `.to` | `HH:MM` | required | Local times; `to` is exclusive, earlier than `from` crosses midnight, equal never opens. |
| `when.windows[].timeZone` | IANA name | `UTC` | Time zone of the window's wall clock, so it follows the local daylight saving. |
| `when.freezes` | list | none | Periods nothing is applied in, windows or not. |
| `when.freezes[].from`, `.to` | RFC 3339 | required | Absolute instants; `to` is exclusive and has to be after `from`. |
| `when.freezes[].reason` | string | none | Shown in the decision and in the UI: `frozen: Sales event (until …)`. |
| `limits.cooldown` | duration | none | Minimum time between two changes to the same workload, counted from the last apply or revert of any hash. |

The applier writes the decision to `status.applyPolicy`:

| Field | Meaning |
|---|---|
| `auto`, `manual` | One decision per kind of apply, computed apart. |
| `state` | `Allowed` or `Blocked`. |
| `policy` | `Kind/namespace/name` of the policy that blocks, when one does. |
| `blockedBy` | The policy field that blocks: `when.freezes`, `when.windows`, `limits.cooldown`, or `policy` when the rule cannot be interpreted (an unknown time zone). |
| `reason` | The same in words, for people: `outside the apply windows`, `frozen: … (until …)`, `changed less than 24h0m0s ago, cooling down until …`. |
| `nextAllowed` | When the block ends, from the end of the last blocker: a freeze ending outside a window resolves to when that window opens. |
| `wouldBlock` | What `Audit`-mode policies would have blocked, most specific first, as `<policy>: <reason>`. |
| `evaluatedAt` | When the decision last changed, not when it was last computed: identical decisions are not written again. |

## What the UI and `kubectl` show

The applier records its decision on every Recommendation. The two kinds of apply are decided apart, because a policy
commonly holds auto-deploy back while leaving a person free, or the other way around:

```yaml
status:
  applyPolicy:
    auto:
      state: Blocked                # Allowed | Blocked
      policy: ApplyPolicy/payments/checkout
      blockedBy: when.windows       # when.freezes | when.windows | limits.cooldown | policy
      reason: outside the apply windows
      nextAllowed: "2026-11-03T14:00:00Z"
      wouldBlock: []                # what Audit-mode policies would have blocked
    manual:
      state: Allowed
    evaluatedAt: "2026-11-02T18:00:00Z"
```

`blockedBy` is a field path, so a client can name the rule instead of parsing a message. The API refuses an apply or a
revert that a policy holds with `409` and the reason (`ApplyPolicy/payments/checkout holds this apply: frozen: …`), which
is what the UI shows when someone clicks; the auto-deploy card says when a policy holds the automatic apply. A blocked
apply is reconsidered when the policy changes and when `nextAllowed` arrives, so a window opening does not wait for the
next regeneration of the recommendation.

```bash
kubectl get clusterapplypolicies
kubectl get applypolicies -A
kubectl -n payments get rec deployment-checkout -o jsonpath='{.status.applyPolicy}'
```

The kinds ship in the chart's `crds/`, which Helm installs on a fresh install but **not** on an upgrade: apply
`charts/resourcerer/crds` first, then restart the applier. Without them the applier runs and logs that applies are
not gated, instead of refusing to start.

## What it does not do

- **How much one apply may change** (`maxDecrease`, steps over several applies) is not here yet.
- **Verification after an apply** (revert when pods start getting OOM-killed) is not here yet.
- **Disruption budgets**: an apply that a PodDisruptionBudget will not let roll is not detected yet.
