<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue'
import type { ComponentPublicInstance } from 'vue'

type OtpStatus = 'idle' | 'success' | 'error'

const props = withDefaults(
  defineProps<{
    modelValue: string
    /** Cell count; the code is modelled as one contiguous string of this length. */
    length?: number
    label?: string
    status?: OtpStatus
    disabled?: boolean
    autofocus?: boolean
    /** Accessible name for a single cell — the caller owns the i18n string. */
    slotLabel?: (index: number, total: number) => string
  }>(),
  { length: 8, label: '', status: 'idle', disabled: false, autofocus: false },
)

const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

/** Codes are A-Z / 0-9; everything else is dropped before it reaches the model. */
const CODE_CHAR = /[A-Z0-9]/

const uid = useId()
const labelId = `${uid}-label`

/** Codes are uppercase letters/digits, so a parent may hand us `ABCD-EFGH` verbatim. */
function toSlots(raw: string) {
  const chars = raw.toUpperCase().replace(/[^A-Z0-9]/g, '')
  return Array.from({ length: props.length }, (_, index) => chars[index] ?? '')
}

/**
 * Cells own the display state: a mid-code delete leaves a gap instead of shifting every
 * following digit into a fresh entry animation. The emitted value stays contiguous, and
 * `emitted` lets the echo of our own update be ignored so the gap survives it.
 */
const slots = ref<string[]>(toSlots(props.modelValue))
let emitted = ''

watch(
  () => props.modelValue,
  (value) => {
    if (value === emitted) return
    slots.value = toSlots(value)
  },
)
/** Even-length codes breathe after the first half, mirroring the printed `ABCD-EFGH`. */
const splitAfter = computed(() => (props.length % 2 === 0 ? props.length / 2 - 1 : -1))

const cleared = ref(false)
const focused = ref<number | null>(null)
const caretX = ref(0)
const rowEl = ref<HTMLElement | null>(null)
const cellEls: (HTMLElement | null)[] = []
const inputEls: (HTMLInputElement | null)[] = []
/** The slot the user navigated to on purpose, so a full code only changes deliberately. */
let editingAt: number | null = null
let resizeObserver: ResizeObserver | undefined

const caretVisible = computed(() => {
  const index = focused.value
  return index !== null && !slots.value[index]
})

function inputId(index: number) {
  return `${uid}-${index}`
}

function setInput(el: Element | ComponentPublicInstance | null, index: number) {
  inputEls[index] = (el as HTMLInputElement | null) ?? null
}

function setCell(el: Element | ComponentPublicInstance | null, index: number) {
  cellEls[index] = (el as HTMLElement | null) ?? null
}

function commit(next: string[]) {
  slots.value = next
  const code = next.join('')
  emitted = code
  emit('update:modelValue', code)
}

function focusAt(index: number) {
  const el = inputEls[Math.min(Math.max(index, 0), props.length - 1)]
  el?.focus()
  el?.select()
}

function setCharAt(index: number, char: string) {
  cleared.value = char === ''
  commit(slots.value.map((slot, i) => (i === index ? char : slot)))
}

/** Writes a run of characters (paste / autofill) starting at `index`. */
function fill(index: number, chars: string[]) {
  const room = Math.min(chars.length, props.length - index)
  const next = [...slots.value]
  chars.slice(0, room).forEach((char, offset) => {
    next[index + offset] = char
  })
  cleared.value = false
  commit(next)
  editingAt = null
  focusAt(index + room)
}

/** The input is uncontrolled between commits, so rejected keystrokes need a manual reset. */
function syncInput(index: number) {
  const el = inputEls[index]
  if (el) el.value = slots.value[index] ?? ''
}

function onInput(index: number, event: Event) {
  const chars = (event.target as HTMLInputElement).value
    .toUpperCase()
    .split('')
    .filter((char) => CODE_CHAR.test(char))

  if (chars.length === 0) {
    syncInput(index)
    return
  }

  // Typing into a filled slot appends, so keep only the new character.
  const typed =
    chars.length === 1
      ? chars[0]
      : chars.length === 2 && chars[0] === slots.value[index]
        ? chars[1]
        : null

  if (typed === null) {
    fill(index, chars)
    return
  }

  if (slots.value.every(Boolean) && editingAt !== index) {
    syncInput(index)
    return
  }

  setCharAt(index, typed)
  editingAt = null
  focusAt(index + 1)
}

function onKeyDown(index: number, event: KeyboardEvent) {
  switch (event.key) {
    case 'ArrowLeft':
      editingAt = Math.max(index - 1, 0)
      focusAt(index - 1)
      break
    case 'ArrowRight':
      editingAt = Math.min(index + 1, props.length - 1)
      focusAt(index + 1)
      break
    case 'Backspace':
      if (slots.value[index]) {
        setCharAt(index, '')
      } else if (index > 0) {
        setCharAt(index - 1, '')
        focusAt(index - 1)
      }
      break
    case 'Delete':
      if (!slots.value[index]) return
      setCharAt(index, '')
      break
    default:
      return
  }
  event.preventDefault()
}

function onPaste(index: number, event: ClipboardEvent) {
  event.preventDefault()
  const pasted = (event.clipboardData?.getData('text') ?? '')
    .toUpperCase()
    .split('')
    .filter((char) => CODE_CHAR.test(char))
  if (pasted.length) fill(index, pasted)
}

/** Clicking past the first gap lands on the gap, so the code stays contiguous. */
function onPointerDown(index: number, event: PointerEvent) {
  const firstEmpty = slots.value.findIndex((slot) => !slot)
  const target = firstEmpty === -1 ? index : Math.min(index, firstEmpty)
  editingAt = target
  if (target === index) return
  event.preventDefault()
  focusAt(target)
}

function onFocus(event: FocusEvent) {
  ;(event.target as HTMLInputElement).select()
}

function onFocusIn(event: FocusEvent) {
  const index = inputEls.indexOf(event.target as HTMLInputElement)
  if (index >= 0) focused.value = index
}

function onFocusOut(event: FocusEvent) {
  const next = event.relatedTarget as Node | null
  if (next && rowEl.value?.contains(next)) return
  focused.value = null
}

function measureCaret() {
  const index = focused.value
  if (index === null) return
  const cell = cellEls[index]
  if (!cell) return
  caretX.value = cell.offsetLeft + cell.offsetWidth / 2
}

watch([focused, slots], () => void nextTick(measureCaret), { flush: 'post' })

onMounted(() => {
  if (rowEl.value && typeof ResizeObserver !== 'undefined') {
    resizeObserver = new ResizeObserver(() => measureCaret())
    resizeObserver.observe(rowEl.value)
  }
  if (props.autofocus && !props.disabled) focusAt(0)
})

onBeforeUnmount(() => resizeObserver?.disconnect())
</script>

<template>
  <div class="base-otp" :class="{ 'base-otp--disabled': disabled }">
    <span v-if="label" :id="labelId" class="base-otp__label">{{ label }}</span>
    <div
      ref="rowEl"
      class="base-otp__row"
      :class="{
        'base-otp__row--clearing': cleared,
        'base-otp__row--success': status === 'success',
        'base-otp__row--error': status === 'error',
      }"
      role="group"
      :aria-labelledby="label ? labelId : undefined"
      @focusin="onFocusIn"
      @focusout="onFocusOut"
    >
      <div
        v-for="(char, index) in slots"
        :key="index"
        :ref="(el) => setCell(el, index)"
        class="base-otp__cell"
        :class="{
          'base-otp__cell--split': index === splitAfter,
          'base-otp__cell--filled': Boolean(char),
        }"
      >
        <input
          :id="inputId(index)"
          :ref="(el) => setInput(el, index)"
          class="base-otp__input"
          type="text"
          inputmode="text"
          autocapitalize="characters"
          :autocomplete="index === 0 ? 'one-time-code' : 'off'"
          spellcheck="false"
          :value="char"
          :disabled="disabled"
          :aria-label="slotLabel ? slotLabel(index, length) : undefined"
          @input="onInput(index, $event)"
          @keydown="onKeyDown(index, $event)"
          @paste="onPaste(index, $event)"
          @pointerdown="onPointerDown(index, $event)"
          @focus="onFocus"
        />
        <span class="base-otp__glyph" aria-hidden="true">
          <Transition name="otp-roll">
            <span v-if="char" :key="char" class="base-otp__char">{{ char }}</span>
          </Transition>
        </span>
        <svg
          v-if="status === 'success'"
          class="base-otp__ring"
          viewBox="0 0 40 50"
          aria-hidden="true"
        >
          <rect
            x="1"
            y="1"
            width="38"
            height="48"
            rx="11"
            pathLength="1"
            :style="{ '--otp-index': index }"
          />
        </svg>
      </div>
      <span
        v-if="caretVisible"
        class="base-otp__caret"
        :style="{ transform: `translateX(${caretX - 1}px)` }"
        aria-hidden="true"
      />
    </div>
  </div>
</template>

<style scoped>
.base-otp {
  display: flex;
  flex-direction: column;
}

.base-otp__label {
  display: block;
  margin-bottom: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.base-otp__row {
  position: relative;
  display: flex;
  gap: 0.375rem;
}

.base-otp__cell {
  position: relative;
  flex: 1 1 0;
  min-width: 0;
  aspect-ratio: 4 / 5;
  border: 1px solid var(--input-border-color);
  border-radius: var(--radius-md);
  background: var(--input-bg-color);
  box-shadow: var(--shadow-sm);
  transition:
    border-color 160ms ease,
    box-shadow 160ms ease;
}

.base-otp__cell--split {
  margin-left: 0.25rem;
}

.base-otp__cell--filled {
  border-color: var(--color-border-strong);
}

.base-otp__cell:focus-within {
  border-color: var(--color-accent);
  box-shadow: 0 0 0 3px var(--input-focus-ring-color);
}

.base-otp--disabled .base-otp__cell {
  opacity: 0.55;
}

.base-otp--disabled .base-otp__input {
  cursor: not-allowed;
}

.base-otp__row--error .base-otp__cell {
  border-color: var(--color-danger);
}

/* The cell keeps its focus affordance, tinted so it never reads as "valid again". */
.base-otp__row--error .base-otp__cell:focus-within {
  border-color: var(--color-danger);
  box-shadow: 0 0 0 3px color-mix(in oklab, var(--color-danger) 28%, transparent);
}

/* Success drops the CSS outline: the drawn SVG ring is the only outline left. */
.base-otp__row--success .base-otp__cell,
.base-otp__row--success .base-otp__cell:focus-within {
  border-color: transparent;
  box-shadow: none;
}

/* 32px / 0.32s: a short recoil, not a wobble. Re-triggered each time the status flips. */
.base-otp__row--error {
  animation: otp-shake 320ms ease-out;
}

.base-otp__input {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  padding: 0;
  border: 0;
  border-radius: inherit;
  outline: none;
  background: transparent;
  color: transparent;
  caret-color: transparent;
  font-family: var(--font-family-mono);

  /* 16px keeps iOS from zooming the page when the cell is focused. */
  font-size: 16px;
  text-align: center;
  cursor: text;
}

.base-otp__input::selection {
  background: transparent;
}

.base-otp__glyph {
  position: absolute;
  inset: 0;
  overflow: hidden;
  border-radius: inherit;
  pointer-events: none;
}

/* Each glyph owns the whole cell, so the outgoing and incoming characters overlap in
   one line while they exchange places — and 110% of this box is a full cell of travel,
   which is what makes the outgoing one actually leave instead of parking mid-cell. */
.base-otp__char {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  font-size: 1.125rem;
  font-weight: 600;
  line-height: 1;
}

/* Slot-reel: the incoming glyph rises from below, the outgoing one leaves upward
   (downward when the user cleared the cell, so the motion points the way it went). */
.otp-roll-enter-active,
.otp-roll-leave-active {
  will-change: transform;
  transition: transform 200ms cubic-bezier(0.23, 1, 0.32, 1);
}

.otp-roll-enter-from {
  transform: translateY(110%);
}

.otp-roll-leave-to {
  transform: translateY(-110%);
}

.base-otp__row--clearing .otp-roll-leave-to {
  transform: translateY(110%);
}

.base-otp__caret {
  position: absolute;
  top: 50%;
  left: 0;
  width: 2px;
  height: 1.5rem;
  margin-top: -0.75rem;
  border-radius: 999px;
  background: var(--color-accent);
  pointer-events: none;
  transition: transform 220ms cubic-bezier(0.23, 1, 0.32, 1);
  animation: otp-caret-blink 1.1s linear infinite;
}

.base-otp__ring {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  overflow: visible;
  pointer-events: none;
}

/* pathLength="1" normalises the perimeter, so one dash covers one rounded rect. */
.base-otp__ring rect {
  fill: none;
  stroke: var(--color-success);
  stroke-width: 2;
  stroke-dasharray: 1;
  stroke-dashoffset: 1;
  animation: otp-ring-draw 320ms cubic-bezier(0.23, 1, 0.32, 1) forwards;
  animation-delay: calc(120ms + var(--otp-index) * 30ms);
}

@keyframes otp-shake {
  0% {
    transform: translateX(0);
  }

  25% {
    transform: translateX(-5px);
  }

  55% {
    transform: translateX(4px);
  }

  80% {
    transform: translateX(-2px);
  }

  100% {
    transform: translateX(0);
  }
}

@keyframes otp-caret-blink {
  0%,
  49.9% {
    opacity: 1;
  }

  50%,
  100% {
    opacity: 0;
  }
}

@keyframes otp-ring-draw {
  to {
    stroke-dashoffset: 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .base-otp__row--error {
    animation: none;
  }

  .otp-roll-enter-active,
  .otp-roll-leave-active {
    transition: opacity 120ms ease;
  }

  .otp-roll-enter-from,
  .otp-roll-leave-to,
  .base-otp__row--clearing .otp-roll-leave-to {
    transform: none;
    opacity: 0;
  }

  .base-otp__caret {
    opacity: 1;
    transition: none;
    animation: none;
  }

  .base-otp__ring rect {
    animation-duration: 1ms;
    animation-delay: 0ms;
  }
}
</style>
