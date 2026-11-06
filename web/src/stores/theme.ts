// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { defineStore } from 'pinia'
import { ref } from 'vue'
import { localStg } from '@/utils/storage'

export type ThemeMode = 'light' | 'dark' | 'sunny'
type ThemeType = 'light' | 'dark' | 'sunny'
const THEME_COLOR_META_NAME = 'theme-color'
const THEME_COLOR_FALLBACK: Record<ThemeType, string> = {
  light: '#f4f1ec',
  dark: '#333333',
  sunny: '#eeece6',
}

export const useThemeStore = defineStore('themeStore', () => {
  const savedThemeMode = localStg.getItem('themeMode')
  const savedTheme = localStg.getItem('theme')

  const mode = ref<ThemeMode>(
    savedThemeMode === 'light' || savedThemeMode === 'dark' || savedThemeMode === 'sunny'
      ? savedThemeMode
      : 'light',
  )
  const theme = ref<ThemeType>(
    savedTheme === 'light' || savedTheme === 'dark' || savedTheme === 'sunny'
      ? savedTheme
      : 'light',
  )

  const applyThemeToggle = () => {
    if (mode.value === 'light') {
      mode.value = 'sunny'
    } else if (mode.value === 'sunny') {
      mode.value = 'dark'
    } else {
      mode.value = 'light'
    }

    applyTheme()
    localStg.setItem('themeMode', mode.value)
  }

  let isTransitioning = false

  const toggleTheme = async () => {
    if (isTransitioning) return

    const prefersReducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches

    type ViewTransitionLike = {
      ready: Promise<void>
      finished: Promise<void>
      updateCallbackDone?: Promise<void>
    }
    const startViewTransition = (
      document as Document & {
        startViewTransition?: (callback: () => void) => ViewTransitionLike
      }
    ).startViewTransition?.bind(document)

    if (prefersReducedMotion || !startViewTransition) {
      applyThemeToggle()
      return
    }

    isTransitioning = true
    try {
      const transition = startViewTransition(() => {
        applyThemeToggle()
      })
      await transition.finished
    } finally {
      isTransitioning = false
    }
  }

  const applyTheme = () => {
    switch (mode.value) {
      case 'light':
        theme.value = 'light'
        break
      case 'dark':
        theme.value = 'dark'
        break
      case 'sunny':
        theme.value = 'sunny'
        break
    }

    const rootEl = document.documentElement
    rootEl.classList.remove('light', 'dark', 'sunny')
    rootEl.classList.add(theme.value)
    // The chosen theme can differ from the OS preference; mirror it onto color-scheme so
    // native scrollbars, form controls and the UA canvas follow the app, not the device.
    rootEl.style.colorScheme = theme.value === 'dark' ? 'dark' : 'light'
    syncThemeColorMeta()
    localStg.setItem('theme', theme.value)
  }

  const syncThemeColorMeta = () => {
    const rootStyles = getComputedStyle(document.documentElement)
    const chromeColor = rootStyles.getPropertyValue('--color-chrome-theme').trim()
    const canvasColor = rootStyles.getPropertyValue('--color-bg-canvas').trim()
    const nextThemeColor = chromeColor || canvasColor || THEME_COLOR_FALLBACK[theme.value]

    const metas = document.querySelectorAll<HTMLMetaElement>(
      `meta[name="${THEME_COLOR_META_NAME}"]`,
    )
    if (metas.length === 0) {
      const meta = document.createElement('meta')
      meta.setAttribute('name', THEME_COLOR_META_NAME)
      meta.setAttribute('content', nextThemeColor)
      document.head.appendChild(meta)
      return
    }

    // Every tag must agree: a tag left at a stale value can win under its own media query.
    metas.forEach((meta) => {
      meta.removeAttribute('media')
      meta.setAttribute('content', nextThemeColor)
    })
  }

  const init = () => {
    applyTheme()
  }

  return {
    theme,
    mode,
    toggleTheme,
    applyTheme,
    init,
  }
})
