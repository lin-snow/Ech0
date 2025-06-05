<template>
  <div class="w-full px-2">
    <!-- 一个月内的热力图 -->
    <div class="flex justify-center items-center">
      <TheHeatMap />
    </div>

    <div class="flex justify-center my-5">
      <div class="text-gray-400 text-md">
        <!-- 系统管理员 -->
        <div>
          <h1>
            当前系统管理员：
            <span class="ml-2">{{ status?.username }}</span>
          </h1>
        </div>
        <!-- 当前登录用户 -->
        <div>
          <h1>
            当前登录的用户：
            <span class="ml-2">
              {{ userStore?.user?.username ? userStore.user.username : '当前未登录' }}
            </span>
          </h1>
        </div>
        <!-- 当前共有Ech0 -->
        <div>
          <h1>
            当前Ech0总共有：
            <span class="ml-2">{{ status?.total_messages }}</span>
            <span class="ml-2">条</span>
          </h1>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { fetchGetStatus } from '@/service/api'
import { onMounted, ref } from 'vue'
import { useUserStore } from '@/stores/user'
import TheHeatMap from '@/components/advanced/TheHeatMap.vue'

const status = ref<App.Api.Ech0.Status>()
const userStore = useUserStore()

onMounted(async () => {
  await fetchGetStatus().then((res) => {
    status.value = res.data
  })
})
</script>
