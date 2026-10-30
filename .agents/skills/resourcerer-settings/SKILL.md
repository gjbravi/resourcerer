---
name: resourcerer-settings
description: Add or change a resourcerer setting, the per-workload options users set through chart defaults and resourcerer.io/<key> annotations (goals, learning period, schedule, auto-deploy, limits policies, minimums…). Use when a behaviour should be configurable per namespace or workload.
---

# Adding a setting

Settings are how users tune resourcerer per workload. Each one resolves per key: built-in default → chart values
(`defaults`, in the defaults ConfigMap) → namespace annotation → workload annotation (`resourcerer.io/<key>`).

## Steps

1. **Define it** in `internal/settings/settings.go`:
   - a constant for the key (dotted, lowercase: `cpu.requests.min`, `auto-deploy.threshold`);
   - an entry in `All` (the order is the display order) with `Default` and either `Options` (closed set) or
     `Validate` (open values: durations, percentages, quantities, ratios). An invalid annotation is ignored with a
     warning in the Recommendation status, so validation is what keeps bad values out.
   - a typed accessor on `Effective` if the value needs parsing (`Duration`, `Float`, `Quantity`, `SkipList`).
2. **Use it** where the behaviour lives: the engine reads it from `Input.Settings`; the controller or applier from
   the resolved `Effective` (`settings.Resolve`).
3. **UI:** the server lists every setting (`settingInfos`), so it shows up in the Settings tab and the cluster
   defaults automatically, with `Options` as choices. Add its label, group, help and suggested values to `SETTINGS`
   in `web/src/settings.ts`.
4. **Docs:** add a row to the settings table in `README.md` (annotation, default, values).
5. **Chart:** `defaults` in the values is a free map rendered into the ConfigMap, so no schema change is needed; set
   it in `hack/resourcerer-values.yaml` if kind or e2e should use a non-default value.

## Behaviour to keep in mind

- A change to the effective settings regenerates the recommendation (`sameSettings` in the recommender), so editing an
  annotation shows its effect within a reconcile.
- The Recommendation status records each value and its source (`cluster`, `namespace`, `workload`); the UI shows where
  a value came from.
- The server writes settings as annotations through `PUT .../settings`, refused in read-only mode.
- Policies by selector are planned (#30); a new setting should work as a plain key and value so a policy can set it.

## Tests

- `internal/settings`: the default, valid and invalid values, and resolution order.
- The behaviour the setting controls, in the package that uses it (engine tests with `input(annotations, …)`).
- e2e when it changes something observable in kind: annotate a demo workload and `wait_for` the effect, as the
  `memory.optimization-goal` check in `test/e2e/run.sh` does.
