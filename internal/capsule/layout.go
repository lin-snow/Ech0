// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"fmt"
	"strings"
	"time"

	"github.com/lin-snow/ech0/pkg/virefs"
)

// MediaPath is where a capsule keeps one file's bytes: files/<key>, mirroring
// the object key the source instance used. Because keys are year/month-nested
// now, the media tree is nested too.
func MediaPath(key string) string {
	return FilesDir + "/" + key
}

// ValidateKey accepts a clean relative object key. The key names bytes inside
// the capsule, so it may be nested; it may not escape the capsule.
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty key")
	}
	if strings.ContainsRune(key, '\\') {
		return fmt.Errorf("key must use forward slashes: %q", key)
	}
	if strings.HasPrefix(key, "/") {
		return fmt.Errorf("key must be relative: %q", key)
	}
	cleaned, err := virefs.CleanKey(key)
	if err != nil {
		return err
	}
	if cleaned != key {
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
