<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-col gap-3 md:flex-row md:items-center md:justify-between">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">
            {{ t('webhookSetting.title') }}
          </h1>
          <p class="mt-1 text-sm text-[var(--color-text-muted)]">
            {{ t('webhookSetting.description') }}
          </p>
        </div>
        <div class="flex items-center gap-2">
          <button
            v-if="panelMode === 'manage'"
            type="button"
            class="inline-flex h-8 w-8 items-center justify-center rounded-[var(--btn-radius)] bg-transparent p-1.5 text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-[var(--color-bg-muted)]"
            v-tooltip="t('webhookSetting.openGuide')"
            @click="openGuide"
          >
            <InfoIcon class="h-full w-full" />
          </button>
          <BaseButton
            v-if="panelMode === 'guide'"
            class="top-action-btn top-action-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
            @click="backToManage"
          >
            {{ t('webhookSetting.backToManage') }}
          </BaseButton>
          <BaseButton
            v-if="panelMode === 'manage' && !isFormOpen"
            class="top-action-btn top-action-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
            @click="openCreateForm"
          >
            {{ t('webhookSetting.createWebhook') }}
          </BaseButton>
        </div>
      </div>

      <div
        v-if="panelMode === 'guide'"
        class="mt-4 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/40 p-4"
      >
        <div class="guide-hero">
          <p class="guide-kicker">{{ t('webhookSetting.guideKicker') }}</p>
          <h2 class="guide-title">{{ t('webhookSetting.guideTitle') }}</h2>
          <p class="guide-desc">{{ t('webhookSetting.guideDescription') }}</p>
        </div>

        <div class="mt-4 grid gap-3 lg:grid-cols-2">
          <section class="guide-section lg:col-span-2">
            <h3 class="guide-section-title">{{ t('webhookSetting.guideEventsTitle') }}</h3>
            <p class="guide-section-desc">{{ t('webhookSetting.guideEventsDesc') }}</p>
            <div class="mt-3 flex flex-wrap gap-2">
              <span v-for="topic in webhookGuideTopics" :key="topic" class="topic-chip">
                {{ topic }}
              </span>
            </div>
          </section>

          <section class="guide-section">
            <h3 class="guide-section-title">{{ t('webhookSetting.guideHeadersTitle') }}</h3>
            <ul class="guide-list mt-2">
              <li v-for="header in webhookGuideHeaders" :key="header.key">
                <code>{{ header.key }}</code>
                <span>{{ header.desc }}</span>
              </li>
            </ul>
          </section>

          <section class="guide-section">
            <h3 class="guide-section-title">{{ t('webhookSetting.guideBodyTitle') }}</h3>
            <ul class="guide-list mt-2">
              <li v-for="field in webhookGuideBodyFields" :key="field.key">
                <code>{{ field.key }}</code>
                <span>{{ field.desc }}</span>
              </li>
            </ul>
          </section>

          <section class="guide-section lg:col-span-2">
            <h3 class="guide-section-title">{{ t('webhookSetting.guideExampleTitle') }}</h3>
            <p class="guide-section-desc">{{ t('webhookSetting.guideExampleDesc') }}</p>
            <pre class="guide-code mt-3">{{ webhookPayloadExample }}</pre>
          </section>
        </div>
      </div>

      <div v-else>
        <div
          v-if="isFormOpen"
          class="mt-4 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/40 p-4"
        >
          <div class="grid gap-3 md:grid-cols-2">
            <div class="md:col-span-1">
              <div class="mb-1 text-sm text-[var(--color-text-primary)]">
                {{ t('webhookSetting.name') }}
              </div>
              <BaseInput
                v-model="webhookForm.name"
                class="w-full"
                :placeholder="t('webhookSetting.namePlaceholder')"
              />
              <p v-if="formErrors.name" class="mt-1 text-xs text-[var(--color-danger)]">
                {{ formErrors.name }}
              </p>
            </div>

            <div class="md:col-span-1">
              <div class="mb-1 text-sm text-[var(--color-text-primary)]">
                {{ t('webhookSetting.url') }}
              </div>
              <BaseInput
                v-model="webhookForm.url"
                class="w-full font-mono"
                :placeholder="t('webhookSetting.urlPlaceholder')"
              />
              <p v-if="formErrors.url" class="mt-1 text-xs text-[var(--color-danger)]">
                {{ formErrors.url }}
              </p>
            </div>

            <div class="md:col-span-2">
              <div class="mb-1 text-sm text-[var(--color-text-primary)]">
                {{ t('webhookSetting.secret') }}
              </div>
              <BaseInput
                v-model="webhookForm.secret"
                class="w-full font-mono"
                :placeholder="t('webhookSetting.secretPlaceholder')"
              />
            </div>

            <div
              class="md:col-span-2 flex items-center justify-between rounded-md border border-[var(--color-border-subtle)] px-3 py-2"
            >
              <div>
                <p class="text-sm text-[var(--color-text-primary)]">
                  {{ t('webhookSetting.enableWebhook') }}
                </p>
                <p class="text-xs text-[var(--color-text-muted)]">
                  {{ t('webhookSetting.enableWebhookHint') }}
                </p>
              </div>
              <BaseSwitch
                :model-value="webhookForm.is_active"
                @update:model-value="onFormActiveChange"
              />
            </div>
          </div>

          <div class="mt-4 flex flex-nowrap items-center justify-center gap-2">
            <BaseButton class="h-9 rounded-md px-4" :disabled="formSubmitting" @click="resetForm">
              {{ t('commonUi.cancel') }}
            </BaseButton>
            <BaseButton class="h-9 rounded-md px-4" :loading="formSubmitting" @click="submitForm">
              {{ isEditMode ? t('webhookSetting.saveWebhook') : t('webhookSetting.createWebhook') }}
            </BaseButton>
          </div>
        </div>

        <div class="mt-4">
          <div
            v-if="webhooksLoading"
            class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-6 text-center text-sm text-[var(--color-text-muted)]"
          >
            {{ t('webhookSetting.loading') }}
          </div>

          <div
            v-else-if="webhooksError"
            class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-6 text-center"
          >
            <p class="text-sm text-[var(--color-danger)]">
              {{ t('webhookSetting.loadFailed') }}
            </p>
            <p class="mt-1 text-xs text-[var(--color-text-muted)]">
              {{ webhooksError }}
            </p>
            <BaseButton class="mt-3 h-8 rounded-md px-3" @click="refreshWebhooks">
              {{ t('webhookSetting.retryLoad') }}
            </BaseButton>
          </div>

          <div
            v-else-if="Webhooks.length === 0"
            class="rounded-lg border border-[var(--color-border-subtle)] px-4 py-8 text-center text-sm text-[var(--color-text-muted)]"
          >
            {{ t('webhookSetting.empty') }}
          </div>

          <div
            v-else
            class="x-scrollbar overflow-x-auto rounded-lg border border-[var(--color-border-subtle)]"
          >
            <table class="w-full min-w-[644px] table-fixed text-sm">
              <thead>
                <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
                  <th class="w-[44px] px-2 py-2 whitespace-nowrap">#</th>
                  <th class="w-[100px] px-2 py-2 whitespace-nowrap">
                    {{ t('webhookSetting.name') }}
                  </th>
                  <th class="w-[190px] px-2 py-2 whitespace-nowrap">URL</th>
                  <th class="w-[88px] px-1 py-2 whitespace-nowrap">
                    {{ t('webhookSetting.lastStatus') }}
                  </th>
                  <th class="w-[72px] px-2 py-2 whitespace-nowrap">
                    {{ t('webhookSetting.enableWebhook') }}
                  </th>
                  <th class="w-[76px] px-1 py-2 text-right whitespace-nowrap">
                    {{ t('commonUi.actions') }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="(webhook, index) in Webhooks"
                  :key="webhook.id"
                  class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
                >
                  <td class="px-2 py-3 text-[var(--color-text-primary)]">{{ index + 1 }}</td>
                  <td class="px-2 py-3 text-[var(--color-text-primary)]">
                    {{ webhook.name }}
                  </td>
                  <td
                    class="max-w-[260px] truncate px-2 py-3 font-mono text-[var(--color-text-primary)]"
                    v-tooltip="webhook.url"
                  >
                    {{ webhook.url }}
                  </td>
                  <td class="px-2 py-3">
                    <span class="status-pill" :class="statusClass(webhook.last_status)">
                      {{ statusLabel(webhook.last_status) }}
                    </span>
                  </td>
                  <td class="py-3">
                    <BaseSwitch
                      :model-value="webhook.is_active"
                      :disabled="isRowBusy(webhook.id)"
                      @update:model-value="() => handleToggleWebhook(webhook)"
                    />
                  </td>
                  <td class="px-2 py-3">
                    <div class="flex items-center justify-end gap-1">
                      <BaseButton
                        class="h-8 w-8 !p-1.5"
                        :icon="EditIcon"
                        :disabled="isRowBusy(webhook.id)"
                        :tooltip="t('commonUi.edit')"
                        @click="openEditForm(webhook)"
                      />
                      <BaseButton
                        class="h-8 w-8 !p-1.5"
                        :icon="Trashbin"
                        :disabled="isRowBusy(webhook.id)"
                        :tooltip="t('webhookSetting.delete')"
                        @click="handleDeleteWebhook(webhook.id)"
                      />
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import BaseButton from '@/components/common/BaseButton.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import EditIcon from '@/components/icons/edit.vue'
import InfoIcon from '@/components/icons/info.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import PanelCard from '@/layout/PanelCard.vue'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { fetchCreateWebhook, fetchDeleteWebhook, fetchUpdateWebhook } from '@/service/api'
import { useSettingStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'

type FormMode = 'create' | 'edit'
type PanelMode = 'manage' | 'guide'

const { t } = useI18n()
const { openConfirm } = useBaseDialog()
const settingStore = useSettingStore()
const { Webhooks, webhooksError, webhooksLoading } = storeToRefs(settingStore)

const panelMode = ref<PanelMode>('manage')
const isFormOpen = ref(false)
const formMode = ref<FormMode>('create')
const editingWebhookId = ref<string | null>(null)
const formSubmitting = ref(false)
const togglingId = ref<string | null>(null)
const deletingId = ref<string | null>(null)

const webhookForm = ref<App.Api.Setting.WebhookDto>({
  name: '',
  url: '',
  secret: '',
  is_active: true,
})
const formErrors = ref<{ name: string; url: string }>({
  name: '',
  url: '',
})

const isEditMode = computed(() => formMode.value === 'edit')
const webhookGuideTopics = [
  'user.created',
  'user.updated',
  'user.deleted',
  'echo.created',
  'echo.updated',
  'echo.deleted',
  'comment.created',
  'comment.status.updated',
  'comment.deleted',
  'resource.uploaded',
  'system.snapshot',
  'system.export',
  'system.snapshot_schedule.updated',
]
const webhookGuideHeaders = computed(() => [
  { key: 'X-Ech0-Event', desc: String(t('webhookSetting.guideHeaderEvent')) },
  { key: 'X-Ech0-Event-ID', desc: String(t('webhookSetting.guideHeaderEventId')) },
  { key: 'X-Ech0-Timestamp', desc: String(t('webhookSetting.guideHeaderTimestamp')) },
  { key: 'X-Ech0-Signature', desc: String(t('webhookSetting.guideHeaderSignature')) },
  { key: 'User-Agent', desc: String(t('webhookSetting.guideHeaderUserAgent')) },
])
const webhookGuideBodyFields = computed(() => [
  { key: 'topic', desc: String(t('webhookSetting.guideBodyTopic')) },
  { key: 'event_name', desc: String(t('webhookSetting.guideBodyEventName')) },
  { key: 'payload_raw', desc: String(t('webhookSetting.guideBodyPayloadRaw')) },
  { key: 'metadata', desc: String(t('webhookSetting.guideBodyMetadata')) },
  { key: 'occurred_at', desc: String(t('webhookSetting.guideBodyOccurredAt')) },
])
const webhookPayloadExample = `{
  "topic": "echo.created",
  "event_name": "EchoCreatedEvent",
  "payload_raw": {
    "echo": {
      "id": "018f5e24-0fb7-7af0-a31b-a7ac0ad5e731",
      "content": "Hello from Ech0 webhook"
    },
    "user": {
      "id": "018f5e24-12f7-70e5-8a87-cf03a12bf10c",
      "username": "admin"
    }
  },
  "metadata": null,
  "occurred_at": 1710000000
}`
const onFormActiveChange = (value: boolean) => {
  webhookForm.value.is_active = value
}

const openGuide = () => {
  panelMode.value = 'guide'
  isFormOpen.value = false
}

const backToManage = () => {
  panelMode.value = 'manage'
}

const resetForm = () => {
  isFormOpen.value = false
  formMode.value = 'create'
  editingWebhookId.value = null
  formErrors.value = { name: '', url: '' }
  webhookForm.value = {
    name: '',
    url: '',
    secret: '',
    is_active: true,
  }
}

const openCreateForm = () => {
  resetForm()
  isFormOpen.value = true
}

const openEditForm = (webhook: App.Api.Setting.Webhook) => {
  isFormOpen.value = true
  formMode.value = 'edit'
  editingWebhookId.value = webhook.id
  formErrors.value = { name: '', url: '' }
  webhookForm.value = {
    name: webhook.name,
    url: webhook.url,
    secret: '',
    is_active: webhook.is_active,
  }
}

const refreshWebhooks = async () => {
  await settingStore.getAllWebhooks()
}

const isValidWebhookUrl = (value: string) => {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

const validateForm = () => {
  formErrors.value = { name: '', url: '' }
  let valid = true
  if (!webhookForm.value.name.trim()) {
    formErrors.value.name = String(t('webhookSetting.fieldNameRequired'))
    valid = false
  }
  if (!webhookForm.value.url.trim()) {
    formErrors.value.url = String(t('webhookSetting.fieldUrlRequired'))
    valid = false
  } else if (!isValidWebhookUrl(webhookForm.value.url)) {
    formErrors.value.url = String(t('webhookSetting.invalidUrl'))
    valid = false
  }
  return valid
}

const submitForm = async () => {
  if (formSubmitting.value) return
  if (!validateForm()) return
  formSubmitting.value = true
  const payload: App.Api.Setting.WebhookDto = {
    name: webhookForm.value.name.trim(),
    url: webhookForm.value.url.trim(),
    secret: webhookForm.value.secret?.trim() || '',
    is_active: webhookForm.value.is_active,
  }
  try {
    const res =
      isEditMode.value && editingWebhookId.value
        ? await fetchUpdateWebhook(editingWebhookId.value, payload)
        : await fetchCreateWebhook(payload)
    if (res.code === 1) {
      theToast.success(
        String(t(isEditMode.value ? 'webhookSetting.updateSuccess' : 'webhookSetting.addSuccess')),
      )
      await refreshWebhooks()
      resetForm()
      return
    }
    theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
  } finally {
    formSubmitting.value = false
  }
}

const isRowBusy = (webhookId: string) =>
  togglingId.value === webhookId || deletingId.value === webhookId

const handleToggleWebhook = async (webhook: App.Api.Setting.Webhook) => {
  if (isRowBusy(webhook.id)) return
  togglingId.value = webhook.id
  try {
    const res = await fetchUpdateWebhook(webhook.id, {
      name: webhook.name,
      url: webhook.url,
      secret: '',
      is_active: !webhook.is_active,
    })
    if (res.code === 1) {
      theToast.success(String(t('webhookSetting.updateSuccess')))
      await refreshWebhooks()
      return
    }
    theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
  } finally {
    togglingId.value = null
  }
}

const handleDeleteWebhook = (id: string) => {
  if (isRowBusy(id)) return
  openConfirm({
    title: String(t('webhookSetting.deleteConfirmTitle')),
    description: String(t('webhookSetting.deleteConfirmDesc')),
    onConfirm: async () => {
      deletingId.value = id
      try {
        const res = await fetchDeleteWebhook(id)
        if (res.code === 1) {
          theToast.success(String(t('webhookSetting.deleteSuccess')))
          await refreshWebhooks()
          return
        }
        theToast.error(String(res.msg || t('webhookSetting.operateFailed')))
      } finally {
        deletingId.value = null
      }
    },
  })
}

const statusLabel = (status: string) => {
  if (status === 'success') return String(t('webhookSetting.statusSuccess'))
  if (status === 'failed') return String(t('webhookSetting.statusFailed'))
  return String(t('webhookSetting.statusUnknown'))
}

const statusClass = (status: string) => {
  if (status === 'success') return 'status-success'
  if (status === 'failed') return 'status-failed'
  return 'status-unknown'
}

onMounted(async () => {
  await refreshWebhooks()
})
</script>

<style scoped>
.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.125rem 0.5rem;
  font-size: 0.75rem;
  line-height: 1rem;
}

.status-success {
  color: rgb(34 197 94);
  background: rgb(34 197 94 / 12%);
}

.status-failed {
  color: rgb(239 68 68);
  background: rgb(239 68 68 / 12%);
}

.status-unknown {
  color: rgb(245 158 11);
  background: rgb(245 158 11 / 14%);
}

.guide-hero {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  padding: 0.9rem 1rem;
  background: var(--webhook-guide-hero-bg);
  min-width: 0;
}

.guide-kicker {
  font-size: 12px;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  white-space: normal;
  overflow-wrap: anywhere;
}

.guide-title {
  margin-top: 2px;
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.guide-desc {
  margin-top: 0.25rem;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  overflow-wrap: anywhere;
}

.guide-section {
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.75rem;
  padding: 0.85rem;
  background: var(--webhook-guide-section-bg);
  min-width: 0;
}

.guide-section-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--color-text-primary);
  overflow-wrap: anywhere;
}

.guide-section-desc {
  margin-top: 0.2rem;
  font-size: 0.8rem;
  color: var(--color-text-muted);
  overflow-wrap: anywhere;
}

.topic-chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  border-radius: 9999px;
  border: 1px solid var(--color-border-subtle);
  padding: 0.2rem 0.55rem;
  font-size: 12px;
  line-height: 1rem;
  color: var(--color-text-secondary);
  background: var(--webhook-topic-chip-bg);
  white-space: normal;
  overflow-wrap: anywhere;
}

.guide-list {
  display: grid;
  gap: 0.42rem;
}

.guide-list li {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  align-items: flex-start;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  min-width: 0;
  overflow-wrap: anywhere;
}

.guide-list code {
  border-radius: 0.4rem;
  border: 1px solid var(--color-border-subtle);
  padding: 0.08rem 0.36rem;
  font-size: 0.75rem;
  line-height: 1.1rem;
  color: var(--color-text-primary);
  background: var(--webhook-guide-list-code-bg);
  white-space: normal;
  overflow-wrap: anywhere;
}

.guide-checklist {
  display: grid;
  gap: 0.4rem;
}

.guide-checklist li {
  position: relative;
  padding-left: 0.85rem;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.guide-checklist li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.45rem;
  width: 0.35rem;
  height: 0.35rem;
  border-radius: 9999px;
  background: var(--webhook-checklist-dot-bg);
}

.guide-code {
  border-radius: 0.6rem;
  border: 1px solid var(--color-border-subtle);
  padding: 0.75rem;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.4;
  color: var(--color-text-secondary);
  background: var(--webhook-guide-code-bg);
}

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
</style>
