<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-[var(--color-text-primary)] font-bold text-lg">Ech0 Connect</h1>
          <p class="mt-1 text-sm text-[var(--color-text-muted)]">
            {{ t('connectSetting.description') }}
          </p>
        </div>
        <BaseButton
          v-if="!connectsEdit"
          class="top-action-btn top-action-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
          @click="connectsEdit = true"
        >
          {{ t('connectSetting.addConnect') }}
        </BaseButton>
      </div>

      <div
        v-if="connectsEdit"
        class="mt-4 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/40 p-4 text-[var(--color-text-secondary)]"
      >
        <div class="mb-1 text-sm text-[var(--color-text-primary)]">
          {{ t('connectSetting.connectUrl') }}
        </div>
        <div class="flex items-center gap-2">
          <BaseInput
            v-model="connectUrl"
            type="text"
            :placeholder="t('connectSetting.urlPlaceholder')"
            class="h-9 flex-1"
          />
        </div>
        <p v-if="connectUrlError" class="mt-1 text-xs text-[var(--color-danger)]">
          {{ connectUrlError }}
        </p>
        <p class="mt-1 text-xs text-[var(--color-text-muted)]">
          {{ t('connectSetting.connectHint') }}
        </p>

        <div class="mt-4 flex flex-nowrap items-center justify-center gap-2">
          <BaseButton
            class="h-9 rounded-md px-4 whitespace-nowrap"
            :disabled="isSubmitting"
            @click="handleCancelConnect"
          >
            {{ t('commonUi.cancel') }}
          </BaseButton>
          <BaseButton
            class="h-9 rounded-md px-4 whitespace-nowrap"
            :loading="isSubmitting"
            @click="handleAddConnect"
          >
            {{ t('connectSetting.connect') }}
          </BaseButton>
        </div>
      </div>

      <div v-if="connects.length === 0" class="flex flex-col items-center justify-center mt-4">
        <span class="text-[var(--color-text-muted)]">{{ t('connectSetting.empty') }}</span>
      </div>

      <div
        v-else
        class="mt-4 x-scrollbar overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
      >
        <table class="w-full min-w-[640px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-[56px] px-2 py-2 whitespace-nowrap">#</th>
              <th class="px-2 py-2 whitespace-nowrap">
                {{ t('connectSetting.connectUrl') }}
              </th>
              <th class="w-[110px] px-2 py-2 whitespace-nowrap">
                {{ t('connectSetting.status') }}
              </th>
              <th class="w-[120px] px-2 py-2 whitespace-nowrap">
                {{ t('connectSetting.version') }}
              </th>
              <th class="w-[88px] px-2 py-2 text-right whitespace-nowrap">
                {{ t('commonUi.actions') }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="(connect, index) in connects"
              :key="connect.id"
              class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
            >
              <td class="px-2 py-2 text-[var(--color-text-primary)]">{{ index + 1 }}</td>
              <td
                class="px-2 py-2 text-[var(--color-text-primary)] font-mono truncate"
                v-tooltip="connect.connect_url"
              >
                {{ connect.connect_url }}
              </td>
              <td class="px-1 py-2">
                <span :class="['status-pill', statusClass(connect.id)]">
                  {{ statusLabel(connect.id) }}
                </span>
              </td>
              <td class="px-1 py-2">
                <span
                  :class="['status-pill', 'font-mono', versionPillClass(connect.id)]"
                  :title="versionText(connect.id)"
                >
                  {{ versionText(connect.id) }}
                </span>
              </td>
              <td class="px-2 py-2 text-right">
                <BaseButton
                  class="h-8 w-8 !p-1.5"
                  :icon="Disconnect"
                  @click="handleDisconnect(connect.id)"
                  :tooltip="t('connectSetting.disconnect')"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import Disconnect from '@/components/icons/disconnect.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchAddConnect, fetchDeleteConnect, fetchGetConnectsHealth } from '@/service/api'
import { theToast } from '@/utils/toast'

import { useConnectStore } from '@/stores'
import { storeToRefs } from 'pinia'

import { useBaseDialog } from '@/composables/useBaseDialog'
const { openConfirm } = useBaseDialog()

const connectStore = useConnectStore()
const { t } = useI18n()
const { getConnect } = connectStore
const { connects } = storeToRefs(connectStore)
const connectsEdit = ref<boolean>(false)
const connectUrl = ref<string>('')
const connectUrlError = ref<string>('')
const isSubmitting = ref<boolean>(false)
const healthById = ref<Record<string, { status: 'online' | 'offline'; version: string }>>({})
const healthLoading = ref(false)

const isValidConnectUrl = (value: string) => {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const handleCancelConnect = () => {
  connectUrl.value = ''
  connectUrlError.value = ''
  connectsEdit.value = false
}

const rowStatus = (id: string): 'checking' | 'online' | 'offline' => {
  if (healthLoading.value) return 'checking'
  const row = healthById.value[id]
  if (!row) return 'checking'
  return row.status
}

const statusClass = (id: string) => {
  const current = rowStatus(id)
  if (current === 'online') return 'status-success'
  if (current === 'offline') return 'status-failed'
  return 'status-checking'
}

const statusLabel = (id: string) => {
  const current = rowStatus(id)
  if (current === 'online') return t('connectSetting.statusOnline')
  if (current === 'offline') return t('connectSetting.statusOffline')
  return t('connectSetting.statusChecking')
}

const versionText = (id: string) => {
  if (healthLoading.value) return '—'
  const row = healthById.value[id]
  if (!row || row.status !== 'online' || !row.version?.trim()) return '—'
  return row.version.trim()
}

const versionPillClass = (id: string) => {
  if (healthLoading.value) return 'version-pill-muted'
  const row = healthById.value[id]
  const hasVersion = row?.status === 'online' && Boolean(row.version?.trim())
  return hasVersion ? 'version-pill-yellow' : 'version-pill-muted'
}

const refreshConnectivityStatus = async () => {
  if (connects.value.length === 0) {
    healthById.value = {}
    return
  }
  healthLoading.value = true
  try {
    const res = await fetchGetConnectsHealth()
    if (res.code !== 1 || !res.data) {
      healthById.value = {}
      return
    }
    const next: Record<string, { status: 'online' | 'offline'; version: string }> = {}
    for (const row of res.data) {
      next[row.id] = {
        status: row.status === 'online' ? 'online' : 'offline',
        version: row.version ?? '',
      }
    }
    healthById.value = next
  } catch {
    healthById.value = {}
  } finally {
    healthLoading.value = false
  }
}

const refreshConnectData = async () => {
  await getConnect({ force: true })
  await refreshConnectivityStatus()
}

const handleAddConnect = async () => {
  if (isSubmitting.value) return
  connectUrlError.value = ''
  const target = connectUrl.value.trim()
  if (!target) {
    connectUrlError.value = String(t('connectSetting.enterAddress'))
    return
  }
  if (!isValidConnectUrl(target)) {
    connectUrlError.value = String(t('connectSetting.invalidUrl'))
    return
  }
  isSubmitting.value = true
  await fetchAddConnect(target)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        connectUrl.value = ''
        connectUrlError.value = ''
        connectsEdit.value = false
        refreshConnectData()
      }
    })
    .finally(() => {
      isSubmitting.value = false
    })
}

const handleDisconnect = async (connect_id: string) => {
  openConfirm({
    title: String(t('connectSetting.disconnectConfirmTitle')),
    description: '',
    onConfirm: async () => {
      await fetchDeleteConnect(connect_id).then((res) => {
        if (res.code === 1) {
          theToast.success(res.msg)
          refreshConnectData()
        }
      })
    },
  })
}

onMounted(() => {
  refreshConnectData()
})
</script>

<style scoped>
.top-action-btn {
  border: 1px solid var(--color-border-subtle) !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
}

.top-action-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.top-action-btn-primary {
  border-color: var(--color-border-strong) !important;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  padding: 2px 8px;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 500;
  white-space: nowrap;
  border: 1px solid transparent;
}

.status-success {
  color: #166534;
  border-color: rgb(34 197 94 / 35%);
  background: rgb(34 197 94 / 14%);
}

.status-failed {
  color: #991b1b;
  border-color: rgb(239 68 68 / 35%);
  background: rgb(239 68 68 / 14%);
}

.status-checking {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: var(--connect-status-checking-bg);
}

.version-pill-yellow,
.version-pill-muted {
  max-width: 100%;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: middle;
}

.version-pill-yellow {
  color: #a16207;
  border-color: rgb(234 179 8 / 40%);
  background: rgb(250 204 21 / 16%);
}

.version-pill-muted {
  color: var(--color-text-muted);
  border-color: rgb(234 179 8 / 22%);
  background: rgb(250 204 21 / 7%);
}
</style>
