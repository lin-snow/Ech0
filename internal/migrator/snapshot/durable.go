// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package snapshot

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/internal/database"
	"github.com/lin-snow/ech0/internal/database/pagefile"
	"github.com/lin-snow/ech0/internal/job"
	jobModel "github.com/lin-snow/ech0/internal/model/job"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"github.com/lin-snow/ech0/pkg/virefs"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Wake struct {
	Cleanup     string `json:"cleanup,omitempty"`
	SnapshotDue int64  `json:"snapshot_due,omitempty"`
	Schedule    string `json:"schedule,omitempty"`
	Run         string `json:"run,omitempty"`
	Revision    int64  `json:"revision"`
}

type checkpoint struct {
	Cursor    database.ImageCursor
	Size      int64
	Name      string
	Upload    string
	Offset    int64
	CRC       uint32
	Parts     []string
	Written   int64
	Scheduled bool
	Started   bool
}

func setupExports(db *gorm.DB) error {
	for _, q := range []string{
		"CREATE TABLE IF NOT EXISTS ech0_export_state(slot INTEGER PRIMARY KEY CHECK(slot=1),id TEXT NOT NULL,rev INTEGER NOT NULL,state TEXT NOT NULL,active INTEGER NOT NULL,lease TEXT NOT NULL,expires INTEGER NOT NULL,failures INTEGER NOT NULL)",
		"CREATE TABLE IF NOT EXISTS ech0_export_blocks(run TEXT NOT NULL,n INTEGER NOT NULL,ref TEXT NOT NULL,PRIMARY KEY(run,n))",
		"CREATE TABLE IF NOT EXISTS ech0_export_gc(run TEXT PRIMARY KEY,round INTEGER NOT NULL,upload TEXT NOT NULL,name TEXT NOT NULL)",
		"CREATE TABLE IF NOT EXISTS ech0_export_occurrences(id TEXT PRIMARY KEY,due INTEGER NOT NULL)",
		"CREATE TABLE IF NOT EXISTS ech0_export_latest(slot INTEGER PRIMARY KEY CHECK(slot=1),name TEXT NOT NULL)",
	} {
		if err := db.Exec(q).Error; err != nil {
			return err
		}
	}
	return nil
}

func StartExport(ctx context.Context, db *gorm.DB, scheduled bool) (Wake, jobModel.Job, error) {
	id := uuidUtil.NewV7()
	now := time.Now().UTC()
	state := checkpoint{Scheduled: scheduled, Name: "ech0_snapshot_" + now.Format("2006-01-02_15-04-05") + "_" + id + ".zip"}
	encoded, err := json.Marshal(state)
	if err != nil {
		return Wake{}, jobModel.Job{}, err
	}
	started := now.Unix()
	item := jobModel.Job{Type: "export", Status: jobModel.StatusPending, Phase: "packing", StartedAt: &started, Payload: `{"format":"snapshot"}`}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := setupExports(tx); err != nil {
			return err
		}
		reserve := tx.Exec(`INSERT INTO jobs(type,status,phase,error,payload,started_at,finished_at,updated_at)
            VALUES('export','pending','packing','',?,?,NULL,?) ON CONFLICT(type) DO UPDATE SET status=excluded.status,phase=excluded.phase,error='',payload=excluded.payload,started_at=excluded.started_at,finished_at=NULL,updated_at=excluded.updated_at WHERE jobs.status NOT IN ('pending','running')`, item.Payload, started, started)
		if reserve.Error != nil {
			return reserve.Error
		}
		if reserve.RowsAffected == 0 {
			return job.ErrAlreadyRunning
		}
		return tx.Exec(`INSERT INTO ech0_export_state VALUES(1,?,0,?,1,'',0,0)
            ON CONFLICT(slot) DO UPDATE SET id=excluded.id,rev=0,state=excluded.state,active=1,lease='',expires=0,failures=0`, id, string(encoded)).Error
	})
	return Wake{Run: id}, item, err
}

func PendingExport(ctx context.Context, db *gorm.DB) (*Wake, error) {
	if !db.WithContext(ctx).Migrator().HasTable("ech0_export_state") {
		return nil, nil
	}
	var w Wake
	err := db.WithContext(ctx).Raw("SELECT id,rev FROM ech0_export_state WHERE active=1").Row().Scan(&w.Run, &w.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &w, err
}

var ErrBusy = errors.New("snapshot worker already has the lease")

// AdvanceExport owns at most one checkpoint. A stopped function cannot publish
// with an expired fencing token once another worker has reclaimed the job.
func AdvanceExport(ctx context.Context, db *gorm.DB, cfg config.StorageConfig, w Wake, completed func(context.Context, *gorm.DB, bool) error) (*Wake, error) {
	db = db.WithContext(ctx)
	current, err := PendingExport(ctx, db)
	if err != nil || current == nil || current.Run != w.Run {
		return nil, err
	}
	if current.Revision != w.Revision {
		return current, nil
	}
	lease := uuidUtil.NewV7()
	var raw string
	err = db.Raw("UPDATE ech0_export_state SET lease=?,expires=unixepoch()+90 WHERE id=? AND rev=? AND active=1 AND expires<=unixepoch() RETURNING state", lease, w.Run, w.Revision).Row().Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBusy
	}
	if err != nil {
		return nil, err
	}
	var c checkpoint
	err = json.Unmarshal([]byte(raw), &c)
	fs, fsErr := BuildS3FS(cfg)
	if err == nil {
		err = fsErr
	}
	refs := map[int64]string{}
	unused := []string{}
	done := false
	history := database.SnapshotHistory{ID: w.Run}
	if err == nil {
		switch {
		case !c.Started:
			err = history.Begin(ctx, db)
			c.Started = err == nil
		case !c.Cursor.Ready:
			err = c.copyBlocks(ctx, db, history, fs, w.Revision, refs)
		default:
			if err = history.Finish(ctx, db); err == nil {
				done, err = c.uploadPart(ctx, db, cfg, fs, w.Run)
			}
		}
	}
	if err != nil {
		if ctx.Err() == nil {
			_ = recordFailure(db, w, lease, c)

		}
		return nil, err
	}
	encoded, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		r := tx.Exec("UPDATE ech0_export_state SET rev=rev+1,state=?,active=?,lease='',expires=0,failures=0 WHERE id=? AND rev=? AND lease=? AND active=1", string(encoded), !done, w.Run, w.Revision, lease)
		if r.Error != nil {
			return r.Error
		}
		if r.RowsAffected != 1 {
			return ErrBusy
		}
		if c.Started && !c.Cursor.Ready {
			if err := renewHistory(tx, w.Run); err != nil {
				return err
			}
		}
		if len(refs) > 0 {
			type blockRef struct {
				Run string
				N   int64
				Ref string
			}
			numbers := make([]int64, 0, len(refs))
			rows := make([]blockRef, 0, len(refs))
			for n, ref := range refs {
				numbers = append(numbers, n)
				rows = append(rows, blockRef{w.Run, n, ref})
			}
			if err := tx.Table("ech0_export_blocks").Where("run=? AND n IN ?", w.Run, numbers).Pluck("ref", &unused).Error; err != nil {
				return err
			}
			if err := tx.Table("ech0_export_blocks").Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "run"}, {Name: "n"}}, DoUpdates: clause.AssignmentColumns([]string{"ref"})}).CreateInBatches(&rows, 120).Error; err != nil {
				return err
			}
		}
		if done {
			if err := completed(ctx, tx, c.Scheduled); err != nil {
				return err
			}
			if err := markCleanup(tx, w.Run, "", c.Name); err != nil {
				return err
			}
			if err := tx.Exec("INSERT INTO ech0_export_latest VALUES(1,?) ON CONFLICT(slot) DO UPDATE SET name=excluded.name", c.Name).Error; err != nil {
				return err
			}
			payload, err := json.Marshal(map[string]any{"format": "snapshot", "file_name": c.Name, "size": c.Written})
			if err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM ech0_export_blocks WHERE run=?", w.Run).Error; err != nil {
				return err
			}
			return tx.Exec("UPDATE jobs SET status='success',phase='completed',payload=?,finished_at=unixepoch(),updated_at=unixepoch() WHERE type='export'", string(payload)).Error
		}
		return tx.Exec("UPDATE jobs SET status='running',phase='packing',updated_at=unixepoch() WHERE type='export'").Error
	})
	if err != nil {
		return nil, err
	}
	for _, ref := range unused {
		_ = fs.Delete(ctx, ref)
	}
	if done {
		return &Wake{Cleanup: w.Run}, nil
	}
	return &Wake{Run: w.Run, Revision: w.Revision + 1}, nil
}

func blockReader(ctx context.Context, db *gorm.DB, fs virefs.FS, run string) func(int64) ([]byte, error) {
	return func(n int64) ([]byte, error) {
		var ref string
		if err := db.Raw("SELECT ref FROM ech0_export_blocks WHERE run=? AND n=?", run, n).Row().Scan(&ref); err != nil {
			return nil, err
		}
		return fetchBlock(ctx, fs, ref)
	}
}

func fetchBlock(ctx context.Context, fs virefs.FS, ref string) ([]byte, error) {
	obj, err := fs.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	defer func() { _ = obj.Close() }()
	data, err := io.ReadAll(io.LimitReader(obj, pagefile.BlockSize+1))
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(data)
	if len(data) != pagefile.BlockSize || !strings.HasSuffix(ref, "-"+hex.EncodeToString(hash[:])) {
		return nil, errors.New("snapshot block integrity check failed")
	}
	return data, nil
}

// The manifest is stable while this worker owns the checkpoint. Fetch all
// references in one indexed query and bound independent R2 reads to eight.
func readBlockRange(ctx context.Context, db *gorm.DB, fs virefs.FS, run string, offset, count int64) ([]byte, error) {
	first, last := offset/pagefile.BlockSize, (offset+count-1)/pagefile.BlockSize
	var rows []struct {
		N   int64
		Ref string
	}
	if err := db.WithContext(ctx).Table("ech0_export_blocks").Select("n,ref").Where("run=? AND n BETWEEN ? AND ?", run, first, last).Order("n").Scan(&rows).Error; err != nil {
		return nil, err
	}
	if int64(len(rows)) != last-first+1 {
		return nil, errors.New("incomplete snapshot block range")
	}
	result := make([]byte, count)
	group, groupCtx := errgroup.WithContext(ctx)
	group.SetLimit(8)
	for _, row := range rows {
		group.Go(func() error {
			data, err := fetchBlock(groupCtx, fs, row.Ref)
			if err != nil {
				return err
			}
			blockStart := row.N * pagefile.BlockSize
			start, end := max(offset, blockStart), min(offset+count, blockStart+pagefile.BlockSize)
			copy(result[start-offset:end-offset], data[start-blockStart:end-blockStart])
			return nil
		})
	}
	return result, group.Wait()
}

func (c *checkpoint) copyBlocks(ctx context.Context, db *gorm.DB, h database.SnapshotHistory, fs virefs.FS, revision int64, refs map[int64]string) error {
	file := pagefile.New(c.Size, blockReader(ctx, db, fs, h.ID))
	cursor, err := h.CopyImagePage(ctx, db, file, c.Cursor)
	if err != nil {
		return err
	}
	for n, data := range file.Dirty {
		hash := sha256.Sum256(data)
		key := fmt.Sprintf("ech0-staging/%s/%d-%d-%x", h.ID, n, revision, hash)
		if err := fs.Put(ctx, key, bytes.NewReader(data)); err != nil {
			return err
		}
		refs[n] = key
	}
	c.Size, c.Cursor = file.Size, cursor
	return nil
}

const minPart = 8 << 20

func (c *checkpoint) uploadPart(ctx context.Context, db *gorm.DB, cfg config.StorageConfig, fs virefs.FS, run string) (bool, error) {
	client, err := virefs.NewS3Client(ctx, virefsS3Config(cfg))
	if err != nil {
		return false, err
	}
	key := aws.String(path.Join(cfg.PathPrefix, s3SnapshotPrefix, c.Name))
	bucket := aws.String(cfg.BucketName)
	// This separate checkpoint records the S3 upload ID before writing parts.
	if c.Upload == "" {
		out, err := client.CreateMultipartUpload(ctx, &s3.CreateMultipartUploadInput{Bucket: bucket, Key: key, ContentType: aws.String("application/zip"), ContentDisposition: aws.String("attachment; filename=\"" + c.Name + "\"")})
		if err != nil {
			return false, err
		}
		c.Upload = aws.ToString(out.UploadId)
		return false, nil
	}
	size := uint64(c.Size)
	header := snapshotHeader(size)
	total := int64(len(header)+len(snapshotFooter(size, 0))) + c.Size
	if c.Written == total {
		parts := make([]types.CompletedPart, len(c.Parts))
		for i, etag := range c.Parts {
			parts[i] = types.CompletedPart{PartNumber: aws.Int32(int32(i + 1)), ETag: aws.String(etag)}
		}
		_, err = client.CompleteMultipartUpload(ctx, &s3.CompleteMultipartUploadInput{Bucket: bucket, Key: key, UploadId: &c.Upload, MultipartUpload: &types.CompletedMultipartUpload{Parts: parts}})
		if err != nil {
			// Completion may have reached the store before a worker died.
			stat, headErr := fs.Stat(ctx, s3SnapshotPrefix+c.Name)
			if headErr != nil || stat.Size != total {
				return false, err
			}
		}
		return true, nil
	}
	// Choose equal-sized non-final parts, including the ZIP header and tail.
	partSize := max(int64(minPart), (total+9998)/9999)
	var part bytes.Buffer
	if c.Written == 0 {
		part.Write(header)
	}
	remaining := min(c.Size-c.Offset, partSize-int64(part.Len()))
	if remaining > 0 {
		data, err := readBlockRange(ctx, db, fs, run, c.Offset, remaining)
		if err != nil {
			return false, err
		}
		c.CRC = crc32.Update(c.CRC, crc32.IEEETable, data)
		part.Write(data)
		c.Offset += int64(len(data))
	}
	if c.Offset == c.Size {
		footer := snapshotFooter(size, c.CRC)
		used := max(int64(0), c.Written+int64(part.Len())-int64(len(header))-c.Size)
		count := min(int64(len(footer))-used, partSize-int64(part.Len()))
		part.Write(footer[used : used+count])
	}
	number := int32(len(c.Parts) + 1)
	out, err := client.UploadPart(ctx, &s3.UploadPartInput{Bucket: bucket, Key: key, UploadId: &c.Upload, PartNumber: &number, Body: bytes.NewReader(part.Bytes()), ContentLength: aws.Int64(int64(part.Len()))})
	if err != nil {
		return false, err
	}
	if out.ETag == nil {
		return false, fmt.Errorf("missing ETag for snapshot part %d", number)
	}
	c.Parts = append(c.Parts, *out.ETag)
	c.Written += int64(part.Len())
	return false, nil
}

func PublishedDownload(ctx context.Context, db *gorm.DB, cfg config.StorageConfig) (Download, error) {
	var hasPointer bool
	if err := db.WithContext(ctx).Raw("SELECT EXISTS(SELECT 1 FROM sqlite_schema WHERE type='table' AND name='ech0_export_latest')").Row().Scan(&hasPointer); err != nil {
		return Download{}, err
	}
	if !hasPointer {
		return LatestDownload(ctx, cfg)
	}
	var name string
	if err := db.WithContext(ctx).Raw("SELECT name FROM ech0_export_latest WHERE slot=1").Row().Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Download{}, ErrNoSnapshot
		}
		return Download{}, err
	}
	fs, err := BuildS3FS(cfg)
	if err != nil {
		return Download{}, err
	}
	key := s3SnapshotPrefix + name
	info, err := fs.Stat(ctx, key)
	if err != nil {
		return Download{}, err
	}
	return signDownload(ctx, fs, key, info.Size)
}

func CancelDurableExport(ctx context.Context, db *gorm.DB, cfg config.StorageConfig) (bool, error) {
	w, err := PendingExport(ctx, db)
	if err != nil || w == nil {
		return false, err
	}
	var raw string
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Raw("UPDATE ech0_export_state SET active=0,rev=rev+1,lease='',expires=0 WHERE id=? AND active=1 RETURNING state", w.Run).Row().Scan(&raw); err != nil {
			return err
		}
		var current checkpoint
		if err := json.Unmarshal([]byte(raw), &current); err != nil {
			return err
		}
		if err := markCleanup(tx, w.Run, current.Upload, current.Name); err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM ech0_export_blocks WHERE run=?", w.Run).Error; err != nil {
			return err
		}
		return tx.Exec("UPDATE jobs SET status='cancelled',finished_at=unixepoch(),updated_at=unixepoch() WHERE type='export'").Error
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
