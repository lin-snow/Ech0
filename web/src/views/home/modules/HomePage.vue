<template>
  <div
    class="max-w-sm sm:max-w-full px-2 pb-4 py-2 mt-4 sm:mt-0 mb-10 sm:mb-0 mx-auto flex flex-col sm:flex-row justify-center items-start sm:items-stretch sm:gap-8 sm:h-[100dvh] sm:overflow-hidden"
  >
    <div class="sm:max-w-sm w-full sm:min-h-0 sm:h-full sm:overflow-y-auto">
      <TheTop class="sm:hidden" />
      <TheEditor v-if="isLogin" />
      <TheBoard v-else />
    </div>
    <div
      ref="mainColumn"
      class="[--echo-date-sticky-top:0px] sm:[--echo-date-sticky-top:56px] sm:max-w-lg w-full sm:mt-1 sm:min-h-0 sm:h-full sm:overflow-y-auto sm:[overscroll-behavior:contain]"
    >
      <div
        class="hidden sm:block sticky top-0 z-20 relative -mx-2 sm:-mx-4 md:-mx-6 px-2 sm:px-4 md:px-6 pt-2 pb-2 bg-[var(--bg-color)]"
      >
        <TheTop class="sm:px-4" />
      </div>
      <TheEchos v-if="!todoMode && !isFilteringMode && !inboxMode" :scroll-target="mainColumn" />
      <TheFilteredEchos
        v-else-if="!todoMode && isFilteringMode && !inboxMode"
        :scroll-target="mainColumn"
      />
      <TheTodos v-else-if="todoMode && !inboxMode" />
      <TheInbox v-else />
    </div>
    <div class="hidden xl:block sm:max-w-sm w-full px-6 sm:min-h-0 sm:h-full sm:overflow-y-auto">
      <TheHeatMap class="mb-2" />
      <TheStatusCard v-if="isLogin" class="mb-2" />
      <div v-if="isLogin" class="mb-2 px-11">
        <TheTodoCard :todo="todos[0]" :index="0" :operative="false" @refresh="getTodos" />
      </div>
      <div class="mb-2">
        <TheAudioCard />
      </div>
      <TheConnects class="mb-2" />
      <TheRecentCard v-if="AgentSetting.enable" />
    </div>
  </div>
</template>

<script setup lang="ts">
import TheTop from './TheTop.vue'
import TheEditor from './TheEditor.vue'
import TheBoard from './TheBoard.vue'
import TheEchos from './TheEchos.vue'
import TheFilteredEchos from './TheFilteredEchos.vue'
import TheTodos from './TheTodos.vue'
import TheInbox from './TheInbox.vue'
import TheConnects from '@/views/connect/modules/TheConnects.vue'
import TheTodoCard from '@/components/advanced/TheTodoCard.vue'
import TheRecentCard from '@/components/advanced/TheRecentCard.vue'
import TheStatusCard from '@/components/advanced/TheStatusCard.vue'
import TheHeatMap from '@/components/advanced/TheHeatMap.vue'
import { onMounted, ref, onBeforeUnmount } from 'vue'
import { useUserStore, useTodoStore, useEchoStore, useSettingStore, useInboxStore } from '@/stores'
import { storeToRefs } from 'pinia'
import TheAudioCard from '@/components/advanced/TheAudioCard.vue'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'

const todoStore = useTodoStore()
const userStore = useUserStore()
const echoStore = useEchoStore()
const settingStore = useSettingStore()
const inboxStore = useInboxStore()
const { getTodos } = todoStore
const { todoMode, todos } = storeToRefs(todoStore)
const { isLogin } = storeToRefs(userStore)
const { isFilteringMode } = storeToRefs(echoStore)
const { AgentSetting } = storeToRefs(settingStore)
const { inboxMode } = storeToRefs(inboxStore)

const mainColumn = ref<HTMLElement | null>(null)
const backTopStyle = ref({ right: '100px' }) // 默认 fallback
const TIMELINE_SCROLL_KEY = 'home:timeline:scrollTop'
let timelineScrollRaf: number | null = null

const updatePosition = () => {
  if (mainColumn.value) {
    const rect = mainColumn.value.getBoundingClientRect()
    const rightOffset = window.innerWidth - rect.right
    backTopStyle.value = {
      right: `${rightOffset - 160}px`,
    }
  }
}

const schedulePositionUpdate = () => {
  runWithBfCacheGuard(updatePosition, 120)
}

const saveTimelineScrollPosition = () => {
  if (!mainColumn.value || timelineScrollRaf !== null) return

  timelineScrollRaf = window.requestAnimationFrame(() => {
    timelineScrollRaf = null
    if (!mainColumn.value) return
    sessionStorage.setItem(TIMELINE_SCROLL_KEY, String(mainColumn.value.scrollTop))
  })
}

const restoreTimelineScrollPosition = () => {
  if (!mainColumn.value) return
  const raw = sessionStorage.getItem(TIMELINE_SCROLL_KEY)
  if (!raw) return
  const scrollTop = Number(raw)
  if (!Number.isFinite(scrollTop) || scrollTop < 0) return
  mainColumn.value.scrollTop = scrollTop
}

const { runWithBfCacheGuard } = useBfCacheRestore({
  onRestore: () => {
    schedulePositionUpdate()
  },
})

onMounted(async () => {
  // 监听窗口大小变化
  schedulePositionUpdate()
  window.addEventListener('resize', schedulePositionUpdate)
  if (mainColumn.value) {
    mainColumn.value.addEventListener('scroll', saveTimelineScrollPosition, { passive: true })
  }
  window.requestAnimationFrame(() => {
    restoreTimelineScrollPosition()
  })
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', schedulePositionUpdate)
  if (mainColumn.value) {
    mainColumn.value.removeEventListener('scroll', saveTimelineScrollPosition)
  }
  if (timelineScrollRaf !== null) {
    window.cancelAnimationFrame(timelineScrollRaf)
    timelineScrollRaf = null
  }
})
</script>
