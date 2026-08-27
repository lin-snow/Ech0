// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

func (s *Service) SiteVerify(siteKey string, req SiteVerifyRequest) (*SiteVerifyResponse, error) {
	if req.Secret == "" || req.Response == "" {
		return nil, NewBadRequest("Missing required parameters")
	}

	site, ok := s.Store.GetSite(siteKey)
	if !ok {
		return nil, NewNotFound("Invalid site key")
	}

	if !SecureSecretEqual(req.Secret, site.SecretHash, s.SecretPepper) {
		return nil, NewForbidden("Invalid site key or secret")
	}

	found, expired, err := s.Store.ConsumeRedeemToken(siteKey, req.Response, s.Now())
	if err != nil {
		return nil, NewInternal("Failed to consume redeem token")
	}
	if !found {
		return nil, NewNotFound("Token not found")
	}
	if expired {
		return nil, NewForbidden("Token expired")
	}

	return &SiteVerifyResponse{Success: true}, nil
}
