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

    <div v-if="exportFormat === 'capsule'" class="export-capsule-extra">
      <p class="export-capsule-warning">{{ t('exportSetting.capsuleWarning') }}</p>
      <BaseSwitch v-model="exportIncludePrivate" :disabled="isExporting">
        {{ t('exportSetting.includePrivate') }}
      </BaseSwitch>
      <p class="export-capsule-hint">{{ t('exportSetting.includePrivateHint') }}</p>
    </div>

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

// 导出 = 异步 export 作业产出快照或胶囊（与导入对称：重活在 job 里，有进度/可取消），完成后取回产物。
const isExporting = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
)

// 快照与胶囊不可互换：快照是唯一能灾难恢复的完整备份，胶囊只是内容。故快照恒为默认项，
// 描述里把「含凭据 / 不含凭据」讲清楚，避免被当成两个平级中性选项。
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

// 进度卡与产物行描述的是「已发起的作业」,一律读 snapshotFormat 而非界面选择器。
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

// 步进器对齐后端真实阶段：排队(pending) → 打包(packing) → 完成(completed)。
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

// 鉴权下载：经 fetchDownloadExport（credentials + Authorization header）取回 blob 触发下载，
// token 不进 URL（避免出现在浏览器历史/日志/Referer）。
const downloadSnapshot = async () => {
  try {
    // 下载跟随作业实际产物的格式：用户导完胶囊后可能把选择器拨回快照，此时仍须取回胶囊。
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
    // 作业完成后由下方 watch(snapshotStatus) 自动触发下载。
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
  // 进入页面时若有进行中的导出作业则接管轮询（完成会自动下载）。
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

/* 与导入 tab 的 migration-source-grid 保持一致的卡片式选择器，两页视觉对称。 */
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

.export-capsule-extra {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  align-items: flex-start;
}

.export-capsule-warning {
  padding: 0.5rem 0.65rem;
  color: var(--color-danger);
  background: color-mix(in srgb, var(--color-danger) 10%, var(--color-bg-surface));
  border: 1px solid color-mix(in srgb, var(--color-danger) 35%, transparent);
  border-radius: var(--radius-md);
  font-size: 0.83rem;
}

.export-capsule-hint {
  color: var(--color-text-muted);
  font-size: 0.8rem;
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
