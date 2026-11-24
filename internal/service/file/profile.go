// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/lin-snow/ech0/internal/config"
	fileModel "github.com/lin-snow/ech0/internal/model/file"
)

// Profile is a creation-time contract between a route and the file system. It
// is never stored: it decides the access rule, the byte limit, the allowlist
// and whether a presigned upload is allowed, once, where the bytes arrive.
type Profile struct {
	Name        string
	MaxBytes    int64
	Access      fileModel.Access
	ImageOnly   bool
	AllowDirect bool
}

const (
	profileAvatar     = "avatar"
	profileInstance   = "instance"
	profileAttachment = "attachment"
	profileDraft      = "draft"

	avatarMaxBytes   = 2 << 20
	instanceMaxBytes = 512 << 10
)

// profileFor resolves the profile of a request. An unknown name is a client
// error; an empty name means attachment, the only profile whose access rule
// may be attached to any owner.
func profileFor(name string) (Profile, error) {
	switch strings.TrimSpace(name) {
	case "", profileAttachment:
		return Profile{Name: profileAttachment, Access: fileModel.AccessOwner, AllowDirect: true}, nil
	case profileDraft:
		return Profile{Name: profileDraft, Access: fileModel.AccessPrivate}, nil
	case profileAvatar:
		return Profile{Name: profileAvatar, MaxBytes: avatarMaxBytes, Access: fileModel.AccessPublic, ImageOnly: true}, nil
	case profileInstance:
		return Profile{Name: profileInstance, MaxBytes: instanceMaxBytes, Access: fileModel.AccessPublic, ImageOnly: true}, nil
	default:
		return Profile{}, ErrInvalid
	}
}

// Kind reports how the profile stores bytes. Instance assets are tiny and must
// survive a storage outage, so they live in the database; everything else is
// written to the default store.
func (p Profile) Kind() fileModel.Kind {
	if p.Name == profileInstance {
		return fileModel.KindBlob
	}
	return fileModel.KindOwned
}

// limit is the byte ceiling of a body, given its media category.
func (p Profile) limit(category string) int64 {
	if p.MaxBytes > 0 {
		return p.MaxBytes
	}
	upload := config.Config().Upload
	switch category {
	case "audio":
		return int64(upload.AudioMaxSize)
	case "video":
		return int64(upload.VideoMaxSize)
	default:
		return int64(upload.ImageMaxSize)
	}
}

// allows reports whether a content type may enter the profile.
func (p Profile) allows(contentType string) bool {
	if p.ImageOnly {
		return strings.HasPrefix(contentType, "image/") && contentType != "image/svg+xml"
	}
	return slices.Contains(config.Config().Upload.AllowedTypes, contentType)
}

// upload is one inspected body: what it claims to be, what it was sniffed as,
// and how it must be stored.
type upload struct {
	Filename    string
	Size        int64
	ContentType string
	Width       int
	Height      int
	fh          *multipart.FileHeader
}

// inspect validates one multipart body against a profile and reports what it
// is. It reads the header bytes twice (sniff + extension check) and leaves the
// body seekable, so the caller never buffers a large file to learn its type.
func inspect(fh *multipart.FileHeader, prof Profile) (upload, error) {
	if fh == nil {
		return upload{}, ErrInvalid
	}
	filename := filepath.Base(strings.TrimSpace(fh.Filename))
	if filename == "" || filename == "." || filename == string(filepath.Separator) {
		return upload{}, ErrInvalid
	}

	file, err := fh.Open()
	if err != nil {
		return upload{}, err
	}
	defer func() { _ = file.Close() }()

	detected, err := detectContentType(file)
	if err != nil {
		return upload{}, err
	}
	contentType := resolveContentType(filename, detected)
	if !prof.allows(contentType) {
		return upload{}, ErrTypeNotAllowed
	}
	if err := validateFileUpload(filename, detected); err != nil {
		return upload{}, ErrTypeNotAllowed
	}

	out := upload{Filename: filename, Size: fh.Size, ContentType: contentType, fh: fh}
	if strings.HasPrefix(contentType, "image/") {
		width, height, sizeErr := imgSize(fh)
		if sizeErr == nil {
			out.Width, out.Height = width, height
		}
	}
	return out, nil
}

// detectContentType sniffs the first block and rewinds, so the same handle can
// be streamed afterwards.
func detectContentType(r io.ReadSeeker) (string, error) {
	buf := make([]byte, 512)
	n, err := io.ReadFull(r, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

// validateFileUpload pairs the extension with the sniffed type. The extension
// is not routing — keys carry it for human readability only — but a mismatch
// means the client lied about the file, which the allowlist above already
// trusted once.
func validateFileUpload(filename, detected string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, blocked := dangerousExtensions[ext]; blocked {
		return ErrTypeNotAllowed
	}
	expected, ok := safeExtMIME[ext]
	if !ok {
		return ErrTypeNotAllowed
	}
	if _, exec := executableMIMEs[detected]; exec {
		return ErrTypeNotAllowed
	}
	if detected == "application/octet-stream" || slices.Contains(expected, detected) {
		return nil
	}
	return ErrTypeNotAllowed
}

func resolveContentType(filename, detected string) string {
	if detected != "" && detected != "application/octet-stream" {
		return detected
	}
	if mimes, ok := safeExtMIME[strings.ToLower(filepath.Ext(filename))]; ok && len(mimes) > 0 {
		return mimes[0]
	}
	return "application/octet-stream"
}

var dangerousExtensions = map[string]struct{}{
	".html": {}, ".htm": {}, ".xhtml": {},
	".svg": {},
	".xml": {}, ".xsl": {}, ".xslt": {},
	".js": {}, ".mjs": {}, ".cjs": {},
	".php": {}, ".asp": {}, ".aspx": {}, ".jsp": {},
}

var safeExtMIME = map[string][]string{
	".jpg":  {"image/jpeg"},
	".jpeg": {"image/jpeg"},
	".png":  {"image/png"},
	".gif":  {"image/gif"},
	".webp": {"image/webp"},
	".avif": {"image/avif"},
	".mp3":  {"audio/mpeg"},
	".flac": {"audio/flac"},
	".wav":  {"audio/wav", "audio/x-wav"},
	".m4a":  {"audio/mp4", "audio/x-m4a"},
	".ogg":  {"audio/ogg"},
	".mp4":  {"video/mp4"},
}

var executableMIMEs = map[string]struct{}{
	"text/html":               {},
	"text/xml":                {},
	"application/xhtml+xml":   {},
	"application/xml":         {},
	"image/svg+xml":           {},
	"application/javascript":  {},
	"text/javascript":         {},
	"application/x-httpd-php": {},
}
