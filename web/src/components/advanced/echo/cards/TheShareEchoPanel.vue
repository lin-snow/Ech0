<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <Popover class="relative flex items-center">
    <PopoverButton as="template">
      <button
        v-tooltip="t('echoDetail.share')"
        type="button"
        :aria-label="t('echoDetail.share')"
        :class="[
          'flex items-center justify-center transform transition-transform duration-150 outline-none',
          isShareAnimating ? 'scale-160' : 'scale-100',
        ]"
        @click="triggerAnimation"
      >
        <Share class="w-4 h-4" />
      </button>
    </PopoverButton>

    <transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="opacity-0 scale-95"
      enter-to-class="opacity-100 scale-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100 scale-100"
      leave-to-class="opacity-0 scale-95"
    >
      <PopoverPanel
        class="absolute right-0 z-40 mt-2 w-60 origin-top-right rounded-[var(--radius-md)] bg-[var(--color-bg-muted)] ring-1 ring-[var(--color-border-subtle)] shadow-[var(--shadow-md)] p-2"
      >
        <div class="flex items-center gap-1.5 px-1 pt-0.5 pb-1.5 text-[var(--color-text-muted)]">
          <Share class="w-3 h-3" />
          <span class="text-[10px] font-semibold tracking-[0.08em] uppercase">
            {{ t('echoDetail.sharePanelTitle') }}
          </span>
        </div>

        <div class="grid gap-1.5" :class="canSystemShare ? 'grid-cols-2' : 'grid-cols-1'">
          <button v-if="canSystemShare" type="button" class="share-tile" @click="handleSystemShare">
            <Share class="w-4 h-4" />
            <span>{{ t('echoDetail.shareSystem') }}</span>
          </button>
          <button type="button" class="share-tile" @click="handleCopyMarkdown">
            <Markdown class="w-4 h-4" />
            <span>{{ t('echoDetail.shareCopyMarkdown') }}</span>
          </button>
        </div>

        <div
          class="mt-1.5 flex items-center gap-2 h-8 rounded-[var(--radius-sm)] bg-[var(--color-bg-surface)] ring-1 ring-[var(--color-border-subtle)] px-2.5"
        >
          <span
            class="flex-1 truncate text-[11px] text-[var(--color-text-muted)]"
            :title="shareUrl"
          >
            {{ shareUrl }}
          </span>
          <button
            type="button"
            v-tooltip="t('echoDetail.shareCopyLink')"
            :aria-label="t('echoDetail.shareCopyLink')"
            class="shrink-0 text-[var(--color-text-muted)] transition-colors duration-150 hover:text-[var(--color-text-primary)] focus:outline-none"
            @click="handleCopyLink"
          >
            <Clipboard v-if="linkJustCopied" class="w-3.5 h-3.5" />
            <Link v-else class="w-3.5 h-3.5" />
          </button>
        </div>
      </PopoverPanel>
    </transition>
  </Popover>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import { useI18n } from 'vue-i18n'
import Share from '@/components/icons/share.vue'
import Link from '@/components/icons/link.vue'
import Markdown from '@/components/icons/markdown.vue'
import Clipboard from '@/components/icons/clipboard.vue'
import { theToast } from '@/utils/toast'

const props = defineProps<{
  echoId: string
  echoContent?: string
}>()

const { t } = useI18n()

const isShareAnimating = ref(false)
const linkJustCopied = ref(false)

const shareUrl = computed(() => `${window.location.origin}/echo/${props.echoId}`)

const canSystemShare = computed(() => typeof navigator !== 'undefined' && !!navigator.share)

const triggerAnimation = () => {
  isShareAnimating.value = true
  setTimeout(() => {
    isShareAnimating.value = false
  }, 250)
}

const writeToClipboard = async (text: string) => {
  await navigator.clipboard.writeText(text)
}

const handleCopyLink = async () => {
  try {
    await writeToClipboard(shareUrl.value)
    linkJustCopied.value = true
    theToast.info(String(t('echoDetail.copied')))
    setTimeout(() => {
      linkJustCopied.value = false
    }, 1500)
  } catch {
    theToast.error(String(t('echoDetail.shareCopyFailed')))
  }
}

const handleCopyMarkdown = async () => {
  const body = (props.echoContent ?? '').trim()
  const quoted = body
    ? body
        .split('\n')
        .map((line) => `> ${line}`)
        .join('\n')
    : shareUrl.value
  try {
    await writeToClipboard(quoted)
    theToast.info(String(t('echoDetail.shareMarkdownCopied')))
  } catch {
    theToast.error(String(t('echoDetail.shareCopyFailed')))
  }
}

const handleSystemShare = async () => {
  if (!canSystemShare.value) return
  try {
    await navigator.share({
      title: t('echoDetail.shareSuffix'),
      text: props.echoContent ?? '',
      url: shareUrl.value,
    })
  } catch {}
}
</script>

<style scoped>
.share-tile {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: 54px;
  padding: 8px 6px;
  border-radius: var(--radius-sm);
  background: var(--color-bg-surface);
  box-shadow: inset 0 0 0 1px var(--color-border-subtle);
  color: var(--color-text-secondary);
  font-size: 11px;
  font-weight: 500;
  line-height: 1.3;
  text-align: center;
  transition:
    color 150ms ease,
    background-color 150ms ease,
    box-shadow 150ms ease,
    transform 150ms ease;
}

.share-tile:hover {
  color: var(--color-text-primary);
  background: var(--color-bg-canvas);
  box-shadow: inset 0 0 0 1px var(--color-border-strong);
}

.share-tile:active {
  transform: scale(0.97);
}

.share-tile:focus-visible {
  outline: none;
  box-shadow: inset 0 0 0 1.5px var(--color-border-strong);
}
</style>
