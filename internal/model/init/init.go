// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

type Status struct {
	Initialized bool `json:"initialized"`
	OwnerExists bool `json:"owner_exists"`
}

// InitOwnerDto is the first-run registration payload. It is intentionally
// separate from RegisterDto: site_title has meaning only while the instance is
// being claimed, and must not become part of the public signup contract.
type InitOwnerDto struct {
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email"`
	Locale    string `json:"locale,omitempty"`
	SiteTitle string `json:"site_title,omitempty"`
}
