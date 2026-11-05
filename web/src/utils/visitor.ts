// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const VISITOR_ID_KEY = 'ech0_visitor_id'

function isValidVisitorId(value: string | null): value is string {
  return !!value && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(value)
}

function newVisitorId(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  // Extremely old browsers: build a v4-shaped id from random bytes.
  const bytes = new Uint8Array(16)
  crypto.getRandomValues(bytes)
  bytes[6] = (bytes[6]! & 0x0f) | 0x40
  bytes[8] = (bytes[8]! & 0x3f) | 0x80
  const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
  return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
}

/**
 * Anonymous identity for emoji reactions. It is not a credential, only a
 * stable per-browser key so one person's reactions stay separate.
 */
export function getVisitorId(): string {
  if (typeof localStorage === 'undefined') {
    return newVisitorId()
  }
  const stored = localStorage.getItem(VISITOR_ID_KEY)
  if (isValidVisitorId(stored)) {
    return stored
  }
  const created = newVisitorId()
  localStorage.setItem(VISITOR_ID_KEY, created)
  return created
}
