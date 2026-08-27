// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cap

import "time"

type SiteRegistration struct {
	SiteKey        string
	Secret         string
	Difficulty     int
	ChallengeCount int
	SaltSize       int
}

type RateLimitConfig struct {
	Max    int
	Window time.Duration
	Scope  string
}
