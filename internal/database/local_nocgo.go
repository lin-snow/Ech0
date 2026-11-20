// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build !cgo

package database

import "errors"

func enableLocalSQLite() error { return errors.New("local SQLite requires a build with CGO") }
