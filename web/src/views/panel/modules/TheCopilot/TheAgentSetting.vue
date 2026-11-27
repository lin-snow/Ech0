<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full text-[var(--color-text-secondary)]">
    <div class="flex items-center justify-between mb-4">
      <h2 class="font-semibold">{{ t('agentSetting.enableAgent') }}</h2>
      <BaseSwitch v-model="AgentSetting.enable" :disabled="!editMode" />
    </div>

    <div class="flex items-center justify-between mb-4">
      <h2 class="font-semibold">{{ t('agentSetting.protocol') }}</h2>
      <BaseSelect
        v-model="AgentSetting.protocol"
        :options="agentProtocolOptions"
        :disabled="!editMode"
        class="w-40 h-8"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.modelName') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80" v-tooltip="AgentSetting.model">
        {{ AgentSetting.model || t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="AgentSetting.model"
        type="text"
        :placeholder="t('agentSetting.modelPlaceholder')"
        class="w-full"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.apiKey') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ AgentSetting.api_key ? '********' : t('commonUi.none') }}
      </span>
      <BaseInput
        v-else
        v-model="AgentSetting.api_key"
        type="password"
        :placeholder="t('agentSetting.apiKeyPlaceholder')"
        class="w-full"
      />
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.baseUrl') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ AgentSetting.base_url.length === 0 ? t('commonUi.none') : AgentSetting.base_url }}
      </span>
      <template v-else>
        <BaseInput
          v-model="AgentSetting.base_url"
          :placeholder="t('agentSetting.baseUrlPlaceholder')"
          class="w-full"
        />
      </template>
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.contextWindow') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ formatTokenSize(AgentSetting.context_window) || t('commonUi.none') }}
      </span>
      <template v-else>
        <BaseInput
          v-model="contextWindowRaw"
          type="text"
          :placeholder="t('agentSetting.contextWindowPlaceholder')"
          class="w-full"
          @input="onContextWindowInput"
          @blur="onContextWindowBlur"
        />
        <p class="text-xs opacity-70 mt-1">{{ t('agentSetting.contextWindowHint') }}</p>
      </template>
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.temperature') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ AgentSetting.temperature ?? t('agentSetting.temperatureDefault') }}
      </span>
      <template v-else>
        <BaseInput
          v-model="temperatureRaw"
          type="text"
          inputmode="decimal"
          :placeholder="t('agentSetting.temperaturePlaceholder')"
          class="w-full"
          @input="onTemperatureInput"
          @blur="onTemperatureBlur"
        />
        <p class="text-xs opacity-70 mt-1">{{ t('agentSetting.temperatureHint') }}</p>
      </template>
    </div>

    <div class="mb-4">
      <h2 class="font-semibold mb-1.5">{{ t('agentSetting.prompt') }}</h2>
      <span v-if="!editMode" class="block truncate opacity-80">
        {{ AgentSetting.prompt.length === 0 ? t('commonUi.none') : AgentSetting.prompt }}
      </span>
      <template v-else>
        <BaseTextArea
          v-model="AgentSetting.prompt"
          :placeholder="t('agentSetting.promptPlaceholder')"
          class="w-full"
          :rows="4"
        />
        <p class="text-xs opacity-70 mt-1">{{ t('agentSetting.promptHint') }}</p>
      </template>
    </div>

    <div class="mb-1">
      <div class="flex items-center justify-between">
        <h2 class="font-semibold">{{ t('agentSetting.multimodal') }}</h2>
        <BaseSwitch v-model="AgentSetting.multimodal" :disabled="!editMode" />
      </div>
      <p class="text-xs opacity-70 mt-1">{{ t('agentSetting.multimodalHint') }}</p>
    </div>

    <div class="flex justify-end mt-6">
      <BaseButton
        class="px-3 text-sm bg-transparent"
        :loading="agentTesting"
        :title="t('agentSetting.testConnection')"
        @click="handleTestAgentConnection"
      >
        {{ t('agentSetting.testConnection') }}
      </BaseButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import BaseSwitch from '@/components/common/BaseSwitch.vue'
import BaseSelect from '@/components/common/BaseSelect.vue'
import BaseTextArea from '@/components/common/BaseTextArea.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchUpdateAgentSettings, fetchTestAgentConnection } from '@/service/api'
import { theToast } from '@/utils/toast'
import { i18n } from '@/locales'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { AgentProtocol } from '@/enums/enums'
import { parseTokenSize, formatTokenSize } from '@/utils/tokenSize'

defineProps<{ editMode: boolean }>()

const settingStore = useSettingStore()
const { t } = useI18n()
const { getAgentSetting } = settingStore
const { AgentSetting } = storeToRefs(settingStore)

const contextWindowRaw = ref('')
watch(
  () => AgentSetting.value.context_window,
  (tokens) => {
    contextWindowRaw.value = formatTokenSize(tokens)
  },
  { immediate: true },
)
const onContextWindowInput = () => {
  AgentSetting.value.context_window = parseTokenSize(contextWindowRaw.value)
}
const onContextWindowBlur = () => {
  contextWindowRaw.value = formatTokenSize(AgentSetting.value.context_window)
}

// Empty means "not sent": the model runs at its own default, which is the only
// value every model accepts. Anything that is not a number in 0–2 is treated
// as empty rather than sent to fail at the provider.
const temperatureRaw = ref('')
watch(
  () => AgentSetting.value.temperature,
  (value) => {
    temperatureRaw.value = value === undefined ? '' : String(value)
  },
  { immediate: true },
)
const parseTemperature = (raw: string): number | undefined => {
  const text = raw.trim()
  if (text === '') return undefined
  const value = Number(text)
  return Number.isFinite(value) && value >= 0 && value <= 2 ? value : undefined
}
const onTemperatureInput = () => {
  AgentSetting.value.temperature = parseTemperature(temperatureRaw.value)
}
const onTemperatureBlur = () => {
  const value = AgentSetting.value.temperature
  temperatureRaw.value = value === undefined ? '' : String(value)
}

const agentProtocolOptions = computed<{ label: string; value: AgentProtocol }[]>(() => [
  { label: t('agentSetting.protocolOpenAI'), value: AgentProtocol.OPENAI },
  { label: t('agentSetting.protocolOpenAIResponses'), value: AgentProtocol.OPENAI_RESPONSES },
  { label: t('agentSetting.protocolAnthropic'), value: AgentProtocol.ANTHROPIC },
])

const save = async () => {
  await fetchUpdateAgentSettings(settingStore.AgentSetting)
    .then((res) => {
      if (res) {
        theToast.success(String(i18n.global.t('common.success')))
      }
    })
    .finally(() => {
      getAgentSetting()
    })
}

defineExpose({ save })

const agentTesting = ref<boolean>(false)
const handleTestAgentConnection = async () => {
  agentTesting.value = true
  try {
    const res = await fetchTestAgentConnection(settingStore.AgentSetting)
    if (res) {
      theToast.success(t('agentSetting.testSuccess'))
    }
  } finally {
    agentTesting.value = false
  }
}

onMounted(() => {
  getAgentSetting()
})
</script>

<style scoped></style>
