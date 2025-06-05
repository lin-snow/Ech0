<template>
  <div class="mx-auto px-2 sm:px-4 md:px-6 my-4 sm:my-5 md:my-6">
    <!-- Echos -->
    <div>
      <div v-for="echo in echoStore.echoList" :key="echo.id">
        <TheEchoCard :echo="echo" @refresh="handleRefresh" />
      </div>
    </div>
    <!-- 加载更多 -->
    <div v-if="echoStore.hasMore" class="mb-4 mt-1 -ml-2">
      <BaseButton @click="handleLoadMore" class="rounded-full bg-white !hover:bg-gray-100">
        <span class="text-gray-600 text-md text-center px-2 py-1">继续装填</span>
      </BaseButton>
    </div>
    <!-- 没有更多 -->
    <div v-else class="mx-auto my-5 text-center">
      <span class="text-xl text-gray-400">没有啦！🎉</span>
    </div>
    <!-- 备案号 -->
    <div class="text-center">
      <a href="https://beian.miit.gov.cn/" target="_blank">
        <span class="text-gray-400 text-sm">
          {{ SystemSetting.ICP_number }}
        </span>
      </a>
    </div>
  </div>
</template>

<script setup lang="ts">
import TheEchoCard from '@/components/advanced/TheEchoCard.vue'
import { onMounted } from 'vue'
import { useEchoStore } from '@/stores/echo'
import { useSettingStore } from '@/stores/settting'
import BaseButton from '@/components/common/BaseButton.vue'
import { storeToRefs } from 'pinia'

const echoStore = useEchoStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const handleLoadMore = async () => {
  echoStore.current = echoStore.current + 1
  await echoStore.getEchosByPage()
}

// 刷新数据
const handleRefresh = () => {
  echoStore.refreshEchos()
}

onMounted(async () => {
  // 获取数据
  await echoStore.getEchosByPage()
})
</script>
