<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="actions">
    <span v-if="hint" class="actions__hint">{{ hint }}</span>

    <button
      v-if="text.trim().length > 0"
      class="actions__btn"
      :title="copyLabel"
      :aria-label="copyLabel"
      @click="handleCopy"
    >
      <svg
        v-if="copied"
        class="actions__icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="m20 6-11 11-5-5" />
      </svg>
      <svg
        v-else
        class="actions__icon"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <rect x="9" y="9" width="12" height="12" rx="2.5" />
        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
      </svg>
    </button>

    <button
      v-if="canRetry"
      class="actions__btn"
      :title="t('chatPanel.retry')"
      :aria-label="t('chatPanel.retry')"
      @click="emit('retry')"
    >
      <svg
        class="actions__icon actions__icon--retry"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="M21 12a9 9 0 1 1-2.64-6.36" />
        <path d="M21 3v6h-6" />
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'

const props = defineProps<{
  text: string
  /** Only the newest turn can be re-run: retrying rewrites that turn in place. */
  canRetry?: boolean
  /** Explains an answer that came back empty; blank the rest of the time. */
  hint?: string
}>()

const emit = defineEmits<{
  (e: 'retry'): void
}>()

const { t } = useI18n()

const copied = ref<boolean>(false)
let resetTimer = 0

const copyLabel = computed<string>(() =>
  copied.value ? String(t('chatPanel.copied')) : String(t('chatPanel.copy')),
)

const handleCopy = async () => {
  try {
    await navigator.clipboard.writeText(props.text)
  } catch {
    theToast.error(String(t('chatPanel.copyFailed')))
    return
  }
  copied.value = true
  if (resetTimer) clearTimeout(resetTimer)
  resetTimer = window.setTimeout(() => {
    copied.value = false
    resetTimer = 0
  }, 1600)
}

onBeforeUnmount(() => {
  if (resetTimer) clearTimeout(resetTimer)
})
</script>

<style scoped>
.actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.25rem;
  margin-left: -0.3rem;
}

.actions__hint {
  margin-left: 0.3rem;
  margin-right: 0.25rem;
  color: var(--color-text-muted);
  font-size: 0.8rem;
  line-height: 1.5;
}

.actions__btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.55rem;
  height: 1.55rem;
  border: none;
  border-radius: 0.4rem;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0.5;
  transition:
    opacity 0.18s ease,
    background 0.18s ease,
    color 0.18s ease;
}

.actions__btn:hover {
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  opacity: 1;
}

.actions__icon {
  width: 0.9rem;
  height: 0.9rem;
  transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1);
}

.actions__btn:hover .actions__icon--retry {
  transform: rotate(180deg);
}

@media (prefers-reduced-motion: reduce) {
  .actions__icon {
    transition: none;
  }
}
</style>
