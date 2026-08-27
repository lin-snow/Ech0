// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import (
	"strconv"
	"time"
)

func (s *Service) Redeem(siteKey string, req RedeemRequest) (*RedeemResponse, error) {
	if req.Token == "" || len(req.Solutions) == 0 {
		return nil, NewBadRequest("Missing required fields")
	}

	site, ok := s.Store.GetSite(siteKey)
	if !ok {
		return nil, NewNotFound("Invalid site key")
	}

	claims, valid := VerifyChallengeToken(req.Token, site.JWTSecret)
	if !valid {
		return nil, NewForbidden("Invalid challenge token")
	}

	if claims.SiteKey != siteKey {
		return nil, NewForbidden("Challenge token does not match site key")
	}

	now := s.Now()
	if claims.ExpiresAtMS <= now.UnixMilli() {
		return nil, NewForbidden("Challenge expired")
	}

	sig := TokenSignatureHash(req.Token)

	if claims.ChallengeCount <= 0 || claims.ChallengeCount > 500 {
		return nil, NewBadRequest("Invalid challenge count")
	}
	if claims.Difficulty <= 0 || claims.Difficulty > 8 {
		return nil, NewBadRequest("Invalid difficulty")
	}
	if claims.SaltSize <= 0 || claims.SaltSize > 128 {
		return nil, NewBadRequest("Invalid salt size")
	}

	if len(req.Solutions) != claims.ChallengeCount {
		return nil, NewBadRequest("Invalid solutions")
	}

	isValid := VerifySolutions(
		req.Token,
		claims.ChallengeCount,
		claims.SaltSize,
		claims.Difficulty,
		req.Solutions,
	)
	if !isValid {
		return nil, NewForbidden("Invalid solution")
	}

	ttl := max(time.Until(time.UnixMilli(claims.ExpiresAtMS)), time.Second)
	marked, err := s.Store.TryMarkChallengeSigUsed(sig, now.Add(ttl), now)
	if err != nil {
		return nil, NewInternal("Failed to mark challenge token as used")
	}
	if !marked {
		return nil, NewForbidden("Challenge already redeemed")
	}

	redeemRandom, err := randomHex(s.RNG, 24)
	if err != nil {
		return nil, NewInternal("Failed to generate redeem token")
	}
	redeemToken := siteKey + "_" + redeemRandom + "_" + strconv.FormatInt(now.UnixNano(), 36)
	redeemExp := now.Add(s.RedeemTTL)
	if err := s.Store.StoreRedeemToken(siteKey, redeemToken, redeemExp); err != nil {
		return nil, NewInternal("Failed to persist redeem token")
	}

	return &RedeemResponse{
		Success: true,
		Token:   redeemToken,
		Expires: redeemExp.UnixMilli(),
	}, nil
}
