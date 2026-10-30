// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

/** Scope catalogue shared by the token editor and the device-approval page. */
export const ACCESS_SCOPE_GROUPS = [
  {
    labelKey: 'accessTokenSetting.scopeGroupContent',
    items: [
      { value: 'echo:read', labelKey: 'accessTokenSetting.scopeEchoRead' },
      { value: 'echo:write', labelKey: 'accessTokenSetting.scopeEchoWrite' },
      { value: 'comment:read', labelKey: 'accessTokenSetting.scopeCommentRead' },
      { value: 'comment:write', labelKey: 'accessTokenSetting.scopeCommentWrite' },
      { value: 'comment:moderate', labelKey: 'accessTokenSetting.scopeCommentModerate' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupFile',
    items: [
      { value: 'file:read', labelKey: 'accessTokenSetting.scopeFileRead' },
      { value: 'file:write', labelKey: 'accessTokenSetting.scopeFileWrite' },
    ],
  },
  {
    labelKey: 'accessTokenSetting.scopeGroupConnect',
    items: [
      { value: 'connect:read', labelKey: 'accessTokenSetting.scopeConnectRead' },
      { value: 'connect:write', labelKey: 'accessTokenSetting.scopeConnectWrite' },
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
      { value: 'admin:settings', labelKey: 'accessTokenSetting.scopeAdminSettings' },
      { value: 'admin:user', labelKey: 'accessTokenSetting.scopeAdminUser' },
      { value: 'admin:token', labelKey: 'accessTokenSetting.scopeAdminToken' },
    ],
  },
] as const

/** Scope value → i18n label key, for callers that only need a one-off label. */
export const ACCESS_SCOPE_LABEL_KEYS: Record<string, string> = Object.fromEntries(
  ACCESS_SCOPE_GROUPS.flatMap((group) =>
    group.items.map((item) => [item.value, item.labelKey] as const),
  ),
)
