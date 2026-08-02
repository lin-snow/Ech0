// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package export

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	echoModel "github.com/lin-snow/ech0/internal/model/echo"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/virefs"
)

// writeCapsule 把快照写进暂存 FS，返回写入的相对路径（按写入顺序），供 zip 打包
// 复用为条目名。
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

	// 空评论不落文件：comments.yaml 是可选的（spec §5），一个空列表只会给消费者
	// 和 diff 添噪。
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
			FavCount:  echo.FavCount,
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

// uniquePath 消解文件名撞车：路径只取 id 前 8 位，同日两条 id 前缀相同就会重名。
// 命名本就只为浏览友好（身份以 frontmatter 的 id 为准），加序号即可。
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

// fileRefs 把 Echo 的媒体关联转成 frontmatter 引用。三种存储形态在胶囊里被归一化
// 成两种：external 用 url 透传，local/object 统一用 key——托管文件的 URL 是运行时
// 拓扑（AfterFind 按当前配置重算），带进胶囊只会在迁移后指回原实例（spec §11.1）。
func fileRefs(links []fileModel.EchoFile) []capsule.FileRef {
	if len(links) == 0 {
		return nil
	}
	refs := make([]capsule.FileRef, 0, len(links))
	for i := range links {
		refs = append(refs, fileRef(links[i].File))
	}
	return refs
}

// fileRef 把一行 File 转成胶囊引用。storage_type/provider/bucket/user_id/created_at
// 不入胶囊：它们是运行时拓扑或行元数据，由目标实例按自己的配置重建。
func fileRef(file fileModel.File) capsule.FileRef {
	ref := capsule.FileRef{
		ID:          file.ID,
		Category:    file.Category,
		Name:        file.Name,
		ContentType: file.ContentType,
		Size:        file.Size,
		Width:       file.Width,
		Height:      file.Height,
	}
	if storage.NormalizeStorageType(file.StorageType) == storage.StorageTypeExternal {
		ref.URL = file.URL
	} else {
		ref.Key = file.Key
	}
	return ref
}

// unattachedRefs 收集没挂在任何导出 Echo 上的文件行（站点 logo、上传后没用上的
// 附件）。它们的字节本来就随记录驱动导出进了胶囊，但 frontmatter 只能表达
// 「挂在某条 Echo 上的文件」，元数据没有落脚点——导入侧因此无法还原这些行，
// 最直接的后果是搬家之后 site.server_logo 变成死链。清单里的 files 块就是它们的位置。
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

// mediaFailure 记录一条取不回字节的托管文件。
type mediaFailure struct {
	key         string
	storageType string
	err         error
}

// writeMedia 把托管文件的字节搬进胶囊。自包含是硬承诺（spec §11.2）：任何一条取不
// 回都不能静默略过，但也不在第一条就中断——一次跑完再把完整清单交给用户，比让他
// 修一条再跑一次快得多。
func writeMedia(ctx context.Context, deps Deps, stage virefs.FS, data *dataset) ([]string, error) {
	keys := make([]string, 0, len(data.files))
	var failures []mediaFailure

	for i := range data.files {
		file := &data.files[i]
		if storage.NormalizeStorageType(file.StorageType) == storage.StorageTypeExternal {
			continue // 外链的权威表示是 URL，字节不属于本实例
		}
		if err := capsule.ValidateKey(file.Key); err != nil {
			failures = append(failures, mediaFailure{key: file.Key, storageType: file.StorageType, err: err})
			continue
		}

		// 传扁平 key：selector 的底层 FS 已挂 schema.Resolve，这里再 Resolve 一次会
		// 变成 images/images/x.png。
		reader, err := deps.Selector.Get(ctx, storage.StorageType(file.StorageType), file.Key)
		if err != nil {
			failures = append(failures, mediaFailure{key: file.Key, storageType: file.StorageType, err: err})
			continue
		}
		key := capsule.MediaPath(file.Key)
		err = stage.Put(ctx, key, reader)
		_ = reader.Close()
		if err != nil {
			return nil, fmt.Errorf("capsule export: write %s: %w", key, err)
		}
		keys = append(keys, key)
	}

	if len(failures) > 0 {
		return nil, unreadableMediaError(failures)
	}
	return keys, nil
}

func unreadableMediaError(failures []mediaFailure) error {
	var b strings.Builder
	fmt.Fprintf(&b, "capsule export: %d managed file(s) unreadable, capsule would not be self-contained:", len(failures))
	for _, f := range failures {
		fmt.Fprintf(&b, "\n  - key=%q storage=%s: %v", f.key, f.storageType, f.err)
	}
	return errors.New(b.String())
}

func put(ctx context.Context, stage virefs.FS, key string, body []byte) error {
	if err := stage.Put(ctx, key, bytes.NewReader(body)); err != nil {
		return fmt.Errorf("capsule export: write %s: %w", key, err)
	}
	return nil
}
