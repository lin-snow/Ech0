<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="listEl" class="echo-timeline-list" :aria-busy="busy">
    <div
      v-for="(item, index) in items"
      :key="item.id"
      :data-echo-id="item.id"
      class="echo-timeline-row"
    >
      <TheEchoCard :echo="item" :index="index" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import {
  HOME_TIMELINE_ANCHOR_KEY,
  parseTimelineAnchor,
  serializeTimelineAnchor,
} from '@/utils/timeline'

const props = defineProps<{
  items: App.Api.Ech0.Echo[]
  scrollTarget: HTMLElement | null
  useWindow: boolean
  busy?: boolean
  /** Filter signature of the loaded stream; anchors from another query are discarded. */
  signature?: string | null
}>()

const listEl = ref<HTMLElement | null>(null)

// Top edge of a row in the coordinates of the scroller that owns it: document
// space for the window, scroller content space for the inner column.
const rowTop = (el: HTMLElement): number => {
  if (props.useWindow) return el.getBoundingClientRect().top + window.scrollY
  const root = props.scrollTarget
  if (!root) return el.offsetTop
  return el.getBoundingClientRect().top - root.getBoundingClientRect().top + root.scrollTop
}

const rowElements = (): HTMLElement[] =>
  Array.from(listEl.value?.querySelectorAll<HTMLElement>('[data-echo-id]') ?? [])

// ---------------------------------------------------------------------------
// Refresh restore
//
// One anchor (first visible Echo + how far it is scrolled past the top) is
// kept in sessionStorage and restored as a single offset write once the page
// it belongs to is rendered. The page index itself lives in the store.
// ---------------------------------------------------------------------------

const initialAnchor = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))

// The startup loader keeps <html> at viewport height with `overflow: hidden`
// until the first paint settles, so on the mobile layout the page reports a
// zero max offset while the feed is already mounted. The browser clamps every
// scroll command to that stale maximum, which would turn the restore into a
// no-op at the top. Wait for the release instead of losing the anchor.
const RESTORE_RETRY_MS = 50
const RESTORE_RETRY_LIMIT = 40
let restoreTimer: number | undefined

const scrollerCanScroll = (): boolean => {
  if (props.useWindow) {
    const scroller = document.scrollingElement
    return !!scroller && scroller.scrollHeight > window.innerHeight + 1
  }
  const root = props.scrollTarget
  return !!root && root.scrollHeight > root.clientHeight + 1
}

const restoreAnchor = (attempt = 0) => {
  const anchor = initialAnchor
  if (!anchor) return
  if (anchor.signature === undefined || anchor.signature !== (props.signature ?? '')) {
    // The stored position belongs to another filter set; restoring it would
    // resume the new stream at an arbitrary row.
    sessionStorage.removeItem(HOME_TIMELINE_ANCHOR_KEY)
    return
  }
  const row = rowElements().find((el) => el.dataset.echoId === anchor.id)
  if (!row) {
    // The anchored Echo is not on this page (or was deleted): start at the top.
    sessionStorage.removeItem(HOME_TIMELINE_ANCHOR_KEY)
    return
  }
  if (props.useWindow && !scrollerCanScroll() && attempt < RESTORE_RETRY_LIMIT) {
    restoreTimer = window.setTimeout(() => {
      // A user scroll while we waited rewrites or drops the anchor; that
      // position is the newer truth, so the retry must not drag it back.
      const stored = parseTimelineAnchor(sessionStorage.getItem(HOME_TIMELINE_ANCHOR_KEY))
      if (stored?.id !== anchor.id || stored.offset !== anchor.offset) return
      restoreAnchor(attempt + 1)
    }, RESTORE_RETRY_MS)
    return
  }
  const target = rowTop(row) + anchor.offset
  if (props.useWindow) window.scrollTo({ top: target })
  else props.scrollTarget?.scrollTo({ top: target })
}

// ---------------------------------------------------------------------------
// Anchor persistence (throttled to at most one write per 150 ms)
// ---------------------------------------------------------------------------

let anchorTimer: number | undefined

// A scroll-mode switch collapses the scroller this instance reads (the page
// becomes fixed-height, or the old container loses its content). Its zero
// offset is a layout artifact then, not a position the user chose.
const scrollerCollapsed = (): boolean => {
  if (props.useWindow) {
    return document.documentElement.scrollHeight <= window.innerHeight + 1
  }
  const root = props.scrollTarget
  return !root || root.scrollHeight <= root.clientHeight + 1
}

const persistAnchor = () => {
  const rows = rowElements()
  if (!rows.length) return
  const scroll = props.useWindow ? window.scrollY : (props.scrollTarget?.scrollTop ?? 0)
  // The viewport sits above the list (header, banner, or the very top): there
  // is no Echo to anchor. Drop the anchor so a refresh starts where the user
  // left off instead of jumping back down to an old Echo; a scroller that is
  // momentarily collapsed keeps it for the restore that follows the remount.
  if (scroll < rowTop(rows[0]!)) {
    if (!scrollerCollapsed()) sessionStorage.removeItem(HOME_TIMELINE_ANCHOR_KEY)
    return
  }
  const row = rows.find((el) => rowTop(el) + el.offsetHeight > scroll) ?? rows[rows.length - 1]!
  const id = row.dataset.echoId
  if (!id) return
  sessionStorage.setItem(
    HOME_TIMELINE_ANCHOR_KEY,
    serializeTimelineAnchor({
      id,
      offset: Math.max(0, Math.round(scroll - rowTop(row))),
      signature: props.signature ?? undefined,
    }),
  )
}

const schedulePersist = () => {
  if (anchorTimer !== undefined) return
  anchorTimer = window.setTimeout(() => {
    anchorTimer = undefined
    persistAnchor()
  }, 150)
}

const flushPersist = () => {
  window.clearTimeout(anchorTimer)
  anchorTimer = undefined
  persistAnchor()
}

let scrollBoundTo: HTMLElement | Window | null = null
const bindScroll = () => {
  scrollBoundTo?.removeEventListener('scroll', schedulePersist)
  scrollBoundTo = props.useWindow ? window : props.scrollTarget
  scrollBoundTo?.addEventListener('scroll', schedulePersist, { passive: true })
}

onMounted(async () => {
  bindScroll()
  window.addEventListener('pagehide', flushPersist)
  await nextTick()
  restoreAnchor()
})

onBeforeUnmount(() => {
  flushPersist()
  scrollBoundTo?.removeEventListener('scroll', schedulePersist)
  window.removeEventListener('pagehide', flushPersist)
  window.clearTimeout(restoreTimer)
})
</script>

<style scoped>
.echo-timeline-list {
  position: relative;
  width: 100%;
}

.echo-timeline-row {
  width: 100%;
}
</style>
