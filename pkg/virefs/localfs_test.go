// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustNewLocalFS(t *testing.T, root string, opts ...LocalOption) *LocalFS {
	t.Helper()
	fs, err := NewLocalFS(root, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return fs
}

func TestLocalFS_PutGetDeleteStat(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	if err := fs.Put(ctx, "hello.txt", strings.NewReader("world")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	rc, err := fs.Get(ctx, "hello.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "world" {
		t.Fatalf("Get content = %q, want %q", data, "world")
	}

	info, err := fs.Stat(ctx, "hello.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("Stat size = %d, want 5", info.Size)
	}

	if err := fs.Delete(ctx, "hello.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = fs.Get(ctx, "hello.txt")
	if err == nil {
		t.Fatal("Get after delete should fail")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after delete error = %v, want ErrNotFound", err)
	}
}

func TestLocalFS_List(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "sub/b.txt", strings.NewReader("b"))

	result, err := fs.List(ctx, "")
	if err != nil {
		t.Fatalf("List root: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("List root got %d entries, want 2", len(result.Files))
	}

	result, err = fs.List(ctx, "sub")
	if err != nil {
		t.Fatalf("List sub: %v", err)
	}
	if len(result.Files) != 1 {
		t.Fatalf("List sub got %d entries, want 1", len(result.Files))
	}
	if result.Files[0].Key != "sub/b.txt" {
		t.Fatalf("List sub key = %q, want %q", result.Files[0].Key, "sub/b.txt")
	}
}

func TestLocalFS_NestedPut(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	if err := fs.Put(ctx, "a/b/c/d.txt", strings.NewReader("deep")); err != nil {
		t.Fatalf("nested Put: %v", err)
	}
	rc, err := fs.Get(ctx, "a/b/c/d.txt")
	if err != nil {
		t.Fatalf("nested Get: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "deep" {
		t.Fatalf("nested Get content = %q, want %q", data, "deep")
	}
}

func TestLocalFS_WithKeyFunc(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir, WithLocalKeyFunc(func(key string) string {
		return "transformed/" + key
	}))
	ctx := context.Background()

	if err := fs.Put(ctx, "note.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("Put with KeyFunc: %v", err)
	}

	rc, err := fs.Get(ctx, "note.txt")
	if err != nil {
		t.Fatalf("Get with KeyFunc: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Fatalf("Get content = %q, want %q", data, "hello")
	}

	plain := mustNewLocalFS(t, dir)
	rc, err = plain.Get(ctx, "transformed/note.txt")
	if err != nil {
		t.Fatalf("plain Get transformed path: %v", err)
	}
	data, _ = io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Fatalf("plain Get content = %q, want %q", data, "hello")
	}
}

func TestLocalFS_Access(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "doc/readme.txt", strings.NewReader("hello"))

	info, err := fs.Access(ctx, "doc/readme.txt")
	if err != nil {
		t.Fatalf("Access: %v", err)
	}
	if info.Path == "" {
		t.Fatal("Access.Path should be non-empty for LocalFS")
	}
	if info.URL != "" {
		t.Fatal("Access.URL should be empty for LocalFS without AccessFunc")
	}
	if !strings.HasSuffix(info.Path, "doc/readme.txt") {
		t.Fatalf("Access.Path = %q, want suffix doc/readme.txt", info.Path)
	}
}

func TestLocalFS_AccessWithAccessFunc(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir, WithLocalAccessFunc(func(key string) *AccessInfo {
		return &AccessInfo{URL: "https://cdn.example.com/files/" + key}
	}))
	ctx := context.Background()

	_ = fs.Put(ctx, "images/logo.png", strings.NewReader("png"))

	info, err := fs.Access(ctx, "images/logo.png")
	if err != nil {
		t.Fatalf("Access with AccessFunc: %v", err)
	}
	if info.Path == "" {
		t.Fatal("Access.Path should still be set with AccessFunc")
	}
	if !strings.HasSuffix(info.Path, "images/logo.png") {
		t.Fatalf("Access.Path = %q, want suffix images/logo.png", info.Path)
	}
	wantURL := "https://cdn.example.com/files/images/logo.png"
	if info.URL != wantURL {
		t.Fatalf("Access.URL = %q, want %q", info.URL, wantURL)
	}
}

func TestLocalFS_AccessFuncWithKeyFunc(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir,
		WithLocalKeyFunc(func(key string) string { return "v2/" + key }),
		WithLocalAccessFunc(func(key string) *AccessInfo {
			return &AccessInfo{URL: "https://cdn.example.com/" + key}
		}),
	)
	ctx := context.Background()

	info, err := fs.Access(ctx, "doc.txt")
	if err != nil {
		t.Fatalf("Access: %v", err)
	}
	if !strings.HasSuffix(info.Path, "v2/doc.txt") {
		t.Fatalf("Access.Path = %q, want suffix v2/doc.txt", info.Path)
	}
	wantURL := "https://cdn.example.com/v2/doc.txt"
	if info.URL != wantURL {
		t.Fatalf("Access.URL = %q, want %q", info.URL, wantURL)
	}
}

func TestLocalFS_ExistsMethod(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "found.txt", strings.NewReader("yes"))

	ok, err := fs.Exists(ctx, "found.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("Exists should return true for existing key")
	}

	ok, err = fs.Exists(ctx, "missing.txt")
	if err != nil {
		t.Fatalf("Exists missing: %v", err)
	}
	if ok {
		t.Fatal("Exists should return false for missing key")
	}

	_, err = fs.Exists(ctx, "../../etc/passwd")
	if err == nil {
		t.Fatal("Exists with traversal should return error")
	}
}

func TestLocalFS_AccessWithKeyFunc(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir, WithLocalKeyFunc(func(key string) string {
		return "v2/" + key
	}))
	ctx := context.Background()

	info, err := fs.Access(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Access with KeyFunc: %v", err)
	}
	if !strings.HasSuffix(info.Path, "v2/file.txt") {
		t.Fatalf("Access.Path = %q, want suffix v2/file.txt", info.Path)
	}
}

func TestLocalFS_TraversalRejected(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_, err := fs.Get(ctx, "../../etc/passwd")
	if err == nil {
		t.Fatal("traversal should be rejected")
	}
	if !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("traversal error = %v, want ErrInvalidKey", err)
	}
}

func TestLocalFS_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir, WithAtomicWrite())
	ctx := context.Background()

	if err := fs.Put(ctx, "atomic.txt", strings.NewReader("safe")); err != nil {
		t.Fatalf("AtomicWrite Put: %v", err)
	}
	rc, err := fs.Get(ctx, "atomic.txt")
	if err != nil {
		t.Fatalf("AtomicWrite Get: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "safe" {
		t.Fatalf("AtomicWrite content = %q, want %q", data, "safe")
	}
}

func TestLocalFS_AtomicWriteNested(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir, WithAtomicWrite())
	ctx := context.Background()

	if err := fs.Put(ctx, "a/b/c.txt", strings.NewReader("deep")); err != nil {
		t.Fatalf("AtomicWrite nested Put: %v", err)
	}
	rc, err := fs.Get(ctx, "a/b/c.txt")
	if err != nil {
		t.Fatalf("AtomicWrite nested Get: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "deep" {
		t.Fatalf("AtomicWrite nested content = %q, want %q", data, "deep")
	}
}

func TestLocalFS_Copy(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "original.txt", strings.NewReader("data"))

	if err := fs.Copy(ctx, "original.txt", "copied.txt"); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	rc, err := fs.Get(ctx, "copied.txt")
	if err != nil {
		t.Fatalf("Get copied: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "data" {
		t.Fatalf("Copy content = %q, want %q", data, "data")
	}

	rc, err = fs.Get(ctx, "original.txt")
	if err != nil {
		t.Fatalf("original should still exist: %v", err)
	}
	rc.Close()
}

func TestLocalFS_CopyNested(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "src/file.txt", strings.NewReader("nested"))

	if err := fs.Copy(ctx, "src/file.txt", "dst/dir/file.txt"); err != nil {
		t.Fatalf("Copy nested: %v", err)
	}

	rc, err := fs.Get(ctx, "dst/dir/file.txt")
	if err != nil {
		t.Fatalf("Get nested copy: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "nested" {
		t.Fatalf("Nested copy content = %q, want %q", data, "nested")
	}
}

func TestLocalFS_Exists(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "exists.txt", strings.NewReader("yes"))

	ok, err := Exists(ctx, fs, "exists.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("Exists should return true")
	}

	ok, err = Exists(ctx, fs, "nope.txt")
	if err != nil {
		t.Fatalf("Exists missing: %v", err)
	}
	if ok {
		t.Fatal("Exists should return false for missing key")
	}
}

func TestLocalFS_CopyHelper_SameBackend(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("hello"))

	if err := Copy(ctx, fs, "a.txt", fs, "b.txt"); err != nil {
		t.Fatalf("Copy helper same backend: %v", err)
	}

	rc, err := fs.Get(ctx, "b.txt")
	if err != nil {
		t.Fatalf("Get b.txt: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Fatalf("content = %q, want %q", data, "hello")
	}
}

func TestLocalFS_DeleteNotFound(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	err := fs.Delete(ctx, "nonexistent.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Delete missing key error = %v, want ErrNotFound", err)
	}
}

func TestLocalFS_WithCreateRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", "dir")
	fs := mustNewLocalFS(t, root, WithCreateRoot())
	ctx := context.Background()

	if err := fs.Put(ctx, "test.txt", strings.NewReader("created")); err != nil {
		t.Fatalf("Put after WithCreateRoot: %v", err)
	}
	rc, err := fs.Get(ctx, "test.txt")
	if err != nil {
		t.Fatalf("Get after WithCreateRoot: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "created" {
		t.Fatalf("content = %q, want %q", data, "created")
	}
}

func TestLocalFS_WithDirPerm(t *testing.T) {
	root := t.TempDir()
	perm := os.FileMode(0o700)
	fs := mustNewLocalFS(t, root, WithDirPerm(perm))
	ctx := context.Background()

	_ = fs.Put(ctx, "sub/file.txt", strings.NewReader("data"))

	info, err := os.Stat(filepath.Join(root, "sub"))
	if err != nil {
		t.Fatalf("Stat sub dir: %v", err)
	}
	got := info.Mode().Perm()
	if got != perm {
		t.Fatalf("dir perm = %o, want %o", got, perm)
	}
}

func TestLocalFS_StatContentType(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "photo.jpg", strings.NewReader("fake-jpeg"))
	_ = fs.Put(ctx, "readme.txt", strings.NewReader("text"))
	_ = fs.Put(ctx, "noext", strings.NewReader("no extension"))

	info, err := fs.Stat(ctx, "photo.jpg")
	if err != nil {
		t.Fatalf("Stat photo.jpg: %v", err)
	}
	if info.ContentType != "image/jpeg" {
		t.Fatalf("Stat photo.jpg ContentType = %q, want %q", info.ContentType, "image/jpeg")
	}

	info, err = fs.Stat(ctx, "readme.txt")
	if err != nil {
		t.Fatalf("Stat readme.txt: %v", err)
	}
	if !strings.HasPrefix(info.ContentType, "text/plain") {
		t.Fatalf("Stat readme.txt ContentType = %q, want prefix %q", info.ContentType, "text/plain")
	}

	info, err = fs.Stat(ctx, "noext")
	if err != nil {
		t.Fatalf("Stat noext: %v", err)
	}
	if info.ContentType != "" {
		t.Fatalf("Stat noext ContentType = %q, want empty", info.ContentType)
	}
}

func TestLocalFS_ListShallow(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "sub/b.txt", strings.NewReader("b"))
	_ = fs.Put(ctx, "sub/deep/c.txt", strings.NewReader("c"))

	result, err := fs.List(ctx, "")
	if err != nil {
		t.Fatalf("List root: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("List root got %d entries, want 2 (a.txt + sub/)", len(result.Files))
	}

	result, err = fs.List(ctx, "sub")
	if err != nil {
		t.Fatalf("List sub: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("List sub got %d entries, want 2 (b.txt + deep/)", len(result.Files))
	}
}
