// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type UserRole = 'owner' | 'admin' | 'member'

/** Administrators and the owner share every administrative surface. */
export function isAdminRole(role?: string | null): boolean {
  return role === 'admin' || role === 'owner'
}

/** Ownership is unique: only the owner assigns roles or hands over the space. */
export function isOwnerRole(role?: string | null): boolean {
  return role === 'owner'
}
