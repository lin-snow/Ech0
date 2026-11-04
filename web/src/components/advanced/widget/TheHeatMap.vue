<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <WidgetShell title="Daily Log" :fluid="fluid">
    <template #icon>
      <CalendarIcon />
    </template>
    <div class="flex justify-center items-start">
      <div class="flex gap-1">
        <div v-for="col in 10" :key="col" class="flex flex-col gap-1">
          <div
            v-for="row in 3"
            :key="row"
            class="relative w-5 h-5 rounded-[6px] transition-colors duration-300 ease ring-1 ring-[var(--color-border-subtle)] hover:ring-[var(--color-border-strong)] hover:shadow-sm"
            :style="{ backgroundColor: getColor(getCell(row - 1, col - 1)?.count ?? 0) }"
            @mouseenter="showTooltip(row - 1, col - 1, $event)"
            @mouseleave="hideTooltip"
          ></div>
        </div>
      </div>
    </div>
  </WidgetShell>
  <div
    v-if="tooltip.visible"
    class="fixed z-50 px-2 py-1 bg-orange-500 text-white text-xs rounded shadow"
    :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
  >
    {{ tooltip.text }}
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import CalendarIcon from '@/components/icons/calendar.vue'
import WidgetShell from './WidgetShell.vue'
import { fetchGetHeatMap } from '@/service/api'
import { useI18n } from 'vue-i18n'

defineProps<{ fluid?: boolean }>()

const heatmapData = ref<App.Api.Ech0.HeatMap>([])
const { t } = useI18n()

const grid = computed(() => {
  const cells = [...heatmapData.value]
  const total = 3 * 10
  while (cells.length < total) cells.push({ date: '', count: 0 } as App.Api.Ech0.HeatMap[0])
  const result: (App.Api.Ech0.HeatMap[0] | null)[][] = []
  for (let row = 0; row < 3; row++) {
    result.push(cells.slice(row * 10, (row + 1) * 10))
  }
  return result
})

const getCell = (row: number, col: number) => {
  return grid.value[row]?.[col] ?? null
}

const getColor = (count: number): string => {
  if (count >= 4) return 'var(--heatmap-bg-color-4)'
  if (count >= 3) return 'var(--heatmap-bg-color-3)'
  if (count >= 2) return 'var(--heatmap-bg-color-2)'
  if (count >= 1) return 'var(--heatmap-bg-color-1)'
  return 'var(--heatmap-bg-color-0)'
}

const tooltip = ref({
  visible: false,
  text: '',
  x: 0,
  y: 0,
})

function showTooltip(row: number, col: number, event: MouseEvent) {
  const cell = getCell(row, col)
  if (cell) {
    tooltip.value.text = t('heatmap.tooltip', { date: cell.date ?? '', count: cell.count ?? 0 })
    tooltip.value.visible = true

    const target = event.target as HTMLElement
    const rect = target.getBoundingClientRect()

    tooltip.value.x = rect.left
    if (rect.top < 40) {
      tooltip.value.y = rect.bottom + 10
    } else {
      tooltip.value.y = rect.top - 30
    }
  }
}

function hideTooltip() {
  tooltip.value.visible = false
}

onMounted(() => {
  fetchGetHeatMap().then((res) => {
    heatmapData.value = res.data
  })
})
</script>
