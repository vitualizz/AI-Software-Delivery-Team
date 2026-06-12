import type { APIRoute } from 'astro'
import { getCollection } from 'astro:content'

export const GET: APIRoute = async () => {
  const docs = await getCollection('docs')
  const index = docs.map((entry) => ({
    id: entry.id,
    slug: entry.id.replace(/^(en|es)\//, ''),
    locale: entry.data.locale,
    title: entry.data.title,
    description: entry.data.description,
  }))
  return new Response(JSON.stringify(index), {
    headers: { 'Content-Type': 'application/json' },
  })
}
