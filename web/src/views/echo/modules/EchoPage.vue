<template>
  <div class="px-3 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <div class="w-full sm:max-w-lg mx-auto">
      <div class="mx-auto max-w-sm">
        <!-- 返回上一页 -->
        <BaseButton
          @click="router.push({ name: 'home' })"
          class="text-gray-600 rounded-md !shadow-none !border-none !ring-0 !bg-transparent group"
          title="返回首页"
        >
          <Arrow
            class="w-9 h-9 rotate-180 transition-transform duration-200 group-hover:-translate-x-1"
          />
        </BaseButton>
      </div>

      <div v-if="echo" class="w-full sm:mt-1 mx-auto">
        <TheEchoDetail :echo="echo" @update-like-count="handleUpdateLikeCount" />
        <TheComment class="my-2" />
      </div>
      <div v-else class="w-full sm:mt-1 text-gray-300">
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

const router = useRouter()
const route = useRoute()
const echoId = route.params.echoId as string
const echo = ref<App.Api.Ech0.Echo | null>(null)
const isLoading = ref(true)

// 刷新点赞数据
const handleUpdateLikeCount = () => {
  if (echo.value) {
    // 更新 Echo 的点赞数量
    echo.value.fav_count += 1
  }
}

onMounted(() => {
  // 在这里可以添加获取Echo详情的逻辑
  fetchGetEchoById(echoId).then((res) => {
    if (res.code === 1) {
      echo.value = res.data
      isLoading.value = false
    }
  })
})
</script>
