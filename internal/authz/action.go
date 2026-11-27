// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

// Action is the verb of an authorization decision: what the subject wants to
// do. Actions are a closed, compile-time vocabulary. Adding one without a rule
// in Authorize fails closed at runtime and is caught by TestEveryActionClassified.
type Action string

const (
	// EchoWrite edits an echo; EchoDelete removes it. Both are allowed to its
	// author and to administrators.
	EchoWrite  Action = "echo.write"
	EchoDelete Action = "echo.delete"

	// FileRead opens a file (metadata, stream); FileDelete removes the record
	// and its stored object. Both are allowed to the uploader and administrators.
	FileRead   Action = "file.read"
	FileDelete Action = "file.delete"

	// FileManage lists every file of the instance (admin file browser).
	FileManage Action = "file.manage"

	// UserList reads the user directory; UserManage disables, enables, deletes
	// or re-roles a user; UserTransfer moves ownership. Manage is bounded by
	// rank: only a strictly higher role may act, and may only grant a role
	// below its own.
	UserList     Action = "user.list"
	UserManage   Action = "user.manage"
	UserTransfer Action = "user.transfer"

	// SettingWrite changes instance settings (system, storage, auth, agent,
	// webhook, snapshot schedule, …).
	SettingWrite Action = "setting.write"

	// TokenManage mints, lists and revokes access tokens.
	TokenManage Action = "token.manage"

	// DataManage imports/exports/migrates instance data.
	DataManage Action = "data.manage"

	// ConnectManage manages federation/notification connectors.
	ConnectManage Action = "connect.manage"

	// CopilotUse runs the agent on behalf of the instance.
	CopilotUse Action = "copilot.use"

	// OAuthManage binds and inspects OAuth providers, and consents to OAuth
	// client authorization requests.
	OAuthManage Action = "oauth.manage"

	// CommentManage moderates comments (delete, batch approve/reject, settings).
	CommentManage Action = "comment.manage"

	// TagManage creates or deletes tags.
	TagManage Action = "tag.manage"
)

// AllActions lists every action; the policy must classify all of them.
func AllActions() []Action {
	return []Action{
		EchoWrite, EchoDelete,
		FileRead, FileDelete, FileManage,
		UserList, UserManage, UserTransfer,
		SettingWrite, TokenManage, DataManage, ConnectManage,
		CopilotUse, OAuthManage, CommentManage, TagManage,
	}
}
