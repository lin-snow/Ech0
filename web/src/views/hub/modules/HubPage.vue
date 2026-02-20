<template>
  <div class="w-full px-2 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <!-- Ech0s Hub -->
    <div class="mx-auto px-2 text-[var(--text-color-next-200)] w-full">
      <h1
        class="text-4xl md:text-6xl italic font-bold font-serif text-center text-[var(--text-color-next-300)]"
      >
        Ech0 Hub
      </h1>

      <div class="w-full max-w-sm mx-auto">
        <!-- 返回首页 -->
        <BaseButton @click="router.push('/')" :class="getButtonClasses('', true)" title="返回首页">
          <Arrow
            class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
          />
        </BaseButton>
      </div>

      <div v-if="echoList.length > 0 && !isPreparing" class="space-y-6">
        <div v-for="echo in echoList" :key="echo.id" class="flex justify-center items-center">
          <TheHubEcho :echo="echo" class="hover:shadow-md" />
        </div>
      </div>

      <div v-if="isLoading || isPreparing" class="my-6">
        <p class="text-[var(--text-color-500)] text-center">加载中...</p>
      </div>
      <div
        v-else-if="echoList.length === 0 && hasTriedInitialLoad && !isPreparing && !isLoading"
        class="my-6"
      >
        <p class="text-[var(--text-color-500)] text-center">暂无数据，快去添加Connect吧🙃</p>
      </div>

      <div v-if="echoList.length > 0 && !hasMore" class="my-6">
        <p class="text-[var(--text-color-500)] text-center flex items-center justify-center">
          没有更多数据了<Flowers />
        </p>
      </div>
    </div>

    <div
      v-show="showBackTop"
      :style="backTopStyle"
      class="hidden xl:block fixed bottom-6 z-50 transition-all duration-500 animate-fade-in"
    >
      <TheBackTop class="w-8 h-8 p-1" />
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import TheHubEcho from '@/components/advanced/TheHubEcho.vue'
import Flowers from '@/components/icons/flowers.vue'
import { onMounted, watch, computed, ref, onBeforeUnmount, nextTick } from 'vue'
import { useHubStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useRouter, useRoute } from 'vue-router'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'

const router = useRouter()
const route = useRoute()

const currentRoute = computed(() => route.name as string)

// 统一的按钮样式计算函数
const getButtonClasses = (routeName: string, isBackButton = false) => {
  const baseClasses = isBackButton
    ? 'text-[var(--text-color-next-600)] rounded-md transition-all duration-300 border-none !shadow-none !ring-0 hover:opacity-75 p-2 group bg-transparent'
    : 'flex items-center gap-2 pl-3 py-1 rounded-md transition-all duration-300 border-none !shadow-none !ring-0 justify-start bg-transparent'

  const activeClasses =
    currentRoute.value === routeName
      ? 'text-stone-800 bg-orange-200'
      : 'text-[var(--text-color-next-600)] hover:opacity-75'

  return `${baseClasses} ${activeClasses}`
}

const hubStore = useHubStore()
const { echoList, isLoading, isPreparing, hasMore, hasTriedInitialLoad } = storeToRefs(hubStore)

const mainColumn = ref<HTMLElement | null>(null)
const backTopStyle = ref({ right: '100px' }) // 默认 fallback
const showBackTop = ref(true) // 自定义条件

// 监听窗口滚动事件，判断是否显示回到顶部按钮
const updateShowBackTop = () => {
  showBackTop.value = window.scrollY > 300
}
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

const { runWithBfCacheGuard } = useBfCacheRestore({
  onRestore: () => {
    schedulePositionUpdate()
  },
})

// --- 滚动到底部检测 ---
let ticking = false
const onScroll = () => {
  if (ticking) return
  ticking = true
  requestAnimationFrame(() => {
    updateShowBackTop()

    if (isLoading.value || !hasMore.value) {
      ticking = false
      return
    }

    const scrollPosition = window.scrollY + window.innerHeight
    const fullHeight = document.documentElement.scrollHeight
    const threshold = 300

    if (scrollPosition + threshold >= fullHeight) {
      hubStore.loadEchoListPage()
    }

    ticking = false
  })
}

// --- 自动加载补全 ---
const ensureScrollable = async () => {
  await nextTick()
  const fullHeight = document.documentElement.scrollHeight
  const windowHeight = window.innerHeight

  // 如果内容高度太短，继续加载直到可滚动或无更多数据
  if (fullHeight <= windowHeight + 10 && hasMore.value && !isLoading.value) {
    await hubStore.loadEchoListPage()
    ensureScrollable()
  }
}

onMounted(async () => {
  // 监听窗口大小变化
  updateShowBackTop()
  schedulePositionUpdate()
  window.addEventListener('scroll', onScroll)
  window.addEventListener('resize', schedulePositionUpdate)

  // 获取 Hub 数据
  await hubStore.getHubList()
  await hubStore.getHubInfoList()
  hubStore.loadEchoListPage()

  // 自动填充内容不足的情况
  ensureScrollable()
})

// 当 echoList 变化时，自动检测是否需要补充加载
watch(echoList, () => {
  ensureScrollable()
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', onScroll)
  window.removeEventListener('resize', schedulePositionUpdate)
})
</script>
