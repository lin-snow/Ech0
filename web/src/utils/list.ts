// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/**
 * List endpoints answer `{ items }`, older ones a bare array. Both collapse to
 * the same array here so callers never have to branch on the shape.
 */
export function listItems<T>(body: unknown): T[] {
  if (Array.isArray(body)) return body as T[]
  if (body && typeof body === 'object' && 'items' in body) {
    const items: unknown = body.items
    if (Array.isArray(items)) return items as T[]
  }
  return []
}
