---
name: resourcerer-docs
description: Write or change resourcerer's user documentation (docs/) and its Starlight site (website/), published on GitHub Pages. Use when behaviour, a setting, a chart value, a CRD field or the UI changes in a way users see, or when adding, moving or renaming a page.
---

# Documentation

The pages are Markdown in `docs/`, readable on GitHub as they are. `website/` is an Astro + Starlight project that
turns them into the site at https://gjbravi.github.io/resourcerer/, built and deployed by `.github/workflows/docs.yaml`
on every push to `main` that touches `docs/` or `website/`.

## Where a page goes

| Section | Directory | For |
|---|---|---|
| Getting started | `docs/getting-started/` | Install, first recommendation, upgrade: the path of someone new |
| Concepts | `docs/concepts/` | What the UI shows and why: one page per idea (phases, verdict, costs…) |
| Configuration | `docs/configuration/` | How to choose settings and chart values, with the trade-offs |
| Guides | `docs/guides/` | Tasks and setups: apply, auto-deploy, apply policies, GitOps, Karpenter… |
| Reference | `docs/reference/` | Complete tables: every setting, value, CRD field, endpoint, metric |
| Project | `docs/architecture.md`, `docs/roadmap.md` | How it works under the hood (contributors too), what comes next |

Explain a behaviour once, in its concept or guide page, and link to it; reference pages list and link, they do not
explain again. `docs/architecture.md` keeps the component view and the sequence diagrams for contributors.

## When a change needs a docs update

| Change | Pages |
|---|---|
| A setting (`internal/settings`) | `reference/settings.md`, and its configuration page |
| A chart value | `reference/helm-values.md`, and its configuration or guide page |
| A CRD field (`api/v1alpha1`) | `reference/custom-resources.md`, and the concept or guide that uses it |
| An HTTP endpoint | `reference/http-api.md` |
| A metric or event | `reference/metrics-and-events.md` |
| RBAC in the chart | `reference/permissions.md` |
| A recommendation rule (`internal/engine`) | `concepts/recommendations.md`, `configuration/goals-and-limits.md`, `architecture.md` |
| What the UI shows | `guides/using-the-ui.md`, and the concept page behind it |
| A breaking change | `getting-started/upgrade-and-uninstall.md` when an upgrade needs a step |

## Writing a page

- **Front matter:** `title` (required by Starlight) and `description` (one sentence, used by search engines and link
  cards). Quote the description if it contains `: `. The page starts below it, without a `#` heading: Starlight renders
  the title.
- **Add it to the sidebar** in `website/src/sidebar.mjs`, in reading order. The build fails when a page in `docs/` is
  missing from the sidebar or an entry has no page.
- **`.md` by default.** Asides need no MDX: `:::note`, `:::tip`, `:::caution`, `:::danger`, optionally with a title
  (`:::caution[Beta]`). Use `.mdx` only for components (`import { Tabs, TabItem } from '@astrojs/starlight/components'`),
  as `docs/index.mdx` does.
- **Links are relative paths to the `.md` file** (`[Phases](../concepts/phases.md#what-to-do-in-each-phase)`), so they
  work on GitHub too. `website/src/markdown-links.mjs` turns them into site URLs, turns links to other files of the
  repository (`../../charts/resourcerer/values.yaml`) into GitHub URLs, and fails the build when the target does not
  exist. Anchors are not checked: keep them in step with the headings.
- **Moving or renaming a page** breaks external links: add the old path to `redirects` in `website/astro.config.mjs`.
- **Diagrams** are ` ```mermaid ` blocks, rendered in the browser and following the theme. `make lint-docs` with
  `MERMAID=1` checks that they render.
- **Code blocks** take a title: ` ```yaml title="values.yaml" `. PromQL blocks use ` ```promql ` (shown as plain text).
- **Tone:** concrete, short sentences, the user's decision first. Show the setting, its default and the trade-off; give
  a command or a manifest the reader can copy. Use fictional names (`payments/checkout`, `demo/api`) and nothing
  specific to a company's environment, as everywhere in the repository.
- **Other products' docs** can inspire structure and ideas, never text, and are not named.

## Preview and check

```bash
make docs-dev      # http://localhost:4321/resourcerer/ with live reload
make docs-build    # what CI runs: fails on broken links, missing sidebar entries and invalid front matter
make lint-docs     # paths and make targets in backticks exist (MERMAID=1 also renders the diagrams)
```

Check new pages in both themes (the selector is in the header) and at a phone width.

## The site itself

- `website/astro.config.mjs`: site URL and base (`/resourcerer`), redirects, Starlight options.
- `website/src/content.config.ts`: loads the pages from `../docs`.
- `website/src/sidebar.mjs`: the sidebar and the check that it matches `docs/`.
- `website/src/styles/custom.css`: the theme, taken from the UI's `web/src/styles.css`.
- Dependencies are pinned in `website/package-lock.json` and updated by Dependabot.
