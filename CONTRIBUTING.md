# Contributing

Thanks for your interest in resourcerer.

## Workflow

1. Open an issue for anything beyond a small fix, so we can agree on the approach first.
2. Fork, create a branch, and keep each pull request focused on one change.
3. Run `make lint test` before pushing. For controller or chart changes, also run `make kind-up kind-deploy e2e`.
4. Use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages and PR titles,
   for example `feat(ui): show usage percentiles` or `fix(applier): keep user limits`.
   Release notes are generated from them.

## Code guidelines

[docs/architecture.md](docs/architecture.md) explains how the components work, with diagrams.
[AGENTS.md](AGENTS.md) describes the layout, commands, conventions and design decisions in more detail, and
`.agents/skills/` has step-by-step guides for recurring tasks (dev loop on kind, PromQL, UI, CRD and API changes,
chart values, releases, design issues). They are written for coding agents and are just as useful to people.

- Go: `gofmt`/`goimports`, errors wrapped with context, table-driven tests for logic.
- Recommendation rules live in `internal/engine` (pure, no I/O) and `internal/settings`; controllers and handlers stay thin.
- UI: decisions about what to show live in pure, tested modules (`web/src/verdict.ts`, `placement.ts`, `reliability.ts`); components reuse the `rz-` styles and theme variables; actions that change the cluster ask for confirmation.
- Chart: every new value is documented in `values.yaml` and validated in `values.schema.json`.

## Reporting security issues

See [SECURITY.md](SECURITY.md). Please do not open public issues for vulnerabilities.
