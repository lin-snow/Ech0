<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="reasoning" :class="{ 'reasoning--active': active }">
    <button class="reasoning__header" :aria-expanded="!collapsed" @click="collapsed = !collapsed">
      <Reasoning class="reasoning__glyph" aria-hidden="true" />
      <span class="reasoning__label">{{ label }}</span>
      <svg
        class="reasoning__chevron"
        :class="{ 'reasoning__chevron--open': !collapsed }"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="m6 9 6 6 6-6" />
      </svg>
    </button>
    <Transition name="reasoning-fade">
      <div v-if="!collapsed" class="reasoning__body">
        <TheMdPreview :content="text" />
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { TheMdPreview } from '@/components/advanced/md'
import Reasoning from '@/components/icons/reasoning.vue'

const props = defineProps<{
  text: string
  active?: boolean
  durationMs?: number
}>()

const { t } = useI18n()

const collapsed = ref<boolean>(props.active !== true)

watch(
  () => props.active,
  (now, prev) => {
    if (prev && !now) collapsed.value = true
  },
)

const label = computed<string>(() => {
  if (props.active) return t('chatPanel.reasoningThinking')
  const seconds = Math.max(0, Math.round((props.durationMs ?? 0) / 1000))
  return t('chatPanel.reasoningDone', { seconds })
})
</script>

<style scoped>
.reasoning {
  width: 100%;
  margin-bottom: 0.55rem;
}

.reasoning__header {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  max-width: 100%;
  padding: 0.18rem 0.5rem 0.18rem 0.4rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 0.78rem;
  line-height: 1.5;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease;
}

.reasoning__header:hover {
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
}

.reasoning__glyph {
  width: 1.05rem;
  height: 1.05rem;
  flex-shrink: 0;
  opacity: 0.85;
}

.reasoning__glyph :deep(path) {
  fill: currentColor;
}

.reasoning--active .reasoning__glyph {
  color: var(--color-accent);
  opacity: 1;
  animation: reasoning-pulse 1.4s ease-in-out infinite;
}

@keyframes reasoning-pulse {
  0%,
  100% {
    opacity: 0.5;
  }

  50% {
    opacity: 1;
  }
}

.reasoning__label {
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.reasoning__chevron {
  width: 0.85rem;
  height: 0.85rem;
  flex-shrink: 0;
  opacity: 0.7;
  transition: transform 0.22s ease;
}

.reasoning__chevron--open {
  transform: rotate(180deg);
}

.reasoning__body {
  margin-top: 0.35rem;
  padding: 0.1rem 0 0.1rem 0.85rem;
  border-left: 2px solid var(--color-border-strong);
  color: var(--color-text-secondary);
  font-size: 0.86rem;
}

.reasoning__body :deep(.echo-markdown) {
  line-height: 1.7;
  color: var(--color-text-secondary);
}

.reasoning-fade-enter-active,
.reasoning-fade-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.2s ease;
}

.reasoning-fade-enter-from,
.reasoning-fade-leave-to {
  opacity: 0;
  transform: translateY(-2px);
}

@media (prefers-reduced-motion: reduce) {
  .reasoning--active .reasoning__glyph {
    animation: none;
  }

  .reasoning__chevron,
  .reasoning-fade-enter-active,
  .reasoning-fade-leave-active {
    transition: none;
  }
}
</style>
