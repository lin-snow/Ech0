// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

import "github.com/lin-snow/ech0/pkg/role"

// Principal is the concrete Context. It is immutable by convention: the
// middleware builds one per request and everything downstream reads it.
type Principal struct {
	kind     Kind
	userID   string
	role     role.Role
	scopes   []string
	audience []string
	tokenID  string
}

// Anonymous is the principal of an unauthenticated request: no identity, no
// role, no scopes. It is a value, so callers never branch on nil.
var Anonymous Context = Principal{}

// System is the principal of an internal writer. It is explicit rather than an
// empty user ID, so "no user" is a decision input and not an accident.
var System Context = Principal{kind: KindSystem}

// NewSession builds an interactive-session principal. Scopes are derived from
// the role by the caller (middleware), so a role change takes effect on the
// next request without touching the token.
func NewSession(userID string, r role.Role, scopes []string) Principal {
	return Principal{kind: KindSession, userID: userID, role: r, scopes: scopes}
}

// NewAccess builds an access-token principal from the claims it was minted
// with. The role still comes from the database, never from the token.
func NewAccess(userID, tokenID string, r role.Role, scopes, audience []string) Principal {
	return Principal{
		kind:     KindAccess,
		userID:   userID,
		role:     r,
		scopes:   append([]string(nil), scopes...),
		audience: append([]string(nil), audience...),
		tokenID:  tokenID,
	}
}

func (p Principal) Kind() Kind         { return p.kind }
func (p Principal) UserID() string     { return p.userID }
func (p Principal) Role() role.Role    { return p.role }
func (p Principal) Scopes() []string   { return p.scopes }
func (p Principal) Audience() []string { return p.audience }
func (p Principal) TokenID() string    { return p.tokenID }

var _ Context = Principal{}
