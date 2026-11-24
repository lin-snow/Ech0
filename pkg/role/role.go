// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package role is the permission vocabulary shared by the domain model and the
// request principal. It is deliberately free of IO so both layers can import it
// without a dependency cycle.
package role

// Role is a user's standing in the single space. It is stored once, in
// users.role, and never copied into a token: a demotion must apply to the next
// request, not to the next refresh.
type Role string

const (
	// Member can read the space and write its own content. It is the default
	// for every account created after the owner exists.
	Member Role = "member"
	// Admin can administer the instance but not its other admins. Two admins
	// of equal standing cannot disable each other, which is what keeps the
	// instance governable without a last-admin guard.
	Admin Role = "admin"
	// Owner is the single account the space belongs to. Exactly one user holds
	// it, enforced by a partial unique index; only the owner manages admins and
	// transfers ownership.
	Owner Role = "owner"
)

func (r Role) Valid() bool {
	return r == Member || r == Admin || r == Owner
}

// Rank orders roles for management decisions only. An unknown or empty role
// ranks below every real role, so an anonymous principal can never manage.
func (r Role) Rank() int {
	switch r {
	case Owner:
		return 3
	case Admin:
		return 2
	case Member:
		return 1
	default:
		return 0
	}
}

// AtLeast reports whether r is at or above min. It answers "may this principal
// enter an admin-only surface", never "may it act on this particular record".
func (r Role) AtLeast(min Role) bool {
	return r.Rank() >= min.Rank()
}
