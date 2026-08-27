// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package util

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
)

func GetImageSizeFromPath(path string) (width, height int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = f.Close()
	}()

	return GetImageSizeFromReader(f)
}

func GetImageSizeFromFile(file *multipart.FileHeader) (width, height int, err error) {
	reader, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = reader.Close()
	}()
	return GetImageSizeFromReader(reader)
}

func GetImageSizeFromReader(reader io.Reader) (width, height int, err error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return 0, 0, err
	}
	if len(data) == 0 {
		return 0, 0, fmt.Errorf("empty image data")
	}

	if cfg, _, stdErr := image.DecodeConfig(bytes.NewReader(data)); stdErr == nil {
		return cfg.Width, cfg.Height, nil
	}

	return 0, 0, nil
}
