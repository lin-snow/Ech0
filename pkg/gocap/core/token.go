// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package core

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
)

var b64url = base64.RawURLEncoding

func SignChallengeToken(claims ChallengeClaims, secret []byte) (string, error) {
	headerJSON := []byte(`{"alg":"HS256","typ":"JWT"}`)
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := b64url.EncodeToString(headerJSON)
	payload := b64url.EncodeToString(payloadJSON)
	signingInput := header + "." + payload

	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(signingInput))
	sig := b64url.EncodeToString(h.Sum(nil))

	return signingInput + "." + sig, nil
}

func VerifyChallengeToken(token string, secret []byte) (*ChallengeClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, false
	}
	signingInput := parts[0] + "." + parts[1]

	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(signingInput))
	expectedSig := h.Sum(nil)

	actualSig, err := b64url.DecodeString(parts[2])
	if err != nil {
		return nil, false
	}
	if !hmac.Equal(expectedSig, actualSig) {
		return nil, false
	}

	payloadJSON, err := b64url.DecodeString(parts[1])
	if err != nil {
		return nil, false
	}
	var claims ChallengeClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, false
	}
	return &claims, true
}

func TokenSignatureHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
