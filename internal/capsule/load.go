// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package capsule

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/lin-snow/ech0/pkg/virefs"
	vizip "github.com/lin-snow/ech0/pkg/virefs/plugin/zip"
)

type Source struct {
	Path string
	FS   virefs.FS

	closer io.Closer
}

func Open(location string) (*Source, error) {
	info, err := os.Stat(location)
	if err != nil {
		return nil, fmt.Errorf("open capsule %q: %w", location, err)
	}
	if info.IsDir() {
		fsys, err := virefs.NewLocalFS(location)
		if err != nil {
			return nil, fmt.Errorf("open capsule %q: %w", location, err)
		}
		return &Source{Path: location, FS: fsys}, nil
	}
	zfs, err := vizip.OpenFS(location)
	if err != nil {
		return nil, fmt.Errorf("open capsule archive %q: %w", location, err)
	}
	return &Source{Path: location, FS: zfs, closer: zfs}, nil
}

func (s *Source) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

func (s *Source) ReadFile(ctx context.Context, key string) ([]byte, error) {
	rc, err := s.FS.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
}

type LoadedEcho struct {
	Path    string
	Doc     *EchoDoc
	Unknown []string
	Err     error
}

type Loaded struct {
	Source *Source

	Manifest        *Manifest
	ManifestUnknown []string
	ManifestErr     error

	Echoes []LoadedEcho

	Comments        *CommentsDoc
	CommentsRaw     []byte
	CommentsUnknown []string
	CommentsErr     error
	HasComments     bool

	MediaPaths   map[string]int64
	UnknownPaths []string
}

func Load(ctx context.Context, src *Source) (*Loaded, error) {
	l := &Loaded{Source: src, MediaPaths: make(map[string]int64)}

	if data, err := src.ReadFile(ctx, ManifestPath); err != nil {
		l.ManifestErr = fmt.Errorf("read %s: %w", ManifestPath, err)
	} else {
		m := &Manifest{}
		unknown, decErr := DecodeYAML(data, m)
		l.ManifestUnknown = unknown
		if decErr != nil {
			l.ManifestErr = fmt.Errorf("parse %s: %w", ManifestPath, decErr)
		} else {
			l.Manifest = m
		}
	}

	if data, err := src.ReadFile(ctx, CommentsPath); err == nil {
		l.HasComments = true
		l.CommentsRaw = data
		doc := &CommentsDoc{}
		unknown, decErr := DecodeYAML(data, doc)
		l.CommentsUnknown = unknown
		if decErr != nil {
			l.CommentsErr = fmt.Errorf("parse %s: %w", CommentsPath, decErr)
		} else {
			l.Comments = doc
		}
	} else if !errors.Is(err, virefs.ErrNotFound) {
		l.HasComments = true
		l.CommentsErr = fmt.Errorf("read %s: %w", CommentsPath, err)
	}

	if err := l.scanTree(ctx, src); err != nil {
		return nil, err
	}
	sort.Slice(l.Echoes, func(i, j int) bool { return l.Echoes[i].Path < l.Echoes[j].Path })
	sort.Strings(l.UnknownPaths)
	return l, nil
}

func (l *Loaded) scanTree(ctx context.Context, src *Source) error {
	return virefs.Walk(ctx, src.FS, "", func(key string, info virefs.FileInfo, err error) error {
		if err != nil {
			if key == "" {
				return err
			}
			l.UnknownPaths = append(l.UnknownPaths, key)
			return nil
		}
		if info.IsDir {
			return nil
		}
		switch {
		case key == ManifestPath || key == CommentsPath:
			return nil
		case strings.HasPrefix(key, FilesDir+"/"):
			l.MediaPaths[key] = info.Size
			return nil
		case IsEchoPath(key):
			l.Echoes = append(l.Echoes, loadEcho(ctx, src, key))
			return nil
		default:
			l.UnknownPaths = append(l.UnknownPaths, key)
			return nil
		}
	})
}

func loadEcho(ctx context.Context, src *Source, key string) LoadedEcho {
	data, err := src.ReadFile(ctx, key)
	if err != nil {
		return LoadedEcho{Path: key, Err: fmt.Errorf("read: %w", err)}
	}
	doc, unknown, err := DecodeEcho(data)
	if err != nil {
		return LoadedEcho{Path: key, Unknown: unknown, Err: err}
	}
	return LoadedEcho{Path: key, Doc: doc, Unknown: unknown}
}

func RawComments(data []byte) ([]map[string]any, error) {
	var raw struct {
		Comments []map[string]any `yaml:"comments"`
	}
	dec := newLaxDecoder(data)
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	return raw.Comments, nil
}

func EchoDir(p string) string { return path.Dir(p) }
