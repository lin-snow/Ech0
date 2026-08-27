// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package middleware

import (
	"github.com/google/wire"
	authService "github.com/lin-snow/ech0/internal/service/auth"
)

type Deps struct {
	TokenRevoker authService.TokenRevoker
}

func NewDeps(revoker authService.TokenRevoker) *Deps {
	return &Deps{TokenRevoker: revoker}
}

var ProviderSet = wire.NewSet(NewDeps)
