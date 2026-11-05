// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import 'virtual:uno.css'
import '@/themes/index.scss'

import { createApp } from 'vue'
import { setupI18n } from '@/locales'
import ChatEffectsPage from './ChatEffectsPage.vue'

document.documentElement.className = 'light'

const i18n = await setupI18n('zh-CN')
createApp(ChatEffectsPage).use(i18n).mount('#effects')
