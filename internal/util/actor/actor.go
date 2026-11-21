// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package actor carries the request-scoped reaction identity.
//
// A reaction actor is a single opaque string:
//
//	u:<userID>   an authenticated user
//	v:<visitor>  a client-generated visitor id (X-Ech0-Visitor header)
//	ip:<hash>    the fallback when no visitor id is supplied
//
// The service prefers the authenticated user from the viewer context and
// falls back to whatever the HTTP middleware stored here.
package actor

import "context"

type contextKey struct{}

// UserActor renders the actor string for an authenticated user. Callers that
// filter reactions by owner must use this instead of concatenating the prefix
// themselves, so the identity format has one definition.
func UserActor(userID string) string {
	return "u:" + userID
}

// WithContext stores the request's fallback actor.
func WithContext(ctx context.Context, value string) context.Context {
	if ctx == nil || value == "" {
		return ctx
	}
	return context.WithValue(ctx, contextKey{}, value)
}

// FromContext returns the request's fallback actor, or "" when unset.
func FromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(contextKey{}).(string)
	return value
}
