<template>
  <div
    class="relative pointer-events-auto scale-90 md:scale-100 origin-bottom transition-transform duration-300"
  >
    <div
      class="relative w-[380px] md:w-[440px] bg-gradient-to-b from-lime-500 to-lime-600 rounded-[3rem] p-6 shadow-2xl border-t border-white/30"
      style="
        box-shadow:
          0 50px 60px -20px rgba(0, 0, 0, 0.6),
          inset 0 2px 10px rgba(255, 255, 255, 0.4),
          inset 0 -10px 20px rgba(0, 0, 0, 0.2);
      "
    >
      <div
        class="absolute inset-0 rounded-[3rem] bg-[url('https://www.transparenttextures.com/patterns/noise-lines.png')] opacity-20 pointer-events-none mix-blend-overlay"
      ></div>
      <div
        ref="paperSlotRef"
        class="absolute -top-3 left-1/2 -translate-x-1/2 w-64 h-5 bg-gray-900 rounded-full shadow-inner border-b border-gray-700 z-0"
      ></div>
      <div
        class="absolute top-2 left-10 right-10 h-1 bg-gradient-to-r from-transparent via-white/60 to-transparent rounded-full blur-[1px]"
      ></div>

      <div
        class="relative bg-[#8ec93e] rounded-3xl p-4 shadow-[inset_0_4px_8px_rgba(0,0,0,0.3),0_2px_4px_rgba(255,255,255,0.2)] border border-lime-700/20"
      >
        <div class="flex justify-between items-end mb-3 px-2">
          <div class="flex flex-col">
            <div class="flex items-center gap-1">
              <div
                class="w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse shadow-[0_0_5px_red]"
              ></div>
              <span
                class="text-[10px] font-black tracking-widest text-lime-900 uppercase font-['VT323']"
              >
                Auto-Feed
              </span>
            </div>
            <div class="text-lime-800/60 text-[9px] font-bold tracking-[0.2em] uppercase mt-0.5">
              Series 9000
            </div>
          </div>
          <div class="flex items-center gap-1 opacity-60 text-lime-900">
            <span class="text-xs font-['VT323']">5G</span>
          </div>
        </div>

        <div
          class="bg-[#0d160d] rounded-xl p-1 pb-0 shadow-[inset_0_0_20px_rgba(0,0,0,1)] border-b-2 border-white/10 relative overflow-hidden"
        >
          <div
            class="absolute inset-0 bg-[linear-gradient(rgba(18,16,16,0)_50%,rgba(0,0,0,0.25)_50%),linear-gradient(90deg,rgba(255,0,0,0.06),rgba(0,255,0,0.02),rgba(0,0,255,0.06))] z-10 bg-[length:100%_2px,3px_100%] pointer-events-none"
          ></div>

          <div class="relative z-20 p-2">
            <div
              class="flex justify-between text-[#1a5c1a] text-xs mb-1 font-['VT323'] border-b border-[#1a5c1a]/30 pb-1"
            >
              <span>COMPOSE_MODE</span>
              <span>{{ modelLength }} chars</span>
            </div>

            <textarea
              :value="modelValue"
              class="w-full h-20 bg-transparent resize-none outline-none font-['VT323'] text-xl text-[#4aff4a] placeholder-[#1a5c1a] tracking-wider leading-tight"
              placeholder="TYPE MESSAGE HERE..."
              spellcheck="false"
              style="text-shadow: 0 0 5px rgba(74, 255, 74, 0.5)"
              @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
            />
          </div>

          <div
            v-if="deleteConfirm"
            class="absolute inset-0 bg-[#0d160d]/95 z-30 flex flex-col items-center justify-center"
          >
            <span class="text-red-500 font-['VT323'] text-lg animate-pulse"
              >CLICK AGAIN TO CLEAR ALL PRINTS...</span
            >
            <div class="text-red-400 font-['VT323'] text-3xl mt-2 font-bold">{{ countdown }}</div>
          </div>
        </div>

        <div class="mt-5 grid grid-cols-5 gap-3 items-center">
          <button
            class="col-span-1 aspect-square rounded-full bg-zinc-800 shadow-[0_4px_0_#000,0_5px_10px_rgba(0,0,0,0.5)] active:shadow-none active:translate-y-1 transition-all border-t border-zinc-600 flex items-center justify-center group relative overflow-hidden"
            :title="stampEnabled ? 'Stamp Enabled' : 'Enable Stamp'"
            @click="stampEnabled = !stampEnabled"
          >
            <div
              class="absolute inset-0 bg-gradient-to-tr from-transparent to-white/10 rounded-full"
            ></div>
            <span
              :class="[
                'text-base transition-all duration-200',
                stampEnabled
                  ? 'text-yellow-400 drop-shadow-[0_0_8px_rgba(250,204,21,0.8)]'
                  : 'text-gray-300',
              ]"
            >
              印
            </span>
          </button>

          <button
            class="col-span-1 aspect-square rounded-full bg-zinc-800 shadow-[0_4px_0_#000,0_5px_10px_rgba(0,0,0,0.5)] active:shadow-none active:translate-y-1 transition-all border-t border-zinc-600 flex items-center justify-center group relative overflow-hidden"
            :title="deleteConfirm ? 'Click again to confirm' : 'Clear all prints'"
            @click="handleDeleteClick"
          >
            <div
              class="absolute inset-0 bg-gradient-to-tr from-transparent to-white/10 rounded-full"
            ></div>
            <span
              :class="[
                'text-base transition-all duration-200',
                deleteConfirm
                  ? 'text-red-500 drop-shadow-[0_0_8px_rgba(239,68,68,0.8)]'
                  : 'text-gray-300',
              ]"
            >
              删
            </span>
          </button>

          <div class="col-span-1 flex flex-col items-center gap-1 px-2">
            <div class="w-full h-1 bg-lime-900/20 rounded-full"></div>
            <div class="w-full h-1 bg-lime-900/20 rounded-full"></div>
            <div class="w-full h-1 bg-lime-900/20 rounded-full"></div>
            <div class="w-full h-1 bg-lime-900/20 rounded-full"></div>
          </div>

          <button
            class="col-span-2 h-14 bg-orange-600 rounded-lg shadow-[0_5px_0_#9a3412,0_8px_15px_rgba(0,0,0,0.4)] active:shadow-none active:translate-y-[5px] transition-all border-t border-orange-400 flex items-center justify-center gap-2 relative overflow-hidden disabled:opacity-50 disabled:cursor-not-allowed"
            :disabled="!canPrint"
            @click="handlePrint"
          >
            <div
              class="absolute top-0 left-0 w-full h-1/2 bg-gradient-to-b from-white/20 to-transparent"
            ></div>
            <span class="font-['VT323'] text-2xl text-orange-950 font-bold drop-shadow-sm mt-1"
              >PRINT</span
            >
          </button>
        </div>
      </div>

      <div
        class="absolute bottom-3 left-1/2 -translate-x-1/2 bg-black/80 px-3 py-0.5 rounded text-[8px] text-lime-500 font-mono tracking-widest border border-lime-500/30 shadow-sm"
      >
        ECH0
      </div>
    </div>

    <div
      class="absolute -bottom-4 left-10 right-10 h-8 bg-lime-500/30 blur-xl rounded-full z-[-1]"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  print: [text: string, withStamp: boolean]
  'clear-all': []
}>()

const stampEnabled = ref(false)
const deleteConfirm = ref(false)
const countdown = ref(5)
const timer = ref<number | null>(null)
const paperSlotRef = ref<HTMLDivElement | null>(null)

const modelLength = computed(() => props.modelValue.length)
const canPrint = computed(() => props.modelValue.trim().length > 0)

const clearTimer = () => {
  if (timer.value !== null) {
    clearTimeout(timer.value)
    timer.value = null
  }
}

watch(
  () => deleteConfirm.value,
  (active) => {
    clearTimer()
    if (!active) return

    const tick = () => {
      if (!deleteConfirm.value) return
      if (countdown.value <= 0) {
        deleteConfirm.value = false
        countdown.value = 5
        return
      }
      countdown.value -= 1
      timer.value = window.setTimeout(tick, 1000)
    }

    timer.value = window.setTimeout(tick, 1000)
  },
)

const handlePrint = () => {
  const text = props.modelValue.trim()
  if (!text) return
  emit('print', text, stampEnabled.value)
  emit('update:modelValue', '')
  stampEnabled.value = false
}

const handleDeleteClick = () => {
  if (deleteConfirm.value) {
    emit('clear-all')
    deleteConfirm.value = false
    countdown.value = 5
    return
  }

  deleteConfirm.value = true
  countdown.value = 5
}

onBeforeUnmount(() => {
  clearTimer()
})

const getPaperOrigin = () => {
  const slot = paperSlotRef.value
  if (!slot) return null
  const rect = slot.getBoundingClientRect()
  return {
    x: rect.left + rect.width / 2,
    y: rect.top + Math.min(12, rect.height / 2),
  }
}

defineExpose({
  getPaperOrigin,
})
</script>
