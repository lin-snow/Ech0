<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootRef" class="echo-markdown">
    <div v-if="!rendererReady && props.content" class="markdown-renderer-fallback">
      {{ props.content }}
    </div>
    <div v-else v-html="html"></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { loadedMarkdown, preloadMarkdown } from '../core/loader'
import type { MarkdownModule } from '../core/loader'
import type { MarkdownLabels } from '../core/markdown'
import '../styles/markdown.scss'

const props = defineProps<{
  content: string
}>()

const rootRef = ref<HTMLElement | null>(null)
const { t } = useI18n()
const copyResetTimers = new WeakMap<HTMLButtonElement, ReturnType<typeof setTimeout>>()
const expandLabel = computed(() => String(t('markdown.expand')))
const collapseLabel = computed(() => String(t('markdown.collapse')))
const copyLabel = computed(() => String(t('markdown.copy')))
const copiedLabel = computed(() => String(t('markdown.copied')))
const taskCheckboxLabel = computed(() => String(t('markdown.taskCheckboxLabel')))
const html = ref('')
const rendererReady = ref(false)
let markdownCore: MarkdownModule | null = loadedMarkdown()
let renderSequence = 0

const currentLabels = (): MarkdownLabels => ({
  expandLabel: expandLabel.value,
  collapseLabel: collapseLabel.value,
  copyLabel: copyLabel.value,
  copiedLabel: copiedLabel.value,
  taskCheckboxLabel: taskCheckboxLabel.value,
})

const renderContent = async () => {
  const currentSequence = ++renderSequence
  const content = props.content
  if (!content) {
    html.value = ''
    rendererReady.value = true
    return
  }

  if (!markdownCore) {
    markdownCore = await preloadMarkdown().catch(() => null)
    if (!markdownCore || currentSequence !== renderSequence) return
  }

  // Paint final line boxes on the first frame: a raw-text fallback that is
  // replaced after paint changes row height and reflows the virtual timeline.
  const labels = currentLabels()
  html.value = markdownCore.renderMarkdownSync(content, labels)
  rendererReady.value = true

  if (!markdownCore.isHighlighterReady() && markdownCore.needsHighlight(content)) {
    // The coloured swap keeps the same text and line boxes, so it cannot move
    // the row; it only adds token spans.
    const highlighted = await markdownCore.renderMarkdown(content, labels)
    if (currentSequence !== renderSequence) return
    html.value = highlighted
  }
}

function onRootClick(event: Event) {
  if (!rendererReady.value) return
  const target = event.target
  if (!(target instanceof HTMLElement)) return

  const toggleButton = target.closest<HTMLButtonElement>('.code-block-toggle')
  if (toggleButton && rootRef.value?.contains(toggleButton)) {
    const block = toggleButton.closest<HTMLElement>('.code-block--collapsible')
    if (!block) return

    const isCollapsed = block.classList.toggle('code-block--collapsed')
    const expandLabel = toggleButton.dataset.expandLabel ?? String(t('markdown.expand'))
    const collapseLabel = toggleButton.dataset.collapseLabel ?? String(t('markdown.collapse'))

    toggleButton.setAttribute('aria-expanded', String(!isCollapsed))
    toggleButton.textContent = isCollapsed ? expandLabel : collapseLabel
    return
  }

  const copyButton = target.closest<HTMLButtonElement>('.code-block-copy')
  if (copyButton && rootRef.value?.contains(copyButton)) {
    const block = copyButton.closest<HTMLElement>('.code-block')
    const codeEl = block?.querySelector<HTMLElement>('pre code')
    if (!codeEl) return

    const text = codeEl.textContent ?? ''
    const copyLabel = copyButton.dataset.copyLabel ?? String(t('markdown.copy'))
    const copiedLabel = copyButton.dataset.copiedLabel ?? String(t('markdown.copied'))

    const markCopied = () => {
      copyButton.classList.add('is-copied')
      copyButton.textContent = copiedLabel
      const prev = copyResetTimers.get(copyButton)
      if (prev) clearTimeout(prev)
      copyResetTimers.set(
        copyButton,
        setTimeout(() => {
          copyButton.classList.remove('is-copied')
          copyButton.textContent = copyLabel
          copyResetTimers.delete(copyButton)
        }, 1800),
      )
    }

    if (navigator.clipboard?.writeText) {
      navigator.clipboard
        .writeText(text)
        .then(markCopied)
        .catch(() => {})
    }
  }
}

onMounted(() => {
  rootRef.value?.addEventListener('click', onRootClick)
})

onBeforeUnmount(() => {
  rootRef.value?.removeEventListener('click', onRootClick)
})

watch(
  [() => props.content, expandLabel, collapseLabel, taskCheckboxLabel],
  () => {
    void renderContent()
  },
  { immediate: true },
)
</script>

<style scoped>
.markdown-renderer-fallback {
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  font: inherit;
  color: inherit;
}
</style>
