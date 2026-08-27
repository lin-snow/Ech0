// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const THEME_LIGHT = 'light' as const

const THEME_COLOR_FALLBACK_LIGHT = '#f4f1ec'

function syncThemeColorMeta(): void {
  const apply = () => {
    const rootStyles = getComputedStyle(document.documentElement)
    const chromeColor = rootStyles.getPropertyValue('--color-chrome-theme').trim()
    const canvasColor = rootStyles.getPropertyValue('--color-bg-canvas').trim()
    const next = chromeColor || canvasColor || THEME_COLOR_FALLBACK_LIGHT
    let meta = document.querySelector<HTMLMetaElement>('meta[name="theme-color"]')
    if (!meta) {
      meta = document.createElement('meta')
      meta.setAttribute('name', 'theme-color')
      document.head.appendChild(meta)
    }
    meta.setAttribute('content', next)
  }
  requestAnimationFrame(apply)
}

export function applyWebRootThemeClass(): typeof THEME_LIGHT {
  const root = document.documentElement
  root.classList.remove('light', 'dark', 'sunny')
  root.classList.add(THEME_LIGHT)
  syncThemeColorMeta()
  return THEME_LIGHT
}

if (typeof document !== 'undefined') {
  applyWebRootThemeClass()
}
