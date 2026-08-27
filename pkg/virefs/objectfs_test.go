// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type fakeS3 struct {
	objects      map[string][]byte
	contentTypes map[string]string
	metadata     map[string]map[string]string
	maxKeys      int
}

func newFakeS3() *fakeS3 {
	return &fakeS3{
		objects:      make(map[string][]byte),
		contentTypes: make(map[string]string),
		metadata:     make(map[string]map[string]string),
	}
}

func (f *fakeS3) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	data, err := io.ReadAll(in.Body)
	if err != nil {
		return nil, err
	}
	key := aws.ToString(in.Key)
	f.objects[key] = data
	if in.ContentType != nil {
		f.contentTypes[key] = aws.ToString(in.ContentType)
	}
	if in.Metadata != nil {
		f.metadata[key] = in.Metadata
	}
	return &s3.PutObjectOutput{}, nil
}

func (f *fakeS3) CopyObject(_ context.Context, in *s3.CopyObjectInput, _ ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	src := aws.ToString(in.CopySource)
	parts := strings.SplitN(src, "/", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid copy source: %s", src)
	}
	srcKey := parts[1]
	data, ok := f.objects[srcKey]
	if !ok {
		return nil, &types.NoSuchKey{Message: aws.String("no such key: " + srcKey)}
	}
	dstKey := aws.ToString(in.Key)
	f.objects[dstKey] = make([]byte, len(data))
	copy(f.objects[dstKey], data)
	return &s3.CopyObjectOutput{}, nil
}

func (f *fakeS3) GetObject(_ context.Context, in *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	key := aws.ToString(in.Key)
	data, ok := f.objects[key]
	if !ok {
		return nil, &types.NoSuchKey{Message: aws.String("no such key: " + key)}
	}
	return &s3.GetObjectOutput{
		Body:          io.NopCloser(bytes.NewReader(data)),
		ContentLength: aws.Int64(int64(len(data))),
	}, nil
}

func (f *fakeS3) DeleteObject(_ context.Context, in *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	delete(f.objects, aws.ToString(in.Key))
	return &s3.DeleteObjectOutput{}, nil
}

func (f *fakeS3) DeleteObjects(_ context.Context, in *s3.DeleteObjectsInput, _ ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	for _, obj := range in.Delete.Objects {
		delete(f.objects, aws.ToString(obj.Key))
	}
	return &s3.DeleteObjectsOutput{}, nil
}

func (f *fakeS3) HeadObject(_ context.Context, in *s3.HeadObjectInput, _ ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	key := aws.ToString(in.Key)
	data, ok := f.objects[key]
	if !ok {
		return nil, &types.NotFound{Message: aws.String("not found: " + key)}
	}
	now := time.Now()
	out := &s3.HeadObjectOutput{
		ContentLength: aws.Int64(int64(len(data))),
		LastModified:  &now,
	}
	if ct, ok := f.contentTypes[key]; ok {
		out.ContentType = aws.String(ct)
	}
	return out, nil
}

func (f *fakeS3) ListObjectsV2(_ context.Context, in *s3.ListObjectsV2Input, _ ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	prefix := aws.ToString(in.Prefix)
	delimiter := aws.ToString(in.Delimiter)

	type entry struct {
		key  string
		data []byte
	}
	var matched []entry
	commonPrefixSet := make(map[string]struct{})
	var commonPrefixes []types.CommonPrefix

	keys := make([]string, 0, len(f.objects))
	for k := range f.objects {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if !strings.HasPrefix(k, prefix) {
			continue
		}
		rest := k[len(prefix):]
		if delimiter != "" {
			if idx := strings.Index(rest, delimiter); idx >= 0 {
				cp := prefix + rest[:idx+len(delimiter)]
				if _, seen := commonPrefixSet[cp]; !seen {
					commonPrefixSet[cp] = struct{}{}
					commonPrefixes = append(commonPrefixes, types.CommonPrefix{
						Prefix: aws.String(cp),
					})
				}
				continue
			}
		}
		matched = append(matched, entry{key: k, data: f.objects[k]})
	}

	startIdx := 0
	if tok := aws.ToString(in.ContinuationToken); tok != "" {
		startIdx, _ = strconv.Atoi(tok)
	}

	limit := len(matched)
	if f.maxKeys > 0 && f.maxKeys < limit-startIdx {
		limit = startIdx + f.maxKeys
	}

	var contents []types.Object
	for i := startIdx; i < limit && i < len(matched); i++ {
		now := time.Now()
		contents = append(contents, types.Object{
			Key:          aws.String(matched[i].key),
			Size:         aws.Int64(int64(len(matched[i].data))),
			LastModified: &now,
		})
	}

	truncated := limit < len(matched)
	out := &s3.ListObjectsV2Output{
		Contents:       contents,
		CommonPrefixes: commonPrefixes,
		IsTruncated:    aws.Bool(truncated),
	}
	if truncated {
		nextToken := strconv.Itoa(limit)
		out.NextContinuationToken = aws.String(nextToken)
	}
	return out, nil
}

func TestObjectFS_PutGetDeleteStat(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "test-bucket")
	ctx := context.Background()

	if err := fs.Put(ctx, "doc.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("Put: %v", err)
	}

	rc, err := fs.Get(ctx, "doc.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "hello" {
		t.Fatalf("Get content = %q, want %q", data, "hello")
	}

	info, err := fs.Stat(ctx, "doc.txt")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != 5 {
		t.Fatalf("Stat size = %d, want 5", info.Size)
	}

	if err := fs.Delete(ctx, "doc.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = fs.Get(ctx, "doc.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after delete error = %v, want ErrNotFound", err)
	}
}

func TestObjectFS_BasePrefix(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("data/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))

	if _, ok := fake.objects["data/a.txt"]; !ok {
		t.Fatal("expected object at data/a.txt in fake store")
	}

	rc, err := fs.Get(ctx, "a.txt")
	if err != nil {
		t.Fatalf("Get with prefix: %v", err)
	}
	d, _ := io.ReadAll(rc)
	rc.Close()
	if string(d) != "a" {
		t.Fatalf("Get content = %q, want %q", d, "a")
	}
}

func TestObjectFS_List(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "dir/x.txt", strings.NewReader("x"))
	_ = fs.Put(ctx, "dir/y.txt", strings.NewReader("y"))
	_ = fs.Put(ctx, "other.txt", strings.NewReader("o"))

	result, err := fs.List(ctx, "dir")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Files) != 2 {
		t.Fatalf("List got %d entries, want 2", len(result.Files))
	}
}

func TestObjectFS_WithKeyFunc(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("base/"), WithObjectKeyFunc(func(key string) string {
		return "2026/03/06/" + key
	}))
	ctx := context.Background()

	_ = fs.Put(ctx, "photo.jpg", strings.NewReader("img"))

	wantKey := "base/2026/03/06/photo.jpg"
	if _, ok := fake.objects[wantKey]; !ok {
		keys := make([]string, 0, len(fake.objects))
		for k := range fake.objects {
			keys = append(keys, k)
		}
		t.Fatalf("expected object at %q, got keys %v", wantKey, keys)
	}

	rc, err := fs.Get(ctx, "photo.jpg")
	if err != nil {
		t.Fatalf("Get with KeyFunc: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "img" {
		t.Fatalf("Get content = %q, want %q", data, "img")
	}
}

func TestObjectFS_NotFound(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_, err := fs.Get(ctx, "nope.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get missing key error = %v, want ErrNotFound", err)
	}

	_, err = fs.Stat(ctx, "nope.txt")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Stat missing key error = %v, want ErrNotFound", err)
	}
}

type fakePresign struct{}

func (fp *fakePresign) PresignGetObject(_ context.Context, in *s3.GetObjectInput, opts ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	var po s3.PresignOptions
	for _, fn := range opts {
		fn(&po)
	}
	return &v4.PresignedHTTPRequest{
		URL:          fmt.Sprintf("https://s3.example.com/%s/%s?expires=%s", aws.ToString(in.Bucket), aws.ToString(in.Key), po.Expires),
		Method:       "GET",
		SignedHeader: http.Header{"Host": []string{"s3.example.com"}},
	}, nil
}

func (fp *fakePresign) PresignPutObject(_ context.Context, in *s3.PutObjectInput, opts ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	var po s3.PresignOptions
	for _, fn := range opts {
		fn(&po)
	}
	return &v4.PresignedHTTPRequest{
		URL:          fmt.Sprintf("https://s3.example.com/%s/%s?expires=%s", aws.ToString(in.Bucket), aws.ToString(in.Key), po.Expires),
		Method:       "PUT",
		SignedHeader: http.Header{"Host": []string{"s3.example.com"}},
	}, nil
}

func TestObjectFS_PresignGet(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"), WithPresignClient(&fakePresign{}))
	ctx := context.Background()

	req, err := fs.PresignGet(ctx, "report.pdf", 15*time.Minute)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	if req.Method != "GET" {
		t.Fatalf("PresignGet method = %q, want GET", req.Method)
	}
	if !strings.Contains(req.URL, "pfx/report.pdf") {
		t.Fatalf("PresignGet URL should contain prefixed key, got %q", req.URL)
	}
	if !strings.Contains(req.URL, "15m0s") {
		t.Fatalf("PresignGet URL should contain expiry, got %q", req.URL)
	}
}

func TestObjectFS_PresignPut(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPresignClient(&fakePresign{}))
	ctx := context.Background()

	req, err := fs.PresignPut(ctx, "upload.zip", 30*time.Minute)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if req.Method != "PUT" {
		t.Fatalf("PresignPut method = %q, want PUT", req.Method)
	}
	if !strings.Contains(req.URL, "upload.zip") {
		t.Fatalf("PresignPut URL should contain key, got %q", req.URL)
	}
}

func TestObjectFS_PresignWithKeyFunc(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("data/"),
		WithObjectKeyFunc(func(key string) string { return "v2/" + key }),
		WithPresignClient(&fakePresign{}),
	)
	ctx := context.Background()

	req, err := fs.PresignGet(ctx, "file.txt", 5*time.Minute)
	if err != nil {
		t.Fatalf("PresignGet with KeyFunc: %v", err)
	}
	if !strings.Contains(req.URL, "data/v2/file.txt") {
		t.Fatalf("PresignGet URL should contain transformed key, got %q", req.URL)
	}
}

func TestObjectFS_PresignWithoutClient(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_, err := fs.PresignGet(ctx, "file.txt", 5*time.Minute)
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("PresignGet without client error = %v, want ErrNotSupported", err)
	}

	_, err = fs.PresignPut(ctx, "file.txt", 5*time.Minute)
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("PresignPut without client error = %v, want ErrNotSupported", err)
	}
}

func TestObjectFS_AccessWithPresign(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"), WithPresignClient(&fakePresign{}))
	ctx := context.Background()

	info, err := fs.Access(ctx, "report.pdf")
	if err != nil {
		t.Fatalf("Access with presign: %v", err)
	}
	if info.URL == "" {
		t.Fatal("Access.URL should be non-empty")
	}
	if info.Path != "" {
		t.Fatal("Access.Path should be empty for ObjectFS")
	}
	if !strings.Contains(info.URL, "pfx/report.pdf") {
		t.Fatalf("Access.URL should contain prefixed key, got %q", info.URL)
	}
}

func TestObjectFS_AccessWithBaseURL(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("data/"), WithBaseURL("https://cdn.example.com"))
	ctx := context.Background()

	info, err := fs.Access(ctx, "img/logo.png")
	if err != nil {
		t.Fatalf("Access with base URL: %v", err)
	}
	want := "https://cdn.example.com/data/img/logo.png"
	if info.URL != want {
		t.Fatalf("Access.URL = %q, want %q", info.URL, want)
	}
}

func TestObjectFS_AccessPresignPriority(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket",
		WithPresignClient(&fakePresign{}),
		WithBaseURL("https://cdn.example.com"),
	)
	ctx := context.Background()

	info, err := fs.Access(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Access: %v", err)
	}
	if strings.HasPrefix(info.URL, "https://cdn.example.com") {
		t.Fatalf("presign client should take priority over base URL, got %q", info.URL)
	}
}

func TestObjectFS_AccessNotConfigured(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_, err := fs.Access(ctx, "file.txt")
	if !errors.Is(err, ErrNotSupported) {
		t.Fatalf("Access without config error = %v, want ErrNotSupported", err)
	}
}

func TestObjectFS_AccessFunc_CDN(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("assets/"),
		WithAccessFunc(func(key string) *AccessInfo {
			return &AccessInfo{URL: "https://cdn.example.com/" + key}
		}),
	)
	ctx := context.Background()

	info, err := fs.Access(ctx, "img/logo.png")
	if err != nil {
		t.Fatalf("Access with AccessFunc: %v", err)
	}
	want := "https://cdn.example.com/assets/img/logo.png"
	if info.URL != want {
		t.Fatalf("Access.URL = %q, want %q", info.URL, want)
	}
}

func TestObjectFS_AccessFunc_PriorityOverPresign(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket",
		WithPresignClient(&fakePresign{}),
		WithBaseURL("https://s3.example.com"),
		WithAccessFunc(func(key string) *AccessInfo {
			return &AccessInfo{URL: "https://cdn.fast.io/" + key}
		}),
	)
	ctx := context.Background()

	info, err := fs.Access(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Access: %v", err)
	}
	if !strings.HasPrefix(info.URL, "https://cdn.fast.io/") {
		t.Fatalf("AccessFunc should take priority over presign and baseURL, got %q", info.URL)
	}
}

func TestObjectFS_AccessFunc_WithKeyFunc(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("data/"),
		WithObjectKeyFunc(func(key string) string { return "v2/" + key }),
		WithAccessFunc(func(key string) *AccessInfo {
			return &AccessInfo{URL: "https://cdn.example.com/" + key}
		}),
	)
	ctx := context.Background()

	info, err := fs.Access(ctx, "config.yaml")
	if err != nil {
		t.Fatalf("Access: %v", err)
	}
	want := "https://cdn.example.com/data/v2/config.yaml"
	if info.URL != want {
		t.Fatalf("AccessFunc should receive full s3 key, got URL %q, want %q", info.URL, want)
	}
}

func TestObjectFS_PutWithContentType(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_ = fs.Put(ctx, "image.png", strings.NewReader("png-data"), WithContentType("image/png"))

	if ct := fake.contentTypes["image.png"]; ct != "image/png" {
		t.Fatalf("ContentType = %q, want %q", ct, "image/png")
	}
}

func TestObjectFS_StatContentType(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_ = fs.Put(ctx, "image.png", strings.NewReader("png-data"), WithContentType("image/png"))

	info, err := fs.Stat(ctx, "image.png")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.ContentType != "image/png" {
		t.Fatalf("Stat ContentType = %q, want %q", info.ContentType, "image/png")
	}
}

func TestObjectFS_StatContentTypeEmpty(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_ = fs.Put(ctx, "data.bin", strings.NewReader("binary"))

	info, err := fs.Stat(ctx, "data.bin")
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.ContentType != "" {
		t.Fatalf("Stat ContentType = %q, want empty", info.ContentType)
	}
}

func TestObjectFS_PutWithMetadata(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	meta := map[string]string{"author": "test", "version": "1"}
	_ = fs.Put(ctx, "doc.pdf", strings.NewReader("pdf"), WithMetadata(meta))

	got := fake.metadata["doc.pdf"]
	if got["author"] != "test" || got["version"] != "1" {
		t.Fatalf("Metadata = %v, want %v", got, meta)
	}
}

func TestObjectFS_Copy(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "src.txt", strings.NewReader("content"))

	if err := fs.Copy(ctx, "src.txt", "dst.txt"); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	if _, ok := fake.objects["pfx/dst.txt"]; !ok {
		t.Fatal("expected dst object after Copy")
	}

	rc, err := fs.Get(ctx, "dst.txt")
	if err != nil {
		t.Fatalf("Get dst: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "content" {
		t.Fatalf("Copy content = %q, want %q", data, "content")
	}
}

func TestObjectFS_Exists(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket")
	ctx := context.Background()

	_ = fs.Put(ctx, "yes.txt", strings.NewReader("y"))

	ok, err := Exists(ctx, fs, "yes.txt")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("Exists should return true for existing key")
	}

	ok, err = Exists(ctx, fs, "no.txt")
	if err != nil {
		t.Fatalf("Exists missing: %v", err)
	}
	if ok {
		t.Fatal("Exists should return false for missing key")
	}
}

func TestObjectFS_ExistsMethod(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "data.bin", strings.NewReader("bin"))

	ok, err := fs.Exists(ctx, "data.bin")
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if !ok {
		t.Fatal("Exists should return true")
	}

	ok, err = fs.Exists(ctx, "missing.bin")
	if err != nil {
		t.Fatalf("Exists missing: %v", err)
	}
	if ok {
		t.Fatal("Exists should return false")
	}
}

func TestObjectFS_CrossBackendCopy(t *testing.T) {
	fake1 := newFakeS3()
	fake2 := newFakeS3()
	src := NewObjectFS(fake1, "src-bucket")
	dst := NewObjectFS(fake2, "dst-bucket")
	ctx := context.Background()

	_ = src.Put(ctx, "file.txt", strings.NewReader("cross"))

	if err := Copy(ctx, src, "file.txt", dst, "file.txt"); err != nil {
		t.Fatalf("Cross-backend Copy: %v", err)
	}

	rc, err := dst.Get(ctx, "file.txt")
	if err != nil {
		t.Fatalf("Get from dst: %v", err)
	}
	data, _ := io.ReadAll(rc)
	rc.Close()
	if string(data) != "cross" {
		t.Fatalf("Cross copy content = %q, want %q", data, "cross")
	}
}

func TestObjectFS_ListPagination(t *testing.T) {
	fake := newFakeS3()
	fake.maxKeys = 2
	fs := NewObjectFS(fake, "bucket", WithPrefix("p/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "b.txt", strings.NewReader("b"))
	_ = fs.Put(ctx, "c.txt", strings.NewReader("c"))
	_ = fs.Put(ctx, "d.txt", strings.NewReader("d"))
	_ = fs.Put(ctx, "e.txt", strings.NewReader("e"))

	result, err := fs.List(ctx, "")
	if err != nil {
		t.Fatalf("List with pagination: %v", err)
	}
	if len(result.Files) != 5 {
		t.Fatalf("List got %d entries, want 5", len(result.Files))
	}
}

func TestObjectFS_ListShallow(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("root/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "file.txt", strings.NewReader("top"))
	_ = fs.Put(ctx, "sub/a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "sub/deep/b.txt", strings.NewReader("b"))

	result, err := fs.List(ctx, "")
	if err != nil {
		t.Fatalf("List root: %v", err)
	}

	var files, dirs []string
	for _, f := range result.Files {
		if f.IsDir {
			dirs = append(dirs, f.Key)
		} else {
			files = append(files, f.Key)
		}
	}
	if len(files) != 1 || files[0] != "file.txt" {
		t.Fatalf("files = %v, want [file.txt]", files)
	}
	if len(dirs) != 1 || dirs[0] != "sub" {
		t.Fatalf("dirs = %v, want [sub]", dirs)
	}
}

func TestObjectFS_BatchDelete(t *testing.T) {
	fake := newFakeS3()
	fs := NewObjectFS(fake, "bucket", WithPrefix("pfx/"))
	ctx := context.Background()

	_ = fs.Put(ctx, "a.txt", strings.NewReader("a"))
	_ = fs.Put(ctx, "b.txt", strings.NewReader("b"))
	_ = fs.Put(ctx, "c.txt", strings.NewReader("c"))

	if err := BatchDelete(ctx, fs, []string{"a.txt", "b.txt"}); err != nil {
		t.Fatalf("BatchDelete: %v", err)
	}

	if _, ok := fake.objects["pfx/a.txt"]; ok {
		t.Fatal("a.txt should be deleted")
	}
	if _, ok := fake.objects["pfx/b.txt"]; ok {
		t.Fatal("b.txt should be deleted")
	}
	if _, ok := fake.objects["pfx/c.txt"]; !ok {
		t.Fatal("c.txt should still exist")
	}
}

func TestBatchDelete_Fallback(t *testing.T) {
	dir := t.TempDir()
	fs := mustNewLocalFS(t, dir)
	ctx := context.Background()

	_ = fs.Put(ctx, "x.txt", strings.NewReader("x"))
	_ = fs.Put(ctx, "y.txt", strings.NewReader("y"))

	if err := BatchDelete(ctx, fs, []string{"x.txt", "y.txt"}); err != nil {
		t.Fatalf("BatchDelete fallback: %v", err)
	}

	ok, _ := Exists(ctx, fs, "x.txt")
	if ok {
		t.Fatal("x.txt should be deleted")
	}
	ok, _ = Exists(ctx, fs, "y.txt")
	if ok {
		t.Fatal("y.txt should be deleted")
	}
}
