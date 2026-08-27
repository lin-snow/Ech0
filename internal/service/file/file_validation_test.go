// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package service

import (
	"bytes"
	"io"
	"mime/multipart"
	"testing"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMultipartFile struct {
	*bytes.Reader
}

func (fakeMultipartFile) Close() error { return nil }

func newFakeFile(b []byte) multipart.File {
	return fakeMultipartFile{bytes.NewReader(b)}
}

var pngMagic = []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0}

var htmlBody = []byte("<!DOCTYPE html><html><body>hi</body></html>")

func TestValidateFileUpload(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		detectedMIME string
		allowed      []string
		wantErr      bool
	}{
		{
			name:         "legal png passes",
			filename:     "photo.png",
			detectedMIME: "image/png",
			allowed:      []string{"image/png", "image/jpeg"},
			wantErr:      false,
		},
		{
			name:         "octet-stream sniff is accepted for whitelisted ext (avif)",
			filename:     "pic.avif",
			detectedMIME: "application/octet-stream",
			allowed:      []string{"image/avif"},
			wantErr:      false,
		},
		{
			name:         "wav alternate mime matches one of expected set",
			filename:     "clip.wav",
			detectedMIME: "audio/x-wav",
			allowed:      []string{"audio/wav"},
			wantErr:      false,
		},
		{
			name:         "legal mp4 video passes",
			filename:     "clip.mp4",
			detectedMIME: "video/mp4",
			allowed:      []string{"video/mp4"},
			wantErr:      false,
		},
		{
			name:         "mp4 with octet-stream sniff is accepted for whitelisted ext",
			filename:     "clip.mp4",
			detectedMIME: "application/octet-stream",
			allowed:      []string{"video/mp4"},
			wantErr:      false,
		},
		{
			name:         "uppercase extension is normalized",
			filename:     "PHOTO.PNG",
			detectedMIME: "image/png",
			allowed:      []string{"image/png"},
			wantErr:      false,
		},
		{
			name:         "surrounding whitespace in filename is trimmed",
			filename:     "  photo.png  ",
			detectedMIME: "image/png",
			allowed:      []string{"image/png"},
			wantErr:      false,
		},

		{
			name:         "html extension rejected",
			filename:     "evil.html",
			detectedMIME: "image/png",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},
		{
			name:         "svg extension rejected",
			filename:     "logo.svg",
			detectedMIME: "application/octet-stream",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},
		{
			name:         "js extension rejected",
			filename:     "payload.js",
			detectedMIME: "application/octet-stream",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},
		{
			name:         "php extension rejected",
			filename:     "shell.php",
			detectedMIME: "application/octet-stream",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},

		{
			name:         "unknown extension not whitelisted",
			filename:     "doc.pdf",
			detectedMIME: "application/pdf",
			allowed:      []string{"application/pdf"},
			wantErr:      true,
		},
		{
			name:         "no extension rejected",
			filename:     "noext",
			detectedMIME: "image/png",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},

		{
			name:         "svg content disguised as png rejected via magic-byte gate",
			filename:     "fake.png",
			detectedMIME: "image/svg+xml",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},
		{
			name:         "html content disguised as png rejected (mime mismatch)",
			filename:     "fake.png",
			detectedMIME: "text/html",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},

		{
			name:         "gif content under png extension rejected",
			filename:     "photo.png",
			detectedMIME: "image/gif",
			allowed:      []string{"image/png", "image/gif"},
			wantErr:      true,
		},

		{
			name:         "extension valid but mime not on admin allowlist",
			filename:     "photo.png",
			detectedMIME: "image/png",
			allowed:      []string{"image/gif"},
			wantErr:      true,
		},
		{
			name:         "mp4 rejected when video/mp4 not on allowlist",
			filename:     "clip.mp4",
			detectedMIME: "video/mp4",
			allowed:      []string{"image/png"},
			wantErr:      true,
		},
		{
			name:         "html content disguised as mp4 rejected via magic-byte gate",
			filename:     "fake.mp4",
			detectedMIME: "text/html",
			allowed:      []string{"video/mp4"},
			wantErr:      true,
		},
		{
			name:         "empty allowlist rejects everything",
			filename:     "photo.png",
			detectedMIME: "image/png",
			allowed:      nil,
			wantErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFileUpload(tc.filename, tc.detectedMIME, tc.allowed)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, commonModel.FILE_TYPE_NOT_ALLOWED, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateFileUploadByName(t *testing.T) {
	cases := []struct {
		name         string
		filename     string
		declaredMIME string
		allowed      []string
		wantErr      bool
	}{
		{
			name:         "legal png with matching declared mime",
			filename:     "photo.png",
			declaredMIME: "image/png",
			allowed:      []string{"image/png"},
			wantErr:      false,
		},
		{
			name:         "wav with alternate declared mime",
			filename:     "clip.wav",
			declaredMIME: "audio/x-wav",
			allowed:      []string{"audio/x-wav"},
			wantErr:      false,
		},
		{
			name:         "dangerous js extension rejected",
			filename:     "payload.js",
			declaredMIME: "text/javascript",
			allowed:      []string{"text/javascript"},
			wantErr:      true,
		},
		{
			name:         "unknown extension rejected",
			filename:     "doc.pdf",
			declaredMIME: "application/pdf",
			allowed:      []string{"application/pdf"},
			wantErr:      true,
		},
		{
			name:         "declared mime not on allowlist rejected",
			filename:     "photo.png",
			declaredMIME: "image/png",
			allowed:      []string{"image/gif"},
			wantErr:      true,
		},
		{
			name:         "declared mime allowed but mismatches extension (spoof) rejected",
			filename:     "photo.png",
			declaredMIME: "image/gif",
			allowed:      []string{"image/gif", "image/png"},
			wantErr:      true,
		},
		{
			name:         "octet-stream is not accepted by name-only validation",
			filename:     "pic.avif",
			declaredMIME: "application/octet-stream",
			allowed:      []string{"application/octet-stream", "image/avif"},
			wantErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateFileUploadByName(tc.filename, tc.declaredMIME, tc.allowed)
			if tc.wantErr {
				require.Error(t, err)
				assert.Equal(t, commonModel.FILE_TYPE_NOT_ALLOWED, err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestCanonicalMIMEForExt(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"png", "photo.png", "image/png"},
		{"jpeg alias", "photo.jpeg", "image/jpeg"},
		{"uppercase normalized", "PHOTO.JPG", "image/jpeg"},
		{"wav returns first listed", "clip.wav", "audio/wav"},
		{"trim whitespace", "  photo.png  ", "image/png"},
		{"unknown extension", "doc.pdf", ""},
		{"no extension", "noext", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, canonicalMIMEForExt(tc.in))
		})
	}
}

func TestResolveContentType(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		detected string
		want     string
	}{
		{"specific detected wins", "photo.png", "image/png", "image/png"},
		{"octet-stream falls back to canonical ext", "pic.avif", "application/octet-stream", "image/avif"},
		{"empty detected falls back to canonical ext", "photo.png", "", "image/png"},
		{"octet-stream with unknown ext keeps detected", "doc.pdf", "application/octet-stream", "application/octet-stream"},
		{"empty detected unknown ext keeps detected", "doc.xyz", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, resolveContentType(tc.filename, tc.detected))
		})
	}
}

func TestDetectContentType(t *testing.T) {
	t.Run("sniffs png and rewinds", func(t *testing.T) {
		f := newFakeFile(pngMagic)
		got, err := detectContentType(f)
		require.NoError(t, err)
		assert.Equal(t, "image/png", got)

		rest, err := io.ReadAll(f)
		require.NoError(t, err)
		assert.Equal(t, pngMagic, rest)
	})

	t.Run("sniffs html content", func(t *testing.T) {
		f := newFakeFile(htmlBody)
		got, err := detectContentType(f)
		require.NoError(t, err)
		assert.Equal(t, "text/html; charset=utf-8", got)
	})

	t.Run("empty file defaults to text/plain", func(t *testing.T) {
		f := newFakeFile(nil)
		got, err := detectContentType(f)
		require.NoError(t, err)
		assert.Equal(t, "text/plain; charset=utf-8", got)
	})
}
