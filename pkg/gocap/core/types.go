// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

type ChallengeParams struct {
	C int `json:"c"`
	S int `json:"s"`
	D int `json:"d"`
}

type ChallengeResponse struct {
	Challenge ChallengeParams `json:"challenge"`
	Token     string          `json:"token"`
	Expires   int64           `json:"expires"`
}

type RedeemRequest struct {
	Token        string `json:"token"`
	Solutions    []int  `json:"solutions"`
	Instr        any    `json:"instr,omitempty"`
	InstrTimeout bool   `json:"instr_timeout,omitempty"`
	InstrBlocked bool   `json:"instr_blocked,omitempty"`
}

type RedeemResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Expires int64  `json:"expires"`
}

type SiteVerifyRequest struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

type SiteVerifyResponse struct {
	Success bool `json:"success"`
}

type ChallengeClaims struct {
	SiteKey        string `json:"sk"`
	Nonce          string `json:"n"`
	ChallengeCount int    `json:"c"`
	SaltSize       int    `json:"s"`
	Difficulty     int    `json:"d"`
	ExpiresAtMS    int64  `json:"exp"`
	IssuedAtMS     int64  `json:"iat"`
}
