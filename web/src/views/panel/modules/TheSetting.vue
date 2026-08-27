<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <TheSystemSetting v-if="tab === 'system'" />
    <TheAccessTokenSetting v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheSystemSetting from './TheSetting/TheSystemSetting.vue'
import TheAccessTokenSetting from './TheSetting/TheAccessTokenSetting.vue'

const { t } = useI18n()
const tab = ref('system')
const tabOptions = computed(() => [
  { label: String(t('settingManagement.tabSystem')), value: 'system' },
  { label: String(t('settingManagement.tabAccessToken')), value: 'accessToken' },
])
</script>
