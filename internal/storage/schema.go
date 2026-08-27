// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import "github.com/lin-snow/ech0/pkg/virefs"

func NewFileSchema() *virefs.Schema {
	return virefs.NewSchema(
		virefs.RouteByExt("images/", ".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif"),
		virefs.RouteByExt("audios/", ".mp3", ".flac", ".wav", ".m4a", ".ogg"),
		virefs.RouteByExt("videos/", ".mp4", ".avi", ".mkv", ".webm"),
		virefs.RouteByExt("documents/", ".pdf", ".doc", ".docx"),
		virefs.DefaultRoute("files/"),
	)
}
