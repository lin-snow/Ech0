<template>
  <div class="px-9 md:px-11">
    <!-- 列出所有连接（列出每个连接的头像） -->
    <div
      class="widget rounded-md shadow-sm hover:shadow-md ring-1 ring-[var(--ring-color)] ring-inset p-4"
    >
      <h2 class="text-[var(--widget-title-color)] font-bold text-lg mb-2 flex items-center">
        <Connect class="mr-2" />我的连接:
      </h2>
      <div v-if="!loading">
        <div v-if="!connectsInfo.length" class="text-stone-400 text-sm mb-2">当前暂无连接</div>
        <div v-else class="flex flex-wrap gap-3">
          <div
            v-for="(connect, index) in connectsInfo"
            :key="index"
            class="relative flex flex-col items-center justify-center w-8 h-8 min-w-[2rem] min-h-[2rem] flex-none border-2 border-gray-200 shadow-sm rounded-md hover:shadow-md transition duration-200 ease-in-out group"
          >
            <a :href="connect.server_url" target="_blank" class="block w-full h-full">
              <img
                :src="connect.logo || '/favicon.svg'"
                alt="logo"
                class="w-full h-full rounded-md object-cover"
                @error="handleImageError"
              />
              <!-- 热力圆点 -->
              <span
                class="absolute top-0 right-0 w-2.5 h-2.5 border-2 border-white rounded-full"
                :style="{
                  transform: 'translate(35%, -35%)',
                  backgroundColor: getColor(connect.today_echos || 0),
                }"
              ></span>
            </a>
            <!-- Tooltip -->
            <div
              class="absolute z-10 left-1/2 -translate-x-1/2 top-10 min-w-max bg-gray-800 text-white text-xs rounded px-3 py-2 opacity-0 group-hover:opacity-100 pointer-events-none transition-opacity duration-200 shadow-lg"
            >
              <div class="font-bold mb-1">{{ connect.server_name }}</div>
              <div v-if="connect.sys_username">管理员: {{ connect.sys_username }}</div>
              <div v-if="connect.total_echos">共有: {{ connect.total_echos }}</div>
              <div v-if="connect.today_echos">今日: {{ connect.today_echos }}</div>
            </div>
          </div>
        </div>
      </div>
      <div v-else>
        <div class="text-stone-500 text-sm mb-2">加载中...</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Connect from '@/components/icons/connect.vue'
import { useConnectStore } from '@/stores/connect'
import { storeToRefs } from 'pinia'
import { onMounted } from 'vue'

const connectStore = useConnectStore()
const { getConnectInfo } = connectStore
const { loading, connectsInfo } = storeToRefs(connectStore)

const getColor = (count: number): string => {
  if (count >= 4) return '#196127'
  if (count >= 3) return '#239a3b'
  if (count >= 2) return '#7bc96f'
  if (count >= 1) return '#c6e48b'
  return '#b7bbb7' // 默认颜色
}

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = '/favicon.svg'
}

onMounted(() => {
  getConnectInfo()
})
</script>

<style scoped></style>
