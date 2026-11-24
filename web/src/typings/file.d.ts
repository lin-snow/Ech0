// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace File {
      type Category = import('@/constants/file').FileCategory
      type Kind = import('@/constants/file').FileKind
      type Access = import('@/constants/file').FileAccess
      type Profile = 'attachment' | 'draft' | 'avatar' | 'instance'
      type Status = 'pending' | 'ready'

      type Link = {
        owner_kind: string
        owner_id: string
        role: string
        sort_order: number
      }

      type FileDto = {
        id: string
        kind: Kind
        access: Access
        status: Status
        name: string
        content_type: string
        category: Category
        size: number
        width?: number
        height?: number
        url: string
        uploader_id: string
        created_at: number
        links?: Link[]
      }
      type FileListQuery = {
        page: number
        pageSize: number
        search?: string
        kind?: Kind
        access?: Access
      }
      type FileListResult = {
        items: FileDto[]
        total: number
      }
      type UploadTicket = {
        url: string
        method: string
        headers?: Record<string, string>
        expires_at?: number
      }
      type UploadResult = {
        file: FileDto
        upload?: UploadTicket
      }
      type UploadFileDto = {
        file: File
        profile?: Profile
        name?: string
        width?: number
        height?: number
      }
      type RegisterExternalFileDto = {
        url: string
        name?: string
        content_type?: string
        width?: number
        height?: number
      }
      type RenameFileDto = {
        name: string
      }
      type FileDeleteDto = {
        id: string
        force?: boolean
      }
    }
  }
}
