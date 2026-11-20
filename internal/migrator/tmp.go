// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migrator

import (
	"os"

	"github.com/lin-snow/ech0/internal/migrator/spec"
)

const TmpRelativeDir = spec.SourceDirectory

func CleanupTmpDirFromPayload(sourcePayload map[string]any) error {
	tmpDir, ok := spec.SourcePath(sourcePayload)
	if !ok {
		return nil
	}
	return os.RemoveAll(tmpDir)
}
