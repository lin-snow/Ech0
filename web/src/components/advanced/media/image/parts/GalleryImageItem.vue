<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <button
    class="bg-transparent border-0 p-0 cursor-pointer"
    :class="buttonClass"
    type="button"
    @click="handleClick"
  >
    <div class="gallery-image-frame" :class="frameClass" :style="frameStyle">
      <div v-if="!loaded" class="image-skeleton" aria-hidden="true"></div>
      <img
        :src="src"
        :alt="alt"
        :width="image.width || undefined"
        :height="image.height || undefined"
        :loading="effectiveLoading"
        :fetchpriority="priority ? 'high' : undefined"
        decoding="async"
        class="echoimg transition-opacity duration-300"
        :class="[imgClass, loaded ? 'opacity-100' : 'opacity-0']"
        @load="$emit('load')"
        @error="$emit('error')"
      />
      <slot></slot>
    </div>
  </button>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const emit = defineEmits<{
  (e: 'click', sourceElement: HTMLElement | null): void
  (e: 'load'): void
  (e: 'error'): void
}>()

const props = withDefaults(
  defineProps<{
    image: App.Api.Ech0.FileObject
    src: string
    alt: string
    loaded: boolean
    loading?: 'lazy' | 'eager'
    priority?: boolean
    buttonClass?: string
    frameClass?: string
    imgClass?: string
    frameStyle?: Record<string, string>
  }>(),
  {
    loading: 'lazy',
    priority: false,
    buttonClass: 'w-fit',
    frameClass: '',
    imgClass: 'block max-w-full h-auto',
    frameStyle: undefined,
  },
)

const effectiveLoading = computed(() => (props.priority ? 'eager' : props.loading))

const handleClick = (event: MouseEvent) => {
  const sourceElement = event.currentTarget
  emit('click', sourceElement instanceof HTMLElement ? sourceElement : null)
}
</script>

<style scoped>
.echoimg {
  border-radius: 8px;
  box-shadow:
    0 1px 2px rgb(0 0 0 / 2%),
    0 2px 4px rgb(0 0 0 / 2%),
    0 4px 8px rgb(0 0 0 / 2%),
    0 8px 16px rgb(0 0 0 / 2%);
}

.gallery-image-frame {
  position: relative;
  border-radius: 8px;
  overflow: hidden;
}

.image-skeleton {
  position: absolute;
  inset: 0;
  border-radius: 8px;
  pointer-events: none;
  background:
    radial-gradient(
      120% 120% at 50% 30%,
      rgb(160 160 160 / 18%) 0%,
      rgb(160 160 160 / 10%) 55%,
      rgb(160 160 160 / 7%) 100%
    ),
    linear-gradient(180deg, rgb(140 140 140 / 8%), rgb(140 140 140 / 11%));
  animation: skeleton-breathe 2.2s ease-in-out infinite;
}

@keyframes skeleton-breathe {
  0%,
  100% {
    opacity: 0.42;
  }

  50% {
    opacity: 0.68;
  }
}

@media (prefers-reduced-motion: reduce) {
  .image-skeleton {
    animation: none;
    opacity: 0.55;
  }
}
</style>
