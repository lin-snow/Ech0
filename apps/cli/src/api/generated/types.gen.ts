// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export type ClientOptions = {
    baseUrl: `${string}://${string}/api` | (string & {});
};

export type AccessTokenSetting = {
    audience?: string;
    created_at?: number;
    expiry?: number | null;
    id?: string;
    jti?: string;
    last_used_at?: number;
    name?: string;
    scopes?: string;
    token?: string;
    token_type?: string;
    user_id?: string;
    [key: string]: unknown;
};

export type AccessTokenSettingDto = {
    audience?: string;
    expiry?: string;
    name?: string;
    scopes?: Array<string> | null;
    [key: string]: unknown;
};

export type AgentSetting = {
    api_key?: string;
    base_url?: string;
    context_window?: number;
    enable?: boolean;
    model?: string;
    multimodal?: boolean;
    prompt?: string;
    protocol?: string;
    temperature?: number;
    [key: string]: unknown;
};

export type AgentSettingDto = {
    api_key?: string;
    base_url?: string;
    context_window?: number;
    enable?: boolean;
    model?: string;
    multimodal?: boolean;
    prompt?: string;
    protocol?: string;
    /**
     * 采样温度；留空则不发送，使用模型默认值。是否支持以模型官方文档为准
     */
    temperature?: number | null;
    [key: string]: unknown;
};

export type AnswerAskInputBody = {
    /**
     * 每个问题的回答，按问题 ID 对应
     */
    answers?: Array<AskAnswer> | null;
    /**
     * 要回答的提问轮次 ID，来自 SSE 的 ask 事件
     */
    ask_id?: string;
    [key: string]: unknown;
};

export type AskAnswer = {
    custom?: string;
    question_id?: string;
    selected?: Array<string> | null;
    [key: string]: unknown;
};

export type AskExchange = {
    answers?: Array<AskAnswer> | null;
    questions?: Array<AskQuestion> | null;
    [key: string]: unknown;
};

export type AskOption = {
    description?: string;
    label?: string;
    [key: string]: unknown;
};

export type AskQuestion = {
    detail?: string;
    header?: string;
    id?: string;
    multi?: boolean;
    options?: Array<AskOption> | null;
    recommended?: number;
    text?: string;
    [key: string]: unknown;
};

export type BatchCommentActionDto = {
    action?: string;
    ids?: Array<string> | null;
    [key: string]: unknown;
};

export type ChatMessage = {
    asks?: Array<AskExchange> | null;
    content?: string;
    reasoning?: string;
    reasoning_ms?: number;
    role?: string;
    sources?: Array<SearchResult> | null;
    [key: string]: unknown;
};

export type CheckUpdateResponse = {
    current_version?: string;
    has_update?: boolean;
    latest_version?: string;
    [key: string]: unknown;
};

export type Comment = {
    content?: string;
    created_at?: number;
    echo_id?: string;
    email?: string;
    hot?: boolean;
    id?: string;
    nickname?: string;
    parent_id?: string;
    source?: string;
    status?: string;
    updated_at?: number;
    user_id?: string;
    website?: string;
    [key: string]: unknown;
};

export type Connect = {
    logo?: string;
    server_name?: string;
    server_url?: string;
    sys_username?: string;
    today_echos?: number;
    total_echos?: number;
    version?: string;
    [key: string]: unknown;
};

export type Connected = {
    connect_url?: string;
    id?: string;
    [key: string]: unknown;
};

export type ConnectedHealth = {
    connect_url?: string;
    id?: string;
    status?: string;
    version?: string;
    [key: string]: unknown;
};

export type CreateCommentDto = {
    captcha_token?: string;
    content?: string;
    echo_id?: string;
    email?: string;
    form_token?: string;
    hp_field?: string;
    nickname?: string;
    parent_id?: string;
    website?: string;
    [key: string]: unknown;
};

export type CreateCommentResult = {
    id?: string;
    status?: string;
    [key: string]: unknown;
};

export type CreateExternalFileDto = {
    category?: string;
    content_type?: string;
    height?: number;
    name?: string;
    url?: string;
    width?: number;
    [key: string]: unknown;
};

export type CreateIntegrationCommentDto = {
    content?: string;
    echo_id?: string;
    metadata?: string;
    nickname?: string;
    [key: string]: unknown;
};

export type CreateTagDto = {
    /**
     * 标签名
     */
    name?: string;
    [key: string]: unknown;
};

export type DayStat = {
    date?: string;
    pv?: number;
    uv?: number;
    [key: string]: unknown;
};

export type Echo = {
    content?: string;
    created_at?: number;
    echo_files?: Array<EchoFile> | null;
    extension?: EchoExtension;
    fav_count?: number;
    id?: string;
    layout?: string;
    private?: boolean;
    tags?: Array<Tag> | null;
    user_id?: string;
    username?: string;
    [key: string]: unknown;
};

export type EchoExtension = {
    created_at?: number;
    echo_id?: string;
    id?: string;
    payload?: {
        [key: string]: unknown;
    };
    type?: string;
    updated_at?: number;
    [key: string]: unknown;
};

export type EchoExtensionDto = {
    /**
     * 扩展渲染所需的任意字段
     */
    payload?: {
        [key: string]: unknown;
    };
    /**
     * 扩展卡片类型，如 weather / music
     */
    type?: string;
    [key: string]: unknown;
};

export type EchoFile = {
    /**
     * 所属 Echo ID，提交时忽略
     */
    echo_id?: string;
    /**
     * 文件元数据，提交时忽略（服务端按 file_id 填充）
     */
    file?: File;
    /**
     * 已上传文件的 ID（上传接口返回）
     */
    file_id?: string;
    /**
     * 关联记录 ID，提交时忽略
     */
    id?: string;
    /**
     * 同一 Echo 内的排序，从 0 开始
     */
    sort_order?: number;
    [key: string]: unknown;
};

export type EchoQueryDto = {
    dateFrom?: number;
    dateTo?: number;
    page?: number;
    pageSize?: number;
    private?: boolean;
    search?: string;
    sortBy?: string;
    sortOrder?: string;
    tagIds?: Array<string> | null;
    [key: string]: unknown;
};

export type EchoUpsertDto = {
    /**
     * 正文（Markdown）；与附件、扩展至少有一项
     */
    content?: string;
    /**
     * 发布时间（Unix 秒），缺省为当前时间
     */
    created_at?: number;
    /**
     * 附件引用列表：{file_id, sort_order}；文件先通过上传接口获取 file_id
     */
    echo_files?: Array<EchoFile> | null;
    /**
     * 扩展卡片 {type, payload}
     */
    extension?: EchoExtensionDto;
    /**
     * 更新时的 Echo ID；创建时忽略
     */
    id?: string;
    /**
     * 布局：waterfall / grid / horizontal / carousel / stack / none
     */
    layout?: string;
    /**
     * 仅自己可见
     */
    private?: boolean;
    /**
     * 标签列表，元素为 {name}；同名标签复用
     */
    tags?: Array<Tag> | null;
    [key: string]: unknown;
};

export type EmailNotifySetting = {
    enabled?: boolean;
    smtp_host?: string;
    smtp_password?: string;
    smtp_password_set?: boolean;
    smtp_port?: number;
    smtp_sender?: string;
    smtp_username?: string;
    [key: string]: unknown;
};

export type EmbeddingSetting = {
    api_key?: string;
    base_url?: string;
    batch_size?: number;
    dim?: number;
    enable?: boolean;
    model?: string;
    [key: string]: unknown;
};

export type EmbeddingSettingDto = {
    api_key?: string;
    base_url?: string;
    batch_size?: number;
    dim?: number;
    enable?: boolean;
    model?: string;
    [key: string]: unknown;
};

export type ErrorBody = {
    /**
     * 业务状态码，失败为 0
     */
    code?: number;
    /**
     * 失败时为 null
     */
    data?: unknown;
    /**
     * 稳定的业务错误码
     */
    error_code?: string;
    /**
     * i18n 翻译 key
     */
    message_key?: string;
    /**
     * i18n 模板参数
     */
    message_params?: {
        [key: string]: unknown;
    };
    /**
     * 状态描述（回退文案）
     */
    msg?: string;
    [key: string]: unknown;
};

export type ExportStateDto = {
    error_message?: string;
    file_name?: string;
    finished_at?: number;
    format?: string;
    phase?: string;
    size?: number;
    started_at?: number;
    status?: string;
    updated_at?: number;
    version?: number;
    [key: string]: unknown;
};

export type File = {
    bucket?: string;
    category?: string;
    content_type?: string;
    created_at?: number;
    height?: number;
    id?: string;
    key?: string;
    name?: string;
    provider?: string;
    size?: number;
    storage_type?: string;
    url?: string;
    user_id?: string;
    width?: number;
    [key: string]: unknown;
};

export type FileDto = {
    category?: string;
    content_type?: string;
    height?: number;
    id?: string;
    key?: string;
    name?: string;
    size?: number;
    storage_type?: string;
    url?: string;
    width?: number;
    [key: string]: unknown;
};

export type FileListItemDto = {
    content_type?: string;
    created_at?: number;
    id?: string;
    key?: string;
    name?: string;
    size?: number;
    storage_type?: string;
    url?: string;
    [key: string]: unknown;
};

export type FileListResultDto = {
    items?: Array<FileListItemDto> | null;
    total?: number;
    [key: string]: unknown;
};

export type FileTreeNodeDto = {
    content_type?: string;
    file_id?: string;
    has_children?: boolean;
    modified_at?: number;
    name?: string;
    node_type?: string;
    path?: string;
    size?: number;
    [key: string]: unknown;
};

export type FileTreeResultDto = {
    items?: Array<FileTreeNodeDto> | null;
    [key: string]: unknown;
};

export type FormMeta = {
    captcha_api_endpoint?: string;
    captcha_enabled?: boolean;
    enable_comment?: boolean;
    form_token?: string;
    min_submit_ms?: number;
    [key: string]: unknown;
};

export type GetPresignUrlDto = {
    content_type?: string;
    file_name?: string;
    storage_type?: string;
    [key: string]: unknown;
};

export type GlobalMigrationStateDto = {
    error_message?: string;
    finished_at?: number;
    phase?: string;
    source_payload?: {
        [key: string]: unknown;
    };
    source_type?: string;
    started_at?: number;
    status?: string;
    updated_at?: number;
    version?: number;
    [key: string]: unknown;
};

export type Heatmap = {
    count?: number;
    date?: string;
    [key: string]: unknown;
};

export type HelloResponse = {
    author?: string;
    build_time?: string;
    commit?: string;
    copyright?: string;
    hello?: string;
    license?: string;
    repo_url?: string;
    version?: string;
    [key: string]: unknown;
};

export type LogEntry = {
    caller?: string;
    error?: string;
    fields?: {
        [key: string]: unknown;
    };
    level?: string;
    module?: string;
    msg?: string;
    raw?: string;
    time?: string;
    [key: string]: unknown;
};

export type Manifest = {
    audience?: string;
    path?: string;
    protocol_versions?: Array<string> | null;
    resource_templates?: Array<ManifestResourceTemplate> | null;
    resources?: Array<ManifestResource> | null;
    tools?: Array<ManifestTool> | null;
    transport?: string;
    [key: string]: unknown;
};

export type ManifestResource = {
    description?: string;
    name?: string;
    scopes?: Array<string> | null;
    title?: string;
    uri?: string;
    [key: string]: unknown;
};

export type ManifestResourceTemplate = {
    description?: string;
    name?: string;
    scopes?: Array<string> | null;
    title?: string;
    uri_template?: string;
    [key: string]: unknown;
};

export type ManifestTool = {
    description?: string;
    destructive?: boolean;
    name?: string;
    read_only?: boolean;
    scopes?: Array<string> | null;
    title?: string;
    [key: string]: unknown;
};

export type ModelResultCommentSystemSetting = {
    code?: number;
    data?: ModelSystemSetting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ModelSystemSetting = {
    captcha_enabled?: boolean;
    email_notify?: EmailNotifySetting;
    enable_comment?: boolean;
    require_approval?: boolean;
    [key: string]: unknown;
};

export type OAuth2Setting = {
    auth_redirect_allowed_return_urls?: Array<string> | null;
    auth_url?: string;
    client_id?: string;
    client_secret?: string;
    cors_allowed_origins?: Array<string> | null;
    enable?: boolean;
    is_oidc?: boolean;
    issuer?: string;
    jwks_url?: string;
    provider?: string;
    redirect_uri?: string;
    scopes?: Array<string> | null;
    token_url?: string;
    user_info_url?: string;
    [key: string]: unknown;
};

export type OAuth2SettingDto = {
    auth_redirect_allowed_return_urls?: Array<string> | null;
    auth_url?: string;
    client_id?: string;
    client_secret?: string;
    cors_allowed_origins?: Array<string> | null;
    enable?: boolean;
    is_oidc?: boolean;
    issuer?: string;
    jwks_url?: string;
    provider?: string;
    redirect_uri?: string;
    scopes?: Array<string> | null;
    token_url?: string;
    user_info_url?: string;
    [key: string]: unknown;
};

export type OAuth2Status = {
    enabled?: boolean;
    oauth_ready?: boolean;
    provider?: string;
    [key: string]: unknown;
};

export type OAuthBindBody = {
    /**
     * OAuth 回调地址
     */
    redirect_uri?: string;
    [key: string]: unknown;
};

export type OAuthInfoDto = {
    auth_type?: string;
    issuer?: string;
    oauth_id?: string;
    provider?: string;
    user_id?: string;
    [key: string]: unknown;
};

export type PageQueryDto = {
    page?: number;
    pageSize?: number;
    search?: string;
    [key: string]: unknown;
};

export type PageQueryResultListEcho = {
    items?: Array<Echo> | null;
    total?: number;
    [key: string]: unknown;
};

export type PageResultComment = {
    items?: Array<Comment> | null;
    total?: number;
    [key: string]: unknown;
};

export type PasskeyDeviceDto = {
    aaguid?: string;
    created_at?: number;
    device_name?: string;
    id?: string;
    last_used_at?: number;
    [key: string]: unknown;
};

export type PasskeySetting = {
    webauthn_allowed_origins?: Array<string> | null;
    webauthn_rp_id?: string;
    [key: string]: unknown;
};

export type PasskeySettingDto = {
    webauthn_allowed_origins?: Array<string> | null;
    webauthn_rp_id?: string;
    [key: string]: unknown;
};

export type PasskeyStatus = {
    passkey_ready?: boolean;
    [key: string]: unknown;
};

export type PasskeyUpdateDeviceNameReq = {
    device_name?: string;
    [key: string]: unknown;
};

export type PresignDto = {
    content_type?: string;
    file_name?: string;
    file_url?: string;
    id?: string;
    key?: string;
    presign_url?: string;
    [key: string]: unknown;
};

export type PublicComment = {
    content?: string;
    created_at?: number;
    echo_id?: string;
    hot?: boolean;
    id?: string;
    nickname?: string;
    parent_id?: string;
    source?: string;
    status?: string;
    updated_at?: number;
    website?: string;
    [key: string]: unknown;
};

export type RegisterDto = {
    email?: string;
    locale?: string;
    password?: string;
    username?: string;
    [key: string]: unknown;
};

export type ReindexStatusResponse = {
    /**
     * 失败原因（status=failed 时）
     */
    error?: string;
    /**
     * 结束时间（Unix 秒）
     */
    finished_at?: number;
    /**
     * 回填结果 BackfillResult: total/indexed/skipped/failed
     */
    payload?: unknown;
    /**
     * 当前阶段
     */
    phase?: string;
    /**
     * 开始时间（Unix 秒）
     */
    started_at?: number;
    /**
     * 作业状态：idle/pending/running/succeeded/failed/cancelled
     */
    status?: string;
    [key: string]: unknown;
};

export type ResultAgentSetting = {
    code?: number;
    data?: AgentSetting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultCheckUpdateResponse = {
    code?: number;
    data?: CheckUpdateResponse;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultComment = {
    code?: number;
    data?: Comment;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultConnect = {
    code?: number;
    data?: Connect;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultCreateCommentResult = {
    code?: number;
    data?: CreateCommentResult;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultEcho = {
    code?: number;
    data?: Echo;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultEmbeddingSetting = {
    code?: number;
    data?: EmbeddingSetting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultExportStateDto = {
    code?: number;
    data?: ExportStateDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultFileDto = {
    code?: number;
    data?: FileDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultFileListResultDto = {
    code?: number;
    data?: FileListResultDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultFileTreeResultDto = {
    code?: number;
    data?: FileTreeResultDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultFormMeta = {
    code?: number;
    data?: FormMeta;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultGlobalMigrationStateDto = {
    code?: number;
    data?: GlobalMigrationStateDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultHelloResponse = {
    code?: number;
    data?: HelloResponse;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultInterface = {
    code?: number;
    data?: unknown;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListAccessTokenSetting = {
    code?: number;
    data?: Array<AccessTokenSetting> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListChatMessage = {
    code?: number;
    data?: Array<ChatMessage> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListConnect = {
    code?: number;
    data?: Array<Connect> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListConnected = {
    code?: number;
    data?: Array<Connected> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListConnectedHealth = {
    code?: number;
    data?: Array<ConnectedHealth> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListDayStat = {
    code?: number;
    data?: Array<DayStat> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListEcho = {
    code?: number;
    data?: Array<Echo> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListHeatmap = {
    code?: number;
    data?: Array<Heatmap> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListLogEntry = {
    code?: number;
    data?: Array<LogEntry> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListPasskeyDeviceDto = {
    code?: number;
    data?: Array<PasskeyDeviceDto> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListPublicComment = {
    code?: number;
    data?: Array<PublicComment> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListTag = {
    code?: number;
    data?: Array<Tag> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListUser = {
    code?: number;
    data?: Array<User> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultListWebhook = {
    code?: number;
    data?: Array<Webhook> | null;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultManifest = {
    code?: number;
    data?: Manifest;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultOAuth2Setting = {
    code?: number;
    data?: OAuth2Setting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultOAuth2Status = {
    code?: number;
    data?: OAuth2Status;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultOAuthInfoDto = {
    code?: number;
    data?: OAuthInfoDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultPageQueryResultListEcho = {
    code?: number;
    data?: PageQueryResultListEcho;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultPageResultComment = {
    code?: number;
    data?: PageResultComment;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultPasskeySetting = {
    code?: number;
    data?: PasskeySetting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultPasskeyStatus = {
    code?: number;
    data?: PasskeyStatus;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultPresignDto = {
    code?: number;
    data?: PresignDto;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultReindexStatusResponse = {
    code?: number;
    data?: ReindexStatusResponse;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultS3Setting = {
    code?: number;
    data?: S3Setting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultSnapshotSchedule = {
    code?: number;
    data?: SnapshotSchedule;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultStatus = {
    code?: number;
    data?: Status;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultString = {
    code?: number;
    data?: string;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultSystemSetting = {
    code?: number;
    data?: SystemSetting;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultTag = {
    code?: number;
    data?: Tag;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type ResultUser = {
    code?: number;
    data?: User;
    error_code?: string;
    message_key?: string;
    message_params?: {
        [key: string]: unknown;
    };
    msg?: string;
    [key: string]: unknown;
};

export type S3Setting = {
    access_key?: string;
    bucket_name?: string;
    cdn_url?: string;
    enable?: boolean;
    endpoint?: string;
    path_prefix?: string;
    provider?: string;
    public_read?: boolean;
    region?: string;
    secret_key?: string;
    use_path_style?: boolean;
    use_ssl?: boolean;
    [key: string]: unknown;
};

export type S3SettingDto = {
    access_key?: string;
    bucket_name?: string;
    cdn_url?: string;
    enable?: boolean;
    endpoint?: string;
    path_prefix?: string;
    provider?: string;
    public_read?: boolean;
    region?: string;
    secret_key?: string;
    use_path_style?: boolean;
    use_ssl?: boolean;
    [key: string]: unknown;
};

export type SearchResult = {
    content?: string;
    distance?: number;
    echo_created?: number;
    echo_id?: string;
    extension?: EchoExtension;
    files?: Array<File> | null;
    username?: string;
    [key: string]: unknown;
};

export type SnapshotSchedule = {
    cron_expression?: string;
    enable?: boolean;
    [key: string]: unknown;
};

export type SnapshotScheduleDto = {
    cron_expression?: string;
    enable?: boolean;
    [key: string]: unknown;
};

export type StartExportRequest = {
    format?: string;
    include_private?: boolean;
    [key: string]: unknown;
};

export type StartGlobalMigrationRequest = {
    source_payload?: {
        [key: string]: unknown;
    };
    source_type?: string;
    [key: string]: unknown;
};

export type Status = {
    initialized?: boolean;
    owner_exists?: boolean;
    [key: string]: unknown;
};

export type SystemSetting = {
    ICP_number?: string;
    allow_register?: boolean;
    custom_css?: string;
    custom_js?: string;
    default_locale?: string;
    footer_content?: string;
    footer_link?: string;
    meting_api?: string;
    server_logo?: string;
    server_name?: string;
    server_url?: string;
    site_title?: string;
    [key: string]: unknown;
};

export type SystemSettingDto = {
    ICP_number?: string;
    allow_register?: boolean;
    custom_css?: string;
    custom_js?: string;
    default_locale?: string;
    footer_content?: string;
    footer_link?: string;
    meting_api?: string;
    server_logo?: string;
    server_logo_file_id?: string;
    server_name?: string;
    server_url?: string;
    site_title?: string;
    [key: string]: unknown;
};

export type Tag = {
    /**
     * 创建时间，提交时忽略
     */
    created_at?: number;
    /**
     * 标签 ID，提交时忽略
     */
    id?: string;
    /**
     * 标签名；同名标签复用
     */
    name?: string;
    /**
     * 引用次数，提交时忽略
     */
    usage_count?: number;
    [key: string]: unknown;
};

export type TestEmailRequest = {
    setting?: ModelSystemSetting;
    [key: string]: unknown;
};

export type UpdateCommentHotDto = {
    hot?: boolean;
    [key: string]: unknown;
};

export type UpdateCommentStatusDto = {
    status?: string;
    [key: string]: unknown;
};

export type UpdateFileMetaDto = {
    content_type?: string;
    height?: number;
    size?: number;
    width?: number;
    [key: string]: unknown;
};

export type User = {
    avatar?: string;
    email?: string;
    id?: string;
    is_admin?: boolean;
    is_owner?: boolean;
    locale?: string;
    username?: string;
    [key: string]: unknown;
};

export type UserInfoDto = {
    avatar?: string;
    avatar_file_id?: string;
    email?: string;
    is_admin?: boolean;
    is_owner?: boolean;
    locale?: string;
    password?: string;
    username?: string;
    [key: string]: unknown;
};

export type Webhook = {
    created_at?: number;
    id?: string;
    is_active?: boolean;
    last_status?: string;
    last_trigger?: number;
    name?: string;
    updated_at?: number;
    url?: string;
    [key: string]: unknown;
};

export type WebhookDto = {
    is_active?: boolean;
    name?: string;
    secret?: string;
    url?: string;
    [key: string]: unknown;
};

export type AccessTokenListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/access-tokens';
};

export type AccessTokenListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AccessTokenListError = AccessTokenListErrors[keyof AccessTokenListErrors];

export type AccessTokenListResponses = {
    /**
     * OK
     */
    200: ResultListAccessTokenSetting;
};

export type AccessTokenListResponse = AccessTokenListResponses[keyof AccessTokenListResponses];

export type AccessTokenCreateData = {
    body: AccessTokenSettingDto;
    path?: never;
    query?: never;
    url: '/access-tokens';
};

export type AccessTokenCreateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AccessTokenCreateError = AccessTokenCreateErrors[keyof AccessTokenCreateErrors];

export type AccessTokenCreateResponses = {
    /**
     * OK
     */
    200: ResultString;
};

export type AccessTokenCreateResponse = AccessTokenCreateResponses[keyof AccessTokenCreateResponses];

export type AccessTokenDeleteData = {
    body?: never;
    path: {
        /**
         * 资源 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/access-tokens/{id}';
};

export type AccessTokenDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AccessTokenDeleteError = AccessTokenDeleteErrors[keyof AccessTokenDeleteErrors];

export type AccessTokenDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type AccessTokenDeleteResponse = AccessTokenDeleteResponses[keyof AccessTokenDeleteResponses];

export type AgentInfoData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/agent/info';
};

export type AgentInfoErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AgentInfoError = AgentInfoErrors[keyof AgentInfoErrors];

export type AgentInfoResponses = {
    /**
     * OK
     */
    200: ResultAgentSetting;
};

export type AgentInfoResponse = AgentInfoResponses[keyof AgentInfoResponses];

export type CopilotRecentData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/agent/recent';
};

export type CopilotRecentErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CopilotRecentError = CopilotRecentErrors[keyof CopilotRecentErrors];

export type CopilotRecentResponses = {
    /**
     * OK
     */
    200: ResultString;
};

export type CopilotRecentResponse = CopilotRecentResponses[keyof CopilotRecentResponses];

export type AgentSettingsGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/agent/settings';
};

export type AgentSettingsGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AgentSettingsGetError = AgentSettingsGetErrors[keyof AgentSettingsGetErrors];

export type AgentSettingsGetResponses = {
    /**
     * OK
     */
    200: ResultAgentSetting;
};

export type AgentSettingsGetResponse = AgentSettingsGetResponses[keyof AgentSettingsGetResponses];

export type AgentSettingsUpdateData = {
    body: AgentSettingDto;
    path?: never;
    query?: never;
    url: '/agent/settings';
};

export type AgentSettingsUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AgentSettingsUpdateError = AgentSettingsUpdateErrors[keyof AgentSettingsUpdateErrors];

export type AgentSettingsUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type AgentSettingsUpdateResponse = AgentSettingsUpdateResponses[keyof AgentSettingsUpdateResponses];

export type AgentSettingsTestData = {
    body: AgentSettingDto;
    path?: never;
    query?: never;
    url: '/agent/settings/test';
};

export type AgentSettingsTestErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type AgentSettingsTestError = AgentSettingsTestErrors[keyof AgentSettingsTestErrors];

export type AgentSettingsTestResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type AgentSettingsTestResponse = AgentSettingsTestResponses[keyof AgentSettingsTestResponses];

export type CopilotChatAnswerData = {
    body: AnswerAskInputBody;
    path?: never;
    query?: never;
    url: '/chat/answer';
};

export type CopilotChatAnswerErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CopilotChatAnswerError = CopilotChatAnswerErrors[keyof CopilotChatAnswerErrors];

export type CopilotChatAnswerResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CopilotChatAnswerResponse = CopilotChatAnswerResponses[keyof CopilotChatAnswerResponses];

export type CopilotSessionClearData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/chat/session';
};

export type CopilotSessionClearErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CopilotSessionClearError = CopilotSessionClearErrors[keyof CopilotSessionClearErrors];

export type CopilotSessionClearResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CopilotSessionClearResponse = CopilotSessionClearResponses[keyof CopilotSessionClearResponses];

export type CopilotSessionGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/chat/session';
};

export type CopilotSessionGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CopilotSessionGetError = CopilotSessionGetErrors[keyof CopilotSessionGetErrors];

export type CopilotSessionGetResponses = {
    /**
     * OK
     */
    200: ResultListChatMessage;
};

export type CopilotSessionGetResponse = CopilotSessionGetResponses[keyof CopilotSessionGetResponses];

export type CommentListByEchoData = {
    body?: never;
    path?: never;
    query: {
        /**
         * Echo ID
         */
        echo_id: string;
    };
    url: '/comments';
};

export type CommentListByEchoErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentListByEchoError = CommentListByEchoErrors[keyof CommentListByEchoErrors];

export type CommentListByEchoResponses = {
    /**
     * OK
     */
    200: ResultListPublicComment;
};

export type CommentListByEchoResponse = CommentListByEchoResponses[keyof CommentListByEchoResponses];

export type CommentCreateData = {
    body: CreateCommentDto;
    path?: never;
    query?: never;
    url: '/comments';
};

export type CommentCreateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentCreateError = CommentCreateErrors[keyof CommentCreateErrors];

export type CommentCreateResponses = {
    /**
     * OK
     */
    200: ResultCreateCommentResult;
};

export type CommentCreateResponse = CommentCreateResponses[keyof CommentCreateResponses];

export type CommentFormMetaData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/comments/form';
};

export type CommentFormMetaErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentFormMetaError = CommentFormMetaErrors[keyof CommentFormMetaErrors];

export type CommentFormMetaResponses = {
    /**
     * OK
     */
    200: ResultFormMeta;
};

export type CommentFormMetaResponse = CommentFormMetaResponses[keyof CommentFormMetaResponses];

export type CommentCreateIntegrationData = {
    body: CreateIntegrationCommentDto;
    path?: never;
    query?: never;
    url: '/comments/integration';
};

export type CommentCreateIntegrationErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentCreateIntegrationError = CommentCreateIntegrationErrors[keyof CommentCreateIntegrationErrors];

export type CommentCreateIntegrationResponses = {
    /**
     * OK
     */
    200: ResultCreateCommentResult;
};

export type CommentCreateIntegrationResponse = CommentCreateIntegrationResponses[keyof CommentCreateIntegrationResponses];

export type CommentListPublicData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * 返回数量上限
         */
        limit?: number;
    };
    url: '/comments/public';
};

export type CommentListPublicErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentListPublicError = CommentListPublicErrors[keyof CommentListPublicErrors];

export type CommentListPublicResponses = {
    /**
     * OK
     */
    200: ResultListPublicComment;
};

export type CommentListPublicResponse = CommentListPublicResponses[keyof CommentListPublicResponses];

export type ConnectSelfData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/connect';
};

export type ConnectSelfErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectSelfError = ConnectSelfErrors[keyof ConnectSelfErrors];

export type ConnectSelfResponses = {
    /**
     * OK
     */
    200: ResultConnect;
};

export type ConnectSelfResponse = ConnectSelfResponses[keyof ConnectSelfResponses];

export type ConnectListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/connect/list';
};

export type ConnectListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectListError = ConnectListErrors[keyof ConnectListErrors];

export type ConnectListResponses = {
    /**
     * OK
     */
    200: ResultListConnected;
};

export type ConnectListResponse = ConnectListResponses[keyof ConnectListResponses];

export type ConnectAddData = {
    body: Connected;
    path?: never;
    query?: never;
    url: '/connects';
};

export type ConnectAddErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectAddError = ConnectAddErrors[keyof ConnectAddErrors];

export type ConnectAddResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type ConnectAddResponse = ConnectAddResponses[keyof ConnectAddResponses];

export type ConnectHealthData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/connects/health';
};

export type ConnectHealthErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectHealthError = ConnectHealthErrors[keyof ConnectHealthErrors];

export type ConnectHealthResponses = {
    /**
     * OK
     */
    200: ResultListConnectedHealth;
};

export type ConnectHealthResponse = ConnectHealthResponses[keyof ConnectHealthResponses];

export type ConnectInfoData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/connects/info';
};

export type ConnectInfoErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectInfoError = ConnectInfoErrors[keyof ConnectInfoErrors];

export type ConnectInfoResponses = {
    /**
     * OK
     */
    200: ResultListConnect;
};

export type ConnectInfoResponse = ConnectInfoResponses[keyof ConnectInfoResponses];

export type ConnectDeleteData = {
    body?: never;
    path: {
        /**
         * 连接 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/connects/{id}';
};

export type ConnectDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type ConnectDeleteError = ConnectDeleteErrors[keyof ConnectDeleteErrors];

export type ConnectDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type ConnectDeleteResponse = ConnectDeleteResponses[keyof ConnectDeleteResponses];

export type EchoCreateData = {
    body: EchoUpsertDto;
    path?: never;
    query?: never;
    url: '/echo';
};

export type EchoCreateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoCreateError = EchoCreateErrors[keyof EchoCreateErrors];

export type EchoCreateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type EchoCreateResponse = EchoCreateResponses[keyof EchoCreateResponses];

export type EchoUpdateData = {
    body: EchoUpsertDto;
    path?: never;
    query?: never;
    url: '/echo';
};

export type EchoUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoUpdateError = EchoUpdateErrors[keyof EchoUpdateErrors];

export type EchoUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type EchoUpdateResponse = EchoUpdateResponses[keyof EchoUpdateResponses];

export type EchoHotData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * 返回条数，默认 5，最大 20
         */
        limit?: number;
    };
    url: '/echo/hot';
};

export type EchoHotErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoHotError = EchoHotErrors[keyof EchoHotErrors];

export type EchoHotResponses = {
    /**
     * OK
     */
    200: ResultListEcho;
};

export type EchoHotResponse = EchoHotResponses[keyof EchoHotResponses];

export type EchoLikeData = {
    body?: never;
    path: {
        /**
         * Echo ID
         */
        id: string;
    };
    query?: never;
    url: '/echo/like/{id}';
};

export type EchoLikeErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoLikeError = EchoLikeErrors[keyof EchoLikeErrors];

export type EchoLikeResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type EchoLikeResponse = EchoLikeResponses[keyof EchoLikeResponses];

export type EchoOnthisdayData = {
    body?: never;
    headers?: {
        /**
         * 客户端时区（IANA 名）
         */
        'X-Timezone'?: string;
    };
    path?: never;
    query?: never;
    url: '/echo/onthisday';
};

export type EchoOnthisdayErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoOnthisdayError = EchoOnthisdayErrors[keyof EchoOnthisdayErrors];

export type EchoOnthisdayResponses = {
    /**
     * OK
     */
    200: ResultListEcho;
};

export type EchoOnthisdayResponse = EchoOnthisdayResponses[keyof EchoOnthisdayResponses];

export type EchoPageGetData = {
    body?: never;
    path?: never;
    query?: {
        page?: number;
        pageSize?: number;
        search?: string;
    };
    url: '/echo/page';
};

export type EchoPageGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoPageGetError = EchoPageGetErrors[keyof EchoPageGetErrors];

export type EchoPageGetResponses = {
    /**
     * OK
     */
    200: ResultPageQueryResultListEcho;
};

export type EchoPageGetResponse = EchoPageGetResponses[keyof EchoPageGetResponses];

export type EchoPagePostData = {
    body: PageQueryDto;
    path?: never;
    query?: never;
    url: '/echo/page';
};

export type EchoPagePostErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoPagePostError = EchoPagePostErrors[keyof EchoPagePostErrors];

export type EchoPagePostResponses = {
    /**
     * OK
     */
    200: ResultPageQueryResultListEcho;
};

export type EchoPagePostResponse = EchoPagePostResponses[keyof EchoPagePostResponses];

export type EchoQueryData = {
    body: EchoQueryDto;
    path?: never;
    query?: never;
    url: '/echo/query';
};

export type EchoQueryErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoQueryError = EchoQueryErrors[keyof EchoQueryErrors];

export type EchoQueryResponses = {
    /**
     * OK
     */
    200: ResultPageQueryResultListEcho;
};

export type EchoQueryResponse = EchoQueryResponses[keyof EchoQueryResponses];

export type EchoRandomData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/echo/random';
};

export type EchoRandomErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoRandomError = EchoRandomErrors[keyof EchoRandomErrors];

export type EchoRandomResponses = {
    /**
     * OK
     */
    200: ResultEcho;
};

export type EchoRandomResponse = EchoRandomResponses[keyof EchoRandomResponses];

export type EchoByTagData = {
    body?: never;
    path: {
        /**
         * 标签 ID
         */
        tagid: string;
    };
    query?: {
        page?: number;
        pageSize?: number;
        search?: string;
    };
    url: '/echo/tag/{tagid}';
};

export type EchoByTagErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoByTagError = EchoByTagErrors[keyof EchoByTagErrors];

export type EchoByTagResponses = {
    /**
     * OK
     */
    200: ResultPageQueryResultListEcho;
};

export type EchoByTagResponse = EchoByTagResponses[keyof EchoByTagResponses];

export type EchoTodayData = {
    body?: never;
    headers?: {
        /**
         * 客户端时区（IANA 名）
         */
        'X-Timezone'?: string;
    };
    path?: never;
    query?: never;
    url: '/echo/today';
};

export type EchoTodayErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoTodayError = EchoTodayErrors[keyof EchoTodayErrors];

export type EchoTodayResponses = {
    /**
     * OK
     */
    200: ResultListEcho;
};

export type EchoTodayResponse = EchoTodayResponses[keyof EchoTodayResponses];

export type EchoDeleteData = {
    body?: never;
    path: {
        /**
         * Echo ID
         */
        id: string;
    };
    query?: never;
    url: '/echo/{id}';
};

export type EchoDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoDeleteError = EchoDeleteErrors[keyof EchoDeleteErrors];

export type EchoDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type EchoDeleteResponse = EchoDeleteResponses[keyof EchoDeleteResponses];

export type EchoGetData = {
    body?: never;
    path: {
        /**
         * Echo ID
         */
        id: string;
    };
    query?: never;
    url: '/echo/{id}';
};

export type EchoGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EchoGetError = EchoGetErrors[keyof EchoGetErrors];

export type EchoGetResponses = {
    /**
     * OK
     */
    200: ResultEcho;
};

export type EchoGetResponse = EchoGetResponses[keyof EchoGetResponses];

export type EmbeddingReindexData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/embedding/reindex';
};

export type EmbeddingReindexErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EmbeddingReindexError = EmbeddingReindexErrors[keyof EmbeddingReindexErrors];

export type EmbeddingReindexResponses = {
    /**
     * OK
     */
    200: ResultReindexStatusResponse;
};

export type EmbeddingReindexResponse = EmbeddingReindexResponses[keyof EmbeddingReindexResponses];

export type EmbeddingReindexCancelData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/embedding/reindex/cancel';
};

export type EmbeddingReindexCancelErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EmbeddingReindexCancelError = EmbeddingReindexCancelErrors[keyof EmbeddingReindexCancelErrors];

export type EmbeddingReindexCancelResponses = {
    /**
     * OK
     */
    200: ResultReindexStatusResponse;
};

export type EmbeddingReindexCancelResponse = EmbeddingReindexCancelResponses[keyof EmbeddingReindexCancelResponses];

export type EmbeddingReindexStatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/embedding/reindex/status';
};

export type EmbeddingReindexStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EmbeddingReindexStatusError = EmbeddingReindexStatusErrors[keyof EmbeddingReindexStatusErrors];

export type EmbeddingReindexStatusResponses = {
    /**
     * OK
     */
    200: ResultReindexStatusResponse;
};

export type EmbeddingReindexStatusResponse = EmbeddingReindexStatusResponses[keyof EmbeddingReindexStatusResponses];

export type EmbeddingSettingsGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/embedding/settings';
};

export type EmbeddingSettingsGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EmbeddingSettingsGetError = EmbeddingSettingsGetErrors[keyof EmbeddingSettingsGetErrors];

export type EmbeddingSettingsGetResponses = {
    /**
     * OK
     */
    200: ResultEmbeddingSetting;
};

export type EmbeddingSettingsGetResponse = EmbeddingSettingsGetResponses[keyof EmbeddingSettingsGetResponses];

export type EmbeddingSettingsUpdateData = {
    body: EmbeddingSettingDto;
    path?: never;
    query?: never;
    url: '/embedding/settings';
};

export type EmbeddingSettingsUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type EmbeddingSettingsUpdateError = EmbeddingSettingsUpdateErrors[keyof EmbeddingSettingsUpdateErrors];

export type EmbeddingSettingsUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type EmbeddingSettingsUpdateResponse = EmbeddingSettingsUpdateResponses[keyof EmbeddingSettingsUpdateResponses];

export type FileTreeData = {
    body?: never;
    path?: never;
    query: {
        /**
         * 存储类型
         */
        storage_type: string;
        /**
         * 路径前缀
         */
        prefix?: string;
    };
    url: '/file/tree';
};

export type FileTreeErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileTreeError = FileTreeErrors[keyof FileTreeErrors];

export type FileTreeResponses = {
    /**
     * OK
     */
    200: ResultFileTreeResultDto;
};

export type FileTreeResponse = FileTreeResponses[keyof FileTreeResponses];

export type FileDeleteData = {
    body?: never;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: never;
    url: '/file/{id}';
};

export type FileDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileDeleteError = FileDeleteErrors[keyof FileDeleteErrors];

export type FileDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type FileDeleteResponse = FileDeleteResponses[keyof FileDeleteResponses];

export type FileGetData = {
    body?: never;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: never;
    url: '/file/{id}';
};

export type FileGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileGetError = FileGetErrors[keyof FileGetErrors];

export type FileGetResponses = {
    /**
     * OK
     */
    200: ResultFileDto;
};

export type FileGetResponse = FileGetResponses[keyof FileGetResponses];

export type FileUpdateMetaData = {
    body: UpdateFileMetaDto;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: never;
    url: '/file/{id}/meta';
};

export type FileUpdateMetaErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileUpdateMetaError = FileUpdateMetaErrors[keyof FileUpdateMetaErrors];

export type FileUpdateMetaResponses = {
    /**
     * OK
     */
    200: ResultFileDto;
};

export type FileUpdateMetaResponse = FileUpdateMetaResponses[keyof FileUpdateMetaResponses];

export type FileListData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * 页码
         */
        page?: number;
        /**
         * 每页数量
         */
        pageSize?: number;
        /**
         * 搜索关键词
         */
        search?: string;
        /**
         * 存储类型
         */
        storage_type?: string;
    };
    url: '/files';
};

export type FileListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileListError = FileListErrors[keyof FileListErrors];

export type FileListResponses = {
    /**
     * OK
     */
    200: ResultFileListResultDto;
};

export type FileListResponse = FileListResponses[keyof FileListResponses];

export type FileExternalData = {
    body: CreateExternalFileDto;
    path?: never;
    query?: never;
    url: '/files/external';
};

export type FileExternalErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FileExternalError = FileExternalErrors[keyof FileExternalErrors];

export type FileExternalResponses = {
    /**
     * OK
     */
    200: ResultFileDto;
};

export type FileExternalResponse = FileExternalResponses[keyof FileExternalResponses];

export type FilePresignData = {
    body: GetPresignUrlDto;
    path?: never;
    query?: never;
    url: '/files/presign';
};

export type FilePresignErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type FilePresignError = FilePresignErrors[keyof FilePresignErrors];

export type FilePresignResponses = {
    /**
     * OK
     */
    200: ResultPresignDto;
};

export type FilePresignResponse = FilePresignResponses[keyof FilePresignResponses];

export type CommonHeatmapData = {
    body?: never;
    headers?: {
        /**
         * 客户端时区（IANA 名），用于按本地日界对齐热力图
         */
        'X-Timezone'?: string;
    };
    path?: never;
    query?: never;
    url: '/heatmap';
};

export type CommonHeatmapErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommonHeatmapError = CommonHeatmapErrors[keyof CommonHeatmapErrors];

export type CommonHeatmapResponses = {
    /**
     * OK
     */
    200: ResultListHeatmap;
};

export type CommonHeatmapResponse = CommonHeatmapResponses[keyof CommonHeatmapResponses];

export type CommonHelloData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/hello';
};

export type CommonHelloErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommonHelloError = CommonHelloErrors[keyof CommonHelloErrors];

export type CommonHelloResponses = {
    /**
     * OK
     */
    200: ResultHelloResponse;
};

export type CommonHelloResponse = CommonHelloResponses[keyof CommonHelloResponses];

export type InitOwnerData = {
    body: RegisterDto;
    path?: never;
    query?: never;
    url: '/init/owner';
};

export type InitOwnerErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type InitOwnerError = InitOwnerErrors[keyof InitOwnerErrors];

export type InitOwnerResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type InitOwnerResponse = InitOwnerResponses[keyof InitOwnerResponses];

export type InitStatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/init/status';
};

export type InitStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type InitStatusError = InitStatusErrors[keyof InitStatusErrors];

export type InitStatusResponses = {
    /**
     * OK
     */
    200: ResultStatus;
};

export type InitStatusResponse = InitStatusResponses[keyof InitStatusResponses];

export type McpManifestData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/mcp/manifest';
};

export type McpManifestErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type McpManifestError = McpManifestErrors[keyof McpManifestErrors];

export type McpManifestResponses = {
    /**
     * OK
     */
    200: ResultManifest;
};

export type McpManifestResponse = McpManifestResponses[keyof McpManifestResponses];

export type MigrationCancelData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/migration/cancel';
};

export type MigrationCancelErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationCancelError = MigrationCancelErrors[keyof MigrationCancelErrors];

export type MigrationCancelResponses = {
    /**
     * OK
     */
    200: ResultGlobalMigrationStateDto;
};

export type MigrationCancelResponse = MigrationCancelResponses[keyof MigrationCancelResponses];

export type MigrationCleanupData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/migration/cleanup';
};

export type MigrationCleanupErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationCleanupError = MigrationCleanupErrors[keyof MigrationCleanupErrors];

export type MigrationCleanupResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type MigrationCleanupResponse = MigrationCleanupResponses[keyof MigrationCleanupResponses];

export type MigrationExportData = {
    body: StartExportRequest;
    path?: never;
    query?: never;
    url: '/migration/export';
};

export type MigrationExportErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationExportError = MigrationExportErrors[keyof MigrationExportErrors];

export type MigrationExportResponses = {
    /**
     * OK
     */
    200: ResultExportStateDto;
};

export type MigrationExportResponse = MigrationExportResponses[keyof MigrationExportResponses];

export type MigrationExportCancelData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/migration/export/cancel';
};

export type MigrationExportCancelErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationExportCancelError = MigrationExportCancelErrors[keyof MigrationExportCancelErrors];

export type MigrationExportCancelResponses = {
    /**
     * OK
     */
    200: ResultExportStateDto;
};

export type MigrationExportCancelResponse = MigrationExportCancelResponses[keyof MigrationExportCancelResponses];

export type MigrationExportStatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/migration/export/status';
};

export type MigrationExportStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationExportStatusError = MigrationExportStatusErrors[keyof MigrationExportStatusErrors];

export type MigrationExportStatusResponses = {
    /**
     * OK
     */
    200: ResultExportStateDto;
};

export type MigrationExportStatusResponse = MigrationExportStatusResponses[keyof MigrationExportStatusResponses];

export type MigrationStartData = {
    body: StartGlobalMigrationRequest;
    path?: never;
    query?: never;
    url: '/migration/start';
};

export type MigrationStartErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationStartError = MigrationStartErrors[keyof MigrationStartErrors];

export type MigrationStartResponses = {
    /**
     * OK
     */
    200: ResultGlobalMigrationStateDto;
};

export type MigrationStartResponse = MigrationStartResponses[keyof MigrationStartResponses];

export type MigrationStatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/migration/status';
};

export type MigrationStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type MigrationStatusError = MigrationStatusErrors[keyof MigrationStatusErrors];

export type MigrationStatusResponses = {
    /**
     * OK
     */
    200: ResultGlobalMigrationStateDto;
};

export type MigrationStatusResponse = MigrationStatusResponses[keyof MigrationStatusResponses];

export type OauthInfoData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * OAuth2 提供商，默认 github
         */
        provider?: string;
    };
    url: '/oauth/info';
};

export type OauthInfoErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type OauthInfoError = OauthInfoErrors[keyof OauthInfoErrors];

export type OauthInfoResponses = {
    /**
     * OK
     */
    200: ResultOAuthInfoDto;
};

export type OauthInfoResponse = OauthInfoResponses[keyof OauthInfoResponses];

export type OauthBindData = {
    body: OAuthBindBody;
    path: {
        /**
         * OAuth2 提供商 (github/google/qq/custom)
         */
        provider: string;
    };
    query?: never;
    url: '/oauth/{provider}/bind';
};

export type OauthBindErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type OauthBindError = OauthBindErrors[keyof OauthBindErrors];

export type OauthBindResponses = {
    /**
     * OK
     */
    200: ResultString;
};

export type OauthBindResponse = OauthBindResponses[keyof OauthBindResponses];

export type Oauth2GetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/oauth2/settings';
};

export type Oauth2GetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type Oauth2GetError = Oauth2GetErrors[keyof Oauth2GetErrors];

export type Oauth2GetResponses = {
    /**
     * OK
     */
    200: ResultOAuth2Setting;
};

export type Oauth2GetResponse = Oauth2GetResponses[keyof Oauth2GetResponses];

export type Oauth2UpdateData = {
    body: OAuth2SettingDto;
    path?: never;
    query?: never;
    url: '/oauth2/settings';
};

export type Oauth2UpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type Oauth2UpdateError = Oauth2UpdateErrors[keyof Oauth2UpdateErrors];

export type Oauth2UpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type Oauth2UpdateResponse = Oauth2UpdateResponses[keyof Oauth2UpdateResponses];

export type Oauth2StatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/oauth2/status';
};

export type Oauth2StatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type Oauth2StatusError = Oauth2StatusErrors[keyof Oauth2StatusErrors];

export type Oauth2StatusResponses = {
    /**
     * OK
     */
    200: ResultOAuth2Status;
};

export type Oauth2StatusResponse = Oauth2StatusResponses[keyof Oauth2StatusResponses];

export type CommentPanelListData = {
    body?: never;
    path?: never;
    query?: {
        page?: number;
        page_size?: number;
        keyword?: string;
        status?: string;
        echo_id?: string;
        /**
         * true/false 过滤置顶；缺省不过滤
         */
        hot?: string;
    };
    url: '/panel/comments';
};

export type CommentPanelListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelListError = CommentPanelListErrors[keyof CommentPanelListErrors];

export type CommentPanelListResponses = {
    /**
     * OK
     */
    200: ResultPageResultComment;
};

export type CommentPanelListResponse = CommentPanelListResponses[keyof CommentPanelListResponses];

export type CommentPanelBatchData = {
    body: BatchCommentActionDto;
    path?: never;
    query?: never;
    url: '/panel/comments/batch';
};

export type CommentPanelBatchErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelBatchError = CommentPanelBatchErrors[keyof CommentPanelBatchErrors];

export type CommentPanelBatchResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelBatchResponse = CommentPanelBatchResponses[keyof CommentPanelBatchResponses];

export type CommentPanelSettingsGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/panel/comments/settings';
};

export type CommentPanelSettingsGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelSettingsGetError = CommentPanelSettingsGetErrors[keyof CommentPanelSettingsGetErrors];

export type CommentPanelSettingsGetResponses = {
    /**
     * OK
     */
    200: ModelResultCommentSystemSetting;
};

export type CommentPanelSettingsGetResponse = CommentPanelSettingsGetResponses[keyof CommentPanelSettingsGetResponses];

export type CommentPanelSettingsUpdateData = {
    body: ModelSystemSetting;
    path?: never;
    query?: never;
    url: '/panel/comments/settings';
};

export type CommentPanelSettingsUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelSettingsUpdateError = CommentPanelSettingsUpdateErrors[keyof CommentPanelSettingsUpdateErrors];

export type CommentPanelSettingsUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelSettingsUpdateResponse = CommentPanelSettingsUpdateResponses[keyof CommentPanelSettingsUpdateResponses];

export type CommentPanelTestEmailData = {
    body: TestEmailRequest;
    path?: never;
    query?: never;
    url: '/panel/comments/settings/test-email';
};

export type CommentPanelTestEmailErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelTestEmailError = CommentPanelTestEmailErrors[keyof CommentPanelTestEmailErrors];

export type CommentPanelTestEmailResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelTestEmailResponse = CommentPanelTestEmailResponses[keyof CommentPanelTestEmailResponses];

export type CommentPanelDeleteData = {
    body?: never;
    path: {
        /**
         * 评论 ID
         */
        id: string;
    };
    query?: never;
    url: '/panel/comments/{id}';
};

export type CommentPanelDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelDeleteError = CommentPanelDeleteErrors[keyof CommentPanelDeleteErrors];

export type CommentPanelDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelDeleteResponse = CommentPanelDeleteResponses[keyof CommentPanelDeleteResponses];

export type CommentPanelGetData = {
    body?: never;
    path: {
        /**
         * 评论 ID
         */
        id: string;
    };
    query?: never;
    url: '/panel/comments/{id}';
};

export type CommentPanelGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelGetError = CommentPanelGetErrors[keyof CommentPanelGetErrors];

export type CommentPanelGetResponses = {
    /**
     * OK
     */
    200: ResultComment;
};

export type CommentPanelGetResponse = CommentPanelGetResponses[keyof CommentPanelGetResponses];

export type CommentPanelHotData = {
    body: UpdateCommentHotDto;
    path: {
        /**
         * 评论 ID
         */
        id: string;
    };
    query?: never;
    url: '/panel/comments/{id}/hot';
};

export type CommentPanelHotErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelHotError = CommentPanelHotErrors[keyof CommentPanelHotErrors];

export type CommentPanelHotResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelHotResponse = CommentPanelHotResponses[keyof CommentPanelHotResponses];

export type CommentPanelStatusData = {
    body: UpdateCommentStatusDto;
    path: {
        /**
         * 评论 ID
         */
        id: string;
    };
    query?: never;
    url: '/panel/comments/{id}/status';
};

export type CommentPanelStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommentPanelStatusError = CommentPanelStatusErrors[keyof CommentPanelStatusErrors];

export type CommentPanelStatusResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type CommentPanelStatusResponse = CommentPanelStatusResponses[keyof CommentPanelStatusResponses];

export type PasskeyGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/passkey/settings';
};

export type PasskeyGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyGetError = PasskeyGetErrors[keyof PasskeyGetErrors];

export type PasskeyGetResponses = {
    /**
     * OK
     */
    200: ResultPasskeySetting;
};

export type PasskeyGetResponse = PasskeyGetResponses[keyof PasskeyGetResponses];

export type PasskeyUpdateData = {
    body: PasskeySettingDto;
    path?: never;
    query?: never;
    url: '/passkey/settings';
};

export type PasskeyUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyUpdateError = PasskeyUpdateErrors[keyof PasskeyUpdateErrors];

export type PasskeyUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type PasskeyUpdateResponse = PasskeyUpdateResponses[keyof PasskeyUpdateResponses];

export type PasskeyStatusData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/passkey/status';
};

export type PasskeyStatusErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyStatusError = PasskeyStatusErrors[keyof PasskeyStatusErrors];

export type PasskeyStatusResponses = {
    /**
     * OK
     */
    200: ResultPasskeyStatus;
};

export type PasskeyStatusResponse = PasskeyStatusResponses[keyof PasskeyStatusResponses];

export type PasskeyListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/passkeys';
};

export type PasskeyListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyListError = PasskeyListErrors[keyof PasskeyListErrors];

export type PasskeyListResponses = {
    /**
     * OK
     */
    200: ResultListPasskeyDeviceDto;
};

export type PasskeyListResponse = PasskeyListResponses[keyof PasskeyListResponses];

export type PasskeyDeleteData = {
    body?: never;
    path: {
        /**
         * Passkey 设备 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/passkeys/{id}';
};

export type PasskeyDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyDeleteError = PasskeyDeleteErrors[keyof PasskeyDeleteErrors];

export type PasskeyDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type PasskeyDeleteResponse = PasskeyDeleteResponses[keyof PasskeyDeleteResponses];

export type PasskeyUpdateNameData = {
    body: PasskeyUpdateDeviceNameReq;
    path: {
        /**
         * Passkey 设备 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/passkeys/{id}';
};

export type PasskeyUpdateNameErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type PasskeyUpdateNameError = PasskeyUpdateNameErrors[keyof PasskeyUpdateNameErrors];

export type PasskeyUpdateNameResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type PasskeyUpdateNameResponse = PasskeyUpdateNameResponses[keyof PasskeyUpdateNameResponses];

export type UserRegisterData = {
    body: RegisterDto;
    path?: never;
    query?: never;
    url: '/register';
};

export type UserRegisterErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserRegisterError = UserRegisterErrors[keyof UserRegisterErrors];

export type UserRegisterResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type UserRegisterResponse = UserRegisterResponses[keyof UserRegisterResponses];

export type S3GetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/s3/settings';
};

export type S3GetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type S3GetError = S3GetErrors[keyof S3GetErrors];

export type S3GetResponses = {
    /**
     * OK
     */
    200: ResultS3Setting;
};

export type S3GetResponse = S3GetResponses[keyof S3GetResponses];

export type S3UpdateData = {
    body: S3SettingDto;
    path?: never;
    query?: never;
    url: '/s3/settings';
};

export type S3UpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type S3UpdateError = S3UpdateErrors[keyof S3UpdateErrors];

export type S3UpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type S3UpdateResponse = S3UpdateResponses[keyof S3UpdateResponses];

export type S3TestData = {
    body: S3SettingDto;
    path?: never;
    query?: never;
    url: '/s3/settings/test';
};

export type S3TestErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type S3TestError = S3TestErrors[keyof S3TestErrors];

export type S3TestResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type S3TestResponse = S3TestResponses[keyof S3TestResponses];

export type SettingsGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/settings';
};

export type SettingsGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type SettingsGetError = SettingsGetErrors[keyof SettingsGetErrors];

export type SettingsGetResponses = {
    /**
     * OK
     */
    200: ResultSystemSetting;
};

export type SettingsGetResponse = SettingsGetResponses[keyof SettingsGetResponses];

export type SettingsUpdateData = {
    body: SystemSettingDto;
    path?: never;
    query?: never;
    url: '/settings';
};

export type SettingsUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type SettingsUpdateError = SettingsUpdateErrors[keyof SettingsUpdateErrors];

export type SettingsUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type SettingsUpdateResponse = SettingsUpdateResponses[keyof SettingsUpdateResponses];

export type SnapshotScheduleGetData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/snapshot/schedule';
};

export type SnapshotScheduleGetErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type SnapshotScheduleGetError = SnapshotScheduleGetErrors[keyof SnapshotScheduleGetErrors];

export type SnapshotScheduleGetResponses = {
    /**
     * OK
     */
    200: ResultSnapshotSchedule;
};

export type SnapshotScheduleGetResponse = SnapshotScheduleGetResponses[keyof SnapshotScheduleGetResponses];

export type SnapshotScheduleUpdateData = {
    body: SnapshotScheduleDto;
    path?: never;
    query?: never;
    url: '/snapshot/schedule';
};

export type SnapshotScheduleUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type SnapshotScheduleUpdateError = SnapshotScheduleUpdateErrors[keyof SnapshotScheduleUpdateErrors];

export type SnapshotScheduleUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type SnapshotScheduleUpdateResponse = SnapshotScheduleUpdateResponses[keyof SnapshotScheduleUpdateResponses];

export type DashboardCheckUpdateData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/system/check-update';
};

export type DashboardCheckUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type DashboardCheckUpdateError = DashboardCheckUpdateErrors[keyof DashboardCheckUpdateErrors];

export type DashboardCheckUpdateResponses = {
    /**
     * OK
     */
    200: ResultCheckUpdateResponse;
};

export type DashboardCheckUpdateResponse = DashboardCheckUpdateResponses[keyof DashboardCheckUpdateResponses];

export type DashboardSystemLogsData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * 返回最近多少条（默认 200）
         */
        tail?: string;
        /**
         * 日志级别过滤
         */
        level?: string;
        /**
         * 关键词过滤
         */
        keyword?: string;
    };
    url: '/system/logs';
};

export type DashboardSystemLogsErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type DashboardSystemLogsError = DashboardSystemLogsErrors[keyof DashboardSystemLogsErrors];

export type DashboardSystemLogsResponses = {
    /**
     * OK
     */
    200: ResultListLogEntry;
};

export type DashboardSystemLogsResponse = DashboardSystemLogsResponses[keyof DashboardSystemLogsResponses];

export type DashboardVisitorStatsData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/system/visitor-stats';
};

export type DashboardVisitorStatsErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type DashboardVisitorStatsError = DashboardVisitorStatsErrors[keyof DashboardVisitorStatsErrors];

export type DashboardVisitorStatsResponses = {
    /**
     * OK
     */
    200: ResultListDayStat;
};

export type DashboardVisitorStatsResponse = DashboardVisitorStatsResponses[keyof DashboardVisitorStatsResponses];

export type TagCreateData = {
    body: CreateTagDto;
    path?: never;
    query?: never;
    url: '/tag';
};

export type TagCreateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type TagCreateError = TagCreateErrors[keyof TagCreateErrors];

export type TagCreateResponses = {
    /**
     * OK
     */
    200: ResultTag;
};

export type TagCreateResponse = TagCreateResponses[keyof TagCreateResponses];

export type TagDeleteData = {
    body?: never;
    path: {
        /**
         * 标签 ID
         */
        id: string;
    };
    query?: never;
    url: '/tag/{id}';
};

export type TagDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type TagDeleteError = TagDeleteErrors[keyof TagDeleteErrors];

export type TagDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type TagDeleteResponse = TagDeleteResponses[keyof TagDeleteResponses];

export type TagListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/tags';
};

export type TagListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type TagListError = TagListErrors[keyof TagListErrors];

export type TagListResponses = {
    /**
     * OK
     */
    200: ResultListTag;
};

export type TagListResponse = TagListResponses[keyof TagListResponses];

export type UserInfoData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/user';
};

export type UserInfoErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserInfoError = UserInfoErrors[keyof UserInfoErrors];

export type UserInfoResponses = {
    /**
     * OK
     */
    200: ResultUser;
};

export type UserInfoResponse = UserInfoResponses[keyof UserInfoResponses];

export type UserUpdateData = {
    body: UserInfoDto;
    path?: never;
    query?: never;
    url: '/user';
};

export type UserUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserUpdateError = UserUpdateErrors[keyof UserUpdateErrors];

export type UserUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type UserUpdateResponse = UserUpdateResponses[keyof UserUpdateResponses];

export type UserSetAdminData = {
    body?: never;
    path: {
        /**
         * 用户 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/user/admin/{id}';
};

export type UserSetAdminErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserSetAdminError = UserSetAdminErrors[keyof UserSetAdminErrors];

export type UserSetAdminResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type UserSetAdminResponse = UserSetAdminResponses[keyof UserSetAdminResponses];

export type UserDeleteData = {
    body?: never;
    path: {
        /**
         * 用户 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/user/{id}';
};

export type UserDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserDeleteError = UserDeleteErrors[keyof UserDeleteErrors];

export type UserDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type UserDeleteResponse = UserDeleteResponses[keyof UserDeleteResponses];

export type UserListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/users';
};

export type UserListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type UserListError = UserListErrors[keyof UserListErrors];

export type UserListResponses = {
    /**
     * OK
     */
    200: ResultListUser;
};

export type UserListResponse = UserListResponses[keyof UserListResponses];

export type WebhookListData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/webhook';
};

export type WebhookListErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type WebhookListError = WebhookListErrors[keyof WebhookListErrors];

export type WebhookListResponses = {
    /**
     * OK
     */
    200: ResultListWebhook;
};

export type WebhookListResponse = WebhookListResponses[keyof WebhookListResponses];

export type WebhookCreateData = {
    body: WebhookDto;
    path?: never;
    query?: never;
    url: '/webhook';
};

export type WebhookCreateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type WebhookCreateError = WebhookCreateErrors[keyof WebhookCreateErrors];

export type WebhookCreateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type WebhookCreateResponse = WebhookCreateResponses[keyof WebhookCreateResponses];

export type WebhookDeleteData = {
    body?: never;
    path: {
        /**
         * 资源 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/webhook/{id}';
};

export type WebhookDeleteErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type WebhookDeleteError = WebhookDeleteErrors[keyof WebhookDeleteErrors];

export type WebhookDeleteResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type WebhookDeleteResponse = WebhookDeleteResponses[keyof WebhookDeleteResponses];

export type WebhookUpdateData = {
    body: WebhookDto;
    path: {
        /**
         * Webhook ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/webhook/{id}';
};

export type WebhookUpdateErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type WebhookUpdateError = WebhookUpdateErrors[keyof WebhookUpdateErrors];

export type WebhookUpdateResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type WebhookUpdateResponse = WebhookUpdateResponses[keyof WebhookUpdateResponses];

export type WebhookTestData = {
    body?: never;
    path: {
        /**
         * 资源 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/webhook/{id}/test';
};

export type WebhookTestErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type WebhookTestError = WebhookTestErrors[keyof WebhookTestErrors];

export type WebhookTestResponses = {
    /**
     * OK
     */
    200: ResultInterface;
};

export type WebhookTestResponse = WebhookTestResponses[keyof WebhookTestResponses];

export type CommonWebsiteTitleData = {
    body?: never;
    path?: never;
    query: {
        /**
         * 目标网站 URL
         */
        website_url: string;
    };
    url: '/website/title';
};

export type CommonWebsiteTitleErrors = {
    /**
     * Error
     */
    default: ErrorBody;
};

export type CommonWebsiteTitleError = CommonWebsiteTitleErrors[keyof CommonWebsiteTitleErrors];

export type CommonWebsiteTitleResponses = {
    /**
     * OK
     */
    200: ResultString;
};

export type CommonWebsiteTitleResponse = CommonWebsiteTitleResponses[keyof CommonWebsiteTitleResponses];
