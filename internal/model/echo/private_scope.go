// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

// PrivateScope decides which private echoes a listing may return. A signed-in
// viewer sees public echoes plus their own private ones; administrators get no
// extra reach here, because the feed must never surface somebody else's private
// echo. All is reserved for trusted server-side aggregates (instance counters,
// the embedding index) that legitimately need every row.
type PrivateScope struct {
	ViewerID string
	All      bool
}

// CacheTag renders the scope as the cache-key fragment used by echo listings,
// so a cached page can never be replayed to a different viewer.
func (s PrivateScope) CacheTag() string {
	switch {
	case s.All:
		return "all"
	case s.ViewerID == "":
		return "public"
	default:
		return "u:" + s.ViewerID
	}
}
