<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { onClickOutside, useFetch } from '@vueuse/core'
import { useI18n } from 'vue-i18n'
import { getVisitorId } from '@/utils/visitor'

type ReactionCount = App.Api.Ech0.ReactionCount
type ReactionKey = App.Api.Ech0.ReactionKey

const CATALOG: App.Api.Ech0.ReactionCatalogItem[] = [
  { key: 'like', emoji: '👍' },
  { key: 'heart', emoji: '❤️' },
  { key: 'haha', emoji: '😂' },
  { key: 'wow', emoji: '😮' },
  { key: 'sad', emoji: '😢' },
  { key: 'cry', emoji: '😭' },
  { key: 'party', emoji: '🎉' },
]

const props = defineProps<{
  echo: { id: string; server_url: string; reactions?: ReactionCount[] }
}>()

const { t } = useI18n()
const items = ref<ReactionCount[]>((props.echo.reactions ?? []).map((item) => ({ ...item })))
const pickerOpen = ref(false)
const pickerRef = ref<HTMLElement | null>(null)
const catalog = ref<App.Api.Ech0.ReactionCatalogItem[]>([...CATALOG])

const catalogMap = computed(
  () => new Map(catalog.value.map((item) => [item.key, item.emoji] as const)),
)

watch(
  () => props.echo.reactions,
  (next) => {
    items.value = (next ?? []).map((item) => ({ ...item }))
  },
  { deep: true },
)

onMounted(() => {
  // Remote instances may expose more emoji than the fallback catalog; an
  // unreachable or older instance simply keeps the fallback.
  void (async () => {
    try {
      const { data } = await useFetch<App.Api.Response<App.Api.Ech0.ReactionCatalogItem[]>>(
        `${props.echo.server_url}/api/reactions`,
      )
        .get()
        .json()
      if (data.value?.code === 1 && data.value.data?.length) {
        catalog.value = data.value.data
      }
    } catch {
      // keep the fallback catalog
    }
  })()
})

onClickOutside(pickerRef, () => {
  pickerOpen.value = false
})

const label = (key: ReactionKey): string => String(t(`reaction.${key}`))

function snapshot(): ReactionCount[] {
  return items.value.map((item) => ({ ...item }))
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
      { key, emoji: catalogMap.value.get(key) ?? '', count: 1, reacted: true },
    ]
    return
  }
  if (!existing) return
  items.value =
    existing.count > 1
      ? items.value.map((item) =>
          item.key === key ? { ...item, count: item.count - 1, reacted: false } : item,
        )
      : items.value.filter((item) => item.key !== key)
}

async function send(key: ReactionKey, add: boolean) {
  const previous = snapshot()
  optimistic(key, add)
  try {
    const { error, data } = await useFetch<App.Api.Response<ReactionCount[]>>(
      `${props.echo.server_url}/api/echo/${encodeURIComponent(props.echo.id)}/reactions/${encodeURIComponent(key)}`,
      { method: add ? 'PUT' : 'DELETE', headers: { 'X-Ech0-Visitor': getVisitorId() } },
    ).json()

    if (error.value || data.value?.code !== 1) {
      throw new Error('reaction failed')
    }
    items.value = ((data.value?.data ?? []) as ReactionCount[]).map((item: ReactionCount) => ({ ...item }))
  } catch {
    items.value = previous
  }
}

function toggle(key: ReactionKey) {
  send(key, !items.value.find((item) => item.key === key)?.reacted)
}

function pick(key: ReactionKey) {
  pickerOpen.value = false
  toggle(key)
}
</script>

<template>
  <div class="hub-reactions">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="hub-reaction-chip"
      :class="{ 'hub-reaction-chip--active': item.reacted }"
      :aria-pressed="item.reacted"
      :title="label(item.key)"
      @click="toggle(item.key)"
    >
      <span aria-hidden="true">{{ item.emoji }}</span>
      <span class="hub-reaction-count">{{ item.count > 99 ? '99+' : item.count }}</span>
    </button>

    <div ref="pickerRef" class="hub-reaction-picker-wrap">
      <button
        type="button"
        class="hub-reaction-add"
        :title="t('reaction.add')"
        :aria-label="t('reaction.add')"
        @click="pickerOpen = !pickerOpen"
      >
        +
      </button>
      <div v-if="pickerOpen" class="hub-reaction-picker" role="menu">
        <button
          v-for="emoji in catalog"
          :key="emoji.key"
          type="button"
          role="menuitem"
          class="hub-reaction-picker-item"
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

<style scoped>
.hub-reactions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 0.25rem;
}

.hub-reaction-chip,
.hub-reaction-add {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  height: 1.55rem;
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
    background-color 150ms ease;
}

.hub-reaction-chip:hover,
.hub-reaction-add:hover {
  border-color: var(--color-border-strong);
}

.hub-reaction-chip--active {
  border-color: var(--color-accent);
  background: color-mix(in srgb, var(--color-accent) 12%, transparent);
  color: var(--color-text-primary);
}

.hub-reaction-count {
  font-variant-numeric: tabular-nums;
}

.hub-reaction-picker-wrap {
  position: relative;
  display: inline-flex;
}

.hub-reaction-add {
  justify-content: center;
  width: 1.55rem;
  padding: 0;
  border-style: dashed;
}

.hub-reaction-picker {
  position: absolute;
  bottom: calc(100% + 0.35rem);
  right: 0;
  z-index: 20;
  display: flex;
  gap: 0.15rem;
  padding: 0.3rem;
  border: 1px solid var(--color-border-subtle);
  border-radius: 0.6rem;
  background: var(--color-bg-surface);
  box-shadow: var(--shadow-sm);
}

.hub-reaction-picker-item {
  width: 1.8rem;
  height: 1.8rem;
  border: none;
  border-radius: 0.4rem;
  background: transparent;
  font-size: 1.05rem;
  line-height: 1;
  cursor: pointer;
}

.hub-reaction-picker-item:hover {
  background: var(--color-bg-muted);
}
</style>
