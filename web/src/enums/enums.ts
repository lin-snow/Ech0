// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

export enum Mode {
  ECH0 = 0,
  Panel = 1,
  EXTEN = 3,
  Media = 5,
  TagManage = 6,
}

export enum ExtensionType {
  MUSIC = 'MUSIC',
  VIDEO = 'VIDEO',
  GITHUBPROJ = 'GITHUBPROJ',
  WEBSITE = 'WEBSITE',
  LOCATION = 'LOCATION',
  TWEET = 'TWEET',
}

export enum ImageLayout {
  WATERFALL = 'waterfall',
  GRID = 'grid',
  HORIZONTAL = 'horizontal',
  CAROUSEL = 'carousel',
  STACK = 'stack',
}

export enum VideoLayout {
  DEFAULT = 'none',
}

export enum AudioLayout {
  DEFAULT = 'none',
}

export enum S3Provider {
  AWS = 'aws',
  ALIYUN = 'aliyun',
  TENCENT = 'tencent',
  MINIO = 'minio',
  R2 = 'r2',
  OTHER = 'other',
}

export enum OAuth2Provider {
  GITHUB = 'github',
  GOOGLE = 'google',
  QQ = 'qq',
  CUSTOM = 'custom',
}

export enum FollowStatus {
  NONE = 'none',
  PENDING = 'pending',
  ACCEPTED = 'accepted',
  REJECTED = 'rejected',
}

export enum MusicProvider {
  NETEASE = 'netease',
  QQ = 'tencent',
  APPLE = 'apple',
}

export enum AccessTokenExpiration {
  EIGHT_HOUR_EXPIRY = '8_hours',
  ONE_MONTH_EXPIRY = '1_month',
  NEVER_EXPIRY = 'never',
}

export enum AgentProtocol {
  OPENAI = 'openai',
  OPENAI_RESPONSES = 'openai-responses',
  ANTHROPIC = 'anthropic',
}
