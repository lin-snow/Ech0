// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package viewer carries the authenticated principal of a request.
//
// The principal is a value, not a nullable pointer: an unauthenticated request
// has KindAnonymous, a background job has KindSystem. Decision code therefore
// takes one path ("what kind is this?") instead of defending against nil at
// every call site, and a missing auth middleware degrades to anonymous rather
// than to whatever a zero-valued struct happens to allow.
package viewer

import "github.com/lin-snow/ech0/pkg/role"

// Kind says which credential produced the principal. It changes what the
// principal may do (a session carries the scopes of its live role; an access
// token carries the scopes it was minted with) and it is never inferred from
// empty strings.
type Kind uint8

const (
	// KindAnonymous is the zero value: absent credential, unknown context, or
	// an explicit anonymous caller. Zero-valued principals fail closed.
	KindAnonymous Kind = iota
	// KindSession is an interactive browser session. Its scopes are derived
	// from the user's role on every request.
	KindSession
	// KindAccess is a bearer access token. Its scopes were fixed at mint time.
	KindAccess
	// KindSystem is an internal writer (background job, event subscriber). It
	// has no user identity and never satisfies a user-facing authorization
	// decision.
	KindSystem
)

// Context is the authenticated principal attached to a request. Role is read
// from the database on every request and is never carried in a token, so a
// demotion or ownership transfer takes effect immediately.
type Context interface {
	Kind() Kind
	// UserID is empty for anonymous and system principals.
	UserID() string
	// Role is empty for anonymous and system principals.
	Role() role.Role
	// Scopes is the capability set of this credential. Sessions derive it from
	// Role; access tokens carry the scopes they were minted with.
	Scopes() []string
	Audience() []string
	TokenID() string
}

// IsAnonymous reports whether the principal has no identity at all.
func IsAnonymous(v Context) bool {
	return v == nil || v.Kind() == KindAnonymous || v.UserID() == ""
}

// IsSession reports whether the principal is an interactive browser session.
func IsSession(v Context) bool {
	return v != nil && v.Kind() == KindSession
}

// IsSystem reports whether the principal is an internal writer.
func IsSystem(v Context) bool {
	return v != nil && v.Kind() == KindSystem
}
