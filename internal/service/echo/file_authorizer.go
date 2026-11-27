// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"

	"github.com/lin-snow/ech0/internal/authz"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// EchoFileAuthorizer is the echo domain's answer to the file domain's only
// question about a post: may this caller see its attachments?
//
// It is registered with the file domain at boot, so the file package never
// imports the echo package and an echo's visibility can only be decided here.
type EchoFileAuthorizer struct {
	echos Repository
}

func NewEchoFileAuthorizer(echos Repository) *EchoFileAuthorizer {
	return &EchoFileAuthorizer{echos: echos}
}

// Kind is the owner kind whose links this authorizer speaks for.
func (a *EchoFileAuthorizer) Kind() string {
	return fileModel.OwnerKindEcho
}

// Authorize denies by default. A deleted post, or one that no longer exists,
// answers "no" rather than propagating a lookup failure: the caller must not be
// able to tell a missing post from a hidden one.
func (a *EchoFileAuthorizer) Authorize(ctx context.Context, ownerID string, p viewer.Context) (bool, error) {
	echo, err := a.echos.GetEchosById(ctx, ownerID)
	if err != nil || echo == nil || echo.DeletedAt != nil {
		return false, nil
	}
	if !echo.Private {
		return true, nil
	}
	// A private post is readable by every signed-in principal whose role carries
	// echo:private. Sessions derive that scope from their role; an access token
	// must have been minted with it, so echo:read alone cannot widen the audience.
	if viewer.IsAnonymous(p) {
		return false, nil
	}
	return authz.Has(p.Scopes(), authz.ScopeEchoPrivate), nil
}
