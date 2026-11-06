// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { setActivePinia, createPinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { useThemeStore } from '@/stores/theme'

const readMeta = () =>
  [...document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')].map((meta) => ({
    content: meta.getAttribute('content'),
    media: meta.getAttribute('media'),
  }))

let headBackup = ''

describe('theme store chrome colour', () => {
  beforeEach(() => {
    headBackup = document.head.innerHTML
  })

  afterEach(() => {
    document.head.innerHTML = headBackup
  })

  it('flattens legacy media-scoped theme-color tags onto the chosen theme', () => {
    // The shape index.html shipped while the app theme was (wrongly) inferred from the OS:
    // a dark phone kept the #333333 tag active even though the app rendered light.
    document.head.innerHTML =
      '<meta name="theme-color" content="#f4f1ec" media="(prefers-color-scheme: light)" />' +
      '<meta name="theme-color" content="#333333" media="(prefers-color-scheme: dark)" />'

    setActivePinia(createPinia())
    useThemeStore().init()

    expect(readMeta()).toEqual([
      { content: '#f4f1ec', media: null },
      { content: '#f4f1ec', media: null },
    ])
    expect(document.documentElement.style.colorScheme).toBe('light')
  })

  it('follows the app theme, not the OS preference, when switching to dark', () => {
    setActivePinia(createPinia())
    const store = useThemeStore()

    store.mode = 'dark'
    store.applyTheme()

    expect(readMeta()).toEqual([{ content: '#333333', media: null }])
    expect(document.documentElement.style.colorScheme).toBe('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)
  })
})
