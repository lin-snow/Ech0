<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <WidgetShell title="Connect" :fluid="fluid">
    <template #icon>
      <Connect />
    </template>
    <div v-if="!loading">
      <div v-if="!connectsInfo.length" class="text-[var(--color-text-muted)] text-sm mb-2">
        {{ t('connectWidget.noConnections') }}
      </div>
      <div v-else class="flex flex-wrap gap-3">
        <div
          v-for="(connect, index) in connectsInfo"
          :key="index"
          class="relative flex flex-col items-center justify-center w-8 h-8 min-w-[2rem] min-h-[2rem] flex-none border-2 border-[var(--color-border-subtle)] shadow-sm rounded-full hover:shadow-md transition duration-200 ease-in-out group"
        >
          <a
            :href="connect.server_url"
            target="_blank"
            class="block w-full h-full"
            @mouseenter="showTooltip(connect, $event)"
            @mouseleave="hideTooltip"
          >
            <img
              :src="connect.logo"
              alt="avatar"
              loading="lazy"
              decoding="async"
              class="w-full h-full rounded-full object-cover"
            />
            <span
              class="absolute top-0 right-0 w-2.5 h-2.5 border-2 border-[var(--color-bg-surface)] rounded-full"
              :style="{
                transform: 'translate(35%, -35%)',
                backgroundColor: getColor(connect.today_echos || 0),
              }"
            ></span>
          </a>
        </div>
      </div>
    </div>
    <div v-else>
      <div class="text-[var(--color-text-secondary)] text-sm mb-2">
        {{ t('connectWidget.loading') }}
      </div>
    </div>
  </WidgetShell>
  <div
    v-if="tooltip.visible"
    class="fixed z-50 min-w-max bg-gray-800 text-white text-xs rounded px-3 py-2 shadow-lg pointer-events-none"
    :style="{ left: tooltip.x + 'px', top: tooltip.y + 'px' }"
  >
    <div class="font-bold mb-1">{{ tooltip.data?.server_name }}</div>
    <div>{{ t('connectWidget.tooltipOwner') }}: {{ tooltip.data?.sys_username || '-' }}</div>
    <div>{{ t('connectWidget.tooltipTotal') }}: {{ tooltip.data?.total_echos ?? 0 }}</div>
    <div>{{ t('connectWidget.tooltipToday') }}: {{ tooltip.data?.today_echos ?? 0 }}</div>
    <div>{{ t('connectWidget.tooltipVersion') }}: {{ tooltip.data?.version || '-' }}</div>
  </div>
</template>

<script setup lang="ts">
import Connect from '@/components/icons/connect.vue'
import WidgetShell from './WidgetShell.vue'
import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

defineProps<{ fluid?: boolean }>()

const CONNECT_INFO_POLL_MS = 60_000

const connectStore = useConnectStore()
const { t } = useI18n()
const { getConnectInfo } = connectStore
const { loading, connectsInfo } = storeToRefs(connectStore)

const tooltip = ref<{
  visible: boolean
  x: number
  y: number
  data: (typeof connectsInfo.value)[number] | null
}>({ visible: false, x: 0, y: 0, data: null })

function showTooltip(connect: (typeof connectsInfo.value)[number], event: MouseEvent) {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  tooltip.value.data = connect
  tooltip.value.x = rect.left
  tooltip.value.y = rect.bottom + 6
  tooltip.value.visible = true
}

function hideTooltip() {
  tooltip.value.visible = false
}

const getColor = (count: number): string => {
  if (count >= 4) return 'var(--heatmap-bg-color-4)'
  if (count >= 3) return 'var(--heatmap-bg-color-3)'
  if (count >= 2) return 'var(--heatmap-bg-color-2)'
  if (count >= 1) return 'var(--heatmap-bg-color-1)'
  return 'var(--color-border-subtle)'
}

let refreshTimer: number | null = null

const handleVisibilityChange = () => {
  if (document.visibilityState === 'visible') void getConnectInfo()
}

onMounted(() => {
  void getConnectInfo()
  // The store returns cached data instantly and revalidates in the background;
  // polling only triggers a revalidation once the 30-minute cache expires.
  refreshTimer = window.setInterval(() => void getConnectInfo(), CONNECT_INFO_POLL_MS)
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  if (refreshTimer !== null) {
    window.clearInterval(refreshTimer)
    refreshTimer = null
  }
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>
