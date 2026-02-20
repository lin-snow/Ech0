<template>
  <div class="zone-root min-h-[calc(100vh-40px)] relative overflow-hidden px-4 py-6">
    <div class="absolute inset-0 pointer-events-none zone-bg"></div>
    <div
      class="zone-banner absolute top-5 left-1/2 -translate-x-1/2 pointer-events-none z-20 text-center"
    >
      <h1 class="zone-title text-3xl md:text-4xl tracking-[0.2em] uppercase">Ech0 Zone</h1>
      <p class="zone-subtitle text-[10px] md:text-xs tracking-[0.28em] uppercase mt-1">
        Thermal Print Console
      </p>
    </div>

    <div class="relative z-10 h-[calc(100vh-40px)] pb-[320px]" @mousedown="blurTopCard">
      <DraggablePaper
        v-for="card in cards"
        :key="card.id"
        :data="card"
        :z-index="card.zIndex"
        @update="updateCard"
        @delete="deleteCard"
        @focus="() => focusCard(card.id)"
      />
    </div>

    <div class="absolute left-1/2 -translate-x-1/2 bottom-2 z-30 pointer-events-auto">
      <TypewriterConsole
        ref="consoleRef"
        v-model="inputText"
        @print="handlePrint"
        @clear-all="clearAllCards"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { fetchGetEchoById } from '@/service/api'
import { useZoneStore } from '@/stores'
import { theToast } from '@/utils/toast'
import DraggablePaper from '../components/DraggablePaper.vue'
import TypewriterConsole from '../components/TypewriterConsole.vue'
import type { PaperCardData } from '../types'
import { getRandomStamp } from '../utils/stampUtils'

type PaperCardDataWithZ = PaperCardData & { zIndex: number }

const STORAGE_KEY = 'zone-paper-cards'

const route = useRoute()
const zoneStore = useZoneStore()
const cards = ref<PaperCardDataWithZ[]>([])
const inputText = ref('')
const topZIndex = ref(10)
const handledEchoId = ref('')
const consoleRef = ref<{ getPaperOrigin: () => { x: number; y: number } | null } | null>(null)

const serializedCards = computed(() =>
  cards.value.map(({ zIndex, ...rest }) => ({
    ...rest,
    zIndex,
  })),
)

const persistCards = () => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(serializedCards.value))
}

const restoreCards = () => {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return

  try {
    const parsed = JSON.parse(raw) as PaperCardDataWithZ[]
    if (!Array.isArray(parsed)) return
    cards.value = parsed
    const maxZ = parsed.reduce((max, item) => Math.max(max, item.zIndex ?? 10), 10)
    topZIndex.value = maxZ + 1
  } catch {
    localStorage.removeItem(STORAGE_KEY)
  }
}

const makeCardId = () => `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 7)}`

const getFallbackPosition = () => {
  const safeWidth = Math.max(window.innerWidth - 360, 40)
  const safeHeight = Math.max(window.innerHeight - 420, 80)
  return {
    x: Math.floor(Math.random() * safeWidth) + 20,
    y: Math.floor(Math.random() * safeHeight) + 80,
  }
}

const getPrintStartPosition = () => {
  const origin = consoleRef.value?.getPaperOrigin()
  if (!origin) {
    return getFallbackPosition()
  }

  const cardWidth = 280
  const estimatedCardHeight = 230
  const jitterX = Math.floor(Math.random() * 24) - 12
  const x = origin.x - cardWidth / 2 + jitterX
  const y = origin.y - estimatedCardHeight

  return {
    x: Math.max(10, Math.min(x, window.innerWidth - cardWidth - 10)),
    y: Math.max(10, Math.min(y, window.innerHeight - 280)),
  }
}

const handlePrint = (text: string, withStamp: boolean) => {
  const normalized = text.trim()
  if (!normalized) return

  const pos = getPrintStartPosition()
  const stampImage = withStamp ? getRandomStamp() : ''
  const card: PaperCardDataWithZ = {
    id: makeCardId(),
    text: normalized,
    x: pos.x,
    y: pos.y,
    rotation: Math.random() * 8 - 4,
    timestamp: Date.now(),
    isTyping: true,
    width: 280,
    height: 180,
    stampImage: stampImage || undefined,
    stampRotation: stampImage ? Math.random() * 18 - 9 : undefined,
    stampPosition: stampImage
      ? {
          x: 8 + Math.floor(Math.random() * 36),
          y: 4 + Math.floor(Math.random() * 22),
        }
      : undefined,
    zIndex: topZIndex.value++,
  }

  cards.value.push(card)
}

const clearAllCards = () => {
  cards.value = []
  localStorage.removeItem(STORAGE_KEY)
  theToast.info('已清空打印纸条')
}

const updateCard = (id: string, updates: Partial<PaperCardData>) => {
  const idx = cards.value.findIndex((item) => item.id === id)
  if (idx < 0) return
  const current = cards.value[idx]
  if (!current) return
  cards.value[idx] = {
    ...current,
    ...updates,
  }
}

const deleteCard = (id: string) => {
  cards.value = cards.value.filter((item) => item.id !== id)
}

const focusCard = (id: string) => {
  const idx = cards.value.findIndex((item) => item.id === id)
  if (idx < 0) return
  const current = cards.value[idx]
  if (!current) return
  cards.value[idx] = {
    ...current,
    zIndex: topZIndex.value++,
  }
}

const blurTopCard = () => {
  // 空实现用于保留舞台点击行为，避免误拖拽时被浏览器选中文本
}

const tryConsumePendingPrint = () => {
  const payload = zoneStore.consumePendingPrint()
  if (!payload) return
  if (!payload.text) return

  handlePrint(payload.text, false)
  const rawEchoId = route.params.echoId
  const echoId = Array.isArray(rawEchoId) ? rawEchoId[0] : rawEchoId
  if (echoId) {
    handledEchoId.value = echoId
  }
}

const tryPrintFromRouteEchoId = async () => {
  const rawEchoId = route.params.echoId
  const echoId = Array.isArray(rawEchoId) ? rawEchoId[0] : rawEchoId
  if (!echoId || handledEchoId.value === echoId) return

  const res = await fetchGetEchoById(String(echoId))
  if (res.code === 1 && res.data?.content?.trim()) {
    zoneStore.setPendingPrintEcho(res.data)
    tryConsumePendingPrint()
    handledEchoId.value = echoId
    return
  }

  theToast.info('该链接对应的 Echo 没有可打印文本')
}

watch(cards, persistCards, { deep: true })
watch(
  () => route.params.echoId,
  () => {
    void tryPrintFromRouteEchoId()
  },
)

onMounted(() => {
  restoreCards()
  tryConsumePendingPrint()
  void tryPrintFromRouteEchoId()
})
</script>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Special+Elite&family=Courier+Prime:wght@400;700&family=VT323&display=swap');

.zone-root {
  user-select: none;
}

.zone-bg {
  background:
    radial-gradient(circle at 50% 50%, rgba(235, 238, 243, 0.9), rgba(214, 219, 228, 0.95)),
    linear-gradient(rgba(255, 255, 255, 0.2) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.2) 1px, transparent 1px);
  background-size:
    auto,
    24px 24px,
    24px 24px;
}

.zone-title {
  color: rgba(30, 41, 59, 0.7);
  font-family: 'VT323', 'Courier Prime', 'Courier New', monospace;
  text-shadow:
    0 1px 0 rgba(255, 255, 255, 0.5),
    0 0 12px rgba(148, 163, 184, 0.35);
}

.zone-subtitle {
  color: rgba(71, 85, 105, 0.68);
  font-family: 'Courier Prime', 'Courier New', monospace;
}
</style>
