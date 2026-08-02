// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/storage"
	"github.com/lin-snow/ech0/pkg/virefs"
)

// mediaSchema 是媒体目录的路由表，与实例本地存储 DataRoot 用的是同一份
// （internal/storage.NewFileSchema）。胶囊内 files/ 就是它的原样 mirror，
// 所以「胶囊里的位置」永远等于「实例里的位置」，无需在胶囊中存路径。
var mediaSchema = storage.NewFileSchema()

// MediaPath 由存储键派生胶囊内的媒体相对路径（spec §6）：
// files/ + Resolve(key)，例如 "a.png" -> "files/images/a.png"。
func MediaPath(key string) string {
	return FilesDir + "/" + mediaSchema.Resolve(key)
}

// ValidateKey 检查存储键是否满足胶囊约束：非空、不含目录分隔符、无路径穿越。
// File.Key 在库中的约定就是扁平键，位置全交给 Resolve。
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty key")
	}
	if strings.ContainsRune(key, '/') || strings.ContainsRune(key, '\\') {
		return fmt.Errorf("key must be flat (no path separator): %q", key)
	}
	if _, err := virefs.CleanKey(key); err != nil {
		return err
	}
	if key != path.Clean(key) {
		return fmt.Errorf("key must be already clean: %q", key)
	}
	return nil
}

// EchoPath 由 id 与创建时刻派生 Echo 内容文件的相对路径（spec §4.1）：
// echoes/<YYYY>/<YYYY-MM-DD>-<id 前 8 位>.md。命名仅为浏览友好，
// 消费者一律以 frontmatter 为准。
func EchoPath(id string, createdAt time.Time) string {
	utc := createdAt.UTC()
	short := id
	if len(short) > 8 {
		short = short[:8]
	}
	return fmt.Sprintf("%s/%04d/%s-%s.md", EchoesDir, utc.Year(), utc.Format("2006-01-02"), short)
}

// IsEchoPath 报告某个胶囊内路径是否应被当作 Echo 内容文件读取。
func IsEchoPath(p string) bool {
	return strings.HasPrefix(p, EchoesDir+"/") && strings.HasSuffix(p, ".md")
}
