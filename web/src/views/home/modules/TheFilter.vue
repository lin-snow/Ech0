<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="w-full sm:px-2 mb-1 sm:mb-0">
    <div class="w-full flex flex-col gap-2">
      <div class="flex justify-start items-center gap-2 w-full flex-wrap">
        <div
          v-if="!isFilteringMode"
          class="home-filter__search-shell"
          :class="{
            'home-filter__search-shell--with-chat': showChatTrigger && chatAvailable,
            'home-filter__search-shell--win': !isMac,
          }"
        >
          <BaseInput
            v-tooltip="t('homeTop.searchTitle')"
            type="text"
            v-model="searchContent"
            :placeholder="t('homeTop.searchPlaceholder')"
            class="h-9 w-full! max-w-none rounded-[var(--radius-xs)]! border-[var(--filter-search-input-border)]! bg-transparent! shadow-none! text-[var(--color-text-secondary)] focus:ring-0!"
            @keyup.enter="($event.target as HTMLInputElement).blur()"
            @blur="handleSearch"
          />
          <div class="home-filter__shell-keys">
            <button
              type="button"
              class="home-filter__kbd-hint"
              :aria-label="t('commandPalette.title')"
              v-tooltip="t('commandPalette.title')"
              @click="emit('openPalette')"
            >
              {{ shortcutBadge }}
            </button>
            <button
              v-if="showChatTrigger && chatAvailable"
              type="button"
              class="home-filter__chat-trigger"
              :aria-label="t('chatLauncher.title')"
              v-tooltip="chatTooltip"
              @click="emit('openChat')"
            >
              <Chat class="home-filter__chat-icon" />
            </button>
          </div>
        </div>
        <Filter v-if="isFilteringMode" class="w-7 h-7" />
        <div
          v-if="isFilteringMode && filteredTag"
          @click="handleCancelTapFilter"
          class="home-filter__chip"
        >
          <span class="home-filter__chip-hash">#</span>
          <p class="text-nowrap truncate">{{ filteredTag.name }}</p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
      </div>

      <div
        v-if="isDateRangeActive || isVisibilityFilterActive || selectedTagChips.length > 0"
        class="flex flex-wrap gap-1.5"
      >
        <div
          v-if="isDateRangeActive"
          class="home-filter__chip"
          v-tooltip="t('commandPalette.reset')"
          @click="handleClearDateRange"
        >
          <p class="text-nowrap truncate">
            {{ t('commandPalette.activeChipDatePrefix') }} · {{ dateRangeSummary }}
          </p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
        <div
          v-if="isVisibilityFilterActive"
          class="home-filter__chip"
          v-tooltip="t('commandPalette.reset')"
          @click="handleClearVisibility"
        >
          <p class="text-nowrap truncate">
            {{ t('commandPalette.activeChipVisibilityPrefix') }} · {{ visibilitySummary }}
          </p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
        <div
          v-for="chip in selectedTagChips"
          :key="chip.id"
          class="home-filter__chip"
          @click="handleRemoveTag(chip.id)"
        >
          <span class="home-filter__chip-hash">#</span>
          <p class="text-nowrap truncate">{{ chip.name }}</p>
          <Close class="inline w-4 h-4 ml-1 shrink-0" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import BaseInput from '@/components/common/BaseInput.vue'
import { useEchoStore, useSettingStore, useUserStore } from '@/stores'
import { storeToRefs } from 'pinia'
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import Close from '@/components/icons/close.vue'
import Filter from '@/components/icons/filter.vue'
import Chat from '@/components/icons/chat.vue'

withDefaults(
  defineProps<{
    showChatTrigger?: boolean
  }>(),
  { showChatTrigger: false },
)

const emit = defineEmits<{
  (e: 'openPalette'): void
  (e: 'openChat'): void
}>()

const echoStore = useEchoStore()
const userStore = useUserStore()
const settingStore = useSettingStore()
const { isLogin } = storeToRefs(userStore)
const { AgentSetting } = storeToRefs(settingStore)
const {
  refreshForSearch,
  fetchCurrentPage,
  refreshEchos,
  resetDateRange,
  resetVisibilityFilter,
  removeSelectedTag,
  ensureTagsLoaded,
} = echoStore
const {
  searchingMode,
  filteredTag,
  isFilteringMode,
  searchValue,
  dateFrom,
  dateTo,
  isDateRangeActive,
  visibilityFilter,
  isVisibilityFilterActive,
  selectedTagIds,
  tagList,
} = storeToRefs(echoStore)
const { t } = useI18n()

const searchContent = ref<string>(searchValue.value)

const isMac =
  typeof navigator !== 'undefined' && /Mac|iPhone|iPad|iPod/.test(navigator.platform || '')
const shortcutBadge = computed(() => (isMac ? '⌘K' : 'Ctrl+K'))

const chatAvailable = computed(() => isLogin.value && AgentSetting.value.enable)
const chatTooltip = computed(() => `${t('chatLauncher.title')} · ${isMac ? '⌘J' : 'Ctrl+J'}`)

const formatDate = (sec: number | null): string => {
  if (sec === null) return '…'
  const d = new Date(sec * 1000)
  if (Number.isNaN(d.getTime())) return '…'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}

const dateRangeSummary = computed(() => {
  const from = formatDate(dateFrom.value)
  const to = formatDate(dateTo.value)
  return `${from} ${t('commandPalette.rangeSeparator')} ${to}`
})

const visibilitySummary = computed(() =>
  visibilityFilter.value === 'private'
    ? t('commandPalette.visibilityPrivate')
    : t('commandPalette.visibilityPublic'),
)

const selectedTagChips = computed(() => {
  if (!selectedTagIds.value.length) return []
  const byId = new Map(tagList.value.map((tag) => [tag.id, tag.name]))
  return selectedTagIds.value.map((id) => ({ id, name: byId.get(id) ?? id }))
})

const handleSearch = () => {
  echoStore.searchValue = searchContent.value
  if (searchingMode.value) {
    refreshForSearch()
    fetchCurrentPage()
  }
}

const handleCancelTapFilter = () => {
  echoStore.isFilteringMode = false
}

const handleClearDateRange = () => {
  resetDateRange()
  refreshEchos()
}

const handleClearVisibility = () => {
  resetVisibilityFilter()
  refreshEchos()
}

const handleRemoveTag = (id: string) => {
  removeSelectedTag(id)
  refreshEchos()
}

watch(searchValue, (value) => {
  if (value !== searchContent.value) {
    searchContent.value = value
  }
})

onMounted(() => {
  if (selectedTagIds.value.length > 0) {
    void ensureTagsLoaded()
  }
})
watch(selectedTagIds, (ids) => {
  if (ids.length > 0 && tagList.value.length === 0) {
    void ensureTagsLoaded()
  }
})
</script>

<style scoped>
.home-filter__search-shell {
  --kbd-reserve: 3.25rem;

  position: relative;
  flex: 1 1 auto;
  min-width: 0;
  padding: 0.3rem;

  padding-right: var(--kbd-reserve);
  border-radius: var(--radius-xs);
  background: var(--filter-search-shell-bg);
  box-shadow: inset 0 0 0 1px var(--color-border-subtle);
}

.home-filter__search-shell--with-chat {
  --kbd-reserve: 5.6rem;
}

.home-filter__search-shell--win {
  --kbd-reserve: 4.5rem;
}

.home-filter__search-shell--win.home-filter__search-shell--with-chat {
  --kbd-reserve: 7rem;
}

.home-filter__shell-keys {
  position: absolute;
  top: 50%;
  right: 0.4rem;
  transform: translateY(-50%);
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.home-filter__kbd-hint {
  box-sizing: border-box;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 1.75rem;
  min-width: 2.25rem;
  padding: 0.25rem 0.5rem;
  font-family: var(--font-family-mono, monospace);
  font-size: 0.72rem;
  font-weight: 500;
  line-height: 1;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
  border-bottom-width: 2px;
  border-radius: 6px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background 0.15s ease,
    transform 0.08s ease;
}

.home-filter__kbd-hint:hover {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
}

.home-filter__kbd-hint:active {
  transform: translateY(1px);
}

.home-filter__chat-trigger {
  box-sizing: border-box;
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 2.25rem;
  height: 1.75rem;
  padding: 0.25rem 0.5rem;
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  border: 1px solid var(--color-border-subtle);
  border-bottom-width: 2px;
  border-radius: 6px;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background 0.15s ease,
    transform 0.08s ease,
    box-shadow 0.15s ease;
}

.home-filter__chat-trigger:hover {
  color: var(--color-text-primary);
  border-color: var(--color-border-strong);
  background: var(--color-bg-surface);
  box-shadow: 0 1px 2px rgb(0 0 0 / 6%);
}

.home-filter__chat-trigger:active {
  transform: translateY(1px);
}

.home-filter__chat-icon {
  width: 1.1rem;
  height: 1.1rem;
}

:deep(.home-filter__chat-trigger path) {
  fill: currentColor;
}

.home-filter__chip {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  padding: 0.125rem 0.5rem;
  color: var(--color-text-muted);
  border: 1px dashed var(--color-border-strong);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition:
    color 0.15s ease,
    background 0.15s ease,
    text-decoration-color 0.15s ease;
}

.home-filter__chip:hover {
  color: var(--color-text-secondary);
  background: var(--color-bg-muted);
  text-decoration: line-through;
}

.home-filter__chip-hash {
  color: inherit;
  opacity: 0.6;
  margin-right: 0.15rem;
  font-weight: 500;
}
</style>
