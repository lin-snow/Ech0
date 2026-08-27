// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import { ref, shallowRef, nextTick } from 'vue'

const title = ref('')
const description = ref('')
const onConfirmCallback = shallowRef<(() => void) | null>(null)
const onCancelCallback = shallowRef<(() => void) | null>(null)
import type { ComponentPublicInstance } from 'vue'
import type BaseDialog from '@/components/common/BaseDialog.vue'

type ConfirmDialogInstance = ComponentPublicInstance<typeof BaseDialog>
type ConfirmOptions = {
  title: string
  description: string
  onConfirm?: () => void
  onCancel?: () => void
}

let confirmDialogRef: ConfirmDialogInstance | null = null
let pendingOpenOptions: ConfirmOptions | null = null

export function useBaseDialog() {
  function register(refInstance: ConfirmDialogInstance) {
    confirmDialogRef = refInstance
    if (pendingOpenOptions) {
      const cached = pendingOpenOptions
      pendingOpenOptions = null
      openConfirm(cached)
    }
  }

  function openConfirm(options: ConfirmOptions) {
    title.value = options.title
    description.value = options.description
    onConfirmCallback.value = options.onConfirm || null
    onCancelCallback.value = options.onCancel || null

    if (!confirmDialogRef) {
      pendingOpenOptions = options
      return
    }

    nextTick(() => {
      confirmDialogRef?.open()
    })
  }

  function handleConfirm() {
    onConfirmCallback.value?.()
    onConfirmCallback.value = null
    onCancelCallback.value = null
  }

  function handleCancel() {
    onCancelCallback.value?.()
    onConfirmCallback.value = null
    onCancelCallback.value = null
  }

  return {
    register,
    openConfirm,
    title,
    description,
    handleConfirm,
    handleCancel,
  }
}
