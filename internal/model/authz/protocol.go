// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package authz

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// CodeType marks authorization codes signed by this server.
	CodeType = "oauth_code"

	// CodeTTL is deliberately short: a code is single-use and only needs to
	// survive one redirect plus one token request.
	CodeTTL = 60 * time.Second

	// AccessTokenTTL is the lifetime granted to OAuth-issued MCP tokens. The
	// panel still offers shorter presets for manual tokens.
	AccessTokenTTL = 30 * 24 * time.Hour

	AuthorizationPath               = "/oauth/authorize"
	TokenPath                       = "/oauth/token"
	RegistrationPath                = "/oauth/register"
	ProtectedResourcePath           = "/.well-known/oauth-protected-resource"
	ProtectedResourcePathForMCP     = ProtectedResourcePath + "/mcp"
	AuthorizationServerMetadataPath = "/.well-known/oauth-authorization-server"
	MCPResourcePath                 = "/mcp"
	ConsentPath                     = "/oauth/authorize"
)

// DefaultScopes is the least-privilege grant used when a client omits scope.
var DefaultScopes = []string{"echo:read"}

// Error is an RFC 6749 error response. Status is the HTTP status the endpoint
// should return; the OAuth layer never uses Ech0's result envelope.
type Error struct {
	Code        string
	Description string
	Status      int
}

func (e *Error) Error() string { return e.Code + ": " + e.Description }

func errorWith(code, description string, status int) *Error {
	return &Error{Code: code, Description: description, Status: status}
}

func ErrInvalidRequest(description string) *Error {
	return errorWith("invalid_request", description, 400)
}

func ErrInvalidClient(description string) *Error {
	return errorWith("invalid_client", description, 401)
}

func ErrInvalidGrant(description string) *Error {
	return errorWith("invalid_grant", description, 400)
}

func ErrUnauthorizedClient(description string) *Error {
	return errorWith("unauthorized_client", description, 400)
}

func ErrUnsupportedGrantType(description string) *Error {
	return errorWith("unsupported_grant_type", description, 400)
}

func ErrUnsupportedResponseType(description string) *Error {
	return errorWith("unsupported_response_type", description, 400)
}

func ErrInvalidScope(description string) *Error {
	return errorWith("invalid_scope", description, 400)
}

func ErrInvalidRedirectURI(description string) *Error {
	return errorWith("invalid_redirect_uri", description, 400)
}

func ErrInvalidClientMetadata(description string) *Error {
	return errorWith("invalid_client_metadata", description, 400)
}

func ErrAccessDenied(description string) *Error {
	return errorWith("access_denied", description, 403)
}

// ClientMetadata is both the Client ID Metadata Document shape and the
// Dynamic Client Registration request/response shape.
type ClientMetadata struct {
	ClientID                string   `json:"client_id,omitempty"`
	ClientName              string   `json:"client_name,omitempty"`
	ClientURI               string   `json:"client_uri,omitempty"`
	LogoURI                 string   `json:"logo_uri,omitempty"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types,omitempty"`
	ResponseTypes           []string `json:"response_types,omitempty"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method,omitempty"`
	Scope                   string   `json:"scope,omitempty"`
	ApplicationType         string   `json:"application_type,omitempty"`
	ClientIDIssuedAt        int64    `json:"client_id_issued_at,omitempty"`
}

// AuthorizationRequest is a validated /oauth/authorize query. ClientName is
// resolved from CIMD or the registration table so the consent screen never
// renders an unvalidated client's self-asserted identity.
type AuthorizationRequest struct {
	ClientID            string
	ClientName          string
	ClientURI           string
	RedirectURI         string
	ResponseType        string
	Scopes              []string
	State               string
	Resource            string
	CodeChallenge       string
	CodeChallengeMethod string
}

// ConsentContext is what the SPA needs to render the approval screen.
type ConsentContext struct {
	ClientName  string   `json:"client_name"`
	ClientURI   string   `json:"client_uri,omitempty"`
	RedirectURI string   `json:"redirect_uri"`
	Scopes      []string `json:"scopes"`
	Resource    string   `json:"resource"`
	State       string   `json:"state,omitempty"`
}

// CodeClaims is the signed, stateless authorization code. It carries
// everything the token endpoint needs, so no server-side session is required.
type CodeClaims struct {
	Type          string   `json:"typ"`
	UserID        string   `json:"uid"`
	Username      string   `json:"username,omitempty"`
	ClientID      string   `json:"client_id"`
	ClientName    string   `json:"client_name,omitempty"`
	RedirectURI   string   `json:"redirect_uri"`
	Scopes        []string `json:"scope"`
	Resource      string   `json:"resource"`
	CodeChallenge string   `json:"code_challenge"`
	jwt.RegisteredClaims
}

// TokenResponse is an RFC 6749 token endpoint success body.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope,omitempty"`
}

// ProtectedResourceMetadata is the RFC 9728 document served at
// /.well-known/oauth-protected-resource[/mcp].
type ProtectedResourceMetadata struct {
	Resource               string   `json:"resource"`
	AuthorizationServers   []string `json:"authorization_servers"`
	ScopesSupported        []string `json:"scopes_supported,omitempty"`
	BearerMethodsSupported []string `json:"bearer_methods_supported,omitempty"`
	ResourceName           string   `json:"resource_name,omitempty"`
}

// AuthorizationServerMetadata is the RFC 8414 document consumed by MCP
// clients before starting the authorization code flow.
type AuthorizationServerMetadata struct {
	Issuer                                     string   `json:"issuer"`
	AuthorizationEndpoint                      string   `json:"authorization_endpoint"`
	TokenEndpoint                              string   `json:"token_endpoint"`
	RegistrationEndpoint                       string   `json:"registration_endpoint,omitempty"`
	ScopesSupported                            []string `json:"scopes_supported"`
	ResponseTypesSupported                     []string `json:"response_types_supported"`
	GrantTypesSupported                        []string `json:"grant_types_supported"`
	TokenEndpointAuthMethodsSupported          []string `json:"token_endpoint_auth_methods_supported"`
	CodeChallengeMethodsSupported              []string `json:"code_challenge_methods_supported"`
	AuthorizationResponseIssParameterSupported bool     `json:"authorization_response_iss_parameter_supported"`
	ClientIDMetadataDocumentSupported          bool     `json:"client_id_metadata_document_supported"`
	ServiceDocumentation                       string   `json:"service_documentation,omitempty"`
}
