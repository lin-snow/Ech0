<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import BaseOtpInput from '@/components/common/BaseOtpInput.vue'

const THEMES = ['light', 'dark', 'sunny'] as const
const theme = ref<(typeof THEMES)[number]>('light')

const code = ref('')
const client = ref('ech0-cli · macOS')
const expiresIn = ref(300)
const lookupFailed = ref(false)
const approving = ref(false)
const succeeded = ref(false)
const approved = ref(false)

const prepared = computed(() => code.value.length === 8)
const formatted = computed(() =>
  prepared.value ? `${code.value.slice(0, 4)}-${code.value.slice(4)}` : code.value,
)
const status = computed<'idle' | 'success' | 'error'>(() => {
  if (succeeded.value) return 'success'
  return lookupFailed.value ? 'error' : 'idle'
})
const remaining = computed(
  () => `${Math.floor(expiresIn.value / 60)}:${String(expiresIn.value % 60).padStart(2, '0')}`,
)

let timer: ReturnType<typeof setTimeout> | undefined
const ticker = setInterval(() => {
  if (expiresIn.value > 0) expiresIn.value -= 1
}, 1000)

function setTheme(next: (typeof THEMES)[number]) {
  theme.value = next
  document.documentElement.className = next
}

function onCodeUpdate(value: string) {
  code.value = value
  lookupFailed.value = false
}

/** Same choreography as DeviceView: rings draw for ~650ms, then the receipt swaps in. */
function approve() {
  if (!prepared.value || approving.value) return
  approving.value = true
  timer = setTimeout(() => {
    approving.value = false
    succeeded.value = true
    timer = setTimeout(() => {
      approved.value = true
    }, 760)
  }, 450)
}

function reset() {
  clearTimeout(timer)
  code.value = ''
  lookupFailed.value = false
  approving.value = false
  succeeded.value = false
  approved.value = false
  expiresIn.value = 300
}

onBeforeUnmount(() => {
  clearTimeout(timer)
  clearInterval(ticker)
})
</script>

<template>
  <main class="flex justify-center items-start min-h-screen px-4 py-10">
    <div class="w-[min(92vw,25rem)]">
      <div class="flex justify-end gap-1.5">
        <button
          v-for="name in THEMES"
          :key="name"
          type="button"
          class="cursor-pointer rounded-full px-2.5 py-1 text-[11px] transition-colors"
          :class="
            theme === name
              ? 'bg-[var(--color-accent)] text-[var(--dialog-confirm-text-color)]'
              : 'bg-[var(--color-bg-muted)] text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
          "
          @click="setTheme(name)"
        >
          {{ name }}
        </button>
      </div>

      <header class="mt-6 flex flex-col items-center text-center">
        <span class="text-lg italic font-bold font-serif text-[var(--color-text-muted)]">Ech0</span>
      </header>

      <div
        class="mt-5 flex items-center gap-2.5 rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] px-3 py-2"
      >
        <span class="h-6 w-6 shrink-0 rounded-full bg-[var(--color-accent-soft)]" />
        <p class="flex min-w-0 flex-1 items-baseline gap-1.5 text-xs">
          <span class="shrink-0 text-[var(--color-text-muted)]">Signed in as</span>
          <span class="truncate font-medium text-[var(--color-text-primary)]">lin-snow</span>
        </p>
        <span class="shrink-0 text-xs text-[var(--color-accent)]">Switch account</span>
      </div>

      <div v-if="approved" class="mt-10 text-center">
        <p class="text-sm text-[var(--color-text-primary)]">
          Approved. Return to your terminal to continue.
        </p>
        <button
          type="button"
          class="mt-4 cursor-pointer rounded-[var(--radius-md)] border border-[var(--input-border-color)] bg-[var(--input-bg-color)] px-3 py-1.5 text-sm text-[var(--input-text-color)] hover:border-[var(--input-hover-border-color)]"
          @click="reset"
        >
          Run it again
        </button>
      </div>

      <div v-else class="mt-9">
        <BaseOtpInput
          :model-value="code"
          label="Device code"
          :length="8"
          :status="status"
          :disabled="approving || succeeded"
          autofocus
          :slot-label="(index, total) => `Character ${index + 1} of ${total}`"
          @update:model-value="onCodeUpdate"
        />
        <p v-if="lookupFailed" class="mt-2 text-xs text-[var(--color-danger)]">
          The device code is invalid or expired.
        </p>
        <p v-else class="mt-2 text-xs text-[var(--color-text-muted)]">
          Enter the 8-character code shown in your terminal.
        </p>

        <section v-if="prepared && !lookupFailed" class="mt-6 space-y-4">
          <div class="flex items-baseline gap-1.5 text-sm">
            <span class="shrink-0 text-[var(--color-text-muted)]">Client:</span>
            <span class="font-medium text-[var(--color-text-primary)] break-all">{{ client }}</span>
          </div>
          <div class="flex items-center justify-between text-xs">
            <span class="text-[var(--color-text-muted)]">Expires in</span>
            <span class="font-mono tabular-nums text-[var(--color-text-secondary)]">
              {{ remaining }}
            </span>
          </div>
          <button
            type="button"
            :disabled="approving || succeeded"
            class="w-full cursor-pointer rounded-[var(--radius-md)] bg-[var(--color-accent)] px-4 py-2.5 text-sm font-semibold text-[var(--dialog-confirm-text-color)] transition-opacity hover:opacity-90 disabled:cursor-not-allowed disabled:opacity-60"
            @click="approve"
          >
            {{ approving ? 'Approving…' : 'Approve' }}
          </button>
        </section>

        <div
          class="mt-8 flex flex-wrap items-center gap-2 border-t border-[var(--color-border-subtle)] pt-4"
        >
          <span class="text-[11px] text-[var(--color-text-muted)]">Dev:</span>
          <button
            type="button"
            class="cursor-pointer rounded-full bg-[var(--color-bg-muted)] px-2.5 py-1 text-[11px] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]"
            @click="lookupFailed = true"
          >
            invalid code
          </button>
          <button
            type="button"
            class="cursor-pointer rounded-full bg-[var(--color-bg-muted)] px-2.5 py-1 text-[11px] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)]"
            @click="reset"
          >
            reset
          </button>
          <span class="w-full font-mono text-[11px] text-[var(--color-text-muted)]">
            model = "{{ formatted }}" ({{ code.length }}/8)
          </span>
        </div>

        <p class="mt-6 text-xs leading-relaxed text-[var(--color-text-muted)]">
          Try: type 8 characters, paste <code class="font-mono">ABCD-EFGH</code>, delete a digit in
          the middle (it leaves a gap you can refill), press Backspace on an empty cell, click past
          the first gap, then flip the OS "reduce motion" switch.
        </p>
      </div>
    </div>
  </main>
</template>
