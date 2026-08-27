<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div v-if="images.length" class="image-gallery-container">
    <GalleryWaterfall
      v-if="layoutValue === ImageLayout.WATERFALL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-aspect-ratio-style="getAspectRatioStyle"
      :priority="props.priority"
    />

    <GalleryGrid
      v-if="layoutValue === ImageLayout.GRID"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :priority="props.priority"
    />

    <GalleryCarousel
      v-if="layoutValue === ImageLayout.CAROUSEL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-aspect-ratio-style="getAspectRatioStyle"
      :priority="props.priority"
    />

    <GalleryHorizontal
      v-if="layoutValue === ImageLayout.HORIZONTAL"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :scroll-hint-text="t('imageGallery.scrollHint')"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :get-horizontal-aspect-style="getHorizontalAspectStyle"
      :priority="props.priority"
    />

    <GalleryStack
      v-if="layoutValue === ImageLayout.STACK"
      :images="images"
      :resolved-srcs="resolvedSrcs"
      :get-alt="getAlt"
      :get-image-key="getImageKey"
      :is-loaded="isImageLoaded"
      :mark-loaded="markImageLoaded"
      :open="openGallery"
      :priority="props.priority"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { getImageUrl, getHubImageUrl } from '@/utils/other'
import { ImageLayout } from '@/enums/enums'
import { useI18n } from 'vue-i18n'
import { usePhotoSwipeGallery } from './composables/usePhotoSwipeGallery'
import GalleryWaterfall from './layouts/GalleryWaterfall.vue'
import GalleryGrid from './layouts/GalleryGrid.vue'
import GalleryCarousel from './layouts/GalleryCarousel.vue'
import GalleryHorizontal from './layouts/GalleryHorizontal.vue'
import GalleryStack from './layouts/GalleryStack.vue'

const props = withDefaults(
  defineProps<{
    images?: App.Api.Ech0.FileObject[]
    baseUrl?: string
    layout?: ImageLayout | string | undefined
    priority?: boolean
  }>(),
  { priority: false },
)

const { t } = useI18n()

const images = computed(() => props.images || [])
const isValidLayout = (layout?: ImageLayout | string): layout is ImageLayout => {
  if (!layout) return false
  return (Object.values(ImageLayout) as string[]).includes(layout)
}

const layoutValue = computed(() =>
  isValidLayout(props.layout) ? props.layout : ImageLayout.WATERFALL,
)
const loadedImages = ref<Record<string, boolean>>({})
const resolvedSrcs = computed(() =>
  images.value.map((image) =>
    props.baseUrl ? getHubImageUrl(image, props.baseUrl) : getImageUrl(image),
  ),
)

const galleryItems = computed(() =>
  images.value.map((image, idx) => ({
    src: resolvedSrcs.value[idx] || '',
    width: image.width,
    height: image.height,
    alt: getAlt(idx),
  })),
)

const { open } = usePhotoSwipeGallery(galleryItems)

const getAlt = (idx: number) => t('imageGallery.previewImage', { index: idx + 1 })

const openGallery = (index: number, sourceElement?: HTMLElement | null) => {
  open(index, sourceElement)
}

const getImageKey = (image: App.Api.Ech0.FileObject, idx: number) =>
  image.id || `${image.url}-${idx}`

const FALLBACK_ASPECT_RATIO = '4 / 3'

const getAspectRatioStyle = (
  image: App.Api.Ech0.FileObject,
  loaded: boolean,
): Record<string, string> | undefined => {
  if (image.width && image.height) return { aspectRatio: `${image.width} / ${image.height}` }
  if (loaded) return undefined
  return { aspectRatio: FALLBACK_ASPECT_RATIO }
}

const getHorizontalAspectStyle = (
  image: App.Api.Ech0.FileObject,
  loaded: boolean,
): Record<string, string> | undefined => {
  if (image.width && image.height) {
    return { aspectRatio: `${image.width} / ${image.height}`, width: 'auto' }
  }
  if (loaded) return undefined
  return { aspectRatio: FALLBACK_ASPECT_RATIO, width: 'auto' }
}

const isImageLoaded = (image: App.Api.Ech0.FileObject, idx: number) =>
  Boolean(loadedImages.value[getImageKey(image, idx)])

const markImageLoaded = (image: App.Api.Ech0.FileObject, idx: number) => {
  loadedImages.value[getImageKey(image, idx)] = true
}

watch(
  images,
  (nextImages) => {
    const nextKeys = new Set(nextImages.map((image, idx) => getImageKey(image, idx)))
    const nextLoadedMap: Record<string, boolean> = {}

    Object.keys(loadedImages.value).forEach((key) => {
      if (nextKeys.has(key)) {
        nextLoadedMap[key] = true
      }
    })

    loadedImages.value = nextLoadedMap
  },
  { immediate: true },
)
</script>

<style scoped>
.image-gallery-container {
  position: relative;
}
</style>
