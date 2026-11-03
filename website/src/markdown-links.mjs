// Pages link with relative paths, so they also work when read on GitHub. This rehype plugin turns a link to another
// page into its site URL and a link to any other file of the repository into its GitHub URL, and fails the build when
// the target does not exist. Anchors are kept as they are.
import { existsSync } from 'node:fs';
import path from 'node:path';

const walk = (node, visit) => {
  visit(node);
  for (const child of node.children ?? []) walk(child, visit);
};

const posix = (p) => p.split(path.sep).join('/');

export default function markdownLinks({ docsDir, repoDir, base, repoUrl }) {
  const root = base.replace(/\/$/, '');
  return (tree, file) => {
    const from = file.path ?? file.history?.[0];
    if (!from || !path.resolve(from).startsWith(docsDir + path.sep)) return;
    walk(tree, (node) => {
      const href = node.type === 'element' && node.tagName === 'a' ? node.properties?.href : undefined;
      if (typeof href !== 'string' || href === '' || /^[a-z][a-z0-9+.-]*:|^\/|^#/i.test(href)) return;
      const [target, hash] = href.split('#');
      const resolved = path.resolve(path.dirname(from), decodeURIComponent(target));
      if (!existsSync(resolved)) {
        throw new Error(`${posix(path.relative(repoDir, from))}: broken link to ${href}`);
      }
      const anchor = hash ? `#${hash}` : '';
      if (resolved.startsWith(docsDir + path.sep) && /\.mdx?$/.test(resolved)) {
        const slug = posix(path.relative(docsDir, resolved))
          .replace(/\.mdx?$/, '')
          .replace(/(^|\/)index$/, '')
          .toLowerCase();
        node.properties.href = `${root}/${slug}${slug ? '/' : ''}${anchor}`;
      } else {
        node.properties.href = `${repoUrl}/blob/main/${posix(path.relative(repoDir, resolved))}${anchor}`;
      }
    });
  };
}
