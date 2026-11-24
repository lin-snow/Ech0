// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type UserRole = "owner" | "admin" | "member";

/** Administrators and the space owner share every administrative surface. */
export function isAdminRole(role: string | null | undefined): boolean {
  return role === "admin" || role === "owner";
}
