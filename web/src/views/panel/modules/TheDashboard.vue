<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { usePreferredReducedMotion, useTransition } from '@vueuse/core'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchCheckUpdate, fetchGetEchosByPage, fetchGetTodayEchos } from '@/service/api'
import { useConnectStore, useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { TheActivityLog, TheVisitorStatsWidget } from '@/components/advanced/widget'
import Box from '@/components/icons/box.vue'
import ConnectedIcon from '@/components/icons/connected-icon.vue'
import DateIcon from '@/components/icons/date-icon.vue'
import TotalIcon from '@/components/icons/total.vue'

type StatCard = {
  key: string
  label: string
  value: string
  icon: string
}

const settingStore = useSettingStore()
const connectStore = useConnectStore()
const { t, locale } = useI18n()

const loading = ref(true)
const echoTotal = ref<number | null>(null)
const todayEchoCount = ref<number | null>(null)
const connectCount = ref<number | null>(null)

const prefersReducedMotion = usePreferredReducedMotion()
const statAnimDuration = computed(() => (prefersReducedMotion.value === 'reduce' ? 0 : 820))

const echoAnimTarget = ref(0)
const todayAnimTarget = ref(0)
const connectAnimTarget = ref(0)

const echoAnimated = useTransition(echoAnimTarget, {
  duration: statAnimDuration,
})
const todayAnimated = useTransition(todayAnimTarget, {
  duration: statAnimDuration,
})
const connectAnimated = useTransition(connectAnimTarget, {
  duration: statAnimDuration,
})
const hasUpdate = ref(false)
const latestVersion = ref('')
const checkingUpdate = ref(false)

const formatMetric = (value: number | null) => {
  if (value === null) return '--'
  return new Intl.NumberFormat(locale.value).format(value)
}

const formatAnimatedMetric = (key: string) => {
  if (loading.value) return '--'
  if (key === 'echos' && echoTotal.value === null) return '--'
  if (key === 'today-echo' && todayEchoCount.value === null) return '--'
  if (key === 'connect' && connectCount.value === null) return '--'
  let n = 0
  if (key === 'echos') n = Math.round(echoAnimated.value)
  else if (key === 'today-echo') n = Math.round(todayAnimated.value)
  else if (key === 'connect') n = Math.round(connectAnimated.value)
  return new Intl.NumberFormat(locale.value).format(n)
}

const dashboardStats = computed<StatCard[]>(() => [
  {
    key: 'echos',
    label: String(t('dashboard.echoTotal')),
    value: formatMetric(echoTotal.value),
    icon: 'echos',
  },
  {
    key: 'today-echo',
    label: String(t('dashboard.todayEchoCount')),
    value: formatMetric(todayEchoCount.value),
    icon: 'today',
  },
  {
    key: 'connect',
    label: String(t('dashboard.connectedNodes')),
    value: formatMetric(connectCount.value),
    icon: 'connect',
  },
  {
    key: 'version',
    label: String(t('dashboard.currentVersion')),
    value: settingStore.hello?.version || '--',
    icon: 'version',
  },
])

const statValueText = (item: StatCard) => {
  if (item.key === 'version') return item.value
  return formatAnimatedMetric(item.key)
}

const todayText = computed(() => {
  return new Intl.DateTimeFormat(locale.value, {
    month: '2-digit',
    day: '2-digit',
    weekday: 'long',
  }).format(new Date())
})

const loadDashboardStats = async () => {
  loading.value = true
  const [echoRes, , todayRes] = await Promise.allSettled([
    fetchGetEchosByPage({ page: 1, pageSize: 1, search: '' }),
    connectStore.getConnect(),
    fetchGetTodayEchos(),
  ])

  if (echoRes.status === 'fulfilled' && echoRes.value.code === 1) {
    echoTotal.value = echoRes.value.data?.total ?? 0
    echoAnimTarget.value = echoTotal.value
  }

  if (todayRes.status === 'fulfilled' && todayRes.value.code === 1) {
    const data = todayRes.value.data
    todayEchoCount.value = Array.isArray(data) ? data.length : 0
    todayAnimTarget.value = todayEchoCount.value
  }

  connectCount.value = connectStore.connects.length
  connectAnimTarget.value = connectCount.value
  loading.value = false
}

const CHECK_UPDATE_ERR_TOAST_ID = 'dashboard-check-update-error'

const handleCheckUpdate = async () => {
  if (checkingUpdate.value) return
  checkingUpdate.value = true
  let failed = false
  try {
    const res = await fetchCheckUpdate()
    if (res.code === 1 && res.data) {
      hasUpdate.value = res.data.has_update
      latestVersion.value = res.data.latest_version
      if (res.data.has_update) {
        theToast.info(String(t('dashboard.updateAvailable', { version: res.data.latest_version })))
      } else {
        theToast.info(String(t('dashboard.alreadyLatest')))
      }
    } else {
      failed = true
    }
  } catch {
    failed = true
  } finally {
    checkingUpdate.value = false
  }
  if (failed) {
    theToast.error(String(t('dashboard.checkUpdateFailed')), { id: CHECK_UPDATE_ERR_TOAST_ID })
  }
}

const handleStatCardClick = (key: string) => {
  if (key === 'version') void handleCheckUpdate()
}

onMounted(() => {
  void loadDashboardStats()
})
</script>

<template>
  <div class="dashboard-page">
    <section class="dashboard-meta">
      <span class="meta-item meta-item-strong">PANEL DASHBOARD</span>
      <span class="meta-item">DATE {{ todayText }}</span>
    </section>

    <section class="stats-grid">
      <div
        v-for="(item, statIndex) in dashboardStats"
        :key="item.key"
        :class="['stat-card', item.key === 'version' ? 'stat-card--clickable' : '']"
        :style="{ '--stat-enter-delay': `${statIndex * 72}ms` }"
        v-tooltip="item.key === 'version' ? t('dashboard.clickToCheckUpdate') : undefined"
        @click="handleStatCardClick(item.key)"
      >
        <div class="stat-icon-wrap">
          <TotalIcon v-if="item.icon === 'echos'" class="stat-card-icon stat-card-icon--fill" />
          <DateIcon v-else-if="item.icon === 'today'" class="stat-card-icon stat-card-icon--fill" />
          <ConnectedIcon
            v-else-if="item.icon === 'connect'"
            class="stat-card-icon stat-card-icon--stroke"
          />
          <Box v-else-if="item.icon === 'version'" class="stat-card-icon stat-card-icon--stroke" />
        </div>
        <div class="stat-body">
          <p
            class="stat-value"
            :class="{
              'is-loading': loading && item.key !== 'version' && item.value === '--',
              'stat-value--numeric': item.key !== 'version',
            }"
          >
            {{ statValueText(item) }}
            <span
              v-if="item.key === 'version' && hasUpdate"
              class="update-dot"
              :title="t('dashboard.updateAvailable', { version: latestVersion })"
            />
            <svg
              v-if="item.key === 'version' && checkingUpdate"
              class="stat-spinner"
              xmlns="http://www.w3.org/2000/svg"
              width="14"
              height="14"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2.5"
              stroke-linecap="round"
            >
              <path d="M21 12a9 9 0 1 1-6.219-8.56" />
            </svg>
          </p>
          <p class="stat-label">{{ item.label }}</p>
        </div>
      </div>
    </section>

    <section class="dashboard-panels">
      <div class="panel-widget-wrap">
        <TheActivityLog />
      </div>
      <div class="panel-widget-wrap panel-widget-wrap--divider">
        <TheVisitorStatsWidget />
      </div>
    </section>
  </div>
</template>

<style scoped>
.dashboard-page {
  width: 100%;
  max-width: 100%;
  overflow: hidden;
  padding: 0.35rem 0 1rem;
  display: grid;
  gap: 0.85rem;
}

.dashboard-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.5rem;
  border-bottom: 1px dashed var(--color-border-subtle);
  padding: 0.35rem 0.2rem 0.55rem;
}

.meta-item {
  font-size: 0.73rem;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
}

.meta-item-strong {
  color: var(--color-text-secondary);
  font-weight: 700;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.1rem 0.6rem;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 0.7rem;
  padding: 0.65rem 0.4rem;
  border-bottom: 1px dashed var(--color-border-subtle);
  transition: opacity 0.2s ease;
  animation: stat-card-enter 0.58s cubic-bezier(0.22, 1, 0.36, 1) both;
  animation-delay: var(--stat-enter-delay, 0ms);
}

.stat-card:hover {
  opacity: 0.75;
}

.stat-card--clickable {
  cursor: pointer;
}

.stat-icon-wrap {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 2.2rem;
  height: 2.2rem;
  border-radius: var(--radius-md);
  border: 1px dashed var(--color-border-subtle);
  color: var(--color-text-muted);
}

.stat-card-icon {
  display: block;
  width: 18px;
  height: 18px;
}

.stat-card-icon--fill :deep(path) {
  fill: currentColor;
}

.stat-card-icon--stroke :deep(path) {
  stroke: currentColor;
}

.stat-card-icon--stroke :deep(g) {
  stroke: currentColor;
}

.stat-body {
  min-width: 0;
}

.stat-value {
  margin: 0;
  font-size: clamp(1.15rem, 2.4vw, 1.4rem);
  line-height: 1.2;
  color: var(--color-text-primary);
  font-weight: 600;
  font-family: var(--font-family-display);
  letter-spacing: -0.01em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.stat-value--numeric {
  font-variant-numeric: tabular-nums;
}

.stat-value.is-loading {
  opacity: 0.5;
}

.stat-label {
  margin: 0.05rem 0 0;
  font-size: 0.72rem;
  color: var(--color-text-muted);
  line-height: 1.3;
  font-family: var(--font-family-display);
  letter-spacing: 0.01em;
}

.update-dot {
  display: inline-block;
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 50%;
  background: var(--color-accent);
  margin-left: 0.3rem;
  vertical-align: middle;
  animation: dot-pulse 2s ease-in-out infinite;
}

.stat-spinner {
  display: inline-block;
  margin-left: 0.25em;
  vertical-align: middle;
  color: var(--color-text-muted);
  animation: spin 0.8s linear infinite;
}

.dashboard-panels {
  display: grid;
  grid-template-columns: 1fr;
  gap: 0;
}

.panel-widget-wrap {
  padding: 0.45rem 0.1rem;
}

.panel-widget-wrap--divider {
  margin-top: 0.2rem;
  padding-top: 0.9rem;
  border-top: 0.5px dashed var(--color-border-subtle);
}

@keyframes dot-pulse {
  0%,
  100% {
    opacity: 1;
  }

  50% {
    opacity: 0.35;
  }
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

@keyframes stat-card-enter {
  from {
    opacity: 0;
    transform: translateY(0.45rem);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (prefers-reduced-motion: reduce) {
  .stat-card {
    animation: none;
  }
}

@media (width >= 768px) {
  .stats-grid {
    grid-template-columns: repeat(4, 1fr);
  }

  .dashboard-panels {
    position: relative;
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .dashboard-panels::before {
    content: '';
    position: absolute;
    top: 0.35rem;
    bottom: 0.35rem;
    left: 50%;
    border-left: 0.5px dashed var(--color-border-subtle);
    transform: translateX(-50%);
    pointer-events: none;
  }

  .panel-widget-wrap--divider {
    margin-top: 0;
    padding-top: 0.45rem;
    border-top: none;
  }
}

@media (width <= 400px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
}
</style>
