<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <WidgetShell v-if="AgentSetting.enable" title="Recent AI" :fluid="fluid">
    <template #icon>
      <RecentIcon />
    </template>

    <div class="recent-body">
      <div class="recent-card">
        <div v-if="!loading" class="recent-content">
          <TheMdPreview :content="recent" />
        </div>
        <div v-else>
          <div class="recent-loading">{{ t('recentCard.generating') }}</div>
        </div>
      </div>
    </div>
  </WidgetShell>
</template>
<script setup lang="ts">
import { fetchGetRecent } from '@/service/api'
import { onMounted, ref } from 'vue'
import RecentIcon from '@/components/icons/recent.vue'
import WidgetShell from './WidgetShell.vue'
import { TheMdPreview } from '@/components/advanced/md'
import { useSettingStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'

defineProps<{ fluid?: boolean }>()

const settingStore = useSettingStore()
const { AgentSetting } = storeToRefs(settingStore)
const { t } = useI18n()

const recent = ref<string>(String(t('recentCard.mysteriousRecent')))
const loading = ref<boolean>(true)

onMounted(() => {
  if (AgentSetting.value.enable) {
    fetchGetRecent()
      .then((res) => {
        if (res.code === 1) {
          recent.value = res.data
        }
      })
      .catch((err) => {
        console.error(err)
      })
      .finally(() => {
        loading.value = false
      })
  }
})
</script>
<style scoped>
.recent-body {
  width: 100%;
}

.recent-card {
  position: relative;
  width: 100%;
  border: 1px solid var(--color-border-subtle);
  background-color: var(--recent-card-bg);
  box-shadow: 0 8px 18px rgb(20 20 20 / 4%);
  padding: 14px 12px 12px;
  transform: rotate(-1.1deg);
  transform-origin: top center;
  transition: transform 220ms ease;
}

.recent-card:hover {
  transform: rotate(-0.4deg);
}

.recent-card::before {
  content: '';
  position: absolute;
  left: 50%;
  top: -7px;
  transform: translateX(-50%) rotate(-1.5deg);
  width: 42px;
  height: 12px;
  border-radius: 2px;
  background: var(--recent-card-before-bg);
  box-shadow:
    0 1px 0 rgb(255 255 255 / 30%) inset,
    0 1px 2px rgb(0 0 0 / 8%);
  opacity: 0.95;
}

.recent-content {
  color: var(--color-text-secondary);
  font-size: 13px;
  line-height: 1.65;
  white-space: normal;
  overflow-wrap: anywhere;
}

.recent-loading {
  color: var(--color-text-secondary);
  font-size: 13px;
}

:deep(.echo-markdown p) {
  color: var(--color-text-secondary) !important;
  font-size: 13px !important;
  line-height: 1.65 !important;
}
</style>
