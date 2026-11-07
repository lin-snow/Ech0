<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="echo-timeline group w-full">
    <div class="echo-header-sticky flex justify-between items-center">
      <div class="flex justify-start items-center h-9">
        <div class="flex items-center h-full pr-1">
          <div class="timeline-marker" :class="{ 'is-first': props.index === 0 }">
            <div class="w-2 h-2 rounded-full bg-[var(--color-accent)]"></div>
          </div>
          <div
            @click="handleExpandEcho(echo.id)"
            class="flex items-center h-full justify-start leading-none text-sm font-semibold text-nowrap text-[var(--color-accent)] cursor-pointer hover:underline hover:decoration-offset-3 hover:decoration-1 mr-1"
          >
            {{ formatDate(props.echo.created_at) }}
          </div>
          <button
            type="button"
            class="echo-open-btn flex items-center justify-center w-6 h-6 rounded-sm text-[var(--color-text-muted)] opacity-0 transition-opacity duration-150 hover:text-[var(--color-text-primary)] focus-visible:opacity-100 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-[var(--color-border-subtle)]"
            :aria-label="t('echoCard.openDetail')"
            v-tooltip="t('echoCard.openDetail')"
            @click="handleExpandEcho(echo.id)"
          >
            <Open class="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      <div
        v-if="!userStore.isLogin && props.echo.private"
        v-tooltip="t('echoCard.privateStatus')"
        class="w-7 h-7 flex items-center justify-center bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] ring-inset rounded-full shadow-sm"
      >
        <Lock class="w-4 h-4" />
      </div>

      <EchoActionMenu v-else-if="userStore.isLogin" :echo="props.echo" />
    </div>

    <div class="timeline-content">
      <div class="px-4 py-3">
        <template v-if="isContentLeadingEcho(props.echo)">
          <div class="mx-auto w-11/12 pl-1 mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>

          <div :class="{ 'mx-auto w-11/12 pl-1': isAvEcho }">
            <TheMediaPlayer
              :echo="props.echo"
              :layout="props.echo.layout"
              :priority="props.index === 0"
            />
          </div>
        </template>

        <template v-else>
          <div :class="{ 'mx-auto w-11/12 pl-1': isAvEcho }">
            <TheMediaPlayer
              :echo="props.echo"
              :layout="props.echo.layout"
              :priority="props.index === 0"
            />
          </div>

          <div class="mx-auto w-11/12 pl-1 mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>

        <div v-if="props.echo.extension" class="my-2">
          <TheExtensionRenderer :echo="props.echo" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { useUserStore } from '@/stores'
import { TheMdPreview } from '@/components/advanced/md'
import Lock from '@/components/icons/lock.vue'
import Open from '@/components/icons/open.vue'
import { useRouter } from 'vue-router'
import { getEchoFilesBy, isContentLeadingEcho } from '@/utils/echo'
import { formatDate } from '@/utils/other'
import { useI18n } from 'vue-i18n'
import EchoActionMenu from '@/components/advanced/echo/EchoActionMenu.vue'

const TheMediaPlayer = defineAsyncComponent(
  () => import('@/components/advanced/media/TheMediaPlayer.vue'),
)
const TheExtensionRenderer = defineAsyncComponent(
  () => import('@/components/advanced/extension/TheExtensionRenderer.vue'),
)

const { t } = useI18n()

type Echo = App.Api.Ech0.Echo

const props = defineProps<{
  echo: Echo
  index?: number
}>()

const isAvEcho = computed(
  () => getEchoFilesBy(props.echo, { categories: ['audio', 'video'] }).length > 0,
)

const userStore = useUserStore()

const router = useRouter()

const handleExpandEcho = (echoId: string) => {
  router.push({
    name: 'echo',
    params: { echoId: echoId },
  })
}
</script>

<style scoped lang="css">
.echo-header-sticky {
  position: relative;
  z-index: 1;
  background-color: var(--color-bg-canvas);
}

.echo-timeline {
  --axis-offset: calc(0.25rem + 1px);
  --axis-line-width: 2px;
  --axis-dot-size: 0.5rem;
  --axis-dot-gap: 0.3rem;

  max-width: 100%;
  overflow: clip visible;
}

.timeline-marker {
  position: relative;
  width: var(--axis-dot-size);
  height: 100%;
  margin-right: 0.5rem;
  margin-left: calc(var(--axis-offset) - (var(--axis-dot-size) / 2));
  display: flex;
  align-items: center;
  justify-content: center;
}

.timeline-content {
  position: relative;
  margin-left: var(--axis-offset);
  max-width: 100%;
  min-width: 0;
  overflow: clip visible;
}

.timeline-content::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 0;
  width: var(--axis-line-width);
  background-color: var(--color-border-subtle);
  pointer-events: none;
}

@media (hover: hover) {
  .echo-timeline:hover .echo-open-btn {
    opacity: 1;
    animation: echo-open-nudge-left 1200ms ease-out both;
  }
}

.echo-open-btn:focus-visible {
  animation: echo-open-nudge-left 1200ms ease-out both;
}

@keyframes echo-open-nudge-left {
  0% {
    transform: translateX(0);
  }

  16% {
    transform: translateX(-4px);
  }

  28% {
    transform: translateX(1px);
  }

  42% {
    transform: translateX(-3px);
  }

  54% {
    transform: translateX(1px);
  }

  68% {
    transform: translateX(-2px);
  }

  80% {
    transform: translateX(0);
  }

  100% {
    transform: translateX(0);
  }
}
</style>
