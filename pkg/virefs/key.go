// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"fmt"
	"path"
	"strings"
)

func CleanKey(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return "", nil
	}
	cleaned := path.Clean(raw)
	if cleaned == "." {
		return "", nil
	}
	if strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "/../") || strings.HasSuffix(cleaned, "/..") {
		return "", fmt.Errorf("%w: path traversal not allowed: %q", ErrInvalidKey, raw)
	}
	return cleaned, nil
}
