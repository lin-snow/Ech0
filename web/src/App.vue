<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router'
import { onMounted, ref } from 'vue'
import { watch } from 'vue'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Toaster } from 'vue-sonner'
import { getApiUrl } from './service/request/shared'
import 'vue-sonner/style.css'
import BaseDialog from './components/common/BaseDialog.vue'

import { useBaseDialog } from '@/composables/useBaseDialog'

const { register, title, description, handleConfirm } = useBaseDialog()
const dialogRef = ref()

// 路由切换动画
const router = useRouter()
const transitionName = ref('fade')

// 监听路由变化，根据导航方向选择动画
router.afterEach((to, from) => {
  // Panel 子页面之间切换不使用动画
  const toName = to.name as string
  const fromName = from.name as string
  if (toName?.startsWith('panel-') && fromName?.startsWith('panel-')) {
    transitionName.value = 'none'
    return
  }

  // 定义路由层级（用于判断前进/后退）
  const routeDepth: Record<string, number> = {
    home: 0,
    echo: 1,
    panel: 1,
    auth: 1,
    connect: 1,
    hub: 1,
    widget: 1,
    'not-found': 2,
  }

  const toDepth = routeDepth[toName] ?? 1
  const fromDepth = routeDepth[fromName] ?? 1

  if (toDepth > fromDepth) {
    transitionName.value = 'slide-left'
  } else if (toDepth < fromDepth) {
    transitionName.value = 'slide-right'
  } else {
    transitionName.value = 'fade'
  }
})

const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const DEFAULT_FAVICON = '/favicon.ico'
const API_URL = getApiUrl()

const updateFavicon = (logo?: string) => {
  const head = document.head
  if (!head) return

  const href = logo?.trim() ? API_URL + logo : DEFAULT_FAVICON
  const iconLinks = head.querySelectorAll<HTMLLinkElement>('link[rel*="icon"]')

  if (iconLinks.length > 0) {
    iconLinks.forEach((link) => {
      link.href = href
    })
    return
  }

  const newFavicon = document.createElement('link')
  newFavicon.rel = 'icon'
  newFavicon.href = href
  head.appendChild(newFavicon)
}

watch(
  () => SystemSetting.value.site_title,
  (title) => {
    if (title) document.title = title
  },
  { immediate: true },
)

watch(
  () => SystemSetting.value.server_logo,
  (logo) => {
    updateFavicon(logo)
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

onMounted(() => {
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
  <!-- 路由视图 - 带切换动画 -->
  <RouterView v-slot="{ Component, route }">
    <Transition :name="transitionName" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </RouterView>
  <!-- 通知组件 -->
  <Toaster theme="light" position="top-right" :expand="false" richColors />
  <!-- 全局弹窗对话框 -->
  <BaseDialog ref="dialogRef" :title="title" :description="description" @confirm="handleConfirm" />
</template>

<style scoped>
/* 路由切换动画 - 淡入淡出 + 轻微滑动 */
.fade-enter-active,
.fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

/* 滑动动画 - 用于前进后退 */
.slide-left-enter-active,
.slide-left-leave-active,
.slide-right-enter-active,
.slide-right-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.slide-left-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.slide-left-leave-to {
  opacity: 0;
  transform: translateX(-20px);
}

.slide-right-enter-from {
  opacity: 0;
  transform: translateX(-20px);
}

.slide-right-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
