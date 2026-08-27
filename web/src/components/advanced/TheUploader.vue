<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->
<!-- Copyright (C) 2025-2026 lin-snow -->
<template>
  <div class="rounded-md overflow-hidden">
    <button
      type="button"
      class="w-full flex flex-col items-center justify-center gap-1 py-5 px-3 rounded-md border-1 border-dashed transition-colors cursor-pointer text-center select-none"
      :class="[
        isDropOverZone
          ? 'border-[var(--color-accent)] bg-[var(--color-accent-soft)]'
          : 'border-[var(--md-editor-mini-border)] bg-[var(--color-bg-surface-alt,transparent)] hover:border-[var(--color-border-strong)]',
      ]"
      @click="openFilePicker"
      @dragover.prevent="onZoneDragOver"
      @dragenter.prevent="onZoneDragOver"
      @dragleave.prevent="isDropOverZone = false"
      @drop.prevent="handleDrop"
    >
      <span class="text-[var(--color-text-primary)] font-medium">
        <slot name="drop-title">{{ dropTitle }}</slot>
      </span>
      <span class="text-xs text-[var(--color-text-muted)]">
        <slot name="drop-hint" :max="maxFiles" :max-file-size="maxFileSize">
          {{ dropHintText }}
        </slot>
      </span>
    </button>

    <input
      ref="fileInputRef"
      type="file"
      multiple
      :accept="acceptAttr"
      class="hidden"
      @change="handleInputChange"
    />

    <ul v-if="items.length > 0" class="mt-2 flex flex-col gap-1.5">
      <li
        v-for="item in items"
        :key="item.id"
        class="flex items-center gap-2 px-2 py-1.5 rounded-md ring-1 ring-inset ring-[var(--md-editor-mini-border)] bg-[var(--color-bg-surface-alt,transparent)] transition-colors"
        :class="[
          isReorderable(item) ? 'cursor-grab active:cursor-grabbing' : '',
          dragOverId === item.id && draggingId !== item.id
            ? 'ring-[var(--color-accent)] bg-[var(--color-accent-soft)]'
            : '',
          draggingId === item.id ? 'opacity-50' : '',
        ]"
        :draggable="isReorderable(item)"
        @dragstart="(e) => onItemDragStart(item, e)"
        @dragenter.prevent="onItemDragEnter(item)"
        @dragover.prevent="onItemDragOver(item, $event)"
        @dragleave="onItemDragLeave(item)"
        @drop.prevent="onItemDrop(item)"
        @dragend="onItemDragEnd"
      >
        <img
          v-if="isImageItem(item)"
          :src="item.preview"
          alt=""
          class="w-10 h-10 object-cover rounded-sm shrink-0 bg-[var(--color-bg-muted)]"
        />
        <span
          v-else
          class="w-10 h-10 flex items-center justify-center rounded-sm shrink-0 bg-[var(--color-bg-muted)] text-[var(--color-text-muted)]"
        >
          <component :is="isVideoItem(item) ? VideoIcon : AudioIcon" class="w-5 h-5" />
        </span>

        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between gap-2">
            <span
              class="flex-1 min-w-0 text-sm truncate text-[var(--color-text-primary)]"
              :title="item.file.name"
            >
              {{ item.file.name }}
            </span>
            <div class="flex items-center gap-1.5 shrink-0">
              <span class="text-xs text-[var(--color-text-muted)]">
                {{ formatBytes(item.file.size) }}
              </span>
              <template v-for="badge in [compressionBadge(item)]" :key="badge?.label">
                <span
                  v-if="badge"
                  class="text-xs px-1 rounded"
                  :class="
                    badge.tone === 'savings'
                      ? 'bg-[var(--color-accent-soft)] text-[var(--color-accent)]'
                      : 'bg-[var(--color-border-subtle)] text-[var(--color-text-muted)]'
                  "
                  :title="badge.tooltip"
                >
                  {{ badge.label }}
                </span>
              </template>
              <span class="text-xs" :class="statusColorClass(item.status)">
                {{ statusLabel(item) }}
              </span>
            </div>
          </div>

          <div
            v-if="
              item.status === UPLOAD_STATUS.UPLOADING || item.status === UPLOAD_STATUS.COMPRESSING
            "
            class="mt-1 h-1 rounded-full bg-[var(--color-border-subtle,#e5e7eb)] overflow-hidden"
          >
            <div
              class="h-full bg-[var(--color-accent)] transition-all"
              :style="{
                width: (item.status === UPLOAD_STATUS.COMPRESSING ? 100 : item.progress) + '%',
              }"
            />
          </div>
          <div
            v-else-if="item.status === UPLOAD_STATUS.ERROR && item.error"
            class="mt-0.5 text-xs text-[var(--color-danger,#dc2626)] truncate"
            :title="item.error"
          >
            {{ item.error }}
          </div>
        </div>

        <div class="flex items-center gap-1 shrink-0">
          <button
            v-if="item.status === UPLOAD_STATUS.ERROR || item.status === UPLOAD_STATUS.CANCELLED"
            type="button"
            class="text-xs px-2 py-0.5 rounded cursor-pointer text-[var(--color-accent)] hover:bg-[var(--color-accent-soft)]"
            @click.stop="retry(item.id)"
          >
            {{ t('uploader.retry') }}
          </button>
          <button
            type="button"
            class="text-xs px-2 py-0.5 rounded cursor-pointer text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-border-subtle)]"
            @click.stop="handleRemoveItem(item)"
            :title="t('uploader.removeItem')"
          >
            ✕
          </button>
        </div>
      </li>
    </ul>

    <div
      v-if="isUploading"
      class="mt-2 flex items-center justify-between text-xs text-[var(--color-text-muted)]"
    >
      <span>{{ t('uploader.uploadingProgress', { active: totalActive }) }}</span>
      <button
        type="button"
        class="cursor-pointer hover:text-[var(--color-danger,#dc2626)]"
        @click="cancelAll"
      >
        {{ t('uploader.cancelAll') }}
      </button>
    </div>

    <div
      v-else-if="hasReorderable"
      class="mt-1.5 text-xs text-[var(--color-text-muted)] text-center"
    >
      {{ t('uploader.reorderHint') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, toRef, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { useEditorStore } from '@/stores'
import { useUpload, UPLOAD_STATUS, type QueueItem, type UploadStatus } from '@/lib/file/useUpload'
import { formatBytes } from '@/utils/file'
import { theToast } from '@/utils/toast'
import { deleteFileById } from '@/lib/file'
import AudioIcon from '@/components/icons/audio.vue'
import VideoIcon from '@/components/icons/videomedia.vue'
import { FILE_CATEGORY, FILE_STORAGE_TYPE } from '@/constants/file'

const props = withDefaults(
  defineProps<{
    fileStorageType: App.Api.File.StorageType
    EnableCompressor: boolean
    fileCategory?: App.Api.File.Category
    allowedFileTypes?: string[]
    maxFiles?: number
    maxFileSize?: number
  }>(),
  {
    maxFiles: 6,
    maxFileSize: undefined,
  },
)

const editorStore = useEditorStore()
const { t } = useI18n()

const fileInputRef = ref<HTMLInputElement | null>(null)
const isDropOverZone = ref(false)

const allowedTypes = computed(() =>
  props.allowedFileTypes?.length ? props.allowedFileTypes : ['image/*'],
)
const acceptAttr = computed(() => allowedTypes.value.join(','))

const maxFiles = computed(() => props.maxFiles)
const maxFileSize = computed(() => props.maxFileSize)

const isSingleMedia = computed(
  () => props.fileCategory === FILE_CATEGORY.AUDIO || props.fileCategory === FILE_CATEGORY.VIDEO,
)
const dropTitle = computed(() => {
  switch (props.fileCategory) {
    case FILE_CATEGORY.AUDIO:
      return t('uploader.dropHereAudio')
    case FILE_CATEGORY.VIDEO:
      return t('uploader.dropHereVideo')
    default:
      return t('uploader.dropHere')
  }
})
const dropHintText = computed(() =>
  isSingleMedia.value
    ? t('uploader.dropHintSingle', { max: maxFiles.value })
    : t('uploader.dropHint', { max: maxFiles.value }),
)

const { items, isUploading, totalActive, addFiles, retry, remove, moveItem, cancelAll } = useUpload(
  {
    storageType: toRef(props, 'fileStorageType'),
    enableCompressor: toRef(props, 'EnableCompressor'),
    category: props.fileCategory,
    allowedTypes: allowedTypes.value,
    maxFiles: maxFiles.value,
    maxFileSize: maxFileSize.value,
    concurrency: 3,
    onAllComplete: (results) => {
      editorStore.handleUppyUploaded(results)
    },
  },
)

watch(
  isUploading,
  (v) => {
    editorStore.fileUploading = v
  },
  { immediate: true },
)

async function handleRemoveItem(item: QueueItem) {
  const fileId = item.result?.id
  const storageType = item.result?.storage_type
  remove(item.id)
  if (!fileId) return

  if (storageType === FILE_STORAGE_TYPE.LOCAL || storageType === FILE_STORAGE_TYPE.OBJECT) {
    try {
      await deleteFileById(fileId)
    } catch {
      theToast.error(String(t('editor.removeImageRemoteFailed')))
    }
  }

  const idx = editorStore.filesToAdd.findIndex((f) => f.id === fileId)
  if (idx >= 0) editorStore.removeFileAt(idx)

  if (editorStore.isUpdateMode) {
    await editorStore.handleAddOrUpdateEcho(true)
  }
}

watch(
  () => editorStore.filesToAdd.map((f) => f.id ?? '').join('|'),
  () => {
    const present = new Set(editorStore.filesToAdd.map((f) => f.id).filter(Boolean))
    for (const item of [...items.value]) {
      if (
        item.status === UPLOAD_STATUS.SUCCESS &&
        item.result?.id &&
        !present.has(item.result.id)
      ) {
        remove(item.id)
      }
    }
  },
)

const hasReorderable = computed(
  () => items.value.filter((i) => i.status === UPLOAD_STATUS.SUCCESS).length >= 2,
)

function isReorderable(item: QueueItem): boolean {
  return item.status === UPLOAD_STATUS.SUCCESS
}

function isImageItem(item: QueueItem): boolean {
  return item.file.type.startsWith('image/')
}

function isVideoItem(item: QueueItem): boolean {
  return item.file.type.startsWith('video/')
}

function openFilePicker() {
  fileInputRef.value?.click()
}

function onZoneDragOver(e: DragEvent) {
  if (draggingId.value) return
  if (e.dataTransfer?.types?.includes('Files')) {
    isDropOverZone.value = true
  }
}

function handleInputChange(e: Event) {
  const input = e.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    addFiles(Array.from(input.files))
    input.value = ''
  }
}

function handleDrop(e: DragEvent) {
  isDropOverZone.value = false
  if (draggingId.value) return
  const dt = e.dataTransfer
  if (!dt) return
  const files: File[] = []
  if (dt.items) {
    for (const item of dt.items) {
      if (item.kind === 'file') {
        const f = item.getAsFile()
        if (f) files.push(f)
      }
    }
  } else if (dt.files) {
    for (const f of dt.files) files.push(f)
  }
  if (files.length > 0) addFiles(files)
}

function handlePaste(e: ClipboardEvent) {
  if (!e.clipboardData) return
  const files: File[] = []
  for (const item of e.clipboardData.items) {
    if (item.kind === 'file' && item.type.startsWith('image/')) {
      const f = item.getAsFile()
      if (f) {
        files.push(new File([f], f.name, { type: f.type, lastModified: f.lastModified }))
      }
    }
  }
  if (files.length > 0) addFiles(files)
}

onMounted(() => {
  document.addEventListener('paste', handlePaste)
})
onBeforeUnmount(() => {
  document.removeEventListener('paste', handlePaste)
  cancelAll()
  editorStore.fileUploading = false
})

const draggingId = ref<string | null>(null)
const dragOverId = ref<string | null>(null)

function onItemDragStart(item: QueueItem, e: DragEvent) {
  if (!isReorderable(item)) {
    e.preventDefault()
    return
  }
  draggingId.value = item.id
  if (e.dataTransfer) {
    e.dataTransfer.effectAllowed = 'move'
    e.dataTransfer.setData('text/plain', item.id)
  }
}

function onItemDragEnter(item: QueueItem) {
  if (!draggingId.value || !isReorderable(item)) return
  dragOverId.value = item.id
}

function onItemDragOver(item: QueueItem, e: DragEvent) {
  if (!draggingId.value || !isReorderable(item)) return
  if (e.dataTransfer) e.dataTransfer.dropEffect = 'move'
  dragOverId.value = item.id
}

function onItemDragLeave(item: QueueItem) {
  if (dragOverId.value === item.id) dragOverId.value = null
}

function onItemDrop(target: QueueItem) {
  const fromId = draggingId.value
  draggingId.value = null
  dragOverId.value = null
  if (!fromId || fromId === target.id || !isReorderable(target)) return
  moveItem(fromId, target.id)
  const orderedIds = items.value
    .filter((i) => i.status === UPLOAD_STATUS.SUCCESS && i.result?.id)
    .map((i) => i.result!.id!)
  if (orderedIds.length > 0) editorStore.reorderFilesByIds(orderedIds)
}

function onItemDragEnd() {
  draggingId.value = null
  dragOverId.value = null
}

type CompressionBadge =
  | { tone: 'savings'; label: string; tooltip: string }
  | { tone: 'none'; label: string; tooltip: string }
  | null

function compressionBadge(item: QueueItem): CompressionBadge {
  if (!item.compressionAttempted) return null
  const before = item.originalSize
  const after = item.file.size
  if (before > 0 && after > 0 && after < before) {
    const pct = Math.round((1 - after / before) * 100)
    if (pct >= 1) {
      return {
        tone: 'savings',
        label: `−${pct}%`,
        tooltip: t('uploader.compressionTooltip', {
          from: formatBytes(before),
          to: formatBytes(after),
        }),
      }
    }
  }
  return {
    tone: 'none',
    label: t('uploader.compressionNone'),
    tooltip: t('uploader.compressionNoneTooltip'),
  }
}

function statusLabel(item: QueueItem): string {
  switch (item.status) {
    case UPLOAD_STATUS.PENDING:
      return t('uploader.statusPending')
    case UPLOAD_STATUS.COMPRESSING:
      return t('uploader.statusCompressing')
    case UPLOAD_STATUS.UPLOADING:
      return `${item.progress}%`
    case UPLOAD_STATUS.SUCCESS:
      return t('uploader.statusSuccess')
    case UPLOAD_STATUS.CANCELLED:
      return t('uploader.statusCancelled')
    case UPLOAD_STATUS.ERROR:
      return t('uploader.statusError')
    default:
      return ''
  }
}

function statusColorClass(status: UploadStatus): string {
  switch (status) {
    case UPLOAD_STATUS.SUCCESS:
      return 'text-[var(--color-success,#16a34a)]'
    case UPLOAD_STATUS.ERROR:
      return 'text-[var(--color-danger,#dc2626)]'
    case UPLOAD_STATUS.UPLOADING:
    case UPLOAD_STATUS.COMPRESSING:
      return 'text-[var(--color-accent)]'
    default:
      return 'text-[var(--color-text-muted)]'
  }
}
</script>
