// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"net/url"

	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
)

// Service is the authorization-server surface consumed by the HTTP handler.
type Service interface {
	ProtectedResourceMetadata(ctx context.Context, origin string) oauthserverModel.ProtectedResourceMetadata
	AuthorizationServerMetadata(ctx context.Context, origin string) oauthserverModel.AuthorizationServerMetadata
	Consent(ctx context.Context, origin string, query url.Values, userID string) (oauthserverModel.ConsentContext, *oauthserverModel.Error)
	Approve(ctx context.Context, origin string, query url.Values, userID string) (string, *oauthserverModel.Error)
	Deny(ctx context.Context, origin string, query url.Values) (string, *oauthserverModel.Error)
	RegisterClient(ctx context.Context, meta oauthserverModel.ClientMetadata) (oauthserverModel.ClientMetadata, *oauthserverModel.Error)
	Exchange(ctx context.Context, origin string, form url.Values) (oauthserverModel.TokenResponse, *oauthserverModel.Error)
}

// ClientRepository stores dynamically registered clients. CIMD clients are
// resolved on demand and never persisted.
type ClientRepository interface {
	FindClient(ctx context.Context, clientID string) (oauthserverModel.Client, error)
	CreateClient(ctx context.Context, client *oauthserverModel.Client) error
	CountClients(ctx context.Context) (int64, error)
}

// UserReader resolves the approving principal. Only administrators may mint
// access tokens in Ech0, so the authorization server inherits that rule.
type UserReader interface {
	GetUserByID(ctx context.Context, id string) (userModel.User, error)
}

// TokenIssuer mints an access token for an explicit principal; the token
// endpoint has no session.
type TokenIssuer interface {
	IssueForUser(ctx context.Context, userID string, dto *settingModel.AccessTokenSettingDto) (string, error)
}
