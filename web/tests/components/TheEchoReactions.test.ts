// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { flushPromises, mount, type VueWrapper } from '@vue/test-utils'
import { createI18n } from 'vue-i18n'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import TheEchoReactions from '@/components/advanced/echo/cards/TheEchoReactions.vue'

// jsdom ships no matchMedia, and the picker skips its autofocus under reduced motion.
beforeEach(() => {
  window.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => {},
    removeListener: () => {},
    addEventListener: () => {},
    removeEventListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof window.matchMedia
})

vi.mock('@/service/api', () => ({
  fetchReactEcho: vi.fn(),
  fetchUnreactEcho: vi.fn(),
  fetchReactionCatalog: vi.fn(async () => ({ code: 1, data: [] })),
}))

vi.mock('@/service/request/shared', () => ({ isStaticMode: () => false }))
vi.mock('@/utils/toast', () => ({ theToast: { info: vi.fn(), error: vi.fn() } }))

const i18n = createI18n({
  legacy: false,
  locale: 'en',
  missingWarn: false,
  fallbackWarn: false,
  messages: {
    en: {
      reaction: {
        add: 'Add reaction',
        close: 'Close reaction',
        unavailable: 'Unavailable',
        like: 'Like',
        heart: 'Heart',
        haha: 'Haha',
        wow: 'Wow',
        sad: 'Sad',
        cry: 'Cry',
        party: 'Party',
      },
    },
  },
})

const wrappers: VueWrapper[] = []

const mountReactions = () => {
  const wrapper = mount(TheEchoReactions, {
    props: { echo: { id: 'echo-1', reactions: [] } },
    global: { plugins: [i18n] },
    attachTo: document.body,
  })
  wrappers.push(wrapper)
  return wrapper
}

const triggerButton = () => document.querySelector<HTMLButtonElement>('.reaction-trigger')!

afterEach(() => {
  wrappers.splice(0).forEach((wrapper) => wrapper.unmount())
  document.body.innerHTML = ''
})

describe('reaction picker interaction', () => {
  it('leaves focus outside the picker when opened by a pointer', async () => {
    mountReactions()

    triggerButton().dispatchEvent(
      new PointerEvent('pointerdown', { button: 0, bubbles: true, cancelable: true }),
    )
    await flushPromises()

    const items = [...document.querySelectorAll('.reaction-item')]
    expect(items).toHaveLength(7)
    expect(triggerButton().getAttribute('aria-expanded')).toBe('true')
    expect(items).not.toContain(document.activeElement)
  })

  it('moves focus to the active item when opened from the keyboard', async () => {
    mountReactions()

    // detail === 0 is what Enter/Space activation reports on a button.
    triggerButton().dispatchEvent(
      new MouseEvent('click', { detail: 0, bubbles: true, cancelable: true }),
    )
    await flushPromises()

    const items = [...document.querySelectorAll('.reaction-item')]
    expect(items).toHaveLength(7)
    expect(document.activeElement).toBe(items[0])
  })

  it('closes on the first tap even when the opening tap produced no click', async () => {
    mountReactions()
    const trigger = triggerButton()

    trigger.dispatchEvent(new PointerEvent('pointerdown', { button: 0, bubbles: true }))
    await flushPromises()
    expect(trigger.getAttribute('aria-expanded')).toBe('true')

    // Home-screen WebKit drops the click that trails a touch gesture. A guard that only
    // that click clears would stay set and swallow the next tap.
    trigger.dispatchEvent(new PointerEvent('pointerdown', { button: 0, bubbles: true }))
    await flushPromises()

    expect(trigger.getAttribute('aria-expanded')).toBe('false')
  })

  it('lets a keyboard click close the picker after a lost touch click', async () => {
    mountReactions()
    const trigger = triggerButton()

    trigger.dispatchEvent(new PointerEvent('pointerdown', { button: 0, bubbles: true }))
    await flushPromises()

    trigger.dispatchEvent(new MouseEvent('click', { detail: 0, bubbles: true, cancelable: true }))
    await flushPromises()

    expect(trigger.getAttribute('aria-expanded')).toBe('false')
  })
})
