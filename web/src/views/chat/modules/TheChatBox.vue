<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div ref="rootEl" class="hairline-chat" :class="{ 'hairline-chat--empty': isEmpty }">
    <button class="ghost-ctrl ghost-ctrl--back" :title="t('commonNav.backHome')" @click="goHome">
      <Back class="ghost-ctrl__icon" />
    </button>
    <button
      v-if="messages.length > 0"
      class="ghost-ctrl ghost-ctrl--clear"
      :title="t('chatPanel.clear')"
      @click="handleClear"
    >
      <Close class="ghost-ctrl__icon" />
    </button>

    <div ref="scrollArea" class="transcript">
      <div
        ref="transcriptInner"
        class="transcript__inner"
        :style="{ '--tail-space': tailSpace + 'px' }"
      >
        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="turn"
          :class="msg.role === 'user' ? 'turn--user' : 'turn--ai'"
        >
          <p v-if="msg.role === 'user'" class="bubble">{{ msg.content }}</p>

          <template v-else>
            <ChatReasoning
              v-if="msg.reasoning !== undefined"
              :text="msg.reasoning"
              :active="msg.reasoningActive"
              :duration-ms="msg.reasoning_ms"
            />

            <div v-if="msg.searches && msg.searches.length > 0" class="searching">
              <span
                v-for="(query, qi) in msg.searches"
                :key="qi"
                class="searching__chip"
                :class="{
                  'searching__chip--live': isStreaming(idx) && qi === msg.searches.length - 1,
                }"
              >
                {{ t('chatPanel.searching', { query }) }}
              </span>
            </div>

            <div v-if="msg.coverage" class="searching">
              <span class="searching__chip">
                {{
                  msg.coverage.truncated
                    ? t('chatPanel.coverageTruncated', { returned: msg.coverage.returned })
                    : t('chatPanel.coverage', { total: msg.coverage.total })
                }}
              </span>
            </div>

            <div
              v-if="msg.content.length === 0 && isStreaming(idx) && !msg.reasoningActive"
              class="thinking"
              :aria-label="t('chatPanel.send')"
            >
              <span class="thinking__dot" />
              <span class="thinking__dot" />
              <span class="thinking__dot" />
            </div>
            <div v-else-if="msg.content.length > 0" class="answer">
              <AnimatedMarkdown
                v-if="showAnimated(idx)"
                :content="msg.content"
                :streaming="isStreaming(idx)"
                animation="blurIn"
                @update:revealing="assistantRevealing = $event"
              />
              <TheMdPreview v-else :content="msg.content" />
            </div>

            <div v-if="isRetryable(idx)" class="retry">
              <span v-if="msg.content.trim().length === 0" class="retry__hint">
                {{ t('chatPanel.noResponse') }}
              </span>
              <button class="retry__btn" :title="t('chatPanel.retry')" @click="retryLast">
                <svg
                  class="retry__icon"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M21 12a9 9 0 1 1-2.64-6.36" />
                  <path d="M21 3v6h-6" />
                </svg>
                <span>{{ t('chatPanel.retry') }}</span>
              </button>
            </div>
          </template>

          <ChatSources
            v-if="msg.sources && msg.sources.length > 0"
            :sources="msg.sources"
            @open="goToEcho"
          />
        </div>
      </div>
    </div>

    <nav v-if="questionNav.length > 0" class="qnav" :aria-label="t('chatPanel.navLabel')">
      <ul class="qnav__list">
        <li v-for="item in questionNav" :key="item.idx">
          <button
            class="qnav__item"
            :class="{ 'qnav__item--active': item.idx === activeQuestionIdx }"
            :title="item.content"
            :aria-current="item.idx === activeQuestionIdx ? 'true' : undefined"
            @click="scrollToQuestion(item.idx)"
          >
            <span class="qnav__label">{{ item.content }}</span>
            <span class="qnav__pill" />
          </button>
        </li>
      </ul>
    </nav>

    <div
      ref="composerEl"
      class="composer"
      :class="{
        'composer--active': input.trim().length > 0 || loading,
        'composer--loading': loading,
      }"
    >
      <textarea
        ref="inputEl"
        v-model="input"
        class="composer__field"
        rows="1"
        :placeholder="t('chatPanel.inputPlaceholder')"
        @input="autoGrow"
        @keydown="handleKeydown"
      />
      <Transition name="send-pop">
        <button
          v-if="loading"
          class="composer__action composer__action--stop"
          :title="t('chatPanel.clear')"
          @click="handleStop"
        >
          <span class="composer__stop-glyph" />
        </button>
        <button
          v-else-if="canSend"
          class="composer__action composer__action--send"
          :title="t('chatPanel.send')"
          @click="send(input)"
        >
          <Send class="composer__send-icon" />
        </button>
      </Transition>
    </div>

    <div class="understory">
      <Transition name="understory-fade">
        <div v-if="showSuggestions" class="understory__list">
          <p class="understory__hint">{{ t('chatPanel.suggestionsTitle') }}</p>
          <button
            v-for="(s, i) in suggestions"
            :key="i"
            class="understory__suggestion"
            @click="send(s)"
          >
            {{ s }}
          </button>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import Back from '@/components/icons/back.vue'
import Close from '@/components/icons/close.vue'
import Send from '@/components/icons/send.vue'
import { TheMdPreview } from '@/components/advanced/md'
import AnimatedMarkdown from './AnimatedMarkdown.vue'
import ChatSources from './ChatSources.vue'
import ChatReasoning from './ChatReasoning.vue'
import { ref, computed, nextTick, onBeforeUnmount, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'
import { chatStream } from '@/service/api'
import { getChatSession, clearChatSession } from '@/service/api/chat'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { theToast } from '@/utils/toast'

const { t } = useI18n()
const route = useRoute()
const router = useRouter()
const { openConfirm } = useBaseDialog()

const input = ref<string>('')
const loading = ref<boolean>(false)
const messages = ref<App.Api.Chat.ChatMessage[]>([])
const rootEl = ref<HTMLElement | null>(null)
const scrollArea = ref<HTMLElement | null>(null)
const transcriptInner = ref<HTMLElement | null>(null)
const composerEl = ref<HTMLElement | null>(null)
const inputEl = ref<HTMLTextAreaElement | null>(null)
let abort: (() => void) | null = null

const canSend = computed<boolean>(() => !loading.value && input.value.trim().length > 0)

const assistantRevealing = ref<boolean>(false)

const isStreaming = (idx: number): boolean => loading.value && idx === messages.value.length - 1

const showAnimated = (idx: number): boolean => {
  const last = messages.value.length - 1
  if (idx !== last || messages.value[idx]?.role !== 'assistant') return false
  return loading.value || assistantRevealing.value
}

const suggestions = computed<string[]>(() => [
  t('chatPanel.suggestion1'),
  t('chatPanel.suggestion2'),
  t('chatPanel.suggestion3'),
])

const showSuggestions = computed<boolean>(() => !loading.value && input.value.trim().length === 0)

const isEmpty = computed<boolean>(() => messages.value.length === 0)

const MAX_NAV_QUESTIONS = 7

const questionNav = computed<{ idx: number; content: string }[]>(() =>
  messages.value
    .map((m, idx) => ({ idx, content: m.content, role: m.role }))
    .filter((m) => m.role === 'user')
    .slice(-MAX_NAV_QUESTIONS)
    .map(({ idx, content }) => ({ idx, content })),
)

const activeQuestionIdx = ref<number>(-1)

const NAV_TOP_GUTTER = 84

const NAV_ACTIVE_TOLERANCE = 8

const turnElAt = (idx: number): HTMLElement | null =>
  (transcriptInner.value?.children[idx] as HTMLElement | undefined) ?? null

const updateActiveQuestion = () => {
  const area = scrollArea.value
  const nav = questionNav.value
  if (!area || nav.length === 0) {
    activeQuestionIdx.value = -1
    return
  }
  if (jumping) return
  if (pinned.value) {
    activeQuestionIdx.value = nav[nav.length - 1].idx
    return
  }
  const refLine = area.getBoundingClientRect().top + NAV_TOP_GUTTER + NAV_ACTIVE_TOLERANCE
  let active = nav[0].idx
  for (const item of nav) {
    const el = turnElAt(item.idx)
    if (!el) continue
    if (el.getBoundingClientRect().top <= refLine) active = item.idx
    else break
  }
  activeQuestionIdx.value = active
}

let navRaf = 0
const scheduleActiveUpdate = () => {
  if (navRaf) return
  navRaf = requestAnimationFrame(() => {
    navRaf = 0
    updateActiveQuestion()
  })
}

const scrollToQuestion = async (idx: number) => {
  const area = scrollArea.value
  const el = turnElAt(idx)
  if (!area || !el) return
  pinned.value = false
  activeQuestionIdx.value = idx
  jumping = true
  if (jumpTimer) clearTimeout(jumpTimer)
  const delta = el.getBoundingClientRect().top - area.getBoundingClientRect().top
  const target = area.scrollTop + delta - NAV_TOP_GUTTER
  const realMax = area.scrollHeight - tailSpace.value - area.clientHeight
  tailSpace.value = target > realMax ? area.clientHeight : 0
  await nextTick()
  jumpTimer = window.setTimeout(() => {
    jumping = false
    jumpTimer = 0
  }, 700)
  area.scrollTo({ top: target, behavior: 'smooth' })
}

const STICK_THRESHOLD = 80
const pinned = ref<boolean>(true)
let resizeObserver: ResizeObserver | null = null

const tailSpace = ref<number>(0)
let jumping = false
let jumpTimer = 0

const collapseTailIfSafe = () => {
  const el = scrollArea.value
  if (!el || tailSpace.value === 0) return
  const maxAfter = el.scrollHeight - tailSpace.value - el.clientHeight
  if (el.scrollTop <= maxAfter + 1) tailSpace.value = 0
}

const jumpToBottom = () => {
  nextTick(() => {
    const el = scrollArea.value
    if (el) el.scrollTop = el.scrollHeight - tailSpace.value - el.clientHeight
  })
}

const onScroll = () => {
  const el = scrollArea.value
  if (!el) return
  pinned.value = el.scrollHeight - el.scrollTop - el.clientHeight < STICK_THRESHOLD
  if (!jumping) {
    if (pinned.value) tailSpace.value = 0
    else collapseTailIfSafe()
  }
  scheduleActiveUpdate()
}

const onContentResize = () => {
  const el = scrollArea.value
  if (el && pinned.value) el.scrollTop = el.scrollHeight - tailSpace.value - el.clientHeight
  scheduleActiveUpdate()
}

const syncComposerHeight = () => {
  const root = rootEl.value
  const c = composerEl.value
  if (root && c) root.style.setProperty('--composer-h', `${c.offsetHeight}px`)
}

const autoGrow = () => {
  const el = inputEl.value
  if (!el) return
  el.style.height = 'auto'
  el.style.height = `${el.scrollHeight}px`
  syncComposerHeight()
  onContentResize()
}

const goHome = () => router.push('/')
const goToEcho = (echoId: string) => router.push(`/echo/${echoId}`)

const streamInto = (question: string, assistant: App.Api.Chat.ChatMessage) => {
  loading.value = true
  assistantRevealing.value = true
  pinned.value = true
  tailSpace.value = 0
  jumping = false
  if (jumpTimer) {
    clearTimeout(jumpTimer)
    jumpTimer = 0
  }
  nextTick(autoGrow)
  jumpToBottom()

  abort = chatStream(question, {
    onSearching: (query) => {
      if (query && !assistant.searches?.includes(query)) {
        assistant.searches?.push(query)
      }
    },
    onSources: (sources) => {
      const merged = assistant.sources ? [...assistant.sources] : []
      const seen = new Set(merged.map((s) => s.echo_id))
      for (const src of sources) {
        if (!seen.has(src.echo_id)) {
          seen.add(src.echo_id)
          merged.push(src)
        }
      }
      assistant.sources = merged
    },
    onCoverage: (coverage) => {
      assistant.coverage = coverage
    },
    onReasoning: (text) => {
      if (assistant.reasoning === undefined) {
        assistant.reasoning = ''
        assistant.reasoningActive = true
      }
      assistant.reasoning += text
    },
    onReasoningDone: (durationMs) => {
      assistant.reasoning_ms = durationMs
      assistant.reasoningActive = false
    },
    onDelta: (text) => {
      assistant.content += text
    },
    onError: (message) => {
      loading.value = false
      assistant.failed = true
      theToast.error(message || String(t('chatPanel.errorGeneric')))
    },
    onDone: () => {
      loading.value = false
    },
  })
}

const send = (question: string) => {
  const q = question.trim()
  if (q.length === 0 || loading.value) return

  messages.value.push({ role: 'user', content: q })
  messages.value.push({ role: 'assistant', content: '', sources: [], searches: [] })
  input.value = ''
  streamInto(q, messages.value[messages.value.length - 1])
}

const isRetryable = (idx: number): boolean => {
  const m = messages.value[idx]
  if (!m || m.role !== 'assistant') return false
  if (isStreaming(idx) || idx !== messages.value.length - 1) return false
  if (m.failed === true) return true
  const noText = m.content.trim().length === 0
  const noSources = !m.sources || m.sources.length === 0
  return noText && noSources
}

const retryLast = () => {
  if (loading.value) return
  const n = messages.value.length
  if (n < 2) return
  const assistant = messages.value[n - 1]
  const user = messages.value[n - 2]
  if (assistant.role !== 'assistant' || user.role !== 'user') return

  assistant.content = ''
  assistant.sources = []
  assistant.searches = []
  assistant.coverage = undefined
  assistant.failed = false
  assistant.reasoning = undefined
  assistant.reasoning_ms = undefined
  assistant.reasoningActive = false
  streamInto(user.content, assistant)
}

const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    send(input.value)
  }
}

const handleStop = () => {
  if (abort) abort()
  abort = null
  loading.value = false
}

watch(loading, (now, prev) => {
  if (!prev || now) return
  const last = messages.value[messages.value.length - 1]
  if (last?.role === 'assistant' && last.reasoningActive) last.reasoningActive = false
  if (!jumping) collapseTailIfSafe()
})

const handleClear = () => {
  openConfirm({
    title: t('chatPanel.clearConfirmTitle'),
    description: t('chatPanel.clearConfirmDesc'),
    onConfirm: async () => {
      if (abort) abort()
      abort = null
      loading.value = false
      assistantRevealing.value = false
      pinned.value = true
      tailSpace.value = 0
      jumping = false
      if (jumpTimer) {
        clearTimeout(jumpTimer)
        jumpTimer = 0
      }
      try {
        await clearChatSession()
      } catch {}
      messages.value = []
      theToast.success(String(t('chatPanel.clearSuccess')))
    },
  })
}

onMounted(async () => {
  const area = scrollArea.value
  if (area) area.addEventListener('scroll', onScroll, { passive: true })
  if (typeof ResizeObserver === 'function') {
    resizeObserver = new ResizeObserver((entries) => {
      let composerChanged = false
      for (const entry of entries) {
        if (entry.target === composerEl.value) composerChanged = true
      }
      if (composerChanged) syncComposerHeight()
      onContentResize()
    })
    if (transcriptInner.value) resizeObserver.observe(transcriptInner.value)
    if (composerEl.value) resizeObserver.observe(composerEl.value)
  }
  syncComposerHeight()

  try {
    const res = await getChatSession()
    const history = res.data
    if (Array.isArray(history) && history.length > 0) {
      messages.value = history
      pinned.value = true
      jumpToBottom()
    }
  } catch {}
  scheduleActiveUpdate()

  const initialQuery = route.query.q
  const q = Array.isArray(initialQuery) ? initialQuery[0] : initialQuery
  if (typeof q === 'string' && q.trim().length > 0) {
    void router.replace({ query: {} })
    send(q)
  }
})

onBeforeUnmount(() => {
  resizeObserver?.disconnect()
  resizeObserver = null
  scrollArea.value?.removeEventListener('scroll', onScroll)
  if (navRaf) cancelAnimationFrame(navRaf)
  if (jumpTimer) clearTimeout(jumpTimer)
})
</script>

<style scoped>
.hairline-chat {
  --composer-max: min(8.5rem, 30dvh);

  --composer-h: 1.6rem;

  --line-pos: 25dvh;

  position: relative;
  width: 100%;
  height: 100vh;
  height: 100dvh;
  overflow: hidden;
  background: var(--color-bg-canvas);
  color: var(--color-text-primary);
}

.hairline-chat--empty {
  --line-pos: 52dvh;
}

.ghost-ctrl {
  position: absolute;
  top: 1.25rem;
  z-index: 3;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 2.25rem;
  height: 2.25rem;
  border: none;
  border-radius: 999px;

  background: color-mix(in srgb, var(--color-bg-canvas) 92%, transparent);
  backdrop-filter: blur(8px);
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0.95;
  transition:
    opacity 0.2s ease,
    background 0.2s ease,
    color 0.2s ease;
}

.ghost-ctrl--back {
  left: 1.25rem;
}

.ghost-ctrl--clear {
  right: 1.25rem;
}

.ghost-ctrl:hover {
  opacity: 1;
  background: color-mix(in srgb, var(--color-bg-canvas) 98%, transparent);
  color: var(--color-text-primary);
}

.ghost-ctrl__icon {
  width: 1.2rem;
  height: 1.2rem;
}

.ghost-ctrl__icon :deep(path) {
  fill: currentColor;
}

.transcript {
  position: absolute;

  inset: 0 0 calc(var(--line-pos) + var(--composer-h, 1.6rem) + 1rem);

  display: flex;
  justify-content: center;
  overflow-y: auto;

  overflow-anchor: none;

  mask-image: linear-gradient(
    to bottom,
    transparent 0,
    #000 4.5rem,
    #000 calc(100% - 1rem),
    transparent 100%
  );
}

.transcript__inner {
  width: 100%;
  max-width: 42rem;

  align-self: flex-start;

  padding: 4.5rem 1.5rem 1.5rem;

  padding-bottom: calc(1.5rem + var(--tail-space, 0px));
  display: flex;
  flex-direction: column;
  gap: 1.6rem;
}

.turn {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  animation: turn-in 0.32s ease both;

  contain: layout style;
}

.turn--user {
  align-items: flex-end;
}

.turn--ai {
  align-items: flex-start;
}

@keyframes turn-in {
  from {
    opacity: 0;
    transform: translateY(6px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

.bubble {
  max-width: 85%;
  padding: 0.55rem 0.9rem;
  border-radius: 1.1rem 1.1rem 0.25rem;
  background: var(--color-accent-soft);
  color: var(--color-text-primary);
  font-size: 0.9rem;
  line-height: 1.6;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
}

.answer {
  width: 100%;
  font-size: 1rem;
}

.answer :deep(.echo-markdown) {
  line-height: 1.8;
}

.retry {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem 0.75rem;
  margin-top: 0.15rem;
}

.retry__hint {
  font-size: 0.82rem;
  line-height: 1.5;
  color: var(--color-text-muted);
}

.retry__btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.2rem 0.55rem 0.2rem 0.4rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-secondary);
  font-size: 0.82rem;
  line-height: 1.5;
  cursor: pointer;
  transition:
    color 0.18s ease,
    background 0.18s ease;
}

.retry__btn:hover {
  background: var(--color-accent-soft);
  color: var(--color-accent);
}

.retry__icon {
  width: 0.95rem;
  height: 0.95rem;
  flex-shrink: 0;
  transition: transform 0.4s cubic-bezier(0.22, 1, 0.36, 1);
}

.retry__btn:hover .retry__icon {
  transform: rotate(180deg);
}

@media (prefers-reduced-motion: reduce) {
  .retry__icon {
    transition: none;
  }
}

.thinking {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0;
}

.thinking__dot {
  width: 0.4rem;
  height: 0.4rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.5;
  animation: thinking-bounce 1.2s ease-in-out infinite;
}

.thinking__dot:nth-child(2) {
  animation-delay: 0.16s;
}

.thinking__dot:nth-child(3) {
  animation-delay: 0.32s;
}

@keyframes thinking-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.35;
  }

  40% {
    transform: translateY(-0.28rem);
    opacity: 1;
  }
}

.searching {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
  margin-bottom: 0.5rem;
}

.searching__chip {
  display: inline-flex;
  align-items: center;
  max-width: 22rem;
  padding: 0.1rem 0.5rem;
  border-radius: 999px;
  background: var(--color-accent-soft);
  color: var(--color-text-secondary);
  font-size: 0.72rem;
  line-height: 1.5;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.searching__chip::before {
  content: '🔍';
  margin-right: 0.3rem;
  font-size: 0.7rem;
  font-variant-emoji: text;
}

.searching__chip--live {
  animation: searching-pulse 1.4s ease-in-out infinite;
}

@keyframes searching-pulse {
  0%,
  100% {
    opacity: 0.55;
  }

  50% {
    opacity: 1;
  }
}

.composer {
  position: absolute;
  left: 50%;
  bottom: var(--line-pos);
  transform: translateX(-50%);
  z-index: 2;
  width: min(42rem, calc(100% - 3rem));
  display: flex;
  align-items: flex-end;
  gap: 0.6rem;

  transition: bottom 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.composer::after {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  height: 1px;
  background: linear-gradient(
    to right,
    transparent,
    var(--color-border-strong) 18%,
    var(--color-border-strong) 82%,
    transparent
  );
}

.composer::before {
  content: '';
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 1;
  height: 1px;
  background: var(--color-accent);
  mask-image: linear-gradient(to right, transparent, #000 15%, #000 85%, transparent);
  transform: scaleX(0);
  transform-origin: 50% 50%;
  transition: transform 0.35s cubic-bezier(0.22, 1, 0.36, 1);
}

.composer--active::before,
.composer:focus-within::before {
  transform: scaleX(1);
}

.composer--loading::before {
  background-image: linear-gradient(
    90deg,
    var(--color-accent) 0%,
    color-mix(in srgb, var(--color-accent) 35%, #fff) 50%,
    var(--color-accent) 100%
  );
  background-size: 200% 100%;
  animation: composer-shimmer 1.6s linear infinite;
}

@keyframes composer-shimmer {
  from {
    background-position: 150% 0;
  }

  to {
    background-position: -150% 0;
  }
}

@media (prefers-reduced-motion: reduce) {
  .composer,
  .understory {
    transition: none;
  }

  .composer::before {
    transition: none;
  }

  .composer--loading::before {
    animation: none;
  }
}

.composer__field {
  flex: 1;
  resize: none;
  border: none;
  outline: none;
  background: transparent;
  padding: 0 0 0.7rem;
  max-height: var(--composer-max);
  font-size: 1rem;
  line-height: 1.6;
  color: var(--color-text-primary);
  overflow-y: auto;
}

.composer__field::placeholder {
  color: var(--color-text-muted);
  opacity: 0.7;
}

.composer__action {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  margin-bottom: 0.5rem;
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.composer__action--send {
  color: var(--color-accent);
}

.composer__action:hover {
  transform: translateY(-1px);
  filter: brightness(1.1);
}

.composer__action:active {
  transform: translateY(0);
}

.composer__send-icon {
  width: 1.15rem;
  height: 1.15rem;
}

.composer__stop-glyph {
  width: 0.6rem;
  height: 0.6rem;
  border-radius: 0.14rem;
  background: currentColor;
}

.send-pop-enter-active,
.send-pop-leave-active {
  transition:
    transform 0.18s cubic-bezier(0.34, 1.56, 0.64, 1),
    opacity 0.18s ease;
}

.send-pop-enter-from,
.send-pop-leave-to {
  transform: scale(0.5);
  opacity: 0;
}

.understory {
  position: absolute;
  left: 50%;

  top: calc(100dvh - var(--line-pos));
  transform: translateX(-50%);
  width: min(42rem, calc(100% - 3rem));
  padding-top: 1.5rem;

  transition: top 0.5s cubic-bezier(0.22, 1, 0.36, 1);
}

.understory__list {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 0.35rem;
}

.understory-fade-enter-active,
.understory-fade-leave-active {
  transition:
    opacity 0.25s ease,
    transform 0.25s ease;
}

.understory-fade-enter-from,
.understory-fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}

.understory__hint {
  font-size: 0.72rem;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--color-text-muted);
  opacity: 0.7;
  margin-bottom: 0.25rem;
}

.understory__suggestion {
  border: none;
  background: transparent;
  padding: 0.15rem 0;
  font-size: 0.9rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  text-align: left;
  cursor: pointer;
  transition:
    color 0.18s ease,
    transform 0.18s ease;
}

.understory__suggestion:hover {
  color: var(--color-accent);
  transform: translateX(3px);
}

.qnav {
  position: absolute;
  top: 50%;
  right: 0;
  transform: translateY(-50%);
  z-index: 3;
  display: flex;
  align-items: center;
  max-height: 72dvh;

  padding-right: 0.4rem;
}

.qnav__list {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.3rem;
  max-height: 72dvh;
  margin: 0;
  padding: 0;
  list-style: none;

  overflow-y: auto;
  scrollbar-width: none;
}

.qnav__list::-webkit-scrollbar {
  display: none;
}

.qnav__item {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  width: 100%;
  padding: 0.22rem 0.3rem;
  border: none;
  border-radius: 999px;
  background: transparent;
  cursor: pointer;
  transition: background 0.22s ease;
}

.qnav__label {
  max-width: 0;
  margin-right: 0;
  overflow: hidden;
  font-size: 0.78rem;
  line-height: 1.5;
  color: var(--color-text-secondary);
  white-space: nowrap;
  text-overflow: ellipsis;
  opacity: 0;
  transition:
    max-width 0.28s cubic-bezier(0.22, 1, 0.36, 1),
    margin 0.28s cubic-bezier(0.22, 1, 0.36, 1),
    opacity 0.2s ease;
}

.qnav__pill {
  flex-shrink: 0;
  width: 1.1rem;
  height: 0.26rem;
  border-radius: 999px;
  background: var(--color-text-muted);
  opacity: 0.6;
  transition:
    width 0.22s ease,
    opacity 0.22s ease,
    background 0.22s ease;
}

.qnav__item--active .qnav__pill {
  width: 1.7rem;
  background: var(--color-accent);
  opacity: 1;
}

.qnav__item:hover .qnav__pill {
  opacity: 0.85;
}

.qnav:hover .qnav__item {
  background: color-mix(in srgb, var(--color-bg-canvas) 94%, transparent);
  backdrop-filter: blur(10px);
}

.qnav:hover .qnav__label {
  max-width: min(38vw, 15rem);
  margin-right: 0.5rem;
  opacity: 1;
}

.qnav:hover .qnav__item--active .qnav__label {
  color: var(--color-text-primary);
  font-weight: 500;
}

@media (hover: none), (width <= 768px) {
  .qnav {
    display: none;
  }
}

@media (prefers-reduced-motion: reduce) {
  .qnav__item,
  .qnav__label,
  .qnav__pill {
    transition: none;
  }
}

@media (width <= 640px) {
  .transcript__inner {
    padding-left: 1.25rem;
    padding-right: 1.25rem;
  }
}
</style>
