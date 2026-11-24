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

export type AddConnectDto = {
    /**
     * Peer instance root URL; the server calls <url>/api/connect
     */
    connect_url: string;
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

export type CreateIntegrationCommentDto = {
    /**
     * Comment body, at most 200 characters
     */
    content: string;
    /**
     * Echo ID the comment belongs to
     */
    echo_id: string;
    /**
     * Caller-defined audit string; recorded in server logs only
     */
    metadata?: string;
    /**
     * Display name; defaults to Integration
     */
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
    deleted_at?: string;
    echo_files?: Array<EchoFile> | null;
    extension?: EchoExtension;
    id?: string;
    layout?: string;
    private?: boolean;
    reactions?: Array<ReactionCount> | null;
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
     * Arbitrary fields the card needs to render
     */
    payload?: {
        [key: string]: unknown;
    };
    /**
     * Card type, e.g. weather or music
     */
    type?: string;
    [key: string]: unknown;
};

export type EchoFile = {
    file?: File;
    file_id?: string;
    sort_order?: number;
    [key: string]: unknown;
};

export type EchoQueryDto = {
    /**
     * Keyset pagination cursor; omit for page-based pagination, send an empty string for the first cursor page. Cursor mode ignores page and requires sortBy=created_at
     */
    cursor?: string;
    /**
     * Only Echoes created at or after this Unix time (seconds)
     */
    dateFrom?: number;
    /**
     * Only Echoes created at or before this Unix time (seconds)
     */
    dateTo?: number;
    /**
     * Only Echoes authored by the authenticated caller; requires sign-in
     */
    mine?: boolean;
    /**
     * Page number, 1-based (default 1); ignored when cursor is set
     */
    page?: number;
    /**
     * Items per page, 1-100 (default 10); also the cursor page size
     */
    pageSize?: number;
    /**
     * Visibility filter: true for private only, false for public only; omit for both
     */
    private?: boolean;
    /**
     * Substring match on Echo content
     */
    search?: string;
    /**
     * Sort field (default created_at); cursor pagination requires created_at
     */
    sortBy?: 'created_at' | 'reactions';
    /**
     * Sort direction (default desc)
     */
    sortOrder?: 'asc' | 'desc';
    /**
     * Tag IDs to match; an Echo qualifies if it carries any listed tag
     */
    tagIds?: Array<string> | null;
    [key: string]: unknown;
};

export type EchoUpsertDto = {
    /**
     * Markdown body; supply at least one of content, files, or extension
     */
    content?: string;
    /**
     * Publish time in Unix seconds; defaults to now
     */
    created_at?: number;
    /**
     * Attachments as {file_id, sort_order}; upload files first to get file_id. Omit on update to keep existing attachments; send [] to clear
     */
    echo_files?: Array<EchoFile>;
    /**
     * Card payload: {type, payload}
     */
    extension?: EchoExtensionDto;
    /**
     * Echo ID to update; ignored on create
     */
    id?: string;
    /**
     * Layout: waterfall / grid / horizontal / carousel / stack / none
     */
    layout?: string;
    /**
     * Hidden from anonymous visitors; visible to all signed-in users
     */
    private?: boolean;
    /**
     * Tags as {name}; an existing tag with the same name is reused
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

export type ErrorDetail = {
    /**
     * Where the error occurred, e.g. 'body.items[3].tags' or 'path.thing-id'
     */
    location?: string;
    /**
     * Error message text
     */
    message?: string;
    /**
     * The value at the given location
     */
    value?: unknown;
    [key: string]: unknown;
};

export type ErrorModel = {
    /**
     * A human-readable explanation specific to this occurrence of the problem.
     */
    detail?: string;
    /**
     * Optional list of individual error details
     */
    errors?: Array<ErrorDetail> | null;
    /**
     * A URI reference that identifies the specific occurrence of the problem.
     */
    instance?: string;
    /**
     * HTTP status code
     */
    status?: number;
    /**
     * A short, human-readable summary of the problem type. This value should not change between occurrences of the error.
     */
    title?: string;
    /**
     * A URI reference to human-readable documentation for the error.
     */
    type?: string;
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
    access?: string;
    blob_bytes?: number;
    content_type?: string;
    created_at?: number;
    height?: number;
    id?: string;
    kind?: string;
    name?: string;
    object_key?: string;
    ref_url?: string;
    size?: number;
    status?: string;
    storage_id?: string;
    updated_at?: number;
    uploader_id?: string;
    width?: number;
    [key: string]: unknown;
};

export type FileDto = {
    access?: string;
    category?: string;
    content_type?: string;
    created_at?: number;
    height?: number;
    id?: string;
    kind?: string;
    links?: Array<LinkDto> | null;
    name?: string;
    size?: number;
    status?: string;
    uploader_id?: string;
    url?: string;
    width?: number;
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

export type InitOwnerDto = {
    email?: string;
    locale?: string;
    password?: string;
    site_title?: string;
    username?: string;
    [key: string]: unknown;
};

export type LinkDto = {
    owner_id?: string;
    owner_kind?: string;
    role?: string;
    sort_order?: number;
    [key: string]: unknown;
};

export type ListAccessTokenSetting = {
    /**
     * 当前页数据
     */
    items?: Array<AccessTokenSetting> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListComment = {
    /**
     * 当前页数据
     */
    items?: Array<Comment> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListConnect = {
    /**
     * 当前页数据
     */
    items?: Array<Connect> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListConnected = {
    /**
     * 当前页数据
     */
    items?: Array<Connected> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListConnectedHealth = {
    /**
     * 当前页数据
     */
    items?: Array<ConnectedHealth> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListDayStat = {
    /**
     * 当前页数据
     */
    items?: Array<DayStat> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListEcho = {
    /**
     * 当前页数据
     */
    items?: Array<Echo> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListFileDto = {
    /**
     * 当前页数据
     */
    items?: Array<FileDto> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListHeatmap = {
    /**
     * 当前页数据
     */
    items?: Array<Heatmap> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListLogEntry = {
    /**
     * 当前页数据
     */
    items?: Array<LogEntry> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListPublicComment = {
    /**
     * 当前页数据
     */
    items?: Array<PublicComment> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListReactionCount = {
    /**
     * 当前页数据
     */
    items?: Array<ReactionCount> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListReactionEmoji = {
    /**
     * 当前页数据
     */
    items?: Array<ReactionEmoji> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListTag = {
    /**
     * 当前页数据
     */
    items?: Array<Tag> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListUser = {
    /**
     * 当前页数据
     */
    items?: Array<User> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
    [key: string]: unknown;
};

export type ListWebhook = {
    /**
     * 当前页数据
     */
    items?: Array<Webhook> | null;
    /**
     * 下一页游标，仅游标分页返回；为空表示没有更多
     */
    nextCursor?: string;
    /**
     * 符合查询条件的总条数，仅偏移分页返回
     */
    total?: number;
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

export type ReactionCount = {
    count?: number;
    emoji?: string;
    key?: string;
    reacted?: boolean;
    [key: string]: unknown;
};

export type ReactionEmoji = {
    emoji?: string;
    key?: string;
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

export type RenameFileInputBody = {
    /**
     * 展示名称（可为空）
     */
    name?: string;
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
    enable_agent_discovery?: boolean;
    footer_content?: string;
    footer_link?: string;
    meting_api?: string;
    privacy_policy?: string;
    server_logo?: string;
    server_logo_file_id?: string;
    server_name?: string;
    server_url?: string;
    show_privacy_footer?: boolean;
    site_title?: string;
    [key: string]: unknown;
};

export type SystemSettingDto = {
    ICP_number?: string;
    allow_register?: boolean;
    custom_css?: string;
    custom_js?: string;
    default_locale?: string;
    enable_agent_discovery?: boolean;
    footer_content?: string;
    footer_link?: string;
    meting_api?: string;
    privacy_policy?: string;
    server_logo?: string;
    server_logo_file_id?: string;
    server_name?: string;
    server_url?: string;
    show_privacy_footer?: boolean;
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

export type UpdateUserRoleInputBody = {
    /**
     * 目标角色：admin 或 member
     */
    role?: string;
    [key: string]: unknown;
};

export type UpdateUserStatusInputBody = {
    /**
     * 账户状态：ACTIVE 或 DISABLED
     */
    status?: string;
    [key: string]: unknown;
};

export type User = {
    avatar?: string;
    avatar_file_id?: string;
    email?: string;
    id?: string;
    locale?: string;
    role?: string;
    status?: string;
    username?: string;
    [key: string]: unknown;
};

export type UserInfoDto = {
    avatar?: string;
    avatar_file_id?: string;
    email?: string;
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
    /**
     * Enabled; defaults to true on create, while omitting it on update disables the webhook
     */
    is_active?: boolean;
    /**
     * Display name
     */
    name: string;
    /**
     * Optional signing secret; the server signs payloads with X-Ech0-Signature
     */
    secret?: string;
    /**
     * HTTP(S) endpoint that receives events
     */
    url: string;
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
    default: ErrorModel;
};

export type AccessTokenListError = AccessTokenListErrors[keyof AccessTokenListErrors];

export type AccessTokenListResponses = {
    /**
     * OK
     */
    200: ListAccessTokenSetting;
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
    default: ErrorModel;
};

export type AccessTokenCreateError = AccessTokenCreateErrors[keyof AccessTokenCreateErrors];

export type AccessTokenCreateResponses = {
    /**
     * Created
     */
    201: string;
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
    default: ErrorModel;
};

export type AccessTokenDeleteError = AccessTokenDeleteErrors[keyof AccessTokenDeleteErrors];

export type AccessTokenDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type AgentInfoError = AgentInfoErrors[keyof AgentInfoErrors];

export type AgentInfoResponses = {
    /**
     * OK
     */
    200: AgentSetting;
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
    default: ErrorModel;
};

export type CopilotRecentError = CopilotRecentErrors[keyof CopilotRecentErrors];

export type CopilotRecentResponses = {
    /**
     * OK
     */
    200: string;
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
    default: ErrorModel;
};

export type AgentSettingsGetError = AgentSettingsGetErrors[keyof AgentSettingsGetErrors];

export type AgentSettingsGetResponses = {
    /**
     * OK
     */
    200: AgentSetting;
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
    default: ErrorModel;
};

export type AgentSettingsUpdateError = AgentSettingsUpdateErrors[keyof AgentSettingsUpdateErrors];

export type AgentSettingsUpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type AgentSettingsTestError = AgentSettingsTestErrors[keyof AgentSettingsTestErrors];

export type AgentSettingsTestResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CopilotChatAnswerError = CopilotChatAnswerErrors[keyof CopilotChatAnswerErrors];

export type CopilotChatAnswerResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CopilotSessionClearError = CopilotSessionClearErrors[keyof CopilotSessionClearErrors];

export type CopilotSessionClearResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CopilotSessionGetError = CopilotSessionGetErrors[keyof CopilotSessionGetErrors];

export type CopilotSessionGetResponses = {
    /**
     * OK
     */
    200: Array<ChatMessage> | null;
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
    default: ErrorModel;
};

export type CommentListByEchoError = CommentListByEchoErrors[keyof CommentListByEchoErrors];

export type CommentListByEchoResponses = {
    /**
     * OK
     */
    200: ListPublicComment;
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
    default: ErrorModel;
};

export type CommentCreateError = CommentCreateErrors[keyof CommentCreateErrors];

export type CommentCreateResponses = {
    /**
     * Created
     */
    201: Comment;
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
    default: ErrorModel;
};

export type CommentFormMetaError = CommentFormMetaErrors[keyof CommentFormMetaErrors];

export type CommentFormMetaResponses = {
    /**
     * OK
     */
    200: FormMeta;
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
    default: ErrorModel;
};

export type CommentCreateIntegrationError = CommentCreateIntegrationErrors[keyof CommentCreateIntegrationErrors];

export type CommentCreateIntegrationResponses = {
    /**
     * Created
     */
    201: Comment;
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
    default: ErrorModel;
};

export type CommentListPublicError = CommentListPublicErrors[keyof CommentListPublicErrors];

export type CommentListPublicResponses = {
    /**
     * OK
     */
    200: ListPublicComment;
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
    default: ErrorModel;
};

export type ConnectSelfError = ConnectSelfErrors[keyof ConnectSelfErrors];

export type ConnectSelfResponses = {
    /**
     * OK
     */
    200: Connect;
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
    default: ErrorModel;
};

export type ConnectListError = ConnectListErrors[keyof ConnectListErrors];

export type ConnectListResponses = {
    /**
     * OK
     */
    200: ListConnected;
};

export type ConnectListResponse = ConnectListResponses[keyof ConnectListResponses];

export type ConnectAddData = {
    body: AddConnectDto;
    path?: never;
    query?: never;
    url: '/connects';
};

export type ConnectAddErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type ConnectAddError = ConnectAddErrors[keyof ConnectAddErrors];

export type ConnectAddResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type ConnectHealthError = ConnectHealthErrors[keyof ConnectHealthErrors];

export type ConnectHealthResponses = {
    /**
     * OK
     */
    200: ListConnectedHealth;
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
    default: ErrorModel;
};

export type ConnectInfoError = ConnectInfoErrors[keyof ConnectInfoErrors];

export type ConnectInfoResponses = {
    /**
     * OK
     */
    200: ListConnect;
};

export type ConnectInfoResponse = ConnectInfoResponses[keyof ConnectInfoResponses];

export type ConnectDeleteData = {
    body?: never;
    path: {
        /**
         * Connection ID (UUID)
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
    default: ErrorModel;
};

export type ConnectDeleteError = ConnectDeleteErrors[keyof ConnectDeleteErrors];

export type ConnectDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type EchoCreateError = EchoCreateErrors[keyof EchoCreateErrors];

export type EchoCreateResponses = {
    /**
     * Created
     */
    201: Echo;
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
    default: ErrorModel;
};

export type EchoUpdateError = EchoUpdateErrors[keyof EchoUpdateErrors];

export type EchoUpdateResponses = {
    /**
     * OK
     */
    200: Echo;
};

export type EchoUpdateResponse = EchoUpdateResponses[keyof EchoUpdateResponses];

export type EchoHotData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * Number of Echoes to return; default 5, maximum 20
         */
        limit?: number;
    };
    url: '/echo/hot';
};

export type EchoHotErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type EchoHotError = EchoHotErrors[keyof EchoHotErrors];

export type EchoHotResponses = {
    /**
     * OK
     */
    200: ListEcho;
};

export type EchoHotResponse = EchoHotResponses[keyof EchoHotResponses];

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
    default: ErrorModel;
};

export type EchoOnthisdayError = EchoOnthisdayErrors[keyof EchoOnthisdayErrors];

export type EchoOnthisdayResponses = {
    /**
     * OK
     */
    200: ListEcho;
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
    default: ErrorModel;
};

export type EchoPageGetError = EchoPageGetErrors[keyof EchoPageGetErrors];

export type EchoPageGetResponses = {
    /**
     * OK
     */
    200: ListEcho;
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
    default: ErrorModel;
};

export type EchoPagePostError = EchoPagePostErrors[keyof EchoPagePostErrors];

export type EchoPagePostResponses = {
    /**
     * OK
     */
    200: ListEcho;
};

export type EchoPagePostResponse = EchoPagePostResponses[keyof EchoPagePostResponses];

export type EchoListData = {
    body: EchoQueryDto;
    path?: never;
    query?: never;
    url: '/echo/query';
};

export type EchoListErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type EchoListError = EchoListErrors[keyof EchoListErrors];

export type EchoListResponses = {
    /**
     * OK
     */
    200: ListEcho;
};

export type EchoListResponse = EchoListResponses[keyof EchoListResponses];

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
    default: ErrorModel;
};

export type EchoRandomError = EchoRandomErrors[keyof EchoRandomErrors];

export type EchoRandomResponses = {
    /**
     * OK
     */
    200: Echo;
};

export type EchoRandomResponse = EchoRandomResponses[keyof EchoRandomResponses];

export type EchoByTagData = {
    body?: never;
    path: {
        /**
         * Tag ID
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
    default: ErrorModel;
};

export type EchoByTagError = EchoByTagErrors[keyof EchoByTagErrors];

export type EchoByTagResponses = {
    /**
     * OK
     */
    200: ListEcho;
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
    default: ErrorModel;
};

export type EchoTodayError = EchoTodayErrors[keyof EchoTodayErrors];

export type EchoTodayResponses = {
    /**
     * OK
     */
    200: ListEcho;
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
    default: ErrorModel;
};

export type EchoDeleteError = EchoDeleteErrors[keyof EchoDeleteErrors];

export type EchoDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type EchoGetError = EchoGetErrors[keyof EchoGetErrors];

export type EchoGetResponses = {
    /**
     * OK
     */
    200: Echo;
};

export type EchoGetResponse = EchoGetResponses[keyof EchoGetResponses];

export type ReactionRemoveData = {
    body?: never;
    path: {
        /**
         * Echo ID
         */
        id: string;
        /**
         * Reaction key
         */
        emoji: 'like' | 'heart' | 'haha' | 'wow' | 'sad' | 'cry' | 'party';
    };
    query?: never;
    url: '/echo/{id}/reactions/{emoji}';
};

export type ReactionRemoveErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type ReactionRemoveError = ReactionRemoveErrors[keyof ReactionRemoveErrors];

export type ReactionRemoveResponses = {
    /**
     * OK
     */
    200: ListReactionCount;
};

export type ReactionRemoveResponse = ReactionRemoveResponses[keyof ReactionRemoveResponses];

export type ReactionAddData = {
    body?: never;
    path: {
        /**
         * Echo ID
         */
        id: string;
        /**
         * Reaction key
         */
        emoji: 'like' | 'heart' | 'haha' | 'wow' | 'sad' | 'cry' | 'party';
    };
    query?: never;
    url: '/echo/{id}/reactions/{emoji}';
};

export type ReactionAddErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type ReactionAddError = ReactionAddErrors[keyof ReactionAddErrors];

export type ReactionAddResponses = {
    /**
     * OK
     */
    200: ListReactionCount;
};

export type ReactionAddResponse = ReactionAddResponses[keyof ReactionAddResponses];

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
    default: ErrorModel;
};

export type EmbeddingReindexError = EmbeddingReindexErrors[keyof EmbeddingReindexErrors];

export type EmbeddingReindexResponses = {
    /**
     * OK
     */
    200: ReindexStatusResponse;
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
    default: ErrorModel;
};

export type EmbeddingReindexCancelError = EmbeddingReindexCancelErrors[keyof EmbeddingReindexCancelErrors];

export type EmbeddingReindexCancelResponses = {
    /**
     * OK
     */
    200: ReindexStatusResponse;
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
    default: ErrorModel;
};

export type EmbeddingReindexStatusError = EmbeddingReindexStatusErrors[keyof EmbeddingReindexStatusErrors];

export type EmbeddingReindexStatusResponses = {
    /**
     * OK
     */
    200: ReindexStatusResponse;
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
    default: ErrorModel;
};

export type EmbeddingSettingsGetError = EmbeddingSettingsGetErrors[keyof EmbeddingSettingsGetErrors];

export type EmbeddingSettingsGetResponses = {
    /**
     * OK
     */
    200: EmbeddingSetting;
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
    default: ErrorModel;
};

export type EmbeddingSettingsUpdateError = EmbeddingSettingsUpdateErrors[keyof EmbeddingSettingsUpdateErrors];

export type EmbeddingSettingsUpdateResponses = {
    /**
     * No Content
     */
    204: void;
};

export type EmbeddingSettingsUpdateResponse = EmbeddingSettingsUpdateResponses[keyof EmbeddingSettingsUpdateResponses];

export type FileListData = {
    body?: never;
    path?: never;
    query?: {
        /**
         * 页码，从 1 开始
         */
        page?: number;
        /**
         * 每页数量，1-100
         */
        pageSize?: number;
        /**
         * 文件名称模糊匹配
         */
        search?: string;
        /**
         * owned / blob / ref
         */
        kind?: string;
        /**
         * pending / ready
         */
        status?: string;
        /**
         * public / owner / private
         */
        access?: string;
        /**
         * 引用方类型，如 echo
         */
        ownerKind?: string;
        /**
         * 引用方 ID
         */
        ownerId?: string;
        /**
         * 引用角色，如 attachment
         */
        role?: string;
    };
    url: '/files';
};

export type FileListErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type FileListError = FileListErrors[keyof FileListErrors];

export type FileListResponses = {
    /**
     * OK
     */
    200: ListFileDto;
};

export type FileListResponse = FileListResponses[keyof FileListResponses];

export type FileDeleteData = {
    body?: never;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: {
        /**
         * 管理员强制删除仍有引用的文件
         */
        force?: boolean;
    };
    url: '/files/{id}';
};

export type FileDeleteErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type FileDeleteError = FileDeleteErrors[keyof FileDeleteErrors];

export type FileDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    url: '/files/{id}';
};

export type FileGetErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type FileGetError = FileGetErrors[keyof FileGetErrors];

export type FileGetResponses = {
    /**
     * OK
     */
    200: FileDto;
};

export type FileGetResponse = FileGetResponses[keyof FileGetResponses];

export type FileRenameData = {
    body: RenameFileInputBody;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: never;
    url: '/files/{id}';
};

export type FileRenameErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type FileRenameError = FileRenameErrors[keyof FileRenameErrors];

export type FileRenameResponses = {
    /**
     * OK
     */
    200: FileDto;
};

export type FileRenameResponse = FileRenameResponses[keyof FileRenameResponses];

export type FileCommitData = {
    body?: never;
    path: {
        /**
         * 文件 ID
         */
        id: string;
    };
    query?: never;
    url: '/files/{id}/commit';
};

export type FileCommitErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type FileCommitError = FileCommitErrors[keyof FileCommitErrors];

export type FileCommitResponses = {
    /**
     * OK
     */
    200: FileDto;
};

export type FileCommitResponse = FileCommitResponses[keyof FileCommitResponses];

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
    default: ErrorModel;
};

export type CommonHeatmapError = CommonHeatmapErrors[keyof CommonHeatmapErrors];

export type CommonHeatmapResponses = {
    /**
     * OK
     */
    200: ListHeatmap;
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
    default: ErrorModel;
};

export type CommonHelloError = CommonHelloErrors[keyof CommonHelloErrors];

export type CommonHelloResponses = {
    /**
     * OK
     */
    200: HelloResponse;
};

export type CommonHelloResponse = CommonHelloResponses[keyof CommonHelloResponses];

export type InitOwnerData = {
    body: InitOwnerDto;
    path?: never;
    query?: never;
    url: '/init/owner';
};

export type InitOwnerErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type InitOwnerError = InitOwnerErrors[keyof InitOwnerErrors];

export type InitOwnerResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type InitStatusError = InitStatusErrors[keyof InitStatusErrors];

export type InitStatusResponses = {
    /**
     * OK
     */
    200: Status;
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
    default: ErrorModel;
};

export type McpManifestError = McpManifestErrors[keyof McpManifestErrors];

export type McpManifestResponses = {
    /**
     * OK
     */
    200: Manifest;
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
    default: ErrorModel;
};

export type MigrationCancelError = MigrationCancelErrors[keyof MigrationCancelErrors];

export type MigrationCancelResponses = {
    /**
     * OK
     */
    200: GlobalMigrationStateDto;
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
    default: ErrorModel;
};

export type MigrationCleanupError = MigrationCleanupErrors[keyof MigrationCleanupErrors];

export type MigrationCleanupResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type MigrationExportError = MigrationExportErrors[keyof MigrationExportErrors];

export type MigrationExportResponses = {
    /**
     * OK
     */
    200: ExportStateDto;
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
    default: ErrorModel;
};

export type MigrationExportCancelError = MigrationExportCancelErrors[keyof MigrationExportCancelErrors];

export type MigrationExportCancelResponses = {
    /**
     * OK
     */
    200: ExportStateDto;
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
    default: ErrorModel;
};

export type MigrationExportStatusError = MigrationExportStatusErrors[keyof MigrationExportStatusErrors];

export type MigrationExportStatusResponses = {
    /**
     * OK
     */
    200: ExportStateDto;
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
    default: ErrorModel;
};

export type MigrationStartError = MigrationStartErrors[keyof MigrationStartErrors];

export type MigrationStartResponses = {
    /**
     * OK
     */
    200: GlobalMigrationStateDto;
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
    default: ErrorModel;
};

export type MigrationStatusError = MigrationStatusErrors[keyof MigrationStatusErrors];

export type MigrationStatusResponses = {
    /**
     * OK
     */
    200: GlobalMigrationStateDto;
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
    default: ErrorModel;
};

export type OauthInfoError = OauthInfoErrors[keyof OauthInfoErrors];

export type OauthInfoResponses = {
    /**
     * OK
     */
    200: OAuthInfoDto;
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
    default: ErrorModel;
};

export type OauthBindError = OauthBindErrors[keyof OauthBindErrors];

export type OauthBindResponses = {
    /**
     * OK
     */
    200: string;
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
    default: ErrorModel;
};

export type Oauth2GetError = Oauth2GetErrors[keyof Oauth2GetErrors];

export type Oauth2GetResponses = {
    /**
     * OK
     */
    200: OAuth2Setting;
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
    default: ErrorModel;
};

export type Oauth2UpdateError = Oauth2UpdateErrors[keyof Oauth2UpdateErrors];

export type Oauth2UpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type Oauth2StatusError = Oauth2StatusErrors[keyof Oauth2StatusErrors];

export type Oauth2StatusResponses = {
    /**
     * OK
     */
    200: OAuth2Status;
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
    default: ErrorModel;
};

export type CommentPanelListError = CommentPanelListErrors[keyof CommentPanelListErrors];

export type CommentPanelListResponses = {
    /**
     * OK
     */
    200: ListComment;
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
    default: ErrorModel;
};

export type CommentPanelBatchError = CommentPanelBatchErrors[keyof CommentPanelBatchErrors];

export type CommentPanelBatchResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommentPanelSettingsGetError = CommentPanelSettingsGetErrors[keyof CommentPanelSettingsGetErrors];

export type CommentPanelSettingsGetResponses = {
    /**
     * OK
     */
    200: ModelSystemSetting;
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
    default: ErrorModel;
};

export type CommentPanelSettingsUpdateError = CommentPanelSettingsUpdateErrors[keyof CommentPanelSettingsUpdateErrors];

export type CommentPanelSettingsUpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommentPanelTestEmailError = CommentPanelTestEmailErrors[keyof CommentPanelTestEmailErrors];

export type CommentPanelTestEmailResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommentPanelDeleteError = CommentPanelDeleteErrors[keyof CommentPanelDeleteErrors];

export type CommentPanelDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommentPanelGetError = CommentPanelGetErrors[keyof CommentPanelGetErrors];

export type CommentPanelGetResponses = {
    /**
     * OK
     */
    200: Comment;
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
    default: ErrorModel;
};

export type CommentPanelHotError = CommentPanelHotErrors[keyof CommentPanelHotErrors];

export type CommentPanelHotResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommentPanelStatusError = CommentPanelStatusErrors[keyof CommentPanelStatusErrors];

export type CommentPanelStatusResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type PasskeyGetError = PasskeyGetErrors[keyof PasskeyGetErrors];

export type PasskeyGetResponses = {
    /**
     * OK
     */
    200: PasskeySetting;
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
    default: ErrorModel;
};

export type PasskeyUpdateError = PasskeyUpdateErrors[keyof PasskeyUpdateErrors];

export type PasskeyUpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type PasskeyStatusError = PasskeyStatusErrors[keyof PasskeyStatusErrors];

export type PasskeyStatusResponses = {
    /**
     * OK
     */
    200: PasskeyStatus;
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
    default: ErrorModel;
};

export type PasskeyListError = PasskeyListErrors[keyof PasskeyListErrors];

export type PasskeyListResponses = {
    /**
     * OK
     */
    200: Array<PasskeyDeviceDto> | null;
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
    default: ErrorModel;
};

export type PasskeyDeleteError = PasskeyDeleteErrors[keyof PasskeyDeleteErrors];

export type PasskeyDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type PasskeyUpdateNameError = PasskeyUpdateNameErrors[keyof PasskeyUpdateNameErrors];

export type PasskeyUpdateNameResponses = {
    /**
     * No Content
     */
    204: void;
};

export type PasskeyUpdateNameResponse = PasskeyUpdateNameResponses[keyof PasskeyUpdateNameResponses];

export type ReactionCatalogData = {
    body?: never;
    path?: never;
    query?: never;
    url: '/reactions';
};

export type ReactionCatalogErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type ReactionCatalogError = ReactionCatalogErrors[keyof ReactionCatalogErrors];

export type ReactionCatalogResponses = {
    /**
     * OK
     */
    200: ListReactionEmoji;
};

export type ReactionCatalogResponse = ReactionCatalogResponses[keyof ReactionCatalogResponses];

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
    default: ErrorModel;
};

export type UserRegisterError = UserRegisterErrors[keyof UserRegisterErrors];

export type UserRegisterResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type S3GetError = S3GetErrors[keyof S3GetErrors];

export type S3GetResponses = {
    /**
     * OK
     */
    200: S3Setting;
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
    default: ErrorModel;
};

export type S3UpdateError = S3UpdateErrors[keyof S3UpdateErrors];

export type S3UpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type S3TestError = S3TestErrors[keyof S3TestErrors];

export type S3TestResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type SettingsGetError = SettingsGetErrors[keyof SettingsGetErrors];

export type SettingsGetResponses = {
    /**
     * OK
     */
    200: SystemSetting;
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
    default: ErrorModel;
};

export type SettingsUpdateError = SettingsUpdateErrors[keyof SettingsUpdateErrors];

export type SettingsUpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type SnapshotScheduleGetError = SnapshotScheduleGetErrors[keyof SnapshotScheduleGetErrors];

export type SnapshotScheduleGetResponses = {
    /**
     * OK
     */
    200: SnapshotSchedule;
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
    default: ErrorModel;
};

export type SnapshotScheduleUpdateError = SnapshotScheduleUpdateErrors[keyof SnapshotScheduleUpdateErrors];

export type SnapshotScheduleUpdateResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type DashboardCheckUpdateError = DashboardCheckUpdateErrors[keyof DashboardCheckUpdateErrors];

export type DashboardCheckUpdateResponses = {
    /**
     * OK
     */
    200: CheckUpdateResponse;
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
    default: ErrorModel;
};

export type DashboardSystemLogsError = DashboardSystemLogsErrors[keyof DashboardSystemLogsErrors];

export type DashboardSystemLogsResponses = {
    /**
     * OK
     */
    200: ListLogEntry;
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
    default: ErrorModel;
};

export type DashboardVisitorStatsError = DashboardVisitorStatsErrors[keyof DashboardVisitorStatsErrors];

export type DashboardVisitorStatsResponses = {
    /**
     * OK
     */
    200: ListDayStat;
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
    default: ErrorModel;
};

export type TagCreateError = TagCreateErrors[keyof TagCreateErrors];

export type TagCreateResponses = {
    /**
     * OK
     */
    200: Tag;
};

export type TagCreateResponse = TagCreateResponses[keyof TagCreateResponses];

export type TagDeleteData = {
    body?: never;
    path: {
        /**
         * Tag ID
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
    default: ErrorModel;
};

export type TagDeleteError = TagDeleteErrors[keyof TagDeleteErrors];

export type TagDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type TagListError = TagListErrors[keyof TagListErrors];

export type TagListResponses = {
    /**
     * OK
     */
    200: ListTag;
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
    default: ErrorModel;
};

export type UserInfoError = UserInfoErrors[keyof UserInfoErrors];

export type UserInfoResponses = {
    /**
     * OK
     */
    200: User;
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
    default: ErrorModel;
};

export type UserUpdateError = UserUpdateErrors[keyof UserUpdateErrors];

export type UserUpdateResponses = {
    /**
     * No Content
     */
    204: void;
};

export type UserUpdateResponse = UserUpdateResponses[keyof UserUpdateResponses];

export type UserTransferOwnerData = {
    body?: never;
    path: {
        /**
         * 接任所有者的用户 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/user/owner/{id}';
};

export type UserTransferOwnerErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type UserTransferOwnerError = UserTransferOwnerErrors[keyof UserTransferOwnerErrors];

export type UserTransferOwnerResponses = {
    /**
     * No Content
     */
    204: void;
};

export type UserTransferOwnerResponse = UserTransferOwnerResponses[keyof UserTransferOwnerResponses];

export type UserSetRoleData = {
    body: UpdateUserRoleInputBody;
    path: {
        /**
         * 用户 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/user/role/{id}';
};

export type UserSetRoleErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type UserSetRoleError = UserSetRoleErrors[keyof UserSetRoleErrors];

export type UserSetRoleResponses = {
    /**
     * No Content
     */
    204: void;
};

export type UserSetRoleResponse = UserSetRoleResponses[keyof UserSetRoleResponses];

export type UserSetStatusData = {
    body: UpdateUserStatusInputBody;
    path: {
        /**
         * 用户 ID（UUID）
         */
        id: string;
    };
    query?: never;
    url: '/user/status/{id}';
};

export type UserSetStatusErrors = {
    /**
     * Error
     */
    default: ErrorModel;
};

export type UserSetStatusError = UserSetStatusErrors[keyof UserSetStatusErrors];

export type UserSetStatusResponses = {
    /**
     * No Content
     */
    204: void;
};

export type UserSetStatusResponse = UserSetStatusResponses[keyof UserSetStatusResponses];

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
    default: ErrorModel;
};

export type UserDeleteError = UserDeleteErrors[keyof UserDeleteErrors];

export type UserDeleteResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type UserListError = UserListErrors[keyof UserListErrors];

export type UserListResponses = {
    /**
     * OK
     */
    200: ListUser;
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
    default: ErrorModel;
};

export type WebhookListError = WebhookListErrors[keyof WebhookListErrors];

export type WebhookListResponses = {
    /**
     * OK
     */
    200: ListWebhook;
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
    default: ErrorModel;
};

export type WebhookCreateError = WebhookCreateErrors[keyof WebhookCreateErrors];

export type WebhookCreateResponses = {
    /**
     * No Content
     */
    204: void;
};

export type WebhookCreateResponse = WebhookCreateResponses[keyof WebhookCreateResponses];

export type WebhookDeleteData = {
    body?: never;
    path: {
        /**
         * Webhook ID (UUID)
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
    default: ErrorModel;
};

export type WebhookDeleteError = WebhookDeleteErrors[keyof WebhookDeleteErrors];

export type WebhookDeleteResponses = {
    /**
     * No Content
     */
    204: void;
};

export type WebhookDeleteResponse = WebhookDeleteResponses[keyof WebhookDeleteResponses];

export type WebhookUpdateData = {
    body: WebhookDto;
    path: {
        /**
         * Webhook ID (UUID)
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
    default: ErrorModel;
};

export type WebhookUpdateError = WebhookUpdateErrors[keyof WebhookUpdateErrors];

export type WebhookUpdateResponses = {
    /**
     * No Content
     */
    204: void;
};

export type WebhookUpdateResponse = WebhookUpdateResponses[keyof WebhookUpdateResponses];

export type WebhookTestData = {
    body?: never;
    path: {
        /**
         * Webhook ID (UUID)
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
    default: ErrorModel;
};

export type WebhookTestError = WebhookTestErrors[keyof WebhookTestErrors];

export type WebhookTestResponses = {
    /**
     * No Content
     */
    204: void;
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
    default: ErrorModel;
};

export type CommonWebsiteTitleError = CommonWebsiteTitleErrors[keyof CommonWebsiteTitleErrors];

export type CommonWebsiteTitleResponses = {
    /**
     * OK
     */
    200: string;
};

export type CommonWebsiteTitleResponse = CommonWebsiteTitleResponses[keyof CommonWebsiteTitleResponses];
