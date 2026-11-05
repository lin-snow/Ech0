<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <ChatActivity :label="label" :active="active" :expandable="hasQueries" :detail="detail">
    <template #glyph>
      <!-- Running: a globe and its coverage ring in one coordinate system, so
           the two can never drift apart. Done: the ring hands the slot to the
           check, so the row itself records how the search ended. -->
      <svg v-if="active" class="retrieval__globe" viewBox="0 0 24 24" aria-hidden="true">
        <g transform="rotate(-90 12 12)">
          <circle class="retrieval__ring-track" cx="12" cy="12" r="10.4" pathLength="1" />
          <circle
            class="retrieval__ring-fill"
            cx="12"
            cy="12"
            r="10.4"
            pathLength="1"
            :style="{ strokeDasharray: `${progress} 1` }"
          />
        </g>
        <g class="retrieval__land" fill="none" stroke="currentColor" stroke-linecap="round">
          <circle cx="12" cy="12" r="7.1" stroke-width="1.7" />
          <path
            d="M4.9 12h14.2M12 4.9c2.1 2.3 3.15 4.66 3.15 7.1S14.1 16.8 12 19.1c-2.1-2.3-3.15-4.66-3.15-7.1S9.9 7.2 12 4.9"
            stroke-width="1.5"
          />
        </g>
      </svg>
      <ChatDoneGlyph v-else />
    </template>

    <template #trailing>
      <span v-if="stack.length" class="retrieval__stack" aria-hidden="true">
        <span
          v-for="(src, i) in stack"
          :key="src.echo_id"
          class="retrieval__tile"
          :style="{ zIndex: 10 - i }"
          >{{ i + 1 }}</span
        >
        <span v-if="extra" class="retrieval__tile retrieval__tile--extra">+{{ extra }}</span>
      </span>
    </template>

    <ul class="retrieval__list">
      <li
        v-for="(query, i) in searches"
        :key="query"
        class="retrieval__row"
        :style="{ animationDelay: `${i * 60}ms` }"
      >
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
import ChatDoneGlyph from './ChatDoneGlyph.vue'

const props = defineProps<{
  searches: string[]
  coverage?: App.Api.Chat.ChatCoverage
  active?: boolean
  /** What the run has pulled in so far, for the counting stack beside the pill. */
  sources?: App.Api.Chat.ChatSource[]
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

/** The latest query: enough to say what is being read without a second line. */
const detail = computed<string>(() => {
  const last = props.searches[props.searches.length - 1]
  return last ? `“${last}”` : ''
})

const STACK_LIMIT = 3

const stack = computed<App.Api.Chat.ChatSource[]>(() => (props.sources ?? []).slice(0, STACK_LIMIT))

const extra = computed<number>(() => Math.max(0, (props.sources?.length ?? 0) - STACK_LIMIT))

/**
 * The ring reports how much of the searchable range came back, not how many
 * queries ran: coverage is the honest denominator, and before it arrives the
 * source count fills the ring asymptotically so it never claims completion.
 */
const progress = computed<number>(() => {
  const coverage = props.coverage
  if (coverage && coverage.total > 0) {
    return Math.min(1, Math.max(0.08, coverage.returned / coverage.total))
  }
  const count = props.sources?.length ?? 0
  return Math.min(1, Math.max(0.08, 1 - Math.pow(0.7, count)))
})
</script>

<style scoped>
.retrieval__globe {
  display: inline-block;
  width: 1.05em;
  height: 1.05em;
  overflow: visible;
  vertical-align: -0.16em;
}

.retrieval__ring-track {
  fill: none;
  stroke: var(--color-border-strong);
  stroke-width: 1.4;
  opacity: 0.45;
}

.retrieval__ring-fill {
  fill: none;
  stroke: var(--color-accent);
  stroke-width: 1.4;
  stroke-dasharray: 0 1;
  transition: stroke-dasharray 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

/* Numbered rather than initialed: the source list has no favicon, and the index
   is what the expanded body repeats. */
.retrieval__stack {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
}

.retrieval__tile {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.15rem;
  height: 1.15rem;
  border-radius: 999px;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  font-size: 0.6rem;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  line-height: 1;
  transition: margin-left 0.28s cubic-bezier(0.22, 1, 0.36, 1);
}

.retrieval__tile:not(:first-child) {
  margin-left: -0.35rem;
}

/* Hovering the stack spreads it: the icons slide apart along their own
   margins, so nothing outside the pill reflows. */
.retrieval__stack:hover .retrieval__tile:not(:first-child) {
  margin-left: 0.08rem;
}

.retrieval__tile--extra {
  color: var(--color-text-muted);
  background: color-mix(in srgb, var(--color-accent-soft) 55%, transparent);
}

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

  .retrieval__ring-fill,
  .retrieval__tile {
    transition: none;
  }
}
</style>
