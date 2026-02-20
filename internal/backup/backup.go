package backup

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/lin-snow/ech0/internal/database"
	fileUtil "github.com/lin-snow/ech0/internal/util/file"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
)

const (
	dataDir        = "data"                // 待备份的数据目录
	backupDir      = "backup"              // 备份后存储zip的目录
	backupFileName = "ech0_backup"         // 备份文件名
	excludeFile    = "*.log"               // 排除的文件名
	timeLayout     = "2006-01-02_15-04-05" // 时间格式化布局
)

// ExecuteBackup 执行备份
func ExecuteBackup() (string, string, error) {
	backupTime := time.Now().UTC().Format(timeLayout)
	backupFileName := fmt.Sprintf("%s_%s.zip", backupFileName, backupTime) // 暂时不开启多备份，每次只保留最新的一份备份
	backupPath := fmt.Sprintf("%s/%s", backupDir, backupFileName)

	return backupPath, backupFileName, fileUtil.ZipDirectoryWithOptions(
		dataDir,
		backupPath,
		fileUtil.ZipOptions{
			ExcludePatterns: []string{excludeFile},
		},
	)
}

// ExecuteRestore 执行恢复
func ExecuteRestore(backupFilePath string) error {
	// 检查备份文件是否存在
	if !fileUtil.FileExists(backupFilePath) {
		return errors.New("备份文件不存在: " + backupFilePath)
	}

	previousLock := database.IsWriteLocked()
	if !previousLock {
		database.EnableWriteLock()
		defer database.DisableWriteLock()
	}

	logUtil.CloseLogger()
	defer logUtil.ReopenLogger()

	// 解压备份文件到数据目录
	if err := fileUtil.UnzipFile(backupFilePath, dataDir); err != nil {
		return err
	}

	return nil
}

// ExcuteRestoreOnline 在线恢复备份
func ExcuteRestoreOnline(filePath string, timeStamp int64) error {
	// 检查备份文件是否存在
	if !fileUtil.FileExists(filePath) {
		return errors.New("备份文件不存在: " + filePath)
	}

	// 启用写锁，阻止新的写操作
	previousLock := database.IsWriteLocked()
	if !previousLock {
		database.EnableWriteLock()
		defer database.DisableWriteLock()
	}

	// 关闭 Logger，释放文件句柄
	logUtil.CloseLogger()
	defer logUtil.ReopenLogger()

	// 解压备份文件到数据目录 （./temp/snapshot_时间戳）
	extractPath := fmt.Sprintf("temp/snapshot_%d", timeStamp)
	if err := fileUtil.UnzipFile(filePath, extractPath); err != nil {
		return err
	}

	tempDbPath := filepath.Join(extractPath, "ech0.db")

	// 热切换到临时数据库
	if err := database.HotChangeDatabase(tempDbPath); err != nil {
		return err
	}

	// 复制备份覆盖到正式数据目录
	dataPath := "data"
	if err := fileUtil.CopyDirectory(extractPath, dataPath); err != nil {
		return err
	}

	// 热切换回正式数据库
	if err := database.HotChangeDatabase("data/ech0.db"); err != nil {
		return err
	}

	return nil
}
