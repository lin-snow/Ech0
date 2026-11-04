// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package access owns the lifecycle primitive for Ech0 access tokens: policy
// validation, JWT minting and persistence. The settings domain keeps the admin
// CRUD surface; the device flow and the OAuth authorization server depend on
// this package instead of the whole settings aggregate.
package access

import (
	"context"

	model "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

// TokenRepository persists a minted token. It is bound to the settings
// repository because the settings aggregate owns the access_token_settings
// table.
type TokenRepository interface {
	CreateAccessToken(ctx context.Context, token *model.AccessTokenSetting) error
}

// UserReader resolves the token owner. Only administrators may mint tokens.
type UserReader interface {
	GetUserByID(ctx context.Context, id string) (userModel.User, error)
}
