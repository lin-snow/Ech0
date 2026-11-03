// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"net/url"

	authzModel "github.com/lin-snow/ech0/internal/model/authz"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

// Service is the authorization-server surface consumed by the HTTP handler.
type Service interface {
	ProtectedResourceMetadata(ctx context.Context, origin string) authzModel.ProtectedResourceMetadata
	AuthorizationServerMetadata(ctx context.Context, origin string) authzModel.AuthorizationServerMetadata
	Consent(ctx context.Context, origin string, query url.Values, userID string) (authzModel.ConsentContext, *authzModel.Error)
	Approve(ctx context.Context, origin string, query url.Values, userID string) (string, *authzModel.Error)
	Deny(ctx context.Context, origin string, query url.Values) (string, *authzModel.Error)
	RegisterClient(ctx context.Context, meta authzModel.ClientMetadata) (authzModel.ClientMetadata, *authzModel.Error)
	Exchange(ctx context.Context, origin string, form url.Values) (authzModel.TokenResponse, *authzModel.Error)
}

// ClientRepository stores dynamically registered clients. CIMD clients are
// resolved on demand and never persisted.
type ClientRepository interface {
	FindClient(ctx context.Context, clientID string) (authzModel.Client, error)
	CreateClient(ctx context.Context, client *authzModel.Client) error
	CountClients(ctx context.Context) (int64, error)
}

// UserReader resolves the approving principal. Only administrators may mint
// access tokens in Ech0, so the authorization server inherits that rule.
type UserReader interface {
	GetUserByID(ctx context.Context, id string) (userModel.User, error)
}

// TokenIssuer mints an access token for a user without requiring a viewer in
// the context; the token endpoint has no session.
type TokenIssuer interface {
	IssueAccessTokenForUser(ctx context.Context, userID string, dto *settingModel.AccessTokenSettingDto) (string, error)
}
