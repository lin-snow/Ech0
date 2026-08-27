// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package viewer

type Context interface {
	UserID() string
	TokenType() string
	Scopes() []string
	Audience() []string
	TokenID() string
}
