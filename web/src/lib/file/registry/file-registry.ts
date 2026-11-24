// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import type { FileCategory, FileEntity } from '../types'

export class FileRegistry {
  private byId = new Map<string, FileEntity>()

  upsert(file: FileEntity) {
    if (!file.id) return
    this.byId.set(file.id, file)
  }

  batchUpsert(files: FileEntity[]) {
    files.forEach((file) => this.upsert(file))
  }

  getById(id: string) {
    return this.byId.get(id)
  }

  listByCategory(category: FileCategory) {
    return Array.from(this.byId.values()).filter((file) => file.category === category)
  }

  removeById(id: string) {
    const existing = this.byId.get(id)
    if (!existing) return
    this.byId.delete(id)
  }

  clear() {
    this.byId.clear()
  }
}

export const globalFileRegistry = new FileRegistry()
