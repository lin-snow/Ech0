/* SPDX-License-Identifier: AGPL-3.0-or-later */
/* Copyright (C) 2025-2026 lin-snow */

import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: { port: 5174 },
  preview: { port: 5174 },
})
