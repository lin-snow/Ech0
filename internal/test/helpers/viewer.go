// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"

	"github.com/lin-snow/ech0/pkg/viewer"
)

func CtxAsUser(userID string) context.Context {
	return viewer.WithContext(context.Background(), viewer.NewUserViewer(userID))
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
