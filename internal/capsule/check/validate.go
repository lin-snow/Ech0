// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package check

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/internal/storage"
	uuidUtil "github.com/lin-snow/ech0/internal/util/uuid"
)

// echoFileMode 是回写 Echo 文件的权限；与导出侧产出的文件保持一致。
const echoFileMode = 0o644

// validateManifest 校验 ech0.yaml（spec §3）。清单是胶囊的身份证明：
// 它不可读或版本不认识时，后面所有字段校验都失去意义，故只报这一条。
func validateManifest(r *Report, loaded *capsule.Loaded, site capsule.Site) {
	p := capsule.ManifestPath

	if loaded.ManifestErr != nil {
		r.errorf(p, "", "%v", loaded.ManifestErr)
	}
	for _, u := range loaded.ManifestUnknown {
		r.warnf(p, "", "unknown field ignored: %s", u)
	}
	if loaded.Manifest == nil {
		return
	}

	switch v := loaded.Manifest.SchemaVersion; {
	case v <= 0:
		r.errorf(p, "schema_version", "schema_version is required")
	case v > capsule.SchemaVersion:
		// 高于自身支持的版本必须拒绝（spec §8）：语义可能已变更，
		// 「尽力而为地解析」只会静默产出错数据。
		r.errorf(p, "schema_version", "unsupported schema_version %d, this build supports up to %d", v, capsule.SchemaVersion)
	}

	if loaded.Manifest.Owner.Username == "" {
		r.errorf(p, "owner.username", "owner.username is required")
	}

	if site.CustomJS != "" {
		r.warnf(p, "site.custom_js", "custom_js is not empty: consuming this capsule means running third-party code")
	}
	if site.CustomCSS != "" {
		r.warnf(p, "site.custom_css", "custom_css is not empty: consuming this capsule means applying third-party styles")
	}
	if marker := instanceMarker(site.ServerLogo, site.ServerURL); marker != "" {
		r.warnf(p, "site.server_logo", "embeds source instance URL (%s), the link may break after migration", marker)
	}
}

// validateEchoes 校验全部 Echo 内容文件（spec §4），返回胶囊内的 Echo id 集合
// （评论孤儿判定用）与被引用的媒体路径集合（悬空媒体判定用）。
// 只有 --fix 的写回失败会中断校验：那是 I/O 故障，继续跑只会掩盖它。
func validateEchoes(r *Report, loaded *capsule.Loaded, opts Options, serverURL string) (ids map[string]struct{}, referenced map[string]struct{}, err error) {
	ids = make(map[string]struct{}, len(loaded.Echoes))
	referenced = make(map[string]struct{})
	firstSeen := make(map[string]string, len(loaded.Echoes))

	for i := range loaded.Echoes {
		e := &loaded.Echoes[i]
		if e.Err != nil {
			r.errorf(e.Path, "", "%v", e.Err)
			continue
		}
		for _, u := range e.Unknown {
			r.warnf(e.Path, "", "unknown field ignored: %s", u)
		}
		doc := e.Doc
		if doc == nil {
			continue
		}

		if doc.ID == "" && opts.Fix {
			if err := fixEchoID(r, loaded, e); err != nil {
				return nil, nil, err
			}
		}

		switch {
		case doc.ID == "":
			r.errorf(e.Path, "id", "id is required, run `ech0 check --fix` to generate one")
		case !uuidUtil.IsValid(doc.ID):
			r.errorf(e.Path, "id", "id %q is not a valid UUID", doc.ID)
		default:
			if first, dup := firstSeen[doc.ID]; dup {
				// id 是幂等键与 permalink，重复即无法区分两条 Echo。
				r.errorf(e.Path, "id", "duplicate id %s, already used by %s", doc.ID, first)
			} else {
				firstSeen[doc.ID] = e.Path
			}
			ids[doc.ID] = struct{}{}
		}

		if doc.CreatedAt == "" {
			r.errorf(e.Path, "created_at", "created_at is required")
		} else if _, perr := capsule.ParseTime(doc.CreatedAt); perr != nil {
			r.errorf(e.Path, "created_at", "%v", perr)
		}

		// 表现层枚举不认得的取值只警告，不阻断（spec §7）：内容本身完好，消费者
		// 退回默认值即可。活实例的写路径本来就是这么干的（service/echo 把未知
		// layout 归一成 waterfall），校验器没有理由比它描述的系统更严格。
		if doc.Layout != "" {
			if _, ok := capsule.ValidLayouts[doc.Layout]; !ok {
				r.warnf(e.Path, "layout", "unknown layout %q, consumers fall back to %q", doc.Layout, capsule.DefaultLayout)
			}
		}

		validateExtension(r, e.Path, doc.Extension, serverURL)
		validateFiles(r, loaded, e.Path, doc.Files, referenced)

		if marker := instanceMarker(doc.Content, serverURL); marker != "" {
			r.warnf(e.Path, "content", "embeds source instance URL (%s), the link may break after migration", marker)
		}
	}
	return ids, referenced, nil
}

// validateExtension 校验 extension 块（spec §4.2）：type 与 payload 同时必须，
// payload 内部结构随 type 而异，规格不约束，这里只扫内嵌实例 URL。
func validateExtension(r *Report, echoPath string, ext *capsule.Extension, serverURL string) {
	if ext == nil {
		return
	}
	if ext.Type == "" {
		r.errorf(echoPath, "extension.type", "extension.type is required when extension is present")
	} else if _, ok := capsule.ValidExtensionTypes[ext.Type]; !ok {
		// 同 layout/category：类型不认得只是渲染不出来，正文与 payload 都还在，
		// 不该让整个胶囊无法导入。缺失 type 才是硬错——那时 payload 无从解释。
		r.warnf(echoPath, "extension.type", "unknown extension type %q, consumers skip rendering it", ext.Type)
	}
	if ext.Payload == nil {
		r.errorf(echoPath, "extension.payload", "extension.payload is required when extension is present")
		return
	}
	for _, hit := range scanPayload(ext.Payload, serverURL) {
		r.warnf(echoPath, hit.field, "embeds source instance URL (%s), the link may break after migration", hit.marker)
	}
}

// validateFiles 校验 files[]（spec §4.2 / §6）。托管条目的字节位置是
// MediaPath(key) 的纯函数结果，胶囊不存路径，所以「字节在不在」只能这样比对。
func validateFiles(r *Report, loaded *capsule.Loaded, echoPath string, files []capsule.FileRef, referenced map[string]struct{}) {
	for i := range files {
		f := files[i]
		field := fmt.Sprintf("files[%d]", i)

		switch {
		case f.Key != "" && f.URL != "":
			r.errorf(echoPath, field, "key and url are mutually exclusive")
		case f.Key == "" && f.URL == "":
			r.errorf(echoPath, field, "one of key or url is required")
		}

		if f.Category != "" {
			if _, ok := capsule.ValidCategories[f.Category]; !ok {
				r.warnf(echoPath, field+".category", "unknown category %q, consumers fall back to %q", f.Category, string(storage.CategoryFile))
			}
		}

		if f.Key == "" {
			continue
		}
		if err := capsule.ValidateKey(f.Key); err != nil {
			r.errorf(echoPath, field+".key", "%v", err)
			continue
		}

		media := capsule.MediaPath(f.Key)
		// 即使字节缺失也算「被引用」：悬空判定看的是引用意图，
		// 缺字节已经单独报了 error，不该再牵连别的文件。
		referenced[media] = struct{}{}

		size, ok := loaded.MediaPaths[media]
		if !ok {
			r.errorf(echoPath, field+".key", "capsule is not self-contained: %s is missing for key %q", media, f.Key)
			continue
		}
		if f.Size > 0 && f.Size != size {
			r.warnf(echoPath, field+".size", "declared size %d does not match %s (%d bytes on disk)", f.Size, media, size)
		}
	}
}

// validateComments 校验 comments.yaml（spec §5）。
func validateComments(r *Report, loaded *capsule.Loaded, echoIDs map[string]struct{}) {
	p := capsule.CommentsPath
	if !loaded.HasComments {
		return
	}
	if loaded.CommentsErr != nil {
		r.errorf(p, "", "%v", loaded.CommentsErr)
		return
	}
	for _, u := range loaded.CommentsUnknown {
		r.warnf(p, "", "unknown field ignored: %s", u)
	}

	// 禁止字段必须从原始键检出：Comment 结构体压根没有这些字段，
	// 走结构体解码只会把隐私泄露降级成「未知字段」警告。
	raw, err := capsule.RawComments(loaded.CommentsRaw)
	if err != nil {
		r.errorf(p, "", "parse %s: %v", p, err)
	}
	for i, item := range raw {
		for _, forbidden := range capsule.ForbiddenCommentFields {
			if _, ok := item[forbidden]; ok {
				r.errorf(p, fmt.Sprintf("comments[%d].%s", i, forbidden),
					"forbidden field %q must not appear in a capsule", forbidden)
			}
		}
	}

	if loaded.Comments == nil {
		return
	}
	firstSeen := make(map[string]int, len(loaded.Comments.Comments))
	for i := range loaded.Comments.Comments {
		c := loaded.Comments.Comments[i]
		at := func(name string) string { return fmt.Sprintf("comments[%d].%s", i, name) }

		if c.ID == "" {
			r.errorf(p, at("id"), "id is required")
		} else if first, dup := firstSeen[c.ID]; dup {
			r.errorf(p, at("id"), "duplicate id %s, already used by comments[%d]", c.ID, first)
		} else {
			firstSeen[c.ID] = i
		}

		if c.EchoID == "" {
			r.errorf(p, at("echo_id"), "echo_id is required")
		} else if _, ok := echoIDs[c.EchoID]; !ok {
			// 孤儿评论只是警告：胶囊可能只导出了部分 Echo（如排除私密）。
			r.warnf(p, at("echo_id"), "orphan comment: echo %s is not in this capsule", c.EchoID)
		}

		if c.Nickname == "" {
			r.errorf(p, at("nickname"), "nickname is required")
		}
		if c.Content == "" {
			r.errorf(p, at("content"), "content is required")
		}
		if c.CreatedAt == "" {
			r.errorf(p, at("created_at"), "created_at is required")
		} else if _, perr := capsule.ParseTime(c.CreatedAt); perr != nil {
			r.errorf(p, at("created_at"), "%v", perr)
		}

		if c.Status != "" && c.Status != capsule.DefaultCommentStatus {
			r.warnf(p, at("status"), "status %q is not %q: a capsule should only carry approved comments",
				c.Status, capsule.DefaultCommentStatus)
		}
	}
}

// validateMedia 找出悬空媒体（spec §6）：合法但没人引用，通常是导出侧
// 多拷了东西或引用被删掉了。
func validateMedia(r *Report, loaded *capsule.Loaded, referenced map[string]struct{}, site capsule.Site) {
	logo := logoMedia(site.ServerLogo, loaded.MediaPaths)
	for _, p := range sortedMediaPaths(loaded.MediaPaths) {
		if _, ok := referenced[p]; ok {
			continue
		}
		if p == logo {
			continue
		}
		r.warnf(p, "", "dangling media: not referenced by any echo file or site.server_logo")
	}
}

// validatePaths 处理与具体文件内容无关的路径级规则（spec §2）。
func validatePaths(r *Report, loaded *capsule.Loaded) {
	for i := range loaded.Echoes {
		reportTraversal(r, loaded.Echoes[i].Path)
	}
	for _, p := range sortedMediaPaths(loaded.MediaPaths) {
		reportTraversal(r, p)
	}
	for _, p := range loaded.UnknownPaths {
		reportTraversal(r, p)
		// 未知路径必须被忽略而非拒绝（spec §8），但用户有权知道它们的存在。
		r.warnf(p, "", "unknown path ignored: not defined by the capsule spec")
	}
}

func reportTraversal(r *Report, p string) {
	if hasTraversal(p) {
		r.errorf(p, "", "path traversal (\"..\") is forbidden inside a capsule")
	}
}

func hasTraversal(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// fixEchoID 补全缺失的 id 并回写整个文件（spec §7 唯一的自动修复项）。
// 用 EncodeEcho 重写而非插一行 YAML：正文逐字保留由编码器保证，
// 手工拼字符串迟早会在 CRLF / 无正文这类边角上出错。
func fixEchoID(r *Report, loaded *capsule.Loaded, e *capsule.LoadedEcho) error {
	id := uuidUtil.MustNewV7()
	e.Doc.ID = id

	data, err := capsule.EncodeEcho(e.Doc)
	if err != nil {
		return fmt.Errorf("capsule check: re-encode %s: %w", e.Path, err)
	}
	// 仅目录形态会走到这里（Validate 已挡掉 zip）。
	target := filepath.Join(loaded.Source.Path, filepath.FromSlash(e.Path))
	if err := os.WriteFile(target, data, echoFileMode); err != nil {
		return fmt.Errorf("capsule check: write back %s: %w", e.Path, err)
	}

	// 不重命名文件：文件名里的 id 前缀纯属浏览友好，消费者禁止从中解析语义
	// （spec §4.1），重命名只会制造无谓的路径变更。
	r.Fixed = append(r.Fixed, fmt.Sprintf("%s: generated id %s", e.Path, id))
	return nil
}

// logoMedia 把 site.server_logo 反查成胶囊内的媒体路径。logo 原样导出
// （可能是绝对 URL、也可能是 /api/files/... 相对引用），唯一稳定的锚点
// 是文件名，所以按 basename 比对。
func logoMedia(logo string, media map[string]int64) string {
	if logo == "" || len(media) == 0 {
		return ""
	}
	base := logo
	if u, err := url.Parse(logo); err == nil && u.Path != "" {
		base = u.Path
	}
	base = path.Base(base)
	if base == "" || base == "." || base == "/" {
		return ""
	}
	for p := range media {
		if path.Base(p) == base {
			return p
		}
	}
	return ""
}

func sortedMediaPaths(media map[string]int64) []string {
	paths := make([]string, 0, len(media))
	for p := range media {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
