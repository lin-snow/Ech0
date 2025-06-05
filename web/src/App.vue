<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import { watch } from 'vue'
import { useSettingStore } from '@/stores/settting'
import { storeToRefs } from 'pinia'

const userStore = useUserStore()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

watch(
  () => SystemSetting.value.site_title,
  (title) => {
    if (title) document.title = title
  },
  { immediate: true },
)

onMounted(async () => {
  // 自动登录
  await userStore.autoLogin()

  // 获取系统设置
  settingStore.getSystemReady()
  await settingStore.getSystemSetting()

  // 注入自定义 CSS
  if (SystemSetting.value.custom_css) {
    const styleTag = document.createElement('style')
    styleTag.textContent = SystemSetting.value.custom_css
    document.head.appendChild(styleTag)
  }

  // 注入自定义 JS
  if (SystemSetting.value.custom_js) {
    const scriptTag = document.createElement('script')
    scriptTag.textContent = SystemSetting.value.custom_js
    document.body.appendChild(scriptTag)
  }
})
</script>

<template>
  <RouterView />
</template>

<style scoped></style>
