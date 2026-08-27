<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div
    :class="[
      'w-[88%] mx-auto grid gap-2 mb-4',
      images.length === 1 ? 'grid-cols-1 justify-items-center' : 'grid-cols-2',
    ]"
  >
    <GalleryImageItem
      v-for="(image, idx) in images"
      :key="getImageKey(image, idx)"
      :image="image"
      :src="resolvedSrcs[idx] || ''"
      :alt="getAlt(idx)"
      :loaded="isLoaded(image, idx)"
      :priority="!!priority && idx === 0"
      :button-class="getColSpan(idx, images.length)"
      :frame-style="getAspectRatioStyle(image, isLoaded(image, idx))"
      img-class="block max-w-full h-auto"
      @click="open(idx, $event)"
      @load="markLoaded(image, idx)"
      @error="markLoaded(image, idx)"
    />
  </div>
</template>

<script setup lang="ts">
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryWaterfallProps } from './types'

defineProps<GalleryWaterfallProps>()

const getColSpan = (idx: number, total: number) => {
  if (total === 1) return 'col-span-1 justify-self-center'
  if (idx === 0 && total % 2 !== 0) return 'col-span-2'
  return ''
}
</script>
