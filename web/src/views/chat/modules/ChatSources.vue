<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="sources">
    <div v-for="src in visibleSources" :key="src.echoId" class="sources__item">
      <button class="sources__link" @click="emit('open', src.echoId)">
        <span class="sources__mark">↗</span>
        <span class="sources__day">{{ src.day }}</span>
        <span class="sources__sep" aria-hidden="true">·</span>
        <span class="sources__text" :class="{ 'sources__text--empty': src.empty }">{{
          src.text
        }}</span>
        <span v-if="src.ext" class="sources__ext">
          <span aria-hidden="true">{{ src.ext.icon }}</span>
          {{ src.ext.label }}
        </span>
        <span v-for="badge in src.mediaBadges" :key="badge.label" class="sources__ext">
          <span aria-hidden="true">{{ badge.icon }}</span>
          {{ badge.label }}
        </span>
      </button>
      <div v-if="src.images.length" class="sources__thumbs">
        <button
          v-for="(img, i) in src.images"
          :key="img.id || i"
          class="sources__thumb"
          @click="emit('open', src.echoId)"
        >
          <img :src="getImageUrl(img)" alt="" loading="lazy" />
        </button>
      </div>
    </div>
    <button
      v-if="displaySources.length > LIMIT"
      class="sources__toggle"
      @click="showAll = !showAll"
    >
      {{
        showAll ? t('chatPanel.sourcesLess') : t('chatPanel.sourcesMore', { count: hiddenCount })
      }}
    </button>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { getImageUrl } from '@/utils/other'

const props = defineProps<{
  sources: App.Api.Chat.ChatSource[]
}>()

const emit = defineEmits<{
  (e: 'open', echoId: string): void
}>()

const { t } = useI18n()

const LIMIT = 3
const THUMB_LIMIT = 4
const showAll = ref<boolean>(false)

type DisplaySource = {
  echoId: string
  day: string
  text: string
  empty: boolean
  images: App.Api.Ech0.FileObject[]
  ext?: { icon: string; label: string }
  mediaBadges: { icon: string; label: string }[]
}

const MEDIA_META: Record<string, { icon: string; labelKey: string }> = {
  video: { icon: '🎬', labelKey: 'editor.mediaTypeVideo' },
  audio: { icon: '🎵', labelKey: 'editor.mediaTypeAudio' },
}

const EXT_META: Record<string, { icon: string; labelKey: string }> = {
  MUSIC: { icon: '🎵', labelKey: 'editor.extMusic' },
  VIDEO: { icon: '🎬', labelKey: 'editor.extVideo' },
  GITHUBPROJ: { icon: '💻', labelKey: 'editor.extGithubProject' },
  WEBSITE: { icon: '🔗', labelKey: 'editor.extWebsiteLink' },
  LOCATION: { icon: '📍', labelKey: 'editor.extLocation' },
  TWEET: { icon: '🐦', labelKey: 'editor.extTweet' },
}

const displaySources = computed<DisplaySource[]>(() =>
  props.sources.map((src) => {
    const day = new Date(src.echo_created * 1000).toISOString().slice(0, 10)
    const trimmed = src.content.trim()
    const empty = trimmed.length === 0
    const text = empty
      ? t('chatPanel.sourceNoContent')
      : trimmed.length > 40
        ? trimmed.slice(0, 40) + '…'
        : trimmed
    const images: App.Api.Ech0.FileObject[] = (src.files ?? [])
      .filter((f) => f.category === 'image' && f.url)
      .slice(0, THUMB_LIMIT)
      .map((f) => ({ ...f, echo_id: src.echo_id }))
    const meta = src.extension ? EXT_META[src.extension.type] : undefined
    const ext = meta ? { icon: meta.icon, label: t(meta.labelKey) } : undefined
    const mediaTypes = new Set<string>()
    for (const f of src.files ?? []) {
      if (f.category && f.category in MEDIA_META) mediaTypes.add(f.category)
    }
    const mediaBadges = [...mediaTypes].map((c) => ({
      icon: MEDIA_META[c].icon,
      label: t(MEDIA_META[c].labelKey),
    }))
    return { echoId: src.echo_id, day, text, empty, images, ext, mediaBadges }
  }),
)

const visibleSources = computed<DisplaySource[]>(() =>
  showAll.value ? displaySources.value : displaySources.value.slice(0, LIMIT),
)
const hiddenCount = computed(() => displaySources.value.length - LIMIT)
</script>

<style scoped>
.sources {
  display: flex;
  flex-direction: column;
  gap: 0.15rem;
  margin-top: 0.35rem;
  min-width: 0;
  max-width: 100%;
}

.sources__item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 0;
  max-width: 100%;
}

.sources__thumbs {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
  margin: 0.1rem 0 0.15rem;
}

.sources__thumb {
  width: 2.75rem;
  height: 2.75rem;
  padding: 0;
  border: 1px solid var(--color-border-strong);
  border-radius: 0.4rem;
  overflow: hidden;
  background: var(--color-accent-soft);
  cursor: pointer;
  transition:
    transform 0.18s ease,
    border-color 0.18s ease;
}

.sources__thumb:hover {
  transform: translateY(-1px);
  border-color: var(--color-accent);
}

.sources__thumb img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.sources__link {
  display: inline-flex;
  align-items: baseline;
  gap: 0.3rem;

  max-width: 100%;
  min-width: 0;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--color-text-muted);
  text-align: left;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: color 0.18s ease;
}

.sources__link:hover {
  color: var(--color-accent);
}

.sources__mark {
  color: var(--color-accent);
  opacity: 0.8;
}

.sources__day {
  font-variant-numeric: tabular-nums;
  flex: none;
}

.sources__sep {
  flex: none;
  opacity: 0.5;
}

.sources__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sources__text--empty {
  font-style: italic;
  opacity: 0.7;
}

.sources__ext {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 0.15rem;
  padding: 0 0.35rem;
  border-radius: 999px;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  font-size: 0.66rem;
  line-height: 1.6;
}

.sources__toggle {
  align-self: flex-start;
  margin-top: 0.1rem;
  border: none;
  background: transparent;
  padding: 0;
  font-size: 0.72rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: color 0.18s ease;
}

.sources__toggle:hover {
  color: var(--color-accent);
}
</style>
