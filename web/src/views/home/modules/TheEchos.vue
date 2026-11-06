<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
    :class="compact ? 'pl-1 pr-0 max-w-full' : 'px-2 sm:px-4 md:px-6 max-w-full'"
  >
    <TheOnboardingChecklist v-if="showChecklist" />
    <TheFirstEchoCard v-if="showFirstEchoCard" @dismiss="dismissFirstEchoCard" />
    <TransitionGroup
      v-if="echoStore.echoList"
      name="list"
      tag="div"
      class="relative"
      :aria-busy="echoStore.isLoadingMore"
      @before-enter="onBeforeEnter"
      @enter="onEnter"
      @before-leave="onBeforeLeave"
      @leave="onLeave"
    >
      <div v-for="(echo, index) in echoStore.echoList" :key="echo.id">
        <TheEchoCard :echo="echo" :index="index" />
      </div>
    </TransitionGroup>

    <Transition name="fade">
      <div v-if="echoStore.isLoading && echoStore.echoList.length === 0" class="my-5">
        <TheLoadingIndicator class="echos-toolbar" size="lg" :label="t('homeFeed.loading')" />
      </div>
    </Transition>

    <Transition name="fade">
      <div v-if="showEmptyState" class="mx-auto my-5 text-center echos-toolbar">
        <p class="text-xl text-[var(--color-text-muted)]">
          {{ echoStore.isFilteringMode ? t('homeFeed.noMoreFiltered') : t('homeFeed.noMore') }}
        </p>
      </div>
    </Transition>

    <Transition name="fade">
      <div v-if="echoStore.loadError === 'first'" class="echo-feed-status">
        <p>{{ t('homeFeed.loadMoreFailed') }}</p>
        <button type="button" class="echo-feed-retry" @click="echoStore.retryLoad()">
          {{ t('homeFeed.retry') }}
        </button>
      </div>
    </Transition>

    <Transition name="fade">
      <TheLoadingIndicator
        v-if="echoStore.isLoadingMore"
        class="mx-auto my-5 echos-toolbar"
        size="md"
        :label="t('homeFeed.loadingMore')"
      />
    </Transition>

    <div v-if="echoStore.loadError === 'more'" class="echo-feed-status">
      <p>{{ t('homeFeed.loadMoreFailed') }}</p>
      <button type="button" class="echo-feed-retry" @click="echoStore.retryLoad()">
        {{ t('homeFeed.retry') }}
      </button>
    </div>

    <p
      v-if="
        !echoStore.isLoadingMore &&
        echoStore.loadError === null &&
        echoStore.echoList.length > 0 &&
        echoStore.nextCursor === null
      "
      class="echo-feed-end echos-toolbar"
    >
      {{ t('homeFeed.allLoaded') }}
    </p>

    <div ref="sentinel" class="h-px" aria-hidden="true" />
    <p class="feed-sr-only" aria-live="polite">{{ liveMessage }}</p>

    <div v-if="footerContent" class="mt-6 text-center">
      <a v-if="footerLink" :href="footerLink" target="_blank" rel="noopener noreferrer">
        <span class="text-[var(--color-text-muted)] text-sm">
          {{ footerContent }}
        </span>
      </a>
      <span v-else class="text-[var(--color-text-muted)] text-sm">
        {{ footerContent }}
      </span>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/echo/cards/TheEchoCard.vue'
import TheFirstEchoCard from './TheFirstEchoCard.vue'
import TheOnboardingChecklist from './TheOnboardingChecklist.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useEchoStore, useSettingStore, useUserStore } from '@/stores'
import { FIRST_ECHO_CARD_SEEN_KEY, shouldShowFirstEchoCard } from '@/composables/useSiteOnboarding'
import { useSiteChecklist } from '@/composables/useSiteChecklist'
import { useInfiniteScroll } from '@/composables/useInfiniteScroll'
import { HOME_TIMELINE_DEPTH_KEY } from '@/constants/timeline'
import { localStg } from '@/utils/storage'
import TheLoadingIndicator from '@/components/common/TheLoadingIndicator.vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

const props = defineProps<{
  scrollTarget?: HTMLElement | null
  compact?: boolean
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const route = useRoute()
const router = useRouter()
const { t } = useI18n()
const userStore = useUserStore()
const { SystemSetting } = storeToRefs(settingStore)
const { isFilteringMode, hasActiveQuery } = storeToRefs(echoStore)
const { isLogin } = storeToRefs(userStore)

const firstEchoSeen = ref(localStg.getItem<boolean>(FIRST_ECHO_CARD_SEEN_KEY) === true)

const { visible: showChecklist } = useSiteChecklist()

const showFirstEchoCard = computed(() => {
  const echo = echoStore.echoList?.[0]
  if (!echo) return false
  return shouldShowFirstEchoCard({
    isOwner: isLogin.value,
    filtering: hasActiveQuery.value,
    echoTotal: echoStore.total,
    firstEchoPrivate: echo.private === true,
    dismissed: firstEchoSeen.value,
  })
})

const showEmptyState = computed(
  () =>
    !showChecklist.value &&
    !echoStore.isLoading &&
    echoStore.loadError !== 'first' &&
    echoStore.total === 0,
)

const dismissFirstEchoCard = () => {
  firstEchoSeen.value = true
  localStg.setItem(FIRST_ECHO_CARD_SEEN_KEY, true)
}
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)

const liveMessage = computed(() => {
  if (echoStore.isLoadingMore) return t('homeFeed.loadingMore')
  if (echoStore.loadError === 'more') return t('homeFeed.loadMoreFailed')
  if (echoStore.echoList.length > 0 && echoStore.nextCursor === null) {
    return t('homeFeed.allLoaded')
  }
  return ''
})

const hasInitialRendered = ref(false)

const sentinel = ref<HTMLElement | null>(null)
useInfiniteScroll({
  sentinel,
  scrollTarget: () => props.scrollTarget ?? null,
  canLoad: () => echoStore.canLoadMore,
  load: () => echoStore.loadMore(),
  items: computed(() => echoStore.echoList),
})

const parseLegacyPage = (raw: unknown): number => {
  const value = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isFinite(value) || value < 1) return 1
  return Math.floor(value)
}

const storedDepth = (): number => {
  const value = Number(sessionStorage.getItem(HOME_TIMELINE_DEPTH_KEY))
  return Number.isFinite(value) && value > 1 ? Math.floor(value) : 1
}

watch(
  () => echoStore.echoList.length,
  (length) => {
    const depth = Math.max(1, Math.ceil(length / echoStore.pageSize))
    sessionStorage.setItem(HOME_TIMELINE_DEPTH_KEY, String(depth))
  },
)

watch(isFilteringMode, () => {
  echoStore.refreshEchos()
})

const markInitialRendered = () => {
  window.setTimeout(() => {
    hasInitialRendered.value = true
  }, 500)
}

onMounted(async () => {
  const legacyPage = parseLegacyPage(route.query.page)
  if (route.query.page !== undefined) {
    const nextQuery = { ...route.query }
    delete nextQuery.page
    router.replace({ query: nextQuery })
  }

  // A refresh started elsewhere (publishing returns to this tab) already owns
  // the request; wait for it instead of firing a duplicate.
  if (echoStore.isLoading && echoStore.hasFetched) {
    markInitialRendered()
    return
  }
  if (echoStore.isTimelineCurrent) {
    markInitialRendered()
    return
  }

  const depth = Math.max(storedDepth(), legacyPage)
  await echoStore.refreshEchos(depth > 1 ? depth * echoStore.pageSize : undefined)
  markInitialRendered()
})

const ENTER_DURATION = 420
const ENTER_STAGGER = 100
const ENTER_STAGGER_CAP = 600
const ENTER_EASING = 'cubic-bezier(0.22, 1, 0.36, 1)'

let enterBatchIndex = 0
let enterBatchResetTimer: number | null = null

const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.opacity = '0'
  element.style.transform = 'translateY(-18px)'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  const indexInBatch = enterBatchIndex++
  if (enterBatchResetTimer !== null) {
    window.clearTimeout(enterBatchResetTimer)
  }
  enterBatchResetTimer = window.setTimeout(() => {
    enterBatchIndex = 0
    enterBatchResetTimer = null
  }, 400)

  const baseDelay = hasInitialRendered.value ? 80 : 0
  const staggerDelay = Math.min(indexInBatch * ENTER_STAGGER, ENTER_STAGGER_CAP)

  window.setTimeout(() => {
    element.style.transition = `opacity ${ENTER_DURATION}ms ${ENTER_EASING}, transform ${ENTER_DURATION}ms ${ENTER_EASING}`
    element.style.opacity = '1'
    element.style.transform = 'translateY(0)'
    window.setTimeout(done, ENTER_DURATION)
  }, baseDelay + staggerDelay)
}

const onBeforeLeave = (el: Element) => {
  const element = el as HTMLElement
  const parent = element.parentElement
  if (!parent) return
  const rect = element.getBoundingClientRect()
  const parentRect = parent.getBoundingClientRect()
  element.style.position = 'absolute'
  element.style.left = `${rect.left - parentRect.left}px`
  element.style.top = `${rect.top - parentRect.top}px`
  element.style.width = `${rect.width}px`
}

const onLeave = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  window.requestAnimationFrame(() => {
    element.style.transition = 'opacity 0.18s ease'
    element.style.opacity = '0'
    window.setTimeout(done, 180)
  })
}
</script>

<style scoped>
.echos-toolbar {
  font-family: var(--font-family-display);
}

.echo-feed-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.6rem;
  margin: 1.5rem 0;
  font-family: var(--font-family-display);
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.echo-feed-retry {
  padding: 0.35rem 1.1rem;
  font-family: var(--font-family-display);
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  background: transparent;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  cursor: pointer;
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
}

.echo-feed-retry:hover {
  color: var(--color-text-primary);
  border-color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}

.echo-feed-retry:focus-visible {
  outline: 2px solid var(--color-text-secondary);
  outline-offset: 2px;
}

.echo-feed-end {
  margin: 1.5rem 0 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  text-align: center;
}

.feed-sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}

.list-move {
  transition: transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
