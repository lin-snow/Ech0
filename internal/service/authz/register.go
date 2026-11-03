// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"
	"crypto/rand"
	"encoding/base64"

	authzModel "github.com/lin-snow/ech0/internal/model/authz"
)

// RegisterClient implements RFC 7591 Dynamic Client Registration. It exists for
// backwards compatibility: the spec deprecates DCR in favour of Client ID
// Metadata Documents, but many deployed clients still use it.
func (s *AuthzService) RegisterClient(
	ctx context.Context,
	meta authzModel.ClientMetadata,
) (authzModel.ClientMetadata, *authzModel.Error) {
	if err := validateClientPresentation(meta); err != nil {
		return authzModel.ClientMetadata{}, err
	}
	if len(meta.RedirectURIs) == 0 {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidRedirectURI("redirect_uris is required")
	}
	if len(meta.RedirectURIs) > 10 {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidRedirectURI("too many redirect_uris")
	}
	for _, uri := range meta.RedirectURIs {
		if err := validateRedirectURI(uri); err != nil {
			return authzModel.ClientMetadata{}, authzModel.ErrInvalidRedirectURI(err.Description)
		}
	}

	switch meta.TokenEndpointAuthMethod {
	case "", "none":
		meta.TokenEndpointAuthMethod = "none"
	default:
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata(
			"only public clients (token_endpoint_auth_method=none) are supported")
	}

	if len(meta.GrantTypes) == 0 {
		meta.GrantTypes = []string{"authorization_code"}
	}
	if len(meta.GrantTypes) != 1 || meta.GrantTypes[0] != "authorization_code" {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata(
			"only the authorization_code grant is supported")
	}
	if len(meta.ResponseTypes) == 0 {
		meta.ResponseTypes = []string{"code"}
	}
	if len(meta.ResponseTypes) != 1 || meta.ResponseTypes[0] != "code" {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata(
			"only the code response type is supported")
	}
	switch meta.ApplicationType {
	case "", "native", "web":
	default:
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata(
			"application_type must be native or web")
	}

	count, err := s.clients.CountClients(ctx)
	if err != nil {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata("failed to register client")
	}
	if count >= maxRegisteredApps {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata(
			"client registration limit reached")
	}

	clientID, err := newClientID()
	if err != nil {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata("failed to register client")
	}
	issuedAt := s.now().Unix()

	client, err := authzModel.NewClientFromMetadata(meta, clientID, issuedAt)
	if err != nil {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata("failed to register client")
	}
	if err := s.clients.CreateClient(ctx, &client); err != nil {
		return authzModel.ClientMetadata{}, authzModel.ErrInvalidClientMetadata("failed to register client")
	}

	meta.ClientID = clientID
	meta.ClientIDIssuedAt = issuedAt
	return meta, nil
}

func newClientID() (string, error) {
	raw := make([]byte, 18)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return "ech0-" + base64.RawURLEncoding.EncodeToString(raw), nil
}
