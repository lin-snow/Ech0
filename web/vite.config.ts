// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/// <reference types="vitest/config" />
import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import UnoCSS from 'unocss/vite'
import viteCompression from 'vite-plugin-compression'

import { fingerprintPlugin } from './src/plugins/fingerprint-plugin.ts'
import { welcomePlugin } from './src/plugins/welcome-plugin.ts'

export default defineConfig(({ command }) => ({
  plugins: [
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag === 'meting-js' || tag === 'cap-widget',
        },
      },
    }),
    ...(command === 'serve' ? [vueDevTools()] : []),
    UnoCSS(),
    viteCompression({
      deleteOriginFile: false,
      threshold: 10240,
      filter: (file) => /\.(js|mjs|css|html|svg)$/i.test(file),
    }),
    fingerprintPlugin(),
    welcomePlugin(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/setup.ts'],
    include: ['tests/**/*.{test,spec}.ts'],
    clearMocks: true,
    restoreMocks: true,
  },
  build: {
    outDir: '../template/dist',
    emptyOutDir: true,
    reportCompressedSize: false,
    rollupOptions: {
      checks: {
        invalidAnnotation: false,
        pluginTimings: false,
      },
      output: {
        manualChunks(id) {
          const normalizedId = id.replaceAll('\\', '/')
          if (normalizedId.includes('/node_modules/floating-vue/')) {
            return 'floating-vue'
          }
          if (normalizedId.includes('/node_modules/highlight.js/')) {
            return 'highlight'
          }
          if (
            normalizedId.includes('/node_modules/markdown-it/') ||
            normalizedId.includes('/node_modules/linkify-it/') ||
            normalizedId.includes('/node_modules/mdurl/') ||
            normalizedId.includes('/node_modules/uc.micro/')
          ) {
            return 'markdown'
          }
          return undefined
        },
      },
    },
  },
}))
