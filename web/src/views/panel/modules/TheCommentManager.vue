<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2 comment-manager-page">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <PanelCard v-if="tab === 'setting'">
      <div class="mb-4 flex items-center justify-between gap-3">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">
            {{ t('commentManager.title') }}
          </h1>
          <p class="text-xs text-[var(--color-text-muted)]">
            {{ t('commentManager.subtitle') }}
          </p>
        </div>
        <div class="flex items-center gap-2 shrink-0">
          <BaseButton
            class="comment-btn comment-btn-primary whitespace-nowrap px-2.5 py-1 text-xs"
            @click="saveSetting"
            :disabled="settingSaving"
          >
            {{ settingSaving ? t('common.saving') : t('common.save') }}
          </BaseButton>
          <BaseButton
            v-if="setting.email_notify.enabled"
            class="comment-btn whitespace-nowrap px-2.5 py-1 text-xs"
            @click="testEmail"
            :disabled="settingSaving || testingEmail"
          >
            {{
              testingEmail ? t('commentManager.testEmailSending') : t('commentManager.testEmail')
            }}
          </BaseButton>
        </div>
      </div>

      <div class="space-y-2">
        <div class="setting-row">
          <div>
            <h3 class="setting-title">{{ t('commentManager.enableCommentTitle') }}</h3>
            <p class="setting-desc">{{ t('commentManager.enableCommentDesc') }}</p>
          </div>
          <BaseSwitch v-model="setting.enable_comment" />
        </div>
        <div class="setting-row">
          <div>
            <h3 class="setting-title">{{ t('commentManager.requireApprovalTitle') }}</h3>
            <p class="setting-desc">{{ t('commentManager.requireApprovalDesc') }}</p>
          </div>
          <BaseSwitch v-model="setting.require_approval" :disabled="!setting.enable_comment" />
        </div>
        <div class="setting-row">
          <div>
            <h3 class="setting-title">{{ t('commentManager.enableCaptchaTitle') }}</h3>
            <p class="setting-desc">{{ t('commentManager.enableCaptchaDesc') }}</p>
          </div>
          <BaseSwitch v-model="setting.captcha_enabled" :disabled="!setting.enable_comment" />
        </div>

        <div class="mt-3">
          <div class="setting-row">
            <div>
              <h3 class="setting-title">{{ t('commentManager.emailNotifyTitle') }}</h3>
              <p class="setting-desc">{{ t('commentManager.emailNotifyDesc') }}</p>
            </div>
            <BaseSwitch v-model="setting.email_notify.enabled" />
          </div>
          <div v-if="setting.email_notify.enabled" class="mt-3 grid gap-2 md:grid-cols-2">
            <BaseInput
              v-model.trim="setting.email_notify.smtp_host"
              :placeholder="t('commentManager.smtpHostPlaceholder')"
            />
            <BaseInput
              v-model.number="setting.email_notify.smtp_port"
              type="number"
              :placeholder="t('commentManager.smtpPortPlaceholder')"
            />
            <BaseInput
              v-model.trim="setting.email_notify.smtp_username"
              :placeholder="t('commentManager.smtpUsernamePlaceholder')"
            />
            <BaseInput
              v-model.trim="setting.email_notify.smtp_sender"
              :placeholder="t('commentManager.smtpSenderPlaceholder')"
            />
            <p class="md:col-span-2 text-xs text-[var(--color-text-muted)]">
              {{ t('commentManager.smtpSenderHint') }}
            </p>
            <BaseInput
              v-model="setting.email_notify.smtp_password"
              type="password"
              :placeholder="
                setting.email_notify.smtp_password_set
                  ? t('commentManager.smtpPasswordKeepPlaceholder')
                  : t('commentManager.smtpPasswordPlaceholder')
              "
            />
            <p
              v-if="setting.email_notify.smtp_password_set && !setting.email_notify.smtp_password"
              class="md:col-span-2 text-xs text-[var(--color-text-muted)]"
            >
              {{ t('commentManager.smtpPasswordSavedHint') }}
            </p>
          </div>
        </div>
      </div>
    </PanelCard>

    <PanelCard v-else>
      <div class="mb-4">
        <h1 class="text-lg font-bold text-[var(--color-text-primary)]">
          {{ t('commentManager.listTitle') }}
        </h1>
        <p class="text-xs text-[var(--color-text-muted)]">
          {{ t('commentManager.listSubtitle') }}
        </p>
      </div>
      <div class="mb-3 flex flex-wrap items-center gap-2">
        <BaseInput
          v-model.trim="query.keyword"
          type="text"
          class="min-w-56 md:w-64"
          :placeholder="t('commentManager.searchPlaceholder')"
        />
        <BaseSelect
          v-model="query.status"
          class="h-9 min-w-28"
          :options="statusOptions"
          :placeholder="t('commentManager.statusAll')"
        >
        </BaseSelect>
        <BaseSelect
          v-model="hotFilter"
          class="h-9 min-w-28"
          :options="hotOptions"
          :placeholder="t('commentManager.hotFilter')"
        >
        </BaseSelect>
        <BaseButton class="comment-btn px-3 py-1.5 text-sm" @click="reload">
          {{ t('commentManager.query') }}
        </BaseButton>
        <BaseButton
          class="comment-btn px-3 py-1.5 text-sm"
          @click="runBatch('approve')"
          :disabled="selectedIds.length === 0"
        >
          {{ t('commentManager.batchApprove') }}
        </BaseButton>
        <BaseButton
          class="comment-btn px-3 py-1.5 text-sm"
          @click="runBatch('reject')"
          :disabled="selectedIds.length === 0"
        >
          {{ t('commentManager.batchReject') }}
        </BaseButton>
        <BaseButton
          class="comment-btn-danger px-3 py-1.5 text-sm"
          @click="runBatch('delete')"
          :disabled="selectedIds.length === 0"
        >
          {{ t('commentManager.batchDelete') }}
        </BaseButton>
      </div>

      <div
        class="x-scrollbar overflow-x-auto rounded-lg border border-[var(--color-border-subtle)]"
      >
        <table class="w-full min-w-[700px] text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-10 px-2 py-2 align-middle">
                <div class="flex items-center justify-center">
                  <input v-model="allChecked" type="checkbox" />
                </div>
              </th>
              <th class="min-w-[100px] px-2 py-2">{{ t('commentManager.nickname') }}</th>
              <th class="min-w-[160px] px-2 py-2">{{ t('commentManager.email') }}</th>
              <th class="min-w-[68px] px-2 py-2">{{ t('commentManager.status') }}</th>
              <th class="min-w-[68px] px-2 py-2">{{ t('commentManager.hotColumn') }}</th>
              <th class="min-w-[200px] px-2 py-2">{{ t('commonUi.actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in list.items"
              :key="item.id"
              class="border-t border-[var(--color-border-subtle)] text-[var(--color-text-secondary)]"
            >
              <td class="px-2 py-2 align-middle">
                <div class="flex items-center justify-center">
                  <input v-model="selectedIds" type="checkbox" :value="item.id" />
                </div>
              </td>
              <td class="px-2 py-2">{{ item.nickname }}</td>
              <td class="px-2 py-2">{{ item.email }}</td>
              <td class="px-1 py-2">
                <span class="status-pill" :class="statusClass(item.status)">
                  {{ statusLabelMap[item.status] || item.status }}
                </span>
              </td>
              <td class="px-1 py-2">
                <span class="status-pill" :class="hotClass(item.hot)">
                  {{ item.hot ? t('commentManager.hotPicked') : t('commentManager.hotNormal') }}
                </span>
              </td>
              <td class="px-2 py-2">
                <div class="flex items-center gap-2">
                  <button class="table-action text-cyan-500" @click="openEcho(item.echo_id)">
                    {{ t('commentManager.view') }}
                  </button>
                  <button class="table-action text-sky-500" @click="openDetail(item.id)">
                    {{ t('commentManager.detail') }}
                  </button>
                  <button
                    class="table-action text-violet-500"
                    @click="updateHot(item.id, !item.hot)"
                  >
                    {{ item.hot ? t('commentManager.hotUnset') : t('commentManager.hotSet') }}
                  </button>
                  <button
                    class="table-action text-emerald-500"
                    @click="updateStatus(item.id, 'approved')"
                  >
                    {{ t('commentManager.approve') }}
                  </button>
                  <button
                    class="table-action text-amber-500"
                    @click="updateStatus(item.id, 'rejected')"
                  >
                    {{ t('commentManager.reject') }}
                  </button>
                  <button class="table-action text-red-500" @click="remove(item.id)">
                    {{ t('commentManager.delete') }}
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="list.items.length === 0">
              <td colspan="6" class="px-3 py-8 text-center text-[var(--color-text-muted)]">
                {{ t('commentManager.empty') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="mt-3 flex items-center justify-between text-sm">
        <span class="text-[var(--color-text-muted)]">{{
          t('commentManager.total', { total: list.total })
        }}</span>
        <div class="flex items-center gap-2">
          <BaseButton
            class="comment-btn px-2 py-1 text-sm"
            :disabled="query.page <= 1"
            @click="prevPage"
          >
            {{ t('commentManager.prevPage') }}
          </BaseButton>
          <span class="text-[var(--color-text-secondary)]">{{
            t('commentManager.pageInfo', { page: query.page, total: totalPages })
          }}</span>
          <BaseButton
            class="comment-btn px-2 py-1 text-sm"
            :disabled="query.page * query.page_size >= list.total"
            @click="nextPage"
          >
            {{ t('commentManager.nextPage') }}
          </BaseButton>
        </div>
      </div>
    </PanelCard>

    <div
      v-if="detailOpen && current"
      class="fixed inset-0 z-30 flex items-center justify-center bg-black/30 p-3"
      @click.self="detailOpen = false"
    >
      <div
        class="comment-detail w-full max-w-lg overflow-hidden rounded-xl border border-[var(--color-border-subtle)] bg-[var(--color-bg-canvas)] shadow-[var(--shadow-md)]"
      >
        <div class="comment-detail__header">
          <h3 class="comment-detail__title">{{ t('commentManager.detailTitle') }}</h3>
          <button
            class="comment-detail__close"
            :title="t('commentManager.close')"
            :aria-label="t('commentManager.close')"
            @click="detailOpen = false"
          >
            ✕
          </button>
        </div>

        <div class="comment-detail__body">
          <div class="comment-detail__author">
            <BaseAvatar
              :seed="avatarSeed(current)"
              :size="40"
              :alt="current.nickname"
              class="comment-detail__avatar"
            />
            <div class="min-w-0">
              <p class="comment-detail__name">{{ current.nickname }}</p>
              <div class="comment-detail__badges">
                <span class="status-pill" :class="statusClass(current.status)">
                  {{ statusLabelMap[current.status] || current.status }}
                </span>
                <span v-if="current.hot" class="status-pill status-hot">
                  {{ t('commentManager.hotPicked') }}
                </span>
              </div>
            </div>
          </div>

          <p class="comment-detail__content">{{ current.content }}</p>

          <dl class="comment-detail__meta">
            <div class="comment-detail__meta-row">
              <dt>{{ t('commentManager.email') }}</dt>
              <dd>{{ current.email || '-' }}</dd>
            </div>
            <div class="comment-detail__meta-row">
              <dt>{{ t('commentManager.website') }}</dt>
              <dd>{{ current.website || '-' }}</dd>
            </div>
            <div class="comment-detail__meta-row">
              <dt>{{ t('commentManager.source') }}</dt>
              <dd>{{ current.source || '-' }}</dd>
            </div>
            <div class="comment-detail__meta-row">
              <dt>{{ t('commentManager.time') }}</dt>
              <dd>{{ formatDate(current.created_at) }}</dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import {
  fetchBatchPanelComments,
  fetchDeletePanelComment,
  fetchGetCommentSystemSetting,
  fetchGetPanelCommentById,
  fetchGetPanelComments,
  fetchTestCommentEmail,
  fetchUpdateCommentSystemSetting,
  fetchUpdatePanelCommentHot,
  fetchUpdatePanelCommentStatus,
} from '@/service/api'
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import BaseAvatar from '@/components/common/BaseAvatar.vue'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'

const router = useRouter()
const { t } = useI18n()

const tab = ref('setting')
const tabOptions = computed(() => [
  { label: String(t('commentManager.tabSetting')), value: 'setting' },
  { label: String(t('commentManager.tabManage')), value: 'manage' },
])

const setting = reactive<App.Api.Comment.SystemSetting>({
  enable_comment: true,
  require_approval: true,
  captcha_enabled: false,
  email_notify: {
    enabled: false,
    smtp_host: '',
    smtp_port: 587,
    smtp_username: '',
    smtp_password: '',
    smtp_password_set: false,
    smtp_sender: '',
  },
})
const settingSaving = ref(false)
const testingEmail = ref(false)

const query = reactive<App.Api.Comment.PanelListQuery>({
  page: 1,
  page_size: 10,
  keyword: '',
  status: '',
  echo_id: '',
})

const list = reactive<App.Api.Comment.PanelPageResult>({
  items: [],
  total: 0,
})
const hotFilter = ref('')

const selectedIds = ref<string[]>([])
const detailOpen = ref(false)
const current = ref<App.Api.Comment.CommentItem | null>(null)
const statusOptions = computed(() => [
  { label: t('commentManager.statusAll'), value: '' },
  { label: t('commentManager.statusPending'), value: 'pending' },
  { label: t('commentManager.statusApproved'), value: 'approved' },
  { label: t('commentManager.statusRejected'), value: 'rejected' },
])
const hotOptions = computed(() => [
  { label: t('commentManager.hotAll'), value: '' },
  { label: t('commentManager.hotOnly'), value: 'true' },
  { label: t('commentManager.hotOnlyNot'), value: 'false' },
])
const statusLabelMap = computed<Record<string, string>>(() => ({
  pending: String(t('commentManager.statusPending')),
  approved: String(t('commentManager.statusApproved')),
  rejected: String(t('commentManager.statusRejected')),
}))
const totalPages = computed(() => Math.max(1, Math.ceil(list.total / query.page_size)))

const allChecked = computed({
  get() {
    return list.items.length > 0 && selectedIds.value.length === list.items.length
  },
  set(v: boolean) {
    if (v) {
      selectedIds.value = list.items.map((i) => i.id)
    } else {
      selectedIds.value = []
    }
  },
})

const loadSetting = async () => {
  const res = await fetchGetCommentSystemSetting()
  if (res.code === 1) {
    Object.assign(setting, res.data)
    setting.email_notify = {
      ...setting.email_notify,
      ...(res.data.email_notify || {}),
      smtp_password: '',
    }
  }
}

const saveSetting = async () => {
  settingSaving.value = true
  try {
    const res = await fetchUpdateCommentSystemSetting(buildSettingPayload())
    if (res.code === 1) {
      theToast.success(String(t('commentManager.settingUpdated')))
    }
  } finally {
    settingSaving.value = false
  }
}

const testEmail = async () => {
  if (!setting.email_notify.enabled) {
    theToast.info(String(t('commentManager.emailNotifyEnableFirst')))
    return
  }
  testingEmail.value = true
  try {
    const res = await fetchTestCommentEmail(buildSettingPayload())
    if (res.code === 1) {
      theToast.success(String(t('commentManager.testEmailSuccess')))
    }
  } finally {
    testingEmail.value = false
  }
}

const buildSettingPayload = (): App.Api.Comment.SystemSetting => {
  const rawPort = Number(setting.email_notify.smtp_port)
  const safePort = Number.isFinite(rawPort) && rawPort > 0 ? Math.trunc(rawPort) : 587
  return {
    enable_comment: setting.enable_comment,
    require_approval: setting.require_approval,
    captcha_enabled: setting.captcha_enabled,
    email_notify: {
      enabled: Boolean(setting.email_notify.enabled),
      smtp_host: String(setting.email_notify.smtp_host || '').trim(),
      smtp_port: safePort,
      smtp_username: String(setting.email_notify.smtp_username || '').trim(),
      smtp_password: String(setting.email_notify.smtp_password || ''),
      smtp_password_set: Boolean(setting.email_notify.smtp_password_set),
      smtp_sender: String(setting.email_notify.smtp_sender || '').trim(),
    },
  }
}

const loadList = async () => {
  const hot = hotFilter.value === '' ? undefined : hotFilter.value === 'true'
  const res = await fetchGetPanelComments({
    ...query,
    hot,
  })
  if (res.code === 1) {
    list.items = res.data.items || []
    list.total = res.data.total || 0
    selectedIds.value = []
  }
}

const reload = async () => {
  query.page = 1
  await loadList()
}

const runBatch = async (action: App.Api.Comment.BatchAction) => {
  if (selectedIds.value.length === 0) {
    theToast.info(String(t('commentManager.selectFirst')))
    return
  }
  const res = await fetchBatchPanelComments(action, selectedIds.value)
  if (res.code === 1) {
    theToast.success(String(t('commentManager.batchSuccess')))
    await loadList()
  }
}

const updateStatus = async (id: string, status: App.Api.Comment.CommentStatus) => {
  const res = await fetchUpdatePanelCommentStatus(id, status)
  if (res.code === 1) {
    await loadList()
  }
}

const updateHot = async (id: string, hot: boolean) => {
  const res = await fetchUpdatePanelCommentHot(id, hot)
  if (res.code === 1) {
    await loadList()
  }
}

const remove = async (id: string) => {
  const res = await fetchDeletePanelComment(id)
  if (res.code === 1) {
    await loadList()
  }
}

const openEcho = (echoId: string) => {
  const id = echoId?.trim()
  if (!id) {
    theToast.info(String(t('commentManager.echoIdMissing')))
    return
  }
  const target = router.resolve({
    name: 'echo',
    params: { echoId: id },
  })
  window.open(target.href, '_blank', 'noopener,noreferrer')
}

const openDetail = async (id: string) => {
  const res = await fetchGetPanelCommentById(id)
  if (res.code === 1) {
    current.value = res.data
    detailOpen.value = true
  }
}

const avatarSeed = (item: App.Api.Comment.CommentItem) =>
  `${item.id}-${item.nickname}-${item.source}`

const prevPage = async () => {
  if (query.page <= 1) return
  query.page -= 1
  await loadList()
}

const nextPage = async () => {
  if (query.page * query.page_size >= list.total) return
  query.page += 1
  await loadList()
}

const statusClass = (status: string) => {
  if (status === 'approved') return 'status-approved'
  if (status === 'rejected') return 'status-rejected'
  return 'status-pending'
}

const hotClass = (hot: boolean) => {
  if (hot) return 'status-hot'
  return 'status-normal'
}

onMounted(async () => {
  await Promise.all([loadSetting(), loadList()])
})
</script>

<style scoped>
.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  min-height: 2.6rem;
}

.setting-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.setting-desc {
  margin-top: 0.1rem;
  font-size: 0.78rem;
  color: var(--color-text-muted);
}

.comment-btn {
  border: 1px solid var(--color-border-subtle) !important;
  background: var(--color-bg-surface) !important;
  color: var(--color-text-secondary) !important;
}

.comment-btn:hover {
  border-color: var(--color-border-strong) !important;
  background: var(--color-bg-muted) !important;
}

.comment-btn-primary {
  border-color: var(--color-border-strong) !important;
}

.comment-btn-danger {
  border: 1px solid var(--comment-danger-border) !important;
  background: var(--comment-danger-bg) !important;
  color: #dc2626 !important;
}

.comment-btn-danger:hover {
  border-color: #f87171 !important;
  background: var(--comment-danger-hover-bg) !important;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  border-radius: 9999px;
  padding: 0.1rem 0.5rem;
  font-size: 0.75rem;
  line-height: 1.2;
  border: 1px solid transparent;
}

.status-approved {
  color: #059669;
  border-color: var(--comment-status-approved-border);
  background: var(--comment-status-approved-bg);
}

.status-rejected {
  color: #d97706;
  border-color: var(--comment-status-rejected-border);
  background: var(--comment-status-rejected-bg);
}

.status-pending {
  color: #0369a1;
  border-color: var(--comment-status-pending-border);
  background: var(--comment-status-pending-bg);
}

.status-hot {
  color: #7c3aed;
  border-color: var(--comment-status-hot-border);
  background: var(--comment-status-hot-bg);
}

.status-normal {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: var(--comment-status-normal-bg);
}

.table-action {
  transition: opacity 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.table-action:hover {
  opacity: 0.72;
}

.comment-detail__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.85rem 1.1rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.comment-detail__title {
  font-size: 0.95rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.comment-detail__close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 1.9rem;
  height: 1.9rem;
  flex-shrink: 0;
  border-radius: var(--radius-md);
  color: var(--color-text-muted);
  font-size: 0.9rem;
  line-height: 1;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.comment-detail__close:hover {
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
}

.comment-detail__body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 1.1rem;
}

.comment-detail__author {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.comment-detail__avatar {
  width: 2.5rem;
  height: 2.5rem;
  flex-shrink: 0;
  border-radius: 999px;
  object-fit: cover;
}

.comment-detail__name {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.comment-detail__badges {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.4rem;
  margin-top: 0.35rem;
}

.comment-detail__content {
  margin: 0;
  padding: 0.75rem 0.85rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-md);
  background: var(--color-bg-muted);
  color: var(--color-text-primary);
  font-size: 0.92rem;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  max-height: 16rem;
  overflow-y: auto;
}

.comment-detail__meta {
  margin: 0;
}

.comment-detail__meta-row {
  display: grid;
  grid-template-columns: 4.5rem 1fr;
  gap: 0.75rem;
  align-items: start;
  padding: 0.45rem 0;
  border-top: 1px solid var(--color-border-subtle);
}

.comment-detail__meta-row:first-child {
  border-top: none;
}

.comment-detail__meta-row dt {
  color: var(--color-text-muted);
  font-size: 0.82rem;
}

.comment-detail__meta-row dd {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 0.85rem;
  overflow-wrap: anywhere;
}

@media (width <= 768px) {
  .setting-row {
    align-items: flex-start;
  }
}
</style>
