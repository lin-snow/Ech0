// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Ech0 {
      type EchoExtensionType = 'MUSIC' | 'VIDEO' | 'GITHUBPROJ' | 'WEBSITE' | 'LOCATION' | 'TWEET'
      type EchoExtension =
        | { type: 'MUSIC'; payload: { url: string } }
        | { type: 'VIDEO'; payload: { videoId: string } }
        | { type: 'GITHUBPROJ'; payload: { repoUrl: string } }
        | { type: 'WEBSITE'; payload: { title: string; site: string } }
        | {
            type: 'LOCATION'
            payload: { placeholder: string; latitude: number; longitude: number }
          }
        | {
            type: 'TWEET'
            payload: { url: string; username: string; statusId: string }
          }

      type ParamsByPagination = {
        page: number
        pageSize: number
        search?: string
      }

      type EchoVisibilityFilter = 'all' | 'public' | 'private'

      type EchoQueryParams = {
        page: number
        pageSize: number
        search?: string
        tagIds?: string[]
        sortBy?: string
        sortOrder?: string
        dateFrom?: number
        dateTo?: number
        private?: boolean
      }

      type Echo = {
        id: string
        content: string
        username: string
        echo_files?: EchoFile[]
        layout?: string
        private: boolean
        user_id: string
        extension?: EchoExtension | null
        tags?: Tag[]
        fav_count: number
        created_at: number | string
      }

      type FileObject = {
        id: string
        echo_id: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string
        size?: number
        width?: number
        height?: number
      }

      type Tag = {
        id: string
        name: string
        usage_count: number
        created_at: number | string
      }

      type EchoFile = {
        id: string
        echo_id: string
        file_id: string
        sort_order: number
        file?: {
          id: string
          key: string
          storage_type: File.StorageType
          provider?: string
          bucket?: string
          url: string
          name?: string
          content_type?: string
          size?: number
          category?: File.Category
          user_id?: string
          width?: number
          height?: number
          created_at?: number
        }
      }

      type FileToAdd = {
        id?: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string
        size?: number
        width?: number
        height?: number
      }

      type TagToAdd = {
        id?: string
        name: string
        usage_count?: number
        created_at?: number | string
      }

      type EchoToAdd = {
        content: string
        echo_files?: Array<{ file_id: string; sort_order: number }> | null
        tags?: TagToAdd[] | null
        layout?: string | null
        extension?: EchoExtension | null
        private: boolean
      }

      type EchoToUpdate = {
        id: string
        content: string
        username: string
        echo_files?: Array<{ file_id: string; sort_order: number }> | null
        tags?: TagToAdd[] | null
        layout?: string | null
        private: boolean
        user_id: string
        extension?: EchoExtension | null
        created_at: number | string
      }

      type PaginationResult = {
        items: Echo[]
        total: number
      }

      type HeatMap = {
        date: string
        count: number
      }[]

      type FileToDelete = {
        id: string
      }

      type GithubCardData = {
        name: string
        stargazers_count: number
        forks_count: number
        description: string
        owner: {
          avatar_url: string
        }
      }

      type HelloEch0 = {
        hello: string
        copyright: string
        version: string
        commit: string
        build_time: string
        license: string
        author: string
        repo_url: string
      }

      type PresignResult = {
        id: string
        file_name: string
        content_type: string
        key: string
        presign_url: string
        file_url: string
      }
    }
  }
}
