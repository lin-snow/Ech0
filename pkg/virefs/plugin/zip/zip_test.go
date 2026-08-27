// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package zip

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	virefs "github.com/lin-snow/ech0/pkg/virefs"
)

func makeZipBytes(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(w, body); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func seedLocalFS(t *testing.T, files map[string]string) *virefs.LocalFS {
	t.Helper()
	dir := t.TempDir()
	fs, err := virefs.NewLocalFS(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for k, v := range files {
		if err := fs.Put(ctx, k, strings.NewReader(v)); err != nil {
			t.Fatal(err)
		}
	}
	return fs
}

func readAll(t *testing.T, fsys virefs.FS, key string) string {
	t.Helper()
	rc, err := fsys.Get(context.Background(), key)
	if err != nil {
		t.Fatalf("Get %q: %v", key, err)
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("ReadAll %q: %v", key, err)
	}
	return string(data)
}

func TestFS_GetAndStat(t *testing.T) {
	data := makeZipBytes(t, map[string]string{
		"hello.txt":     "world",
		"sub/nested.md": "# Title",
	})
	zfs, err := NewFSFromBytes(data)
	if err != nil {
		t.Fatalf("NewFSFromBytes: %v", err)
	}

	ctx := context.Background()

	got := readAll(t, zfs, "hello.txt")
	if got != "world" {
		t.Fatalf("Get hello.txt = %q, want %q", got, "world")
	}

	got = readAll(t, zfs, "sub/nested.md")
	if got != "# Title" {
		t.Fatalf("Get sub/nested.md = %q, want %q", got, "# Title")
	}

	info, err := zfs.Stat(ctx, "hello.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("Stat size = %d, want 5", info.Size)
	}
}

func TestFS_StatContentType(t *testing.T) {
	data := makeZipBytes(t, map[string]string{
		"image.jpg": "fake-jpeg",
		"doc.txt":   "text content",
		"noext":     "no extension",
	})
	zfs, err := NewFSFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	info, err := zfs.Stat(ctx, "image.jpg")
	if err != nil {
		t.Fatalf("Stat image.jpg: %v", err)
	}
	if info.ContentType != "image/jpeg" {
		t.Fatalf("Stat image.jpg ContentType = %q, want %q", info.ContentType, "image/jpeg")
	}

	info, err = zfs.Stat(ctx, "doc.txt")
	if err != nil {
		t.Fatalf("Stat doc.txt: %v", err)
	}
	if !strings.HasPrefix(info.ContentType, "text/plain") {
		t.Fatalf("Stat doc.txt ContentType = %q, want prefix %q", info.ContentType, "text/plain")
	}

	info, err = zfs.Stat(ctx, "noext")
	if err != nil {
		t.Fatalf("Stat noext: %v", err)
	}
	if info.ContentType != "" {
		t.Fatalf("Stat noext ContentType = %q, want empty", info.ContentType)
	}
}

func TestFS_GetNotFound(t *testing.T) {
	data := makeZipBytes(t, map[string]string{"a.txt": "a"})
	zfs, err := NewFSFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}

	_, err = zfs.Get(context.Background(), "missing.txt")
	if !errors.Is(err, virefs.ErrNotFound) {
		t.Fatalf("Get missing = %v, want ErrNotFound", err)
	}
}

func TestFS_List(t *testing.T) {
	data := makeZipBytes(t, map[string]string{
		"a.txt":     "a",
		"dir/b.txt": "b",
		"dir/c.txt": "c",
		"other.txt": "o",
	})
	zfs, err := NewFSFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	res, err := zfs.List(ctx, "")
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(res.Files) != 3 {
		t.Fatalf("List root got %d, want 3 (2 files + 1 dir)", len(res.Files))
	}

	res, err = zfs.List(ctx, "dir")
	if err != nil {
		t.Fatalf("List dir: %v", err)
	}
	if len(res.Files) != 2 {
		t.Fatalf("List dir got %d, want 2", len(res.Files))
	}
}

func TestFS_PutDeleteNotSupported(t *testing.T) {
	data := makeZipBytes(t, map[string]string{"a.txt": "a"})
	zfs, err := NewFSFromBytes(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	if err := zfs.Put(ctx, "x.txt", strings.NewReader("x")); !errors.Is(err, virefs.ErrNotSupported) {
		t.Fatalf("Put = %v, want ErrNotSupported", err)
	}
	if err := zfs.Delete(ctx, "a.txt"); !errors.Is(err, virefs.ErrNotSupported) {
		t.Fatalf("Delete = %v, want ErrNotSupported", err)
	}
	if _, err := zfs.Access(ctx, "a.txt"); !errors.Is(err, virefs.ErrNotSupported) {
		t.Fatalf("Access = %v, want ErrNotSupported", err)
	}
}

func TestFS_EmptyArchive(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	zw.Close()

	zfs, err := NewFSFromBytes(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	res, err := zfs.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(res.Files) != 0 {
		t.Fatalf("List empty got %d, want 0", len(res.Files))
	}
}

func TestPack(t *testing.T) {
	src := seedLocalFS(t, map[string]string{
		"a.txt":     "alpha",
		"dir/b.txt": "beta",
	})
	ctx := context.Background()

	var buf bytes.Buffer
	if err := Pack(ctx, src, []string{"a.txt", "dir/b.txt"}, &buf); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	zfs, err := NewFSFromBytes(buf.Bytes())
	if err != nil {
		t.Fatalf("open packed zip: %v", err)
	}

	if got := readAll(t, zfs, "a.txt"); got != "alpha" {
		t.Fatalf("packed a.txt = %q, want %q", got, "alpha")
	}
	if got := readAll(t, zfs, "dir/b.txt"); got != "beta" {
		t.Fatalf("packed dir/b.txt = %q, want %q", got, "beta")
	}
}

func TestPack_ContextCancelled(t *testing.T) {
	src := seedLocalFS(t, map[string]string{"a.txt": "a"})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err := Pack(ctx, src, []string{"a.txt"}, &buf)
	if err == nil {
		t.Fatal("Pack with cancelled context should fail")
	}
}

func TestUnpack(t *testing.T) {
	data := makeZipBytes(t, map[string]string{
		"x.txt":     "X",
		"sub/y.txt": "Y",
	})
	dst, err := virefs.NewLocalFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	r := bytes.NewReader(data)

	if err := Unpack(ctx, r, int64(len(data)), dst, "out"); err != nil {
		t.Fatalf("Unpack: %v", err)
	}

	if got := readAll(t, dst, "out/x.txt"); got != "X" {
		t.Fatalf("unpacked x.txt = %q, want %q", got, "X")
	}
	if got := readAll(t, dst, "out/sub/y.txt"); got != "Y" {
		t.Fatalf("unpacked sub/y.txt = %q, want %q", got, "Y")
	}
}

func TestUnpack_NoPrefix(t *testing.T) {
	data := makeZipBytes(t, map[string]string{"file.txt": "content"})
	dst, err := virefs.NewLocalFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	r := bytes.NewReader(data)

	if err := Unpack(ctx, r, int64(len(data)), dst, ""); err != nil {
		t.Fatalf("Unpack no prefix: %v", err)
	}

	if got := readAll(t, dst, "file.txt"); got != "content" {
		t.Fatalf("unpacked file.txt = %q, want %q", got, "content")
	}
}

func TestUnpack_ContextCancelled(t *testing.T) {
	data := makeZipBytes(t, map[string]string{"a.txt": "a"})
	dst, err := virefs.NewLocalFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := bytes.NewReader(data)

	err = Unpack(ctx, r, int64(len(data)), dst, "")
	if err == nil {
		t.Fatal("Unpack with cancelled context should fail")
	}
}

func TestPackUnpack_RoundTrip(t *testing.T) {
	files := map[string]string{
		"readme.md":          "# Hello",
		"src/main.go":        "package main",
		"src/util/helper.go": "package util",
		"data/中文.txt":        "你好世界",
	}
	src := seedLocalFS(t, files)
	ctx := context.Background()

	keys := make([]string, 0, len(files))
	for k := range files {
		keys = append(keys, k)
	}

	var buf bytes.Buffer
	if err := Pack(ctx, src, keys, &buf); err != nil {
		t.Fatalf("Pack: %v", err)
	}

	dst, err := virefs.NewLocalFS(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	r := bytes.NewReader(data)
	if err := Unpack(ctx, r, int64(len(data)), dst, ""); err != nil {
		t.Fatalf("Unpack: %v", err)
	}

	for k, want := range files {
		got := readAll(t, dst, k)
		if got != want {
			t.Errorf("round-trip %q = %q, want %q", k, got, want)
		}
	}
}
