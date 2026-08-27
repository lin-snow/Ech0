// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"fmt"
	"strings"
	"time"
)

func FormatUnix(sec int64) string {
	return time.Unix(sec, 0).UTC().Format(time.RFC3339)
}

func ParseTime(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty timestamp")
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf("invalid RFC3339 timestamp %q: %w", raw, err)
	}
	return t.Unix(), nil
}
