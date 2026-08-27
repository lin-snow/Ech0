// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package migration_test

import (
	"fmt"
	"testing"

	"github.com/lin-snow/ech0/internal/database"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newLocalAuthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	database.SetDB(db)
	require.NoError(t, database.MigrateDB())
	return db
}

func addLegacyPasswordColumn(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.Exec(`ALTER TABLE users ADD COLUMN password TEXT`).Error)
}

func seedLegacyUser(t *testing.T, db *gorm.DB, id, username, md5pw string) {
	t.Helper()
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, username, password, is_admin, is_owner, locale) VALUES (?,?,?,?,?,?)`,
		id, username, md5pw, false, false, "zh-CN",
	).Error)
}

type localAuthRow struct {
	UserID       string
	PasswordHash string
	PasswordAlgo string
}

func readLocalAuth(t *testing.T, db *gorm.DB, userID string) (localAuthRow, bool) {
	t.Helper()
	var row localAuthRow
	err := db.Raw(
		`SELECT user_id, password_hash, password_algo FROM user_local_auth WHERE user_id = ?`, userID,
	).Scan(&row).Error
	require.NoError(t, err)
	return row, row.UserID != ""
}

func countLocalAuth(t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	require.NoError(t, db.Table("user_local_auth").Count(&n).Error)
	return n
}

func TestUserLocalAuthBackfillMigrator_BackfillsLegacyPassword(t *testing.T) {
	db := newLocalAuthTestDB(t)
	addLegacyPasswordColumn(t, db)
	seedLegacyUser(t, db, "u1", "alice", "5f4dcc3b5aa765d61d8327deb882cf99")

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewUserLocalAuthBackfillMigrator()),
	)

	row, ok := readLocalAuth(t, db, "u1")
	require.True(t, ok, "expected user_local_auth row for u1")
	assert.Equal(t, "5f4dcc3b5aa765d61d8327deb882cf99", row.PasswordHash)
	assert.Equal(t, "md5", row.PasswordAlgo)

	var marker commonModel.KeyValue
	require.NoError(t, db.Where("key = ?", commonModel.UserLocalAuthBackfilledKey).First(&marker).Error)
}

func TestUserLocalAuthBackfillMigrator_Idempotent(t *testing.T) {
	db := newLocalAuthTestDB(t)
	addLegacyPasswordColumn(t, db)
	seedLegacyUser(t, db, "u1", "alice", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	m := dbMigration.NewUserLocalAuthBackfillMigrator()
	require.NoError(t, m.Migrate(db))

	require.NoError(t, db.Exec(
		`UPDATE user_local_auth SET password_hash = ?, password_algo = ? WHERE user_id = ?`,
		"$2a$10$bcrypthashplaceholdervalueaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bcrypt", "u1",
	).Error)

	require.NoError(t, m.Migrate(db))

	assert.Equal(t, int64(1), countLocalAuth(t, db))
	row, ok := readLocalAuth(t, db, "u1")
	require.True(t, ok)
	assert.Equal(t, "bcrypt", row.PasswordAlgo, "已升级的哈希不应被回填覆盖")
}

func TestUserLocalAuthBackfillMigrator_NoOpWhenNoPasswordColumn(t *testing.T) {
	db := newLocalAuthTestDB(t)
	require.NoError(t, db.Exec(
		`INSERT INTO users (id, username, is_admin, is_owner, locale) VALUES ('u1','alice',0,0,'zh-CN')`,
	).Error)

	require.NoError(t, dbMigration.NewUserLocalAuthBackfillMigrator().Migrate(db))
	assert.Equal(t, int64(0), countLocalAuth(t, db), "无 password 列时不应回填")
}

func TestUsersPasswordDropMigrator_DropsColumn(t *testing.T) {
	db := newLocalAuthTestDB(t)
	addLegacyPasswordColumn(t, db)
	assert.True(t, db.Migrator().HasColumn("users", "password"))

	dbMigration.Migrate(
		db,
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(dbMigration.NewUsersPasswordDropMigrator()),
	)

	assert.False(t, db.Migrator().HasColumn("users", "password"), "password 列应被删除")

	require.NoError(t, dbMigration.NewUsersPasswordDropMigrator().Migrate(db))
}
