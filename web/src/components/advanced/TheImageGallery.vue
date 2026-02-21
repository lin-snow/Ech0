<template>
  <div v-if="images?.length" class="image-gallery-container">
    <!-- 瀑布流布局 -->
    <div
      v-if="layout === ImageLayout.WATERFALL || !layout"
      :class="[
        'imgwidth mx-auto grid gap-2 mb-4',
        images.length === 1 ? 'grid-cols-1 justify-items-center' : 'grid-cols-2',
      ]"
    >
      <button
        v-for="(src, idx) in images"
        :key="idx"
        class="bg-transparent border-0 p-0 cursor-pointer w-fit"
        :class="getColSpan(idx, images.length)"
        @click="openFancybox(idx)"
      >
        <img
          :src="baseUrl ? getHubImageUrl(src, baseUrl) : getImageUrl(src)"
          :alt="`预览图片${idx + 1}`"
          loading="lazy"
          class="echoimg block max-w-full h-auto"
        />
      </button>
    </div>

    <!-- 九宫格布局 -->
    <div v-if="layout === ImageLayout.GRID" class="imgwidth mx-auto mb-4">
      <div class="grid grid-cols-3 gap-2">
        <button
          v-for="(src, idx) in displayedImages"
          :key="idx"
          class="bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
          @click="openFancybox(idx)"
        >
          <img
            :src="baseUrl ? getHubImageUrl(src, baseUrl) : getImageUrl(src)"
            :alt="`预览图片${idx + 1}`"
            loading="lazy"
            class="echoimg w-full h-full object-cover"
          />

          <div v-if="extraCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
            +{{ extraCount }}
          </div>
        </button>
      </div>
    </div>

    <!-- 单图轮播布局 -->
    <div v-if="layout === ImageLayout.CAROUSEL" class="imgwidth mx-auto mb-4">
      <div class="carousel-container rounded-lg overflow-hidden">
        <button
          v-if="images[carouselIndex]"
          class="carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full overflow-hidden"
          @click="openFancybox(carouselIndex)"
        >
          <img
            :src="
              baseUrl
                ? getHubImageUrl(images[carouselIndex]!, baseUrl)
                : getImageUrl(images[carouselIndex]!)
            "
            :alt="`预览图片${carouselIndex + 1}`"
            loading="lazy"
            class="echoimg w-full h-auto"
          />
        </button>
      </div>

      <div
        v-if="images.length > 1"
        class="carousel-nav mt-3 flex items-center justify-center gap-3 text-[var(--text-color-500)]"
      >
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="prevCarousel"
          :disabled="carouselIndex === 0"
        >
          <Prev class="w-5 h-5 text-[var(--text-color-600)]" />
        </button>
        <span class="text-sm"> {{ carouselIndex + 1 }} / {{ images.length }} </span>
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="nextCarousel"
          :disabled="carouselIndex === images.length - 1"
        >
          <Next class="w-5 h-5 text-[var(--text-color-600)]" />
        </button>
      </div>
    </div>

    <!-- 水平轮播布局 -->
    <div v-if="layout === ImageLayout.HORIZONTAL" class="imgwidth mx-auto mb-4">
      <div class="horizontal-scroll-container">
        <div class="horizontal-scroll-wrapper">
          <button
            v-for="(src, idx) in images"
            :key="idx"
            class="horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0"
            @click="openFancybox(idx)"
          >
            <img
              :src="baseUrl ? getHubImageUrl(src, baseUrl) : getImageUrl(src)"
              :alt="`预览图片${idx + 1}`"
              loading="lazy"
              class="echoimg h-full w-auto object-contain"
            />
          </button>
        </div>
      </div>
      <div class="scroll-hint">← 左右滑动查看更多 →</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed, watch } from 'vue'
import { getImageUrl, getHubImageUrl } from '@/utils/other'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { ImageLayout } from '@/enums/enums'
import Prev from '@/components/icons/prev.vue'
import Next from '@/components/icons/next.vue'

const props = defineProps<{
  images: App.Api.Ech0.Echo['images']
  baseUrl?: string
  layout?: ImageLayout | string | undefined
}>()

const baseUrl = computed(() => props.baseUrl)

// 布局状态（来自 props.layout）
const layout = computed(() => props.layout || ImageLayout.WATERFALL)

// 轮播索引
const carouselIndex = ref(0)

watch(
  () => [props.images, props.layout, props.baseUrl],
  () => {
    carouselIndex.value = 0
  },
)

// 只显示前 9 张（用于九宫格），第 9 张显示 "+N" 覆盖层
const displayedImages = computed(() => (props.images ? props.images.slice(0, 9) : []))
const extraCount = computed(() =>
  props.images ? (props.images.length > 9 ? props.images.length - 9 : 0) : 0,
)

// 瀑布流布局：获取列跨度
const getColSpan = (idx: number, total: number) => {
  if (total === 1) return 'col-span-1 justify-self-center'
  if (idx === 0 && total % 2 !== 0) return 'col-span-2'
  return ''
}

// 轮播导航
const prevCarousel = () => {
  if (carouselIndex.value > 0) carouselIndex.value--
}
const nextCarousel = () => {
  if (carouselIndex.value < (props.images ? props.images.length - 1 : 0)) carouselIndex.value++
}

function openFancybox(startIndex: number) {
  const items = (props.images || []).map((src) => ({
    src: baseUrl.value ? getHubImageUrl(src, baseUrl.value) : getImageUrl(src),
    type: 'image' as const,
    thumb: baseUrl.value ? getHubImageUrl(src, baseUrl.value) : getImageUrl(src),
  }))

  Fancybox.show(items, {
    theme: 'auto',
    zoomEffect: true,
    fadeEffect: true,
    startIndex,
    backdropClick: 'close',
    dragToClose: true,
    keyboard: {
      Escape: 'close',
      ArrowRight: 'next',
      ArrowLeft: 'prev',
      Delete: 'close',
      Backspace: 'close',
      ArrowDown: 'next',
      ArrowUp: 'prev',
      PageUp: 'close',
      PageDown: 'close',
    },
  })
}

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})

onBeforeUnmount(() => {})
</script>

<style scoped>
.image-gallery-container {
  position: relative;
}

.imgwidth {
  width: 88%;
}

.echoimg {
  border-radius: 8px;
  box-shadow:
    0 1px 2px rgba(0, 0, 0, 0.02),
    0 2px 4px rgba(0, 0, 0, 0.02),
    0 4px 8px rgba(0, 0, 0, 0.02),
    0 8px 16px rgba(0, 0, 0, 0.02);
  transition:
    transform 0.3s ease,
    box-shadow 0.3s ease;
}

button:hover .echoimg {
  transform: scale(1.01);
  box-shadow:
    0 1px 3px rgba(0, 0, 0, 0.03),
    0 2px 6px rgba(0, 0, 0, 0.03),
    0 4px 12px rgba(0, 0, 0, 0.03);
}

/* carousel, horizontal, grid styles (copied/adapted from provided template) */
.carousel-container {
  position: relative;
  width: 100%;
}
.carousel-slide {
  position: relative;
  width: 100%;
  display: block;
}

.horizontal-scroll-container {
  position: relative;
  width: 100%;
  overflow-x: auto;
  overflow-y: hidden;
  scroll-behavior: smooth;
  -webkit-overflow-scrolling: touch;
  scrollbar-width: thin;
  scrollbar-color: rgba(0, 0, 0, 0.1) transparent;
}
.horizontal-scroll-container::-webkit-scrollbar {
  height: 4px;
}
.horizontal-scroll-wrapper {
  display: flex;
  gap: 8px;
  padding: 4px 0;
  align-items: center;
}
.horizontal-item {
  flex-shrink: 0;
  height: 200px;
  width: auto;
  overflow: hidden;
}
.scroll-hint {
  text-align: center;
  font-size: 12px;
  color: #999;
  margin-top: 8px;
  animation: hint-pulse 2s infinite;
}
@keyframes hint-pulse {
  0%,
  100% {
    opacity: 0.5;
  }
  50% {
    opacity: 1;
  }
}

.more-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.45);
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  border-radius: 8px;
}
</style>
