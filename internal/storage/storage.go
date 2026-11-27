// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

// Package storage is the byte layer of the file system: it maps an opaque key
// to bytes and back, and nothing else. It knows no file model, no owner, no
// permission and no URL policy. Those live one layer up.
//
// The whole contract is Store. A driver that cannot delete, cannot seek or
// cannot sign says so through an optional interface (Presigner, RangeOpener) or
// a sentinel error (ErrUnsupported); it never returns a nil error it did not
// earn, and it never grows a capability field that the caller must remember to
// check.
package storage

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

// Driver names. They are persisted in storages.driver, so they are a wire
// format: adding one is cheap, renaming one is a migration.
const (
	DriverLocal = "local"
	DriverS3    = "s3"
	DriverMem   = "mem" // tests only
)

var (
	// ErrNotFound is the only "missing" signal a driver may return. Callers
	// map it to 404 and nothing else.
	ErrNotFound = errors.New("storage: object not found")
	// ErrConflict reports a write that would overwrite an existing object.
	ErrConflict = errors.New("storage: object already exists")
	// ErrUnsupported reports a capability the driver does not implement.
	ErrUnsupported = errors.New("storage: operation unsupported")
	// ErrInvalidKey rejects a key that did not come from KeyFor.
	ErrInvalidKey = errors.New("storage: invalid key")
	// ErrInvalidFileID rejects a file id that is not a UUIDv7.
	ErrInvalidFileID = errors.New("storage: file id must be a UUIDv7")
	// ErrSizeMismatch reports a body that did not match the declared size.
	ErrSizeMismatch = errors.New("storage: body size does not match declaration")
	// ErrUnknownStorage reports a storage id with no registered driver.
	ErrUnknownStorage = errors.New("storage: unknown storage id")
)

// PutResult is what a successful Put observed while streaming, so the caller
// never has to read the body twice to learn its digest.
type PutResult struct {
	Size        int64
	SHA256      string
	ContentType string
}

// BlobInfo is the metadata a driver can report without reading the body.
type BlobInfo struct {
	Size        int64
	ContentType string
	ModTime     time.Time
}

// Store is the complete byte contract: four methods, no class, no visibility,
// no URL. Everything a caller needs to address an object is its key.
type Store interface {
	Put(ctx context.Context, key string, r io.Reader, size int64) (PutResult, error)
	Open(ctx context.Context, key string) (io.ReadCloser, BlobInfo, error)
	Stat(ctx context.Context, key string) (BlobInfo, error)
	Delete(ctx context.Context, key string) error
}

// Presigner is the optional capability of handing out a short-lived direct
// URL. Private content is never signed; the gateway only calls it for
// access=public rows.
type Presigner interface {
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
	// PresignPut signs the declared Content-Type and size into the request and
	// returns the headers the client must send verbatim.
	PresignPut(ctx context.Context, key string, ttl time.Duration, contentType string, size int64) (string, map[string]string, error)
}

// RangeOpener is the optional capability of reading a byte range without
// buffering the object. Local files satisfy the same need through io.Seeker;
// object stores can only do it with a ranged request.
type RangeOpener interface {
	OpenRange(ctx context.Context, key string, off, n int64) (io.ReadCloser, error)
}

// KeyFor is the single place a key is ever built. The layout is
// <year>/<month>/<uuidv7><ext>: the prefix is derived from the id itself, so
// keys sort by creation time without a clock read, and two uploads can never
// collide because the id is unique.
func KeyFor(fileID, ext string) (string, error) {
	created, err := timeOfV7(fileID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d/%02d/%s%s", created.Year(), int(created.Month()), fileID, NormalizeExt(ext)), nil
}

// NormalizeExt keeps a human-readable suffix. It never routes anything: the
// content type comes from sniffing, never from the extension.
func NormalizeExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	ext = strings.TrimPrefix(ext, ".")
	if len(ext) == 0 || len(ext) > 8 {
		return ""
	}
	for _, r := range ext {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') {
			return ""
		}
	}
	return "." + ext
}

// timeOfV7 extracts the embedded millisecond timestamp of a UUIDv7 and
// validates the version nibble. Anything else is rejected: converting a v4 id
// into a key would scatter files into the wrong year forever.
func timeOfV7(id string) (time.Time, error) {
	clean := strings.ToLower(strings.TrimSpace(id))
	clean = strings.ReplaceAll(clean, "-", "")
	if len(clean) != 32 {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidFileID, id)
	}
	raw, err := hex.DecodeString(clean)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidFileID, id)
	}
	if raw[6]>>4 != 7 {
		return time.Time{}, fmt.Errorf("%w: %q", ErrInvalidFileID, id)
	}
	ms := int64(raw[0])<<40 | int64(raw[1])<<32 | int64(raw[2])<<24 |
		int64(raw[3])<<16 | int64(raw[4])<<8 | int64(raw[5])
	return time.UnixMilli(ms).UTC(), nil
}
