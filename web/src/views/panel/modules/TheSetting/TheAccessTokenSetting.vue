<template>
  <PanelCard>
    <!-- Webhook 设置 -->
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
        <table class="w-full min-w-[820px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
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
            <tr v-for="tokenItem in AccessTokens" :key="tokenItem.id">
              <td class="px-2 py-2 font-mono text-[var(--color-text-primary)]">
                <div class="flex items-center gap-1">
                  <span class="truncate" :title="tokenItem.token">{{
                    maskToken(tokenItem.token)
                  }}</span>
                  <button
                    class="p-1 hover:bg-[var(--color-bg-surface)] rounded"
                    @click="copyAccessToken(tokenItem.token)"
                    :title="t('accessTokenSetting.copyToken')"
                  >
                    <Clipboard class="w-4 h-4" />
                  </button>
                </div>
              </td>
              <td class="px-2 py-2 text-[var(--color-text-primary)]">
                <span :title="tokenItem.name" class="truncate block max-w-xs">{{
                  tokenItem.name
                }}</span>
              </td>
              <td class="px-1 py-2 text-[var(--color-text-secondary)] whitespace-nowrap">
                {{ new Date(tokenItem.created_at).toLocaleString() }}
              </td>
              <td class="px-1 py-2 text-[var(--color-text-secondary)] whitespace-nowrap">
                {{
                  tokenItem.expiry
                    ? new Date(tokenItem.expiry).toLocaleString()
                    : t('accessTokenSetting.neverExpire')
                }}
              </td>
              <td class="px-1 py-2">
                <div class="flex items-center justify-end gap-1">
                  <BaseButton
                    class="h-8 rounded-md px-2 text-xs whitespace-nowrap"
                    @click="openTokenDetail(tokenItem)"
                    :title="t('accessTokenSetting.viewDetail')"
                  >
                    <span>{{ t('accessTokenSetting.detail') }}</span>
                  </BaseButton>
                  <BaseButton
                    class="h-8 w-8 !p-1.5"
                    :icon="Trashbin"
                    @click="handleDeleteAccessToken(tokenItem)"
                    :title="t('accessTokenSetting.deleteToken')"
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
            <span class="text-[var(--color-text-primary)]">{{ t('accessTokenSetting.expiry') }}：</span>
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
          <span class="text-[var(--color-text-primary)]">{{ t('accessTokenSetting.scopes') }}：</span>
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
          <span v-if="scopesError" class="text-xs text-[var(--color-danger)]">{{ scopesError }}</span>
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
          :title="t('accessTokenSetting.cancelAdd')"
        >
          <span>{{ t('commonUi.cancel') }}</span>
        </BaseButton>

        <BaseButton
          :loading="isSubmitting"
          @click="handleAddAccessToken"
          class="h-9 rounded-md px-4 bg-[var(--color-bg-surface)]! bg-op-80"
          :title="t('accessTokenSetting.addToken')"
        >
          <span class="text-[var(--color-text-primary)]">{{ t('commonUi.add') }}</span>
        </BaseButton>
      </div>
    </div>

    <div
      v-if="detailModalOpen && selectedToken"
      class="fixed inset-0 z-5000 flex items-center justify-center bg-black/30 p-4"
      @click.self="closeTokenDetail"
    >
      <div
        class="w-full max-w-lg rounded-[var(--radius-lg)] bg-[var(--dialog-bg-color)] p-5 shadow-[var(--shadow-md)] ring-1 ring-inset ring-[var(--color-border-subtle)]"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <h3 class="text-base font-semibold text-[var(--color-text-primary)]">
              {{ t('accessTokenSetting.detailTitle') }}
            </h3>
            <p class="mt-1 text-sm text-[var(--color-text-muted)]">{{ selectedToken.name }}</p>
          </div>
        </div>

        <div class="mt-4 space-y-3 text-sm">
          <div class="rounded-md border border-[var(--color-border-subtle)] p-3">
            <div class="text-xs text-[var(--color-text-muted)]">
              {{ t('accessTokenSetting.audience') }}
            </div>
            <div class="mt-1 text-[var(--color-text-primary)]">
              {{ getAudienceLabel(selectedToken.audience) }}
            </div>
          </div>
          <div class="rounded-md border border-[var(--color-border-subtle)] p-3">
            <div class="text-xs text-[var(--color-text-muted)]">
              {{ t('accessTokenSetting.expiry') }}
            </div>
            <div class="mt-1 text-[var(--color-text-primary)]">
              {{
                selectedToken.expiry
                  ? new Date(selectedToken.expiry).toLocaleString()
                  : t('accessTokenSetting.neverExpire')
              }}
            </div>
          </div>
          <div class="rounded-md border border-[var(--color-border-subtle)] p-3">
            <div class="text-xs text-[var(--color-text-muted)]">
              {{ t('accessTokenSetting.createdAt') }}
            </div>
            <div class="mt-1 text-[var(--color-text-primary)]">
              {{ new Date(selectedToken.created_at).toLocaleString() }}
            </div>
          </div>
          <div class="rounded-md border border-[var(--color-border-subtle)] p-3">
            <div class="text-xs text-[var(--color-text-muted)]">
              {{ t('accessTokenSetting.scopes') }}
            </div>
            <div v-if="selectedTokenScopes.length > 0" class="mt-2 flex flex-wrap gap-2">
              <span
                v-for="scope in selectedTokenScopes"
                :key="scope"
                class="rounded-md border border-[var(--color-border-subtle)] px-2 py-1 text-xs text-[var(--color-text-secondary)]"
              >
                {{ getScopeLabel(scope) }}
              </span>
            </div>
            <div v-else class="mt-1 text-[var(--color-text-muted)]">
              {{ t('accessTokenSetting.scopeEmpty') }}
            </div>
          </div>
        </div>

        <div class="mt-4 flex justify-end">
          <BaseButton class="h-9 rounded-md px-4" @click="closeTokenDetail">
            {{ t('accessTokenSetting.closeDetail') }}
          </BaseButton>
        </div>
      </div>
    </div>
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
    labelKey: 'accessTokenSetting.scopeGroupProfile',
    items: [{ value: 'profile:read', labelKey: 'accessTokenSetting.scopeProfileRead' }],
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
  'profile:read': 'accessTokenSetting.scopeProfileRead',
  'admin:settings': 'accessTokenSetting.scopeAdminSettings',
  'admin:user': 'accessTokenSetting.scopeAdminUser',
  'admin:token': 'accessTokenSetting.scopeAdminToken',
}

const selectedTokenScopes = computed(() => parseTokenScopes(selectedToken.value?.scopes))

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
  return String(t('accessTokenSetting.audiencePublicClient'))
}

function openTokenDetail(item: App.Api.Setting.AccessToken) {
  selectedToken.value = item
  detailModalOpen.value = true
}

function closeTokenDetail() {
  detailModalOpen.value = false
  selectedToken.value = null
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

// 删除 Access Token
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
