<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full space-y-3">
    <div class="flex flex-wrap items-start justify-between gap-3">
      <div class="space-y-1">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('snapshotScheduleSetting.title') }}
        </h1>
        <p class="text-[var(--color-text-secondary)] text-sm">
          {{ t('snapshotScheduleSetting.description') }}
        </p>
      </div>
      <BaseEditCapsule
        :editing="scheduleEditMode"
        :apply-title="t('commonUi.apply')"
        :cancel-title="t('commonUi.cancel')"
        :edit-title="t('commonUi.edit')"
        @apply="handleUpdateSnapshotSchedule"
        @toggle="scheduleEditMode = !scheduleEditMode"
      />
    </div>

    <div class="schedule-row">
      <h2 class="schedule-row__label">
        {{ t('snapshotScheduleSetting.enableAutoSnapshot') }}
      </h2>
      <div class="schedule-row__control">
        <BaseSwitch v-model="SnapshotSchedule.enable" :disabled="!scheduleEditMode" />
      </div>
    </div>

    <div class="schedule-row schedule-row--top">
      <h2 class="schedule-row__label">
        {{ t('snapshotScheduleSetting.crontab') }}
      </h2>
      <div class="schedule-row__control">
        <div v-if="!scheduleEditMode" class="schedule-display">
          <p class="schedule-display__text">
            {{ SnapshotSchedule.cron_expression.length === 0 ? t('commonUi.none') : humanizedCron }}
          </p>
          <code
            v-if="SnapshotSchedule.cron_expression.length > 0"
            class="schedule-display__code"
            v-tooltip="SnapshotSchedule.cron_expression"
          >
            {{ SnapshotSchedule.cron_expression }}
          </code>
        </div>
        <CronScheduleEditor v-else v-model="SnapshotSchedule.cron_expression" />
      </div>
    </div>

    <div class="schedule-row">
      <h2 class="schedule-row__label">
        {{ t('snapshotScheduleSetting.manualLabel') }}
      </h2>
      <div class="schedule-row__control">
        <BaseButton
          class="snapshot-create-btn"
          :disabled="isCreating"
          :tooltip="t('snapshotScheduleSetting.createNow')"
          @click="handleCreateNow"
        >
          {{
            isCreating
              ? t('snapshotScheduleSetting.creating')
              : t('snapshotScheduleSetting.createNow')
          }}
        </BaseButton>
      </div>
    </div>

    <JobProgressCard
      v-if="snapshotStatus !== 'idle'"
      :title="t('snapshotScheduleSetting.jobTitle')"
      :status="snapshotStatus"
      :status-label="statusLabelMap[snapshotStatus] || snapshotStatus"
      :steps="exportSteps"
      :current-key="exportCurrentKey"
      :error-message="snapshotStatus === 'failed' ? snapshotError : ''"
    />
  </div>
</template>

<script setup lang="ts">
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import CronScheduleEditor from './components/CronScheduleEditor.vue'
import JobProgressCard from './components/JobProgressCard.vue'
import { computed, ref, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { humanizeCron } from '@/utils/cron'
import { fetchUpdateSnapshotScheduleSetting } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getSnapshotSchedule, startSnapshotTask, restoreSnapshotTask } = settingStore
const { SnapshotSchedule, snapshotStatus, snapshotError, snapshotPhase } = storeToRefs(settingStore)

const scheduleEditMode = ref<boolean>(false)
const humanizedCron = computed(() => humanizeCron(SnapshotSchedule.value.cron_expression, t))

const isCreating = computed(
  () => snapshotStatus.value === 'pending' || snapshotStatus.value === 'running',
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

const handleCreateNow = async () => {
  if (isCreating.value) return
  try {
    const res = await startSnapshotTask()
    if (!res) return
    if (res.code !== 1) {
      theToast.error(res.msg || String(t('snapshotScheduleSetting.createFailed')))
    }
  } catch (error) {
    console.error(String(t('snapshotScheduleSetting.createFailed')), error)
    theToast.error(String(t('snapshotScheduleSetting.createFailed')))
  }
}

watch(
  () => snapshotStatus.value,
  (status, prevStatus) => {
    if (status === prevStatus) return
    if (status === 'success') {
      theToast.success(String(t('snapshotScheduleSetting.createSuccess')))
    } else if (status === 'failed') {
      theToast.error(snapshotError.value || String(t('snapshotScheduleSetting.createFailed')))
    }
  },
)

const handleUpdateSnapshotSchedule = async () => {
  const res = await fetchUpdateSnapshotScheduleSetting(SnapshotSchedule.value)
  if (res.code === 1) {
    theToast.success(res.msg)
  }

  scheduleEditMode.value = false
  await getSnapshotSchedule()
}

onMounted(async () => {
  await getSnapshotSchedule()
  void restoreSnapshotTask()
})
</script>

<style scoped>
.schedule-row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 0.75rem;
  min-height: 2.5rem;
  color: var(--color-text-secondary);
}

.schedule-row--top {
  align-items: flex-start;
}

.schedule-row__label {
  flex: 0 0 auto;
  width: 9rem;
  font-weight: 600;
  font-size: 0.9rem;
  line-height: 1.4;
}

.schedule-row--top .schedule-row__label {
  padding-top: 0.2rem;
}

.schedule-row__control {
  flex: 1;
  min-width: 0;
}

.schedule-display {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
  min-width: 0;
}

.schedule-display__text {
  font-size: 0.9rem;
  color: var(--color-text-primary);
  line-height: 1.4;
  margin: 0;
  overflow-wrap: anywhere;
}

.schedule-display__code {
  display: inline-block;
  align-self: flex-start;
  max-width: 100%;
  padding: 0.1rem 0.5rem;
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  color: var(--color-text-muted);
  background: var(--color-bg-muted);
  border-radius: var(--radius-sm);
  letter-spacing: 0.02em;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

@media (width < 640px) {
  .schedule-row {
    flex-direction: column;
    align-items: stretch;
    gap: 0.35rem;
  }

  .schedule-row__label {
    width: auto;
    font-size: 0.85rem;
    color: var(--color-text-muted);
  }

  .schedule-row--top .schedule-row__label {
    padding-top: 0;
  }
}
</style>
