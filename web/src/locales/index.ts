import { createI18n } from 'vue-i18n'
import { localStg } from '@/utils/storage'

export const LOCALE_STORAGE_KEY = 'locale'
export const DEFAULT_LOCALE = 'zh-CN'
export const SUPPORTED_LOCALES = ['zh-CN', 'en-US', 'de-DE'] as const

export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

const loadedLocales = new Set<string>()

const normalizeLocale = (raw?: string | null): AppLocale => {
  const value = String(raw || '').trim()
  if (!value) return DEFAULT_LOCALE

  const exact = SUPPORTED_LOCALES.find((locale) => locale.toLowerCase() === value.toLowerCase())
  if (exact) return exact

  const langPrefix = value.slice(0, 2).toLowerCase()
  if (langPrefix === 'en') return 'en-US'
  if (langPrefix === 'zh') return 'zh-CN'
  if (langPrefix === 'de') return 'de-DE'

  return DEFAULT_LOCALE
}

const resolveInitialLocale = (): AppLocale => {
  const fromStorage = localStg.getItem<string>(LOCALE_STORAGE_KEY)
  if (fromStorage) return normalizeLocale(fromStorage)

  const fromNavigator = navigator.languages?.[0] || navigator.language || DEFAULT_LOCALE
  return normalizeLocale(fromNavigator)
}

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: resolveInitialLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
})

async function loadLocaleMessages(locale: AppLocale) {
  if (loadedLocales.has(locale)) return
  const messages = await import(`./messages/${locale}.json`)
  i18n.global.setLocaleMessage(locale, messages.default)
  loadedLocales.add(locale)
}

export async function setI18nLocale(locale: string): Promise<AppLocale> {
  const normalized = normalizeLocale(locale)
  await loadLocaleMessages(normalized)
  i18n.global.locale.value = normalized
  document.documentElement.setAttribute('lang', normalized)
  localStg.setItem(LOCALE_STORAGE_KEY, normalized)
  return normalized
}

export async function setupI18n(defaultLocale?: string) {
  const fromStorage = localStg.getItem<string>(LOCALE_STORAGE_KEY)
  const locale = normalizeLocale(fromStorage || defaultLocale || i18n.global.locale.value)
  await setI18nLocale(locale)
  return i18n
}
