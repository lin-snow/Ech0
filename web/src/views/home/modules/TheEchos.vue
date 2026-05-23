<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="mx-auto mt-1 sm:mt-0 mb-4 sm:mb-5 md:mb-6"
    :class="compact ? 'pl-1 pr-0 max-w-full' : 'px-2 sm:px-4 md:px-6 max-w-full'"
  >
    <!-- Echos - 使用 TransitionGroup 实现入场动画 -->
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
    <!-- 时间线翻页：Before = 更早（左），After = 更新（右）；缺位用空 span 占位以稳定布局 -->
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
    <!-- 没有数据 -->
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
    <!-- 加载中 -->
    <Transition name="fade">
      <TheLoadingIndicator
        v-if="echoStore.isLoading"
        class="mx-auto my-5 echos-toolbar"
        size="lg"
        :label="t('homeFeed.loading')"
      />
    </Transition>
    <!-- 自定义页脚（紧跟时间线内容之后） -->
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
  /** 首页窄栏：减少左右留白以贴合参考图时间线宽度 */
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

// 瀑布式入场：批内计数器决定 stagger 时长，
// 整页换页时新批从 0 重新起跳，不会顶到封顶延迟。
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
  // 起始位置在最终位置上方 18px，配合 stagger 形成"自上而下落入"的视觉。
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

// 离场：先把元素位置快照并切成 absolute 脱离文档流，让留下的卡片
// 通过 .list-move 平滑补位、新进来的卡片不会和旧卡重叠。
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
  const nextQuery = { ...route.query }
  if (target > 1) {
    nextQuery.page = String(target)
  } else {
    delete nextQuery.page
  }
  await router.replace({ query: nextQuery })
  // route watcher 会驱动实际的 fetch + scroll
}

// 刷新数据（点赞 / 编辑 / 删除单条 echo 后刷新当前页）
const handleRefresh = () => {
  echoStore.refreshEchos()
}

// URL ?page=N → store currentPage：单一来源，避免双向同步歪楼
watch(
  () => route.query.page,
  async (raw) => {
    const target = parsePageQuery(raw)
    if (target === echoStore.currentPage && echoStore.echoList.length > 0) return
    echoStore.currentPage = target
    await echoStore.fetchCurrentPage()
    scrollToTop()
  },
)

// 过滤模式切换时（进入/退出/切换标签），刷新列表回到第一页
watch(isFilteringMode, () => {
  if (Number(route.query.page) > 1) {
    const nextQuery = { ...route.query }
    delete nextQuery.page
    router.replace({ query: nextQuery })
    return
  }
  echoStore.refreshEchos()
})

onMounted(async () => {
  const target = parsePageQuery(route.query.page)
  echoStore.currentPage = target
  await echoStore.fetchCurrentPage()
  // 首屏交错动画跑完（最长 ~460ms）后切到淡入模式，避免后续过滤刷新抖动
  window.setTimeout(() => {
    hasInitialRendered.value = true
  }, 500)
})
</script>

<style scoped>
.echos-toolbar {
  font-family: var(--font-family-display);
}

/* 时间线翻页按钮：透明底 + 细边框药丸，贴合纸面色调 */
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

/* 留下的卡片位置变化：用于过滤/刷新时平滑补位 */
.list-move {
  transition: transform 0.24s cubic-bezier(0.2, 0.8, 0.2, 1);
}

/* 离场动画完全在 JS 里处理（onBeforeLeave / onLeave）以便快照位置后脱离文档流。 */

/* 淡入淡出动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
