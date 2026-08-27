// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package cache

import (
	"os"
	"testing"

	"github.com/lin-snow/ech0/pkg/leakcheck"
)

func TestMain(m *testing.M) {
	os.Exit(leakcheck.Run(m))
}
