// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package storage

import "github.com/lin-snow/ech0/pkg/virefs"

// mediaRoute is one top-level category directory under the local storage root.
// Stored keys always live under one of these prefixes, so the same list drives
// both key resolution and the unauthenticated /api/files allowlist.
type mediaRoute struct {
	prefix string
	exts   []string
}

var mediaRoutes = []mediaRoute{
	{"images/", []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".avif"}},
	{"audios/", []string{".mp3", ".flac", ".wav", ".m4a", ".ogg"}},
	{"videos/", []string{".mp4", ".avi", ".mkv", ".webm"}},
	{"documents/", []string{".pdf", ".doc", ".docx"}},
	{"files/", nil}, // DefaultRoute, must stay last.
}

func NewFileSchema() *virefs.Schema {
	routes := make([]virefs.Route, 0, len(mediaRoutes))
	for _, route := range mediaRoutes {
		if len(route.exts) == 0 {
			routes = append(routes, virefs.DefaultRoute(route.prefix))
			continue
		}
		routes = append(routes, virefs.RouteByExt(route.prefix, route.exts...))
	}
	return virefs.NewSchema(routes...)
}

// PublicPathPrefixes lists the only storage paths reachable without
// authentication through /api/files.
func PublicPathPrefixes() []string {
	prefixes := make([]string, 0, len(mediaRoutes))
	for _, route := range mediaRoutes {
		prefixes = append(prefixes, route.prefix)
	}
	return prefixes
}
