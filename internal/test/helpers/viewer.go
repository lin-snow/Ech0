// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"

	authModel "github.com/lin-snow/ech0/internal/model/auth"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// CtxAsUser builds an interactive browser session. Tests that need a delegated
// access token use CtxAsToken; the distinction matters for scope-only gates
// such as echo:private.
func CtxAsUser(userID string) context.Context {
	return viewer.WithContext(
		context.Background(),
		viewer.NewUserViewerWithToken(userID, authModel.TokenTypeSession, nil, nil, ""),
	)
}

func CtxAsToken(userID, tokenType string, scopes, audience []string, jti string) context.Context {
	return viewer.WithContext(
		context.Background(),
		viewer.NewUserViewerWithToken(userID, tokenType, scopes, audience, jti),
	)
}

func CtxAnonymous() context.Context {
	return viewer.WithContext(context.Background(), viewer.NewNoopViewer())
}
