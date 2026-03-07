package repository

import (
	"context"
	"time"

	model "github.com/lin-snow/ech0/internal/model/file"
)

// FileRepositoryInterface defines data-access operations for the files table.
type FileRepositoryInterface interface {
	Create(ctx context.Context, file *model.File) error
	GetByID(ctx context.Context, id uint) (*model.File, error)
	GetByKey(ctx context.Context, key string) (*model.File, error)
	Delete(ctx context.Context, id uint) error
	DeleteByKey(ctx context.Context, key string) error

	// GetOrphanFiles returns files not linked to any echo and older than the
	// given threshold — candidates for garbage collection.
	GetOrphanFiles(ctx context.Context, olderThan time.Time) ([]model.File, error)

	// GetByCategory returns all files matching the given category.
	GetByCategory(ctx context.Context, category string) ([]model.File, error)
}
