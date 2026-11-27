// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/lin-snow/ech0/pkg/virefs"
)

// fsStore adapts a virefs.FS to the Store contract. Both drivers share it, so
// hashing, size enforcement and error translation have exactly one
// implementation.
type fsStore struct {
	fs virefs.FS
}

func newFSStore(fs virefs.FS) *fsStore { return &fsStore{fs: fs} }

var _ Store = (*fsStore)(nil)

// Put streams the body once: the same pass feeds the hash, counts the bytes and
// feeds the driver. A body whose length differs from the declaration is
// rejected after the fact and its partial object is removed, because a stored
// object of unknown length can never be served with a correct Content-Length.
func (s *fsStore) Put(ctx context.Context, key string, r io.Reader, size int64) (PutResult, error) {
	if size < 0 {
		return PutResult{}, ErrSizeMismatch
	}
	hash := sha256.New()
	counter := &countingReader{r: io.TeeReader(r, hash)}

	prefix := make([]byte, 512)
	n, err := io.ReadFull(counter, prefix)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return PutResult{}, err
	}
	prefix = prefix[:n]
	contentType := ContentTypeOf(prefix)
	body := io.MultiReader(bytes.NewReader(prefix), counter)

	putErr := s.fs.Put(ctx, key, body, virefs.WithContentType(contentType))
	if putErr != nil {
		return PutResult{}, translate(putErr)
	}
	if counter.n != size {
		_ = s.fs.Delete(ctx, key)
		return PutResult{}, ErrSizeMismatch
	}
	return PutResult{
		Size:        counter.n,
		SHA256:      hex.EncodeToString(hash.Sum(nil)),
		ContentType: contentType,
	}, nil
}

func (s *fsStore) Open(ctx context.Context, key string) (io.ReadCloser, BlobInfo, error) {
	info, err := s.Stat(ctx, key)
	if err != nil {
		return nil, BlobInfo{}, err
	}
	rc, err := s.fs.Get(ctx, key)
	if err != nil {
		return nil, BlobInfo{}, translate(err)
	}
	return rc, info, nil
}

func (s *fsStore) Stat(ctx context.Context, key string) (BlobInfo, error) {
	info, err := s.fs.Stat(ctx, key)
	if err != nil {
		return BlobInfo{}, translate(err)
	}
	return BlobInfo{Size: info.Size, ContentType: info.ContentType, ModTime: info.LastModified}, nil
}

func (s *fsStore) Delete(ctx context.Context, key string) error {
	if err := s.fs.Delete(ctx, key); err != nil {
		if errors.Is(err, virefs.ErrNotFound) {
			return nil // deleting a missing object is already the goal state
		}
		return translate(err)
	}
	return nil
}

// PresignGet signs a public read. Only reachable for access=public rows.
func (s *fsStore) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	presigner, ok := s.fs.(virefs.Presigner)
	if !ok {
		return "", ErrUnsupported
	}
	req, err := presigner.PresignGet(ctx, key, ttl)
	if err != nil {
		return "", translate(err)
	}
	return req.URL, nil
}

// PresignPut signs a direct upload, including the declared Content-Type, so the
// object store rejects a body whose type differs from the declaration.
func (s *fsStore) PresignPut(ctx context.Context, key string, ttl time.Duration, contentType string, size int64) (string, map[string]string, error) {
	signer, ok := s.fs.(contentPresigner)
	if !ok {
		return "", nil, ErrUnsupported
	}
	req, err := signer.PresignPutContent(ctx, key, ttl, contentType, size)
	if err != nil {
		return "", nil, translate(err)
	}
	headers := map[string]string{"Content-Type": contentType}
	return req.URL, headers, nil
}

// OpenRange serves a byte range without buffering the object when the driver
// can seek. Object stores answer through a ranged request.
func (s *fsStore) OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error) {
	if ranged, ok := s.fs.(rangeOpener); ok {
		rc, err := ranged.OpenRange(ctx, key, off, n)
		if err != nil {
			return nil, translate(err)
		}
		return rc, nil
	}
	rc, err := s.fs.Get(ctx, key)
	if err != nil {
		return nil, translate(err)
	}
	seeker, ok := rc.(io.ReadSeeker)
	if !ok {
		_ = rc.Close()
		return nil, ErrUnsupported
	}
	if _, err := seeker.Seek(off, io.SeekStart); err != nil {
		_ = rc.Close()
		return nil, err
	}
	return &limitedReadCloser{Reader: io.LimitReader(seeker, n), Closer: rc}, nil
}

// contentPresigner is the virefs capability that binds a Content-Type into the
// signature. It is optional: drivers without it are simply not presignable.
type contentPresigner interface {
	PresignPutContent(ctx context.Context, key string, expires time.Duration, contentType string, size int64) (*virefs.PresignedRequest, error)
}

type rangeOpener interface {
	OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error)
}

// ContentTypeOf is the one sniffing rule shared by the file service and the
// drivers. A storage object's stored type and the row's type therefore cannot
// drift: they are produced by the same function from the same bytes.
func ContentTypeOf(prefix []byte) string {
	if len(prefix) == 0 {
		return "application/octet-stream"
	}
	return http.DetectContentType(prefix)
}

func translate(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, virefs.ErrNotFound):
		return ErrNotFound
	case errors.Is(err, virefs.ErrAlreadyExist):
		return ErrConflict
	case errors.Is(err, virefs.ErrNotSupported), errors.Is(err, virefs.ErrInvalidKey):
		return ErrUnsupported
	default:
		return err
	}
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}
