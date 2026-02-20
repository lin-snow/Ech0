<template>
  <div
    ref="cardRef"
    :class="[
      'absolute cursor-move select-none',
      data.isTyping ? 'pointer-events-none' : 'pointer-events-auto',
    ]"
    :style="containerStyle"
    @pointerdown="handlePointerDown"
  >
    <div :style="animationWrapperStyle">
      <div class="relative shadow-md" :style="paperStyle">
        <div class="absolute -top-1.5 left-0 w-full h-3 serrated-top"></div>

        <div class="px-5 py-6 relative overflow-hidden">
          <div
            v-if="!data.isTyping && data.stampImage && data.stampPosition"
            class="absolute z-10"
            :style="stampStyle"
          >
            <img
              :src="data.stampImage"
              alt="stamp"
              draggable="false"
              class="w-[80px] h-[80px] object-contain opacity-60 pointer-events-none"
            />
          </div>

          <button
            v-if="!data.isTyping"
            class="absolute top-2 right-2 text-gray-400 hover:text-red-600 transition-colors z-10 mix-blend-multiply"
            @mousedown.stop
            @pointerdown.stop
            @click.stop="emit('delete', data.id)"
          >
            <svg viewBox="0 0 24 24" fill="none" class="w-4 h-4" aria-hidden="true">
              <path
                d="M18 6L6 18M6 6l12 12"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
              />
            </svg>
          </button>

          <div
            class="flex flex-col items-center border-b border-dashed border-gray-400/50 pb-3 mb-4 opacity-70 font-mono text-[10px] text-gray-600"
          >
            <span class="uppercase tracking-widest font-bold text-[11px] text-gray-800">
              Echo Print
            </span>
            <div class="flex justify-between w-full mt-1.5 px-1 text-[9px] tracking-tight">
              <span>ID: {{ data.id.slice(0, 6).toUpperCase() }}</span>
              <span class="font-semibold">{{ dateStr }} {{ timeStr }}</span>
            </div>
          </div>

          <div
            class="font-serif text-lg leading-relaxed text-gray-900 break-words whitespace-pre-wrap min-h-[2.5rem]"
            style="
              text-shadow: 0 0 1px rgba(0, 0, 0, 0.1);
              font-family:
                'Special Elite', 'Noto Serif SC', 'Source Han Serif SC', 'Songti SC', 'STSong',
                'Times New Roman', serif;
            "
          >
            {{ printableMainText }}
            <span
              v-if="data.isTyping && !hasMetadataBlock"
              class="inline-block w-2.5 h-4 bg-gray-800 ml-0.5 animate-pulse align-middle opacity-80"
            ></span>
          </div>

          <div
            v-if="hasMetadataBlock"
            class="mt-3 pt-2 border-t border-dashed border-gray-300/20 text-[8px] leading-[1.3] text-gray-400 whitespace-pre-wrap break-words font-mono"
          >
            {{ printableMetadataText }}
            <span
              v-if="data.isTyping"
              class="inline-block w-2 h-3 bg-gray-400 ml-0.5 animate-pulse align-middle opacity-45"
            ></span>
          </div>

          <div
            class="mt-5 pt-3 border-t border-gray-400/30 flex justify-between items-end opacity-50 text-gray-700"
          >
            <div class="h-3 w-20 bg-current opacity-30 barcode-mask"></div>
            <span class="text-[8px] font-mono tracking-wide">END OF TRANSMISSION</span>
          </div>
        </div>

        <div class="absolute -bottom-1.5 left-0 w-full h-3 serrated-bottom"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import type { Coordinates, PaperCardData } from '../types'

const PAPER_COLOR = '#ffffff'

const props = defineProps<{
  data: PaperCardData
  zIndex: number
}>()

const emit = defineEmits<{
  update: [id: string, updates: Partial<PaperCardData>]
  delete: [id: string]
  focus: [force?: boolean]
}>()

const isDragging = ref(false)
const dragOffset = ref<Coordinates>({ x: 0, y: 0 })
const displayedText = ref('')
const textIndex = ref(0)
const typingRaf = ref<number | null>(null)
const typingProgress = ref(0)
const typingStartAt = ref(0)
const typingDuration = ref(1)
const cardRef = ref<HTMLDivElement | null>(null)

const HIDDEN_OFFSET_PX = 220
const HOLD_OFFSET_PX = 32
const typingTranslateY = computed(
  () => HIDDEN_OFFSET_PX - typingProgress.value * (HIDDEN_OFFSET_PX - HOLD_OFFSET_PX),
)

const dateObj = computed(() => new Date(props.data.timestamp))
const dateStr = computed(() =>
  dateObj.value.toLocaleDateString('en-US', { month: '2-digit', day: '2-digit', year: '2-digit' }),
)
const timeStr = computed(() =>
  dateObj.value.toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit' }),
)

const containerStyle = computed(() => ({
  left: `${props.data.x}px`,
  top: `${props.data.y}px`,
  zIndex: props.zIndex,
  width: '280px',
  touchAction: 'none',
  transform: `rotate(${props.data.rotation}deg) scale(${isDragging.value ? 1.05 : 1})`,
  transition: isDragging.value
    ? 'none'
    : 'transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), filter 0.2s ease-out',
  filter: isDragging.value ? 'drop-shadow(0 10px 25px rgba(0,0,0,0.3))' : 'none',
}))

const animationWrapperStyle = computed(() => ({
  transform: props.data.isTyping ? `translateY(${typingTranslateY.value}px)` : undefined,
  animation: !props.data.isTyping
    ? 'ejecting 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards'
    : 'none',
  transition: props.data.isTyping ? 'none' : 'none',
  transformOrigin: 'bottom center',
}))

const paperStyle = {
  backgroundColor: PAPER_COLOR,
  filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))',
}

const stampStyle = computed(() => ({
  right: `${props.data.stampPosition?.x ?? 0}px`,
  bottom: `${props.data.stampPosition?.y ?? 0}px`,
  transform: `rotate(${props.data.stampRotation ?? 0}deg)`,
}))

const METADATA_FLAG = '[METADATA]'

const printableMainText = computed(() => {
  const raw = displayedText.value || ''
  const metadataIndex = raw.indexOf(METADATA_FLAG)
  if (metadataIndex < 0) return raw

  const mainPart = raw.slice(0, metadataIndex)
  return mainPart.replace(/\s*---\s*$/u, '').trimEnd()
})

const printableMetadataText = computed(() => {
  const raw = displayedText.value || ''
  const metadataIndex = raw.indexOf(METADATA_FLAG)
  if (metadataIndex < 0) return ''
  return raw.slice(metadataIndex).trimStart()
})

const hasMetadataBlock = computed(() => printableMetadataText.value.length > 0)

const clearTypingRaf = () => {
  if (typingRaf.value !== null) {
    cancelAnimationFrame(typingRaf.value)
    typingRaf.value = null
  }
}

const getTypingDuration = (length: number) => {
  if (length > 150) return Math.max(700, length * 10)
  if (length > 50) return Math.max(900, length * 22)
  return Math.max(1100, length * 45)
}

const startTyping = () => {
  const text = props.data.text
  if (!text.length) {
    typingProgress.value = 1
    emit('update', props.data.id, { isTyping: false })
    return
  }

  displayedText.value = ''
  textIndex.value = 0
  typingProgress.value = 0
  typingStartAt.value = performance.now()
  typingDuration.value = getTypingDuration(text.length)

  const tick = (now: number) => {
    const elapsed = now - typingStartAt.value
    const progress = Math.min(elapsed / typingDuration.value, 1)
    typingProgress.value = progress

    const nextIndex = Math.min(text.length, Math.floor(progress * text.length))
    if (nextIndex !== textIndex.value) {
      textIndex.value = nextIndex
      displayedText.value = text.slice(0, nextIndex)
    }

    if (progress >= 1 && nextIndex >= text.length) {
      emit('update', props.data.id, { isTyping: false })
      typingRaf.value = null
      return
    }

    typingRaf.value = requestAnimationFrame(tick)
  }

  typingRaf.value = requestAnimationFrame(tick)
}

watch(
  () => [props.data.isTyping, props.data.text, props.data.id],
  ([isTyping]) => {
    clearTypingRaf()
    if (isTyping) {
      startTyping()
      return
    }
    typingProgress.value = 1
    displayedText.value = props.data.text
  },
  { immediate: true },
)

watch(
  () => [displayedText.value, props.data.isTyping],
  () => {
    if (!cardRef.value || props.data.isTyping) return
    const rect = cardRef.value.getBoundingClientRect()
    if (props.data.width !== rect.width || props.data.height !== rect.height) {
      emit('update', props.data.id, { width: rect.width, height: rect.height })
    }
  },
)

const onPointerMove = (e: PointerEvent) => {
  if (!isDragging.value) return
  emit('update', props.data.id, {
    x: e.clientX - dragOffset.value.x,
    y: e.clientY - dragOffset.value.y,
  })
}

const onPointerUp = () => {
  isDragging.value = false
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerUp)
}

const handlePointerDown = (e: PointerEvent) => {
  if (e.button !== 0) return
  e.stopPropagation()
  emit('focus', true)
  isDragging.value = true
  dragOffset.value = {
    x: e.clientX - props.data.x,
    y: e.clientY - props.data.y,
  }
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
  window.addEventListener('pointercancel', onPointerUp)
}

onBeforeUnmount(() => {
  clearTypingRaf()
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  window.removeEventListener('pointercancel', onPointerUp)
})
</script>

<style scoped>
@keyframes ejecting {
  0% {
    transform: translateY(32px);
  }
  50% {
    transform: translateY(-8px);
  }
  100% {
    transform: translateY(0);
  }
}

.serrated-top {
  background-color: #fff;
  mask-image: radial-gradient(circle at 5px 0, transparent 5px, black 5.5px);
  mask-size: 10px 10px;
  mask-repeat: repeat-x;
  mask-position: bottom;
  -webkit-mask-image: radial-gradient(circle at 5px 0, transparent 5px, black 5.5px);
  -webkit-mask-size: 10px 10px;
  -webkit-mask-repeat: repeat-x;
  -webkit-mask-position: bottom;
}

.serrated-bottom {
  background-color: #fff;
  mask-image: radial-gradient(circle at 5px 10px, transparent 5px, black 5.5px);
  mask-size: 10px 10px;
  mask-repeat: repeat-x;
  mask-position: top;
  -webkit-mask-image: radial-gradient(circle at 5px 10px, transparent 5px, black 5.5px);
  -webkit-mask-size: 10px 10px;
  -webkit-mask-repeat: repeat-x;
  -webkit-mask-position: top;
}

.barcode-mask {
  mask-image: linear-gradient(90deg, black 50%, transparent 50%);
  mask-size: 3px 100%;
}
</style>
