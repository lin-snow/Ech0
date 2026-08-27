// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/internal/capsule"
)

const spaRoot = "dist"

const (
	indexFile    = "index.html"
	notFoundFile = "404.html"
)

func ensureEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	switch {
	case err == nil:
		if len(entries) > 0 {
			return fmt.Errorf("output directory %s is not empty", dir)
		}
		return nil
	case os.IsNotExist(err):
		return os.MkdirAll(dir, 0o755)
	default:
		return fmt.Errorf("inspect output directory %s: %w", dir, err)
	}
}

func writeFile(dir, rel string, data []byte) error {
	dest := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", rel, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", rel, err)
	}
	return nil
}

func copySPA(dir string, assets fs.FS) error {
	return fs.WalkDir(assets, spaRoot, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, readErr := fs.ReadFile(assets, p)
		if readErr != nil {
			return fmt.Errorf("read embedded %s: %w", p, readErr)
		}
		rel := strings.TrimPrefix(p, spaRoot+"/")
		return writeFile(dir, rel, data)
	})
}

func copyMedia(ctx context.Context, dir string, loaded *capsule.Loaded) (int, error) {
	keys := make([]string, 0, len(loaded.MediaPaths))
	for p := range loaded.MediaPaths {
		keys = append(keys, p)
	}
	sort.Strings(keys)

	prefix := capsule.FilesDir + "/"
	n := 0
	for _, p := range keys {
		rel := strings.TrimPrefix(p, prefix)
		if rel == p || rel == "" || rel != path.Clean(rel) || strings.HasPrefix(rel, "../") {
			return n, fmt.Errorf("unexpected media path %q in capsule", p)
		}
		data, err := loaded.Source.ReadFile(ctx, p)
		if err != nil {
			return n, fmt.Errorf("read media %s: %w", p, err)
		}
		if err := writeFile(dir, "api/files/"+rel, data); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

var rootPathAttr = regexp.MustCompile(`\b(href|src)="/([^/"])`)

var rootPathAttrBare = regexp.MustCompile(`\b(href|src)="/"`)

var firstScriptTag = regexp.MustCompile(`<script[\s>]`)

func renderIndex(raw []byte, baseURL string) []byte {
	html := string(raw)

	if baseURL != "/" {
		html = rootPathAttrBare.ReplaceAllString(html, `$1="`+baseURL+`"`)
		html = rootPathAttr.ReplaceAllString(html, `$1="`+baseURL+`$2`)
	}

	html = strings.Replace(html, `href="`+baseURL+`rss"`, `href="`+baseURL+`rss.xml"`, 1)

	snippet := fmt.Sprintf(
		"<script>window.__ECH0_STATIC__=true;window.__ECH0_STATIC_BASE__=%s;</script>\n    ",
		jsonString(baseURL),
	)

	if loc := firstScriptTag.FindStringIndex(html); loc != nil {
		return []byte(html[:loc[0]] + snippet + html[loc[0]:])
	}
	if i := strings.Index(html, "</head>"); i >= 0 {
		return []byte(html[:i] + snippet + html[i:])
	}
	return []byte(snippet + html)
}

func jsonString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '<':
			b.WriteString(`\u003c`)
		case '>':
			b.WriteString(`\u003e`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func writeEntrypoints(dir, baseURL string, assets fs.FS) error {
	raw, err := fs.ReadFile(assets, spaRoot+"/"+indexFile)
	if err != nil {
		return fmt.Errorf("read embedded %s: %w", indexFile, err)
	}
	rendered := renderIndex(raw, baseURL)
	if err := writeFile(dir, indexFile, rendered); err != nil {
		return err
	}
	return writeFile(dir, notFoundFile, rendered)
}
