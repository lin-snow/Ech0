<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <TheConnectSetting v-if="tab === 'connect'" />
    <TheCopilotSetting v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheConnectSetting from './TheSetting/TheConnectSetting.vue'
import TheCopilotSetting from './TheCopilot/TheCopilotSetting.vue'

const { t } = useI18n()
const tab = ref('connect')
const tabOptions = computed(() => [
  { label: String(t('extensionManagement.tabConnect')), value: 'connect' },
  { label: String(t('extensionManagement.tabCopilot')), value: 'copilot' },
])
</script>

<style scoped></style>
