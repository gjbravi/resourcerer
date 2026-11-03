---
name: resourcerer-docs
description: Write or change resourcerer's user documentation (docs/) and its Zensical site (mkdocs.yml, website/), published on GitHub Pages. Use when behaviour, a setting, a chart value, a CRD field or the UI changes in a way users see, or when adding, moving or renaming a page.
---

# Documentation

The pages are Markdown in `docs/`, readable on GitHub as they are. [Zensical](https://zensical.org) (the Material
theme, `variant: modern`) builds them into the site at https://gjbravi.github.io/resourcerer/, configured in
`mkdocs.yml` at the repository root, and `.github/workflows/docs.yaml` deploys it on every push to `main` that touches
`docs/`, `website/` or `mkdocs.yml`.

## Where a page goes

| Section (tab) | Directory | For |
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

- **Front matter:** `title` and `description` (one sentence, for search engines and link previews). The page starts
  below it, without a `#` heading: the theme renders the title.
- **Add it to `nav` in `mkdocs.yml`**, under its tab, in reading order. `make lint-docs` fails on a page in `docs/`
  that is not in the nav, and on a nav entry without a page.
- **Admonitions** use the Material syntax, with the content indented by four spaces:

  ```markdown
  !!! warning "CRDs left behind"
      A release upgraded without its CRDs keeps working…
  ```

  Kinds in use: `note`, `tip`, `warning`, `danger`. `???` instead of `!!!` makes a collapsible one.
- **Tabs:** `=== "Helm"` followed by indented content (`pymdownx.tabbed`).
- **Links between pages are relative paths to the `.md` file** (`[Phases](../concepts/phases.md#what-to-do-in-each-phase)`),
  so they work on GitHub too; the strict build fails when the target page does not exist. Link to files outside
  `docs/` (the chart, `CONTRIBUTING.md`) with their full GitHub URL: the site only has `docs/`.
- **Moving or renaming a page** breaks links people saved; avoid it, or say so in the PR.
- **Diagrams** are ` ```mermaid ` blocks. `docs/assets/javascripts/diagrams.js` renders them with zoom, pan and full
  screen, in the colors of `docs/assets/stylesheets/extra.css`; the theme's own Mermaid support is bypassed because it
  renders into a closed shadow root that cannot be zoomed. `make lint-docs` with `MERMAID=1` checks that they render.
- **Code blocks** take a title: ` ```yaml title="values.yaml" `. PromQL blocks use ` ```promql `.
- **The home page** (`docs/index.md`) is HTML inside Markdown with the `rz-` classes of `extra.css`, like a landing
  page; keep its claims in step with the concept pages.
- **Tone:** concrete, short sentences, the user's decision first. Show the setting, its default and the trade-off;
  give a command or a manifest the reader can copy. Use fictional names (`payments/checkout`, `demo/api`) and nothing
  specific to a company's environment, as everywhere in the repository.
- **Other products' docs** can inspire structure and ideas, never text, and are not named.

## Preview and check

```bash
make docs-dev      # http://localhost:8000/resourcerer/ with live reload (creates website/.venv on first run)
make docs-build    # what CI runs: zensical build --strict, fails on broken links and other warnings
make lint-docs     # pages in the nav, paths and make targets in backticks exist (MERMAID=1 renders the diagrams)
```

Check new pages in both themes (the toggle is in the header) and at a phone width.

## The site itself

- `mkdocs.yml`: site URL, theme and features (tabs, instant navigation, search), Markdown extensions, and the nav.
- `docs/assets/stylesheets/extra.css`: the lavender theme from the UI's `web/src/styles.css`, the home page and the
  diagram styles.
- `docs/assets/javascripts/diagrams.js`: Mermaid rendering with zoom; it hooks the theme's instant navigation through
  `document$`.
- `website/requirements.txt`: the pinned Zensical version, updated by Dependabot. `website/.venv` and the built
  `website/site` are not committed.
