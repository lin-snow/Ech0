// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type UserInfoDto struct {
	Username string `json:"username"`

	Password string `json:"password"`

	Email string `json:"email"`

	IsAdmin bool `json:"is_admin"`

	IsOwner bool `json:"is_owner"`

	Avatar string `json:"avatar"`

	AvatarFileID string `json:"avatar_file_id"`

	Locale string `json:"locale"`
}

type OAuthInfoDto struct {
	Provider string `json:"provider"`
	UserID   string `json:"user_id"`
	OAuthID  string `json:"oauth_id"`
	Issuer   string `json:"issuer"`
	AuthType string `json:"auth_type"`
}
