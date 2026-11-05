<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="askdone">
    <span class="askdone__status" aria-hidden="true">
      <svg viewBox="0 0 24 24" fill="none">
        <path
          class="askdone__check"
          d="M5 12.5 10 17.5 19 7"
          pathLength="1"
          stroke="currentColor"
          stroke-width="2.2"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
      </svg>
    </span>

    <div class="askdone__rows">
      <div v-for="(row, ri) in rows" :key="ri" class="askdone__row" :title="row.title">
        <span class="askdone__question">{{ row.question }}</span>
        <span class="askdone__answer">{{ row.answer }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  exchange: App.Api.Chat.ChatAskExchange
}>()

type Row = { question: string; answer: string; title: string }

/**
 * A settled round, replayed from what was sent: the question and the words the
 * reader picked or typed. Nothing here is interactive — the decision is made.
 */
const rows = computed<Row[]>(() => {
  const out: Row[] = []
  for (const q of props.exchange.questions) {
    const answer = props.exchange.answers.find((a) => a.question_id === q.id)
    if (!answer) continue
    const picked = answer.selected ?? []
    const text = picked.length > 0 ? picked.join(' · ') : (answer.custom ?? '')
    if (text.length === 0) continue
    const question = q.header && q.header.length > 0 ? q.header : q.text
    out.push({ question, answer: text, title: `${question} — ${text}` })
  }
  return out
})
</script>

<style scoped>
/* The same card surface as the pending ask, one weight down: the check is the
   only saturated mark, so the pair reads as "this was a decision, and it is
   closed". */
.askdone {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  width: 100%;
  max-width: 30rem;
  min-width: 0;
  margin: 0.5rem 0 0.3rem;
  padding: 0.55rem 0.8rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.9rem;
  background: color-mix(in srgb, var(--color-bg-muted) 70%, transparent);
}

.askdone__status {
  display: grid;
  place-items: center;
  width: 1.1rem;
  height: 1.1rem;
  flex-shrink: 0;
  margin-top: 0.08rem;
  color: var(--color-accent);
}

.askdone__check {
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: askdone-check 0.5s cubic-bezier(0.22, 1, 0.36, 1) 0.06s forwards;
}

@keyframes askdone-check {
  to {
    stroke-dashoffset: 0;
  }
}

.askdone__rows {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
  flex: 1;
}

/* Wrapping, not truncating. This is the record of a decision someone made —
   including the irreversible ones — so it stays readable at any width. */
.askdone__row {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 0.15rem 0.4rem;
  min-width: 0;
  max-width: 100%;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  line-height: 1.6;
}

.askdone__question {
  min-width: 0;
  overflow-wrap: anywhere;
}

/* The badge is the punctuation: it is the only thing here in a different
   colour, so it needs no tick and no separator to be found. */
.askdone__answer {
  min-width: 0;
  padding: 0 0.4rem;
  border-radius: 0.45rem;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

@media (prefers-reduced-motion: reduce) {
  .askdone__check {
    animation: none;
    stroke-dashoffset: 0;
  }
}
</style>
