// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"

	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
)

// ProtectedResourceMetadata serves RFC 9728 discovery for the MCP endpoint.
func (s *OAuthServerService) ProtectedResourceMetadata(
	ctx context.Context,
	origin string,
) oauthserverModel.ProtectedResourceMetadata {
	issuer := s.issuer(ctx, origin)
	return oauthserverModel.ProtectedResourceMetadata{
		Resource:               issuer + oauthserverModel.MCPResourcePath,
		AuthorizationServers:   []string{issuer},
		ScopesSupported:        scopesSupported(),
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Ech0 MCP",
	}
}

// AuthorizationServerMetadata serves RFC 8414 metadata. CIMD is advertised as
// supported; Dynamic Client Registration is still offered for clients that
// have not migrated yet.
func (s *OAuthServerService) AuthorizationServerMetadata(
	ctx context.Context,
	origin string,
) oauthserverModel.AuthorizationServerMetadata {
	issuer := s.issuer(ctx, origin)
	return oauthserverModel.AuthorizationServerMetadata{
		Issuer:                                     issuer,
		AuthorizationEndpoint:                      issuer + oauthserverModel.AuthorizationPath,
		TokenEndpoint:                              issuer + oauthserverModel.TokenPath,
		RegistrationEndpoint:                       issuer + oauthserverModel.RegistrationPath,
		ScopesSupported:                            scopesSupported(),
		ResponseTypesSupported:                     []string{"code"},
		GrantTypesSupported:                        []string{"authorization_code"},
		TokenEndpointAuthMethodsSupported:          []string{"none"},
		CodeChallengeMethodsSupported:              []string{"S256"},
		AuthorizationResponseIssParameterSupported: true,
		ClientIDMetadataDocumentSupported:          true,
	}
}
