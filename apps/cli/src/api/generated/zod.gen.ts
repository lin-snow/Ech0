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

export const zAddConnectDto = z.object({
    connect_url: z.string()
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

export const zCreateIntegrationCommentDto = z.object({
    content: z.string(),
    echo_id: z.string(),
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
    cursor: z.string().optional(),
    dateFrom: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    dateTo: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    mine: z.boolean().optional(),
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    private: z.boolean().optional(),
    search: z.string().optional(),
    sortBy: z.enum(['created_at', 'reactions']).optional(),
    sortOrder: z.enum(['asc', 'desc']).optional(),
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

export const zErrorDetail = z.object({
    location: z.string().optional(),
    message: z.string().optional(),
    value: z.unknown().optional()
});

export const zErrorModel = z.object({
    detail: z.string().optional(),
    errors: z.array(zErrorDetail).nullish(),
    instance: z.url().optional(),
    status: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    title: z.string().optional(),
    type: z.url().optional().default('about:blank')
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
    access: z.string().optional(),
    blob_bytes: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    content_type: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    kind: z.string().optional(),
    name: z.string().optional(),
    object_key: z.string().optional(),
    ref_url: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    status: z.string().optional(),
    storage_id: z.string().optional(),
    updated_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    uploader_id: z.string().optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zEchoFile = z.object({
    file: zFile.optional(),
    file_id: z.string().optional(),
    sort_order: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFormMeta = z.object({
    captcha_api_endpoint: z.string().optional(),
    captcha_enabled: z.boolean().optional(),
    enable_comment: z.boolean().optional(),
    form_token: z.string().optional(),
    min_submit_ms: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
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

export const zInitOwnerDto = z.object({
    email: z.string().optional(),
    locale: z.string().optional(),
    password: z.string().optional(),
    site_title: z.string().optional(),
    username: z.string().optional()
});

export const zLinkDto = z.object({
    owner_id: z.string().optional(),
    owner_kind: z.string().optional(),
    role: z.string().optional(),
    sort_order: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zFileDto = z.object({
    access: z.string().optional(),
    category: z.string().optional(),
    content_type: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    height: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    id: z.string().optional(),
    kind: z.string().optional(),
    links: z.array(zLinkDto).nullish(),
    name: z.string().optional(),
    size: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    status: z.string().optional(),
    uploader_id: z.string().optional(),
    url: z.string().optional(),
    width: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListAccessTokenSetting = z.object({
    items: z.array(zAccessTokenSetting).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListComment = z.object({
    items: z.array(zComment).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListConnect = z.object({
    items: z.array(zConnect).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListConnected = z.object({
    items: z.array(zConnected).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListConnectedHealth = z.object({
    items: z.array(zConnectedHealth).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListDayStat = z.object({
    items: z.array(zDayStat).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListFileDto = z.object({
    items: z.array(zFileDto).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListHeatmap = z.object({
    items: z.array(zHeatmap).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
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

export const zListLogEntry = z.object({
    items: z.array(zLogEntry).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
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

export const zListPublicComment = z.object({
    items: z.array(zPublicComment).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zReactionCount = z.object({
    count: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    emoji: z.string().optional(),
    key: z.string().optional(),
    reacted: z.boolean().optional()
});

export const zListReactionCount = z.object({
    items: z.array(zReactionCount).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zReactionEmoji = z.object({
    emoji: z.string().optional(),
    key: z.string().optional()
});

export const zListReactionEmoji = z.object({
    items: z.array(zReactionEmoji).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
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

export const zRenameFileInputBody = z.object({
    name: z.string().optional()
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

export const zSnapshotSchedule = z.object({
    cron_expression: z.string().optional(),
    enable: z.boolean().optional()
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

export const zSystemSetting = z.object({
    ICP_number: z.string().optional(),
    allow_register: z.boolean().optional(),
    custom_css: z.string().optional(),
    custom_js: z.string().optional(),
    default_locale: z.string().optional(),
    enable_agent_discovery: z.boolean().optional(),
    footer_content: z.string().optional(),
    footer_link: z.string().optional(),
    meting_api: z.string().optional(),
    privacy_policy: z.string().optional(),
    server_logo: z.string().optional(),
    server_logo_file_id: z.string().optional(),
    server_name: z.string().optional(),
    server_url: z.string().optional(),
    show_privacy_footer: z.boolean().optional(),
    site_title: z.string().optional()
});

export const zSystemSettingDto = z.object({
    ICP_number: z.string().optional(),
    allow_register: z.boolean().optional(),
    custom_css: z.string().optional(),
    custom_js: z.string().optional(),
    default_locale: z.string().optional(),
    enable_agent_discovery: z.boolean().optional(),
    footer_content: z.string().optional(),
    footer_link: z.string().optional(),
    meting_api: z.string().optional(),
    privacy_policy: z.string().optional(),
    server_logo: z.string().optional(),
    server_logo_file_id: z.string().optional(),
    server_name: z.string().optional(),
    server_url: z.string().optional(),
    show_privacy_footer: z.boolean().optional(),
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
    deleted_at: z.iso.datetime().optional(),
    echo_files: z.array(zEchoFile).nullish(),
    extension: zEchoExtension.optional(),
    id: z.string().optional(),
    layout: z.string().optional(),
    private: z.boolean().optional(),
    reactions: z.array(zReactionCount).nullish(),
    tags: z.array(zTag).nullish(),
    user_id: z.string().optional(),
    username: z.string().optional()
});

export const zEchoUpsertDto = z.object({
    content: z.string().optional(),
    created_at: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    echo_files: z.array(zEchoFile).optional(),
    extension: zEchoExtensionDto.optional(),
    id: z.string().optional(),
    layout: z.string().optional(),
    private: z.boolean().optional(),
    tags: z.array(zTag).nullish()
});

export const zListEcho = z.object({
    items: z.array(zEcho).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zListTag = z.object({
    items: z.array(zTag).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
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

export const zUpdateUserRoleInputBody = z.object({
    role: z.string().optional()
});

export const zUpdateUserStatusInputBody = z.object({
    status: z.string().optional()
});

export const zUser = z.object({
    avatar: z.string().optional(),
    avatar_file_id: z.string().optional(),
    email: z.string().optional(),
    id: z.string().optional(),
    locale: z.string().optional(),
    role: z.string().optional(),
    status: z.string().optional(),
    username: z.string().optional()
});

export const zListUser = z.object({
    items: z.array(zUser).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zUserInfoDto = z.object({
    avatar: z.string().optional(),
    avatar_file_id: z.string().optional(),
    email: z.string().optional(),
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

export const zListWebhook = z.object({
    items: z.array(zWebhook).nullish(),
    nextCursor: z.string().optional(),
    total: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional()
});

export const zWebhookDto = z.object({
    is_active: z.boolean().optional(),
    name: z.string(),
    secret: z.string().optional(),
    url: z.string()
});

/**
 * OK
 */
export const zAccessTokenListResponse = zListAccessTokenSetting;

export const zAccessTokenCreateBody = zAccessTokenSettingDto;

/**
 * Created
 */
export const zAccessTokenCreateResponse = z.string();

export const zAccessTokenDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zAccessTokenDeleteResponse = z.void();

/**
 * OK
 */
export const zAgentInfoResponse = zAgentSetting;

/**
 * OK
 */
export const zCopilotRecentResponse = z.string();

/**
 * OK
 */
export const zAgentSettingsGetResponse = zAgentSetting;

export const zAgentSettingsUpdateBody = zAgentSettingDto;

/**
 * No Content
 */
export const zAgentSettingsUpdateResponse = z.void();

export const zAgentSettingsTestBody = zAgentSettingDto;

/**
 * No Content
 */
export const zAgentSettingsTestResponse = z.void();

export const zCopilotChatAnswerBody = zAnswerAskInputBody;

/**
 * No Content
 */
export const zCopilotChatAnswerResponse = z.void();

/**
 * No Content
 */
export const zCopilotSessionClearResponse = z.void();

/**
 * OK
 */
export const zCopilotSessionGetResponse = z.array(zChatMessage).nullable();

export const zCommentListByEchoQuery = z.object({
    echo_id: z.string()
});

/**
 * OK
 */
export const zCommentListByEchoResponse = zListPublicComment;

export const zCommentCreateBody = zCreateCommentDto;

/**
 * Created
 */
export const zCommentCreateResponse = zComment;

/**
 * OK
 */
export const zCommentFormMetaResponse = zFormMeta;

export const zCommentCreateIntegrationBody = zCreateIntegrationCommentDto;

/**
 * Created
 */
export const zCommentCreateIntegrationResponse = zComment;

export const zCommentListPublicQuery = z.object({
    limit: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(30))
});

/**
 * OK
 */
export const zCommentListPublicResponse = zListPublicComment;

/**
 * OK
 */
export const zConnectSelfResponse = zConnect;

/**
 * OK
 */
export const zConnectListResponse = zListConnected;

export const zConnectAddBody = zAddConnectDto;

/**
 * No Content
 */
export const zConnectAddResponse = z.void();

/**
 * OK
 */
export const zConnectHealthResponse = zListConnectedHealth;

/**
 * OK
 */
export const zConnectInfoResponse = zListConnect;

export const zConnectDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zConnectDeleteResponse = z.void();

export const zEchoCreateBody = zEchoUpsertDto;

/**
 * Created
 */
export const zEchoCreateResponse = zEcho;

export const zEchoUpdateBody = zEchoUpsertDto;

/**
 * OK
 */
export const zEchoUpdateResponse = zEcho;

export const zEchoHotQuery = z.object({
    limit: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional().default(BigInt(5))
});

/**
 * OK
 */
export const zEchoHotResponse = zListEcho;

export const zEchoOnthisdayHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zEchoOnthisdayResponse = zListEcho;

export const zEchoPageGetQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional()
});

/**
 * OK
 */
export const zEchoPageGetResponse = zListEcho;

export const zEchoPagePostBody = zPageQueryDto;

/**
 * OK
 */
export const zEchoPagePostResponse = zListEcho;

export const zEchoListBody = zEchoQueryDto;

/**
 * OK
 */
export const zEchoListResponse = zListEcho;

/**
 * OK
 */
export const zEchoRandomResponse = zEcho;

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
export const zEchoByTagResponse = zListEcho;

export const zEchoTodayHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zEchoTodayResponse = zListEcho;

export const zEchoDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zEchoDeleteResponse = z.void();

export const zEchoGetPath = z.object({
    id: z.uuid()
});

/**
 * OK
 */
export const zEchoGetResponse = zEcho;

export const zReactionRemovePath = z.object({
    id: z.uuid(),
    emoji: z.enum([
        'like',
        'heart',
        'haha',
        'wow',
        'sad',
        'cry',
        'party'
    ])
});

/**
 * OK
 */
export const zReactionRemoveResponse = zListReactionCount;

export const zReactionAddPath = z.object({
    id: z.uuid(),
    emoji: z.enum([
        'like',
        'heart',
        'haha',
        'wow',
        'sad',
        'cry',
        'party'
    ])
});

/**
 * OK
 */
export const zReactionAddResponse = zListReactionCount;

/**
 * OK
 */
export const zEmbeddingReindexResponse = zReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingReindexCancelResponse = zReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingReindexStatusResponse = zReindexStatusResponse;

/**
 * OK
 */
export const zEmbeddingSettingsGetResponse = zEmbeddingSetting;

export const zEmbeddingSettingsUpdateBody = zEmbeddingSettingDto;

/**
 * No Content
 */
export const zEmbeddingSettingsUpdateResponse = z.void();

export const zFileListQuery = z.object({
    page: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    pageSize: z.coerce.bigint().min(BigInt('-9223372036854775808'), { error: 'Invalid value: Expected int64 to be >= -9223372036854775808' }).max(BigInt('9223372036854775807'), { error: 'Invalid value: Expected int64 to be <= 9223372036854775807' }).optional(),
    search: z.string().optional(),
    kind: z.string().optional(),
    status: z.string().optional(),
    access: z.string().optional(),
    ownerKind: z.string().optional(),
    ownerId: z.string().optional(),
    role: z.string().optional()
});

/**
 * OK
 */
export const zFileListResponse = zListFileDto;

export const zFileDeletePath = z.object({
    id: z.string()
});

export const zFileDeleteQuery = z.object({
    force: z.boolean().optional()
});

/**
 * No Content
 */
export const zFileDeleteResponse = z.void();

export const zFileGetPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileGetResponse = zFileDto;

export const zFileRenameBody = zRenameFileInputBody;

export const zFileRenamePath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileRenameResponse = zFileDto;

export const zFileCommitPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zFileCommitResponse = zFileDto;

export const zCommonHeatmapHeaders = z.object({
    'X-Timezone': z.string().optional()
});

/**
 * OK
 */
export const zCommonHeatmapResponse = zListHeatmap;

/**
 * OK
 */
export const zCommonHelloResponse = zHelloResponse;

export const zInitOwnerBody = zInitOwnerDto;

/**
 * No Content
 */
export const zInitOwnerResponse = z.void();

/**
 * OK
 */
export const zInitStatusResponse = zStatus;

/**
 * OK
 */
export const zMcpManifestResponse = zManifest;

/**
 * OK
 */
export const zMigrationCancelResponse = zGlobalMigrationStateDto;

/**
 * No Content
 */
export const zMigrationCleanupResponse = z.void();

export const zMigrationExportBody = zStartExportRequest;

/**
 * OK
 */
export const zMigrationExportResponse = zExportStateDto;

/**
 * OK
 */
export const zMigrationExportCancelResponse = zExportStateDto;

/**
 * OK
 */
export const zMigrationExportStatusResponse = zExportStateDto;

export const zMigrationStartBody = zStartGlobalMigrationRequest;

/**
 * OK
 */
export const zMigrationStartResponse = zGlobalMigrationStateDto;

/**
 * OK
 */
export const zMigrationStatusResponse = zGlobalMigrationStateDto;

export const zOauthInfoQuery = z.object({
    provider: z.string().optional()
});

/**
 * OK
 */
export const zOauthInfoResponse = zOAuthInfoDto;

export const zOauthBindBody = zOAuthBindBody;

export const zOauthBindPath = z.object({
    provider: z.string()
});

/**
 * OK
 */
export const zOauthBindResponse = z.string();

/**
 * OK
 */
export const zOauth2GetResponse = zOAuth2Setting;

export const zOauth2UpdateBody = zOAuth2SettingDto;

/**
 * No Content
 */
export const zOauth2UpdateResponse = z.void();

/**
 * OK
 */
export const zOauth2StatusResponse = zOAuth2Status;

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
export const zCommentPanelListResponse = zListComment;

export const zCommentPanelBatchBody = zBatchCommentActionDto;

/**
 * No Content
 */
export const zCommentPanelBatchResponse = z.void();

/**
 * OK
 */
export const zCommentPanelSettingsGetResponse = zModelSystemSetting;

export const zCommentPanelSettingsUpdateBody = zModelSystemSetting;

/**
 * No Content
 */
export const zCommentPanelSettingsUpdateResponse = z.void();

export const zCommentPanelTestEmailBody = zTestEmailRequest;

/**
 * No Content
 */
export const zCommentPanelTestEmailResponse = z.void();

export const zCommentPanelDeletePath = z.object({
    id: z.string()
});

/**
 * No Content
 */
export const zCommentPanelDeleteResponse = z.void();

export const zCommentPanelGetPath = z.object({
    id: z.string()
});

/**
 * OK
 */
export const zCommentPanelGetResponse = zComment;

export const zCommentPanelHotBody = zUpdateCommentHotDto;

export const zCommentPanelHotPath = z.object({
    id: z.string()
});

/**
 * No Content
 */
export const zCommentPanelHotResponse = z.void();

export const zCommentPanelStatusBody = zUpdateCommentStatusDto;

export const zCommentPanelStatusPath = z.object({
    id: z.string()
});

/**
 * No Content
 */
export const zCommentPanelStatusResponse = z.void();

/**
 * OK
 */
export const zPasskeyGetResponse = zPasskeySetting;

export const zPasskeyUpdateBody = zPasskeySettingDto;

/**
 * No Content
 */
export const zPasskeyUpdateResponse = z.void();

/**
 * OK
 */
export const zPasskeyStatusResponse = zPasskeyStatus;

/**
 * OK
 */
export const zPasskeyListResponse = z.array(zPasskeyDeviceDto).nullable();

export const zPasskeyDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zPasskeyDeleteResponse = z.void();

export const zPasskeyUpdateNameBody = zPasskeyUpdateDeviceNameReq;

export const zPasskeyUpdateNamePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zPasskeyUpdateNameResponse = z.void();

/**
 * OK
 */
export const zReactionCatalogResponse = zListReactionEmoji;

export const zUserRegisterBody = zRegisterDto;

/**
 * No Content
 */
export const zUserRegisterResponse = z.void();

/**
 * OK
 */
export const zS3GetResponse = zS3Setting;

export const zS3UpdateBody = zS3SettingDto;

/**
 * No Content
 */
export const zS3UpdateResponse = z.void();

export const zS3TestBody = zS3SettingDto;

/**
 * No Content
 */
export const zS3TestResponse = z.void();

/**
 * OK
 */
export const zSettingsGetResponse = zSystemSetting;

export const zSettingsUpdateBody = zSystemSettingDto;

/**
 * No Content
 */
export const zSettingsUpdateResponse = z.void();

/**
 * OK
 */
export const zSnapshotScheduleGetResponse = zSnapshotSchedule;

export const zSnapshotScheduleUpdateBody = zSnapshotScheduleDto;

/**
 * No Content
 */
export const zSnapshotScheduleUpdateResponse = z.void();

/**
 * OK
 */
export const zDashboardCheckUpdateResponse = zCheckUpdateResponse;

export const zDashboardSystemLogsQuery = z.object({
    tail: z.string().optional(),
    level: z.string().optional(),
    keyword: z.string().optional()
});

/**
 * OK
 */
export const zDashboardSystemLogsResponse = zListLogEntry;

/**
 * OK
 */
export const zDashboardVisitorStatsResponse = zListDayStat;

export const zTagCreateBody = zCreateTagDto;

/**
 * OK
 */
export const zTagCreateResponse = zTag;

export const zTagDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zTagDeleteResponse = z.void();

/**
 * OK
 */
export const zTagListResponse = zListTag;

/**
 * OK
 */
export const zUserInfoResponse = zUser;

export const zUserUpdateBody = zUserInfoDto;

/**
 * No Content
 */
export const zUserUpdateResponse = z.void();

export const zUserTransferOwnerPath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zUserTransferOwnerResponse = z.void();

export const zUserSetRoleBody = zUpdateUserRoleInputBody;

export const zUserSetRolePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zUserSetRoleResponse = z.void();

export const zUserSetStatusBody = zUpdateUserStatusInputBody;

export const zUserSetStatusPath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zUserSetStatusResponse = z.void();

export const zUserDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zUserDeleteResponse = z.void();

/**
 * OK
 */
export const zUserListResponse = zListUser;

/**
 * OK
 */
export const zWebhookListResponse = zListWebhook;

export const zWebhookCreateBody = zWebhookDto;

/**
 * No Content
 */
export const zWebhookCreateResponse = z.void();

export const zWebhookDeletePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zWebhookDeleteResponse = z.void();

export const zWebhookUpdateBody = zWebhookDto;

export const zWebhookUpdatePath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zWebhookUpdateResponse = z.void();

export const zWebhookTestPath = z.object({
    id: z.uuid()
});

/**
 * No Content
 */
export const zWebhookTestResponse = z.void();

export const zCommonWebsiteTitleQuery = z.object({
    website_url: z.string()
});

/**
 * OK
 */
export const zCommonWebsiteTitleResponse = z.string();
