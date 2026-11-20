// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build !cgo

package pagefile

import (
	"context"
	"database/sql"
	"errors"
)

func UseSQLite(context.Context, *File, func(*sql.DB) error) error {
	return errors.New("materializing native SQLite snapshot pages requires CGO")
}
