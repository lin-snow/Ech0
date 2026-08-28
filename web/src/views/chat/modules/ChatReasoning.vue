<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ChatActivity :label="label" :active="active" :clock="clock">
    <template #glyph>
      <Reasoning class="reasoning__glyph" aria-hidden="true" />
    </template>
    <div ref="bodyEl" class="reasoning__body">
      <TheMdPreview :content="text" />
    </div>
  </ChatActivity>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { TheMdPreview } from '@/components/advanced/md'
import Reasoning from '@/components/icons/reasoning.vue'
import ChatActivity from './ChatActivity.vue'
import { formatDuration, useElapsed } from './useElapsed'

const props = defineProps<{
  text: string
  active?: boolean
  durationMs?: number
}>()

const { t } = useI18n()

const bodyEl = ref<HTMLElement | null>(null)

const active = computed<boolean>(() => props.active === true)
const liveMs = useElapsed(active)

const label = computed<string>(() =>
  active.value ? t('chatPanel.reasoningThinking') : t('chatPanel.reasoningDone'),
)

/**
 * `reasoning_ms` is the window between the first reasoning token and the first
 * answer token, not the time the model spent thinking. A provider that buffers
 * the whole thought and flushes it at once closes that window in milliseconds,
 * and `0.0s` would then assert something we cannot know — so anything under a
 * second is reported as no duration at all rather than as none elapsed.
 */
const MEASURABLE_MS = 1000

const clock = computed<string>(() => {
  if (active.value) return formatDuration(liveMs.value)
  const settled = props.durationMs
  if (settled === undefined || settled < MEASURABLE_MS) return ''
  return formatDuration(settled)
})

// The trace is capped, so without this the reader would be pinned to the first
// few lines of a thought that is still being written.
watch(
  () => [props.text, active.value] as const,
  async () => {
    if (!active.value) return
    await nextTick()
    const el = bodyEl.value
    if (el) el.scrollTop = el.scrollHeight
  },
)
</script>

<style scoped>
.reasoning__glyph :deep(path) {
  fill: currentColor;
}

.reasoning__body {
  max-height: 13rem;
  overflow-y: auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  color: var(--color-text-secondary);
  font-size: 0.86rem;
}

.reasoning__body :deep(.echo-markdown) {
  line-height: 1.7;
  color: var(--color-text-secondary);
}
</style>
