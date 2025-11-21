<template>
  <div v-if="mediaItems.length" class="image-gallery-container">
    <!-- 瀑布流布局 -->
    <div
      v-if="layout === ImageLayout.WATERFALL || !layout"
      :class="[
        'imgwidth mx-auto grid gap-2 mb-4',
        mediaItems.length === 1 ? 'grid-cols-1 justify-items-center' : 'grid-cols-2',
      ]"
    >
      <template v-for="(item, idx) in mediaItems" :key="idx">
        <!-- 图片 -->
        <button
          v-if="!isVideo(item)"
          class="bg-transparent border-0 p-0 cursor-pointer w-fit"
          :class="getColSpan(idx, mediaItems.length)"
          @click="openFancybox(idx)"
        >
          <img
            :src="getMediaUrlCompat(item)"
            :alt="`预览图片${idx + 1}`"
            loading="lazy"
            class="echoimg block max-w-full h-auto"
          />
        </button>
        
        <!-- 视频 -->
        <button
          v-else
          class="video-preview bg-transparent border-0 p-0 cursor-pointer w-fit relative"
          :class="getColSpan(idx, mediaItems.length)"
          @click="openFancybox(idx)"
        >
          <video
            :src="getMediaUrlCompat(item)"
            preload="metadata"
            class="echoimg video-thumb block max-w-full h-auto"
          ></video>
          <div class="play-overlay">
            <Play class="play-icon" color="#ffffff" />
          </div>
        </button>
      </template>
    </div>

    <!-- 九宫格布局 -->
    <div v-if="layout === ImageLayout.GRID" class="imgwidth mx-auto mb-4">
      <div class="grid grid-cols-3 gap-2">
        <template v-for="(item, idx) in displayedImages" :key="idx">
          <!-- 图片 -->
          <button
            v-if="!isVideo(item)"
            class="bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
            @click="openFancybox(idx)"
          >
            <img
              :src="getMediaUrlCompat(item)"
              :alt="`预览图片${idx + 1}`"
              loading="lazy"
              class="echoimg w-full h-full object-cover"
            />

            <div v-if="extraCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraCount }}
            </div>
          </button>
          
          <!-- 视频 -->
          <button
            v-else
            class="video-preview bg-transparent border-0 p-0 cursor-pointer overflow-hidden aspect-square relative"
            @click="openFancybox(idx)"
          >
            <video
              :src="getMediaUrlCompat(item)"
              preload="metadata"
              class="echoimg video-thumb w-full h-full object-cover"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
            
            <div v-if="extraCount > 0 && idx === 8" class="more-overlay" aria-hidden="true">
              +{{ extraCount }}
            </div>
          </button>
        </template>
      </div>
    </div>

    <!-- 单图轮播布局 -->
    <div v-if="layout === ImageLayout.CAROUSEL" class="imgwidth mx-auto mb-4">
      <div class="carousel-container rounded-lg overflow-hidden">
        <template v-if="mediaItems[carouselIndex]">
          <!-- 图片 -->
          <button
            v-if="!isVideo(mediaItems[carouselIndex]!)"
            class="carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full overflow-hidden"
            @click="openFancybox(carouselIndex)"
          >
            <img
              :src="getMediaUrlCompat(mediaItems[carouselIndex]!)"
              :alt="`预览图片${carouselIndex + 1}`"
              loading="lazy"
              class="echoimg w-full h-auto"
            />
          </button>
          
          <!-- 视频 -->
          <button
            v-else
            class="video-preview carousel-slide bg-transparent border-0 p-0 cursor-pointer w-full relative"
            @click="openFancybox(carouselIndex)"
          >
            <video
              :src="getMediaUrlCompat(mediaItems[carouselIndex]!)"
              preload="metadata"
              class="echoimg video-thumb w-full h-auto"
            ></video>
            <div class="play-overlay">
              <Play class="play-icon" color="#ffffff" />
            </div>
          </button>
        </template>
      </div>

      <div
        v-if="mediaItems.length > 1"
        class="carousel-nav mt-3 flex items-center justify-center gap-3 text-gray-500"
      >
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="prevCarousel"
          :disabled="carouselIndex === 0"
        >
          <Prev class="w-5 h-5 text-gray-600" />
        </button>
        <span class="text-sm"> {{ carouselIndex + 1 }} / {{ mediaItems.length }} </span>
        <button
          class="nav-btn flex items-center justify-center w-8 h-8 rounded-full transition disabled:opacity-40 disabled:cursor-not-allowed"
          @click="nextCarousel"
          :disabled="carouselIndex === mediaItems.length - 1"
        >
          <Next class="w-5 h-5 text-gray-600" />
        </button>
      </div>
    </div>

    <!-- 水平轮播布局 -->
    <div v-if="layout === ImageLayout.HORIZONTAL" class="imgwidth mx-auto mb-4">
      <div class="horizontal-scroll-container">
        <div class="horizontal-scroll-wrapper">
          <template v-for="(item, idx) in mediaItems" :key="idx">
            <!-- 图片 -->
            <button
              v-if="!isVideo(item)"
              class="horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0"
              @click="openFancybox(idx)"
            >
              <img
                :src="getMediaUrlCompat(item)"
                :alt="`预览图片${idx + 1}`"
                loading="lazy"
                class="echoimg h-full w-auto object-contain"
              />
            </button>
            
            <!-- 视频 -->
            <button
              v-else
              class="video-preview horizontal-item bg-transparent rounded-lg border-0 p-0 cursor-pointer shrink-0 relative"
              @click="openFancybox(idx)"
            >
              <video
                :src="getMediaUrlCompat(item)"
                preload="metadata"
                class="echoimg video-thumb h-full w-auto object-contain"
              ></video>
              <div class="play-overlay">
                <Play class="play-icon" color="#ffffff" />
              </div>
            </button>
          </template>
        </div>
      </div>
      <div class="scroll-hint">← 左右滑动查看更多 →</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref, computed } from 'vue'
import { getMediaUrl, getHubMediaUrl, getImageUrl, getHubImageUrl } from '@/utils/other'
import { Fancybox } from '@fancyapps/ui'
import '@fancyapps/ui/dist/fancybox/fancybox.css'
import { ImageLayout } from '@/enums/enums'
import Prev from '@/components/icons/prev.vue'
import Next from '@/components/icons/next.vue'
import Play from '@/components/icons/play.vue'

const props = defineProps<{
  media?: App.Api.Ech0.Media[]
  images?: App.Api.Ech0.Image[]  // 向后兼容
  baseUrl?: string
  layout?: ImageLayout | string | undefined
}>()

// 使用 media 或 images（向后兼容）
const mediaItems = computed(() => props.media || props.images || [])

const baseUrl = props.baseUrl

// 布局状态（来自 props.layout）
const layout = props.layout || ImageLayout.WATERFALL

// 辅助函数：获取媒体URL（兼容新旧格式）
const getMediaUrlCompat = (item: any) => {
  // 如果有 media_url 字段，说明是新格式（Media）
  if ('media_url' in item) {
    return baseUrl ? getHubMediaUrl(item, baseUrl) : getMediaUrl(item)
  }
  // 否则是旧格式（Image）
  return baseUrl ? getHubImageUrl(item, baseUrl) : getImageUrl(item)
}

// 检查是否为视频
const isVideo = (item: any) => {
  return item.media_type === 'video'
}

// 轮播索引
const carouselIndex = ref(0)

// 只显示前 9 张（用于九宫格），第 9 张显示 "+N" 覆盖层
const displayedImages = computed(() => mediaItems.value.slice(0, 9))
const extraCount = computed(() =>
  mediaItems.value.length > 9 ? mediaItems.value.length - 9 : 0
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
  if (carouselIndex.value < mediaItems.value.length - 1) carouselIndex.value++
}

// 创建视频HTML内容（用于Fancybox）
function createVideoHTML(src: string): string {
  return `
    <div class="fancybox-video-container">
      <video 
        class="fancybox-video-player"
        playsinline 
        controls
        preload="metadata"
        controlsList="nodownload"
      >
        <source src="${src}" type="video/mp4" />
        您的浏览器不支持视频播放。
      </video>
      <div class="video-error-message" style="display: none;">
        <div class="error-content">
          <svg class="error-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <p class="error-title">视频加载失败</p>
          <p class="error-description">无法加载视频内容，请检查网络连接或稍后重试</p>
          <div class="error-actions">
            <button class="retry-button">重试</button>
            <button class="skip-button">跳过</button>
          </div>
        </div>
      </div>
    </div>
  `
}

// 初始化视频错误处理（在Fancybox中）
function initVideoErrorHandling(slide: any): void {
  try {
    const videoElement = slide.$el?.querySelector('.fancybox-video-player') as HTMLVideoElement
    const errorMessage = slide.$el?.querySelector('.video-error-message')
    
    if (!videoElement) {
      console.error('Video element not found')
      return
    }
    
    // 添加视频错误处理
    const handleVideoError = (event: Event) => {
      console.error('Video loading error:', event)
      
      // 显示错误消息
      if (errorMessage) {
        videoElement.style.display = 'none'
        ;(errorMessage as HTMLElement).style.display = 'flex'
        
        // 设置重试按钮
        const retryButton = errorMessage.querySelector('.retry-button')
        const skipButton = errorMessage.querySelector('.skip-button')
        
        if (retryButton) {
          retryButton.addEventListener('click', () => {
            // 重新加载视频
            videoElement.load()
            ;(errorMessage as HTMLElement).style.display = 'none'
            videoElement.style.display = 'block'
          }, { once: true })
        }
        
        if (skipButton) {
          skipButton.addEventListener('click', () => {
            // 跳到下一个
            const fancybox = Fancybox.getInstance()
            if (fancybox) {
              const carousel = fancybox.getCarousel()
              if (carousel) {
                carousel.next()
              }
            }
          }, { once: true })
        }
      }
    }
    
    videoElement.addEventListener('error', handleVideoError)
    
    // 存储清理函数
    slide.cleanup = () => {
      videoElement.removeEventListener('error', handleVideoError)
      videoElement.pause()
      videoElement.src = ''
      videoElement.load()
    }
    
  } catch (error) {
    console.error('Error initializing video error handling:', error)
  }
}

// 清理视频资源
function cleanupVideo(slide: any): void {
  if (slide?.cleanup) {
    try {
      slide.cleanup()
      slide.cleanup = null
    } catch (error) {
      console.error('Failed to cleanup video:', error)
    }
  }
}

function openFancybox(startIndex: number) {
  // 处理所有媒体类型（图片和视频）
  const items = mediaItems.value.map((item) => {
    const mediaUrl = getMediaUrlCompat(item)
    
    if (isVideo(item)) {
      // 为视频创建HTML类型的Fancybox项（不提供缩略图）
      return {
        src: mediaUrl,
        type: 'html' as const,
        html: createVideoHTML(mediaUrl),
      }
    } else {
      // 为图片创建image类型的Fancybox项
      return {
        src: mediaUrl,
        type: 'image' as const,
        thumb: mediaUrl,
      }
    }
  })

  Fancybox.show(items, {
    theme: 'auto',
    zoomEffect: true,
    fadeEffect: true,
    startIndex: startIndex,
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
    Carousel: {
      Thumbs: {
        type: 'classic',
        showOnStart: true,
      },
    },
    on: {
      // 当幻灯片附加到DOM时初始化视频错误处理
      'Carousel.attachSlideEl': (fancybox: any, carousel: any, slide: any) => {
        if (slide.type === 'html') {
          setTimeout(() => {
            initVideoErrorHandling(slide)
          }, 100)
        }
      },
      // 当Fancybox关闭时清理资源
      destroy: (fancybox: any) => {
        const carousel = fancybox.getCarousel()
        if (carousel) {
          carousel.getSlides().forEach((slide: any) => {
            cleanupVideo(slide)
          })
        }
      },
      // 当切换幻灯片时，清理前一个幻灯片的视频
      'Carousel.change': (fancybox: any, carousel: any, to: number, from?: number) => {
        if (from !== undefined) {
          const slides = carousel.getSlides()
          if (slides[from]) {
            cleanupVideo(slides[from])
          }
        }
      }
    },
  })
}

onMounted(() => {
  Fancybox.bind('[data-fancybox]', {})
})

onBeforeUnmount(() => {
  // 清理所有可能残留的Fancybox实例
  const fancybox = Fancybox.getInstance()
  if (fancybox) {
    const carousel = fancybox.getCarousel()
    if (carousel) {
      carousel.getSlides().forEach((slide: any) => {
        cleanupVideo(slide)
      })
    }
    fancybox.close()
  }
})
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

/* 图片按钮悬停效果 */
button:has(img.echoimg):hover img.echoimg {
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

/* 视频预览容器样式 - 与图片保持一致 */
.video-preview {
  position: relative;
  display: block;
  overflow: hidden;
  cursor: pointer;
}

/* 视频预览悬停效果 - 与图片一致 */
.video-preview:hover .video-thumb {
  transform: translateY(-2px);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
}

.video-preview:hover .play-overlay {
  background: rgba(0, 0, 0, 0.65);
  transform: translate(-50%, -50%) scale(1.1);
}

.video-preview:active .video-thumb {
  transform: translateY(0);
}

.video-thumb {
  display: block;
  pointer-events: none;
  width: 100%;
  height: auto;
}

/* 播放图标覆盖层 */
.play-overlay {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  border-radius: 50%;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  pointer-events: none;
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.play-overlay::before {
  content: '';
  position: absolute;
  inset: -2px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.2);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.video-preview:hover .play-overlay::before {
  opacity: 1;
  animation: pulse-ring 1.5s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}

@keyframes pulse-ring {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.1);
    opacity: 0.5;
  }
}

.play-icon {
  width: 32px;
  height: 32px;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
  transition: transform 0.3s ease;
}

.video-preview:hover .play-icon {
  transform: scale(1.15);
}

/* 确保九宫格中视频和图片大小一致 */
.grid .video-preview {
  width: 100%;
  height: 100%;
}

.grid .video-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

/* 水平轮播中的视频和图片保持一致 */
.horizontal-item {
  position: relative;
}

.horizontal-item .video-thumb,
.horizontal-item img {
  height: 100%;
  width: auto;
  object-fit: contain;
  display: block;
}

button:hover .echoimg {
  transform: scale(1.02);
  box-shadow:
    0 2px 4px rgba(0, 0, 0, 0.04),
    0 4px 8px rgba(0, 0, 0, 0.04),
    0 8px 16px rgba(0, 0, 0, 0.04),
    0 16px 32px rgba(0, 0, 0, 0.04);
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
  border-radius: 8px;
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

/* Fancybox视频容器样式 */
:deep(.fancybox-video-container) {
  width: 100%;
  max-width: 1200px;
  margin: 0 auto;
  aspect-ratio: 16 / 9;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  position: relative;
}

/* 视频错误消息样式 */
:deep(.video-error-message) {
  position: absolute;
  inset: 0;
  display: none;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.95);
  z-index: 100;
  padding: 20px;
}

:deep(.error-content) {
  text-align: center;
  color: #fff;
  max-width: 400px;
}

:deep(.error-icon) {
  margin: 0 auto 16px;
  color: #ef4444;
  opacity: 0.9;
}

:deep(.error-title) {
  font-size: 20px;
  font-weight: 600;
  margin: 0 0 8px;
  color: #fff;
}

:deep(.error-description) {
  font-size: 14px;
  margin: 0 0 24px;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.5;
}

:deep(.error-actions) {
  display: flex;
  gap: 12px;
  justify-content: center;
}

:deep(.retry-button),
:deep(.skip-button) {
  padding: 10px 24px;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  border: none;
  outline: none;
}

:deep(.retry-button) {
  background: #3b82f6;
  color: #fff;
}

:deep(.retry-button:hover) {
  background: #2563eb;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

:deep(.retry-button:active) {
  transform: translateY(0);
}

:deep(.skip-button) {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.2);
}

:deep(.skip-button:hover) {
  background: rgba(255, 255, 255, 0.15);
  border-color: rgba(255, 255, 255, 0.3);
}

:deep(.skip-button:active) {
  transform: scale(0.98);
}

:deep(.fancybox-video-player) {
  width: 100%;
  height: 100%;
  max-height: 80vh;
  background: #000;
  object-fit: contain;
}

/* 移动端优化 */
@media (max-width: 768px) {
  :deep(.fancybox-video-container) {
    max-width: 100%;
    aspect-ratio: auto;
    height: auto;
    border-radius: 0;
  }
  
  :deep(.fancybox-video-player) {
    max-height: 60vh;
  }
}

/* 平板优化 */
@media (min-width: 769px) and (max-width: 1024px) {
  :deep(.fancybox-video-container) {
    max-width: 90%;
  }
}

/* 超小屏幕优化 */
@media (max-width: 480px) {
  :deep(.fancybox-video-container) {
    aspect-ratio: auto;
  }
  
  :deep(.fancybox-video-player) {
    max-height: 50vh;
  }
  
  .play-overlay {
    width: 56px;
    height: 56px;
  }
  
  .play-icon {
    width: 28px;
    height: 28px;
  }
}
</style>
