// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { FileAttachment, FileValidationRule } from '../types'

function keyOf(file: FileAttachment): string {
  return String(file.id || file.url)
}

export class AttachmentManager {
  private files: FileAttachment[] = []

  list() {
    return [...this.files]
  }

  reset(files: FileAttachment[] = []) {
    this.files = [...files]
  }

  add(file: FileAttachment) {
    const dedupKey = keyOf(file)
    const has = this.files.some((item) => keyOf(item) === dedupKey)
    if (!has) this.files.push(file)
  }

  addMany(files: FileAttachment[]) {
    files.forEach((file) => this.add(file))
  }

  remove(index: number) {
    if (index < 0 || index >= this.files.length) return
    this.files.splice(index, 1)
  }

  removeById(fileId: string) {
    const idx = this.files.findIndex((item) => String(item.id || '') === fileId)
    if (idx >= 0) this.files.splice(idx, 1)
  }

  reorder(from: number, to: number) {
    if (from < 0 || to < 0 || from >= this.files.length || to >= this.files.length) return
    const [item] = this.files.splice(from, 1)
    if (item) this.files.splice(to, 0, item)
  }

  validate(rule: FileValidationRule = {}) {
    if (rule.maxCount !== undefined && this.files.length > rule.maxCount) {
      return { valid: false, reason: `附件数量超过限制（${rule.maxCount}）` }
    }
    if (rule.requireId) {
      const missing = this.files.find((item) => !item.id)
      if (missing) return { valid: false, reason: '存在未绑定 file_id 的附件' }
    }
    return { valid: true as const }
  }
}
