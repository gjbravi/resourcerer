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
| `internal/settings` | Settings, their defaults, validation and resolution (values → namespace → workload annotations) |
| `internal/signals` | Reliability signals: built-in toggles and PromQL templates declared in the values |
| `web/` | React 19 + TypeScript + Vite UI, React Query, Recharts 3; `web/src/api.ts` mirrors `internal/server/types.go` |
| `charts/resourcerer` | Helm chart; `values.schema.json` rejects unknown keys |
| `hack/` | kind environment: cluster, Prometheus, demo workloads, a fake Karpenter exporter |
| `test/e2e/run.sh` | End-to-end checks against kind |
| `docs/` | `architecture.md` (components, data flow, diagrams) and user docs such as `reliability-signals.md` |

## Commands

```bash
make lint            # golangci-lint, actionlint, eslint, prettier, tsc, helm lint, lint-docs
make lint-docs       # docs and skills match the repository; MERMAID=1 also renders the diagrams
make test            # Go and UI unit tests
make generate        # deepcopy + CRD into the chart; run after changing api/
make verify-generate # fails when generated files are stale (compares with git: commit first)
make kind-up         # kind cluster with Prometheus, metrics-server, VPA, demo workloads, fake Karpenter
make kind-deploy     # build the image, load it into kind, install the chart with hack/resourcerer-values.yaml
make e2e             # end-to-end checks against kind
make web-dev         # UI dev server proxied to a running server on :8080
```

Behind a TLS-intercepting proxy, pass `EXTRA_CA_CERT=/path/to/ca.pem` to `make kind-up`.

The UI of the kind install: `kubectl --context kind-resourcerer -n resourcerer port-forward svc/resourcerer 8080:80`.
A redeploy replaces the pods, so restart the port-forward afterwards.

## Definition of done

1. `make lint test` passes; after `api/` changes, `make generate` and commit the generated files.
2. Controller, server, chart or query changes: `make kind-deploy e2e` passes, with an e2e check for the new
   behaviour when it can be observed in kind (`test/e2e/run.sh`, using `pass`/`fail`/`wait_for`).
3. UI changes: checked in the browser against kind, light and dark themes.
4. Docs, chart values (`values.yaml` comments, schema, chart README table) and `hack/resourcerer-values.yaml` updated
   when behaviour or configuration changed; `docs/architecture.md`, this file and the skills updated when a
   component, flow or convention they describe changed (`make lint-docs` catches stale paths and targets, not stale
   explanations).

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
must not be stored as valid, and cap windows and points so large clusters stay cheap.

### UI

See the `resourcerer-ui` skill. In short: logic in pure, tested modules (`verdict.ts`, `placement.ts`,
`reliability.ts`, `format.ts`), components reuse the `rz-` classes and theme variables in `styles.css`, one color per
chart series, shimmer skeletons for loading (no loading texts, no mascots), and no new dependencies without a reason.

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
- **The Impact tab leads with a verdict** ("Safe to apply", "Review before applying", "Apply to fix reliability", …)
  built from history coverage, a replay of usage against the recommendation (`status.evidence`) and reliability
  signals. Signals the recommendation itself fixes (OOM kills, CPU throttling from 25%) count for applying.
- **Reliability signals** are built-in or declared as PromQL templates (`docs/reliability-signals.md`). A signal with
  no data does not apply to a workload.
- **Live data** for the workload page (reliability windows, placement) is computed by the server on request and
  cached; the controller's `status` stays small and drives lists and auto-deploy.

## Roadmap

Design issues describe the next steps; read the relevant one before working in that area:
reliability signals as custom resources (#28), `ApplyPolicy` (#29), `OptimizationPolicy` (#30), `CostProfile` (#27),
decision models with Jev/decider (#26), shimmer skeletons (#31).

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
| **Signal** | A reliability check, built-in (throttling, memory headroom, restarts…) or declared as a PromQL template. "Fixed by apply" when the recommendation itself addresses it. |
| **Cluster factor** | Calibrated rates over the configured rates: what capacity costs on this cluster relative to `cost.rates`. |
| **Workload factor** | The same for one workload's nodes, under `cost.allocation: node`. |
| **Coverage** | The share of a workload's requests that ran on nodes with a known price. |
| **Placement** | Where a workload's requests ran over the learning period: capacity types and instance types over time. |
| **Auto-deploy** | The applier applying `Ready`, pending recommendations without a click, once per hash. |

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
