<template>
  <div class="mx-auto px-2 sm:px-4 md:px-6 my-4 sm:my-5 md:my-6">
    <!-- Echos - 使用 TransitionGroup 实现入场动画 -->
    <TransitionGroup
      v-if="echoStore.filteredEchoList"
      name="list"
      tag="div"
      @before-enter="onBeforeEnter"
      @enter="onEnter"
    >
      <div
        v-for="(echo, index) in echoStore.filteredEchoList"
        :key="echo.id"
        :data-index="index"
        class="will-change-transform"
      >
        <TheEchoCard
          :echo="echo"
          @refresh="handleRefresh"
          @update-like-count="handleUpdateLikeCount"
        />
      </div>
    </TransitionGroup>
    <!-- 加载更多 -->
    <Transition name="fade">
      <div
        v-if="echoStore.filteredHasMore && !echoStore.isLoading"
        class="mb-4 mt-1 -ml-2 flex items-center justify-between font-serif"
      >
        <BaseButton
          @click="handleLoadMore"
          class="rounded-full bg-[var(--timeline-load-more-bg-color)] !active:bg-[var(--timeline-load-more-active-bg-color)] mr-2"
        >
          <span class="text-[var(--timeline-load-more-text-color)] text-md text-center px-2 py-1"
            >继续装填</span
          >
        </BaseButton>
        <TheBackTop class="w-8 h-8 p-1" :target="scrollTarget" />
      </div>
    </Transition>
    <!-- 没有更多 -->
    <Transition name="fade">
      <div
        v-if="!echoStore.filteredHasMore && !echoStore.isLoading"
        class="mx-auto my-5 text-center"
      >
        <span class="text-xl text-[var(--text-color-400)]">没有啦！🥲</span>
      </div>
    </Transition>
    <!-- 加载中 -->
    <Transition name="fade">
      <div v-if="echoStore.isLoading" class="mx-auto my-5 text-center">
        <span class="text-xl text-[var(--text-color-400)]">加载中...</span>
      </div>
    </Transition>
    <!-- 备案号 -->
    <div class="text-center">
      <a href="https://beian.miit.gov.cn/" target="_blank">
        <span class="text-[var(--text-color-400)] text-sm">
          {{ SystemSetting.ICP_number }}
        </span>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/TheEchoCard.vue'
import { onMounted } from 'vue'
import { useEchoStore, useSettingStore } from '@/stores'
import BaseButton from '@/components/common/BaseButton.vue'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import { storeToRefs } from 'pinia'

defineProps<{
  scrollTarget?: HTMLElement | null
}>()

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

// 列表入场动画钩子 - 交错入场效果
const onBeforeEnter = (el: Element) => {
  const element = el as HTMLElement
  element.style.opacity = '0'
  element.style.transform = 'translateY(20px)'
}

const onEnter = (el: Element, done: () => void) => {
  const element = el as HTMLElement
  const index = Number(element.dataset.index) || 0
  // 交错延迟：每个元素延迟 50ms，最大延迟 250ms
  const delay = Math.min(index * 50, 250)

  setTimeout(() => {
    element.style.transition = 'opacity 0.3s ease, transform 0.3s ease'
    element.style.opacity = '1'
    element.style.transform = 'translateY(0)'

    // 动画结束后调用 done
    setTimeout(done, 300)
  }, delay)
}

const handleLoadMore = async () => {
  echoStore.filteredCurrent = echoStore.filteredCurrent + 1
  await echoStore.getEchosByPageForFilter()
}

// 刷新数据
const handleRefresh = () => {
  echoStore.refreshEchosForFilter()
}

// 刷新点赞数据
const handleUpdateLikeCount = (echoId: number) => {
  echoStore.updateLikeCount(echoId, 1)
}

onMounted(async () => {
  // 获取数据
  echoStore.getEchosByPageForFilter()
})
</script>

<style scoped>
/* 列表项移动动画 */
.list-move {
  transition: transform 0.3s ease;
}

/* 列表项离开动画 */
.list-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.list-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

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
