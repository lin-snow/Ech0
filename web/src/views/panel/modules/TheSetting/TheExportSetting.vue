<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="export-wrap">
    <div class="export-header">
      <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
        {{ t('exportSetting.title') }}
      </h1>
      <p class="export-desc">{{ t('exportSetting.description') }}</p>
    </div>

    <div class="export-format-grid">
      <button
        v-for="option in formatCards"
        :key="option.value"
        class="export-format-card"
        :class="{ active: exportFormat === option.value, disabled: isExporting }"
        :disabled="isExporting"
        @click="exportFormat = option.value"
      >
        <h3>{{ option.title }}</h3>
        <p>{{ option.desc }}</p>
      </button>
    </div>

    <BaseSwitch
      v-if="exportFormat === 'capsule'"
      v-model="exportIncludePrivate"
      :disabled="isExporting"
      class="export-include-private"
    >
      {{ t('exportSetting.includePrivate') }}
    </BaseSwitch>

    <div class="export-action">
      <BaseButton
        @click="handleExport"
        :disabled="isExporting"
        class="export-download-btn"
        :tooltip="exportActionText"
      >
        {{ isExporting ? t('exportSetting.exporting') : exportActionText }}
      </BaseButton>
    </div>

    <JobProgressCard
      v-if="snapshotStatus !== 'idle'"
      :title="jobTitle"
      :status="snapshotStatus"
      :status-label="statusLabelMap[snapshotStatus] || snapshotStatus"
      :steps="exportSteps"
      :current-key="exportCurrentKey"
      :error-message="snapshotStatus === 'failed' ? snapshotError : ''"
    >
      <template v-if="snapshotStatus === 'success'" #footer>
        <div class="export-artifact">
          <span class="export-artifact__label">{{ t('exportSetting.artifactLabel') }}</span>
          <span class="export-artifact__format">{{ jobFormatTitle }}</span>
          <span class="export-artifact__name" v-tooltip="snapshotFileName">
            {{ snapshotFileName || '—' }}
          </span>
          <span class="export-artifact__size">{{ formatBytes(snapshotSize) }}</span>
          <BaseButton :tooltip="t('exportSetting.redownload')" @click="downloadSnapshot">
            {{ t('exportSetting.redownload') }}
          </BaseButton>
        </div>
      </template>
    </JobProgressCard>
  </div>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import JobProgressCard from './components/JobProgressCard.vue'
import { computed, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { theToast } from '@/utils/toast'
import { useSettingStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { fetchDownloadExport } from '@/service/api'
import type { ExportFormat } from '@/service/api'
import { formatBytes } from '@/utils/file'

const { t } = useI18n()
const settingStore = useSettingStore()
const userStore = useUserStore()
const { startSnapshotTask, restoreSnapshotTask } = settingStore
const {
  snapshotStatus,
  snapshotError,
  snapshotPhase,
  snapshotFileName,
  snapshotSize,
  snapshotFormat,
  exportFormat,
  exportIncludePrivate,
} = storeToRefs(settingStore)
const { isLogin } = storeToRefs(userStore)

const isExporting = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
)

const formatCards = computed<{ value: ExportFormat; title: string; desc: string }[]>(() => [
  {
    value: 'snapshot',
    title: String(t('exportSetting.formatSnapshotTitle')),
    desc: String(t('exportSetting.formatSnapshotDesc')),
  },
  {
    value: 'capsule',
    title: String(t('exportSetting.formatCapsuleTitle')),
    desc: String(t('exportSetting.formatCapsuleDesc')),
  },
])

const exportActionText = computed(() =>
  exportFormat.value === 'capsule'
    ? String(t('exportSetting.exportCapsule'))
    : String(t('exportSetting.exportSnapshot')),
)

const jobTitle = computed(() =>
  snapshotFormat.value === 'capsule'
    ? String(t('exportSetting.jobTitleCapsule'))
    : String(t('exportSetting.jobTitle')),
)

const jobFormatTitle = computed(() =>
  snapshotFormat.value === 'capsule'
    ? String(t('exportSetting.formatCapsuleTitle'))
    : String(t('exportSetting.formatSnapshotTitle')),
)

const exportSteps = computed(() => [
  { key: 'pending', label: String(t('jobProgress.exportPhasePending')) },
  { key: 'packing', label: String(t('jobProgress.exportPhasePacking')) },
  { key: 'completed', label: String(t('jobProgress.exportPhaseCompleted')) },
])

const exportCurrentKey = computed(() => {
  if (snapshotStatus.value === 'pending') return 'pending'
  if (snapshotStatus.value === 'success') return 'completed'
  return snapshotPhase.value || 'packing'
})

const statusLabelMap = computed<Record<string, string>>(() => ({
  idle: String(t('jobProgress.statusIdle')),
  pending: String(t('jobProgress.statusPending')),
  running: String(t('jobProgress.statusRunning')),
  success: String(t('jobProgress.statusSuccess')),
  failed: String(t('jobProgress.statusFailed')),
  cancelled: String(t('jobProgress.statusCancelled')),
}))

const downloadSnapshot = async () => {
  try {
    const format = snapshotFormat.value
    const blob = await fetchDownloadExport(format)
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = snapshotFileName.value || `ech0-${format}-${Date.now()}.zip`
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  } catch (error) {
    console.error(String(t('exportSetting.exportFailed')), error)
    theToast.error(String(t('exportSetting.exportFailed')))
  }
}

const handleExport = async () => {
  if (!isLogin.value) {
    theToast.info(String(t('exportSetting.loginRequired')), { duration: 3000 })
    return
  }
  if (isExporting.value) return
  try {
    theToast.info(String(t('exportSetting.exporting')), { duration: 4000 })
    const res = await startSnapshotTask(exportFormat.value, exportIncludePrivate.value)
    if (!res) return
    if (res.code !== 1) {
      theToast.error(res.msg || String(t('exportSetting.exportFailed')))
    }
  } catch (error) {
    console.error(String(t('exportSetting.exportFailed')), error)
    theToast.error(String(t('exportSetting.exportFailed')))
  }
}

watch(
  () => snapshotStatus.value,
  (status, prevStatus) => {
    if (status === prevStatus) return
    if (status === 'success') {
      theToast.success(String(t('exportSetting.exportStarted')))
      void downloadSnapshot()
      return
    }
    if (status === 'failed') {
      theToast.error(snapshotError.value || String(t('exportSetting.exportFailed')))
    }
  },
)

onMounted(() => {
  void restoreSnapshotTask()
})
</script>

<style scoped>
.export-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
}

.export-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.export-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.export-format-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.75rem;
}

.export-format-card {
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  border-radius: var(--radius-md);
  padding: 0.75rem;
  text-align: left;
  transition: all 0.2s ease;
}

.export-format-card h3 {
  margin-bottom: 0.35rem;
  color: var(--color-text-primary);
  font-weight: 700;
}

.export-format-card p {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
}

.export-format-card.active {
  border-color: var(--color-nav-active-bg);
  box-shadow: inset 0 0 0 1px var(--color-nav-active-bg);
}

.export-format-card.disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.export-include-private {
  align-self: flex-start;
}

.export-action {
  display: flex;
  align-items: center;
}

.export-download-btn {
  border-radius: var(--radius-md);
  color: var(--color-text-primary) !important;
}

.export-artifact {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.82rem;
}

.export-artifact__label {
  color: var(--color-text-muted);
}

.export-artifact__name {
  max-width: 16rem;
  color: var(--color-text-primary);
  font-family: var(--font-family-mono);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.export-artifact__size {
  padding: 0.05rem 0.4rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  font-variant-numeric: tabular-nums;
}

.export-artifact__format {
  padding: 0.05rem 0.4rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
}

@media (width <= 768px) {
  .export-format-grid {
    grid-template-columns: 1fr;
  }
}
</style>
