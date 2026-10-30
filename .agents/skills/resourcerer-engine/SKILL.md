---
name: resourcerer-engine
description: Change how resourcerer computes recommendations (internal/engine): goals, percentiles and margins, limits policies, OOM and throttling rules, HPA handling, phases, the pending threshold or the hash. Use for any change to the recommended values or to when a recommendation is ready or applied.
---

# Changing recommendation rules

`internal/engine` decides the values that get applied to production workloads. It is pure (no I/O), so every rule
is tested with synthetic series; keep it that way and keep the controller (`internal/recommender`) as the only place
that gathers data.

## Where each rule lives

| Rule | Where |
|---|---|
| Percentile and margin per goal | `Goals` in `engine.go` |
| Phase (`Pending`, `Learning`, `Ready`, `Warning`, `Unsupported`) and minimum samples | `Recommend`, `minSamples` |
| Request: quantile × (1 + margin), OOM raise, minimum, rounding | `Recommend`, `round` |
| Limits per policy and ratio, CPU limit raised when throttled ≥ 25% | `Recommend`, `raiseLimit`, `ThrottledShare` |
| HPA: unsupported on CPU and memory, target preserved on CPU (50–90%) | `Recommend` |
| Explanation shown in the UI ("p95 + 10% over 7d · raised after OOM kill") | `ContainerOutput.Basis` |
| Identity of a recommendation | `Hash` |
| Whether a recommendation differs enough to apply | `Pending` (threshold and 10m CPU / 16Mi memory minimum) |
| Totals, monthly cost, score | `Summarize` |
| Replay of usage against the recommendation | `Replay` (`evidence.go`) |

## Things that ripple

- **The hash changes when any recommended value changes.** A rule change produces new hashes for many workloads at
  their next regeneration, and auto-deploy applies each new hash once. A change that moves values broadly is a
  rollout to every auto-deployed workload: call it out in the PR and in the release notes, and prefer changes that
  only move values when the evidence says so.
- **The `Basis` is the explanation users see.** Update it whenever a rule changes a value, so the UI says why.
- **Goals are documented in three places**: `Goals` here, the table in `docs/architecture.md`, and the help text in
  `web/src/settings.ts`. Change all three together.
- **Phases drive auto-deploy** (`Ready` only) and manual apply (`Ready` or `Learning`). A rule that keeps workloads
  in `Learning` longer delays savings; one that makes them `Ready` sooner applies on less history.
- **Usage is the busiest pod**, not the average: rules that compare with requests compare per pod.
- **Never lower memory below what killed the container.** OOM kills cap observed memory at the old limit, so the
  history underestimates; keep the raise above the limit.

## Tests

Table tests in `engine_test.go` with the helpers `series`, `flat` and `input(annotations, container)`:

```go
c := ContainerInput{Name: "app", Requests: Values{CPU: 1}, CPU: flat(0.2), Memory: flat(100 * Mi)}
out := Recommend(input(map[string]string{"resourcerer.io/cpu.optimization-goal": "Savings"}, c))
```

Cover the new rule, its boundaries (exactly at a threshold, just below, just above), and that unrelated settings keep
their behaviour. Then run `make test` and `make kind-deploy e2e`: the e2e applies real recommendations in kind.

## Before merging

Look at what the change does to real recommendations: regenerate in kind and compare `status.containers` and
`status.impact` before and after for the demo workloads, and describe the expected effect (direction and rough size)
in the PR.
