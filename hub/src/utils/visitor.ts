// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

const VISITOR_ID_KEY = 'ech0_visitor_id'

function isValidVisitorId(value: string | null): value is string {
  return !!value && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(value)
}

/** Stable per-browser anonymous identity used for reactions on remote instances. */
export function getVisitorId(): string {
  const stored = localStorage.getItem(VISITOR_ID_KEY)
  if (isValidVisitorId(stored)) {
    return stored
  }
  const created =
    typeof crypto.randomUUID === 'function'
      ? crypto.randomUUID()
      : Array.from(crypto.getRandomValues(new Uint8Array(16)), (b) => b.toString(16).padStart(2, '0')).join('')
  localStorage.setItem(VISITOR_ID_KEY, created)
  return created
}
