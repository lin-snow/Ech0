<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <TheStorageSetting v-if="tab === 'object'" />
    <TheStorageFileList v-else />
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheStorageSetting from './TheSetting/TheStorageSetting.vue'
import TheStorageFileList from './TheSetting/TheStorageFileList.vue'

const { t } = useI18n()
const tab = ref('object')
const tabOptions = computed(() => [
  { label: String(t('storageManagement.tabObjectStorage')), value: 'object' },
  { label: String(t('storageManagement.tabFiles')), value: 'files' },
])
</script>

<style scoped></style>
