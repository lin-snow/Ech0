// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"context"
	"database/sql"
	"errors"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/pkg/virefs"
	"gorm.io/gorm"
)

func markCleanup(tx *gorm.DB, run, upload, name string) error {
	return tx.Exec("INSERT INTO ech0_export_gc(run,round,upload,name) VALUES(?,0,?,?) ON CONFLICT(run) DO NOTHING", run, upload, name).Error
}

func recordFailure(db *gorm.DB, w Wake, lease string, c checkpoint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var failures int
		err := tx.Raw("UPDATE ech0_export_state SET lease='',expires=0,failures=failures+1 WHERE id=? AND rev=? AND lease=? AND active=1 RETURNING failures", w.Run, w.Revision, lease).Row().Scan(&failures)
		if err != nil || failures < 8 {
			return err
		}
		if err := tx.Exec("UPDATE ech0_export_state SET active=0 WHERE id=?", w.Run).Error; err != nil {
			return err
		}
		if err := markCleanup(tx, w.Run, c.Upload, c.Name); err != nil {
			return err
		}
		return tx.Exec("UPDATE jobs SET status='failed',error='Snapshot checkpoint failed repeatedly; check database and object storage availability and quota',finished_at=unixepoch(),updated_at=unixepoch() WHERE type='export'").Error
	})
}

func PendingCleanup(ctx context.Context, db *gorm.DB) (*Wake, error) {
	db = db.WithContext(ctx)
	if !db.Migrator().HasTable("ech0_export_gc") {
		return nil, nil
	}
	var w Wake
	err := db.Raw("SELECT run,round FROM ech0_export_gc ORDER BY run LIMIT 1").Row().Scan(&w.Cleanup, &w.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

func CleanupExport(ctx context.Context, db *gorm.DB, cfg config.StorageConfig, w Wake) (*Wake, error) {
	db = db.WithContext(ctx)
	var upload, name string
	var round int64
	err := db.Raw("SELECT round,upload,name FROM ech0_export_gc WHERE run=?", w.Cleanup).Row().Scan(&round, &upload, &name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if round != w.Revision {
		return &Wake{Cleanup: w.Cleanup, Revision: round}, nil
	}
	if db.Migrator().HasTable("ech0_history_state") {
		if err := (database.SnapshotHistory{ID: w.Cleanup}).Finish(ctx, db); err != nil {
			return nil, err
		}
	}
	store, err := virefs.NewS3Client(ctx, virefsS3Config(cfg))
	if err != nil {
		return nil, err
	}
	if upload != "" {
		_, err := store.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{Bucket: &cfg.BucketName, Key: aws.String(path.Join(cfg.PathPrefix, s3SnapshotPrefix, name)), UploadId: &upload})
		var apiError smithy.APIError
		if err != nil && !(errors.As(err, &apiError) && apiError.ErrorCode() == "NoSuchUpload") {
			return nil, err
		}
	}
	prefix := path.Join(cfg.PathPrefix, "ech0-staging", w.Cleanup) + "/"
	listed, err := store.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &cfg.BucketName, Prefix: &prefix, MaxKeys: aws.Int32(100)})
	if err != nil {
		return nil, err
	}
	if len(listed.Contents) > 0 {
		var keys []types.ObjectIdentifier
		for _, obj := range listed.Contents {
			if obj.Key == nil || !strings.HasPrefix(*obj.Key, prefix) {
				return nil, errors.New("unexpected staging object prefix")
			}
			keys = append(keys, types.ObjectIdentifier{Key: obj.Key})
		}
		result, err := store.DeleteObjects(ctx, &s3.DeleteObjectsInput{Bucket: &cfg.BucketName, Delete: &types.Delete{Objects: keys}})
		if err != nil {
			return nil, err
		}
		if len(result.Errors) > 0 {
			return nil, errors.New("snapshot staging cleanup incomplete")
		}
		if err := db.Exec("UPDATE ech0_export_gc SET round=round+1,upload='' WHERE run=? AND round=?", w.Cleanup, w.Revision).Error; err != nil {
			return nil, err
		}
		return &Wake{Cleanup: w.Cleanup, Revision: w.Revision + 1}, nil
	}
	if err := db.Exec("DELETE FROM ech0_export_blocks WHERE run=?", w.Cleanup).Error; err != nil {
		return nil, err
	}
	fs, err := BuildS3FS(cfg)
	if err != nil {
		return nil, err
	}
	if err := cleanupOldS3Snapshots(ctx, fs, s3SnapshotKeepCount); err != nil {
		return nil, err
	}
	return nil, db.Exec("DELETE FROM ech0_export_gc WHERE run=? AND round=?", w.Cleanup, w.Revision).Error
}

// Renewing only a live generation prevents a redelivered worker from reviving
// an expired snapshot. Long-running exports can make progress across many days.
func renewHistory(db *gorm.DB, id string) error {
	result := db.Exec("UPDATE ech0_history_state SET expires=? WHERE id=? AND expires>unixepoch()", time.Now().Add(24*time.Hour).Unix(), id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("snapshot history lease expired")
	}
	return nil
}
