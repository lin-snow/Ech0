// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"testing"

	"github.com/lin-snow/ech0/internal/storage"
)

func NewTestStorage(t *testing.T) *storage.Manager {
	t.Helper()
	return storage.NewStorageManagerForTest(t.TempDir())
}
