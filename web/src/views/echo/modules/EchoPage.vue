<template>
  <div class="px-3 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <div class="w-full sm:max-w-lg mx-auto">
      <div class="mx-auto max-w-sm">
        <!-- 返回上一页 -->
        <BaseButton
          @click="goBack"
          class="text-[var(--text-color-600)] rounded-md shadow-none! border-none! ring-0! bg-transparent! group"
          title="返回首页"
        >
          <Arrow
            class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
          />
        </BaseButton>
      </div>

      <div v-if="echo" class="w-full sm:mt-1 mx-auto">
        <TheEchoDetail
          :echo="echo"
          @update-like-count="handleUpdateLikeCount"
          @print-echo="handlePrintEcho"
        />
        <TheComment class="my-2" />
      </div>
      <div v-else class="w-full sm:mt-1 text-[var(--text-color-300)]">
        <p class="text-center">正在加载 Echo 详情...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { fetchGetEchoById } from '@/service/api'
import { ref } from 'vue'
import TheEchoDetail from '@/components/advanced/TheEchoDetail.vue'
import TheComment from '@/components/advanced/TheComment.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Arrow from '@/components/icons/arrow.vue'
import { useEchoStore, useZoneStore } from '@/stores'

const router = useRouter()
const route = useRoute()
const echoId = route.params.echoId as string

const echoStore = useEchoStore()
const zoneStore = useZoneStore()
const isLoading = ref(true)
const echo = ref<App.Api.Ech0.Echo | null>(null)

// 从 echoIndexMap 获取对应的 EchoList索引
const getEchoFromStore = (): App.Api.Ech0.Echo | null => {
  const idx = echoStore.echoIndexMap.get(Number(echoId))
  if (idx !== undefined) {
    return echoStore.echoList[idx] ?? null
  }
  return null
}

// 刷新点赞数据
const handleUpdateLikeCount = () => {
  if (echo.value) {
    // 更新 Echo 的点赞数量
    echo.value.fav_count += 1
  }
}

const handlePrintEcho = (targetEcho: App.Api.Ech0.Echo) => {
  const text = targetEcho.content?.trim() || ''
  if (!text) return

  zoneStore.setPendingPrintEcho(targetEcho)

  router.push({
    name: 'zone',
    params: {
      echoId: String(targetEcho.id),
    },
  })
}

const goBack = () => {
  if (window.history.length > 2) {
    window.history.back()
  } else {
    router.push({ name: 'home' }) // 没有历史记录则跳首页
  }
}
onMounted(async () => {
  // 先尝试从 store 获取
  echo.value = getEchoFromStore()

  // 如果 store 里没有，再发请求兜底
  if (!echo.value) {
    const res = await fetchGetEchoById(echoId)
    if (res.code === 1) {
      echo.value = res.data
    }
  }
  isLoading.value = false
})
</script>
