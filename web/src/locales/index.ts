// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { createI18n } from 'vue-i18n'
import { localStg } from '@/utils/storage'

export const LOCALE_STORAGE_KEY = 'locale'
export const DEFAULT_LOCALE = 'zh-CN'
export const FALLBACK_LOCALE = 'en-US'
export const SUPPORTED_LOCALES = ['zh-CN', 'en-US', 'de-DE', 'ja-JP'] as const

export type AppLocale = (typeof SUPPORTED_LOCALES)[number]

export const LOCALE_ENDONYMS: Record<AppLocale, string> = {
  'zh-CN': '简体中文',
  'en-US': 'English',
  'de-DE': 'Deutsch',
  'ja-JP': '日本語',
}

export const LOCALE_OPTIONS = SUPPORTED_LOCALES.map((value) => ({
  value,
  label: LOCALE_ENDONYMS[value],
}))

const loadedLocales = new Set<string>()

const toSupported = (raw?: string | null): AppLocale | null => {
  const value = String(raw || '').trim()
  if (!value) return null

  const exact = SUPPORTED_LOCALES.find((locale) => locale.toLowerCase() === value.toLowerCase())
  if (exact) return exact

  const langPrefix = value.slice(0, 2).toLowerCase()
  if (langPrefix === 'en') return 'en-US'
  if (langPrefix === 'zh') return 'zh-CN'
  if (langPrefix === 'de') return 'de-DE'
  if (langPrefix === 'ja') return 'ja-JP'

  return null
}

const normalizeLocale = (raw?: string | null): AppLocale => toSupported(raw) ?? FALLBACK_LOCALE

const initialLocale =
  toSupported(localStg.getItem<string>(LOCALE_STORAGE_KEY)) ||
  toSupported(navigator.languages?.[0] || navigator.language) ||
  FALLBACK_LOCALE

export const i18n = createI18n({
  legacy: false,
  globalInjection: true,
  locale: initialLocale,
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
  const fromNavigator = navigator.languages?.[0] || navigator.language
  const locale =
    toSupported(fromStorage) ||
    toSupported(fromNavigator) ||
    toSupported(defaultLocale) ||
    FALLBACK_LOCALE
  await setI18nLocale(locale)
  return i18n
}
