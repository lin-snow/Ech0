<template>
  <div
    ref="cardRef"
    :class="[
      'absolute cursor-move select-none',
      data.isTyping ? 'pointer-events-none' : 'pointer-events-auto',
    ]"
    :style="containerStyle"
    @mousedown="handleMouseDown"
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
            class="mt-3 pt-2 border-t border-dashed border-gray-400/25 text-[9px] leading-[1.35] text-gray-500 whitespace-pre-wrap break-words font-mono"
          >
            {{ printableMetadataText }}
            <span
              v-if="data.isTyping"
              class="inline-block w-2 h-3 bg-gray-500 ml-0.5 animate-pulse align-middle opacity-60"
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
const typingTimeout = ref<number | null>(null)
const cardRef = ref<HTMLDivElement | null>(null)

const progress = computed(() => {
  if (!props.data.text.length) return 0
  return displayedText.value.length / props.data.text.length
})

const typingTranslateY = computed(() => 100 - progress.value * 85)

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
  transform: `rotate(${props.data.rotation}deg) scale(${isDragging.value ? 1.05 : 1})`,
  transition: isDragging.value
    ? 'none'
    : 'transform 0.2s cubic-bezier(0.34, 1.56, 0.64, 1), filter 0.2s ease-out',
  filter: isDragging.value ? 'drop-shadow(0 10px 25px rgba(0,0,0,0.3))' : 'none',
}))

const animationWrapperStyle = computed(() => ({
  transform: props.data.isTyping ? `translateY(${typingTranslateY.value}%)` : undefined,
  animation: !props.data.isTyping
    ? 'ejecting 0.5s cubic-bezier(0.175, 0.885, 0.32, 1.275) forwards'
    : 'none',
  transition: props.data.isTyping ? 'transform 0.1s linear' : 'none',
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

const clearTypingTimeout = () => {
  if (typingTimeout.value !== null) {
    clearTimeout(typingTimeout.value)
    typingTimeout.value = null
  }
}

const scheduleTyping = () => {
  const { text } = props.data
  if (textIndex.value >= text.length) {
    emit('update', props.data.id, { isTyping: false })
    return
  }

  displayedText.value += text.charAt(textIndex.value)
  textIndex.value += 1

  const length = text.length
  let minDelay = 30
  let variance = 50

  if (length > 150) {
    minDelay = 5
    variance = 15
  } else if (length > 50) {
    minDelay = 15
    variance = 25
  }

  const delay = Math.random() * variance + minDelay
  typingTimeout.value = window.setTimeout(scheduleTyping, delay)
}

watch(
  () => [props.data.isTyping, props.data.text, props.data.id],
  ([isTyping]) => {
    clearTypingTimeout()
    if (isTyping) {
      displayedText.value = ''
      textIndex.value = 0
      typingTimeout.value = window.setTimeout(scheduleTyping, 100)
      return
    }
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

const onMouseMove = (e: MouseEvent) => {
  if (!isDragging.value) return
  emit('update', props.data.id, {
    x: e.clientX - dragOffset.value.x,
    y: e.clientY - dragOffset.value.y,
  })
}

const onMouseUp = () => {
  isDragging.value = false
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', onMouseUp)
}

const handleMouseDown = (e: MouseEvent) => {
  e.stopPropagation()
  emit('focus', true)
  isDragging.value = true
  dragOffset.value = {
    x: e.clientX - props.data.x,
    y: e.clientY - props.data.y,
  }
  window.addEventListener('mousemove', onMouseMove)
  window.addEventListener('mouseup', onMouseUp)
}

onBeforeUnmount(() => {
  clearTypingTimeout()
  window.removeEventListener('mousemove', onMouseMove)
  window.removeEventListener('mouseup', onMouseUp)
})
</script>

<style scoped>
@keyframes ejecting {
  0% {
    transform: translateY(15%);
  }
  50% {
    transform: translateY(-5%);
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
