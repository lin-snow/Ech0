<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="w-full px-2 pb-4 py-2 mt-4 sm:mt-0 mb-10 sm:mb-0 mx-auto flex justify-center items-start"
  >
    <div ref="mainColumn" class="mx-auto px-2 text-[var(--color-text-muted)] w-full">
      <template v-if="!embedded">
        <h1
          class="text-4xl md:text-6xl italic font-bold font-serif text-center text-[var(--color-text-muted)]"
        >
          Ech0 Hub
        </h1>

        <div class="w-full max-w-sm mx-auto">
          <BaseButton
            @click="router.push('/')"
            :class="getButtonClasses('', true)"
            :tooltip="t('commonNav.backHome')"
          >
            <Arrow
              class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
            />
          </BaseButton>
        </div>
      </template>

      <div v-if="echoList.length > 0 && !isPreparing && !props.embedded">
        <DynamicScroller
          class="hub-dynamic-scroller"
          :items="echoList"
          key-field="virtual_key"
          :min-item-size="320"
          :emit-update="true"
          :page-mode="true"
          @update="onScrollerUpdate"
          v-slot="{ item, index, active }"
        >
          <DynamicScrollerItem
            :item="item"
            :active="active"
            :size-dependencies="[
              item.content?.length ?? 0,
              getEchoFilesBy(item, { categories: ['image'], dedupeBy: 'id' }).length,
              item.extension?.type ?? '',
              item.layout ?? '',
            ]"
            :data-index="index"
          >
            <div class="hub-item-wrap flex justify-center items-center py-3">
              <TheHubEcho :echo="item" />
            </div>
          </DynamicScrollerItem>
        </DynamicScroller>
      </div>

      <div v-else-if="echoList.length > 0 && !isPreparing" class="w-full">
        <div
          v-for="item in echoList"
          :key="item.virtual_key"
          class="hub-item-wrap flex justify-center items-center py-3"
        >
          <TheHubEcho :echo="item" />
        </div>
      </div>

      <div v-if="isLoading || isPreparing" class="my-6">
        <TheLoadingIndicator :label="t('hub.loading')" />
      </div>
      <div
        v-else-if="echoList.length === 0 && hasTriedInitialLoad && !isPreparing && !isLoading"
        class="my-6"
      >
        <p class="text-[var(--color-text-secondary)] text-center">
          {{ t('hub.emptyConnectHint') }}
        </p>
      </div>

      <div v-if="echoList.length > 0 && !hasMore" class="my-6">
        <p class="text-[var(--color-text-secondary)] text-center">
          {{ t('hub.noMoreData') }}
        </p>
      </div>

      <div ref="sentinelRef" class="h-1" />
    </div>

    <div
      v-show="showBackTop"
      :style="backTopStyle"
      class="fixed bottom-6 z-50 transition-all duration-500 animate-fade-in"
    >
      <TheBackTop class="w-8 h-8 p-1" :target="props.embedded ? props.scrollTarget : null" />
    </div>
  </div>
</template>

<script setup lang="ts">
import 'vue-virtual-scroller/index.css'

import BaseButton from '@/components/common/BaseButton.vue'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import Arrow from '@/components/icons/arrow.vue'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import TheHubEcho from '@/components/advanced/echo/cards/TheHubEcho.vue'
import { onMounted, watch, computed, ref, onBeforeUnmount, nextTick } from 'vue'
import { useHubStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useRouter, useRoute } from 'vue-router'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'
import { DynamicScroller, DynamicScrollerItem } from 'vue-virtual-scroller'
import { getEchoFilesBy } from '@/utils/echo'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
  defineProps<{
    embedded?: boolean
    scrollTarget?: HTMLElement | null
  }>(),
  {
    embedded: false,
    scrollTarget: null,
  },
)

const router = useRouter()
const route = useRoute()
const { t } = useI18n()

const currentRoute = computed(() => route.name as string)

const getButtonClasses = (routeName: string, isBackButton = false) => {
  const baseClasses = isBackButton
    ? 'text-[var(--color-text-primary)] rounded-md transition-all duration-300 border-none !shadow-none !ring-0 hover:opacity-75 p-2 group bg-transparent'
    : 'flex items-center gap-2 pl-3 py-1 rounded-md transition-all duration-300 border-none !shadow-none !ring-0 justify-start bg-transparent'

  const activeClasses =
    currentRoute.value === routeName
      ? 'text-stone-800 bg-orange-200'
      : 'text-[var(--color-text-primary)] hover:opacity-75'

  return `${baseClasses} ${activeClasses}`
}

const hubStore = useHubStore()
const { echoList, isLoading, isPreparing, hasMore, hasTriedInitialLoad } = storeToRefs(hubStore)

const mainColumn = ref<HTMLElement | null>(null)
const sentinelRef = ref<HTMLElement | null>(null)
const backTopStyle = ref<Record<string, string>>({ right: '100px' })
const showBackTop = ref(false)
const HUB_SCROLL_KEY = 'hub:timeline:scrollTop'
let saveScrollTimer: number | null = null
let sentinelObserver: IntersectionObserver | null = null

const isScrollable = (el: HTMLElement) => {
  const style = window.getComputedStyle(el)
  const ov = style.overflowY
  return ov === 'auto' || ov === 'scroll' || ov === 'overlay'
}

const getActiveScrollElement = () => {
  if (props.embedded && props.scrollTarget && isScrollable(props.scrollTarget)) {
    return props.scrollTarget
  }
  return null
}

const getScrollMetrics = () => {
  const scrollEl = getActiveScrollElement()
  if (scrollEl) {
    return {
      scrollTop: scrollEl.scrollTop,
      viewportHeight: scrollEl.clientHeight,
      fullHeight: scrollEl.scrollHeight,
    }
  }

  const docEl = document.documentElement
  const body = document.body
  return {
    scrollTop: window.scrollY || docEl.scrollTop || 0,
    viewportHeight: window.innerHeight,
    fullHeight: Math.max(docEl.scrollHeight, body.scrollHeight),
  }
}

const updateShowBackTop = () => {
  showBackTop.value = getScrollMetrics().scrollTop > 300
}

const updatePosition = () => {
  const column = mainColumn.value
  if (!column) return
  const rect = column.getBoundingClientRect?.()
  if (!rect) return

  if (props.embedded) {
    const safeLeft = Math.min(window.innerWidth - 56, rect.right + 24)
    backTopStyle.value = {
      left: `${safeLeft}px`,
    }
    return
  }

  const rightOffset = window.innerWidth - rect.right
  const safeRight = Math.max(24, rightOffset - 160)
  backTopStyle.value = {
    right: `${safeRight}px`,
  }
}

const schedulePositionUpdate = () => {
  runWithBfCacheGuard(updatePosition, 120)
}

const { runWithBfCacheGuard } = useBfCacheRestore({
  onRestore: () => {
    schedulePositionUpdate()
  },
})

const onSentinelVisible = () => {
  if (isLoading.value || isPreparing.value || !hasMore.value) return
  hubStore.loadEchoListPage()
}

const setupSentinelObserver = () => {
  teardownSentinelObserver()

  const el = sentinelRef.value
  if (!el) return

  const root = getActiveScrollElement()
  sentinelObserver = new IntersectionObserver(
    (entries) => {
      if (entries.some((e) => e.isIntersecting)) {
        onSentinelVisible()
      }
    },
    { root: root ?? null, rootMargin: '0px 0px 300px 0px' },
  )
  sentinelObserver.observe(el)
}

const teardownSentinelObserver = () => {
  sentinelObserver?.disconnect()
  sentinelObserver = null
}

const onScrollerUpdate = () => {
  updateShowBackTop()
}

let scrollListenerBound = false
const onScrollForBackTop = () => {
  updateShowBackTop()

  if (saveScrollTimer !== null) window.clearTimeout(saveScrollTimer)
  saveScrollTimer = window.setTimeout(() => {
    const { scrollTop } = getScrollMetrics()
    sessionStorage.setItem(HUB_SCROLL_KEY, String(scrollTop))
    saveScrollTimer = null
  }, 120)
}

const bindScrollListenerForBackTop = () => {
  if (scrollListenerBound) return
  const scrollEl = getActiveScrollElement()
  if (scrollEl) {
    scrollEl.addEventListener('scroll', onScrollForBackTop, { passive: true })
  } else {
    window.addEventListener('scroll', onScrollForBackTop, { passive: true })
  }
  scrollListenerBound = true
}

const unbindScrollListenerForBackTop = () => {
  if (!scrollListenerBound) return
  const scrollEl = getActiveScrollElement()
  if (scrollEl) {
    scrollEl.removeEventListener('scroll', onScrollForBackTop)
  } else {
    window.removeEventListener('scroll', onScrollForBackTop)
  }
  scrollListenerBound = false
}

const restoreHubScrollPosition = () => {
  const raw = sessionStorage.getItem(HUB_SCROLL_KEY)
  if (!raw) return
  const scrollTop = Number(raw)
  if (!Number.isFinite(scrollTop) || scrollTop < 0) return
  const scrollEl = getActiveScrollElement()
  if (scrollEl) {
    scrollEl.scrollTop = scrollTop
    return
  }
  window.scrollTo({ top: scrollTop })
}

onMounted(async () => {
  schedulePositionUpdate()
  window.addEventListener('resize', schedulePositionUpdate)

  await hubStore.getHubList()
  await hubStore.getHubInfoList()
  await hubStore.loadEchoListPage()

  restoreHubScrollPosition()
  updateShowBackTop()

  await nextTick()
  setupSentinelObserver()
  bindScrollListenerForBackTop()
})

watch(
  () => props.scrollTarget,
  async () => {
    await nextTick()
    setupSentinelObserver()
    unbindScrollListenerForBackTop()
    bindScrollListenerForBackTop()
  },
)

watch(isLoading, (loading) => {
  if (loading || !hasMore.value) return
  nextTick(() => {
    onSentinelVisible()
  })
})

watch(
  echoList,
  () => {
    nextTick(() => {
      setupSentinelObserver()
    })
  },
  { flush: 'post' },
)

onBeforeUnmount(() => {
  window.removeEventListener('resize', schedulePositionUpdate)
  teardownSentinelObserver()
  unbindScrollListenerForBackTop()
  sessionStorage.setItem(HUB_SCROLL_KEY, String(getScrollMetrics().scrollTop))
  if (saveScrollTimer !== null) {
    window.clearTimeout(saveScrollTimer)
    saveScrollTimer = null
  }
})
</script>

<style scoped>
.hub-dynamic-scroller {
  width: 100%;
}

.hub-item-wrap {
  contain: layout paint;
}
</style>
