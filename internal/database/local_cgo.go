// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build cgo

package database

import sqlitevec "github.com/asg017/sqlite-vec-go-bindings/cgo"

func enableLocalSQLite() error { sqlitevec.Auto(); return nil }
