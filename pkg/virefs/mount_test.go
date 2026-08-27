// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"testing"
)

func TestMountTable_Routing(t *testing.T) {
	dir := t.TempDir()
	local := mustNewLocalFS(t, dir)
	fake := newFakeS3()
	obj := NewObjectFS(fake, "bucket")

	mt := NewMountTable()
	if err := mt.Mount("local", local); err != nil {
		t.Fatal(err)
	}
	if err := mt.Mount("s3", obj); err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	_ = mt.Put(ctx, "local/greet.txt", strings.NewReader("hi"))
	_ = mt.Put(ctx, "s3/data.bin", strings.NewReader("01"))

	rc, err := mt.Get(ctx, "local/greet.txt")
	if err != nil {
		t.Fatalf("Get local: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hi" {
		t.Fatalf("Get local = %q, want %q", data, "hi")
	}

	rc, err = mt.Get(ctx, "s3/data.bin")
	if err != nil {
		t.Fatalf("Get s3: %v", err)
	}
	data, _ = io.ReadAll(rc)
	rc.Close()
	if string(data) != "01" {
		t.Fatalf("Get s3 = %q, want %q", data, "01")
	}
}

func TestMountTable_UnmountedPrefix(t *testing.T) {
	mt := NewMountTable()
	ctx := context.Background()
	_, err := mt.Get(ctx, "unknown/file.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("unmounted prefix error = %v, want ErrNotFound", err)
	}
}

func TestMountTable_ListRoot(t *testing.T) {
	mt := NewMountTable()
	_ = mt.Mount("a", mustNewLocalFS(t, t.TempDir()))
	_ = mt.Mount("b", mustNewLocalFS(t, t.TempDir()))

	result, err := mt.List(context.Background(), "")
	if err != nil {
		t.Fatalf("List root: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("List root got %d, want 2", len(result.Files))
	}
}

func TestMountTable_Access(t *testing.T) {
	dir := t.TempDir()
	local := mustNewLocalFS(t, dir)
	fake := newFakeS3()
	obj := NewObjectFS(fake, "bucket", WithBaseURL("https://cdn.example.com"))

	mt := NewMountTable()
	_ = mt.Mount("local", local)
	_ = mt.Mount("s3", obj)
	ctx := context.Background()

	info, err := mt.Access(ctx, "local/readme.txt")
	if err != nil {
		t.Fatalf("Access local: %v", err)
	}
	if info.Path == "" {
		t.Fatal("Access local should return Path")
	}

	info, err = mt.Access(ctx, "s3/img/logo.png")
	if err != nil {
		t.Fatalf("Access s3: %v", err)
	}
	if info.URL == "" {
		t.Fatal("Access s3 should return URL")
	}
	if !strings.Contains(info.URL, "img/logo.png") {
		t.Fatalf("Access s3 URL = %q, want to contain key", info.URL)
	}
}

func TestMountTable_InvalidPrefix(t *testing.T) {
	mt := NewMountTable()
	if err := mt.Mount("a/b", mustNewLocalFS(t, t.TempDir())); err == nil {
		t.Fatal("mount with slash should fail")
	}
	if err := mt.Mount("", mustNewLocalFS(t, t.TempDir())); err == nil {
		t.Fatal("mount with empty should fail")
	}
}

func TestMountTable_Unmount(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("data", local)
	ctx := context.Background()

	_ = mt.Put(ctx, "data/file.txt", strings.NewReader("hello"))
	_, err := mt.Get(ctx, "data/file.txt")
	if err != nil {
		t.Fatalf("Get before unmount: %v", err)
	}

	mt.Unmount("data")

	_, err = mt.Get(ctx, "data/file.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after unmount error = %v, want ErrNotFound", err)
	}
}

func TestMountTable_Delete(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("fs", local)
	ctx := context.Background()

	_ = mt.Put(ctx, "fs/del.txt", strings.NewReader("bye"))

	if err := mt.Delete(ctx, "fs/del.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := mt.Get(ctx, "fs/del.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after delete error = %v, want ErrNotFound", err)
	}
}

func TestMountTable_Stat(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("fs", local)
	ctx := context.Background()

	_ = mt.Put(ctx, "fs/stat.txt", strings.NewReader("12345"))

	info, err := mt.Stat(ctx, "fs/stat.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("Stat size = %d, want 5", info.Size)
	}
}

func TestMountTable_ConcurrentAccess(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("c", local)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "c/" + strings.Repeat("a", n%5+1) + ".txt"
			_ = mt.Put(ctx, key, strings.NewReader("data"))
			_, _ = mt.Get(ctx, key)
			_, _ = mt.Stat(ctx, key)
			_, _ = mt.List(ctx, "c")
		}(i)
	}
	wg.Wait()
}

func TestMountTable_Copy_SameBackend(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("fs", local)
	ctx := context.Background()

	_ = mt.Put(ctx, "fs/src.txt", strings.NewReader("copy-me"))

	if err := mt.Copy(ctx, "fs/src.txt", "fs/dst.txt"); err != nil {
		t.Fatalf("Copy same backend: %v", err)
	}

	rc, err := mt.Get(ctx, "fs/dst.txt")
	if err != nil {
		t.Fatalf("Get dst: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "copy-me" {
		t.Fatalf("content = %q, want %q", data, "copy-me")
	}
}

func TestMountTable_Exists(t *testing.T) {
	mt := NewMountTable()
	local := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("fs", local)
	ctx := context.Background()

	_ = mt.Put(ctx, "fs/found.txt", strings.NewReader("yes"))

	ok, err := mt.Exists(ctx, "fs/found.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("Exists should return true")
	}

	ok, err = mt.Exists(ctx, "fs/nope.txt")
	if err != nil {
		t.Fatalf("Exists missing: %v", err)
	}
	if ok {
		t.Fatal("Exists should return false")
	}
}

func TestMountTable_Copy_CrossBackend(t *testing.T) {
	mt := NewMountTable()
	local1 := mustNewLocalFS(t, t.TempDir())
	local2 := mustNewLocalFS(t, t.TempDir())
	_ = mt.Mount("a", local1)
	_ = mt.Mount("b", local2)
	ctx := context.Background()

	_ = mt.Put(ctx, "a/file.txt", strings.NewReader("cross"))

	if err := mt.Copy(ctx, "a/file.txt", "b/file.txt"); err != nil {
		t.Fatalf("Copy cross backend: %v", err)
	}

	rc, err := mt.Get(ctx, "b/file.txt")
	if err != nil {
		t.Fatalf("Get from b: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "cross" {
		t.Fatalf("content = %q, want %q", data, "cross")
	}
}
