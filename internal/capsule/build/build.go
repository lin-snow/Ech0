// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/lin-snow/ech0/internal/capsule"
	"github.com/lin-snow/ech0/template"
)

type Options struct {
	Output  string
	BaseURL string
}

type Result struct {
	Path     string
	Echoes   int
	Files    int
	Comments int
}

func Run(ctx context.Context, loaded *capsule.Loaded, opts Options) (*Result, error) {
	return run(ctx, loaded, opts, template.WebFS)
}

func run(ctx context.Context, loaded *capsule.Loaded, opts Options, assets fs.FS) (*Result, error) {
	if loaded == nil {
		return nil, fmt.Errorf("capsule is not loaded")
	}
	if strings.TrimSpace(opts.Output) == "" {
		return nil, fmt.Errorf("output directory is required")
	}
	dir := filepath.Clean(opts.Output)
	baseURL := NormalizeBaseURL(opts.BaseURL)

	if err := ensureEmptyDir(dir); err != nil {
		return nil, err
	}

	ds, err := bake(bakeInput{loaded: loaded, baseURL: baseURL, generatedAt: time.Now()})
	if err != nil {
		return nil, err
	}

	if err := copySPA(dir, assets); err != nil {
		return nil, err
	}
	if err := writeEntrypoints(dir, baseURL, assets); err != nil {
		return nil, err
	}

	payload, err := json.Marshal(ds)
	if err != nil {
		return nil, fmt.Errorf("marshal dataset: %w", err)
	}
	if err := writeFile(dir, "dataset.json", payload); err != nil {
		return nil, err
	}

	envelope, err := json.Marshal(resultEnvelope{Code: 1, Msg: "", Data: ds.Connect})
	if err != nil {
		return nil, fmt.Errorf("marshal connect payload: %w", err)
	}
	if err := writeFile(dir, "api/connect", envelope); err != nil {
		return nil, err
	}

	files, err := copyMedia(ctx, dir, loaded)
	if err != nil {
		return nil, err
	}

	l := newLinks(ds.Settings.ServerURL, baseURL)
	generatedAt := time.Unix(ds.GeneratedAt, 0).UTC()

	atom, err := renderAtom(ds, l, generatedAt)
	if err != nil {
		return nil, fmt.Errorf("render rss: %w", err)
	}
	if err := writeFile(dir, "rss.xml", []byte(atom)); err != nil {
		return nil, err
	}

	sitemap, err := renderSitemap(ds, l, generatedAt)
	if err != nil {
		return nil, err
	}
	if err := writeFile(dir, "sitemap.xml", sitemap); err != nil {
		return nil, err
	}

	return &Result{
		Path:     dir,
		Echoes:   len(ds.Echos),
		Files:    files,
		Comments: len(ds.Comments),
	}, nil
}

func NormalizeBaseURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "/"
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if !strings.HasSuffix(s, "/") {
		s += "/"
	}
	return s
}
