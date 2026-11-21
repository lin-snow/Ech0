// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// PrivateScope decides which private echoes a listing may return. Private is
// an audience marker, not an isolation boundary: hidden from anonymous
// visitors, visible to every signed-in member. All is reserved for trusted
// server-side aggregates (instance counters, the embedding index) that
// legitimately need every row.
type PrivateScope struct {
	Authenticated bool
	All           bool
}

// CacheTag renders the scope as the cache-key fragment used by echo listings,
// so a cached anonymous page can never be replayed to a signed-in viewer.
func (s PrivateScope) CacheTag() string {
	switch {
	case s.All:
		return "all"
	case s.Authenticated:
		return "authed"
	default:
		return "public"
	}
}
