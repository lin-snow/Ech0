<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <div class="flex flex-row items-center justify-between mb-3">
        <h1 class="text-[var(--color-text-primary)] font-bold text-lg">
          {{ t('storageSetting.title') }}
        </h1>
        <div class="flex flex-row items-center justify-end">
          <BaseEditCapsule
            :editing="storageEditMode"
            :apply-title="t('commonUi.apply')"
            :cancel-title="t('commonUi.cancel')"
            :edit-title="t('commonUi.edit')"
            @apply="handleUpdateS3Setting"
            @toggle="storageEditMode = !storageEditMode"
          />
        </div>
      </div>

      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.enableS3') }}:
        </h2>
        <BaseSwitch v-model="S3Setting.enable" :disabled="!storageEditMode" />
      </div>

      <div class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10">
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.enableSsl') }}:
        </h2>
        <BaseSwitch v-model="S3Setting.use_ssl" :disabled="!storageEditMode" />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.s3Provider') }}:
        </h2>
        <BaseSelect
          v-model="S3Setting.provider"
          :options="S3ServiceOptions"
          :disabled="!storageEditMode"
          class="w-fit h-8"
        />
      </div>

      <div
        v-if="S3Setting.provider === S3Provider.OTHER"
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] h-10"
      >
        <h2
          v-tooltip="t('storageSetting.usePathStyleHint')"
          class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap"
        >
          {{ t('storageSetting.usePathStyle') }}:
        </h2>
        <BaseSwitch v-model="S3Setting.use_path_style" :disabled="!storageEditMode" />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.endpoint') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.endpoint"
          style="vertical-align: middle"
        >
          {{ S3Setting.endpoint.length === 0 ? t('commonUi.none') : S3Setting.endpoint }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.endpoint"
          type="text"
          :placeholder="t('storageSetting.endpointPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.accessKey') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.access_key"
          style="vertical-align: middle"
        >
          {{ S3Setting.access_key.length === 0 ? t('commonUi.none') : S3Setting.access_key }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.access_key"
          type="text"
          :placeholder="t('storageSetting.accessKeyPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.secretKey') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.secret_key"
          style="vertical-align: middle"
        >
          {{ S3Setting.secret_key.length === 0 ? t('commonUi.none') : S3Setting.secret_key }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.secret_key"
          type="text"
          :placeholder="t('storageSetting.secretKeyPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.bucket') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.bucket_name"
          style="vertical-align: middle"
        >
          {{ S3Setting.bucket_name.length === 0 ? t('commonUi.none') : S3Setting.bucket_name }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.bucket_name"
          type="text"
          :placeholder="t('storageSetting.bucketPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.pathPrefix') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.path_prefix"
          style="vertical-align: middle"
        >
          {{ S3Setting.path_prefix.length === 0 ? t('commonUi.none') : S3Setting.path_prefix }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.path_prefix"
          type="text"
          :placeholder="t('storageSetting.pathPrefixPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        v-if="S3Setting.provider !== S3Provider.MINIO"
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.region') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.region"
          style="vertical-align: middle"
        >
          {{ S3Setting.region.length === 0 ? t('commonUi.none') : S3Setting.region }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.region"
          type="text"
          :placeholder="t('storageSetting.regionPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div
        class="flex flex-row items-center justify-start text-[var(--color-text-secondary)] gap-2 h-10"
      >
        <h2 class="font-semibold min-w-30 w-max shrink-0 whitespace-nowrap">
          {{ t('storageSetting.cdnDomain') }}:
        </h2>
        <span
          v-if="!storageEditMode"
          class="flex-1 min-w-0 truncate inline-block align-middle"
          v-tooltip="S3Setting.cdn_url"
          style="vertical-align: middle"
        >
          {{ S3Setting.cdn_url.length === 0 ? t('commonUi.none') : S3Setting.cdn_url }}
        </span>
        <BaseInput
          v-else
          v-model="S3Setting.cdn_url"
          type="text"
          :placeholder="t('storageSetting.cdnPlaceholder')"
          class="w-full py-1!"
        />
      </div>

      <div class="flex justify-end mt-6">
        <BaseButton
          class="px-3 text-sm bg-transparent"
          :loading="s3Testing"
          :title="t('storageSetting.testConnection')"
          @click="handleTestS3Connection"
        >
          {{ t('storageSetting.testConnection') }}
        </BaseButton>
      </div>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import PanelCard from '@/layout/PanelCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { S3Provider } from '@/enums/enums'
import { fetchUpdateS3Settings, fetchTestS3Connection } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'

const settingStore = useSettingStore()
const { t } = useI18n()
const { getS3Setting } = settingStore
const { S3Setting } = storeToRefs(settingStore)

const storageEditMode = ref<boolean>(false)

const S3ServiceOptions = ref<{ label: string; value: S3Provider }[]>([
  { label: 'AWS', value: S3Provider.AWS },
  { label: 'MinIO', value: S3Provider.MINIO },
  { label: 'Cloudflare R2', value: S3Provider.R2 },
  { label: 'Other', value: S3Provider.OTHER },
])

const handleUpdateS3Setting = async () => {
  await fetchUpdateS3Settings(settingStore.S3Setting)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
      }
    })
    .finally(() => {
      storageEditMode.value = false
      getS3Setting()
    })
}

const s3Testing = ref<boolean>(false)
const handleTestS3Connection = async () => {
  s3Testing.value = true
  try {
    const res = await fetchTestS3Connection(settingStore.S3Setting)
    if (res.code === 1) {
      theToast.success(t('storageSetting.testSuccess'))
    } else {
      theToast.error(t('storageSetting.testFailed', { detail: res.msg }))
    }
  } finally {
    s3Testing.value = false
  }
}

onMounted(() => {
  getS3Setting()
})
</script>
