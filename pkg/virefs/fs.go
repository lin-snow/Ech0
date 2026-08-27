// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	ErrNotFound     = errors.New("virefs: not found")
	ErrInvalidKey   = errors.New("virefs: invalid key")
	ErrAlreadyExist = errors.New("virefs: already exists")
	ErrNotSupported = errors.New("virefs: operation not supported")
	ErrPermission   = errors.New("virefs: permission denied")
)

type FileInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
	IsDir        bool
	ContentType  string
}

type ListResult struct {
	Files []FileInfo
}

type FS interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)

	Put(ctx context.Context, key string, r io.Reader, opts ...PutOption) error

	Delete(ctx context.Context, key string) error

	List(ctx context.Context, prefix string) (*ListResult, error)

	Stat(ctx context.Context, key string) (*FileInfo, error)

	Access(ctx context.Context, key string) (*AccessInfo, error)

	Exists(ctx context.Context, key string) (bool, error)
}

type PutOption func(*PutConfig)

type PutConfig struct {
	ContentType string
	Metadata    map[string]string
}

func BuildPutConfig(opts []PutOption) PutConfig {
	var c PutConfig
	for _, o := range opts {
		o(&c)
	}
	return c
}

func WithContentType(ct string) PutOption {
	return func(c *PutConfig) { c.ContentType = ct }
}

func WithMetadata(m map[string]string) PutOption {
	return func(c *PutConfig) { c.Metadata = m }
}

func Exists(ctx context.Context, fs FS, key string) (bool, error) {
	return fs.Exists(ctx, key)
}

type Copier interface {
	Copy(ctx context.Context, srcKey, dstKey string) error
}

type BatchDeleter interface {
	BatchDelete(ctx context.Context, keys []string) error
}

func BatchDelete(ctx context.Context, fsys FS, keys []string) error {
	if bd, ok := fsys.(BatchDeleter); ok {
		return bd.BatchDelete(ctx, keys)
	}
	for _, key := range keys {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := fsys.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func Copy(ctx context.Context, src FS, srcKey string, dst FS, dstKey string, opts ...PutOption) error {
	if src == dst {
		if c, ok := src.(Copier); ok {
			return c.Copy(ctx, srcKey, dstKey)
		}
	}
	rc, err := src.Get(ctx, srcKey)
	if err != nil {
		return fmt.Errorf("copy: get %q: %w", srcKey, err)
	}
	defer rc.Close()
	return dst.Put(ctx, dstKey, rc, opts...)
}

type AccessInfo struct {
	Path string
	URL  string
}

type PresignedRequest struct {
	URL    string
	Method string
	Header http.Header
}

type Presigner interface {
	PresignGet(ctx context.Context, key string, expires time.Duration) (*PresignedRequest, error)

	PresignPut(ctx context.Context, key string, expires time.Duration) (*PresignedRequest, error)
}

type KeyFunc func(key string) string

type AccessFunc func(key string) *AccessInfo

type OpError struct {
	Op  string
	Key string
	Err error
}

func (e *OpError) Error() string {
	return fmt.Sprintf("virefs %s %q: %v", e.Op, e.Key, e.Err)
}

func (e *OpError) Unwrap() error { return e.Err }
