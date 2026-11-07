// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import 'virtual:uno.css'
import '@/themes/index.scss'
import 'floating-vue/dist/style.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { initStores } from './stores/store-init'
import { preloadMarkdown } from './editor/core/loader'
import { useSettingStore } from './stores/setting'
import { useInitStore } from './stores/init'
import { setupI18n } from './locales'

import BaseDialog from '@/components/common/BaseDialog.vue'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// The timeline restores its own row-based anchor. Browser scroll restoration
// would move the document to a pixel offset that the virtual list immediately
// contradicts, so the user sees two jumps on reload.
if ('scrollRestoration' in history) {
  history.scrollRestoration = 'manual'
}

// The timeline renders markdown immediately; both bootstrap tasks run in
// parallel so the renderer is ready before the first row mounts.
await Promise.all([
  initStores().catch((e) => {
    console.error('Failed to initialize stores:', e)
  }),
  preloadMarkdown().catch((e) => {
    console.error('Failed to load the markdown renderer:', e)
  }),
])

const settingStore = useSettingStore()
const initStore = useInitStore()
const siteDefaultLocale = initStore.initialized
  ? settingStore.SystemSetting.default_locale
  : undefined
const i18n = await setupI18n(siteDefaultLocale)
const { default: FloatingVue } = await import('floating-vue')

app.use(router)
app.use(i18n)
app.use(FloatingVue, {
  themes: {
    tooltip: {
      triggers: ['hover'],
      hideTriggers: ['hover', 'click', 'touch'],
      placement: 'top',
      delay: { show: 300, hide: 80 },
      distance: 10,
      container: 'body',
      noAutoFocus: true,
      autoHide: true,
    },
  },
})

app.component('BaseDialog', BaseDialog)

app.mount('#app')

const appLoader = document.getElementById('app-loader')
let loaderCleared = false
const clearStartupLoader = () => {
  if (loaderCleared) return
  loaderCleared = true
  appLoader?.remove()
  document.documentElement.classList.remove('app-loading')
}

const startLoaderFade = () => {
  if (!appLoader) {
    clearStartupLoader()
    return
  }
  window.requestAnimationFrame(() => {
    appLoader.classList.add('fade-out')
  })
  appLoader.addEventListener('transitionend', clearStartupLoader, { once: true })
  window.setTimeout(clearStartupLoader, 800)
}

const loaderTimeout = new Promise<void>((resolve) => {
  window.setTimeout(resolve, 3000)
})
Promise.race([router.isReady().catch(() => undefined), loaderTimeout]).then(() => {
  startLoaderFade()
})
