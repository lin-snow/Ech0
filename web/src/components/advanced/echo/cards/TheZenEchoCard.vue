<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    class="zen-echo-card relative w-full bg-[var(--color-bg-surface)] h-auto p-3 sm:p-3.5 shadow rounded-md"
  >
    <div class="flex flex-row items-center justify-between gap-2 mt-1 mb-3">
      <div class="flex flex-row items-center gap-2 min-w-0">
        <div class="flex-none">
          <img
            :src="avatarUrl"
            alt=""
            loading="lazy"
            decoding="async"
            class="w-6 h-6 sm:w-7 sm:h-7 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] object-cover"
          />
        </div>
        <div class="flex items-center gap-1 min-w-0">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap truncate"
          >
            {{ siteName }}
          </h2>
        </div>
      </div>

      <div class="flex flex-none items-center gap-2">
        <Lock
          v-if="echo.private"
          class="w-4 h-4 sm:w-[1.125rem] sm:h-[1.125rem] text-[var(--color-text-muted)]"
        />
        <button
          type="button"
          v-tooltip="t('zenMode.openDetail')"
          :aria-label="t('zenMode.openDetail')"
          class="flex-none opacity-70 hover:opacity-100 transition-opacity"
          @click="goDetail"
        >
          <img
            src="/Ech0.svg"
            alt="Ech0"
            loading="lazy"
            decoding="async"
            class="w-5 h-5 sm:w-6 sm:h-6 object-contain"
          />
        </button>
      </div>
    </div>

    <div class="zen-echo-body py-1.5">
      <template v-if="isContentLeadingEcho(echo)">
        <div v-if="echo.content" class="mb-2.5">
          <TheMdPreview :content="echo.content" />
        </div>

        <TheMediaPlayer :echo="echo" :layout="echo.layout" />
      </template>

      <template v-else>
        <TheMediaPlayer :echo="echo" :layout="echo.layout" />

        <div v-if="echo.content" :class="hasMedia ? 'mt-2.5' : ''">
          <TheMdPreview :content="echo.content" />
        </div>
      </template>
    </div>

    <div class="mt-1 flex items-center justify-between gap-2">
      <div class="min-w-0 truncate whitespace-nowrap text-xs text-[var(--color-text-muted)]">
        {{ formatDate(echo.created_at) }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { TheMdPreview } from '@/components/advanced/md'
import Lock from '@/components/icons/lock.vue'
import { useSettingStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { formatDate } from '@/utils/other'
import { getEchoFilesBy, isContentLeadingEcho } from '@/utils/echo'

const { t } = useI18n()

const TheMediaPlayer = defineAsyncComponent(
  () => import('@/components/advanced/media/TheMediaPlayer.vue'),
)

const props = defineProps<{
  echo: App.Api.Ech0.Echo
  index?: number
}>()

const router = useRouter()
const settingStore = useSettingStore()
const { SystemSetting } = storeToRefs(settingStore)

const siteName = computed(() => String(SystemSetting.value?.server_name ?? 'Ech0'))
const avatarUrl = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))
const hasMedia = computed(() => getEchoFilesBy(props.echo, { dedupeBy: 'id' }).length > 0)

const goDetail = () => {
  router.push({ name: 'echo', params: { echoId: String(props.echo.id) } })
}
</script>

<style scoped lang="css">
.zen-echo-card {
  contain: layout paint;
}

.zen-echo-card::before {
  content: '';
  position: absolute;

  top: 1.25rem;
  left: 0;
  width: 2px;
  height: 1rem;
  border-radius: 0 2px 2px 0;
  background: var(--color-accent);
  opacity: 0.85;
}

@media (width >= 640px) {
  .zen-echo-card::before {
    top: 1.4375rem;
    height: 1.125rem;
  }
}

.zen-echo-body :deep(p),
.zen-echo-body :deep(li) {
  font-size: 0.9rem;
  line-height: 1.55;
}

.zen-echo-body :deep(p) {
  margin: 0 0 0.55rem;
}

.zen-echo-body :deep(p:last-child) {
  margin-bottom: 0;
}

.zen-echo-body :deep(.image-gallery-container) > div {
  width: 100%;
  margin-left: 0;
  margin-right: 0;
  margin-bottom: 0;
}
</style>
