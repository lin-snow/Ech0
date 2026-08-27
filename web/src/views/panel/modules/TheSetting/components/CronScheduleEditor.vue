<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="cron-editor">
    <div class="cron-editor__grid">
      <div class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.frequency') }}</label>
        <select v-model="frequency" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in frequencyOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'weekly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.weekday') }}</label>
        <select v-model.number="weekday" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in weekdayOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'monthly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.dayOfMonth') }}</label>
        <select v-model.number="monthday" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in monthdayOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div
        v-if="frequency === 'daily' || frequency === 'weekly' || frequency === 'monthly'"
        class="cron-editor__field"
      >
        <label class="cron-editor__label">{{ t('cronEditor.time') }}</label>
        <input
          v-model="timeString"
          type="time"
          :disabled="disabled"
          class="cron-editor__control cron-editor__control--time"
        />
      </div>

      <div v-if="frequency === 'hourly'" class="cron-editor__field">
        <label class="cron-editor__label">{{ t('cronEditor.everyInterval') }}</label>
        <select v-model.number="hourlyInterval" :disabled="disabled" class="cron-editor__control">
          <option v-for="opt in hourlyOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
      </div>

      <div v-if="frequency === 'custom'" class="cron-editor__field cron-editor__field--wide">
        <label class="cron-editor__label">{{ t('cronEditor.customExpression') }}</label>
        <input
          v-model="customExpression"
          type="text"
          :disabled="disabled"
          :placeholder="t('cronEditor.customPlaceholder')"
          class="cron-editor__control cron-editor__control--input"
          spellcheck="false"
          autocomplete="off"
        />
      </div>
    </div>

    <div class="cron-editor__summary">
      <CalendarIcon class="cron-editor__summary-icon" />
      <div class="cron-editor__summary-body">
        <span class="cron-editor__summary-text">{{ humanReadable }}</span>
        <code class="cron-editor__summary-code">{{ modelValue || '—' }}</code>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import CalendarIcon from '@/components/icons/date-icon.vue'
import { buildCron as buildCronExpr, humanizeCron, parseCron as parseCronExpr } from '@/utils/cron'
import type { CronFrequency } from '@/utils/cron'

type Frequency = CronFrequency

const props = defineProps<{
  modelValue: string
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', v: string): void
}>()

const { t } = useI18n()

const frequency = ref<Frequency>('daily')
const hour = ref<number>(2)
const minute = ref<number>(0)
const weekday = ref<number>(0)
const monthday = ref<number>(1)
const hourlyInterval = ref<number>(6)
const customExpression = ref<string>('')

const frequencyOptions = computed(() => [
  { label: t('cronEditor.freqDaily'), value: 'daily' as Frequency },
  { label: t('cronEditor.freqWeekly'), value: 'weekly' as Frequency },
  { label: t('cronEditor.freqMonthly'), value: 'monthly' as Frequency },
  { label: t('cronEditor.freqHourly'), value: 'hourly' as Frequency },
  { label: t('cronEditor.freqCustom'), value: 'custom' as Frequency },
])

const weekdayOptions = computed(() => [
  { label: t('cronEditor.weekdaySun'), value: 0 },
  { label: t('cronEditor.weekdayMon'), value: 1 },
  { label: t('cronEditor.weekdayTue'), value: 2 },
  { label: t('cronEditor.weekdayWed'), value: 3 },
  { label: t('cronEditor.weekdayThu'), value: 4 },
  { label: t('cronEditor.weekdayFri'), value: 5 },
  { label: t('cronEditor.weekdaySat'), value: 6 },
])

const monthdayOptions = computed(() =>
  Array.from({ length: 28 }, (_, i) => ({
    label: t('cronEditor.dayN', { n: i + 1 }),
    value: i + 1,
  })),
)

const hourlyOptions = computed(() =>
  [1, 2, 3, 4, 6, 8, 12].map((n) => ({
    label: t('cronEditor.everyNHours', { n }),
    value: n,
  })),
)

const timeString = computed<string>({
  get: () => `${pad(hour.value)}:${pad(minute.value)}`,
  set: (v: string) => {
    const m = /^(\d{1,2}):(\d{1,2})$/.exec(v || '')
    if (!m) return
    hour.value = clamp(parseInt(m[1], 10), 0, 23)
    minute.value = clamp(parseInt(m[2], 10), 0, 59)
  },
})

const humanReadable = computed(() => humanizeCron(buildCron(), t))

function pad(n: number): string {
  return String(n).padStart(2, '0')
}

function clamp(n: number, min: number, max: number): number {
  if (Number.isNaN(n)) return min
  return Math.min(Math.max(n, min), max)
}

function parseCron(expr: string) {
  const p = parseCronExpr(expr)
  frequency.value = p.frequency
  minute.value = p.minute
  hour.value = p.hour
  weekday.value = p.weekday
  monthday.value = p.monthday
  hourlyInterval.value = p.hourlyInterval
  customExpression.value = p.custom
}

function buildCron(): string {
  return buildCronExpr({
    frequency: frequency.value,
    minute: minute.value,
    hour: hour.value,
    weekday: weekday.value,
    monthday: monthday.value,
    hourlyInterval: hourlyInterval.value,
    custom: customExpression.value,
  })
}

watch(
  () => props.modelValue,
  (v) => {
    if (buildCron() === (v || '').trim()) return
    parseCron(v || '')
  },
  { immediate: true },
)

watch([frequency, hour, minute, weekday, monthday, hourlyInterval, customExpression], () => {
  const next = buildCron()
  if (next !== props.modelValue) {
    emit('update:modelValue', next)
  }
})
</script>

<style scoped>
.cron-editor {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
}

.cron-editor__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.6rem 0.75rem;
}

@media (width < 480px) {
  .cron-editor__grid {
    grid-template-columns: minmax(0, 1fr);
  }
}

.cron-editor__field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
  min-width: 0;
}

.cron-editor__field--wide {
  grid-column: 1 / -1;
}

.cron-editor__label {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.cron-editor__control {
  width: 100%;
  height: 2.1rem;
  padding: 0 0.65rem;
  border-radius: var(--radius-sm);
  border: 1px solid var(--input-border-color);
  background: var(--input-bg-color);
  color: var(--input-text-color);
  font-family: inherit;
  font-size: 0.85rem;
  line-height: 1;
  box-shadow: var(--shadow-sm);
  outline: none;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.cron-editor__control:hover:not(:disabled) {
  border-color: var(--color-border-strong);
}

.cron-editor__control:focus-visible {
  border-color: var(--input-focus-ring-color);
  box-shadow:
    0 0 0 2px var(--input-focus-ring-color),
    var(--shadow-sm);
}

.cron-editor__control:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

select.cron-editor__control {
  appearance: none;
  padding-right: 1.9rem;
  cursor: pointer;
  background-image:
    linear-gradient(45deg, transparent 50%, currentColor 50%),
    linear-gradient(135deg, currentColor 50%, transparent 50%);
  background-position:
    calc(100% - 1rem) calc(50% - 2px),
    calc(100% - 0.65rem) calc(50% - 2px);
  background-size:
    5px 5px,
    5px 5px;
  background-repeat: no-repeat;
  color: var(--input-text-color);
}

select.cron-editor__control::-ms-expand {
  display: none;
}

.cron-editor__control--time {
  font-family: var(--font-family-mono);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.02em;
}

.cron-editor__control--time::-webkit-date-and-time-value {
  text-align: left;
  min-height: 1.2em;
}

.cron-editor__control--input {
  font-family: var(--font-family-mono);
  font-size: 0.82rem;
}

.cron-editor__summary {
  display: flex;
  align-items: flex-start;
  gap: 0.55rem;
  padding: 0.55rem 0.7rem;
  border-radius: var(--radius-md);
  background: color-mix(in srgb, var(--color-accent, #e07020) 6%, var(--color-bg-muted));
  border: 1px solid color-mix(in srgb, var(--color-accent, #e07020) 22%, transparent);
}

.cron-editor__summary-icon {
  width: 1rem;
  height: 1rem;
  margin-top: 0.1rem;
  color: var(--color-accent, #e07020);
  flex-shrink: 0;
}

:deep(.cron-editor__summary-icon path) {
  fill: currentColor;
}

.cron-editor__summary-body {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  min-width: 0;
  flex: 1 1 auto;
}

.cron-editor__summary-text {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-text-primary);
  line-height: 1.35;
}

.cron-editor__summary-code {
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  color: var(--color-text-muted);
  letter-spacing: 0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
