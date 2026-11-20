// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	"github.com/lin-snow/ech0/pkg/egress"
)

const (
	cimdCacheTTL      = 10 * time.Minute
	cimdMaxBodyBytes  = 64 << 10
	maxRegisteredApps = 200
)

type cimdEntry struct {
	metadata oauthserverModel.ClientMetadata
	expires  time.Time
}

// OAuthServerService implements the OAuth 2.1 authorization server used by MCP
// clients. Authorization codes are signed JWTs with a shared redemption ledger.
type OAuthServerService struct {
	clients ClientRepository
	users   UserReader
	tokens  TokenIssuer
	http    *http.Client
	now     func() time.Time

	cimdMu    sync.Mutex
	cimdCache map[string]cimdEntry
}

func NewOAuthServerService(
	clients ClientRepository,
	users UserReader,
	tokens TokenIssuer,
) *OAuthServerService {
	return &OAuthServerService{
		clients:   clients,
		users:     users,
		tokens:    tokens,
		http:      egress.NewClient(egress.Guard(), egress.Timeout(10*time.Second)),
		now:       func() time.Time { return time.Now().UTC() },
		cimdCache: make(map[string]cimdEntry),
	}
}

// issuer derives the identifier from the origin the client actually used.
// RFC 8414 requires the metadata issuer to equal the identifier that built the
// well-known URL, so a separately configured server_url must not override it.
// X-Forwarded-Proto is honoured by the HTTP handlers that build origin.
func (s *OAuthServerService) issuer(_ context.Context, origin string) string {
	return strings.TrimRight(strings.TrimSpace(origin), "/")
}

func (s *OAuthServerService) resourceURI(ctx context.Context, origin string) string {
	return s.issuer(ctx, origin) + s.resource().Path
}

func scopesSupported() []string {
	return authz.AllScopes()
}
