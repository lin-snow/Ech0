// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

declare namespace App {
  namespace Api {
    namespace Setting {
      type SystemSetting = {
        site_title: string
        server_logo: string
        server_logo_file_id?: string
        server_name: string
        server_url: string
        home_layout: 'single' | 'three'
        allow_register: boolean
        default_locale: string
        ICP_number: string
        footer_content: string
        footer_link: string
        meting_api: string
        custom_css: string
        custom_js: string
      }

      type S3Setting = {
        enable: boolean
        provider: string
        endpoint: string
        access_key: string
        secret_key: string
        bucket_name: string
        region: string
        use_ssl: boolean
        cdn_url: string
        path_prefix: string
        public_read: boolean
        use_path_style: boolean
      }

      type OAuth2Setting = {
        enable: boolean
        provider: string
        client_id: string
        client_secret: string
        redirect_uri: string
        scopes: string[]
        auth_url: string
        token_url: string
        user_info_url: string

        is_oidc: boolean
        issuer: string
        jwks_url: string

        auth_redirect_allowed_return_urls: string[]
        cors_allowed_origins: string[]
      }

      type OAuth2Status = {
        enabled: boolean
        provider: string
        oauth_ready: boolean
      }

      type PasskeySetting = {
        webauthn_rp_id: string
        webauthn_allowed_origins: string[]
      }

      type PasskeyStatus = {
        passkey_ready: boolean
      }

      type OAuthInfo = {
        provider: string
        user_id: string
        oauth_id: string
        issuer: string
        auth_type: string
      }

      type Webhook = {
        id: string
        name: string
        url: string
        is_active: boolean
        last_status: string
        last_trigger: number
        created_at: number
        updated_at: number
      }

      type WebhookDto = {
        name: string
        url: string
        secret?: string
        is_active: boolean
      }

      type AccessToken = {
        id: string
        user_id: string
        token: string
        name: string
        token_type?: string
        scopes?: string | string[]
        audience?: 'public-client' | 'cli' | 'integration' | 'mcp-remote'
        jti?: string
        expiry: number | null
        last_used_at?: number | null
        created_at: number
      }

      type AccessTokenDto = {
        name: string
        expiry: string
        scopes: string[]
        audience: 'public-client' | 'cli' | 'integration' | 'mcp-remote'
      }

      type SnapshotSchedule = {
        enable: boolean
        cron_expression: string
      }

      type SnapshotScheduleDto = {
        enable: boolean
        cron_expression: string
      }

      type AgentSetting = {
        enable: boolean
        protocol: string
        model: string
        api_key: string
        prompt: string
        base_url: string
        multimodal: boolean
        context_window: number
      }

      type AgentSettingDto = {
        enable: boolean
        protocol: string
        model: string
        api_key: string
        prompt: string
        base_url: string
        multimodal: boolean
        context_window: number
      }
    }
  }
}
