// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"github.com/google/wire"
	"github.com/lin-snow/ech0/internal/ratelimit"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

type Deps struct {
	TokenRevoker authService.TokenRevoker
	Limiter      *ratelimit.MemoryLimiter
}

func NewDeps(revoker authService.TokenRevoker, limits *ratelimit.MemoryLimiter) *Deps {
	return &Deps{TokenRevoker: revoker, Limiter: limits}
}

var ProviderSet = wire.NewSet(NewDeps)
