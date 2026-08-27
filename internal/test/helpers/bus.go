// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package helpers

import (
	"context"
	"testing"

	"github.com/lin-snow/ech0/pkg/busen"
)

func NewTestBus(t *testing.T) *busen.Bus {
	t.Helper()
	b := busen.New()
	t.Cleanup(func() { _ = b.Close(context.Background()) })
	return b
}
