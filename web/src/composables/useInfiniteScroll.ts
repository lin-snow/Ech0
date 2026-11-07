// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import type { Ref } from 'vue'

type InfiniteScrollOptions = {
  /** Element used as the load trigger; it must sit after the rendered items. */
  sentinel: Ref<HTMLElement | null>
  /** Timeline scroll container; only used when it actually scrolls. */
  scrollTarget: () => HTMLElement | null
  /** False while a request is in flight, the feed is exhausted, or an error is shown. */
  canLoad: () => boolean
  /** Starts the next page; the promise keeps the trigger single-flight. */
  load: () => void | Promise<void>
  /** Re-check after items are appended: the observer does not fire again while the sentinel stays visible. */
  items: Ref<readonly unknown[]>
  rootMargin?: number
}

/**
 * Sentinel-driven pagination that works with either the window or an internal
 * scroll container. All feed state stays in the store; this only decides when
 * to ask for the next page.
 */
export function useInfiniteScroll(options: InfiniteScrollOptions) {
  const rootMargin = options.rootMargin ?? 600
  let observer: IntersectionObserver | null = null
  let root: HTMLElement | null = null
  let busy = false
  let frame = 0
  let resizeTimer: number | null = null
  let disposed = false
  // Set when a page appended no rows; the automatic chain then waits for a
  // user scroll instead of requesting the same position in a loop.
  let stalled = false

  const scrollRoot = (): HTMLElement | null => {
    const target = options.scrollTarget()
    if (!target) return null
    const overflow = getComputedStyle(target).overflowY
    return overflow === 'auto' || overflow === 'scroll' || overflow === 'overlay' ? target : null
  }

  const nearViewport = (): boolean => {
    const el = options.sentinel.value
    if (!el) return false
    const bottom = root ? root.getBoundingClientRect().bottom : window.innerHeight
    return el.getBoundingClientRect().top <= bottom + rootMargin
  }

  // A page that appended nothing keeps the sentinel in place, so the next
  // check would request the same position again, forever: a page of already
  // known rows (dedupe) or a cursor that did not advance resolves in the same
  // layout. `stalled` parks the automatic chain until the user scrolls.
  const onUserScroll = () => {
    stalled = false
  }

  // A short feed keeps the sentinel visible, so the observer never fires a
  // second time; re-check explicitly after each page and while it fills.
  const check = () => {
    if (disposed) return
    cancelAnimationFrame(frame)
    frame = window.requestAnimationFrame(() => {
      if (disposed || busy || stalled) return
      if (scrollRoot() !== root) {
        // The breakpoint moved scrolling to/from the window since the observer
        // was built. The stale root would report the sentinel as permanently
        // near and chain pages without end, so rebuild before deciding.
        setup()
        return
      }
      if (!nearViewport() || !options.canLoad()) return
      busy = true
      const before = options.items.value.length
      Promise.resolve(options.load())
        .catch(() => {})
        .finally(() => {
          busy = false
          if (disposed) return
          if (options.items.value.length <= before) {
            stalled = true
            return
          }
          nextTick(check)
        })
    })
  }

  const setup = () => {
    if (disposed) return
    observer?.disconnect()
    root = scrollRoot()
    const el = options.sentinel.value
    if (!el || typeof IntersectionObserver === 'undefined') return

    observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) check()
      },
      { root, rootMargin: `0px 0px ${rootMargin}px 0px` },
    )
    observer.observe(el)
    check()
  }

  // The root changes when the layout breakpoint moves scrolling to/from the
  // window, or when the first page makes the container scrollable.
  const onResize = () => {
    if (resizeTimer !== null) window.clearTimeout(resizeTimer)
    resizeTimer = window.setTimeout(() => {
      resizeTimer = null
      setup()
    }, 150)
  }

  onMounted(() => {
    nextTick(setup)
    window.addEventListener('resize', onResize, { passive: true })
    // Scroll events do not bubble, but a capturing window listener sees both
    // the page and the inner timeline column.
    window.addEventListener('scroll', onUserScroll, { passive: true, capture: true })
  })

  watch(
    () => options.scrollTarget(),
    () => nextTick(setup),
    { flush: 'post' },
  )
  watch(options.sentinel, () => nextTick(setup), { flush: 'post' })
  watch(
    () => options.items.value.length,
    () => {
      // Rows arrived, so the feed can make progress again (a retry after a
      // failed page also lands here).
      stalled = false
      nextTick(check)
    },
    { flush: 'post' },
  )
  watch(
    () => options.canLoad(),
    (ready) => {
      if (ready) check()
    },
  )

  onBeforeUnmount(() => {
    disposed = true
    observer?.disconnect()
    cancelAnimationFrame(frame)
    window.removeEventListener('resize', onResize)
    window.removeEventListener('scroll', onUserScroll, { capture: true })
    if (resizeTimer !== null) window.clearTimeout(resizeTimer)
  })
}
