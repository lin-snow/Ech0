<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { ref, watch } from 'vue'
import BaseAvatar from '@/components/common/BaseAvatar.vue'
import { resolveHubInstanceLogo } from '../utils/resolveHubLogoUrl'
import type { InstanceConnectSummary } from '../services/connectApi'

const props = defineProps<{
  summaries: InstanceConnectSummary[]
}>()

const failed = ref<Record<string, boolean>>({})

watch(
  () => props.summaries,
  () => {
    failed.value = {}
  },
)

function logoUrl(s: InstanceConnectSummary) {
  return resolveHubInstanceLogo(s.rawLogo, s.urlKey)
}

function seed(s: InstanceConnectSummary) {
  return `${s.urlKey}-${s.username || s.serverName}`
}

function dotColor(todayEchos: number): string {
  const n = todayEchos ?? 0
  if (n >= 4) return 'var(--heatmap-bg-color-4)'
  if (n >= 3) return 'var(--heatmap-bg-color-3)'
  if (n >= 2) return 'var(--heatmap-bg-color-2)'
  if (n >= 1) return 'var(--heatmap-bg-color-1)'
  return 'var(--color-border-subtle)'
}
</script>

<template>
  <section v-if="props.summaries.length > 0" class="mt-12 pt-2">
    <p class="mb-4 text-center text-[0.625rem] font-medium uppercase tracking-widest text-[var(--color-text-muted)]">
      Active creators
    </p>
    <div class="mx-auto flex max-w-4xl flex-wrap justify-center gap-x-5 gap-y-4">
      <a
        v-for="s in props.summaries"
        :key="s.urlKey"
        :href="s.urlKey"
        target="_blank"
        rel="noopener noreferrer"
        class="block shrink-0 no-underline"
        :aria-label="s.serverName"
      >
        <div
          class="relative flex h-8 w-8 items-center justify-center sm:h-9 sm:w-9"
        >
          <img
            v-if="logoUrl(s) && !failed[s.urlKey]"
            :src="logoUrl(s)"
            alt=""
            loading="lazy"
            decoding="async"
            class="h-full w-full rounded-full object-cover ring-1 ring-[var(--color-border-subtle)]"
            @error="failed[s.urlKey] = true"
          />
          <BaseAvatar
            v-else
            :seed="seed(s)"
            :size="40"
            alt=""
            class="h-full w-full rounded-full object-cover ring-1 ring-[var(--color-border-subtle)]"
          />
          <span
            class="pointer-events-none absolute right-0 top-0 h-2 w-2 rounded-full border border-[var(--color-bg-surface)]"
            :style="{
              transform: 'translate(35%, -35%)',
              backgroundColor: dotColor(s.todayEchos),
            }"
            aria-hidden="true"
          />
        </div>
      </a>
    </div>
  </section>
</template>
