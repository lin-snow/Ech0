// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { computed, ref } from 'vue'
import { FILE_CATEGORY, FILE_STORAGE_TYPE, type FileCategory } from '@/constants/file'
import { createExternalFile, globalFileRegistry, useFileAttachments } from '@/lib/file'
import { getImageSize } from '@/utils/image'
import { getFileToAddUrl } from '@/utils/other'
import { theToast } from '@/utils/toast'
import type { Translate } from './types'

type FilesModuleDeps = {
  t: Translate
}

export function useFilesModule({ t }: FilesModuleDeps) {
  const fileUploading = ref<boolean>(false)
  const fileIndex = ref<number>(0)

  const fileToAdd = ref<App.Api.Ech0.FileToAdd>({
    url: '',
    storage_type: FILE_STORAGE_TYPE.LOCAL,
    key: '',
  })

  const {
    files: filesToAdd,
    addAttachment,
    resetAttachments,
    removeAttachment,
    validateAttachments,
  } = useFileAttachments()

  const hasFile = computed(() => filesToAdd.value.length > 0)

  const selectedCategory = ref<FileCategory>(FILE_CATEGORY.IMAGE)

  const mediaCategory = computed<FileCategory | null>(
    () => (filesToAdd.value[0]?.category as FileCategory | undefined) ?? null,
  )

  const effectiveCategory = computed<FileCategory>(
    () => mediaCategory.value ?? selectedCategory.value,
  )

  const setSelectedCategory = (category: FileCategory) => {
    if (mediaCategory.value && mediaCategory.value !== category) {
      theToast.info(t('editor.categoryLockedHint'))
      return
    }
    selectedCategory.value = category
  }

  const handleAddMoreFile = async () => {
    const incomingCategory =
      (fileToAdd.value.category as FileCategory | undefined) ?? effectiveCategory.value
    if (mediaCategory.value && mediaCategory.value !== incomingCategory) {
      theToast.error(t('editor.mixedCategoryRejected'))
      return
    }

    const isSingleClipCategory =
      incomingCategory === FILE_CATEGORY.AUDIO || incomingCategory === FILE_CATEGORY.VIDEO
    if (isSingleClipCategory && filesToAdd.value.length > 0) {
      theToast.error(t('editor.singleMediaLimit'))
      return
    }

    let width: number | undefined = fileToAdd.value.width
    let height: number | undefined = fileToAdd.value.height
    if (incomingCategory === FILE_CATEGORY.IMAGE && (width === undefined || height === undefined)) {
      try {
        const previewUrl = getFileToAddUrl(fileToAdd.value)
        const size = await getImageSize(previewUrl || fileToAdd.value.url)
        width = size.width
        height = size.height
      } catch {}
    }

    if (fileToAdd.value.storage_type === FILE_STORAGE_TYPE.EXTERNAL && !fileToAdd.value.id) {
      const externalUrl = String(fileToAdd.value.url || '').trim()
      if (!externalUrl) {
        theToast.error(t('editor.imageUrlRequired'))
        return
      }

      const created = await createExternalFile({
        url: externalUrl,
        category: incomingCategory,
        width: width,
        height: height,
      })
      if (!created.id) {
        theToast.error(t('editor.externalRegisterFailed'))
        return
      }

      fileToAdd.value.id = created.id
      fileToAdd.value.key = created.key
      fileToAdd.value.url = created.url || externalUrl
      globalFileRegistry.upsert(created)
    }

    addAttachment({
      id: fileToAdd.value.id,
      url: fileToAdd.value.url,
      storage_type: fileToAdd.value.storage_type,
      category: incomingCategory,
      content_type: fileToAdd.value.content_type,
      key: fileToAdd.value.key ? fileToAdd.value.key : '',
      size: fileToAdd.value.size,
      width,
      height,
    })

    fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: fileToAdd.value.storage_type
        ? fileToAdd.value.storage_type
        : FILE_STORAGE_TYPE.LOCAL,
      key: '',
    }
  }

  const setFilesToAdd = (files: App.Api.Ech0.FileToAdd[]) => {
    resetAttachments(
      files.map((file) => ({
        id: file.id,
        key: file.key,
        url: file.url,
        storage_type: file.storage_type,
        category: file.category,
        content_type: file.content_type,
        size: file.size,
        width: file.width,
        height: file.height,
      })),
    )
  }

  const reorderFilesByIds = (orderedIds: string[]) => {
    if (orderedIds.length === 0) return
    const idSet = new Set(orderedIds)
    const rank = new Map(orderedIds.map((id, i) => [id, i]))
    const current = filesToAdd.value
    const positions: number[] = []
    const managed: App.Api.Ech0.FileToAdd[] = []
    current.forEach((file, idx) => {
      if (file.id && idSet.has(file.id)) {
        positions.push(idx)
        managed.push(file)
      }
    })
    if (managed.length === 0) return
    managed.sort((a, b) => (rank.get(a.id!) ?? 0) - (rank.get(b.id!) ?? 0))
    const next = current.slice()
    positions.forEach((pos, i) => {
      next[pos] = managed[i]
    })
    setFilesToAdd(next)
  }

  const removeFileAt = (index: number) => {
    removeAttachment(index)
  }

  const resetFilesState = () => {
    fileToAdd.value = {
      id: undefined,
      url: '',
      storage_type: fileToAdd.value.storage_type
        ? fileToAdd.value.storage_type
        : FILE_STORAGE_TYPE.LOCAL,
      key: '',
    }
    selectedCategory.value = FILE_CATEGORY.IMAGE
    resetAttachments([])
  }

  return {
    fileUploading,
    fileToAdd,
    filesToAdd,
    fileIndex,
    selectedCategory,
    hasFile,
    mediaCategory,
    effectiveCategory,
    handleAddMoreFile,
    setSelectedCategory,
    setFilesToAdd,
    reorderFilesByIds,
    removeFileAt,
    resetFilesState,
    resetAttachments,
    validateAttachments,
  }
}
