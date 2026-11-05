---
name: resourcerer-prior-art
description: Look up how the two mature products resourcerer takes inspiration from (codenames Atlas and Orion) solve a problem, in their local docs. Use when designing a feature, naming a setting, choosing a default, writing an issue or docs, or when the user asks how "Atlas" or "Orion" does something.
---

# Prior art: Atlas and Orion

Local, gitignored copies of the public docs of two Kubernetes resource-optimization products, with brand names, domains
and component names replaced by codenames. Always use the codenames, in code, commits, issues, PRs and docs; never write
the real names into the repository.

| Codename | What it is | Strong on |
| --- | --- | --- |
| **Atlas** | Broad cost-automation platform: node autoscaling and rebalancing, spot, workload autoscaling (requests, HPA/VPA, in-place resize), storage, reporting | node-level automation, cost reporting, workload autoscaler annotations and configuration |
| **Orion** | Focused workload rightsizing: an Agent collects metrics, an ML backend recommends requests/limits, an Applier deploys them on a schedule | optimization goals (savings/balanced/reliability), learning period, auto-deploy, reliability bump-ups on OOM/CPU pressure, cluster/namespace/workload defaults |

## Where things are

- `references/atlas/INDEX.md`, `references/orion/INDEX.md`: section tree with a one-line summary per page. Start here.
- `references/atlas/<slug>.md`: one file per page.
- `references/orion/<section>__<page>.md`: one file per page; `__` stands for `/` in the original path.
- `references/FETCHED`: date of the last download.
- Domains became `atlas.example` / `orion.example`; links between pages point at the local files; links to pages that
  were not downloaded (API reference, unlisted pages) still point at the `.example` domain and cannot be opened.

## How to use

1. Grep both directories for the concept (`rg -il 'learning period|oom' references/`), then read the INDEX entry and
   the page.
2. Compare the two: where they agree is usually the safe default; where they differ, name the trade-off.
3. Bring back the idea, not the text: summarize in your own words and say which product it came from by codename.
   Do not copy pages into the repository.
4. Check what resourcerer already does (`docs/`, `internal/`) before proposing; say what it adds over the current
   behaviour.

## Refresh

`scripts/refresh.sh` downloads both sites again and rebuilds `references/` (a few minutes). It fails if a real name
survives the substitutions; add the missing pattern to `scripts/sanitize.py` and run it again. `scripts/` is the only
place that names the real products.
