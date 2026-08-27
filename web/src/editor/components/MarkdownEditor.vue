<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="markdown-editor-root">
    <div v-if="isFullMode" class="markdown-editor-backdrop" @click="exitFullMode"></div>

    <div v-if="!isFullMode" class="markdown-editor">
      <button
        type="button"
        class="mode-toggle-btn"
        v-tooltip="t('editor.fullscreenEdit')"
        :aria-label="t('editor.fullscreenEdit')"
        @click="enterFullMode"
      >
        <Full class="w-3.5 h-3.5" />
      </button>
      <div class="editor-content">
        <textarea
          ref="textareaRef"
          class="editor-input"
          :placeholder="placeholder"
          :value="modelValue"
          @input="onInput"
        />
      </div>
    </div>

    <div v-else class="markdown-full-shell" :class="{ 'no-preview': isPreviewMode }">
      <div class="markdown-editor is-full">
        <div class="toolbar">
          <button
            v-for="item in toolbarItems"
            :key="item.action"
            type="button"
            class="toolbar-btn"
            v-tooltip="item.label"
            :aria-label="item.label"
            @click="onToolbarClick(item.action)"
          >
            <span class="toolbar-icon" aria-hidden="true">{{ item.icon }}</span>
          </button>
          <div class="toolbar-actions">
            <button
              type="button"
              class="toolbar-btn"
              v-tooltip="isPreviewMode ? t('editor.showPreview') : t('editor.hidePreview')"
              :aria-label="isPreviewMode ? t('editor.showPreview') : t('editor.hidePreview')"
              @click="isPreviewMode = !isPreviewMode"
            >
              <Preview class="w-3.5 h-3.5" />
            </button>
            <button
              type="button"
              class="toolbar-btn mode-toggle-btn-inline"
              v-tooltip="t('editor.exitFullscreen')"
              :aria-label="t('editor.exitFullscreen')"
              @click="exitFullMode"
            >
              <Closefull class="w-3.5 h-3.5" />
            </button>
          </div>
        </div>

        <div class="editor-content">
          <textarea
            ref="textareaRef"
            class="editor-input"
            :placeholder="placeholder"
            :value="modelValue"
            @input="onInput"
          />
        </div>
      </div>

      <div v-if="!isPreviewMode" class="markdown-preview-dock">
        <MarkdownPreviewCard :content="modelValue" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from 'vue'
import MarkdownPreviewCard from './MarkdownPreviewCard.vue'
import { applyMarkdownAction } from '../composables/useMarkdownEditorActions'
import type { MarkdownEditorAction } from '../types'
import Full from '@/components/icons/full.vue'
import Closefull from '@/components/icons/closefull.vue'
import Preview from '@/components/icons/preview.vue'
import { useI18n } from 'vue-i18n'

defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)
const isFullMode = ref(false)
const isPreviewMode = ref(false)
const FULL_MODE_LOCK_CLASS = 'md-editor-full-open'
const { t } = useI18n()

const toolbarItems = computed<Array<{ label: string; icon: string; action: MarkdownEditorAction }>>(
  () => [
    { label: String(t('editor.bold')), icon: 'B', action: 'bold' },
    { label: String(t('editor.italic')), icon: 'I', action: 'italic' },
    { label: String(t('editor.heading')), icon: 'H', action: 'heading' },
    { label: String(t('editor.quote')), icon: '❝', action: 'quote' },
    { label: String(t('editor.unorderedList')), icon: '•', action: 'unorderedList' },
    { label: String(t('editor.codeBlock')), icon: '</>', action: 'codeBlock' },
  ],
)

function onInput(event: Event) {
  const target = event.target as HTMLTextAreaElement
  emit('update:modelValue', target.value)
}

function onToolbarClick(action: MarkdownEditorAction) {
  if (isPreviewMode.value) {
    isPreviewMode.value = false
    nextTick(() => {
      if (!textareaRef.value) return
      applyMarkdownAction(textareaRef.value, action)
    })
    return
  }

  if (!textareaRef.value) return
  applyMarkdownAction(textareaRef.value, action)
}

function syncPageScrollLock(locked: boolean) {
  if (typeof document === 'undefined') return
  const root = document.documentElement
  const body = document.body
  if (locked) {
    root.classList.add(FULL_MODE_LOCK_CLASS)
    body.classList.add(FULL_MODE_LOCK_CLASS)
    return
  }
  root.classList.remove(FULL_MODE_LOCK_CLASS)
  body.classList.remove(FULL_MODE_LOCK_CLASS)
}

function enterFullMode() {
  isFullMode.value = true
  isPreviewMode.value = false
  syncPageScrollLock(true)
}

function exitFullMode() {
  isFullMode.value = false
  isPreviewMode.value = false
  syncPageScrollLock(false)
}

function onWindowKeydown(event: KeyboardEvent) {
  const key = event.key.toLowerCase()
  const isUndo = key === 'z' && event.ctrlKey && !event.metaKey && !event.altKey
  const isRedo =
    (key === 'y' && event.ctrlKey && !event.metaKey && !event.altKey) ||
    (key === 'z' && event.ctrlKey && event.shiftKey && !event.metaKey && !event.altKey)

  if ((isUndo || isRedo) && textareaRef.value) {
    if (document.activeElement !== textareaRef.value) {
      return
    }
    return
  }

  if (event.key === 'Escape' && isFullMode.value) {
    exitFullMode()
  }
}

onMounted(() => {
  window.addEventListener('keydown', onWindowKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onWindowKeydown)
  syncPageScrollLock(false)
})
</script>
