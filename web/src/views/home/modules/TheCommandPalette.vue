<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <Teleport to="body">
    <Transition name="palette">
      <div
        v-if="modelValue"
        class="palette"
        role="dialog"
        aria-modal="true"
        :aria-label="t('commandPalette.title')"
        @mousedown.self="close"
      >
        <div class="palette__panel" role="document" @keydown="onPanelKeydown">
          <div class="palette__search-row">
            <Search class="palette__search-icon" />
            <input
              ref="keywordInputRef"
              v-model="draftKeyword"
              type="search"
              class="palette__search-input"
              :placeholder="t('commandPalette.placeholder')"
              :aria-label="t('commandPalette.keywordLabel')"
            />
            <span
              v-if="activeFilterCount > 0"
              class="palette__badge"
              :aria-label="t('commandPalette.activeCountSuffix', { count: activeFilterCount })"
            >
              {{ activeFilterCount }}
            </span>
            <button
              type="button"
              class="palette__icon-btn"
              :aria-label="t('commandPalette.close')"
              @click="close"
            >
              <Close class="w-4 h-4" />
            </button>
          </div>

          <div class="palette__body">
            <section class="palette__section">
              <header class="palette__section-header">
                <span class="palette__section-label">{{ t('commandPalette.dateRangeLabel') }}</span>
                <button
                  v-if="draftFrom !== null || draftTo !== null"
                  type="button"
                  class="palette__section-action"
                  @click="clearDate"
                >
                  {{ t('commandPalette.reset') }}
                </button>
              </header>
              <div class="palette__preset-row">
                <button
                  v-for="preset in presets"
                  :key="preset.key"
                  type="button"
                  class="palette__pill"
                  :class="{ 'palette__pill--active': activePreset === preset.key }"
                  @mousedown.prevent
                  @click="applyPreset(preset.key)"
                >
                  {{ preset.label }}
                </button>
              </div>
              <div class="palette__date-row">
                <label class="palette__date-field">
                  <span class="palette__date-label">{{ t('commandPalette.dateFromLabel') }}</span>
                  <span
                    class="palette__date-control"
                    :class="{ 'palette__date-control--empty': !draftFromStr }"
                  >
                    <svg
                      class="palette__date-icon"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path
                        fill="currentColor"
                        d="M19 4h-1V2h-2v2H8V2H6v2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2m0 16H5V10h14zm0-12H5V6h14z"
                      />
                    </svg>
                    <input
                      type="date"
                      class="palette__date-input"
                      :value="draftFromStr"
                      :max="draftToStr || undefined"
                      @input="onFromInput(($event.target as HTMLInputElement).value)"
                    />
                  </span>
                </label>
                <svg
                  class="palette__date-arrow"
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="1.6"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  aria-hidden="true"
                >
                  <path d="M5 12h14m0 0l-6-6m6 6l-6 6" />
                </svg>
                <label class="palette__date-field">
                  <span class="palette__date-label">{{ t('commandPalette.dateToLabel') }}</span>
                  <span
                    class="palette__date-control"
                    :class="{ 'palette__date-control--empty': !draftToStr }"
                  >
                    <svg
                      class="palette__date-icon"
                      xmlns="http://www.w3.org/2000/svg"
                      viewBox="0 0 24 24"
                      aria-hidden="true"
                    >
                      <path
                        fill="currentColor"
                        d="M19 4h-1V2h-2v2H8V2H6v2H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2m0 16H5V10h14zm0-12H5V6h14z"
                      />
                    </svg>
                    <input
                      type="date"
                      class="palette__date-input"
                      :value="draftToStr"
                      :min="draftFromStr || undefined"
                      @input="onToInput(($event.target as HTMLInputElement).value)"
                    />
                  </span>
                </label>
              </div>
            </section>

            <section v-if="canFilterVisibility" class="palette__section">
              <header class="palette__section-header">
                <span class="palette__section-label">{{
                  t('commandPalette.visibilityLabel')
                }}</span>
                <button
                  v-if="draftVisibility !== 'all'"
                  type="button"
                  class="palette__section-action"
                  @click="draftVisibility = 'all'"
                >
                  {{ t('commandPalette.reset') }}
                </button>
              </header>
              <div class="palette__preset-row">
                <button
                  v-for="option in visibilityOptions"
                  :key="option.key"
                  type="button"
                  class="palette__pill"
                  :class="{ 'palette__pill--active': draftVisibility === option.key }"
                  @mousedown.prevent
                  @click="draftVisibility = option.key"
                >
                  {{ option.label }}
                </button>
              </div>
            </section>

            <section class="palette__section">
              <header class="palette__section-header">
                <span class="palette__section-label">
                  {{ t('commandPalette.tagsLabel') }}
                  <span class="palette__section-hint">· {{ t('commandPalette.tagsHint') }}</span>
                </span>
                <button
                  v-if="draftTagIds.length > 0"
                  type="button"
                  class="palette__section-action"
                  @click="clearTags"
                >
                  {{ t('commandPalette.tagsClear') }}
                </button>
              </header>
              <div v-if="tagList.length === 0" class="palette__empty">
                {{ t('commandPalette.tagsEmpty') }}
              </div>
              <div v-else class="palette__tag-grid">
                <button
                  v-for="tag in visibleTags"
                  :key="tag.id"
                  type="button"
                  class="palette__pill palette__pill--tag"
                  :class="{ 'palette__pill--active': draftTagIds.includes(tag.id) }"
                  @mousedown.prevent
                  @click="toggleTag(tag.id)"
                >
                  <span class="palette__tag-hash">#</span>{{ tag.name }}
                </button>
                <button
                  v-if="tagList.length > visibleTagLimit"
                  type="button"
                  class="palette__more"
                  @mousedown.prevent
                  @click="showAllTags = !showAllTags"
                >
                  {{
                    showAllTags
                      ? t('commandPalette.tagsShowLess')
                      : t('commandPalette.tagsMoreHidden', {
                          count: tagList.length - visibleTagLimit,
                        })
                  }}
                </button>
              </div>
            </section>
          </div>

          <footer class="palette__footer">
            <div class="palette__kbd-row" aria-hidden="true">
              <span class="palette__kbd-group">
                <kbd class="palette__kbd">↵</kbd>
                <span>{{ t('commandPalette.kbdApply') }}</span>
              </span>
              <span class="palette__kbd-group">
                <kbd class="palette__kbd">esc</kbd>
                <span>{{ t('commandPalette.kbdClose') }}</span>
              </span>
            </div>
            <div class="palette__footer-actions">
              <button
                type="button"
                class="palette__btn palette__btn--ghost"
                :disabled="!isDraftDirty && !isDraftActive"
                @click="handleReset"
              >
                {{ t('commandPalette.reset') }}
              </button>
              <button type="button" class="palette__btn palette__btn--primary" @click="handleApply">
                {{ t('commandPalette.apply') }}
              </button>
            </div>
          </footer>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { storeToRefs } from 'pinia'
import { useI18n } from 'vue-i18n'
import { useEchoStore, useUserStore } from '@/stores'
import Close from '@/components/icons/close.vue'
import Search from '@/components/icons/search.vue'

const props = defineProps<{
  modelValue: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
}>()

const { t } = useI18n()
const echoStore = useEchoStore()
const userStore = useUserStore()
const {
  searchValue,
  dateFrom,
  dateTo,
  selectedTagIds,
  tagList,
  isFilteringMode,
  filteredTag,
  visibilityFilter,
} = storeToRefs(echoStore)

const canFilterVisibility = computed(() => userStore.isLogin && userStore.user?.is_admin === true)

const draftKeyword = ref<string>('')
const draftFrom = ref<number | null>(null)
const draftTo = ref<number | null>(null)
const draftTagIds = ref<string[]>([])
const draftVisibility = ref<App.Api.Ech0.EchoVisibilityFilter>('all')
const keywordInputRef = ref<HTMLInputElement | null>(null)
const showAllTags = ref<boolean>(false)
const visibleTagLimit = 12

const toDateStr = (sec: number | null): string => {
  if (sec === null) return ''
  const d = new Date(sec * 1000)
  if (Number.isNaN(d.getTime())) return ''
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const parseDateStr = (str: string, endOfDay = false): number | null => {
  if (!str) return null
  const [y, m, d] = str.split('-').map(Number)
  if (!y || !m || !d) return null
  const date = endOfDay ? new Date(y, m - 1, d, 23, 59, 59) : new Date(y, m - 1, d, 0, 0, 0)
  if (Number.isNaN(date.getTime())) return null
  return Math.floor(date.getTime() / 1000)
}

const draftFromStr = computed(() => toDateStr(draftFrom.value))
const draftToStr = computed(() => toDateStr(draftTo.value))

const presets = computed(() => [
  { key: 'today', label: t('commandPalette.dateRangeToday') },
  { key: 'last7', label: t('commandPalette.dateRangeLast7Days') },
  { key: 'last30', label: t('commandPalette.dateRangeLast30Days') },
  { key: 'thisYear', label: t('commandPalette.dateRangeThisYear') },
])

const visibilityOptions = computed(() => [
  { key: 'all' as const, label: t('commandPalette.visibilityAll') },
  { key: 'public' as const, label: t('commandPalette.visibilityPublic') },
  { key: 'private' as const, label: t('commandPalette.visibilityPrivate') },
])

const activePreset = computed<string | null>(() => {
  if (draftFrom.value === null && draftTo.value === null) return null

  const now = new Date()
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
  const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
  const toSec = (d: Date) => Math.floor(d.getTime() / 1000)

  const matches = (fromSec: number, toSec2: number) =>
    draftFrom.value === fromSec && draftTo.value === toSec2

  if (matches(toSec(todayStart), toSec(todayEnd))) return 'today'

  const last7Start = new Date(todayStart)
  last7Start.setDate(last7Start.getDate() - 6)
  if (matches(toSec(last7Start), toSec(todayEnd))) return 'last7'

  const last30Start = new Date(todayStart)
  last30Start.setDate(last30Start.getDate() - 29)
  if (matches(toSec(last30Start), toSec(todayEnd))) return 'last30'

  const yearStart = new Date(now.getFullYear(), 0, 1, 0, 0, 0)
  if (matches(toSec(yearStart), toSec(todayEnd))) return 'thisYear'

  return null
})

const applyPreset = (key: string) => {
  if (activePreset.value === key) {
    draftFrom.value = null
    draftTo.value = null
    return
  }

  const now = new Date()
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0)
  const todayEnd = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59)
  const toSec = (d: Date) => Math.floor(d.getTime() / 1000)

  if (key === 'today') {
    draftFrom.value = toSec(todayStart)
    draftTo.value = toSec(todayEnd)
  } else if (key === 'last7') {
    const start = new Date(todayStart)
    start.setDate(start.getDate() - 6)
    draftFrom.value = toSec(start)
    draftTo.value = toSec(todayEnd)
  } else if (key === 'last30') {
    const start = new Date(todayStart)
    start.setDate(start.getDate() - 29)
    draftFrom.value = toSec(start)
    draftTo.value = toSec(todayEnd)
  } else if (key === 'thisYear') {
    const start = new Date(now.getFullYear(), 0, 1, 0, 0, 0)
    draftFrom.value = toSec(start)
    draftTo.value = toSec(todayEnd)
  }
}

const onFromInput = (value: string) => {
  draftFrom.value = parseDateStr(value, false)
}

const onToInput = (value: string) => {
  draftTo.value = parseDateStr(value, true)
}

const toggleTag = (id: string) => {
  const idx = draftTagIds.value.indexOf(id)
  if (idx === -1) {
    draftTagIds.value = [...draftTagIds.value, id]
  } else {
    draftTagIds.value = draftTagIds.value.filter((t) => t !== id)
  }
}

const clearTags = () => {
  draftTagIds.value = []
}

const clearDate = () => {
  draftFrom.value = null
  draftTo.value = null
}

const visibleTags = computed(() => {
  if (showAllTags.value) return tagList.value
  return tagList.value.slice(0, visibleTagLimit)
})

const arraysEqualAsSet = (a: string[], b: string[]) => {
  if (a.length !== b.length) return false
  const setB = new Set(b)
  return a.every((v) => setB.has(v))
}

const currentTagDraftFromStore = (): string[] => {
  const ids = new Set<string>(selectedTagIds.value)
  if (isFilteringMode.value && filteredTag.value?.id) {
    ids.add(filteredTag.value.id)
  }
  return Array.from(ids)
}

const isDraftDirty = computed(() => {
  const trimmed = draftKeyword.value.trim()
  if (trimmed !== (searchValue.value ?? '')) return true
  if (draftFrom.value !== dateFrom.value) return true
  if (draftTo.value !== dateTo.value) return true
  if (draftVisibility.value !== visibilityFilter.value) return true
  if (!arraysEqualAsSet(draftTagIds.value, currentTagDraftFromStore())) return true
  return false
})

const isDraftActive = computed(
  () =>
    draftKeyword.value.trim().length > 0 ||
    draftFrom.value !== null ||
    draftTo.value !== null ||
    draftVisibility.value !== 'all' ||
    draftTagIds.value.length > 0,
)

const activeFilterCount = computed(() => {
  let n = 0
  if (draftKeyword.value.trim().length > 0) n += 1
  if (draftFrom.value !== null || draftTo.value !== null) n += 1
  if (draftVisibility.value !== 'all') n += 1
  if (draftTagIds.value.length > 0) n += draftTagIds.value.length
  return n
})

const syncDraftFromStore = () => {
  draftKeyword.value = searchValue.value ?? ''
  draftFrom.value = dateFrom.value
  draftTo.value = dateTo.value
  draftVisibility.value = visibilityFilter.value
  draftTagIds.value = currentTagDraftFromStore()
}

const close = () => {
  emit('update:modelValue', false)
}

const handleApply = () => {
  const trimmed = draftKeyword.value.trim()
  const keywordChanged = trimmed !== (searchValue.value ?? '')
  const fromChanged = draftFrom.value !== dateFrom.value
  const toChanged = draftTo.value !== dateTo.value
  const visibilityChanged = draftVisibility.value !== visibilityFilter.value
  const tagsChanged = !arraysEqualAsSet(draftTagIds.value, currentTagDraftFromStore())

  if (!keywordChanged && !fromChanged && !toChanged && !visibilityChanged && !tagsChanged) {
    close()
    return
  }

  echoStore.searchValue = trimmed
  echoStore.dateFrom = draftFrom.value
  echoStore.dateTo = draftTo.value
  echoStore.visibilityFilter = draftVisibility.value

  echoStore.selectedTagIds = [...draftTagIds.value]
  echoStore.isFilteringMode = false
  echoStore.filteredTag = null

  const searchingBefore = (searchValue.value ?? '').length > 0
  const searchingAfter = trimmed.length > 0
  const exitsSearchMode = searchingBefore && !searchingAfter
  if (!exitsSearchMode) {
    echoStore.refreshEchos()
  }

  close()
}

const handleReset = () => {
  draftKeyword.value = ''
  draftFrom.value = null
  draftTo.value = null
  draftVisibility.value = 'all'
  draftTagIds.value = []
}

const onPanelKeydown = (e: KeyboardEvent) => {
  if (e.key !== 'Enter' || e.isComposing) return
  if (e.target instanceof HTMLButtonElement) return
  e.preventDefault()
  handleApply()
}

watch(
  () => props.modelValue,
  (open) => {
    if (open) {
      syncDraftFromStore()
      void echoStore.ensureTagsLoaded()
      showAllTags.value = false
      void nextTick(() => {
        keywordInputRef.value?.focus()
        keywordInputRef.value?.select()
      })
    }
  },
)
</script>

<style scoped>
.palette {
  position: fixed;
  inset: 0;
  z-index: 9000;
  display: flex;
  justify-content: center;
  align-items: flex-start;
  padding: clamp(4.5rem, 12vh, 7rem) 1rem 1rem;
  background: var(--color-overlay-strong);
  backdrop-filter: blur(6px);
}

.palette__panel {
  width: 100%;
  max-width: 34rem;
  max-height: calc(100vh - 7rem);
  display: flex;
  flex-direction: column;
  border: 1px solid var(--color-border-subtle);
  border-radius: 14px;
  background: var(--color-bg-surface);
  box-shadow:
    0 1px 1px rgb(0 0 0 / 4%),
    0 20px 45px -18px rgb(0 0 0 / 30%),
    0 40px 80px -30px rgb(0 0 0 / 22%);
  overflow: hidden;
  font-family: var(--font-family-base, inherit);
}

.palette__search-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.85rem 1rem;
  border-bottom: 1px solid var(--color-border-subtle);
}

.palette__search-icon {
  width: 1.1rem;
  height: 1.1rem;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

:deep(.palette__search-icon path) {
  fill: currentColor;
}

.palette__search-input {
  flex: 1 1 auto;
  min-width: 0;
  padding: 0.25rem 0;
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--color-text-primary);
  background: transparent;
  border: 0;
  outline: none;
  letter-spacing: 0.01em;
}

.palette__search-input::placeholder {
  color: var(--color-text-muted);
  font-weight: 400;
}

.palette__search-input::-webkit-search-cancel-button {
  appearance: none;
}

.palette__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.2rem;
  height: 1.2rem;
  padding: 0 0.38rem;
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--color-accent, #e07020);
  background: color-mix(in srgb, var(--color-accent, #e07020) 14%, transparent);
  border-radius: 999px;
  line-height: 1;
  flex-shrink: 0;
  letter-spacing: 0.02em;
}

.palette__icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 1.75rem;
  height: 1.75rem;
  padding: 0;
  border: 0;
  border-radius: 8px;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease;
  flex-shrink: 0;
}

.palette__icon-btn:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.palette__body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding: 0.9rem 1rem 0.6rem;
  overflow-y: auto;
  flex: 1 1 auto;
}

.palette__section {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.palette__section-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
}

.palette__section-label {
  font-size: 0.7rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.palette__section-hint {
  font-weight: 400;
  letter-spacing: 0.02em;
  text-transform: none;
  color: var(--color-text-muted);
  opacity: 0.8;
}

.palette__section-action {
  font-size: 0.72rem;
  color: var(--color-text-muted);
  background: transparent;
  border: 0;
  padding: 0.1rem 0.3rem;
  border-radius: 4px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease;
}

.palette__section-action:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
}

.palette__preset-row,
.palette__tag-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.palette__pill {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  padding: 0.28rem 0.65rem;
  font-size: 0.8rem;
  line-height: 1;
  color: var(--color-text-secondary);
  background: transparent;
  border: 1px solid var(--color-border-subtle);
  border-radius: 999px;
  cursor: pointer;
  transition:
    color 0.12s ease,
    background 0.12s ease,
    border-color 0.12s ease,
    transform 0.08s ease;
}

.palette__pill:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
  border-color: var(--color-border-strong);
}

.palette__pill:active {
  transform: scale(0.97);
}

.palette__pill--active {
  color: var(--color-accent, #e07020);
  background: color-mix(in srgb, var(--color-accent, #e07020) 12%, transparent);
  border-color: color-mix(in srgb, var(--color-accent, #e07020) 35%, transparent);
  font-weight: 500;
}

.palette__pill--active:hover {
  color: var(--color-accent, #e07020);
  background: color-mix(in srgb, var(--color-accent, #e07020) 18%, transparent);
  border-color: color-mix(in srgb, var(--color-accent, #e07020) 50%, transparent);
}

.palette__pill--tag .palette__tag-hash {
  color: inherit;
  opacity: 0.55;
  margin-right: 0.1rem;
  font-weight: 500;
}

.palette__more {
  font-size: 0.78rem;
  color: var(--color-text-muted);
  background: transparent;
  border: 0;
  padding: 0.28rem 0.5rem;
  border-radius: 999px;
  cursor: pointer;
  transition: color 0.12s ease;
}

.palette__more:hover {
  color: var(--color-text-primary);
}

.palette__empty {
  padding: 0.55rem 0.3rem;
  font-size: 0.8rem;
  color: var(--color-text-muted);
  font-style: italic;
}

.palette__date-row {
  display: flex;
  align-items: flex-end;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.palette__date-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  flex: 1 1 9rem;
  min-width: 0;
}

.palette__date-label {
  font-size: 0.66rem;
  font-weight: 500;
  color: var(--color-text-muted);
  letter-spacing: 0.06em;
  text-transform: uppercase;
  padding-left: 0.1rem;
}

.palette__date-control {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.55rem;
  background: var(--color-bg-muted, transparent);
  border: 1px solid var(--color-border-subtle);
  border-radius: 8px;
  cursor: text;
  transition:
    border-color 0.15s ease,
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.palette__date-control:hover {
  border-color: var(--color-border-strong);
}

.palette__date-control:focus-within {
  border-color: var(--color-accent, #e07020);
  background: var(--color-bg-surface);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--color-accent, #e07020) 14%, transparent);
}

.palette__date-icon {
  width: 0.95rem;
  height: 0.95rem;
  color: var(--color-text-muted);
  flex-shrink: 0;
  transition: color 0.15s ease;
}

.palette__date-control:focus-within .palette__date-icon {
  color: var(--color-accent, #e07020);
}

.palette__date-control--empty .palette__date-icon {
  opacity: 0.7;
}

.palette__date-input {
  flex: 1 1 auto;
  min-width: 0;
  appearance: none;
  padding: 0;
  font-size: 0.82rem;
  font-family: var(--font-family-mono, inherit);
  font-variant-numeric: tabular-nums;
  letter-spacing: 0.01em;
  color: var(--color-text-primary);
  background: transparent;
  border: 0;
  outline: none;
}

.palette__date-input::-webkit-date-and-time-value {
  text-align: left;
  min-height: 1.2em;
}

.palette__date-input::-webkit-calendar-picker-indicator {
  opacity: 0.55;
  cursor: pointer;
  transition: opacity 0.15s ease;
}

.palette__date-input::-webkit-calendar-picker-indicator:hover {
  opacity: 1;
}

.palette__date-arrow {
  width: 1rem;
  height: 1rem;
  margin-bottom: 0.6rem;
  color: var(--color-text-muted);
  flex-shrink: 0;
  opacity: 0.7;
}

.palette__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
  padding: 0.7rem 1rem 0.8rem;
  border-top: 1px solid var(--color-border-subtle);
  background: color-mix(in srgb, var(--color-bg-muted) 45%, transparent);
}

.palette__kbd-row {
  display: flex;
  align-items: center;
  gap: 0.85rem;
  font-size: 0.7rem;
  color: var(--color-text-muted);
}

.palette__kbd-group {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
}

.palette__kbd {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 1.2rem;
  height: 1.2rem;
  padding: 0 0.3rem;
  font-family: var(--font-family-mono, monospace);
  font-size: 0.68rem;
  line-height: 1;
  color: var(--color-text-secondary);
  background: var(--color-bg-surface);
  border: 1px solid var(--color-border-subtle);
  border-bottom-width: 2px;
  border-radius: 4px;
}

.palette__footer-actions {
  display: flex;
  gap: 0.4rem;
}

.palette__btn {
  padding: 0.38rem 0.95rem;
  font-size: 0.82rem;
  font-weight: 500;
  border-radius: 8px;
  border: 1px solid transparent;
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease,
    border-color 0.15s ease,
    opacity 0.15s ease;
}

.palette__btn--ghost {
  color: var(--color-text-muted);
  background: transparent;
  border-color: var(--color-border-subtle);
}

.palette__btn--ghost:hover:not(:disabled) {
  color: var(--color-text-primary);
  background: var(--color-bg-muted);
  border-color: var(--color-border-strong);
}

.palette__btn--ghost:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.palette__btn--primary {
  color: #fff;
  background: var(--color-accent, #e07020);
  border-color: transparent;
  box-shadow: 0 1px 2px rgb(0 0 0 / 8%);
  font-weight: 500;
  letter-spacing: 0.01em;
}

.palette__btn--primary:hover {
  color: #fff;
  filter: brightness(0.95);
}

.palette__btn--primary:active {
  transform: translateY(1px);
}

@media (width <= 640px) {
  .palette {
    padding: 3rem 0.75rem 0.75rem;
  }

  .palette__panel {
    max-height: calc(100vh - 4rem);
  }

  .palette__kbd-row {
    display: none;
  }

  .palette__footer {
    justify-content: flex-end;
  }
}

.palette-enter-active,
.palette-leave-active {
  transition: opacity 0.18s ease;
}

.palette-enter-active .palette__panel,
.palette-leave-active .palette__panel {
  transition:
    transform 0.22s cubic-bezier(0.2, 0.9, 0.3, 1.2),
    opacity 0.22s ease;
}

.palette-enter-from,
.palette-leave-to {
  opacity: 0;
}

.palette-enter-from .palette__panel,
.palette-leave-to .palette__panel {
  opacity: 0;
  transform: translateY(-10px) scale(0.97);
}
</style>
