// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package importer

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"path"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
	"gorm.io/gorm"
)

const defaultContentType = "application/octet-stream"

func (s *session) importFiles(ctx context.Context, docPath string, doc *capsule.EchoDoc, userID string) error {
	for idx := range doc.Files {
		ref := doc.Files[idx]
		fileID, err := s.ensureFile(ctx, docPath, ref, userID)
		if err != nil {
			return err
		}
		link := fileModel.Link{
			FileID:    fileID,
			OwnerKind: fileModel.OwnerKindEcho,
			OwnerID:   doc.ID,
			Role:      fileModel.RoleAttachment,
			SortOrder: idx,
		}
		if err := s.db.Table("file_links").Create(&link).Error; err != nil {
			return fmt.Errorf("capsule import: %s: link file %s: %w", docPath, fileID, err)
		}
	}
	return nil
}

func (s *session) importUnattachedFiles(ctx context.Context) error {
	if s.loaded.Manifest == nil {
		return nil
	}
	for _, ref := range s.loaded.Manifest.Files {
		if _, err := s.ensureFile(ctx, capsule.ManifestPath, ref, s.ownerID); err != nil {
			return err
		}
	}
	return nil
}

func (s *session) ensureFile(
	ctx context.Context,
	docPath string,
	ref capsule.FileRef,
	userID string,
) (string, error) {
	if ref.Managed() {
		return s.ensureManagedFile(ctx, docPath, ref, userID)
	}
	return s.ensureRefFile(docPath, ref, userID)
}

func (s *session) ensureManagedFile(
	ctx context.Context,
	docPath string,
	ref capsule.FileRef,
	userID string,
) (string, error) {
	if ref.ID != "" {
		reusedID, found, err := s.reuseByID(ref.ID)
		if err != nil {
			return "", fmt.Errorf("capsule import: %s: probe file %s: %w", docPath, ref.ID, err)
		}
		if found {
			return reusedID, nil
		}
	}

	data, err := s.loaded.Source.ReadFile(ctx, capsule.MediaPath(ref.Key))
	if err != nil {
		return "", fmt.Errorf("capsule import: %s: read media for key %q: %w", docPath, ref.Key, err)
	}

	id := ref.ID
	if id == "" || !uuidUtil.IsValid(id) {
		id = uuidUtil.NewV7()
	}
	ext := path.Ext(ref.Key)
	objectKey, err := storage.KeyFor(id, ext)
	if err != nil {
		// A valid UUID that is not version 7 cannot name an object; fall back
		// to a fresh key rather than lose the bytes.
		id = uuidUtil.NewV7()
		objectKey, err = storage.KeyFor(id, ext)
		if err != nil {
			return "", fmt.Errorf("capsule import: %s: key for %q: %w", docPath, ref.Key, err)
		}
	}

	var existing fileModel.File
	err = s.db.
		Where("kind = ? AND object_key = ?", string(fileModel.KindOwned), objectKey).
		First(&existing).Error
	switch {
	case err == nil:
		s.res.FilesReused++
		return existing.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return "", fmt.Errorf("capsule import: %s: probe object key %q: %w", docPath, objectKey, err)
	}

	storageID, store, err := s.storage.Default()
	if err != nil {
		return "", fmt.Errorf("capsule import: %s: resolve default storage: %w", docPath, err)
	}

	contentType := ref.ContentType
	if contentType == "" {
		contentType = mimeForName(ref.Key)
	}
	if !s.opts.DryRun {
		if _, err := store.Put(ctx, objectKey, bytes.NewReader(data), int64(len(data))); err != nil {
			return "", fmt.Errorf("capsule import: %s: store bytes for key %q: %w", docPath, objectKey, err)
		}
	}

	row := fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindOwned),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessOwner),
		StorageID:   &storageID,
		ObjectKey:   &objectKey,
		Name:        ref.Name,
		ContentType: contentType,
		Size:        int64(len(data)),
		Width:       ref.Width,
		Height:      ref.Height,
		UploaderID:  userID,
	}
	if row.Name == "" {
		row.Name = ref.Key
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", fmt.Errorf("capsule import: %s: create file row for key %q: %w", docPath, objectKey, err)
	}
	s.res.FilesCreated++
	return row.ID, nil
}

func (s *session) ensureRefFile(docPath string, ref capsule.FileRef, userID string) (string, error) {
	if ref.URL == "" {
		return "", fmt.Errorf("capsule import: %s: file ref has neither key nor url", docPath)
	}
	if ref.ID != "" {
		reusedID, found, err := s.reuseByID(ref.ID)
		if err != nil {
			return "", fmt.Errorf("capsule import: %s: probe file %s: %w", docPath, ref.ID, err)
		}
		if found {
			return reusedID, nil
		}
	}

	var existing fileModel.File
	err := s.db.
		Where("kind = ? AND ref_url = ?", string(fileModel.KindRef), ref.URL).
		First(&existing).Error
	switch {
	case err == nil:
		s.res.FilesReused++
		return existing.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return "", fmt.Errorf("capsule import: %s: probe ref %q: %w", docPath, ref.URL, err)
	}

	id := ref.ID
	if id == "" || !uuidUtil.IsValid(id) {
		id = uuidUtil.NewV7()
	}
	url := ref.URL
	name := ref.Name
	if name == "" {
		name = ref.URL
	}
	contentType := ref.ContentType
	if contentType == "" {
		contentType = mimeForName(ref.URL)
	}
	row := fileModel.File{
		ID:          id,
		Kind:        string(fileModel.KindRef),
		Status:      string(fileModel.StatusReady),
		Access:      string(fileModel.AccessOwner),
		RefURL:      &url,
		Name:        name,
		ContentType: contentType,
		Size:        ref.Size,
		Width:       ref.Width,
		Height:      ref.Height,
		UploaderID:  userID,
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", fmt.Errorf("capsule import: %s: create ref file row: %w", docPath, err)
	}
	s.res.FilesCreated++
	return row.ID, nil
}

func (s *session) reuseByID(id string) (string, bool, error) {
	var row fileModel.File
	err := s.db.Where("id = ?", id).First(&row).Error
	switch {
	case err == nil:
		s.res.FilesReused++
		return row.ID, true, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return "", false, nil
	default:
		return "", false, err
	}
}

func nameOnly(raw string) string {
	if i := strings.IndexAny(raw, "?#"); i >= 0 {
		raw = raw[:i]
	}
	return path.Base(raw)
}

func mimeForName(name string) string {
	if ct := mime.TypeByExtension(path.Ext(nameOnly(name))); ct != "" {
		return ct
	}
	return defaultContentType
}
