// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build cgo

package pagefile

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestReopenPublishedBlocks(t *testing.T) {
	original := New(0, nil)
	require.NoError(t, UseSQLite(t.Context(), original, func(db *sql.DB) error {
		_, err := db.Exec("CREATE TABLE sample(id INTEGER PRIMARY KEY, b BLOB); INSERT INTO sample VALUES(1,zeroblob(200000))")
		return err
	}))
	reads := 0
	another := New(original.Size, func(i int64) ([]byte, error) { reads++; return original.Dirty[i], nil })
	require.NoError(t, UseSQLite(t.Context(), another, func(db *sql.DB) error {
		var n int
		if err := db.QueryRow("SELECT length(b) FROM sample").Scan(&n); err != nil {
			return err
		}
		require.Equal(t, 200000, n)
		_, err := db.Exec("INSERT INTO sample VALUES(2,zeroblob(80))")
		return err
	}))
	require.Positive(t, reads)
	require.Less(t, len(another.Dirty), len(original.Dirty))
	require.NoError(t, UseSQLite(t.Context(), original, func(db *sql.DB) error {
		var n int
		err := db.QueryRow("SELECT count(*) FROM sample").Scan(&n)
		require.Equal(t, 1, n)
		return err
	}))
}
