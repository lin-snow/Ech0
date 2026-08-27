<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-[88%] mx-auto mb-4">
    <div class="grid grid-cols-3 gap-2">
      <GalleryImageItem
        v-for="(image, idx) in displayedImages"
        :key="getImageKey(image, idx)"
        :image="image"
        :src="resolvedSrcs[idx] || ''"
        :alt="getAlt(idx)"
        :loaded="isLoaded(image, idx)"
        :priority="!!priority && idx === 0"
        button-class="w-full overflow-hidden aspect-square relative"
        frame-class="h-full"
        img-class="w-full h-full object-cover"
        @click="open(idx, $event)"
        @load="markLoaded(image, idx)"
        @error="markLoaded(image, idx)"
      >
        <div v-if="extraCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
          +{{ extraCount }}
        </div>
      </GalleryImageItem>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryWithImageKeyProps } from './types'

const props = defineProps<GalleryWithImageKeyProps>()

const displayedImages = computed(() => props.images.slice(0, 9))
const extraCount = computed(() => (props.images.length > 9 ? props.images.length - 9 : 0))
</script>

<style scoped>
.more-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-overlay-strong);
  color: var(--color-gallery-overlay-text);
  font-size: 20px;
  font-weight: 600;
  border-radius: 8px;
}
</style>
