package repository

import (
	"context"
	"time"

	model "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/transaction"
	"gorm.io/gorm"
)

type FileRepository struct {
	db func() *gorm.DB
}

func NewFileRepository(dbProvider func() *gorm.DB) FileRepositoryInterface {
	return &FileRepository{db: dbProvider}
}

func (r *FileRepository) getDB(ctx context.Context) *gorm.DB {
	if tx, ok := transaction.TxFromContext(ctx); ok {
		return tx
	}
	return r.db()
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	return r.getDB(ctx).Create(file).Error
}

func (r *FileRepository) GetByID(ctx context.Context, id uint) (*model.File, error) {
	var f model.File
	if err := r.getDB(ctx).First(&f, id).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) GetByKey(ctx context.Context, key string) (*model.File, error) {
	var f model.File
	if err := r.getDB(ctx).Where("key = ?", key).First(&f).Error; err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *FileRepository) Delete(ctx context.Context, id uint) error {
	return r.getDB(ctx).Delete(&model.File{}, id).Error
}

func (r *FileRepository) DeleteByKey(ctx context.Context, key string) error {
	return r.getDB(ctx).Where("key = ?", key).Delete(&model.File{}).Error
}

func (r *FileRepository) GetOrphanFiles(ctx context.Context, olderThan time.Time) ([]model.File, error) {
	var files []model.File
	err := r.getDB(ctx).
		Where("created_at < ?", olderThan).
		Where("id NOT IN (?)",
			r.getDB(ctx).Table("echo_files").Select("file_id"),
		).
		Find(&files).Error
	return files, err
}

func (r *FileRepository) GetByCategory(ctx context.Context, category string) ([]model.File, error) {
	var files []model.File
	err := r.getDB(ctx).Where("category = ?", category).Find(&files).Error
	return files, err
}
