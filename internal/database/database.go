// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package database

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

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
	oauthserverModel "github.com/lin-snow/ech0/internal/model/oauthserver"
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

// ensureDBDir creates the parent directory of the configured SQLite file.
// ECH0_DB_PATH is an arbitrary path — not necessarily data/ech0.db — so the
// directory must come from the path itself rather than a hardcoded suffix.
func ensureDBDir(dbPath string) error {
	if strings.TrimSpace(dbPath) == "" {
		return errors.New("ECH0_DB_PATH is empty")
	}
	dir := filepath.Dir(dbPath)
	if dir == "." || dir == string(filepath.Separator) {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}

func configLogLevel() logger.LogLevel {
	if config.Config().Database.LogMode == "release" {
		return logger.Silent
	}
	return logger.Error
}

func SnapshotTo(dstPath string) error {
	return SnapshotToContext(context.Background(), dstPath)
}

func SnapshotToContext(ctx context.Context, dstPath string) error {
	if config.Config().Database.Type == "libsql" {
		return writeRemoteSnapshot(ctx, GetDB(), dstPath)
	}
	return GetDB().WithContext(ctx).Exec("VACUUM INTO ?", dstPath).Error
}

func InitDatabase() {
	dbType := config.Config().Database.Type
	dbPath := config.Config().Database.Path

	if dbType == "libsql" {
		remote, err := OpenRemoteDatabase(config.Config().Database)
		if err != nil {
			util.HandlePanicError(&commonModel.ServerError{Msg: commonModel.INIT_DATABASE_PANIC, Err: err})
		}
		SetDB(remote)
		return
	}
	if dbType != "sqlite" {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.INIT_DATABASE_PANIC,
			Err: errors.New("unsupported ECH0_DB_TYPE " + strconv.Quote(dbType) + ` (expected "sqlite" or "libsql")`),
		})
	}

	if err := ensureDBDir(dbPath); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.CREATE_DB_PATH_PANIC,
			Err: err,
		})
	}

	if err := enableLocalSQLite(); err != nil {
		util.HandlePanicError(&commonModel.ServerError{Msg: commonModel.INIT_DATABASE_PANIC, Err: err})
	}
	SQLiteDB, err := openSQLite(dbPath, configLogLevel())
	if err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.INIT_DATABASE_PANIC,
			Err: err,
		})
	}
	SetDB(SQLiteDB)

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
			dbMigration.NewEchosFavCountDropMigrator(),
			dbMigration.NewEchosPrivateCreatedIndexDropMigrator(),
		),
	)
}

func MigrateDB() error { return migrateModels(GetDB()) }

func migrateModels(db *gorm.DB) error {
	models := []any{
		&userModel.User{},
		&userModel.UserLocalAuth{},
		&userModel.UserExternalIdentity{},
		&userModel.WebAuthnCredential{},
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&echoModel.Reaction{},
		&embeddingModel.EchoEmbedding{},
		&fileModel.File{},
		&fileModel.EchoFile{},
		&fileModel.TempFile{},
		&fileModel.UploadIntent{},
		&commonModel.KeyValue{},
		&connectModel.Connected{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&commentModel.Comment{},
		&webhookModel.Webhook{},
		&jobModel.Job{},
		&settingModel.AccessTokenSetting{},
		&authModel.Passkey{},
		&authModel.RevokedToken{},
		&authModel.Challenge{},
		&oauthserverModel.Client{},
		&oauthserverModel.ConsumedCode{},
		&visitorModel.DailyStat{},
		&visitorModel.VisitorSeen{},
	}

	return db.AutoMigrate(
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
