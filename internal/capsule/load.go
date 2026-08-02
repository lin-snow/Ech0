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

// Source 是一个已打开的胶囊：目录形态与 .zip 形态在这里被抹平成同一个
// virefs.FS 视图，check / import / build 三个消费者共用。
type Source struct {
	// Path 是用户给出的胶囊位置，仅用于报告。
	Path string
	// FS 以胶囊根为根，键即 spec §2 的相对路径。
	FS virefs.FS

	closer io.Closer
}

// Open 打开目录或 zip 形态的胶囊。调用方必须 Close。
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

// Close 释放底层句柄（zip 形态才有）。
func (s *Source) Close() error {
	if s.closer == nil {
		return nil
	}
	return s.closer.Close()
}

// ReadFile 读取胶囊内一个相对路径的全部字节。
func (s *Source) ReadFile(ctx context.Context, key string) ([]byte, error) {
	rc, err := s.FS.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rc.Close() }()
	return io.ReadAll(rc)
}

// LoadedEcho 是一个 Echo 内容文件的解析结果。Err 非空表示该文件自身解析失败，
// 其余字段无意义——加载不因单文件失败而中断，好让 check 一次报全。
type LoadedEcho struct {
	Path    string
	Doc     *EchoDoc
	Unknown []string
	Err     error
}

// Loaded 是一个胶囊的完整解析结果（不含媒体字节）。
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

	// MediaPaths 是 files/ 下实际存在的相对路径集合（含 files/ 前缀）。
	MediaPaths map[string]int64
	// UnknownPaths 是规格未定义的顶层条目——消费者必须忽略，check 告警。
	UnknownPaths []string
}

// Load 解析整个胶囊。除「胶囊根不可读」外不返回 error：单点失败记录在对应
// 字段里，由 check 统一分级上报。
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

// scanTree 遍历胶囊，收集 Echo 文件、媒体文件与未知路径。
func (l *Loaded) scanTree(ctx context.Context, src *Source) error {
	return virefs.Walk(ctx, src.FS, "", func(key string, info virefs.FileInfo, err error) error {
		if err != nil {
			// 根不可读才致命；子目录读失败降级成未知路径。
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

// RawComments 解析 comments.yaml 的原始键，用于检出 spec §5 的禁止字段。
// 单独走一遍宽松解码，避免把「禁止字段」和「未知字段」混为一谈。
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

// EchoDir 返回某个 Echo 文件所属的年份目录，供报告使用。
func EchoDir(p string) string { return path.Dir(p) }
