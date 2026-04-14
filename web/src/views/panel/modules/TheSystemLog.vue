<template>
  <div class="w-full px-2">
    <div class="mb-3 log-toolbar">
      <div class="log-toolbar-grid">
        <label class="toolbar-field">
          <span class="field-label">{{ t('systemLog.level') }}</span>
          <select
            v-model="level"
            class="h-9 rounded-[var(--radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--color-text-secondary)] border border-[var(--color-border-subtle)]"
          >
            <option value="all">{{ t('systemLog.allLevels') }}</option>
            <option value="debug">debug</option>
            <option value="info">info</option>
            <option value="warn">warn</option>
            <option value="error">error</option>
          </select>
        </label>

        <label class="toolbar-field field-grow">
          <span class="field-label">{{ t('systemLog.keyword') }}</span>
          <input
            v-model="keyword"
            :placeholder="t('systemLog.keywordPlaceholder')"
            class="h-9 rounded-[var(--radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--color-text-secondary)] border border-[var(--color-border-subtle)]"
          />
        </label>

        <label class="toolbar-field field-tail">
          <span class="field-label">{{ t('systemLog.tail') }}</span>
          <input
            v-model.number="tail"
            type="number"
            min="50"
            max="1000"
            step="50"
            v-tooltip="t('systemLog.tailTitle')"
            @blur="normalizeTail"
            class="h-9 rounded-[var(--radius-md)] px-2 bg-[var(--input-bg-color)] text-[var(--color-text-secondary)] border border-[var(--color-border-subtle)]"
          />
          <span class="field-hint">{{ t('systemLog.tailHint') }}</span>
        </label>
      </div>

      <div class="log-toolbar-actions">
        <button class="h-9 px-3 rounded-md log-btn" @click="reload">
          {{ t('systemLog.applyFilter') }}
        </button>
        <button class="h-9 px-3 rounded-md log-btn" @click="clearLogs">
          {{ t('systemLog.clear') }}
        </button>
        <label class="inline-flex items-center gap-1 text-sm text-[var(--color-text-secondary)]">
          <input v-model="autoScroll" type="checkbox" />
          {{ t('systemLog.autoScroll') }}
        </label>
        <span class="text-xs text-[var(--color-text-muted)]"
          >{{ t('systemLog.connection') }}: {{ connectionText }}</span
        >
      </div>
    </div>

    <div ref="logContainer" class="log-container text-xs md:text-sm">
      <template v-if="logs.length > 0">
        <div v-for="(line, idx) in logs" :key="`${line.time}-${idx}`" class="log-line">
          <div class="log-line-main">
            <span class="time">[{{ line.time || '-' }}]</span>
            <span class="level">[{{ line.level || 'info' }}]</span>
            <span class="msg">{{ line.msg }}</span>
          </div>
          <div v-if="line.error" class="log-line-detail">{{ line.error }}</div>
          <details v-if="line.raw" class="log-line-raw">
            <summary>{{ t('systemLog.rawLine') }}</summary>
            <pre class="log-line-raw-pre">{{ line.raw }}</pre>
          </details>
        </div>
      </template>
      <div v-else class="text-[var(--color-text-muted)]">{{ t('systemLog.empty') }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { fetchSystemLogs } from '@/service/api'
import { useOWebSocket } from '@/service/request/websocket'
import { getApiUrl, getWsUrl } from '@/service/request/shared'
import { useAuthStore } from '@/stores/auth'
import { theToast } from '@/utils/toast'

const logs = ref<App.Api.SystemLog.Entry[]>([])
const tail = ref(200)
const level = ref('all')
const keyword = ref('')
const autoScroll = ref(true)
const logContainer = ref<HTMLElement>()
const transport = ref<'ws' | 'sse'>('ws')
let es: EventSource | null = null
const { t } = useI18n()
const authStore = useAuthStore()

const { onMessage, open, close, status } = useOWebSocket<App.Api.Response<App.Api.SystemLog.Entry>>(
  {
    url: getWsUrl('/ws/system/logs'),
    autoReconnect: {
      retries: 5,
      delay: 1000,
      onFailed: () => {
        startSSE()
      },
    },
    // 与后端协议对齐，避免 heartbeat ping/pong 不一致导致反复断连
    heartbeat: false,
  },
)

const connectionText = computed(() => {
  if (transport.value === 'sse') return String(t('systemLog.sseFallback'))
  return String(status.value)
})

const normalizeTail = () => {
  const value = Number(tail.value)
  if (!Number.isFinite(value)) {
    tail.value = 200
    return
  }
  tail.value = Math.min(1000, Math.max(50, Math.round(value)))
}

const normalizeLevel = (value: string) => value.trim().toLowerCase()

const hitFilter = (entry: App.Api.SystemLog.Entry) => {
  const currentLevel = normalizeLevel(level.value)
  const currentKeyword = keyword.value.trim().toLowerCase()
  if (currentLevel !== 'all' && normalizeLevel(entry.level || '') !== currentLevel) {
    return false
  }
  if (!currentKeyword) {
    return true
  }
  const haystack = `${entry.msg || ''} ${entry.raw || ''} ${entry.error || ''}`.toLowerCase()
  return haystack.includes(currentKeyword)
}

const scrollToBottom = async () => {
  if (!autoScroll.value) return
  await nextTick()
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

const pushLog = async (entry: App.Api.SystemLog.Entry) => {
  if (!hitFilter(entry)) return
  logs.value.push(entry)
  if (logs.value.length > 1000) {
    logs.value = logs.value.slice(logs.value.length - 1000)
  }
  await scrollToBottom()
}

const loadHistory = async () => {
  normalizeTail()
  const res = await fetchSystemLogs({
    tail: tail.value,
    level: level.value,
    keyword: keyword.value,
  })
  if (res.code !== 1) {
    return
  }
  logs.value = Array.isArray(res.data) ? res.data : []
  await scrollToBottom()
}

const reload = async () => {
  await loadHistory()
  restartStream()
}

const clearLogs = () => {
  logs.value = []
}

const buildSSEUrl = () => {
  const token = authStore.accessToken
  const query = new URLSearchParams()
  query.set('token', token)
  if (level.value !== 'all') {
    query.set('level', level.value)
  }
  const trimmedKeyword = keyword.value.trim()
  if (trimmedKeyword) {
    query.set('keyword', trimmedKeyword)
  }
  const base = getApiUrl().replace(/\/+$/, '')
  return `${base}/system/logs/stream?${query.toString()}`
}

const stopSSE = () => {
  if (es) {
    es.close()
    es = null
  }
}

const startSSE = () => {
  if (transport.value === 'sse') return
  stopSSE()
  close()
  transport.value = 'sse'
  es = new EventSource(buildSSEUrl())
  es.onmessage = async (event) => {
    try {
      const payload = JSON.parse(event.data) as App.Api.Response<App.Api.SystemLog.Entry>
      if (payload.code === 1 && payload.data) {
        await pushLog(payload.data)
      }
    } catch {
      theToast.error(String(t('systemLog.sseParseFailed')))
    }
  }
  es.onerror = () => {
    stopSSE()
  }
}

const startWS = () => {
  transport.value = 'ws'
  open()
}

const restartStream = () => {
  stopSSE()
  close()
  startWS()
}

onMounted(async () => {
  await loadHistory()
  startWS()
  onMessage(async (payload) => {
    if (transport.value !== 'ws') return
    if (payload.code !== 1 || !payload.data) return
    try {
      await pushLog(payload.data)
    } catch {
      theToast.error(String(t('systemLog.processFailed')))
    }
  })
})

onUnmounted(() => {
  stopSSE()
  close()
})
</script>

<style scoped>
.log-toolbar {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.log-toolbar-grid {
  display: grid;
  grid-template-columns: repeat(1, minmax(0, 1fr));
  gap: 0.6rem;
}

.toolbar-field {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.field-label {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  line-height: 1.2;
}

.field-tail input {
  width: 100%;
}

.field-hint {
  font-size: 0.72rem;
  line-height: 1.2;
  color: var(--color-text-muted);
}

.log-toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.5rem;
}

.log-container {
  height: 42vh;
  overflow-y: auto;
  border: 1px solid var(--color-border-subtle);
  border-radius: var(--radius-lg);
  padding: 12px;
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

@media (min-width: 1024px) {
  .log-container {
    height: 48vh;
  }
}

.log-line {
  font-family: var(--font-family-mono);
  line-height: 1.55;
  margin-bottom: 0.65rem;
  word-break: break-word;
  color: var(--color-text-secondary);
}

.log-line-main {
  display: block;
}

.time {
  color: var(--color-text-muted);
}

.level {
  margin: 0 6px;
  color: var(--color-accent);
}

.msg {
  color: var(--color-text-primary);
}

.log-line-detail {
  display: block;
  margin-top: 0.25rem;
  padding-left: 0.35rem;
  border-left: 2px solid var(--color-border-strong);
  font-size: 0.92em;
  line-height: 1.45;
  color: var(--color-text-muted);
  white-space: pre-wrap;
}

.log-line-raw {
  margin-top: 0.35rem;
  padding-left: 0;
  overflow: visible;
}

/* outside 会被 .log-container 的 overflow 裁切；inside 与首行时间左缘对齐 */
.log-line-raw > summary {
  cursor: pointer;
  user-select: none;
  display: list-item;
  list-style-position: inside;
  list-style-type: disclosure-closed;
  padding-left: 0;
  color: var(--color-text-muted);
  font-size: 0.88em;
  line-height: 1.55;
}

.log-line-raw[open] > summary {
  list-style-type: disclosure-open;
}

.log-line-raw > summary::-webkit-details-marker {
  display: list-item;
}

.log-line-raw > summary::marker {
  color: var(--color-text-muted);
}

.log-line-raw-pre {
  margin: 0.35rem 0 0;
  padding: 0.5rem 0.6rem;
  max-height: 12rem;
  overflow: auto;
  font-size: 0.82em;
  line-height: 1.4;
  white-space: pre-wrap;
  word-break: break-word;
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border-subtle);
  background: var(--color-bg-muted);
  color: var(--color-text-secondary);
}

.log-btn {
  border: 1px solid var(--color-border-subtle);
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

.log-btn:hover {
  background: var(--color-bg-muted);
  border-color: var(--color-border-strong);
}

@media (min-width: 768px) {
  .log-toolbar-grid {
    grid-template-columns: minmax(8rem, 10rem) minmax(12rem, 1fr) minmax(9rem, 11rem);
    align-items: start;
  }
}
</style>
