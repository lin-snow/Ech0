<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
    :class="compact ? 'pl-1 pr-0 max-w-full' : 'px-2 sm:px-4 md:px-6 max-w-full'"
  >
    <TransitionGroup
      v-if="echoStore.echoList"
      name="list"
      tag="div"
      class="relative"
      @before-enter="onBeforeEnter"
      @enter="onEnter"
      @before-leave="onBeforeLeave"
      @leave="onLeave"
    >
      <div v-for="(echo, index) in echoStore.echoList" :key="echo.id">
        <TheEchoCard :echo="echo" :index="index" @refresh="handleRefresh" />
      </div>
    </TransitionGroup>
    <Transition name="fade">
      <div
        v-if="!echoStore.isLoading && echoStore.total > 0 && echoStore.totalPages > 1"
        class="echos-toolbar mb-2 mt-1 -ml-1 flex items-center justify-between"
      >
        <button
          v-if="canGoOlder"
          type="button"
          class="echos-pager echos-pager--older"
          @click="handleGoToPage(echoStore.currentPage + 1)"
        >
          {{ t('homeFeed.older') }}
        </button>
        <span v-else aria-hidden="true" />
        <button
          v-if="canGoNewer"
          type="button"
          class="echos-pager echos-pager--newer"
          @click="handleGoToPage(echoStore.currentPage - 1)"
        >
          {{ t('homeFeed.newer') }}
        </button>
        <span v-else aria-hidden="true" />
      </div>
    </Transition>
    <Transition name="fade">
      <div
        v-if="!echoStore.isLoading && echoStore.total === 0"
        class="mx-auto my-5 text-center echos-toolbar"
      >
        <p class="text-xl text-[var(--color-text-muted)]">
          {{ echoStore.isFilteringMode ? t('homeFeed.noMoreFiltered') : t('homeFeed.noMore') }}
        </p>
      </div>
    </Transition>
    <Transition name="fade">
      <TheLoadingIndicator
        v-if="echoStore.isLoading"
        class="mx-auto my-5 echos-toolbar"
        size="lg"
        :label="t('homeFeed.loading')"
      />
    </Transition>
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
import { computed, onMounted, ref, watch } from 'vue'
import { useEchoStore, useSettingStore } from '@/stores'
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
const { SystemSetting } = storeToRefs(settingStore)
const { isFilteringMode } = storeToRefs(echoStore)
const footerContent = computed(
  () => SystemSetting.value.footer_content || SystemSetting.value.ICP_number,
)
const footerLink = computed(() => SystemSetting.value.footer_link)

const canGoNewer = computed(() => echoStore.currentPage > 1)
const canGoOlder = computed(() => echoStore.currentPage < echoStore.totalPages)

const hasInitialRendered = ref(false)

let enterBatchIndex = 0
let enterBatchResetTimer: number | null = null

const ENTER_DURATION = 420
const ENTER_STAGGER = 100
const ENTER_STAGGER_CAP = 600
const ENTER_EASING = 'cubic-bezier(0.22, 1, 0.36, 1)'

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

const scrollToTop = () => {
  const container = props.scrollTarget
  if (container) {
    container.scrollTo({ top: 0, behavior: 'smooth' })
  } else if (typeof window !== 'undefined') {
    window.scrollTo({ top: 0, behavior: 'smooth' })
  }
}

const parsePageQuery = (raw: unknown): number => {
  const value = Number(Array.isArray(raw) ? raw[0] : raw)
  if (!Number.isFinite(value) || value < 1) return 1
  return Math.floor(value)
}

const handleGoToPage = async (page: number) => {
  const target = Math.max(1, Math.min(page, echoStore.totalPages || page))
  if (target === echoStore.currentPage) return
  await echoStore.goToPage(target)
  scrollToTop()
}

const handleRefresh = () => {
  echoStore.refreshEchos()
}

watch(
  () => echoStore.currentPage,
  (page) => {
    if (parsePageQuery(route.query.page) === page) return
    const nextQuery = { ...route.query }
    if (page > 1) {
      nextQuery.page = String(page)
    } else {
      delete nextQuery.page
    }
    router.replace({ query: nextQuery })
  },
)

watch(
  () => route.query.page,
  async (raw) => {
    const target = parsePageQuery(raw)
    if (target === echoStore.currentPage) return
    await echoStore.goToPage(target)
    scrollToTop()
  },
)

watch(isFilteringMode, () => {
  echoStore.refreshEchos()
})

onMounted(async () => {
  const target = parsePageQuery(route.query.page)
  echoStore.currentPage = target
  await echoStore.fetchCurrentPage()
  window.setTimeout(() => {
    hasInitialRendered.value = true
  }, 500)
})
</script>

<style scoped>
.echos-toolbar {
  font-family: var(--font-family-display);
}

.echos-pager {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 5rem;
  padding: 0.4rem 1.1rem;
  font-family: var(--font-family-display);
  font-size: 0.8125rem;
  font-weight: 500;
  letter-spacing: 0.01em;
  color: var(--color-text-secondary);
  background: transparent;
  border: 1px solid var(--color-border-strong);
  border-radius: 9999px;
  cursor: pointer;
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease,
    transform 0.08s ease;
}

.echos-pager:hover {
  color: var(--color-text-primary);
  border-color: var(--color-text-secondary);
  background: var(--color-border-subtle);
}

.echos-pager:active {
  transform: translateY(1px);
}

.echos-pager:focus-visible {
  outline: 2px solid var(--color-text-secondary);
  outline-offset: 2px;
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
