// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace File {
      type Category = import('@/constants/file').FileCategory
      type StorageType = import('@/constants/file').FileStorageType

      type FileDto = {
        id: string
        name?: string
        key: string
        url: string
        content_type?: string
        category?: Category
        storage_type?: StorageType
        size?: number
        width?: number
        height?: number
      }
      type FileListQuery = {
        page: number
        pageSize: number
        search?: string
        storage_type?: StorageType
      }
      type FileListItem = {
        id: string
        name: string
        key: string
        storage_type: StorageType
        url: string
        content_type?: string
        size?: number
        created_at: number
      }
      type FileListResult = {
        items: FileListItem[]
        total: number
      }
      type FileTreeQuery = {
        storage_type: StorageType
        prefix?: string
      }
      type FilePathStreamQuery = {
        storage_type: StorageType
        path: string
        name?: string
        content_type?: string
      }
      type FileTreeNode = {
        name: string
        path: string
        node_type: 'file' | 'folder'
        has_children: boolean
        file_id?: string
        size?: number
        content_type?: string
        modified_at?: number
      }
      type FileTreeResult = {
        items: FileTreeNode[]
      }
      type FileDeleteDto = {
        id: string
      }
      type CreateExternalFileDto = {
        url: string
        content_type?: string
        category?: Category
        width?: number
        height?: number
        name?: string
      }
      type UpdateFileMetaDto = {
        size: number
        width?: number
        height?: number
        content_type?: string
      }
    }
  }
}
