// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/pkg/virefs"
)

func writeCapsule(
	ctx context.Context,
	deps Deps,
	stage virefs.FS,
	data *dataset,
	opts Options,
) ([]string, error) {
	keys := make([]string, 0, 1+len(data.echoes)+1+len(data.files))

	manifest := &capsule.Manifest{
		SchemaVersion: capsule.SchemaVersion,
		Generator:     opts.Generator,
		ExportedAt:    capsule.FormatUnix(time.Now().Unix()),
		Site:          data.site,
		Owner:         data.owner,
		Connects:      data.connects,
		Files:         unattachedRefs(data),
	}
	body, err := capsule.EncodeYAML(manifest)
	if err != nil {
		return nil, fmt.Errorf("capsule export: encode %s: %w", capsule.ManifestPath, err)
	}
	if err := put(ctx, stage, capsule.ManifestPath, body); err != nil {
		return nil, err
	}
	keys = append(keys, capsule.ManifestPath)

	echoKeys, err := writeEchoes(ctx, stage, data)
	if err != nil {
		return nil, err
	}
	keys = append(keys, echoKeys...)

	if len(data.comments) > 0 {
		doc := &capsule.CommentsDoc{SchemaVersion: capsule.SchemaVersion, Comments: data.comments}
		body, err := capsule.EncodeYAML(doc)
		if err != nil {
			return nil, fmt.Errorf("capsule export: encode %s: %w", capsule.CommentsPath, err)
		}
		if err := put(ctx, stage, capsule.CommentsPath, body); err != nil {
			return nil, err
		}
		keys = append(keys, capsule.CommentsPath)
	}

	mediaKeys, err := writeMedia(ctx, deps, stage, data)
	if err != nil {
		return nil, err
	}
	return append(keys, mediaKeys...), nil
}

func writeEchoes(ctx context.Context, stage virefs.FS, data *dataset) ([]string, error) {
	keys := make([]string, 0, len(data.echoes))
	used := make(map[string]struct{}, len(data.echoes))

	for i := range data.echoes {
		echo := &data.echoes[i]
		doc := &capsule.EchoDoc{
			ID:        echo.ID,
			CreatedAt: capsule.FormatUnix(echo.CreatedAt),
			Username:  echo.Username,
			Tags:      tagNames(echo.Tags),
			Layout:    echo.Layout,
			Private:   echo.Private,
			Files:     fileRefs(echo.EchoFiles),
			Extension: extension(echo.Extension),
			Content:   echo.Content,
		}
		body, err := capsule.EncodeEcho(doc)
		if err != nil {
			return nil, fmt.Errorf("capsule export: encode echo %s: %w", echo.ID, err)
		}

		key := uniquePath(used, capsule.EchoPath(echo.ID, time.Unix(echo.CreatedAt, 0)))
		if err := put(ctx, stage, key, body); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}

func uniquePath(used map[string]struct{}, base string) string {
	candidate := base
	for n := 2; ; n++ {
		if _, taken := used[candidate]; !taken {
			used[candidate] = struct{}{}
			return candidate
		}
		candidate = fmt.Sprintf("%s-%d.md", strings.TrimSuffix(base, ".md"), n)
	}
}

func tagNames(tags []echoModel.Tag) []string {
	if len(tags) == 0 {
		return nil
	}
	names := make([]string, 0, len(tags))
	for i := range tags {
		names = append(names, tags[i].Name)
	}
	return names
}

func extension(ext *echoModel.EchoExtension) *capsule.Extension {
	if ext == nil {
		return nil
	}
	return &capsule.Extension{Type: ext.Type, Payload: ext.Payload}
}

func fileRefs(links []echoModel.EchoFile) []capsule.FileRef {
	if len(links) == 0 {
		return nil
	}
	refs := make([]capsule.FileRef, 0, len(links))
	for i := range links {
		refs = append(refs, fileRef(links[i].File))
	}
	return refs
}

func fileRef(file fileModel.File) capsule.FileRef {
	ref := capsule.FileRef{
		ID:          file.ID,
		Category:    fileModel.CategoryOf(file.ContentType),
		Name:        file.Name,
		ContentType: file.ContentType,
		Size:        file.Size,
		Width:       file.Width,
		Height:      file.Height,
	}
	switch {
	case file.IsRef():
		if file.RefURL != nil {
			ref.URL = *file.RefURL
		}
	case file.IsBlob():
		ref.Key = blobKey(file.ID)
	case file.IsOwned():
		if file.ObjectKey != nil {
			ref.Key = *file.ObjectKey
		}
	}
	return ref
}

func unattachedRefs(data *dataset) []capsule.FileRef {
	attached := make(map[string]struct{})
	for i := range data.echoes {
		for _, link := range data.echoes[i].EchoFiles {
			attached[link.FileID] = struct{}{}
		}
	}

	refs := make([]capsule.FileRef, 0)
	for i := range data.files {
		if _, ok := attached[data.files[i].ID]; ok {
			continue
		}
		refs = append(refs, fileRef(data.files[i]))
	}
	if len(refs) == 0 {
		return nil
	}
	return refs
}

type mediaFailure struct {
	key string
	err error
}

func writeMedia(ctx context.Context, deps Deps, stage virefs.FS, data *dataset) ([]string, error) {
	keys := make([]string, 0, len(data.files))
	var failures []mediaFailure

	for i := range data.files {
		file := &data.files[i]
		if file.IsRef() {
			data.externalFiles++
			continue
		}

		key, reader, err := openMedia(ctx, deps, file)
		if err != nil {
			if reader != nil {
				_ = reader.Close()
			}
			failures = append(failures, mediaFailure{key: key, err: err})
			continue
		}
		if err := capsule.ValidateKey(key); err != nil {
			_ = reader.Close()
			failures = append(failures, mediaFailure{key: key, err: err})
			continue
		}

		path := capsule.MediaPath(key)
		err = stage.Put(ctx, path, reader)
		_ = reader.Close()
		if err != nil {
			return nil, fmt.Errorf("capsule export: write %s: %w", path, err)
		}
		keys = append(keys, path)
	}

	if len(failures) > 0 {
		return nil, unreadableMediaError(failures)
	}
	return keys, nil
}

func openMedia(ctx context.Context, deps Deps, file *fileModel.File) (string, io.ReadCloser, error) {
	switch {
	case file.IsOwned():
		if file.StorageID == nil || file.ObjectKey == nil {
			return file.ID, nil, errors.New("owned file is missing its storage locator")
		}
		store, err := deps.Storage.For(*file.StorageID)
		if err != nil {
			return *file.ObjectKey, nil, err
		}
		reader, _, err := store.Open(ctx, *file.ObjectKey)
		return *file.ObjectKey, reader, err
	case file.IsBlob():
		key := blobKey(file.ID)
		var data []byte
		if err := deps.DB.WithContext(ctx).
			Raw("SELECT data FROM file_blobs WHERE file_id = ?", file.ID).
			Row().Scan(&data); err != nil {
			return key, nil, fmt.Errorf("read file_blobs: %w", err)
		}
		return key, io.NopCloser(bytes.NewReader(data)), nil
	default:
		return file.ID, nil, fmt.Errorf("file kind %q has no managed bytes", file.Kind)
	}
}

func blobKey(fileID string) string {
	return "blob/" + fileID
}

func unreadableMediaError(failures []mediaFailure) error {
	var b strings.Builder
	fmt.Fprintf(&b, "capsule export: %d managed file(s) unreadable, capsule would not be self-contained:", len(failures))
	for _, f := range failures {
		fmt.Fprintf(&b, "\n  - key=%q: %v", f.key, f.err)
	}
	return errors.New(b.String())
}

func put(ctx context.Context, stage virefs.FS, key string, body []byte) error {
	if err := stage.Put(ctx, key, bytes.NewReader(body)); err != nil {
		return fmt.Errorf("capsule export: write %s: %w", key, err)
	}
	return nil
}
