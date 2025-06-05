<template>
  <div class="flex justify-center items-center p-2">
    <div class="">
      <div class="flex">
        <div v-for="col in 10" :key="col" class="flex flex-col gap-1 mr-1">
          <div
            v-for="row in 3"
            :key="row"
            class="relative w-5 h-5 rounded-[6px] transition-colors duration-300 ease ring-1 ring-gray-100 hover:ring-gray-300 hover:shadow-sm"
            :style="{ backgroundColor: getColor(getCell(row - 1, col - 1)?.count ?? 0) }"
            @mouseenter="showTooltip(row - 1, col - 1, $event)"
            @mouseleave="hideTooltip"
          ></div>
        </div>
      </div>
    </div>
    <!-- 自定义 tooltip -->
    <div
      v-if="tooltip.visible"
      class="absolute z-50 px-2 py-1 bg-orange-500 text-white text-xs rounded shadow"
      :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
    >
      {{ tooltip.text }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { fetchGetHeatMap } from '@/service/api'

// const props = defineProps<{
//   heatmapData: (App.Api.Ech0.HeatMap[0] | null)[]
// }>()

const heatmapData = ref<App.Api.Ech0.HeatMap>([])

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
  if (count >= 4) return '#196127'
  if (count >= 3) return '#239a3b'
  if (count >= 2) return '#7bc96f'
  if (count >= 1) return '#c6e48b'
  return '#fff'
}

// Tooltip 相关
const tooltip = ref({
  visible: false,
  text: '',
  x: 0,
  y: 0,
})

function showTooltip(row: number, col: number, event: MouseEvent) {
  const cell = getCell(row, col)
  if (cell) {
    tooltip.value.text = `${cell.date ?? ''}: ${cell.count ?? 0} 条`
    tooltip.value.visible = true

    // 获取触发事件的目标元素
    const target = event.target as HTMLElement
    const rect = target.getBoundingClientRect()

    // 计算 tooltip 的位置
    tooltip.value.x = rect.left
    tooltip.value.y = rect.top - 30 // 位于 cell 上方
  }
}

function hideTooltip() {
  tooltip.value.visible = false
}

onMounted(async () => {
  await fetchGetHeatMap().then((res) => {
    heatmapData.value = res.data
  })
})
</script>
