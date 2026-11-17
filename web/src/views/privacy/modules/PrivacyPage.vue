<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <NotFoundPage v-if="!policy" />
  <div v-else class="privacy-page">
    <div class="privacy-frame">
      <button
        type="button"
        class="privacy-back"
        :aria-label="t('commonNav.backHome')"
        @click="$router.push({ name: 'home' })"
      >
        <Arrow class="privacy-back__icon" />
        <span>{{ t('commonNav.backHome') }}</span>
      </button>

      <article class="privacy-doc">
        <MarkdownRenderer :content="policy" />
      </article>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { storeToRefs } from 'pinia'
import { MarkdownRenderer } from '@/editor/preview'
import { useSettingStore } from '@/stores'
import Arrow from '@/components/icons/arrow.vue'
import NotFoundPage from '@/views/404/modules/NotFoundPage.vue'

const { t } = useI18n()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

// An empty policy is not "an empty page": it is an unconfigured page, which
// must render the same 404 the server already returned for this path.
const policy = computed(() => SystemSetting.value.privacy_policy?.trim() ?? '')
</script>

<style scoped>
.privacy-page {
  display: flex;
  justify-content: center;
  width: 100%;
  min-height: 100vh;
  padding: 4rem 1.5rem;
}

.privacy-frame {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 46rem;
}

.privacy-back {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  align-self: flex-start;
  padding: 0;
  border: 0;
  background: none;
  color: var(--color-text-muted);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: color 0.15s ease;
}

.privacy-back:hover {
  color: var(--color-text-primary);
}

.privacy-back__icon {
  width: 0.95rem;
  height: 0.95rem;
  transform: rotate(180deg);
  transition: transform 0.15s ease;
}

.privacy-back:hover .privacy-back__icon {
  transform: rotate(180deg) translateX(2px);
}

.privacy-doc {
  margin-top: 2.5rem;
}

@media (width <= 480px) {
  .privacy-page {
    padding: 3rem 1.25rem;
  }
}
</style>
