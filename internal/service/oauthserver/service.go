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
// clients. It is deliberately stateless for authorization codes: the code is a
// signed JWT, and the only in-memory state is a short-lived single-use ledger.
type OAuthServerService struct {
	clients ClientRepository
	users   UserReader
	tokens  TokenIssuer
	http    *http.Client
	now     func() time.Time

	cimdMu    sync.Mutex
	cimdCache map[string]cimdEntry

	codeMu   sync.Mutex
	consumed map[string]time.Time
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
		consumed:  make(map[string]time.Time),
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

// consumeCode enforces single use of an authorization code. The ledger is
// in-memory and keyed by jti; codes live for CodeTTL, so the map stays tiny.
func (s *OAuthServerService) consumeCode(jti string) bool {
	now := s.now()
	s.codeMu.Lock()
	defer s.codeMu.Unlock()

	for key, expiry := range s.consumed {
		if now.After(expiry) {
			delete(s.consumed, key)
		}
	}
	if _, used := s.consumed[jti]; used {
		return false
	}
	s.consumed[jti] = now.Add(2 * oauthserverModel.CodeTTL)
	return true
}
