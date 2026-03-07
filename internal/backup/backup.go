package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	virefs "github.com/lin-snow/VireFS"
	vizip "github.com/lin-snow/VireFS/plugin/zip"
	"github.com/lin-snow/ech0/internal/database"
	logUtil "github.com/lin-snow/ech0/internal/util/log"
	"go.uber.org/zap"
)

const (
	dataDir        = "data"
	backupDir      = "backup"
	backupFileName = "ech0_backup"
	excludePattern = ".log"
	timeLayout     = "2006-01-02_15-04-05"
)

// ExecuteBackup packs the data/ directory into a zip archive using VireFS.
func ExecuteBackup() (string, string, error) {
	backupTime := time.Now().UTC().Format(timeLayout)
	fileName := fmt.Sprintf("%s_%s.zip", backupFileName, backupTime)
	backupPath := filepath.Join(backupDir, fileName)

	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create backup dir: %w", err)
	}

	dataFS, err := virefs.NewLocalFS(dataDir)
	if err != nil {
		return "", "", fmt.Errorf("open data dir: %w", err)
	}

	ctx := context.Background()
	var keys []string
	if err := virefs.Walk(ctx, dataFS, "", func(key string, info virefs.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir {
			return nil
		}
		if strings.HasSuffix(key, excludePattern) {
			return nil
		}
		keys = append(keys, key)
		return nil
	}); err != nil {
		return "", "", fmt.Errorf("walk data dir: %w", err)
	}

	f, err := os.Create(backupPath)
	if err != nil {
		return "", "", fmt.Errorf("create zip file: %w", err)
	}

	if err := vizip.Pack(ctx, dataFS, keys, f); err != nil {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close backup zip after pack error",
				zap.String("path", backupPath), zap.String("error", closeErr.Error()))
		}
		return "", "", fmt.Errorf("pack zip: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", "", fmt.Errorf("close zip file: %w", err)
	}

	return backupPath, fileName, nil
}

// ExecuteRestore unpacks a backup zip into the data directory.
func ExecuteRestore(backupFilePath string) error {
	if _, err := os.Stat(backupFilePath); err != nil {
		return errors.New("备份文件不存在: " + backupFilePath)
	}

	previousLock := database.IsWriteLocked()
	if !previousLock {
		database.EnableWriteLock()
		defer database.DisableWriteLock()
	}

	logUtil.CloseLogger()
	defer logUtil.ReopenLogger()

	return unpackZipToDir(backupFilePath, dataDir)
}

// ExcuteRestoreOnline performs an online restore from an uploaded zip.
func ExcuteRestoreOnline(filePath string, timeStamp int64) error {
	if _, err := os.Stat(filePath); err != nil {
		return errors.New("备份文件不存在: " + filePath)
	}

	previousLock := database.IsWriteLocked()
	if !previousLock {
		database.EnableWriteLock()
		defer database.DisableWriteLock()
	}

	logUtil.CloseLogger()
	defer logUtil.ReopenLogger()

	extractPath := fmt.Sprintf("temp/snapshot_%d", timeStamp)
	defer func() {
		if err := os.RemoveAll(extractPath); err != nil {
			logUtil.GetLogger().Warn("Failed to cleanup extracted snapshot temp directory",
				zap.String("path", extractPath), zap.String("error", err.Error()))
		}
		if err := os.Remove(filePath); err != nil {
			logUtil.GetLogger().Warn("Failed to cleanup uploaded snapshot zip",
				zap.String("path", filePath), zap.String("error", err.Error()))
		}
	}()

	if err := unpackZipToDir(filePath, extractPath); err != nil {
		return err
	}

	tempDbPath := filepath.Join(extractPath, "ech0.db")
	if err := database.HotChangeDatabase(tempDbPath); err != nil {
		return err
	}

	if err := copyDirViaVireFS(extractPath, dataDir); err != nil {
		return err
	}

	if err := database.HotChangeDatabase("data/ech0.db"); err != nil {
		return err
	}

	return nil
}

func unpackZipToDir(zipPath, destDir string) error {
	f, err := os.Open(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			logUtil.GetLogger().Warn("Failed to close backup zip reader",
				zap.String("path", zipPath), zap.String("error", closeErr.Error()))
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat zip: %w", err)
	}

	dstFS, err := virefs.NewLocalFS(destDir, virefs.WithCreateRoot())
	if err != nil {
		return fmt.Errorf("open dest dir: %w", err)
	}

	return vizip.Unpack(context.Background(), f, info.Size(), dstFS, "")
}

func copyDirViaVireFS(srcDir, dstDir string) error {
	srcFS, err := virefs.NewLocalFS(srcDir)
	if err != nil {
		return fmt.Errorf("open src dir: %w", err)
	}
	dstFS, err := virefs.NewLocalFS(dstDir, virefs.WithCreateRoot())
	if err != nil {
		return fmt.Errorf("open dst dir: %w", err)
	}

	_, err = virefs.Migrate(context.Background(), srcFS, "", dstFS, "",
		virefs.WithConflictPolicy(virefs.ConflictOverwrite),
	)
	return err
}
