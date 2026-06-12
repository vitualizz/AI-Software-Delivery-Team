import type { UIStrings } from '@i18n/types'
import { en } from '@i18n/locales/en'
import { es } from '@i18n/locales/es'

export type Locale = 'en' | 'es'
export const defaultLocale: Locale = 'en'

const translations: Record<Locale, UIStrings> = { en, es }

export function useTranslations(lang: Locale): UIStrings {
  return translations[lang] ?? translations[defaultLocale]
}

export function getLangFromUrl(url: URL): Locale {
  const base = import.meta.env.BASE_URL
  const path = url.pathname.replace(base, '')
  const [first] = path.split('/').filter(Boolean)
  if (first === 'es') return 'es'
  return 'en'
}

export function getBaseHref(path: string): string {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  const clean = path.startsWith('/') ? path : `/${path}`
  return `${base}${clean}`
}

export function getLocalePath(path: string, lang: Locale): string {
  const base = import.meta.env.BASE_URL.replace(/\/$/, '')
  const clean = path.replace(/^\//, '')
  if (lang === 'en') return `${base}/${clean}`
  return `${base}/es/${clean}`
}
