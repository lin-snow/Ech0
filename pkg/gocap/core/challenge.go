// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

func (s *Service) CreateChallenge(siteKey string) (*ChallengeResponse, error) {
	site, ok := s.Store.GetSite(siteKey)
	if !ok {
		return nil, NewNotFound("Invalid site key")
	}

	c := site.ChallengeCount
	if c <= 0 {
		c = 80
	}
	saltSize := site.SaltSize
	if saltSize <= 0 {
		saltSize = 32
	}
	difficulty := site.Difficulty
	if difficulty <= 0 {
		difficulty = 4
	}

	nonce, err := randomHex(s.RNG, 25)
	if err != nil {
		return nil, NewInternal("Failed to generate challenge nonce")
	}

	now := s.Now()
	exp := now.Add(s.ChallengeTTL).UnixMilli()

	claims := ChallengeClaims{
		SiteKey:        siteKey,
		Nonce:          nonce,
		ChallengeCount: c,
		SaltSize:       saltSize,
		Difficulty:     difficulty,
		ExpiresAtMS:    exp,
		IssuedAtMS:     now.UnixMilli(),
	}

	token, err := SignChallengeToken(claims, site.JWTSecret)
	if err != nil {
		return nil, NewInternal("Failed to sign challenge token")
	}

	return &ChallengeResponse{
		Challenge: ChallengeParams{
			C: c,
			S: saltSize,
			D: difficulty,
		},
		Token:   token,
		Expires: exp,
	}, nil
}
