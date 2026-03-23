declare namespace App {
  /**
   * Namespace Api
   */
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
      error_code?: string
      message_key?: string
      message_params?: Record<string, unknown>
      data: T
    }

    namespace Auth {
      type LoginParams = {
        username: string
        password: string
      }

      type LoginResponse = {
        token: string
      }

      type SignupParams = {
        username: string
        password: string
        email?: string
      }

      // Passkey / WebAuthn
      type PasskeyRegisterBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyLoginBeginResp = {
        nonce: string
        publicKey: unknown
      }

      type PasskeyDevice = {
        id: string
        device_name: string
        aaguid: string
        last_used_at: string
        created_at: string
      }
    }

    namespace User {
      type User = {
        id: string
        username: string
        email?: string
        password?: string
        is_admin: boolean
        is_owner?: boolean
        avatar?: string
        locale: string
      }

      type UserInfo = {
        username: string
        password: string
        email?: string
        is_admin: boolean
        is_owner?: boolean
        avatar: string
        locale: string
      }
    }

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
        created_at: string
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
        modified_at?: string
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

    namespace Ech0 {
      type EchoExtensionType = 'MUSIC' | 'VIDEO' | 'GITHUBPROJ' | 'WEBSITE'
      type EchoExtension =
        | { type: 'MUSIC'; payload: { url: string } }
        | { type: 'VIDEO'; payload: { videoId: string } }
        | { type: 'GITHUBPROJ'; payload: { repoUrl: string } }
        | { type: 'WEBSITE'; payload: { title: string; site: string } }

      type ParamsByPagination = {
        page: number
        pageSize: number
        search?: string
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
        created_at: string
      }

      type FileObject = {
        id: string
        echo_id: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string // 对应后端 file.key
        size?: number // 文件大小（字节）
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type Tag = {
        id: string
        name: string
        usage_count: number
        created_at: string
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
          created_at?: string
        }
      }

      type FileToAdd = {
        id?: string
        url: string
        storage_type: File.StorageType
        category?: File.Category
        content_type?: string
        key?: string // 对应后端 file.key
        size?: number // 文件大小（字节）
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type TagToAdd = {
        id?: string
        name: string
        usage_count?: number
        created_at?: string
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
        created_at: string
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
        version: string
        github: string
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

    namespace Setting {
      type SystemSetting = {
        site_title: string
        server_logo: string
        server_name: string
        server_url: string
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
        last_trigger: string
        created_at: string
        updated_at: string
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
        audience?: 'public-client' | 'cli' | 'integration'
        jti?: string
        expiry: string | null
        last_used_at?: string | null
        created_at: string
      }

      type AccessTokenDto = {
        name: string
        expiry: string
        scopes: string[]
        audience: 'public-client' | 'cli' | 'integration'
      }

      type BackupSchedule = {
        enable: boolean
        cron_expression: string
      }

      type BackupScheduleDto = {
        enable: boolean
        cron_expression: string
      }

      type AgentSetting = {
        enable: boolean
        provider: string
        model: string
        api_key: string
        prompt: string
        base_url: string
      }

      type AgentSettingDto = {
        enable: boolean
        provider: string
        model: string
        api_key: string
        prompt: string
        base_url: string
      }
    }

    namespace Comment {
      type CommentStatus = 'pending' | 'approved' | 'rejected'
      type BatchAction = 'approve' | 'reject' | 'delete'

      type CommentItem = {
        id: string
        echo_id: string
        user_id?: string
        nickname: string
        email: string
        website?: string
        content: string
        status: CommentStatus
        hot: boolean
        source: 'guest' | 'system'
        created_at: string
        updated_at: string
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
        }
      }
    }

    namespace Init {
      type Status = {
        initialized: boolean
        owner_exists: boolean
      }
    }

    namespace Connect {
      type Connect = {
        server_name: string
        server_url: string
        logo: string
        total_echos: number
        today_echos: number
        sys_username: string
        version: string
      }

      type Connected = {
        id: string
        connect_url: string
      }
    }

    namespace SystemLog {
      type Entry = {
        time: string
        level: string
        msg: string
        module?: string
        caller?: string
        error?: string
        raw?: string
        fields?: Record<string, unknown>
      }

      type QueryParams = {
        tail?: number
        level?: string
        keyword?: string
      }
    }

    namespace Hub {
      type HubItem = string | { id: string; connect_url: string }
      type HubList = HubItem[]
      type HubItemInfo = Connect.Connect
      type HubInfoList = HubItemInfo[]

      type Echo = {
        id: string
        content: string
        username: string
        echo_files?: Ech0.EchoFile[]
        tags?: Tag[]
        layout?: string
        private: boolean
        user_id: string
        extension?: Ech0.EchoExtension | null
        fav_count: number
        created_at: string
        createdTs: number
        virtual_key: string
        server_name: string
        server_url: string
        logo: string
      }
    }

    namespace Inbox {
      type Inbox = {
        id: string
        source: string
        content: string
        type: string
        read: boolean
        read_count: number
        meta?: string
        read_at?: number // Unix时间戳（秒）
        created_at: number // Unix时间戳（秒）
      }

      type InboxListResult = {
        items: Inbox[]
        total: number
      }

      type InboxListParams = {
        page: number
        pageSize: number
        search?: string
      }
    }
  }
}
