import './assets/main.css'
import 'virtual:uno.css'
import 'vue-toastification/dist/index.css'

// Md-Editor Start
import { config } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

import screenfull from 'screenfull'

import katex from 'katex'
import 'katex/dist/katex.min.css'

import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'

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
})

// Md-Editor End

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import Toast from 'vue-toastification'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(Toast, {
  transition: 'Vue-Toastification__fade',
  maxToasts: 10,
  newestOnTop: true,
})

app.mount('#app')
