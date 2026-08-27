<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <article class="echo-detail w-full max-w-sm mx-auto">
    <header class="echo-detail-head">
      <img
        :src="logo"
        alt="logo"
        loading="lazy"
        decoding="async"
        class="w-10 h-10 sm:w-12 sm:h-12 rounded-full ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)] object-cover"
      />
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            {{ SystemSetting.server_name }}
          </h2>
          <Verified class="text-sky-500 w-5 h-5" />
        </div>
        <span class="echo-username text-[var(--color-text-secondary)]">@ {{ echo.username }}</span>
      </div>
      <button
        type="button"
        class="echo-detail-back ml-auto cursor-pointer rounded-full px-4 py-1 text-sm font-medium ring-1 ring-inset ring-[var(--color-border-subtle)] text-[var(--color-text-primary)] bg-transparent transition-colors duration-200 hover:bg-[var(--color-bg-muted)] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--color-accent)]"
        :aria-label="t('commonNav.back')"
        @click="goBack"
      >
        {{ t('commonNav.back') }}
      </button>
    </header>

    <section class="echo-detail-body">
      <template v-if="isContentLeadingEcho(props.echo)">
        <div class="mb-3">
          <TheMdPreview :content="props.echo.content" />
        </div>

        <TheMediaPlayer :echo="props.echo" :layout="props.echo.layout" priority />
      </template>

      <template v-else>
        <TheMediaPlayer :echo="props.echo" :layout="props.echo.layout" priority />

        <div class="mt-3">
          <TheMdPreview :content="props.echo.content" />
        </div>
      </template>

      <div v-if="props.echo.extension" class="my-4">
        <TheExtensionRenderer :echo="props.echo" />
      </div>
    </section>

    <TheEchoMeta :echo="props.echo" @update-like-count="(id) => emit('updateLikeCount', id)" />
  </article>
</template>

<script setup lang="ts">
import Verified from '@/components/icons/verified.vue'
import TheEchoMeta from '@/components/advanced/echo/cards/TheEchoMeta.vue'
import { computed, defineAsyncComponent } from 'vue'
import { storeToRefs } from 'pinia'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSettingStore } from '@/stores'
import { resolveAvatarUrl } from '@/service/request/shared'
import { isContentLeadingEcho } from '@/utils/echo'
import { TheMdPreview } from '@/components/advanced/md'

const TheMediaPlayer = defineAsyncComponent(
  () => import('@/components/advanced/media/TheMediaPlayer.vue'),
)
const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
}>()

const emit = defineEmits<{
  (e: 'updateLikeCount', echoId: string): void
}>()

const settingStore = useSettingStore()

const { SystemSetting } = storeToRefs(settingStore)
const logo = computed(() => resolveAvatarUrl(SystemSetting.value?.server_logo))

const router = useRouter()
const { t } = useI18n()

const goBack = () => {
  if (window.history.length > 2) {
    window.history.back()
  } else {
    router.push({ name: 'home' })
  }
}
</script>

<style scoped lang="css">
.echo-detail {
  background-color: transparent;
  padding: 0.5rem 0.25rem;
}

.echo-detail-head {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  background-image: linear-gradient(
    to right,
    var(--color-border-subtle) 0,
    var(--color-border-subtle) 5px,
    transparent 5px,
    transparent 8px
  );
  background-size: 8px 1px;
  background-repeat: repeat-x;
  background-position: left bottom;
}

.echo-detail-body {
  padding: 2rem 0;
}

.echo-detail-body :deep(.image-gallery-container > div) {
  width: 100%;
}

.echo-username {
  font-family: var(--font-family-display);
}
</style>
