<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router'
import { computed, onMounted, ref, watch } from 'vue'
import { useSettingStore, useThemeStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { Toaster } from 'vue-sonner'
import { resolveAvatarUrl } from './service/request/shared'
import 'vue-sonner/style.css'
import BaseDialog from './components/common/BaseDialog.vue'
import TheRouteProgress from './components/common/TheRouteProgress.vue'

import { useBaseDialog } from '@/composables/useBaseDialog'
import { useBfCacheRestore } from '@/composables/useBfCacheRestore'
import { useSeoHead } from '@/composables/useSeoHead'

const { register, title, description, handleConfirm, handleCancel } = useBaseDialog()
const dialogRef = ref()

const router = useRouter()
const transitionName = ref('fade')
const { isBfCacheRestore } = useBfCacheRestore({
  debug: true,
  onRestore: () => {
    transitionName.value = 'none'
  },
})

router.afterEach((to, from) => {
  if (isBfCacheRestore.value) {
    transitionName.value = 'none'
    return
  }

  const toName = to.name as string
  const fromName = from.name as string
  if (toName?.startsWith('panel-') && fromName?.startsWith('panel-')) {
    transitionName.value = 'none'
    return
  }

  const routeDepth: Record<string, number> = {
    home: 0,
    echo: 1,
    panel: 1,
    auth: 1,
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
const themeStore = useThemeStore()
const { theme } = storeToRefs(themeStore)
const toasterTheme = computed(() => (theme.value === 'dark' ? 'dark' : 'light'))
const sunnyVideoSrc = ref('')
let sunnyVideoLoading = false

const DEFAULT_FAVICON = '/favicon.ico'
const CUSTOM_STYLE_ID = 'ech0-custom-style'
const CUSTOM_SCRIPT_ID = 'ech0-custom-script'
useSeoHead(SystemSetting)

const updateFavicon = (logo?: string) => {
  const head = document.head
  if (!head) return

  const href = resolveAvatarUrl(logo, DEFAULT_FAVICON)
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
  () => SystemSetting.value.server_logo,
  (logo) => {
    updateFavicon(logo)
  },
  { immediate: true },
)

const upsertCustomStyle = (css: string) => {
  const head = document.head
  if (!head) return

  const normalized = css.trim()
  const existing = document.getElementById(CUSTOM_STYLE_ID) as HTMLStyleElement | null

  if (!normalized) {
    existing?.remove()
    return
  }

  if (existing) {
    existing.textContent = normalized
    return
  }

  const styleTag = document.createElement('style')
  styleTag.id = CUSTOM_STYLE_ID
  styleTag.textContent = normalized
  head.appendChild(styleTag)
}

const upsertCustomScript = (script: string) => {
  const body = document.body
  if (!body) return

  const normalized = script.trim()
  const existing = document.getElementById(CUSTOM_SCRIPT_ID)
  existing?.remove()

  if (!normalized) {
    return
  }

  const scriptTag = document.createElement('script')
  scriptTag.id = CUSTOM_SCRIPT_ID
  scriptTag.textContent = normalized
  body.appendChild(scriptTag)
}

watch(
  () => SystemSetting.value.custom_css,
  (css) => {
    upsertCustomStyle(css || '')
  },
  { immediate: true },
)

watch(
  () => SystemSetting.value.custom_js,
  (script) => {
    upsertCustomScript(script || '')
  },
  { immediate: true },
)

watch(
  theme,
  async (nextTheme) => {
    if (nextTheme !== 'sunny') return
    if (sunnyVideoSrc.value || sunnyVideoLoading) return

    sunnyVideoLoading = true
    try {
      const mod = await import('@/assets/leaves.mp4')
      sunnyVideoSrc.value = mod.default
    } finally {
      sunnyVideoLoading = false
    }
  },
  { immediate: true },
)

onMounted(() => {
  register(dialogRef.value)
})
</script>

<template>
  <div v-if="theme === 'sunny' && sunnyVideoSrc" class="sunny-atmosphere" aria-hidden="true">
    <video
      class="sunny-atmosphere__leaves"
      :src="sunnyVideoSrc"
      autoplay
      muted
      loop
      playsinline
      preload="metadata"
    />
  </div>
  <div class="app-stack">
    <TheRouteProgress />
    <RouterView v-slot="{ Component }">
      <Transition :name="transitionName" mode="out-in">
        <component :is="Component" />
      </Transition>
    </RouterView>
    <Toaster :theme="toasterTheme" position="top-right" :expand="false" richColors />
    <BaseDialog
      ref="dialogRef"
      :title="title"
      :description="description"
      @confirm="handleConfirm"
      @cancel="handleCancel"
    />
  </div>
</template>

<style scoped>
.sunny-atmosphere {
  position: fixed;
  inset: 0;
  z-index: 3;
  pointer-events: none;
  overflow: hidden;
  mix-blend-mode: multiply;
}

.sunny-atmosphere__leaves {
  width: 100%;
  height: 100%;
  object-fit: cover;
  opacity: var(--sunny-video-opacity, 0.52);
  filter: saturate(1.02) contrast(1.01);
}

.app-stack {
  position: relative;
  z-index: 2;
  min-height: 100%;
  isolation: isolate;
}

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
