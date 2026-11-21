// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package model

import fileModel "github.com/lin-snow/ech0/internal/model/file"

// PurgeResult reports what a permanent user deletion removed, so the service
// can finish the parts that live outside SQL: object storage and caches.
type PurgeResult struct {
	EchoIDs []string
	Files   []fileModel.File
}
