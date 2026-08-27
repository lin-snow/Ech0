<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="jpc" :class="`jpc--${tone}`">
    <div class="jpc__header">
      <div class="jpc__title-wrap">
        <h3 class="jpc__title">{{ title }}</h3>
        <p v-if="subtitle" class="jpc__subtitle">{{ subtitle }}</p>
      </div>
      <span class="jpc__pill" :class="`jpc__pill--${tone}`">
        <span class="jpc__pill-dot" />
        {{ statusLabel }}
      </span>
    </div>

    <div v-if="steps.length" class="jpc__steps">
      <div v-for="(s, i) in steps" :key="s.key" class="jpc__step" :class="`is-${stepState(i)}`">
        <span class="jpc__node">
          <svg
            v-if="stepState(i) === 'done'"
            class="jpc__node-glyph"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              d="M5 13l4 4L19 7"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
          <svg
            v-else-if="stepState(i) === 'error'"
            class="jpc__node-glyph"
            viewBox="0 0 24 24"
            aria-hidden="true"
          >
            <path
              d="M6 6l12 12M18 6L6 18"
              fill="none"
              stroke="currentColor"
              stroke-width="3"
              stroke-linecap="round"
            />
          </svg>
          <template v-else>{{ i + 1 }}</template>
        </span>
        <span class="jpc__step-label">{{ s.label }}</span>
      </div>
    </div>

    <div
      class="jpc__bar"
      role="progressbar"
      :aria-valuenow="percent"
      aria-valuemin="0"
      aria-valuemax="100"
    >
      <div
        class="jpc__bar-fill"
        :class="{ 'is-active': isActive }"
        :style="{ width: percent + '%' }"
      />
    </div>

    <p v-if="errorMessage" class="jpc__error">{{ errorMessage }}</p>

    <div v-if="metrics && metrics.length" class="jpc__metrics">
      <div v-for="m in metrics" :key="m.label" class="jpc__metric">
        <span class="jpc__metric-label">{{ m.label }}</span>
        <span class="jpc__metric-value" :class="m.tone ? `is-${m.tone}` : ''">{{ m.value }}</span>
      </div>
    </div>

    <div v-if="$slots.footer" class="jpc__footer"><slot name="footer" /></div>

    <div v-if="meta && meta.length" class="jpc__meta">
      <p v-for="line in meta" :key="line.label">
        <span class="jpc__meta-label">{{ line.label }}</span>
        <span class="jpc__meta-value">{{ line.value }}</span>
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

type JobStatus = 'idle' | 'pending' | 'running' | 'success' | 'failed' | 'cancelled'

interface Step {
  key: string
  label: string
}
interface Metric {
  label: string
  value: string | number
  tone?: 'success' | 'danger'
}
interface MetaLine {
  label: string
  value: string
}

const props = defineProps<{
  title: string
  status: JobStatus
  statusLabel: string
  steps: Step[]
  currentKey?: string
  subtitle?: string
  metrics?: Metric[]
  meta?: MetaLine[]
  errorMessage?: string
}>()

const isActive = computed(() => props.status === 'pending' || props.status === 'running')
const isSuccess = computed(() => props.status === 'success')
const isError = computed(() => props.status === 'failed' || props.status === 'cancelled')

const tone = computed(() => {
  if (isSuccess.value) return 'success'
  if (isError.value) return 'error'
  if (isActive.value) return 'running'
  return 'idle'
})

const activeIndex = computed(() => {
  if (isSuccess.value) return props.steps.length - 1
  const i = props.steps.findIndex((s) => s.key === props.currentKey)
  if (i >= 0) return i
  return isActive.value ? 0 : -1
})

const percent = computed(() => {
  const n = props.steps.length || 1
  if (isSuccess.value) return 100
  if (props.status === 'idle' || activeIndex.value < 0) return 0
  return Math.min(100, Math.round(((activeIndex.value + 1) / n) * 100))
})

const stepState = (i: number): 'done' | 'active' | 'error' | 'pending' => {
  if (isSuccess.value) return 'done'
  if (i < activeIndex.value) return 'done'
  if (i === activeIndex.value) return isError.value ? 'error' : 'active'
  return 'pending'
}
</script>

<style scoped>
.jpc {
  --jpc-accent: var(--color-accent);
  --jpc-danger: var(--color-danger);
  --jpc-success: #15a06a;

  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  padding: 1rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

.jpc__header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 0.6rem;
}

.jpc__title {
  font-size: 0.95rem;
  font-weight: 700;
  line-height: 1.3;
  color: var(--color-text-primary);
}

.jpc__subtitle {
  margin-top: 0.2rem;
  font-size: 0.8rem;
  color: var(--color-text-secondary);
}

.jpc__pill {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 700;
  line-height: 1.2;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
}

.jpc__pill-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 999px;
  background: currentColor;
}

.jpc__pill--running {
  color: var(--jpc-accent);
  background: color-mix(in srgb, var(--jpc-accent) 10%, var(--color-bg-surface));
  border-color: color-mix(in srgb, var(--jpc-accent) 35%, transparent);
}

.jpc__pill--running .jpc__pill-dot {
  animation: jpc-pulse 1.2s ease-in-out infinite;
}

.jpc__pill--success {
  color: var(--jpc-success);
  background: color-mix(in srgb, var(--jpc-success) 10%, var(--color-bg-surface));
  border-color: color-mix(in srgb, var(--jpc-success) 35%, transparent);
}

.jpc__pill--error {
  color: var(--jpc-danger);
  background: color-mix(in srgb, var(--jpc-danger) 10%, var(--color-bg-surface));
  border-color: color-mix(in srgb, var(--jpc-danger) 35%, transparent);
}

.jpc__steps {
  display: flex;
  padding: 0 0.25rem;
}

.jpc__step {
  position: relative;
  flex: 1 1 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.4rem;
  text-align: center;
}

.jpc__step:not(:first-child)::before {
  content: '';
  position: absolute;
  top: calc(0.9rem - 1px);
  right: 50%;
  left: -50%;
  height: 2px;
  background: var(--color-border-subtle);
  z-index: 0;
  transition: background 0.3s ease;
}

.jpc__step.is-done::before,
.jpc__step.is-active::before {
  background: var(--jpc-accent);
}

.jpc__step.is-error::before {
  background: var(--jpc-danger);
}

.jpc__node {
  position: relative;
  z-index: 1;
  width: 1.8rem;
  height: 1.8rem;
  border-radius: 999px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.78rem;
  font-weight: 700;
  color: var(--color-text-muted);
  background: var(--color-bg-surface);
  border: 2px solid var(--color-border-strong);
  transition:
    color 0.25s ease,
    background 0.25s ease,
    border-color 0.25s ease;
}

.jpc__node-glyph {
  width: 1rem;
  height: 1rem;
}

.jpc__step.is-done .jpc__node {
  color: var(--color-bg-surface);
  background: var(--jpc-accent);
  border-color: var(--jpc-accent);
}

.jpc__step.is-active .jpc__node {
  color: var(--jpc-accent);
  background: color-mix(in srgb, var(--jpc-accent) 12%, var(--color-bg-surface));
  border-color: var(--jpc-accent);
  animation: jpc-ring 1.6s ease-out infinite;
}

.jpc__step.is-error .jpc__node {
  color: var(--jpc-danger);
  background: color-mix(in srgb, var(--jpc-danger) 12%, var(--color-bg-surface));
  border-color: var(--jpc-danger);
}

.jpc__step-label {
  max-width: 100%;
  font-size: 0.72rem;
  line-height: 1.2;
  color: var(--color-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.jpc__step.is-done .jpc__step-label,
.jpc__step.is-active .jpc__step-label {
  color: var(--color-text-secondary);
}

.jpc__bar {
  height: 0.45rem;
  border-radius: 999px;
  background: var(--color-bg-muted);
  overflow: hidden;
}

.jpc__bar-fill {
  height: 100%;
  border-radius: 999px;
  background-color: var(--jpc-accent);
  transition:
    width 0.45s cubic-bezier(0.4, 0, 0.2, 1),
    background-color 0.3s ease;
}

.jpc--success .jpc__bar-fill {
  background-color: var(--jpc-success);
}

.jpc--error .jpc__bar-fill {
  background-color: var(--jpc-danger);
}

.jpc__bar-fill.is-active {
  background-image: linear-gradient(
    90deg,
    transparent,
    color-mix(in srgb, #fff 45%, transparent),
    transparent
  );
  background-repeat: no-repeat;
  background-size: 35% 100%;
  animation: jpc-shimmer 1.3s linear infinite;
}

.jpc__error {
  font-size: 0.83rem;
  color: var(--jpc-danger);
  overflow-wrap: anywhere;
}

.jpc__metrics {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.55rem;
}

.jpc__metric {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  padding: 0.5rem 0.6rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-sm);
  background: var(--color-bg-muted);
}

.jpc__metric-label {
  font-size: 0.72rem;
  color: var(--color-text-muted);
}

.jpc__metric-value {
  font-size: 1.05rem;
  font-weight: 700;
  color: var(--color-text-primary);
  font-variant-numeric: tabular-nums;
}

.jpc__metric-value.is-success {
  color: var(--jpc-success);
}

.jpc__metric-value.is-danger {
  color: var(--jpc-danger);
}

.jpc__footer {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
}

.jpc__meta {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  padding-top: 0.6rem;
  border-top: 1px dashed var(--color-border-subtle);
}

.jpc__meta p {
  display: flex;
  gap: 0.5rem;
  font-size: 0.76rem;
}

.jpc__meta-label {
  flex-shrink: 0;
  color: var(--color-text-muted);
}

.jpc__meta-value {
  min-width: 0;
  color: var(--color-text-secondary);
  font-family: var(--font-family-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@keyframes jpc-pulse {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.35;
  }
}

@keyframes jpc-ring {
  0% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--jpc-accent) 45%, transparent);
  }

  70% {
    box-shadow: 0 0 0 0.32rem transparent;
  }

  100% {
    box-shadow: 0 0 0 0 transparent;
  }
}

@keyframes jpc-shimmer {
  from {
    background-position: -35% 0;
  }

  to {
    background-position: 135% 0;
  }
}

@media (width <= 640px) {
  .jpc__metrics {
    grid-template-columns: 1fr;
  }
}

@media (prefers-reduced-motion: reduce) {
  .jpc__pill--running .jpc__pill-dot,
  .jpc__step.is-active .jpc__node,
  .jpc__bar-fill.is-active {
    animation: none;
  }
}
</style>
