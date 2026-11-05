<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="px-3 pb-4 py-2 mt-4 sm:mt-6 mb-10 mx-auto flex justify-center items-center">
    <div class="w-full sm:max-w-lg mx-auto">
      <div v-if="echo" class="w-full sm:mt-1 mx-auto">
        <TheEchoDetail :echo="echo" @update-like-count="handleUpdateLikeCount" />
        <TheEchoInteractions />
      </div>
      <div v-else class="w-full sm:mt-1 text-[var(--color-text-muted)]">
        <p class="text-center">{{ t('echoPage.loadingDetail') }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import TheEchoDetail from '@/components/advanced/echo/cards/TheEchoDetail.vue'
import TheEchoInteractions from '@/components/advanced/echo/cards/TheEchoInteractions.vue'
import { clearSeoOverride, setSeoOverride } from '@/composables/useSeoOverride'
import { DEFAULT_SEO_DESCRIPTION } from '@/composables/useSeoHead'
import { DEFAULT_SITE_TITLE } from '@/constants/site'
import { useEchoStore } from '@/stores'
import { getEchoImages } from '@/utils/echo'
import { absoluteSeoUrl, seoPlainText, seoTruncate } from '@/utils/seo'
import { useI18n } from 'vue-i18n'

const route = useRoute()
const { t, locale } = useI18n()

const echoStore = useEchoStore()
const isLoading = ref(true)
const echo = ref<App.Api.Ech0.Echo | null>(null)

const getEchoFromStore = (echoId: string): App.Api.Ech0.Echo | null => {
  const idx = echoStore.echoIndexMap.get(echoId)
  if (idx !== undefined) {
    return echoStore.echoList[idx] ?? null
  }
  return null
}

const handleUpdateLikeCount = () => {
  if (echo.value) {
    echo.value.fav_count += 1
  }
}

const currentEchoId = () => String(route.params.echoId ?? '').trim()

const loadEcho = async () => {
  const echoId = currentEchoId()
  if (!echoId) {
    isLoading.value = false
    return
  }

  isLoading.value = true
  echo.value = getEchoFromStore(echoId)

  if (!echo.value) {
    echo.value = await echoStore.prefetchEcho(echoId)
  }
  isLoading.value = false
}

// Mirrors the server-rendered head so client navigation and the crawlable
// document agree on the card, canonical URL and BlogPosting payload.
const applyEchoSeo = () => {
  const value = echo.value
  if (!value) return

  const text = seoPlainText(value.content)
  const title = seoTruncate(text.split('\n')[0] || 'Echo', 60)
  const description = seoTruncate(text.replace(/\n/g, ' ') || DEFAULT_SEO_DESCRIPTION, 160)
  const image = absoluteSeoUrl(getEchoImages(value)[0]?.url)
  const createdAt = Number(value.created_at)
  const publishedAt =
    Number.isFinite(createdAt) && createdAt > 0
      ? new Date(createdAt * 1000).toISOString()
      : undefined

  setSeoOverride({
    path: route.path,
    title,
    description,
    type: 'article',
    image,
    publishedAt,
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'BlogPosting',
      headline: title,
      description,
      datePublished: publishedAt,
      dateModified: publishedAt,
      inLanguage: locale.value,
      image,
      author: { '@type': 'Person', name: value.username || DEFAULT_SITE_TITLE },
    },
  })
}

watch(echo, applyEchoSeo)

watch(
  () => route.params.echoId,
  () => {
    clearSeoOverride()
    loadEcho()
  },
)

onMounted(loadEcho)

onBeforeUnmount(() => clearSeoOverride(route.path))
</script>
