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
import { onMounted, ref } from 'vue'
import { fetchGetStatus, fetchHelloEch0 } from '@/service/api'
import { useSettingStore } from '@/stores/setting'
import { useThemeStore } from '@/stores/theme'
import { getApiUrl } from '@/service/request/shared'
import { theToast } from '@/utils/toast'

const settingStore = useSettingStore()
const themeStore = useThemeStore()

const { SystemSetting } = storeToRefs(settingStore)

const apiUrl = getApiUrl()
const logo = ref<string>('/favicon.svg')

const modeTextMap = { light: '浅色', dark: '深色', auto: '自动' }

const getModeText = () => {
  return modeTextMap[themeStore.mode]
}

const handleHello = () => {
  themeStore.toggleTheme()

  const hello = ref<App.Api.Ech0.HelloEch0>()

  fetchHelloEch0().then((res) => {
    if (res.code === 1) {
      hello.value = res.data
      theToast.success('你好呀！ 👋', {
        description: `当前版本：v${hello.value.version} | ${getModeText()}`,
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
  fetchGetStatus().then((res) => {
    if (res.code === 1) {
      const theLogo = res.data.logo
      if (theLogo && theLogo !== '') {
        logo.value = `${apiUrl}${theLogo}`
      }
    }
  })
})
</script>

<style scoped></style>
