package database

import (
	"errors"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
)

// fixOldEchoLayoutData 为旧数据补充默认的布局值（layout 为 NULL 或空字符串时设为 'waterfall'）
func fixOldEchoLayoutData() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	// 更新所有 layout 为 NULL 或空字符串的 echo 记录为 'waterfall'
	if err := db.Model(&echoModel.Echo{}).
		Where("layout IS NULL OR layout = ''").
		Update("layout", "waterfall").Error; err != nil {
		return err
	}

	return nil
}

// MigrateImageToMedia 将images表迁移到media表
func MigrateImageToMedia() error {
	db := GetDB()
	if db == nil {
		return errors.New(commonModel.DATABASE_NOT_INITED)
	}

	// 检查是否需要迁移（images表是否存在）
	if !db.Migrator().HasTable("images") {
		// images表不存在，说明是新安装或已经迁移过，无需处理
		return nil
	}

	// images表存在，需要迁移数据到media表
	// 注意：media表已经由MigrateDB()的AutoMigrate创建

	// 复制数据，将所有现有图片的media_type设置为'image'
	if err := db.Exec(`
		INSERT INTO media (id, message_id, media_url, media_type, media_source, object_key, width, height)
		SELECT id, message_id, image_url, 'image', image_source, object_key, width, height
		FROM images
	`).Error; err != nil {
		return err
	}

	// 删除旧表
	if err := db.Migrator().DropTable("images"); err != nil {
		return err
	}

	return nil
}

// UpdateMigration 执行旧数据库迁移和数据修复任务
func UpdateMigration() error {
	var err error

	err = fixOldEchoLayoutData()
	if err != nil {
		return err
	}

	err = MigrateImageToMedia()
	if err != nil {
		return err
	}

	return nil
}
