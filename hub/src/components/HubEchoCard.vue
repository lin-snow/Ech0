<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import Verified from '@/components/icons/verified.vue'
import GrayLike from '@/components/icons/graylike.vue'
import LinkTo from '@/components/icons/linkto.vue'
import BaseAvatar from '@/components/common/BaseAvatar.vue'
import TheMdPreview from '@/components/advanced/md/TheMdPreview.vue'
import { computed, defineAsyncComponent, ref, watch } from 'vue'
import { ImageLayout } from '@/enums/enums'
import { useFetch } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { localStg } from '../utils/storage'
import { formatHubDate } from '../utils/formatHubDate'

const TheMediaPlayer = defineAsyncComponent({
  loader: () => import('@/components/advanced/media/TheMediaPlayer.vue'),
  onError(error, retry, fail, attempts) {
    if (attempts <= 3) {
      retry()
    } else {
      fail()
    }
  },
})

type Echo = App.Api.Hub.Echo

const props = withDefaults(
  defineProps<{
    echo: Echo
    variant?: 'default' | 'masonry'
  }>(),
  { variant: 'default' },
)

const { t } = useI18n()

const fav_count = ref<number>(props.echo.fav_count)
const server_url = computed(() => props.echo.server_url)
const echo_id = computed(() => props.echo.id)
const isLikeAnimating = ref(false)
const LIKE_LIST_KEY = computed(() => `${server_url.value}_liked_echo_ids`)

watch(
  () => props.echo.fav_count,
  (next) => {
    fav_count.value = next
  },
)

const logoUrl = computed(() => props.echo.logo?.trim() ?? '')
const avatarFailed = ref(false)
const avatarSeed = computed(
  () => `${props.echo.server_url}-${props.echo.username}-${props.echo.id}`,
)

watch(
  () => props.echo.logo,
  () => {
    avatarFailed.value = false
  },
)

const handleLikeEcho = async () => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250)

  const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY.value) || []
  if (likedEchoIds.includes(echo_id.value)) {
    return
  }

  const { error, data } = await useFetch<App.Api.Response<null>>(
    `${server_url.value}/api/echo/like/${echo_id.value}`,
  )
    .put()
    .json()

  if (!error.value && data.value?.code === 1) {
    fav_count.value += 1
    likedEchoIds.push(echo_id.value)
    localStg.setItem(LIKE_LIST_KEY.value, likedEchoIds)
  }
}
</script>

<template>
  <div
    :class="[
      'w-full rounded-sm border border-[var(--color-border-strong)] h-auto px-3 py-3 sm:px-3.5 sm:py-3.5',
      props.variant === 'masonry' ? 'max-w-none' : 'max-w-sm mx-auto',
    ]"
  >
    <div class="flex flex-row items-center gap-2 mb-3">
      <div class="shrink-0">
        <img
          v-if="logoUrl && !avatarFailed"
          :src="logoUrl"
          alt=""
          loading="lazy"
          decoding="async"
          class="h-10 w-10 sm:h-12 sm:w-12 rounded-full object-cover ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)]"
          @error="avatarFailed = true"
        />
        <BaseAvatar
          v-else
          :seed="avatarSeed"
          :size="48"
          alt="avatar"
          class="h-10 w-10 sm:h-12 sm:w-12 rounded-full object-cover ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-sm)]"
        />
      </div>
      <div class="flex flex-col">
        <div class="flex items-center gap-1">
          <h2
            class="text-[var(--color-text-primary)] font-bold overflow-hidden whitespace-nowrap text-center"
          >
            <a :href="echo.server_url" target="_blank" rel="noopener noreferrer">{{
              echo.server_name
            }}</a>
          </h2>
          <div>
            <Verified class="text-sky-500 w-5 h-5" />
          </div>
        </div>
        <span class="hub-echo-username text-[var(--color-text-secondary)]"
          >@ {{ echo.username }}</span
        >
      </div>
    </div>

    <div>
      <div class="py-2">
        <template
          v-if="
            props.echo.layout === ImageLayout.GRID ||
            props.echo.layout === ImageLayout.HORIZONTAL ||
            props.echo.layout === ImageLayout.STACK
          "
        >
          <div class="mx-auto w-11/12 pl-1 mb-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
          <TheMediaPlayer
            :echo="props.echo"
            :base-url="echo.server_url"
            :layout="props.echo.layout"
          />
        </template>
        <template v-else>
          <TheMediaPlayer
            :echo="props.echo"
            :base-url="echo.server_url"
            :layout="props.echo.layout"
          />
          <div class="mx-auto w-11/12 pl-1 mt-3">
            <TheMdPreview :content="props.echo.content" />
          </div>
        </template>
      </div>
    </div>

    <div class="flex items-center justify-between gap-2">
      <div class="min-w-0 flex flex-1 items-center overflow-hidden">
        <div class="min-w-0 truncate whitespace-nowrap text-sm text-slate-500">
          {{ formatHubDate(props.echo.created_at) }}
        </div>
        <div
          v-if="props.echo.tags?.[0]?.name"
          class="hidden min-w-0 flex-shrink truncate whitespace-nowrap text-xs text-[var(--color-text-muted)] sm:block sm:ml-1"
        >
          #{{ props.echo.tags[0]?.name }}
        </div>
      </div>

      <div class="relative flex h-auto flex-none items-center justify-center gap-2">
        <a
          :href="`${server_url}/echo/${echo_id}`"
          target="_blank"
          rel="noopener noreferrer"
          v-tooltip="t('hubEcho.jumpToEcho')"
        >
          <LinkTo class="w-4 h-4" />
        </a>

        <div class="flex items-center justify-end" v-tooltip="t('hubEcho.like')">
          <div class="flex items-center gap-1">
            <button
              type="button"
              @click="handleLikeEcho()"
              :class="[
                'transform transition-transform duration-150',
                isLikeAnimating ? 'scale-160' : 'scale-100',
              ]"
            >
              <GrayLike class="w-4 h-4" />
            </button>
            <span class="text-sm text-[var(--color-text-muted)]">
              {{ fav_count > 99 ? '99+' : fav_count }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="css">
.hub-echo-username {
  font-family: var(--font-family-display);
}
</style>
