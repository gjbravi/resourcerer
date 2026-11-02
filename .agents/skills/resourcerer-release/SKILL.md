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
  passed CI (an artifact `ci-passed-<tree hash>` exists).
- `passed` uploads that artifact.
- e2e builds the image in the background while kind starts.
- `docs` runs on every change, docs-only ones included: `make lint-docs` with `MERMAID=1` checks that paths, make
  targets and skills mentioned in the docs exist and that the Mermaid diagrams render.

## Releases (release-please)

1. Merging `feat`/`fix` commits makes release-please open or update the PR `chore: release X.Y.Z` on
   `release-please--branches--main`. Versions bump minor for features while below 1.0.
2. release-please runs with a token of the repository's GitHub App (`vars.RELEASE_APP_ID`,
   `secrets.RELEASE_APP_PRIVATE_KEY`), so its PR triggers CI and the title check like any other; a PR opened with
   `GITHUB_TOKEN` would trigger no workflows.
3. **Merge the release PR only once its checks pass** (the ruleset requires them). The release job's `verify` step
   looks for `ci-passed-<tree>` of the merged commit; merging early, with a bypass, makes it fail.
4. After the merge, `verify` passes, then `image` and `chart` publish to GHCR with signatures.

### Recovering a failed `verify`

```bash
tree=$(git rev-parse '<merge commit>^{tree}')
gh run list --branch main --workflow CI --limit 1   # CI on the push of the merge commit; wait for it to finish
gh api "repos/<owner>/<repo>/actions/artifacts?name=ci-passed-$tree" -q '.artifacts | length'   # must be 1
gh run rerun <release run id> --failed
```

Never re-tag or delete a published release; fix forward with a new version.
