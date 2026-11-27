// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package repository

import (
	"context"
	"strings"

	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FileRepository is the persistence of the five file-system tables. It holds
// no policy: every authorization decision is made by the service before a
// method here is called.
type FileRepository struct {
	db func() *gorm.DB
}

func NewFileRepository(dbProvider func() *gorm.DB) *FileRepository {
	return &FileRepository{db: dbProvider}
}

func (r *FileRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

// ---- files -------------------------------------------------------------

func (r *FileRepository) CreateFile(ctx context.Context, file *fileModel.File) error {
	return r.getDB(ctx).Create(file).Error
}

func (r *FileRepository) GetFile(ctx context.Context, id string) (*fileModel.File, error) {
	var f fileModel.File
	if err := r.getDB(ctx).Where("id = ?", id).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

// UpdateName renames a file. It is the only mutable column.
func (r *FileRepository) UpdateName(ctx context.Context, id, name string) error {
	return r.getDB(ctx).Model(&fileModel.File{}).
		Where("id = ?", id).
		Update("name", name).Error
}

// DeleteFile removes the row; links and blob bytes go with it through foreign
// keys. Callers that must enqueue a GC task do so in the same transaction.
func (r *FileRepository) DeleteFile(ctx context.Context, id string) (int64, error) {
	res := r.getDB(ctx).Where("id = ?", id).Delete(&fileModel.File{})
	return res.RowsAffected, res.Error
}

// DeletePendingIfStill deletes a pending row only if it is still pending: a
// commit may have turned it ready between the sweep's SELECT and DELETE.
func (r *FileRepository) DeletePendingIfStill(ctx context.Context, id string) (int64, error) {
	res := r.getDB(ctx).
		Where("id = ? AND status = ?", id, fileModel.StatusPending).
		Delete(&fileModel.File{})
	return res.RowsAffected, res.Error
}

// MarkReady promotes a pending row. It returns 0 rows when a commit or a
// sweep already moved it, which is the caller's signal to stop.
func (r *FileRepository) MarkReady(ctx context.Context, id string) (int64, error) {
	res := r.getDB(ctx).Model(&fileModel.File{}).
		Where("id = ? AND status = ?", id, fileModel.StatusPending).
		Update("status", fileModel.StatusReady)
	return res.RowsAffected, res.Error
}

// PendingOlderThan lists rows a direct upload never committed.
func (r *FileRepository) PendingOlderThan(ctx context.Context, cutoff int64, limit int) ([]fileModel.File, error) {
	var rows []fileModel.File
	err := r.getDB(ctx).
		Where("status = ? AND created_at < ?", fileModel.StatusPending, cutoff).
		Order("created_at").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FileRepository) ListFiles(ctx context.Context, q fileModel.ListFilter) ([]fileModel.File, int64, error) {
	db := r.getDB(ctx).Model(&fileModel.File{})
	if q.OwnerKind != "" && q.OwnerID != "" {
		db = db.Where(
			"files.id IN (?)",
			r.getDB(ctx).Model(&fileModel.Link{}).
				Select("file_id").
				Where("owner_kind = ? AND owner_id = ?", q.OwnerKind, q.OwnerID),
		)
	}
	if q.Role != "" {
		db = db.Where(
			"files.id IN (?)",
			r.getDB(ctx).Model(&fileModel.Link{}).Select("file_id").Where("role = ?", q.Role),
		)
	}
	if q.UploaderID != "" {
		db = db.Where("uploader_id = ?", q.UploaderID)
	}
	if q.Kind != "" {
		db = db.Where("kind = ?", q.Kind)
	}
	if q.Status != "" {
		db = db.Where("status = ?", q.Status)
	}
	if q.Access != "" {
		db = db.Where("access = ?", q.Access)
	}
	if search := strings.TrimSpace(q.Search); search != "" {
		like := "%" + search + "%"
		db = db.Where("name LIKE ? OR object_key LIKE ?", like, like)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []fileModel.File{}, 0, nil
	}

	page, pageSize := q.Page, q.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	var rows []fileModel.File
	err := db.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&rows).Error
	return rows, total, err
}

// ---- file_blobs --------------------------------------------------------

func (r *FileRepository) PutBlob(ctx context.Context, fileID string, data []byte) error {
	return r.getDB(ctx).Create(&fileModel.Blob{FileID: fileID, Data: data}).Error
}

func (r *FileRepository) GetBlob(ctx context.Context, fileID string) ([]byte, error) {
	var blob fileModel.Blob
	if err := r.getDB(ctx).Where("file_id = ?", fileID).First(&blob).Error; err != nil {
		return nil, err
	}
	return blob.Data, nil
}

// ---- file_links --------------------------------------------------------

func (r *FileRepository) AddLinks(ctx context.Context, links []fileModel.Link) error {
	if len(links) == 0 {
		return nil
	}
	return r.getDB(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&links).Error
}

func (r *FileRepository) DeleteLink(ctx context.Context, fileID, ownerKind, ownerID, role string) error {
	return r.getDB(ctx).
		Where("file_id = ? AND owner_kind = ? AND owner_id = ? AND role = ?", fileID, ownerKind, ownerID, role).
		Delete(&fileModel.Link{}).Error
}

func (r *FileRepository) DeleteOwnerLinks(ctx context.Context, ownerKind, ownerID, role string) error {
	db := r.getDB(ctx).Where("owner_kind = ? AND owner_id = ?", ownerKind, ownerID)
	if role != "" {
		db = db.Where("role = ?", role)
	}
	return db.Delete(&fileModel.Link{}).Error
}

func (r *FileRepository) LinksOfFile(ctx context.Context, fileID string) ([]fileModel.Link, error) {
	var links []fileModel.Link
	err := r.getDB(ctx).Where("file_id = ?", fileID).Find(&links).Error
	return links, err
}

func (r *FileRepository) LinksOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.Link, error) {
	db := r.getDB(ctx).Where("owner_kind = ? AND owner_id = ?", ownerKind, ownerID)
	if role != "" {
		db = db.Where("role = ?", role)
	}
	var links []fileModel.Link
	err := db.Order("role, sort_order").Find(&links).Error
	return links, err
}

func (r *FileRepository) LinksForFiles(ctx context.Context, ids []string) (map[string][]fileModel.Link, error) {
	if len(ids) == 0 {
		return map[string][]fileModel.Link{}, nil
	}
	var links []fileModel.Link
	if err := r.getDB(ctx).Where("file_id IN ?", ids).Find(&links).Error; err != nil {
		return nil, err
	}
	out := make(map[string][]fileModel.Link, len(ids))
	for _, link := range links {
		out[link.FileID] = append(out[link.FileID], link)
	}
	return out, nil
}

// FilesOfOwner returns the files a domain attached, ordered by role then slot
// order. It is the echo hydration path.
func (r *FileRepository) FilesOfOwner(ctx context.Context, ownerKind, ownerID, role string) ([]fileModel.File, error) {
	db := r.getDB(ctx).
		Table("files").
		Select("files.*").
		Joins("JOIN file_links ON file_links.file_id = files.id").
		Where("file_links.owner_kind = ? AND file_links.owner_id = ?", ownerKind, ownerID)
	if role != "" {
		db = db.Where("file_links.role = ?", role)
	}
	var rows []fileModel.File
	err := db.Order("file_links.role, file_links.sort_order").Find(&rows).Error
	return rows, err
}

func (r *FileRepository) CountLinks(ctx context.Context, fileID string) (int64, error) {
	var n int64
	err := r.getDB(ctx).Model(&fileModel.Link{}).Where("file_id = ?", fileID).Count(&n).Error
	return n, err
}

// UnlinkedFilesOfUploader returns the files a user uploaded that no link
// mentions any more. They are the only files a user deletion removes: a file
// still attached to someone else's echo is theirs to keep.
func (r *FileRepository) UnlinkedFilesOfUploader(ctx context.Context, uploaderID string) ([]fileModel.File, error) {
	var rows []fileModel.File
	err := r.getDB(ctx).
		Where("uploader_id = ?", uploaderID).
		Where("NOT EXISTS (SELECT 1 FROM file_links WHERE file_links.file_id = files.id)").
		Find(&rows).Error
	return rows, err
}

// ---- file_gc -----------------------------------------------------------

func (r *FileRepository) AddGCTask(ctx context.Context, task *fileModel.GCTask) error {
	return r.getDB(ctx).Create(task).Error
}

func (r *FileRepository) DueGCTasks(ctx context.Context, now int64, limit int) ([]fileModel.GCTask, error) {
	var rows []fileModel.GCTask
	err := r.getDB(ctx).
		Where("next_try_at <= ?", now).
		Order("next_try_at").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *FileRepository) UpdateGCTask(ctx context.Context, task *fileModel.GCTask) error {
	return r.getDB(ctx).Model(&fileModel.GCTask{}).
		Where("id = ?", task.ID).
		Updates(map[string]any{
			"attempts":    task.Attempts,
			"last_error":  task.LastError,
			"next_try_at": task.NextTryAt,
		}).Error
}

func (r *FileRepository) DeleteGCTask(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&fileModel.GCTask{}).Error
}

func (r *FileRepository) CountGCTasks(ctx context.Context) (int64, error) {
	var n int64
	err := r.getDB(ctx).Model(&fileModel.GCTask{}).Count(&n).Error
	return n, err
}

// ---- storages ----------------------------------------------------------

func (r *FileRepository) ListStorages(ctx context.Context) ([]fileModel.StorageConfig, error) {
	var rows []fileModel.StorageConfig
	err := r.getDB(ctx).Order("created_at").Find(&rows).Error
	return rows, err
}

func (r *FileRepository) SaveStorage(ctx context.Context, cfg *fileModel.StorageConfig) error {
	return r.getDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"driver", "name", "config", "is_default", "updated_at"}),
	}).Create(cfg).Error
}

func (r *FileRepository) DeleteStorage(ctx context.Context, id string) error {
	return r.getDB(ctx).Where("id = ?", id).Delete(&fileModel.StorageConfig{}).Error
}

// SetDefaultStorage moves the default flag in one transaction. The partial
// unique index allows a single default row, so the clear has to land before
// the set.
func (r *FileRepository) SetDefaultStorage(ctx context.Context, id string) error {
	db := r.getDB(ctx)
	if err := db.Model(&fileModel.StorageConfig{}).
		Where("is_default = ? AND id <> ?", true, id).
		Update("is_default", false).Error; err != nil {
		return err
	}
	return db.Model(&fileModel.StorageConfig{}).Where("id = ?", id).
		Update("is_default", true).Error
}
