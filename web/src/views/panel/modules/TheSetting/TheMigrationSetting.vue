<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="migration-wrap">
    <div class="migration-header">
      <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
        {{ t('migrationSetting.title') }}
      </h1>
      <p class="migration-desc">{{ t('migrationSetting.description') }}</p>
    </div>

    <div class="migration-source-grid">
      <button
        v-for="source in sourceCards"
        :key="source.value"
        class="migration-source-card"
        :class="{ active: sourceType === source.value, disabled: source.inDevelopment }"
        @click="handleSelectSource(source)"
      >
        <div class="migration-source-title-wrap">
          <h3>{{ source.title }}</h3>
          <span v-if="source.inDevelopment" class="migration-dev-badge">{{
            t('migrationSetting.inDevelopment')
          }}</span>
        </div>
        <p>{{ source.desc }}</p>
      </button>
    </div>

    <div v-if="sourceType === 'capsule'" class="migration-capsule-panel">
      <p class="migration-capsule-note">{{ t('migrationSetting.capsuleNoteAppend') }}</p>
      <p class="migration-capsule-note">{{ t('migrationSetting.capsuleNoteNoCredential') }}</p>
      <div class="migration-capsule-switch">
        <BaseSwitch v-model="capsuleIncludePrivate" :disabled="isSubmittingMigration">
          {{ t('migrationSetting.capsuleIncludePrivate') }}
        </BaseSwitch>
        <p class="migration-capsule-hint">{{ t('migrationSetting.capsuleIncludePrivateHint') }}</p>
      </div>
    </div>

    <div class="migration-form">
      <div class="migration-row migration-row-top">
        <span class="migration-label">{{ t('migrationSetting.sourceZip') }}</span>
        <div class="migration-upload-wrap">
          <BaseButton
            :tooltip="t('migrationSetting.pickZip')"
            :disabled="isSubmittingMigration"
            @click="handlePickZip"
          >
            {{ t('migrationSetting.pickZip') }}
          </BaseButton>
          <p class="migration-file-name">
            {{ selectedZipName || t('migrationSetting.noFileSelected') }}
          </p>
        </div>
      </div>
    </div>

    <div class="migration-actions">
      <BaseButton
        :tooltip="t('migrationSetting.startMigration')"
        :disabled="isSubmittingMigration"
        @click="handleStartMigration"
      >
        {{ startActionText }}
      </BaseButton>
      <BaseButton
        :tooltip="t('migrationSetting.refreshStatus')"
        :disabled="isSubmittingMigration"
        @click="handleRefreshJob"
      >
        {{ t('migrationSetting.refreshStatus') }}
      </BaseButton>
      <BaseButton
        v-if="migrationStore.isRunning"
        :tooltip="t('migrationSetting.cancelJob')"
        :disabled="isSubmittingMigration"
        @click="handleCancelJob"
      >
        {{ t('migrationSetting.cancelJob') }}
      </BaseButton>
      <BaseButton
        v-if="migrationStore.canCleanup"
        :tooltip="migrationStore.isSuccess ? t('commonUi.done') : t('migrationSetting.cleanup')"
        :disabled="isSubmittingMigration"
        @click="handleCleanupMigration"
      >
        {{ migrationStore.isSuccess ? t('commonUi.done') : t('migrationSetting.cleanup') }}
      </BaseButton>
    </div>
    <p v-if="isUploadingZip" class="migration-progress-tip">
      {{ t('migrationSetting.uploadingTip') }}
    </p>
    <p v-else-if="isCreatingMigration" class="migration-progress-tip">
      {{ t('migrationSetting.creatingTip') }}
    </p>

    <JobProgressCard
      v-if="migrationStore.hasJob"
      :title="t('migrationSetting.jobTitle')"
      :subtitle="jobSubtitle"
      :status="migrationStore.state.status"
      :status-label="statusLabelMap[migrationStore.state.status] || migrationStore.state.status"
      :steps="importSteps"
      :current-key="migrationStore.state.phase"
      :metrics="jobMetrics"
      :meta="jobMeta"
      :error-message="migrationStore.state.error_message"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import JobProgressCard from './components/JobProgressCard.vue'
import { fetchUploadMigrationSourceZip, type MigrationSourceType } from '@/service/api'
import { useMigrationStore } from '@/stores'
import { theToast } from '@/utils/toast'

// SourceCard 的 value 直接复用 API 契约里的来源枚举,避免两处各写一份。
interface SourceCard {
  value: MigrationSourceType
  title: string
  desc: string
  inDevelopment?: boolean
}

const sourceCards = computed<SourceCard[]>(() => [
  { value: 'ech0', title: 'Ech0', desc: String(t('migrationSetting.sourceEch0')) },
  {
    value: 'capsule',
    title: String(t('migrationSetting.sourceCapsuleTitle')),
    desc: String(t('migrationSetting.sourceCapsule')),
  },
  {
    value: 'memos',
    title: 'Memos',
    desc: String(t('migrationSetting.sourceMemos')),
    inDevelopment: true,
  },
])

const sourceType = ref<MigrationSourceType>('ech0')
const selectedZip = ref<File | null>(null)
const selectedZipName = ref('')
const capsuleIncludePrivate = ref(false)
const isUploadingZip = ref(false)
const isCreatingMigration = ref(false)
const migrationStore = useMigrationStore()
const { t, locale } = useI18n()
const statusLabelMap = computed<Record<string, string>>(() => ({
  idle: String(t('migrationSetting.statusIdle')),
  pending: String(t('migrationSetting.statusPending')),
  running: String(t('migrationSetting.statusRunning')),
  success: String(t('migrationSetting.statusSuccess')),
  failed: String(t('migrationSetting.statusFailed')),
  cancelled: String(t('migrationSetting.statusCancelled')),
}))
const sourceLabelMap = computed<Record<string, string>>(() => ({
  ech0: 'Ech0',
  memos: 'Memos',
  capsule: String(t('migrationSetting.sourceCapsuleTitle')),
}))
const migrationReport = computed(
  () => (migrationStore.state.source_payload?.report as Record<string, unknown> | undefined) ?? {},
)
const migrationJobId = computed(
  () =>
    (migrationStore.state.source_payload?.migration_job_id as string | undefined) ||
    (migrationReport.value.job_id as string | undefined),
)
const migrationProcessed = computed(() => migrationReport.value.processed)
const migrationSuccess = computed(() => migrationReport.value.success_count)
const migrationFail = computed(() => migrationReport.value.fail_count)
const hasMetrics = computed(
  () =>
    migrationProcessed.value !== undefined ||
    migrationSuccess.value !== undefined ||
    migrationFail.value !== undefined,
)
const formattedStartedAt = computed(() => formatTime(migrationStore.state.started_at))
const formattedFinishedAt = computed(() => formatTime(migrationStore.state.finished_at))
const isSubmittingMigration = computed(() => isUploadingZip.value || isCreatingMigration.value)
const startActionText = computed(() => {
  if (isUploadingZip.value) return String(t('migrationSetting.uploading'))
  if (isCreatingMigration.value) return String(t('migrationSetting.creating'))
  return String(t('migrationSetting.startMigration'))
})

// ---- job 进度卡 ----
const jobSubtitle = computed(
  () =>
    `${t('migrationSetting.source')} ${sourceLabelMap.value[migrationStore.state.source_type] || migrationStore.state.source_type}`,
)

// 步进器阶段名由具体 importer 决定:ech0 走 解析→写入→汇总→完成,
// 胶囊 importer 只上报 校验→导入→完成,两套不能混用。
const importSteps = computed(() => {
  if (migrationStore.state.source_type === 'capsule') {
    return [
      { key: 'checking', label: String(t('jobProgress.importPhaseChecking')) },
      { key: 'importing', label: String(t('jobProgress.importPhaseImporting')) },
      { key: 'completed', label: String(t('jobProgress.importPhaseCompleted')) },
    ]
  }
  return [
    { key: 'extracting', label: String(t('jobProgress.importPhaseExtracting')) },
    { key: 'loading', label: String(t('jobProgress.importPhaseLoading')) },
    { key: 'reporting', label: String(t('jobProgress.importPhaseReporting')) },
    { key: 'completed', label: String(t('jobProgress.importPhaseCompleted')) },
  ]
})

const metricText = (v: unknown) => (v === undefined || v === null ? '—' : String(v))
const failIsPositive = computed(() => {
  const n = Number(migrationFail.value)
  return Number.isFinite(n) && n > 0
})

const jobMetrics = computed(() => {
  if (!hasMetrics.value) return []
  return [
    {
      label: String(t('migrationSetting.totalProcessed')),
      value: metricText(migrationProcessed.value),
    },
    {
      label: String(t('migrationSetting.success')),
      value: metricText(migrationSuccess.value),
      tone: 'success' as const,
    },
    {
      label: String(t('migrationSetting.failed')),
      value: metricText(migrationFail.value),
      tone: failIsPositive.value ? ('danger' as const) : undefined,
    },
  ]
})

const jobMeta = computed(() => {
  const lines: { label: string; value: string }[] = []
  if (migrationJobId.value) {
    lines.push({ label: String(t('migrationSetting.jobId')), value: String(migrationJobId.value) })
  }
  if (formattedStartedAt.value) {
    lines.push({ label: String(t('migrationSetting.startedAt')), value: formattedStartedAt.value })
  }
  if (formattedFinishedAt.value) {
    lines.push({
      label: String(t('migrationSetting.finishedAt')),
      value: formattedFinishedAt.value,
    })
  }
  return lines
})

const resetSelectedZip = () => {
  selectedZip.value = null
  selectedZipName.value = ''
}

const handleSelectSource = (source: SourceCard) => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  if (source.inDevelopment) {
    theToast.info(String(t('migrationSetting.sourceInDevelopment', { source: source.title })))
    return
  }
  sourceType.value = source.value
  resetSelectedZip()
}

const handlePickZip = () => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.zip,application/zip'
  input.onchange = (event: Event) => {
    const target = event.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return
    if (!file.name.toLowerCase().endsWith('.zip')) {
      theToast.error(String(t('migrationSetting.onlyZip')))
      return
    }
    selectedZip.value = file
    selectedZipName.value = file.name
  }
  input.click()
}

const handleStartMigration = async () => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  if (sourceType.value === 'memos') {
    theToast.info(String(t('migrationSetting.memosUnavailable')))
    return
  }
  if (migrationStore.hasJob) {
    theToast.info(String(t('migrationSetting.cleanupFirst')))
    return
  }
  if (!selectedZip.value) {
    theToast.info(String(t('migrationSetting.selectZipFirst')))
    return
  }

  try {
    isUploadingZip.value = true
    theToast.info(String(t('migrationSetting.uploadingRequest')))
    const uploadRes = await fetchUploadMigrationSourceZip(sourceType.value, selectedZip.value)
    if (uploadRes.code !== 1) {
      theToast.error(uploadRes.msg || String(t('migrationSetting.uploadFailed')))
      return
    }
    const sourcePayload: Record<string, unknown> = { ...(uploadRes.data?.source_payload ?? {}) }
    // 私密内容开关只对胶囊有意义,随 source_payload 交给后端作业。
    if (sourceType.value === 'capsule') {
      sourcePayload.include_private = capsuleIncludePrivate.value
    }

    isUploadingZip.value = false
    isCreatingMigration.value = true
    const res = await migrationStore.startMigration({
      source_type: sourceType.value,
      source_payload: sourcePayload,
    })
    if (res.code !== 1) {
      theToast.error(res.msg || String(t('migrationSetting.createJobFailed')))
      return
    }
    theToast.success(String(t('migrationSetting.started')))
  } catch (error) {
    console.error('Upload migration source zip failed:', error)
    theToast.error(String(t('migrationSetting.requestFailed')))
  } finally {
    isUploadingZip.value = false
    isCreatingMigration.value = false
  }
}

const handleRefreshJob = async () => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  const ok = await migrationStore.fetchStatus()
  if (!ok) {
    theToast.error(String(t('migrationSetting.refreshFailed')))
    return
  }
  theToast.success(String(t('migrationSetting.refreshed')))
}

const handleCancelJob = async () => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  if (!migrationStore.isRunning) {
    theToast.info(String(t('migrationSetting.noRunningJob')))
    return
  }
  const res = await migrationStore.cancelMigration()
  if (res.code !== 1) {
    theToast.error(res.msg || String(t('migrationSetting.cancelFailed')))
    return
  }
  theToast.success(String(t('migrationSetting.cancelled')))
}

const handleCleanupMigration = async () => {
  if (isSubmittingMigration.value) {
    theToast.info(String(t('migrationSetting.processing')))
    return
  }
  const res = await migrationStore.cleanupMigration()
  if (res.code !== 1) {
    theToast.error(res.msg || String(t('migrationSetting.cleanupFailed')))
    return
  }
  theToast.success(String(t('migrationSetting.cleaned')))
}

const formatTime = (ts?: number) => {
  if (!ts) return ''
  const dt = new Date(ts * 1000)
  if (Number.isNaN(dt.getTime())) return String(ts)
  return dt.toLocaleString(locale.value, { hour12: false })
}

void migrationStore.init()
</script>

<style scoped>
.migration-wrap {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.migration-header {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.migration-desc {
  color: var(--color-text-secondary);
  font-size: 0.9rem;
}

.migration-source-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 0.75rem;
}

.migration-source-card {
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-surface);
  border-radius: var(--radius-md);
  padding: 0.75rem;
  text-align: left;
  transition: all 0.2s ease;
}

.migration-source-title-wrap {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  margin-bottom: 0.35rem;
}

.migration-source-card h3 {
  color: var(--color-text-primary);
  font-weight: 700;
}

.migration-dev-badge {
  display: inline-flex;
  align-items: center;
  border-radius: 999px;
  padding: 0.05rem 0.4rem;
  font-size: 0.72rem;
  line-height: 1.2;
  color: #a56900;
  background: #fff5db;
  border: 1px solid #f0d59a;
}

.migration-source-card p {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
}

.migration-source-card.disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.migration-source-card.active {
  border-color: var(--color-nav-active-bg);
  box-shadow: inset 0 0 0 1px var(--color-nav-active-bg);
}

.migration-capsule-panel {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-surface);
  padding: 0.75rem;
}

.migration-capsule-note {
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  line-height: 1.5;
}

.migration-capsule-switch {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.3rem;
  margin-top: 0.25rem;
}

.migration-capsule-hint {
  color: var(--color-text-secondary);
  font-size: 0.8rem;
}

.migration-form {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.migration-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.migration-row-top {
  align-items: flex-start;
}

.migration-label {
  min-width: 6.2rem;
  color: var(--color-text-secondary);
  font-weight: 600;
}

.migration-upload-wrap {
  width: 100%;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.5rem;
}

.migration-file-name {
  color: var(--color-text-secondary);
  font-size: 0.82rem;
}

.migration-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
}

.migration-progress-tip {
  color: var(--color-text-secondary);
  font-size: 0.83rem;
}

@media (width <= 768px) {
  .migration-source-grid {
    grid-template-columns: 1fr;
  }
}
</style>
