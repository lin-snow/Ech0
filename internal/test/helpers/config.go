// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"testing"

	"github.com/lin-snow/ech0/internal/config"
)

func SetJWTSecret(t *testing.T, secret string) {
	t.Helper()
	cfg := config.Config()
	prev := cfg.Security.JWTSecret
	cfg.Security.JWTSecret = []byte(secret)
	t.Cleanup(func() { cfg.Security.JWTSecret = prev })
}
