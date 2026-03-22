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
        <table class="w-full min-w-[860px] table-fixed text-sm">
          <thead>
            <tr class="bg-[var(--color-bg-muted)]/70 text-left text-[var(--color-text-muted)]">
              <th class="w-[170px] px-2 py-2 whitespace-nowrap">Token</th>
              <th class="w-[100px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.name') }}
              </th>
              <th class="w-[120px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.scope') }}
              </th>
              <th class="w-[156px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.createdAt') }}
              </th>
              <th class="w-[156px] px-2 py-2 whitespace-nowrap">
                {{ t('accessTokenSetting.expiry') }}
              </th>
              <th class="w-[56px] px-2 py-2 text-right whitespace-nowrap">
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
              <td class="px-2 py-2 text-[var(--color-text-secondary)] whitespace-nowrap">
                {{ formatTokenScope(tokenItem.scope) }}
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
              <td class="px-1 py-2 text-right">
                <BaseButton
                  class="h-8 w-8 !p-1.5"
                  :icon="Trashbin"
                  @click="handleDeleteAccessToken(tokenItem)"
                  :title="t('accessTokenSetting.deleteToken')"
                />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <div v-else class="text-[var(--color-text-secondary)]">
      <!-- 添加 AccessToken -->

      <div class="flex flex-col gap-2 mb-2">
        <span>{{ t('accessTokenSetting.name') }}：</span>
        <BaseInput
          class="w-full"
          v-model="accessTokenToAdd.name"
          :placeholder="t('accessTokenSetting.namePlaceholder')"
        />
      </div>

      <div class="flex flex-col gap-2">
        <span>{{ t('accessTokenSetting.expiry') }}：</span>
        <BaseSelect
          v-model="accessTokenToAdd.expiry"
          :options="ExpirationOptions"
          class="w-34 h-8 bg-[var(--color-bg-surface)]! bg-op-80 mt-2 mb-4"
        />
      </div>

      <div class="flex flex-col gap-2">
        <span>{{ t('accessTokenSetting.scope') }}：</span>
        <BaseSelect
          v-model="accessTokenToAdd.scope"
          :options="ScopeOptions"
          class="w-48 h-8 bg-[var(--color-bg-surface)]! bg-op-80 mt-2 mb-4"
        />
      </div>

      <div class="flex items-center justify-center my-2">
        <BaseButton
          :disabled="isSubmitting"
          @click="handleCancelAddAccessToken"
          class="w-1/4 h-8 rounded-md flex justify-center mr-2 bg-[var(--color-bg-surface)]! bg-op-80"
          :title="t('accessTokenSetting.cancelAdd')"
        >
          <span>{{ t('commonUi.cancel') }}</span>
        </BaseButton>

        <BaseButton
          :loading="isSubmitting"
          @click="handleAddAccessToken"
          class="w-1/4 h-8 rounded-md flex justify-center bg-[var(--color-bg-surface)]! bg-op-80"
          :title="t('accessTokenSetting.addToken')"
        >
          <span class="text-[var(--color-text-primary)]">{{ t('commonUi.add') }}</span>
        </BaseButton>
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
import { ref, onMounted } from 'vue'
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

const accessTokenToAdd = ref<App.Api.Setting.AccessTokenDto>({
  name: '',
  expiry: AccessTokenExpiration.EIGHT_HOUR_EXPIRY,
  scope: '',
})
const ExpirationOptions = [
  { label: '8 Hours', value: AccessTokenExpiration.EIGHT_HOUR_EXPIRY },
  { label: '1 Month', value: AccessTokenExpiration.ONE_MONTH_EXPIRY },
  { label: 'Never', value: AccessTokenExpiration.NEVER_EXPIRY },
]
const ScopeOptions = [
  { label: String(t('accessTokenSetting.scopeFull')), value: '' },
  { label: String(t('accessTokenSetting.scopeIntegration')), value: 'integration' },
]

const isSubmitting = ref<boolean>(false)
const handleAddAccessToken = async () => {
  if (!accessTokenToAdd.value?.name) {
    theToast.error(String(t('accessTokenSetting.fillName')))
    return
  }

  isSubmitting.value = true

  const res = await fetchCreateAccessToken({
    name: accessTokenToAdd.value.name,
    expiry: accessTokenToAdd.value.expiry || AccessTokenExpiration.NEVER_EXPIRY,
    scope: accessTokenToAdd.value.scope || '',
  })
  if (res.code === 1) {
    theToast.success(String(t('accessTokenSetting.createSuccess')))
    accessTokenToAdd.value = {
      name: '',
      expiry: AccessTokenExpiration.EIGHT_HOUR_EXPIRY,
      scope: '',
    }
    await useSetting.getAllAccessTokens()
    accessTokenEdit.value = false
  }
  isSubmitting.value = false
}

const handleCancelAddAccessToken = () => {
  accessTokenToAdd.value = { name: '', expiry: AccessTokenExpiration.EIGHT_HOUR_EXPIRY, scope: '' }
  accessTokenEdit.value = false
}

const formatTokenScope = (scope: string) => {
  return scope === 'integration'
    ? String(t('accessTokenSetting.scopeIntegration'))
    : String(t('accessTokenSetting.scopeFull'))
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
