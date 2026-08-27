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
        <TheShareEchoPanel :echo-id="props.echo.id" :echo-content="props.echo.content" />
        <button
          type="button"
          class="echo-meta-like"
          v-tooltip="t('echoDetail.like')"
          @click="handleLikeEcho(props.echo.id)"
        >
          <span
            :class="[
              'transform transition-transform duration-150 inline-flex',
              isLikeAnimating ? 'scale-160' : 'scale-100',
            ]"
          >
            <GrayLike class="w-4 h-4" />
          </span>
          <span class="text-xs text-[var(--color-text-muted)]">
            {{ props.echo.fav_count > 99 ? '99+' : props.echo.fav_count }}
          </span>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import Lock from '@/components/icons/lock.vue'
import GrayLike from '@/components/icons/graylike.vue'
import TheShareEchoPanel from '@/components/advanced/echo/cards/TheShareEchoPanel.vue'
import { formatDateTime } from '@/utils/other'
import { countWords } from '@/utils/echo'
import { fetchLikeEcho } from '@/service/api'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { isStaticMode } from '@/service/request/shared'

const { t } = useI18n()

const props = defineProps<{
  echo: App.Api.Ech0.Echo
}>()

const emit = defineEmits<{
  (e: 'updateLikeCount', echoId: string): void
}>()

const wordCount = computed(() => countWords(props.echo.content))
const tags = computed(() => props.echo.tags ?? [])

const isLikeAnimating = ref(false)

const LIKE_LIST_KEY = 'likedEchoIds'
const likedEchoIds: string[] = localStg.getItem(LIKE_LIST_KEY) || []
const hasLikedEcho = (echoId: string): boolean => likedEchoIds.includes(echoId)

const handleLikeEcho = (echoId: string) => {
  isLikeAnimating.value = true
  setTimeout(() => {
    isLikeAnimating.value = false
  }, 250)

  if (isStaticMode()) {
    theToast.info(String(t('staticSite.likeUnavailable')))
    return
  }

  if (hasLikedEcho(echoId)) {
    theToast.info(String(t('echoDetail.alreadyLiked')))
    return
  }

  fetchLikeEcho(echoId).then((res) => {
    if (res.code === 1) {
      likedEchoIds.push(echoId)
      localStg.setItem(LIKE_LIST_KEY, likedEchoIds)
      emit('updateLikeCount', echoId)
      theToast.info(String(t('echoDetail.likeSuccess')))
    }
  })
}
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
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem;
  color: var(--color-text-muted);
  font-family: var(--font-family-mono);
  font-size: 0.72rem;
  letter-spacing: 0.02em;
}

.echo-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
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
  align-items: center;
  gap: 0.5rem;
}

.echo-meta-like {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
}
</style>
