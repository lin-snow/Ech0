<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-[88%] mx-auto mb-4">
    <div class="carousel-container rounded-lg overflow-hidden">
      <GalleryImageItem
        v-if="images[carouselIndex]"
        :image="images[carouselIndex]!"
        :src="resolvedSrcs[carouselIndex] || ''"
        :alt="getAlt(carouselIndex)"
        :loaded="isLoaded(images[carouselIndex]!, carouselIndex)"
        loading="eager"
        :priority="!!priority && carouselIndex === 0"
        button-class="carousel-slide w-full overflow-hidden"
        frame-class="w-full"
        img-class="w-full h-auto"
        :frame-style="
          getAspectRatioStyle(
            images[carouselIndex]!,
            isLoaded(images[carouselIndex]!, carouselIndex),
          )
        "
        @click="open(carouselIndex, $event)"
        @load="markLoaded(images[carouselIndex]!, carouselIndex)"
        @error="markLoaded(images[carouselIndex]!, carouselIndex)"
      />
    </div>

    <div
      v-if="images.length > 1"
      class="carousel-nav mt-3 flex items-center justify-center gap-3 text-[var(--color-text-muted)]"
    >
      <button
        class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
        type="button"
        :aria-label="t('imageGallery.prevImage')"
        @click.stop="prevCarousel"
        :disabled="carouselIndex === 0"
      >
        <Prev class="w-5 h-5 text-[var(--color-text-secondary)]" aria-hidden="true" />
      </button>
      <span class="text-sm"> {{ carouselIndex + 1 }} / {{ images.length }} </span>
      <button
        class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
        type="button"
        :aria-label="t('imageGallery.nextImage')"
        @click.stop="nextCarousel"
        :disabled="carouselIndex === images.length - 1"
      >
        <Next class="w-5 h-5 text-[var(--color-text-secondary)]" aria-hidden="true" />
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Prev from '@/components/icons/prev.vue'
import Next from '@/components/icons/next.vue'
import GalleryImageItem from '../parts/GalleryImageItem.vue'
import type { GalleryWithAspectRatioProps } from './types'

const props = defineProps<GalleryWithAspectRatioProps>()
const { t } = useI18n()

const carouselIndex = ref(0)
const imagesLength = computed(() => props.images.length)
const galleryIdentity = computed(() => {
  const first = props.images[0]
  const last = props.images[props.images.length - 1]
  const firstKey = first ? first.id || first.url || 'first' : 'empty'
  const lastKey = last ? last.id || last.url || 'last' : 'empty'
  return `${firstKey}|${lastKey}|${props.images.length}`
})

watch(
  imagesLength,
  (nextLength, prevLength) => {
    const safePrevLength = prevLength ?? nextLength

    if (nextLength <= 0) {
      carouselIndex.value = 0
      return
    }

    if (nextLength < safePrevLength && carouselIndex.value > nextLength - 1) {
      carouselIndex.value = nextLength - 1
    }
  },
  { immediate: true },
)

watch(galleryIdentity, (nextIdentity, prevIdentity) => {
  if (!prevIdentity || nextIdentity === prevIdentity) return
  carouselIndex.value = 0
})

const prevCarousel = () => {
  if (carouselIndex.value > 0) carouselIndex.value--
}

const nextCarousel = () => {
  if (carouselIndex.value < props.images.length - 1) carouselIndex.value++
}
</script>

<style scoped>
.carousel-container {
  position: relative;
  width: 100%;
}

.carousel-slide {
  position: relative;
  width: 100%;
  display: block;
}
</style>
