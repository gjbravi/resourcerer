// The sidebar lists every page in order. Starlight can only generate groups for pages under src/content/docs, and
// the pages live in the repository's docs/ directory, so the list is explicit; checkSidebar fails the build when a
// page is missing from it or an entry has no page.
import { readdirSync } from 'node:fs';
import path from 'node:path';

const repoUrl = 'https://github.com/gjbravi/resourcerer';

export const sidebar = [
  {
    label: 'Getting started',
    items: [
      'getting-started/introduction',
      'getting-started/quickstart',
      'getting-started/requirements',
      'getting-started/installation',
      'getting-started/upgrade-and-uninstall',
    ],
  },
  {
    label: 'Concepts',
    items: [
      'concepts/how-it-works',
      'concepts/recommendations',
      'concepts/phases',
      'concepts/verdict-and-evidence',
      'concepts/costs',
      'concepts/workloads',
      'concepts/glossary',
    ],
  },
  {
    label: 'Configuration',
    items: [
      { slug: 'configuration/overview', label: 'Overview' },
      'configuration/goals-and-limits',
      'configuration/learning-period-and-schedule',
      'configuration/prometheus',
      'configuration/cost-estimates',
    ],
  },
  {
    label: 'Guides',
    items: [
      'guides/using-the-ui',
      'guides/apply-and-revert',
      'guides/auto-deploy',
      'guides/apply-policies',
      'guides/reliability-signals',
      'guides/gitops',
      'guides/argo-rollouts',
      'guides/karpenter-costs',
      'guides/exposing-the-ui',
      'guides/large-clusters',
      'guides/try-on-kind',
    ],
  },
  {
    label: 'Reference',
    items: [
      'reference/settings',
      'reference/helm-values',
      'reference/custom-resources',
      'reference/http-api',
      'reference/metrics-and-events',
      'reference/permissions',
      'reference/troubleshooting',
      'reference/faq',
    ],
  },
  {
    label: 'Project',
    items: [
      'architecture',
      'roadmap',
      { label: 'Contributing', link: `${repoUrl}/blob/main/CONTRIBUTING.md`, attrs: { target: '_blank' } },
      { label: 'Changelog', link: `${repoUrl}/blob/main/CHANGELOG.md`, attrs: { target: '_blank' } },
    ],
  },
];

const slugs = (items) =>
  items.flatMap((i) => (typeof i === 'string' ? [i] : i.slug ? [i.slug] : i.items ? slugs(i.items) : []));

export function checkSidebar(docsDir) {
  const pages = readdirSync(docsDir, { recursive: true })
    .filter((f) => /\.mdx?$/.test(f) && !path.basename(f).startsWith('_'))
    .map((f) => f.split(path.sep).join('/').replace(/\.mdx?$/, ''))
    .filter((p) => p !== 'index');
  const listed = slugs(sidebar);
  const problems = [
    ...pages.filter((p) => !listed.includes(p)).map((p) => `docs/${p} is not in website/src/sidebar.mjs`),
    ...listed.filter((s) => !pages.includes(s)).map((s) => `website/src/sidebar.mjs lists ${s}, which has no page`),
  ];
  if (problems.length) throw new Error(problems.join('\n'));
}
