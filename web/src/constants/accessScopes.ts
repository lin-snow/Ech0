// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** Scope catalogue shared by the token editor and the device-approval page. */
export const ACCESS_SCOPE_GROUPS = [
  {
    labelKey: 'accessTokenSetting.scopeGroupContent',
    items: [
      { value: 'echo:read', labelKey: 'accessTokenSetting.scopeEchoRead' },
      { value: 'echo:write', labelKey: 'accessTokenSetting.scopeEchoWrite' },
      { value: 'echo:delete', labelKey: 'accessTokenSetting.scopeEchoDelete' },
      { value: 'echo:private', labelKey: 'accessTokenSetting.scopeEchoPrivate' },
      { value: 'comment:read', labelKey: 'accessTokenSetting.scopeCommentRead' },
      { value: 'comment:write', labelKey: 'accessTokenSetting.scopeCommentWrite' },
      { value: 'comment:review', labelKey: 'accessTokenSetting.scopeCommentReview' },
      { value: 'comment:delete', labelKey: 'accessTokenSetting.scopeCommentDelete' },
      { value: 'comment:settings', labelKey: 'accessTokenSetting.scopeCommentSettings' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupFile',
    items: [
      { value: 'file:read', labelKey: 'accessTokenSetting.scopeFileRead' },
      { value: 'file:write', labelKey: 'accessTokenSetting.scopeFileWrite' },
      { value: 'file:delete', labelKey: 'accessTokenSetting.scopeFileDelete' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupConnect',
    items: [
      { value: 'connect:read', labelKey: 'accessTokenSetting.scopeConnectRead' },
      { value: 'connect:write', labelKey: 'accessTokenSetting.scopeConnectWrite' },
      { value: 'connect:delete', labelKey: 'accessTokenSetting.scopeConnectDelete' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupProfile',
    items: [
      { value: 'profile:read', labelKey: 'accessTokenSetting.scopeProfileRead' },
      { value: 'profile:write', labelKey: 'accessTokenSetting.scopeProfileWrite' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupAdmin',
    items: [
      { value: 'admin:site', labelKey: 'accessTokenSetting.scopeAdminSite' },
      { value: 'admin:storage', labelKey: 'accessTokenSetting.scopeAdminStorage' },
      { value: 'admin:auth', labelKey: 'accessTokenSetting.scopeAdminAuth' },
      { value: 'admin:webhook', labelKey: 'accessTokenSetting.scopeAdminWebhook' },
      { value: 'admin:agent', labelKey: 'accessTokenSetting.scopeAdminAgent' },
      { value: 'admin:data', labelKey: 'accessTokenSetting.scopeAdminData' },
      { value: 'admin:ops', labelKey: 'accessTokenSetting.scopeAdminOps' },
      { value: 'admin:user', labelKey: 'accessTokenSetting.scopeAdminUser' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupMcp',
    items: [{ value: 'mcp:read', labelKey: 'accessTokenSetting.scopeMcpRead' }],
  },
] as const

/** Scope value → i18n label key, for callers that only need a one-off label. */
export const ACCESS_SCOPE_LABEL_KEYS: Record<string, string> = Object.fromEntries(
  ACCESS_SCOPE_GROUPS.flatMap((group) =>
    group.items.map((item) => [item.value, item.labelKey] as const),
  ),
)
