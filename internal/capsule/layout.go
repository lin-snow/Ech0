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

var mediaSchema = storage.NewFileSchema()

func MediaPath(key string) string {
	return FilesDir + "/" + mediaSchema.Resolve(key)
}

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

func EchoPath(id string, createdAt time.Time) string {
	utc := createdAt.UTC()
	short := strings.ReplaceAll(id, "-", "")
	if len(short) > 8 {
		short = short[len(short)-8:]
	}
	return fmt.Sprintf("%s/%04d/%s-%s.md", EchoesDir, utc.Year(), utc.Format("2006-01-02"), short)
}

func IsEchoPath(p string) bool {
	return strings.HasPrefix(p, EchoesDir+"/") && strings.HasSuffix(p, ".md")
}
