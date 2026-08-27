// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type Heatmap struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type (
	UploadFileType string
	S3Provider     string
	OAuth2Provider string
	AgentProtocol  string
	Locale         string
)

const (
	ImageType UploadFileType = "image"
	AudioType UploadFileType = "audio"
	VideoType UploadFileType = "video"
)

const (
	AWS     S3Provider = "aws"
	ALIYUN  S3Provider = "aliyun"
	TENCENT S3Provider = "tencent"
	R2      S3Provider = "r2"
	MINIO   S3Provider = "minio"
	OTHER   S3Provider = "other"
)

const (
	OAuth2GITHUB OAuth2Provider = "github"
	OAuth2GOOGLE OAuth2Provider = "google"
	OAuth2QQ     OAuth2Provider = "qq"
	OAuth2CUSTOM OAuth2Provider = "custom"
)

const (
	OpenAI          AgentProtocol = "openai"
	OpenAIResponses AgentProtocol = "openai-responses"
	Anthropic       AgentProtocol = "anthropic"
)

const (
	LocaleZhCN     Locale = "zh-CN"
	LocaleEnUS     Locale = "en-US"
	LocaleDeDE     Locale = "de-DE"
	DefaultLocale         = LocaleZhCN
	FallbackLocale        = LocaleEnUS
)

type KeyValue struct {
	Key   string `json:"key"   gorm:"primaryKey"`
	Value string `json:"value"`
}

const (
	SystemSettingsKey              = "system_settings"
	CommentSettingKey              = "comment_setting"
	S3SettingKey                   = "s3_setting"
	OAuth2SettingKey               = "oauth2_setting"
	PasskeySettingKey              = "passkey_setting"
	ServerURLKey                   = "server_url"
	SnapshotScheduleKey            = "snapshot_schedule"
	AgentSettingKey                = "agent_setting"
	EmbeddingSettingKey            = "embedding_setting"
	EmbeddingIndexStateKey         = "embedding_index_state"
	ReleaseVersionKey              = "release_version"
	InstallInitializedKey          = "install_initialized"
	MigrationGlobalJobStateKey     = "migration_global_job_state"
	StorageTimeUTCNormalizedKey    = "storage_time_utc_normalized_v1"
	StorageTimeSanitizedKey        = "storage_time_sanitized_v1"
	StorageTimeValidatedKey        = "storage_time_validated_v1"
	StorageTimeUnixMigratedKey     = "storage_time_unix_migrated_v1"
	StorageTimeSchemaRebuiltKey    = "storage_time_schema_rebuilt_v1"
	OAuthBindingsDroppedKey        = "oauth_bindings_dropped_v1"
	LegacyInboxesDroppedKey        = "legacy_inboxes_dropped_v1"
	EchoExtensionOrphansCleanedKey = "echo_extension_orphans_cleaned_v1"
	AgentProtocolCollapsedKey      = "agent_provider_collapsed_v1"
	AgentSettingProtocolRenamedKey = "agent_setting_protocol_renamed_v1"
	UserLocalAuthBackfilledKey     = "user_local_auth_backfilled_v1"
	UsersPasswordColumnDroppedKey  = "users_password_column_dropped_v1"
	ChatSessionKeyPrefix           = "chat_session:"
)

type PageQueryResult[T any] struct {
	Total int64 `json:"total"`
	Items T     `json:"items"`
}
