// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"context"

	authzModel "github.com/lin-snow/ech0/internal/model/authz"
)

// ProtectedResourceMetadata serves RFC 9728 discovery for the MCP endpoint.
func (s *AuthzService) ProtectedResourceMetadata(
	ctx context.Context,
	origin string,
) authzModel.ProtectedResourceMetadata {
	issuer := s.issuer(ctx, origin)
	return authzModel.ProtectedResourceMetadata{
		Resource:               issuer + authzModel.MCPResourcePath,
		AuthorizationServers:   []string{issuer},
		ScopesSupported:        scopesSupported(),
		BearerMethodsSupported: []string{"header"},
		ResourceName:           "Ech0 MCP",
	}
}

// AuthorizationServerMetadata serves RFC 8414 metadata. CIMD is advertised as
// supported; Dynamic Client Registration is still offered for clients that
// have not migrated yet.
func (s *AuthzService) AuthorizationServerMetadata(
	ctx context.Context,
	origin string,
) authzModel.AuthorizationServerMetadata {
	issuer := s.issuer(ctx, origin)
	return authzModel.AuthorizationServerMetadata{
		Issuer:                                     issuer,
		AuthorizationEndpoint:                      issuer + authzModel.AuthorizationPath,
		TokenEndpoint:                              issuer + authzModel.TokenPath,
		RegistrationEndpoint:                       issuer + authzModel.RegistrationPath,
		ScopesSupported:                            scopesSupported(),
		ResponseTypesSupported:                     []string{"code"},
		GrantTypesSupported:                        []string{"authorization_code"},
		TokenEndpointAuthMethodsSupported:          []string{"none"},
		CodeChallengeMethodsSupported:              []string{"S256"},
		AuthorizationResponseIssParameterSupported: true,
		ClientIDMetadataDocumentSupported:          true,
	}
}
