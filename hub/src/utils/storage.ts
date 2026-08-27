// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const localStg = {
  setItem<T>(key: string, obj: T) {
    localStorage.setItem(key, JSON.stringify(obj))
  },
  getItem<T>(key: string): T | null {
    const item = localStorage.getItem(key)
    if (!item) return null
    try {
      return JSON.parse(item) as T
    } catch {
      return null
    }
  },
  removeItem(key: string) {
    localStorage.removeItem(key)
  },
  clear() {
    localStorage.clear()
  },
}
