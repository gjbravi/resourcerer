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

## Rules

- **Windows** are the periods an apply may start in. They are a union: inside any of them an apply may start. `to` is
  exclusive, `days` empty means every day, and a window whose `to` is earlier than its `from` crosses midnight and
  belongs to the day it opens on. `timeZone` is an IANA name; empty means UTC.
- **Freezes** are periods nothing is applied in, whatever the windows say. `to` is exclusive. An apply that is due
  waits for the freeze to end, and then still has to be inside a window.
- **Cooldown** is the minimum time between two changes to the same workload, counted from the last apply or revert.
- **`appliesTo`** says which applies the policy governs. Empty means `Auto`: a policy never blocks a person unless it
  says so.
- **Every policy that selects a workload has to allow the apply.** A namespaced policy can therefore be stricter than
  the cluster's, never looser: with the platform allowing `10:00–16:00` and the team `14:00–17:00`, applies happen
  between 14:00 and 16:00.
- **`mode: Audit`** lets the apply through and records what the policy would have blocked. Use it to try a policy on
  real traffic before enforcing it.
- A policy resourcerer cannot interpret, such as an unknown time zone, blocks: a policy that cannot be read must not
  open the door. If the policies cannot be read at all (a malformed duration, say), no apply of either kind goes
  through until the problem is fixed, the reason is recorded on every recommendation, and the applier retries every
  minute instead of stopping.

## What the UI and `kubectl` show

The applier records its decision on every Recommendation:

```yaml
status:
  applyPolicy:
    state: Blocked                  # Allowed | Blocked
    blockedBy: when.windows         # the policy field that blocks
    policy: ApplyPolicy/payments/checkout
    reason: outside the apply windows
    nextAllowed: "2026-11-03T14:00:00Z"
    wouldBlock: []                  # what Audit-mode policies would have blocked
    evaluatedAt: "2026-11-02T18:00:00Z"
```

`blockedBy` is a field path, so the UI points at the rule instead of asking you to parse a message. A blocked apply
is reconsidered when the policy changes and when `nextAllowed` arrives, so a window opening does not wait for the next
regeneration of the recommendation.

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
