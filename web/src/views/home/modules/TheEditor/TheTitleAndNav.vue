<template>
  <div class="flex justify-between items-center py-1 px-3">
    <!-- 未登录状态：显示站点Logo和服务名称 -->
    <div v-if="!isLogin" class="flex flex-row items-center gap-2 justify-between">
      <div>
        <img
          :src="siteLogo"
          alt="站点Logo"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-lg ring-1 ring-gray-200 shadow-sm object-cover"
          @error="handleImageError"
        />
      </div>
      <h1 class="text-[var(--editor-title-color)] font-bold sm:text-xl">
        {{ SystemSetting.server_name }}
      </h1>
    </div>

    <!-- 已登录状态：显示用户头像和用户名 -->
    <div v-else class="flex flex-row items-center gap-2 justify-between">
      <div>
        <img
          :src="userAvatar"
          alt="用户头像"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-full ring-1 ring-gray-200 shadow-sm object-cover"
          @error="handleImageError"
        />
      </div>
      <h1 class="text-[var(--editor-title-color)] font-bold sm:text-xl">
        {{ user?.username }}
      </h1>
    </div>

    <div class="flex flex-row items-center gap-2">
      <!-- Hello -->
      <div
        class="p-1 ring-1 ring-inset ring-[var(--ring-color)] rounded-full transition-colors duration-200 cursor-pointer"
      >
        <Hello @click="handleHello" class="w-6 h-6" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Hello from '@/components/icons/hello.vue'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref } from 'vue'
import { fetchHelloEch0 } from '@/service/api'
import { useSettingStore } from '@/stores/setting'
import { useUserStore } from '@/stores/user'
import { useThemeStore } from '@/stores/theme'
import { getApiUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const settingStore = useSettingStore()
const userStore = useUserStore()
const themeStore = useThemeStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)

const apiUrl = getApiUrl()

// 站点Logo（未登录时显示）
const siteLogo = computed(() => {
  const logo = SystemSetting.value.logo
  if (!logo || logo.length === 0) {
    return '/favicon.svg'
  }
  return logo.startsWith('http') ? logo : `${apiUrl}${logo}`
})

// 用户头像（已登录时显示）
const userAvatar = computed(() => {
  const avatar = user.value?.avatar
  if (!avatar || avatar.length === 0) {
    return '/favicon.svg'
  }
  return avatar.startsWith('http') ? avatar : `${apiUrl}${avatar}`
})

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = '/favicon.svg'
}

const handleHello = () => {
  themeStore.toggleTheme()

  const hello = ref<App.Api.Ech0.HelloEch0>()

  fetchHelloEch0().then((res) => {
    if (res.code === 1) {
      hello.value = res.data
      theToast.success('你好呀！ 👋', {
        description: `当前版本：v${hello.value.version}`,
        duration: 2000,
        action: {
          label: 'Github',
          onClick: () => {
            window.open(hello.value?.github, '_blank')
          },
        },
      })
    }
  })
}

onMounted(() => {
  // 获取系统设置
  settingStore.getSystemSetting()
})
</script>

<style scoped></style>
