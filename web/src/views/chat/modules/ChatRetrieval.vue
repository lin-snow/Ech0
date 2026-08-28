<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ChatActivity :label="label" :active="active" :expandable="hasQueries">
    <template #glyph>
      <Search class="retrieval__glyph" aria-hidden="true" />
    </template>
    <ul class="retrieval__list">
      <li v-for="query in searches" :key="query" class="retrieval__row">
        <Search class="retrieval__row-glyph" aria-hidden="true" />
        <span class="retrieval__query">{{ query }}</span>
      </li>
      <li v-if="coverageLabel" class="retrieval__row retrieval__row--note">
        {{ coverageLabel }}
      </li>
    </ul>
  </ChatActivity>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Search from '@/components/icons/search.vue'
import ChatActivity from './ChatActivity.vue'

const props = defineProps<{
  searches: string[]
  coverage?: App.Api.Chat.ChatCoverage
  active?: boolean
}>()

const { t } = useI18n()

/* No clock here: the run's own elapsed time is reported once, by the waiting
   row, and a second number ticking beside it only invites comparison. */
const active = computed<boolean>(() => props.active === true)

const hasQueries = computed<boolean>(() => props.searches.length > 0)

const coverageLabel = computed<string>(() => {
  const coverage = props.coverage
  if (!coverage) return ''
  return coverage.truncated
    ? t('chatPanel.coverageTruncated', { returned: coverage.returned })
    : t('chatPanel.coverage', { total: coverage.total })
})

/**
 * The queries are the summary once there are any. A run whose only tool call
 * carried no query hint still reports its coverage, and then that line is all
 * this row has to say — so it becomes the label and there is nothing to open.
 */
const label = computed<string>(() => {
  if (active.value) return t('chatPanel.retrievalActive')
  if (hasQueries.value) return t('chatPanel.retrievalDone', { count: props.searches.length })
  return coverageLabel.value
})
</script>

<style scoped>
.retrieval__glyph :deep(path),
.retrieval__row-glyph :deep(path) {
  fill: currentColor;
}

.retrieval__list {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  margin: 0;
  padding: 0;
  list-style: none;
}

.retrieval__row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
  color: var(--color-text-secondary);
  font-size: 0.78rem;
  line-height: 1.6;
  animation: retrieval-row-in 0.32s cubic-bezier(0.23, 1, 0.32, 1) both;
}

.retrieval__row-glyph {
  flex-shrink: 0;
  font-size: 0.85rem;
  color: var(--color-text-muted);
  opacity: 0.8;
}

.retrieval__query {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
}

.retrieval__row--note {
  padding-left: 1.2rem;
  color: var(--color-text-muted);
  font-size: 0.74rem;
}

@keyframes retrieval-row-in {
  from {
    opacity: 0;
    transform: translateY(4px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .retrieval__row {
    animation: none;
  }
}
</style>
