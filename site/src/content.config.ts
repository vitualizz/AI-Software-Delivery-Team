import { defineCollection, z } from 'astro:content'
import { glob } from 'astro/loaders'

const docs = defineCollection({
  loader: glob({ pattern: '**/*.{md,mdx}', base: './src/content/docs' }),
  schema: z.object({
    title: z.string(),
    description: z.string(),
    order: z.number(),
    locale: z.enum(['en', 'es']),
  }),
})

export const collections = { docs }
