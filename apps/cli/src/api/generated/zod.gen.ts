// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

import * as z from 'zod';

export const zAccessTokenSetting = z.object({
    audience: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    expiry: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).nullish(),
    id: z.string().optional(),
    jti: z.string().optional(),
    last_used_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    name: z.string().optional(),
    scopes: z.string().optional(),
    token: z.string().optional(),
    token_type: z.string().optional(),
    user_id: z.string().optional()
});

export const zAccessTokenSettingDto = z.object({
    audience: z.string().optional(),
    expiry: z.string().optional(),
    name: z.string().optional(),
    scopes: z.array(z.string()).nullish()
});

export const zAgentSetting = z.object({
    api_key: z.string().optional(),
    base_url: z.string().optional(),
    context_window: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    enable: z.boolean().optional(),
    model: z.string().optional(),
    multimodal: z.boolean().optional(),
    prompt: z.string().optional(),
    protocol: z.string().optional(),
    temperature: z.number().optional()
});

export const zAgentSettingDto = z.object({
    api_key: z.string().optional(),
    base_url: z.string().optional(),
    context_window: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    enable: z.boolean().optional(),
    model: z.string().optional(),
    multimodal: z.boolean().optional(),
    prompt: z.string().optional(),
    protocol: z.string().optional(),
    temperature: z.number().gte(0).lte(2).nullish()
});

export const zAskAnswer = z.object({
    custom: z.string().optional(),
    question_id: z.string().optional(),
    selected: z.array(z.string()).nullish()
});

export const zAnswerAskInputBody = z.object({
    answers: z.array(zAskAnswer).min(1).nullish(),
    ask_id: z.string().min(1).optional()
});

export const zAskOption = z.object({
    description: z.string().optional(),
    label: z.string().optional()
});

export const zAskQuestion = z.object({
    detail: z.string().optional(),
    header: z.string().optional(),
    id: z.string().optional(),
    multi: z.boolean().optional(),
    options: z.array(zAskOption).nullish(),
    recommended: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    text: z.string().optional()
});

export const zAskExchange = z.object({
    answers: z.array(zAskAnswer).nullish(),
    questions: z.array(zAskQuestion).nullish()
});

export const zBatchCommentActionDto = z.object({
    action: z.string().optional(),
    ids: z.array(z.string()).nullish()
});

export const zCheckUpdateResponse = z.object({
    current_version: z.string().optional(),
    has_update: z.boolean().optional(),
    latest_version: z.string().optional()
});

export const zComment = z.object({
    content: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_id: z.string().optional(),
    email: z.string().optional(),
    hot: z.boolean().optional(),
    id: z.string().optional(),
    nickname: z.string().optional(),
    parent_id: z.string().optional(),
    source: z.string().optional(),
    status: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    user_id: z.string().optional(),
    website: z.string().optional()
});

export const zConnect = z.object({
    logo: z.string().optional(),
    server_name: z.string().optional(),
    server_url: z.string().optional(),
    sys_username: z.string().optional(),
    today_echos: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    total_echos: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    version: z.string().optional()
});

export const zConnected = z.object({
    connect_url: z.string().optional(),
    id: z.string().optional()
});

export const zConnectedHealth = z.object({
    connect_url: z.string().optional(),
    id: z.string().optional(),
    status: z.string().optional(),
    version: z.string().optional()
});

export const zCreateCommentDto = z.object({
    captcha_token: z.string().optional(),
    content: z.string().optional(),
    echo_id: z.string().optional(),
    email: z.string().optional(),
    form_token: z.string().optional(),
    hp_field: z.string().optional(),
    nickname: z.string().optional(),
    parent_id: z.string().optional(),
    website: z.string().optional()
});

export const zCreateCommentResult = z.object({
    id: z.string().optional(),
    status: z.string().optional()
});

export const zCreateExternalFileDto = z.object({
    category: z.string().optional(),
    content_type: z.string().optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    name: z.string().optional(),
    url: z.string().optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zCreateIntegrationCommentDto = z.object({
    content: z.string().optional(),
    echo_id: z.string().optional(),
    metadata: z.string().optional(),
    nickname: z.string().optional()
});

export const zCreateTagDto = z.object({
    name: z.string().optional()
});

export const zDayStat = z.object({
    date: z.string().optional(),
    pv: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    uv: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zEchoExtension = z.object({
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_id: z.string().optional(),
    id: z.string().optional(),
    payload: z.record(z.string(), z.unknown()).optional(),
    type: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zEchoExtensionDto = z.object({
    payload: z.record(z.string(), z.unknown()).optional(),
    type: z.string().optional()
});

export const zEchoQueryDto = z.object({
    dateFrom: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    dateTo: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    private: z.boolean().optional(),
    search: z.string().optional(),
    sortBy: z.string().optional(),
    sortOrder: z.string().optional(),
    tagIds: z.array(z.string()).nullish()
});

export const zEmailNotifySetting = z.object({
    enabled: z.boolean().optional(),
    smtp_host: z.string().optional(),
    smtp_password: z.string().optional(),
    smtp_password_set: z.boolean().optional(),
    smtp_port: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    smtp_sender: z.string().optional(),
    smtp_username: z.string().optional()
});

export const zEmbeddingSetting = z.object({
    api_key: z.string().optional(),
    base_url: z.string().optional(),
    batch_size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    dim: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    enable: z.boolean().optional(),
    model: z.string().optional()
});

export const zEmbeddingSettingDto = z.object({
    api_key: z.string().optional(),
    base_url: z.string().optional(),
    batch_size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    dim: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    enable: z.boolean().optional(),
    model: z.string().optional()
});

export const zErrorBody = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.unknown().optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zExportStateDto = z.object({
    error_message: z.string().optional(),
    file_name: z.string().optional(),
    finished_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    format: z.string().optional(),
    phase: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    started_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    status: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    version: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFile = z.object({
    bucket: z.string().optional(),
    category: z.string().optional(),
    content_type: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    key: z.string().optional(),
    name: z.string().optional(),
    provider: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    storage_type: z.string().optional(),
    url: z.string().optional(),
    user_id: z.string().optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zEchoFile = z.object({
    echo_id: z.string().optional(),
    file: zFile.optional(),
    file_id: z.string().optional(),
    id: z.string().optional(),
    sort_order: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFileDto = z.object({
    category: z.string().optional(),
    content_type: z.string().optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    key: z.string().optional(),
    name: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    storage_type: z.string().optional(),
    url: z.string().optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFileListItemDto = z.object({
    content_type: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    key: z.string().optional(),
    name: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    storage_type: z.string().optional(),
    url: z.string().optional()
});

export const zFileListResultDto = z.object({
    items: z.array(zFileListItemDto).nullish(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFileTreeNodeDto = z.object({
    content_type: z.string().optional(),
    file_id: z.string().optional(),
    has_children: z.boolean().optional(),
    modified_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    name: z.string().optional(),
    node_type: z.string().optional(),
    path: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFileTreeResultDto = z.object({
    items: z.array(zFileTreeNodeDto).nullish()
});

export const zFormMeta = z.object({
    captcha_api_endpoint: z.string().optional(),
    captcha_enabled: z.boolean().optional(),
    enable_comment: z.boolean().optional(),
    form_token: z.string().optional(),
    min_submit_ms: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zGetPresignUrlDto = z.object({
    content_type: z.string().optional(),
    file_name: z.string().optional(),
    storage_type: z.string().optional()
});

export const zGlobalMigrationStateDto = z.object({
    error_message: z.string().optional(),
    finished_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    phase: z.string().optional(),
    source_payload: z.record(z.string(), z.unknown()).optional(),
    source_type: z.string().optional(),
    started_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    status: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    version: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zHeatmap = z.object({
    count: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    date: z.string().optional()
});

export const zHelloResponse = z.object({
    author: z.string().optional(),
    build_time: z.string().optional(),
    commit: z.string().optional(),
    copyright: z.string().optional(),
    hello: z.string().optional(),
    license: z.string().optional(),
    repo_url: z.string().optional(),
    version: z.string().optional()
});

export const zLogEntry = z.object({
    caller: z.string().optional(),
    error: z.string().optional(),
    fields: z.record(z.string(), z.unknown()).optional(),
    level: z.string().optional(),
    module: z.string().optional(),
    msg: z.string().optional(),
    raw: z.string().optional(),
    time: z.string().optional()
});

export const zManifestResource = z.object({
    description: z.string().optional(),
    name: z.string().optional(),
    scopes: z.array(z.string()).nullish(),
    title: z.string().optional(),
    uri: z.string().optional()
});

export const zManifestResourceTemplate = z.object({
    description: z.string().optional(),
    name: z.string().optional(),
    scopes: z.array(z.string()).nullish(),
    title: z.string().optional(),
    uri_template: z.string().optional()
});

export const zManifestTool = z.object({
    description: z.string().optional(),
    destructive: z.boolean().optional(),
    name: z.string().optional(),
    read_only: z.boolean().optional(),
    scopes: z.array(z.string()).nullish(),
    title: z.string().optional()
});

export const zManifest = z.object({
    audience: z.string().optional(),
    path: z.string().optional(),
    protocol_versions: z.array(z.string()).nullish(),
    resource_templates: z.array(zManifestResourceTemplate).nullish(),
    resources: z.array(zManifestResource).nullish(),
    tools: z.array(zManifestTool).nullish(),
    transport: z.string().optional()
});

export const zModelSystemSetting = z.object({
    captcha_enabled: z.boolean().optional(),
    email_notify: zEmailNotifySetting.optional(),
    enable_comment: z.boolean().optional(),
    require_approval: z.boolean().optional()
});

export const zModelResultCommentSystemSetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zModelSystemSetting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zOAuth2Setting = z.object({
    auth_redirect_allowed_return_urls: z.array(z.string()).nullish(),
    auth_url: z.string().optional(),
    client_id: z.string().optional(),
    client_secret: z.string().optional(),
    cors_allowed_origins: z.array(z.string()).nullish(),
    enable: z.boolean().optional(),
    is_oidc: z.boolean().optional(),
    issuer: z.string().optional(),
    jwks_url: z.string().optional(),
    provider: z.string().optional(),
    redirect_uri: z.string().optional(),
    scopes: z.array(z.string()).nullish(),
    token_url: z.string().optional(),
    user_info_url: z.string().optional()
});

export const zOAuth2SettingDto = z.object({
    auth_redirect_allowed_return_urls: z.array(z.string()).nullish(),
    auth_url: z.string().optional(),
    client_id: z.string().optional(),
    client_secret: z.string().optional(),
    cors_allowed_origins: z.array(z.string()).nullish(),
    enable: z.boolean().optional(),
    is_oidc: z.boolean().optional(),
    issuer: z.string().optional(),
    jwks_url: z.string().optional(),
    provider: z.string().optional(),
    redirect_uri: z.string().optional(),
    scopes: z.array(z.string()).nullish(),
    token_url: z.string().optional(),
    user_info_url: z.string().optional()
});

export const zOAuth2Status = z.object({
    enabled: z.boolean().optional(),
    oauth_ready: z.boolean().optional(),
    provider: z.string().optional()
});

export const zOAuthBindBody = z.object({
    redirect_uri: z.string().optional()
});

export const zOAuthInfoDto = z.object({
    auth_type: z.string().optional(),
    issuer: z.string().optional(),
    oauth_id: z.string().optional(),
    provider: z.string().optional(),
    user_id: z.string().optional()
});

export const zPageQueryDto = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional()
});

export const zPageResultComment = z.object({
    items: z.array(zComment).nullish(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zPasskeyDeviceDto = z.object({
    aaguid: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    device_name: z.string().optional(),
    id: z.string().optional(),
    last_used_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zPasskeySetting = z.object({
    webauthn_allowed_origins: z.array(z.string()).nullish(),
    webauthn_rp_id: z.string().optional()
});

export const zPasskeySettingDto = z.object({
    webauthn_allowed_origins: z.array(z.string()).nullish(),
    webauthn_rp_id: z.string().optional()
});

export const zPasskeyStatus = z.object({
    passkey_ready: z.boolean().optional()
});

export const zPasskeyUpdateDeviceNameReq = z.object({
    device_name: z.string().optional()
});

export const zPresignDto = z.object({
    content_type: z.string().optional(),
    file_name: z.string().optional(),
    file_url: z.string().optional(),
    id: z.string().optional(),
    key: z.string().optional(),
    presign_url: z.string().optional()
});

export const zPublicComment = z.object({
    content: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_id: z.string().optional(),
    hot: z.boolean().optional(),
    id: z.string().optional(),
    nickname: z.string().optional(),
    parent_id: z.string().optional(),
    source: z.string().optional(),
    status: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    website: z.string().optional()
});

export const zRegisterDto = z.object({
    email: z.string().optional(),
    locale: z.string().optional(),
    password: z.string().optional(),
    username: z.string().optional()
});

export const zReindexStatusResponse = z.object({
    error: z.string().optional(),
    finished_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    payload: z.unknown().optional(),
    phase: z.string().optional(),
    started_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    status: z.string().optional()
});

export const zResultAgentSetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zAgentSetting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultCheckUpdateResponse = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zCheckUpdateResponse.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultComment = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zComment.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultConnect = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zConnect.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultCreateCommentResult = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zCreateCommentResult.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultEmbeddingSetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zEmbeddingSetting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultExportStateDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zExportStateDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultFileDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zFileDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultFileListResultDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zFileListResultDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultFileTreeResultDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zFileTreeResultDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultFormMeta = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zFormMeta.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultGlobalMigrationStateDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zGlobalMigrationStateDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultHelloResponse = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zHelloResponse.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultInterface = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.unknown().optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListAccessTokenSetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zAccessTokenSetting).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListConnect = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zConnect).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListConnected = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zConnected).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListConnectedHealth = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zConnectedHealth).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListDayStat = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zDayStat).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListHeatmap = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zHeatmap).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListLogEntry = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zLogEntry).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListPasskeyDeviceDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zPasskeyDeviceDto).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListPublicComment = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zPublicComment).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultManifest = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zManifest.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultOAuth2Setting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zOAuth2Setting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultOAuth2Status = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zOAuth2Status.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultOAuthInfoDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zOAuthInfoDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultPageResultComment = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zPageResultComment.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultPasskeySetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zPasskeySetting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultPasskeyStatus = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zPasskeyStatus.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultPresignDto = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zPresignDto.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultReindexStatusResponse = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zReindexStatusResponse.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultString = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.string().optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zS3Setting = z.object({
    access_key: z.string().optional(),
    bucket_name: z.string().optional(),
    cdn_url: z.string().optional(),
    enable: z.boolean().optional(),
    endpoint: z.string().optional(),
    path_prefix: z.string().optional(),
    provider: z.string().optional(),
    public_read: z.boolean().optional(),
    region: z.string().optional(),
    secret_key: z.string().optional(),
    use_path_style: z.boolean().optional(),
    use_ssl: z.boolean().optional()
});

export const zResultS3Setting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zS3Setting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zS3SettingDto = z.object({
    access_key: z.string().optional(),
    bucket_name: z.string().optional(),
    cdn_url: z.string().optional(),
    enable: z.boolean().optional(),
    endpoint: z.string().optional(),
    path_prefix: z.string().optional(),
    provider: z.string().optional(),
    public_read: z.boolean().optional(),
    region: z.string().optional(),
    secret_key: z.string().optional(),
    use_path_style: z.boolean().optional(),
    use_ssl: z.boolean().optional()
});

export const zSearchResult = z.object({
    content: z.string().optional(),
    distance: z.number().optional(),
    echo_created: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_id: z.string().optional(),
    extension: zEchoExtension.optional(),
    files: z.array(zFile).nullish(),
    username: z.string().optional()
});

export const zChatMessage = z.object({
    asks: z.array(zAskExchange).nullish(),
    content: z.string().optional(),
    reasoning: z.string().optional(),
    reasoning_ms: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    role: z.string().optional(),
    sources: z.array(zSearchResult).nullish()
});

export const zResultListChatMessage = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zChatMessage).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zSnapshotSchedule = z.object({
    cron_expression: z.string().optional(),
    enable: z.boolean().optional()
});

export const zResultSnapshotSchedule = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zSnapshotSchedule.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zSnapshotScheduleDto = z.object({
    cron_expression: z.string().optional(),
    enable: z.boolean().optional()
});

export const zStartExportRequest = z.object({
    format: z.string().optional(),
    include_private: z.boolean().optional()
});

export const zStartGlobalMigrationRequest = z.object({
    source_payload: z.record(z.string(), z.unknown()).optional(),
    source_type: z.string().optional()
});

export const zStatus = z.object({
    initialized: z.boolean().optional(),
    owner_exists: z.boolean().optional()
});

export const zResultStatus = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zStatus.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zSystemSetting = z.object({
    ICP_number: z.string().optional(),
    allow_register: z.boolean().optional(),
    custom_css: z.string().optional(),
    custom_js: z.string().optional(),
    default_locale: z.string().optional(),
    footer_content: z.string().optional(),
    footer_link: z.string().optional(),
    meting_api: z.string().optional(),
    server_logo: z.string().optional(),
    server_name: z.string().optional(),
    server_url: z.string().optional(),
    site_title: z.string().optional()
});

export const zResultSystemSetting = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zSystemSetting.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zSystemSettingDto = z.object({
    ICP_number: z.string().optional(),
    allow_register: z.boolean().optional(),
    custom_css: z.string().optional(),
    custom_js: z.string().optional(),
    default_locale: z.string().optional(),
    footer_content: z.string().optional(),
    footer_link: z.string().optional(),
    meting_api: z.string().optional(),
    server_logo: z.string().optional(),
    server_logo_file_id: z.string().optional(),
    server_name: z.string().optional(),
    server_url: z.string().optional(),
    site_title: z.string().optional()
});

export const zTag = z.object({
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    name: z.string().optional(),
    usage_count: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zEcho = z.object({
    content: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_files: z.array(zEchoFile).nullish(),
    extension: zEchoExtension.optional(),
    fav_count: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    layout: z.string().optional(),
    private: z.boolean().optional(),
    tags: z.array(zTag).nullish(),
    user_id: z.string().optional(),
    username: z.string().optional()
});

export const zEchoUpsertDto = z.object({
    content: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_files: z.array(zEchoFile).nullish(),
    extension: zEchoExtensionDto.optional(),
    id: z.string().optional(),
    layout: z.string().optional(),
    private: z.boolean().optional(),
    tags: z.array(zTag).nullish()
});

export const zPageQueryResultListEcho = z.object({
    items: z.array(zEcho).nullish(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zResultEcho = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zEcho.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListEcho = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zEcho).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultListTag = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zTag).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultPageQueryResultListEcho = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zPageQueryResultListEcho.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultTag = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zTag.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zTestEmailRequest = z.object({
    setting: zModelSystemSetting.optional()
});

export const zUpdateCommentHotDto = z.object({
    hot: z.boolean().optional()
});

export const zUpdateCommentStatusDto = z.object({
    status: z.string().optional()
});

export const zUpdateFileMetaDto = z.object({
    content_type: z.string().optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zUser = z.object({
    avatar: z.string().optional(),
    email: z.string().optional(),
    id: z.string().optional(),
    is_admin: z.boolean().optional(),
    is_owner: z.boolean().optional(),
    locale: z.string().optional(),
    username: z.string().optional()
});

export const zResultListUser = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zUser).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zResultUser = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: zUser.optional(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zUserInfoDto = z.object({
    avatar: z.string().optional(),
    avatar_file_id: z.string().optional(),
    email: z.string().optional(),
    is_admin: z.boolean().optional(),
    is_owner: z.boolean().optional(),
    locale: z.string().optional(),
    password: z.string().optional(),
    username: z.string().optional()
});

export const zWebhook = z.object({
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    is_active: z.boolean().optional(),
    last_status: z.string().optional(),
    last_trigger: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    name: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    url: z.string().optional()
});

export const zResultListWebhook = z.object({
    code: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    data: z.array(zWebhook).nullish(),
    error_code: z.string().optional(),
    message_key: z.string().optional(),
    message_params: z.record(z.string(), z.unknown()).optional(),
    msg: z.string().optional()
});

export const zWebhookDto = z.object({
    is_active: z.boolean().optional(),
    name: z.string().optional(),
    secret: z.string().optional(),
    url: z.string().optional()
});

/**
 * OK
 */
export const zAccessTokenListResponse = zResultListAccessTokenSetting;

export const zAccessTokenCreateBody = zAccessTokenSettingDto;

/**
 * OK
 */
export const zAccessTokenCreateResponse = zResultString;

export const zAccessTokenDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zAccessTokenDeleteResponse = zResultInterface;

/**
 * OK
 */
export const zAgentInfoResponse = zResultAgentSetting;

/**
 * OK
 */
export const zCopilotRecentResponse = zResultString;

/**
 * OK
 */
export const zAgentSettingsGetResponse = zResultAgentSetting;

export const zAgentSettingsUpdateBody = zAgentSettingDto;

/**
 * OK
 */
export const zAgentSettingsUpdateResponse = zResultInterface;

export const zAgentSettingsTestBody = zAgentSettingDto;

/**
 * OK
 */
export const zAgentSettingsTestResponse = zResultInterface;

export const zCopilotChatAnswerBody = zAnswerAskInputBody;

/**
 * OK
 */
export const zCopilotChatAnswerResponse = zResultInterface;

/**
 * OK
 */
export const zCopilotSessionClearResponse = zResultInterface;

/**
 * OK
 */
export const zCopilotSessionGetResponse = zResultListChatMessage;

export const zCommentListByEchoQuery = z.object({
    echo_id: z.string()
});

/**
 * OK
 */
export const zCommentListByEchoResponse = zResultListPublicComment;

export const zCommentCreateBody = zCreateCommentDto;

/**
 * OK
 */
export const zCommentCreateResponse = zResultCreateCommentResult;

/**
 * OK
 */
export const zCommentFormMetaResponse = zResultFormMeta;

export const zCommentCreateIntegrationBody = zCreateIntegrationCommentDto;

/**
 * OK
 */
export const zCommentCreateIntegrationResponse = zResultCreateCommentResult;

export const zCommentListPublicQuery = z.object({
    limit: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(30))
});

/**
 * OK
 */
export const zCommentListPublicResponse = zResultListPublicComment;

/**
 * OK
 */
export const zConnectSelfResponse = zResultConnect;

/**
 * OK
 */
export const zConnectListResponse = zResultListConnected;

export const zConnectAddBody = zConnected;

/**
 * OK
 */
export const zConnectAddResponse = zResultInterface;

/**
 * OK
 */
export const zConnectHealthResponse = zResultListConnectedHealth;

/**
 * OK
 */
export const zConnectInfoResponse = zResultListConnect;

export const zConnectDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zConnectDeleteResponse = zResultInterface;

export const zEchoCreateBody = zEchoUpsertDto;

/**
 * OK
 */
export const zEchoCreateResponse = zResultEcho;

export const zEchoUpdateBody = zEchoUpsertDto;

/**
 * OK
 */
export const zEchoUpdateResponse = zResultEcho;

export const zEchoHotQuery = z.object({
    limit: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(5))
});

/**
 * OK
 */
export const zEchoHotResponse = zResultListEcho;

export const zEchoLikePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zEchoLikeResponse = zResultInterface;

export const zEchoOnthisdayHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zEchoOnthisdayResponse = zResultListEcho;

export const zEchoPageGetQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional()
});

/**
 * OK
 */
export const zEchoPageGetResponse = zResultPageQueryResultListEcho;

export const zEchoPagePostBody = zPageQueryDto;

/**
 * OK
 */
export const zEchoPagePostResponse = zResultPageQueryResultListEcho;

export const zEchoListBody = zEchoQueryDto;

/**
 * OK
 */
export const zEchoListResponse = zResultPageQueryResultListEcho;

/**
 * OK
 */
export const zEchoRandomResponse = zResultEcho;

export const zEchoByTagPath = z.object({
    tagid: z.uuid()
});

export const zEchoByTagQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional()
});

/**
 * OK
 */
export const zEchoByTagResponse = zResultPageQueryResultListEcho;

export const zEchoTodayHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zEchoTodayResponse = zResultListEcho;

export const zEchoDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zEchoDeleteResponse = zResultInterface;

export const zEchoGetPath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zEchoGetResponse = zResultEcho;

/**
 * OK
 */
export const zEmbeddingReindexResponse = zResultReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingReindexCancelResponse = zResultReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingReindexStatusResponse = zResultReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingSettingsGetResponse = zResultEmbeddingSetting;

export const zEmbeddingSettingsUpdateBody = zEmbeddingSettingDto;

/**
 * OK
 */
export const zEmbeddingSettingsUpdateResponse = zResultInterface;

export const zFileTreeQuery = z.object({
    storage_type: z.string(),
    prefix: z.string().optional()
});

/**
 * OK
 */
export const zFileTreeResponse = zResultFileTreeResultDto;

export const zFileDeletePath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileDeleteResponse = zResultInterface;

export const zFileGetPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileGetResponse = zResultFileDto;

export const zFileUpdateMetaBody = zUpdateFileMetaDto;

export const zFileUpdateMetaPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileUpdateMetaResponse = zResultFileDto;

export const zFileListQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional(),
    storage_type: z.string().optional()
});

/**
 * OK
 */
export const zFileListResponse = zResultFileListResultDto;

export const zFileExternalBody = zCreateExternalFileDto;

/**
 * OK
 */
export const zFileExternalResponse = zResultFileDto;

export const zFilePresignBody = zGetPresignUrlDto;

/**
 * OK
 */
export const zFilePresignResponse = zResultPresignDto;

export const zCommonHeatmapHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zCommonHeatmapResponse = zResultListHeatmap;

/**
 * OK
 */
export const zCommonHelloResponse = zResultHelloResponse;

export const zInitOwnerBody = zRegisterDto;

/**
 * OK
 */
export const zInitOwnerResponse = zResultInterface;

/**
 * OK
 */
export const zInitStatusResponse = zResultStatus;

/**
 * OK
 */
export const zMcpManifestResponse = zResultManifest;

/**
 * OK
 */
export const zMigrationCancelResponse = zResultGlobalMigrationStateDto;

/**
 * OK
 */
export const zMigrationCleanupResponse = zResultInterface;

export const zMigrationExportBody = zStartExportRequest;

/**
 * OK
 */
export const zMigrationExportResponse = zResultExportStateDto;

/**
 * OK
 */
export const zMigrationExportCancelResponse = zResultExportStateDto;

/**
 * OK
 */
export const zMigrationExportStatusResponse = zResultExportStateDto;

export const zMigrationStartBody = zStartGlobalMigrationRequest;

/**
 * OK
 */
export const zMigrationStartResponse = zResultGlobalMigrationStateDto;

/**
 * OK
 */
export const zMigrationStatusResponse = zResultGlobalMigrationStateDto;

export const zOauthInfoQuery = z.object({
    provider: z.string().optional()
});

/**
 * OK
 */
export const zOauthInfoResponse = zResultOAuthInfoDto;

export const zOauthBindBody = zOAuthBindBody;

export const zOauthBindPath = z.object({
    provider: z.string()
});

/**
 * OK
 */
export const zOauthBindResponse = zResultString;

/**
 * OK
 */
export const zOauth2GetResponse = zResultOAuth2Setting;

export const zOauth2UpdateBody = zOAuth2SettingDto;

/**
 * OK
 */
export const zOauth2UpdateResponse = zResultInterface;

/**
 * OK
 */
export const zOauth2StatusResponse = zResultOAuth2Status;

export const zCommentPanelListQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(1)),
    page_size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(20)),
    keyword: z.string().optional(),
    status: z.string().optional(),
    echo_id: z.string().optional(),
    hot: z.string().optional()
});

/**
 * OK
 */
export const zCommentPanelListResponse = zResultPageResultComment;

export const zCommentPanelBatchBody = zBatchCommentActionDto;

/**
 * OK
 */
export const zCommentPanelBatchResponse = zResultInterface;

/**
 * OK
 */
export const zCommentPanelSettingsGetResponse = zModelResultCommentSystemSetting;

export const zCommentPanelSettingsUpdateBody = zModelSystemSetting;

/**
 * OK
 */
export const zCommentPanelSettingsUpdateResponse = zResultInterface;

export const zCommentPanelTestEmailBody = zTestEmailRequest;

/**
 * OK
 */
export const zCommentPanelTestEmailResponse = zResultInterface;

export const zCommentPanelDeletePath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zCommentPanelDeleteResponse = zResultInterface;

export const zCommentPanelGetPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zCommentPanelGetResponse = zResultComment;

export const zCommentPanelHotBody = zUpdateCommentHotDto;

export const zCommentPanelHotPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zCommentPanelHotResponse = zResultInterface;

export const zCommentPanelStatusBody = zUpdateCommentStatusDto;

export const zCommentPanelStatusPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zCommentPanelStatusResponse = zResultInterface;

/**
 * OK
 */
export const zPasskeyGetResponse = zResultPasskeySetting;

export const zPasskeyUpdateBody = zPasskeySettingDto;

/**
 * OK
 */
export const zPasskeyUpdateResponse = zResultInterface;

/**
 * OK
 */
export const zPasskeyStatusResponse = zResultPasskeyStatus;

/**
 * OK
 */
export const zPasskeyListResponse = zResultListPasskeyDeviceDto;

export const zPasskeyDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zPasskeyDeleteResponse = zResultInterface;

export const zPasskeyUpdateNameBody = zPasskeyUpdateDeviceNameReq;

export const zPasskeyUpdateNamePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zPasskeyUpdateNameResponse = zResultInterface;

export const zUserRegisterBody = zRegisterDto;

/**
 * OK
 */
export const zUserRegisterResponse = zResultInterface;

/**
 * OK
 */
export const zS3GetResponse = zResultS3Setting;

export const zS3UpdateBody = zS3SettingDto;

/**
 * OK
 */
export const zS3UpdateResponse = zResultInterface;

export const zS3TestBody = zS3SettingDto;

/**
 * OK
 */
export const zS3TestResponse = zResultInterface;

/**
 * OK
 */
export const zSettingsGetResponse = zResultSystemSetting;

export const zSettingsUpdateBody = zSystemSettingDto;

/**
 * OK
 */
export const zSettingsUpdateResponse = zResultInterface;

/**
 * OK
 */
export const zSnapshotScheduleGetResponse = zResultSnapshotSchedule;

export const zSnapshotScheduleUpdateBody = zSnapshotScheduleDto;

/**
 * OK
 */
export const zSnapshotScheduleUpdateResponse = zResultInterface;

/**
 * OK
 */
export const zDashboardCheckUpdateResponse = zResultCheckUpdateResponse;

export const zDashboardSystemLogsQuery = z.object({
    tail: z.string().optional(),
    level: z.string().optional(),
    keyword: z.string().optional()
});

/**
 * OK
 */
export const zDashboardSystemLogsResponse = zResultListLogEntry;

/**
 * OK
 */
export const zDashboardVisitorStatsResponse = zResultListDayStat;

export const zTagCreateBody = zCreateTagDto;

/**
 * OK
 */
export const zTagCreateResponse = zResultTag;

export const zTagDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zTagDeleteResponse = zResultInterface;

/**
 * OK
 */
export const zTagListResponse = zResultListTag;

/**
 * OK
 */
export const zUserInfoResponse = zResultUser;

export const zUserUpdateBody = zUserInfoDto;

/**
 * OK
 */
export const zUserUpdateResponse = zResultInterface;

export const zUserSetAdminPath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zUserSetAdminResponse = zResultInterface;

export const zUserDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zUserDeleteResponse = zResultInterface;

/**
 * OK
 */
export const zUserListResponse = zResultListUser;

/**
 * OK
 */
export const zWebhookListResponse = zResultListWebhook;

export const zWebhookCreateBody = zWebhookDto;

/**
 * OK
 */
export const zWebhookCreateResponse = zResultInterface;

export const zWebhookDeletePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zWebhookDeleteResponse = zResultInterface;

export const zWebhookUpdateBody = zWebhookDto;

export const zWebhookUpdatePath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zWebhookUpdateResponse = zResultInterface;

export const zWebhookTestPath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zWebhookTestResponse = zResultInterface;

export const zCommonWebsiteTitleQuery = z.object({
    website_url: z.string()
});

/**
 * OK
 */
export const zCommonWebsiteTitleResponse = zResultString;
