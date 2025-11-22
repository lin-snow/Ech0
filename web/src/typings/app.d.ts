declare namespace App {
  /**
   * Namespace Api
   */
  namespace Api {
    type Response<T> = {
      code: number
      msg: string
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
      }
    }

    namespace User {
      type User = {
        id: number
        username: string
        password?: string
        is_admin: boolean
        avatar?: string
      }

      type UserInfo = {
        username: string
        password: string
        is_admin: boolean
        avatar: string
      }

      type UserStatus = {
        user_id: number
        username: string
        is_admin: boolean
      }
    }

    namespace Ech0 {
      type ParamsByPagination = {
        page: number
        pageSize: number
        search?: string
      }

      type Echo = {
        id: number
        content: string
        username: string
        image_url: string
        image_source: string
        images: Image[]
        layout?: string
        private: boolean
        user_id: number
        extension: string
        extension_type: string
        tags?: Tag[]
        fav_count: number
        created_at: string
        user?: {
          id: number
          username: string
          avatar: string
        }
      }

      type Image = {
        id: number
        message_id: number
        image_url: string
        image_source: string
        object_key?: string // 对象存储的Key (如果是本地存储则为空)
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type Tag = {
        id: number
        name: string
        usage_count: number
        created_at: string
      }

      type ImageToAdd = {
        image_url: string
        image_source: string
        object_key?: string // 对象存储的Key (如果是本地存储则为空)
        width?: number // 图片宽度
        height?: number // 图片高度
      }

      type TagToAdd = {
        id?: number
        name: string
        usage_count?: number
        created_at?: string
      }

      type EchoToAdd = {
        content: string
        images?: ImageToAdd[] | null
        tags?: TagToAdd[] | null
        layout?: string | null
        extension?: string | null
        extension_type?: string | null
        private: boolean
      }

      type EchoToUpdate = {
        id: number
        content: string
        username: string
        images?: ImageToAdd[] | null
        tags?: TagToAdd[] | null
        layout?: string | null
        private: boolean
        user_id: number
        extension?: string | null
        extension_type?: string | null
        created_at: string
      }

      type PaginationResult = {
        items: Echo[]
        total: number
      }

      type Status = {
        sys_admin_id: number // 系统管理员ID
        username: string // 系统管理员用户名
        logo: string // 站点Logo
        users: App.Api.User.UserStatus[] // 用户列表
        total_echos: number // Echo总数
      }

      type HeatMap = {
        date: string
        count: number
      }[]

      type ImageToDelete = {
        url: string
        source: string
        object_key?: string // 对象存储的 Key, 用于删除 S3/R2 上的图片
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
        file_name: string
        content_type: string
        object_key: string
        presign_url: string
        file_url: string
      }
    }

    namespace Setting {
      type SystemSetting = {
        site_title: string
        server_name: string
        server_url: string
        allow_register: boolean
        ICP_number: string
        meting_api: string
        custom_css: string
        custom_js: string
        logo: string
      }

      type CommentSetting = {
        enable_comment: boolean
        provider: string // 评论提供者
        comment_api: string // 评论 API 地址
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
      }

      type OAuth2Status = {
        enabled: boolean
        provider: string
      }

      type OAuthInfo = {
        provider: string
        user_id: number
        oauth_id: string
      }

      type FediverseSetting = {
        enable: boolean
        server_url: string
      }

      type Webhook = {
        id: number
        name: string
        url: string
        secret: string
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
        id: number
        user_id: number
        token: string
        name: string
        expiry: string
        created_at: string
      }

      type AccessTokenDto = {
        name: string
        expiry: string
      }

      type BackupSchedule = {
        enable: boolean
        cron_expression: string
      }

      type BackupScheduleDto = {
        enable: boolean
        cron_expression: string
      }
    }

    namespace Connect {
      type Connect = {
        server_name: string
        server_url: string
        logo: string // 站点Logo
        total_echos: number
        today_echos: number
        sys_username: string
      }

      type Connected = {
        id: number
        connect_url: string
      }
    }

    namespace Fediverse {
      type Actor = Record<string, unknown>

      type FollowActionRequest = {
        targetActor: string
      }

      type FollowResponse = {
        activityId: string
      }

      // type LikeActionRequest = {
      //   targetActor: string
      //   object: string
      //   objectType?: string
      // }

      type UnfollowResponse = {
        activityId: string
        followActivityId?: string
      }

      // type LikeResponse = {
      //   activityId: string
      // }

      // type UndoLikeResponse = {
      //   activityId: string
      //   likeActivityId?: string
      // }

      type TimelineItem = {
        id: number
        activityId: string
        actorId: string
        actorPreferredUsername: string
        actorDisplayName: string
        actorAvatar: string
        objectId: string
        objectType: string
        objectAttributedTo: string
        summary: string
        content: string
        to: string[]
        cc: string[]
        rawActivity?: unknown
        rawObject?: unknown
        publishedAt: string
        createdAt: string
        updatedAt: string
      }

      type TimelineResult = {
        total: number
        items: TimelineItem[]
      }
    }

    namespace Todo {
      type Todo = {
        id: number
        content: string
        user_id: number
        username: string
        status: number
        created_at: string
      }

      type TodoToAdd = {
        content: string
      }
    }

    namespace Dashboard {
      // CpuMetric cpu监控指标
      type CpuMetric = {
        UsagePercent: number // CPU 使用率百分比
        Cores: number // CPU 核心数
        FrequencyMHz: number // CPU 主频，单位 MHz
      }
      // MemoryMetric 内存监控指标
      type MemoryMetric = {
        Total: number // 总内存大小
        Used: number // 已使用内存大小
        Available: number // 可用内存大小
        Percentage: number // 内存使用率百分比
      }
      // DiskMetric 磁盘监控指标
      type DiskMetric = {
        Total: number // 磁盘总大小
        Used: number // 已使用磁盘大小
        Available: number // 可用磁盘大小
        Percentage: number // 磁盘使用率百分比
      }

      // NetworkMetric 网络监控指标
      type NetworkMetric = {
        TotalBytesSent: number // 总发送字节数
        TotalBytesReceived: number // 总接收字节数
        BytesSentPerSecond: number // 每秒发送字节数 (B/s)
        BytesReceivedPerSecond: number // 每秒接收字节数 (B/s)
      }

      // SystemMetric 系统监控指标
      type SystemMetric = {
        Hostname: string // 主机名
        OsName: string // 操作系统名称
        Uptime: number // 系统运行时长
        KernelVersion: string // 内核版本
        KernelArch: string // 内核架构
        Time: string // 采样时间
        TimeZone: string // 采样时区
        ProcessCount: number // 当前进程数
        ThreadCount: number // 当前线程数
        GolangVersion: string // Golang 版本
        GoRoutineCount: number // 当前 Goroutine 数量
      }

      // Metrics 综合监控指标
      type Metrics = {
        CPU: CpuMetric // CPU 监控指标
        Memory: MemoryMetric // 内存监控指标
        Disk: DiskMetric // 磁盘监控指标
        Network: NetworkMetric // 网络监控指标
        System: SystemMetric // 系统监控指标
      }
    }

    namespace Hub {
      type HubItem = string | { id: number; connect_url: string }
      type HubList = HubItem[]
      type HubItemInfo = Connect.Connect
      type HubInfoList = HubItemInfo[]

      type Echo = {
        id: number
        content: string
        username: string
        image_url: string
        image_source: string
        images: Image[]
        tags?: Tag[]
        layout?: string
        private: boolean
        user_id: number
        extension: string
        extension_type: string
        fav_count: number
        created_at: string
        createdTs: number
        server_name: string
        server_url: string
        logo: string
        user?: {
          id: number
          username: string
          avatar: string
        }
      }
    }
  }
}
