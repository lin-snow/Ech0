<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="py-4">
    <h2 class="text-[var(--color-text-secondary)] font-bold mb-2">
      {{ t('editor.tagManagerTitle') }}
    </h2>
    <p class="text-xs text-[var(--color-text-muted)] mb-3">{{ t('editor.tagManagerHint') }}</p>

    <form v-if="isLogin" class="flex items-center gap-2 mb-4" @submit.prevent="handleCreateTag">
      <div
        class="flex items-center gap-1 flex-1 border border-dashed border-[var(--color-border-subtle)] rounded-sm px-2 py-1 focus-within:border-[var(--color-text-secondary)] transition-colors"
      >
        <span class="text-[var(--color-text-muted)] select-none">#</span>
        <input
          v-model="newTagName"
          type="text"
          maxlength="50"
          :placeholder="t('editor.createTagPlaceholder')"
          class="flex-1 bg-transparent text-sm outline-none text-[var(--color-text-primary)] placeholder-[var(--color-text-muted)]"
        />
      </div>
      <button
        type="submit"
        :disabled="isCreating || newTagName.trim() === ''"
        class="text-sm font-medium px-3 py-1 rounded-sm border border-[var(--color-border-subtle)] text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:border-[var(--color-text-secondary)] disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
      >
        {{ t('editor.createTagButton') }}
      </button>
    </form>

    <div
      v-if="tagList.length === 0"
      class="text-sm text-[var(--color-text-muted)] py-4 text-center"
    >
      {{ t('editor.tagManagerEmpty') }}
    </div>
    <div v-else ref="listRef" class="flex flex-wrap gap-2">
      <Popover
        v-for="tag in tagList"
        :key="tag.id"
        class="relative overflow-visible"
        v-slot="{ close }"
      >
        <PopoverButton
          class="flex items-center gap-1 border rounded-sm border-[var(--color-border-subtle)] border-dashed py-0.5 px-1 mb-1 outline-none transition-colors duration-150 hover:text-[var(--color-text-secondary)]"
          style="white-space: nowrap"
          @click="resolvePanelSide($event, tag.id)"
        >
          <div
            class="hover:cursor-pointer text-[var(--color-text-muted)] flex items-center justify-start gap-2"
          >
            <div>#</div>
            {{ tag.name }}
          </div>
        </PopoverButton>

        <transition
          enter-active-class="transition duration-200 ease-out"
          enter-from-class="translate-y-1 opacity-0"
          enter-to-class="translate-y-0 opacity-100"
          leave-active-class="transition duration-150 ease-in"
          leave-from-class="translate-y-0 opacity-100"
          leave-to-class="translate-y-1 opacity-0"
        >
          <PopoverPanel :class="getPopoverPanelClass(tag.id)">
            <div class="overflow-hidden rounded-md shadow-sm ring-black/5">
              <div
                class="relative flex justify-around gap-2 bg-[var(--color-bg-surface)] p-1 text-[var(--color-text-secondary)]"
              >
                <button
                  @click="
                    () => {
                      handleFilterByTag(tag)
                      close()
                    }
                  "
                  v-tooltip="t('editor.filterByTag')"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:text-[var(--color-text-primary)] focus:outline-none focus-visible:ring focus-visible:ring-[var(--input-focus-color-border-subtle)]"
                >
                  <Filter class="w-5 h-5" />
                </button>
                <div v-if="isLogin" class="w-px bg-[var(--color-bg-muted)]"></div>
                <button
                  v-if="isLogin"
                  @click="
                    () => {
                      handleDeleteTag(tag.id)
                      close()
                    }
                  "
                  v-tooltip="t('editor.deleteTag')"
                  class="flex items-center justify-center rounded-md p-1 transition duration-150 ease-in-out hover:text-[var(--color-danger)] focus:outline-none focus-visible:ring focus-visible:ring-[var(--input-focus-color-border-subtle)]"
                >
                  <Trashbin class="w-5 h-5" />
                </button>
              </div>
            </div>
          </PopoverPanel>
        </transition>
      </Popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useEchoStore, useUserStore } from '@/stores'
import { fetchDeleteTagById } from '@/service/api'
import { storeToRefs } from 'pinia'
import { useBaseDialog } from '@/composables/useBaseDialog'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import Trashbin from '@/components/icons/trashbin.vue'
import Filter from '@/components/icons/filter.vue'
import { useI18n } from 'vue-i18n'
import { useRouter } from 'vue-router'
import { theToast } from '@/utils/toast'

const echoStore = useEchoStore()
const userStore = useUserStore()
const { tagList } = storeToRefs(echoStore)
const { isLogin } = storeToRefs(userStore)
const { t } = useI18n()
const router = useRouter()

const newTagName = ref('')
const isCreating = ref(false)

const handleCreateTag = async () => {
  const name = newTagName.value.trim().replace(/^#+/, '').trim()
  if (!name) {
    theToast.warning(String(t('editor.createTagEmpty')))
    return
  }
  isCreating.value = true
  try {
    const tag = await echoStore.createTag(name)
    if (tag) {
      newTagName.value = ''
      theToast.success(String(t('editor.createTagSuccess')))
    }
  } finally {
    isCreating.value = false
  }
}

onMounted(() => {
  echoStore.ensureTagsLoaded()
})

const { openConfirm } = useBaseDialog()

const listRef = ref<HTMLElement | null>(null)
const panelSides = reactive<Record<string, 'left' | 'center' | 'right'>>({})

const resolvePanelSide = (event: MouseEvent, tagId: string) => {
  const button = event.currentTarget as HTMLElement | null
  const container = listRef.value
  if (!button || !container) {
    panelSides[tagId] = 'center'
    return
  }
  const buttonRect = button.getBoundingClientRect()
  const containerRect = container.getBoundingClientRect()
  const halfPanel = 96 / 2
  const buttonCenter = buttonRect.left + buttonRect.width / 2
  if (buttonCenter - halfPanel < containerRect.left) {
    panelSides[tagId] = 'left'
  } else if (buttonCenter + halfPanel > containerRect.right) {
    panelSides[tagId] = 'right'
  } else {
    panelSides[tagId] = 'center'
  }
}

const getPopoverPanelClass = (tagId: string) => {
  const side = panelSides[tagId] ?? 'center'
  if (side === 'left') return 'absolute left-0 z-[60] mt-1'
  if (side === 'right') return 'absolute right-0 z-[60] mt-1'
  return 'absolute left-1/2 z-[60] mt-1 -translate-x-1/2 transform'
}

const handleFilterByTag = (tag: App.Api.Ech0.Tag) => {
  if (!tag) return

  echoStore.filteredTag = tag
  echoStore.isFilteringMode = true
  echoStore.refreshEchos()
  router.push({ name: 'home' })
}

const handleDeleteTag = (tagId: string) => {
  openConfirm({
    title: String(t('editor.deleteTagConfirmTitle')),
    description: String(t('editor.deleteTagConfirmDesc')),
    onConfirm: () => {
      fetchDeleteTagById(tagId).then((res) => {
        if (res.code === 1) {
          echoStore.getTags()
        }
      })
    },
  })
}
</script>

<style scoped></style>
