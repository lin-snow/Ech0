// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export const FILE_CATEGORY = {
  IMAGE: 'image',
  AUDIO: 'audio',
  VIDEO: 'video',
  PDF: 'pdf',
  MARKDOWN: 'markdown',
  FILE: 'file',
} as const

export type FileCategory = (typeof FILE_CATEGORY)[keyof typeof FILE_CATEGORY]

export const FILE_CATEGORY_VALUES = Object.values(FILE_CATEGORY) as FileCategory[]

export const FILE_KIND = {
  OWNED: 'owned',
  BLOB: 'blob',
  REF: 'ref',
} as const

export type FileKind = (typeof FILE_KIND)[keyof typeof FILE_KIND]

export const FILE_KIND_VALUES = Object.values(FILE_KIND) as FileKind[]

export const FILE_ACCESS = {
  PUBLIC: 'public',
  OWNER: 'owner',
  PRIVATE: 'private',
} as const

export type FileAccess = (typeof FILE_ACCESS)[keyof typeof FILE_ACCESS]

export const FILE_ACCESS_VALUES = Object.values(FILE_ACCESS) as FileAccess[]
