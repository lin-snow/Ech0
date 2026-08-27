// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"errors"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/lin-snow/ech0/internal/config"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	embeddingModel "github.com/lin-snow/ech0/internal/model/embedding"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	util "github.com/lin-snow/ech0/internal/util/err"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db atomic.Value

var writeLocked atomic.Bool

func GetDB() *gorm.DB {
	return db.Load().(*gorm.DB)
}

func SetDB(newDB *gorm.DB) {
	db.Store(newDB)
}

func EnableWriteLock() {
	writeLocked.Store(true)
}

func DisableWriteLock() {
	writeLocked.Store(false)
}

func SetWriteLock(enabled bool) {
	writeLocked.Store(enabled)
}

func IsWriteLocked() bool {
	return writeLocked.Load()
}

func buildGormConfig(logLevel logger.LogLevel) *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}
}

const sqliteConnParams = "_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_txlock=immediate"

func openSQLite(dbPath string, logLevel logger.LogLevel) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath+"?"+sqliteConnParams), buildGormConfig(logLevel))
}

func configLogLevel() logger.LogLevel {
	if config.Config().Database.LogMode == "release" {
		return logger.Silent
	}
	return logger.Error
}

func SnapshotTo(dstPath string) error {
	return GetDB().Exec("VACUUM INTO ?", dstPath).Error
}

func InitDatabase() {
	dbType := config.Config().Database.Type
	dbPath := config.Config().Database.Path

	dir := dbPath[:len(dbPath)-len("/ech0.db")]
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.CREATE_DB_PATH_PANIC,
			Err: err,
		})
	}

	if dbType == "sqlite" {
		sqlite_vec.Auto()
		SQLiteDB, err := openSQLite(dbPath, configLogLevel())
		if err != nil {
			util.HandlePanicError(&commonModel.ServerError{
				Msg: commonModel.INIT_DATABASE_PANIC,
				Err: err,
			})
		}
		SetDB(SQLiteDB)
	}

	if err := MigrateDB(); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.MIGRATE_DB_PANIC,
			Err: err,
		})
	}

	dbMigration.Migrate(
		GetDB(),
		dbMigration.WithStopOnError(),
		dbMigration.WithMigrators(
			dbMigration.NewLegacyTimeNormalizerMigrator(dbMigration.DefaultLegacySourceTimezone),
			dbMigration.NewStorageTimeSanitizeMigrator(),
			dbMigration.NewStorageTimeValidateMigrator(),
			dbMigration.NewStorageTimeUnixMigrator(),
			dbMigration.NewStorageTimeSchemaRebuildMigrator(),
			dbMigration.NewOAuthBindingsDropMigrator(),
			dbMigration.NewLegacyInboxesDropMigrator(),
			dbMigration.NewAgentProtocolCollapseMigrator(),
			dbMigration.NewAgentSettingProtocolRenameMigrator(),
			dbMigration.NewUserLocalAuthBackfillMigrator(),
			dbMigration.NewUsersPasswordDropMigrator(),
			dbMigration.NewEchoExtensionOrphansMigrator(),
		),
	)
}

func MigrateDB() error {
	models := []any{
		&userModel.User{},
		&userModel.UserLocalAuth{},
		&userModel.UserExternalIdentity{},
		&userModel.WebAuthnCredential{},
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&embeddingModel.EchoEmbedding{},
		&fileModel.File{},
		&fileModel.EchoFile{},
		&fileModel.TempFile{},
		&commonModel.KeyValue{},
		&connectModel.Connected{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&commentModel.Comment{},
		&webhookModel.Webhook{},
		&jobModel.Job{},
		&settingModel.AccessTokenSetting{},
		&authModel.Passkey{},
		&visitorModel.DailyStat{},
	}

	return GetDB().AutoMigrate(
		models...,
	)
}

func HotChangeDatabase(newDBPath string) error {
	oldDB := GetDB()

	if oldDB != nil {
		if err := CloseDatabaseFully(oldDB); err != nil {
			return err
		}
	}

	newDB, err := openSQLite(newDBPath, configLogLevel())
	if err != nil {
		return err
	}

	SetDB(newDB)
	return nil
}

func CloseDatabaseFully(db *gorm.DB) error {
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		if err := sqlDB.Close(); err != nil {
			return err
		}
		SetDB(nil)

		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		return nil
	}

	return errors.New(commonModel.DATABASE_CLOSE_FAILED)
}
