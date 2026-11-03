import { fileURLToPath } from 'node:url';
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import mermaid from 'astro-mermaid';
import markdownLinks from './src/markdown-links.mjs';
import { checkSidebar, sidebar } from './src/sidebar.mjs';

const repoUrl = 'https://github.com/gjbravi/resourcerer';
const base = '/resourcerer';
const repoDir = fileURLToPath(new URL('..', import.meta.url)).replace(/\/$/, '');
const docsDir = `${repoDir}/docs`;

checkSidebar(docsDir);

// MDX pages in docs/ import packages installed here: resolve their bare imports as if from this directory.
const docsImports = {
  name: 'docs-imports',
  enforce: 'pre',
  resolveId(source, importer, options) {
    if (!importer?.startsWith(docsDir) || /^[./]|^\0|^astro:|^virtual:/.test(source)) return null;
    return this.resolve(source, fileURLToPath(new URL('./astro.config.mjs', import.meta.url)), {
      ...options,
      skipSelf: true,
    });
  },
};

export default defineConfig({
  site: 'https://gjbravi.github.io',
  base,
  markdown: {
    rehypePlugins: [[markdownLinks, { docsDir, repoDir, base, repoUrl }]],
  },
  vite: { plugins: [docsImports] },
  redirects: {
    '/apply-policy': `${base}/guides/apply-policies/`,
    '/reliability-signals': `${base}/guides/reliability-signals/`,
  },
  integrations: [
    mermaid({ theme: 'default', autoTheme: true }),
    starlight({
      title: 'resourcerer',
      description:
        'Self-hosted Kubernetes rightsizing: CPU and memory recommendations from Prometheus, applied with confidence.',
      logo: { src: '../docs/logo.svg' },
      favicon: '/favicon.svg',
      customCss: ['./src/styles/custom.css'],
      social: [{ icon: 'github', label: 'GitHub', href: repoUrl }],
      // Page paths are relative to this directory (../docs/…), so the links resolve to docs/.
      editLink: { baseUrl: `${repoUrl}/edit/main/website/` },
      lastUpdated: true,
      markdown: { processedDirs: ['../docs/'] },
      // No PromQL grammar ships with the highlighter; plain text beats a warning per block.
      expressiveCode: { shiki: { langAlias: { promql: 'txt' } } },
      sidebar,
    }),
  ],
});
