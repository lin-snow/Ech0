<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="ask" role="group" :aria-label="t('chatPanel.askGroupLabel')">
    <div v-if="question.header || total > 1" class="ask__head">
      <span v-if="question.header" class="ask__eyebrow">{{ question.header }}</span>
      <span v-if="total > 1" class="ask__step">
        {{ t('chatPanel.askProgress', { current: index + 1, total }) }}
      </span>
    </div>

    <p class="ask__text">{{ question.text }}</p>

    <pre v-if="question.detail" class="ask__detail">{{ question.detail }}</pre>

    <div v-if="options.length > 0" class="ask__options" :class="{ 'ask__options--list': listMode }">
      <button
        v-for="(opt, oi) in options"
        :key="oi"
        type="button"
        class="ask__option"
        :class="{ 'ask__option--picked': picked.has(opt.label) }"
        :style="{ '--i': oi }"
        :disabled="pending"
        :aria-pressed="question.multi === true ? picked.has(opt.label) : undefined"
        @click="choose(opt.label)"
      >
        <span class="ask__option-head">
          <span
            v-if="oi === question.recommended"
            class="ask__tip"
            role="img"
            :aria-label="t('chatPanel.askRecommended')"
            :title="t('chatPanel.askRecommended')"
          />
          <span class="ask__option-label">{{ opt.label }}</span>
          <svg
            v-if="question.multi === true && picked.has(opt.label)"
            class="ask__check"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="3"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="m5 13 4 4L19 7" />
          </svg>
        </span>
        <span v-if="opt.description" class="ask__option-desc">{{ opt.description }}</span>
      </button>
    </div>

    <div v-if="question.multi === true || index > 0" class="ask__foot">
      <button
        v-if="question.multi === true"
        type="button"
        class="ask__submit"
        :disabled="pending || picked.size === 0"
        @click="commitMulti"
      >
        {{ pending ? t('chatPanel.askSending') : t('chatPanel.askConfirm') }}
      </button>
      <button
        v-if="index > 0"
        type="button"
        class="ask__back"
        :disabled="pending"
        @click="emit('back')"
      >
        {{ t('chatPanel.askBack') }}
      </button>
    </div>

    <p class="ask__hint">{{ t('chatPanel.askTypeHint') }}</p>

    <p v-if="error" class="ask__error" role="alert">{{ error }}</p>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const props = defineProps<{
  ask: App.Api.Chat.ChatAsk
  /** Cursor into `ask.questions`; the round is answered one question at a time. */
  index: number
  /** The answer for the whole round is in flight. */
  pending?: boolean
  error?: string
}>()

const emit = defineEmits<{
  (e: 'answer', answer: App.Api.Chat.ChatAskAnswer): void
  (e: 'back'): void
}>()

const { t } = useI18n()

const total = computed<number>(() => props.ask.questions.length)

const question = computed<App.Api.Chat.ChatAskQuestion>(
  () => props.ask.questions[props.index] ?? props.ask.questions[0],
)

const options = computed<App.Api.Chat.ChatAskOption[]>(() => question.value.options ?? [])

/**
 * Descriptions do two different jobs, and they want different shapes. When
 * several options carry one they are comparison material — three Echo previews
 * to read against each other — so the options become a column you scan. When at
 * most one does, it is a note about that single choice (「删除后无法恢复」), and
 * the options stay a compact row of buttons rather than letting the destructive
 * one balloon into the largest target on screen.
 */
const listMode = computed<boolean>(() => options.value.filter((o) => o.description).length > 1)

/**
 * Multi-select scratch space. `recommended` never seeds it: a mark the model
 * left is not a choice the reader made.
 */
const picked = ref<Set<string>>(new Set())

watch(
  () => [props.ask.ask_id, props.index] as const,
  () => {
    picked.value = new Set()
  },
)

const choose = (label: string) => {
  if (props.pending === true) return
  if (question.value.multi !== true) {
    emit('answer', { question_id: question.value.id, selected: [label], custom: '' })
    return
  }
  const next = new Set(picked.value)
  if (next.has(label)) next.delete(label)
  else next.add(label)
  picked.value = next
}

const commitMulti = () => {
  if (props.pending === true || picked.value.size === 0) return
  // Ordered by the option list, not by click order: the payload mirrors what the
  // reader saw rather than the path they took through it.
  const selected = options.value.map((o) => o.label).filter((label) => picked.value.has(label))
  emit('answer', { question_id: question.value.id, selected, custom: '' })
}
</script>

<style scoped>
/* The transcript has no cards in it: traces, sources and actions are all plain
   text hanging off the left edge. So an ask is not a dialog dropped into the
   conversation either — it is the same hanging indent the expanded traces use,
   with the rule promoted to accent to say this one is waiting for someone. */
.ask {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  width: 100%;
  min-width: 0;
  margin: 0.5rem 0 0.3rem;
  padding-left: 0.85rem;
}

.ask::before {
  content: '';
  position: absolute;
  top: 0.1rem;
  bottom: 0.1rem;
  left: 0;
  width: 2px;
  border-radius: 1px;
  background: var(--color-accent);
  transform: scaleY(0);
  transform-origin: top;
  animation: ask-rule 0.42s cubic-bezier(0.23, 1, 0.32, 1) both;
}

.ask__head {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  min-width: 0;
}

/* The header names which decision this is — 删除确认, 修改确认 — so it is a
   category label, and an eyebrow is what a category label looks like. */
.ask__eyebrow {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: var(--color-text-muted);
  font-size: 0.68rem;
  letter-spacing: 0.06em;
  line-height: 1.5;
  text-transform: uppercase;
}

.ask__step {
  flex-shrink: 0;
  margin-left: auto;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: 0.68rem;
  font-variant-numeric: tabular-nums;
}

.ask__text {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.94rem;
  line-height: 1.65;
}

/* Preformatted, not code: what the backend puts here is the Echo's own prose
   under 内容：, and a mono face turns someone's writing into log output. The
   line breaks matter, the typewriter does not.

   No rule of its own either — a second vertical line 0.85rem inside the accent
   one just draws a bracket. Size and colour already separate the evidence from
   the question above it. */
.ask__detail {
  margin: 0.15rem 0 0.2rem;
  max-height: 9rem;
  overflow: auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  color: var(--color-text-secondary);
  font-family: inherit;
  font-size: 0.82rem;
  line-height: 1.75;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.ask__options {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-top: 0.2rem;
}

/* Squarer than the pills elsewhere in the transcript, because these hold two
   lines when an option explains itself, and because this is the one place in a
   conversation that is actually a form. */
.ask__option {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 0.1rem;
  max-width: min(100%, 16rem);
  min-width: 0;
  padding: 0.3rem 0.7rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.6rem;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.84rem;
  line-height: 1.55;
  text-align: left;
  cursor: pointer;
  animation: ask-option-in 0.34s cubic-bezier(0.23, 1, 0.32, 1) both;
  animation-delay: calc(120ms + var(--i) * 45ms);
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
}

/* Comparison material reads down the page, one row each, on a measure short
   enough that the eye finds the next label without travelling. */
.ask__options--list .ask__option {
  flex-basis: 100%;
  max-width: min(100%, 26rem);
}

.ask__option:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.ask__option:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}

.ask__option:disabled {
  cursor: default;
  opacity: 0.55;
}

/* Selection has to be legible next to hover, or a multi-select cannot be read:
   hover moves the outline and the text to accent, so selection needs a channel
   hover does not use — the fill, plus a check that only exists when chosen. */
.ask__option--picked {
  border-color: var(--color-accent);
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
}

.ask__option-head {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
}

.ask__option-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* A dot, not a star: the model's suggestion is a hint worth one quiet channel,
   and a dashed border would read as "unavailable" on the very option it means
   to point at. */
.ask__tip {
  flex-shrink: 0;
  width: 0.3rem;
  height: 0.3rem;
  border-radius: 999px;
  background: var(--color-accent);
}

.ask__check {
  width: 0.7rem;
  height: 0.7rem;
  flex-shrink: 0;
  color: var(--color-accent);
}

.ask__option-desc {
  min-width: 0;
  color: var(--color-text-muted);
  font-size: 0.74rem;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.ask__option--picked .ask__option-desc {
  color: var(--color-text-secondary);
}

.ask__foot {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem 0.7rem;
  margin-top: 0.25rem;
}

.ask__submit {
  padding: 0.24rem 0.75rem;
  border: 1px solid var(--color-accent);
  border-radius: 0.6rem;
  background: transparent;
  color: var(--color-accent);
  font-size: 0.82rem;
  line-height: 1.5;
  cursor: pointer;
  transition:
    background 0.18s ease,
    opacity 0.18s ease;
}

.ask__submit:hover:not(:disabled) {
  background: var(--color-accent-soft);
}

.ask__submit:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}

.ask__submit:disabled {
  cursor: default;
  opacity: 0.4;
}

.ask__back {
  border: none;
  background: transparent;
  padding: 0;
  color: var(--color-text-secondary);
  font-size: 0.78rem;
  line-height: 1.5;
  cursor: pointer;
  transition: color 0.18s ease;
}

.ask__back:hover:not(:disabled) {
  color: var(--color-accent);
}

.ask__back:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}

.ask__back:disabled {
  cursor: default;
  opacity: 0.5;
}

/* Its own line: this is a standing note about the input box, not a third
   control competing with the buttons for the same row. */
.ask__hint {
  margin: 0.2rem 0 0;
  min-width: 0;
  color: var(--color-text-muted);
  font-size: 0.72rem;
  line-height: 1.5;
}

.ask__error {
  margin: 0.1rem 0 0;
  color: var(--color-danger);
  font-size: 0.76rem;
  line-height: 1.5;
}

@keyframes ask-rule {
  from {
    transform: scaleY(0);
  }

  to {
    transform: scaleY(1);
  }
}

@keyframes ask-option-in {
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
  .ask::before,
  .ask__option {
    animation: none;
  }

  .ask::before {
    transform: none;
  }

  .ask__option {
    transition: none;
  }
}
</style>
