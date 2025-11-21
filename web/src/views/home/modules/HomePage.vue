<template>
  <div
    class="max-w-sm sm:max-w-full px-2 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex flex-col sm:flex-row justify-center items-start sm:gap-8"
  >
    <div class="sm:max-w-sm w-full">
      <TheTop class="sm:hidden" />
      <TheEditor v-if="isLogin" />
      <TheBoard v-else />
    </div>
    <div ref="mainColumn" class="sm:max-w-lg w-full sm:mt-1">
      <TheTop class="hidden sm:block sm:px-4" />
      <TheEchos v-if="!todoMode && !isFilteringMode" />
      <TheFilteredEchos v-else-if="!todoMode && isFilteringMode" />
      <TheTodos v-else />
    </div>
    <div class="hidden xl:block sm:max-w-sm w-full px-6 h-screen">
      <TheHeatMap class="mb-2" />
      <TheStatusCard v-if="isLogin" class="mb-2" />
      <div v-if="isLogin" class="mb-2 px-11">
        <TheTodoCard :todo="todos[0]" :index="0" :operative="false" @refresh="getTodos" />
      </div>
      <div class="mb-2">
        <TheAudioCard />
      </div>
      <TheConnects />
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
import TheTop from './TheTop.vue'
import TheEditor from './TheEditor.vue'
import TheBoard from './TheBoard.vue'
import TheEchos from './TheEchos.vue'
import TheFilteredEchos from './TheFilteredEchos.vue'
import TheTodos from './TheTodos.vue'
import TheConnects from '@/views/connect/modules/TheConnects.vue'
import TheTodoCard from '@/components/advanced/TheTodoCard.vue'
import TheStatusCard from '@/components/advanced/TheStatusCard.vue'
import TheHeatMap from '@/components/advanced/TheHeatMap.vue'
import TheBackTop from '@/components/advanced/TheBackTop.vue'
import { onMounted, ref, onBeforeUnmount } from 'vue'
import { useUserStore } from '@/stores/user'
import { useTodoStore } from '@/stores/todo'
import { useEchoStore } from '@/stores/echo'
import { useEditorStore } from '@/stores/editor'
import { storeToRefs } from 'pinia'
import TheAudioCard from '@/components/advanced/TheAudioCard.vue'

const todoStore = useTodoStore()
const userStore = useUserStore()
const echoStore = useEchoStore()
const editorStore = useEditorStore()
const { getTodos } = todoStore
const { todoMode, todos } = storeToRefs(todoStore)
const { isLogin } = storeToRefs(userStore)
const { isFilteringMode } = storeToRefs(echoStore)

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

onMounted(async () => {
  // 监听窗口大小变化
  updateShowBackTop()
  updatePosition()
  window.addEventListener('scroll', updateShowBackTop)
  window.addEventListener('resize', updatePosition)
})

onBeforeUnmount(() => {
  window.removeEventListener('scroll', updateShowBackTop)
  window.removeEventListener('resize', updatePosition)
})
</script>
