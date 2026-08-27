<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="home-page">
    <div class="home-shell">
      <!-- 左栏：品牌 + 主导航 + 快捷操作（≥820px 显示） -->
      <HomeLeftRail class="home-shell__left" />

      <!-- 中栏：主信息流 / 编辑器 / 标签 / 广场 / 状态 -->
      <main class="home-shell__main">
        <!-- 移动端品牌 + 导航：仅 <820px 显示 -->
        <div class="home-shell__mobile-top">
          <HomeHeader class="home-shell__mobile-header" />
          <HomeSidebarNav
            v-model:mobile-search-open="mobileSearchOpen"
            class="home-shell__mobile-nav"
            @open-palette="paletteOpen = true"
            @open-chat="chatLauncherOpen = true"
          />
        </div>

        <div
          v-if="activeTab === 'publish'"
          class="home-content-block home-content-block--publish"
        >
          <TheEditor />
        </div>
        <div v-else-if="activeTab === 'tags'" class="home-content-block">
          <TheTagsManager />
        </div>

        <template v-else-if="activeTab === 'home'">
          <HomeBanner :class="{ 'home-banner--mobile-hidden': shouldHideBannerOnMobile }" />
          <TheEchos compact />
        </template>

        <div v-else-if="activeTab === 'status'" class="home-content-block home-status-widgets">
          <TheHeatMap />
          <TheRecentCard v-if="AgentSetting.enable" />
          <TheConnectWidget />
          <TheCommentWidget />
        </div>

        <HubPage v-else embedded />
      </main>

      <!-- 右栏：搜索 + 常驻 widgets + version（≥1100px 显示） -->
      <HomeRightRail
        class="home-shell__right"
        @open-palette="paletteOpen = true"
        @open-chat="chatLauncherOpen = true"
      />
    </div>
    <TheCommandPalette v-model="paletteOpen" />
    <TheChatLauncher v-model="chatLauncherOpen" />
  </div>
</template>

<script setup lang="ts">
import HomeHeader from './HomeHeader.vue'
import HomeBanner from './HomeBanner.vue'
import HomeSidebarNav from './HomeSidebarNav.vue'
import HomeLeftRail from './HomeLeftRail.vue'
import HomeRightRail from './HomeRightRail.vue'
import TheEchos from './TheEchos.vue'
import TheCommandPalette from './TheCommandPalette.vue'
import TheChatLauncher from './TheChatLauncher.vue'
import { defineAsyncComponent, onMounted, ref, onBeforeUnmount, computed, watch } from 'vue'
import { useEchoStore, useUserStore, useSettingStore } from '@/stores'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import {
  TheCommentWidget,
  TheConnectWidget,
  TheHeatMap,
  TheRecentCard,
} from '@/components/advanced/widget'

const route = useRoute()
const TheEditor = defineAsyncComponent(() => import('./TheEditor.vue'))
const TheTagsManager = defineAsyncComponent(() => import('./TheEditor/TheTagsManager.vue'))
const HubPage = defineAsyncComponent(() => import('@/views/hub/modules/HubPage.vue'))

const userStore = useUserStore()
const settingStore = useSettingStore()
const echoStore = useEchoStore()
const { isLogin } = storeToRefs(userStore)
const { AgentSetting } = storeToRefs(settingStore)
const { searchingMode, isFilteringMode } = storeToRefs(echoStore)
const mobileSearchOpen = ref(false)
const activeTab = computed<'home' | 'publish' | 'status' | 'tags' | 'hub'>(() => {
  if (route.query.tab === 'publish' && isLogin.value) return 'publish'
  if (route.query.tab === 'status') return 'status'
  if (route.query.tab === 'tags') return 'tags'
  if (route.query.tab === 'hub') return 'hub'
  return 'home'
})
const shouldHideBannerOnMobile = computed(
  () => mobileSearchOpen.value || searchingMode.value || isFilteringMode.value,
)

const WINDOW_SCROLL_KEY = 'home:window:scrollTop'
let windowScrollRaf: number | null = null

const paletteOpen = ref<boolean>(false)
const chatLauncherOpen = ref<boolean>(false)
const chatAvailable = computed(() => isLogin.value && AgentSetting.value.enable)

const handleGlobalKeydown = (event: KeyboardEvent) => {
  const withModifier = (event.metaKey || event.ctrlKey) && !event.altKey && !event.shiftKey
  const isSearchShortcut = withModifier && event.key === 'k'
  if (isSearchShortcut) {
    event.preventDefault()
    paletteOpen.value = !paletteOpen.value
    return
  }
  const isChatShortcut = withModifier && event.key === 'j'
  if (isChatShortcut && chatAvailable.value) {
    event.preventDefault()
    chatLauncherOpen.value = !chatLauncherOpen.value
    return
  }
  if (event.key === 'Escape') {
    if (paletteOpen.value) paletteOpen.value = false
    if (chatLauncherOpen.value) chatLauncherOpen.value = false
  }
}

const saveWindowScrollPosition = () => {
  if (windowScrollRaf !== null) return
  windowScrollRaf = window.requestAnimationFrame(() => {
    windowScrollRaf = null
    sessionStorage.setItem(WINDOW_SCROLL_KEY, String(window.scrollY))
  })
}

const restoreWindowScrollPosition = () => {
  const rawWindow = sessionStorage.getItem(WINDOW_SCROLL_KEY)
  const windowTop = rawWindow ? Number(rawWindow) : 0
  if (Number.isFinite(windowTop) && windowTop > 0) {
    window.scrollTo({ top: windowTop, behavior: 'auto' })
  }
}

const prefetchHeavyChunks = () => {
  const trigger = () => {
    import('@/views/echo/EchoView.vue').catch(() => {})
    import('@/editor/core/markdown').catch(() => {})
  }
  const ric = (window as Window & { requestIdleCallback?: typeof requestIdleCallback })
    .requestIdleCallback
  if (typeof ric === 'function') {
    ric(trigger, { timeout: 2000 })
  } else {
    window.setTimeout(trigger, 1500)
  }
}

onMounted(async () => {
  window.addEventListener('scroll', saveWindowScrollPosition, { passive: true })
  // 等首批 echo 渲染后再恢复滚动位置，否则文档高度还没撑开，scrollY 会被夹到 0。
  const stopScrollRestoreWatch = watch(
    () => echoStore.echoList.length > 0 && !echoStore.isLoading,
    (ready) => {
      if (!ready) return
      stopScrollRestoreWatch()
      window.requestAnimationFrame(() => {
        restoreWindowScrollPosition()
      })
    },
    { immediate: true, flush: 'post' },
  )
  window.addEventListener('keydown', handleGlobalKeydown)
  prefetchHeavyChunks()
})

onBeforeUnmount(() => {
  if (windowScrollRaf !== null) {
    window.cancelAnimationFrame(windowScrollRaf)
    windowScrollRaf = null
  }
  window.removeEventListener('scroll', saveWindowScrollPosition)
  window.removeEventListener('keydown', handleGlobalKeydown)
})
</script>

<style scoped>
.home-page {
  --home-canvas: var(--color-bg-canvas, #f5f3ef);
  --home-rail-left: 15rem;
  --home-rail-right: 19rem;
  --home-main-max: 36rem;

  min-height: 100dvh;
  background: var(--home-canvas);
  color: var(--color-text-primary);
}

/* === 移动端 (<820px)：单栏 === */
.home-shell {
  margin: 0 auto;
  padding: 0 0.75rem;
  max-width: 36rem;
}

@media (width >= 640px) {
  .home-shell {
    padding: 0 1rem;
  }
}

.home-shell__left,
.home-shell__right {
  display: none;
}

.home-shell__mobile-top {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 1rem;
  padding-bottom: 0.5rem;
}

.home-shell__main {
  width: 100%;
  min-width: 0;
  padding-bottom: 2rem;
}

/* === 平板 (820-1099px)：左栏 sticky + 中栏自然滚动 === */
@media (width >= 820px) {
  .home-shell {
    display: grid;
    grid-template-columns: var(--home-rail-left) minmax(0, 1fr);
    gap: 0;
    max-width: calc(var(--home-rail-left) + var(--home-main-max));
    padding: 0;
    align-items: start;
  }

  .home-shell__left {
    display: flex;
    position: sticky;
    top: 0;
    height: 100dvh;
    overflow-y: auto;
    overscroll-behavior: contain;
    border-right: 1px solid var(--color-border-subtle);
  }

  /* 中栏跟随整页滚动；不再独立 overflow 容器 */
  .home-shell__main {
    padding: 1.25rem 1.5rem 2rem;
    border-right: 1px solid var(--color-border-subtle);
  }

  /* 桌面布局下隐藏中栏顶部移动端品牌 */
  .home-shell__mobile-top {
    display: none;
  }
}

/* === 桌面 (≥1100px)：完整三栏，右栏自然跟随页面滚动 === */
@media (width >= 1100px) {
  .home-shell {
    grid-template-columns: var(--home-rail-left) minmax(0, 1fr) var(--home-rail-right);
    max-width: calc(var(--home-rail-left) + var(--home-main-max) + var(--home-rail-right));
  }

  .home-shell__main {
    border-right: 1px solid var(--color-border-subtle);
  }

  /* 右栏不再 sticky / 独立滚动；与中栏一起跟随页面滚动 */
  .home-shell__right {
    display: flex;
    flex-direction: column;
  }
}

/* === 主内容容器内边距：移动端紧凑 === */
.home-content-block {
  width: 100%;
}

.home-content-block--publish {
  padding-inline: 0;
}

.home-status-widgets {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

@media (width >= 768px) {
  .home-content-block--publish {
    padding-inline: 1.25rem;
  }
}

@media (width >= 1024px) {
  .home-content-block--publish {
    padding-inline: 1.5rem;
  }
}

@media (width <= 819.98px) {
  .home-banner--mobile-hidden {
    display: none;
  }
}
</style>
