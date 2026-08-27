<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="editor-actions">
    <div class="editor-actions__left">
      <BaseButton
        :icon="currentMode === Mode.ECH0 ? Advance : Back"
        @click="handleChangeMode"
        :class="['w-8 h-8 sm:w-9 sm:h-9 rounded-xs'].join(' ')"
        :tooltip="currentMode === Mode.ECH0 ? t('editor.more') : t('editor.backToEditor')"
      />
      <BaseButton
        v-if="currentMode === Mode.ECH0"
        :icon="AttachmentIcon"
        @click="handleAddMediaMode"
        class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs"
        :tooltip="t('editor.addImage')"
      />
      <div v-if="currentMode === Mode.ECH0" class="editor-actions__tag">
        <Popover>
          <PopoverButton
            :aria-label="tagTriggerTooltip"
            class="cursor-pointer p-1.5 w-8 h-8 sm:w-9 sm:h-9 rounded-xs ring-inset ring-1 ring-[var(--btn-ring-color)] text-[var(--btn-text-color)] outline-none shadow-[var(--btn-shadow)] bg-[var(--btn-bg-color)] hover:bg-[var(--btn-hover-bg-color)] hover:ring-[var(--btn-hover-border-color)] focus-visible:ring-2 focus-visible:ring-[var(--btn-focus-ring-color)] transition-colors duration-200 relative inline-flex items-center justify-center"
          >
            <TagSetting class="w-full h-full" />
          </PopoverButton>

          <transition
            enter-active-class="transition duration-150 ease-out"
            enter-from-class="opacity-0 -translate-y-1"
            enter-to-class="opacity-100 translate-y-0"
            leave-active-class="transition duration-100 ease-in"
            leave-from-class="opacity-100 translate-y-0"
            leave-to-class="opacity-0 -translate-y-1"
          >
            <PopoverPanel class="editor-actions__tag-panel">
              <div v-if="tagOptions.length === 0" class="editor-actions__tag-empty">
                <p class="editor-actions__tag-empty-text">{{ t('editor.tagPickerEmpty') }}</p>
                <button
                  type="button"
                  class="editor-actions__tag-empty-link"
                  @click="goToTagManager"
                >
                  {{ t('editor.tagPickerGoToManage') }} →
                </button>
              </div>
              <div v-else class="editor-actions__tag-chip-list">
                <button
                  v-for="name in tagOptions"
                  :key="name"
                  type="button"
                  class="editor-actions__tag-chip"
                  :class="{
                    'editor-actions__tag-chip--selected': tagToAdd.includes(name),
                    'editor-actions__tag-chip--disabled': isTagChipDisabled(name),
                  }"
                  @click="toggleTag(name)"
                >
                  #{{ name }}
                </button>
              </div>
            </PopoverPanel>
          </transition>
        </Popover>
      </div>
    </div>

    <div class="editor-actions__right">
      <div v-if="hasContent || hasFile || hasExtension" class="relative group">
        <Info class="w-6 h-6 text-[var(--color-text-muted)] hover:text-[var(--color-text-muted)]" />
        <div class="editor-actions__info-pop">
          <div v-if="infoTooltipLines.length > 0">
            <div class="mb-1 font-medium text-[var(--color-text-secondary)]">
              {{ t('editor.addedLabel') }}
            </div>
            <div
              v-for="line in infoTooltipLines"
              :key="line.label"
              class="flex items-center gap-1 text-[var(--color-text-muted)]"
            >
              <component v-if="line.icon" :is="line.icon" class="w-3.5 h-3.5" />
              <span>{{ line.label }}</span>
            </div>
          </div>
          <div v-else class="text-[var(--color-text-muted)]">{{ t('editor.noContentAdded') }}</div>
        </div>
      </div>

      <div v-if="currentMode !== Mode.Panel && isUpdateMode === true">
        <BaseButton
          :icon="ExitUpdate"
          @click="handleExitUpdateMode"
          class="w-8 h-8 sm:w-9 sm:h-9 rounded-xs editor-actions__cta"
          :tooltip="t('editor.exitUpdateMode')"
        />
      </div>
      <Popover
        v-if="currentMode !== Mode.Panel && currentMode !== Mode.TagManage"
        class="editor-actions__publish"
      >
        <PopoverButton
          :aria-label="publishTriggerTooltip"
          :disabled="isSubmitting"
          :class="[
            'cursor-pointer p-1.5 w-8 h-8 sm:w-9 sm:h-9 rounded-xs ring-inset ring-1 ring-[var(--btn-ring-color)] text-[var(--btn-text-color)] outline-none shadow-[var(--btn-shadow)] bg-[var(--btn-bg-color)] transition-colors duration-200 relative inline-flex items-center justify-center editor-actions__cta',
            isSubmitting
              ? 'cursor-not-allowed opacity-70'
              : 'hover:bg-[var(--btn-hover-bg-color)] hover:ring-[var(--btn-hover-border-color)] focus-visible:ring-2 focus-visible:ring-[var(--btn-focus-ring-color)]',
          ]"
        >
          <component :is="isUpdateMode ? Update : Publish" class="w-full h-full" />
        </PopoverButton>

        <transition
          enter-active-class="transition duration-150 ease-out"
          enter-from-class="opacity-0 -translate-y-1"
          enter-to-class="opacity-100 translate-y-0"
          leave-active-class="transition duration-100 ease-in"
          leave-from-class="opacity-100 translate-y-0"
          leave-to-class="opacity-0 -translate-y-1"
        >
          <PopoverPanel v-slot="{ close }" class="editor-actions__publish-panel">
            <button
              type="button"
              class="editor-actions__publish-option"
              @click="handlePublishWithPrivacy(false, close)"
            >
              <Public class="editor-actions__publish-option-icon" />
              <span>{{ t('editor.publishEchoPublic') }}</span>
            </button>
            <button
              type="button"
              class="editor-actions__publish-option"
              @click="handlePublishWithPrivacy(true, close)"
            >
              <Private class="editor-actions__publish-option-icon" />
              <span>{{ t('editor.publishEchoPrivate') }}</span>
            </button>
          </PopoverPanel>
        </transition>
      </Popover>
    </div>
  </div>
</template>

<script setup lang="ts">
import Advance from '@/components/icons/advance.vue'
import AttachmentIcon from '@/components/icons/attachment.vue'
import ImageIcon from '@/components/icons/image.vue'
import Public from '@/components/icons/public.vue'
import Private from '@/components/icons/private.vue'
import Publish from '@/components/icons/publish.vue'
import Update from '@/components/icons/update.vue'
import ExitUpdate from '@/components/icons/exitupdate.vue'
import Back from '@/components/icons/back.vue'
import Info from '@/components/icons/info.vue'
import Write from '@/components/icons/write.vue'
import Music from '@/components/icons/music.vue'
import Video from '@/components/icons/video.vue'
import GithubProj from '@/components/icons/githubproj.vue'
import Website from '@/components/icons/website.vue'
import MapPin from '@/components/icons/mappin.vue'
import XBrand from '@/components/icons/tweet.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import TagSetting from '@/components/icons/tagsetting.vue'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'
import { Mode, ExtensionType } from '@/enums/enums'
import { FILE_STORAGE_TYPE } from '@/constants/file'
import { storeToRefs } from 'pinia'
import { useEditorStore, useEchoStore } from '@/stores'
import { theToast } from '@/utils/toast'
import { localStg } from '@/utils/storage'
import { computed, onMounted, type Component } from 'vue'
import { useI18n } from 'vue-i18n'

const editorStore = useEditorStore()
const {
  currentMode,
  isUpdateMode,
  echoToAdd,
  fileToAdd,
  tagToAdd,
  hasContent,
  hasFile,
  hasExtension,
  extensionToAdd,
  isSubmitting,
} = storeToRefs(editorStore)
const echoStore = useEchoStore()
const { tagOptions } = storeToRefs(echoStore)
const { t } = useI18n()

onMounted(() => {
  echoStore.ensureTagsLoaded()
})

type TooltipLine = { label: string; icon?: Component }

const infoTooltipLines = computed<TooltipLine[]>(() => {
  const extType = extensionToAdd.value.extension_type || echoToAdd.value.extension?.type
  const extMap: Record<ExtensionType, { label: string; icon: Component }> = {
    [ExtensionType.MUSIC]: { label: String(t('editor.extMusic')), icon: Music },
    [ExtensionType.VIDEO]: { label: String(t('editor.extVideo')), icon: Video },
    [ExtensionType.GITHUBPROJ]: { label: String(t('editor.extGithubProject')), icon: GithubProj },
    [ExtensionType.WEBSITE]: { label: String(t('editor.extWebsiteLink')), icon: Website },
    [ExtensionType.LOCATION]: { label: String(t('editor.extLocation')), icon: MapPin },
    [ExtensionType.TWEET]: { label: String(t('editor.extTweet')), icon: XBrand },
  }

  const parts: TooltipLine[] = []
  if (hasContent.value) parts.push({ label: String(t('editor.extText')), icon: Write })
  if (hasFile.value) parts.push({ label: String(t('editor.extImage')), icon: ImageIcon })
  if (hasExtension.value)
    parts.push({
      label:
        extType && extMap[extType as ExtensionType]?.label
          ? extMap[extType as ExtensionType].label
          : String(t('editor.extGeneric')),
      icon:
        extType && extMap[extType as ExtensionType]?.icon
          ? extMap[extType as ExtensionType].icon
          : undefined,
    })

  return parts
})

const handlePublishWithPrivacy = (isPrivate: boolean, close: () => void) => {
  if (isSubmitting.value) return
  echoToAdd.value.private = isPrivate
  close()
  editorStore.handleAddOrUpdate()
}

const publishTriggerTooltip = computed(() =>
  isUpdateMode.value ? t('editor.updateEcho') : t('editor.publishEcho'),
)

const handleChangeMode = () => {
  editorStore.toggleMode()
}

const handleAddMediaMode = () => {
  fileToAdd.value.storage_type = FILE_STORAGE_TYPE.LOCAL

  const rememberedSource = localStg.getItem<App.Api.File.StorageType>('file_storage_type')
  if (rememberedSource) {
    fileToAdd.value.storage_type = rememberedSource
  }

  editorStore.setMode(Mode.Media)
}

const handleExitUpdateMode = () => {
  editorStore.handleExitUpdateMode()
}

const MAX_TAGS = 3

const tagTriggerTooltip = computed(() =>
  tagToAdd.value.length > 0
    ? `${t('editor.tagPickerTooltip')} · ${tagToAdd.value.length}/${MAX_TAGS}`
    : t('editor.tagPickerTooltip'),
)

const isTagChipDisabled = (name: string) =>
  tagToAdd.value.length >= MAX_TAGS && !tagToAdd.value.includes(name)

const toggleTag = (name: string) => {
  const next = [...tagToAdd.value]
  const idx = next.indexOf(name)
  if (idx >= 0) {
    next.splice(idx, 1)
  } else {
    if (next.length >= MAX_TAGS) {
      theToast.warning(String(t('editor.tagPickerLimit', { max: MAX_TAGS })))
      return
    }
    next.push(name)
  }
  tagToAdd.value = next
}

const goToTagManager = () => {
  editorStore.setMode(Mode.TagManage)
}
</script>

<style scoped>
.editor-actions {
  --btn-bg-color: var(--md-editor-mini-btn-bg);
  --btn-ring-color: var(--md-editor-actions-ring-color);
  --btn-hover-bg-color: var(--md-editor-actions-hover-bg);
  --btn-hover-border-color: var(--md-editor-actions-hover-border);

  position: relative;
  z-index: 50;
  display: flex;
  justify-content: space-between;
  flex-wrap: nowrap;
  gap: 0.6rem;
  align-items: center;
  padding: 0.3rem 0.35rem 0.1rem;
}

.editor-actions__left {
  min-width: 0;
  flex: 1 1 auto;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: nowrap;
  overflow: visible;
}

.editor-actions__tag {
  position: relative;
  flex: 0 0 auto;
  isolation: isolate;
}

.editor-actions__tag-panel {
  position: absolute;
  z-index: 100;
  top: calc(100% + 0.4rem);
  left: 0;
  width: 16rem;
  max-height: 16rem;
  overflow-y: auto;
  padding: 0.55rem;
  border-radius: var(--radius-xs);
  border: 1px solid var(--color-border-subtle);
  background: var(--md-editor-mini-bg);
  box-shadow: var(--md-editor-mini-shell-shadow);
}

.editor-actions__tag-empty {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  padding: 0.5rem;
  text-align: center;
}

.editor-actions__tag-empty-text {
  font-size: 0.8rem;
  color: var(--color-text-muted);
  margin: 0;
}

.editor-actions__tag-empty-link {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--color-accent);
  background: transparent;
  border: none;
  padding: 0;
  cursor: pointer;
}

.editor-actions__tag-empty-link:hover {
  text-decoration: underline;
}

.editor-actions__tag-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.editor-actions__tag-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.18rem 0.55rem;
  border-radius: var(--radius-sm);
  border: 1px dashed var(--color-border-subtle);
  background: transparent;
  color: var(--color-text-muted);
  font-size: 0.75rem;
  line-height: 1.3;
  cursor: pointer;
  transition:
    color 0.15s ease,
    border-color 0.15s ease,
    background-color 0.15s ease;
}

.editor-actions__tag-chip:hover {
  color: var(--color-text-primary);
  border-color: var(--color-text-secondary);
}

.editor-actions__tag-chip--selected {
  color: var(--color-bg-canvas);
  background: var(--color-text-primary);
  border-style: solid;
  border-color: var(--color-text-primary);
}

.editor-actions__tag-chip--selected:hover {
  color: var(--color-bg-canvas);
}

.editor-actions__tag-chip--disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.editor-actions__tag-chip--disabled:hover {
  color: var(--color-text-muted);
  border-color: var(--color-border-subtle);
  background: transparent;
}

.editor-actions__right {
  flex: 0 0 auto;
  display: flex;
  align-items: center;
  gap: 0.45rem;
  white-space: nowrap;
}

.editor-actions__info-pop {
  position: absolute;
  right: 0;
  top: 100%;
  z-index: 10;
  margin-top: 0.5rem;
  white-space: nowrap;
  border-radius: var(--radius-xs);
  border: 1px dashed var(--md-editor-mini-btn-border);
  background: var(--md-editor-mini-bg);
  padding: 0.4rem 0.55rem;
  font-size: 0.75rem;
  box-shadow: var(--md-editor-mini-shell-shadow);
  opacity: 0;
  transform: translateY(0.25rem) scale(0.96);
  pointer-events: none;
  transition: all 0.2s ease-out;
}

.group:hover .editor-actions__info-pop {
  opacity: 1;
  transform: translateY(0) scale(1);
  pointer-events: auto;
}

.editor-actions__cta {
  box-shadow: var(--md-editor-mini-shell-shadow);
}

.editor-actions__publish {
  position: relative;
  flex: 0 0 auto;
  isolation: isolate;
}

.editor-actions__publish-panel {
  position: absolute;
  z-index: 100;
  top: calc(100% + 0.4rem);
  right: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  min-width: 11rem;
  padding: 0.4rem;
  border-radius: var(--radius-xs);
  border: 1px solid var(--color-border-subtle);
  background: var(--md-editor-mini-bg);
  box-shadow: var(--md-editor-mini-shell-shadow);
}

.editor-actions__publish-option {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.4rem 0.6rem;
  border-radius: var(--radius-xs);
  border: none;
  background: transparent;
  color: var(--color-text-primary);
  font-size: 0.8rem;
  line-height: 1.3;
  text-align: left;
  cursor: pointer;
  transition:
    background-color 0.15s ease,
    color 0.15s ease;
}

.editor-actions__publish-option:hover {
  background: var(--md-editor-actions-hover-bg);
  color: var(--color-text-primary);
}

.editor-actions__publish-option-icon {
  width: 1rem;
  height: 1rem;
  flex: 0 0 auto;
}

@media (width <= 639.98px) {
  .editor-actions {
    gap: 0.45rem;
  }

  .editor-actions__tag {
    position: static;
  }

  .editor-actions__tag-panel {
    left: 0.35rem;
    right: 0.35rem;
    width: auto;
    max-width: none;
  }

  .editor-actions__publish-panel {
    right: 0.35rem;
  }
}
</style>
