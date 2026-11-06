# AGENTS.md

Guidance for coding agents (and humans) working on resourcerer. Read this first; the skills in `.agents/skills/`
cover the recurring tasks step by step, and [docs/architecture.md](docs/architecture.md) explains how the
components work under the hood, with diagrams.

## What resourcerer is

A self-hosted Kubernetes rightsizing tool. It reads usage history from an existing Prometheus, writes one
`Recommendation` per workload, shows it in a web UI, and applies it on request or automatically. It is open source
and in beta: breaking changes are fine when the changelog says so.

The product's focus is **helping someone apply a recommendation with confidence**: how much it saves, what the risk
is, and why the numbers can be trusted. Prefer changes that serve that decision over more data on screen.

## Layout

| Path | What lives there |
|---|---|
| `cmd/resourcerer` | One binary; `controller`, `applier`, `server` or `all` as the first argument |
| `api/v1alpha1` | The `Recommendation` CRD types; deepcopy and the CRD in the chart are generated |
| `internal/engine` | Recommendation rules: **pure**, no I/O, table-driven tests with synthetic series |
| `internal/recommender` | Controller: gathers history, runs the engine, writes `status` on each workload's schedule |
| `internal/applier` | The only component that writes workloads: apply, revert, auto-deploy, history |
| `internal/server` | HTTP API (`/api/v1/...`) and the embedded UI; live, cached queries for the workload page |
| `internal/prom` | PromQL builders and the Prometheus client |
| `internal/cluster` | Cluster config: defaults ConfigMap, cost pricing (`Pricer`), node allocation and placement |
| `internal/policy` | When an apply may start, from the apply policies that select the workload: **pure**, table-driven tests |
| `internal/settings` | Settings, their defaults, validation and resolution (values → namespace → workload annotations) |
| `internal/signals` | Reliability signals: built-in toggles and PromQL templates declared in the values |
| `internal/decision` | The decision model's state from a Recommendation's status and usage, and the check of its answer: **pure**, plus one HTTP call (`Ask`) |
| `web/` | React 19 + TypeScript + Vite UI, React Query, Recharts 3; `web/src/api.ts` mirrors `internal/server/types.go` |
| `charts/resourcerer` | Helm chart; `values.schema.json` rejects unknown keys |
| `charts/resourcerer-crds` | The CRDs as a chart of their own, a separate release installed before `resourcerer`; `files/crds.yaml` is generated |
| `hack/` | kind environment: cluster, Prometheus, Grafana, demo workloads, a fake Karpenter exporter, a fake decider for e2e; the Grafana dashboard's generator |
| `test/e2e/run.sh` | End-to-end checks against kind |
| `docs/` | The user docs, by section (`getting-started/`, `concepts/`, `configuration/`, `guides/`, `reference/`), and `architecture.md` (components, data flow, diagrams) |
| `mkdocs.yml`, `website/` | The documentation site: Zensical builds `docs/` (config in `mkdocs.yml`, pinned version in `website/requirements.txt`), published to GitHub Pages by `.github/workflows/docs.yaml` |

## Commands

```bash
make lint            # golangci-lint, actionlint, eslint, prettier, tsc, helm lint, lint-docs
make lint-docs       # docs and skills match the repository; MERMAID=1 also renders the diagrams
make test            # Go and UI unit tests
make generate        # deepcopy + the CRDs into charts/resourcerer-crds/files/crds.yaml; run after changing api/
make dashboard       # the Grafana dashboard into the chart, from hack/grafana-dashboard.py
make verify-generate # fails when generated files are stale (compares with git: commit first)
make kind-up         # kind cluster with Prometheus, Grafana, metrics-server, VPA, demo workloads, fake Karpenter
make kind-deploy     # build the image, load it into kind, install the chart with hack/resourcerer-values.yaml
make e2e             # end-to-end checks against kind
make web-dev         # UI dev server proxied to a running server on :8080
make docs-dev        # documentation site with live reload on :8000
make docs-build      # build the documentation site strictly; fails on broken links
```

Behind a TLS-intercepting proxy, pass `EXTRA_CA_CERT=/path/to/ca.pem` to `make kind-up`.

The UI of the kind install: `kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80`.
A redeploy replaces the pods, so restart the port-forward afterwards.

## Definition of done

1. `make lint test` passes; after `api/` changes, `make generate` and commit the generated files.
2. Controller, server, chart or query changes: `make kind-deploy e2e` passes, with an e2e check for the new
   behaviour when it can be observed in kind (`test/e2e/run.sh`, using `pass`/`fail`/`wait_for`).
3. UI changes: checked in the browser against kind, light and dark themes.
4. Docs (`docs/`, see the `resourcerer-docs` skill; `make docs-build` passes), chart values (`values.yaml` comments,
   schema, chart README table) and `hack/resourcerer-values.yaml` updated when behaviour or configuration changed;
   `docs/architecture.md`, this file and the skills updated when a component, flow or convention they describe
   changed (`make lint-docs` catches stale paths and targets, not stale explanations).

## Conventions

### Go

- Keep controllers and handlers thin; put decisions in pure functions (`engine`, `cluster.combine`, `cluster.place`,
  `signals.Parse`) and test those with tables.
- Comments state intent, constraints or trade-offs the code cannot show. No comments that narrate the next line,
  reference tickets or justify a change. Match the density of the file you edit.
- Wrap errors with context (`fmt.Errorf("query Karpenter prices: %w", err)`).
- A failing optional feature (a reliability signal, node allocation, a decision model) degrades to the previous
  behaviour and records why; it never blocks a recommendation.

### PromQL and Prometheus

See the `resourcerer-promql` skill. In short: scope every query to the workload (`prom.Owner` joins pods through
ReplicaSets), deduplicate series that can overlap (`topk by (...) (1, ...)`), read Karpenter metrics with
`last_over_time(...[5m])` because they can arrive a minute late, use `RangeComplete` where a partial Thanos response
must not be stored as valid, cap windows and points so large clusters stay cheap, and tag the context with
`prom.For(ctx, purpose)` so the query is counted under what it is for (the server tags each route's).

### UI

See the `resourcerer-ui` skill. In short: logic in pure, tested modules (`verdict.ts`, `decision.ts`, `apply.ts`,
`placement.ts`, `reliability.ts`, `format.ts`), components reuse the `rz-` classes and theme variables in
`styles.css`, one color per chart series, shimmer skeletons for loading (no loading texts, no mascots), and no new
dependencies without a reason.

### Chart

Every value is documented with a `# --` comment in `values.yaml`, validated in `values.schema.json`
(`additionalProperties: false`), listed in the chart README when user-facing, and rendered into the defaults
ConfigMap when the binary reads it. See the `resourcerer-chart-values` skill.

### Commits, PRs and releases

- Conventional Commits for commits and PR titles (`feat(ui): ...`, `fix(applier): ...`); the PR title check
  enforces it and release notes come from it. Subjects are short, imperative, lowercase.
- No AI attribution trailers in commits or PRs.
- PRs are squash-merged. Fill the template: what and why, how it was tested (commands and results).
- Releases are cut by release-please. See the `resourcerer-release` skill before merging a release PR.

### Open source hygiene

The repository is public. Code, comments, tests, fixtures, commit messages, PRs, issues and release notes must not
mention a specific company's environment: cluster names, internal hostnames, account IDs, IPs, internal service or
namespace names, or numbers measured on them. Use generic wording ("a cluster running mostly on spot", "a real
cluster") and fictional examples (`payments/checkout`, `demo/api`). Grep the diff and the PR body before pushing.

## Design decisions to keep

- **Settings** resolve chart values → namespace annotations → workload annotations (`resourcerer.io/<key>`); the
  Recommendation status records each value and its source.
- **Costs** are requests × average replicas × hourly rates × 730. With `cost.source: karpenter` the rates are
  calibrated to Karpenter's price estimates; with `cost.allocation: node` each workload is priced at the nodes it ran
  on (time- and request-weighted over the learning period), falling back to the cluster average.
- **The workload page shows a verdict under the apply button** ("Safe to apply", "Review before applying", "Apply to
  fix reliability", …) built from history coverage, a replay of usage against the recommendation (`status.evidence`) and reliability
  signals. Signals the recommendation itself fixes (OOM kills, CPU throttling from 25%) count for applying.
- **Apply policies** (`ClusterApplyPolicy`, `ApplyPolicy`) decide **when** an apply may start: windows, freezes and
  a rate. Every policy that selects the workload must allow it, so a namespace can be stricter but never looser.
  The applier records the decision in `status.applyPolicy`, with the rule that blocked and when it stops; the server
  answers a held click with 409, a revert is never held, and a policy that cannot be read holds.
- **Decision models** give a second opinion on the verdict, never a decision: the server asks the `DecisionModel`
  named `default` only when a user clicks Evaluate, about a recommendation with something to apply, and caches the
  answer per recommendation hash. The controller never calls a model and nothing is stored in the status; the rules
  verdict stays beside it, unchanged. The state is built by `internal/decision` from the status and the usage series,
  never names the workload, its namespace or people, and tags the signals the recommendation fixes. See
  `docs/guides/decision-models.md`.
- **Reliability signals** are built-in or declared as PromQL templates (`docs/guides/reliability-signals.md`). A
  signal with no data does not apply to a workload. Containers in `containers.skip` are left out of every signal that
  names a container, in the status and the live checks: applying never changes them.
- **Live data** for the workload page (reliability windows, placement) is computed by the server on request and
  cached; the controller's `status` stays small and drives lists and auto-deploy.
- **The CRDs ship in their own chart** (`charts/resourcerer-crds`), a release installed before `resourcerer` and
  upgraded with it, never a chart dependency: the main chart renders `ClusterApplyPolicy` objects from
  `applyPolicies.cluster`, which cannot exist in the release that creates their CRD. They are templates, so
  `helm upgrade` maintains them, and every CRD carries `helm.sh/resource-policy: keep`, so uninstalling keeps them.

## Roadmap

Design issues describe the next steps; read the relevant one before working in that area:
reliability signals as custom resources (#28), `ApplyPolicy` (#29: windows, freezes and rate limits are in; the
disruption budget, `freezeWhen`, verification with revert and change limits as settings are next), `OptimizationPolicy`
(#30), `CostProfile` (#27), decision models with Jev/decider (#26: a second opinion on request is in; an auto-deploy
gate requiring the rules and the model, and calibration on past applies, are next), shimmer skeletons (#31).

## Glossary

Words the code and the UI use with a specific meaning:

| Term | Meaning |
|---|---|
| **Learning period** | The history a recommendation is computed from (`learning-period`, default `P7D`). |
| **Busiest pod** | Usage is taken per container from the pod using the most at each step, not the average: the pod that would be throttled or OOM-killed first. |
| **Goal** | `Savings`, `Balanced` or `Reliability`: the percentile and margin applied to usage, per resource. |
| **Phase** | How far a recommendation can be trusted: `Pending` (not enough samples), `Learning` (preliminary), `Ready`, `Warning` (cannot apply now), `Unsupported`, `Error`. |
| **Pending** (status field) | The recommendation differs from the current requests by more than `auto-deploy.threshold`. Not the same as the `Pending` phase. |
| **Hash** | Identity of a set of recommended values; an apply carries the hash the user reviewed. |
| **Basis** | The per-value explanation shown in the UI ("p95 + 10% over 7d"). |
| **Evidence** | The replay of usage against the recommendation (`status.evidence`): peak and time above it, p95 usage priced. |
| **Verdict** | The Impact tab's answer on whether to apply, from phase, evidence and reliability signals. |
| **Decision model** | A model the server asks, when a user clicks Evaluate, whether a recommendation is safe to apply: apply, review or wait, with a probability each. A second opinion beside the verdict; it never applies anything. |
| **Signal** | A reliability check, built-in (throttling, memory headroom, restarts…) or declared as a PromQL template. "Fixed by apply" when the recommendation itself addresses it. |
| **Cluster factor** | Calibrated rates over the configured rates: what capacity costs on this cluster relative to `cost.rates`. |
| **Workload factor** | The same for one workload's nodes, under `cost.allocation: node`. |
| **Coverage** | The share of a workload's requests that ran on nodes with a known price. |
| **Placement** | Where a workload's requests ran over the learning period: capacity types and instance types, with what each cost. |
| **Auto-deploy** | The applier applying `Ready`, pending recommendations without a click, once per hash; `auto-deploy.containers` limits it to some containers (`status.autoDeployScope`). |
| **Apply policy** | A `ClusterApplyPolicy` or `ApplyPolicy` deciding when an apply may start: windows, freezes and a rate. Every policy that selects a workload must allow the apply. |
| **Window / Freeze** | A recurring period in a time zone an apply may start in / a one-off period nothing is applied in, both kinds of apply included. |
| **Hold** | The applier's name for an apply a policy refuses now: it is deferred, not cancelled, and the decision carries the rule and the instant it ends. |
| **Overwritten / drifted** | An applied recommendation whose resources another field manager changed afterwards, such as a GitOps sync (`RecommendationOverwritten`); the workload stays drifted until it is applied again or reverted. |
| **In progress** | An apply of ours whose rollout has not finished. A rate limit counts those, and a workload is never applied on top of a rollout, ours or not. |

## Skills

| Skill | Use it to |
|---|---|
| `resourcerer-dev-loop` | Run the kind environment, deploy a change, run e2e, troubleshoot kind |
| `resourcerer-promql` | Write or change a Prometheus query |
| `resourcerer-ui` | Change the web UI |
| `resourcerer-engine` | Change how recommendations are computed |
| `resourcerer-settings` | Add or change a per-workload setting |
| `resourcerer-api-change` | Add a field to the CRD status or the HTTP API |
| `resourcerer-chart-values` | Add or change a chart value |
| `resourcerer-release` | Open, merge and release changes; recover a failed release |
| `resourcerer-design-issue` | Write a design issue for a larger feature |
| `resourcerer-docs` | Write or change the user docs and the documentation site |
