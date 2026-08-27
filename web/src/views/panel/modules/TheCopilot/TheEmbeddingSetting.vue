<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full text-[var(--color-text-secondary)]">
    <p class="text-xs opacity-70 mb-4">{{ t('embeddingSetting.optionalHint') }}</p>

    <div class="flex items-center justify-between mb-4">
      <h2 class="font-semibold">{{ t('embeddingSetting.enable') }}</h2>
      <BaseSwitch v-model="setting.enable" :disabled="!editMode" />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.modelName') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80" v-tooltip="setting.model">
        {{ setting.model || t('commonUi.none') }}
      </span>
      <BaseCombobox
        v-else
        v-model="setting.model"
        :options="modelOptions"
        :allow-create="true"
        :placeholder="t('embeddingSetting.modelPlaceholder')"
        wrapper-class="w-full"
        class="w-full"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.dim') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.dim || t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model.number="setting.dim"
        type="number"
        :placeholder="t('embeddingSetting.dimPlaceholder')"
        class="w-full"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.apiKey') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.api_key ? '********' : t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="setting.api_key"
        type="password"
        :placeholder="t('embeddingSetting.apiKeyPlaceholder')"
        class="w-full"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.baseUrl') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.base_url.length === 0 ? t('commonUi.none') : setting.base_url }}
      </span>
      <BaseInput
        v-else
        v-model="setting.base_url"
        :placeholder="t('embeddingSetting.baseUrlPlaceholder')"
        class="w-full"
      />
      <p v-if="editMode" class="text-xs opacity-70 mt-1">{{ t('embeddingSetting.baseUrlHint') }}</p>
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('embeddingSetting.batchSize') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ setting.batch_size || t('embeddingSetting.batchSizeDefault') }}
      </span>
      <BaseInput
        v-else
        v-model.number="setting.batch_size"
        type="number"
        :placeholder="t('embeddingSetting.batchSizePlaceholder')"
        class="w-full"
      />
      <p v-if="editMode" class="text-xs opacity-70 mt-1">
        {{ t('embeddingSetting.batchSizeHint') }}
      </p>
    </div>

    <div
      class="flex flex-row items-center justify-between gap-2 mt-5 pt-4 border-t border-[var(--color-border-subtle)]"
    >
      <div class="min-w-0">
        <h2 class="font-semibold">{{ t('embeddingSetting.reindex') }}</h2>
        <p class="text-xs opacity-70 mt-1">{{ t('embeddingSetting.reindexHint') }}</p>
      </div>
      <div class="flex items-center gap-2 shrink-0">
        <BaseButton v-if="reindex.isRunning" @click="handleCancelReindex">
          {{ t('embeddingSetting.reindexCancel') }}
        </BaseButton>
        <BaseButton
          :loading="reindex.isRunning"
          :disabled="reindex.isRunning"
          @click="handleReindex"
        >
          {{ t('embeddingSetting.reindexAction') }}
        </BaseButton>
      </div>
    </div>
    <p v-if="reindex.isRunning" class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2">
      {{
        reindex.result
          ? t('embeddingSetting.reindexProgress', {
              indexed: reindex.result.indexed,
              total: reindex.result.total,
            })
          : t('embeddingSetting.reindexRunning')
      }}
    </p>
    <p
      v-else-if="reindex.status === 'success' && reindex.result"
      class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2"
    >
      {{
        t('embeddingSetting.reindexResult', {
          indexed: reindex.result.indexed,
          total: reindex.result.total,
          failed: reindex.result.failed,
        })
      }}
    </p>
    <p
      v-else-if="reindex.status === 'failed'"
      class="text-xs text-[var(--color-danger,#dc2626)] mt-2"
    >
      {{ t('embeddingSetting.reindexFailed', { error: reindex.error }) }}
    </p>
    <p
      v-else-if="reindex.status === 'cancelled'"
      class="text-xs text-[var(--color-text-secondary)] opacity-80 mt-2"
    >
      {{ t('embeddingSetting.reindexCancelled') }}
    </p>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseCombobox from '@/components/common/BaseCombobox.vue'
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchGetEmbeddingSettings, fetchUpdateEmbeddingSettings } from '@/service/api'
import { theToast } from '@/utils/toast'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { useReindexStore } from '@/stores/reindex'

const props = defineProps<{ editMode: boolean }>()

const { t } = useI18n()
const { openConfirm } = useBaseDialog()

const MODEL_DIM_PRESETS: Record<string, number> = {
  'text-embedding-3-small': 1536,
  'text-embedding-3-large': 3072,
  'text-embedding-ada-002': 1536,
  'text-embedding-v3': 1024,
  'bge-m3': 1024,
  'bge-large-zh-v1.5': 1024,
  'jina-embeddings-v3': 1024,
  'nomic-embed-text': 768,
  'mxbai-embed-large': 1024,
}
const modelOptions = Object.keys(MODEL_DIM_PRESETS)

const reindex = useReindexStore()

const setting = ref<App.Api.Embedding.EmbeddingSetting>({
  enable: false,
  model: '',
  api_key: '',
  base_url: '',
  dim: 0,
  batch_size: 0,
})

const originalModel = ref<string>('')
const originalDim = ref<number>(0)

watch(
  () => setting.value.model,
  (next) => {
    if (!props.editMode) return
    const preset = MODEL_DIM_PRESETS[next]
    if (preset) {
      setting.value.dim = preset
    }
  },
)

const getSetting = async () => {
  const res = await fetchGetEmbeddingSettings()
  if (res.code === 1 && res.data) {
    setting.value = res.data
    originalModel.value = res.data.model
    originalDim.value = res.data.dim
  }
}

const save = async () => {
  const changed =
    setting.value.model !== originalModel.value || setting.value.dim !== originalDim.value
  await fetchUpdateEmbeddingSettings(setting.value)
    .then((res) => {
      if (res.code === 1) {
        theToast.success(res.msg)
        if (changed && setting.value.enable) {
          openConfirm({
            title: t('embeddingSetting.reindexConfirmTitle'),
            description: t('embeddingSetting.reindexConfirmDesc'),
            onConfirm: () => handleReindex(),
          })
        }
      }
    })
    .finally(() => {
      getSetting()
    })
}

defineExpose({ save })

const handleReindex = async () => {
  const res = await reindex.start()
  if (res.code === 1) {
    theToast.success(res.msg)
  }
}

const handleCancelReindex = async () => {
  const res = await reindex.cancel()
  if (res.code === 1) {
    theToast.success(res.msg)
  }
}

onMounted(() => {
  getSetting()
  reindex.init()
})

onUnmounted(() => {
  reindex.stopPolling()
})
</script>

<style scoped></style>
