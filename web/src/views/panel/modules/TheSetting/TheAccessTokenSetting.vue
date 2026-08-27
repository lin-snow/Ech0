<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-4">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('accessTokenSetting.title') }}
        </h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="accessTokenEdit"
            :apply-title="t('commonUi.done')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="accessTokenEdit = false"
            @toggle="accessTokenEdit = !accessTokenEdit"
          />
        </div>
      </div>
    </div>

    <div v-if="!accessTokenEdit">
      <div v-if="AccessTokens.length === 0" class="flex flex-col items-center justify-center mt-2">
        <span class="text-[var(--color-text-muted)]">{{ t('accessTokenSetting.empty') }}</span>
      </div>
      <div
        v-else
        class="mt-2 x-scrollbar overflow-x-auto border border-[var(--color-border-subtle)] rounded-lg"
      >
        <table class="w-full min-w-[864px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-[44px] px-2 py-2 whitespace-nowrap">#</th>
              <th class="w-[170px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.token') }}
              </th>
              <th class="w-[100px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.name') }}
              </th>
              <th class="w-[156px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.createdAt') }}
              </th>
              <th class="w-[156px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.expiry') }}
              </th>
              <th class="w-[120px] px-2 py-2 text-right whitespace-nowrap">
                {{ t('commonUi.actions') }}
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(tokenItem, index) in AccessTokens" :key="tokenItem.id">
              <td class="px-2 py-2 text-[var(--color-text-primary)]">{{ index + 1 }}</td>
              <td class="px-2 py-2 font-mono text-[var(--color-text-primary)]">
                <div class="flex items-center gap-1">
                  <span class="truncate" v-tooltip="tokenItem.token">{{
                    maskToken(tokenItem.token)
                  }}</span>
                  <button
                    class="p-1 hover:bg-[var(--color-bg-surface)] rounded"
                    @click="copyAccessToken(tokenItem.token)"
                    v-tooltip="t('accessTokenSetting.copyToken')"
                  >
                    <Clipboard class="w-4 h-4" />
                  </button>
                </div>
              </td>
              <td class="px-2 py-2 text-[var(--color-text-primary)]">
                <span v-tooltip="tokenItem.name" class="truncate block max-w-xs">{{
                  tokenItem.name
                }}</span>
              </td>
              <td class="px-1 py-2 text-[var(--color-text-secondary)] whitespace-nowrap">
                {{ new Date(tokenItem.created_at * 1000).toLocaleString() }}
              </td>
              <td class="px-1 py-2 text-[var(--color-text-secondary)] whitespace-nowrap">
                {{
                  tokenItem.expiry
                    ? new Date(tokenItem.expiry * 1000).toLocaleString()
                    : t('accessTokenSetting.neverExpire')
                }}
              </td>
              <td class="px-1 py-2">
                <div class="flex items-center justify-end gap-1">
                  <BaseButton
                    class="h-8 rounded-md px-2 text-xs whitespace-nowrap"
                    @click="openTokenDetail(tokenItem)"
                    :tooltip="t('accessTokenSetting.viewDetail')"
                  >
                    <span>{{ t('accessTokenSetting.detail') }}</span>
                  </BaseButton>
                  <BaseButton
                    class="h-8 w-8 !p-1.5"
                    :icon="Trashbin"
                    @click="handleDeleteAccessToken(tokenItem)"
                    :tooltip="t('accessTokenSetting.deleteToken')"
                  />
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-else class="text-[var(--color-text-secondary)]">
      <div class="rounded-lg border border-[var(--color-border-subtle)] p-4 space-y-4">
        <div class="space-y-2">
          <span class="text-[var(--color-text-primary)]">{{ t('accessTokenSetting.name') }}：</span>
          <BaseInput
            class="w-full"
            v-model="accessTokenToAdd.name"
            :placeholder="t('accessTokenSetting.namePlaceholder')"
          />
          <span v-if="nameError" class="text-xs text-[var(--color-danger)]">{{ nameError }}</span>
        </div>

        <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
          <div class="space-y-2">
            <span class="text-[var(--color-text-primary)]"
              >{{ t('accessTokenSetting.expiry') }}：</span
            >
            <BaseSelect
              v-model="accessTokenToAdd.expiry"
              :options="expirationOptions"
              class="w-full h-9 bg-[var(--color-bg-surface)]! bg-op-80"
            />
          </div>
          <div class="space-y-2">
            <span class="text-[var(--color-text-primary)]"
              >{{ t('accessTokenSetting.audience') }}：</span
            >
            <BaseSelect
              v-model="accessTokenToAdd.audience"
              :options="audienceOptions"
              class="w-full h-9 bg-[var(--color-bg-surface)]! bg-op-80"
              :placeholder="t('accessTokenSetting.audiencePlaceholder')"
            />
          </div>
        </div>

        <div class="space-y-2">
          <span class="text-[var(--color-text-primary)]"
            >{{ t('accessTokenSetting.scopes') }}：</span
          >
          <p class="text-xs text-[var(--color-text-muted)]">
            {{ t('accessTokenSetting.scopesHint') }}
          </p>
          <div
            v-for="group in scopeGroups"
            :key="group.labelKey"
            class="rounded-md border border-[var(--color-border-subtle)] p-3"
          >
            <div class="text-sm font-medium text-[var(--color-text-primary)] mb-2">
              {{ t(group.labelKey) }}
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="scope in group.items"
                :key="scope.value"
                type="button"
                class="px-2.5 py-1.5 rounded-md border text-xs transition"
                :class="
                  hasScope(scope.value)
                    ? 'border-[var(--color-accent)] text-[var(--color-accent)] bg-[var(--color-bg-muted)]'
                    : 'border-[var(--color-border-subtle)] text-[var(--color-text-secondary)] hover:border-[var(--color-border-strong)]'
                "
                @click="toggleScope(scope.value)"
              >
                {{ t(scope.labelKey) }}
              </button>
            </div>
          </div>
          <span v-if="scopesError" class="text-xs text-[var(--color-danger)]">{{
            scopesError
          }}</span>
        </div>

        <p class="text-xs text-[var(--color-text-muted)]">
          {{ t('accessTokenSetting.securityHint') }}
        </p>
      </div>

      <div class="flex items-center justify-center gap-2 mt-4">
        <BaseButton
          :disabled="isSubmitting"
          @click="handleCancelAddAccessToken"
          class="h-9 rounded-md px-4 bg-[var(--color-bg-surface)]! bg-op-80"
          :tooltip="t('accessTokenSetting.cancelAdd')"
        >
          <span>{{ t('commonUi.cancel') }}</span>
        </BaseButton>

        <BaseButton
          :loading="isSubmitting"
          @click="handleAddAccessToken"
          class="h-9 rounded-md px-4 bg-[var(--color-bg-surface)]! bg-op-80"
          :tooltip="t('accessTokenSetting.addToken')"
        >
          <span class="text-[var(--color-text-primary)]">{{ t('commonUi.add') }}</span>
        </BaseButton>
      </div>
    </div>

    <TransitionRoot appear :show="detailModalOpen" as="template">
      <Dialog as="div" class="relative z-5000" @close="closeTokenDetail">
        <TransitionChild
          as="template"
          enter="duration-200 ease-out"
          enter-from="opacity-0"
          enter-to="opacity-100"
          leave="duration-150 ease-in"
          leave-from="opacity-100"
          leave-to="opacity-0"
        >
          <div class="fixed inset-0 bg-black/30 backdrop-blur-sm" />
        </TransitionChild>

        <div class="fixed inset-0 overflow-y-auto">
          <div class="flex min-h-full items-center justify-center p-4">
            <TransitionChild
              as="template"
              enter="duration-200 ease-out"
              enter-from="opacity-0 scale-95"
              enter-to="opacity-100 scale-100"
              leave="duration-150 ease-in"
              leave-from="opacity-100 scale-100"
              leave-to="opacity-0 scale-95"
            >
              <DialogPanel
                class="w-full max-w-lg transform rounded-[var(--radius-lg)] bg-[var(--dialog-bg-color)] p-5 text-left align-middle shadow-[var(--shadow-md)] ring-1 ring-inset ring-[var(--color-border-subtle)] transition-all"
              >
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <DialogTitle class="text-base font-semibold text-[var(--color-text-primary)]">
                      {{ t('accessTokenSetting.detailTitle') }}
                    </DialogTitle>
                    <p class="mt-1 truncate text-sm text-[var(--color-text-muted)]">
                      {{ selectedToken?.name }}
                    </p>
                  </div>
                  <button
                    type="button"
                    class="-mt-1 -mr-1 shrink-0 cursor-pointer rounded-md p-1.5 text-[var(--color-text-muted)] transition-colors hover:bg-[var(--color-bg-muted)] hover:text-[var(--color-text-primary)]"
                    :aria-label="t('accessTokenSetting.closeDetail')"
                    @click="closeTokenDetail"
                  >
                    <Close class="h-4 w-4" />
                  </button>
                </div>

                <section
                  class="mt-4 rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/40 p-4"
                >
                  <h4
                    class="text-xs font-medium tracking-wide text-[var(--color-text-muted)] uppercase"
                  >
                    {{ t('accessTokenSetting.detailOverview') }}
                  </h4>
                  <dl class="mt-3 space-y-2.5 text-sm">
                    <div class="flex items-center justify-between gap-3">
                      <dt class="shrink-0 text-[var(--color-text-muted)]">
                        {{ t('accessTokenSetting.audience') }}
                      </dt>
                      <dd>
                        <span
                          class="inline-flex items-center rounded-full bg-[var(--color-accent)]/12 px-2.5 py-0.5 text-xs font-medium text-[var(--color-accent)] ring-1 ring-inset ring-[var(--color-accent)]/25"
                        >
                          {{ getAudienceLabel(selectedToken?.audience) }}
                        </span>
                      </dd>
                    </div>
                    <div class="flex items-center justify-between gap-3">
                      <dt class="shrink-0 text-[var(--color-text-muted)]">
                        {{ t('accessTokenSetting.expiry') }}
                      </dt>
                      <dd class="text-right text-[var(--color-text-primary)]">
                        {{ selectedTokenExpiryText }}
                      </dd>
                    </div>
                    <div class="flex items-center justify-between gap-3">
                      <dt class="shrink-0 text-[var(--color-text-muted)]">
                        {{ t('accessTokenSetting.createdAt') }}
                      </dt>
                      <dd class="text-right text-[var(--color-text-primary)]">
                        {{ selectedTokenCreatedText }}
                      </dd>
                    </div>
                  </dl>
                </section>

                <section
                  class="mt-3 rounded-[var(--radius-md)] border border-[var(--color-border-subtle)] bg-[var(--color-bg-muted)]/40 p-4"
                >
                  <h4
                    class="text-xs font-medium tracking-wide text-[var(--color-text-muted)] uppercase"
                  >
                    {{ t('accessTokenSetting.scopes') }}
                  </h4>
                  <div v-if="selectedTokenScopeGroups.length > 0" class="mt-3 space-y-3">
                    <div v-for="group in selectedTokenScopeGroups" :key="group.labelKey">
                      <div class="text-xs text-[var(--color-text-secondary)]">
                        {{ t(group.labelKey) }}
                      </div>
                      <div class="mt-1.5 flex flex-wrap gap-1.5">
                        <span
                          v-for="item in group.items"
                          :key="item.value"
                          class="rounded-md bg-[var(--color-bg-surface)] px-2 py-1 text-xs text-[var(--color-text-primary)] ring-1 ring-inset ring-[var(--color-border-subtle)]"
                        >
                          {{ getScopeLabel(item.value) }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <p v-else class="mt-2 text-sm text-[var(--color-text-muted)]">
                    {{ t('accessTokenSetting.scopeEmpty') }}
                  </p>
                </section>

                <div class="mt-4 flex justify-end">
                  <BaseButton class="h-9 rounded-md px-4" @click="closeTokenDetail">
                    {{ t('accessTokenSetting.closeDetail') }}
                  </BaseButton>
                </div>
              </DialogPanel>
            </TransitionChild>
          </div>
        </div>
      </Dialog>
    </TransitionRoot>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import Clipboard from '@/components/icons/clipboard.vue'
import Trashbin from '@/components/icons/trashbin.vue'
import Close from '@/components/icons/close.vue'
import { Dialog, DialogPanel, DialogTitle, TransitionChild, TransitionRoot } from '@headlessui/vue'
import { ref, onMounted, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { fetchCreateAccessToken, fetchDeleteAccessToken } from '@/service/api'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'
import { AccessTokenExpiration } from '@/enums/enums'

const { openConfirm } = useBaseDialog()
const { t } = useI18n()

const accessTokenEdit = ref<boolean>(false)
const useSetting = useSettingStore()
const { AccessTokens } = storeToRefs(useSetting)
const detailModalOpen = ref(false)
const selectedToken = ref<App.Api.Setting.AccessToken | null>(null)

const scopeGroups = [
  {
    labelKey: 'accessTokenSetting.scopeGroupContent',
    items: [
      { value: 'echo:read', labelKey: 'accessTokenSetting.scopeEchoRead' },
      { value: 'echo:write', labelKey: 'accessTokenSetting.scopeEchoWrite' },
      { value: 'comment:read', labelKey: 'accessTokenSetting.scopeCommentRead' },
      { value: 'comment:write', labelKey: 'accessTokenSetting.scopeCommentWrite' },
      { value: 'comment:moderate', labelKey: 'accessTokenSetting.scopeCommentModerate' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupFile',
    items: [
      { value: 'file:read', labelKey: 'accessTokenSetting.scopeFileRead' },
      { value: 'file:write', labelKey: 'accessTokenSetting.scopeFileWrite' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupConnect',
    items: [
      { value: 'connect:read', labelKey: 'accessTokenSetting.scopeConnectRead' },
      { value: 'connect:write', labelKey: 'accessTokenSetting.scopeConnectWrite' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupProfile',
    items: [
      { value: 'profile:read', labelKey: 'accessTokenSetting.scopeProfileRead' },
      { value: 'profile:write', labelKey: 'accessTokenSetting.scopeProfileWrite' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupAdmin',
    items: [
      { value: 'admin:settings', labelKey: 'accessTokenSetting.scopeAdminSettings' },
      { value: 'admin:user', labelKey: 'accessTokenSetting.scopeAdminUser' },
      { value: 'admin:token', labelKey: 'accessTokenSetting.scopeAdminToken' },
    ],
  },
] as const

const audienceOptions = computed(() => [
  {
    label: t('accessTokenSetting.audiencePublicClient'),
    value: 'public-client' as const,
  },
  {
    label: t('accessTokenSetting.audienceCli'),
    value: 'cli' as const,
  },
  {
    label: t('accessTokenSetting.audienceIntegration'),
    value: 'integration' as const,
  },
  {
    label: t('accessTokenSetting.audienceMcpRemote'),
    value: 'mcp-remote' as const,
  },
])

const expirationOptions = computed(() => [
  {
    label: t('accessTokenSetting.expiryOption8Hours'),
    value: AccessTokenExpiration.EIGHT_HOUR_EXPIRY,
  },
  {
    label: t('accessTokenSetting.expiryOption1Month'),
    value: AccessTokenExpiration.ONE_MONTH_EXPIRY,
  },
  {
    label: t('accessTokenSetting.expiryOptionNever'),
    value: AccessTokenExpiration.NEVER_EXPIRY,
  },
])

const accessTokenToAdd = ref<App.Api.Setting.AccessTokenDto>({
  name: '',
  expiry: AccessTokenExpiration.EIGHT_HOUR_EXPIRY,
  scopes: [],
  audience: 'public-client',
})
const nameError = ref('')
const scopesError = ref('')

const isSubmitting = ref<boolean>(false)
const scopeLabelMap: Record<string, string> = {
  'echo:read': 'accessTokenSetting.scopeEchoRead',
  'echo:write': 'accessTokenSetting.scopeEchoWrite',
  'comment:read': 'accessTokenSetting.scopeCommentRead',
  'comment:write': 'accessTokenSetting.scopeCommentWrite',
  'comment:moderate': 'accessTokenSetting.scopeCommentModerate',
  'file:read': 'accessTokenSetting.scopeFileRead',
  'file:write': 'accessTokenSetting.scopeFileWrite',
  'connect:read': 'accessTokenSetting.scopeConnectRead',
  'connect:write': 'accessTokenSetting.scopeConnectWrite',
  'profile:read': 'accessTokenSetting.scopeProfileRead',
  'profile:write': 'accessTokenSetting.scopeProfileWrite',
  'admin:settings': 'accessTokenSetting.scopeAdminSettings',
  'admin:user': 'accessTokenSetting.scopeAdminUser',
  'admin:token': 'accessTokenSetting.scopeAdminToken',
}

const selectedTokenScopes = computed(() => parseTokenScopes(selectedToken.value?.scopes))

const selectedTokenExpiryText = computed(() =>
  selectedToken.value?.expiry
    ? new Date(selectedToken.value.expiry * 1000).toLocaleString()
    : t('accessTokenSetting.neverExpire'),
)
const selectedTokenCreatedText = computed(() =>
  selectedToken.value ? new Date(selectedToken.value.created_at * 1000).toLocaleString() : '',
)

const selectedTokenScopeGroups = computed(() => {
  const owned = new Set(selectedTokenScopes.value)
  return scopeGroups
    .map((group) => ({
      labelKey: group.labelKey,
      items: group.items.filter((item) => owned.has(item.value)),
    }))
    .filter((group) => group.items.length > 0)
})

function parseTokenScopes(rawScopes: App.Api.Setting.AccessToken['scopes']) {
  if (Array.isArray(rawScopes)) {
    return rawScopes
  }
  if (typeof rawScopes !== 'string' || rawScopes.trim() === '') {
    return []
  }
  try {
    const parsed = JSON.parse(rawScopes)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

function getScopeLabel(scope: string) {
  const key = scopeLabelMap[scope]
  return key ? String(t(key)) : scope
}

function getAudienceLabel(audience?: App.Api.Setting.AccessToken['audience']) {
  if (audience === 'cli') {
    return String(t('accessTokenSetting.audienceCli'))
  }
  if (audience === 'integration') {
    return String(t('accessTokenSetting.audienceIntegration'))
  }
  if (audience === 'mcp-remote') {
    return String(t('accessTokenSetting.audienceMcpRemote'))
  }
  return String(t('accessTokenSetting.audiencePublicClient'))
}

function openTokenDetail(item: App.Api.Setting.AccessToken) {
  selectedToken.value = item
  detailModalOpen.value = true
}

function closeTokenDetail() {
  detailModalOpen.value = false
}

function hasScope(scope: string) {
  return accessTokenToAdd.value.scopes.includes(scope)
}

function toggleScope(scope: string) {
  const scopes = accessTokenToAdd.value.scopes
  if (scopes.includes(scope)) {
    accessTokenToAdd.value.scopes = scopes.filter((item) => item !== scope)
    return
  }
  accessTokenToAdd.value.scopes = [...scopes, scope]
}

function resetAccessTokenForm() {
  accessTokenToAdd.value = {
    name: '',
    expiry: AccessTokenExpiration.EIGHT_HOUR_EXPIRY,
    scopes: [],
    audience: 'public-client',
  }
  nameError.value = ''
  scopesError.value = ''
}

const handleAddAccessToken = async () => {
  nameError.value = ''
  scopesError.value = ''
  const normalizedName = accessTokenToAdd.value.name.trim()
  if (!normalizedName) {
    nameError.value = String(t('accessTokenSetting.fillName'))
    theToast.error(nameError.value)
    return
  }
  if (accessTokenToAdd.value.scopes.length === 0) {
    scopesError.value = String(t('accessTokenSetting.selectScopes'))
    theToast.error(scopesError.value)
    return
  }

  isSubmitting.value = true
  try {
    const res = await fetchCreateAccessToken({
      name: normalizedName,
      expiry: accessTokenToAdd.value.expiry || AccessTokenExpiration.NEVER_EXPIRY,
      scopes: accessTokenToAdd.value.scopes,
      audience: accessTokenToAdd.value.audience,
    })
    if (res.code === 1) {
      theToast.success(String(t('accessTokenSetting.createSuccess')))
      resetAccessTokenForm()
      await useSetting.getAllAccessTokens()
      accessTokenEdit.value = false
    } else {
      theToast.error(res.msg || String(t('accessTokenSetting.createFailed')))
    }
  } catch (error) {
    console.error(error)
    theToast.error(String(t('accessTokenSetting.createFailed')))
  } finally {
    isSubmitting.value = false
  }
}

const handleCancelAddAccessToken = () => {
  resetAccessTokenForm()
  accessTokenEdit.value = false
}

const maskToken = (token: string) => {
  if (!token) return ''
  if (token.length <= 10) {
    const left = Math.max(1, Math.floor(token.length / 3))
    const right = Math.max(1, Math.floor(token.length / 3))
    return `${token.slice(0, left)}***${token.slice(token.length - right)}`
  }
  return `${token.slice(0, 6)}...${token.slice(-4)}`
}

const copyAccessToken = async (token: string) => {
  if (!token) {
    theToast.error(String(t('accessTokenSetting.tokenEmpty')))
    return
  }

  try {
    await navigator.clipboard.writeText(token)
    theToast.success(String(t('accessTokenSetting.copySuccess')))
  } catch {
    theToast.error(String(t('accessTokenSetting.copyFailed')))
  }
}

const handleDeleteAccessToken = async (item: App.Api.Setting.AccessToken) => {
  if (!item) return

  openConfirm({
    title: String(t('accessTokenSetting.deleteConfirmTitle')),
    description: `${String(t('accessTokenSetting.name'))}：${item.name}`,
    onConfirm: async () => {
      const res = await fetchDeleteAccessToken(item.id)
      if (res.code === 1) {
        theToast.success(String(t('accessTokenSetting.deleteSuccess')))
        await useSetting.getAllAccessTokens()
      }
    },
  })
}

onMounted(async () => {
  await useSetting.getAllAccessTokens()
})
</script>

<style scoped></style>
