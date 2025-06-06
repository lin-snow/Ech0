package database

import (
	"os"

	"github.com/lin-snow/ech0/internal/config"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	connectModel "github.com/lin-snow/ech0/internal/model/connect"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	todoModel "github.com/lin-snow/ech0/internal/model/todo"
	userModel "github.com/lin-snow/ech0/internal/model/user"

	util "github.com/lin-snow/ech0/internal/util/err"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	// 读取数据库类型和保存路径
	dbType := config.Config.Database.Type
	dbPath := config.Config.Database.Path

	dir := dbPath[:len(dbPath)-len("/ech0.db")] // 提取目录部分
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.CREATE_DB_PATH_PANIC,
			Err: err,
		})
	}

	if dbType == "sqlite" {
		var err error
		DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			util.HandlePanicError(&commonModel.ServerError{
				Msg: commonModel.INIT_DATABASE_PANIC,
				Err: err,
			})
		}
	}

	if err := MigrateDB(); err != nil {
		util.HandlePanicError(&commonModel.ServerError{
			Msg: commonModel.MIGRATE_DB_PANIC,
			Err: err,
		})
	}
}

func MigrateDB() error {
	models := []interface{}{
		&userModel.User{},
		&echoModel.Echo{},
		&echoModel.Image{},
		&commonModel.KeyValue{},
		&todoModel.Todo{},
		&connectModel.Connected{},
	}

	return DB.AutoMigrate(
		models...,
	)
}
