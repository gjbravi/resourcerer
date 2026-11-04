---
name: resourcerer-release
description: Open, check and merge pull requests in resourcerer, and cut or recover a release. Use when opening a PR, merging, waiting on CI, or when the Release workflow fails.
---

# PRs and releases

## Pull requests

- Branch from `main`; one change per PR.
- Title and commits follow Conventional Commits (`feat(scope): ...`, `fix(scope): ...`, `docs: ...`, `ci: ...`);
  `pr-title.yaml` checks it. No AI attribution trailers.
- Body: the template's sections, concretely: what changed and why, then the commands run and their results
  (`make lint test`, `make e2e`, manual checks in kind). Link the issue (`Closes #N`).
- Before pushing, grep the diff and the body for anything specific to a company's environment (cluster names,
  hostnames, account IDs): the repository is public.
- `main` requires the checks `go`, `web`, `chart` and `e2e` and a squash merge.

## CI

`.github/workflows/ci.yaml`:

- `gate` skips `go`/`web`/`chart`/`e2e` when only docs or metadata changed, and on a push to `main` whose tree already
  passed CI (an artifact `ci-passed-<tree hash>` exists). It skips only `e2e` when nothing but the UI changed
  (`web/` besides `web/embed.go`): e2e checks the API and the cluster, not the UI. Such a PR uploads no
  `ci-passed` artifact, so its push to `main` runs `go`, `web` and `chart` again; the release PR's dispatched run
  always runs e2e.
- `passed` uploads that artifact, and on `workflow_dispatch` runs also sets the commit statuses `go`, `web`, `chart`
  and `e2e`, so dispatched runs can satisfy the branch rules.
- e2e builds the image in the background while kind starts, and `make kind-up` installs metrics-server and
  Prometheus in the background while Argo Rollouts and the demo go in.
- `docs` runs on every change, docs-only ones included: `make lint-docs` with `MERMAID=1` checks that paths, make
  targets and skills mentioned in the docs exist and that the Mermaid diagrams render.

## Releases (release-please)

1. Merging `feat`/`fix` commits makes release-please open or update the PR `chore: release X.Y.Z` on
   `release-please--branches--main`. Versions bump minor for features while below 1.0.
2. The Release workflow dispatches CI on that branch (PRs created by the workflow token do not trigger workflows).
3. **Wait until that CI run finishes and the PR shows its checks before merging the release PR.** The release job's
   `verify` step looks for `ci-passed-<tree>` of the merged commit; merging early makes it fail.
4. After the merge, `verify` passes, then `image` and `chart` publish the image and both charts to GHCR with
   signatures. The release attaches `resourcerer-crds-X.Y.Z.yaml` and the install steps for the two releases.

### Release candidates

For an RC line, `release-please-config.json` uses the `prerelease` strategy: `versioning: prerelease`,
`prerelease-type: rc` and `prerelease: true` produce `X.Y.Z-rc.N` and mark the GitHub release as a prerelease. A
`Release-As` footer in a commit on `main` sets the next version regardless of the commits:

```text
chore: start the 1.0 release candidates

Release-As: 1.0.0-rc.1
```

Squash merges have an empty commit body by default, so pass the footer when merging the PR that starts the line:

```bash
gh pr merge <number> --squash --body "Release-As: 1.0.0-rc.1"
```

To cut the stable release, remove those three keys from the config and land a commit with `Release-As: 1.0.0`.
Never re-tag or delete an RC; fix forward with the next `rc.N`.

### Recovering a failed `verify`

```bash
tree=$(git rev-parse '<merge commit>^{tree}')
gh run list --branch release-please--branches--main --workflow CI --limit 1   # wait for it to finish
gh api "repos/<owner>/<repo>/actions/artifacts?name=ci-passed-$tree" -q '.artifacts | length'   # must be 1
gh run rerun <release run id> --failed
```

Never re-tag or delete a published release; fix forward with a new version.
