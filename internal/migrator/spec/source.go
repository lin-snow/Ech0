// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package spec

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
)

const SourceDirectory = "files/tmp"

// SourcePath accepts one upload directory beneath the shared staging area.
func SourcePath(payload map[string]any) (string, bool) {
	name, _ := payload["tmp_dir"].(string)
	name = strings.TrimSpace(name)
	if !fs.ValidPath(name) || strings.ContainsAny(name, `\:`) || path.Dir(name) != SourceDirectory {
		return "", false
	}
	return filepath.Join("data", filepath.FromSlash(name)), true
}
