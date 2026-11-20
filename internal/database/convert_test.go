// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

//go:build cgo

package database

import (
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"path/filepath"
	"testing"
)

func TestConvertSQLite(t *testing.T) {
	require.NoError(t, enableLocalSQLite())
	source := filepath.Join(t.TempDir(), "source.db")
	target := filepath.Join(t.TempDir(), "target.db")
	src, err := openSQLite(source, logger.Silent)
	require.NoError(t, err)
	db, _ := src.DB()
	t.Cleanup(func() { db.Close() })
	require.NoError(t, migrateModels(src))
	require.NoError(t, src.Create(&commonModel.KeyValue{Key: commonModel.EchosPrivateCreatedDroppedKey, Value: "done"}).Error)
	require.NoError(t, src.Exec("CREATE VIRTUAL TABLE vec_echo USING vec0(echo_id TEXT PRIMARY KEY, embedding FLOAT[2])").Error)
	require.NoError(t, src.Exec("INSERT INTO vec_echo(echo_id,embedding) VALUES('v','[1.0,2.0]')").Error)
	require.NoError(t, ConvertSQLite(t.Context(), source, target))
	require.Error(t, ConvertSQLite(t.Context(), source, target))
	require.False(t, src.Migrator().HasTable("ech0_schema"))
	dest, err := gorm.Open(sqlite.Open(target), buildGormConfig(logger.Silent))
	require.NoError(t, err)
	destPool, _ := dest.DB()
	t.Cleanup(func() { destPool.Close() })
	require.NoError(t, checkRemoteSchema(dest))
	var value string
	require.NoError(t, dest.Raw("SELECT hex(embedding) FROM vec_echo WHERE echo_id='v'").Row().Scan(&value))
	require.Equal(t, "0000803F00000040", value)
}
