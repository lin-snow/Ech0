// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"

	"github.com/lin-snow/ech0/internal/authz"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	"github.com/lin-snow/ech0/pkg/role"
	"github.com/lin-snow/ech0/pkg/viewer"
)

// CtxAsUser builds an interactive browser session. Tests that need a delegated
// access token use CtxAsToken; the distinction matters because a session's
// scopes are derived from its role, not chosen by the test. The role defaults
// to member; pass role.Admin or role.Owner only when the test's subject is
// meant to hold that rank, because the live role — not a stored flag — is what
// authorization reads.
func CtxAsUser(userID string, roles ...role.Role) context.Context {
	r := roleOrMember(roles)
	return viewer.WithContext(
		context.Background(),
		viewer.NewSession(userID, r, authz.ScopesForRole(r)),
	)
}

// CtxAsToken builds an access-token principal with an explicit, frozen scope
// set. A session type is accepted for tests that model an older transport, and
// is upgraded to the role-derived scope set so the helper cannot create a
// capability state the middleware never produces.
func CtxAsToken(userID, tokenType string, scopes, audience []string, jti string, roles ...role.Role) context.Context {
	r := roleOrMember(roles)
	if tokenType == authModel.TokenTypeSession {
		return viewer.WithContext(
			context.Background(),
			viewer.NewSession(userID, r, authz.ScopesForRole(r)),
		)
	}
	return viewer.WithContext(
		context.Background(),
		viewer.NewAccess(userID, jti, r, scopes, audience),
	)
}

func roleOrMember(roles []role.Role) role.Role {
	if len(roles) > 0 {
		return roles[0]
	}
	return role.Member
}

func CtxAnonymous() context.Context {
	return viewer.WithContext(context.Background(), viewer.Anonymous)
}
