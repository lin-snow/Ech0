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

  // A short feed keeps the sentinel visible, so the observer never fires a
  // second time; re-check explicitly after each page and while it fills.
  const check = () => {
    if (disposed) return
    cancelAnimationFrame(frame)
    frame = window.requestAnimationFrame(() => {
      if (disposed || busy || !nearViewport() || !options.canLoad()) return
      busy = true
      Promise.resolve(options.load())
        .catch(() => {})
        .finally(() => {
          busy = false
          if (!disposed) nextTick(check)
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
  })

  watch(
    () => options.scrollTarget(),
    () => nextTick(setup),
    { flush: 'post' },
  )
  watch(options.sentinel, () => nextTick(setup), { flush: 'post' })
  watch(
    () => options.items.value.length,
    () => nextTick(check),
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
    if (resizeTimer !== null) window.clearTimeout(resizeTimer)
  })
}
