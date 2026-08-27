// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package url

import (
	"strings"
)

func TrimURL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSpace(url)
	url = strings.TrimPrefix(url, "/")
	url = strings.TrimSuffix(url, "/")
	return url
}
