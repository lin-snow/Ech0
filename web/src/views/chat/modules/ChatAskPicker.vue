<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="ask" role="group" :aria-label="t('chatPanel.askGroupLabel')" :aria-busy="pending">
    <div class="ask__head">
      <span class="ask__status" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none">
          <circle cx="12" cy="12" r="9" stroke="currentColor" stroke-width="1.7" />
          <path
            d="M9.6 9.4a2.5 2.5 0 1 1 3.6 2.3c-.75.4-1.2.95-1.2 1.75v.35"
            stroke="currentColor"
            stroke-width="1.7"
            stroke-linecap="round"
          />
          <circle cx="12" cy="17.1" r="1" fill="currentColor" />
        </svg>
      </span>

      <div class="ask__heading">
        <span v-if="question.header" class="ask__eyebrow">{{ question.header }}</span>
        <Transition name="ask-swap" mode="out-in">
          <p :key="question.id" class="ask__text">{{ question.text }}</p>
        </Transition>
      </div>

      <span v-if="total > 1" class="ask__step">{{ index + 1 }} / {{ total }}</span>
    </div>

    <pre v-if="question.detail" class="ask__detail">{{ question.detail }}</pre>

    <div
      v-if="options.length > 0"
      class="ask__options"
      :role="question.multi === true ? undefined : 'radiogroup'"
    >
      <button
        v-for="(opt, oi) in options"
        :key="oi"
        type="button"
        class="ask__option"
        :class="{
          'ask__option--picked': isPicked(opt.label),
          'ask__option--multi': question.multi === true,
        }"
        :style="{ '--i': oi }"
        :disabled="pending"
        :role="question.multi === true ? undefined : 'radio'"
        :aria-checked="isPicked(opt.label)"
        :aria-pressed="question.multi === true ? isPicked(opt.label) : undefined"
        @click="choose(opt.label)"
      >
        <span
          class="ask__mark"
          :class="{
            'ask__mark--on': isPicked(opt.label),
            'ask__mark--multi': question.multi === true,
          }"
          aria-hidden="true"
        >
          <svg
            v-if="question.multi === true"
            class="ask__mark-check"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="3.2"
            stroke-linecap="round"
            stroke-linejoin="round"
          >
            <path d="m5 13 4 4L19 7" />
          </svg>
          <span v-else class="ask__mark-dot" />
        </span>

        <span class="ask__option-body">
          <span class="ask__option-head">
            <span
              v-if="oi === question.recommended"
              class="ask__tip"
              role="img"
              :aria-label="t('chatPanel.askRecommended')"
              :title="t('chatPanel.askRecommended')"
            />
            <span class="ask__option-label">{{ opt.label }}</span>
          </span>
          <span v-if="opt.description" class="ask__option-desc">{{ opt.description }}</span>
        </span>
      </button>
    </div>

    <span v-if="total > 1" class="ask__dots">
      <span class="ask__sr-only">{{
        t('chatPanel.askProgress', { current: index + 1, total })
      }}</span>
      <span
        v-for="(q, qi) in ask.questions"
        :key="q.id"
        class="ask__dot"
        aria-hidden="true"
        :class="{
          'ask__dot--done': qi < index,
          'ask__dot--current': qi === index,
        }"
      />
    </span>

    <div v-if="question.multi === true || index > 0" class="ask__foot">
      <button
        v-if="question.multi === true"
        type="button"
        class="ask__submit"
        :disabled="pending || picked.size === 0"
        @click="commitMulti"
      >
        <span v-if="pending" class="ask__spinner" aria-hidden="true" />
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
import { computed, onBeforeUnmount, ref, watch } from 'vue'
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
 * Multi-select scratch space. `recommended` never seeds it: a mark the model
 * left is not a choice the reader made.
 */
const picked = ref<Set<string>>(new Set())

/**
 * A single choice is a decision, and the card should show it happened before
 * the question changes underneath it. The picked row lights up, then the
 * answer leaves a beat later — short enough to feel like the tap, long enough
 * to be seen.
 */
const pickedSingle = ref<string>('')
let advanceTimer = 0

const clearAdvance = () => {
  if (!advanceTimer) return
  clearTimeout(advanceTimer)
  advanceTimer = 0
}

const isPicked = (label: string): boolean =>
  question.value.multi === true ? picked.value.has(label) : pickedSingle.value === label

watch(
  () => [props.ask.ask_id, props.index] as const,
  () => {
    picked.value = new Set()
    pickedSingle.value = ''
    clearAdvance()
  },
)

const choose = (label: string) => {
  if (props.pending === true) return
  if (question.value.multi === true) {
    const next = new Set(picked.value)
    if (next.has(label)) next.delete(label)
    else next.add(label)
    picked.value = next
    return
  }
  if (advanceTimer) return
  pickedSingle.value = label
  advanceTimer = window.setTimeout(() => {
    advanceTimer = 0
    pickedSingle.value = ''
    emit('answer', { question_id: question.value.id, selected: [label], custom: '' })
  }, 220)
}

const commitMulti = () => {
  if (props.pending === true || picked.value.size === 0) return
  // Ordered by the option list, not by click order: the payload mirrors what the
  // reader saw rather than the path they took through it.
  const selected = options.value.map((o) => o.label).filter((label) => picked.value.has(label))
  emit('answer', { question_id: question.value.id, selected, custom: '' })
}

onBeforeUnmount(clearAdvance)
</script>

<style scoped>
/* The one card in the transcript: a decision is a different kind of object
   from a trace, so it gets a surface, a status glyph and a quiet shadow while
   everything around it stays hairlines on paper. */
.ask {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  width: 100%;
  max-width: 30rem;
  min-width: 0;
  margin: 0.55rem 0 0.35rem;
  padding: 0.8rem 0.9rem 0.7rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 1rem;
  background: var(--color-bg-muted);
  animation: ask-in 0.32s cubic-bezier(0.23, 1, 0.32, 1) both;
}

@keyframes ask-in {
  from {
    opacity: 0;
    transform: translateY(6px) scale(0.99);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

.ask__head {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  min-width: 0;
}

.ask__status {
  display: grid;
  place-items: center;
  width: 1.1rem;
  height: 1.1rem;
  flex-shrink: 0;
  margin-top: 0.08rem;
  color: var(--color-accent);
}

.ask__heading {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.12rem;
}

.ask__eyebrow {
  min-width: 0;
  overflow: hidden;
  white-space: nowrap;
  text-overflow: ellipsis;
  color: var(--color-text-muted);
  font-size: 0.64rem;
  letter-spacing: 0.07em;
  line-height: 1.5;
  text-transform: uppercase;
}

.ask__text {
  margin: 0;
  color: var(--color-text-primary);
  font-size: 0.92rem;
  font-weight: 550;
  line-height: 1.55;
  overflow-wrap: anywhere;
}

/* The question slides in from the side the flow is moving toward, so a
   multi-step round reads as progress instead of content shuffling. */
.ask-swap-enter-active,
.ask-swap-leave-active {
  transition:
    opacity 0.2s ease,
    transform 0.24s cubic-bezier(0.22, 1, 0.36, 1),
    filter 0.24s ease;
}

.ask-swap-enter-from {
  opacity: 0;
  transform: translateX(8px);
  filter: blur(3px);
}

.ask-swap-leave-to {
  opacity: 0;
  transform: translateX(-6px);
  filter: blur(3px);
}

.ask__step {
  flex-shrink: 0;
  margin-top: 0.12rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: 0.68rem;
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
}

/* Preformatted, not code: what the backend puts here is the Echo's own prose
   under 内容：, and a mono face turns someone's writing into log output. */
.ask__detail {
  margin: 0.05rem 0 0 1.65rem;
  max-height: 8rem;
  overflow: auto;
  overscroll-behavior: contain;
  scrollbar-width: thin;
  color: var(--color-text-secondary);
  font-family: inherit;
  font-size: 0.8rem;
  line-height: 1.7;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.ask__options {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  margin-top: 0.15rem;
}

/* One row per choice, with the control on the left: the shape a form takes
   when it is asking for a decision rather than offering shortcuts. */
.ask__option {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  width: 100%;
  min-width: 0;
  padding: 0.42rem 0.55rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.7rem;
  background: var(--color-bg-surface);
  color: var(--color-text-secondary);
  font-size: 0.84rem;
  line-height: 1.55;
  text-align: left;
  cursor: pointer;
  animation: ask-option-in 0.34s cubic-bezier(0.23, 1, 0.32, 1) both;
  animation-delay: calc(80ms + var(--i) * 35ms);
  transition:
    color 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
}

.ask__option:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-text-primary);
}

.ask__option:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}

.ask__option:disabled {
  cursor: default;
  opacity: 0.55;
}

.ask__option--picked {
  border-color: var(--color-accent);
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
}

.ask__mark {
  display: grid;
  place-items: center;
  width: 1rem;
  height: 1rem;
  flex-shrink: 0;
  margin-top: 0.16rem;
  border: 1.5px solid var(--color-border-strong);
  border-radius: 999px;
  transition:
    border-color 0.18s ease,
    background 0.18s ease;
}

.ask__mark--multi {
  border-radius: 0.3rem;
}

.ask__mark--on {
  border-color: var(--color-accent);
}

.ask__mark--multi.ask__mark--on {
  background: var(--color-accent);
  color: var(--dialog-confirm-text-color);
}

.ask__mark-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 999px;
  background: var(--color-accent);
  transform: scale(0);
  transition: transform 0.18s cubic-bezier(0.22, 1, 0.36, 1);
}

.ask__mark--on .ask__mark-dot {
  transform: scale(1);
}

.ask__mark-check {
  width: 0.7rem;
  height: 0.7rem;
  opacity: 0;
  transform: scale(0.6);
  transition:
    opacity 0.16s ease,
    transform 0.18s cubic-bezier(0.22, 1, 0.36, 1);
}

.ask__mark--on .ask__mark-check {
  opacity: 1;
  transform: scale(1);
}

.ask__option-body {
  display: flex;
  flex-direction: column;
  gap: 0.1rem;
  min-width: 0;
  flex: 1;
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

.ask__option-desc {
  min-width: 0;
  color: var(--color-text-muted);
  font-size: 0.73rem;
  line-height: 1.5;
  overflow-wrap: anywhere;
}

.ask__option--picked .ask__option-desc {
  color: var(--color-text-secondary);
}

/* Dots rather than a number alone: where the round is going is a shape the eye
   can read without counting. */
.ask__dots {
  position: relative;
  display: flex;
  gap: 0.3rem;
  margin: 0.1rem 0 0 1.65rem;
}

.ask__sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip-path: inset(50%);
  white-space: nowrap;
  border: 0;
}

.ask__dot {
  width: 0.3rem;
  height: 0.3rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.3;
  transition:
    transform 0.24s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.24s ease;
}

.ask__dot--done {
  opacity: 0.8;
}

.ask__dot--current {
  opacity: 1;
  background: var(--color-accent);
  transform: scale(1.25);
}

.ask__foot {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.35rem 0.7rem;
  margin-top: 0.15rem;
  margin-left: 1.65rem;
}

.ask__submit {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.32rem 0.85rem;
  border: none;
  border-radius: 999px;
  background: var(--dialog-confirm-btn-bg-color);
  color: var(--dialog-confirm-text-color);
  font-size: 0.8rem;
  line-height: 1.5;
  cursor: pointer;
  transition:
    opacity 0.18s ease,
    transform 0.18s ease;
}

.ask__submit:hover:not(:disabled) {
  transform: translateY(-1px);
  opacity: 0.92;
}

.ask__submit:active:not(:disabled) {
  transform: translateY(0);
}

.ask__submit:focus-visible {
  outline: 2px solid var(--color-focus-ring);
  outline-offset: 2px;
}

.ask__submit:disabled {
  cursor: default;
  opacity: 0.35;
}

.ask__spinner {
  width: 0.7rem;
  height: 0.7rem;
  border: 1.5px solid currentcolor;
  border-top-color: transparent;
  border-radius: 999px;
  animation: ask-spin 0.7s linear infinite;
}

@keyframes ask-spin {
  to {
    transform: rotate(360deg);
  }
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
  margin: 0 0 0 1.65rem;
  min-width: 0;
  color: var(--color-text-muted);
  font-size: 0.7rem;
  line-height: 1.5;
}

.ask__error {
  margin: 0.05rem 0 0 1.65rem;
  color: var(--color-danger);
  font-size: 0.76rem;
  line-height: 1.5;
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
  .ask,
  .ask__option {
    animation: none;
  }

  .ask__option,
  .ask-swap-enter-active,
  .ask-swap-leave-active,
  .ask__dot,
  .ask__mark,
  .ask__mark-dot,
  .ask__mark-check {
    transition: none;
  }

  .ask-swap-enter-from,
  .ask-swap-leave-to {
    transform: none;
    filter: none;
  }

  .ask__spinner {
    animation: none;
  }
}
</style>
