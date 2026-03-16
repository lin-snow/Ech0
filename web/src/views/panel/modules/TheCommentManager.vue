<template>
  <div class="w-full px-2 comment-manager-page">
    <PanelCard class="mb-4">
      <div class="mb-4 flex items-center justify-between gap-3">
        <div>
          <h1 class="text-lg font-bold text-[var(--color-text-primary)]">
            {{ t('commentManager.title') }}
          </h1>
          <p class="text-xs text-[var(--color-text-muted)]">
            {{ t('commentManager.subtitle') }}
          </p>
        </div>
        <BaseButton
          class="comment-btn comment-btn-primary shrink-0 whitespace-nowrap px-2.5 py-1 text-xs"
          @click="saveSetting"
          :disabled="settingSaving"
        >
          {{ settingSaving ? t('common.saving') : t('common.save') }}
        </BaseButton>
      </div>

      <div
        class="space-y-2 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-surface)]/50 p-3"
      >
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
        <div class="setting-row">
          <div>
            <h3 class="setting-title">{{ t('commentManager.enableEmailNotificationTitle') }}</h3>
            <p class="setting-desc">{{ t('commentManager.enableEmailNotificationDesc') }}</p>
          </div>
          <BaseSwitch v-model="setting.enable_email_notification" :disabled="!setting.enable_comment" />
        </div>
      </div>

      <div v-if="setting.captcha_enabled" class="mt-3 grid gap-2">
        <BaseInput
          v-model.trim="setting.captcha_verify_url"
          type="text"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.captchaVerifyPlaceholder')"
        />
        <BaseInput
          v-model.trim="setting.captcha_secret"
          type="text"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.captchaSecretPlaceholder')"
        />
      </div>
      <div
        v-else
        class="mt-3 rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/60 px-3 py-2 text-xs text-[var(--color-text-muted)]"
      >
        {{ t('commentManager.captchaDisabledPrefix') }}
        <a
          href="https://github.com/tiagozip/cap"
          target="_blank"
          rel="noreferrer"
          class="underline underline-offset-2 hover:text-[var(--color-text-secondary)]"
        >
          cap
        </a>
        {{ t('commentManager.captchaDisabledSuffix') }}
      </div>

      <div v-if="setting.enable_email_notification" class="mt-3 grid gap-2">
        <BaseInput
          v-model.trim="setting.smtp_host"
          type="text"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.smtpHostPlaceholder')"
        />
        <BaseInput
          v-model.number="setting.smtp_port"
          type="number"
          min="1"
          max="65535"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.smtpPortPlaceholder')"
        />
        <BaseInput
          v-model.trim="setting.smtp_user"
          type="text"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.smtpUserPlaceholder')"
        />
        <BaseInput
          v-model.trim="setting.smtp_password"
          type="password"
          :disabled="!setting.enable_comment"
          :placeholder="t('commentManager.smtpPasswordPlaceholder')"
        />
      </div>
    </PanelCard>

    <PanelCard>
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
        <table class="w-full min-w-[820px] text-sm">
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
              <th class="min-w-[120px] px-2 py-2 whitespace-nowrap">
                {{ t('commentManager.time') }}
              </th>
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
              <td class="px-2 py-2 whitespace-nowrap">{{ formatDate(item.created_at) }}</td>
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
              <td colspan="7" class="px-3 py-8 text-center text-[var(--color-text-muted)]">
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
      class="fixed inset-0 z-30 flex items-end justify-center bg-black/30 p-3 md:items-center"
      @click.self="detailOpen = false"
    >
      <div
        class="w-full max-w-lg rounded-xl border border-[var(--color-border-subtle)] bg-[var(--color-bg-canvas)] p-4 shadow-[var(--shadow-md)]"
      >
        <div class="mb-2 flex items-center justify-between">
          <h3 class="text-base font-semibold text-[var(--color-text-primary)]">
            {{ t('commentManager.detailTitle') }}
          </h3>
          <button
            class="table-action text-sm text-[var(--color-text-muted)]"
            @click="detailOpen = false"
          >
            {{ t('commentManager.close') }}
          </button>
        </div>
        <div class="space-y-1 text-sm text-[var(--color-text-secondary)]">
          <p>
            <b>{{ t('commentManager.nickname') }}：</b>{{ current.nickname }}
          </p>
          <p>
            <b>{{ t('commentManager.email') }}：</b>{{ current.email }}
          </p>
          <p>
            <b>{{ t('commentManager.website') }}：</b>{{ current.website || '-' }}
          </p>
          <p>
            <b>{{ t('commentManager.status') }}：</b
            >{{ statusLabelMap[current.status] || current.status }}
          </p>
          <p>
            <b>{{ t('commentManager.hotColumn') }}：</b
            >{{ current.hot ? t('commentManager.yes') : t('commentManager.no') }}
          </p>
          <p>
            <b>{{ t('commentManager.source') }}：</b>{{ current.source }}
          </p>
          <p>
            <b>{{ t('commentManager.time') }}：</b>{{ formatDate(current.created_at) }}
          </p>
          <p
            class="mt-2 whitespace-pre-wrap break-words rounded-lg border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/50 p-2"
          >
            {{ current.content }}
          </p>
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
  fetchUpdateCommentSystemSetting,
  fetchUpdatePanelCommentHot,
  fetchUpdatePanelCommentStatus,
} from '@/service/api'
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { theToast } from '@/utils/toast'
import { formatDate } from '@/utils/other'

const router = useRouter()
const { t } = useI18n()

const setting = reactive<App.Api.Comment.SystemSetting>({
  enable_comment: true,
  require_approval: true,
  captcha_enabled: false,
  captcha_verify_url: '',
  captcha_secret: '',
  enable_email_notification: false,
  smtp_host: '',
  smtp_port: 465,
  smtp_user: '',
  smtp_password: '',
})
const settingSaving = ref(false)

const query = reactive<App.Api.Comment.PanelListQuery>({
  page: 1,
  page_size: 20,
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
  }
}

const saveSetting = async () => {
  settingSaving.value = true
  try {
    const res = await fetchUpdateCommentSystemSetting(setting)
    if (res.code === 1) {
      theToast.success(String(t('commentManager.settingUpdated')))
    }
  } finally {
    settingSaving.value = false
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
  border: 1px solid color-mix(in srgb, #ef4444 35%, var(--color-border-subtle) 65%) !important;
  background: color-mix(in srgb, #ef4444 9%, var(--color-bg-surface) 91%) !important;
  color: #dc2626 !important;
}

.comment-btn-danger:hover {
  border-color: #f87171 !important;
  background: color-mix(in srgb, #ef4444 14%, var(--color-bg-surface) 86%) !important;
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
  border-color: color-mix(in srgb, #10b981 45%, transparent 55%);
  background: color-mix(in srgb, #10b981 12%, transparent 88%);
}

.status-rejected {
  color: #d97706;
  border-color: color-mix(in srgb, #f59e0b 45%, transparent 55%);
  background: color-mix(in srgb, #f59e0b 14%, transparent 86%);
}

.status-pending {
  color: #0369a1;
  border-color: color-mix(in srgb, #38bdf8 45%, transparent 55%);
  background: color-mix(in srgb, #38bdf8 14%, transparent 86%);
}

.status-hot {
  color: #7c3aed;
  border-color: color-mix(in srgb, #8b5cf6 45%, transparent 55%);
  background: color-mix(in srgb, #8b5cf6 14%, transparent 86%);
}

.status-normal {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-muted) 65%, transparent 35%);
}

.table-action {
  transition: opacity 0.2s ease;
  white-space: nowrap;
  flex-shrink: 0;
}

.table-action:hover {
  opacity: 0.72;
}

@media (max-width: 768px) {
  .setting-row {
    align-items: flex-start;
  }
}
</style>
