<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="visitor-widget">
    <WidgetShell :title="t('dashboard.visitorTitle')" meta="7 DAYS" :fluid="fluid">
      <div class="visitor-legend">
        <span class="legend-item">
          <span class="legend-dot legend-dot--pv" />
          {{ t('dashboard.visitorPv') }}
        </span>
        <span class="legend-item">
          <span class="legend-dot legend-dot--uv" />
          {{ t('dashboard.visitorUv') }}
        </span>
      </div>

      <div class="bar-chart-wrap">
        <svg
          v-if="last7Days.length >= 1"
          class="bar-chart-svg"
          :viewBox="`0 0 ${chartW} ${chartH}`"
          preserveAspectRatio="xMidYMid meet"
        >
          <line
            :x1="chartPadX"
            :y1="chartH - chartPadY"
            :x2="chartW - chartPadX"
            :y2="chartH - chartPadY"
            class="bar-chart-baseline"
          />
          <g v-for="(bar, i) in barGroups" :key="`group-${i}`">
            <rect
              :x="bar.pvX"
              :y="bar.pvY"
              :width="barWidth"
              :height="bar.pvHeight"
              rx="2"
              class="bar-rect bar-rect--pv"
            />
            <rect
              :x="bar.uvX"
              :y="bar.uvY"
              :width="barWidth"
              :height="bar.uvHeight"
              rx="2"
              class="bar-rect bar-rect--uv"
            />
            <rect
              :x="bar.hitX"
              :y="chartPadY"
              :width="bar.hitW"
              :height="chartH - chartPadY * 2"
              class="bar-hit-area"
              @mouseenter="showTooltip(i, $event)"
              @mouseleave="hideTooltip"
            />
          </g>
        </svg>
        <div v-else class="visitor-empty">{{ t('dashboard.visitorEmpty') }}</div>
        <div v-if="last7Days.length >= 1" class="bar-chart-labels">
          <span v-for="(d, i) in last7Days" :key="i" class="line-chart-label">
            {{ (d.date || '').slice(5) }}
          </span>
        </div>
      </div>
    </WidgetShell>

    <div
      v-if="tooltip.visible"
      class="chart-tooltip"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
    >
      {{ tooltip.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import WidgetShell from './WidgetShell.vue'
import { fetchGetVisitorStats } from '@/service/api'

defineProps<{ fluid?: boolean }>()

const { t } = useI18n()

const visitorData = ref<App.Api.Dashboard.VisitorDayStat[]>([])

const last7Days = computed(() => {
  const all = visitorData.value
  return all.length >= 7 ? all.slice(-7) : all
})

const chartMax = computed(() =>
  Math.max(...last7Days.value.flatMap((d) => [d.pv ?? 0, d.uv ?? 0]), 1),
)

const chartW = 600
const chartH = 200
const chartPadX = 20
const chartPadY = 24

const slotW = computed(() => {
  const days = last7Days.value
  if (days.length <= 0) return 0
  return (chartW - chartPadX * 2) / days.length
})

const yByValue = (value: number) => {
  return chartPadY + (chartH - chartPadY * 2) * (1 - value / chartMax.value)
}

const groupW = computed(() => {
  if (slotW.value <= 0) return 0
  return Math.max(slotW.value - 10, 12)
})

const barGap = 4
const barWidth = computed(() => {
  if (groupW.value <= 0) return 0
  return Math.max((groupW.value - barGap) / 2, 4)
})

const barGroups = computed(() => {
  const days = last7Days.value
  if (days.length <= 0 || slotW.value <= 0) return []
  const baselineY = chartH - chartPadY
  return days.map((d, i) => {
    const groupX = chartPadX + i * slotW.value + (slotW.value - groupW.value) / 2
    const pvY = yByValue(d.pv ?? 0)
    const uvY = yByValue(d.uv ?? 0)
    return {
      pvX: groupX,
      pvY,
      pvHeight: Math.max(baselineY - pvY, 1),
      uvX: groupX + barWidth.value + barGap,
      uvY,
      uvHeight: Math.max(baselineY - uvY, 1),
      hitX: groupX - 2,
      hitW: groupW.value + 4,
    }
  })
})

const tooltip = ref({ visible: false, text: '', x: 0, y: 0 })

function showTooltip(index: number, event: MouseEvent) {
  const row = last7Days.value[index]
  if (!row) return
  tooltip.value.text = `${(row.date || '').slice(5)} · ${t('dashboard.visitorPv')}: ${row.pv ?? 0} · ${t('dashboard.visitorUv')}: ${row.uv ?? 0}`
  const rect = (event.target as SVGElement).getBoundingClientRect()
  tooltip.value.x = rect.left + rect.width / 2
  tooltip.value.y = rect.top - 28
  tooltip.value.visible = true
}

function hideTooltip() {
  tooltip.value.visible = false
}

onMounted(async () => {
  try {
    const res = await fetchGetVisitorStats()
    if (res.code === 1 && Array.isArray(res.data)) {
      visitorData.value = res.data
    }
  } catch {}
})
</script>

<style scoped>
.visitor-widget {
  min-width: 0;
}

.visitor-legend {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.65rem;
  color: var(--color-text-muted);
  font-size: 0.68rem;
}

.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-family: var(--font-family-mono);
}

.legend-dot {
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
}

.legend-dot--pv {
  background: var(--color-accent);
}

.legend-dot--uv {
  background: var(--color-text-muted);
}

.bar-chart-wrap {
  width: 100%;
}

.bar-chart-svg {
  width: 100%;
  height: 7rem;
  display: block;
}

.bar-chart-baseline {
  stroke: var(--color-border-subtle);
  stroke-width: 1;
}

.bar-rect {
  transition: opacity 0.15s ease;
}

.bar-rect--pv {
  fill: var(--color-accent);
}

.bar-rect--uv {
  fill: var(--color-text-muted);
}

.bar-hit-area {
  fill: transparent;
}

.bar-chart-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 0.3rem;
}

.line-chart-label {
  font-size: 0.62rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  letter-spacing: -0.02em;
}

.visitor-empty {
  height: 7rem;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border-subtle);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.chart-tooltip {
  position: fixed;
  z-index: 50;
  padding: 0.15rem 0.45rem;
  background: var(--color-text-primary);
  color: var(--color-bg-surface);
  font-size: 0.68rem;
  border-radius: var(--radius-sm);
  pointer-events: none;
  font-family: var(--font-family-mono);
  transform: translateX(-50%);
}
</style>
