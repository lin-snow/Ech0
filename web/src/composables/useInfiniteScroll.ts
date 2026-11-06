// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import type { Ref } from 'vue'

type InfiniteScrollOptions = {
  /** Element used as the load trigger; it must sit after the rendered items. */
  sentinel: Ref<HTMLElement | null>
  /** Timeline scroll container; only used when it actually scrolls. */
  scrollTarget: () => HTMLElement | null
  /** Return false while a request is in flight, the feed is exhausted, or an error is shown. */
  canLoad: () => boolean
  /** Starts the next page; returning a promise keeps the trigger single-flight. */
  load: () => void | Promise<void>
  /** Re-check after items are appended: the observer does not fire again while the sentinel stays visible. */
  items: Ref<readonly unknown[]>
  rootMargin?: number
}

const SCROLLABLE_OVERFLOW = ['auto', 'scroll', 'overlay']

/**
 * Sentinel-driven pagination that works with either the window (mobile) or an
 * internal scroll container (desktop). It is deliberately stateless about the
 * feed: all pagination state stays in the store.
 */
export function useInfiniteScroll(options: InfiniteScrollOptions) {
  const rootMargin = options.rootMargin ?? 600
  let observer: IntersectionObserver | null = null
  let root: HTMLElement | null = null
  let inFlight = false
  let scheduled = false

  const activeRoot = (): HTMLElement | null => {
    const target = options.scrollTarget()
    if (!target) return null
    const overflowY = window.getComputedStyle(target).overflowY
    return SCROLLABLE_OVERFLOW.includes(overflowY) ? target : null
  }

  const trigger = async () => {
    if (inFlight || !options.canLoad()) return
    inFlight = true
    try {
      await options.load()
    } finally {
      inFlight = false
      scheduleCheck()
    }
  }

  const isNearViewport = (): boolean => {
    const el = options.sentinel.value
    if (!el) return false
    const rect = el.getBoundingClientRect()
    if (root) {
      return rect.top <= root.getBoundingClientRect().bottom + rootMargin
    }
    return rect.top <= window.innerHeight + rootMargin
  }

  // IntersectionObserver is enough once the sentinel is off-screen, but a
  // short feed keeps it visible and never fires again. This explicit check
  // fills the viewport and survives appended pages.
  const scheduleCheck = () => {
    if (scheduled) return
    scheduled = true
    window.requestAnimationFrame(() => {
      scheduled = false
      if (isNearViewport()) {
        void trigger()
      }
    })
  }

  const teardown = () => {
    observer?.disconnect()
    observer = null
  }

  const setup = () => {
    teardown()
    root = activeRoot()
    const el = options.sentinel.value
    if (!el || typeof IntersectionObserver === 'undefined') return

    observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          void trigger()
        }
      },
      { root, rootMargin: `0px 0px ${rootMargin}px 0px` },
    )
    observer.observe(el)
    scheduleCheck()
  }

  onMounted(() => {
    nextTick(setup)
  })

  watch(
    () => options.scrollTarget(),
    () => nextTick(setup),
    { flush: 'post' },
  )

  watch(
    () => options.sentinel.value,
    () => nextTick(setup),
    { flush: 'post' },
  )

  watch(
    () => options.items.value.length,
    () => nextTick(scheduleCheck),
    { flush: 'post' },
  )

  watch(
    () => options.canLoad(),
    (ready) => {
      if (ready) nextTick(scheduleCheck)
    },
  )

  let resizeTimer: number | null = null
  const onResize = () => {
    if (resizeTimer !== null) window.clearTimeout(resizeTimer)
    resizeTimer = window.setTimeout(() => {
      resizeTimer = null
      // The desktop root stops scrolling below the layout breakpoint; keeping
      // the stale root would make the sentinel permanently intersect.
      setup()
    }, 150)
  }
  window.addEventListener('resize', onResize, { passive: true })

  onBeforeUnmount(() => {
    window.removeEventListener('resize', onResize)
    if (resizeTimer !== null) {
      window.clearTimeout(resizeTimer)
      resizeTimer = null
    }
    teardown()
  })

  return { check: scheduleCheck }
}
