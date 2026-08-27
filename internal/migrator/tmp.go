// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"os"
	"path/filepath"
	"strings"
)

const TmpRelativeDir = "files/tmp"

func CleanupTmpDirFromPayload(sourcePayload map[string]any) error {
	tmpDir, ok := resolveTmpDir(sourcePayload)
	if !ok {
		return nil
	}
	return os.RemoveAll(tmpDir)
}

func resolveTmpDir(sourcePayload map[string]any) (string, bool) {
	if len(sourcePayload) == 0 {
		return "", false
	}
	tmpDirRaw, ok := sourcePayload["tmp_dir"].(string)
	if !ok || strings.TrimSpace(tmpDirRaw) == "" {
		return "", false
	}
	cleanRelPath := filepath.Clean(filepath.FromSlash(strings.TrimSpace(tmpDirRaw)))
	if cleanRelPath == "." || cleanRelPath == "" || filepath.IsAbs(cleanRelPath) || strings.HasPrefix(cleanRelPath, "..") {
		return "", false
	}

	allowedBaseDir := filepath.Clean(filepath.Join("data", TmpRelativeDir))
	targetDir := filepath.Clean(filepath.Join("data", cleanRelPath))
	if targetDir != allowedBaseDir && !strings.HasPrefix(targetDir, allowedBaseDir+string(os.PathSeparator)) {
		return "", false
	}
	return targetDir, true
}
