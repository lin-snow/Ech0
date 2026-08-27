<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full px-2">
    <BaseSegmented v-model="tab" :options="tabOptions" />

    <PanelCard>
      <TheMigrationSetting v-if="tab === 'import'" />
      <TheExportSetting v-else-if="tab === 'export'" />
      <TheSnapshotScheduleSetting v-else />
    </PanelCard>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import PanelCard from '@/layout/PanelCard.vue'
import BaseSegmented from '@/components/common/BaseSegmented.vue'
import TheMigrationSetting from './TheSetting/TheMigrationSetting.vue'
import TheExportSetting from './TheSetting/TheExportSetting.vue'
import TheSnapshotScheduleSetting from './TheSetting/TheSnapshotScheduleSetting.vue'

const { t } = useI18n()
const tab = ref('import')
const tabOptions = computed(() => [
  { label: String(t('dataManagement.tabImport')), value: 'import' },
  { label: String(t('dataManagement.tabExport')), value: 'export' },
  { label: String(t('dataManagement.tabSnapshot')), value: 'snapshot' },
])
</script>

<style scoped></style>
