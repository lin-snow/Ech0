import 'virtual:uno.css'
import '@/themes/index.scss'

// Google Model Viewer for 3D models
import '@google/model-viewer'

import { initStores } from './stores/store-init'

// Md-Editor Start
import { config } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

import screenfull from 'screenfull'

import katex from 'katex'
import 'katex/dist/katex.min.css'

import Cropper from 'cropperjs'

import mermaid from 'mermaid'

import highlight from 'highlight.js'
import 'highlight.js/styles/atom-one-dark.css'

// <3.0
// import prettier from 'prettier';
// import parserMarkdown from 'prettier/parser-markdown';
// >=3.0
import * as prettier from 'prettier'
import parserMarkdown from 'prettier/plugins/markdown'

config({
  editorExtensions: {
    prettier: {
      prettierInstance: prettier,
      parserMarkdownInstance: parserMarkdown,
    },
    highlight: {
      instance: highlight,
    },
    screenfull: {
      instance: screenfull,
    },
    katex: {
      instance: katex,
    },
    cropper: {
      instance: Cropper,
    },
    mermaid: {
      instance: mermaid,
    },
  },
  codeMirrorExtensions(extensions) {
    return [
      // 移除 linkShortener
      ...extensions.filter((ext) => ext.type !== 'linkShortener'),
    ]
  },
})

// Md-Editor End

// 自定义组件
import BaseDialog from '@/components/common/BaseDialog.vue'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)

// init
await initStores().catch((e) => {
  console.error('Failed to initialize stores:', e)
})

app.use(router)

// 全局注册组件
app.component('BaseDialog', BaseDialog)

app.mount('#app')

// 移除启动加载动画
const appLoader = document.getElementById('app-loader')
if (appLoader) {
  appLoader.remove()
}
