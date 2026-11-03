import { defineCollection } from 'astro:content';
import { glob } from 'astro/loaders';
import { docsSchema } from '@astrojs/starlight/schema';

// The pages live in the repository's docs/ directory, so they stay readable on GitHub and there is one copy.
export const collections = {
  docs: defineCollection({
    loader: glob({ base: '../docs', pattern: '**/[^_]*.{md,mdx}' }),
    schema: docsSchema(),
  }),
};
