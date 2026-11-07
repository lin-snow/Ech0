// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, ref } from 'vue'
import type { Ref } from 'vue'
import { mount } from '@vue/test-utils'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'

const advanceFrames = () => vi.advanceTimersByTimeAsync(50)

class NoopIntersectionObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
  takeRecords() {
    return []
  }
}

// jsdom has no IntersectionObserver; the composable bails out of setup without
// one, and these tests exercise the post-page re-check rather than the
// observer callback.
const mountFeed = (items: Ref<unknown[]>, load: () => Promise<void>, canLoad: () => boolean) =>
  mount(
    defineComponent({
      setup() {
        const sentinel = ref<HTMLElement | null>(null)
        useInfiniteScroll({
          sentinel,
          scrollTarget: () => null,
          canLoad,
          load,
          items,
        })
        return () => h('div', [h('div', { ref: sentinel })])
      },
    }),
  )

// jsdom reports the sentinel at the top of the window, so every check sees it
// inside the prefetch band and the composable's own chaining decides the rest.
describe('useInfiniteScroll', () => {
  beforeEach(() => {
    vi.useFakeTimers({
      toFake: ['setTimeout', 'clearTimeout', 'requestAnimationFrame', 'cancelAnimationFrame'],
    })
    vi.stubGlobal('IntersectionObserver', NoopIntersectionObserver)
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('stops chaining when a page appends no rows', async () => {
    const items = ref<unknown[]>([])
    const load = vi.fn(async () => {})
    const feed = mountFeed(items, load, () => true)

    await advanceFrames()
    expect(load).toHaveBeenCalledTimes(1)

    // Without the stall guard each completed page re-armed the check and
    // requested the same cursor position again, forever.
    await advanceFrames()
    await advanceFrames()
    expect(load).toHaveBeenCalledTimes(1)

    feed.unmount()
  })

  it('keeps filling while pages make progress and stops at the last page', async () => {
    const items = ref<unknown[]>([])
    const load = vi.fn(async () => {
      items.value.push(items.value.length)
    })
    const feed = mountFeed(items, load, () => items.value.length < 3)

    await advanceFrames()
    expect(load).toHaveBeenCalledTimes(3)
    expect(items.value).toHaveLength(3)

    feed.unmount()
  })
})
