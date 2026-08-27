// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { Plugin } from 'vite'
import { printWelcome } from '../scripts/welcome.ts'

export function welcomePlugin(): Plugin {
  let hasShown = false

  return {
    name: 'welcome-banner',
    configureServer(server) {
      server.middlewares.use('/', (req, res, next) => {
        if (!hasShown) {
          printWelcome()
          hasShown = true
        }
        next()
      })
    },
  }
}
