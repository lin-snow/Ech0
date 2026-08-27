// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func encodePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}
	return buf.Bytes()
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestGetImageSizeFromReader(t *testing.T) {
	t.Run("valid PNG returns dimensions", func(t *testing.T) {
		data := encodePNG(t, 7, 13)
		w, h, err := GetImageSizeFromReader(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w != 7 || h != 13 {
			t.Errorf("got (%d,%d), want (7,13)", w, h)
		}
	})

	t.Run("empty data returns error", func(t *testing.T) {
		w, h, err := GetImageSizeFromReader(bytes.NewReader(nil))
		if err == nil {
			t.Fatalf("expected error for empty data, got nil")
		}
		if w != 0 || h != 0 {
			t.Errorf("got (%d,%d), want (0,0)", w, h)
		}
	})

	t.Run("unknown format silently falls back to (0,0,nil)", func(t *testing.T) {
		w, h, err := GetImageSizeFromReader(bytes.NewReader([]byte("this is plainly not an image payload")))
		if err != nil {
			t.Fatalf("expected nil error on unknown format, got %v", err)
		}
		if w != 0 || h != 0 {
			t.Errorf("got (%d,%d), want (0,0)", w, h)
		}
	})

	t.Run("reader error is propagated", func(t *testing.T) {
		w, h, err := GetImageSizeFromReader(errReader{})
		if err == nil {
			t.Fatalf("expected error from failing reader, got nil")
		}
		if w != 0 || h != 0 {
			t.Errorf("got (%d,%d), want (0,0)", w, h)
		}
	})
}

func TestGetImageSizeFromPath(t *testing.T) {
	t.Run("valid PNG file returns dimensions", func(t *testing.T) {
		data := encodePNG(t, 20, 11)
		path := filepath.Join(t.TempDir(), "sample.png")
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatalf("write temp png: %v", err)
		}

		w, h, err := GetImageSizeFromPath(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w != 20 || h != 11 {
			t.Errorf("got (%d,%d), want (20,11)", w, h)
		}
	})

	t.Run("missing file returns error", func(t *testing.T) {
		w, h, err := GetImageSizeFromPath(filepath.Join(t.TempDir(), "does-not-exist.png"))
		if err == nil {
			t.Fatalf("expected error for missing file, got nil")
		}
		if w != 0 || h != 0 {
			t.Errorf("got (%d,%d), want (0,0)", w, h)
		}
	})
}

func TestGetImageSizeFromFile(t *testing.T) {
	data := encodePNG(t, 5, 9)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("file", "upload.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("write part: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(&body, mw.Boundary())
	form, err := reader.ReadForm(int64(len(data) + 1024))
	if err != nil {
		t.Fatalf("ReadForm: %v", err)
	}
	headers := form.File["file"]
	if len(headers) == 0 {
		t.Fatalf("no file header parsed")
	}

	w, h, err := GetImageSizeFromFile(headers[0])
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w != 5 || h != 9 {
		t.Errorf("got (%d,%d), want (5,9)", w, h)
	}
}
