package database

import (
	"errors"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	dbMigration "github.com/lin-snow/ech0/internal/database/migration"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commentModel "github.com/lin-snow/ech0/internal/model/comment"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	migrationModel "github.com/lin-snow/ech0/internal/model/migration"
	queueModel "github.com/lin-snow/ech0/internal/model/queue"
	settingModel "github.com/lin-snow/ech0/internal/model/setting"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	visitorModel "github.com/lin-snow/ech0/internal/model/visitor"
	webhookModel "github.com/lin-snow/ech0/internal/model/webhook"
	util "github.com/lin-snow/ech0/internal/util/err"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB 全局数据库连接变量
// var DB *gorm.DB

// 使用 atomic.Value 来存储 *gorm.DB，确保线程安全和支持热更新
var db atomic.Value // 用于存储 *gorm.DB

var writeLocked atomic.Bool

func GetDB() *gorm.DB {
	return db.Load().(*gorm.DB)
}

func SetDB(newDB *gorm.DB) {
	db.Store(newDB)
}

// func DBProvider() func() *gorm.DB {
// 	return GetDB
// }

// EnableWriteLock 启用写锁，阻止新的写操作
func EnableWriteLock() {
	writeLocked.Store(true)
}

// DisableWriteLock 关闭写锁，允许写操作
func DisableWriteLock() {
	writeLocked.Store(false)
}

// SetWriteLock 手动设置写锁状态
func SetWriteLock(enabled bool) {
	writeLocked.Store(enabled)
}

// IsWriteLocked 判断当前是否启用了写锁
func IsWriteLocked() bool {
	return writeLocked.Load()
}

func buildGormConfig(logLevel logger.LogLevel) *gorm.Config {
	return &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	}
}

// InitDatabase 初始化数据库连接
func InitDatabase() {
	// 读取数据库类型和保存路径
	dbType := config.Config().Database.Type
	dbPath := config.Config().Database.Path

	dir := dbPath[:len(dbPath)-len("/ech0.db")] // 提取目录部分
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.CREATE_DB_PATH_PANIC,
			Err: err,
		})
	}

	if dbType == "sqlite" {
		var err error
		ll := logger.LogLevel(logger.Error)
		if config.Config().Database.LogMode == "release" {
			ll = logger.LogLevel(logger.Silent)
		}
		SQLiteDB, err := gorm.Open(sqlite.Open(dbPath), buildGormConfig(ll))
		if err != nil {
			util.HandlePanicError(&commonModel.ServerError{
				Msg: commonModel.INIT_DATABASE_PANIC,
				Err: err,
			})
		}
		SetDB(SQLiteDB)
	}

	// 自动建表
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
		),
	)
}

// MigrateDB 执行数据库迁移
func MigrateDB() error {
	models := []interface{}{
		&userModel.User{},
		&userModel.UserLocalAuth{},
		&userModel.UserExternalIdentity{},
		&userModel.WebAuthnCredential{},
		&echoModel.Echo{},
		&echoModel.EchoExtension{},
		&fileModel.File{},
		&fileModel.EchoFile{},
		&fileModel.TempFile{},
		&commonModel.KeyValue{},
		&connectModel.Connected{},
		&echoModel.Tag{},
		&echoModel.EchoTag{},
		&commentModel.Comment{},
		&webhookModel.Webhook{},
		&queueModel.DeadLetter{},
		&migrationModel.MigrationJob{},
		&settingModel.AccessTokenSetting{},
		&authModel.Passkey{},
		&visitorModel.DailyStat{},
	}

	return GetDB().AutoMigrate(
		models...,
	)
}

// HotChangeDatabase 热切换数据库连接
func HotChangeDatabase(newDBPath string) error {
	// 获取当前数据库连接
	oldDB := GetDB()

	// 彻底关闭旧连接
	if oldDB != nil {
		if err := CloseDatabaseFully(oldDB); err != nil {
			return err
		}
	}

	// 打开新连接
	ll := logger.LogLevel(logger.Error)
	if config.Config().Database.LogMode == "release" {
		ll = logger.LogLevel(logger.Silent)
	}

	newDB, err := gorm.Open(sqlite.Open(newDBPath), buildGormConfig(ll))
	if err != nil {
		return err
	}

	SetDB(newDB)
	return nil
}

// CloseDatabaseFully 彻底关闭数据库连接，释放资源
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

		// 强制 GC 回收
		runtime.GC()
		time.Sleep(100 * time.Millisecond)

		return nil
	}

	return errors.New(commonModel.DATABASE_CLOSE_FAILED)
}
