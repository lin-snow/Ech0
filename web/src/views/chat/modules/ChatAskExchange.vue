<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="askdone">
    <div v-for="(row, ri) in rows" :key="ri" class="askdone__row" :title="row.title">
      <span class="askdone__question">{{ row.question }}</span>
      <span class="askdone__answer">{{ row.answer }}</span>
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
/* The same hanging indent as the pending ask, one weight down: a hairline in
   the neutral border colour instead of the accent rule. The shape says "this
   was an ask", the weight says "it is settled". */
.askdone {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
  max-width: 100%;
  margin: 0.4rem 0 0.25rem;
  padding-left: 0.85rem;
  border-left: 1px solid var(--color-border-subtle);
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
</style>
