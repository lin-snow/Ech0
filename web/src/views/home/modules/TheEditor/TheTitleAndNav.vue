<template>
  <div class="flex justify-between items-center py-1 px-3">
    <div class="flex flex-row items-center gap-2 justify-between">
      <!-- <div class="text-xl">👾</div> -->
      <div>
        <img
          :src="logo"
          alt="logo"
          class="w-6 sm:w-7 h-6 sm:h-7 rounded-full ring-1 ring-[var(--ring-color)] shadow-sm object-cover"
        />
      </div>
      <h1 class="text-[var(--editor-title-color)] font-bold sm:text-xl">
        {{ SystemSetting.server_name }}
      </h1>
    </div>

    <div class="flex flex-row items-center gap-2">
      <!-- Hello -->
      <div
        class="p-1 ring-1 ring-inset ring-[var(--ring-color)] rounded-full transition-colors duration-200 cursor-pointer"
      >
        <Hello @click="handleHello" class="w-6 h-6" />
      </div>
      <!-- Github -->
      <!--
      <div>
        <a href="https://github.com/lin-snow/Ech0" target="_blank" title="Github">
          <Github class="w-6 sm:w-7 h-6 sm:h-7 text-[var(--text-color-400)]" />
        </a>
      </div>
      -->
    </div>
  </div>
</template>

<script setup lang="ts">
import Hello from '@/components/icons/hello.vue'
import { storeToRefs } from 'pinia'
import { ref } from 'vue'
import { fetchHelloEch0 } from '@/service/api'
import { useSettingStore, useThemeStore, useUserStore } from '@/stores'
import { getApiUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const settingStore = useSettingStore()
const themeStore = useThemeStore()
const userStore = useUserStore()

const { SystemSetting } = storeToRefs(settingStore)
const { user, isLogin } = storeToRefs(userStore)

const apiUrl = getApiUrl()
const logo = ref<string>('/Ech0.svg')
if (isLogin.value && user.value?.avatar && user.value?.avatar !== '') {
  logo.value = `${apiUrl}${user.value?.avatar}`
} else if (
  SystemSetting.value.server_logo &&
  SystemSetting.value.server_logo !== '' &&
  SystemSetting.value.server_logo !== '/Ech0.svg'
) {
  logo.value = `${apiUrl}${SystemSetting.value.server_logo}`
}

const handleHello = (event: MouseEvent) => {
  themeStore.toggleTheme(event)
  const modeText =
    themeStore.mode === 'system' ? 'Auto' : themeStore.mode === 'light' ? 'Light' : 'Dark'

  const hello = ref<App.Api.Ech0.HelloEch0>()

  fetchHelloEch0().then((res) => {
    if (res.code === 1) {
      hello.value = res.data
      theToast.success('你好呀！ 👋', {
        description: `当前版本：v${hello.value.version} | ${modeText}`,
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
</script>

<style scoped></style>
