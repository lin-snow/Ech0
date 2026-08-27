// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package virefs

import (
	"context"
	"io"
)

type Hooks struct {
	WrapGet func(key string, rc io.ReadCloser) io.ReadCloser

	WrapPut func(key string, r io.Reader) io.Reader

	AfterStat func(key string, info *FileInfo)

	OnDelete func(key string)
}

func WithHooks(inner FS, hooks Hooks) *hookFS {
	return &hookFS{inner: inner, hooks: hooks}
}

type hookFS struct {
	inner FS
	hooks Hooks
}

func (h *hookFS) Unwrap() FS { return h.inner }

func (h *hookFS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	rc, err := h.inner.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if h.hooks.WrapGet != nil {
		rc = h.hooks.WrapGet(key, rc)
	}
	return rc, nil
}

func (h *hookFS) Put(ctx context.Context, key string, r io.Reader, opts ...PutOption) error {
	if h.hooks.WrapPut != nil {
		r = h.hooks.WrapPut(key, r)
	}
	return h.inner.Put(ctx, key, r, opts...)
}

func (h *hookFS) Delete(ctx context.Context, key string) error {
	err := h.inner.Delete(ctx, key)
	if err != nil {
		return err
	}
	if h.hooks.OnDelete != nil {
		h.hooks.OnDelete(key)
	}
	return nil
}

func (h *hookFS) List(ctx context.Context, prefix string) (*ListResult, error) {
	return h.inner.List(ctx, prefix)
}

func (h *hookFS) Stat(ctx context.Context, key string) (*FileInfo, error) {
	info, err := h.inner.Stat(ctx, key)
	if err != nil {
		return nil, err
	}
	if h.hooks.AfterStat != nil {
		h.hooks.AfterStat(key, info)
	}
	return info, nil
}

func (h *hookFS) Exists(ctx context.Context, key string) (bool, error) {
	return h.inner.Exists(ctx, key)
}

func (h *hookFS) Access(ctx context.Context, key string) (*AccessInfo, error) {
	return h.inner.Access(ctx, key)
}

var _ FS = (*hookFS)(nil)

type Middleware func(FS) FS

func Chain(fs FS, mw ...Middleware) FS {
	for _, m := range mw {
		fs = m(fs)
	}
	return fs
}

type BaseFS struct{ Inner FS }

func (b BaseFS) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	return b.Inner.Get(ctx, key)
}

func (b BaseFS) Put(ctx context.Context, key string, r io.Reader, opts ...PutOption) error {
	return b.Inner.Put(ctx, key, r, opts...)
}

func (b BaseFS) Delete(ctx context.Context, key string) error {
	return b.Inner.Delete(ctx, key)
}

func (b BaseFS) List(ctx context.Context, prefix string) (*ListResult, error) {
	return b.Inner.List(ctx, prefix)
}

func (b BaseFS) Stat(ctx context.Context, key string) (*FileInfo, error) {
	return b.Inner.Stat(ctx, key)
}

func (b BaseFS) Access(ctx context.Context, key string) (*AccessInfo, error) {
	return b.Inner.Access(ctx, key)
}

func (b BaseFS) Exists(ctx context.Context, key string) (bool, error) {
	return b.Inner.Exists(ctx, key)
}

var _ FS = BaseFS{}
