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

// 外链文件在 files 表里也要占一行，而 key 是 not null 且参与唯一索引 idx_file_route。
// 这里复刻 file service 对 external 的既有派生（纯函数部分，不调 service——它要
// viewer 上下文还会发事件），好让同一外链无论从哪条路进来都收敛到同一行。
const (
	externalKeyPrefix   = "external/"
	externalDefaultName = "external"
	defaultContentType  = "application/octet-stream"
)

// fileEntry 是一条已落定的文件行的最小画像，用于撞 key 时的内容比对。
type fileEntry struct {
	id   string
	key  string
	size int64
	// sum 是存储字节的 sha256（hex）。按需计算并缓存：同一个 key 在一份胶囊里
	// 可能被多篇 Echo 引用，没必要反复回读整个对象。
	sum string
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
		// SortOrder 即数组下标：胶囊用顺序表达展示序，这里把它还原成列。
		link := fileModel.EchoFile{EchoID: doc.ID, FileID: fileID, SortOrder: idx}
		if err := s.db.Omit(clause.Associations).Create(&link).Error; err != nil {
			return fmt.Errorf("capsule import: %s: link file %s: %w", docPath, fileID, err)
		}
	}
	return nil
}

// importUnattachedFiles 落地清单 files 块里的文件行——它们不挂在任何 Echo 上
// （站点 logo、上传后没用上的附件），所以只建 files 行、不建 echo_files 关联。
//
// 少了这一步，logo 的字节虽然躺在胶囊里却没人把它写进目标实例的存储，
// site.server_logo 指向的 /api/files/... 迁移后就是死链。
// 归属挂 owner：胶囊不携带 user_id，而这些行没有 Echo 可以继承归属。
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

// ensureManagedFile 落地一个 key 条目（spec §11.1）。字节源恒为胶囊内的
// files/ + Resolve(key)，落点恒为当前默认后端。
func (s *session) ensureManagedFile(
	ctx context.Context,
	docPath string,
	ref capsule.FileRef,
	userID string,
) (string, error) {
	// a. id 是幂等锚点：同 id 行已在库中就直接复用，字节一个字节都不重写。
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
			// b-2. 同 key 同内容：复用既有行，字节不动。
			s.res.FilesReused++
			return entry.id, nil
		}
		// b-3. 同 key 不同内容（手写胶囊撞名）：另起一个 key 落盘，绝不覆盖既有字节。
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

// ensureExternalFile 落地一个 url 条目：URL 即权威表示，原样透传，
// AfterFind 对 external 不会重算它。
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
		// 外链字节不随胶囊走，size 原样透传胶囊值——export 侧对 external 行导出的
		// 就是 File.Size（实践中为 0），此处不做任何推导。
		Size:     ref.Size,
		Width:    ref.Width,
		Height:   ref.Height,
		Category: string(category),
		UserID:   userID,
	}
	if err := s.db.Create(&row).Error; err != nil {
		return "", fmt.Errorf("capsule import: %s: create external file row: %w", docPath, err)
	}
	s.res.FilesCreated++
	return row.ID, nil
}

// reuseByID 查 id 幂等锚点。found=false 表示该 id 尚未在库中，调用方继续走 key 路径。
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

// lookupRoute 按 files 表的唯一索引四元组（storage_type, provider, bucket, key）
// 定位既有行；nil 表示当前后端下这个 key 还空着。
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

// rememberRoute 让本次新建的行对后续引用可见。dry-run 下这是唯一的可见途径
// （行没真写进库），同一胶囊内的重复 key 因此依然能被识别。
func (s *session) rememberRoute(key string, entry *fileEntry) {
	s.routeCache[routeKey(string(s.storageType), s.provider, s.bucket, key)] = entry
}

// sameContent 判断既有行与胶囊字节是否同一份内容：先比 size（便宜），再比 sha256。
// 字节读不回来时一律判为「不同」——那样最坏是多落一份改名副本，而覆盖既有 key
// 会不可逆地毁掉别的记录引用的字节。
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

// putBytes 把胶囊里的字节写进当前后端。dry-run 在这里断开：写盘/写对象存储不在
// 事务内，哨兵回滚兜不住它。
func (s *session) putBytes(ctx context.Context, key string, data []byte, contentType string) error {
	if s.opts.DryRun {
		return nil
	}
	return s.selector.Put(ctx, s.storageType, key, bytes.NewReader(data), virefs.WithContentType(contentType))
}

// newFileRow 把 FileRef 1:1 摊成 files 行。URL 留空：托管文件的直链由 AfterFind
// 按目标实例的当前配置重算，把源实例的 URL 带过来只会造出跨站死链。
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

// fileCategory 取胶囊值，缺失（或不在枚举内）时按扩展名推导。
// storage.NormalizeCategory 只认已知枚举名、无法从扩展名反推，所以这里另备一张
// 与 storage.NewFileSchema 的路由表同源的映射。
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

// nameOnly 剥掉 URL 的 query / fragment，好让扩展名推导对 "a.png?v=1" 免疫。
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
