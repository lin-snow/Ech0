<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="echo-reactions">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="reaction-chip"
      :class="{ 'reaction-chip--active': item.reacted }"
      :aria-pressed="item.reacted"
      :title="label(item.key)"
      @click="toggle(item.key)"
    >
      <span class="reaction-chip-emoji" aria-hidden="true">{{ item.emoji }}</span>
      <span class="reaction-chip-count">{{ item.count > 99 ? '99+' : item.count }}</span>
    </button>

    <div ref="pickerRef" class="reaction-picker-wrap">
      <button
        type="button"
        class="reaction-add"
        :disabled="disabled"
        :title="t('reaction.add')"
        :aria-label="t('reaction.add')"
        @click="pickerOpen = !pickerOpen"
      >
        +
      </button>

      <div v-if="pickerOpen" class="reaction-picker" role="menu">
        <button
          v-for="emoji in catalog"
          :key="emoji.key"
          type="button"
          role="menuitem"
          class="reaction-picker-item"
          :title="label(emoji.key)"
          :aria-label="label(emoji.key)"
          @click="pick(emoji.key)"
        >
          {{ emoji.emoji }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { onClickOutside } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { fetchReactEcho, fetchReactionCatalog, fetchUnreactEcho } from '@/service/api'
import { useEchoStore } from '@/stores'
import { isStaticMode } from '@/service/request/shared'
import { theToast } from '@/utils/toast'
import { getVisitorId } from '@/utils/visitor'

type ReactionCount = App.Api.Ech0.ReactionCount
type ReactionKey = App.Api.Ech0.ReactionKey

type ReactionEcho = {
  id: string
  reactions?: ReactionCount[]
}

const FALLBACK_CATALOG: App.Api.Ech0.ReactionCatalogItem[] = [
  { key: 'like', emoji: '👍' },
  { key: 'heart', emoji: '❤️' },
  { key: 'haha', emoji: '😂' },
  { key: 'wow', emoji: '😮' },
  { key: 'sad', emoji: '😢' },
  { key: 'cry', emoji: '😭' },
  { key: 'party', emoji: '🎉' },
]

const props = defineProps<{
  echo: ReactionEcho
  /** When set, reactions are sent to this remote instance instead of the local API. */
  baseUrl?: string
}>()

const { t } = useI18n()

const staticMode = isStaticMode()
const disabled = computed(() => staticMode)
const catalog = ref<App.Api.Ech0.ReactionCatalogItem[]>(FALLBACK_CATALOG)
const pickerOpen = ref(false)
const pickerRef = ref<HTMLElement | null>(null)

const catalogMap = computed(
  () => new Map(catalog.value.map((item) => [item.key, item.emoji] as const)),
)

const snapshotOf = (value: ReactionCount[]): ReactionCount[] => value.map((item) => ({ ...item }))

const echoStore = useEchoStore()

const applyServer = (next: ReactionCount[] | null | undefined) => {
  items.value = (next ?? []).map((item) => ({ ...item }))
  // Keep the cached feed row in sync; harmless for remote hub cards.
  echoStore.setEchoReactions(props.echo.id, items.value)
}

const items = ref<ReactionCount[]>(snapshotOf(props.echo.reactions ?? []))

watch(
  () => props.echo.reactions,
  (next) => {
    items.value = snapshotOf(next ?? [])
  },
  { deep: true },
)

onClickOutside(pickerRef, () => {
  pickerOpen.value = false
})

// Literal labelKey entries double as the i18n usage index for the unused-key
// check, which cannot resolve a template-literal key.
const REACTION_LABELS: { key: ReactionKey; labelKey: string }[] = [
  { key: 'like', labelKey: 'reaction.like' },
  { key: 'heart', labelKey: 'reaction.heart' },
  { key: 'haha', labelKey: 'reaction.haha' },
  { key: 'wow', labelKey: 'reaction.wow' },
  { key: 'sad', labelKey: 'reaction.sad' },
  { key: 'cry', labelKey: 'reaction.cry' },
  { key: 'party', labelKey: 'reaction.party' },
]

const label = (key: ReactionKey): string => {
  const entry = REACTION_LABELS.find((item) => item.key === key)
  return entry ? String(t(entry.labelKey)) : key
}

const toggle = (key: ReactionKey) => {
  const current = items.value.find((item) => item.key === key)
  request(key, !current?.reacted)
}

const pick = (key: ReactionKey) => {
  pickerOpen.value = false
  const current = items.value.find((item) => item.key === key)
  request(key, !current?.reacted)
}

async function request(key: ReactionKey, add: boolean) {
  if (staticMode) {
    theToast.info(String(t('reaction.unavailable')))
    return
  }

  const snapshot = snapshotOf(items.value)
  optimistic(key, add)

  try {
    const next = props.baseUrl
      ? await callRemote(key, add ? 'PUT' : 'DELETE')
      : add
        ? (await fetchReactEcho(props.echo.id, key)).data
        : (await fetchUnreactEcho(props.echo.id, key)).data
    applyServer(next)
  } catch {
    items.value = snapshot
    if (props.baseUrl) {
      theToast.error(String(t('reaction.failed')))
    }
  }
}

async function callRemote(key: ReactionKey, method: 'PUT' | 'DELETE'): Promise<ReactionCount[]> {
  const response = await fetch(
    `${props.baseUrl}/api/echo/${encodeURIComponent(props.echo.id)}/reactions/${encodeURIComponent(key)}`,
    {
      method,
      credentials: 'omit',
      headers: { 'X-Ech0-Visitor': getVisitorId() },
    },
  )
  const payload = (await response.json()) as App.Api.Response<ReactionCount[]>
  if (!response.ok || payload.code !== 1) {
    throw new Error(payload.msg || 'reaction failed')
  }
  return payload.data ?? []
}

function optimistic(key: ReactionKey, add: boolean) {
  const existing = items.value.find((item) => item.key === key)

  if (add) {
    if (existing) {
      items.value = items.value.map((item) =>
        item.key === key ? { ...item, count: item.count + 1, reacted: true } : item,
      )
      return
    }
    items.value = [
      ...items.value,
      {
        key,
        emoji: catalogMap.value.get(key) ?? '',
        count: 1,
        reacted: true,
      },
    ]
    return
  }

  if (!existing) {
    return
  }
  items.value =
    existing.count > 1
      ? items.value.map((item) =>
          item.key === key ? { ...item, count: item.count - 1, reacted: false } : item,
        )
      : items.value.filter((item) => item.key !== key)
}

onMounted(async () => {
  if (props.baseUrl || staticMode) {
    return
  }
  const res = await fetchReactionCatalog()
  if (res.code === 1 && res.data?.length) {
    catalog.value = res.data
  }
})
</script>

<style scoped>
.echo-reactions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.25rem;
  position: relative;
}

.reaction-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  height: 1.6rem;
  padding: 0 0.45rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 0.72rem;
  line-height: 1;
  cursor: pointer;
  transition:
    border-color 150ms ease,
    background-color 150ms ease,
    transform 150ms ease;
}

.reaction-chip:hover {
  border-color: var(--color-border-strong);
  transform: translateY(-1px);
}

.reaction-chip--active {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  color: var(--color-text-primary);
}

.reaction-chip-emoji {
  font-size: 0.9rem;
}

.reaction-chip-count {
  font-family: var(--font-family-mono);
}

.reaction-picker-wrap {
  position: relative;
  display: inline-flex;
}

.reaction-add {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.6rem;
  height: 1.6rem;
  border: 1px dashed var(--color-border-subtle);
  border-radius: 999px;
  background: transparent;
  color: var(--color-text-muted);
  font-size: 0.9rem;
  line-height: 1;
  cursor: pointer;
  transition:
    border-color 150ms ease,
    color 150ms ease;
}

.reaction-add:hover:not(:disabled) {
  border-color: var(--color-accent);
  color: var(--color-accent);
}

.reaction-add:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.reaction-picker {
  position: absolute;
  bottom: calc(100% + 0.35rem);
  left: 50%;
  z-index: 20;
  display: flex;
  gap: 0.15rem;
  padding: 0.3rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.6rem;
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
  transform: translateX(-50%);
}

.reaction-picker-item {
  width: 1.8rem;
  height: 1.8rem;
  border: none;
  border-radius: 0.4rem;
  background: transparent;
  font-size: 1.05rem;
  line-height: 1;
  cursor: pointer;
  transition: background-color 150ms ease;
}

.reaction-picker-item:hover {
  background: var(--color-bg-muted);
}
</style>
