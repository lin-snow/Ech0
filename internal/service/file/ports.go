// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	userModel "github.com/lin-snow/ech0/internal/model/user"
	"github.com/lin-snow/ech0/internal/storage"
)

type Service interface {
	UploadFile(
		ctx context.Context,
		file *multipart.FileHeader,
		category storage.Category,
		storageType storage.StorageType,
	) (commonModel.FileDto, error)
	CreateExternalFile(ctx context.Context, dto commonModel.CreateExternalFileDto) (commonModel.FileDto, error)
	DeleteFile(ctx context.Context, id string) error
	GetFileByID(ctx context.Context, id string) (commonModel.FileDto, error)
	GetFilesByIDs(ctx context.Context, ids []string) ([]commonModel.FileDto, error)
	ListFiles(ctx context.Context, query commonModel.FileListQueryDto) (commonModel.FileListResultDto, error)
	ListFileTree(ctx context.Context, query commonModel.FileTreeQueryDto) (commonModel.FileTreeResultDto, error)
	UpdateFileMeta(ctx context.Context, id string, dto commonModel.UpdateFileMetaDto) (commonModel.FileDto, error)
	StreamFileByID(ctx *gin.Context, id string)
	StreamFileByPath(ctx *gin.Context, query commonModel.FilePathStreamQueryDto)
	GetFilePresignURL(ctx context.Context, dto *commonModel.GetPresignURLDto) (commonModel.PresignDto, error)
	CleanupOrphanFiles() error
	ConfirmTempFiles(ctx context.Context, fileIDs []string) error
	DeleteFileRecord(ctx context.Context, id string) error
	DeleteStoredFile(storageType string, key string) error
}

type CommonRepository interface {
	GetUserByUserId(ctx context.Context, id string) (userModel.User, error)
}

type FileRepository interface {
	Create(ctx context.Context, file *fileModel.File) error
	GetByID(ctx context.Context, id string) (*fileModel.File, error)
	ListByIDs(ctx context.Context, ids []string) ([]fileModel.File, error)
	GetByKey(ctx context.Context, key string) (*fileModel.File, error)
	GetByRoute(ctx context.Context, storageType, provider, bucket, key string) (*fileModel.File, error)
	ListByStorageTypeAndSearch(
		ctx context.Context,
		storageType string,
		search string,
		page int,
		pageSize int,
	) ([]fileModel.File, int64, error)
	ListByStorageTypeAndKeys(ctx context.Context, storageType string, keys []string) ([]fileModel.File, error)
	ListByStorageTypeAndURLs(ctx context.Context, storageType string, urls []string) ([]fileModel.File, error)
	UpdateMetaByID(
		ctx context.Context,
		id string,
		size int64,
		width *int,
		height *int,
		contentType *string,
	) (*fileModel.File, error)
	CreateTemp(ctx context.Context, temp *fileModel.TempFile) error
	DeleteTempByFileID(ctx context.Context, fileID string) error
	DeleteTempByID(ctx context.Context, id string) error
	ListExpiredTemps(ctx context.Context, before int64) ([]fileModel.TempFile, error)
	Delete(ctx context.Context, id string) error
	DeleteByRoute(ctx context.Context, storageType, provider, bucket, key string) error
}
