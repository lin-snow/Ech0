<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="echo-meta">
    <div v-if="tags.length > 0" class="echo-meta-tags">
      <span v-for="tag in tags" :key="tag.id" class="echo-meta-chip"> #{{ tag.name }} </span>
    </div>

    <div class="echo-meta-line">
      <time class="echo-meta-item" :datetime="String(props.echo.created_at)">
        {{ formatDateTime(props.echo.created_at) }}
      </time>
      <span class="echo-meta-dot" aria-hidden="true">·</span>
      <span class="echo-meta-item">
        {{ t('echoDetail.metaWordCountValue', { count: wordCount }) }}
      </span>
      <span v-if="props.echo.private" class="echo-meta-dot" aria-hidden="true">·</span>
      <span v-if="props.echo.private" class="echo-meta-item echo-meta-item--lock">
        <Lock class="w-3 h-3" />
        {{ t('echoDetail.metaPrivate') }}
      </span>
      <div class="echo-meta-actions">
        <TheEchoReactions :echo="props.echo" />
        <TheShareEchoPanel :echo-id="props.echo.id" :echo-content="props.echo.content" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Lock from '@/components/icons/lock.vue'
import TheShareEchoPanel from '@/components/advanced/echo/cards/TheShareEchoPanel.vue'
import TheEchoReactions from '@/components/advanced/echo/cards/TheEchoReactions.vue'
import { formatDateTime } from '@/utils/other'
import { countWords } from '@/utils/echo'

const { t } = useI18n()

const props = defineProps<{
  echo: App.Api.Ech0.Echo
}>()

const wordCount = computed(() => countWords(props.echo.content))
const tags = computed(() => props.echo.tags ?? [])
</script>

<style scoped>
.echo-meta {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.5rem 0 0;
}

.echo-meta-line {
  display: flex;
  flex-wrap: nowrap;
  align-items: center;
  gap: 0.35rem;
  min-width: 0;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  letter-spacing: 0.02em;
}

.echo-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  white-space: nowrap;
}

.echo-meta-item--lock {
  color: var(--color-text-secondary);
}

.echo-meta-dot {
  color: var(--color-border-strong);
}

.echo-meta-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.echo-meta-chip {
  display: inline-flex;
  align-items: center;
  border: 1px dashed var(--color-border-subtle);
  border-radius: var(--radius-sm);
  padding: 0.08rem 0.4rem;
  color: var(--color-text-muted);
  font-size: 0.7rem;
  line-height: 1.3;
}

.echo-meta-actions {
  margin-left: auto;
  display: inline-flex;
  flex: none;
  align-items: center;
  gap: 0.55rem;
}
</style>
