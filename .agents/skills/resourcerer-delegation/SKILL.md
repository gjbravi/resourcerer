---
name: resourcerer-delegation
description: Decide what to hand to subagents or parallel sessions in resourcerer and how, so features land faster without merge pain. Use at the start of a feature that touches more than one layer (API types, engine, controller, applier, server, UI, chart), when a task starts with broad exploration or prior-art lookup, or when several features or fixes could run at once.
---

# Delegating work in resourcerer

Local-only skill (gitignored). It is built on how features here actually look, so check the facts below before
splitting work differently.

## What the history says

- **Features are vertical.** Almost every `feat` touches `api/v1alpha1`, `internal/server`, `web/src` and
  `charts/resourcerer` together: 20–40 files, 500–4,700 lines. The UI is usually the largest share (7–13 files).
- **The layers meet in a few hot files**: `web/src/pages/Details.tsx`, `web/src/styles.css`, `web/src/api.ts`,
  `internal/server/server.go`, `internal/server/types.go`, `internal/recommender/recommender.go`,
  `test/e2e/run.sh`, `charts/resourcerer/values.yaml`, the CRD types and their generated files.
- **Unit checks are fast**: `go test ./...` ~8s, vitest ~3s, `tsc` ~2s. Running them in the background saves nothing.
- **kind is the slow part, mostly when cold**: on a warm local cluster `make kind-deploy e2e` is ~1.5 min (image
  ~5–15s, install ~18s, e2e ~48s). A fresh cluster costs ~5 min more (`make kind-up` ~2.5 min, then ~2 min waiting
  for Prometheus history). CI always starts cold: ~6.5 min. Never delete the local cluster between tasks.

So parallelism pays in three places: **reading** (no conflicts), **the UI against a fixed contract** (the largest,
most separable chunk), and **work that overlaps the kind wait**. Splitting the Go side further does not pay: the engine,
recommender, applier and server changes of one feature are small and coupled through the same status fields.

## What to parallelise

| Do in parallel | How | Why it pays |
|---|---|---|
| Exploration: code paths, the design issue (#26–#31), how Atlas/Orion do it (`resourcerer-prior-art`) | 2–3 `explore` subagents in one message, before any design | Read-only, no conflicts; keeps reference docs out of the main context |
| UI of a feature, once the contract is fixed | One `general` subagent owning `web/src`, while the main agent does the Go side | UI is the biggest chunk and only depends on `web/src/api.ts` |
| Review while kind runs | One `general` subagent reviewing `git diff main` against `AGENTS.md`, while the main agent runs `make kind-deploy e2e` | Turns idle minutes into a second pair of eyes |
| Docs sweep while kind runs | Same window: `docs/architecture.md`, chart README table, `AGENTS.md`/skills when a flow changed | Mechanical, read the diff, write prose; no code conflicts |
| Independent features or fixes | Separate sessions in separate worktrees (OpenChamber "new worktree", or `git worktree add`) | No shared context needed; only the hot files can conflict |

## What not to parallelise

- **The contract**: CRD fields (`api/v1alpha1` + `make generate`), `internal/server/types.go` and `web/src/api.ts`,
  setting names and defaults, chart value names. The main agent writes these first; everything else follows them.
- **Design and decision rules** (engine, policy, verdict): never delegated.
- **The Go side of one feature** split across subagents: coupled, small, shares `recommender.go` and `server.go`.
- **Chart values separately from the Go that reads them**: the value, the schema, the defaults ConfigMap and the
  setting in `internal/settings` change together; keep them with the Go side.
- **`make lint test` in the background**: seconds; just run it.
- **Two features that both change the CRD types or `Details.tsx`** in parallel worktrees: they will conflict in
  generated files and the busiest UI file. Sequence them, or land the shared contract first and rebase.
- **A second kind cluster by default**: three nodes plus Prometheus each. Only when two worktrees need e2e at the
  same moment (see below).

## The flow of a multi-layer feature

1. **Explore** (parallel, read-only). Ask for `file:line` answers and short summaries, never file dumps.
2. **Fix the contract** (main agent): types, `make generate`, `types.go` ↔ `api.ts`, setting and value names.
   Commit it locally so the UI subagent and you read the same types.
3. **Build** (parallel): UI subagent on `web/src` (`background: true`); main agent on Go, chart and the e2e check.
   If the UI needs a server field the contract missed, it stops and asks; it never edits Go.
4. **Integrate**: read the subagent's diff, not its summary. Run `make lint test`.
5. **Verify** (parallel): main agent runs `make kind-deploy e2e`; meanwhile one subagent reviews the diff and one
   does the docs sweep. Then the UI in the browser, light and dark.

A feature that touches a single layer, or under ~300 lines, is faster done alone: the prompt would be as long as
the change. Explore in parallel anyway when the code is unfamiliar.

## Prompts

A subagent starts with no context. Every prompt carries:

```text
Repository: /Users/guilherme.bravi/projects/resourcerer. Read AGENTS.md first.
Skill: .agents/skills/<skill>/SKILL.md.
Goal: <one sentence: what the user sees or the code does>.
Contract (fixed; tell me if it is wrong instead of changing it): <types and fields, with file paths>.
You own: <paths>. Edit nothing else; if you must, stop and say why.
Done when: <tests to add> and `<command>` passes; include its output.
Do not commit, push, open PRs or run make kind-deploy/e2e.
Public repository: no company, cluster or host names, no real numbers; use payments/checkout, demo/api.
Reply with: files changed (one line each), commands run with results, open questions.
```

Additions per role:

- **Explore**: "Answer only, edit nothing. Paths as `file:line`, at most 10 quoted lines. Thoroughness: <quick |
  medium | very thorough>."
- **UI** (`resourcerer-ui`): "You own `web/src` except `api.ts`. Logic in pure tested modules (`verdict.ts`,
  `format.ts`, …), `rz-` classes and theme variables from `styles.css`, shimmer skeletons, no new dependencies.
  Done when `cd web && npm run lint && npm run typecheck && npm test` passes."
- **Review**: "Review `git diff main` against AGENTS.md: conventions, comment density, error wrapping, degrade-don't-
  block, tests for every decision, chart value documented in values, schema and README, open-source hygiene (grep
  for company names, hostnames, IPs). Edit nothing; list findings as `file:line` with a suggested fix, most
  important first."
- **Docs sweep**: "From `git diff main`, update `docs/architecture.md`, the chart README table and any skill or
  `AGENTS.md` section whose explanation is now stale. Match the existing tone. Done when `make lint-docs` reports
  nothing new."

## Parallel worktrees and kind

One kind cluster (`kind-resourcerer`) is enough almost always: worktrees develop in parallel and take turns on e2e.
When two must run e2e at the same time, give the second its own cluster, image tag and ports:

```bash
CLUSTER_NAME=rz-<task> IMG=ghcr.io/gjbravi/resourcerer:<task> E2E_PORT=18180 E2E_METRICS_PORT=18191 \
  make kind-up kind-deploy e2e
CLUSTER_NAME=rz-<task> make kind-down   # when done
```

A shared `IMG` tag lets one worktree's deploy load the other's image; a shared port-forward port answers for the
wrong cluster.

## Things that go wrong

| Symptom | Fix |
|---|---|
| UI and server disagree on a field | The contract was not fixed before step 3: fix `types.go` and `api.ts` yourself, then resume the UI. |
| `make verify-generate` fails after integrating | Someone touched `api/` after step 2: regenerate once, in the main agent. |
| Two worktrees conflict in `Details.tsx` or the CRD | They should have been sequenced: land the one with the contract change first, rebase the other. |
| A subagent says it is done but nothing was run | The prompt lacked "Done when … include its output": ask for the output. |
| The summary and the diff disagree | Trust the diff. |
