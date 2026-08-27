// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/config"
	"github.com/lin-snow/ech0/pkg/virefs"
)

type StorageSelector struct {
	localFS        virefs.FS
	objectFS       virefs.FS
	localResolve   URLResolver
	objectResolve  URLResolver
	localPathURL   URLResolver
	objectPathURL  URLResolver
	objectEnabled  bool
	objectProvider string
	objectBucket   string
	objectPrefix   string
	localRoot      string
}

type ListNode struct {
	Name         string
	Path         string
	IsDir        bool
	Size         int64
	ContentType  string
	LastModified time.Time
}

func NewStorageSelector(cfg config.StorageConfig) *StorageSelector {
	schema := NewFileSchema()
	localRoot := strings.TrimSpace(cfg.DataRoot)
	if localRoot == "" {
		localRoot = "data/files"
	}

	localFS := buildLocalFS(cfg, schema)
	localResolve := buildLocalURLResolver(schema)
	localPathResolve := buildLocalPathURLResolver()

	objectFS, objectResolve, objectPathResolve, objectEnabled := buildOptionalObjectFSAndResolver(cfg, schema)

	return &StorageSelector{
		localFS:        localFS,
		objectFS:       objectFS,
		localResolve:   localResolve,
		objectResolve:  objectResolve,
		localPathURL:   localPathResolve,
		objectPathURL:  objectPathResolve,
		objectEnabled:  objectEnabled,
		objectProvider: strings.ToLower(strings.TrimSpace(cfg.Provider)),
		objectBucket:   strings.TrimSpace(cfg.BucketName),
		objectPrefix:   strings.Trim(strings.TrimSpace(cfg.PathPrefix), "/"),
		localRoot:      localRoot,
	}
}

func (r *StorageSelector) ObjectEnabled() bool {
	return r != nil && r.objectEnabled && r.objectFS != nil
}

func (r *StorageSelector) ObjectRoute() (provider string, bucket string) {
	if r == nil || !r.ObjectEnabled() {
		return "", ""
	}
	return r.objectProvider, r.objectBucket
}

func (r *StorageSelector) Put(
	ctx context.Context,
	storageType StorageType,
	key string,
	reader io.Reader,
	opts ...virefs.PutOption,
) error {
	fs, err := r.getFS(storageType)
	if err != nil {
		return err
	}
	return fs.Put(ctx, key, reader, opts...)
}

func (r *StorageSelector) Get(ctx context.Context, storageType StorageType, key string) (io.ReadCloser, error) {
	fs, err := r.getFS(storageType)
	if err != nil {
		return nil, err
	}
	return fs.Get(ctx, key)
}

func (r *StorageSelector) Delete(ctx context.Context, storageType StorageType, key string) error {
	fs, err := r.getFS(storageType)
	if err != nil {
		return err
	}
	return fs.Delete(ctx, key)
}

func (r *StorageSelector) GetByStoragePath(
	ctx context.Context,
	storageType StorageType,
	filePath string,
) (io.ReadCloser, error) {
	if r == nil {
		return nil, errors.New("storage selector is not initialized")
	}
	cleanPath := strings.Trim(strings.TrimSpace(filePath), "/")
	if cleanPath == "" {
		return nil, errors.New("file path is empty")
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if !r.ObjectEnabled() {
			return nil, errors.New("object storage is not enabled")
		}
		return r.objectFS.Get(ctx, cleanPath)
	case StorageTypeExternal:
		return nil, errors.New("external storage does not support filesystem operations")
	default:
		cleanRel := filepath.Clean(filepath.FromSlash(cleanPath))
		if cleanRel == "." || cleanRel == "" || strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
			return nil, errors.New("invalid file path")
		}
		return os.Open(filepath.Join(r.localRoot, cleanRel))
	}
}

func (r *StorageSelector) ResolveURL(storageType StorageType, key string) string {
	if r == nil {
		return ""
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if r.ObjectEnabled() && r.objectResolve != nil {
			return r.objectResolve(key)
		}
		return ""
	default:
		if r.localResolve != nil {
			return r.localResolve(key)
		}
		return ""
	}
}

func (r *StorageSelector) ResolveURLByPath(storageType StorageType, filePath string) string {
	if r == nil {
		return ""
	}
	cleanPath := strings.Trim(strings.TrimSpace(filePath), "/")
	if cleanPath == "" {
		return ""
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if r.ObjectEnabled() && r.objectPathURL != nil {
			return r.objectPathURL(cleanPath)
		}
		return ""
	default:
		if r.localPathURL != nil {
			return r.localPathURL(cleanPath)
		}
		return ""
	}
}

func (r *StorageSelector) ResolveKeyByPath(storageType StorageType, filePath string) string {
	candidates := r.ResolveKeyCandidatesByPath(storageType, filePath)
	if len(candidates) == 0 {
		return ""
	}
	return candidates[0]
}

func (r *StorageSelector) ResolveKeyCandidatesByPath(storageType StorageType, filePath string) []string {
	cleanPath := strings.Trim(strings.TrimSpace(filePath), "/")
	if cleanPath == "" {
		return nil
	}
	appendUnique := func(dst []string, seen map[string]struct{}, candidate string) []string {
		c := strings.Trim(strings.TrimSpace(candidate), "/")
		if c == "" {
			return dst
		}
		if _, ok := seen[c]; ok {
			return dst
		}
		seen[c] = struct{}{}
		return append(dst, c)
	}
	seen := make(map[string]struct{}, 8)
	candidates := make([]string, 0, 8)
	candidates = appendUnique(candidates, seen, cleanPath)

	if NormalizeStorageType(string(storageType)) == StorageTypeObject && r != nil && r.objectPrefix != "" {
		prefix := r.objectPrefix + "/"
		if after, ok := strings.CutPrefix(cleanPath, prefix); ok {
			candidates = appendUnique(candidates, seen, after)
		}
	}

	routePrefixes := []string{"images/", "audios/", "videos/", "documents/", "files/"}
	baseSnapshot := append([]string(nil), candidates...)
	for _, base := range baseSnapshot {
		for _, route := range routePrefixes {
			if after, ok := strings.CutPrefix(base, route); ok {
				candidates = appendUnique(candidates, seen, after)
			}
		}
	}

	candidates = appendUnique(candidates, seen, path.Base(cleanPath))
	for _, c := range append([]string(nil), candidates...) {
		candidates = appendUnique(candidates, seen, path.Base(c))
	}
	return candidates
}

func (r *StorageSelector) ListNodes(
	ctx context.Context,
	storageType StorageType,
	prefix string,
) ([]ListNode, error) {
	fs, err := r.getFS(storageType)
	if err != nil {
		return nil, err
	}
	cleanPrefix := strings.Trim(strings.TrimSpace(prefix), "/")
	result, err := fs.List(ctx, cleanPrefix)
	if err != nil {
		return nil, err
	}

	nodes := make([]ListNode, 0, len(result.Files))
	for _, item := range result.Files {
		cleanPath := strings.Trim(strings.TrimSpace(item.Key), "/")
		if cleanPath == "" {
			continue
		}
		nodes = append(nodes, ListNode{
			Name:         path.Base(cleanPath),
			Path:         cleanPath,
			IsDir:        item.IsDir,
			Size:         item.Size,
			ContentType:  item.ContentType,
			LastModified: item.LastModified,
		})
	}

	sort.SliceStable(nodes, func(i, j int) bool {
		if nodes[i].IsDir != nodes[j].IsDir {
			return nodes[i].IsDir
		}
		return nodes[i].Name < nodes[j].Name
	})
	return nodes, nil
}

func (r *StorageSelector) PresignPutURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	if !r.ObjectEnabled() {
		return "", errors.New("backend does not support presigned URLs")
	}
	p, ok := r.objectFS.(virefs.Presigner)
	if !ok {
		return "", errors.New("backend does not support presigned URLs")
	}
	req, err := p.PresignPut(ctx, key, expires)
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func (r *StorageSelector) getFS(storageType StorageType) (virefs.FS, error) {
	if r == nil {
		return nil, errors.New("storage selector is not initialized")
	}
	switch NormalizeStorageType(string(storageType)) {
	case StorageTypeObject:
		if !r.ObjectEnabled() {
			return nil, errors.New("object storage is not enabled")
		}
		return r.objectFS, nil
	case StorageTypeExternal:
		return nil, errors.New("external storage does not support filesystem operations")
	default:
		return r.localFS, nil
	}
}

func buildOptionalObjectFSAndResolver(
	cfg config.StorageConfig,
	schema *virefs.Schema,
) (virefs.FS, URLResolver, URLResolver, bool) {
	if !cfg.ObjectEnabled {
		return nil, nil, nil, false
	}

	var opts []virefs.ObjectOption
	if cfg.PathPrefix != "" {
		opts = append(opts, virefs.WithPrefix(strings.Trim(cfg.PathPrefix, "/")+"/"))
	}
	opts = append(opts, virefs.WithObjectKeyFunc(schema.Resolve))

	fs, err := virefs.NewObjectFSFromConfig(context.Background(), virefsS3ConfigFromStorage(cfg), opts...)
	if err != nil {
		return nil, nil, nil, false
	}

	return fs, buildS3URLResolver(cfg, schema), buildS3PathURLResolver(cfg), true
}

func (r *StorageSelector) CapabilityText() string {
	if r.ObjectEnabled() {
		return "local+object"
	}
	return "local"
}

func (r *StorageSelector) String() string {
	return fmt.Sprintf("StorageSelector(%s)", r.CapabilityText())
}
