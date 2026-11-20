// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/internal/authz"
	"github.com/lin-snow/ech0/internal/config"
	authModel "github.com/lin-snow/ech0/internal/model/auth"
	commonModel "github.com/lin-snow/ech0/internal/model/common"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/viewer"
	"gorm.io/gorm"
)

func uploadDTO(f *fileModel.File) commonModel.FileDto {
	return commonModel.FileDto{ID: f.ID, Name: f.Name, Key: f.Key, StorageType: f.StorageType, URL: f.URL, ContentType: f.ContentType, Category: f.Category, Size: f.Size, Width: f.Width, Height: f.Height}
}

func (s *FileService) finalizedUpload(ctx context.Context, owner, id string) (commonModel.FileDto, error) {
	f, err := s.fileRepository.GetByID(ctx, id)
	if err != nil {
		return commonModel.FileDto{}, err
	}
	if f.UserID != owner || storage.NormalizeStorageType(f.StorageType) != storage.StorageTypeObject {
		return commonModel.FileDto{}, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	return uploadDTO(f), nil
}

func (s *FileService) finalizeUpload(ctx context.Context, owner, id string, dto commonModel.UpdateFileMetaDto) (commonModel.FileDto, error) {
	var empty commonModel.FileDto
	intent, err := s.fileRepository.GetIntent(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.finalizedUpload(ctx, owner, id)
	}
	if err != nil {
		return empty, err
	}
	f := intent.Metadata
	if owner != f.UserID {
		return empty, errors.New(commonModel.NO_PERMISSION_DENIED)
	}
	selector := s.getSelector()
	provider, bucket := selector.ObjectRoute()
	if intent.ExpiresAt <= time.Now().Unix() || f.Provider != provider || f.Bucket != bucket || dto.Size != f.Size {
		return empty, errors.New(commonModel.INVALID_PARAMS)
	}
	info, err := selector.StatObject(ctx, f.Key)
	if err != nil {
		return empty, err
	}
	if info.Size != f.Size || info.ContentType != f.ContentType {
		return empty, errors.New(commonModel.INVALID_PARAMS)
	}
	body, err := selector.Get(ctx, storage.StorageTypeObject, f.Key)
	if err != nil {
		return empty, err
	}
	prefix, readErr := io.ReadAll(io.LimitReader(body, 512))
	_ = body.Close()
	if readErr != nil {
		return empty, readErr
	}
	if err := validateFileUpload(f.Name, http.DetectContentType(prefix), config.Config().Upload.AllowedTypes); err != nil {
		return empty, err
	}
	if dto.Width != nil {
		f.Width = *dto.Width
	}
	if dto.Height != nil {
		f.Height = *dto.Height
	}
	err = s.transactor.Run(ctx, func(tx context.Context) error {
		if err := s.fileRepository.ConsumeIntent(tx, id, time.Now().Unix()); err != nil {
			return err
		}
		if err := s.fileRepository.Create(tx, &f); err != nil {
			return err
		}
		return s.fileRepository.CreateTemp(tx, &fileModel.TempFile{FileID: id, UploaderID: owner, ExpireAt: time.Now().Add(tempFileTTL).Unix()})
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.finalizedUpload(ctx, owner, id)
	}
	if err != nil {
		return empty, err
	}
	return uploadDTO(&f), nil
}

func (s *FileService) cleanupExpiredIntents(ctx context.Context, now int64) error {
	intents, err := s.fileRepository.ExpiredIntents(ctx, now)
	if err != nil {
		return err
	}
	selector := s.getSelector()
	provider, bucket := selector.ObjectRoute()
	for _, intent := range intents {
		f := intent.Metadata
		if f.Provider != provider || f.Bucket != bucket {
			return errors.New("expired upload belongs to a different object store")
		}
		claimed, err := s.fileRepository.ClaimExpiredIntent(ctx, intent.ID, now)
		if err != nil {
			return err
		}
		if !claimed {
			continue
		}
		if err := selector.Delete(ctx, storage.StorageTypeObject, f.Key); err != nil {
			return err
		}
		if err := s.fileRepository.DeleteIntent(ctx, intent.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileService) MediaDownload(ctx context.Context, key string) (string, error) {
	if !config.Config().Upload.VerifiedDirect {
		return "", gorm.ErrRecordNotFound
	}
	selector := s.getSelector()
	route, provider, bucket := currentStorageRoute(selector, storage.StorageTypeObject)
	f, err := s.fileRepository.GetByRoute(ctx, route, provider, bucket, key)
	if err != nil {
		return "", err
	}
	public, err := s.fileRepository.IsPublicMedia(ctx, f)
	if err != nil {
		return "", err
	}
	if !public {
		v := viewer.MustFromContext(ctx)
		if v.UserID() == "" || (v.TokenType() != authModel.TokenTypeSession && !(authz.Has(v.Scopes(), authz.ScopeFileRead) && authz.Has(v.Scopes(), authz.ScopeEchoPrivate))) {
			return "", gorm.ErrRecordNotFound
		}
		user, err := s.commonRepository.GetUserByUserId(ctx, v.UserID())
		if err != nil {
			return "", err
		}
		if !user.IsAdmin {
			return "", gorm.ErrRecordNotFound
		}
	}
	return selector.PresignDownload(ctx, f.Key)
}
