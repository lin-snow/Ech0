// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package importer

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"path"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	logUtil "github.com/lin-snow/ech0/pkg/log"
	"github.com/lin-snow/ech0/pkg/virefs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	externalKeyPrefix   = "external/"
	externalDefaultName = "external"
	defaultContentType  = "application/octet-stream"
)

type fileEntry struct {
	id   string
	key  string
	size int64
	sum  string
}

func routeKey(storageType, provider, bucket, key string) string {
	return storageType + "|" + provider + "|" + bucket + "|" + key
}

func (s *session) importFiles(ctx context.Context, docPath string, doc *capsule.EchoDoc, userID string) error {
	for idx := range doc.Files {
		ref := doc.Files[idx]
		fileID, err := s.ensureFile(ctx, docPath, ref, userID)
		if err != nil {
			return err
		}
		link := fileModel.EchoFile{EchoID: doc.ID, FileID: fileID, SortOrder: idx}
		if err := s.db.Omit(clause.Associations).Create(&link).Error; err != nil {
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
	return s.ensureExternalFile(docPath, ref, userID)
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

	entry, err := s.lookupRoute(ref.Key)
	if err != nil {
		return "", fmt.Errorf("capsule import: %s: probe file key %q: %w", docPath, ref.Key, err)
	}

	key := ref.Key
	renamed := false
	if entry != nil {
		if s.sameContent(ctx, entry, data) {
			s.res.FilesReused++
			return entry.id, nil
		}
		generated, gerr := s.keygen.GenerateKey(fileCategory(ref), userID, ref.Key)
		if gerr != nil {
			return "", fmt.Errorf("capsule import: %s: generate key for %q: %w", docPath, ref.Key, gerr)
		}
		key = generated
		renamed = true
	}

	row := s.newFileRow(ref, key, data, userID)
	if err := s.putBytes(ctx, key, data, row.ContentType); err != nil {
		return "", fmt.Errorf("capsule import: %s: store bytes for key %q: %w", docPath, key, err)
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", fmt.Errorf("capsule import: %s: create file row for key %q: %w", docPath, key, err)
	}
	s.rememberRoute(key, &fileEntry{id: row.ID, key: key, size: row.Size, sum: sha256Hex(data)})

	if renamed {
		s.res.FilesRenamed++
		s.res.Renames = append(s.res.Renames, ref.Key+" -> "+key)
	} else {
		s.res.FilesCreated++
	}
	return row.ID, nil
}

func (s *session) ensureExternalFile(docPath string, ref capsule.FileRef, userID string) (string, error) {
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

	category := fileCategory(ref)
	hash := sha256.Sum256([]byte(ref.URL))
	key := externalKeyPrefix + string(category) + "/" + hex.EncodeToString(hash[:])
	const (
		externalStorageType = string(storage.StorageTypeExternal)
		externalProvider    = string(storage.StorageTypeExternal)
		externalBucket      = ""
	)

	var existing fileModel.File
	err := s.db.Where("storage_type = ? AND provider = ? AND bucket = ? AND key = ?",
		externalStorageType, externalProvider, externalBucket, key).First(&existing).Error
	switch {
	case err == nil:
		s.res.FilesReused++
		return existing.ID, nil
	case errors.Is(err, gorm.ErrRecordNotFound):
	default:
		return "", fmt.Errorf("capsule import: %s: probe external file: %w", docPath, err)
	}

	name := ref.Name
	if name == "" {
		name = externalDefaultName
	}
	contentType := ref.ContentType
	if contentType == "" {
		contentType = mimeForName(ref.URL)
	}
	row := fileModel.File{
		ID:          ref.ID,
		Key:         key,
		StorageType: externalStorageType,
		Provider:    externalProvider,
		Bucket:      externalBucket,
		URL:         ref.URL,
		Name:        name,
		ContentType: contentType,
		Size:        ref.Size,
		Width:       ref.Width,
		Height:      ref.Height,
		Category:    string(category),
		UserID:      userID,
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", fmt.Errorf("capsule import: %s: create external file row: %w", docPath, err)
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

func (s *session) lookupRoute(key string) (*fileEntry, error) {
	rk := routeKey(string(s.storageType), s.provider, s.bucket, key)
	if entry, ok := s.routeCache[rk]; ok {
		return entry, nil
	}

	var row fileModel.File
	err := s.db.Where("storage_type = ? AND provider = ? AND bucket = ? AND key = ?",
		string(s.storageType), s.provider, s.bucket, key).First(&row).Error
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return nil, nil
	case err != nil:
		return nil, err
	}
	entry := &fileEntry{id: row.ID, key: row.Key, size: row.Size}
	s.routeCache[rk] = entry
	return entry, nil
}

func (s *session) rememberRoute(key string, entry *fileEntry) {
	s.routeCache[routeKey(string(s.storageType), s.provider, s.bucket, key)] = entry
}

func (s *session) sameContent(ctx context.Context, entry *fileEntry, data []byte) bool {
	if entry.size != int64(len(data)) {
		return false
	}
	if entry.sum == "" {
		stored, err := s.readStored(ctx, entry.key)
		if err != nil {
			logUtil.GetLogger().Warn("capsule import cannot read stored object for comparison",
				slog.String("module", logModule),
				slog.String("key", entry.key),
				logUtil.Err(err),
			)
			return false
		}
		entry.sum = sha256Hex(stored)
	}
	return entry.sum == sha256Hex(data)
}

func (s *session) readStored(ctx context.Context, key string) ([]byte, error) {
	rc, err := s.selector.Get(ctx, s.storageType, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
}

func (s *session) putBytes(ctx context.Context, key string, data []byte, contentType string) error {
	if s.opts.DryRun {
		return nil
	}
	return s.selector.Put(ctx, s.storageType, key, bytes.NewReader(data), virefs.WithContentType(contentType))
}

func (s *session) newFileRow(ref capsule.FileRef, key string, data []byte, userID string) fileModel.File {
	name := ref.Name
	if name == "" {
		name = ref.Key
	}
	contentType := ref.ContentType
	if contentType == "" {
		contentType = mimeForName(ref.Key)
	}
	size := ref.Size
	if size == 0 {
		size = int64(len(data))
	}
	return fileModel.File{
		ID:          ref.ID,
		Key:         key,
		StorageType: string(s.storageType),
		Provider:    s.provider,
		Bucket:      s.bucket,
		URL:         "",
		Name:        name,
		ContentType: contentType,
		Size:        size,
		Width:       ref.Width,
		Height:      ref.Height,
		Category:    string(fileCategory(ref)),
		UserID:      userID,
	}
}

func fileCategory(ref capsule.FileRef) storage.Category {
	if _, ok := capsule.ValidCategories[ref.Category]; ok {
		return storage.Category(ref.Category)
	}
	name := ref.Key
	if name == "" {
		name = ref.URL
	}
	return categoryForExt(path.Ext(nameOnly(name)))
}

func categoryForExt(ext string) storage.Category {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif", ".bmp", ".ico":
		return storage.CategoryImage
	case ".mp3", ".flac", ".wav", ".m4a", ".ogg":
		return storage.CategoryAudio
	case ".mp4", ".avi", ".mkv", ".webm", ".mov":
		return storage.CategoryVideo
	case ".pdf":
		return storage.CategoryPDF
	case ".md", ".markdown":
		return storage.CategoryMarkdown
	default:
		return storage.CategoryFile
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

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
