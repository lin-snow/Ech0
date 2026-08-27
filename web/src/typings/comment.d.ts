// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Comment {
      type CommentStatus = 'pending' | 'approved' | 'rejected'
      type BatchAction = 'approve' | 'reject' | 'delete'

      type CommentItem = {
        id: string
        echo_id: string
        parent_id?: string | null
        user_id?: string
        nickname: string
        email: string
        website?: string
        content: string
        status: CommentStatus
        hot: boolean
        source: 'guest' | 'system'
        created_at: number
        updated_at: number
      }

      type FormMeta = {
        form_token: string
        min_submit_ms: number
        captcha_enabled: boolean
        captcha_api_endpoint: string
        enable_comment: boolean
      }

      type CreateCommentDto = {
        echo_id: string
        parent_id?: string
        nickname: string
        email: string
        website: string
        content: string
        hp_field: string
        form_token: string
        captcha_token: string
      }

      type CreateCommentResult = {
        id: string
        status: CommentStatus
      }

      type PanelListQuery = {
        page: number
        page_size: number
        keyword?: string
        status?: string
        echo_id?: string
        hot?: boolean
      }

      type PanelPageResult = {
        items: CommentItem[]
        total: number
      }

      type SystemSetting = {
        enable_comment: boolean
        require_approval: boolean
        captcha_enabled: boolean
        email_notify: {
          enabled: boolean
          smtp_host: string
          smtp_port: number
          smtp_username: string
          smtp_password?: string
          smtp_password_set?: boolean
          smtp_sender: string
        }
      }
    }
  }
}
