<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <TheUserSetting v-if="tab === 'account'" />
    <TheUserManager v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheUserSetting from './TheSetting/TheUserSetting.vue'
import TheUserManager from './TheSetting/TheUserManager.vue'
import { useUserStore } from '@/stores'

const { t } = useI18n()
const userStore = useUserStore()
const tab = ref('account')
const tabOptions = computed(() => {
  const options = [{ label: String(t('userManagement.tabAccount')), value: 'account' }]
  if (userStore.user?.is_admin) {
    options.push({ label: String(t('userManagement.tabManage')), value: 'manage' })
  }
  return options
})
</script>
