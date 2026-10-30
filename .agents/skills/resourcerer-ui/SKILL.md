---
name: resourcerer-ui
description: Change resourcerer's web UI (web/src): pages, components, charts, styles and the logic behind what is shown. Use for any UI feature, layout change or visual fix.
---

# resourcerer UI

React 19, TypeScript, Vite, React Query (`web/src/hooks.ts`), Recharts 3 (`web/src/chart.tsx`), one stylesheet
(`web/src/styles.css`). No component library.

## Where things go

- **Types:** `web/src/api.ts` mirrors `internal/server/types.go` and the CRD status. Change both together.
- **Logic:** decisions about what to show live in pure modules with vitest tests next to them: `verdict.ts`,
  `placement.ts`, `reliability.ts`, `format.ts`. Components render; they do not decide.
- **Data:** one hook per endpoint in `hooks.ts`. Polling only where the data changes on its own (workload 5s, list
  15s, charts 30s); expensive live queries get a `staleTime` and are fetched only when shown (`enabled`).
- **Components:** shared pieces in `components.tsx` (`Delta`, `Modal`, `StatusBadge`, `ErrorBox`, …), drawings in
  `art.tsx`, pages in `pages/`.

## Look

- Theme "Arcane lavender": use the variables, never raw colors in components: `--accent`, `--accent-soft`,
  `--accent-line`, `--text`, `--ink2`, `--dim`, `--faint`, `--panel`, `--border`, `--track`, `--bad`, and the chart
  series `--c-usage`, `--c-usage50`, `--c-cur`, `--c-rec`, `--c-curlim`, `--c-reclim`, `--c-alloc`. Dark theme comes
  for free.
- Classes are prefixed `rz-`; reuse `rz-card`, `rz-cardhd`, `rz-sechd`, `rz-zoom` (pill toggles), `rz-tablewrap`,
  `rz-table`, `rz-link`, `rz-dim`, `rz-xs` before adding new ones.
- Charts: one color per series, no area fills except the usage band, a light tooltip that follows the cursor
  (`rz-tt`). Measure width with `useWidth`; a chart inside a grid needs `minmax(0, 1fr)` or it overflows.
- Money with `fmtMoney`, CPU with `fmtCores`/`fmtCpu`, memory with `fmtMem`.
- Loading: shimmer skeletons shaped like the content (#31); no "Loading…" texts and no animated mascots.
- Copy is short, sentence case, in English; units always shown.

## UX rules learned so far

- Lead with the decision, then the evidence: the Impact tab opens with the verdict, then before/after, then usage,
  then the cost breakdown collapsed.
- Do not repeat a number that is already visible on the same screen; if two places need it, one of them links.
- Put controls next to what they control (chart zoom in the chart's header).
- Every estimate says what it is based on (window, rates, coverage) where the user can see it.
- Tooltips that can be clipped by a scroll container use `position: fixed` and flip near the bottom edge.
- `prefers-reduced-motion` turns animations off.

## Check your work

```bash
cd web && npm run typecheck && npm run lint && npm run test
```

Then deploy to kind and look at the page in light and dark themes, at a narrow width, and with throttled network for
loading states. Do not leave `tsc -b` output (`*.js` next to the sources) in the tree: use `npm run typecheck`.
