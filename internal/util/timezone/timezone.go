// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import "time"

const (
	DefaultTimezoneHeader = "X-Timezone"
	defaultTimezone       = "UTC"
)

func NormalizeTimezone(tz string) string {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return defaultTimezone
	}
	return loc.String()
}

func LoadLocationOrUTC(tz string) *time.Location {
	normalized := NormalizeTimezone(tz)
	loc, err := time.LoadLocation(normalized)
	if err != nil {
		return time.UTC
	}
	return loc
}
