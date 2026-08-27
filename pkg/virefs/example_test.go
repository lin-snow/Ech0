// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	virefs "github.com/lin-snow/ech0/pkg/virefs"
)

func ExampleNewLocalFS() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	fs, err := virefs.NewLocalFS(dir)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	if err := fs.Put(ctx, "hello.txt", strings.NewReader("world")); err != nil {
		log.Fatal(err)
	}
	rc, err := fs.Get(ctx, "hello.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
}

func ExampleNewLocalFS_withKeyFunc() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	fs, err := virefs.NewLocalFS(dir, virefs.WithLocalKeyFunc(func(key string) string {
		return "v2/" + key
	}))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	_ = fs.Put(ctx, "note.txt", strings.NewReader("hello"))

	info, err := fs.Access(ctx, "note.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(strings.Contains(info.Path, "v2/note.txt"))
}

func ExampleCopy() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	fs, err := virefs.NewLocalFS(dir)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	_ = fs.Put(ctx, "src.txt", strings.NewReader("data"))

	if err := virefs.Copy(ctx, fs, "src.txt", fs, "dst.txt"); err != nil {
		log.Fatal(err)
	}

	rc, _ := fs.Get(ctx, "dst.txt")
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
}

func ExampleWalk() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	fs, err := virefs.NewLocalFS(dir)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "sub/b.txt", strings.NewReader("b"))

	var count int
	_ = virefs.Walk(ctx, fs, "", func(key string, info virefs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir {
			count++
		}
		return nil
	})
	fmt.Println(count)
}

func ExampleNewSchema() {
	schema := virefs.NewSchema(
		virefs.RouteByExt("images/", ".jpg", ".png"),
		virefs.RouteByExt("docs/", ".pdf"),
		virefs.DefaultRoute("other/"),
	)

	fmt.Println(schema.Resolve("cat.jpg"))
	fmt.Println(schema.Resolve("report.pdf"))
	fmt.Println(schema.Resolve("readme.txt"))
}

func ExampleNewMountTable() {
	dir1, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir1)
	dir2, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir2)

	fs1, _ := virefs.NewLocalFS(dir1)
	fs2, _ := virefs.NewLocalFS(dir2)

	mt := virefs.NewMountTable()
	_ = mt.Mount("data", fs1)
	_ = mt.Mount("cache", fs2)

	ctx := context.Background()
	_ = mt.Put(ctx, "data/file.txt", strings.NewReader("hello"))

	rc, _ := mt.Get(ctx, "data/file.txt")
	defer rc.Close()
	data, _ := io.ReadAll(rc)
	fmt.Println(string(data))
}

func ExampleNewLocalFS_withAccessFunc() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	fs, err := virefs.NewLocalFS(dir, virefs.WithLocalAccessFunc(func(key string) *virefs.AccessInfo {
		return &virefs.AccessInfo{URL: "https://cdn.example.com/files/" + key}
	}))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	_ = fs.Put(ctx, "photo.jpg", strings.NewReader("jpeg-data"))

	info, err := fs.Access(ctx, "photo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(info.URL)
	fmt.Println(info.Path != "")
}

func ExampleChain() {
	dir, _ := os.MkdirTemp("", "virefs-example-*")
	defer os.RemoveAll(dir)

	base, _ := virefs.NewLocalFS(dir)
	ctx := context.Background()

	_ = base.Put(ctx, "secret.txt", strings.NewReader("classified"))

	fs := virefs.Chain(base,
		func(next virefs.FS) virefs.FS {
			return &uppercaseFS{virefs.BaseFS{Inner: next}}
		},
		func(next virefs.FS) virefs.FS {
			return virefs.WithHooks(next, virefs.Hooks{
				WrapGet: func(key string, rc io.ReadCloser) io.ReadCloser {
					fmt.Println("audit: read", key)
					return rc
				},
			})
		},
	)

	rc, _ := fs.Get(ctx, "secret.txt")
	data, _ := io.ReadAll(rc)
	rc.Close()
	fmt.Println(string(data))
}

type uppercaseFS struct{ virefs.BaseFS }

func (u *uppercaseFS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	rc, err := u.Inner.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	return io.NopCloser(strings.NewReader(strings.ToUpper(string(data)))), nil
}

func ExampleMigrate() {
	srcDir, _ := os.MkdirTemp("", "virefs-src-*")
	defer os.RemoveAll(srcDir)
	dstDir, _ := os.MkdirTemp("", "virefs-dst-*")
	defer os.RemoveAll(dstDir)

	src, _ := virefs.NewLocalFS(srcDir)
	dst, _ := virefs.NewLocalFS(dstDir)
	ctx := context.Background()

	_ = src.Put(ctx, "a.txt", strings.NewReader("aaa"))
	_ = src.Put(ctx, "b.txt", strings.NewReader("bbb"))

	result, err := virefs.Migrate(ctx, src, "", dst, "backup",
		virefs.WithConflictPolicy(virefs.ConflictSkip),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("copied=%d skipped=%d total=%d\n", result.Copied, result.Skipped, result.Total)
}
