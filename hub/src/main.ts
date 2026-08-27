// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import 'virtual:uno.css'
import '../../web/src/themes/index.scss'
import './syncWebTheme'
import 'floating-vue/dist/style.css'

import { createApp } from 'vue'
import { createHubI18n } from './i18n'
import './style.css'
import App from './App.vue'
import router from './router'

async function bootstrap() {
  const FloatingVue = (await import('floating-vue')).default
  const app = createApp(App)
  app.use(createHubI18n())
  app.use(router)
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
  app.mount('#app')
}

void bootstrap()
