import { defineConfig } from 'astro/config'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath } from 'node:url'
import mdx from '@astrojs/mdx'
import preact from '@astrojs/preact'

export default defineConfig({
  site: 'https://vitualizz.github.io',
  base: '/asdt',
  output: 'static',
  integrations: [mdx(), preact()],
  vite: {
    plugins: [tailwindcss()],
    resolve: {
      alias: {
        '@components': fileURLToPath(new URL('./src/components', import.meta.url)),
        '@layouts': fileURLToPath(new URL('./src/layouts', import.meta.url)),
        '@i18n': fileURLToPath(new URL('./src/i18n', import.meta.url)),
        '@data': fileURLToPath(new URL('./src/data', import.meta.url)),
      },
    },
  },
  i18n: {
    defaultLocale: 'en',
    locales: ['en', 'es'],
    fallback: { es: 'en' },
    routing: {
      prefixDefaultLocale: false,
      fallbackType: 'rewrite',
    },
  },
})
