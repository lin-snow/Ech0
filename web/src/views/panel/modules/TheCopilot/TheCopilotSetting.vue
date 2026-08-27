<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <PanelCard>
    <div class="w-full">
      <h1 class="text-[var(--color-text-primary)] font-bold text-lg mb-3">
        {{ t('copilotSetting.title') }}
      </h1>

      <div
        class="flex flex-row items-center justify-between gap-2 mb-4 border-b"
        :style="{ borderColor: 'var(--color-border-subtle)' }"
      >
        <div class="flex flex-row gap-1">
          <button
            type="button"
            class="px-3 py-2 text-sm font-semibold border-b-2 -mb-px transition-colors"
            :class="
              tab === 'model'
                ? 'text-[var(--color-text-primary)] border-[var(--color-text-primary)]'
                : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text-primary)]'
            "
            @click="tab = 'model'"
          >
            {{ t('agentSetting.title') }}
          </button>
          <button
            type="button"
            class="px-3 py-2 text-sm font-semibold border-b-2 -mb-px transition-colors"
            :class="
              tab === 'embedding'
                ? 'text-[var(--color-text-primary)] border-[var(--color-text-primary)]'
                : 'text-[var(--color-text-secondary)] border-transparent hover:text-[var(--color-text-primary)]'
            "
            @click="tab = 'embedding'"
          >
            <span class="inline-flex items-center gap-1.5">
              {{ t('embeddingSetting.title') }}
              <span
                class="px-1.5 py-0.5 text-[10px] font-medium leading-none rounded-full bg-[var(--color-bg-muted)] text-[var(--color-text-muted)]"
              >
                {{ t('commonUi.optional') }}
              </span>
            </span>
          </button>
        </div>

        <BaseEditCapsule
          class="shrink-0 self-center"
          :editing="activeEdit"
          :apply-title="t('commonUi.apply')"
          :cancel-title="t('commonUi.cancel')"
          :edit-title="t('commonUi.edit')"
          @apply="handleApply"
          @toggle="activeEdit = !activeEdit"
        />
      </div>

      <KeepAlive>
        <TheAgentSetting v-if="tab === 'model'" ref="agentRef" :edit-mode="modelEdit" />
        <TheEmbeddingSetting v-else ref="embeddingRef" :edit-mode="embeddingEdit" />
      </KeepAlive>
    </div>
  </PanelCard>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import PanelCard from '@/layout/PanelCard.vue'
import BaseEditCapsule from '@/components/common/BaseEditCapsule.vue'
import TheAgentSetting from './TheAgentSetting.vue'
import TheEmbeddingSetting from './TheEmbeddingSetting.vue'

const { t } = useI18n()

const tab = ref<'model' | 'embedding'>('model')

const modelEdit = ref<boolean>(false)
const embeddingEdit = ref<boolean>(false)

const activeEdit = computed<boolean>({
  get: () => (tab.value === 'model' ? modelEdit.value : embeddingEdit.value),
  set: (v) => {
    if (tab.value === 'model') modelEdit.value = v
    else embeddingEdit.value = v
  },
})

const agentRef = ref<InstanceType<typeof TheAgentSetting> | null>(null)
const embeddingRef = ref<InstanceType<typeof TheEmbeddingSetting> | null>(null)

const handleApply = async () => {
  const target = tab.value === 'model' ? agentRef.value : embeddingRef.value
  await target?.save()
  activeEdit.value = false
}
</script>

<style scoped></style>
