// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package oauthserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
	"github.com/lin-snow/ech0/pkg/egress"
)

// resolveClient returns validated client metadata for a client_id. HTTPS URLs
// are Client ID Metadata Documents (fetched and cached); anything else is a
// client registered through /oauth/register.
func (s *OAuthServerService) resolveClient(
	ctx context.Context,
	clientID string,
) (oauthserverModel.ClientMetadata, *oauthserverModel.Error) {
	if strings.TrimSpace(clientID) == "" {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidRequest("client_id is required")
	}

	var meta oauthserverModel.ClientMetadata
	if isMetadataDocumentURL(clientID) {
		resolved, err := s.fetchClientMetadata(ctx, clientID)
		if err != nil {
			return oauthserverModel.ClientMetadata{}, err
		}
		meta = resolved
	} else {
		client, err := s.clients.FindClient(ctx, clientID)
		if err != nil {
			return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient("unknown client_id")
		}
		meta = client.Metadata()
	}

	if err := validateClientMetadata(meta, clientID); err != nil {
		return oauthserverModel.ClientMetadata{}, err
	}
	return meta, nil
}

func isMetadataDocumentURL(clientID string) bool {
	return strings.HasPrefix(strings.ToLower(clientID), "https://")
}

// fetchClientMetadata implements the Client ID Metadata Document lookup with
// SSRF protection, a hard body limit and HTTP cache-awareness.
func (s *OAuthServerService) fetchClientMetadata(
	ctx context.Context,
	clientID string,
) (oauthserverModel.ClientMetadata, *oauthserverModel.Error) {
	if err := validateMetadataDocumentURL(clientID); err != nil {
		return oauthserverModel.ClientMetadata{}, err
	}

	now := s.now()
	s.cimdMu.Lock()
	if cached, ok := s.cimdCache[clientID]; ok && now.Before(cached.expires) {
		s.cimdMu.Unlock()
		return cached.metadata, nil
	}
	s.cimdMu.Unlock()

	// The URL was validated above; the guarded client below also enforces
	// private/reserved-IP and redirect checks at dial time.
	// codeql[go/request-forgery]
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, clientID, nil)
	if err != nil {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient("invalid client metadata URL")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "ech0-oauth-client")

	// clientID is validated by validateMetadataDocumentURL (scheme, path,
	// egress.Validate) and s.http is an egress.NewClient(egress.Guard()): its
	// dialer rejects private/reserved IPs and its redirect hook re-validates
	// every hop.
	// codeql[go/request-forgery]
	resp, err := s.http.Do(req)
	if err != nil {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient("failed to fetch client metadata")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient(
			fmt.Sprintf("client metadata endpoint returned %d", resp.StatusCode))
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, cimdMaxBodyBytes+1))
	if err != nil || len(body) > cimdMaxBodyBytes {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient("client metadata document is too large")
	}

	var meta oauthserverModel.ClientMetadata
	if err := json.Unmarshal(body, &meta); err != nil {
		return oauthserverModel.ClientMetadata{}, oauthserverModel.ErrInvalidClient("client metadata document is not valid JSON")
	}
	if err := validateClientMetadata(meta, clientID); err != nil {
		return oauthserverModel.ClientMetadata{}, err
	}

	if cacheFor := cacheLifetime(resp.Header.Get("Cache-Control")); cacheFor > 0 {
		s.cacheClientMetadata(clientID, meta, now.Add(cacheFor))
	}
	return meta, nil
}

func validateMetadataDocumentURL(raw string) *oauthserverModel.Error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme != "https" {
		return oauthserverModel.ErrInvalidClient("client_id must be an https URL")
	}
	if parsed.Host == "" || parsed.Path == "" || parsed.Path == "/" {
		return oauthserverModel.ErrInvalidClient("client_id URL must contain a path component")
	}
	if parsed.Fragment != "" || parsed.User != nil {
		return oauthserverModel.ErrInvalidClient("client_id URL must not contain a fragment or userinfo")
	}
	if err := egress.Validate(raw); err != nil {
		return oauthserverModel.ErrInvalidClient("client_id URL is not an allowed outbound target")
	}
	return nil
}

func validateClientPresentation(meta oauthserverModel.ClientMetadata) *oauthserverModel.Error {
	if len([]rune(meta.ClientName)) > 200 {
		return oauthserverModel.ErrInvalidClientMetadata("client_name is too long")
	}
	if len(meta.ClientURI) > 512 || len(meta.LogoURI) > 512 {
		return oauthserverModel.ErrInvalidClientMetadata("client_uri or logo_uri is too long")
	}
	return nil
}

func validateClientMetadata(meta oauthserverModel.ClientMetadata, clientID string) *oauthserverModel.Error {
	if err := validateClientPresentation(meta); err != nil {
		return err
	}
	if meta.ClientID != clientID {
		return oauthserverModel.ErrInvalidClient("client metadata client_id does not match the document URL")
	}
	if len(meta.RedirectURIs) == 0 {
		return oauthserverModel.ErrInvalidClient("client metadata is missing redirect_uris")
	}
	if len(meta.RedirectURIs) > 10 {
		return oauthserverModel.ErrInvalidClient("client registered too many redirect_uris")
	}
	for _, uri := range meta.RedirectURIs {
		if err := validateRedirectURI(uri); err != nil {
			return err
		}
	}
	switch meta.TokenEndpointAuthMethod {
	case "", "none":
	default:
		return oauthserverModel.ErrUnauthorizedClient("only public clients (token_endpoint_auth_method=none) are supported")
	}
	if len(meta.GrantTypes) > 0 && !contains(meta.GrantTypes, "authorization_code") {
		return oauthserverModel.ErrUnauthorizedClient("client does not support the authorization_code grant")
	}
	if len(meta.ResponseTypes) > 0 && !contains(meta.ResponseTypes, "code") {
		return oauthserverModel.ErrUnauthorizedClient("client does not support the code response type")
	}
	return nil
}

func validateRedirectURI(raw string) *oauthserverModel.Error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return oauthserverModel.ErrInvalidRequest("redirect_uri must be an absolute URL")
	}
	if parsed.Fragment != "" {
		return oauthserverModel.ErrInvalidRequest("redirect_uri must not contain a fragment")
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if parsed.Scheme == "http" && isLoopbackHost(parsed.Hostname()) {
		return nil
	}
	return oauthserverModel.ErrInvalidRequest("redirect_uri must use https (http is allowed for loopback only)")
}

func isLoopbackHost(host string) bool {
	switch strings.ToLower(host) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

func cacheLifetime(cacheControl string) time.Duration {
	for _, directive := range strings.Split(cacheControl, ",") {
		trimmed := strings.TrimSpace(directive)
		if strings.EqualFold(trimmed, "no-store") {
			return 0
		}
		key, value, ok := strings.Cut(trimmed, "=")
		if !ok || !strings.EqualFold(key, "max-age") {
			continue
		}
		seconds, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			break
		}
		lifetime := time.Duration(seconds) * time.Second
		if lifetime < time.Minute {
			return time.Minute
		}
		if lifetime > time.Hour {
			return time.Hour
		}
		return lifetime
	}
	return cimdCacheTTL
}

// cacheClientMetadata keeps the CIMD cache bounded: expired entries are
// pruned first, and a single arbitrary entry is evicted at the cap.
func (s *OAuthServerService) cacheClientMetadata(clientID string, meta oauthserverModel.ClientMetadata, expires time.Time) {
	s.cimdMu.Lock()
	defer s.cimdMu.Unlock()

	now := s.now()
	for key, entry := range s.cimdCache {
		if now.After(entry.expires) {
			delete(s.cimdCache, key)
		}
	}
	const maxCachedDocuments = 64
	if len(s.cimdCache) >= maxCachedDocuments {
		for key := range s.cimdCache {
			delete(s.cimdCache, key)
			break
		}
	}
	s.cimdCache[clientID] = cimdEntry{metadata: meta, expires: expires}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
