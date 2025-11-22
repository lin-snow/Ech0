<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted, ref } from 'vue'
import { watch } from 'vue'
import { useSettingStore } from '@/stores/setting'
import { storeToRefs } from 'pinia'
import { Toaster } from 'vue-sonner'
import 'vue-sonner/style.css'
import BaseDialog from './components/common/BaseDialog.vue'
import { getApiUrl } from '@/service/request/shared'

import { useBaseDialog } from '@/composables/useBaseDialog'

const { register, title, description, handleConfirm } = useBaseDialog()
const dialogRef = ref()

const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

watch(
  () => SystemSetting.value.site_title,
  (title) => {
    if (title) document.title = title
  },
  { immediate: true },
)

const injectCustomContent = () => {
  // 注入自定义 CSS
  if (SystemSetting.value.custom_css && SystemSetting.value.custom_css.length > 0) {
    const styleTag = document.createElement('style')
    styleTag.textContent = SystemSetting.value.custom_css
    document.head.appendChild(styleTag)
  }

  // 注入自定义 JS
  if (SystemSetting.value.custom_js && SystemSetting.value.custom_js.length > 0) {
    const scriptTag = document.createElement('script')
    scriptTag.textContent = SystemSetting.value.custom_js
    document.body.appendChild(scriptTag)
  }
}

onMounted(async () => {
  // 获取系统设置
  await settingStore.getSystemSetting()

  const apiUrl = getApiUrl()

  // 初始设置 favicon
  const logo = SystemSetting.value.logo || '/favicon.svg'
  const fullLogoUrl = logo.startsWith('http') ? logo : `${apiUrl}${logo}`

  // 更新所有 favicon 相关的 link 标签
  const links = document.querySelectorAll("link[rel~='icon']")
  links.forEach((link) => {
    ;(link as HTMLLinkElement).href = fullLogoUrl
  })

  // 监听 logo 变化
  watch(
    () => SystemSetting.value.logo,
    (newLogo) => {
      const logoUrl = newLogo || '/favicon.svg'
      const fullLogoUrl = logoUrl.startsWith('http') ? logoUrl : `${apiUrl}${logoUrl}`

      // 更新所有 favicon 相关的 link 标签
      const links = document.querySelectorAll("link[rel~='icon']")
      links.forEach((link) => {
        ;(link as HTMLLinkElement).href = fullLogoUrl
      })
    },
  )

  // 注入自定义CSS 和 JS
  watch(
    () => SystemSetting.value.custom_css || SystemSetting.value.custom_js,
    (newSetting) => {
      if (newSetting) {
        injectCustomContent()
      }
    },
    { immediate: true },
  )

  // 初始注入
  register(dialogRef.value) // 全局注册弹窗对话框
})
</script>

<template>
  <!-- 路由视图 -->
  <RouterView />
  <!-- 通知组件 -->
  <Toaster theme="light" position="top-right" :expand="false" richColors />
  <!-- 全局弹窗对话框 -->
  <BaseDialog ref="dialogRef" :title="title" :description="description" @confirm="handleConfirm" />
</template>

<style scoped></style>
